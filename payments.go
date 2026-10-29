package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	qrcode "github.com/skip2/go-qrcode"
)

// UPI payments: free, direct to the company's UPI ID. The trekker pays from any
// UPI app (amount pre-filled); the team checks their UPI app and marks the
// booking paid in Admin. Each booking has a secret pay token, so its payment
// link can be shared on WhatsApp without the trekker logging in.

func migratePayments(db *sql.DB) {
	db.Exec(`ALTER TABLE bookings ADD COLUMN pay_token TEXT NOT NULL DEFAULT ''`)
	db.Exec(`ALTER TABLE bookings ADD COLUMN paid_at TEXT NOT NULL DEFAULT ''`)
	db.Exec(`ALTER TABLE bookings ADD COLUMN paid_amount INTEGER NOT NULL DEFAULT 0`)
	// Bookings made before payments existed get a pay token too.
	db.Exec(`UPDATE bookings SET pay_token = lower(hex(randomblob(8))) WHERE pay_token = ''`)
}

// Amount is what the trekker owes: trek price × people.
func (k Booking) Amount() int { return k.Price * k.People }

// UPILink is a upi://pay deep link that opens any UPI app with the amount filled in.
func (k Booking) UPILink() string {
	q := url.Values{}
	q.Set("pa", site.UPI)
	q.Set("pn", site.Name)
	q.Set("am", strconv.Itoa(k.Amount())+".00")
	q.Set("cu", "INR")
	q.Set("tn", fmt.Sprintf("Booking %d %s", k.ID, k.TrekName))
	return "upi://pay?" + strings.ReplaceAll(q.Encode(), "+", "%20")
}

// PayPath is the shareable payment page for this booking.
func (k Booking) PayPath() string { return fmt.Sprintf("/pay/%d/%s", k.ID, k.PayToken) }

func (a *App) bookingByPayToken(r *http.Request) (Booking, error) {
	list, err := a.queryBookings(`WHERE k.id = ? AND k.pay_token = ? AND k.pay_token != ''`, pathID(r), r.PathValue("token"))
	if err != nil {
		return Booking{}, err
	}
	if len(list) == 0 {
		return Booking{}, sql.ErrNoRows
	}
	return list[0], nil
}

// payPage shows the UPI button and QR code for one booking.
func (a *App) payPage(w http.ResponseWriter, r *http.Request) {
	a.expireHolds()
	k, err := a.bookingByPayToken(r)
	if errors.Is(err, sql.ErrNoRows) {
		a.notFound(w, r)
		return
	} else if err != nil {
		a.serverError(w, err)
		return
	}
	a.render(w, r, "pay.html", map[string]any{"Booking": k})
}

// payQR is the QR code for the booking's UPI link, for paying from a laptop screen.
func (a *App) payQR(w http.ResponseWriter, r *http.Request) {
	k, err := a.bookingByPayToken(r)
	if err != nil {
		a.notFound(w, r)
		return
	}
	png, err := qrcode.Encode(k.UPILink(), qrcode.Medium, 480)
	if err != nil {
		a.serverError(w, err)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "private, max-age=600")
	w.Write(png)
}

// adminSetPaid marks a booking paid (with the amount received) or unpaid.
func (a *App) adminSetPaid(w http.ResponseWriter, r *http.Request) {
	back := safeNext(r.FormValue("back"))
	if back == "/" {
		back = "/admin/registrations"
	}
	sep := "?"
	if strings.Contains(back, "?") {
		sep = "&"
	}
	var err error
	if r.FormValue("paid") == "1" {
		_, err = a.db.Exec(`UPDATE bookings SET paid_at = datetime('now'), paid_amount = ?, hold_until = '' WHERE id = ?`, formInt(r, "amount"), pathID(r))
	} else {
		// Not paid after all: forget the trekker's UPI reference too, and don't let an old hold time cancel it.
		_, err = a.db.Exec(`UPDATE bookings SET paid_at = '', paid_amount = 0, utr = '', claimed_at = '', hold_until = '' WHERE id = ?`, pathID(r))
	}
	if err != nil {
		a.serverError(w, err)
		return
	}
	redirect(w, r, "%s%smsg=status", back, sep)
}

// payRequestLink opens WhatsApp to the customer with their payment details.
// The UPI ID and amount are in the text itself, so it works even before the
// site has a public address; the link adds the one-tap button and QR.
func payRequestLink(k Booking, base string) string {
	return whatsAppTo(k.UserPhone, payRequestText(k, base))
}

// payRequestText is the WhatsApp message asking a customer to pay.
func payRequestText(k Booking, base string) string {
	return fmt.Sprintf("Hi %s! Your seat on %s (%s) is confirmed.\n\n"+
		"Please pay ₹%s by UPI to *%s* (note: Booking %d). You can also scan the QR code I am sending with this message.\n"+
		"Pay page: %s%s\n\n"+
		"Once paid, reply here with the screenshot and we'll add you to the trek group.",
		strings.Fields(k.UserName + " there")[0], k.TrekName, formatDate(k.StartDate), formatINR(k.Amount()), site.UPI, k.ID, base, k.PayPath())
}

// adminPayRequest shows the QR and the buttons to send it to the customer.
func (a *App) adminPayRequest(w http.ResponseWriter, r *http.Request) {
	list, err := a.queryBookings(`WHERE k.id = ?`, pathID(r))
	if err != nil {
		a.serverError(w, err)
		return
	}
	if len(list) == 0 {
		a.notFound(w, r)
		return
	}
	k := list[0]
	a.render(w, r, "admin_payreq.html", map[string]any{
		"Booking": k, "Message": payRequestText(k, baseURL(r)), "WALink": payRequestLink(k, baseURL(r)),
	})
}
