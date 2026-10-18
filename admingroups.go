package main

import (
	"net/http"
	"sort"
)

// Admin views that group information the way the team works: bookings by trek
// and date, and trek leads with their full profile and the dates they lead.

type DateGroup struct {
	Batch    Batch
	Bookings []Booking
	Due      int // unpaid amount on approved bookings
}

type TrekGroup struct {
	Name, Slug string
	Dates      []*DateGroup
	Count      int
}

// groupBookings arranges bookings as trek → date (soonest first) → bookings.
func (a *App) groupBookings(bookings []Booking) ([]*TrekGroup, error) {
	batches := map[int64]Batch{}
	byTrek := map[string]*TrekGroup{}
	byDate := map[int64]*DateGroup{}
	var treks []*TrekGroup
	for _, k := range bookings {
		if _, ok := batches[k.BatchID]; !ok {
			b, err := a.batchByID(k.BatchID)
			if err != nil {
				return nil, err
			}
			batches[k.BatchID] = b
		}
		tg := byTrek[k.TrekSlug]
		if tg == nil {
			tg = &TrekGroup{Name: k.TrekName, Slug: k.TrekSlug}
			byTrek[k.TrekSlug] = tg
			treks = append(treks, tg)
		}
		dg := byDate[k.BatchID]
		if dg == nil {
			dg = &DateGroup{Batch: batches[k.BatchID]}
			byDate[k.BatchID] = dg
			tg.Dates = append(tg.Dates, dg)
		}
		dg.Bookings = append(dg.Bookings, k)
		if k.Status == StatusApproved && k.PaidAt == "" {
			dg.Due += k.Amount()
		}
		tg.Count++
	}
	today := today()
	for _, tg := range treks {
		// Upcoming dates first (soonest at the top), then past dates (most recent first).
		sort.SliceStable(tg.Dates, func(i, j int) bool {
			a, b := tg.Dates[i].Batch.StartDate, tg.Dates[j].Batch.StartDate
			au, bu := a >= today, b >= today
			if au != bu {
				return au
			}
			if au {
				return a < b
			}
			return a > b
		})
	}
	sort.SliceStable(treks, func(i, j int) bool { return treks[i].Name < treks[j].Name })
	return treks, nil
}

// AssignableBatch is an upcoming live trek date an admin can give to a lead.
type AssignableBatch struct {
	ID                  int64
	TrekName, StartDate string
	LeadID              int64
	LeadName            string
}

func (a *App) assignableBatches() ([]AssignableBatch, error) {
	rows, err := a.db.Query(`SELECT b.id, t.name, b.start_date, COALESCE(b.lead_id, 0), COALESCE(l.name, '')
FROM batches b JOIN treks t ON t.id = b.trek_id LEFT JOIN leads l ON l.id = b.lead_id
WHERE b.status = 'live' AND b.start_date >= date('now') ORDER BY b.start_date, t.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []AssignableBatch
	for rows.Next() {
		var b AssignableBatch
		if err := rows.Scan(&b.ID, &b.TrekName, &b.StartDate, &b.LeadID, &b.LeadName); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

// adminAssignLead puts a lead in charge of an upcoming trek date.
func (a *App) adminAssignLead(w http.ResponseWriter, r *http.Request) {
	leadID := pathID(r)
	var ok int
	a.db.QueryRow(`SELECT COUNT(*) FROM leads WHERE id = ? AND status = 'approved'`, leadID).Scan(&ok)
	if ok == 0 {
		redirect(w, r, "/admin/leads?err=lead_not_approved")
		return
	}
	if _, err := a.db.Exec(`UPDATE batches SET lead_id = ? WHERE id = ? AND status = 'live' AND start_date >= date('now')`,
		leadID, formInt(r, "batch_id")); err != nil {
		a.serverError(w, err)
		return
	}
	redirect(w, r, "/admin/leads?msg=lead_assigned#lead-%d", leadID)
}
