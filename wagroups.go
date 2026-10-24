package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
)

// WhatsApp trek groups and channel.
//
// Each trek has one permanent WhatsApp group (an invite link pasted in Admin →
// Treks), and the site has one WhatsApp channel for announcements (Admin →
// Treks → Trek defaults). Once a booking is paid, the admin console offers a
// one-tap "Send group invite" that opens WhatsApp to the trekker with the
// message written, and the trekker's My bookings shows "Join the trek group".
// After the trek, "Send thank-you" asks them to review, leave the group and
// follow the channel. The group link is only ever shown to admins and to
// trekkers who have paid for that trek.

const (
	groupLinkPrefix   = "https://chat.whatsapp.com/"
	channelLinkPrefix = "https://whatsapp.com/channel/"
)

func migrateWAGroups(db *sql.DB) {
	db.Exec(`ALTER TABLE treks ADD COLUMN whatsapp_group TEXT NOT NULL DEFAULT ''`)
}

// cleanWALink keeps a link only if it starts with the expected WhatsApp prefix.
func cleanWALink(v, prefix string) (string, bool) {
	v = strings.TrimSpace(v)
	if v == "" {
		return "", true
	}
	if !strings.HasPrefix(v, prefix) || len(v) > 300 {
		return "", false
	}
	return v, true
}

// saveTrekGroup stores a trek's WhatsApp group link (after the trek itself is saved).
func (a *App) saveTrekGroup(id int64, r *http.Request) error {
	link, ok := cleanWALink(r.FormValue("whatsapp_group"), groupLinkPrefix)
	if !ok {
		return fmt.Errorf("bad group link")
	}
	_, err := a.db.Exec(`UPDATE treks SET whatsapp_group = ? WHERE id = ?`, link, id)
	return err
}

func (a *App) trekGroup(id int64) string {
	var v string
	a.db.QueryRow(`SELECT whatsapp_group FROM treks WHERE id = ?`, id).Scan(&v)
	return v
}

func firstName(name string) string {
	if f := strings.Fields(name); len(f) > 0 {
		return f[0]
	}
	return name
}

// GroupLink is the trek group invite, only once the booking is paid and live.
func (k Booking) GroupLink() string {
	if k.PaidAt == "" || (k.Status != StatusApproved && k.Status != "completed") {
		return ""
	}
	return k.WhatsAppGroup
}

// GroupInviteWA opens WhatsApp to the trekker with the group invite written.
func (k Booking) GroupInviteWA() string {
	link := k.GroupLink()
	if link == "" || k.Status != StatusApproved {
		return ""
	}
	text := fmt.Sprintf("Hi %s! We've received your payment for %s on %s. Your seat is confirmed.\n\n"+
		"Please join the %s WhatsApp group for pickup details and updates:\n%s\n\n"+
		"We'll approve your request to join shortly. See you on the trail!",
		firstName(k.UserName), k.TrekName, formatDate(k.StartDate), k.TrekName, link)
	return whatsAppTo(k.UserPhone, text)
}

// ThankYouWA opens WhatsApp to the trekker after the trek: review, leave the
// group, follow the channel.
func (k Booking) ThankYouWA(baseURL string) string {
	if k.Status != "completed" {
		return ""
	}
	lines := []string{fmt.Sprintf("Hi %s! Thank you for trekking %s with us. We hope you had a great time!", firstName(k.UserName), k.TrekName)}
	if !k.Reviewed {
		lines = append(lines, "", "We'd love a quick review: "+baseURL+"/my")
	}
	if k.ChannelLink != "" {
		lines = append(lines, "", "For future treks and announcements, follow our WhatsApp channel:", k.ChannelLink,
			"", fmt.Sprintf("You're welcome to leave the %s group now.", k.TrekName))
	}
	return whatsAppTo(k.UserPhone, strings.Join(lines, "\n"))
}

// publicURL is the public website's address, for links in messages sent from the console.
func publicURL() string { return strings.TrimRight(envOr("PUBLIC_URL", "http://localhost:8080"), "/") }
