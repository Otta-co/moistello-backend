# Moistello Backend — Enterprise Implementation Plan

---

## 1. Architecture Overview

```
                              ┌──────────────────────────┐
                              │     STELLAR BLOCKCHAIN     │
                              │  ┌────────┐ ┌──────────┐  │
                              │  │Horizon │ │Soroban   │  │
                              │  │  API   │ │   RPC    │  │
                              │  └───┬────┘ └────┬─────┘  │
                              │      │            │        │
                              │  ┌───┴────────────┴───┐   │
                              │  │  Soroban Contracts   │   │
                              │  │  (Rust)              │   │
                              │  └──────────────────────┘   │
                              └──────────┬───────────────────┘
                                         │
                              ┌──────────▼───────────────────┐
                              │       INDEXER SERVICE        │
                              │  (Go)                         │
                              │  Horizon polling + event sync │
                              │  Cursor tracking + reconciler │
                              └──────────┬───────────────────┘
                                         │ (writes to DB)
                              ┌──────────▼───────────────────┐
                              │       POSTGRESQL 16          │
                              │  users, circles, contributions│
                              │  payouts, notifications,      │
                              │  audit_log, webhooks, etc.    │
                              └──────────┬───────────────────┘
                                         │
                    ┌────────────────────┼────────────────────┐
                    │                    │                    │
           ┌────────▼────────┐  ┌───────▼────────┐  ┌───────▼────────┐
           │   API SERVER    │  │  NOTIFICATION   │  │    WEBHOOK     │
           │   (Go/Gin)      │  │    WORKER       │  │    DISPATCHER  │
           │                  │  │  (Go + RabbitMQ)│  │  (Go + RabbitMQ)│
           │  - Auth          │  │                  │  │                │
           │  - Users         │  │  - Email jobs    │  │  - HTTP POST    │
           │  - Circles       │  │  - SMS jobs      │  │  - HMAC sig    │
           │  - Contributions │  │  - Push jobs     │  │  - Retry logic │
           │  - Payouts       │  │  - In-app (WS)   │  │                │
           │  - Admin         │  │                  │  │                │
           └────────┬─────────┘  └──────────────────┘  └────────────────┘
                    │
           ┌────────▼─────────┐
           │  REDIS (cache +   │
           │  rate limit +     │
           │  session + queue) │
           └──────────────────┘
                    │
           ┌────────▼─────────┐
           │  WEBSOCKET SERVER │
           │  (Go/gorilla)     │
           │  Real-time push:   │
           │  circles, payouts, │
           │  notifications     │
           └──────────────────┘

External:
  Cloudflare → Nginx → API Server (port 1100)
  Frontend (port 1110) → CORS → API Server
```

### Service Topology

| Service | Port | Process | Responsibility |
|---|---|---|---|
| API Server | 1100 | Single binary, horizontally scaled | All REST endpoints, JWT auth |
| Indexer | — | Background worker | Sync Stellar → PostgreSQL |
| Notification Worker | — | RabbitMQ consumer | Email, SMS, push delivery |
| Webhook Dispatcher | — | RabbitMQ consumer | Webhook HTTP delivery |
| WebSocket Server | 1100/ws | Same binary or sidecar | Real-time event push |
| Redis | 6379 | External | Caching, rate limiting, sessions, job queue |
| PostgreSQL | 5432 | External | Primary data store |

---

## 2. Project Structure

