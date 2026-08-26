#!/usr/bin/env bash
# Backs up the production database, builds the Go binary for Linux, deploys
# it to the EC2 server, then restarts it via systemd.
#
# Config is read from server/.env (see server/.env.example), falling back
# to whatever is already set in the environment.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SERVER_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
ENV_FILE="$SERVER_DIR/.env"

if [[ -f "$ENV_FILE" ]]; then
  set -a
  # shellcheck disable=SC1090
  source "$ENV_FILE"
  set +a
fi

: "${SERVER_IP:?SERVER_IP is not set. Add it to $ENV_FILE or export it in the environment.}"

SSH_USER="${SSH_USER:-root}"
SSH_KEY="${SSH_KEY:-}"
REMOTE_DIR="${REMOTE_DIR:-/home/$SSH_USER}"
SERVICE_NAME="${SERVICE_NAME:-domain}"
BINARY_NAME="${BINARY_NAME:-domain}"
GOARCH_TARGET="${GOARCH_TARGET:-amd64}"

SSH_OPTS=(-o BatchMode=yes -o PreferredAuthentications=publickey)
if [[ -n "$SSH_KEY" ]]; then
  SSH_OPTS+=(-i "$SSH_KEY")
fi

DB_NAME="${DB_NAME:-portfolio}"
BACKUP_DIR="$SERVER_DIR/backups"

# ─────────────────────────────────────────────
# Back up the production database before touching anything else. This dump
# is taken via `sudo -u postgres pg_dump` over SSH (peer-auth as the
# `postgres` OS/DB role on the server), so it needs no DB password and works
# regardless of what app-level DB user DATABASE_URL uses. If the backup
# fails or comes back empty, the deploy aborts before the server is touched.
# ─────────────────────────────────────────────
echo "==> Backing up database '$DB_NAME' from $SSH_USER@$SERVER_IP"
mkdir -p "$BACKUP_DIR"
BACKUP_FILE="$BACKUP_DIR/${DB_NAME}_$(date +%Y%m%d_%H%M%S).sql"
if ! ssh "${SSH_OPTS[@]}" "$SSH_USER@$SERVER_IP" "sudo -u postgres pg_dump --no-owner --no-privileges '$DB_NAME'" > "$BACKUP_FILE"; then
  rm -f "$BACKUP_FILE"
  echo "==> Database backup failed — aborting deploy. Nothing on the server was touched." >&2
  exit 1
fi
if [[ ! -s "$BACKUP_FILE" ]]; then
  rm -f "$BACKUP_FILE"
  echo "==> Database backup came back empty — aborting deploy." >&2
  exit 1
fi
echo "==> Backup saved to $BACKUP_FILE ($(du -h "$BACKUP_FILE" | cut -f1))"

cd "$SERVER_DIR"

echo "==> Building $BINARY_NAME for linux/$GOARCH_TARGET"
GOOS=linux GOARCH="$GOARCH_TARGET" go build -o "$BINARY_NAME" ./cmd

echo "==> Copying $BINARY_NAME to $SSH_USER@$SERVER_IP:$REMOTE_DIR"
REMOTE_TMP="$REMOTE_DIR/$BINARY_NAME.new"
scp "${SSH_OPTS[@]}" "$BINARY_NAME" "$SSH_USER@$SERVER_IP:$REMOTE_TMP"
rm -f "$BINARY_NAME"

echo "==> Restarting $SERVICE_NAME via systemd"
ssh "${SSH_OPTS[@]}" "$SSH_USER@$SERVER_IP" "chmod +x '$REMOTE_TMP' && mv -f '$REMOTE_TMP' '$REMOTE_DIR/$BINARY_NAME' && sudo systemctl restart $SERVICE_NAME"

echo "==> Deploy complete"
