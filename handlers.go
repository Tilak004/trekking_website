package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func (a *App) home(w http.ResponseWriter, r *http.Request) {
	treks, err := a.listTreks(true)
	if err != nil {
		a.serverError(w, err)
		return
	}
	leads, err := a.listLeads(true)
	if err != nil {
		a.serverError(w, err)
		return
	}
	lead, err := a.nextLeads()
	if err != nil {
		a.serverError(w, err)
		return
	}
	finder := []FinderTrek{}
	for i := range treks {
		if l, ok := lead[treks[i].ID]; ok {
			treks[i].Lead = &l
		}
	}
	var upcoming, soon []Trek
	for _, t := range treks {
		if t.NextBatch != "" {
			upcoming = append(upcoming, t)
		} else {
			soon = append(soon, t)
		}
	}
	minPrice := 0
	for _, t := range treks {
		if t.Price > 0 && (minPrice == 0 || t.Price < minPrice) {
			minPrice = t.Price
		}
	}
	// Bookable treks first, the same order the server-rendered grid uses.
	for _, t := range append(append([]Trek{}, upcoming...), soon...) {
		finder = append(finder, finderTrek(t))
	}
	stats, err := a.trustStats()
	if err != nil {
		a.serverError(w, err)
		return
	}
	photoReviews, err := a.queryReviews(`r.photo != ''`, nil)
	if err != nil {
		a.serverError(w, err)
		return
	}
	if len(photoReviews) > 6 {
		photoReviews = photoReviews[:6]
	}
	a.render(w, r, "home.html", map[string]any{
		"FinderData": finder, "Stats": stats, "PhotoReviews": photoReviews,
		"Treks": treks, "Upcoming": upcoming, "Soon": soon, "Leads": leads, "MinPrice": minPrice, "HeroPhoto": a.heroPhoto(), "HeroSlides": a.heroSlides(treks), "HeroLayers": a.heroLayers(), "HeroStyle": a.heroStyle(), "Gallery": a.listPhotos("gallery"),
	})
}

func (a *App) trekPage(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	t, err := a.trekWhere(`t.slug = ?`, r.PathValue("slug"))
	if errors.Is(err, sql.ErrNoRows) || (err == nil && !t.Active && (u == nil || !u.IsAdmin)) {
		a.notFound(w, r)
		return
	} else if err != nil {
		a.serverError(w, err)
		return
	}
	batches, err := a.batchesForTrek(t.ID, true)
	if err != nil {
		a.serverError(w, err)
		return
	}
	reviews, err := a.queryReviews(`b.trek_id = ?`, t.ID)
	if err != nil {
		a.serverError(w, err)
		return
	}
	data := map[string]any{"Trek": t, "Batches": batches, "Reviews": reviews, "Existing": nil, "MetaDesc": t.Summary,
		"BookingData": bookingData(t, batches, u)}
	if t.Photo != "" {
		data["MetaImage"] = "/photos/" + t.Photo
	}
	if u != nil {
		existing, err := a.queryBookings(`WHERE k.user_id = ? AND t.id = ? AND k.status IN ('requested', 'approved') LIMIT 1`, u.ID, t.ID)
		if err != nil {
			a.serverError(w, err)
			return
		}
		if len(existing) > 0 {
			data["Existing"] = existing[0]
		}
	}
	a.render(w, r, "trek.html", data)
}

