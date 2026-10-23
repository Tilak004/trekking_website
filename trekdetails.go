package main

import (
	"database/sql"
	"strings"
)

// Columns added to treks for the detailed trek page. They're added by
// migrateTrekDetails so existing databases pick them up on the next start.
var trekDetailColumns = []string{"best_season", "itinerary", "inclusions", "exclusions", "things_to_carry", "faqs"}

// migrateTrekDetails adds the detail columns if missing. When a column is newly
// added, treks with known slugs get sample content (edit it in the admin panel).
func migrateTrekDetails(db *sql.DB) error {
	added := false
	for _, col := range trekDetailColumns {
		if _, err := db.Exec(`ALTER TABLE treks ADD COLUMN ` + col + ` TEXT NOT NULL DEFAULT ''`); err == nil {
			added = true
		}
	}
	if !added {
		return nil
	}
	rows, err := db.Query(`SELECT id, slug, difficulty FROM treks`)
	if err != nil {
		return err
	}
	type row struct {
		id               int64
		slug, difficulty string
	}
	var treks []row
	for rows.Next() {
		var t row
		if err := rows.Scan(&t.id, &t.slug, &t.difficulty); err != nil {
			rows.Close()
			return err
		}
		treks = append(treks, t)
	}
	rows.Close()
	for _, t := range treks {
		d := sampleDetails(t.slug, t.difficulty)
		if _, err := db.Exec(`UPDATE treks SET best_season = ?, itinerary = ?, inclusions = ?, exclusions = ?, things_to_carry = ?, faqs = ?
WHERE id = ?`, d.BestSeason, d.Itinerary, d.Inclusions, d.Exclusions, d.ThingsToCarry, d.FAQs, t.id); err != nil {
			return err
		}
	}
	return nil
}

// ---- sample content (approximate; review before going live) ----

func dayTrip(trekLine, summitLine string) string {
	return "Night before · 10:00 PM\nDepart from Bangalore. Overnight drive to the base village.\n\n" +
		"Day 1 · 6:00 AM\nReach the base, freshen up and have breakfast. Safety briefing by your trek lead.\n\n" +
		"Day 1 · 7:30 AM\nStart the trek. " + trekLine + "\n\n" +
		"Day 1 · 12:30 PM\n" + summitLine + " Packed lunch with a view.\n\n" +
		"Day 1 · 5:00 PM\nBack at the base. Start the drive home.\n\n" +
		"Day 1 · 11:30 PM\nArrive in Bangalore."
}

var sampleItineraries = map[string]string{
	"kudremukh":      dayTrip("Walk through rolling grassland and shola patches, with several stream crossings.", "Reach the Kudremukh summit."),
	"netravati-peak": dayTrip("A steady climb through grassland onto the ridge.", "Reach Netravati Peak and take in the Kudremukh range."),
	"tadiandamol":    dayTrip("Climb past coffee estates into shola forest.", "Reach the open summit meadows of Coorg's highest peak."),
	"bandaje-falls":  dayTrip("Hike through forest and open grassland ridges.", "Reach the top of Bandaje Arbi falls."),
	"kumara-parvatha": "Night before · 9:00 PM\nDepart from Bangalore. Overnight drive to Kukke Subramanya.\n\n" +
		"Day 1 · 6:00 AM\nReach Kukke, freshen up and have breakfast. Safety briefing by your trek lead.\n\n" +
		"Day 1 · 7:00 AM\nStart the climb through dense forest to Bhattara Mane. Lunch break.\n\n" +
		"Day 1 · 3:00 PM\nContinue to the campsite. Set up camp, dinner and an early night.\n\n" +
		"Day 2 · 5:00 AM\nSummit push to Shesha Parvatha and Kumara Parvatha for sunrise.\n\n" +
		"Day 2 · 9:00 AM\nDescend to Kukke Subramanya. Lunch.\n\n" +
		"Day 2 · 4:00 PM\nStart the drive back. Arrive in Bangalore around 11:30 PM.",
	"skandagiri-night-trek": "11:00 PM\nDepart from Bangalore.\n\n" +
		"1:30 AM\nReach the base village. Headlamps on, safety briefing.\n\n" +
		"2:00 AM\nStart the night climb up the rocky trail.\n\n" +
		"5:30 AM\nReach the summit. Watch the sunrise over a sea of clouds.\n\n" +
		"7:30 AM\nDescend and have breakfast at the base.\n\n" +
		"10:30 AM\nArrive in Bangalore.",
	"huttari-betta": "5:00 AM\nDepart from Bangalore.\n\n" +
		"7:00 AM\nReach the base. Safety briefing and start the climb.\n\n" +
		"8:30 AM\nReach the boulder-topped summit. Time for photos and snacks.\n\n" +
		"10:00 AM\nDescend and have breakfast.\n\n" +
		"12:30 PM\nArrive in Bangalore.",
}

var sampleSeasons = map[string]string{
	"kudremukh":             "September – February",
	"netravati-peak":        "September – February",
	"tadiandamol":           "September – February",
	"bandaje-falls":         "September – January",
	"kumara-parvatha":       "October – February",
	"skandagiri-night-trek": "October – March",
	"huttari-betta":         "Year-round (start early)",
}

func sampleDetails(slug, difficulty string) Trek {
	overnight := slug == "kumara-parvatha" || slug == "skandagiri-night-trek"

	inclusions := []string{
		"Return transport from Bangalore (for Bangalore pickups)",
		"Breakfast and lunch on the trek day",
		"Forest permits and entry fees",
		"Experienced trek lead and first-aid kit",
	}
	if slug == "kumara-parvatha" {
		inclusions = append(inclusions, "Tents, sleeping mats and dinner at the campsite")
	}
	carry := []string{
		"Trekking shoes with good grip",
		"2 litres of water",
		"Rain jacket or poncho (June – November)",
		"Light snacks: nuts, fruit, energy bars",
		"Cap, sunglasses and sunscreen",
		"Small backpack (20–30 L)",
		"Government photo ID",
		"Personal medicines",
	}
	if overnight {
		carry = append(carry, "Headlamp with spare batteries", "Warm jacket for the night")
	}

	var beginner string
	switch difficulty {
	case "Easy", "Easy–Moderate":
		beginner = "Yes. If you can walk for a few hours with short breaks, you'll be fine."
	case "Difficult", "Very Difficult":
		beginner = "We recommend doing at least one moderate trek first. You should be able to jog 5 km comfortably."
	default:
		beginner = "Yes, with a little preparation. Regular walks or jogging for 2–3 weeks before the trek will help a lot."
	}
	faqs := []string{
		"Is this trek suitable for beginners?\n" + beginner,
		"Do I need to pay to request a seat?\nNo. Booking requests are free. We confirm your seat over WhatsApp or a call and share payment details then.",
		newBoardFAQ,
		"What if it rains heavily or the trail is closed?\nYour safety comes first. If the forest department closes the trail or conditions are unsafe, we'll move you to another date or trek. See our cancellation policy for refunds.",
	}

	return Trek{
		BestSeason: sampleSeasons[slug],
		Itinerary:  sampleItineraries[slug],
		Inclusions: strings.Join(inclusions, "\n"),
		Exclusions: strings.Join([]string{
			"Food and drinks during travel, other than the meals listed",
			"Personal expenses and tips",
			"Anything not listed under inclusions",
		}, "\n"),
		ThingsToCarry: strings.Join(carry, "\n"),
		FAQs:          strings.Join(faqs, "\n\n"),
	}
}
