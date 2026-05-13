# Moistello — Enterprise Build Plan

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           STELLAR BLOCKCHAIN                             │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────────────┐        │
│  │ Circle   │  │ Reputation│  │Governance│  │  Treasury/Escrow │        │
│  │ Factory  │  │ Registry  │  │  Token   │  │                  │        │
│  └──────────┘  └──────────┘  └──────────┘  └──────────────────┘        │
│                                                                          │
│  Stellar Horizon API ◄──── Stellar RPC ────► Soroban Contracts           │
└──────────────────────────┬──────────────────────────────────────────────┘
                           │
                    Indexer/Event Stream
                           │
┌──────────────────────────▼──────────────────────────────────────────────┐
│                         BACKEND (Go)                                     │
│                                                                          │
│  ┌─────────┐     ┌──────────────────┐     ┌──────────────────────┐      │
│  │  Nginx   │────►│  API Gateway     │────►│  Rate Limiter        │      │
│  │  (TLS)   │     │  (Gin/Echo)      │     │  (Redis token bucket)│      │
│  └─────────┘     └──────┬───────────┘     └──────────────────────┘      │
│                         │                                                │
│         ┌───────────────┼───────────────┬───────────────┐               │
│         ▼               ▼               ▼               ▼               │
│  ┌──────────────┐ ┌──────────┐ ┌──────────────┐ ┌──────────────┐       │
│  │ Circle       │ │ User     │ │ Notification │ │ Analytics    │       │
│  │ Service      │ │ Service  │ │ Service      │ │ Service      │       │
│  └──────┬───────┘ └────┬─────┘ └──────┬───────┘ └──────┬───────┘       │
│         │              │              │                │               │
│  ┌──────┴──────┐ ┌─────┴──────┐ ┌─────┴──────┐ ┌───────┴──────┐       │
│  │ Payment     │ │ Webhook    │ │ Indexer    │ │ Report Gen   │       │
│  │ Service     │ │ Service    │ │ Service    │ │ Service      │       │
│  └─────────────┘ └────────────┘ └────────────┘ └──────────────┘       │
│                                                                          │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────────────────┐    │
│  │PostgreSQL│  │  Redis   │  │ RabbitMQ │  │  IPFS Node           │    │
│  │ (Primary)│  │(cache/Q) │  │ (events) │  │  (metadata / proofs) │    │
│  └──────────┘  └──────────┘  └──────────┘  └──────────────────────┘    │
└──────────────────────────────────────────────────────────────────────────┘
                           │
                    REST API / WebSocket
                           │
┌──────────────────────────▼──────────────────────────────────────────────┐
│                      FRONTEND (React/TypeScript)                         │
│                                                                          │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │                        Next.js 14 (App Router)                    │    │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────────────┐    │    │
│  │  │ Server   │ │ Client   │ │ API      │ │   Middleware      │    │    │
│  │  │Components│ │Components│ │ Routes   │ │ (Auth, i18n, CSRF)│    │    │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────────────┘    │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                                                          │
│  ┌──────────────────────┐  ┌──────────────────────┐                     │
│  │   State Management    │  │   UI Layer            │                    │
│  │  (Zustand + React     │  │  (Radix UI + Tailwind │                    │
│  │   Query + WebSocket)  │  │   + Framer Motion)    │                    │
│  └──────────────────────┘  └──────────────────────┘                     │
│                                                                          │
│  ┌──────────────────────────────────────────────────────────────┐       │
│  │ PWA + Mobile-First + i18n + Accessibility (WCAG 2.1 AA)      │       │
│  └──────────────────────────────────────────────────────────────┘       │
└──────────────────────────────────────────────────────────────────────────┘
```

---

## Phase 0: Foundation (Week 1-2)

### Smart Contracts (Rust / Soroban SDK)

```
contracts/
├── Cargo.toml                    # Workspace config
├── packages/
│   ├── circle-factory/           # Deploys circle instances
│   │   ├── Cargo.toml
│   │   └── src/
│   │       ├── lib.rs            # Contract entry
│   │       ├── contract.rs       # Factory implementation
│   │       ├── types.rs          # CircleConfig, CircleType, etc.
│   │       ├── errors.rs         # Error enum
│   │       ├── events.rs         # CircleCreated, CircleFunded events
│   │       └── tests/
│   │           ├── test.rs       # Unit tests
│   │           └── integration/  # Integration test fixtures
│   ├── circle/                   # Individual circle logic
│   │   ├── Cargo.toml
│   │   └── src/
│   │       ├── lib.rs
│   │       ├── contract.rs       # init, contribute, payout, default, exit
│   │       ├── payout.rs         # Random, fixed, auction, vote logic
│   │       ├── penalties.rs      # Late fees, strikes, slashing
│   │       ├── collateral.rs     # Staking and release
│   │       ├── types.rs          # Circle, Member, Round, PayoutType
│   │       ├── errors.rs
│   │       ├── events.rs
│   │       └── tests/
│   ├── reputation-registry/      # MoiScore tracking
│   │   ├── Cargo.toml
│   │   └── src/
│   │       ├── lib.rs
│   │       ├── contract.rs       # record_payment, record_default, get_score
│   │       ├── scoring.rs        # Score algorithm (streak, frequency, volume)
│   │       ├── types.rs
│   │       ├── errors.rs
│   │       └── tests/
│   ├── governance-token/         # MOI token (CAP-46 standard)
│   │   ├── Cargo.toml
│   │   └── src/
│   │       ├── lib.rs
│   │       ├── contract.rs
│   │       ├── types.rs
│   │       └── tests/
│   └── treasury/                 # Protocol fee management
│       ├── Cargo.toml
│       └── src/
│           ├── lib.rs
│           ├── contract.rs
│           ├── types.rs
│           └── tests/
├── common/                       # Shared library
│   ├── Cargo.toml
│   └── src/
│       ├── lib.rs
│       ├── vrf.rs                # Verifiable random function
│       ├── math.rs               # Fixed-point math for percentages
│       └── auth.rs               # Authorization helpers
└── deploy/
    ├── Makefile                   # Build + deploy commands
    ├── testnet.toml              # Testnet deployment config
    └── mainnet.toml              # Mainnet deployment config
