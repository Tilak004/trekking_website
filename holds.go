package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Seat holds. Booking is instant: the trekker's seat is held for holdMinutes while
// they pay by UPI, and they send us the 12-digit UPI reference (UTR) from their
// payment. The console then shows "Payment claimed": the admin checks the UPI
// app, taps Confirm, and WhatsApp opens with the confirmation message written.
// A hold nobody pays within holdMinutes is released automatically. A booking that
// has a UTR waiting for the admin never expires.
//
//	approved + hold_until set, no UTR   → seat held, countdown running
//	approved + UTR, not paid            → payment claimed, waiting for the admin
//	approved + paid_at                  → confirmed
//
// Older bookings (status "requested", or approved by hand) have no hold_until
// and keep working as before.

const holdMinutes = 10

func migrateHolds(db *sql.DB) {
	db.Exec(`ALTER TABLE bookings ADD COLUMN hold_until TEXT NOT NULL DEFAULT ''`)
	db.Exec(`ALTER TABLE bookings ADD COLUMN utr TEXT NOT NULL DEFAULT ''`)
	db.Exec(`ALTER TABLE bookings ADD COLUMN claimed_at TEXT NOT NULL DEFAULT ''`)
}

// RefundExpired marks a booking cancelled because its hold ran out.
const RefundExpired = "expired"

// expireHolds cancels unpaid bookings whose hold ran out, freeing their seats.
func (a *App) expireHolds() {
	res, err := a.db.Exec(`UPDATE bookings SET status = 'cancelled', refund_reason = ?
WHERE status = 'approved' AND paid_at = '' AND utr = '' AND hold_until != '' AND hold_until < datetime('now')`, RefundExpired)
	if err != nil {
		log.Printf("expire holds: %v", err)
		return
	}
	if n, _ := res.RowsAffected(); n > 0 {
		log.Printf("released %d expired seat hold(s)", n)
	}
}

// startHoldSweeper releases expired holds every 15 seconds.
func (a *App) startHoldSweeper() {
	go func() {
		for range time.Tick(15 * time.Second) {
			a.expireHolds()
		}
	}()
}

// Held is true while the seat is reserved and waiting for the trekker to pay.
func (k Booking) Held() bool {
	return k.Status == StatusApproved && k.PaidAt == "" && k.Utr == "" && k.HoldUntil != ""
}

// Claimed is true once the trekker has sent a UTR that the admin has not verified yet.
func (k Booking) Claimed() bool {
	return k.Status == StatusApproved && k.PaidAt == "" && k.Utr != ""
}

// HoldSecondsLeft is how long the hold has left (0 when not held or already over).
func (k Booking) HoldSecondsLeft() int {
	if !k.Held() {
		return 0
	}
	until, err := time.ParseInLocation("2006-01-02 15:04:05", k.HoldUntil, time.UTC)
	if err != nil {
		return 0
	}
	return max(0, int(time.Until(until).Seconds()))
}

// UPI reference numbers (UTR / transaction ID) are 12 digits.
var utrPattern = regexp.MustCompile(`^[0-9]{12}$`)

// payClaim is the trekker telling us they paid, with the UPI reference.
func (a *App) payClaim(w http.ResponseWriter, r *http.Request) {
	a.expireHolds()
	k, err := a.bookingByPayToken(r)
	if err != nil {
		a.notFound(w, r)
		return
	}
	back := k.PayPath()
	if k.Status != StatusApproved || k.PaidAt != "" {
		redirect(w, r, "%s", back)
		return
	}
	utr := strings.Join(strings.Fields(r.FormValue("utr")), "")
	if !utrPattern.MatchString(utr) {
		redirect(w, r, "%s?err=bad_utr", back)
		return
	}
	var used int
	if err := a.db.QueryRow(`SELECT COUNT(*) FROM bookings WHERE utr = ? AND id != ?`, utr, k.ID).Scan(&used); err != nil {
		a.serverError(w, err)
		return
	}
	if used > 0 {
		redirect(w, r, "%s?err=utr_used", back)
		return
	}
	if _, err := a.db.Exec(`UPDATE bookings SET utr = ?, claimed_at = datetime('now') WHERE id = ? AND status = 'approved' AND paid_at = ''`,
		utr, k.ID); err != nil {
		a.serverError(w, err)
		return
	}
	redirect(w, r, "%s?msg=utr_sent", back)
}

// adminRejectClaim is for a UTR that doesn't match a payment: the claim is cleared
// and the trekker gets another hold period to pay and send the right reference.
func (a *App) adminRejectClaim(w http.ResponseWriter, r *http.Request) {
	back, sep := backTo(r)
	if _, err := a.db.Exec(`UPDATE bookings SET utr = '', claimed_at = '', hold_until = datetime('now', '+`+strconv.Itoa(holdMinutes)+` minutes')
WHERE id = ? AND status = 'approved' AND paid_at = ''`, pathID(r)); err != nil {
		a.serverError(w, err)
		return
	}
	redirect(w, r, "%s%smsg=claim_rejected", back, sep)
}

// ConfirmWA opens WhatsApp to the trekker with their booking confirmation (and the
// trek group invite when the trek has one). The console opens it as it confirms.
func (k Booking) ConfirmWA() string {
	if k.Status != StatusApproved {
		return ""
	}
	paid := k.Amount()
	if k.PaidAmount > 0 {
		paid = k.PaidAmount
	}
	lines := []string{
		fmt.Sprintf("Hi %s! We've received your payment of ₹%s. Your seat is confirmed ✅", firstName(k.UserName), formatINR(paid)),
		"",
		fmt.Sprintf("Booking #%d: %s", k.ID, k.TrekName),
		"Date: " + formatDate(k.StartDate),
		fmt.Sprintf("People: %d", k.People),
	}
	if k.Pickup != "" {
		lines = append(lines, "Pickup: "+k.Pickup)
	}
	if k.WhatsAppGroup != "" {
		lines = append(lines, "", fmt.Sprintf("Please join the %s WhatsApp group for pickup details and updates:", k.TrekName), k.WhatsAppGroup,
			"(We'll approve your request to join shortly.)")
	}
	lines = append(lines, "", "See you on the trail!")
	return whatsAppTo(k.UserPhone, strings.Join(lines, "\n"))
}
