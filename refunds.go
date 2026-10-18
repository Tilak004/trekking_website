package main

import (
	"database/sql"
	"net/http"
	"strings"
	"time"
)

// Cancellations follow the cancellation policy (frontend/src/public/pages/Static.jsx):
//
//	customer cancels 7+ days before the trek  → full refund minus ₹200
//	customer cancels 3–6 days before          → 50% refund
//	customer cancels under 3 days before      → no refund
//	we cancel (weather, closures, too few)    → full refund
//
// A cancelled paid booking records the refund owed; the team sends it by UPI
// and then marks it refunded.

const (
	RefundCustomer = "customer" // the trekker cancelled: policy refund
	RefundUs       = "us"       // we cancelled: full refund
)

const policyFee = 200 // ₹ kept when cancelling 7+ days ahead

func migrateRefunds(db *sql.DB) {
	db.Exec(`ALTER TABLE bookings ADD COLUMN refund_amount INTEGER NOT NULL DEFAULT 0`)
	db.Exec(`ALTER TABLE bookings ADD COLUMN refund_reason TEXT NOT NULL DEFAULT ''`)
	db.Exec(`ALTER TABLE bookings ADD COLUMN refunded_at TEXT NOT NULL DEFAULT ''`)
}

// DaysToTrek is the number of days from today until the trek starts.
func (k Booking) DaysToTrek() int {
	start, err := time.Parse("2006-01-02", k.StartDate)
	if err != nil {
		return 0
	}
	now, _ := time.Parse("2006-01-02", today())
	return int(start.Sub(now).Hours() / 24)
}

// TrekPassed reports whether the trek date has arrived, so it can be marked done.
func (k Booking) TrekPassed() bool { return k.StartDate <= today() }

// CustomerRefund is what the policy refunds if the trekker cancels today.
func (k Booking) CustomerRefund() int {
	switch d := k.DaysToTrek(); {
	case d >= 7:
		return max(0, k.PaidAmount-policyFee)
	case d >= 3:
		return k.PaidAmount / 2
	default:
		return 0
	}
}

// RefundRule explains the customer refund in a few words.
func (k Booking) RefundRule() string {
	switch d := k.DaysToTrek(); {
	case d >= 7:
		return "7+ days before: full refund minus ₹200"
	case d >= 3:
		return "3–6 days before: 50% refund"
	default:
		return "under 3 days before: no refund"
	}
}

// cancelWithRefund cancels a booking and records the refund owed.
func (a *App) cancelWithRefund(k Booking, reason string) error {
	refund := 0
	if k.PaidAt != "" {
		refund = k.CustomerRefund()
		if reason == RefundUs {
			refund = k.PaidAmount
		}
	}
	_, err := a.db.Exec(`UPDATE bookings SET status = 'cancelled', refund_amount = ?, refund_reason = ? WHERE id = ?`,
		refund, reason, k.ID)
	return err
}

func backTo(r *http.Request) (string, string) {
	back := safeNext(r.FormValue("back"))
	if back == "/" {
		back = "/admin/registrations"
	}
	if strings.Contains(back, "?") {
		return back, "&"
	}
	return back, "?"
}

// adminCancelBooking cancels a requested or approved booking; for paid ones the
// reason decides the refund (customer → policy, us → full).
func (a *App) adminCancelBooking(w http.ResponseWriter, r *http.Request) {
	back, sep := backTo(r)
	list, err := a.queryBookings(`WHERE k.id = ?`, pathID(r))
	if err != nil {
		a.serverError(w, err)
		return
	}
	if len(list) == 0 || (list[0].Status != StatusRequested && list[0].Status != StatusApproved) {
		http.Error(w, "That booking can't be cancelled.", http.StatusBadRequest)
		return
	}
	reason := RefundCustomer
	if r.FormValue("reason") == RefundUs {
		reason = RefundUs
	}
	if err := a.cancelWithRefund(list[0], reason); err != nil {
		a.serverError(w, err)
		return
	}
	redirect(w, r, "%s%smsg=cancelled", back, sep)
}

// adminMarkRefunded records that the refund has been sent.
func (a *App) adminMarkRefunded(w http.ResponseWriter, r *http.Request) {
	back, sep := backTo(r)
	if _, err := a.db.Exec(`UPDATE bookings SET refunded_at = datetime('now') WHERE id = ? AND status = 'cancelled' AND refund_amount > 0`,
		pathID(r)); err != nil {
		a.serverError(w, err)
		return
	}
	redirect(w, r, "%s%smsg=refunded", back, sep)
}
