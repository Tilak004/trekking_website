package main

import (
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// The admin console runs as its own process ("northstar -mode admin", port 8081 by
// default) on the same database as the public site. It has its own login and
// session cookie, so nothing here is reachable from the public server.
//
// Money: there are no payments in the app yet, so every rupee figure is an
// ESTIMATE = trek price x people on bookings that are approved or completed.
// When a payment gateway is added, replace confirmedValue below with real
// amounts and the console's numbers become actual collections.

const BatchDraft = "draft" // saved by an admin but not yet published

func init() {
	// waTo turns a phone number into the digits wa.me expects (10-digit numbers get India's +91).
	templateFuncs["waTo"] = func(phone string) string {
		var d strings.Builder
		for _, c := range phone {
			if c >= '0' && c <= '9' {
				d.WriteRune(c)
			}
		}
		if d.Len() == 10 {
			return "91" + d.String()
		}
		return d.String()
	}
}

func isConsolePage(name string) bool {
	return strings.HasPrefix(name, "admin_") || strings.HasPrefix(name, "console_")
}

const confirmedStatuses = `'approved', 'completed'`

func (a *App) adminRoutes() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.FileServerFS(staticFS))
	mux.Handle("GET /photos/", a.servePhotos())

	mux.HandleFunc("GET /login", a.adminLoginForm)
	mux.HandleFunc("POST /login", a.adminLogin)
	mux.HandleFunc("POST /logout", a.adminLogout)
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) { redirect(w, r, "/admin") })
	mux.HandleFunc("/", http.NotFound)

	adm := a.requireAdmin
	mux.HandleFunc("GET /admin", adm(a.consoleDashboard))

	mux.HandleFunc("GET /admin/events", adm(a.consoleEvents))
	mux.HandleFunc("GET /admin/events/new", adm(a.consoleEvent))
	mux.HandleFunc("POST /admin/events/new", adm(a.consoleSaveEvent))
	mux.HandleFunc("GET /admin/events/{id}", adm(a.consoleEvent))
	mux.HandleFunc("POST /admin/events/{id}", adm(a.consoleSaveEvent))
	mux.HandleFunc("POST /admin/events/{id}/status", adm(a.consoleEventStatus))
	mux.HandleFunc("POST /admin/events/{id}/delete", adm(a.consoleDeleteEvent))
	mux.HandleFunc("POST /admin/batches/{id}/review", adm(a.adminReviewBatch))

	mux.HandleFunc("GET /admin/registrations", adm(a.consoleRegistrations))
	mux.HandleFunc("GET /admin/registrations.csv", adm(a.consoleRegistrationsCSV))
	mux.HandleFunc("POST /admin/bookings/{id}/status", adm(a.adminSetBookingStatus))

	mux.HandleFunc("GET /admin/analytics", adm(a.consoleAnalytics))

	mux.HandleFunc("GET /admin/treks", adm(a.adminTreks))
	mux.HandleFunc("GET /admin/treks/new", adm(a.adminTrekForm))
	mux.HandleFunc("POST /admin/treks/new", adm(a.adminSaveTrek))
	mux.HandleFunc("GET /admin/treks/{id}", adm(a.adminTrekForm))
	mux.HandleFunc("POST /admin/treks/{id}", adm(a.adminSaveTrek))
	mux.HandleFunc("POST /admin/treks/{id}/batches", adm(a.adminAddBatch))
	mux.HandleFunc("POST /admin/batches/{id}/lead", adm(a.adminSetBatchLead))
	mux.HandleFunc("POST /admin/batches/{id}/delete", adm(a.adminDeleteBatch))
	mux.HandleFunc("GET /admin/leads", adm(a.adminLeads))
	mux.HandleFunc("GET /admin/leads/new", adm(a.adminLeadForm))
	mux.HandleFunc("POST /admin/leads/new", adm(a.adminSaveLead))
	mux.HandleFunc("GET /admin/leads/{id}", adm(a.adminLeadForm))
	mux.HandleFunc("POST /admin/leads/{id}", adm(a.adminSaveLead))
	mux.HandleFunc("POST /admin/leads/{id}/status", adm(a.adminSetLeadStatus))
	mux.HandleFunc("GET /admin/users", adm(a.adminUsers))
	return a.withUser(mux)
}

// ---- login ----

func (a *App) adminLoginForm(w http.ResponseWriter, r *http.Request) {
	if u := currentUser(r); u != nil && u.IsAdmin {
		redirect(w, r, "/admin")
		return
	}
	a.render(w, r, "console_login.html", map[string]any{"Next": r.URL.Query().Get("next")})
}

