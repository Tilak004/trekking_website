package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const schema = `
CREATE TABLE IF NOT EXISTS users (
  id            INTEGER PRIMARY KEY,
  name          TEXT NOT NULL,
  email         TEXT NOT NULL UNIQUE,
  phone         TEXT NOT NULL DEFAULT '',
  password_hash TEXT NOT NULL,
  is_admin      INTEGER NOT NULL DEFAULT 0,
  created_at    TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE TABLE IF NOT EXISTS sessions (
  token      TEXT PRIMARY KEY,
  user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  expires_at INTEGER NOT NULL
);
CREATE TABLE IF NOT EXISTS leads (
  id           INTEGER PRIMARY KEY,
  slug         TEXT NOT NULL UNIQUE,
  name         TEXT NOT NULL,
  headline     TEXT NOT NULL DEFAULT '',
  bio          TEXT NOT NULL DEFAULT '',
  achievements TEXT NOT NULL DEFAULT '',
  phone        TEXT NOT NULL DEFAULT '',
  active       INTEGER NOT NULL DEFAULT 1
);
CREATE TABLE IF NOT EXISTS treks (
  id            INTEGER PRIMARY KEY,
  slug          TEXT NOT NULL UNIQUE,
  name          TEXT NOT NULL,
  region        TEXT NOT NULL DEFAULT '',
  duration      TEXT NOT NULL DEFAULT '',
  distance_km   INTEGER NOT NULL DEFAULT 0,
  max_altitude  INTEGER NOT NULL DEFAULT 0,
  difficulty    TEXT NOT NULL DEFAULT 'Moderate',
  price         INTEGER NOT NULL DEFAULT 0,
  summary       TEXT NOT NULL DEFAULT '',
  description   TEXT NOT NULL DEFAULT '',
  pickup_points TEXT NOT NULL DEFAULT '',
  photo         TEXT NOT NULL DEFAULT '',
  active        INTEGER NOT NULL DEFAULT 1
);
CREATE TABLE IF NOT EXISTS batches (
  id         INTEGER PRIMARY KEY,
  trek_id    INTEGER NOT NULL REFERENCES treks(id),
  lead_id    INTEGER REFERENCES leads(id) ON DELETE SET NULL,
  start_date TEXT NOT NULL,
  seats      INTEGER NOT NULL
);
CREATE TABLE IF NOT EXISTS bookings (
  id              INTEGER PRIMARY KEY,
  user_id         INTEGER NOT NULL REFERENCES users(id),
  batch_id        INTEGER NOT NULL REFERENCES batches(id),
  status          TEXT NOT NULL DEFAULT 'requested',
  people          INTEGER NOT NULL DEFAULT 1,
  pickup          TEXT NOT NULL DEFAULT '',
  emergency_name  TEXT NOT NULL DEFAULT '',
  emergency_phone TEXT NOT NULL DEFAULT '',
  age             INTEGER NOT NULL DEFAULT 0,
  medical         TEXT NOT NULL DEFAULT '',
  note            TEXT NOT NULL DEFAULT '',
  created_at      TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE TABLE IF NOT EXISTS reviews (
  id          INTEGER PRIMARY KEY,
  booking_id  INTEGER NOT NULL UNIQUE REFERENCES bookings(id),
  trek_rating INTEGER NOT NULL,
  lead_rating INTEGER NOT NULL DEFAULT 0, -- 0 when the batch had no lead
  comment     TEXT NOT NULL DEFAULT '',
  created_at  TEXT NOT NULL DEFAULT (datetime('now'))
);
`

// Booking statuses. Seats are only counted as taken once a booking is approved.
const (
	StatusRequested = "requested"
	StatusApproved  = "approved"
	StatusRejected  = "rejected"
	StatusCompleted = "completed"
	StatusCancelled = "cancelled"
)

var bookingStatuses = []string{StatusRequested, StatusApproved, StatusCompleted, StatusRejected, StatusCancelled}

