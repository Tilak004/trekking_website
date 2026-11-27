# Website analytics (Google Analytics 4)

Property: **Beyond Expeditions**, Measurement ID `G-52R5MKQDDF`. The tag loads on the
live site only (`frontend/index.html`), never on the admin console. Counts are per
browser, not per person: the same person on two devices is two users.

## One-time settings in Google Analytics (not in the code)

1. **Admin → Data streams → your website → Enhanced measurement → ⚙ → turn OFF "Outbound clicks".**
   It sends the full address of every link clicked, which would include WhatsApp links
   with phone numbers and message text, and the private trek group invite.
2. **Admin → Data display → Events → mark as key event** (so they show as conversions):
   `begin_checkout`, `booking_requested`, `payment_claimed`, `booking_confirmed`.
3. **Admin → Data streams → Configure tag settings → Define internal traffic**, add your own
   IP, then Data settings → Data filters → activate "Internal Traffic". Keeps your own
   visits out of the numbers.

## Events the site sends

| Event | When | Extra info |
|---|---|---|
| `page_view` | each page opened (also moving between pages) | cleaned address: `/pay` and `/my/bookings` without token or id |
| `view_item` | a trek page opened (not secret treks) | `item_id` = trek slug, `item_name` |
| `begin_checkout` | picked a date and went to the contact-details step | `item_id` |
| `booking_requested` | booking request sent (seat held) | `item_id`, `people` |
| `payment_claimed` | UPI reference sent ("I've paid") | `item_id` |
| `booking_confirmed` | trekker opened a booking we marked paid, once per booking per browser | `item_id`, `value`, `currency` |
| `whatsapp_click` | any wa.me link (enquire buttons) | `link_location`, `item_id` |
| `join_whatsapp_group` | trek group invite link | same |
| `follow_whatsapp_channel` | WhatsApp channel link | same |
| `call_click`, `email_click`, `instagram_click` | tel:, mailto:, Instagram links | same |

Never sent: names, phone numbers, e-mail addresses, booking numbers, payment tokens, link addresses.

`booking_confirmed` is counted when the trekker looks at their booking page after we
confirm payment, so it can lag or miss trekkers who never come back. The authoritative
count of confirmed bookings is the admin console.

## Reading it

- **Funnel:** Explore → Funnel exploration, steps `view_item` → `begin_checkout` →
  `booking_requested` → `payment_claimed` → `booking_confirmed`.
- **Which treks get attention:** Reports → Engagement → Events → `view_item` → `item_id`, or
  Explore → Free form with `item_id` against the booking events.
- **New vs returning:** Reports → Retention.
- **Where visitors came from:** Reports → Acquisition → Traffic acquisition.
- **Contact taps:** Reports → Engagement → Events, the `*_click` events.

## Tagged links (use a different one for each place you post)

Add `?utm_source=…&utm_medium=…&utm_campaign=…` to any page address. Keep names lower case.

| Where | Link |
|---|---|
| Instagram bio | `https://beyondexpeditions.in/?utm_source=instagram&utm_medium=bio&utm_campaign=profile` |
| Instagram story | `https://beyondexpeditions.in/?utm_source=instagram&utm_medium=story&utm_campaign=general` |
| Instagram reel or post caption | `https://beyondexpeditions.in/?utm_source=instagram&utm_medium=reel&utm_campaign=general` |
| WhatsApp status | `https://beyondexpeditions.in/?utm_source=whatsapp&utm_medium=status&utm_campaign=general` |
| WhatsApp group or broadcast | `https://beyondexpeditions.in/?utm_source=whatsapp&utm_medium=group&utm_campaign=general` |
| Poster or QR code | `https://beyondexpeditions.in/?utm_source=poster&utm_medium=qr&utm_campaign=general` |
| A trek, e.g. Bananthimari Betta from an Instagram story | `https://beyondexpeditions.in/treks/bananthimari-betta?utm_source=instagram&utm_medium=story&utm_campaign=bananthimari` |

Change `utm_campaign` to name a push (a trek, a festival offer). Results show under
Reports → Acquisition → Traffic acquisition → add the secondary dimension "Session campaign".

## Google Search Console (what people searched to find the site)

The site is already verified (the `google…html` file is served at the root). At
https://search.google.com/search-console choose the property for `https://beyondexpeditions.in`,
then **Sitemaps → add `sitemap.xml`** if it isn't listed. **Performance** shows the searches,
clicks and ranking. Data starts about two days after the sitemap is read. Linking it to
Analytics (Admin → Product links → Search Console links) shows the searches inside GA too.
