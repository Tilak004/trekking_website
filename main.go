package main

import (
	"bytes"
	"database/sql"
	"embed"
	"flag"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

//go:embed templates/*.html
var templateFS embed.FS

//go:embed static
var staticFS embed.FS

type App struct {
	demo      bool // running on a demo database: shows a banner
	db        *sql.DB
	pages     map[string]*template.Template
	photosDir string
}

func main() {
	addr := flag.String("addr", envOr("ADDR", ":8080"), "HTTP listen address")
	dbPath := flag.String("db", envOr("DB_PATH", "beyond.db"), "path to the SQLite database file")
	photosDir := flag.String("photos", envOr("PHOTOS_DIR", "photos"), "folder with site photos")
	resetAdmin := flag.Bool("reset-admin-password", false, "set a new random admin password, print it and exit")
	flag.Parse()
	os.MkdirAll(*photosDir+"/gallery", 0o755)

	db, err := openDB(*dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if *resetAdmin {
		if err := resetAdminPassword(db); err != nil {
			log.Fatal(err)
		}
		return
	}

	base := strings.TrimSuffix(filepath.Base(*dbPath), filepath.Ext(*dbPath))
	sessionCookie = "session_" + slugify(base)
	app := &App{db: db, pages: loadPages(), photosDir: *photosDir, demo: strings.Contains(base, "demo")}
	srv := &http.Server{Addr: *addr, Handler: app.routes(), ReadHeaderTimeout: 10 * time.Second}
	log.Printf("%s running on %s", site.Name, *addr)
	log.Fatal(srv.ListenAndServe())
}

func (a *App) routes() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.FileServerFS(staticFS))
	mux.Handle("GET /photos/", a.servePhotos())

	mux.HandleFunc("GET /{$}", a.home)
	mux.HandleFunc("GET /treks/{slug}", a.trekPage)
	mux.HandleFunc("POST /treks/{slug}/book", a.bookTrek)
	mux.HandleFunc("GET /leads", a.leadsPage)
	mux.HandleFunc("GET /leads/{slug}", a.leadPage)
	mux.HandleFunc("GET /lead-with-us", a.leadApplyPage)
	mux.HandleFunc("GET /lead", a.requireUser(a.leadDashboard))
	mux.HandleFunc("POST /lead/dates", a.requireUser(a.leadProposeDate))
	mux.HandleFunc("POST /lead/dates/{id}/withdraw", a.requireUser(a.leadWithdrawDate))
	mux.HandleFunc("POST /lead-with-us", a.requireUser(a.leadApply))
	mux.HandleFunc("GET /about", a.staticPage("about.html"))
	mux.HandleFunc("GET /contact", a.staticPage("contact.html"))
	mux.HandleFunc("GET /gallery", a.galleryPage)
	mux.HandleFunc("GET /terms", a.staticPage("terms.html"))
	mux.HandleFunc("GET /privacy", a.staticPage("privacy.html"))
	mux.HandleFunc("GET /cancellation-policy", a.staticPage("cancellation.html"))
	mux.HandleFunc("/", a.notFound)

	mux.HandleFunc("GET /signup", a.signupForm)
	mux.HandleFunc("POST /signup", a.signup)
	mux.HandleFunc("GET /login", a.loginForm)
	mux.HandleFunc("POST /login", a.login)
	mux.HandleFunc("POST /logout", a.logout)

	mux.HandleFunc("GET /my", a.requireUser(a.myBookings))
	mux.HandleFunc("GET /my/bookings/{id}/sent", a.requireUser(a.bookingSent))
	mux.HandleFunc("POST /my/bookings/{id}/cancel", a.requireUser(a.cancelBooking))
	mux.HandleFunc("POST /my/bookings/{id}/review", a.requireUser(a.reviewBooking))

	mux.HandleFunc("GET /admin", a.requireAdmin(a.adminBookings))
	mux.HandleFunc("POST /admin/bookings/{id}/status", a.requireAdmin(a.adminSetBookingStatus))
	mux.HandleFunc("GET /admin/treks", a.requireAdmin(a.adminTreks))
	mux.HandleFunc("GET /admin/treks/new", a.requireAdmin(a.adminTrekForm))
	mux.HandleFunc("POST /admin/treks/new", a.requireAdmin(a.adminSaveTrek))
	mux.HandleFunc("GET /admin/treks/{id}", a.requireAdmin(a.adminTrekForm))
	mux.HandleFunc("POST /admin/treks/{id}", a.requireAdmin(a.adminSaveTrek))
	mux.HandleFunc("POST /admin/treks/{id}/batches", a.requireAdmin(a.adminAddBatch))
	mux.HandleFunc("POST /admin/batches/{id}/lead", a.requireAdmin(a.adminSetBatchLead))
	mux.HandleFunc("POST /admin/batches/{id}/delete", a.requireAdmin(a.adminDeleteBatch))
	mux.HandleFunc("POST /admin/batches/{id}/review", a.requireAdmin(a.adminReviewBatch))
	mux.HandleFunc("GET /admin/leads", a.requireAdmin(a.adminLeads))
	mux.HandleFunc("GET /admin/leads/new", a.requireAdmin(a.adminLeadForm))
	mux.HandleFunc("POST /admin/leads/new", a.requireAdmin(a.adminSaveLead))
	mux.HandleFunc("GET /admin/leads/{id}", a.requireAdmin(a.adminLeadForm))
	mux.HandleFunc("POST /admin/leads/{id}", a.requireAdmin(a.adminSaveLead))
	mux.HandleFunc("POST /admin/leads/{id}/status", a.requireAdmin(a.adminSetLeadStatus))
	mux.HandleFunc("GET /admin/users", a.requireAdmin(a.adminUsers))

	return a.withUser(mux)
}