```
moistello-backend/
├── go.mod
├── go.sum
├── Makefile
├── Dockerfile
├── docker-compose.yml              # Local dev: PG + Redis + RabbitMQ
├── .env.example
├── config/
│   ├── config.go                   # Viper-based config loader
│   ├── config.yaml                 # Default configuration
│   └── config.production.yaml      # Production overrides
├── cmd/
│   ├── api-server/
│   │   └── main.go                 # API server entry point
│   ├── indexer/
│   │   └── main.go                 # Indexer entry point
│   ├── notification-worker/
│   │   └── main.go                 # Notification consumer
│   ├── webhook-dispatcher/
│   │   └── main.go                 # Webhook delivery worker
│   └── migrate/
│       └── main.go                 # Database migration runner
├── internal/
│   ├── domain/                     # Domain models & business logic
│   │   ├── user/
│   │   │   ├── model.go
│   │   │   ├── repository.go       # Interface
│   │   │   ├── repository_pg.go    # PostgreSQL implementation
│   │   │   ├── service.go          # Business logic
│   │   │   ├── errors.go
│   │   │   └── service_test.go
│   │   ├── circle/
│   │   │   ├── model.go
│   │   │   ├── repository.go
│   │   │   ├── repository_pg.go
│   │   │   ├── service.go
│   │   │   ├── payout.go           # Random, fixed, auction, vote logic
│   │   │   ├── penalties.go        # Late fees, strikes, slashing
│   │   │   ├── errors.go
│   │   │   └── service_test.go
│   │   ├── contribution/
│   │   │   ├── model.go
│   │   │   ├── repository.go
│   │   │   ├── repository_pg.go
│   │   │   └── service.go
│   │   ├── payout/
│   │   │   ├── model.go
│   │   │   ├── repository.go
│   │   │   ├── repository_pg.go
│   │   │   └── service.go
│   │   ├── reputation/
│   │   │   ├── model.go
│   │   │   ├── repository.go
│   │   │   ├── repository_pg.go
│   │   │   ├── service.go          # MoiScore computation
│   │   │   └── scoring.go          # Score algorithm
│   │   ├── notification/
│   │   │   ├── model.go
│   │   │   ├── repository.go
│   │   │   ├── repository_pg.go
│   │   │   ├── service.go
│   │   │   ├── email.go            # SMTP/SendGrid
│   │   │   ├── sms.go              # Twilio/Africa's Talking
│   │   │   ├── push.go             # FCM / APNs
│   │   │   └── inapp.go            # WebSocket push
│   │   ├── auth/
│   │   │   ├── model.go            # Nonce, Session, JWT claims
│   │   │   ├── service.go          # Nonce generation, signature verification
│   │   │   ├── jwt.go              # Token creation/validation
│   │   │   ├── wallet.go           # Ed25519 signature verification
│   │   │   └── service_test.go
│   │   ├── invite/
│   │   │   ├── model.go
│   │   │   ├── repository.go
│   │   │   ├── repository_pg.go
│   │   │   └── service.go
│   │   ├── audit/
│   │   │   ├── model.go
│   │   │   ├── repository.go
│   │   │   └── repository_pg.go
│   │   └── webhook/
│   │       ├── model.go
│   │       ├── repository.go
│   │       ├── repository_pg.go
│   │       ├── service.go
│   │       └── dispatcher.go
│   ├── api/                        # HTTP layer
│   │   ├── server.go               # Gin server setup
│   │   ├── router.go               # Route registration
│   │   ├── middleware/
│   │   │   ├── auth.go             # JWT extraction + validation
│   │   │   ├── cors.go
│   │   │   ├── ratelimit.go        # Redis token bucket
│   │   │   ├── logging.go          # Request ID + structured logging
│   │   │   ├── recovery.go         # Panic recovery
│   │   │   ├── csrf.go
│   │   │   └── requestid.go        # X-Request-ID header
│   │   └── handler/
│   │       ├── auth_handler.go
│   │       ├── user_handler.go
│   │       ├── circle_handler.go
│   │       ├── contribution_handler.go
│   │       ├── payout_handler.go
│   │       ├── invite_handler.go
│   │       ├── notification_handler.go
│   │       ├── admin_handler.go
│   │       ├── webhook_handler.go
│   │       └── health_handler.go
│   ├── indexer/
│   │   ├── engine.go               # Main event loop
│   │   ├── horizon.go              # Horizon API client
│   │   ├── soroban.go              # Soroban RPC client
│   │   ├── event_processor.go       # Map events → DB writes
│   │   ├── cursor.go               # Last processed ledger tracker
│   │   └── reconciler.go           # Gap detection + repair
│   ├── websocket/
│   │   ├── hub.go                  # Connection manager
│   │   ├── client.go               # Individual WS client
│   │   ├── message.go              # Message types
│   │   └── auth.go                 # WS auth (token validation)
│   ├── pkg/                        # Shared utilities
│   │   ├── stellar/
│   │   │   ├── client.go           # Horizon + RPC wrapper
│   │   │   ├── signer.go           # Transaction signing
│   │   │   └── verify.go           # Ed25519 verification
│   │   ├── postgres/
│   │   │   ├── postgres.go         # Connection pool
│   │   │   └── transaction.go      # Tx helper
│   │   ├── redis/
│   │   │   └── redis.go            # Redis client wrapper
│   │   ├── rabbitmq/
│   │   │   └── rabbitmq.go         # Publisher + consumer
│   │   ├── logger/
│   │   │   └── logger.go           # Structured logging (zerolog)
│   │   ├── validator/
│   │   │   └── validator.go        # go-playground/validator setup
│   │   ├── metrics/
│   │   │   └── metrics.go          # Prometheus metrics
│   │   └── pagination/
│   │       └── pagination.go       # Offset/limit → meta
│   └── database/
│       └── migrations/             # golang-migrate SQL files
│           ├── 001_create_users.up.sql
│           ├── 001_create_users.down.sql
│           ├── 002_create_circles.up.sql
│           ├── 002_create_circles.down.sql
│           ├── 003_create_circle_members.up.sql
│           ├── 004_create_contributions.up.sql
│           ├── 005_create_payouts.up.sql
│           ├── 006_create_penalties.up.sql
│           ├── 007_create_invites.up.sql
│           ├── 008_create_notifications.up.sql
│           ├── 009_create_audit_log.up.sql
│           ├── 010_create_webhooks.up.sql
│           ├── 011_create_api_keys.up.sql
│           ├── 012_create_sessions.up.sql
│           ├── 013_create_indexer_cursor.up.sql
│           ├── 014_create_reputation_snapshots.up.sql
│           ├── 015_create_feature_flags.up.sql
│           └── ...down.sql files
├── pkg/                            # Exported packages
│   ├── dto/
│   │   ├── auth_dto.go
│   │   ├── user_dto.go
│   │   ├── circle_dto.go
│   │   ├── contribution_dto.go
│   │   └── pagination.go
│   ├── apperrors/
│   │   └── errors.go               # Domain error types
│   └── response/
│       └── response.go             # Standard API response wrapper
├── tests/
│   ├── integration/
│   │   ├── auth_test.go
│   │   ├── circle_test.go
│   │   ├── contribution_test.go
│   │   └── helpers/
│   │       ├── testutil.go         # Test setup/teardown
│   │       └── fixtures.go         # Test data
│   ├── e2e/
│   │   ├── circle_lifecycle_test.go
│   │   └── auth_flow_test.go
│   └── contract/
│       └── (tests as submodule)
├── contracts/                      # Git submodule → smart contracts repo
│   └── (Soroban contracts in Rust)
├── deployments/
│   ├── kubernetes/
│   │   ├── namespace.yaml
│   │   ├── api-server/
│   │   │   ├── deployment.yaml
│   │   │   ├── service.yaml
│   │   │   ├── ingress.yaml
│   │   │   └── hpa.yaml
│   │   ├── indexer/
│   │   │   └── deployment.yaml
│   │   ├── notification-worker/
│   │   │   └── deployment.yaml
│   │   └── postgres/
│   │       ├── statefulset.yaml
│   │       └── backup-cronjob.yaml
│   ├── terraform/                  # AWS/GCP infra
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   ├── rds.tf
│   │   ├── elasticache.tf
│   │   └── eks.tf
│   └── prometheus/
│       ├── prometheus.yml
│       ├── alert-rules.yml
│       └── grafana-dashboards/
│           ├── api-overview.json
│           ├── business-metrics.json
│           └── contracts.json
├── scripts/
│   ├── seed.go                     # Dev seed data
│   ├── migrate.sh                  # Migration runner
│   └── deploy.sh                   # Deployment script
└── docs/
    ├── api.md                      # API documentation
    └── architecture.md             # Architecture decisions
```