func (a *App) adminLogin(w http.ResponseWriter, r *http.Request) {
	email := strings.ToLower(strings.TrimSpace(r.FormValue("email")))
	next := r.FormValue("next")
	var id int64
	var hash string
	var isAdmin bool
	err := a.db.QueryRow(`SELECT id, password_hash, is_admin FROM users WHERE email = ?`, email).Scan(&id, &hash, &isAdmin)
	// One message for every failure, so the form never says which part was wrong.
	if err != nil || !isAdmin || bcrypt.CompareHashAndPassword([]byte(hash), []byte(r.FormValue("password"))) != nil {
		a.render(w, r, "console_login.html", map[string]any{"Email": email, "Next": next, "Err": "Wrong email or password, or this account is not an admin."})
		return
	}
	if err := a.startSession(w, r, id); err != nil {
		a.serverError(w, err)
		return
	}
	dest := safeNext(next)
	if dest == "/" {
		dest = "/admin"
	}
	http.Redirect(w, r, dest, http.StatusSeeOther)
}

func (a *App) adminLogout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(sessionCookie); err == nil {
		a.db.Exec(`DELETE FROM sessions WHERE token = ?`, c.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: "", Path: "/", MaxAge: -1})
	redirect(w, r, "/login")
}

// ---- events (a trek on a date) ----

type Event struct {
	ID, TrekID         int64
	TrekName, TrekSlug string
	Region             string
	Price              int
	StartDate          string
	Seats, Taken       int // taken = people on approved/completed bookings
	Active             int // registrations still alive (not rejected/cancelled)
	Requested          int // registrations waiting for approval
	Status             string
	LeadID             int64
	LeadName           string
	TrekHidden         bool
}

func (e Event) Left() int       { return e.Seats - e.Taken }
func (e Event) Revenue() int    { return e.Taken * e.Price }
func (e Event) Past() bool      { return e.StartDate < today() }
func (e Event) Fill() int       { return pct(e.Taken, e.Seats) }
func (e Event) IsDraft() bool   { return e.Status == BatchDraft }
func (e Event) IsLive() bool    { return e.Status == BatchLive }
func (e Event) IsPending() bool { return e.Status == BatchPending }

func pct(n, of int) int {
	if of <= 0 {
		return 0
	}
	return n * 100 / of
}

const eventSelect = `
SELECT b.id, b.trek_id, t.name, t.slug, t.region, t.price, b.start_date, b.seats, b.status,
       COALESCE(b.lead_id, 0), COALESCE(l.name, ''), t.active = 0,
       COALESCE((SELECT SUM(k.people) FROM bookings k WHERE k.batch_id = b.id AND k.status IN (` + confirmedStatuses + `)), 0),
       (SELECT COUNT(*) FROM bookings k WHERE k.batch_id = b.id AND k.status NOT IN ('rejected', 'cancelled')),
       (SELECT COUNT(*) FROM bookings k WHERE k.batch_id = b.id AND k.status = 'requested')
FROM batches b JOIN treks t ON t.id = b.trek_id LEFT JOIN leads l ON l.id = b.lead_id`

func scanEvent(s scanner) (Event, error) {
	var e Event
	err := s.Scan(&e.ID, &e.TrekID, &e.TrekName, &e.TrekSlug, &e.Region, &e.Price, &e.StartDate, &e.Seats, &e.Status,
		&e.LeadID, &e.LeadName, &e.TrekHidden, &e.Taken, &e.Active, &e.Requested)
	return e, err
}

