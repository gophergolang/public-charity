# public.charity

Two services:

- `auth/` — Go (stdlib + Resend) magic-link auth service. SQLite via `modernc.org/sqlite`, backed up to S3 with Litestream.
- `dashboard/` — Next.js app router, minimal scaffold.

## Auth flow

1. User visits `/login`, enters email.
2. Dashboard proxies `POST /api/auth/request` → auth's `POST /auth/request`.
3. Auth generates a single-use magic token (15 min TTL), stores it in SQLite, emails a link via Resend.
4. User clicks link → `{DASHBOARD_URL}/auth/callback?token=...`.
5. Dashboard server-side POSTs the token to auth's `POST /auth/verify`.
6. Auth consumes the token (single-use), returns `{jwt, email}`.
7. Dashboard sets `pc_session` HttpOnly cookie, redirects to `/`.

## Data pattern

Both services follow the same pattern:

- **SQLite** on a Fly volume (WAL mode, single-writer) — `modernc.org/sqlite` in Go, `better-sqlite3` / Drizzle in Node
- **Litestream** continuously replicates the SQLite WAL to any S3-compatible bucket (R2, B2, Tigris, MinIO, S3)
- **Tailscale** for inter-service communication — services join the tailnet, reach each other at stable DNS names

On deploy, the entrypoint restores from the latest S3 snapshot (if any), then starts Litestream as a supervisor around the app binary. WAL frames are shipped to S3 in near-real-time.

## Run locally

### Auth service

```bash
cd auth
export JWT_SECRET=$(openssl rand -hex 32)
export RESEND_API_KEY=re_xxx
export EMAIL_FROM="Public Charity <auth@your-verified-domain.com>"
export DASHBOARD_URL=http://localhost:3000
export DATABASE_PATH=./data/auth.db
go run ./cmd/auth
```

### Dashboard

```bash
cd dashboard
cp .env.local.example .env.local
npm install
npm run dev
```

Visit http://localhost:3000/login.

## Deploy to Fly

### 1. Create the auth app

```bash
cd auth
fly apps create public-charity-auth
fly volumes create auth_data --region lhr --size 1
```

### 2. Set secrets

```bash
fly secrets set \
  JWT_SECRET=$(openssl rand -hex 32) \
  RESEND_API_KEY=re_xxx \
  EMAIL_FROM="Public Charity <auth@your-verified-domain.com>" \
  DASHBOARD_URL=https://dashboard.public.charity \
  LITESTREAM_BUCKET=your-bucket \
  LITESTREAM_ENDPOINT=https://s3.us-east-1.amazonaws.com \
  LITESTREAM_REGION=us-east-1 \
  LITESTREAM_ACCESS_KEY_ID=AKxxx \
  LITESTREAM_SECRET_ACCESS_KEY=xxx
```

### 3. Deploy

```bash
fly deploy
```

### 4. Tailscale (inter-service access)

Enable [Fly + Tailscale](https://fly.io/docs/networking/tailscale/) so auth + dashboard can reach each other at `{app}.flycast` over the tailnet instead of the public internet. Set `AUTH_URL` in dashboard's env to `http://public-charity-auth.flycast:8080`.

## Endpoints

| Method | Path             | Body               | Response                    |
| ------ | ---------------- | ------------------- | --------------------------- |
| GET    | `/health`        | —                   | `ok`                        |
| POST   | `/auth/request`  | `{"email": "..."}`  | `{"status": "sent"}`        |
| POST   | `/auth/verify`   | `{"token": "..."}`  | `{"jwt": "...", "email":…}` |
| POST   | `/auth/validate` | `{"jwt": "..."}`    | claims, or 401              |