**Approximate file count:** ~120 Go files, ~30 SQL migrations, ~20 config/misc files

---

## 3. Database Schema (PostgreSQL)

### 3.1 Complete ERD

```
users ──1:N── circle_members ──N:1── circles
  │                  │                   │
  │                  │                   ├──1:N── contributions
  │                  │                   ├──1:N── payouts
  │                  │                   ├──1:N── penalties
  │                  │                   └──1:N── invites
  │                  │
  ├──1:N── contributions (direct)
  ├──1:N── payouts (direct)
  ├──1:N── penalties (direct)
  ├──1:N── notifications
  ├──1:N── webhooks
  ├──1:N── api_keys
  ├──1:N── sessions
  └──1:N── audit_log (actor)
```

### 3.2 Migration Sequence

| Order | Migration | Tables Created |
|---|---|---|
| 001 | `create_users` | `users` — wallet_address (UNIQUE), email, phone, display_name, avatar_ipfs_hash, kyc_status, country_code, preferred_language, moi_score, role |
| 002 | `create_circles` | `circles` — contract_id (UNIQUE), name, description, circle_type, payout_type, contribution_amount (NUMERIC), currency, frequency, max_members, min_moi_score, collateral_percent, late_fee_percent, grace_period_hours, max_strikes, start_date, end_date, status, current_round, total_contributions, organizer_id (FK→users) |
| 003 | `create_circle_members` | `circle_members` — circle_id (FK), user_id (FK), position, status, joined_at; UNIQUE(circle_id, user_id) |
| 004 | `create_contributions` | `contributions` — circle_id (FK), user_id (FK), round_number, amount, txn_hash, status, on_time; UNIQUE(circle_id, user_id, round_number) |
| 005 | `create_payouts` | `payouts` — circle_id (FK), recipient_id (FK→users), round_number, amount, fee_amount, txn_hash, payout_type |
| 006 | `create_penalties` | `penalties` — circle_id (FK), user_id (FK), round_number, penalty_type, amount, strikes_applied, reason |
| 007 | `create_invites` | `invites` — circle_id (FK), code (UNIQUE), created_by (FK→users), max_uses, use_count, expires_at |
| 008 | `create_notifications` | `notifications` — user_id (FK), type, title, body, data (JSONB), is_read, channel, sent_at |
| 009 | `create_audit_log` | `audit_log` — actor_id (FK→users, nullable), action, resource_type, resource_id, details (JSONB), ip_address (INET), user_agent |
| 010 | `create_webhooks` | `webhooks` — user_id (FK), url, events (TEXT[]), secret_hash, is_active, last_delivery_at, failure_count |
| 011 | `create_api_keys` | `api_keys` — user_id (FK), name, key_hash (UNIQUE), scopes (TEXT[]), rate_limit, expires_at, last_used_at |
| 012 | `create_sessions` | `sessions` — user_id (FK), token_hash (UNIQUE), expires_at, created_at |
| 013 | `create_indexer_cursor` | `indexer_cursor` — chain, last_ledger, last_processed_at |
| 014 | `create_reputation_snapshots` | `reputation_snapshots` — user_id (FK), score, level, breakdown (JSONB), month (DATE); UNIQUE(user_id, month) |
| 015 | `create_feature_flags` | `feature_flags` — flag (UNIQUE), enabled, description, updated_at |

### 3.3 Index Strategy

