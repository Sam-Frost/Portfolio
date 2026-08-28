# One-time server setup: friend's domain

These are manual, one-time steps to run yourself on the EC2 box
(`2.29.3.177`) to prepare it for a second, isolated instance of this app —
her own database, her own backend process on its own port, her own Caddy
site block serving her frontend as static files. I did **not** run any of
this — per your instructions I only inspected the server (read-only) and
am handing you the exact commands. Everything here is additive; nothing
touches the existing `portfolio` database, the `domain` systemd service,
or the `api.sat0ru.dev` Caddy block.

Do these in order. After they're done, `scripts/deploy-friend.sh` handles
every future deploy (frontend + backend) in one command.

## 0. Point DNS at the server first

Her domain's A (and AAAA, if you use IPv6) record needs to point at
`2.29.3.177` *before* you touch Caddy in step 4 — Caddy auto-provisions a
Let's Encrypt cert on first request and that fails if DNS isn't live yet.

## 1. Database — a dedicated role + database for her

Isolated from the main `portfolio` database/role on purpose, so nothing
she does (or a bug in her instance) can touch your data.

```bash
ssh root@2.29.3.177

sudo -u postgres psql -c "CREATE ROLE anjalibansal WITH LOGIN PASSWORD 'CHOOSE_A_STRONG_PASSWORD_HERE';"
sudo -u postgres psql -c "CREATE DATABASE portfolio_anjali OWNER anjalibansal;"
```

Pick a real password and keep it — it goes into her env file next. (The
existing `pg_hba.conf` already allows password auth over `localhost`, so
no auth config changes are needed.)

## 2. Spotify app (only if she wants the Spotify widget)

The backend currently treats `SPOTIFY_CLIENT_ID` / `SPOTIFY_CLIENT_SECRET`
/ `SPOTIFY_REDIRECT_URI` / `SPOTIFY_FRONTEND_REDIRECT_URL` /
`SPOTIFY_TOKEN_KEY` as **required at startup** (`main.go` calls
`requireEnv` on each — the binary won't boot without them, even if she
never uses the feature). Register a separate app at
developer.spotify.com/dashboard for her domain (redirect URI
`https://HER_DOMAIN/api/spotify/callback`), or if you'd rather she not
have Spotify at all, that's a small code change (make those `requireEnv`
calls conditional) — say the word and I'll do it in a follow-up rather
than guessing at it here. For now, the setup below assumes you'll fill in
real values.

## 3. Backend env file — `/etc/domain-anjali/domain.env`

```bash
sudo mkdir -p /etc/domain-anjali
sudo tee /etc/domain-anjali/domain.env > /dev/null <<'EOF'
PORT=8081
DOMAIN_PASSWORD=CHOOSE_A_PASSWORD_FOR_HER_GATE
JWT_SECRET=REPLACE_WITH_output_of_openssl_rand_-base64_32
DATABASE_URL=postgres://anjalibansal:CHOOSE_A_STRONG_PASSWORD_HERE@localhost:5432/portfolio_anjali?sslmode=disable
ALLOWED_ORIGIN=https://HER_DOMAIN

SPOTIFY_CLIENT_ID=
SPOTIFY_CLIENT_SECRET=
SPOTIFY_REDIRECT_URI=https://HER_DOMAIN/api/spotify/callback
SPOTIFY_FRONTEND_REDIRECT_URL=https://HER_DOMAIN/settings
SPOTIFY_TOKEN_KEY=REPLACE_WITH_output_of_openssl_rand_-base64_32

# Document Storage — no S3 for her instance, so bytes live on the box's
# disk and are served back through the backend. PUBLIC_API_URL must be her
# real domain so the signed blob URLs the browser hits resolve.
DOCUMENTS_LOCAL_DIR=/var/lib/domain-anjali/documents
PUBLIC_API_URL=https://HER_DOMAIN
EOF
sudo chmod 600 /etc/domain-anjali/domain.env
sudo mkdir -p /var/lib/domain-anjali/documents
```

Generate the two secrets with `openssl rand -base64 32` each — use
**different** values than the main app's `JWT_SECRET`/`SPOTIFY_TOKEN_KEY`;
never reuse those across instances.

## 4. systemd service — `domain-anjali`

```bash
sudo tee /etc/systemd/system/domain-anjali.service > /dev/null <<'EOF'
[Unit]
Description=Domain Server (Anjali)
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/root
ExecStart=/root/domain-anjali

Restart=always
RestartSec=3

EnvironmentFile=/etc/domain-anjali/domain.env

[Install]
WantedBy=multi-user.target
EOF
sudo systemctl daemon-reload
sudo systemctl enable domain-anjali
```

Leave it stopped for now — it won't start successfully until the binary
actually exists at `/root/domain-anjali`, which the first
`deploy-friend.sh` run provides.

## 5. Static site directory for Caddy

```bash
sudo mkdir -p /var/www/HER_DOMAIN
sudo chown root:root /var/www/HER_DOMAIN
```

## 6. Caddy site block

Append to `/etc/caddy/Caddyfile` (don't remove the existing
`api.sat0ru.dev` block above it):

```
HER_DOMAIN {
	encode gzip

	handle /api/* {
		reverse_proxy localhost:8081
	}
	handle /health {
		reverse_proxy localhost:8081
	}
	handle {
		root * /var/www/HER_DOMAIN
		try_files {path} /index.html
		file_server
	}
}
```

Then:

```bash
sudo systemctl reload caddy
```

Caddy will request a cert for `HER_DOMAIN` on the first real request,
which needs DNS from step 0 to already be live.

## 7. First deploy

From your machine, copy `scripts/deploy-friend.env.example` to
`scripts/deploy-friend.env`, fill in `FRIEND_DOMAIN` /
`FRIEND_STATIC_REMOTE_DIR` (`/var/www/HER_DOMAIN`) / `SERVER_IP`, then run
`scripts/deploy-friend.sh`. That builds and ships both the backend binary
and the frontend build, and starts `domain-anjali` for the first time.
