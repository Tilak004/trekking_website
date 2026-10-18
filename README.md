# Beyond Expeditions

A trek listing and booking website. The backend is Go (one binary, everything stored in one SQLite file) and the frontend is a React app, for both the public website and the admin console.

## What you need

- **Go 1.26 or newer** (https://go.dev/dl/). This is all you need to run the site.
- **Node.js 20 or newer** only if you want to change the frontend (`frontend/`). The built frontend is already committed in `static/app/`.

## Run it

Two programs run from the same folder: the public website, and the admin console. They share one database.

```
go build -o beyond.exe .

./beyond.exe                  # public website   -> http://localhost:8080
./beyond.exe -mode admin      # admin console    -> http://localhost:8081   (in a second terminal)
```

On Windows PowerShell use `.\beyond.exe`. For quick testing you can skip the build step and use `go run .` and `go run . -mode admin` instead.

On the very first run the log of the public site prints the admin email and a generated password. **Note it down**, it is shown once. Sign in to the admin console at http://localhost:8081 with that email and password.

Forgot the admin password, or want to choose your own?

```
./beyond.exe -reset-admin-password                       # sets a new random password and prints it
$env:ADMIN_EMAIL="you@example.com"; $env:ADMIN_PASSWORD="a-long-password"; ./beyond.exe -reset-admin-password
```

If you changed anything in `frontend/`, rebuild it **before** building the Go binary:

```
cd frontend
npm install               # first time only
npm run build             # writes ../static/app
cd ..
go build -o beyond.exe .
```

| Setting | Flag | Env var | Default |
|---|---|---|---|
| Which server to run | `-mode` (`public` or `admin`) | `MODE` | `public` |
| Listen address | `-addr` | `ADDR` | `:8080` (`:8081` in admin mode) |
| Database file | `-db` | `DB_PATH` | `beyond.db` |
| Photos folder | `-photos` | `PHOTOS_DIR` | `photos` |
| First admin email | | `ADMIN_EMAIL` | `admin@beyondexpeditions.local` |
| First admin password | | `ADMIN_PASSWORD` | random, printed once |
| Public site address (link previews) | | `BASE_URL` | taken from the request |
| "View public site" link in the console | | `PUBLIC_URL` | `http://localhost:8080` |
| "Admin console" link for admins on the public site | | `ADMIN_URL` | none (link hidden) |

Settings can be given as flags (`./beyond.exe -addr :9000`) or as environment variables. Backup settings are listed under Backups below.

## Brand and contact details

The company name, phone, WhatsApp number, email and Instagram handle all live in `site.go`. Every line marked `PLACEHOLDER` must be replaced before going live. Rebuild after changing them.

Set `BASE_URL` (e.g. `https://yourdomain.in`) in production so WhatsApp/Instagram link previews use the right address.

## Policies

`frontend/src/public/pages/Static.jsx` holds the cancellation, terms and privacy text; they are starting templates. Review the refund rules and amounts, and have the legal pages checked, before going live.

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
./beyond.exe                  # public site, http://localhost:8080
./beyond.exe -mode admin      # admin console, http://localhost:8081
```

Start both from the same folder (they share the database and the `photos/` folder). Sign in with an admin account; non-admin accounts are refused. The admin login uses its own session cookie, separate from the public site's. Settings: `-mode` / `MODE`, `-addr` / `ADDR` (default `:8081` in admin mode), and `PUBLIC_URL` for the "View public site" link, and (on the public server) `ADMIN_URL` to show admins an "Admin console" link in their account menu. In production, keep the admin port off the public internet (firewall, VPN, or a private reverse-proxy route).

- **Dashboard:** key numbers, upcoming events and what needs attention (requests to approve, proposed dates, drafts, lead applications).
- **Events:** an event is a trek on a date. Create it as a **draft** (invisible to trekkers), fill in the date, group size and lead, then **Publish**. Unpublishing is blocked while an event has active registrations. Dates proposed by trek leads show up here for approval.
- **Registrations:** everyone who registered, with phone/WhatsApp, group size, age, pickup, emergency contact and medical notes. Filter by event and status, approve/reject/complete, and download a CSV.
- **Analytics:** money (collected, still to collect, booked value) and people: by month, bookings by status, age and pickup split, new accounts, and per-trek and per-lead tables.
- **Treks, Trek leads, Users:** add or edit treks (itinerary, price, photos), lead profiles and applications, and the user list.

**Money:** payments are UPI (see `payments.go`). "Collected" is the amount you've marked paid on bookings; "booked value" is trek price × people on approved or completed bookings, and "still to collect" is the booked value not yet marked paid. Registrations show the payment buttons (request payment, mark paid).

## Backups

The public-site process backs up automatically every night at 02:00. Each backup is one file, `backups/beyond-YYYY-MM-DD_HHMMSS.tar.gz`, containing a **complete** snapshot of the database plus the photos people uploaded (`photos/leads`, `photos/reviews`). The newest 30 are kept.

```
./beyond.exe -backup-now          # make a backup right now (e.g. before an update)
```

| Setting | Default | Meaning |
|---|---|---|
| `BACKUP_DIR` | `backups` | where backups go; `off` disables them |
| `BACKUP_AT` | `02:00` | time of the nightly backup |
| `BACKUP_KEEP` | `30` | how many backups to keep |
| `BACKUP_UPLOAD` | – | command to copy each backup to Google Drive: `rclone copy {file} gdrive:beyond-backups` |

**Restore:** stop both processes, extract the archive (`tar -xzf beyond-….tar.gz`), put `beyond.db` and the `photos/` folders in place, and start again.

Keep a copy **off the server** in Google Drive (set `BACKUP_UPLOAD`); a backup on the same disk is lost if the server is. Setup on the server: install rclone, run `rclone config` to add a Google Drive remote named `gdrive` (authorise it with `rclone authorize "drive"` on a laptop, since the server has no browser), then run `beyond -backup-now` and check the file appears in Drive.

## Frontend (React)

The whole frontend is a React app in `frontend/`: the public website and the admin console, each its own entry (`index.html` and `admin.html`) sharing one codebase (React 18, React Router, Vite, plus Leaflet for the map). The Go server no longer renders HTML pages. It answers JSON, and the React app draws it:

- Every page's data comes from `GET /api/<page path>` (for example `/api/treks/kudremukh`). Forms `POST` to `/api/...` and get `{"redirect": "/where/next"}` back. The Go handlers are the same ones as before; `render()` and `redirect()` simply answer in JSON for `/api` requests (see `api.go`).
- Any other browser request gets the app's HTML shell, so refreshing or sharing a deep link works. The server fills in the page title and link-preview tags (WhatsApp/Instagram show the right trek photo) before sending it.
- Private data never reaches public pages: `api.go` strips fields such as exact trailhead coordinates, lead phone numbers and other people's emails from public responses.

The built app is committed in `static/app/` and embedded in the Go binary, so **Node is only needed to change the frontend**, not to run the site:

```
cd frontend
npm install
npm run build        # writes ../static/app; then rebuild the Go binary
```

Source layout: `src/public/` (website pages), `src/admin/` (console pages), `src/lib/` (API client, helpers), `src/components/` (shared pieces). Styling is still the hand-written `static/style.css` and `static/console.css`.

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

- `main.go`: startup, routes, flash messages
- `db.go`: schema, queries, sample data
- `auth.go`: signup, login, sessions
- `handlers.go`: public and trekker pages
- `admin.go`: trek, lead and user admin handlers; `console.go`, `analytics.go`: the admin console (events, registrations, dashboard, analytics)
- `api.go`: the JSON API layer and the React app shell
- `frontend/`: the React app; `static/`: CSS, logos and the built app, embedded into the binary

## Sample data

On first run the app adds five Karnataka treks and three placeholder leads ("Founder One/Two/Three"). Distances, altitudes and prices are approximate. Edit them, or hide them, in the admin panel before going live.

## Deploying

Copy the binary (and the `photos/` folder) to a server and run **both** programs, the public site and `-mode admin`, from the same folder so they share the database. Put the public site behind a reverse proxy that handles HTTPS (e.g. Caddy), and keep the admin port (8081) private: firewall it, or reach it over a VPN or SSH tunnel. Set `BASE_URL` to the public address. Back up the `.db` file regularly (see Backups).

To build for a Linux server from Windows (run `npm run build` in `frontend/` first if you changed it):

```
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o beyond .
```
