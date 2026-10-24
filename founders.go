package main

import "database/sql"

// The three founders as trek leads. New databases are seeded with them
// (sampleLeads), and databases still holding the old "Founder One/Two/Three"
// placeholders are switched to the real founders on startup. A lead that has
// already been renamed in the admin console is left alone.

type founderLead struct {
	placeholder                                   string
	name, slug, headline, bio, achievements, photo string
}

var founderLeads = []founderLead{
	{"Founder One", "Prahladh", "prahladh", "Ironman finisher",
		"Ironman finisher. Swims, cycles and runs further than most people drive on a weekend.",
		"Full Ironman finisher", "founders/prahladh.jpg"},
	{"Founder Two", "Lepakshi", "lepakshi", "24-hour adventure racer",
		"Races 24-hour adventure races. Happiest when the trail goes on longer than planned.",
		"24-hour adventure races\nUltra trail races", "founders/lepakshi.jpg"},
	{"Founder Three", "Tilak", "tilak", "Western Ghats and Himalayan trekker",
		"Knows the Western Ghats trail by trail, and has trekked the Himalayas too. Happiest planning the next one.",
		"Most major Western Ghats treks\nHimalayan treks\nHalf marathon runner", "founders/tilak.jpg"},
}

func migrateFounderLeads(db *sql.DB) {
	for _, f := range founderLeads {
		db.Exec(`UPDATE leads SET name = ?, slug = ?, headline = ?, bio = ?, achievements = ?, photo = ?
WHERE name = ? AND NOT EXISTS (SELECT 1 FROM leads WHERE slug = ?)`,
			f.name, f.slug, f.headline, f.bio, f.achievements, f.photo, f.placeholder, f.slug)
	}
}

// removeSampleEvents deletes the ten sample dates the very first start used to
// create (two per sample trek, ids 1–10, 20 seats). Only dates that nobody has
// booked are removed, so no booking or payment is ever lost; dates created in
// the admin console have later ids and are never touched.
func removeSampleEvents(db *sql.DB) {
	db.Exec(`DELETE FROM batches
WHERE id <= 10 AND seats = 20 AND status = 'live'
  AND trek_id IN (SELECT id FROM treks WHERE slug IN ('kumara-parvatha', 'kudremukh', 'netravati-peak', 'tadiandamol', 'skandagiri-night-trek'))
  AND NOT EXISTS (SELECT 1 FROM bookings k WHERE k.batch_id = batches.id)`)
}
