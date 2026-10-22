package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strings"
)

// Enquiries: "Enquire now" on a trek page, for people who want to know more
// before booking. They leave a name, phone (email optional) and a question. Each
// enquiry is saved for the admin console's Enquiries page and, when SMTP is set
// up, emailed to the team.
//
// Email settings (environment variables, public-site process):
//
//	SMTP_USER      Gmail address that sends the alerts
//	SMTP_PASS      a Gmail app password for it (not the normal password)
//	SMTP_HOST      default smtp.gmail.com
//	SMTP_PORT      default 587
//	ENQUIRY_EMAIL  where alerts go (default: the site's email)

type Enquiry struct {
	ID                          int64
	TrekName, TrekSlug          string
	Name, Phone, Email, Message string
	CreatedAt, RepliedAt        string
}

func migrateEnquiries(db *sql.DB) {
	db.Exec(`CREATE TABLE IF NOT EXISTS enquiries (
		id INTEGER PRIMARY KEY,
		trek_id INTEGER REFERENCES treks(id),
		name TEXT NOT NULL,
		phone TEXT NOT NULL,
		email TEXT NOT NULL DEFAULT '',
		message TEXT NOT NULL DEFAULT '',
		created_at TEXT NOT NULL DEFAULT (datetime('now')),
		replied_at TEXT NOT NULL DEFAULT ''
	)`)
}

// tenDigits returns the last 10 digits of an Indian phone number, or "" if it
// doesn't have 10.
func tenDigits(phone string) string {
	d := strings.Map(func(c rune) rune {
		if c >= '0' && c <= '9' {
			return c
		}
		return -1
	}, phone)
	if len(d) > 10 {
		d = d[len(d)-10:]
	}
	if len(d) != 10 {
		return ""
	}
	return d
}

func (a *App) enquire(w http.ResponseWriter, r *http.Request) {
	t, err := a.trekWhere(`t.slug = ? AND t.active = 1`, r.PathValue("slug"))
	if errors.Is(err, sql.ErrNoRows) {
		a.notFound(w, r)
		return
	} else if err != nil {
		a.serverError(w, err)
		return
	}
	a.saveEnquiry(w, r, t.ID, t.Name, "/treks/"+t.Slug)
}

// contactEnquire is the general question form on the Contact page (no trek).
func (a *App) contactEnquire(w http.ResponseWriter, r *http.Request) {
	a.saveEnquiry(w, r, nil, "", "/contact")
}

// saveEnquiry stores an enquiry (trekID may be nil) and sends the email alert.
func (a *App) saveEnquiry(w http.ResponseWriter, r *http.Request, trekID any, trekName, back string) {
	// "website" is a hidden field people never see; bots fill it in.
	if r.FormValue("website") != "" {
		redirect(w, r, "%s?msg=enquired#enquire", back)
		return
	}
	e := Enquiry{
		TrekName: or(trekName, "General question"),
		Name:     trimTo(r.FormValue("name"), 100),
		Phone:    tenDigits(r.FormValue("phone")),
		Email:    trimTo(r.FormValue("email"), 200),
		Message:  trimTo(r.FormValue("message"), 2000),
	}
	if e.Name == "" || e.Phone == "" {
		redirect(w, r, "%s?err=bad_enquiry#enquire", back)
		return
	}
	// The same person pressing send twice only counts once.
	var dup int
	a.db.QueryRow(`SELECT COUNT(*) FROM enquiries WHERE phone = ? AND trek_id IS ? AND message = ? AND created_at >= datetime('now', '-10 minutes')`,
		e.Phone, trekID, e.Message).Scan(&dup)
	if dup == 0 {
		if _, err := a.db.Exec(`INSERT INTO enquiries (trek_id, name, phone, email, message) VALUES (?, ?, ?, ?, ?)`,
			trekID, e.Name, e.Phone, e.Email, e.Message); err != nil {
			a.serverError(w, err)
			return
		}
		go emailEnquiry(e)
	}
	redirect(w, r, "%s?msg=enquired#enquire", back)
}

