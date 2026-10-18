package main

import (
	"database/sql"
	"net/http"
	"time"
)

// Lead dashboard: approved trek leads see the treks they lead (with the
// participant list they need on the trail) and propose new dates on existing
// treks. Proposed dates stay hidden until an admin approves them; the price is
// always the trek's own price.

// Batch statuses. Only "live" batches are shown publicly or bookable.
const (
	BatchLive     = "live"
	BatchPending  = "pending"
	BatchRejected = "rejected"
)

func migrateLeadDates(db *sql.DB) {
	db.Exec(`ALTER TABLE batches ADD COLUMN status TEXT NOT NULL DEFAULT 'live'`)
	db.Exec(`ALTER TABLE batches ADD COLUMN proposed_at TEXT NOT NULL DEFAULT ''`)
}

type DashBatch struct {
	Batch
	TrekName, TrekSlug string
	Participants       []Booking
}

// approvedLead returns the logged-in user's lead profile if it's approved.
func (a *App) approvedLead(r *http.Request) (Lead, bool, error) {
	l, ok, err := a.myLead(currentUser(r).ID)
	if err != nil || !ok || l.Status != LeadApproved {
		return Lead{}, false, err
	}
	return l, true, nil
}

func (a *App) leadDashboard(w http.ResponseWriter, r *http.Request) {
	lead, ok, err := a.approvedLead(r)
	if err != nil {
		a.serverError(w, err)
		return
	}
	if !ok {
		redirect(w, r, "/lead-with-us")
		return
	}
	treks, err := a.listTreks(true)
	if err != nil {
		a.serverError(w, err)
		return
	}
	all, err := a.listTreks(false)
	if err != nil {
		a.serverError(w, err)
		return
	}
	names := map[int64]Trek{}
	for _, t := range all {
		names[t.ID] = t
	}

	batches, err := a.queryBatches(batchSelect+` WHERE b.lead_id = ? ORDER BY b.start_date`, lead.ID)
	if err != nil {
		a.serverError(w, err)
		return
	}
	var upcoming, past []DashBatch
	for _, b := range batches {
		d := DashBatch{Batch: b, TrekName: names[b.TrekID].Name, TrekSlug: names[b.TrekID].Slug}
		if b.StartDate < today() {
			past = append([]DashBatch{d}, past...)
			continue
		}
		if b.Status == BatchLive {
			d.Participants, err = a.queryBookings(`WHERE k.batch_id = ? AND k.status IN ('requested', 'approved') ORDER BY k.status, k.id`, b.ID)
			if err != nil {
				a.serverError(w, err)
				return
			}
		}
		upcoming = append(upcoming, d)
	}
	if len(past) > 10 {
		past = past[:10]
	}
	a.render(w, r, "lead_dashboard.html", map[string]any{
		"Lead": lead, "Upcoming": upcoming, "Past": past, "Treks": treks,
		"MinDate": time.Now().AddDate(0, 0, 3).Format("2006-01-02"),
	})
}

func (a *App) leadProposeDate(w http.ResponseWriter, r *http.Request) {
	lead, ok, err := a.approvedLead(r)
	if err != nil {
		a.serverError(w, err)
		return
	}
	if !ok {
		redirect(w, r, "/lead-with-us")
		return
	}
	trekID := int64(formInt(r, "trek_id"))
	date := r.FormValue("start_date")
	seats := formInt(r, "seats")
	d, perr := time.Parse("2006-01-02", date)
	switch {
	case perr != nil || d.Before(time.Now().AddDate(0, 0, 2)):
		redirect(w, r, "/lead?err=bad_date")
		return
	case seats < 2 || seats > 40:
		redirect(w, r, "/lead?err=bad_seats")
		return
	}
	if _, err := a.trekWhere(`t.id = ? AND t.active = 1`, trekID); err != nil {
		redirect(w, r, "/lead?err=bad_trek")
		return
	}
	var dup int
	a.db.QueryRow(`SELECT COUNT(*) FROM batches WHERE trek_id = ? AND lead_id = ? AND start_date = ? AND status != 'rejected'`,
		trekID, lead.ID, date).Scan(&dup)
	if dup > 0 {
		redirect(w, r, "/lead?err=dup_date")
		return
	}
	if _, err := a.db.Exec(`INSERT INTO batches (trek_id, lead_id, start_date, seats, status, proposed_at)
VALUES (?, ?, ?, ?, 'pending', datetime('now'))`, trekID, lead.ID, date, seats); err != nil {
		a.serverError(w, err)
		return
	}
	redirect(w, r, "/lead?msg=proposed")
}

// leadWithdrawDate lets a lead take back a date that hasn't been approved yet.
func (a *App) leadWithdrawDate(w http.ResponseWriter, r *http.Request) {
	lead, ok, err := a.approvedLead(r)
	if err != nil {
		a.serverError(w, err)
		return
	}
	if !ok {
		redirect(w, r, "/lead-with-us")
		return
	}
	if _, err := a.db.Exec(`DELETE FROM batches WHERE id = ? AND lead_id = ? AND status IN ('pending', 'rejected')
AND NOT EXISTS (SELECT 1 FROM bookings WHERE batch_id = batches.id)`, pathID(r), lead.ID); err != nil {
		a.serverError(w, err)
		return
	}
	redirect(w, r, "/lead?msg=withdrawn")
}

// pendingBatches are lead-proposed dates waiting for an admin.
type PendingBatch struct {
	Batch
	TrekName, ProposedAt string
}

func (a *App) pendingBatches() ([]PendingBatch, error) {
	rows, err := a.db.Query(`SELECT b.id, b.start_date, b.seats, COALESCE(l.name, ''), COALESCE(l.slug, ''), t.name, b.proposed_at
FROM batches b JOIN treks t ON t.id = b.trek_id LEFT JOIN leads l ON l.id = b.lead_id
WHERE b.status = 'pending' ORDER BY b.start_date`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []PendingBatch
	for rows.Next() {
		var p PendingBatch
		if err := rows.Scan(&p.ID, &p.StartDate, &p.Seats, &p.LeadName, &p.LeadSlug, &p.TrekName, &p.ProposedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (a *App) adminReviewBatch(w http.ResponseWriter, r *http.Request) {
	status := r.FormValue("status")
	if status != BatchLive && status != BatchRejected {
		http.Error(w, "Unknown status.", http.StatusBadRequest)
		return
	}
	if _, err := a.db.Exec(`UPDATE batches SET status = ? WHERE id = ? AND status = 'pending'`, status, pathID(r)); err != nil {
		a.serverError(w, err)
		return
	}
	redirect(w, r, "/admin/events?show=pending&msg=date_reviewed")
}
