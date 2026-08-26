#!/usr/bin/env bash
#
# deploy-friend.sh — deploys BOTH the frontend and backend for the friend's
# domain in one go: backs up her database, builds+ships the Go backend to
# its own systemd service on the shared EC2 box, builds the domain-client
# frontend and rsyncs it to the directory Caddy serves her domain from.
#
# One-time server setup (Postgres role/DB, systemd unit, Caddy site block,
# /etc/domain-friend/domain.env) is NOT done by this script — see
# scripts/friend-server-setup.md.
#
# Config is read from scripts/deploy-friend.env (copy from
# scripts/deploy-friend.env.example).
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
SERVER_DIR="$REPO_ROOT/server"
CLIENT_DIR="$REPO_ROOT/domain-client"
ENV_FILE="$SCRIPT_DIR/deploy-friend.env"

if [[ ! -f "$ENV_FILE" ]]; then
  echo "Error: $ENV_FILE not found. Copy scripts/deploy-friend.env.example to scripts/deploy-friend.env and fill in your values." >&2
  exit 1
fi

set -a
# shellcheck disable=SC1090
source "$ENV_FILE"
set +a

: "${SERVER_IP:?SERVER_IP is not set in $ENV_FILE}"
: "${FRIEND_DOMAIN:?FRIEND_DOMAIN is not set in $ENV_FILE}"
: "${FRIEND_STATIC_REMOTE_DIR:?FRIEND_STATIC_REMOTE_DIR is not set in $ENV_FILE}"

SSH_USER="${SSH_USER:-root}"
SSH_KEY="${SSH_KEY:-}"
FRIEND_REMOTE_DIR="${FRIEND_REMOTE_DIR:-/root}"
FRIEND_SERVICE_NAME="${FRIEND_SERVICE_NAME:-domain-friend}"
FRIEND_BINARY_NAME="${FRIEND_BINARY_NAME:-domain-friend}"
FRIEND_DB_NAME="${FRIEND_DB_NAME:-portfolio_friend}"
FRIEND_PUBLIC_SITE_URL="${FRIEND_PUBLIC_SITE_URL:-https://$FRIEND_DOMAIN}"

SSH_OPTS=(-o BatchMode=yes -o PreferredAuthentications=publickey)
if [[ -n "$SSH_KEY" ]]; then
  SSH_OPTS+=(-i "$SSH_KEY")
fi

GREEN='\033[0;32m'; RED='\033[0;31m'; NC='\033[0m'
info()  { echo -e "${GREEN}==>${NC} $1"; }
error() { echo -e "${RED}==>${NC} $1" >&2; }

# ─────────────────────────────────────────────
# 1. Back up her database before touching anything, same safety net as the
#    main deploy — pg_dump via `sudo -u postgres` over SSH, no DB password
#    needed, aborts the deploy if it fails or comes back empty.
# ─────────────────────────────────────────────
BACKUP_DIR="$SERVER_DIR/backups"
info "Backing up database '$FRIEND_DB_NAME' from $SSH_USER@$SERVER_IP"
mkdir -p "$BACKUP_DIR"
BACKUP_FILE="$BACKUP_DIR/${FRIEND_DB_NAME}_$(date +%Y%m%d_%H%M%S).sql"
if ! ssh "${SSH_OPTS[@]}" "$SSH_USER@$SERVER_IP" "sudo -u postgres pg_dump --no-owner --no-privileges '$FRIEND_DB_NAME'" > "$BACKUP_FILE"; then
  rm -f "$BACKUP_FILE"
  error "Database backup failed — aborting deploy. Nothing on the server was touched."
  exit 1
fi
if [[ ! -s "$BACKUP_FILE" ]]; then
  rm -f "$BACKUP_FILE"
  error "Database backup came back empty — aborting deploy."
  exit 1
fi
info "Backup saved to $BACKUP_FILE ($(du -h "$BACKUP_FILE" | cut -f1))"

# ─────────────────────────────────────────────
# 2. Backend: build, ship, restart her systemd service
# ─────────────────────────────────────────────
info "Building $FRIEND_BINARY_NAME for linux/amd64"
(cd "$SERVER_DIR" && GOOS=linux GOARCH=amd64 go build -o "$FRIEND_BINARY_NAME" ./cmd)

info "Copying $FRIEND_BINARY_NAME to $SSH_USER@$SERVER_IP:$FRIEND_REMOTE_DIR"
REMOTE_TMP="$FRIEND_REMOTE_DIR/$FRIEND_BINARY_NAME.new"
scp "${SSH_OPTS[@]}" "$SERVER_DIR/$FRIEND_BINARY_NAME" "$SSH_USER@$SERVER_IP:$REMOTE_TMP"
rm -f "$SERVER_DIR/$FRIEND_BINARY_NAME"

info "Restarting $FRIEND_SERVICE_NAME via systemd"
ssh "${SSH_OPTS[@]}" "$SSH_USER@$SERVER_IP" "chmod +x '$REMOTE_TMP' && mv -f '$REMOTE_TMP' '$FRIEND_REMOTE_DIR/$FRIEND_BINARY_NAME' && sudo systemctl restart $FRIEND_SERVICE_NAME"

# ─────────────────────────────────────────────
# 3. Frontend: build against her domain (same-origin API via Caddy's
#    reverse proxy, so VITE_API_BASE_URL is empty rather than a cross-
#    origin URL), rsync the static output to the directory Caddy serves.
#    Real process.env values passed here override .env.production, so
#    this never touches domain-client's own .env files.
# ─────────────────────────────────────────────
info "Building domain-client frontend for https://$FRIEND_DOMAIN"
(
  cd "$CLIENT_DIR"
  VITE_API_BASE_URL="" VITE_PUBLIC_SITE_URL="$FRIEND_PUBLIC_SITE_URL" bun run build
)

if [[ ! -d "$CLIENT_DIR/dist" ]]; then
  error "Build output $CLIENT_DIR/dist not found."
  exit 1
fi

info "Syncing frontend build to $SSH_USER@$SERVER_IP:$FRIEND_STATIC_REMOTE_DIR"
rsync -az --delete -e "ssh ${SSH_OPTS[*]}" "$CLIENT_DIR/dist/" "$SSH_USER@$SERVER_IP:$FRIEND_STATIC_REMOTE_DIR/"

info "Deploy complete! Backend on port via $FRIEND_SERVICE_NAME, frontend served by Caddy at https://$FRIEND_DOMAIN"
