# Moistello — Stellar Savings Circles (ROSCA)

## Overview
A decentralized rotating savings and credit association platform built on Stellar/Soroban. Members form savings circles, contribute USDC or XLM each cycle, and one member receives the pooled payout per round. Smart contracts enforce rules, handle randomization/payouts, apply default penalties, and track on-chain reputation. The culturally-universal savings model (esusu, tontine, chit fund, tanda, hui) meets blockchain transparency.

---

## COMPLETE BACKEND REQUIREMENTS

### Port Configuration
- **Backend API**: Port **1100** (`http://localhost:1100`)
- **Frontend App**: Port **1110** (`http://localhost:1110`)
- **Frontend → Backend**: CORS configured, API calls to `http://localhost:1100/v1/*`

---

### Tech Stack (Recommended)

| Layer | Technology |
|---|---|
| Language | Go 1.22+ |
| Framework | Gin or Echo |
| Database | PostgreSQL 16 |
| Cache/Queue | Redis (sessions, rate limiting, job queue) |
| Message Queue | RabbitMQ or NATS (async jobs: emails, SMS, webhooks) |
| Blockchain | Stellar Horizon API + Soroban RPC |
| Auth | JWT (RS256), Freighter wallet signature verification |
| File Storage | Local filesystem for uploaded pages (content/pages/) |

---

### 1. AUTH API (`/v1/auth/*`)

| Method | Endpoint | Request | Response | Notes |
|---|---|---|---|---|
| `POST` | `/auth/nonce` | `{ walletAddress: string }` | `{ nonce: string }` | Generate random nonce (store in Redis, TTL 5min). User signs this with Freighter. |
| `POST` | `/auth/verify` | `{ walletAddress: string, signature: string }` | `{ token: string, refreshToken: string, user: User }` | Verify signature against stored nonce. Create user if first login. Generate JWT + refresh token. |
| `POST` | `/auth/register` | `{ walletAddress: string, signature: string, displayName?: string, email?: string, countryCode?: string, language?: string }` | `{ token: string, refreshToken: string, user: User }` | Same as verify + stores profile fields. |
| `POST` | `/auth/refresh` | `{ refreshToken: string }` | `{ token: string, refreshToken: string }` | Rotate tokens. Invalidate old refresh token. |
| `POST` | `/auth/me` | Header: `Authorization: Bearer <token>` | `{ user: User }` | Return current authenticated user from JWT. |
| `POST` | `/auth/logout` | Header: `Authorization: Bearer <token>` | `{ success: true }` | Invalidate current session. |

**JWT Payload:** `{ sub: userId, wallet: walletAddress, iat, exp }`
**Nonce Flow:** Backend generates random nonce → stores in Redis (key: `nonce:<walletAddress>`) with 5min TTL → user signs with Freighter → backend verifies Ed25519 signature against nonce → if valid, issue JWT.

---

### 2. USER API (`/v1/users/*`)

| Method | Endpoint | Request | Response | Notes |
|---|---|---|---|---|
| `GET` | `/users/me` | Auth header | `{ user: User }` | Full profile with KYC status, MoiScore |
| `PATCH` | `/users/me` | `{ displayName?, email?, phone?, countryCode?, preferredLanguage? }` | `{ user: User }` | Update profile |
| `GET` | `/users/:id` | Public | `{ user: PublicUser }` | Public profile (no email/phone) |
| `GET` | `/users/me/reputation` | Auth header | `{ score: number, level: string, breakdown: { streaks, completions, volume, recency }, history: MonthlyScore[] }` | MoiScore data |
| `POST` | `/users/me/kyc` | Auth header | `{ kycLink: string, status: "pending" }` | Initiate KYC via Sumsub |
| `GET` | `/users/me/kyc/status` | Auth header | `{ status: "unverified"\|"pending"\|"verified"\|"rejected", reason?: string }` | Current KYC status |
| `GET` | `/users/me/circles` | Auth header | `{ circles: Circle[] }` | All circles user is a member of |

