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

// CarryGroup is a titled part of the packing list (the title may be empty).
type CarryGroup struct {
	Title string
	Items []string
}

func carryGroups(text string) []CarryGroup {
	var out []CarryGroup
	for _, line := range splitLines(text) {
		if title, ok := strings.CutSuffix(line, ":"); ok {
			out = append(out, CarryGroup{Title: strings.TrimSpace(title)})
			continue
		}
		if len(out) == 0 {
			out = append(out, CarryGroup{})
		}
		out[len(out)-1].Items = append(out[len(out)-1].Items, line)
	}
	return out
}

func init() {
	templateFuncs["pickups"] = parsePickups
	templateFuncs["carryGroups"] = carryGroups
}
