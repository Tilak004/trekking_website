# Deploying to Oracle Cloud (Always Free) with beyondexpeditions.in

This puts the website on a free Oracle Cloud server, with your own domain and automatic HTTPS.

> **Using Microsoft Azure instead of Oracle?** Follow [DEPLOY-AZURE.md](DEPLOY-AZURE.md) for creating the server, then come back here for DNS, deploying and everything after. Any Ubuntu server works with this kit.

```
visitor  ->  https://beyondexpeditions.in  ->  Caddy (HTTPS)  ->  website (127.0.0.1:8080)
you      ->  SSH tunnel                    ->  admin console (127.0.0.1:8081)   never public
```

Everything lives on one server: the website, the admin console and the database file. The
website and the admin console are separate programs, and only the website is reachable from the
internet.

You do the account and server steps in the Oracle console (steps 1 to 5). After that, one
command deploys the site (step 6), and the same command updates it later.

> Oracle's console is renamed and rearranged from time to time. If a menu name below isn't an
> exact match, look for the closest one.

---

## Before you start

- A credit or debit card (Oracle verifies it with a small temporary charge; the free resources are not billed).
- A phone number for verification, and access to the account where you registered **beyondexpeditions.in**.
- On this computer: Go, Node.js, and the Windows "OpenSSH Client" (already present on Windows 11; check with `ssh -V`).

## 1. Create the Oracle Cloud account

1. Go to https://www.oracle.com/cloud/free/ and sign up.
2. **Choose your home region carefully, it cannot be changed later.** Pick **India South (Mumbai)** or **India South (Hyderabad)**: your trekkers are close, and "Always Free" resources only exist in the home region.
3. Finish card verification. You get a free trial credit as well; ignore it, the site only needs the Always Free resources.

## 2. Create the server

1. Make an SSH key on this computer (PowerShell):
   ```powershell
   ssh-keygen -t ed25519 -f $HOME\.ssh\oracle.key
   ```
   Press Enter for no passphrase (or set one). This creates `oracle.key` (private, never share it) and `oracle.key.pub` (public).
2. In the console: **Compute > Instances > Create instance**.
3. **Image:** Canonical **Ubuntu 24.04** (22.04 also works). Do not pick Oracle Linux: the scripts are written for Ubuntu.
4. **Shape:** click *Change shape*.
   - **VM.Standard.E2.1.Micro** (AMD, "Always Free-eligible", 1 GB RAM). Always available, and plenty for this site. Use `-Arch amd64`.
   - **VM.Standard.A1.Flex** (Ampere/Arm) is also free, but Oracle often says "out of host capacity", and its free allowance was reduced in 2026. If it works, use 1 or 2 OCPUs with 6 to 12 GB RAM and `-Arch arm64`.
5. **Networking:** create a new virtual cloud network with a **public subnet**, and keep *Assign a public IPv4 address* ticked.
6. **SSH keys:** *Upload public key file* and choose `oracle.key.pub`.
7. Create the instance. After a minute it shows a **public IP**. The login name is `ubuntu`.

Test the login (answer `yes` to the fingerprint question):
```powershell
ssh -i $HOME\.ssh\oracle.key ubuntu@<the-ip>
```
Type `exit` to leave.

## 3. Open the web ports in Oracle's network