---

### 3. CIRCLES API (`/v1/circles/*`)

| Method | Endpoint | Request | Response | Notes |
|---|---|---|---|---|
| `GET` | `/circles` | Query: `?search=&status=&type=&page=&limit=` | `{ circles: Circle[], meta: PaginationMeta }` | Discover/browse circles. Filters: public only unless authenticated. |
| `POST` | `/circles` | `CreateCirclePayload` | `{ circle: Circle }` | Create new circle. Validates with Zod schema. Creates on-chain representation. |
| `GET` | `/circles/:id` | Public/Auth | `{ circle: Circle }` | Full circle details. |
| `PATCH` | `/circles/:id` | `{ name?, description? }` | `{ circle: Circle }` | Organizer-only. |
| `DELETE` | `/circles/:id` | Auth header | `{ success: true }` | Organizer-only. Only if status is "pending". |
| `POST` | `/circles/:id/join` | `{ inviteCode? }` | `{ member: CircleMember }` | Join circle (checks invite for private circles, MoiScore for public). |
| `POST` | `/circles/:id/contribute` | `{ amount, txnHash, roundNumber? }` | `{ contribution: Contribution }` | Record on-chain contribution. Verifies txnHash on Stellar. |
| `POST` | `/circles/:id/exit` | Auth header | `{ success: true, penalty?: number }` | Early exit with penalty. |
| `GET` | `/circles/:id/members` | Auth header | `{ members: CircleMember[] }` | All members with reputation scores. |
| `GET` | `/circles/:id/rounds` | Auth header | `{ rounds: Round[] }` | Round history with contributions + payouts. |
| `GET` | `/circles/:id/payouts` | Auth header | `{ payouts: Payout[] }` | Payout history. |
| `POST` | `/circles/:id/dispute` | `{ reason, evidence? }` | `{ dispute: Dispute }` | Raise a dispute. Freezes circle until resolved. |
| `POST` | `/circles/:id/vote` | `{ voteFor: string }` | `{ success: true }` | Vote-based payout: cast vote for current round. |
| `POST` | `/circles/:id/auction-bid` | `{ discountBips: number }` | `{ bid: AuctionBid }` | Auction payout: submit bid for current round. |

**CreateCirclePayload Schema:**
```go
type CreateCirclePayload struct {
    Name              string  `json:"name" validate:"required,min=3,max=100"`
    Description       string  `json:"description" validate:"max=500"`
    CircleType        string  `json:"circleType" validate:"required,oneof=public private org community premium"`
    PayoutType        string  `json:"payoutType" validate:"required,oneof=random fixed auction vote"`
    ContributionAmount float64 `json:"contributionAmount" validate:"required,gt=0"`
    Currency          string  `json:"currency" validate:"required,oneof=USDC XLM"`
    Frequency         string  `json:"frequency" validate:"required,oneof=daily weekly biweekly monthly"`
    MaxMembers        int     `json:"maxMembers" validate:"required,min=2,max=100"`
    MinMoiScore       int     `json:"minMoiScore" validate:"min=0,max=1000"`
    CollateralPercent float64 `json:"collateralPercent" validate:"min=0,max=100"`
    LateFeePercent    float64 `json:"lateFeePercent" validate:"min=0,max=50"`
    GracePeriodHours  int     `json:"gracePeriodHours" validate:"min=1,max=168"`
    MaxStrikes        int     `json:"maxStrikes" validate:"min=1,max=10"`
    StartDate         string  `json:"startDate"` // ISO date
}
```

---

### 4. CONTRIBUTIONS API (`/v1/contributions/*`)

| Method | Endpoint | Request | Response | Notes |
|---|---|---|---|---|
| `GET` | `/contributions` | Query: `?circleId=&page=&limit=` | `{ contributions: Contribution[], meta: PaginationMeta }` | User's contribution history. |
| `GET` | `/contributions/:id` | Auth header | `{ contribution: Contribution }` | Single contribution detail + Stellar txn link. |

