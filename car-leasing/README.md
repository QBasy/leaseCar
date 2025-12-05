# Car Leasing Service — Production-Ready Monorepo

Complete microservices architecture for a car leasing platform with payment processing and blockchain integration. Built with **Fastify** (API Gateway), **Go Fiber** (microservices), **PostgreSQL**, **Redis**, **MeiliSearch**, and **TON blockchain**.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                         CLIENT REQUESTS                              │
└──────────────────────────┬──────────────────────────────────────────┘
                           │
                ┌──────────▼──────────┐
                │   core-api:3000     │  (Fastify API Gateway)
                │   - JWT Auth        │  - Routes aggregation
                │   - Redis Sessions  │  - MeiliSearch queries
                └───┬────────┬────────┘
                    │        │
        ┌───────────┘        └──────────────┐
        │                                   │
    ┌───▼─────────────┐        ┌───────────▼──────┐
    │ lease-service   │        │payment-service   │  (Go Fiber microservices)
    │ :3001           │        │ :3002            │
    │ - Lease CRUD    │        │ - Strategies     │
    │ - MeiliSearch   │        │ - Webhooks       │
    │ - Indexing      │        │ - State machine  │
    └─────────────────┘        └───────┬──────────┘
                                       │ publishes to Redis
                        ┌──────────────▼───────────┐
                        │blockchain-service :3003  │ (Observer pattern)
                        │- Redis subscriber        │ (Go Fiber)
                        │- TON contract calls      │
                        │- Tx confirmation polling │
                        └──────────────────────────┘

INFRASTRUCTURE:
─ PostgreSQL:15 (relational DB, migrations auto-run)
─ Redis:7 (pub/sub channels, sessions, cache)
─ MeiliSearch:1.5 (full-text search for leases)
```

## Quick Start (Docker Compose)

### Prerequisites
- Docker & Docker Compose
- (Optional) Go 1.21, Node.js 20 for local dev

### Run Full Stack

```bash
# Navigate to project root
cd car-leasing

# Create .env from template (already exists, modify if needed)
cat .env  # review POSTGRES_PASSWORD, JWT_SECRET, TON_* vars

# Start all services
docker-compose up -d

# View logs
docker-compose logs -f core-api
docker-compose logs -f lease-service
docker-compose logs -f payment-service
docker-compose logs -f blockchain-service

# Stop everything
docker-compose down
```

### Health Checks

```bash
# Check all services are running
curl http://localhost:3000/health    # core-api
curl http://localhost:3001/health    # lease-service
curl http://localhost:3002/health    # payment-service
curl http://localhost:3003/health    # blockchain-service

# MeiliSearch UI
open http://localhost:7700

# PostgreSQL
docker-compose exec postgres psql -U leasing_user -d leasing_db

# Redis
docker-compose exec redis redis-cli
```

## Service Details

### 1. **core-api** (Fastify, Node.js)
**Responsibility:** API Gateway, authentication, routing, search aggregation

**Key endpoints:**
- `POST /auth/login` — User authentication (returns JWT token)
- `GET /api/v1/leases?q=search_term` — Search leases via MeiliSearch
- `GET /api/v1/leases/:id` — Proxy to lease-service

**Architecture:**
- Routes → Controllers → Services → Repositories
- Plugin-based: JWT auth, Redis sessions, MeiliSearch client
- Env vars: `JWT_SECRET`, `SESSION_SECRET`, `REDIS_*`, `MEILISEARCH_*`

**Local dev:**
```bash
cd core-api
npm install
npm run dev  # starts on :3000
```

---

### 2. **lease-service** (Go Fiber)
**Responsibility:** Lease CRUD, payment schedule generation, MeiliSearch indexing

**Key endpoints:**
- `POST /leases` — Create new lease
- `GET /leases/:id` — Fetch lease details
- `GET /leases?q=query` — Search leases (via MeiliSearch)

**Architecture:**
- Repository → Service → Controller pattern
- Async indexing to MeiliSearch after CRUD ops
- Config: `/config/config.yaml` loaded at startup

**Patterns used:**
- **Repository Pattern** — `LeaseRepository` for DB persistence
- **DTO Layer** — Structured request/response validation
- **Adapter Pattern** — `MeiliAdapter` for search operations

**Local dev:**
```bash
cd lease-service
go mod tidy
go run ./cmd/lease-service  # starts on :3001
```

---

### 3. **payment-service** (Go Fiber)
**Responsibility:** Process payments via multiple strategies, emit events to blockchain

**Key endpoints:**
- `POST /payments` — Create payment (accepts provider: "stripe" | "bank_api")
- `POST /webhooks/:provider` — Receive provider webhooks (Stripe, Bank API)

**Architecture:**
- **Strategy Pattern** — `PaymentStrategy` interface with implementations (Stripe, Bank)
- **Factory Pattern** — `PaymentFactory` returns correct strategy by provider
- **Adapter Pattern** — `BankAdapter` wraps external Bank API calls
- **Observer Pattern** — Publishes `payment.completed` event to Redis `payments` channel
- **State Machine** — Payment status: PENDING → PROCESSING → COMPLETED

**Strategies:**
```go
// Add new payment strategy:
// 1. Create internal/strategies/newstrategy_strategy.go
type NewStrategy struct { /* ... */ }
func (s *NewStrategy) Process(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) { /* ... */ }

