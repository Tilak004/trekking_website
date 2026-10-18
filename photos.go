package main

import (
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

// Photos live on disk (not embedded) so they can be added or swapped without a rebuild.
//
//	photos/hero.jpg      big background image on the start page
//	photos/gallery/*     "From the trail" strip on the start page
//	photos/**            anything else can be picked as a trek's cover photo in the admin panel

var imageExts = map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".avif": true}

var videoExts = map[string]bool{".mp4": true, ".webm": true}

func isImage(name string) bool { return imageExts[strings.ToLower(path.Ext(name))] }

// servePhotos serves image and video files only, with no directory listings.
func (a *App) servePhotos() http.Handler {
	files := http.StripPrefix("/photos/", http.FileServer(http.Dir(a.photosDir)))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isImage(r.URL.Path) && !videoExts[strings.ToLower(path.Ext(r.URL.Path))] {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Cache-Control", "public, max-age=86400")
		files.ServeHTTP(w, r)
	})
}

// listPhotos returns every image under the photos folder as slash-separated relative paths.
func (a *App) listPhotos(sub string) []string {
	root := filepath.Join(a.photosDir, sub)
	var out []string
	filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !isImage(d.Name()) {
			return nil
		}
		rel, err := filepath.Rel(a.photosDir, p)
		if err == nil {
			out = append(out, filepath.ToSlash(rel))
		}
		return nil
	})
	sort.Strings(out)
	return out
}

type Slide struct{ Photo, Num string }

// heroSlides is the home-page slideshow: hero.jpg first, then every gallery
// photo that isn't already some trek's cover (so the banner never repeats a
// trek card). A gallery copy of hero.jpg is skipped too.
func (a *App) heroSlides(treks []Trek) []Slide {
	used := map[string]bool{}
	for _, t := range treks {
		used[t.Photo] = true
	}
	hero := a.heroPhoto()
	var heroSize int64 = -1
	if fi, err := os.Stat(filepath.Join(a.photosDir, hero)); hero != "" && err == nil {
		heroSize = fi.Size()
	}
	photos := []string{}
	if hero != "" {
		photos = append(photos, hero)
	}
	for _, g := range a.listPhotos("gallery") {
		if used[g] {
			continue
		}
		if fi, err := os.Stat(filepath.Join(a.photosDir, filepath.FromSlash(g))); err == nil && fi.Size() == heroSize {
			continue
		}
		photos = append(photos, g)
	}
	out := make([]Slide, 0, len(photos))
	for _, ph := range photos {
		out = append(out, Slide{Photo: ph, Num: fmt.Sprintf("%02d", len(out)+1)})
	}
	return out
}

// HeroLayers is a hero photo split into the full image and a cut-out of the
// land (sky removed), so the title can sit between them, behind the hills.
// Made from photos/hero-layers/{background.jpg,foreground.png,skyline.txt}.
type HeroLayers struct {
	Aspect  float64 // image height / width
	Skyline float64 // typical skyline height, as a fraction of image height
	Peak    float64 // highest point of the land, as a fraction of image height
}

func (a *App) heroLayers() *HeroLayers {
	dir := filepath.Join(a.photosDir, "hero-layers")
	for _, f := range []string{"background.webp", "foreground.webp"} {
		if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
			return nil
		}
	}
	raw, err := os.ReadFile(filepath.Join(dir, "skyline.txt"))
	if err != nil {
		return nil
	}
	var top, median, h, w float64
	if n, _ := fmt.Sscan(string(raw), &top, &median, &h, &w); n < 3 || h == 0 {
		return nil
	}
	if w == 0 {
		w = h * 4 / 3
	}
	return &HeroLayers{Aspect: h / w, Skyline: median / h, Peak: top / h}
}

// Style is the CSS custom properties the layered hero is positioned with.
func (l *HeroLayers) Style() string {
	return (fmt.Sprintf("--ar:%.4f;--sky:%.4f;--peak:%.4f", l.Aspect, l.Skyline, l.Peak))
}

// heroPhoto finds photos/hero.<ext>, or returns "" if there isn't one.
func (a *App) heroPhoto() string { return a.findHero(imageExts) }

func (a *App) findHero(exts map[string]bool) string {
	for ext := range exts {
		if _, err := os.Stat(filepath.Join(a.photosDir, "hero"+ext)); err == nil {
			return "hero" + ext
		}
	}
	return ""
}

// heroStyle is the layered hero's CSS custom properties, for the React frontend
// (which can't call the Style method).
func (a *App) heroStyle() string {
	if l := a.heroLayers(); l != nil {
		return l.Style()
	}
	return ""
}