// Flash messages are passed as short codes in the query string (?msg=booked)
// so arbitrary text can never be injected into a page through a link.
var flashMessages = map[string]string{
	"welcome":       "Welcome aboard! Pick a trek and send us a booking request.",
	"booked":        "Booking request sent. We'll confirm your seat shortly.",
	"cancelled":     "Booking cancelled.",
	"reviewed":      "Thanks for the review!",
	"saved":         "Saved.",
	"batch_added":   "Batch added.",
	"batch_updated": "Batch updated.",
	"batch_deleted": "Batch deleted.",
	"status":        "Booking updated.",
	"applied":       "Application received! We'll review it and get back to you.",
	"lead_updated":  "Your trek-lead profile is updated.",
	"lead_status":   "Trek lead updated.",
	"proposed":      "Date proposed! It goes live once our team approves it.",
	"withdrawn":     "Date withdrawn.",
	"date_reviewed": "Proposed date updated.",
}

var flashErrors = map[string]string{
	"full":               "Not enough seats left in that batch.",
	"bad_batch":          "That batch is not available.",
	"already_booked":     "You already have an active booking for this trek.",
	"batch_has_bookings": "This batch has bookings, so it can't be deleted.",
	"bad_review":         "Please give a rating from 1 to 5.",
	"cannot_review":      "You can only review a completed trek, once.",
	"bad_contact":        "Please enter your name and a valid 10-digit phone number.",
	"bad_photo":          "Please upload a JPG, PNG or WebP photo under 5 MB.",
	"bad_date":           "Please pick a date at least 3 days from today.",
	"bad_seats":          "Group size must be between 2 and 40.",
	"bad_trek":           "Please pick a trek from the list.",
	"dup_date":           "You have already proposed that trek on that date.",
}