```sql
-- users
CREATE INDEX idx_users_wallet ON users(wallet_address);
CREATE INDEX idx_users_moi_score ON users(moi_score DESC);
CREATE INDEX idx_users_kyc_status ON users(kyc_status);

-- circles
CREATE INDEX idx_circles_status ON circles(status);
CREATE INDEX idx_circles_type ON circles(circle_type);
CREATE INDEX idx_circles_organizer ON circles(organizer_id);
CREATE INDEX idx_circles_currency ON circles(currency);
CREATE FULLTEXT INDEX idx_circles_search ON circles USING GIN(to_tsvector('english', name || ' ' || COALESCE(description, '')));

-- circle_members
CREATE INDEX idx_cm_circle ON circle_members(circle_id);
CREATE INDEX idx_cm_user ON circle_members(user_id);
CREATE INDEX idx_cm_status ON circle_members(status);

-- contributions
CREATE INDEX idx_contrib_circle ON contributions(circle_id);
CREATE INDEX idx_contrib_user ON contributions(user_id);
CREATE INDEX idx_contrib_round ON contributions(circle_id, round_number);
CREATE INDEX idx_contrib_txn ON contributions(txn_hash);

-- payouts
CREATE INDEX idx_payouts_circle ON payouts(circle_id);
CREATE INDEX idx_payouts_recipient ON payouts(recipient_id);

-- notifications
CREATE INDEX idx_notifs_user ON notifications(user_id, is_read);
CREATE INDEX idx_notifs_created ON notifications(created_at DESC);

-- audit_log
CREATE INDEX idx_audit_resource ON audit_log(resource_type, resource_id);
CREATE INDEX idx_audit_actor ON audit_log(actor_id);
CREATE INDEX idx_audit_created ON audit_log(created_at DESC);

-- sessions
CREATE INDEX idx_sessions_token ON sessions(token_hash);
CREATE INDEX idx_sessions_expires ON sessions(expires_at);
```

---

## 4. API Route Implementation

### 4.1 Response Format (Standard)

```json
{
  "success": true,
  "data": { ... },
  "error": "string (only if success=false)",
  "meta": {
    "page": 1,
    "limit": 20,
    "total": 150,
    "totalPages": 8
  }
}
```

### 4.2 Auth Flow (Detailed)

```
1. Frontend → POST /v1/auth/nonce { walletAddress: "GABC..." }
2. Backend → Redis SET nonce:GABC... = "random-64-hex" EX 300
3. Backend → { nonce: "random-64-hex" }

4. Frontend → Freighter signs nonce → signature (Ed25519)
5. Frontend → POST /v1/auth/verify { walletAddress, signature }

6. Backend → Redis GET nonce:GABC... → verify exists
7. Backend → Ed25519 Verify(signature, nonce, publicKey=walletAddress)
8. Backend → IF first login: INSERT INTO users (wallet_address)
9. Backend → Generate JWT (RS256):
   {
     sub: userId,
     wallet: walletAddress,
     role: user.role,
     iat: now,
     exp: now + 15min
   }
10. Backend → Generate refresh token (random 64-byte hex):
    INSERT INTO sessions (user_id, token_hash, expires_at = now + 7d)
11. Backend → { token, refreshToken, user }
```

**Token refresh:** `POST /v1/auth/refresh { refreshToken }` → rotate both tokens, invalidate old session.

**JWT middleware:** Every request extracts `Authorization: Bearer <token>`, validates JWT, injects `User` into `gin.Context`.

### 4.3 Full Endpoint Implementation Order

**Phase 1 — Foundation (Week 1-2)**
- [ ] Project scaffold: Go modules, Gin skeleton, Docker Compose
- [ ] Config system (Viper) + structured logging (zerolog)
- [ ] PostgreSQL connection pool + migration runner
- [ ] Redis connection
- [ ] Migration 001-012 (all tables)
- [ ] `POST /auth/nonce` — generate nonce, store in Redis
- [ ] `POST /auth/verify` — verify Ed25519 sig, create user, issue JWT
- [ ] `POST /auth/register` — verify + store profile fields
- [ ] `POST /auth/refresh` — rotate tokens
- [ ] `POST /auth/me` — get current user from JWT
- [ ] `POST /auth/logout` — invalidate session
- [ ] JWT middleware
- [ ] Rate limiting middleware
- [ ] `GET /health`, `GET /metrics`

**Phase 2 — Users & Circles (Week 3-4)**
- [ ] `GET /users/me` — profile
- [ ] `PATCH /users/me` — update profile
- [ ] `GET /users/:id` — public profile
- [ ] `GET /circles` — list with search, filters, pagination
- [ ] `POST /circles` — create circle (validate payload)
- [ ] `GET /circles/:id` — circle detail
- [ ] `PATCH /circles/:id` — organizer update
- [ ] `DELETE /circles/:id` — organizer cancel (pending only)
- [ ] `POST /circles/:id/join` — join circle
- [ ] `GET /circles/:id/members` — member list
- [ ] `POST /circles/:id/exit` — exit with penalty calc

**Phase 3 — Contributions & Payouts (Week 5-6)**
- [ ] `POST /circles/:id/contribute` — record contribution (verify txn hash on Stellar)
- [ ] `GET /circles/:id/rounds` — round history
- [ ] `GET /circles/:id/payouts` — payout history
- [ ] `GET /contributions` — user contribution history
- [ ] `GET /contributions/:id` — contribution detail
- [ ] `GET /payouts` — user payout history
- [ ] `GET /payouts/:id` — payout detail
- [ ] `POST /circles/:id/vote` — vote-based payout
- [ ] `POST /circles/:id/auction-bid` — auction payout
- [ ] Payout calculator (random VRF, fixed order, auction, vote)

