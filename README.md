# public.charity

A platform that connects people who need help with people who can help, based on geography, needs, offers, and availability. An AI agent matches users and sends warm introductions to their inbox.

## Architecture

Three services, all on [Fly.io](https://fly.io) (London region):

| Service | Stack | Purpose |
|---------|-------|---------|
| `auth/` | Go, SQLite, Resend | Magic-link login, RS256 JWT issuance |
| `dashboard/` | Next.js 16, SQLite (Drizzle), Resend | User profiles, need sliders, availability, inbox |
| `agent/` | Go, SQLite, Gemini Flash Lite | Matching engine — finds complementary pairs and sends introductions |

```
┌──────────┐     magic link      ┌──────────┐
│          │◄───────────────────►│          │
│  Auth    │   RS256 JWT         │ Dashboard│  ◄── users interact here
│ :8080    │                     │ :3000    │
└──────────┘                     └────┬─────┘
     signs JWT with                   │
     private key                      │ GET /api/users
                                      │ POST /api/messages
                                 ┌────┴─────┐
                                 │  Agent   │  ◄── runs every 4h
                                 │ (no HTTP)│
                                 └──────────┘
```

## How it works

### User journey

1. User visits https://public.charity/login, enters email
2. Magic link arrives via Resend, user clicks it
3. Dashboard sets an RS256 JWT as an HttpOnly cookie
4. User lands on their profile dashboard where they can:
   - Set their name and bio
   - Share their location (browser geolocation)
   - Adjust 20 need sliders across 5 categories (Social, Health, Daily Life, Economic, Practical)
   - Add interests
   - Add things they can offer
   - Tap availability slots (7-day x morning/afternoon/evening grid)
5. All changes auto-save

### Matching (agent service)

Every 4 hours the agent:

1. Fetches all users via `GET /api/users` (API-key protected)
2. Groups users by geographic cell (+ adjacent cells, ~1 mile grid)
3. For each pair in a cluster, scores compatibility:
   - **Need↔Offer**: User A scores high on a category where User B has an offer (and vice versa)
   - **Availability overlap**: both must share at least one free time slot
   - **Interest bonus**: shared interests add to the score
4. Ranks candidates, takes top 20 per run
5. Generates a warm introduction message per user via Gemini Flash Lite (falls back to template)
6. Sends both messages via `POST /api/messages` (auto-forwarded to email)
7. Records the match — same pair won't be re-matched for 7 days

### Need categories

| Section | Categories |
|---------|-----------|
| Social & Connection | companionship, community, family, language |
| Health & Wellbeing | mental_health, physical_health, nutrition, substance_recovery |
| Daily Life | housing, transport, errands, personal_care |
| Economic Opportunities | employment, training, benefits, budgeting |
| Practical Support | digital, admin, legal, childcare |

Each scored 0.0–1.0 by the user. Scores above 0.4 are considered "active needs" for matching.

## Database schema

### Auth service (`auth/data/auth.db`)

- `magic_tokens` — single-use login tokens (value, email, expires_at)
- `schema_migrations` — migration tracking

### Dashboard (`dashboard/data/dashboard.db`)

- `users` — id, email, display_name, bio, lat/lng, cell_id, account_type, onboarding_step
- `need_scores` — (user_id, category) → score (0.0–1.0)
- `interests` — user_id, label
- `offers` — user_id, category, description, available
- `surplus` — user_id, category, description, expires_at
- `availability` — (user_id, day, slot) presence = free
- `messages` — inbox: recipient_id, sender_type (user/system/ai_agent), subject, body, read, email_sent

### Agent (`agent/data/agent.db`)

- `match_history` — (user_a, user_b, matched_at) dedup tracking

## Security

- **RS256 JWT** — auth service signs with an RSA private key, dashboard verifies with the public key only. Dashboard cannot forge tokens even if compromised.
- **Timing-safe API key comparison** — `timingSafeEqual` on all API-key-protected endpoints
- **HttpOnly cookies** — JWT stored as `pc_session`, secure in production, SameSite=lax
- **Single-use magic tokens** — consumed atomically with `DELETE...RETURNING`, expired tokens stay for sweep
- **No passwords** — email-only authentication via Resend

## Run locally

### 1. Generate RSA keys

```bash
openssl genrsa -out private.pem 2048
openssl rsa -in private.pem -pubout -out public.pem
```

### 2. Auth service

```bash
cd auth
export JWT_PRIVATE_KEY="$(cat ../private.pem)"
export RESEND_API_KEY=re_xxx
export EMAIL_FROM="noreply@yourdomain.com"
export DASHBOARD_URL=http://localhost:3000
go run ./cmd/auth
```

### 3. Dashboard

```bash
cd dashboard
cp .env.local.example .env.local
# Edit .env.local: set JWT_PUBLIC_KEY, RESEND_API_KEY, DASHBOARD_API_KEY
npm install
npm run dev
```

### 4. Agent

```bash
cd agent
export DASHBOARD_URL=http://localhost:3000
export DASHBOARD_API_KEY=<same key as dashboard's DASHBOARD_API_KEY>
export MATCH_INTERVAL=30s  # faster for local testing
# Optional: export GEMINI_API_KEY=xxx
go run ./cmd/agent
```

Visit http://localhost:3000/login.

## Deploy to Fly

### Auth

```bash
cd auth
fly apps create public-charity-auth
fly volumes create auth_data --region lhr --size 1 --yes

# Generate keys
openssl genrsa -out /tmp/jwt_private.pem 2048
openssl rsa -in /tmp/jwt_private.pem -pubout -out /tmp/jwt_public.pem

fly secrets set \
  "JWT_PRIVATE_KEY=$(cat /tmp/jwt_private.pem)" \
  RESEND_API_KEY=re_xxx \
  EMAIL_FROM="noreply@yourdomain.com" \
  DASHBOARD_URL=https://public.charity

# Create Tigris bucket (auto-sets AWS_* secrets)
fly storage create --app public-charity-auth --name pc-auth-backups --yes

fly deploy
```

### Dashboard

```bash
cd dashboard
fly apps create public-charity-dashboard
fly volumes create dashboard_data --region lhr --size 1 --yes

fly secrets set \
  "JWT_PUBLIC_KEY=$(cat /tmp/jwt_public.pem)" \
  RESEND_API_KEY=re_xxx \
  EMAIL_FROM="noreply@yourdomain.com" \
  DASHBOARD_API_KEY=$(openssl rand -hex 32)

fly storage create --app public-charity-dashboard --name pc-dashboard-backups --yes
fly deploy

# Scale to 1 machine (SQLite = single writer)
fly scale count 1 --app public-charity-dashboard --yes
```

### Agent

```bash
cd agent
fly apps create public-charity-agent
fly volumes create agent_data --region lhr --size 1 --yes

# Get the dashboard API key
DASH_KEY=$(fly ssh console --app public-charity-dashboard -C "printenv DASHBOARD_API_KEY")

fly secrets set \
  DASHBOARD_URL=https://public.charity \
  DASHBOARD_API_KEY=$DASH_KEY \
  --app public-charity-agent
# Optional: fly secrets set GEMINI_API_KEY=xxx --app public-charity-agent

fly deploy
```

### Custom domain

```bash
fly certs create public.charity --app public-charity-dashboard
# Point DNS A/AAAA records to the dashboard's Fly IPs
```

## API reference

### Auth service (`:8080`)

| Method | Path | Auth | Body | Response |
|--------|------|------|------|----------|
| GET | `/health` | — | — | `ok` |
| POST | `/auth/request` | — | `{"email":"..."}` | `{"status":"sent"}` |
| POST | `/auth/verify` | — | `{"token":"..."}` | `{"jwt":"...","email":"..."}` |
| POST | `/auth/validate` | — | `{"jwt":"..."}` | claims or 401 |

### Dashboard (`:3000`)

| Method | Path | Auth | Body | Response |
|--------|------|------|------|----------|
| GET | `/api/users` | Bearer API key | — | Array of user objects with scores, offers, interests, availability |
| POST | `/api/messages` | Bearer API key | `{"recipient_email","subject","body","sender_type?","category?","rule_id?"}` | `{"id","email_sent"}` |
| POST | `/api/auth/request` | Session cookie | `{"email":"..."}` | Proxied to auth |
| POST | `/api/auth/logout` | Session cookie | — | Clears cookie, 303 redirect |

## Project structure

```
public-charity/
├── auth/                           # Go auth service
│   ├── cmd/auth/main.go            # Entrypoint, graceful shutdown
│   ├── internal/
│   │   ├── config/config.go        # Env vars (JWT_PRIVATE_KEY, RESEND_API_KEY, etc.)
│   │   ├── jwt/jwt.go              # RS256 sign + verify
│   │   ├── token/token.go          # Magic link tokens (random 32 bytes, TTL)
│   │   ├── store/store.go          # SQLite token store (single-use consume)
│   │   ├── email/email.go          # Resend HTTP client
│   │   ├── server/server.go        # HTTP handlers + CORS
│   │   └── db/                     # SQLite connection + migrations
│   ├── Dockerfile                  # Go build + Litestream
│   ├── entrypoint.sh               # Restore from S3 → litestream replicate -exec auth
│   ├── litestream.yml              # SQLite → Tigris replication config
│   └── fly.toml
│
├── dashboard/                      # Next.js dashboard
│   ├── app/
│   │   ├── page.tsx                # Landing / redirect to dashboard
│   │   ├── login/page.tsx          # Email form
│   │   ├── auth/callback/route.ts  # Magic link landing → set JWT cookie
│   │   ├── dashboard/
│   │   │   ├── page.tsx            # Main profile: sliders, basics, interests
│   │   │   ├── availability/       # Tap-to-toggle free times grid
│   │   │   ├── inbox/              # Message list + detail view
│   │   │   ├── offers/             # Manage what you can offer
│   │   │   └── surplus/            # Manage surplus items
│   │   └── api/
│   │       ├── users/route.ts      # GET all users (agent API, key-protected)
│   │       └── messages/route.ts   # POST inject message (agent API)
│   ├── components/
│   │   ├── dashboard/
│   │   │   ├── NeedSliders.tsx     # 20 categories, 5 sections, dot scale
│   │   │   ├── ProfileBasics.tsx   # Name, bio, account type, location
│   │   │   ├── InterestEditor.tsx  # Token-style tag input
│   │   │   └── AvailabilityGrid.tsx# 7x3 toggle grid
│   │   └── Nav.tsx                 # Top nav with inbox badge
│   ├── lib/
│   │   ├── db/schema.ts            # Drizzle ORM table definitions
│   │   ├── db/index.ts             # SQLite connection (lazy singleton)
│   │   ├── actions/                # Server actions (profile, needs, interests, etc.)
│   │   ├── session.ts              # RS256 JWT verification (public key only)
│   │   └── email/forward.ts        # Resend email forwarding
│   ├── Dockerfile                  # Node build + better-sqlite3 + Litestream
│   ├── entrypoint.sh
│   ├── litestream.yml
│   └── fly.toml
│
├── agent/                          # Go matching agent
│   ├── cmd/agent/main.go           # Ticker loop, runs every 4h
│   ├── internal/
│   │   ├── config/config.go        # Env vars
│   │   ├── client/client.go        # HTTP client for dashboard APIs
│   │   ├── matcher/match.go        # Cell clustering, pair scoring, ranking
│   │   ├── history/history.go      # SQLite dedup (7-day suppression)
│   │   └── llm/gemini.go           # Gemini Flash Lite message generation
│   ├── Dockerfile
│   └── fly.toml
│
└── README.md
```

## Data pattern

All three services follow the same infrastructure pattern:

- **SQLite** on a Fly volume (WAL mode) — `modernc.org/sqlite` in Go, `better-sqlite3` + Drizzle in Node
- **Litestream** continuously replicates WAL frames to Tigris (S3-compatible) — restore on boot if DB missing
- **Fly.io** single-machine deployment per service (SQLite = single writer)
- On deploy: `entrypoint.sh` restores from S3 if needed, then Litestream wraps the app process
