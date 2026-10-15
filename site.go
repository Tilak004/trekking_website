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
	Phone       string // shown as text
	WhatsApp    string // digits only, with country code, for wa.me links
	Email       string
	Instagram   string // handle without @
	City        string
}

var site = Site{
	Name:        "Northstar Treks", // PLACEHOLDER
	Tagline:     "Weekend treks in the Western Ghats, led by athletes.",
	Description: "Small-group weekend treks across Karnataka's Western Ghats, led by an Ironman finisher, a 24-hour adventure racer and a trekker who has done every major trail in Karnataka.",
	Phone:       "+91 94817 40319",
	WhatsApp:    "919481740319",
	Email:       "hello@northstartreks.in", // PLACEHOLDER
	Instagram:   "northstartreks",          // PLACEHOLDER
	City:        "Bangalore, Karnataka",
}

// defaultTheme is the theme every visitor sees. Others can be previewed with ?theme=<name>.
const defaultTheme = "ghats"

type Theme struct {
	Name, Label, Fonts string // Fonts is a Google Fonts css2 query
}

var themes = []Theme{
	{"ghats", "Western Ghats", "family=Outfit:wght@500;600;700;800&family=Inter:wght@400;500;600;700"},
	{"explore", "Explore", "family=Montserrat:wght@500;600;700;800&family=Inter:wght@400;500;600;700&family=Bebas+Neue&family=Kaushan+Script"},
	{"forest", "Forest & sunrise", "family=Oswald:wght@500;600;700&family=Inter:wght@400;500;600;700"},
	{"alpine", "Alpine", "family=Montserrat:wght@600;700&family=Inter:wght@400;500;600;700"},
	{"earth", "Earth", "family=Fraunces:opsz,wght@9..144,500;9..144,600;9..144,700&family=DM+Sans:wght@400;500;600;700"},
	{"bold", "Bold", "family=Barlow+Condensed:wght@600;700&family=Barlow:wght@400;500;600;700"},
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

// pickTheme returns the theme for this request. ?theme=<name> starts a preview
// (remembered in a cookie); ?theme=off ends it.
func pickTheme(w http.ResponseWriter, r *http.Request) (Theme, bool) {
	def, _ := findTheme(defaultTheme)
	q := r.URL.Query().Get("theme")
	if q == "off" {
		http.SetCookie(w, &http.Cookie{Name: "theme", Path: "/", MaxAge: -1})
		return def, false
	}
	if t, ok := findTheme(q); ok {
		http.SetCookie(w, &http.Cookie{Name: "theme", Value: t.Name, Path: "/", MaxAge: 7 * 24 * 3600, SameSite: http.SameSiteLaxMode})
		return t, true
	}
	if c, err := r.Cookie("theme"); err == nil {
		if t, ok := findTheme(c.Value); ok {
			return t, true
		}
	}
	return def, false
}

// TelLink is the phone number without spaces, for tel: links.
func (s Site) TelLink() string { return strings.ReplaceAll(s.Phone, " ", "") }

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
