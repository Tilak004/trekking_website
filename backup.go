package main

import (
	"archive/tar"
	"compress/gzip"
	"database/sql"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Automatic backups. The public-site process makes one every night (and the
// admin can run "beyond.exe -backup-now" at any time). Each backup is a single
// .tar.gz holding a consistent snapshot of the database plus the photos people
// uploaded (trek-lead profiles, review photos), which aren't in git.
//
// Settings (environment variables):
//
//	BACKUP_DIR     where backups go             (default "backups"; "off" disables)
//	BACKUP_AT      daily time, 24h local clock  (default "02:00")
//	BACKUP_KEEP    how many backups to keep     (default 30)
//	BACKUP_UPLOAD  optional command run after each backup to copy it off the
//	               server; "{file}" is replaced by the backup's path, e.g.
//	               rclone copy {file} gdrive:beyond-backups

type backupConfig struct {
	dir, at, upload string
	keep            int
}

func loadBackupConfig() backupConfig {
	keep, err := strconv.Atoi(envOr("BACKUP_KEEP", "30"))
	if err != nil || keep < 1 {
		keep = 30
	}
	return backupConfig{
		dir:    envOr("BACKUP_DIR", "backups"),
		at:     envOr("BACKUP_AT", "02:00"),
		upload: os.Getenv("BACKUP_UPLOAD"),
		keep:   keep,
	}
}

// uploadDirs are the photo folders filled by website users (not committed to git).
var uploadDirs = []string{"leads", "reviews"}

// runBackup writes one backup archive and returns its path.
func runBackup(db *sql.DB, photosDir string, cfg backupConfig) (string, error) {
	if err := os.MkdirAll(cfg.dir, 0o755); err != nil {
		return "", err
	}
	stamp := time.Now().Format("2006-01-02_150405")
	snap := filepath.Join(cfg.dir, "snapshot-"+stamp+".db")
	os.Remove(snap)
	// VACUUM INTO writes a consistent copy even while the site is taking bookings.
	if _, err := db.Exec(`VACUUM INTO ?`, snap); err != nil {
		return "", fmt.Errorf("snapshot database: %w", err)
	}
	defer os.Remove(snap)

	out := filepath.Join(cfg.dir, "beyond-"+stamp+".tar.gz")
	tmp := out + ".partial"
	f, err := os.Create(tmp)
	if err != nil {
		return "", err
	}
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)
	err = addFile(tw, snap, "beyond.db")
	for _, d := range uploadDirs {
		if err != nil {
			break
		}
		err = addDir(tw, filepath.Join(photosDir, d), "photos/"+d)
	}
	if err == nil {
		err = tw.Close()
	}
	if err == nil {
		err = gz.Close()
	}
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		os.Remove(tmp)
		return "", err
	}
	if err := os.Rename(tmp, out); err != nil {
		return "", err
	}
	pruneBackups(cfg)
	if cfg.upload != "" {
		uploadBackup(cfg.upload, out)
	}
	return out, nil
}

func addFile(tw *tar.Writer, path, name string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return err
	}
	hdr, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}
	hdr.Name = name
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}
	_, err = io.Copy(tw, f)
	return err
}

func addDir(tw *tar.Writer, dir, prefix string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}
	return filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, err := filepath.Rel(dir, p)
		if err != nil {
			return err
		}
		return addFile(tw, p, prefix+"/"+filepath.ToSlash(rel))
	})
}

// pruneBackups keeps the newest cfg.keep archives.
func pruneBackups(cfg backupConfig) {
	files, _ := filepath.Glob(filepath.Join(cfg.dir, "beyond-*.tar.gz"))
	sort.Strings(files) // names sort by date
	for len(files) > cfg.keep {
		os.Remove(files[0])
		files = files[1:]
	}
}

func uploadBackup(command, file string) {
	cmdline := strings.ReplaceAll(command, "{file}", file)
	parts := strings.Fields(cmdline)
	if len(parts) == 0 {
		return
	}
	if out, err := exec.Command(parts[0], parts[1:]...).CombinedOutput(); err != nil {
		log.Printf("backup upload failed: %v: %s", err, strings.TrimSpace(string(out)))
		return
	}
	log.Printf("backup uploaded: %s", file)
}

// startBackups runs a backup every day at cfg.at, in the background.
func startBackups(db *sql.DB, photosDir string, cfg backupConfig) {
	if cfg.dir == "off" {
		log.Printf("automatic backups are off")
		return
	}
	at, err := time.Parse("15:04", cfg.at)
	if err != nil {
		log.Printf("BACKUP_AT %q is not HH:MM; using 02:00", cfg.at)
		at, _ = time.Parse("15:04", "02:00")
	}
	go func() {
		for {
			now := time.Now()
			next := time.Date(now.Year(), now.Month(), now.Day(), at.Hour(), at.Minute(), 0, 0, now.Location())
			if !next.After(now) {
				next = next.AddDate(0, 0, 1)
			}
			time.Sleep(time.Until(next))
			if path, err := runBackup(db, photosDir, cfg); err != nil {
				log.Printf("backup FAILED: %v", err)
			} else {
				log.Printf("backup written: %s", path)
			}
		}
	}()
	log.Printf("automatic backups: daily at %s into %s (keeping %d)", at.Format("15:04"), cfg.dir, cfg.keep)
}
