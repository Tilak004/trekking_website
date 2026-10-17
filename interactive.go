package main

import (
	"database/sql"
	"strings"
)

// Data behind the interactive (Preact) parts of the site: trek finder, map,
// booking flow and trust stats. The server still renders every page; these
// structs are embedded in the page as JSON for the islands to pick up.

// migrateInteractive adds trek coordinates, finder tags, elevation profiles and
// review photos. Known treks get approximate coordinates and tags the first time.
func migrateInteractive(db *sql.DB) error {
	added := false
	for _, col := range []string{
		`lat REAL NOT NULL DEFAULT 0`, `lng REAL NOT NULL DEFAULT 0`,
		`tags TEXT NOT NULL DEFAULT ''`, `elevation TEXT NOT NULL DEFAULT ''`,
	} {
		if _, err := db.Exec(`ALTER TABLE treks ADD COLUMN ` + col); err == nil {
			added = true
		}
	}
	db.Exec(`ALTER TABLE reviews ADD COLUMN photo TEXT NOT NULL DEFAULT ''`)
	if !added {
		return nil
	}
	for slug, g := range trekGeo {
		if _, err := db.Exec(`UPDATE treks SET lat = ?, lng = ?, tags = ? WHERE slug = ?`, g.lat, g.lng, g.tags, slug); err != nil {
			return err
		}
	}
	return nil
}

// Approximate trailhead/summit coordinates; fine-tune them in Admin → Treks.
var trekGeo = map[string]struct {
	lat, lng float64
	tags     string
}{
	"bananthimari-betta":    {12.52, 77.45, "sunrise, rocky"},
	"bandaje-falls":         {13.11, 75.42, "waterfalls, grasslands, forest"},
	"huttari-betta":         {12.93, 77.10, "rocky, sunrise"},
	"kudremukh":             {13.13, 75.27, "grasslands, summit"},
	"kumara-parvatha":       {12.66, 75.69, "forest, summit"},
	"netravati-peak":        {13.18, 75.32, "grasslands, summit, sunrise"},
	"skandagiri-night-trek": {13.42, 77.68, "sunrise, rocky"},
	"tadiandamol":           {12.22, 75.65, "forest, grasslands, summit"},
}

type LeadMini struct {
	Name     string `json:"name"`
	Slug     string `json:"slug"`
	Photo    string `json:"photo"`
	Verified bool   `json:"verified"`
}

// nextLeads returns the lead of each trek's next upcoming batch.
func (a *App) nextLeads() (map[int64]LeadMini, error) {
	rows, err := a.db.Query(`
SELECT b.trek_id, l.name, l.slug, l.photo, l.verified
FROM batches b JOIN leads l ON l.id = b.lead_id
WHERE b.start_date >= date('now') AND b.status = 'live' AND l.active = 1 AND l.status = 'approved'
ORDER BY b.start_date`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[int64]LeadMini{}
	for rows.Next() {
		var id int64
		var l LeadMini
		if err := rows.Scan(&id, &l.Name, &l.Slug, &l.Photo, &l.Verified); err != nil {
			return nil, err
		}
		if _, seen := out[id]; !seen {
			out[id] = l
		}
	}
	return out, rows.Err()
}

// FinderTrek is one trek as the finder and map see it.
type FinderTrek struct {
	Slug       string    `json:"slug"`
	Name       string    `json:"name"`
	Region     string    `json:"region"`
	Difficulty string    `json:"difficulty"`
	Duration   string    `json:"duration"`
	DistanceKm int       `json:"distanceKm"`
	Altitude   int       `json:"altitude"`
	Photo      string    `json:"photo"`
	Price      int       `json:"price"` // 0 when no dates are open
	Next       string    `json:"next"`
	Tags       []string  `json:"tags"`
	Lat        float64   `json:"lat"`
	Lng        float64   `json:"lng"`
	Rating     float64   `json:"rating"`
	Ratings    int       `json:"ratings"`
	Lead       *LeadMini `json:"lead"`
}

func finderTrek(t Trek) FinderTrek {
	f := FinderTrek{
		Slug: t.Slug, Name: t.Name, Region: t.Region, Difficulty: t.Difficulty, Duration: t.Duration,
		DistanceKm: t.DistanceKm, Altitude: t.MaxAltitude, Photo: t.Photo, Tags: splitTags(t.Tags),
		Lat: t.Lat, Lng: t.Lng, Rating: t.AvgRating, Ratings: t.Ratings, Lead: t.Lead,
	}
	if t.NextBatch != "" {
		f.Price, f.Next = t.Price, formatDate(t.NextBatch)
	}
	return f
}

func splitTags(s string) []string {
	out := []string{}
	for _, t := range strings.Split(s, ",") {
		if t = strings.ToLower(strings.TrimSpace(t)); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// TrustStats are real numbers from completed bookings and reviews.
type TrustStats struct {
	TreksDone, Trekkers, Reviews int
	AvgRating                    float64
}

func (a *App) trustStats() (TrustStats, error) {
	var s TrustStats
	if err := a.db.QueryRow(`SELECT COUNT(DISTINCT batch_id), COALESCE(SUM(people), 0) FROM bookings WHERE status = 'completed'`).
		Scan(&s.TreksDone, &s.Trekkers); err != nil {
		return s, err
	}
	err := a.db.QueryRow(`SELECT COUNT(*), COALESCE(AVG(trek_rating), 0.0) FROM reviews`).Scan(&s.Reviews, &s.AvgRating)
	return s, err
}

// BookingData drives the step-by-step booking flow on a trek page.
type BookingData struct {
	Slug    string         `json:"slug"`
	Price   int            `json:"price"`
	Pickups []string       `json:"pickups"`
	Batches []BookingBatch `json:"batches"`
}

type BookingBatch struct {
	ID    int64  `json:"id"`
	Label string `json:"label"`
	Left  int    `json:"left"`
	Lead  string `json:"lead"`
}

func bookingData(t Trek, batches []Batch) BookingData {
	d := BookingData{Slug: t.Slug, Price: t.Price, Pickups: splitLines(t.PickupPoints), Batches: []BookingBatch{}}
	if d.Pickups == nil {
		d.Pickups = []string{}
	}
	for _, b := range batches {
		d.Batches = append(d.Batches, BookingBatch{ID: b.ID, Label: formatDate(b.StartDate), Left: b.Left(), Lead: b.LeadName})
	}
	return d
}
