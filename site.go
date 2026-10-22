package main

import (
	"html/template"
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

// defaultTheme is the theme every visitor sees. Others can be previewed with ?theme=<name>.
const defaultTheme = "sage"

type Theme struct {
	Name, Label, Fonts string // Fonts is a Google Fonts css2 query
	Base               string // optional theme whose layout this one builds on
}

// Attr is the value of <html data-theme>, e.g. "ghats night".
func (t Theme) Attr() string { return strings.TrimSpace(t.Base + " " + t.Name) }

var themes = []Theme{
	{"editorial", "Editorial", "family=Source+Sans+3:wght@400;500;600;700", "ghats sky"},
	{"mist", "Mist", "family=Source+Sans+3:wght@400;500;600;700", "ghats sky editorial"},
	{"sage", "Sage", "family=Source+Sans+3:wght@400;500;600;700", "ghats sky editorial"},
	{"sky", "Sky", "family=Source+Sans+3:wght@400;500;600;700", "ghats"},
	{"night", "Shola Night", "family=Marcellus&family=Outfit:wght@600;700;800&family=Inter:wght@300;400;500;600", "ghats"},
	{"ghats", "Western Ghats", "family=Outfit:wght@500;600;700;800&family=Inter:wght@400;500;600;700", ""},
	{"explore", "Explore", "family=Montserrat:wght@500;600;700;800&family=Inter:wght@400;500;600;700&family=Bebas+Neue&family=Kaushan+Script", ""},
	{"forest", "Forest & sunrise", "family=Oswald:wght@500;600;700&family=Inter:wght@400;500;600;700", ""},
	{"alpine", "Alpine", "family=Montserrat:wght@600;700&family=Inter:wght@400;500;600;700", ""},
	{"earth", "Earth", "family=Fraunces:opsz,wght@9..144,500;9..144,600;9..144,700&family=DM+Sans:wght@400;500;600;700", ""},
	{"bold", "Bold", "family=Barlow+Condensed:wght@600;700&family=Barlow:wght@400;500;600;700", ""},
}

// FontsHref is the stylesheet link for the theme's fonts. It's a fixed string from
// the themes list above, so it's safe to mark as a trusted URL.
func (t Theme) FontsHref() template.URL {
	return template.URL("https://fonts.googleapis.com/css2?" + t.Fonts + "&display=swap")
}

func findTheme(name string) (Theme, bool) {
	for _, t := range themes {
		if t.Name == name {
			return t, true
		}
	}
	return Theme{}, false
}

// pickTheme returns the theme for this request: always the default.
func pickTheme(w http.ResponseWriter, r *http.Request) (Theme, bool) {
	// The site is locked to the default theme (Sage). Theme previews with
	// ?theme= are switched off; clear any preview cookie left from before.
	if _, err := r.Cookie("theme"); err == nil {
		http.SetCookie(w, &http.Cookie{Name: "theme", Path: "/", MaxAge: -1})
	}
	def, _ := findTheme(defaultTheme)
	return def, false
}

// TelLink is the phone number without spaces, for tel: links.
func (s Site) TelLink() string { return strings.ReplaceAll(s.Phone, " ", "") }

// TelLink2 is the secondary number without spaces, for tel: links.
func (s Site) TelLink2() string { return strings.ReplaceAll(s.Phone2, " ", "") }

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

// Block is a titled chunk of text: the first line is the title, the rest the body.
// Admins write itineraries and FAQs as blocks separated by a blank line.
type Block struct {
	Title string
	Body  string
}

func splitBlocks(s string) []Block {
	var out []Block
	for _, chunk := range strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n\n") {
		chunk = strings.TrimSpace(chunk)
		if chunk == "" {
			continue
		}
		title, body, _ := strings.Cut(chunk, "\n")
		out = append(out, Block{Title: strings.TrimSpace(title), Body: strings.TrimSpace(body)})
	}
	return out
}

// splitParas splits text into paragraphs on blank lines.
func splitParas(s string) []string {
	var out []string
	for _, p := range strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n\n") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// staticPage renders a template that needs no data (About, Terms, ...).
func (a *App) staticPage(page string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) { a.render(w, r, page, nil) }
}

func (a *App) notFound(w http.ResponseWriter, r *http.Request) {
	a.render(w, r, "notfound.html", map[string]any{"Status": http.StatusNotFound})
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