// 2. Register in factory
func (f *PaymentFactory) GetStrategy(provider string) PaymentStrategy {
    case "new_provider":
        return strategies.NewNewStrategy(...)
}

// 3. Call /payments with provider="new_provider"
```

**Event Publishing:**
```
payment-service → Redis publish("payments", {
  "event": "payment.completed",
  "payment_id": "uuid",
  "provider_tx": "stripe_tx_xxx",
  "status": "COMPLETED"
})
```

**Local dev:**
```bash
cd payment-service
go mod tidy
go run ./cmd/payment-service  # starts on :3002
```

---

### 4. **blockchain-service** (Go Fiber)
**Responsibility:** Listen for payment events, process TON blockchain transactions, poll for confirmation

**Architecture:**
- **Observer Pattern** — Redis subscriber on `payments` channel
- **Adapter Pattern** — `TONAdapter` for blockchain interactions
- **Async Processing** — Background polling for transaction confirmations

**Flow:**
```
1. payment-service emits "payment.completed" → Redis
2. blockchain-service subscriber receives event
3. TONAdapter.SendTransaction() → TON blockchain
4. Save blockchain_transaction record in DB
5. Background goroutine polls TON API for status
6. Update blockchain_transactions.confirmed = true (when CONFIRMED)
```

**TON Integration:**
- Replace `TON_WALLET_ADDRESS` and `TON_PRIVATE_KEY` in `.env` with real credentials
- `TONAdapter.SendTransaction()` uses TON API to sign and broadcast transactions
- In production: integrate proper TON SDK (tonweb-go or similar)

**Local dev:**
```bash
cd blockchain-service
go mod tidy
go run ./cmd/blockchain-service  # starts on :3003, subscribes to Redis
```

---

## Database Schema

**Main tables:**
- `users` — User accounts (email, password hash)
- `vehicles` — Available cars (make, model, price, etc.)
- `leases` — Lease agreements (user_id, vehicle_id, dates, status)
- `payments` — Payment records (lease_id, amount, status, blockchain_tx_hash)
- `blockchain_transactions` — TON tx history (tx_hash, status, confirmation)

**Auto-run migrations:**
All `.sql` files in `/migrations` are automatically executed on DB startup via `docker-compose` volume mount.

---

## Configuration

### Environment Variables (.env)
```bash
# Database
POSTGRES_USER=leasing_user
POSTGRES_PASSWORD=secure_pass_change_in_prod
POSTGRES_DB=leasing_db

# Redis
REDIS_HOST=redis
REDIS_PORT=6379

# JWT & Sessions
JWT_SECRET=your-secret-key-change-in-prod
SESSION_SECRET=session-secret-change-in-prod

# Payment Methods
STRIPE_API_KEY=sk_test_...
BANK_API_URL=https://bank-api.example.com
BANK_API_KEY=...