---

### 5. PAYOUTS API (`/v1/payouts/*`)

| Method | Endpoint | Request | Response | Notes |
|---|---|---|---|---|
| `GET` | `/payouts` | Query: `?circleId=&page=&limit=` | `{ payouts: Payout[], meta: PaginationMeta }` | User's payout history. |
| `GET` | `/payouts/:id` | Auth header | `{ payout: Payout }` | Single payout detail. |

---

### 6. INVITES API (`/v1/invites/*`)

| Method | Endpoint | Request | Response | Notes |
|---|---|---|---|---|
| `GET` | `/circles/:id/invites` | Auth header (organizer) | `{ invites: Invite[] }` | List active invites. |
| `POST` | `/circles/:id/invites` | `{ maxUses?, expiresAt? }` | `{ invite: Invite }` | Generate invite code. |
| `DELETE` | `/invites/:code` | Auth header | `{ success: true }` | Revoke invite. |

---

### 7. NOTIFICATIONS API (`/v1/notifications/*`)

| Method | Endpoint | Request | Response | Notes |
|---|---|---|---|---|
| `GET` | `/notifications` | Query: `?page=&limit=&unreadOnly=` | `{ notifications: Notification[], meta: PaginationMeta }` | User's notifications. |
| `PATCH` | `/notifications/:id/read` | Auth header | `{ notification: Notification }` | Mark as read. |
| `PATCH` | `/notifications/read-all` | Auth header | `{ success: true }` | Mark all as read. |
| `PUT` | `/notifications/preferences` | `{ email?, push?, inApp? }` | `{ preferences: NotificationPrefs }` | Update channel preferences. |

---

### 8. ADMIN API (`/v1/admin/*`) — Role-protected

| Method | Endpoint | Request | Response | Notes |
|---|---|---|---|---|
| `GET` | `/admin/users` | Query: `?search=&page=&limit=` | `{ users: User[], meta }` | User management. |
| `GET` | `/admin/circles` | Query: `?status=&page=&limit=` | `{ circles: Circle[], meta }` | Circle management. |
| `GET` | `/admin/audit-log` | Query: `?page=&limit=` | `{ logs: AuditLog[], meta }` | Activity audit trail (immutable, append-only). |
| `GET` | `/admin/metrics` | Auth header | `{ users, circles, contributions, payouts, activeUsers }` | Platform metrics. |
| `POST` | `/admin/feature-flags` | `{ flag: string, enabled: boolean }` | `{ flag, enabled }` | Manage feature flags. |

---

### 9. SMART CONTRACT INTEGRATION (Soroban / Rust)

| Contract | Responsibility |
|---|---|
| `CircleFactory` | Deploy new circle instances, track all circles |
| `Circle` | Core logic: join, contribute, payout (random/fixed/auction/vote), penalties, exits |
| `ReputationRegistry` | Record contributions/defaults, compute MoiScore (0-1000) |
| `GovernanceToken` | MOI token (CAP-46 standard) for protocol governance |
| `Treasury` | Protocol fee collection (0.5% per payout) |

**Key events emitted (for indexer to consume):**
- `CircleCreated`, `MemberJoined`, `ContributionReceived`, `PayoutExecuted`
- `LateReported`, `MemberExited`, `DefaultRecorded`, `CircleCompleted`
- `AuctionBid`, `VoteCast`, `DisputeRaised`

---

### 10. INDEXER SERVICE

A background worker that listens to Stellar ledger events and syncs on-chain data to PostgreSQL:

- **Source**: Stellar Horizon API (transaction history) + Soroban RPC (contract events)
- **Target**: PostgreSQL (normalized tables)
- **Idempotent**: Track last processed ledger cursor, handle reorgs
- **Reconciler**: Periodic scan to find missed events

**Synced data:** Circles, members, contributions, payouts, penalties, reputation scores.

---

### 11. WEBSOCKET SERVER

