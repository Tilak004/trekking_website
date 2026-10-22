#!/usr/bin/env bash
# One-time setup of a fresh Ubuntu 22.04 or 24.04 server (Oracle Cloud or any other).
# It installs Caddy (HTTPS), creates the "beyond" service user and folders, opens the
# firewall for web traffic, and installs the two systemd services. It does NOT start
# the website: deploy.ps1 uploads the program next and starts it.
#
#   sudo DOMAIN=beyondexpeditions.in bash server-setup.sh
#
# Safe to run again: it never overwrites an existing /etc/beyond/beyond.env.
set -euo pipefail

DOMAIN="${DOMAIN:-}"
[ -n "$DOMAIN" ] || { echo "Set your domain, e.g.: sudo DOMAIN=beyondexpeditions.in bash $0"; exit 1; }
[ "$(id -u)" = 0 ] || { echo "Run with sudo."; exit 1; }
HERE="$(cd "$(dirname "$0")" && pwd)"

echo "==> Time zone (backups run at 02:00 local time)"
timedatectl set-timezone Asia/Kolkata || true

echo "==> Service user and folders"
id beyond >/dev/null 2>&1 || useradd --system --home /opt/beyond --shell /usr/sbin/nologin beyond
mkdir -p /opt/beyond/photos /var/lib/beyond/backups /etc/beyond
chown -R beyond:beyond /opt/beyond /var/lib/beyond

echo "==> Swap (the small free machine has only 1 GB of memory)"
if ! swapon --show | grep -q .; then
  fallocate -l 1G /swapfile && chmod 600 /swapfile && mkswap /swapfile && swapon /swapfile
  grep -q '/swapfile' /etc/fstab || echo '/swapfile none swap sw 0 0' >> /etc/fstab
fi

echo "==> Packages (Caddy web server, firewall persistence)"
export DEBIAN_FRONTEND=noninteractive
apt-get update -y
apt-get install -y debian-keyring debian-archive-keyring apt-transport-https curl gpg iptables-persistent
if ! command -v caddy >/dev/null 2>&1; then
  curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
  curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' > /etc/apt/sources.list.d/caddy-stable.list
  apt-get update -y
  apt-get install -y caddy
fi

echo "==> Firewall: allow web traffic (Oracle's Ubuntu images block everything except SSH by default)"
for port in 80 443; do
  if ! iptables -C INPUT -p tcp --dport "$port" -j ACCEPT 2>/dev/null; then
    pos="$(iptables -L INPUT --line-numbers -n | awk '/REJECT/ {print $1; exit}')"
    if [ -n "$pos" ]; then iptables -I INPUT "$pos" -p tcp --dport "$port" -j ACCEPT; else iptables -A INPUT -p tcp --dport "$port" -j ACCEPT; fi
  fi
done
netfilter-persistent save

echo "==> Settings file"
if [ ! -f /etc/beyond/beyond.env ]; then
  sed "s/DOMAIN/$DOMAIN/g" "$HERE/beyond.env.example" > /etc/beyond/beyond.env
  chmod 640 /etc/beyond/beyond.env
  chown root:beyond /etc/beyond/beyond.env
  echo "    created /etc/beyond/beyond.env"
else
  echo "    kept your existing /etc/beyond/beyond.env"
fi

echo "==> Caddy config"
sed "s/DOMAIN/$DOMAIN/g" "$HERE/Caddyfile" > /etc/caddy/Caddyfile
systemctl enable caddy
systemctl restart caddy

echo "==> Services"
install -m 644 "$HERE/beyond-public.service" /etc/systemd/system/beyond-public.service
install -m 644 "$HERE/beyond-admin.service" /etc/systemd/system/beyond-admin.service
systemctl daemon-reload
systemctl enable beyond-public beyond-admin

echo
echo "Server ready. Next, from your computer, upload the website with deploy\\deploy.ps1."