```

**Key Contract Functions:**

| Contract | Function | Description |
|---|---|---|
| CircleFactory | `initialize()` | Set admin, fee rate, treasury address |
| CircleFactory | `create_circle(config)` | Deploy new circle instance, emit event |
| CircleFactory | `get_circles_by_member(addr)` | List active circles for address |
| Circle | `initialize(config)` | Set name, amount, frequency, members, payout type |
| Circle | `join_circle()` | Accept invite, deposit first contribution |
| Circle | `contribute()` | Submit cycle contribution to escrow |
| Circle | `trigger_payout()` | Admin or timed: calculate + distribute payout |
| Circle | `report_late(member)` | Mark member as late, apply penalty |
| Circle | `exit_circle()` | Emergency exit with penalty |
| Circle | `auction_bid(discount_bips)` | Submit auction bid for current round |
| Circle | `vote_payout(member)` | Vote for payout recipient |
| Circle | `dispute(evidence)` | Raise dispute, freeze funds |
| ReputationRegistry | `record_contribution(addr, circle)` | Log on-time payment |
| ReputationRegistry | `record_default(addr, circle)` | Log missed payment |
| ReputationRegistry | `get_moi_score(addr)` | Return 0-1000 score |
| GovernanceToken | `mint(to, amount)` | Admin mint (initial distribution) |
| GovernanceToken | `transfer(from, to, amount)` | Standard transfer |
| Treasury | `deposit_fees()` | Accept protocol fees |
| Treasury | `withdraw(to, amount)` | Governance-controlled withdrawal |

**Events (for Indexer to consume):**
- `CircleCreated(circle_id, creator, config_hash)`
- `MemberJoined(circle_id, member, contribution_amount)`
- `ContributionReceived(circle_id, member, amount, round)`
- `PayoutExecuted(circle_id, recipient, amount, round, payout_type)`
- `LateReported(circle_id, member, penalty_amount, strikes)`
- `MemberExited(circle_id, member, penalty)`
- `DefaultRecorded(circle_id, member)`
- `CircleCompleted(circle_id, total_contributions)`
- `AuctionBid(circle_id, bidder, discount_bips, round)`
- `VoteCast(circle_id, voter, vote_for, round)`
- `DisputeRaised(circle_id, member, evidence_hash)`

### Backend (Go)

```
backend/
├── go.mod
├── go.sum
├── Makefile
├── Dockerfile
├── docker-compose.yml            # Local dev: PG + Redis + RabbitMQ
├── cmd/
│   ├── api-server/               # Main API entry point
│   │   └── main.go
│   ├── indexer/                  # Blockchain event indexer
│   │   └── main.go
│   ├── notification-worker/      # Notification queue consumer
│   │   └── main.go
│   └── migration-runner/         # DB migration tool
│       └── main.go
├── internal/
│   ├── config/
│   │   ├── config.go             # Viper-based config loader
│   │   └── config.yaml           # Default configuration
│   ├── api/
│   │   ├── server.go             # Gin/Echo server setup
│   │   ├── middleware/
│   │   │   ├── auth.go           # JWT + wallet signature verification
│   │   │   ├── cors.go
│   │   │   ├── ratelimit.go      # Token bucket per IP/wallet
│   │   │   ├── logging.go        # Request/response logging
│   │   │   ├── recovery.go       # Panic recovery
│   │   │   ├── csrf.go
│   │   │   └── i18n.go           # Accept-Language header parsing
│   │   ├── routes/
│   │   │   ├── router.go         # Central route registration
│   │   │   ├── circle_routes.go
│   │   │   ├── user_routes.go
│   │   │   ├── payment_routes.go
│   │   │   ├── notification_routes.go
│   │   │   ├── analytics_routes.go
│   │   │   ├── admin_routes.go
│   │   │   └── webhook_routes.go
│   │   └── handlers/
│   │       ├── circle_handler.go
│   │       ├── user_handler.go
│   │       ├── auth_handler.go
│   │       ├── payment_handler.go
│   │       ├── notification_handler.go
│   │       ├── analytics_handler.go
│   │       ├── admin_handler.go
│   │       └── webhook_handler.go
│   ├── domain/
│   │   ├── circle/
│   │   │   ├── model.go          # Circle, Member, Round entities
│   │   │   ├── repository.go     # DB interface
│   │   │   ├── service.go        # Business logic
│   │   │   └── repository_pg.go  # PostgreSQL implementation
│   │   ├── user/
│   │   │   ├── model.go
│   │   │   ├── repository.go
│   │   │   ├── service.go
│   │   │   └── repository_pg.go
│   │   ├── payment/
│   │   │   ├── model.go
│   │   │   ├── repository.go
│   │   │   ├── service.go
│   │   │   └── repository_pg.go
│   │   ├── notification/
│   │   │   ├── model.go
│   │   │   ├── repository.go
│   │   │   ├── service.go
│   │   │   ├── email.go          # SMTP/SendGrid
│   │   │   ├── sms.go            # Twilio/Africa's Talking
│   │   │   ├── push.go           # Firebase / APNs
│   │   │   └── inapp.go          # WebSocket real-time
│   │   ├── analytics/
│   │   │   ├── model.go
│   │   │   ├── repository.go
│   │   │   ├── service.go
│   │   │   └── reports.go
│   │   └── auth/
│   │       ├── model.go
│   │       ├── repository.go
│   │       ├── service.go
│   │       └── jwt.go
│   ├── indexer/
│   │   ├── stellar_client.go     # Horizon + RPC client
│   │   ├── event_processor.go    # Map on-chain events → DB records
│   │   ├── cursor.go             # Track last processed ledger
│   │   └── reconciler.go         # Repair missing events
│   ├── pkg/
│   │   ├── stellar/              # Stellar utilities
│   │   │   ├── client.go         # Horizon API wrapper
│   │   │   ├── soroban.go        # Soroban RPC wrapper
│   │   │   └── keypair.go        # Key management
│   │   ├── ipfs/
│   │   │   └── client.go         # IPFS pinning service
│   │   ├── kyc/
│   │   │   └── sumsub.go         # Sumsub API integration
│   │   ├── websocket/
│   │   │   ├── hub.go            # WebSocket connection manager
│   │   │   ├── client.go         # Individual client connection
│   │   │   └── message.go        # Message types
│   │   ├── queue/
│   │   │   ├── rabbitmq.go       # RabbitMQ publisher/consumer
│   │   │   └── jobs.go           # Job type definitions
│   │   ├── validator/
│   │   │   └── validator.go      # Request validation (go-playground)
│   │   ├── logger/
│   │   │   └── logger.go         # Structured logging (zerolog/zap)
│   │   └── metrics/
│   │       └── metrics.go        # Prometheus metrics
│   └── database/
│       ├── postgres.go           # Connection pool setup
│       ├── migrations/           # golang-migrate SQL files
│       │   ├── 001_create_users.up.sql
│       │   ├── 001_create_users.down.sql
│       │   ├── 002_create_circles.up.sql
│       │   └── ...
│       └── seeds/                # Dev/test seed data
│           └── seed.go
├── pkg/                          # Shared packages (can be imported externally)
│   ├── dtos/
│   │   ├── circle_dto.go
│   │   ├── user_dto.go
│   │   └── payment_dto.go
│   └── errors/
│       └── app_errors.go         # Domain-specific error types
└── tests/
    ├── integration/
    │   ├── circle_test.go
    │   ├── payment_test.go
    │   └── helpers/
    │       └── testutil.go
    └── e2e/
        └── circle_lifecycle_test.go