// emailEnquiry sends the team an alert. Without SMTP settings it only logs.
func emailEnquiry(e Enquiry) {
	user, pass := os.Getenv("SMTP_USER"), os.Getenv("SMTP_PASS")
	if user == "" || pass == "" {
		log.Printf("new enquiry from %s about %s (email alerts are off: set SMTP_USER and SMTP_PASS)", e.Name, e.TrekName)
		return
	}
	host := envOr("SMTP_HOST", "smtp.gmail.com")
	to := envOr("ENQUIRY_EMAIL", site.Email)
	clean := func(s string) string { return strings.NewReplacer("\r", " ", "\n", " ").Replace(s) }
	var b strings.Builder
	fmt.Fprintf(&b, "From: %s <%s>\r\n", site.Name, user)
	fmt.Fprintf(&b, "To: %s\r\n", to)
	if e.Email != "" {
		fmt.Fprintf(&b, "Reply-To: %s\r\n", clean(e.Email))
	}
	fmt.Fprintf(&b, "Subject: New enquiry: %s (%s)\r\n", clean(e.TrekName), clean(e.Name))
	b.WriteString("Content-Type: text/plain; charset=utf-8\r\n\r\n")
	fmt.Fprintf(&b, "Trek:    %s\nName:    %s\nPhone:   +91 %s\nWhatsApp: https://wa.me/91%s\n", e.TrekName, e.Name, e.Phone, e.Phone)
	if e.Email != "" {
		fmt.Fprintf(&b, "Email:   %s\n", e.Email)
	}
	fmt.Fprintf(&b, "\n%s\n\nSee all enquiries in the admin console (Enquiries).\n", or(e.Message, "(no message)"))
	err := smtp.SendMail(host+":"+envOr("SMTP_PORT", "587"), smtp.PlainAuth("", user, pass, host), user, []string{to}, []byte(b.String()))
	if err != nil {
		log.Printf("enquiry email failed: %v", err)
	}
}

func or(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

// consoleEnquiries lists enquiries, open ones first.
func (a *App) consoleEnquiries(w http.ResponseWriter, r *http.Request) {
	rows, err := a.db.Query(`SELECT e.id, COALESCE(t.name, ''), COALESCE(t.slug, ''), e.name, e.phone, e.email, e.message, e.created_at, e.replied_at
FROM enquiries e LEFT JOIN treks t ON t.id = e.trek_id
ORDER BY e.replied_at != '', e.id DESC LIMIT 500`)
	if err != nil {
		a.serverError(w, err)
		return
	}
	defer rows.Close()
	var list []Enquiry
	open := 0
	for rows.Next() {
		var e Enquiry
		if err := rows.Scan(&e.ID, &e.TrekName, &e.TrekSlug, &e.Name, &e.Phone, &e.Email, &e.Message, &e.CreatedAt, &e.RepliedAt); err != nil {
			a.serverError(w, err)
			return
		}
		if e.RepliedAt == "" {
			open++
		}
		list = append(list, e)
	}
	if err := rows.Err(); err != nil {
		a.serverError(w, err)
		return
	}
	a.render(w, r, "console_enquiries.html", map[string]any{"Enquiries": list, "Open": open})
}

// consoleEnquiryReplied marks an enquiry as answered (or reopens it).
func (a *App) consoleEnquiryReplied(w http.ResponseWriter, r *http.Request) {
	q := `UPDATE enquiries SET replied_at = datetime('now') WHERE id = ?`
	if r.FormValue("undo") != "" {
		q = `UPDATE enquiries SET replied_at = '' WHERE id = ?`
	}
	if _, err := a.db.Exec(q, pathID(r)); err != nil {
		a.serverError(w, err)
		return
	}
	redirect(w, r, "/admin/enquiries?msg=saved")
}
