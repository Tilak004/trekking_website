package main

import (
	"net/http"
	"slices"
)

// Analytics for the admin console. Rupee figures are estimates (price x people
// on approved/completed bookings); see the note at the top of console.go.

type Bar struct {
	Label string
	Sub   string
	Value int
	Pct   int // bar length, 0-100, relative to the largest bar in its chart
}

func bars(items []Bar) []Bar {
	most := 0
	for _, b := range items {
		most = max(most, b.Value)
	}
	for i := range items {
		items[i].Pct = pct(items[i].Value, most)
	}
	return items
}

type StatusRow struct {
	Status         string
	Bookings       int
	People         int
	Value          int
	Pct            int
}

type TrekRow struct {
	Name           string
	Events         int
	People, Value  int
	Fill           int
	Rating         float64
	Ratings        int
}

type LeadRow struct {
	Name           string
	Events, People int
	Value          int
}

func (a *App) consoleAnalytics(w http.ResponseWriter, r *http.Request) {
	fail := func(err error) { a.serverError(w, err) }

	// Money and bookings by status.
	rows, err := a.db.Query(`SELECT k.status, COUNT(*), COALESCE(SUM(k.people), 0), COALESCE(SUM(k.people * t.price), 0)
FROM bookings k JOIN batches b ON b.id = k.batch_id JOIN treks t ON t.id = b.trek_id GROUP BY k.status`)
	if err != nil {
		fail(err)
		return
	}
	byStatus := map[string]StatusRow{}
	for rows.Next() {
		var s StatusRow
		if err := rows.Scan(&s.Status, &s.Bookings, &s.People, &s.Value); err != nil {
			rows.Close()
			fail(err)
			return
		}
		byStatus[s.Status] = s
	}
	rows.Close()
	var funnel []StatusRow
	total := 0
	for _, s := range bookingStatuses {
		total += byStatus[s].Bookings
	}
	for _, s := range bookingStatuses {
		row := byStatus[s]
		row.Status, row.Pct = s, pct(row.Bookings, total)
		funnel = append(funnel, row)
	}
	confirmed, completed := byStatus[StatusApproved], byStatus[StatusCompleted]
	confirmedValue := confirmed.Value + completed.Value
	confirmedPeople := confirmed.People + completed.People
	confirmedBookings := confirmed.Bookings + completed.Bookings
	lost := byStatus[StatusCancelled].Value + byStatus[StatusRejected].Value

	var unique, repeat int
	a.db.QueryRow(`SELECT COUNT(DISTINCT user_id) FROM bookings WHERE status IN (` + confirmedStatuses + `)`).Scan(&unique)
	a.db.QueryRow(`SELECT COUNT(*) FROM (SELECT user_id FROM bookings WHERE status IN (` + confirmedStatuses + `) GROUP BY user_id HAVING COUNT(*) >= 2)`).Scan(&repeat)
	var seats, taken int
	a.db.QueryRow(`SELECT COALESCE(SUM(seats), 0) FROM batches WHERE status = 'live'`).Scan(&seats)
	a.db.QueryRow(`SELECT COALESCE(SUM(k.people), 0) FROM bookings k JOIN batches b ON b.id = k.batch_id WHERE b.status = 'live' AND k.status IN (` + confirmedStatuses + `)`).Scan(&taken)

	avgGroup, avgValue := 0.0, 0
	if confirmedBookings > 0 {
		avgGroup = float64(confirmedPeople) / float64(confirmedBookings)
		avgValue = confirmedValue / confirmedBookings
	}

	// Money by month of the trek date (last 12 months with bookings).
	monthly, err := a.barQuery(`SELECT substr(b.start_date, 1, 7), COALESCE(SUM(k.people * t.price), 0), COALESCE(SUM(k.people), 0)
FROM bookings k JOIN batches b ON b.id = k.batch_id JOIN treks t ON t.id = b.trek_id
WHERE k.status IN (`+confirmedStatuses+`) GROUP BY 1 ORDER BY 1 DESC LIMIT 12`, func(label string, v, p int) Bar {
		return Bar{Label: monthLabel(label), Value: v, Sub: plural(p, "person", "people")}
	})
	if err != nil {
		fail(err)
		return
	}
	slices.Reverse(monthly)

	// Who the trekkers are.
	ages, err := a.barQuery(`SELECT CASE WHEN k.age <= 0 THEN 'Not given' WHEN k.age < 18 THEN 'Under 18'
WHEN k.age < 25 THEN '18-24' WHEN k.age < 35 THEN '25-34' WHEN k.age < 45 THEN '35-44' ELSE '45+' END AS grp,
COUNT(*), 0 FROM bookings k WHERE k.status IN (`+confirmedStatuses+`) GROUP BY grp`, func(label string, v, _ int) Bar {
		return Bar{Label: label, Value: v}
	})
	if err != nil {
		fail(err)
		return
	}
	order := []string{"Under 18", "18-24", "25-34", "35-44", "45+", "Not given"}
	slices.SortFunc(ages, func(x, y Bar) int { return slices.Index(order, x.Label) - slices.Index(order, y.Label) })

	pickups, err := a.barQuery(`SELECT CASE WHEN pickup = '' THEN 'Not given' ELSE pickup END, COALESCE(SUM(people), 0), 0
FROM bookings WHERE status IN (`+confirmedStatuses+`) GROUP BY 1 ORDER BY 2 DESC LIMIT 8`, func(label string, v, _ int) Bar {
		return Bar{Label: label, Value: v}
	})
	if err != nil {
		fail(err)
		return
	}
	signups, err := a.barQuery(`SELECT substr(created_at, 1, 7), COUNT(*), SUM(email LIKE 'guest-%') FROM users WHERE is_admin = 0
GROUP BY 1 ORDER BY 1 DESC LIMIT 6`, func(label string, v, guests int) Bar {
		return Bar{Label: monthLabel(label), Value: v, Sub: plural(guests, "guest", "guests")}
	})
	if err != nil {
		fail(err)
		return
	}
	slices.Reverse(signups)

	// Per trek and per lead.
	var treks []TrekRow
	rows, err = a.db.Query(`SELECT t.name,
  (SELECT COUNT(*) FROM batches b WHERE b.trek_id = t.id AND b.status = 'live'),
  COALESCE((SELECT SUM(k.people) FROM bookings k JOIN batches b ON b.id = k.batch_id WHERE b.trek_id = t.id AND k.status IN (`+confirmedStatuses+`)), 0),
  COALESCE((SELECT SUM(b.seats) FROM batches b WHERE b.trek_id = t.id AND b.status = 'live'), 0),
  t.price,
  COALESCE((SELECT AVG(r.trek_rating) FROM reviews r JOIN bookings k ON k.id = r.booking_id JOIN batches b ON b.id = k.batch_id WHERE b.trek_id = t.id), 0),
  (SELECT COUNT(*) FROM reviews r JOIN bookings k ON k.id = r.booking_id JOIN batches b ON b.id = k.batch_id WHERE b.trek_id = t.id)
FROM treks t`)
	if err != nil {
		fail(err)
		return
	}
	for rows.Next() {
		var t TrekRow
		var seats, price int
		if err := rows.Scan(&t.Name, &t.Events, &t.People, &seats, &price, &t.Rating, &t.Ratings); err != nil {
			rows.Close()
			fail(err)
			return
		}
		t.Value, t.Fill = t.People*price, pct(t.People, seats)
		treks = append(treks, t)
	}
	rows.Close()
	slices.SortFunc(treks, func(x, y TrekRow) int { return y.Value - x.Value })

	var leads []LeadRow
	rows, err = a.db.Query(`SELECT l.name, COUNT(DISTINCT b.id), COALESCE(SUM(k.people), 0), COALESCE(SUM(k.people * t.price), 0)
FROM leads l JOIN batches b ON b.lead_id = l.id AND b.status = 'live'
JOIN treks t ON t.id = b.trek_id LEFT JOIN bookings k ON k.batch_id = b.id AND k.status IN (`+confirmedStatuses+`)
GROUP BY l.id ORDER BY 4 DESC LIMIT 15`)
	if err != nil {
		fail(err)
		return
	}
	for rows.Next() {
		var l LeadRow
		if err := rows.Scan(&l.Name, &l.Events, &l.People, &l.Value); err != nil {
			rows.Close()
			fail(err)
			return
		}
		leads = append(leads, l)
	}
	rows.Close()

	a.render(w, r, "console_analytics.html", map[string]any{
		"ConfirmedValue": confirmedValue, "CompletedValue": completed.Value, "PipelineValue": byStatus[StatusRequested].Value,
		"LostValue": lost, "People": confirmedPeople, "Bookings": confirmedBookings, "AvgGroup": avgGroup, "AvgValue": avgValue,
		"Unique": unique, "Repeat": repeat, "RepeatPct": pct(repeat, unique), "FillPct": pct(taken, seats),
		"Funnel": funnel, "Monthly": bars(monthly), "Ages": bars(ages), "Pickups": bars(pickups), "Signups": bars(signups),
		"Treks": treks, "Leads": leads,
	})
}

// barQuery runs a query that returns (label, value, extra) rows.
func (a *App) barQuery(q string, mk func(label string, v, extra int) Bar) ([]Bar, error) {
	rows, err := a.db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Bar
	for rows.Next() {
		var label string
		var v, extra int
		if err := rows.Scan(&label, &v, &extra); err != nil {
			return nil, err
		}
		out = append(out, mk(label, v, extra))
	}
	return out, rows.Err()
}

func monthLabel(ym string) string {
	if len(ym) != 7 {
		return ym
	}
	months := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
	m := int(ym[5]-'0')*10 + int(ym[6]-'0')
	if m < 1 || m > 12 {
		return ym
	}
	return months[m-1] + " " + ym[2:4]
}

func plural(n int, one, many string) string {
	if n == 1 {
		return "1 " + one
	}
	return itoa(n) + " " + many
}

func itoa(n int) string { return formatINR(n) }
