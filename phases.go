package main

import (
	"database/sql"
	"time"
)

// Every event is in one of four phases, worked out from its dates and seats (nothing
// to switch by hand):
//
//	opening soon     bookings open at a date and time an admin sets (opens_at)
//	booking open     opens_at has passed and seats are left
//	sold out         no seats left
//	event completed  the start date is over
//
// opens_at is stored in UTC ("2006-01-02 15:04:05") like hold_until. Admins type it in
// Indian time. Events saved before this existed have no opens_at and count as open.
const (
	PhaseOpeningSoon = "opening_soon"
	PhaseOpen        = "open"
	PhaseSoldOut     = "sold_out"
	PhaseCompleted   = "completed"
)

const opensAtLayout = "2006-01-02 15:04:05"

var istZone = time.FixedZone("IST", 5*3600+1800)

// SQL fragment: the batch (aliased b) has opened for booking.
const batchOpenedSQL = `(b.opens_at = '' OR b.opens_at <= datetime('now'))`

func migratePhases(db *sql.DB) {
	db.Exec(`ALTER TABLE batches ADD COLUMN opens_at TEXT NOT NULL DEFAULT ''`)
}

func phaseOf(startDate, opensAt string, left int) string {
	switch {
	case startDate < today():
		return PhaseCompleted
	case opensAt != "" && opensAt > time.Now().UTC().Format(opensAtLayout):
		return PhaseOpeningSoon
	case left <= 0:
		return PhaseSoldOut
	}
	return PhaseOpen
}

var phaseLabels = map[string]string{
	PhaseOpeningSoon: "Opening soon",
	PhaseOpen:        "Booking open",
	PhaseSoldOut:     "Sold out",
	PhaseCompleted:   "Event completed",
}

// opensAtFromForm turns the admin's datetime-local value (Indian time) into the stored
// UTC value. ok is false for anything that isn't a date and time.
func opensAtFromForm(v string) (string, bool) {
	t, err := time.ParseInLocation("2006-01-02T15:04", v, istZone)
	if err != nil {
		return "", false
	}
	return t.UTC().Format(opensAtLayout), true
}

func opensAtInIST(stored string) (time.Time, bool) {
	t, err := time.ParseInLocation(opensAtLayout, stored, time.UTC)
	if err != nil {
		return time.Time{}, false
	}
	return t.In(istZone), true
}

// opensAtField is the value for a datetime-local input.
func opensAtField(stored string) string {
	if t, ok := opensAtInIST(stored); ok {
		return t.Format("2006-01-02T15:04")
	}
	return ""
}

// opensAtLabel is how trekkers and admins read it, e.g. "15 Nov 2026, 6:00 AM".
func opensAtLabel(stored string) string {
	if t, ok := opensAtInIST(stored); ok {
		return t.Format("2 Jan 2006, 3:04 PM")
	}
	return ""
}

// Phase is the trek's status on the website, taken from its next dates: open if any
// date can be booked now, else opening soon, else sold out. Empty when it has no
// upcoming published date at all.
func (t Trek) Phase() string {
	switch {
	case t.NextBatch != "":
		return PhaseOpen
	case t.NextOpens != "":
		return PhaseOpeningSoon
	case t.Upcoming > 0:
		return PhaseSoldOut
	}
	return ""
}