Real-time push for:
- `circle.updated` — Circle status changes, new rounds
- `contribution.confirmed` — On-chain contribution verified
- `payout.received` — User received payout
- `notification.new` — New notification for user

**Protocol**: Authenticated WebSocket (token in initial handshake). URL: `ws://localhost:1100/ws`

---

### 12. WEBHOOK SYSTEM

Third-party integrations subscribe to events:
- Register: `POST /webhooks/register { url, events[], secret }`
- Delivery: HTTP POST with HMAC signature header
- Retry: 3 retries with exponential backoff
- Dashboard: `GET /webhooks`, `DELETE /webhooks/:id`

---

### 13. ADMIN USER MANAGEMENT (Upload System)

Documented in the product-spec. Summary:

```bash
node scripts/create-user.js         # CLI on server
→ Generates token → User visits /setup?token=...
→ POST /api/auth/setup              # Frontend API route (in Next.js)
→ User sets username + password
→ POST /api/auth/login              # Session cookie, 7-day expiry
→ POST /api/upload                  # Upload .md/.html → /p/<slug>
```

Storage: `content/users.json`, `content/setup-tokens.json`, `content/sessions.json`, `content/pages/`

---

### 14. DATABASE SCHEMA

**Core Tables:**

| Table | Purpose |
|---|---|
| `users` | id, wallet_address, email, phone, display_name, avatar_ipfs_hash, kyc_status, country_code, preferred_language, moi_score, created_at |
| `circles` | id, contract_id, name, description, circle_type, payout_type, contribution_amount, currency, frequency, max_members, status, current_round, organizer_id, start_date, end_date |
| `circle_members` | id, circle_id, user_id, position, status, joined_at |
| `contributions` | id, circle_id, user_id, round_number, amount, txn_hash, status, on_time, submitted_at |
| `payouts` | id, circle_id, recipient_id, round_number, amount, fee_amount, txn_hash, payout_type, executed_at |
| `penalties` | id, circle_id, user_id, round_number, penalty_type, amount, strikes_applied, created_at |
| `invites` | id, circle_id, code, created_by, max_uses, use_count, expires_at |
| `notifications` | id, user_id, type, title, body, data, is_read, channel, sent_at |
| `audit_log` | id, actor_id, action, resource_type, resource_id, details, ip_address, user_agent, created_at |
| `webhooks` | id, user_id, url, events, secret_hash, is_active |
| `api_keys` | id, user_id, name, key_hash, scopes, rate_limit, expires_at |
| `sessions` | id, user_id, token_hash, expires_at, created_at |

---

### 15. MIDDLEWARE STACK

| Layer | Implementation |
|---|---|
| CORS | Allow origin: `http://localhost:1110` (dev), `https://app.moistello.io` (prod). Methods: GET, POST, PATCH, DELETE. Credentials: true. |
| Rate Limiting | Redis token bucket: 100 req/min per IP, 300 req/min per authenticated user |
| Auth | JWT validation via middleware. Extract user from token, inject into request context |
| Logging | Structured JSON logs (zerolog or zap). Request ID per request. |
| Recovery | Panic recovery middleware, returns 500 |
| Request Validation | Input validation via go-playground/validator |
| CSRF | Double-submit cookie for state-changing operations |

---

### 16. INFRASTRUCTURE

| Component | Technology |
|---|---|
| Container | Docker |
| Orchestration | Kubernetes (EKS/GKE) or Docker Compose for dev |
| Database | PostgreSQL 16 (RDS or managed) |
| Cache | Redis (ElastiCache) |
| Message Queue | RabbitMQ |
| CDN | Cloudflare |
| Monitoring | Prometheus + Grafana (dashboards for API latency, error rates, business metrics) |
| Logging | Loki or ELK stack |
| CI/CD | GitHub Actions (lint, test, build, deploy) |
| Secrets | AWS Secrets Manager or HashiCorp Vault |

---

