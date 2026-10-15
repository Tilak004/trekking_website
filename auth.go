package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	sessionCookie = "session"
	sessionTTL    = 30 * 24 * time.Hour
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
		HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: r.TLS != nil,
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
	http.Redirect(w, r, safeNext(next), http.StatusSeeOther)
}

func (a *App) logout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(sessionCookie); err == nil {
		a.db.Exec(`DELETE FROM sessions WHERE token = ?`, c.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: "", Path: "/", MaxAge: -1})
	redirect(w, r, "/")
}