```

**Database Schema (Core Tables):**

```sql
-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_address VARCHAR(56) UNIQUE NOT NULL,   -- Stellar public key
    email VARCHAR(255) UNIQUE,
    phone VARCHAR(20) UNIQUE,
    display_name VARCHAR(100),
    avatar_ipfs_hash VARCHAR(64),
    kyc_status kyc_status_enum DEFAULT 'unverified',
    kyc_provider_ref VARCHAR(255),
    country_code CHAR(2),
    preferred_language VARCHAR(10) DEFAULT 'en',
    moi_score INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Circles table
CREATE TABLE circles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    contract_id VARCHAR(64) UNIQUE NOT NULL,      -- Soroban contract address
    name VARCHAR(200) NOT NULL,
    description TEXT,
    circle_type circle_type_enum NOT NULL,         -- public, private, org, community, premium
    payout_type payout_type_enum NOT NULL,          -- random, fixed, auction, vote
    contribution_amount NUMERIC(18,7) NOT NULL,    -- In stroops (1 XLM = 10^7 stroops)
    currency currency_enum NOT NULL DEFAULT 'USDC',
    frequency frequency_enum NOT NULL,              -- daily, weekly, biweekly, monthly
    max_members INTEGER NOT NULL,
    min_moi_score INTEGER DEFAULT 0,
    collateral_percent NUMERIC(5,2) DEFAULT 0,     -- % of total circle value
    late_fee_percent NUMERIC(5,2) DEFAULT 5,
    grace_period_hours INTEGER DEFAULT 24,
    max_strikes INTEGER DEFAULT 3,
    start_date TIMESTAMPTZ,
    end_date TIMESTAMPTZ,
    status circle_status_enum DEFAULT 'pending',   -- pending, active, completed, cancelled, disputed
    current_round INTEGER DEFAULT 0,
    total_contributions NUMERIC(18,7) DEFAULT 0,
    organizer_id UUID REFERENCES users(id),
    metadata_ipfs_hash VARCHAR(64),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Circle members (join table)
CREATE TABLE circle_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    circle_id UUID REFERENCES circles(id),
    user_id UUID REFERENCES users(id),
    position INTEGER NOT NULL,                     -- payout order position
    status member_status_enum DEFAULT 'pending',   -- pending, active, completed, defaulted, exited
    joined_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(circle_id, user_id)
);

-- Contributions
CREATE TABLE contributions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    circle_id UUID REFERENCES circles(id),
    user_id UUID REFERENCES users(id),
    round_number INTEGER NOT NULL,
    amount NUMERIC(18,7) NOT NULL,
    txn_hash VARCHAR(64) NOT NULL,                 -- Stellar transaction hash
    status contribution_status_enum DEFAULT 'pending',
    on_time BOOLEAN DEFAULT TRUE,
    submitted_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(circle_id, user_id, round_number)
);

-- Payouts
CREATE TABLE payouts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    circle_id UUID REFERENCES circles(id),
    recipient_id UUID REFERENCES users(id),
    round_number INTEGER NOT NULL,
    amount NUMERIC(18,7) NOT NULL,
    fee_amount NUMERIC(18,7) DEFAULT 0,
    txn_hash VARCHAR(64),
    payout_type VARCHAR(20),
    executed_at TIMESTAMPTZ DEFAULT NOW()
);

