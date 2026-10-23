package main

import (
	"database/sql"
	"embed"
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

//go:embed static
var staticFS embed.FS

// version is the commit this program was built from. deploy.ps1 sets it
// (-ldflags "-X main.version=..."); a plain "go build" leaves it as "dev".
var version = "dev"

type App struct {
	demo      bool // running on a demo database: shows a banner
	db        *sql.DB
	photosDir string
	adminMode bool // the separate admin console process (see console.go)
}

func main() {
	mode := flag.String("mode", envOr("MODE", "public"), `"public" serves the website, "admin" serves the admin console`)
	addr := flag.String("addr", os.Getenv("ADDR"), "HTTP listen address (default :8080, or :8081 in admin mode)")
	dbPath := flag.String("db", envOr("DB_PATH", "beyond.db"), "path to the SQLite database file")
	photosDir := flag.String("photos", envOr("PHOTOS_DIR", "photos"), "folder with site photos")
	resetAdmin := flag.Bool("reset-admin-password", false, "set a new random admin password, print it and exit")
	backupNow := flag.Bool("backup-now", false, "write a backup of the database and uploaded photos, then exit")
	showVersion := flag.Bool("version", false, "print the version this program was built from and exit")
	flag.Parse()
	if *showVersion {
		fmt.Println(version)
		return
	}
	if *mode != "public" && *mode != "admin" {
		log.Fatalf("unknown -mode %q (use public or admin)", *mode)
	}
	adminMode := *mode == "admin"
	if *addr == "" {
		*addr = ":8080"
		if adminMode {
			*addr = ":8081"
		}
	}
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
	if *backupNow {
		path, err := runBackup(db, *photosDir, loadBackupConfig())
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("backup written:", path)
		return
	}

	base := strings.TrimSuffix(filepath.Base(*dbPath), filepath.Ext(*dbPath))
	sessionCookie = "session_" + slugify(base)
	if adminMode {
		sessionCookie = "admin_" + sessionCookie // never shared with the public site's login
	}
	app := &App{db: db, photosDir: *photosDir, demo: strings.Contains(base, "demo"), adminMode: adminMode}
	srv := &http.Server{Addr: *addr, Handler: app.routes(), ReadHeaderTimeout: 10 * time.Second}
	if adminMode {
		log.Printf("%s admin console running on %s (version %s)", site.Name, *addr, version)
	} else {
		log.Printf("%s running on %s (version %s)", site.Name, *addr, version)
		// Only the public process runs the nightly backup, so it happens once.
		if !strings.Contains(base, "demo") {
			startBackups(db, *photosDir, loadBackupConfig())
		}
	}
	log.Fatal(srv.ListenAndServe())
}

func (a *App) routes() http.Handler {
	if a.adminMode {
		return a.adminRoutes()
	}
	mux := http.NewServeMux()
	mux.Handle("GET /static/", staticFiles())
	mux.Handle("GET /photos/", a.servePhotos())

	mux.HandleFunc("GET /{$}", a.home)
	mux.HandleFunc("GET /treks/{slug}", a.trekPage)
	mux.HandleFunc("POST /treks/{slug}/book", a.bookTrek)
	mux.HandleFunc("POST /treks/{slug}/enquire", a.enquire)
	mux.HandleFunc("GET /leads", a.leadsPage)
	mux.HandleFunc("GET /leads/{slug}", a.leadPage)
	mux.HandleFunc("GET /lead-with-us", a.leadApplyPage)
	mux.HandleFunc("GET /lead", a.requireUser(a.leadDashboard))
	mux.HandleFunc("POST /lead/dates", a.requireUser(a.leadProposeDate))
	mux.HandleFunc("POST /lead/dates/{id}/withdraw", a.requireUser(a.leadWithdrawDate))
	mux.HandleFunc("POST /lead-with-us", a.requireUser(a.leadApply))
	mux.HandleFunc("GET /about", a.staticPage("about.html"))
	mux.HandleFunc("GET /contact", a.staticPage("contact.html"))
	mux.HandleFunc("POST /contact", a.contactEnquire)
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
	mux.HandleFunc("GET /pay/{id}/{token}", a.payPage)
	mux.HandleFunc("GET /pay/{id}/{token}/qr.png", a.payQR)
	mux.HandleFunc("POST /my/bookings/{id}/cancel", a.requireUser(a.cancelBooking))
	mux.HandleFunc("POST /my/bookings/{id}/review", a.requireUser(a.reviewBooking))

	return a.spa(a.withUser(mux))
}

// Flash messages are passed as short codes in the query string (?msg=booked)
// so arbitrary text can never be injected into a page through a link.
var flashMessages = map[string]string{
	"welcome":           "Welcome aboard! Pick a trek and send us a booking request.",
	"booked":            "Booking request sent. We'll confirm your seat shortly.",
	"refunded":          "Refund marked as sent.",
	"enquired":          "Thanks! We've got your enquiry and will call or WhatsApp you soon.",
	"cancelled":         "Booking cancelled.",
	"reviewed":          "Thanks for the review!",
	"saved":             "Saved.",
	"batch_added":       "Batch added.",
	"batch_updated":     "Batch updated.",
	"batch_deleted":     "Batch deleted.",
	"status":            "Booking updated.",
	"applied":           "Application received! We'll review it and get back to you.",
	"lead_updated":      "Your trek-lead profile is updated.",
	"lead_status":       "Trek lead updated.",
	"lead_assigned":     "Lead assigned to that trek date.",
	"proposed":          "Date proposed! It goes live once our team approves it.",
	"withdrawn":         "Date withdrawn.",
	"date_reviewed":     "Proposed date updated.",
	"event_created":     "Event saved as a draft. Publish it when you're ready.",
	"event_saved":       "Event updated.",
	"event_published":   "Event published. Trekkers can now book it.",
	"event_unpublished": "Event moved back to drafts. Trekkers can no longer see it.",
}

var flashErrors = map[string]string{
	"full":               "Not enough seats left in that batch.",
	"bad_batch":          "That batch is not available.",
	"already_booked":     "You already have an active booking for this trek.",
	"batch_has_bookings": "This batch has bookings, so it can't be deleted.",
	"bad_review":         "Please give a rating from 1 to 5.",
	"cannot_review":      "You can only review a completed trek, once.",
	"lead_not_approved":  "Only approved leads can be assigned to treks.",
	"bad_contact":        "Please enter your name and a valid 10-digit phone number.",
	"bad_enquiry":        "Enquiry not sent: please enter your name and a valid 10-digit phone number.",
	"bad_photo":          "Please upload a JPG, PNG or WebP photo under 5 MB.",
	"bad_date":           "Please pick a date at least 3 days from today.",
	"bad_seats":          "Group size must be between 2 and 40.",
	"bad_trek":           "Please pick a trek from the list.",
	"dup_date":           "You have already proposed that trek on that date.",
	"event_has_bookings": "This event has active registrations, so it can't go back to draft. Cancel or reject them first.",
}

func (a *App) render(w http.ResponseWriter, r *http.Request, page string, data map[string]any) {
	if data == nil {
		data = map[string]any{}
	}
	data["Brand"] = site.Name
	data["Site"] = site
	data["Demo"] = a.demo
	data["PublicURL"] = envOr("PUBLIC_URL", "http://localhost:8080")
	data["AdminURL"] = os.Getenv("ADMIN_URL") // shown to admins on the public site only if set
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
	if _, ok := data["Msg"]; !ok {
		data["Msg"] = flashMessages[r.URL.Query().Get("msg")]
	}
	if _, ok := data["Err"]; !ok {
		data["Err"] = flashErrors[r.URL.Query().Get("err")]
	}
	a.renderJSON(w, r, page, data)
}

func (a *App) serverError(w http.ResponseWriter, err error) {
	log.Printf("error: %v", err)
	http.Error(w, "Something went wrong. Please try again.", http.StatusInternalServerError)
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
	redirectTo(w, r, fmt.Sprintf(format, args...))
}

// redirectTo sends the browser to url. For /api requests (the React app) it
// answers {"redirect": url} instead, and the app navigates there itself.
func redirectTo(w http.ResponseWriter, r *http.Request, url string) {
	if isAPI(r) {
		apiJSON(w, http.StatusOK, map[string]any{"redirect": url})
		return
	}
	http.Redirect(w, r, url, http.StatusSeeOther)
}
