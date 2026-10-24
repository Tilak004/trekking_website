package main

import (
	"database/sql"
	"errors"
	"net/http"
	"slices"
	"strings"
	"time"
)

// today returns the current UTC date, matching SQLite's date('now').
func today() string { return time.Now().UTC().Format("2006-01-02") }

func (a *App) adminSetBookingStatus(w http.ResponseWriter, r *http.Request) {
	status := r.FormValue("status")
	back := safeNext(r.FormValue("back"))
	if back == "/" {
		back = "/admin/registrations"
	}
	sep := "?"
	if strings.Contains(back, "?") {
		sep = "&"
	}
	list, err := a.queryBookings(`WHERE k.id = ?`, pathID(r))
	if err != nil {
		a.serverError(w, err)
		return
	}
	if len(list) == 0 || !slices.Contains(nextActions(list[0].Status), status) {
		http.Error(w, "That status change isn't allowed.", http.StatusBadRequest)
		return
	}
	k := list[0]
	if status == StatusApproved {
		batch, err := a.batchByID(k.BatchID)
		if err != nil {
			a.serverError(w, err)
			return
		}
		if k.People > batch.Left() {
			redirect(w, r, "%s%serr=full", back, sep)
			return
		}
	}
	if _, err := a.db.Exec(`UPDATE bookings SET status = ? WHERE id = ?`, status, k.ID); err != nil {
		a.serverError(w, err)
		return
	}
	if status == StatusApproved && r.FormValue("send_pay") == "1" {
		redirect(w, r, "/admin/bookings/%d/payreq", k.ID)
		return
	}
	redirect(w, r, "%s%smsg=status", back, sep)
}

// ---- treks ----

func (a *App) adminTreks(w http.ResponseWriter, r *http.Request) {
	treks, err := a.listTreks(false)
	if err != nil {
		a.serverError(w, err)
		return
	}
	pending, err := a.pendingBatches()
	if err != nil {
		a.serverError(w, err)
		return
	}
	a.render(w, r, "admin_treks.html", map[string]any{"Treks": treks, "Pending": pending})
}

func (a *App) adminTrekForm(w http.ResponseWriter, r *http.Request) {
	t := Trek{Active: true, Difficulty: "Moderate",
		PickupPoints: defaultPickups}
	var batches []Batch
	if id := pathID(r); id != 0 {
		var err error
		t, err = a.trekWhere(`t.id = ?`, id)
		if errors.Is(err, sql.ErrNoRows) {
			http.NotFound(w, r)
			return
		} else if err != nil {
			a.serverError(w, err)
			return
		}
		if batches, err = a.batchesForTrek(id, false); err != nil {
			a.serverError(w, err)
			return
		}
	} else if r.PathValue("id") != "" {
		http.NotFound(w, r)
		return
	} else {
		// A new trek starts with the common text from Trek defaults.
		d := a.trekDefaults()
		t.Exclusions, t.ThingsToCarry, t.FAQs = d.Exclusions, d.ThingsToCarry, d.FAQs
	}
	a.renderTrekForm(w, r, t, batches, "")
}

func (a *App) renderTrekForm(w http.ResponseWriter, r *http.Request, t Trek, batches []Batch, errMsg string) {
	leads, err := a.listLeads(true)
	if err != nil {
		a.serverError(w, err)
		return
	}
	data := map[string]any{"Trek": t, "Batches": batches, "Leads": leads, "Today": today(), "Photos": a.listPhotos(""), "Defaults": a.trekDefaults(), "Library": trekLibrary, "WhatsAppGroup": a.trekGroup(t.ID),
		"Difficulties": []string{"Easy", "Easy–Moderate", "Moderate", "Difficult", "Very Difficult"}}
	if errMsg != "" {
		data["Err"] = errMsg
	}
	a.render(w, r, "admin_trek_form.html", data)
}