**Phase 4 — Reputation & Social (Week 7-8)**
- [ ] `GET /users/me/reputation` — MoiScore data
- [ ] Reputation scoring engine (CRON or event-driven):
  - Contribution streaks → streaks score
  - Circle completions → completions score
  - Total volume → volume score
  - Recent activity → recency score
  - Weighted combination → MoiScore (0-1000)
- [ ] `GET /users/me/circles` — user's active circles list
- [ ] `POST /users/me/kyc` — KYC initiation (Sumsub API)
- [ ] `GET /users/me/kyc/status` — KYC status
- [ ] `POST /circles/:id/dispute` — dispute creation

**Phase 5 — Invites & Notifications (Week 9-10)**
- [ ] `POST /circles/:id/invites` — generate invite
- [ ] `GET /circles/:id/invites` — list invites
- [ ] `DELETE /invites/:code` — revoke invite
- [ ] `GET /notifications` — list with filters
- [ ] `PATCH /notifications/:id/read` — mark read
- [ ] `PATCH /notifications/read-all` — mark all read
- [ ] `PUT /notifications/preferences` — channel prefs
- [ ] Notification creation triggers (event hooks in services):
  - Circle created → notify organizer
  - Member joined → notify organizer
  - Contribution due → notify member
  - Contribution late → notify member + organizer
  - Payout received → notify recipient
  - Circle completed → notify all members

**Phase 6 — Admin & Enterprise (Week 11-12)**
- [ ] `GET /admin/users` — user management
- [ ] `GET /admin/circles` — circle management
- [ ] `GET /admin/audit-log` — activity trail
- [ ] `GET /admin/metrics` — platform KPI dashboard data
- [ ] `POST /admin/feature-flags` — feature toggle
- [ ] `POST /webhooks/register` — register webhook
- [ ] `GET /webhooks` — list webhooks
- [ ] `DELETE /webhooks/:id` — remove webhook
- [ ] Webhook dispatcher (RabbitMQ consumer)
- [ ] API key management
- [ ] Full API documentation (OpenAPI/Swagger via swaggo)

---

## 5. Smart Contract Integration Strategy

### 5.1 Contract Repository (Separate)

```
moistello-contracts/      ← Separate repo, Rust/Soroban
├── Cargo.toml
├── packages/
│   ├── circle-factory/
│   ├── circle/
│   ├── reputation-registry/
│   ├── governance-token/
│   └── treasury/
└── tests/
```

Referenced as a Git submodule in the backend: `contracts/` → `github.com/moistello/contracts`

### 5.2 Contract ↔ Backend Communication

**Pattern:** Off-chain indexer reads on-chain events, writes to PostgreSQL. API server reads from PostgreSQL (fast), writes to both PostgreSQL and Stellar.

```
WRITE PATH:
  Client → API Server → PostgreSQL (record) → Stellar (transaction)
                                ↓
                          Indexer picks up
                          Stellar event → reconcile

READ PATH:
  Client → API Server → PostgreSQL (always, sub-ms)
```

### 5.3 Event Processing (Indexer)

```
Every 3 seconds:
  1. Read cursor.last_ledger from DB
  2. Horizon: GET /ledgers?cursor=<cursor>&limit=50
  3. For each ledger:
     a. GET /ledgers/:seq/transactions
     b. Filter: transactions containing our contract IDs
     c. For each matching tx:
        - Decode Soroban event topics + data
        - Map to domain event (CircleCreated, ContributionReceived, etc.)
        - Write to PostgreSQL (idempotent via txn_hash)
  4. Update cursor.last_ledger
  5. Emit WebSocket events for real-time updates
```

### 5.4 Transaction Signing

Backend signs transactions using a managed keypair (via AWS KMS or a hot wallet for testnet):

```
1. API Server constructs Soroban transaction
2. Signs with backend keypair
3. Submits to Stellar RPC
4. Waits for confirmation (up to 10s timeout)
5. Returns txn_hash to client
6. Indexer eventually confirms the event
```

---

## 6. Notification System

### 6.1 Architecture

```
Service Layer (any domain event)
  │
  ▼
notification.Service.CreateNotification(notification)
  │
  ├──► PostgreSQL (INSERT INTO notifications)
  │
  ├──► RabbitMQ (publish to "moistello.notifications")
  │     │
  │     ├──► Email Worker → SendGrid/Twilio
  │     ├──► SMS Worker → Africa's Talking/Twilio
  │     ├──► Push Worker → FCM/APNs
  │     └──► In-App Worker → WebSocket Hub
  │
  └──► WebSocket Hub (real-time push to online users)
```

### 6.2 Event Triggers

| Domain Event | Notification | Channel |
|---|---|---|
| `CircleCreated` | "Circle '{name}' created" | In-app |
| `MemberJoined` | "{user} joined '{circle}'" | In-app + email |
| `ContributionDue` | "Contribution due in {n} hours" | In-app + email + SMS |
| `ContributionReceived` | "Contribution confirmed" | In-app |
| `ContributionLate` | "Payment is late for '{circle}'" | In-app + email |
| `PayoutReceived` | "You received {amount} from '{circle}'" | In-app + email + push |
| `CircleCompleted` | "Circle '{name}' completed!" | In-app + email |
| `MemberExited` | "{user} left '{circle}'" | In-app |
| `DisputeRaised` | "Dispute raised in '{circle}'" | In-app + email |

---

