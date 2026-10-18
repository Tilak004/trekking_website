package main

import (
	"net/http"
	"net/url"
	"strings"
)

// Site holds the brand and contact details shown across the site.
// Everything marked PLACEHOLDER must be replaced before going live.
type Site struct {
	Name        string
	Tagline     string
	Description string // used for search engines and link previews
	Phone       string // primary number, shown as text
	Phone2      string // secondary number
	WhatsApp    string // digits only, with country code, for wa.me links
	UPI         string // UPI ID payments go to
	BookingsWA  string // WhatsApp number (digits, with country code) that receives booking requests and payment screenshots
	Email       string
	Instagram   string // handle without @
	City        string
}

var site = Site{
	Name:        "Beyond Expeditions",
	Tagline:     "Weekend treks in the Western Ghats, led by athletes.",
	Description: "Small-group weekend treks across Karnataka's Western Ghats, led by an Ironman finisher, a 24-hour adventure racer and a trekker who has done every major trail in Karnataka.",
	Phone:       "+91 77604 24521",
	Phone2:      "+91 94817 40319",
	WhatsApp:    "917760424521",
	UPI:         "9481740319@ibl",
	BookingsWA:  "919481740319",
	Email:       "beyond.expiditions.kpt@gmail.com", // shown only as an "Email us" link
	Instagram:   "beyond___expeditions",
	City:        "Bangalore, Karnataka",
}

// whatsAppLink opens a WhatsApp chat with an optional pre-filled message.
func whatsAppLink(text string) string {
	link := "https://wa.me/" + site.WhatsApp
	if text != "" {
		link += "?text=" + strings.ReplaceAll(url.QueryEscape(text), "+", "%20")
	}
	return link
}

// baseURL is used for absolute links in link previews (og:image needs one).
func baseURL(r *http.Request) string {
	if u := envOr("BASE_URL", ""); u != "" {
		return strings.TrimRight(u, "/")
	}
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}

// whatsAppTo opens a WhatsApp chat with someone else's number (e.g. a trekker),
// with a message filled in. Indian numbers without a country code get +91.
func whatsAppTo(phone, text string) string {
	d := strings.Map(func(c rune) rune {
		if c >= '0' && c <= '9' {
			return c
		}
		return -1
	}, phone)
	if len(d) == 10 {
		d = "91" + d
	}
	return "https://wa.me/" + d + "?text=" + strings.ReplaceAll(url.QueryEscape(text), "+", "%20")
}

// bookingsWALink opens a WhatsApp chat with the bookings desk (Site.BookingsWA).
func bookingsWALink(text string) string {
	return whatsAppTo(site.BookingsWA, text)
}

// staticPage answers pages that need no data of their own (About, Terms, ...). The
// React app draws them; this call gives it the site details and the signed-in user.
func (a *App) staticPage(page string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) { a.render(w, r, page, nil) }
}

func (a *App) notFound(w http.ResponseWriter, r *http.Request) {
	a.render(w, r, "notfound.html", map[string]any{"Status": http.StatusNotFound})
}