### 17. ENDPOINT SUMMARY (29 endpoints)

| Group | Count | Endpoints |
|---|---|---|
| Auth | 6 | nonce, verify, register, refresh, me, logout |
| Users | 7 | me (GET/PATCH), :id, reputation, kyc, kyc/status, me/circles |
| Circles | 15 | list, create, :id (GET/PATCH/DELETE), join, contribute, exit, members, rounds, payouts, dispute, vote, auction-bid |
| Contributions | 2 | list, :id |
| Payouts | 2 | list, :id |
| Invites | 3 | list, create, delete |
| Notifications | 4 | list, read, read-all, preferences |
| Admin | 5 | users, circles, audit-log, metrics, feature-flags |
| Webhooks | 3 | register, list, delete |
| **Total** | **47** | |

## Problem
- 1.7 billion unbanked adults globally rely on informal savings circles with zero legal protection
- Organizers can abscond with pooled funds — no transparency, no recourse
- Late or missed payments rely purely on social pressure — no enforceable penalties
- No portable credit history for participants who reliably contribute
- Existing ROSCA apps (Esusu Africa, Ajala) are centralized, region-locked, and lack smart contract guarantees

## Target Ecosystem
Stellar/Soroban — Stellar's sub-cent fees and 3-5 second settlement make contribution cycles feasible. Targets underbanked communities in Africa, Latin America, South Asia, and Southeast Asia. Directly applicable for Drips Wave.

## Port Configuration
- **Backend API**: Port **1100** (`http://localhost:1100`)
- **Frontend App**: Port **1110** (`http://localhost:1110`)
- **Frontend → Backend**: CORS configured, API calls to `http://localhost:1100/v1/*`

## Admin User Management (Upload System)

### Creating an Admin User
Admin users for the `/upload` page are created via a CLI script on the server:

```bash
node scripts/create-user.js
```

This generates a one-time setup token (valid for 24 hours) and prints a setup URL:
```
http://<server>:1110/setup?token=<64-char-hex-token>
```

The administrator shares this URL with the user. The user visits the URL, sets their username and password, and is automatically logged in.

### Storage
- `content/users.json` — user accounts (username, hashed password via PBKDF2-SHA512, salt, role)
- `content/setup-tokens.json` — pending setup tokens (token, expiresAt, used flag)
- `content/sessions.json` — active sessions (token, userId, createdAt, 7-day expiry)
- `content/pages/` — uploaded page files (.md format)

### Auth Flow
1. Server admin runs `node scripts/create-user.js`
2. Setup token generated → URL shared with user
3. User visits `/setup?token=...` → sets username + password
4. `POST /api/auth/setup` — validates token, creates user, sets session cookie
5. User visits `/upload` → enters username + password
6. `POST /api/auth/login` — validates credentials, sets `moistello_session` httpOnly cookie
7. All `/api/upload` requests check `moistello_session` cookie for authorization
8. `POST /api/auth/logout` — clears session cookie

### Security
- Passwords hashed with PBKDF2-SHA512 (100,000 iterations)
- Sessions stored as httpOnly cookies (not accessible to JavaScript)
- Setup tokens: one-time use, 24-hour expiry
- Sessions: 7-day expiry
- All JSON files stored server-side only (not in public directory)

---

## Product Pillars

### 1. Circle Lifecycle
- **Creation**: Organizer sets circle name, contribution amount, currency (USDC/XLM), cycle frequency (daily/weekly/bi-weekly/monthly), member count, payout order type (random/fixed/auction/vote), start date
- **Invite**: Share invite link/code; members join by accepting invite + depositing first contribution
- **Active**: Each cycle, all members contribute by deadline; one member receives full pool
- **Completion**: After all members receive their payout, circle archives with on-chain record
- **Early Exit**: Emergency withdrawal with circle vote; penalty applied