Oracle blocks everything except SSH until you allow it (the setup script also opens the server's own firewall).

1. **Networking > Virtual cloud networks** > your network > your public subnet > its **Security List** (usually "Default Security List").
2. **Add ingress rules**, one for each of these:
   - Source CIDR `0.0.0.0/0`, protocol **TCP**, destination port **80**
   - Source CIDR `0.0.0.0/0`, protocol **TCP**, destination port **443**

Do **not** open 8080 or 8081. The website and the admin console only listen on the server itself.

## 4. Reserve the public IP (so it never changes)

An automatically assigned IP can change if the server is ever stopped. Reserve it:

1. **Networking > IP management > Reserved public IPs > Reserve public IP address.**
2. Open the instance > **Attached VNICs** > the VNIC > **IPv4 addresses** > *Edit*, choose *No public IP* and save, then edit again and choose *Reserved public IP* and pick the one you made.

Note the final IP. The next step uses it.

## 5. Point beyondexpeditions.in at the server

At the registrar where you bought the domain, open **DNS settings** and add:

| Type | Name / Host | Value | TTL |
|---|---|---|---|
| A | `@` | the Oracle public IP | 300 (or lowest) |
| A | `www` | the same IP | 300 |

Remove any parking-page records the registrar added for `@` and `www`. Check it from this computer (it can take a few minutes up to an hour):
```powershell
nslookup beyondexpeditions.in
```
It should show your Oracle IP. **Do not continue until it does**: the HTTPS certificate is requested on the first visit and needs the domain to already point at the server.

## 6. Deploy

From the project folder on this computer (`C:\Prahladh\Beyond Expeditions\trekking_website`):

```powershell
.\deploy\deploy.ps1 -Server ubuntu@<the-ip> -Key $HOME\.ssh\oracle.key -Domain beyondexpeditions.in -Setup -Photos
```

Add `-Arch arm64` if you chose an Ampere shape. This single command:
1. installs Caddy (HTTPS), sets the server's time zone to India, opens the server firewall, and creates the services;
2. builds the frontend and the Linux program;
3. uploads the program and the site photos (hero, gallery, logo. Lead and review photos are never uploaded);
4. starts everything.

Then open **https://beyondexpeditions.in**. The first load takes a few seconds while the certificate is issued.

## 7. Sign in to the admin console

The very first start creates the admin account with a random password. Read it:
```powershell
ssh -i $HOME\.ssh\oracle.key ubuntu@<the-ip> "sudo journalctl -u beyond-public --no-pager | grep -i password"
```
Then set your own email and password (replace the values):
```powershell
ssh -i $HOME\.ssh\oracle.key ubuntu@<the-ip> "sudo -u beyond env ADMIN_EMAIL='you@example.com' ADMIN_PASSWORD='a-long-password' /opt/beyond/beyond -db /var/lib/beyond/beyond.db -reset-admin-password"
```

The admin console is not on the internet. Reach it through an SSH tunnel and keep this window open while you work:
```powershell
ssh -i $HOME\.ssh\oracle.key -L 8081:127.0.0.1:8081 ubuntu@<the-ip>
```
Now open **http://localhost:8081** in your browser and sign in.

## 8. Before you tell anyone about it

- **Sample data.** The first start adds five sample treks and three placeholder leads ("Founder One/Two/Three"). Edit or hide them in the admin console (Treks, Trek leads).
- **Contact and payment details** in `site.go` (WhatsApp number, UPI ID, bookings number, email, Instagram). Fix anything wrong, then redeploy (step 9).
- **Legal pages:** terms, privacy and cancellation text are starting templates. Have them reviewed.
- **Test the whole journey:** make a booking, approve it in the console, open the payment page, mark it paid.
- **Email alerts for enquiries** (optional): set the `SMTP_*` and `ENQUIRY_EMAIL` lines in `/etc/beyond/beyond.env` on the server, then `sudo systemctl restart beyond-public`.

## 9. Updating the site later

```powershell
git pull
.\deploy\deploy.ps1 -Server ubuntu@<the-ip> -Key $HOME\.ssh\oracle.key
```
This rebuilds and replaces the program and restarts it. It never touches the database or uploaded photos. Add `-Photos` when you have added new site photos locally.

## 10. Backups (important)

- **Nightly automatic backup** at 02:00 India time: a single `.tar.gz` with the whole database and the uploaded photos, in `/var/lib/beyond/backups`. The newest 30 are kept.
- **A backup on the same server is lost if the server is.** Copy backups to Google Drive with rclone:
  1. On the server: `sudo apt-get install -y rclone`
  2. Set it up for the service user: `sudo -u beyond rclone config` and add a Google Drive remote named `gdrive`. The server has no browser, so use `rclone authorize "drive"` on this computer and paste the result when asked.
  3. In `/etc/beyond/beyond.env` uncomment `BACKUP_UPLOAD=rclone copy {file} gdrive:beyond-backups`, then `sudo systemctl restart beyond-public`.
  4. Test: `sudo -u beyond /opt/beyond/beyond -db /var/lib/beyond/beyond.db -backup-now`, and check the file appears in Drive.
- **Take a backup right before any risky change:** the `-backup-now` command above.
- **Restore:** stop the services (`sudo systemctl stop beyond-public beyond-admin`), unpack the `.tar.gz`, put the database back at `/var/lib/beyond/beyond.db` and the photos in `/opt/beyond/photos`, then `sudo chown -R beyond:beyond /var/lib/beyond /opt/beyond/photos` and start the services.

## 11. Stop Oracle from reclaiming the free server

Oracle may stop an Always Free server it considers **idle** (roughly: under about 20% CPU for 7 days). A quiet website can look idle. The usual protection is to **upgrade the account to "Pay As You Go"** (Billing > Upgrade). It still costs nothing while you stay inside the free limits, and a stopped instance can otherwise be restarted only if capacity remains. Also set a budget alert of about 1 US dollar so any accidental charge is flagged. Check Oracle's current Always Free page, as these rules have changed before.

## Troubleshooting

| Problem | What to check |
|---|---|
| Browser can't connect / times out | Step 3 (both rules saved?) and that DNS points at the server (step 5). |
| "Your connection is not private" | DNS not pointing at the server yet, or the certificate is still being issued. Look at `sudo journalctl -u caddy -n 50`. |
| 502 Bad Gateway | The website isn't running: `sudo systemctl status beyond-public` and `sudo journalctl -u beyond-public -n 50`. |
| Admin page won't open | The SSH tunnel window must stay open; then use http://localhost:8081. |
| Deploy says "Permission denied (publickey)" | The `-Key` path is wrong, or the public key wasn't added in step 2. |
| A change isn't visible | Hard-refresh the page (Ctrl+F5). |

Useful commands on the server:
```
sudo systemctl status beyond-public beyond-admin caddy
sudo journalctl -u beyond-public -f        # live log of the website
sudo systemctl restart beyond-public beyond-admin
```

## Costs

Oracle's Always Free server, storage and traffic allowance: free. HTTPS certificates (Let's Encrypt via Caddy): free. The only recurring cost is the domain's yearly renewal.