# Blockchain (TON)
TON_API_URL=https://testnet.toncenter.com/api/v2
TON_WALLET_ADDRESS=0:...
TON_PRIVATE_KEY=...
```

### Per-Service Config (YAML)
Each service loads `config/config.yaml` (with env var overrides):
- `core-api/config/core-api.yaml` — Server, DB, Redis, JWT settings
- `lease-service/config/lease-service.yaml` — Server, DB, MeiliSearch
- `payment-service/config/payment-service.yaml` — Payment providers
- `blockchain-service/config/blockchain-service.yaml` — TON settings

---

## Development Workflow

### Adding a New Payment Strategy

1. Create file: `payment-service/internal/strategies/newprovider_strategy.go`
   ```go
   package strategies
   
   type NewProviderStrategy struct { /* config */ }
   
   func (s *NewProviderStrategy) Validate(req *PaymentRequest) error { /* ... */ }
   func (s *NewProviderStrategy) Process(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
       // Implement provider logic
       return &PaymentResponse{...}, nil
   }
   ```

2. Update factory: `payment-service/internal/factory/payment_factory.go`
   ```go
   case "new_provider":
       adapter := adapters.NewNewProviderAdapter(...)
       return strategies.NewNewProviderStrategy(adapter)
   ```

3. Call API:
   ```bash
   curl -X POST http://localhost:3002/payments \
     -H "Content-Type: application/json" \
     -d '{
       "lease_id": "uuid",
       "user_id": "uuid",
       "amount": 1000,
       "currency": "USD",
       "method": "CARD",
       "provider": "new_provider"
     }'
   ```

---

## TON Blockchain Integration

### Current Implementation (Simulation)
- `TONAdapter` mocks blockchain calls
- Useful for testing before real TON credentials

### Production Integration
1. **Set real TON credentials:**
   ```bash
   # .env
   TON_WALLET_ADDRESS=0:your_real_address
   TON_PRIVATE_KEY=your_real_private_key
   TON_API_URL=https://toncenter.com/api/v2  # mainnet or testnet
   ```

2. **Implement TON contract calls:**
   ```go
   // blockchain-service/internal/adapters/ton_adapter.go
   // Replace SendTransaction() with actual TON SDK calls:
   
   import "github.com/xssnick/tonutils-go/tlb"
   import "github.com/xssnick/tonutils-go/address"
   
   func (t *TONAdapter) SendTransaction(ctx context.Context, toAddr, amount string) (*BlockchainTransaction, error) {
       // Sign transaction
       // Send via TON blockchain
       // Return tx hash
   }
   ```

3. **Smart Contract (optional):**
   Write FunC contract to handle lease payments automatically on TON blockchain.

---

## Testing

### Unit Tests
```bash
# core-api
cd core-api && npm test

# lease-service
cd lease-service && go test ./...

# payment-service
cd payment-service && go test ./...

# blockchain-service
cd blockchain-service && go test ./...
```

### Integration Tests (Full Stack)
```bash
docker-compose up -d

# Test auth
curl -X POST http://localhost:3000/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "password": "password123"}'

# Test lease creation
curl -X POST http://localhost:3001/leases \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "uuid",
    "vehicle_id": "uuid",
    "start_date": "2024-01-01",
    "end_date": "2024-12-31",
    "monthly_payment": 500
  }'

# Test payment (triggers blockchain observer)
curl -X POST http://localhost:3002/payments \
  -H "Content-Type: application/json" \
  -d '{
    "lease_id": "uuid",
    "user_id": "uuid",
    "amount": 500,
    "provider": "stripe"
  }'
```

---

## Troubleshooting

### Services won't start
```bash
# Check logs
docker-compose logs [service-name]

# Rebuild images
docker-compose build --no-cache

# Verify env file
cat .env

# Ensure ports are free
lsof -i :3000  # check port 3000
```

### Database connection issues
```bash
# Verify DB is running
docker-compose ps postgres

# Test connection
docker-compose exec postgres psql -U leasing_user -d leasing_db -c "SELECT 1"

# Check migrations applied
docker-compose exec postgres psql -U leasing_user -d leasing_db -c "\dt"
```

### Redis issues
```bash
# Test connection
docker-compose exec redis redis-cli ping
# Should return: PONG