// nextActions lists the statuses an admin can move a booking to from its current one.
func nextActions(status string) []string {
	switch status {
	case StatusRequested:
		return []string{StatusApproved, StatusRejected}
	case StatusApproved:
		return []string{StatusCompleted, StatusCancelled}
	case StatusRejected, StatusCancelled:
		return []string{StatusRequested}
	}
	return nil
}

var actionLabels = map[string]string{
	StatusApproved:  "Approve",
	StatusRejected:  "Reject",
	StatusCompleted: "Mark completed",
	StatusCancelled: "Cancel",
	StatusRequested: "Reopen",
}

type User struct {
	ID        int64
	Name      string
	Email     string
	Phone     string
	IsAdmin   bool
	CreatedAt string
	Bookings  int
}

type Lead struct {
	ID             int64
	Slug           string
	Name           string
	Headline       string
	Bio            string
	Achievements   string
	Phone          string
	Photo          string
	UserID         int64 // the account that applied, 0 for leads added by an admin
	Status         string
	Certifications string
	City           string
	Verified       bool
	AppliedAt      string
	Active         bool
	TreksLed       int
	TrekkersGuided int
	AvgRating      float64
	Ratings        int
}

type Trek struct {
	ID            int64
	Slug          string
	Name          string
	Region        string
	Duration      string
	DistanceKm    int
	MaxAltitude   int
	Difficulty    string
	Price         int
	Summary       string
	Description   string
	PickupPoints  string
	Photo         string
	BestSeason    string
	Itinerary     string
	Inclusions    string
	Exclusions    string
	ThingsToCarry string
	FAQs          string
	Lat           float64
	Lng           float64
	Tags          string
	Elevation     string
	Active        bool
	NextBatch     string
	AvgRating     float64
	Ratings       int
	Lead          *LeadMini // next batch's lead, filled in by handlers that need it
}

type Batch struct {
	ID        int64
	TrekID    int64
	StartDate string
	Seats     int
	Taken     int
	LeadID    int64
	LeadName  string
	LeadSlug  string
	Bookings  int
	Status    string
}

func (b Batch) Left() int { return b.Seats - b.Taken }

type Booking struct {
	ID             int64
	Status         string
	People         int
	Pickup         string
	EmergencyName  string
	EmergencyPhone string
	Age            int
	Medical        string
	Note           string
	CreatedAt      string
	BatchID        int64
	StartDate      string
	TrekName       string
	TrekSlug       string
	LeadName       string
	HasLead        bool
	UserName       string
	UserEmail      string
	UserPhone      string
	Reviewed       bool
	Price          int
	PayToken       string
	PaidAt         string
	PaidAmount     int
	RefundAmount   int
	RefundReason   string
	RefundedAt     string
}

type Review struct {
	TrekRating int
	LeadRating int
	Comment    string
	CreatedAt  string
	UserName   string
	TrekName   string
	Photo      string
}

type scanner interface{ Scan(dest ...any) error }

func openDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path+"?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, err
	}
	// One connection keeps SQLite writes simple and is plenty for this traffic.
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("create schema: %w", err)
	}
	// Columns added after the first release; ignore the error if they already exist.
	db.Exec(`ALTER TABLE treks ADD COLUMN photo TEXT NOT NULL DEFAULT ''`)
	db.Exec(`ALTER TABLE leads ADD COLUMN photo TEXT NOT NULL DEFAULT ''`)
	// Trek-lead applications (see leadapply.go).
	db.Exec(`ALTER TABLE leads ADD COLUMN user_id INTEGER REFERENCES users(id)`)
	db.Exec(`ALTER TABLE leads ADD COLUMN status TEXT NOT NULL DEFAULT 'approved'`)
	db.Exec(`ALTER TABLE leads ADD COLUMN certifications TEXT NOT NULL DEFAULT ''`)
	db.Exec(`ALTER TABLE leads ADD COLUMN city TEXT NOT NULL DEFAULT ''`)
	db.Exec(`ALTER TABLE leads ADD COLUMN verified INTEGER NOT NULL DEFAULT 0`)
	db.Exec(`ALTER TABLE leads ADD COLUMN applied_at TEXT NOT NULL DEFAULT ''`)
	if _, err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS leads_user_id ON leads(user_id) WHERE user_id IS NOT NULL`); err != nil {
		return nil, err
	}
	if _, err := db.Exec(`DELETE FROM sessions WHERE expires_at < ?`, time.Now().Unix()); err != nil {
		return nil, err
	}
	if err := seed(db); err != nil {
		return nil, fmt.Errorf("seed: %w", err)
	}
	if err := migrateTrekDetails(db); err != nil {
		return nil, fmt.Errorf("migrate trek details: %w", err)
	}
	migrateLeadDates(db)
	migratePayments(db)
	migrateRefunds(db)
	migrateEnquiries(db)
	if err := migrateInteractive(db); err != nil {
		return nil, fmt.Errorf("migrate interactive: %w", err)
	}
	return db, nil
}

// ---- treks ----

const trekSelect = `
SELECT t.id, t.slug, t.name, t.region, t.duration, t.distance_km, t.max_altitude, t.difficulty,
       t.price, t.summary, t.description, t.pickup_points, t.photo, t.active,
       t.best_season, t.itinerary, t.inclusions, t.exclusions, t.things_to_carry, t.faqs,
       t.lat, t.lng, t.tags, t.elevation,
       COALESCE((SELECT MIN(b.start_date) FROM batches b WHERE b.trek_id = t.id AND b.status = 'live' AND b.start_date >= date('now')), ''),
       (SELECT COALESCE(AVG(r.trek_rating), 0.0) FROM reviews r JOIN bookings k ON k.id = r.booking_id
          JOIN batches b ON b.id = k.batch_id WHERE b.trek_id = t.id),
       (SELECT COUNT(*) FROM reviews r JOIN bookings k ON k.id = r.booking_id
          JOIN batches b ON b.id = k.batch_id WHERE b.trek_id = t.id)
