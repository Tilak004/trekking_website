package main

import (
	"bytes"
	"context"
	"encoding/json"
	"html"
	"io/fs"
	"net/http"
	"net/url"
	"strings"
)

// The React frontend (frontend/, built into static/app/) talks to the same Go
// handlers as the old server-rendered pages. Anything requested under /api/ is
// routed to the normal handler with the "/api" prefix removed, and the handler's
// render() and redirect() answer in JSON instead of HTML:
//
//	GET  /api/treks/kudremukh   -> the data the trek page used to be rendered from
//	POST /api/treks/x/book      -> {"redirect": "/my/bookings/7/sent"}
//
// Every other browser request gets the React app's HTML shell, so deep links and
// refreshes work (with per-page link-preview tags filled in by the server).

type apiKey struct{}

func isAPI(r *http.Request) bool {
	v, _ := r.Context().Value(apiKey{}).(bool)
	return v
}

// spa wraps the site's handler: /api/* goes to the handlers as JSON, assets go
// straight through, and any other page request is answered with the React shell.
func (a *App) spa(next http.Handler) http.Handler {
	shell := "static/app/index.html"
	if a.adminMode {
		shell = "static/app/admin.html"
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		switch {
		case strings.HasPrefix(p, "/api/"):
			r2 := r.WithContext(context.WithValue(r.Context(), apiKey{}, true))
			u := *r.URL
			u.Path, u.RawPath = strings.TrimPrefix(p, "/api"), ""
			r2.URL = &u
			w.Header().Set("Cache-Control", "no-store")
			next.ServeHTTP(w, r2)
		case strings.HasPrefix(p, "/static/"), strings.HasPrefix(p, "/photos/"),
			strings.HasSuffix(p, "/qr.png"), strings.HasSuffix(p, ".csv"):
			next.ServeHTTP(w, r)
		case r.Method == http.MethodGet || r.Method == http.MethodHead:
			a.serveShell(w, r, shell)
		default:
			http.NotFound(w, r)
		}
	})
}

func (a *App) serveShell(w http.ResponseWriter, r *http.Request, shell string) {
	raw, err := fs.ReadFile(staticFS, shell)
	if err != nil {
		http.Error(w, "The frontend isn't built yet. Run: cd frontend && npm run build", http.StatusServiceUnavailable)
		return
	}
	title, desc, image := a.shellMeta(r)
	base := baseURL(r)
	var meta bytes.Buffer
	esc := html.EscapeString
	meta.WriteString("<title>" + esc(title) + "</title>\n")
	meta.WriteString(`<meta name="description" content="` + esc(desc) + `">` + "\n")
	meta.WriteString(`<meta property="og:type" content="website">` + "\n")
	meta.WriteString(`<meta property="og:site_name" content="` + esc(site.Name) + `">` + "\n")
	meta.WriteString(`<meta property="og:title" content="` + esc(title) + `">` + "\n")
	meta.WriteString(`<meta property="og:description" content="` + esc(desc) + `">` + "\n")
	meta.WriteString(`<meta property="og:url" content="` + esc(base+r.URL.Path) + `">` + "\n")
	if image != "" {
		meta.WriteString(`<meta property="og:image" content="` + esc(base+(&url.URL{Path: image}).EscapedPath()) + `">` + "\n")
		meta.WriteString(`<meta name="twitter:card" content="summary_large_image">` + "\n")
	}
	out := bytes.Replace(raw, []byte("<!--META-->"), meta.Bytes(), 1)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Write(out)
}

