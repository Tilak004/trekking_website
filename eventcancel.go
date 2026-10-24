package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
)

// Cancelling an event (a trek on a date) from our side, e.g. when too few
// seats fill to book transport. Every live registration is cancelled as
// "we cancelled", so paid trekkers are owed a full refund (tracked with
// "Mark refunded"), the date disappears from the website, and each trekker
// gets a one-tap WhatsApp message explaining it. Bookings are kept, unlike
// deleting, so payments and refunds stay on record.

const BatchCancelled = "cancelled"

var cancelReasons = map[string]string{
	"low":     "we didn't get enough trekkers to run it",
	"weather": "of bad weather on the trail",
	"closed":  "the forest department has closed the trail",
	"other":   "of reasons beyond our control",
}

func migrateEventCancel(db *sql.DB) {
	db.Exec(`ALTER TABLE batches ADD COLUMN cancel_reason TEXT NOT NULL DEFAULT ''`)
}

func (a *App) consoleCancelEvent(w http.ResponseWriter, r *http.Request) {
	e, err := a.eventByID(pathID(r))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if e.Status == BatchCancelled {
		redirect(w, r, "/admin/events/%d", e.ID)
		return
	}
	reason := r.FormValue("reason")
	if _, ok := cancelReasons[reason]; !ok {
		reason = "other"
	}
	list, err := a.queryBookings(`WHERE k.batch_id = ? AND k.status IN ('requested', 'approved')`, e.ID)
	if err != nil {
		a.serverError(w, err)
		return
	}
	for _, k := range list {
		if err := a.cancelWithRefund(k, RefundUs); err != nil {
			a.serverError(w, err)
			return
		}
	}
	if _, err := a.db.Exec(`UPDATE batches SET status = ?, cancel_reason = ? WHERE id = ?`, BatchCancelled, reason, e.ID); err != nil {
		a.serverError(w, err)
		return
	}
	redirect(w, r, "/admin/events/%d?msg=event_cancelled", e.ID)
}

// CancelNoticeWA opens WhatsApp to a trekker whose trek we cancelled, with the
// reason and what happens to their money written.
func (k Booking) CancelNoticeWA() string {
	if k.Status != "cancelled" || k.RefundReason != RefundUs {
		return ""
	}
	why := cancelReasons[k.CancelReason]
	if why == "" {
		why = cancelReasons["other"]
	}
	lines := []string{
		fmt.Sprintf("Hi %s, we're really sorry: the %s trek on %s has been cancelled because %s.",
			firstName(k.UserName), k.TrekName, formatDate(k.StartDate), why),
		"",
	}
	switch {
	case k.RefundedAt != "":
		lines = append(lines, fmt.Sprintf("Your full refund of ₹%s has been sent.", formatINR(k.RefundAmount)))
	case k.RefundAmount > 0:
		lines = append(lines, fmt.Sprintf("You'll get a full refund of ₹%s to the account you paid from, within 7 working days.", formatINR(k.RefundAmount)))
	default:
		lines = append(lines, "Nothing was paid, so there's nothing to refund.")
	}
	lines = append(lines, "", "We'd love to have you on another date or trek:", publicURL()+"/treks/"+k.TrekSlug)
	return whatsAppTo(k.UserPhone, strings.Join(lines, "\n"))
}