FROM treks t`

func scanTrek(s scanner) (Trek, error) {
	var t Trek
	err := s.Scan(&t.ID, &t.Slug, &t.Name, &t.Region, &t.Duration, &t.DistanceKm, &t.MaxAltitude, &t.Difficulty,
		&t.Price, &t.Summary, &t.Description, &t.PickupPoints, &t.Photo, &t.Active,
		&t.BestSeason, &t.Itinerary, &t.Inclusions, &t.Exclusions, &t.ThingsToCarry, &t.FAQs,
		&t.Lat, &t.Lng, &t.Tags, &t.Elevation, &t.NextBatch, &t.AvgRating, &t.Ratings)
	return t, err
}

func (a *App) listTreks(activeOnly bool) ([]Trek, error) {
	q := trekSelect
	if activeOnly {
		q += ` WHERE t.active = 1`
	}
	rows, err := a.db.Query(q + ` ORDER BY t.active DESC, t.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Trek
	for rows.Next() {
		t, err := scanTrek(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (a *App) trekWhere(where string, arg any) (Trek, error) {
	return scanTrek(a.db.QueryRow(trekSelect+` WHERE `+where, arg))
}

// ---- batches ----

const batchSelect = `
SELECT b.id, b.trek_id, b.start_date, b.seats,
       (SELECT COALESCE(SUM(k.people), 0) FROM bookings k WHERE k.batch_id = b.id AND k.status IN ('approved', 'completed')),
       COALESCE(b.lead_id, 0), COALESCE(l.name, ''), COALESCE(l.slug, ''),
       (SELECT COUNT(*) FROM bookings k WHERE k.batch_id = b.id),
       b.status
FROM batches b LEFT JOIN leads l ON l.id = b.lead_id`

func scanBatch(s scanner) (Batch, error) {
	var b Batch
	err := s.Scan(&b.ID, &b.TrekID, &b.StartDate, &b.Seats, &b.Taken, &b.LeadID, &b.LeadName, &b.LeadSlug, &b.Bookings, &b.Status)
	return b, err
}

func (a *App) batchesForTrek(trekID int64, upcomingOnly bool) ([]Batch, error) {
	q := batchSelect + ` WHERE b.trek_id = ?`
	if upcomingOnly {
		// The public view: only approved (live) dates that haven't happened yet.
		q += ` AND b.start_date >= date('now') AND b.status = 'live'`
	}
	return a.queryBatches(q+` ORDER BY b.start_date`, trekID)
}

func (a *App) queryBatches(q string, args ...any) ([]Batch, error) {
	rows, err := a.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Batch
	for rows.Next() {
		b, err := scanBatch(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (a *App) batchByID(id int64) (Batch, error) {
	return scanBatch(a.db.QueryRow(batchSelect+` WHERE b.id = ?`, id))
}

// ---- leads ----

const leadSelect = `
SELECT l.id, l.slug, l.name, l.headline, l.bio, l.achievements, l.phone, l.photo, l.active,
       COALESCE(l.user_id, 0), l.status, l.certifications, l.city, l.verified, l.applied_at,
       (SELECT COUNT(DISTINCT b.id) FROM batches b JOIN bookings k ON k.batch_id = b.id
          WHERE b.lead_id = l.id AND k.status = 'completed'),
       (SELECT COALESCE(SUM(k.people), 0) FROM batches b JOIN bookings k ON k.batch_id = b.id
          WHERE b.lead_id = l.id AND k.status = 'completed'),
       (SELECT COALESCE(AVG(r.lead_rating), 0.0) FROM reviews r JOIN bookings k ON k.id = r.booking_id
          JOIN batches b ON b.id = k.batch_id WHERE b.lead_id = l.id AND r.lead_rating > 0),
       (SELECT COUNT(*) FROM reviews r JOIN bookings k ON k.id = r.booking_id
          JOIN batches b ON b.id = k.batch_id WHERE b.lead_id = l.id AND r.lead_rating > 0)
FROM leads l`

func scanLead(s scanner) (Lead, error) {
	var l Lead
	err := s.Scan(&l.ID, &l.Slug, &l.Name, &l.Headline, &l.Bio, &l.Achievements, &l.Phone, &l.Photo, &l.Active,
		&l.UserID, &l.Status, &l.Certifications, &l.City, &l.Verified, &l.AppliedAt,
		&l.TreksLed, &l.TrekkersGuided, &l.AvgRating, &l.Ratings)
	return l, err
}

func (a *App) listLeads(activeOnly bool) ([]Lead, error) {
	q := leadSelect
	if activeOnly {
		q += ` WHERE l.active = 1 AND l.status = 'approved'`
	}
	// Pending applications first, so admins see them straight away.
	rows, err := a.db.Query(q + ` ORDER BY CASE l.status WHEN 'pending' THEN 0 ELSE 1 END, l.active DESC, l.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Lead
	for rows.Next() {
		l, err := scanLead(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

func (a *App) leadWhere(where string, arg any) (Lead, error) {
	return scanLead(a.db.QueryRow(leadSelect+` WHERE `+where, arg))
}

// ---- bookings ----

const bookingSelect = `
SELECT k.id, k.status, k.people, k.pickup, k.emergency_name, k.emergency_phone, k.age, k.medical, k.note, k.created_at,
       b.id, b.start_date, t.name, t.slug, COALESCE(l.name, ''), b.lead_id IS NOT NULL,
       u.name, u.email, u.phone,
       EXISTS (SELECT 1 FROM reviews r WHERE r.booking_id = k.id),
       t.price, k.pay_token, k.paid_at, k.paid_amount, k.refund_amount, k.refund_reason, k.refunded_at
FROM bookings k
JOIN batches b ON b.id = k.batch_id
JOIN treks t ON t.id = b.trek_id
JOIN users u ON u.id = k.user_id
LEFT JOIN leads l ON l.id = b.lead_id`

func scanBooking(s scanner) (Booking, error) {
	var k Booking
	err := s.Scan(&k.ID, &k.Status, &k.People, &k.Pickup, &k.EmergencyName, &k.EmergencyPhone, &k.Age, &k.Medical,
		&k.Note, &k.CreatedAt, &k.BatchID, &k.StartDate, &k.TrekName, &k.TrekSlug, &k.LeadName, &k.HasLead,
		&k.UserName, &k.UserEmail, &k.UserPhone, &k.Reviewed, &k.Price, &k.PayToken, &k.PaidAt, &k.PaidAmount, &k.RefundAmount, &k.RefundReason, &k.RefundedAt)
	return k, err
}

func (a *App) queryBookings(where string, args ...any) ([]Booking, error) {
	rows, err := a.db.Query(bookingSelect+` `+where, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Booking
	for rows.Next() {
		k, err := scanBooking(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, k)
	}
	return out, rows.Err()
}

// ---- reviews ----

func (a *App) queryReviews(where string, arg any) ([]Review, error) {
	rows, err := a.db.Query(`
SELECT r.trek_rating, r.lead_rating, r.comment, r.created_at, u.name, t.name, r.photo
FROM reviews r
JOIN bookings k ON k.id = r.booking_id
JOIN users u ON u.id = k.user_id
JOIN batches b ON b.id = k.batch_id
JOIN treks t ON t.id = b.trek_id
WHERE `+where+` ORDER BY r.id DESC LIMIT 50`, args(arg)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Review
	for rows.Next() {
		var r Review
		if err := rows.Scan(&r.TrekRating, &r.LeadRating, &r.Comment, &r.CreatedAt, &r.UserName, &r.TrekName, &r.Photo); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// ---- seed data ----

func seed(db *sql.DB) error {
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM users WHERE is_admin = 1`).Scan(&n); err != nil {
		return err
	}
	if n == 0 {
		email := envOr("ADMIN_EMAIL", "admin@beyondexpeditions.local")
		pw := os.Getenv("ADMIN_PASSWORD")
		if pw == "" {
			pw = randomToken(6)
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		if _, err := db.Exec(`INSERT INTO users (name, email, password_hash, is_admin) VALUES ('Admin', ?, ?, 1)`,
			email, string(hash)); err != nil {
			return fmt.Errorf("create admin: %w", err)
		}
		log.Printf("created admin account  email: %s  password: %s  (shown once, note it down)", email, pw)
	}

	if err := db.QueryRow(`SELECT COUNT(*) FROM treks`).Scan(&n); err != nil || n > 0 {
		return err
	}
	log.Printf("adding sample treks and leads (edit or deactivate them in the admin panel)")

	var leadIDs []int64
	for _, l := range sampleLeads {
		res, err := db.Exec(`INSERT INTO leads (slug, name, headline, bio, achievements) VALUES (?, ?, ?, ?, ?)`,
			slugify(l.Name), l.Name, l.Headline, l.Bio, l.Achievements)
		if err != nil {
			return err
		}
		id, _ := res.LastInsertId()
		leadIDs = append(leadIDs, id)
	}

	pickups := "Bangalore - Majestic\nBangalore - Silk Board\nReach base village directly"
	for i, t := range sampleTreks {
		res, err := db.Exec(`INSERT INTO treks (slug, name, region, duration, distance_km, max_altitude, difficulty, price, summary, description, pickup_points)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, slugify(t.Name), t.Name, t.Region, t.Duration, t.DistanceKm, t.MaxAltitude,
			t.Difficulty, t.Price, t.Summary, t.Description, pickups)
		if err != nil {
			return err
		}
		trekID, _ := res.LastInsertId()
		// Two upcoming weekend batches per trek, rotating through the leads.
		for j, weeks := range []int{1, 3} {
			start := nextSaturday().AddDate(0, 0, 7*(weeks+i%2))
			lead := leadIDs[(i+j)%len(leadIDs)]
			if _, err := db.Exec(`INSERT INTO batches (trek_id, lead_id, start_date, seats) VALUES (?, ?, ?, 20)`,
				trekID, lead, start.Format("2006-01-02")); err != nil {
				return err
			}
		}
	}
	return nil
}

func nextSaturday() time.Time {
	d := time.Now()
	for d.Weekday() != time.Saturday {
		d = d.AddDate(0, 0, 1)
	}
	return d
}

// Sample content only: verify distances, altitudes, prices and permits before going live.
var sampleLeads = []Lead{
	{Name: "Founder One", Headline: "Ironman finisher · Endurance coach",
		Bio:          "Completed a full Ironman and brings that endurance mindset to every group: steady pacing, nobody left behind.",
		Achievements: "Full Ironman finisher\nMarathon runner\nFirst-aid certified (update as applicable)"},
	{Name: "Founder Two", Headline: "Trail runner · 24-hour adventure racer",
		Bio:          "Races 24-hour adventure races and runs the Western Ghats trails most weekends. Knows these routes in every season.",
		Achievements: "24-hour adventure race finisher\nUltra trail runner"},
	{Name: "Founder Three", Headline: "Himalayan & Western Ghats trekker",
		Bio:          "Has trekked most of Karnataka's classic trails and several Himalayan routes. Looks after planning and logistics.",
		Achievements: "Several Himalayan treks completed\nMost major Karnataka treks completed"},
}

var sampleTreks = []Trek{
	{Name: "Kumara Parvatha", Region: "Kukke Subramanya, Dakshina Kannada", Duration: "2 days / 1 night",
		DistanceKm: 22, MaxAltitude: 1712, Difficulty: "Difficult", Price: 3499,
		Summary:     "One of Karnataka's toughest treks: dense forest, steep grassland and a summit above the clouds.",
		Description: "Start from Kukke Subramanya and climb through thick shola forest to Bhattara Mane, then on to Shesha Parvatha and the Kumara Parvatha summit. Camp overnight and descend the next day.\n\nExpect long, steep sections. Good fitness is a must."},
	{Name: "Kudremukh", Region: "Chikkamagaluru", Duration: "1 day (overnight travel)",
		DistanceKm: 20, MaxAltitude: 1894, Difficulty: "Moderate", Price: 2799,
		Summary:     "Rolling green hills, streams and the famous horse-face peak inside Kudremukh National Park.",
		Description: "A long but gradual walk through grasslands and shola patches with several stream crossings. Forest department permits are included.\n\nBest from September to February."},
	{Name: "Netravati Peak", Region: "Chikkamagaluru", Duration: "1 day (overnight travel)",
		DistanceKm: 10, MaxAltitude: 1520, Difficulty: "Easy–Moderate", Price: 2299,
		Summary:     "A short ridge walk with sweeping views of the Western Ghats, perfect for first-timers.",
		Description: "A gentle climb through grassland onto a ridge with views across the Kudremukh range. Great for beginners and photographers."},
	{Name: "Tadiandamol", Region: "Kodagu (Coorg)", Duration: "1 day (overnight travel)",
		DistanceKm: 16, MaxAltitude: 1748, Difficulty: "Moderate", Price: 2499,
		Summary:     "Coorg's highest peak: coffee estates, shola forest and open summit meadows.",
		Description: "Starts near the Nalknad Palace and climbs steadily through forest to the open summit. A classic monsoon and winter trek."},
	{Name: "Skandagiri Night Trek", Region: "Chikkaballapur", Duration: "Overnight (sunrise trek)",
		DistanceKm: 8, MaxAltitude: 1450, Difficulty: "Easy", Price: 1299,
		Summary:     "Climb by headlamp and watch the sunrise over a sea of clouds, an hour from Bangalore.",
		Description: "Leave Bangalore around midnight, trek up in the dark and reach the top for sunrise. Back in the city by late morning."},
}

// args turns an optional single query argument into a parameter list.
func args(a any) []any {
	if a == nil {
		return nil
	}
	return []any{a}
}
