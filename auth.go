package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// sessionCookie is named after the database (see main), so the real site and the
// demo site running on the same computer never share a login.
var sessionCookie = "session"

const (
	sessionTTL = 30 * 24 * time.Hour
)

type ctxKey struct{}

func randomToken(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

func currentUser(r *http.Request) *User {
	u, _ := r.Context().Value(ctxKey{}).(*User)
	return u
}

// withUser loads the logged-in user (if any) into the request context and
// rejects cross-site form posts.
func (a *App) withUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && !sameOrigin(r) {
			http.Error(w, "Cross-site request rejected.", http.StatusForbidden)
			return
		}
		if c, err := r.Cookie(sessionCookie); err == nil {
			var u User
			err := a.db.QueryRow(`
SELECT u.id, u.name, u.email, u.phone, u.is_admin, u.created_at
FROM sessions s JOIN users u ON u.id = s.user_id
WHERE s.token = ? AND s.expires_at > ?`, c.Value, time.Now().Unix()).
				Scan(&u.ID, &u.Name, &u.Email, &u.Phone, &u.IsAdmin, &u.CreatedAt)
			if err == nil {
				r = r.WithContext(context.WithValue(r.Context(), ctxKey{}, &u))
			}
		}
		next.ServeHTTP(w, r)
	})
}

func sameOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" || origin == "null" {
		return origin == ""
	}
	u, err := url.Parse(origin)
	return err == nil && u.Host == r.Host
}

func (a *App) requireUser(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if currentUser(r) == nil {
			next := ""
			if r.Method == http.MethodGet {
				next = "?next=" + url.QueryEscape(r.URL.RequestURI())
			}
			if isAPI(r) { // the React app sends people to the login page itself
				apiJSON(w, http.StatusUnauthorized, map[string]any{"error": "login", "redirect": "/login" + next})
				return
			}
			http.Redirect(w, r, "/login"+next, http.StatusSeeOther)
			return
		}
		h(w, r)
	}
}

func (a *App) requireAdmin(h http.HandlerFunc) http.HandlerFunc {
	return a.requireUser(func(w http.ResponseWriter, r *http.Request) {
		if !currentUser(r).IsAdmin {
			http.NotFound(w, r)
			return
		}
		h(w, r)
	})
}

func (a *App) startSession(w http.ResponseWriter, r *http.Request, userID int64) error {
	token := randomToken(32)
	expires := time.Now().Add(sessionTTL)
	if _, err := a.db.Exec(`INSERT INTO sessions (token, user_id, expires_at) VALUES (?, ?, ?)`,
		token, userID, expires.Unix()); err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name: sessionCookie, Value: token, Path: "/", Expires: expires,
		HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: isHTTPS(r),
	})
	return nil
}

