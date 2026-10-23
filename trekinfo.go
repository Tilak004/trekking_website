package main

import (
	"database/sql"
	"net/http"
	"strings"
)

// Extra trek details shown on the trek page:
//
//   - trail status: admins mark a trail open or closed (forest department,
//     monsoon) with a short note; the trek page and cards show it.
//   - pickup times: each pickup line can carry a time after a "|",
//     e.g. "Majestic, Shanthala Silks | 11:00 PM".
//   - packing list by category: a line ending in ":" in "What to carry"
//     starts a group, e.g. "Clothing:".

const (
	TrailOpen   = "open"
	TrailClosed = "closed"
)

// Pickups are two Bangalore stops; there is no "reach the base directly" option.
// Treks still on the old default list, and the matching default FAQ, are updated
// on startup.
const (
	oldDefaultPickups = "Bangalore - Majestic\nBangalore - Silk Board\nReach base village directly"
	defaultPickups    = "Near BMS College, Basavanagudi\nBanashankari bus stand"
	oldBaseFAQ        = "Can I join directly at the base village?\nYes. Choose \"Reach base village directly\" as your pickup point and we'll share the meeting point and time."
	newBoardFAQ       = "Where do we board?\nAt one of two pickup points in Bangalore: near BMS College, Basavanagudi, or Banashankari bus stand. Pick one when you book and we'll share the exact time on WhatsApp."
)

func migratePickups(db *sql.DB) {
	db.Exec(`UPDATE treks SET pickup_points = ? WHERE pickup_points = ?`, defaultPickups, oldDefaultPickups)
	db.Exec(`UPDATE treks SET faqs = replace(faqs, ?, ?) WHERE instr(faqs, ?) > 0`, oldBaseFAQ, newBoardFAQ, oldBaseFAQ)
}

func migrateTrailStatus(db *sql.DB) {
	db.Exec(`ALTER TABLE treks ADD COLUMN trail_status TEXT NOT NULL DEFAULT ''`)
	db.Exec(`ALTER TABLE treks ADD COLUMN trail_note TEXT NOT NULL DEFAULT ''`)
	db.Exec(`ALTER TABLE treks ADD COLUMN trail_updated TEXT NOT NULL DEFAULT ''`)
}

// adminSetTrailStatus saves a trek's trail status and note, stamped with today's date.
func (a *App) adminSetTrailStatus(w http.ResponseWriter, r *http.Request) {
	status := r.FormValue("trail_status")
	if status != TrailOpen && status != TrailClosed {
		status = ""
	}
	id := pathID(r)
	if _, err := a.db.Exec(`UPDATE treks SET trail_status = ?, trail_note = ?, trail_updated = date('now') WHERE id = ?`,
		status, trimTo(r.FormValue("trail_note"), 160), id); err != nil {
		a.serverError(w, err)
		return
	}
	redirect(w, r, "/admin/treks/%d?msg=saved", id)
}

// Pickup is one pickup point, with an optional time.
type Pickup struct{ Place, Time string }

// Label is how a pickup is shown in the booking form and stored on a booking.
func (p Pickup) Label() string {
	if p.Time == "" {
		return p.Place
	}
	return p.Place + " · " + p.Time
}

func parsePickups(text string) []Pickup {
	var out []Pickup
	for _, line := range splitLines(text) {
		place, at, _ := strings.Cut(line, "|")
		if p := (Pickup{strings.TrimSpace(place), strings.TrimSpace(at)}); p.Place != "" {
			out = append(out, p)
		}
	}
	return out
}
