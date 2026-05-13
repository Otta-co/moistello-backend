# Moistello Backend

Enterprise Go backend for the Moistello decentralized savings circles platform on Stellar/Soroban.

## Architecture

```
cmd/api-server          — REST API server (port 1100)
cmd/indexer             — Stellar event indexer
cmd/notification-worker — Email/SMS notification consumer
cmd/webhook-dispatcher  — Webhook delivery worker
cmd/migrate             — Database migration runner

internal/
├── api/                — Gin handlers + middleware (47 endpoints)
├── domain/             — Business logic (11 domains, 8 repositories)
├── indexer/            — Chain event sync engine
├── websocket/          — Real-time push (per-circle rooms)
└── database/           — SQL migrations (15 up + 15 down)

pkg/                    — Shared packages
├── stellar/            — Horizon + Soroban RPC client
└── stellar/soroban/    — Contract bindings + deployment
```

## Quick Start

```bash
# Infrastructure
make docker-up          # Start PostgreSQL + Redis + RabbitMQ

# Database
make migrate-up         # Apply all migrations

# Run API server
make run                # Starts on port 1100

# Run indexer
make run-indexer        # Syncs Stellar events → DB

# Run tests
make test               # 127 tests across 11 packages
make test-cover         # With coverage report
```

## API (47 endpoints)

See `BACKEND-IMPLEMENTATION-PLAN.md` for full endpoint documentation.

## Configuration

Copy `.env.example` to `.env` and customize. See `config/config.yaml` for all options.

## Testnet

- Account: GAX23V3...T27RC (10,000 XLM, funded)
- Horizon: https://horizon-testnet.stellar.org
- RPC: https://soroban-testnet.stellar.org

## Security

- Ed25519 wallet signature verification
- JWT auth (RS256, 15min access + 7d refresh)
- Rate limiting (Redis token bucket)
- CORS configuration
- Thread-safe account sequence management
- Pre-flight transaction simulation
- Error classification with retryable detection

## License

Apache 2.0
