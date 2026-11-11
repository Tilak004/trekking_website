package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

// A trek can stay a secret until a set time (reveal_at, stored in UTC like opens_at,
// set in Admin > Treks). Until then it has a mystery page at a neutral address with its
// photos but no name, place or price: its real page is "not found", bookings and
// enquiries are refused, and it is left out of the sitemap, FAQs, lead pages and gallery.
// At reveal_at everything appears by itself. The trek's own details are never changed.
const trekRevealedSQL = `(t.reveal_at = '' OR t.reveal_at <= datetime('now'))`

// Bookings for the first trek open with its reel on 21 Nov 2026, 6:00 AM Indian time.
const firstTrekSlug, firstTrekRevealAt = "bananthimari-betta", "2026-11-21 00:30:00"

// A secret trek's public address, so the link doesn't name it.
const teaserSlugFormat = "coming-soon-%d"

func migrateReveal(db *sql.DB) {
	// Only the first time the column is added, so a later change in the admin stays.
	if _, err := db.Exec(`ALTER TABLE treks ADD COLUMN reveal_at TEXT NOT NULL DEFAULT ''`); err == nil {
		db.Exec(`UPDATE treks SET reveal_at = ? WHERE slug = ?`, firstTrekRevealAt, firstTrekSlug)
	}
}

// Teased reports whether the trek is still a secret.
func (t Trek) Teased() bool {
	return t.RevealAt > time.Now().UTC().Format(opensAtLayout)
}

// splitTeased separates the treks the public can see from the secret ones.
func splitTeased(treks []Trek) (public, teased []Trek) {
	for _, t := range treks {
		if t.Teased() {
			teased = append(teased, t)
		} else {
			public = append(public, t)
		}
	}
	return public, teased
}

// teasedTrek finds the secret trek behind a neutral address like "coming-soon-0".
func (a *App) teasedTrek(slug string) (Trek, bool) {
	var i int
	if _, err := fmt.Sscanf(slug, teaserSlugFormat, &i); err != nil || slug != fmt.Sprintf(teaserSlugFormat, i) {
		return Trek{}, false
	}
	treks, err := a.listTreks(true)
	if err != nil {
		return Trek{}, false
	}
	_, teased := splitTeased(treks)
	if i < 0 || i >= len(teased) {
		return Trek{}, false
	}
	return teased[i], true
}

func revealLabel(t Trek) string {
	if at, ok := opensAtInIST(t.RevealAt); ok {
		return at.Format("2 Jan")
	}
	return ""
}

// teaserPage is a secret trek's page: its photos and reveal date, nothing that names it.
func (a *App) teaserPage(w http.ResponseWriter, r *http.Request, slug string, t Trek) {
	photos := []string{}
	for n := range a.listPhotos("treks/" + t.Slug) {
		photos = append(photos, fmt.Sprintf("/treks/%s/photos/%d.jpg", slug, n))
	}
	a.render(w, r, "trek.html", map[string]any{"Teaser": true, "Photos": photos, "Opens": revealLabel(t),
		"Follow": "https://www.instagram.com/" + site.Instagram + "/"})
}

// teaserPhoto serves a secret trek's photo by its neutral address.
func (a *App) teaserPhoto(w http.ResponseWriter, r *http.Request) {
	var n int
	t, ok := a.teasedTrek(r.PathValue("slug"))
	if _, err := fmt.Sscanf(r.PathValue("file"), "%d.jpg", &n); !ok || err != nil {
		http.NotFound(w, r)
		return
	}
	photos := a.listPhotos("treks/" + t.Slug)
	if n < 0 || n >= len(photos) {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	http.ServeFile(w, r, filepath.Join(a.photosDir, filepath.FromSlash(photos[n])))
}

// teaserCard is the finder card for a secret trek: a photo, but nothing that names it.
func (a *App) teaserCard(t Trek, n int) FinderTrek {
	slug := fmt.Sprintf(teaserSlugFormat, n)
	photo := ""
	if len(a.listPhotos("treks/"+t.Slug)) > 0 {
		photo = "/treks/" + slug + "/photos/0.jpg"
	}
	return FinderTrek{Slug: slug, Photo: photo, Tags: []string{}, Status: PhaseOpeningSoon, Opens: revealLabel(t), Teaser: true}
}

// publicGallery is the gallery without photos that would give a secret trek away,
// by its cover photo or a file name with the trek's name in it.
func (a *App) publicGallery() []string {
	treks, err := a.listTreks(true)
	if err != nil {
		return nil
	}
	_, teased := splitTeased(treks)
	out := []string{}
	for _, p := range a.listPhotos("gallery") {
		hidden := false
		for _, t := range teased {
			word, _, _ := strings.Cut(t.Slug, "-")
			hidden = hidden || p == t.Photo || strings.Contains(strings.ToLower(p), word)
		}
		if !hidden {
			out = append(out, p)
		}
	}
	return out
}

// saveTrekReveal stores the admin's "keep secret until" time (Indian time); blank makes the trek public.
func (a *App) saveTrekReveal(id int64, v string) error {
	at := ""
	if v != "" {
		var ok bool
		if at, ok = opensAtFromForm(v); !ok {
			return nil // not a date and time: keep what was saved
		}
	}
	_, err := a.db.Exec(`UPDATE treks SET reveal_at = ? WHERE id = ?`, at, id)
	return err
}