func (a *App) render(w http.ResponseWriter, r *http.Request, page string, data map[string]any) {
	if data == nil {
		data = map[string]any{}
	}
	data["Brand"] = site.Name
	data["Site"] = site
	data["Demo"] = a.demo
	if u := currentUser(r); u != nil {
		role := "Trekker"
		if u.IsAdmin {
			role = "Admin"
		} else if l, ok, _ := a.myLead(u.ID); ok && l.Status == LeadApproved {
			role = "Trek lead"
		}
		data["Role"] = role
	}
	data["User"] = currentUser(r)
	data["Path"] = r.URL.Path
	data["BaseURL"] = baseURL(r)
	theme, previewing := pickTheme(w, r)
	data["Theme"] = theme
	data["ThemePreview"] = previewing
	data["Themes"] = themes
	if _, ok := data["MetaDesc"]; !ok {
		data["MetaDesc"] = site.Description
	}
	if _, ok := data["MetaImage"]; !ok && a.heroPhoto() != "" {
		data["MetaImage"] = "/photos/" + a.heroPhoto()
	}
	if _, ok := data["Msg"]; !ok {
		data["Msg"] = flashMessages[r.URL.Query().Get("msg")]
	}
	if _, ok := data["Err"]; !ok {
		data["Err"] = flashErrors[r.URL.Query().Get("err")]
	}
	var buf bytes.Buffer
	if err := a.pages[page].ExecuteTemplate(&buf, "layout.html", data); err != nil {
		a.serverError(w, err)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if status, ok := data["Status"].(int); ok {
		w.WriteHeader(status)
	}
	buf.WriteTo(w)
}

func (a *App) serverError(w http.ResponseWriter, err error) {
	log.Printf("error: %v", err)
	http.Error(w, "Something went wrong. Please try again.", http.StatusInternalServerError)
}

func loadPages() map[string]*template.Template {
	files, err := fs.Glob(templateFS, "templates/*.html")
	if err != nil {
		log.Fatal(err)
	}
	pages := map[string]*template.Template{}
	for _, f := range files {
		name := path.Base(f)
		if name == "layout.html" {
			continue
		}
		pages[name] = template.Must(template.New("layout.html").Funcs(templateFuncs).
			ParseFS(templateFS, "templates/layout.html", f))
	}
	return pages
}

var templateFuncs = template.FuncMap{
	"inr":     formatINR,
	"fmtDate": formatDate,
	"lines":   splitLines,
	"stars":   stars,
	"starsN":  func(n int) string { return stars(float64(n)) },
	"initial": func(s string) string {
		r := []rune(strings.TrimSpace(s))
		if len(r) == 0 {
			return "?"
		}
		return strings.ToUpper(string(r[0]))
	},
	"rating": func(f float64) string { return strconv.FormatFloat(f, 'f', 1, 64) },
	"seq": func(n int) []int {
		s := make([]int, n)
		for i := range s {
			s[i] = i + 1
		}
		return s
	},
	"nextActions": nextActions,
	"blocks":      splitBlocks,
	"slug":        slugify,
	"paras":       splitParas,
	"bookingWA":   func(k Booking) string { return whatsAppLink(bookingMessage(k)) },
	"hasPrefix":   strings.HasPrefix,
	"firstWord": func(s string) string {
		f := strings.Fields(s)
		if len(f) == 0 {
			return ""
		}
		return f[0]
	},
	"restWords": func(s string) string {
		f := strings.Fields(s)
		if len(f) < 2 {
			return ""
		}
		return strings.Join(f[1:], " ")
	},
	"urlpath":     func(p string) string { return (&url.URL{Path: p}).EscapedPath() },
	"whatsapp":    whatsAppLink,
	"year":        func() int { return time.Now().Year() },
	"actionLabel": func(s string) string { return actionLabels[s] },
}

// formatINR formats a rupee amount with Indian digit grouping (1,25,000).
func formatINR(n int) string {
	s := strconv.Itoa(n)
	if len(s) <= 3 {
		return s
	}
	head, tail := s[:len(s)-3], s[len(s)-3:]
	var parts []string
	for len(head) > 2 {
		parts = append([]string{head[len(head)-2:]}, parts...)
		head = head[:len(head)-2]
	}
	parts = append([]string{head}, parts...)
	return strings.Join(parts, ",") + "," + tail
}

func formatDate(s string) string {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return s
	}
	return t.Format("Mon, 2 Jan 2006")
}

func splitLines(s string) []string {
	var out []string
	for _, l := range strings.Split(s, "\n") {
		if l = strings.TrimSpace(l); l != "" {
			out = append(out, l)
		}
	}
	return out
}

func stars(f float64) string {
	n := int(math.Round(f))
	return strings.Repeat("★", n) + strings.Repeat("☆", 5-n)
}

var nonSlug = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(s string) string {
	return strings.Trim(nonSlug.ReplaceAllString(strings.ToLower(s), "-"), "-")
}

func formInt(r *http.Request, key string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(r.FormValue(key)))
	return n
}

func formFloat(r *http.Request, key string) float64 {
	f, _ := strconv.ParseFloat(strings.TrimSpace(r.FormValue(key)), 64)
	return f
}

func pathID(r *http.Request) int64 {
	n, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	return n
}

func trimTo(s string, max int) string {
	s = strings.TrimSpace(s)
	if len([]rune(s)) > max {
		s = string([]rune(s)[:max])
	}
	return s
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func redirect(w http.ResponseWriter, r *http.Request, format string, args ...any) {
	http.Redirect(w, r, fmt.Sprintf(format, args...), http.StatusSeeOther)
}