## 7. WebSocket Server

### 7.1 Protocol

```
Client connects: ws://localhost:1100/ws?token=<jwt>

Server validates JWT → adds client to hub

Messages (JSON):

→ Server → Client:
  { type: "circle.updated", payload: { circleId, changes } }
  { type: "contribution.confirmed", payload: { circleId, amount, round } }
  { type: "payout.received", payload: { circleId, amount } }
  { type: "notification.new", payload: { notification } }
  { type: "member.joined", payload: { circleId, member } }

→ Client → Server (heartbeat):
  { type: "ping" }

→ Server → Client (heartbeat):
  { type: "pong" }
```

### 7.2 Hub Design

```go
type Hub struct {
    clients    map[string]*Client           // key: userId
    rooms      map[string]map[string]*Client // key: circleId, sub: userId
    broadcast  chan Message
    register   chan *Client
    unregister chan *Client
    mu         sync.RWMutex
}
```

Each client subscribes to:
- Their personal channel (user-{userId})
- Each circle they're a member of (circle-{circleId})

---

## 8. Reputation Engine (MoiScore)

### 8.1 Scoring Algorithm

```go
type ScoreBreakdown struct {
    Streaks      float64 // 0-350  (35%)
    Completions  float64 // 0-300  (30%)
    Volume       float64 // 0-200  (20%)
    Recency      float64 // 0-150  (15%)
    Total        float64 // 0-1000
    Level        string  // Bronze, Silver, Gold, Platinum, Diamond
}

func CalculateMoiScore(userId string) ScoreBreakdown {
    // Streaks: consecutive on-time contributions
    //   score = min(consecutiveStreaks * 35, 350)
    
    // Completions: circles completed
    //   score = min(completedCircles * 50, 300)
    
    // Volume: total amount contributed across all circles
    //   score = min(log(totalVolumeUSD) * 30, 200)
    
    // Recency: days since last contribution
    //   score = max(0, 150 - daysSinceLast * 5)
    
    total := streaks + completions + volume + recency
    
    level := "Bronze"   // 0-200
    if total > 800 { level = "Diamond" }
    else if total > 600 { level = "Platinum" }
    else if total > 400 { level = "Gold" }
    else if total > 200 { level = "Silver" }
    
    return ScoreBreakdown{...}
}
```

### 8.2 Score Updates

- **Real-time:** On every contribution recorded, recalculate immediately
- **Batch:** CRON job runs nightly to recalculate all active users
- **History:** Monthly snapshots stored in `reputation_snapshots` for trend charts

---

## 9. Security Architecture

### 9.1 Authentication

| Layer | Method |
|---|---|
| Transport | TLS 1.3 (Cloudflare → Nginx → Go) |
| Auth | Ed25519 wallet signature verification |
| Token | JWT (RS256), 15min access + 7d refresh |
| Sessions | Stored in PostgreSQL + Redis cache |
| CSRF | Double-submit cookie for state-changing ops |

### 9.2 Authorization

| Resource | Permission |
|---|---|
| `/users/me` | Authenticated |
| `/users/:id` | Public (limited fields) |
| `/circles` GET | Public (filters applied) |
| `/circles` POST | Authenticated |
| `/circles/:id` PATCH/DELETE | Organizer only |
| `/circles/:id/join` | Authenticated (MoiScore check for public circles) |
| `/admin/*` | Role == "admin" |

### 9.3 Input Validation

```go
type Validator struct {
    validate *validator.Validate
}

func (v *Validator) ValidateCreateCircle(req CreateCirclePayload) []FieldError {
    // name: required, 3-100 chars, alphanumeric + spaces + hyphens
    // contributionAmount: required, > 0
    // currency: required, oneof=USDC XLM
    // frequency: required, oneof=daily weekly biweekly monthly
    // maxMembers: required, 2-100
    // etc.
}
```

### 9.4 Rate Limiting

```go
// Redis-based token bucket
// Key: ratelimit:<ip_or_userId>
// Default: 100 req/min per IP
// Authenticated: 300 req/min per user
// Auth endpoints: 10 req/min per IP (brute force protection)
```

### 9.5 SQL Injection Prevention

- All queries use parameterized statements (`$1`, `$2` via `database/sql`)
- No dynamic SQL concatenation
- Input sanitization at validation layer

---

## 10. Testing Strategy

### 10.1 Levels

| Level | Tool | Coverage Target | Scope |
|---|---|---|---|
| Unit | `go test` | 90% | Individual functions, algorithms |
| Integration | `go test` + Testcontainers for PG + Redis | 80% | Repository + Service layers |
| API | `go test` + httptest | 85% | All endpoint handlers |
| E2E | Custom Go test suites | Happy paths | Circle lifecycle, Auth flow |
| Contract | `cargo test` (in contract repo) | 95% | All contract functions |
| Load | k6 | 1000 req/s sustained | API under load |

### 10.2 Test Database

Use `testcontainers-go` to spin up real PostgreSQL and Redis in Docker for integration tests. Each test suite gets isolated containers, auto-destroyed after.

```go
func TestMain(m *testing.M) {
    pgContainer, _ := testcontainers.GenericContainer(...)
    redisContainer, _ := testcontainers.GenericContainer(...)
    defer pgContainer.Terminate()
    defer redisContainer.Terminate()
    
    db = connectToPostgres(pgContainer)
    redis = connectToRedis(redisContainer)
    
    runMigrations(db)
    os.Exit(m.Run())
}
```

