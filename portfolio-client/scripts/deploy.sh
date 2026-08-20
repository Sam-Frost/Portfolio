#!/usr/bin/env bash
#
# deploy.sh — Build a React app and deploy it to S3 + CloudFront.
#
# Usage:
#   ./deploy.sh
#
# Prerequisites:
#   - AWS CLI v2 installed and configured (`aws configure`)
#   - Node/npm installed
#   - A `.env` file in the project root (see .env.example)
#
# .env variables:
#   BUCKET_NAME      (required) S3 bucket name
#   DISTRIBUTION_ID  (required) CloudFront distribution ID
#   BUILD_DIR        (optional) build output dir; default "build"
#   AWS_PROFILE      (optional) AWS CLI profile; default "default"
#   BUILD_CMD        (optional) build command; default "npm run build"
#   S3_PREFIX        (optional) subfolder in the bucket; must match the
#                    CloudFront origin path. Empty = bucket root.

set -euo pipefail

# ─────────────────────────────────────────────
# Load configuration from .env
# ─────────────────────────────────────────────
ENV_FILE="../.env"

if [ ! -f "$ENV_FILE" ]; then
  echo "Error: $ENV_FILE not found. Copy .env.example to .env and fill in your values." >&2
  exit 1
fi

# Export all variables defined in .env
set -a
# shellcheck disable=SC1090
source "$ENV_FILE"
set +a

# ─────────────────────────────────────────────
# Defaults for optional values
# ─────────────────────────────────────────────
BUILD_DIR="${BUILD_DIR:-build}"
AWS_PROFILE="${AWS_PROFILE:-default}"
BUILD_CMD="${BUILD_CMD:-npm run build}"
# Subfolder inside the bucket where the site lives (must match the
# CloudFront origin path). Leave empty to deploy to the bucket root.
S3_PREFIX="${S3_PREFIX:-sat0ru}"

# Build the S3 destination, handling an empty prefix cleanly.
if [ -n "$S3_PREFIX" ]; then
  S3_DEST="s3://$BUCKET_NAME/$S3_PREFIX"
else
  S3_DEST="s3://$BUCKET_NAME"
fi

# Colors
GREEN='\033[0;32m'; YELLOW='\033[1;33m'; RED='\033[0;31m'; NC='\033[0m'
info()  { echo -e "${GREEN}==>${NC} $1"; }
warn()  { echo -e "${YELLOW}==>${NC} $1"; }
error() { echo -e "${RED}==>${NC} $1" >&2; }

# ─────────────────────────────────────────────
# Validate required variables
# ─────────────────────────────────────────────
: "${BUCKET_NAME:?Set BUCKET_NAME in .env}"
: "${DISTRIBUTION_ID:?Set DISTRIBUTION_ID in .env}"

# ─────────────────────────────────────────────
# Pre-flight checks
# ─────────────────────────────────────────────
command -v aws >/dev/null 2>&1 || { error "AWS CLI not found. Install it first."; exit 1; }
command -v npm >/dev/null 2>&1 || { error "npm not found. Install Node.js first."; exit 1; }

export AWS_PROFILE

# ─────────────────────────────────────────────
# 1. Build
# ─────────────────────────────────────────────
info "Building the React app..."
$BUILD_CMD

if [ ! -d "$BUILD_DIR" ]; then
  error "Build directory '$BUILD_DIR' not found. Check BUILD_DIR in .env."
  exit 1
fi

# ─────────────────────────────────────────────
# 2. Sync to S3
#    - Long cache for hashed static assets
#    - No cache for index.html so users get new deploys immediately
# ─────────────────────────────────────────────
info "Uploading hashed assets to $S3_DEST ..."
aws s3 sync "$BUILD_DIR" "$S3_DEST" \
  --delete \
  --exclude "index.html" \
  --exclude "*.map" \
  --cache-control "public,max-age=31536000,immutable"

info "Uploading index.html (no-cache)..."
aws s3 cp "$BUILD_DIR/index.html" "$S3_DEST/index.html" \
  --cache-control "no-cache,no-store,must-revalidate" \
  --content-type "text/html"

# ─────────────────────────────────────────────
# 3. Invalidate CloudFront cache
#    Paths are relative to the CloudFront root (origin path is stripped),
#    so we always invalidate "/*".
# ─────────────────────────────────────────────
info "Creating CloudFront invalidation..."
INVALIDATION_ID=$(aws cloudfront create-invalidation \
  --distribution-id "$DISTRIBUTION_ID" \
  --paths "/*" \
  --query 'Invalidation.Id' \
  --output text)

info "Invalidation created: $INVALIDATION_ID"
warn "It may take a few minutes to propagate."
info "Deployment complete! 🚀"
