# Northstar Treks (placeholder name)

A trek listing and booking website written in Go. It builds to a single binary and stores everything in one SQLite file.

## Run it

```
go build -o northstar.exe .
./northstar.exe
```

Open http://localhost:8080. On first run the log prints the admin email and a generated password, so note it down.

| Setting | Flag | Env var | Default |
|---|---|---|---|
| Listen address | `-addr` | `ADDR` | `:8080` |
| Database file | `-db` | `DB_PATH` | `northstar.db` |
| First admin email | | `ADMIN_EMAIL` | `admin@northstar.local` |
| First admin password | | `ADMIN_PASSWORD` | random, printed once |

## Brand and contact details

The company name, phone, WhatsApp number, email and Instagram handle all live in `site.go`. Every line marked `PLACEHOLDER` must be replaced before going live. Rebuild after changing them.

Set `BASE_URL` (e.g. `https://yourdomain.in`) in production so WhatsApp/Instagram link previews use the right address.

## Policies

`templates/cancellation.html`, `terms.html` and `privacy.html` are starting templates. Review the refund rules and amounts, and have the legal pages checked, before going live.

## Trek details

Each trek has a best season, itinerary, inclusions, exclusions, what-to-carry list and FAQs, all edited in Admin → Treks → Edit. The sample content added on first run is generic, so adjust it per trek.

## Photos

Photos are read from the `photos/` folder next to the app (change it with `-photos` or `PHOTOS_DIR`). You don't need to rebuild after adding them:

- `photos/hero.jpg`: big background image on the start page
- `photos/gallery/*`: shown in the "From the trail" section
- any other image: pick it as a trek's cover photo in Admin → Treks → Edit

Without photos, the site falls back to a green mountain graphic.

## What it does

**Trekkers:** sign up, browse treks, request a booking (date, number of people, pickup point, age, emergency contact, medical notes), cancel a booking, and rate the trek and its lead once the trek is completed.

**Trek leads:** each lead has a public profile page. Its stats come from real data: treks led and trekkers guided count only *completed* bookings, and the rating is the average of trekker reviews. This is the starting point for the lead "score".

**Admin** (`/admin`):
- Bookings: approve or reject requests, then mark them completed or cancelled. Seats are only counted as taken once a booking is approved.
- Treks & batches: add or edit treks, add dates, assign a lead and change seats on each batch.
- Leads: add or edit lead profiles.
- Users: view the list.

There are no payments. Booking is request-only, and you confirm each request manually.

## Trek-lead applications

Anyone with an account can apply to lead treks at `/lead-with-us` (profile, treks done, certifications, photo upload). Applications stay hidden until an admin approves them in Admin → Trek Leads. Admins can also tick **Verified** once they've checked a lead's ID and experience, which adds a ✓ badge to the lead's profile. Uploaded photos are saved in `photos/leads/`.

## Booking status flow

```
requested ──approve──▶ approved ──complete──▶ completed (trekker can now review)
    │                     │
  reject               cancel
    ▼                     ▼
 rejected ◀──reopen── cancelled
```

## Files

- `main.go`: startup, routes, template helpers
- `db.go`: schema, queries, sample data
- `auth.go`: signup, login, sessions
- `handlers.go`: public and trekker pages
- `admin.go`: admin panel
- `templates/`, `static/`: HTML and CSS, embedded into the binary

## Sample data

On first run the app adds five Karnataka treks and three placeholder leads ("Founder One/Two/Three"). Distances, altitudes and prices are approximate. Edit them, or hide them, in the admin panel before going live.

## Deploying

Copy the binary to a server and run it behind a reverse proxy that handles HTTPS (e.g. Caddy). Back up the `.db` file regularly. To build for a Linux server from Windows:

```
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o northstar .
```