func (a *App) queryEvents(where string, args ...any) ([]Event, error) {
	rows, err := a.db.Query(eventSelect+` `+where, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Event
	for rows.Next() {
		e, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (a *App) eventByID(id int64) (Event, error) {
	return scanEvent(a.db.QueryRow(eventSelect+` WHERE b.id = ?`, id))
}

var eventFilters = []struct{ Key, Label, Where string }{
	{"upcoming", "Upcoming", `WHERE b.start_date >= date('now') AND b.status IN ('live', 'draft', 'pending') ORDER BY b.start_date, b.id`},
	{"draft", "Drafts", `WHERE b.status = 'draft' ORDER BY b.start_date, b.id`},
	{"pending", "Proposed by leads", `WHERE b.status = 'pending' ORDER BY b.start_date, b.id`},
	{"past", "Past", `WHERE b.start_date < date('now') AND b.status = 'live' ORDER BY b.start_date DESC, b.id DESC`},
	{"all", "All", `ORDER BY b.start_date DESC, b.id DESC`},
}

func (a *App) consoleEvents(w http.ResponseWriter, r *http.Request) {
	filter := r.URL.Query().Get("show")
	where := eventFilters[0].Where
	found := false
	for _, f := range eventFilters {
		if f.Key == filter {
			where, found = f.Where, true
		}
	}
	if !found {
		filter = eventFilters[0].Key
	}
	events, err := a.queryEvents(where + ` LIMIT 300`)
	if err != nil {
		a.serverError(w, err)
		return
	}
	a.render(w, r, "console_events.html", map[string]any{"Events": events, "Filter": filter, "Filters": eventFilters})
}

func (a *App) renderEvent(w http.ResponseWriter, r *http.Request, e Event, errMsg string) {
	treks, err := a.listTreks(false)
	if err != nil {
		a.serverError(w, err)
		return
	}
	leads, err := a.listLeads(true)
	if err != nil {
		a.serverError(w, err)
		return
	}
	data := map[string]any{"Event": e, "Treks": treks, "Leads": leads, "Today": today(), "Back": r.URL.Path}
	if e.ID != 0 {
		if data["Bookings"], err = a.queryBookings(`WHERE k.batch_id = ? ORDER BY CASE k.status WHEN 'requested' THEN 0 WHEN 'approved' THEN 1 ELSE 2 END, k.id`, e.ID); err != nil {
			a.serverError(w, err)
			return
		}
	}
	if errMsg != "" {
		data["Err"] = errMsg
	}
	a.render(w, r, "console_event.html", data)
}

func (a *App) consoleEvent(w http.ResponseWriter, r *http.Request) {
	e := Event{Seats: 15, Status: BatchDraft}
	if id := pathID(r); id != 0 {
		var err error
		if e, err = a.eventByID(id); errors.Is(err, sql.ErrNoRows) {
			http.NotFound(w, r)
			return
		} else if err != nil {
			a.serverError(w, err)
			return
		}
	} else if r.PathValue("id") != "" {
		http.NotFound(w, r)
		return
	} else if t := formInt(r, "trek"); t > 0 {
		e.TrekID = int64(t) // /admin/events/new?trek=3 pre-selects a trek
	}
	a.renderEvent(w, r, e, "")
}

// consoleSaveEvent creates a draft event, or edits an existing one. The
// "publish" button saves and publishes in one go.
func (a *App) consoleSaveEvent(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	publish := r.FormValue("action") == "publish"
	e := Event{ID: id, TrekID: int64(formInt(r, "trek_id")), StartDate: r.FormValue("start_date"),
		Seats: formInt(r, "seats"), LeadID: int64(formInt(r, "lead_id")), Status: BatchDraft}
	if id != 0 {
		old, err := a.eventByID(id)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		e.TrekID, e.Status, e.Taken, e.Active = old.TrekID, old.Status, old.Taken, old.Active
		e.TrekName, e.Price = old.TrekName, old.Price
	}
	if _, err := time.Parse("2006-01-02", e.StartDate); err != nil {
		a.renderEvent(w, r, e, "Please pick a valid start date.")
		return
	}
	if e.Seats < 1 || e.Seats > 500 || e.Seats < e.Taken {
		msg := "Group size must be between 1 and 500."
		if e.Seats < e.Taken {
			msg = fmt.Sprintf("%d seats are already confirmed, so the group size can't go below that.", e.Taken)
		}
		a.renderEvent(w, r, e, msg)
		return
	}
	var n int
	if err := a.db.QueryRow(`SELECT COUNT(*) FROM treks WHERE id = ?`, e.TrekID).Scan(&n); err != nil || n == 0 {
		a.renderEvent(w, r, e, "Please pick a trek from the list.")
		return
	}
	if id == 0 {
		status := BatchDraft
		if publish {
			status = BatchLive
		}
		res, err := a.db.Exec(`INSERT INTO batches (trek_id, lead_id, start_date, seats, status) VALUES (?, ?, ?, ?, ?)`,
			e.TrekID, nullID(int(e.LeadID)), e.StartDate, e.Seats, status)
		if err != nil {
			a.serverError(w, err)
			return
		}
		id, _ = res.LastInsertId()
		if publish {
			redirect(w, r, "/admin/events/%d?msg=event_published", id)
		} else {
			redirect(w, r, "/admin/events/%d?msg=event_created", id)
		}
		return
	}
	if _, err := a.db.Exec(`UPDATE batches SET lead_id = ?, start_date = ?, seats = ? WHERE id = ?`,
		nullID(int(e.LeadID)), e.StartDate, e.Seats, id); err != nil {
		a.serverError(w, err)
		return
	}
	if publish && e.Status == BatchDraft {
		a.db.Exec(`UPDATE batches SET status = 'live' WHERE id = ? AND status = 'draft'`, id)
		redirect(w, r, "/admin/events/%d?msg=event_published", id)
		return
	}
	redirect(w, r, "/admin/events/%d?msg=event_saved", id)
}

func (a *App) consoleEventStatus(w http.ResponseWriter, r *http.Request) {
	e, err := a.eventByID(pathID(r))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	to := r.FormValue("status")
	switch {
	case to == BatchLive && (e.IsDraft() || e.IsPending()):
	case to == BatchDraft && e.IsLive():
		if e.Active > 0 {
			redirect(w, r, "/admin/events/%d?err=event_has_bookings", e.ID)
			return
		}
	default:
		http.Error(w, "That status change isn't allowed.", http.StatusBadRequest)
		return
	}
	if _, err := a.db.Exec(`UPDATE batches SET status = ? WHERE id = ?`, to, e.ID); err != nil {
		a.serverError(w, err)
		return
	}
	if to == BatchLive {
		redirect(w, r, "/admin/events/%d?msg=event_published", e.ID)
	} else {
		redirect(w, r, "/admin/events/%d?msg=event_unpublished", e.ID)
	}
}

func (a *App) consoleDeleteEvent(w http.ResponseWriter, r *http.Request) {
	e, err := a.eventByID(pathID(r))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	var n int
	a.db.QueryRow(`SELECT COUNT(*) FROM bookings WHERE batch_id = ?`, e.ID).Scan(&n)
	if n > 0 {
		redirect(w, r, "/admin/events/%d?err=batch_has_bookings", e.ID)
		return
	}
	if _, err := a.db.Exec(`DELETE FROM batches WHERE id = ?`, e.ID); err != nil {
		a.serverError(w, err)
		return
	}
	redirect(w, r, "/admin/events?msg=batch_deleted")
}

// ---- registrations ----

// registrationFilter reads ?event=, ?status= and ?scope= and returns the SQL
// that selects the matching bookings. By default only upcoming events show.
func registrationFilter(r *http.Request) (where string, args []any, f map[string]any) {
	q := r.URL.Query()
	event, status, scope := int64(formInt(r, "event")), q.Get("status"), q.Get("scope")
	if scope != "all" {
		scope = "current"
	}
	var conds []string
	switch {
	case event > 0:
		conds, args = append(conds, "b.id = ?"), append(args, event)
	case scope == "current":
		conds = append(conds, "b.start_date >= date('now')")
	}
	for _, s := range bookingStatuses {
		if s == status {
			conds, args = append(conds, "k.status = ?"), append(args, status)
		}
	}
	if len(conds) > 0 {
		where = "WHERE " + strings.Join(conds, " AND ")
	}
	return where, args, map[string]any{"Event": event, "Status": status, "Scope": scope}
}

func (a *App) consoleRegistrations(w http.ResponseWriter, r *http.Request) {
	where, args, f := registrationFilter(r)
	bookings, err := a.queryBookings(where+` ORDER BY b.start_date, CASE k.status WHEN 'requested' THEN 0 WHEN 'approved' THEN 1 ELSE 2 END, k.id LIMIT 1000`, args...)
	if err != nil {
		a.serverError(w, err)
		return
	}
	events, err := a.queryEvents(`WHERE b.start_date >= date('now') AND b.status = 'live' ORDER BY b.start_date, b.id`)
	if err != nil {
		a.serverError(w, err)
		return
	}
	people := 0
	for _, k := range bookings {
		if k.Status == StatusApproved || k.Status == StatusCompleted || k.Status == StatusRequested {
			people += k.People
		}
	}
	f["Bookings"], f["Events"], f["Statuses"], f["People"] = bookings, events, bookingStatuses, people
	f["Back"], f["ShowEvent"] = r.URL.RequestURI(), true
	f["Query"] = r.URL.RawQuery
	a.render(w, r, "console_registrations.html", f)
}

func (a *App) consoleRegistrationsCSV(w http.ResponseWriter, r *http.Request) {
	where, args, _ := registrationFilter(r)
	bookings, err := a.queryBookings(where+` ORDER BY b.start_date, k.id LIMIT 5000`, args...)
	if err != nil {
		a.serverError(w, err)
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="registrations-`+today()+`.csv"`)
	c := csv.NewWriter(w)
	c.Write([]string{"Booking", "Status", "Trek", "Date", "Name", "Phone", "Email", "People", "Age", "Pickup",
		"Emergency contact", "Emergency phone", "Medical", "Note", "Booked on"})
	for _, k := range bookings {
		email := k.UserEmail
		if strings.HasPrefix(email, "guest-") { // placeholder address of a guest booking
			email = ""
		}
		c.Write([]string{fmt.Sprint(k.ID), k.Status, k.TrekName, k.StartDate, csvSafe(k.UserName), k.UserPhone, email,
			fmt.Sprint(k.People), fmt.Sprint(k.Age), csvSafe(k.Pickup), csvSafe(k.EmergencyName), k.EmergencyPhone,
			csvSafe(k.Medical), csvSafe(k.Note), k.CreatedAt})
	}
	c.Flush()
}

// csvSafe stops spreadsheet apps from running a cell as a formula.
func csvSafe(s string) string {
	if s != "" && strings.ContainsRune("=+-@\t\r", rune(s[0])) {
		return "'" + s
	}
	return s
}

// ---- dashboard ----

func (a *App) consoleDashboard(w http.ResponseWriter, r *http.Request) {
	var d struct {
		Requested, Pending, Drafts, LeadApps, UpcomingLive int
		ConfirmedValue, UpcomingValue, Trekkers, WeekRegs  int
		SeatsTaken, SeatsTotal                             int
	}
	q := func(dest *int, sqlText string) { a.db.QueryRow(sqlText).Scan(dest) }
	q(&d.Requested, `SELECT COUNT(*) FROM bookings k JOIN batches b ON b.id = k.batch_id WHERE k.status = 'requested' AND b.start_date >= date('now')`)
	q(&d.Pending, `SELECT COUNT(*) FROM batches WHERE status = 'pending'`)
	q(&d.Drafts, `SELECT COUNT(*) FROM batches WHERE status = 'draft'`)
	q(&d.LeadApps, `SELECT COUNT(*) FROM leads WHERE status = 'pending'`)
	q(&d.UpcomingLive, `SELECT COUNT(*) FROM batches WHERE status = 'live' AND start_date >= date('now')`)
	q(&d.ConfirmedValue, `SELECT COALESCE(SUM(k.people * t.price), 0) FROM bookings k JOIN batches b ON b.id = k.batch_id JOIN treks t ON t.id = b.trek_id WHERE k.status IN (`+confirmedStatuses+`)`)
	q(&d.UpcomingValue, `SELECT COALESCE(SUM(k.people * t.price), 0) FROM bookings k JOIN batches b ON b.id = k.batch_id JOIN treks t ON t.id = b.trek_id WHERE k.status IN (`+confirmedStatuses+`) AND b.start_date >= date('now')`)
	q(&d.Trekkers, `SELECT COALESCE(SUM(people), 0) FROM bookings WHERE status IN (`+confirmedStatuses+`)`)
	q(&d.WeekRegs, `SELECT COUNT(*) FROM bookings WHERE created_at >= datetime('now', '-7 days')`)
	q(&d.SeatsTotal, `SELECT COALESCE(SUM(seats), 0) FROM batches WHERE status = 'live' AND start_date >= date('now')`)
	q(&d.SeatsTaken, `SELECT COALESCE(SUM(k.people), 0) FROM bookings k JOIN batches b ON b.id = k.batch_id WHERE k.status IN (`+confirmedStatuses+`) AND b.status = 'live' AND b.start_date >= date('now')`)

	upcoming, err := a.queryEvents(`WHERE b.start_date >= date('now') AND b.status IN ('live', 'draft') ORDER BY b.start_date, b.id LIMIT 6`)
	if err != nil {
		a.serverError(w, err)
		return
	}
	waiting, err := a.queryBookings(`WHERE k.status = 'requested' AND b.start_date >= date('now') ORDER BY k.id DESC LIMIT 6`)
	if err != nil {
		a.serverError(w, err)
		return
	}
	a.render(w, r, "console_dashboard.html", map[string]any{"D": d, "Upcoming": upcoming, "Bookings": waiting, "Back": "/admin",
		"Fill": pct(d.SeatsTaken, d.SeatsTotal)})
}
