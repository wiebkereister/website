#!/usr/bin/env bash
# Runs on the server as user "deploy" (called by the CI workflow):
#   deploy.sh <dev|prod> <release.tar.gz>
# The tarball holds the `wiebkereister` binary and the `seed/` directory.
set -euo pipefail

ENV="${1:?usage: deploy.sh <dev|prod> <tarball>}"
TARBALL="${2:?usage: deploy.sh <dev|prod> <tarball>}"
[[ "$ENV" == dev || "$ENV" == prod ]] || { echo "unknown environment: $ENV"; exit 1; }

ROOT="/opt/wiebkereister/$ENV"
DATA="/var/lib/wiebkereister/$ENV"
STAMP="$(date +%Y%m%d-%H%M%S)"
REL="$ROOT/releases/$STAMP"
PORT="$(grep -oP '(?<=:)\d+$' <(grep '^BASEPLATE_ADDR=' /etc/wiebkereister/$ENV.env))"

mkdir -p "$REL"
tar -xzf "$TARBALL" -C "$REL"
rm -f "$TARBALL"
chmod +x "$REL/wiebkereister"

# first deployment: import the initial content into the empty data directory
if [[ ! -e "$DATA/.seeded" ]]; then
  ( cd "$REL" && set -a && . "/etc/wiebkereister/$ENV.env" && set +a && ./wiebkereister seed )
  touch "$DATA/.seeded"
fi

PREV="$(readlink -f "$ROOT/current" || true)"
ln -sfn "$REL" "$ROOT/current.new" && mv -Tf "$ROOT/current.new" "$ROOT/current"
sudo /usr/bin/systemctl restart "wiebkereister@$ENV"

# health check; roll back to the previous release on failure
ok=0
for _ in $(seq 1 20); do
  code="$(curl -s -o /dev/null -w '%{http_code}' "http://127.0.0.1:$PORT/" || true)"
  # 401 = site password protection is active, which still proves the app is up
  if [[ "$code" == 200 || "$code" == 401 ]]; then ok=1; break; fi
  sleep 1
done

if [[ "$ok" != 1 ]]; then
  echo "health check failed (last HTTP status: $code)"
  if [[ -n "$PREV" && -d "$PREV" ]]; then
    ln -sfn "$PREV" "$ROOT/current.new" && mv -Tf "$ROOT/current.new" "$ROOT/current"
    sudo /usr/bin/systemctl restart "wiebkereister@$ENV"
    echo "rolled back to $PREV"
  fi
  exit 1
fi

echo "deployed $STAMP to $ENV"
# keep the last 5 releases
ls -1dt "$ROOT"/releases/* | tail -n +6 | xargs -r rm -rf