func (a *App) bookTrek(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	t, err := a.trekWhere(`t.slug = ? AND t.active = 1`, r.PathValue("slug"))
	if errors.Is(err, sql.ErrNoRows) {
		a.notFound(w, r)
		return
	} else if err != nil {
		a.serverError(w, err)
		return
	}
	back := "/treks/" + t.Slug

	// No account needed: guests book with just a name and phone number. They get a
	// lightweight guest account (and are logged in) so "My bookings" still works.
	if u == nil {
		u, err = a.guestUser(w, r, trimTo(r.FormValue("name"), 100), r.FormValue("phone"))
		if err != nil {
			redirect(w, r, "%s?err=bad_contact", back)
			return
		}
	}

	a.expireHolds() // seats from timed-out holds are free again
	batch, err := a.batchByID(int64(formInt(r, "batch_id")))
	if err != nil || batch.TrekID != t.ID || batch.Status != BatchLive || batch.StartDate < today() {
		redirect(w, r, "%s?err=bad_batch", back)
		return
	}

	people := formInt(r, "people")
	pickup := trimTo(r.FormValue("pickup"), 100)
	emName := trimTo(r.FormValue("emergency_name"), 100)
	emPhone := trimTo(r.FormValue("emergency_phone"), 30)
	age := formInt(r, "age")
	if people < 1 || people > 10 || (age != 0 && (age < 5 || age > 100)) ||
		(pickup != "" && !validPickup(t.PickupPoints, pickup)) {
		// The form enforces these with "required" fields, so this only happens on tampered input.
		http.Error(w, "Please fill in all required booking details.", http.StatusBadRequest)
		return
	}
	if people > batch.Left() {
		redirect(w, r, "%s?err=full", back)
		return
	}
	var active int
	if err := a.db.QueryRow(`SELECT COUNT(*) FROM bookings k JOIN batches b ON b.id = k.batch_id
WHERE k.user_id = ? AND b.trek_id = ? AND k.status IN ('requested', 'approved')`, u.ID, t.ID).Scan(&active); err != nil {
		a.serverError(w, err)
		return
	}
	if active > 0 {
		redirect(w, r, "%s?err=already_booked", back)
		return
	}

	// Instant booking: the seat is held for holdHours while the trekker pays. The seat
	// check and the insert are one statement, so two people can't take the last seat.
	token := randomToken(8)
	res, err := a.db.Exec(`INSERT INTO bookings (user_id, batch_id, people, pickup, emergency_name, emergency_phone, age, medical, note,
                      status, hold_until, pay_token)
SELECT ?, b.id, ?, ?, ?, ?, ?, ?, ?, 'approved', datetime('now', '+`+strconv.Itoa(holdHours)+` hours'), ?
FROM batches b
WHERE b.id = ? AND b.seats - (SELECT COALESCE(SUM(k.people), 0) FROM bookings k
                              WHERE k.batch_id = b.id AND k.status IN ('approved', 'completed')) >= ?`,
		u.ID, people, pickup, emName, emPhone, age, trimTo(r.FormValue("medical"), 1000), trimTo(r.FormValue("note"), 1000),
		token, batch.ID, people)
	if err != nil {
		a.serverError(w, err)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		redirect(w, r, "%s?err=full", back)
		return
	}
	id, _ := res.LastInsertId()
	redirect(w, r, "/pay/%d/%s?msg=held", id, token)
}

// bookingSent asks the trekker to forward their new request to us on WhatsApp,
// which is how the team gets notified for now.
func (a *App) bookingSent(w http.ResponseWriter, r *http.Request) {
	list, err := a.queryBookings(`WHERE k.id = ? AND k.user_id = ?`, pathID(r), currentUser(r).ID)
	if err != nil {
		a.serverError(w, err)
		return
	}
	if len(list) == 0 {
		a.notFound(w, r)
		return
	}
	a.render(w, r, "booking_sent.html", map[string]any{"Booking": list[0]})
}

// bookingMessage is the pre-filled WhatsApp text for a booking request.
func bookingMessage(k Booking) string {
	lines := []string{
		fmt.Sprintf("Hi %s! I've requested a booking (#%d).", site.Name, k.ID),
		"",
		"Trek: " + k.TrekName,
		"Date: " + formatDate(k.StartDate),
		fmt.Sprintf("People: %d", k.People),
	}
	if k.Pickup != "" {
		lines = append(lines, "Pickup: "+k.Pickup)
	}
	lines = append(lines, "Name: "+k.UserName, "Phone: "+k.UserPhone)
	if k.Medical != "" {
		lines = append(lines, "Medical notes: "+k.Medical)
	}
	if k.Note != "" {
		lines = append(lines, "Note: "+k.Note)
	}
	lines = append(lines, "", "Please confirm my seat. Thanks!")
	return strings.Join(lines, "\n")
}

func (a *App) galleryPage(w http.ResponseWriter, r *http.Request) {
	a.render(w, r, "gallery.html", map[string]any{"Gallery": a.listPhotos("gallery")})
}

func (a *App) leadsPage(w http.ResponseWriter, r *http.Request) {
	leads, err := a.listLeads(true)
	if err != nil {
		a.serverError(w, err)
		return
	}
	a.render(w, r, "leads.html", map[string]any{"Leads": leads})
}