// safeNext only allows redirects back to a path on this site.
func safeNext(next string) string {
	if strings.HasPrefix(next, "/") && !strings.HasPrefix(next, "//") && !strings.Contains(next, `\`) {
		return next
	}
	return "/"
}

func (a *App) signupForm(w http.ResponseWriter, r *http.Request) {
	a.render(w, r, "signup.html", map[string]any{"Form": map[string]string{"Name": "", "Email": "", "Phone": ""}})
}

func (a *App) signup(w http.ResponseWriter, r *http.Request) {
	form := map[string]string{
		"Name":  trimTo(r.FormValue("name"), 100),
		"Email": strings.ToLower(trimTo(r.FormValue("email"), 200)),
		"Phone": trimTo(r.FormValue("phone"), 30),
	}
	pw := r.FormValue("password")
	fail := func(msg string) { a.render(w, r, "signup.html", map[string]any{"Form": form, "Err": msg}) }

	switch {
	case form["Name"] == "" || form["Email"] == "" || form["Phone"] == "":
		fail("Please fill in your name, email and phone number.")
		return
	case !strings.Contains(form["Email"], "@"):
		fail("Please enter a valid email address.")
		return
	case len(pw) < 8 || len(pw) > 72:
		fail("Password must be 8 to 72 characters long.")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		a.serverError(w, err)
		return
	}
	res, err := a.db.Exec(`INSERT INTO users (name, email, phone, password_hash) VALUES (?, ?, ?, ?)`,
		form["Name"], form["Email"], form["Phone"], string(hash))
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			fail("An account with this email already exists. Try logging in.")
			return
		}
		a.serverError(w, err)
		return
	}
	id, _ := res.LastInsertId()
	if err := a.startSession(w, r, id); err != nil {
		a.serverError(w, err)
		return
	}
	redirect(w, r, "/?msg=welcome")
}

func (a *App) loginForm(w http.ResponseWriter, r *http.Request) {
	a.render(w, r, "login.html", map[string]any{"Email": "", "Next": r.URL.Query().Get("next")})
}

func (a *App) login(w http.ResponseWriter, r *http.Request) {
	email := strings.ToLower(strings.TrimSpace(r.FormValue("email")))
	next := r.FormValue("next")
	var id int64
	var hash string
	err := a.db.QueryRow(`SELECT id, password_hash FROM users WHERE email = ?`, email).Scan(&id, &hash)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(hash), []byte(r.FormValue("password"))) != nil {
		a.render(w, r, "login.html", map[string]any{"Email": email, "Next": next, "Err": "Wrong email or password."})
		return
	}
	if err := a.startSession(w, r, id); err != nil {
		a.serverError(w, err)
		return
	}
	redirectTo(w, r, safeNext(next))
}

func (a *App) logout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(sessionCookie); err == nil {
		a.db.Exec(`DELETE FROM sessions WHERE token = ?`, c.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: "", Path: "/", MaxAge: -1})
	redirect(w, r, "/")
}

// resetAdminPassword gives the first admin account a new password and prints the
// login. By default the password is random; set ADMIN_PASSWORD (and optionally
// ADMIN_EMAIL) to choose them. Run it with: northstar.exe -reset-admin-password
func resetAdminPassword(db *sql.DB) error {
	var id int64
	var email string
	if err := db.QueryRow(`SELECT id, email FROM users WHERE is_admin = 1 ORDER BY id LIMIT 1`).Scan(&id, &email); err != nil {
		return fmt.Errorf("no admin account found: %w", err)
	}
	if e := strings.ToLower(strings.TrimSpace(os.Getenv("ADMIN_EMAIL"))); e != "" && e != email {
		if !strings.Contains(e, "@") {
			return fmt.Errorf("ADMIN_EMAIL %q is not a valid email", e)
		}
		if _, err := db.Exec(`UPDATE users SET email = ? WHERE id = ?`, e, id); err != nil {
			return fmt.Errorf("set admin email: %w", err)
		}
		email = e
	}
	pw := os.Getenv("ADMIN_PASSWORD")
	if pw == "" {
		pw = randomToken(6)
	} else if len(pw) < 8 || len(pw) > 72 {
		return fmt.Errorf("ADMIN_PASSWORD must be 8 to 72 characters long")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if _, err := db.Exec(`UPDATE users SET password_hash = ? WHERE id = ?`, string(hash), id); err != nil {
		return err
	}
	db.Exec(`DELETE FROM sessions WHERE user_id = ?`, id) // log out old sessions
	fmt.Printf("Admin login updated\n  email: %s\n", email)
	if os.Getenv("ADMIN_PASSWORD") == "" {
		fmt.Printf("  password: %s\n", pw)
	}
	return nil
}

// guestUser finds or creates a guest account for someone booking without
// signing up, keyed by their phone number, and logs them in.
func (a *App) guestUser(w http.ResponseWriter, r *http.Request, name, phone string) (*User, error) {
	digits := strings.Map(func(c rune) rune {
		if c >= '0' && c <= '9' {
			return c
		}
		return -1
	}, phone)
	if len(digits) > 10 {
		digits = digits[len(digits)-10:] // drop +91 / 0 prefixes
	}
	if name == "" || len(digits) != 10 {
		return nil, fmt.Errorf("guest booking needs a name and a 10-digit phone")
	}
	email := "guest-" + digits + "@guest.beyondexpeditions.local"
	var id int64
	err := a.db.QueryRow(`SELECT id FROM users WHERE email = ?`, email).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		hash, herr := bcrypt.GenerateFromPassword([]byte(randomToken(16)), bcrypt.DefaultCost)
		if herr != nil {
			return nil, herr
		}
		res, ierr := a.db.Exec(`INSERT INTO users (name, email, phone, password_hash) VALUES (?, ?, ?, ?)`,
			name, email, "+91 "+digits[:5]+" "+digits[5:], string(hash))
		if ierr != nil {
			return nil, ierr
		}
		id, _ = res.LastInsertId()
	} else if err != nil {
		return nil, err
	} else {
		a.db.Exec(`UPDATE users SET name = ? WHERE id = ?`, name, id)
	}
	if err := a.startSession(w, r, id); err != nil {
		return nil, err
	}
	return &User{ID: id, Name: name, Phone: "+91 " + digits[:5] + " " + digits[5:]}, nil
}