# Monitor pub/sub
docker-compose exec redis redis-cli SUBSCRIBE payments
```

### Payment events not reaching blockchain-service
```bash
# Check blockchain-service logs
docker-compose logs blockchain-service

# Manually publish test event
docker-compose exec redis redis-cli PUBLISH payments '{"event":"payment.completed","payment_id":"test"}'

# Monitor subscriber in blockchain-service logs
```

---

## File Structure

```
car-leasing/
├── .env                              # Environment variables
├── docker-compose.yml                # Orchestration
├── migrations/                        # SQL migrations (auto-run)
│   ├── 001_init.sql                 # Users, audit log
│   ├── 002_leases.sql               # Vehicles, leases, lease_payments
│   └── 003_payments.sql             # Payments, blockchain_transactions
│
├── utils/                            # Shared Go libraries
│   ├── go.mod
│   ├── logger/logger.go             # Zap wrapper
│   ├── config/config.go             # YAML config loader
│   └── redis/redis.go               # Redis client wrapper
│
├── core-api/                         # Fastify API Gateway (Node.js)
│   ├── package.json
│   ├── tsconfig.json
│   └── src/
│       ├── main.ts                  # Bootstrap
│       ├── config/index.ts          # Config loader
│       ├── plugins/{jwt,redis,meilisearch}.ts
│       ├── routes/{auth,leases}.ts
│       ├── controllers/{authController,leaseController}.ts
│       ├── services/authService.ts
│       └── repositories/userRepository.ts
│
├── lease-service/                    # Go Fiber - Leases
│   ├── go.mod
│   ├── config/config.yaml
│   ├── cmd/lease-service/main.go
│   └── internal/
│       ├── dtos/lease_dto.go
│       ├── repositories/lease_repository.go
│       ├── services/lease_service.go
│       ├── adapters/meili_adapter.go
│       └── controllers/lease_controller.go
│
├── payment-service/                  # Go Fiber - Payments
│   ├── go.mod
│   ├── config/config.yaml
│   ├── cmd/payment-service/main.go
│   └── internal/
│       ├── dtos/payment_dto.go
│       ├── strategies/{strategy,stripe_strategy,bank_strategy}.go
│       ├── adapters/bank_adapter.go
│       ├── factory/payment_factory.go
│       ├── repositories/payment_repository.go
│       ├── services/payment_service.go
│       └── controllers/{payment_controller,webhook_controller}.go
│
├── blockchain-service/               # Go Fiber - TON Blockchain
│   ├── go.mod
│   ├── config/config.yaml
│   ├── cmd/blockchain-service/
│   │   ├── main.go                  # Redis subscriber + bootstrap
│   │   └── factory.go               # DI
│   └── internal/
│       ├── dtos/blockchain_dto.go
│       ├── adapters/ton_adapter.go
│       ├── repositories/blockchain_repository.go
│       └── services/blockchain_service.go
│
└── docker/                           # Dockerfiles
    ├── Dockerfile.core-api
    ├── Dockerfile.lease-service
    ├── Dockerfile.payment-service
    └── Dockerfile.blockchain-service
```

---

## Production Checklist

- [ ] Update `JWT_SECRET`, `SESSION_SECRET` in `.env`
- [ ] Update `POSTGRES_PASSWORD` in `.env`
- [ ] Set real `STRIPE_API_KEY`, `BANK_API_URL`, `BANK_API_KEY`
- [ ] Configure TON credentials: `TON_WALLET_ADDRESS`, `TON_PRIVATE_KEY`
- [ ] Use production TON network: `TON_API_URL=https://toncenter.com/api/v2`
- [ ] Enable HTTPS for core-api (nginx reverse proxy recommended)
- [ ] Set up database backups
- [ ] Configure monitoring/alerting for services
- [ ] Test payment workflows end-to-end
- [ ] Load test microservices

---

## Support & Contributing

For issues or contributions:
1. Check logs: `docker-compose logs [service]`
2. Verify config: `.env` and `config/*.yaml`
3. Check database state: psql or MeiliSearch UI
4. Run integration tests

---

**Architecture designed by Senior Software Architect with production-ready patterns: Strategy, Factory, Adapter, Observer, Repository, DTO, Dependency Injection, Configuration Management, and Clean Layering.**
