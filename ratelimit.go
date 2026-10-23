package main

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Login throttling. It slows down password guessing, which matters most for the
// admin console now that it is reachable from the internet.
//
//   - 5 wrong passwords for the same email from the same address, or 20 from one
//     address overall, within 15 minutes block that address for 15 minutes.
//   - Blocking is per address, never per email alone, so a stranger cannot lock
//     the real admin out by typing their email wrongly.
//
// The counters live in memory (a restart clears them) which is fine for this purpose.

const (
	failsPerPair  = 5
	failsPerIP    = 20
	failWindow    = 15 * time.Minute
	blockDuration = 15 * time.Minute
)

type failRecord struct {
	times        []time.Time
	blockedUntil time.Time
}

type loginLimiter struct {
	mu sync.Mutex
	m  map[string]*failRecord
}

var limiter = &loginLimiter{m: map[string]*failRecord{}}

// clientIP is the visitor's address. Behind our reverse proxy (Caddy, which
// connects from this machine) the real address is in X-Forwarded-For.
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	if ip := net.ParseIP(host); ip != nil && ip.IsLoopback() {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			if first := strings.TrimSpace(strings.Split(xff, ",")[0]); net.ParseIP(first) != nil {
				return first
			}
		}
	}
	return host
}

func pairKey(ip, email string) string { return "pair|" + ip + "|" + email }
func ipKey(ip string) string          { return "ip|" + ip }

// blocked reports how long this address must still wait, if it is blocked.
func (l *loginLimiter) blocked(ip, email string) (time.Duration, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	for _, k := range []string{pairKey(ip, email), ipKey(ip)} {
		if rec := l.m[k]; rec != nil && now.Before(rec.blockedUntil) {
			return rec.blockedUntil.Sub(now), true
		}
	}
	return 0, false
}

// fail records a wrong password and starts a block once a limit is reached.
func (l *loginLimiter) fail(ip, email string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	l.record(pairKey(ip, email), failsPerPair, now)
	l.record(ipKey(ip), failsPerIP, now)
	if len(l.m) > 5000 { // forget old entries so memory stays small
		for k, rec := range l.m {
			if now.After(rec.blockedUntil) && (len(rec.times) == 0 || now.Sub(rec.times[len(rec.times)-1]) > failWindow) {
				delete(l.m, k)
			}
		}
	}
}

func (l *loginLimiter) record(key string, limit int, now time.Time) {
	rec := l.m[key]
	if rec == nil {
		rec = &failRecord{}
		l.m[key] = rec
	}
	keep := rec.times[:0]
	for _, t := range rec.times {
		if now.Sub(t) < failWindow {
			keep = append(keep, t)
		}
	}
	rec.times = append(keep, now)
	if len(rec.times) >= limit {
		rec.blockedUntil = now.Add(blockDuration)
		rec.times = nil
	}
}

// success clears the counters for this address and email.
func (l *loginLimiter) success(ip, email string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.m, pairKey(ip, email))
}

// tooMany is the message shown when someone is blocked.
func tooMany(wait time.Duration) string {
	mins := int(wait.Minutes()) + 1
	return fmt.Sprintf("Too many wrong attempts. Please wait %d minute%s and try again.", mins, map[bool]string{true: "", false: "s"}[mins == 1])
}