-- Penalties
CREATE TABLE penalties (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    circle_id UUID REFERENCES circles(id),
    user_id UUID REFERENCES users(id),
    round_number INTEGER NOT NULL,
    penalty_type penalty_type_enum NOT NULL,       -- late, default, early_exit
    amount NUMERIC(18,7) NOT NULL,
    strikes_applied INTEGER DEFAULT 0,
    reason TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Invites
CREATE TABLE invites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    circle_id UUID REFERENCES circles(id),
    code VARCHAR(20) UNIQUE NOT NULL,
    created_by UUID REFERENCES users(id),
    max_uses INTEGER DEFAULT 1,
    use_count INTEGER DEFAULT 0,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Notifications
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    type notification_type_enum NOT NULL,
    title VARCHAR(200) NOT NULL,
    body TEXT,
    data JSONB,
    is_read BOOLEAN DEFAULT FALSE,
    channel channel_enum NOT NULL,                 -- email, sms, push, inapp
    sent_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Audit log (immutable, append-only)
CREATE TABLE audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_id UUID REFERENCES users(id),
    action VARCHAR(100) NOT NULL,                  -- circle.created, member.joined, payout.executed
    resource_type VARCHAR(50) NOT NULL,            -- circle, user, payment
    resource_id UUID NOT NULL,
    details JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_audit_log_resource ON audit_log(resource_type, resource_id);
CREATE INDEX idx_audit_log_created ON audit_log(created_at DESC);

-- API keys (for webhook consumers + third-party integrations)
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    name VARCHAR(100),
    key_hash VARCHAR(64) UNIQUE NOT NULL,
    scopes TEXT[],
    rate_limit INTEGER DEFAULT 100,                -- requests per minute
    expires_at TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

**API Design (RESTful, versioned):**

```
Base URL: https://api.moistello.io/v1

Auth:
  POST   /auth/nonce                  # Get nonce for wallet signature
  POST   /auth/verify                 # Verify wallet signature → JWT
  POST   /auth/refresh                # Refresh JWT token
  POST   /auth/logout                 # Invalidate session
  GET    /auth/session                # Current session info

Users:
  GET    /users/me                    # Current user profile
  PATCH  /users/me                    # Update profile
  GET    /users/:id                   # Public user profile
  GET    /users/:id/reputation        # MoiScore + history
  POST   /users/me/kyc               # Initiate KYC verification
  GET    /users/me/kyc/status         # Check KYC status
  GET    /users/me/circles            # My active circles
  GET    /users/me/contributions      # Contribution history

Circles (RESTful):
  GET    /circles                     # List/discover circles (paginated, filterable)
  POST   /circles                     # Create new circle
  GET    /circles/:id                 # Circle details
  PATCH  /circles/:id                 # Update circle settings (organizer only)
  DELETE /circles/:id                 # Cancel circle (organizer, pre-start only)
  POST   /circles/:id/join            # Join circle (with invite code)
  POST   /circles/:id/contribute      # Submit contribution
  POST   /circles/:id/exit            # Exit circle
  GET    /circles/:id/members         # Circle members list
  GET    /circles/:id/rounds          # Round history
  GET    /circles/:id/rounds/:round   # Specific round details
  GET    /circles/:id/payouts         # Payout history
  POST   /circles/:id/dispute         # Raise dispute
  POST   /circles/:id/vote            # Cast vote (voting circles)
  POST   /circles/:id/auction-bid     # Submit auction bid (auction circles)

Invites:
  GET    /circles/:id/invites         # List invites for circle
  POST   /circles/:id/invites         # Generate invite link/code
  DELETE /invites/:code               # Revoke invite

Payments:
  GET    /payments                    # User payment history
  GET    /payments/:id                # Payment details + Stellar txn link
  POST   /payments/estimate           # Estimate fees for contribution

Notifications:
  GET    /notifications               # User notifications (paginated)
  PATCH  /notifications/:id/read      # Mark as read
  PATCH  /notifications/read-all      # Mark all as read
  PUT    /notifications/preferences   # Update notification preferences

Analytics:
  GET    /analytics/overview          # Platform-wide stats
  GET    /analytics/circles/:id       # Single circle analytics
  GET    /analytics/user              # Personal contribution analytics

Admin (rate-limited, role-gated):
  GET    /admin/users                 # User management
  GET    /admin/circles               # Circle management
  GET    /admin/audit-log             # Activity audit trail
  GET    /admin/metrics               # Platform metrics
  POST   /admin/feature-flags         # Manage feature flags

Webhooks:
  POST   /webhooks/register           # Register webhook endpoint
  GET    /webhooks                    # List registered webhooks
  DELETE /webhooks/:id                # Remove webhook
  GET    /webhooks/:id/deliveries     # Delivery history
  POST   /webhooks/:id/retry          # Retry failed delivery

Health:
  GET    /health                      # Liveness probe
  GET    /health/ready                # Readiness probe (DB + Stellar connection)
  GET    /metrics                     # Prometheus metrics endpoint
```

**Webhook Event Types:**
- `circle.created`
- `circle.started`
- `circle.completed`
- `circle.cancelled`
- `member.joined`
- `member.exited`
- `member.defaulted`
- `contribution.received`
- `contribution.late`
- `payout.executed`
- `round.started`
- `round.completed`
- `dispute.raised`
- `dispute.resolved`

### Frontend (React/TypeScript — Next.js 14)

```
frontend/
├── package.json
├── tsconfig.json
├── next.config.js                 # PWA, i18n, image domains, env
├── tailwind.config.ts
├── postcss.config.js
├── Dockerfile
├── .env.example
├── public/
│   ├── manifest.json              # PWA manifest
│   ├── sw.js                      # Service worker
│   ├── icons/                     # PWA icons (192, 512)
│   └── locales/                   # Static i18n files
│       ├── en/
│       │   └── common.json
│       ├── fr/
│       │   └── common.json
│       ├── sw/
│       │   └── common.json
│       ├── es/
│       │   └── common.json
│       ├── pt/
│       │   └── common.json
│       └── hi/
│           └── common.json
├── src/
│   ├── app/                       # Next.js App Router
│   │   ├── layout.tsx             # Root layout (providers, i18n, theme)
│   │   ├── page.tsx               # Landing page
│   │   ├── (auth)/                # Auth route group
│   │   │   ├── layout.tsx         # Auth layout (minimal)
│   │   │   ├── login/
│   │   │   │   └── page.tsx
│   │   │   └── register/
│   │   │       └── page.tsx
│   │   ├── (dashboard)/           # Authenticated route group
│   │   │   ├── layout.tsx         # Dashboard layout (sidebar + header)
│   │   │   ├── dashboard/
│   │   │   │   └── page.tsx       # Home dashboard
│   │   │   ├── circles/
│   │   │   │   ├── page.tsx       # Circle discovery
│   │   │   │   ├── [id]/
│   │   │   │   │   ├── page.tsx   # Circle detail
│   │   │   │   │   ├── members/
│   │   │   │   │   │   └── page.tsx
│   │   │   │   │   ├── rounds/
│   │   │   │   │   │   └── page.tsx
│   │   │   │   │   └── settings/
│   │   │   │   │       └── page.tsx  # Organizer settings
│   │   │   │   └── create/
│   │   │   │       └── page.tsx   # Circle creation wizard
│   │   │   ├── contributions/
│   │   │   │   └── page.tsx       # Contribution history
│   │   │   ├── payouts/
│   │   │   │   └── page.tsx       # Payout history
│   │   │   ├── reputation/
│   │   │   │   └── page.tsx       # MoiScore + reputation detail
│   │   │   ├── notifications/
│   │   │   │   └── page.tsx       # Notification center
│   │   │   ├── settings/
│   │   │   │   └── page.tsx       # User settings + KYC
│   │   │   └── wallet/
│   │   │       └── page.tsx       # Wallet connection / balance
│   │   └── api/                   # Next.js API routes (for client-side proxy)
│   │       ├── auth/
│   │       └── health/
│   ├── components/
│   │   ├── ui/                    # Primitive UI components
│   │   │   ├── button.tsx
│   │   │   ├── input.tsx
│   │   │   ├── select.tsx
│   │   │   ├── modal.tsx
│   │   │   ├── toast.tsx
│   │   │   ├── badge.tsx
│   │   │   ├── card.tsx
│   │   │   ├── table.tsx
│   │   │   ├── tabs.tsx
│   │   │   ├── tooltip.tsx
│   │   │   ├── dropdown.tsx
│   │   │   ├── skeleton.tsx
│   │   │   ├── progress.tsx
│   │   │   ├── avatar.tsx
│   │   │   ├── checkbox.tsx
│   │   │   ├── radio-group.tsx
│   │   │   ├── switch.tsx
│   │   │   ├── dialog.tsx
│   │   │   ├── sheet.tsx
│   │   │   ├── accordion.tsx
│   │   │   ├── date-picker.tsx
│   │   │   └── file-upload.tsx
│   │   ├── layout/
│   │   │   ├── sidebar.tsx        # Dashboard sidebar navigation
│   │   │   ├── header.tsx         # Top header with wallet + notifications
│   │   │   ├── footer.tsx
│   │   │   ├── dashboard-layout.tsx
│   │   │   └── mobile-nav.tsx     # Bottom nav for mobile
│   │   ├── circles/
│   │   │   ├── circle-card.tsx    # Circle in list view
│   │   │   ├── circle-detail-header.tsx
│   │   │   ├── circle-members-list.tsx
│   │   │   ├── circle-round-timeline.tsx
│   │   │   ├── circle-payout-history.tsx
│   │   │   ├── circle-create-form.tsx  # Multi-step wizard
│   │   │   ├── circle-invite-modal.tsx
│   │   │   ├── circle-contribute-form.tsx
│   │   │   ├── circle-auction-panel.tsx
│   │   │   ├── circle-vote-panel.tsx
│   │   │   └── circle-status-badge.tsx
│   │   ├── reputation/
│   │   │   ├── moi-score-gauge.tsx     # 0-1000 radial gauge
│   │   │   ├── reputation-breakdown.tsx
│   │   │   └── reputation-history.tsx  # Timeline chart
│   │   ├── wallet/
│   │   │   ├── wallet-connect.tsx      # Freighter/xBull connection
│   │   │   ├── wallet-balance.tsx
│   │   │   ├── wallet-transaction-list.tsx
│   │   │   └── wallet-sign-message.tsx
│   │   ├── analytics/
│   │   │   ├── overview-stats.tsx
│   │   │   ├── contribution-chart.tsx
│   │   │   ├── circle-progress.tsx
│   │   │   └── platform-metrics.tsx
│   │   └── shared/
│   │       ├── empty-state.tsx
│   │       ├── error-boundary.tsx
│   │       ├── loading-spinner.tsx
│   │       ├── page-header.tsx
│   │       ├── confirm-dialog.tsx
│   │       ├── copy-button.tsx
│   │       └── stellar-address-display.tsx
│   ├── hooks/
│   │   ├── use-auth.ts
│   │   ├── use-wallet.ts           # Stellar wallet connection hook
│   │   ├── use-circles.ts          # TanStack Query hooks for circles
│   │   ├── use-contributions.ts
│   │   ├── use-payouts.ts
│   │   ├── use-reputation.ts
│   │   ├── use-notifications.ts
│   │   ├── use-websocket.ts        # WebSocket connection hook
│   │   ├── use-i18n.ts
│   │   ├── use-media-query.ts
│   │   ├── use-debounce.ts
│   │   └── use-intersection-observer.ts
│   ├── stores/                     # Zustand stores
│   │   ├── auth-store.ts
│   │   ├── wallet-store.ts
│   │   ├── ui-store.ts            # Theme, sidebar state, modals
│   │   └── notification-store.ts   # In-app notification state
│   ├── lib/
│   │   ├── api-client.ts          # Axios instance with interceptors
│   │   ├── stellar.ts             # Horizon RPC + Freighter helpers
│   │   ├── validators.ts          # Zod schemas for forms
│   │   ├── formatters.ts          # Currency, date, address formatting
│   │   ├── constants.ts           # App-wide constants
│   │   ├── utils.ts               # General utilities
│   │   └── cn.ts                  # Tailwind class merging (clsx + twMerge)
│   ├── types/
│   │   ├── circle.ts
│   │   ├── user.ts
│   │   ├── payment.ts
│   │   ├── notification.ts
│   │   ├── api.ts                 # API response types
│   │   └── stellar.ts             # Stellar-specific types
│   ├── providers/
│   │   ├── auth-provider.tsx       # Auth context
│   │   ├── wallet-provider.tsx     # Stellar wallet context
│   │   ├── theme-provider.tsx      # Dark/light theme
│   │   ├── i18n-provider.tsx       # Internationalization
│   │   ├── query-provider.tsx      # TanStack Query
│   │   └── websocket-provider.tsx  # WebSocket context
│   └── styles/
│       ├── globals.css             # Tailwind base + custom vars
│       └── theme.ts                # Theme tokens (light/dark)
├── tests/
│   ├── unit/
│   │   ├── components/
│   │   └── hooks/
│   ├── integration/
│   │   └── pages/
│   └── e2e/
│       ├── circle-lifecycle.spec.ts
│       └── auth-flow.spec.ts
└── cypress/                        # E2E test config
    ├── cypress.config.ts
    └── fixtures/
```

**Frontend Routes Map:**

| Route | Page | Auth | Description |
|---|---|---|---|
| `/` | LandingPage | No | Hero, features, CTA, supported currencies, testimonials |
| `/login` | LoginPage | No | Wallet connect + optional email |
| `/register` | RegisterPage | No | Create profile after wallet connect |
| `/dashboard` | DashboardHome | Yes | Active circles overview, recent activity, MoiScore |
| `/circles` | CircleDiscovery | Yes | Browse/filter public circles, join |
| `/circles/create` | CircleCreate | Yes | Multi-step wizard (details, members, payout type, review) |
| `/circles/[id]` | CircleDetail | Yes | Overview, round tracker, member list, actions |
| `/circles/[id]/members` | CircleMembers | Yes | Full member list with status, reputation |
| `/circles/[id]/rounds` | CircleRounds | Yes | Round-by-round history |
| `/circles/[id]/settings` | CircleSettings | Yes | Organizer-only: edit circle, manage invites |
| `/contributions` | ContributionHistory | Yes | All contributions across circles |
| `/payouts` | PayoutHistory | Yes | All payouts received |
| `/reputation` | ReputationPage | Yes | Detailed MoiScore breakdown + history chart |
| `/notifications` | NotificationCenter | Yes | All notifications, mark read, preferences |
| `/settings` | UserSettings | Yes | Profile, KYC status, notification prefs, language |
| `/wallet` | WalletPage | Yes | Connect wallet, view balance, transaction list |

---

## Phase 1: Core Circle Engine (Week 1-3)

### Smart Contracts (Priority Order)
1. **common/** crate — shared math, VRF, auth helpers (Day 1-2)
2. **circle-factory/** — factory pattern for deploying circle instances (Day 2-3)
3. **circle/** — core circle logic: join, contribute, payout (random + fixed only) (Day 3-8)
4. **reputation-registry/** — basic score recording (Day 8-10)

### Backend (Priority Order)
1. Project scaffold: Go modules, Docker, Postgres, Redis (Day 1)
2. Database schema + migrations (Day 2)
3. Config system + structured logging (Day 2)
4. API server skeleton with middleware stack (Day 3)
5. Auth flow: nonce → wallet signature → JWT (Day 3-4)
6. User CRUD + KYC integration (Day 4-5)
7. Circle CRUD with Postgres (Day 5-7)
8. Indexer service: consume Stellar events → DB (Day 7-9)
9. Contribution tracking (Day 9-10)
10. WebSocket server for real-time updates (Day 10-11)

### Frontend (Priority Order)
1. Project scaffold: Next.js 14, Tailwind, Radix UI, Zustand (Day 1)
2. UI component library (all primitives) (Day 2-4)
3. Layout components: sidebar, header, mobile nav (Day 4-5)
4. Auth pages: login, register (Day 5-6)
5. Dashboard home page (Day 6-7)
6. Circle creation wizard (Day 7-9)
7. Circle detail page (Day 9-11)
8. Wallet connection + Freighter integration (Day 11-12)

---

## Phase 2: Advanced Circle Features (Week 4-5)

### Smart Contracts
1. Auction payout logic (Day 1-3)
2. Vote-based payout logic (Day 3-5)
3. Collateral staking + slashing (Day 5-7)
4. Circle governance token (MOI) (Day 7-8)
5. Dispute resolution contract (Day 8-10)

### Backend
1. Auction/vote API endpoints (Day 1-2)
2. Collateral tracking + release logic (Day 2-3)
3. Penalty engine: late fees, strikes, early exit (Day 3-5)
4. Invite system: code generation, validation (Day 5-6)
5. Notification service: email (SendGrid) + SMS (Twilio) + in-app (Day 6-8)
6. Webhook registration + delivery system (Day 8-10)

### Frontend
1. Circle auction panel (Day 1-2)
2. Circle vote panel (Day 2-3)
3. Circle contribute flow with Freighter signing (Day 3-4)
4. Circle members page with reputation badges (Day 4-5)
5. Circle round timeline component (Day 5-6)
6. Notification center page (Day 6-7)
7. Notification preferences (Day 7)
8. Invite modal + share flow (Day 8)
9. WebSocket real-time updates integration (Day 8-10)

---

## Phase 3: Enterprise & Scale (Week 6-7)

### Smart Contracts
1. Treasury/protocol fee contract (Day 1)
2. Contract upgrade patterns (proxy/delegate) (Day 1-2)
3. Gas optimization pass (Day 2-3)
4. Comprehensive test suite (Day 3-5)

### Backend
1. Analytics service: aggregation queries, report generation (Day 1-2)
2. Rate limiting: per-wallet, per-IP, per-endpoint (Day 2)
3. API key management for third-party integrations (Day 3)
4. Admin dashboard endpoints (Day 3-4)
5. Feature flag system (Day 4)
6. Audit log service (Day 4-5)
7. Performance optimization: query tuning, caching (Day 5-6)
8. API documentation (OpenAPI/Swagger) (Day 6-7)

### Frontend
1. Analytics pages: overview stats, charts (Day 1-2)
2. Reputation page: MoiScore gauge + breakdown + history (Day 2-3)
3. Admin pages: user management, circle management, audit log (Day 3-5)
4. i18n: translation files for 6 languages (Day 5-6)
5. PWA: service worker, offline support, install prompt (Day 6)
6. Accessibility audit + fixes (WCAG 2.1 AA) (Day 6-7)
7. Performance: code splitting, image optimization, bundle analysis (Day 7)

---

## Phase 4: Polish & Launch (Week 8)

### Operations
1. End-to-end tests (Cypress + Stellar testnet) (Day 1-3)
2. Load testing (k6) — target: 10k concurrent circles (Day 3-4)
3. Security audit (smart contracts external auditor) (Day 4-5)
4. Documentation: developer docs, API reference, user guide (Day 5-7)
5. CI/CD pipeline finalization (Day 7)
6. Production deployment (Day 8)

---

## Infrastructure & DevOps

```
infrastructure/
├── docker/
│   ├── backend/
│   │   ├── Dockerfile.prod
│   │   └── Dockerfile.dev
│   ├── frontend/
│   │   ├── Dockerfile.prod
│   │   └── Dockerfile.dev
│   └── indexer/
│       └── Dockerfile
├── kubernetes/
│   ├── namespace.yaml
│   ├── backend/
│   │   ├── deployment.yaml
│   │   ├── service.yaml
│   │   ├── ingress.yaml
│   │   └── hpa.yaml                  # Horizontal Pod Autoscaler
│   ├── frontend/
│   │   ├── deployment.yaml
│   │   ├── service.yaml
│   │   └── ingress.yaml
│   ├── postgres/
│   │   ├── statefulset.yaml
│   │   ├── service.yaml
│   │   └── backup-cronjob.yaml
│   ├── redis/
│   │   └── deployment.yaml
│   └── rabbitmq/
│       └── deployment.yaml
├── terraform/                         # Cloud infrastructure (AWS/GCP)
│   ├── main.tf
│   ├── variables.tf
│   ├── outputs.tf
│   ├── rds.tf                        # Managed PostgreSQL
│   ├── elasticache.tf                # Managed Redis
│   ├── eks.tf                        # Kubernetes cluster
│   └── cloudflare.tf                 # DNS + CDN + WAF
├── monitoring/
│   ├── prometheus/
│   │   └── prometheus.yml
│   ├── grafana/
│   │   └── dashboards/
│   │       ├── api-overview.json
│   │       ├── business-metrics.json
│   │       └── smart-contracts.json
│   ├── loki/
│   │   └── loki-config.yaml
│   └── alertmanager/
│       └── alert-rules.yml
└── ci/
    └── .github/
        └── workflows/
            ├── backend-ci.yml         # Build, test, lint, docker push
            ├── frontend-ci.yml        # Build, test, lint, docker push
            ├── contracts-ci.yml       # Build, test, deploy testnet
            ├── e2e-tests.yml          # Cypress E2E on staging
            └── deploy-prod.yml        # Production deployment
```

**CI/CD Pipeline:**

```
Pull Request:
  1. Lint (golangci-lint, eslint, prettier)
  2. Unit tests (go test, vitest)
  3. Contract tests (cargo test)
  4. Build Docker images
  5. Deploy to preview environment

Merge to main:
  1. All PR steps
  2. Integration tests
  3. Deploy to staging
  4. Run E2E tests against staging
  5. Manual approval gate

Release tag:
  1. Build production images
  2. Deploy to production (blue/green)
  3. Smoke tests
  4. Rollback on failure
```

---

## Communication Between Frontend & Backend

Since they're hosted separately:

```
                   CORS configured on backend
Frontend ──────────► Backend API ────────► PostgreSQL
(Next.js)           (Go/Gin)              Redis
  │                    │                  RabbitMQ
  │ WebSocket          │                  Stellar Horizon
  └────────────────────┘                  Stellar RPC
```

**Authentication Flow:**
1. Frontend requests nonce from `POST /auth/nonce`
2. User signs nonce with Freighter wallet
3. Frontend sends signed message to `POST /auth/verify`
4. Backend verifies signature, returns JWT (access + refresh tokens)
5. Frontend stores JWT in httpOnly cookie or secure localStorage
6. All subsequent requests include `Authorization: Bearer <token>` header
7. WebSocket connection authenticated via token in initial handshake

**CORS Configuration (Backend):**
- Allow origin: `https://app.moistello.io` (production), `http://localhost:3000` (dev)
- Allow methods: GET, POST, PATCH, DELETE, OPTIONS
- Allow headers: Authorization, Content-Type, X-Request-ID
- Allow credentials: true (for cookies)
- Max age: 86400 (24h preflight cache)

**API Versioning Strategy:**
- URL path versioning: `/v1/circles`, `/v2/circles`
- Deprecation header: `X-API-Deprecated: true` with `Sunset` header
- Support N-1 version (current + previous)

---

## Security Architecture

| Layer | Measure |
|---|---|
| Transport | TLS 1.3 mandatory, HSTS preload |
| Auth | JWT (RS256) short-lived access (15min) + refresh (7d) |
| Wallet | Nonce-based signature verification, no private key ever stored |
| API | Rate limiting (Redis token bucket: 100 req/min per IP, 300 req/min per wallet) |
| CSRF | Double-submit cookie pattern for state-changing operations |
| Input | Zod (frontend) + go-playground/validator (backend) validation on all inputs |
| SQL | Parameterized queries only, no raw SQL concat |
| XSS | Content-Security-Policy header, React's built-in escaping |
| Secrets | Vault/HashiCorp or AWS Secrets Manager, never in env vars |
| Smart Contracts | Overflow checks (Rust), reentrancy guards, access control modifiers |
| DDoS | Cloudflare WAF + rate limiting at CDN layer |
| Audit | Immutable audit log table, append-only |

---

## Monitoring & Observability

| Signal | Tool | Metric Examples |
|---|---|---|
| Infrastructure | Prometheus + Grafana | CPU, memory, disk, pod count |
| API | Prometheus + Grafana | Request rate, latency (p50/p95/p99), error rate by endpoint |
| Business | Custom metrics | Circles created/day, contributions/day, payout volume, active users |
| Logs | Loki + Grafana | Structured JSON logs with trace IDs |
| Traces | OpenTelemetry + Jaeger | Request tracing across services |
| Alerts | Alertmanager → PagerDuty | API error rate > 1%, payout failure, indexer lag > 5min |
| Uptime | UptimeRobot / Checkly | 5-min interval checks on health endpoint |
| Errors | Sentry (frontend + backend) | Error grouping, release tracking |

---

## Testing Strategy

| Level | Tool | Coverage Target | What |
|---|---|---|---|
| Unit - Contracts | cargo test | 95% | Every contract function, all branches |
| Unit - Backend | go test | 90% | Services, handlers with mocked repos |
| Unit - Frontend | Vitest | 85% | Components, hooks, utilities |
| Integration - Backend | go test + testcontainers | 80% | API endpoints with real PG/Redis |
| Integration - Contracts | Stellar testnet | N/A | Full deploy and invoke flow |
| E2E | Cypress | Happy paths | Auth → create circle → contribute → payout |
| Load | k6 | N/A | 1000 concurrent users, 100 active circles |
| Security | Slither + manual review | N/A | Contract audit |
| A11y | axe-core + pa11y | N/A | WCAG 2.1 AA compliance |

---

## Key Technical Decisions & Rationale

| Decision | Rationale |
|---|---|
| **Go for backend** | Strong concurrency (goroutines for indexer/websocket), single binary deploy, excellent Stellar SDK support |
| **PostgreSQL over MongoDB** | ACID transactions for financial data, rich JSONB for semi-structured, strong type system |
| **Gin/Echo over Fiber** | Mature ecosystem, middleware chaining, battle-tested |
| **RabbitMQ over Kafka** | Lighter ops, sufficient for notification/event workloads, no need for log replay |
| **Next.js App Router** | Server components for SEO (landing page), client components for interactive dashboard — best of both |
| **Zustand over Redux** | Minimal boilerplate, works with React Query for server state |
| **TanStack Query** | Deduplication, caching, background refetch, optimistic updates |
| **Radix UI over MUI** | Headless, accessible, unstyled — full control with Tailwind |
| **Soroban SDK** | Native Stellar smart contracts, Rust safety, low gas costs |
| **Separate indexer service** | Blockchain events are the source of truth; indexer syncs them to fast DB for queries |
| **WebSocket for real-time** | Round status changes, contribution confirmations, notifications — push not poll |

---

## Directory Summary

```
/root/
├── moistello-backend/          # Go backend (API + indexer + workers)
│   └── product-spec.md         # Product specification (this document's source)
├── moistello-frontend/         # Next.js frontend (React/TypeScript)
├── ab-lumeo/product-spec.md
├── ab-tractos/product-spec.md
├── ab-forgeon/product-spec.md
└── ab-vortic/product-spec.md
```

---

## First Steps (After Plan Approval)

1. Scaffold smart contract workspace (`cargo new --lib` + workspace Cargo.toml)
2. Scaffold Go backend (`go mod init`, Gin server skeleton, Docker compose)
3. Scaffold Next.js frontend (`create-next-app`, Tailwind, Radix UI, Zustand)
4. Write `contracts/common/` crate (shared types, math, VRF, auth)
5. Write `backend/internal/config/` and `backend/internal/database/` (foundation)
6. Write `frontend/src/components/ui/` component library (all primitives)
7. Deploy first contract to Stellar testnet
8. Implement `POST /auth/nonce` and `POST /auth/verify` (first API endpoints)
9. Build login page + wallet connection (first frontend page)