func (a *App) adminSaveTrek(w http.ResponseWriter, r *http.Request) {
	t := Trek{
		ID:            pathID(r),
		Name:          trimTo(r.FormValue("name"), 120),
		Slug:          slugify(r.FormValue("slug")),
		Region:        trimTo(r.FormValue("region"), 120),
		Duration:      trimTo(r.FormValue("duration"), 60),
		DistanceKm:    formInt(r, "distance_km"),
		MaxAltitude:   formInt(r, "max_altitude"),
		Difficulty:    trimTo(r.FormValue("difficulty"), 40),
		Price:         formInt(r, "price"),
		Summary:       trimTo(r.FormValue("summary"), 300),
		Description:   trimTo(r.FormValue("description"), 10000),
		PickupPoints:  strings.Join(splitLines(r.FormValue("pickup_points")), "\n"),
		Photo:         r.FormValue("photo"),
		BestSeason:    trimTo(r.FormValue("best_season"), 80),
		Itinerary:     trimTo(r.FormValue("itinerary"), 10000),
		Inclusions:    strings.Join(splitLines(r.FormValue("inclusions")), "\n"),
		Exclusions:    strings.Join(splitLines(r.FormValue("exclusions")), "\n"),
		ThingsToCarry: strings.Join(splitLines(r.FormValue("things_to_carry")), "\n"),
		FAQs:          trimTo(r.FormValue("faqs"), 10000),
		Lat:           formFloat(r, "lat"),
		Lng:           formFloat(r, "lng"),
		Tags:          strings.Join(splitTags(r.FormValue("tags")), ", "),
		Elevation:     trimTo(r.FormValue("elevation"), 5000),
		Active:        r.FormValue("active") == "on",
	}
	if t.Photo != "" && !slices.Contains(a.listPhotos(""), t.Photo) {
		t.Photo = ""
	}
	if t.Slug == "" {
		t.Slug = slugify(t.Name)
	}
	if t.Name == "" || t.Slug == "" || t.Price < 0 {
		a.renderTrekForm(w, r, t, nil, "Name is required and price can't be negative.")
		return
	}
	args := []any{t.Slug, t.Name, t.Region, t.Duration, t.DistanceKm, t.MaxAltitude, t.Difficulty, t.Price,
		t.Summary, t.Description, t.PickupPoints, t.Photo, t.Active,
		t.BestSeason, t.Itinerary, t.Inclusions, t.Exclusions, t.ThingsToCarry, t.FAQs, t.Lat, t.Lng, t.Tags, t.Elevation}
	var err error
	if t.ID == 0 {
		var res sql.Result
		res, err = a.db.Exec(`INSERT INTO treks (slug, name, region, duration, distance_km, max_altitude, difficulty, price,
summary, description, pickup_points, photo, active, best_season, itinerary, inclusions, exclusions, things_to_carry, faqs, lat, lng, tags, elevation)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, args...)
		if err == nil {
			t.ID, _ = res.LastInsertId()
		}
	} else {
		_, err = a.db.Exec(`UPDATE treks SET slug = ?, name = ?, region = ?, duration = ?, distance_km = ?, max_altitude = ?,
difficulty = ?, price = ?, summary = ?, description = ?, pickup_points = ?, photo = ?, active = ?,
best_season = ?, itinerary = ?, inclusions = ?, exclusions = ?, things_to_carry = ?, faqs = ?,
lat = ?, lng = ?, tags = ?, elevation = ? WHERE id = ?`, append(args, t.ID)...)
	}
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			batches, _ := a.batchesForTrek(t.ID, false)
			a.renderTrekForm(w, r, t, batches, "Another trek already uses this URL slug. Pick a different one.")
			return
		}
		a.serverError(w, err)
		return
	}
	if err := a.saveTrekGroup(t.ID, r); err != nil {
		redirect(w, r, "/admin/treks/%d?err=bad_group_link", t.ID)
		return
	}
	redirect(w, r, "/admin/treks/%d?msg=saved", t.ID)
}

func (a *App) adminAddBatch(w http.ResponseWriter, r *http.Request) {
	trekID := pathID(r)
	date := r.FormValue("start_date")
	seats := formInt(r, "seats")
	if _, err := time.Parse("2006-01-02", date); err != nil || seats < 1 {
		http.Error(w, "Please give a valid date and at least 1 seat.", http.StatusBadRequest)
		return
	}
	// New dates start as drafts; publish them from the event page.
	res, err := a.db.Exec(`INSERT INTO batches (trek_id, lead_id, start_date, seats, status) VALUES (?, ?, ?, ?, 'draft')`,
		trekID, nullID(formInt(r, "lead_id")), date, seats)
	if err != nil {
		a.serverError(w, err)
		return
	}
	id, _ := res.LastInsertId()
	redirect(w, r, "/admin/events/%d?msg=event_created", id)
}

func (a *App) adminSetBatchLead(w http.ResponseWriter, r *http.Request) {
	b, err := a.batchByID(pathID(r))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	seats := formInt(r, "seats")
	if seats < b.Taken || seats < 1 {
		seats = b.Seats
	}
	if _, err := a.db.Exec(`UPDATE batches SET lead_id = ?, seats = ? WHERE id = ?`,
		nullID(formInt(r, "lead_id")), seats, b.ID); err != nil {
		a.serverError(w, err)
		return
	}
	redirect(w, r, "/admin/treks/%d?msg=batch_updated", b.TrekID)
}

func (a *App) adminDeleteBatch(w http.ResponseWriter, r *http.Request) {
	b, err := a.batchByID(pathID(r))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if b.Bookings > 0 {
		redirect(w, r, "/admin/treks/%d?err=batch_has_bookings", b.TrekID)
		return
	}
	if _, err := a.db.Exec(`DELETE FROM batches WHERE id = ?`, b.ID); err != nil {
		a.serverError(w, err)
		return
	}
	redirect(w, r, "/admin/treks/%d?msg=batch_deleted", b.TrekID)
}

// adminDeleteTrek deletes a trek and its dates. Treks with any bookings can't be
// deleted (that would lose booking and payment records); hide them instead.
func (a *App) adminDeleteTrek(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	var bookings int
	if err := a.db.QueryRow(`SELECT COUNT(*) FROM bookings k JOIN batches b ON b.id = k.batch_id WHERE b.trek_id = ?`, id).Scan(&bookings); err != nil {
		a.serverError(w, err)
		return
	}
	if bookings > 0 {
		redirect(w, r, "/admin/treks/%d?err=trek_has_bookings", id)
		return
	}
	tx, err := a.db.Begin()
	if err != nil {
		a.serverError(w, err)
		return
	}
	defer tx.Rollback()
	for _, q := range []string{
		`UPDATE enquiries SET trek_id = NULL WHERE trek_id = ?`,
		`DELETE FROM batches WHERE trek_id = ?`,
		`DELETE FROM treks WHERE id = ?`,
	} {
		if _, err := tx.Exec(q, id); err != nil {
			a.serverError(w, err)
			return
		}
	}
	if err := tx.Commit(); err != nil {
		a.serverError(w, err)
		return
	}
	redirect(w, r, "/admin/treks?msg=trek_deleted")
}

func nullID(id int) any {
	if id <= 0 {
		return nil
	}
	return id
}

// ---- leads ----

func (a *App) adminLeads(w http.ResponseWriter, r *http.Request) {
	leads, err := a.listLeads(false)
	if err != nil {
		a.serverError(w, err)
		return
	}
	batches, err := a.assignableBatches()
	if err != nil {
		a.serverError(w, err)
		return
	}
	a.render(w, r, "admin_leads.html", map[string]any{"Leads": leads, "Batches": batches})
}

func (a *App) adminLeadForm(w http.ResponseWriter, r *http.Request) {
	l := Lead{Active: true}
	if id := pathID(r); id != 0 {
		var err error
		l, err = a.leadWhere(`l.id = ?`, id)
		if errors.Is(err, sql.ErrNoRows) {
			http.NotFound(w, r)
			return
		} else if err != nil {
			a.serverError(w, err)
			return
		}
	} else if r.PathValue("id") != "" {
		http.NotFound(w, r)
		return
	}
	a.render(w, r, "admin_lead_form.html", map[string]any{"Lead": l, "Photos": a.listPhotos("")})
}

func (a *App) adminSaveLead(w http.ResponseWriter, r *http.Request) {
	l := Lead{
		ID:             pathID(r),
		Name:           trimTo(r.FormValue("name"), 100),
		Slug:           slugify(r.FormValue("slug")),
		Headline:       trimTo(r.FormValue("headline"), 150),
		Bio:            trimTo(r.FormValue("bio"), 5000),
		Achievements:   strings.Join(splitLines(r.FormValue("achievements")), "\n"),
		Phone:          trimTo(r.FormValue("phone"), 30),
		Photo:          r.FormValue("photo"),
		Active:         r.FormValue("active") == "on",
		City:           trimTo(r.FormValue("city"), 80),
		Certifications: strings.Join(splitLines(r.FormValue("certifications")), "\n"),
		Verified:       r.FormValue("verified") == "on",
	}
	if l.Slug == "" {
		l.Slug = slugify(l.Name)
	}
	if l.Photo != "" && !slices.Contains(a.listPhotos(""), l.Photo) {
		l.Photo = ""
	}
	fail := func(msg string) {
		a.render(w, r, "admin_lead_form.html", map[string]any{"Lead": l, "Err": msg, "Photos": a.listPhotos("")})
	}
	if l.Name == "" || l.Slug == "" {
		fail("Name is required.")
		return
	}
	args := []any{l.Slug, l.Name, l.Headline, l.Bio, l.Achievements, l.Phone, l.Photo, l.Active, l.City, l.Certifications, l.Verified}
	var err error
	if l.ID == 0 {
		var res sql.Result
		res, err = a.db.Exec(`INSERT INTO leads (slug, name, headline, bio, achievements, phone, photo, active, city, certifications, verified)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, args...)
		if err == nil {
			l.ID, _ = res.LastInsertId()
		}
	} else {
		_, err = a.db.Exec(`UPDATE leads SET slug = ?, name = ?, headline = ?, bio = ?, achievements = ?, phone = ?, photo = ?, active = ?,
city = ?, certifications = ?, verified = ? WHERE id = ?`,
			append(args, l.ID)...)
	}
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			fail("Another lead already uses this URL slug. Pick a different one.")
			return
		}
		a.serverError(w, err)
		return
	}
	redirect(w, r, "/admin/leads/%d?msg=saved", l.ID)
}

// ---- users ----

func (a *App) adminUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := a.db.Query(`SELECT u.id, u.name, u.email, u.phone, u.is_admin, u.created_at,
(SELECT COUNT(*) FROM bookings k WHERE k.user_id = u.id) FROM users u WHERE u.is_admin = 0 ORDER BY u.id DESC LIMIT 1000`)
	if err != nil {
		a.serverError(w, err)
		return
	}
	defer rows.Close()
	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Phone, &u.IsAdmin, &u.CreatedAt, &u.Bookings); err != nil {
			a.serverError(w, err)
			return
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		a.serverError(w, err)
		return
	}
	a.render(w, r, "admin_users.html", map[string]any{"Users": users})
}
