package main

import (
	"database/sql"
	"net/http"
	"strings"
)

// Trek defaults: text that is the same for most treks ("Not included", "What
// to carry", FAQs), written once in Admin → Treks → Trek defaults. New treks
// start with it filled in, and the trek form has buttons to pull it into an
// existing trek. Each trek keeps its own copy, so it can still be changed.

var defaultTrekFields = []string{"exclusions", "things_to_carry", "faqs"}

func migrateSettings(db *sql.DB) {
	db.Exec(`CREATE TABLE IF NOT EXISTS settings (key TEXT PRIMARY KEY, value TEXT NOT NULL DEFAULT '')`)
	starter := map[string]string{
		"default_exclusions": strings.Join([]string{
			"Food and drinks during travel, other than the meals listed",
			"Personal expenses and tips",
			"Insurance of any kind (travel, accident or medical)",
			"Anything not listed under inclusions",
		}, "\n"),
		"default_things_to_carry": strings.Join([]string{
			"Trekking shoes with good grip",
			"2 litres of water",
			"Rain jacket or poncho (June – November)",
			"Light snacks: nuts, fruit, energy bars",
			"Cap, sunglasses and sunscreen",
			"Small backpack (20–30 L)",
			"Government photo ID",
			"Personal medicines",
		}, "\n"),
		"default_faqs": strings.Join([]string{
			"Do I need to pay to request a seat?\nNo. Booking requests are free. We confirm your seat over WhatsApp or a call and share payment details then.",
			newBoardFAQ,
			"What if it rains heavily or the trail is closed?\nYour safety comes first. If the forest department closes the trail or conditions are unsafe, we'll move you to another date or trek. See our cancellation policy for refunds.",
			"Is insurance included?\nNo. Our price does not include travel, accident or medical insurance, and any medical or evacuation costs are yours. If you want cover, you can buy personal accident insurance on your own before the trek.",
		}, "\n\n"),
	}
	for k, v := range starter {
		db.Exec(`INSERT OR IGNORE INTO settings (key, value) VALUES (?, ?)`, k, v)
	}
}

// TrekDefaults is the common trek text, keyed like the trek form fields.
type TrekDefaults struct {
	Exclusions    string `json:"exclusions"`
	ThingsToCarry string `json:"things_to_carry"`
	FAQs          string `json:"faqs"`
}

func (a *App) trekDefaults() TrekDefaults {
	get := func(k string) string {
		var v string
		a.db.QueryRow(`SELECT value FROM settings WHERE key = ?`, "default_"+k).Scan(&v)
		return v
	}
	return TrekDefaults{Exclusions: get("exclusions"), ThingsToCarry: get("things_to_carry"), FAQs: get("faqs")}
}

func (a *App) adminTrekDefaults(w http.ResponseWriter, r *http.Request) {
	var channel string
	a.db.QueryRow(`SELECT value FROM settings WHERE key = 'whatsapp_channel'`).Scan(&channel)
	a.render(w, r, "admin_trek_defaults.html", map[string]any{"Defaults": a.trekDefaults(), "Channel": channel})
}

func (a *App) adminSaveTrekDefaults(w http.ResponseWriter, r *http.Request) {
	for _, f := range defaultTrekFields {
		v := r.FormValue(f)
		if f == "faqs" {
			v = trimTo(v, 10000)
		} else {
			v = strings.Join(splitLines(v), "\n")
		}
		if _, err := a.db.Exec(`INSERT INTO settings (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value`,
			"default_"+f, v); err != nil {
			a.serverError(w, err)
			return
		}
	}
	channel, ok := cleanWALink(r.FormValue("whatsapp_channel"), channelLinkPrefix)
	if !ok {
		redirect(w, r, "/admin/trek-defaults?err=bad_channel_link")
		return
	}
	if _, err := a.db.Exec(`INSERT INTO settings (key, value) VALUES ('whatsapp_channel', ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value`, channel); err != nil {
		a.serverError(w, err)
		return
	}
	redirect(w, r, "/admin/trek-defaults?msg=saved")
}