---

## 11. DevOps & Infrastructure

### 11.1 Docker Compose (Local Dev)

```yaml
services:
  postgres:
    image: postgres:16
    ports: ["5432:5432"]
    environment:
      POSTGRES_DB: moistello
      POSTGRES_USER: moistello
      POSTGRES_PASSWORD: moistello_dev
  
  redis:
    image: redis:7-alpine
    ports: ["6379:6379"]
  
  rabbitmq:
    image: rabbitmq:3-management-alpine
    ports: ["5672:5672", "15672:15672"]
  
  api-server:
    build: .
    ports: ["1100:1100"]
    depends_on: [postgres, redis, rabbitmq]
    environment:
      DATABASE_URL: postgres://moistello:moistello_dev@postgres:5432/moistello?sslmode=disable
      REDIS_URL: redis://redis:6379
      RABBITMQ_URL: amqp://rabbitmq:5672
      STELLAR_NETWORK: testnet
  
  indexer:
    build:
      dockerfile: Dockerfile.indexer
    depends_on: [postgres, redis]
  
  notification-worker:
    build:
      dockerfile: Dockerfile.worker
    depends_on: [postgres, rabbitmq, redis]
```

### 11.2 Production Deployment (Kubernetes)

```
moistello-prod/
├── namespace: moistello
├── secrets: db-url, redis-url, jwt-private-key, sendgrid-key
├── api-server: 3 replicas, HPA (min 3, max 10, CPU 70%)
├── indexer: 1 replica (singleton — leader election via Redis)
├── notification-worker: 2 replicas
├── webhook-dispatcher: 2 replicas
├── postgres: StatefulSet, 100GB PVC, daily backups to S3
├── redis: 2 replicas + sentinel
├── rabbitmq: 3-node cluster
└── ingress: Nginx + cert-manager for TLS

Monitoring:
├── Prometheus: scrapes /metrics on all services
├── Grafana: dashboards (API, business, contracts)
├── Loki: log aggregation
└── Alertmanager: PagerDuty integration
```

### 11.3 CI/CD (GitHub Actions)

```yaml
name: Backend CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.22' }
      - run: golangci-lint run ./...

  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: go test ./... -race -coverprofile=coverage.out
      - run: go tool cover -func=coverage.out

  build:
    runs-on: ubuntu-latest
    needs: [lint, test]
    steps:
      - run: go build -o moistello-api ./cmd/api-server
      - run: docker build -t moistello-api:latest .

  deploy-staging:
    if: github.ref == 'refs/heads/main'
    needs: build
    steps:
      - run: kubectl apply -f deployments/kubernetes/ --context staging

  deploy-prod:
    if: startsWith(github.ref, 'refs/tags/v')
    needs: build
    steps:
      - run: kubectl apply -f deployments/kubernetes/ --context production
```

---

## 12. Implementation Phases (14 Weeks)

### Phase 1: Foundation — Week 1-2
```
Goal: API server running, auth working, DB schema up
Deliverables:
  - Go project scaffolded
  - Docker Compose for local dev
  - All 15 migrations applied
  - Auth endpoints (nonce, verify, refresh, me, logout)
  - JWT middleware deployed
  - Rate limiting active
  - Health check + Prometheus metrics
```

### Phase 2: Users & Circles — Week 3-5
```
Goal: Full circle lifecycle via API
Deliverables:
  - User profile CRUD
  - Circle CRUD
  - Join/exit circles
  - Member management
  - Search + filtering + pagination
  - Input validation (go-playground/validator)
  - Full test suite (unit + integration)
```

### Phase 3: Contributions & Payments — Week 6-8
```
Goal: Money moving through system
Deliverables:
  - Contribution recording (with Stellar txn verification)
  - Payout calculation (random VRF, fixed, auction, vote)
  - Contribution/payout history APIs
  - Penalty engine (late fees, strikes, slashing)
  - Dispute mechanism
  - Stellar Horizon + RPC integration
```

### Phase 4: Indexer — Week 9-10
```
Goal: On-chain events synced to PostgreSQL
Deliverables:
  - Indexer engine (cursor tracking)
  - Event processor (all contract events → DB)
  - Reconciler (gap detection)
  - Idempotent writes
  - Smart contract deployment to testnet
```

### Phase 5: Real-Time & Notifications — Week 11-12
```
Goal: Real-time updates + multi-channel notifications
Deliverables:
  - WebSocket server (authenticated)
  - Notification service (creation + delivery)
  - Email worker (SendGrid)
  - SMS worker (Africa's Talking / Twilio)
  - Push notification worker (FCM)
  - Webhook registration + delivery
```

### Phase 6: Reputation & Enterprise — Week 13-14
```
Goal: MoiScore + admin + production readiness
Deliverables:
  - Reputation scoring engine
  - Monthly snapshot CRON
  - Admin endpoints (users, circles, audit, metrics)
  - Feature flag system
  - API key management
  - API documentation (OpenAPI/Swagger)
  - Load testing (k6)
  - Production deployment
```

---

## 13. Go Module Dependencies

