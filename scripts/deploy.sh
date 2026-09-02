#!/usr/bin/env bash
#
# deploy.sh — deploys BOTH halves of the main (sat0ru) domain instance in
# one go by running each project's own deploy script in sequence:
#
#   1. server/scripts/deploy.sh
#        Backs up the production database, builds the Go binary for Linux,
#        ships it to the EC2 box, restarts the `domain` systemd service.
#        The whole deploy aborts here if the backup or build fails, before
#        anything on the server is touched.
#
#   2. domain-client/scripts/deploy.sh
#        Builds the React app (production mode ⇒ points at
#        https://api.sat0ru.dev via .env.production) and syncs it to
#        S3 + invalidates CloudFront.
#
# This wrapper introduces NO new configuration: each sub-script reads its
# own env file — server/.env and domain-client/.env respectively (both
# gitignored, see the matching .env.example files).
#
# Usage:
#   scripts/deploy.sh                 # backend, then frontend
#   scripts/deploy.sh --backend-only  # just server/scripts/deploy.sh
#   scripts/deploy.sh --frontend-only # just domain-client/scripts/deploy.sh
#
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

DO_BACKEND=1
DO_FRONTEND=1
case "${1:-}" in
  "")               ;;
  --backend-only)   DO_FRONTEND=0 ;;
  --frontend-only)  DO_BACKEND=0 ;;
  -h|--help)
    sed -n '2,25p' "${BASH_SOURCE[0]}" | sed 's/^# \{0,1\}//'
    exit 0
    ;;
  *)
    echo "Unknown argument: $1 (expected --backend-only or --frontend-only)" >&2
    exit 2
    ;;
esac

GREEN='\033[0;32m'; RED='\033[0;31m'; NC='\033[0m'
info()  { echo -e "${GREEN}==>${NC} $1"; }
error() { echo -e "${RED}==>${NC} $1" >&2; }

BACKEND_SCRIPT="$REPO_ROOT/server/scripts/deploy.sh"
FRONTEND_SCRIPT="$REPO_ROOT/domain-client/scripts/deploy.sh"

if [[ "$DO_BACKEND" == 1 && ! -x "$BACKEND_SCRIPT" ]]; then
  error "$BACKEND_SCRIPT not found or not executable."
  exit 1
fi
if [[ "$DO_FRONTEND" == 1 && ! -x "$FRONTEND_SCRIPT" ]]; then
  error "$FRONTEND_SCRIPT not found or not executable."
  exit 1
fi

if [[ "$DO_BACKEND" == 1 ]]; then
  info "Deploying backend (server/) …"
  # deploy.sh resolves its own paths via BASH_SOURCE, so CWD doesn't matter.
  "$BACKEND_SCRIPT"
  info "Backend deploy done."
fi

if [[ "$DO_FRONTEND" == 1 ]]; then
  info "Deploying frontend (domain-client/) …"
  # domain-client/scripts/deploy.sh reads "../.env" and writes "../dist"
  # relative to its own directory, so it must run from there.
  ( cd "$REPO_ROOT/domain-client/scripts" && ./deploy.sh )
  info "Frontend deploy done."
fi

info "All done. 🚀"