// shellMeta picks the title, description and share image for a page, so link
// previews (WhatsApp, Instagram) show the right trek or lead.
func (a *App) shellMeta(r *http.Request) (title, desc, image string) {
	title, desc = site.Name+" · "+site.Tagline, site.Description
	if h := a.heroPhoto(); h != "" {
		image = "/photos/" + h
	}
	if a.adminMode {
		return "Admin · " + site.Name, desc, ""
	}
	p := r.URL.Path
	switch {
	case strings.HasPrefix(p, "/treks/"):
		if t, err := a.trekWhere(`t.slug = ? AND t.active = 1`, strings.TrimPrefix(p, "/treks/")); err == nil {
			title, desc = t.Name+" · "+site.Name, t.Summary
			if t.Photo != "" {
				image = "/photos/" + t.Photo
			}
		}
	case strings.HasPrefix(p, "/leads/"):
		if l, err := a.leadWhere(`l.slug = ? AND l.active = 1 AND l.status = 'approved'`, strings.TrimPrefix(p, "/leads/")); err == nil {
			title, desc = l.Name+" · "+site.Name, l.Headline
			if l.Photo != "" {
				image = "/photos/" + l.Photo
			}
		}
	}
	return
}

// apiJSON writes v as JSON with the given status.
func apiJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// MarshalJSON adds the values templates used to compute with Booking's methods
// (amount owed, payment page, refund rules), so the React app doesn't have to
// repeat that logic.
func (k Booking) MarshalJSON() ([]byte, error) {
	type plain Booking
	return json.Marshal(struct {
		plain
		Amount         int
		PayPath        string
		UPI            string
		CustomerRefund int
		RefundRule     string
		DaysToTrek     int
		TrekPassed     bool
		WALink         string
	}{plain(k), k.Amount(), k.PayPath(), k.UPILink(), k.CustomerRefund(), k.RefundRule(), k.DaysToTrek(), k.TrekPassed(),
		bookingsWALink(bookingMessage(k))})
}

// scrubKeys are fields that exist on our structs but were never shown on public
// pages (private trailhead coordinates, other people's emails and payment
// links). The JSON API must not start publishing them.
func scrubKeys(page string, admin bool) map[string]bool {
	if admin || isConsolePage(page) {
		return nil
	}
	drop := map[string]bool{"Lat": true, "Lng": true, "UserID": true}
	switch page {
	case "lead_apply.html", "lead_dashboard.html":
		// a lead sees their own phone number
	default:
		drop[leadPhoneKey] = true // a lead's phone is only for our team
	}
	switch page {
	case "my.html", "booking_sent.html", "pay.html":
		// a trekker's own bookings keep their payment link
	default:
		drop["PayToken"], drop["PayPath"], drop["UPI"], drop["UserEmail"] = true, true, true, true
	}
	return drop
}

// leadPhoneKey is a pseudo-key: it drops "Phone" only from lead profiles (objects
// with a Headline and Bio), never from the site's own contact number.
const leadPhoneKey = "lead:Phone"

// scrub deletes the dropped keys from decoded JSON, at any depth.
func scrub(v any, drop map[string]bool) {
	switch x := v.(type) {
	case map[string]any:
		if drop[leadPhoneKey] {
			if _, ok := x["Headline"]; ok {
				if _, ok := x["Bio"]; ok {
					delete(x, "Phone")
				}
			}
		}
		for k, val := range x {
			if drop[k] {
				delete(x, k)
				continue
			}
			scrub(val, drop)
		}
	case []any:
		for _, item := range x {
			scrub(item, drop)
		}
	}
}

// renderJSON answers an API request with the page's data as JSON.
func (a *App) renderJSON(w http.ResponseWriter, r *http.Request, page string, data map[string]any) {
	u := currentUser(r)
	drop := scrubKeys(page, u != nil && u.IsAdmin)
	raw, err := json.Marshal(data)
	if err != nil {
		a.serverError(w, err)
		return
	}
	var generic map[string]any
	if err := json.Unmarshal(raw, &generic); err != nil {
		a.serverError(w, err)
		return
	}
	if drop != nil {
		scrub(generic, drop)
	}
	generic["Page"] = page
	status := http.StatusOK
	if s, ok := data["Status"].(int); ok {
		status = s
	}
	apiJSON(w, status, generic)
}