### 2. Payout Order Types
- **Random**: Smart contract VRF (verifiable random function) selects payout order at circle start
- **Fixed**: Organizer pre-defines order (e.g., seniority, need-based)
- **Auction (Chit Fund)**: Members bid discount amount; lowest bidder (biggest discount) gets payout; discount distributed to all members
- **Vote-based**: Members vote each round on who receives payout (consensus or majority)

### 3. Enforcement & Penalties
- **Late Payment Grace Period**: Configurable (default 24h for daily cycles, 3d for weekly)
- **Late Fee**: Percentage of contribution (default 5%) deducted from payout share
- **Default Strike**: Missed payment without cure → 1 strike; 3 strikes = removal
- **Collateral Staking**: Members stake % of total circle value as good-behavior bond
- **Guarantor System**: Each member can designate a guarantor who covers missed payments

### 4. On-Chain Reputation (MoiScore)
- Contribution streak (consecutive on-time payments)
- Circle completion count
- Late payment frequency
- Default history
- Circle organizer rating (by members)
- Score range: 0-1000, decaying if inactive
- Portable across Moistello circles and future Lumeo integration

### 5. Circle Types
- **Public Circles**: Open to anyone meeting MoiScore threshold
- **Private Circles**: Invite-only via link/code
- **Organizational Circles**: Company-backed with admin controls, bulk member management
- **Community Circles**: DAO-gated (only members of DAO X can join)
- **Premium Circles**: Higher limits, priority support, collateralized

### 6. Multi-Currency Support
- USDC on Stellar (primary)
- XLM (native)
- Future: other Stellar-issued tokens (EURC, BRL, NGN stablecoins)
- On-chain exchange rate oracle for multi-currency circles

### 7. Mobile Money Bridge (Phase 2)
- Off-ramp to M-Pesa, MTN Mobile Money, Airtel Money via partner APIs
- On-ramp from mobile money to USDC on Stellar
- Programmable via Soroban contract + oracle attestation

## Technical Stack
- **Smart Contracts**: Rust (Soroban SDK) — CircleFactory, Circle, ReputationRegistry, GovernanceToken, Treasury, Escrow, VRF
- **Backend**: Go (+ TypeScript for notification/indexer services) — microservices architecture
- **Frontend**: React/TypeScript (Next.js) — PWA with mobile-first design
- **Database**: PostgreSQL (primary), Redis (caching/sessions/queues)
- **Message Queue**: RabbitMQ or BullMQ
- **Monitoring**: Prometheus + Grafana + Loki
- **Storage**: IPFS for decentralized metadata, S3-compatible for assets
- **Wallet**: Freighter, Lobstr, XBull — WalletConnect-compatible

## Drips Wave Alignment
- **Stellar/Soroban native**: 100%
- **Open source**: Apache 2.0
- **Repository structure**: Smart contracts (Rust) + Backend (Go) + Frontend (React/TS)
- **Well-scoped issues**: Smart contract development, contribution tracking, payout logic, reputation algos, UI components, mobile responsiveness, i18n, notification system
- **Difficulty spread**: Trivial (typo fixes, button styles, i18n translations), Medium (API endpoints, notification service, payment tracking), High (VRF randomization, auction logic, governance contracts, collateralization)

## Competitive Landscape
- **Esusu Africa** (Nigeria): Centralized, app-only, no smart contracts
- **Tontine Trust** (Francophone Africa): WhatsApp-based, no automation
- **ChitMonks** (India): Regional, fiat-only, centralized
- **No on-chain ROSCA exists on Stellar** — first mover

## Open Source & Community
- Apache 2.0 license
- Community governance token (MOI) for protocol decisions
- Grant funding target: Stellar Community Fund, Drips Wave rewards
- Integration partners: Freighter, Lobstr, MoneyGram (Stellar anchor), mobile money APIs

## Monetization
- Protocol fee: 0.5% per payout (governance-adjustable)
- Premium circle creation fee (one-time, for organizations)
- Featured circles (paid promotion in discover feed)
- Optional: Yield on deposited funds via Neurowealth/Nester integration (Phase 3)
