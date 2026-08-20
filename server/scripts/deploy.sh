#!/usr/bin/env bash
# Builds the Go binary for Linux and deploys it to the EC2 server, then
# restarts it via systemd.
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