func (a *App) leadPage(w http.ResponseWriter, r *http.Request) {
	l, err := a.leadWhere(`l.slug = ? AND l.active = 1 AND l.status = 'approved'`, r.PathValue("slug"))
	if errors.Is(err, sql.ErrNoRows) {
		a.notFound(w, r)
		return
	} else if err != nil {
		a.serverError(w, err)
		return
	}
	upcoming, err := a.queryUpcomingForLead(l.ID)
	if err != nil {
		a.serverError(w, err)
		return
	}
	reviews, err := a.queryReviews(`b.lead_id = ? AND r.lead_rating > 0`, l.ID)
	if err != nil {
		a.serverError(w, err)
		return
	}
	a.render(w, r, "lead.html", map[string]any{"Lead": l, "Upcoming": upcoming, "Reviews": reviews, "MetaDesc": l.Headline})
}

type LeadBatch struct {
	TrekName, TrekSlug, StartDate string
}

func (a *App) queryUpcomingForLead(leadID int64) ([]LeadBatch, error) {
	rows, err := a.db.Query(`SELECT t.name, t.slug, b.start_date FROM batches b JOIN treks t ON t.id = b.trek_id
WHERE b.lead_id = ? AND t.active = 1 AND b.status = 'live' AND b.start_date >= date('now') ORDER BY b.start_date LIMIT 20`, leadID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []LeadBatch
	for rows.Next() {
		var lb LeadBatch
		if err := rows.Scan(&lb.TrekName, &lb.TrekSlug, &lb.StartDate); err != nil {
			return nil, err
		}
		out = append(out, lb)
	}
	return out, rows.Err()
}

func (a *App) myBookings(w http.ResponseWriter, r *http.Request) {
	a.expireHolds()
	bookings, err := a.queryBookings(`WHERE k.user_id = ? ORDER BY b.start_date DESC`, currentUser(r).ID)
	if err != nil {
		a.serverError(w, err)
		return
	}
	a.render(w, r, "my.html", map[string]any{"Bookings": bookings})
}

func (a *App) cancelBooking(w http.ResponseWriter, r *http.Request) {
	list, err := a.queryBookings(`WHERE k.id = ? AND k.user_id = ? AND k.status IN ('requested', 'approved')`, pathID(r), currentUser(r).ID)
	if err != nil {
		a.serverError(w, err)
		return
	}
	if len(list) > 0 {
		if err := a.cancelWithRefund(list[0], RefundCustomer); err != nil {
			a.serverError(w, err)
			return
		}
	}
	redirect(w, r, "/my?msg=cancelled")
}

func (a *App) reviewBooking(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	r.Body = http.MaxBytesReader(w, r.Body, maxPhotoBytes+(1<<20))
	if err := r.ParseMultipartForm(maxPhotoBytes); err != nil && !errors.Is(err, http.ErrNotMultipart) {
		redirect(w, r, "/my?err=bad_photo")
		return
	}
	list, err := a.queryBookings(`WHERE k.id = ? AND k.user_id = ?`, pathID(r), u.ID)
	if err != nil {
		a.serverError(w, err)
		return
	}
	if len(list) == 0 || list[0].Status != StatusCompleted || list[0].Reviewed {
		redirect(w, r, "/my?err=cannot_review")
		return
	}
	k := list[0]
	trekRating, leadRating := formInt(r, "trek_rating"), 0
	if k.HasLead {
		leadRating = formInt(r, "lead_rating")
		if leadRating < 1 || leadRating > 5 {
			redirect(w, r, "/my?err=bad_review")
			return
		}
	}
	if trekRating < 1 || trekRating > 5 {
		redirect(w, r, "/my?err=bad_review")
		return
	}
	photo := ""
	if file, _, err := r.FormFile("photo"); err == nil {
		defer file.Close()
		if photo, err = a.savePhoto(file, "reviews", k.TrekSlug); err != nil {
			redirect(w, r, "/my?err=bad_photo")
			return
		}
	}
	if _, err := a.db.Exec(`INSERT INTO reviews (booking_id, trek_rating, lead_rating, comment, photo) VALUES (?, ?, ?, ?, ?)`,
		k.ID, trekRating, leadRating, trimTo(r.FormValue("comment"), 1000), photo); err != nil {
		a.serverError(w, err)
		return
	}
	redirect(w, r, "/my?msg=reviewed")
}

// validPickup reports whether pickup is one of the trek's pickup points. A point
// may carry a time ("Majestic | 11:00 PM"), and the booking form sends its label
// ("Majestic · 11:00 PM"), so both the label and the bare place are accepted.
func validPickup(pickupPoints, pickup string) bool {
	for _, p := range parsePickups(pickupPoints) {
		if pickup == p.Label() || pickup == p.Place {
			return true
		}
	}
	return false
}