```go
require (
    // Framework
    github.com/gin-gonic/gin           v1.10.x      // HTTP router
    github.com/gin-contrib/cors        v1.7.x       // CORS middleware
    
    // Database
    github.com/lib/pq                  v1.10.x      // PostgreSQL driver
    github.com/jmoiron/sqlx            v1.4.x       // SQL extensions
    github.com/golang-migrate/migrate  v4.x         // Migrations
    
    // Cache & Queue
    github.com/redis/go-redis/v9       v9.x         // Redis client
    github.com/rabbitmq/amqp091-go     v1.x         // RabbitMQ client
    
    // Auth
    github.com/golang-jwt/jwt/v5       v5.x         // JWT
    github.com/stellar/go              (latest)     // Stellar SDK
    
    // Utils
    github.com/spf13/viper             v1.x         // Config
    github.com/rs/zerolog              v1.x         // Logging
    github.com/go-playground/validator v10.x        // Validation
    github.com/google/uuid             v1.x         // UUID generation
    
    // WebSocket
    github.com/gorilla/websocket       v1.x         // WebSocket
    
    // Monitoring
    github.com/prometheus/client_golang v1.x        // Prometheus metrics
    
    // Testing
    github.com/stretchr/testify        v1.x         // Assertions
    github.com/testcontainers/testcontainers-go v0.x // Test containers
    
    // Documentation
    github.com/swaggo/swag             v1.x         // Swagger gen
    github.com/swaggo/gin-swagger      v1.x         // Gin Swagger UI
)
```

---

## 14. Configuration File (config.yaml)

```yaml
server:
  port: 1100
  host: "0.0.0.0"
  readTimeout: 10s
  writeTimeout: 30s
  maxHeaderBytes: 1048576

database:
  url: "postgres://moistello:password@localhost:5432/moistello?sslmode=disable"
  maxOpenConns: 50
  maxIdleConns: 10
  connMaxLifetime: 30m
  migrationPath: "file://internal/database/migrations"

redis:
  url: "redis://localhost:6379"
  password: ""
  db: 0
  poolSize: 20

rabbitmq:
  url: "amqp://guest:guest@localhost:5672/"
  exchange: "moistello.events"
  queues:
    notifications: "moistello.notifications"
    webhooks: "moistello.webhooks"

stellar:
  network: "testnet"
  horizonUrl: "https://horizon-testnet.stellar.org"
  sorobanRpcUrl: "https://soroban-testnet.stellar.org"
  networkPassphrase: "Test SDF Network ; September 2015"
  masterPublicKey: ""  # Backend signing keypair

auth:
  jwtPrivateKeyPath: "/etc/moistello/jwt-private.pem"
  jwtPublicKeyPath: "/etc/moistello/jwt-public.pem"
  accessTokenTTL: 15m
  refreshTokenTTL: 168h    # 7 days
  nonceTTL: 5m

indexer:
  pollInterval: 3s
  batchSize: 50
  startLedger: 0  # 0 = from latest

notification:
  email:
    provider: "sendgrid"
    apiKey: ""
    fromAddress: "notifications@moistello.io"
  sms:
    provider: "twilio"
    accountSid: ""
    authToken: ""
    fromNumber: ""
  push:
    fcmServerKey: ""

cors:
  allowedOrigins:
    - "http://localhost:1110"
    - "https://app.moistello.io"

rateLimit:
  global: 100      # req/min per IP
  authenticated: 300  # req/min per user
  auth: 10         # req/min per IP on auth endpoints

logging:
  level: "debug"
  format: "json"    # json or console
  output: "stdout"
```

---

## 15. Key Architectural Decisions (ADR)

| # | Decision | Rationale |
|---|---|---|
| 1 | **Go over Node.js** | Single binary deploy, goroutines for indexer/WS concurrency, strong type safety for financial data |
| 2 | **Gin over Echo/Chi** | Most mature ecosystem, middleware chaining, best documentation, context injection |
| 3 | **PostgreSQL over MongoDB** | ACID for financial transactions, JSONB for flexible data, full-text search, CITEXT for wallet addresses |
| 4 | **Redis as cache + rate limiter** | In-memory speed, TTL support for nonces/sessions, atomic operations for rate limits |
| 5 | **RabbitMQ for async jobs** | Reliable delivery, retry logic, dead-letter queues for failed notifications |
| 6 | **Indexer as separate process** | Decouples blockchain sync from API availability. API never waits for chain |
| 7 | **Testcontainers for integration tests** | Real PostgreSQL/Redis in CI, no mocks, deterministic tests |
| 8 | **JWT in localStorage over httpOnly cookie** | Frontend already built this way. XSS protection via CSP headers |
| 9 | **golang-migrate for migrations** | Schema-as-code, rollbacks supported, works with any deployment strategy |
| 10 | **Viper for config** | 12-factor app compliance, env var overrides, YAML + env + flags |

---

## File Count Summary

| Category | Files | Lines (est.) |
|---|---|---|
| Domain services | 30 | 4,500 |
| API handlers | 12 | 2,400 |
| Middleware | 7 | 700 |
| Indexer | 6 | 1,200 |
| WebSocket | 4 | 600 |
| Workers (notification, webhook) | 6 | 900 |
| Config + CLI entry points | 5 | 400 |
| Database migrations | 30 | 1,500 |
| Tests | 20 | 3,000 |
| Infrastructure (k8s, terraform, docker) | 15 | 1,200 |
| Docs + scripts | 8 | 800 |
| **Total** | **~143** | **~17,200** |
