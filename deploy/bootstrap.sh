#!/usr/bin/env bash
# One-time (idempotent) server setup. Run as root on a fresh Debian/Ubuntu:
#   sudo bash bootstrap.sh "<ssh public key of the CI deploy key>"
# Safe to re-run (e.g. to roll out a changed Caddyfile).
set -euo pipefail

PUBKEY="${1:?usage: bootstrap.sh <deploy-public-key>}"
HERE="$(cd "$(dirname "$0")" && pwd)"

export DEBIAN_FRONTEND=noninteractive
apt-get update -qq
apt-get install -y -qq debian-keyring debian-archive-keyring apt-transport-https curl gpg ufw unattended-upgrades

# Caddy from the official repo
if ! command -v caddy >/dev/null; then
  curl -fsSL https://dl.cloudsmith.io/public/caddy/stable/gpg.key | gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
  curl -fsSL https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt >/etc/apt/sources.list.d/caddy-stable.list
  apt-get update -qq && apt-get install -y -qq caddy
fi

# deploy user: runs the services and receives CI deployments
id deploy >/dev/null 2>&1 || useradd --create-home --shell /bin/bash deploy
install -d -m 700 -o deploy -g deploy /home/deploy/.ssh
echo "$PUBKEY" >/home/deploy/.ssh/authorized_keys
chown deploy:deploy /home/deploy/.ssh/authorized_keys && chmod 600 /home/deploy/.ssh/authorized_keys

# deploy may only restart/inspect the two services
cat >/etc/sudoers.d/deploy <<'SUDO'
deploy ALL=(root) NOPASSWD: /usr/bin/systemctl restart wiebkereister@dev, /usr/bin/systemctl restart wiebkereister@prod, /usr/bin/systemctl is-active wiebkereister@dev, /usr/bin/systemctl is-active wiebkereister@prod
SUDO
chmod 440 /etc/sudoers.d/deploy && visudo -cf /etc/sudoers.d/deploy

# per-environment layout, config and ports (backends listen on localhost only)
install -d -m 755 /etc/wiebkereister
for spec in dev:18081 prod:18080; do
  env="${spec%%:*}"; port="${spec##*:}"
  install -d -o deploy -g deploy /opt/wiebkereister/$env /var/lib/wiebkereister/$env
  if [[ ! -f /etc/wiebkereister/$env.env ]]; then
    cat >/etc/wiebkereister/$env.env <<ENV
BASEPLATE_ADDR=127.0.0.1:$port
BASEPLATE_DATA_DIR=/var/lib/wiebkereister/$env
BASEPLATE_DIST_DIR=/var/lib/wiebkereister/$env/dist
BASEPLATE_BUILD_ON_PUBLISH=1
BASEPLATE_TRUSTED_PROXY_HEADER=X-Forwarded-For
# Set to 1 once the domain runs over HTTPS and the IP access is no longer used for the admin:
BASEPLATE_SECURE_COOKIES=0
ENV
  fi
done

install -m 644 "$HERE/wiebkereister@.service" /etc/systemd/system/wiebkereister@.service
systemctl daemon-reload
systemctl enable wiebkereister@dev wiebkereister@prod

# reverse proxy
install -m 644 "$HERE/Caddyfile" /etc/caddy/Caddyfile
caddy validate --config /etc/caddy/Caddyfile --adapter caddyfile
systemctl enable --now caddy
systemctl reload caddy || systemctl restart caddy

# firewall
ufw allow 22/tcp >/dev/null
ufw allow 80/tcp >/dev/null
ufw allow 443/tcp >/dev/null
ufw allow 8081/tcp >/dev/null  # dev via IP; remove once test.wiebkereister.de works
ufw --force enable >/dev/null

echo "bootstrap done. Services start with the first deployment."
