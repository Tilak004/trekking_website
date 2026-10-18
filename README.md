# Beyond Expeditions

A trek listing and booking website written in Go. It builds to a single binary and stores everything in one SQLite file.

## Run it

```
go build -o beyond.exe .
./beyond.exe
```

Open http://localhost:8080. On first run the log prints the admin email and a generated password, so note it down.

| Setting | Flag | Env var | Default |
|---|---|---|---|
| Listen address | `-addr` | `ADDR` | `:8080` |
| Database file | `-db` | `DB_PATH` | `beyond.db` |
| First admin email | | `ADMIN_EMAIL` | `admin@beyondexpeditions.local` |
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

## Admin console (separate server)

The admin area is not part of the public website. It runs as its own process on its own port, using the same database, so `/admin` does not exist on the public server at all.

```
./northstar.exe                  # public site, http://localhost:8080
./northstar.exe -mode admin      # admin console, http://localhost:8081
```

Start both from the same folder (they share the database and the `photos/` folder). Sign in with an admin account; non-admin accounts are refused. The admin login uses its own session cookie, separate from the public site's. Settings: `-mode` / `MODE`, `-addr` / `ADDR` (default `:8081` in admin mode), and `PUBLIC_URL` for the "View public site" link, and (on the public server) `ADMIN_URL` to show admins an "Admin console" link in their account menu. In production, keep the admin port off the public internet (firewall, VPN, or a private reverse-proxy route).

- **Dashboard:** key numbers, upcoming events and what needs attention (requests to approve, proposed dates, drafts, lead applications).
- **Events:** an event is a trek on a date. Create it as a **draft** (invisible to trekkers), fill in the date, group size and lead, then **Publish**. Unpublishing is blocked while an event has active registrations. Dates proposed by trek leads show up here for approval.
- **Registrations:** everyone who registered, with phone/WhatsApp, group size, age, pickup, emergency contact and medical notes. Filter by event and status, approve/reject/complete, and download a CSV.
- **Analytics:** money (collected, still to collect, booked value) and people: by month, bookings by status, age and pickup split, new accounts, and per-trek and per-lead tables.
- **Treks, Trek leads, Users:** add or edit treks (itinerary, price, photos), lead profiles and applications, and the user list.

**Money:** payments are UPI (see `payments.go`). "Collected" is the amount you've marked paid on bookings; "booked value" is trek price × people on approved or completed bookings, and "still to collect" is the booked value not yet marked paid. Registrations show the payment buttons (request payment, mark paid).

There are no payments. Booking is request-only, and you confirm each request manually.

## Interactive parts (Preact islands)

The server renders every page. A few parts are upgraded in the browser by small Preact components ("islands"): the trek finder, the Western Ghats map, the elevation profile and the step-by-step booking flow. Each one loads only on pages that use it (the map only when scrolled into view), and the plain HTML still works if JavaScript fails.

Source is in `web/src/`; the built files in `static/js/` are committed and embedded in the Go binary, so **Node is only needed to change the islands**, not to run the site:

```
cd web
npm install
npm run build              # writes ../static/js
node test/islands.test.mjs # quick smoke tests
```

Trek coordinates, finder tags and elevation profiles are edited in Admin → Treks.

## Demo data

To see the site full of sample trek leads, trekkers, trek dates, bookings and reviews without touching real data:

```
cd web && npm run seed:demo      # copies beyond.db → beyond-demo.db and adds sample people (made with Faker)
cd .. && ./beyond.exe -addr :8092 -db beyond-demo.db
```

Every demo account uses the password `demo1234` and an `@example.com` email, and demo phone numbers start with `00000`. **Never run the live site on the demo database**: the sample leads and reviews are made up.

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
- `admin.go`: trek, lead and user admin handlers; `console.go`, `analytics.go`: the admin console (events, registrations, dashboard, analytics)
- `templates/`, `static/`: HTML and CSS, embedded into the binary

## Sample data

On first run the app adds five Karnataka treks and three placeholder leads ("Founder One/Two/Three"). Distances, altitudes and prices are approximate. Edit them, or hide them, in the admin panel before going live.

## Deploying

Copy the binary to a server and run it behind a reverse proxy that handles HTTPS (e.g. Caddy). Back up the `.db` file regularly. To build for a Linux server from Windows:

```
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o beyond .
```
