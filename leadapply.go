package main

import (
	"database/sql"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Trek-lead applications: any logged-in user can apply to lead treks from
// /lead-with-us. Applications stay hidden from the public until an admin
// approves them in Admin → Trek Leads.

const (
	LeadPending  = "pending"
	LeadApproved = "approved"
	LeadRejected = "rejected"
)

const maxPhotoBytes = 5 << 20

var photoExts = map[string]string{"image/jpeg": ".jpg", "image/png": ".png", "image/webp": ".webp"}

// myLead returns the lead profile owned by a user, if they have applied.
func (a *App) myLead(userID int64) (Lead, bool, error) {
	l, err := a.leadWhere(`l.user_id = ?`, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return Lead{}, false, nil
	}
	return l, err == nil, err
}

func (a *App) leadApplyPage(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{"Form": Lead{}, "HasLead": false}
	if u := currentUser(r); u != nil {
		l, ok, err := a.myLead(u.ID)
		if err != nil {
			a.serverError(w, err)
			return
		}
		if ok {
			data["Form"], data["HasLead"] = l, true
		} else {
			data["Form"] = Lead{Name: u.Name, Phone: u.Phone}
		}
	}
	a.render(w, r, "lead_apply.html", data)
}

func (a *App) leadApply(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	existing, has, err := a.myLead(u.ID)
	if err != nil {
		a.serverError(w, err)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxPhotoBytes+(1<<20))
	parseErr := r.ParseMultipartForm(maxPhotoBytes)

	f := Lead{
		ID:             existing.ID,
		Slug:           existing.Slug,
		Status:         existing.Status,
		Photo:          existing.Photo,
		Name:           trimTo(r.FormValue("name"), 100),
		Phone:          trimTo(r.FormValue("phone"), 30),
		City:           trimTo(r.FormValue("city"), 80),
		Headline:       trimTo(r.FormValue("headline"), 150),
		Bio:            trimTo(r.FormValue("bio"), 3000),
		Achievements:   strings.Join(splitLines(r.FormValue("achievements")), "\n"),
		Certifications: strings.Join(splitLines(r.FormValue("certifications")), "\n"),
	}
	fail := func(msg string) {
		a.render(w, r, "lead_apply.html", map[string]any{"Form": f, "HasLead": has, "Err": msg})
	}

	switch {
	case parseErr != nil:
		fail("That photo is too large. Please upload one under 5 MB.")
		return
	case f.Name == "" || f.Phone == "" || f.City == "" || f.Headline == "":
		fail("Please fill in your name, phone, city and headline.")
		return
	case len([]rune(f.Bio)) < 40:
		fail("Please write a few lines about yourself (at least 40 characters).")
		return
	case f.Achievements == "":
		fail("Please list the treks you've completed.")
		return
	case !has && r.FormValue("agree") != "on":
		fail("Please confirm your details are true and that you're 18 or older.")
		return
	}

	if file, _, err := r.FormFile("photo"); err == nil {
		defer file.Close()
		photo, err := a.saveLeadPhoto(file, f.Name)
		if err != nil {
			fail(err.Error())
			return
		}
		f.Photo = photo
	}

	if has {
		// Re-submitting after a rejection puts the application back in the queue.
		status := existing.Status
		if status == LeadRejected {
			status = LeadPending
		}
		_, err = a.db.Exec(`UPDATE leads SET name = ?, phone = ?, city = ?, headline = ?, bio = ?, achievements = ?,
certifications = ?, photo = ?, status = ? WHERE id = ?`,
			f.Name, f.Phone, f.City, f.Headline, f.Bio, f.Achievements, f.Certifications, f.Photo, status, existing.ID)
		if err != nil {
			a.serverError(w, err)
			return
		}
		redirect(w, r, "/lead-with-us?msg=lead_updated")
		return
	}

	slug, err := a.uniqueLeadSlug(f.Name)
	if err != nil {
		a.serverError(w, err)
		return
	}
	_, err = a.db.Exec(`INSERT INTO leads (slug, name, phone, city, headline, bio, achievements, certifications, photo,
user_id, status, active, applied_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'pending', 0, datetime('now'))`,
		slug, f.Name, f.Phone, f.City, f.Headline, f.Bio, f.Achievements, f.Certifications, f.Photo, u.ID)
	if err != nil {
		a.serverError(w, err)
		return
	}
	redirect(w, r, "/lead-with-us?msg=applied")
}

// saveLeadPhoto stores an uploaded profile photo under photos/leads after
// checking that the file really is an image.
func (a *App) saveLeadPhoto(file multipart.File, name string) (string, error) {
	head := make([]byte, 512)
	n, _ := io.ReadFull(file, head)
	ext := photoExts[http.DetectContentType(head[:n])]
	if ext == "" {
		return "", errors.New("Please upload your photo as a JPG, PNG or WebP image.")
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", err
	}
	base := slugify(name)
	if base == "" {
		base = "lead"
	}
	rel := "leads/" + base + "-" + randomToken(4) + ext
	full := filepath.Join(a.photosDir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return "", err
	}
	out, err := os.Create(full)
	if err != nil {
		return "", err
	}
	defer out.Close()
	if _, err := io.Copy(out, file); err != nil {
		return "", err
	}
	return rel, nil
}

// uniqueLeadSlug turns a name into a URL slug that no other lead uses yet.
func (a *App) uniqueLeadSlug(name string) (string, error) {
	base := slugify(name)
	if base == "" {
		base = "lead"
	}
	slug := base
	for i := 2; ; i++ {
		var n int
		if err := a.db.QueryRow(`SELECT COUNT(*) FROM leads WHERE slug = ?`, slug).Scan(&n); err != nil {
			return "", err
		}
		if n == 0 {
			return slug, nil
		}
		slug = base + "-" + strconv.Itoa(i)
	}
}

// adminSetLeadStatus approves or rejects a trek-lead application.
func (a *App) adminSetLeadStatus(w http.ResponseWriter, r *http.Request) {
	status := r.FormValue("status")
	active := 0
	switch status {
	case LeadApproved:
		active = 1
	case LeadRejected, LeadPending:
	default:
		http.Error(w, "Unknown status.", http.StatusBadRequest)
		return
	}
	if _, err := a.db.Exec(`UPDATE leads SET status = ?, active = ? WHERE id = ?`, status, active, pathID(r)); err != nil {
		a.serverError(w, err)
		return
	}
	redirect(w, r, "/admin/leads?msg=lead_status")
}
