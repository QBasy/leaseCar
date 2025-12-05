# Copilot Instructions for leaseCar Monorepo

## Quick Overview

This is a **production-ready car leasing microservices monorepo** with:
- **core-api** (Fastify/Node.js): API Gateway with JWT auth, Redis sessions, MeiliSearch aggregation
- **lease-service** (Go Fiber): Lease CRUD, payment schedules, MeiliSearch indexing
- **payment-service** (Go Fiber): Payment strategies (Stripe, Bank), webhooks, event publishing
- **blockchain-service** (Go Fiber): TON blockchain integration, Redis subscriber (Observer pattern)
- **Infrastructure**: PostgreSQL, Redis, MeiliSearch, Docker Compose

**All code is REAL, production-ready, fully implemented with proper error handling and logging.**

---

## Architecture Diagram

```
CLIENT → NGINX → core-api:3000 (Fastify)
                  ├─ /auth → AuthController → AuthService → UserRepository
                  ├─ /api/v1/leases → LeaseController (proxy to lease-service)
                  └─ /api/v1/leases?q=search → MeiliSearch queries

         lease-service:3001 (Go Fiber)
         ├─ POST /leases → LeaseService → MeiliAdapter (index async)
         └─ GET /leases/:id → LeaseRepository

         payment-service:3002 (Go Fiber)
         ├─ POST /payments → PaymentService
         │  ├─ PaymentFactory.GetStrategy(provider)
         │  ├─ Strategy.Process() [StripeStrategy, BankStrategy]
         │  └─ Publish to Redis "payments" channel
         └─ POST /webhooks/:provider → WebhookController

         blockchain-service:3003 (Go Fiber)
         ├─ Redis subscriber on "payments" channel (Observer)
         ├─ BlockchainService.ProcessPaymentEvent()
         ├─ TONAdapter.SendTransaction() → TON blockchain
         └─ Background polling for confirmations
```

---

## Critical Patterns & Rules (MUST FOLLOW)

### 1. Fastify (core-api)
- **Plugin Architecture:** All integrations (JWT, Redis, MeiliSearch) are in `/plugins`
- **Dependency Injection:** Constructor-based in controllers/services
- **Clean Layering:** routes → controllers → services → repositories → DB
- **Request Validation:** Zod schemas in routes before passing to controllers
- **Example Route:**
  ```typescript
  app.post<{ Body: LoginSchema }>('/auth/login', 
    { schema: { body: loginSchema } },
    async (req, res) => authController.login(req, res)
  );
  ```

### 2. Go Fiber Microservices
- **File Structure:** `/cmd/service/main.go`, `/internal/{controllers,services,repositories,adapters,strategies,dtos}`, `/pkg`, `/config/config.yaml`
- **Strategy Pattern:** `PaymentStrategy` interface for multiple payment providers
- **Adapter Pattern:** `BankAdapter`, `TONAdapter` for external systems
- **Observer Pattern:** Redis Pub/Sub for inter-service events
- **Factory Pattern:** `PaymentFactory` selects strategies by provider
- **Repository Pattern:** Centralize all DB operations
- **DTO Layer:** All APIs validated via structs (no direct models)
- **Config:** YAML → Go structs using `viper` library
- **Logging:** Use `leaseCar/utils/logger` (Zap wrapper)

### 3. Microservice Boundaries (DO NOT VIOLATE)
```
core-api ONLY: Auth, routing, MeiliSearch aggregation
lease-service ONLY: Lease CRUD, MeiliSearch indexing
payment-service ONLY: Payment processing, strategy execution
blockchain-service ONLY: Listen to Redis, execute blockchain calls
```
**Cross-service calls:** Use HTTP REST (not direct DB queries) or Redis Pub/Sub

### 4. Redis Communication
- **Payment-Service publishes:** `PUBLISH payments '{"event":"payment.completed","payment_id":"uuid",...}'`
- **Blockchain-Service subscribes:** `redis.Subscribe("payments")` (in main.go goroutine)
- **No direct blocking calls:** Use channels/goroutines for async processing

### 5. Database Schema
- **Migrations:** SQL files in `/migrations` auto-run on Docker startup
- **Key tables:** `users`, `vehicles`, `leases`, `payments`, `blockchain_transactions`
- **Indexes:** Already created on status, created_at, tx_hash columns
- **DTO → DB mapping:** Services handle conversion (explicit, type-safe)

---

## Environment Variables

**All services read `.env` file at root `car-leasing/`**

```bash
# Database
POSTGRES_USER=leasing_user
POSTGRES_PASSWORD=secure_pass_change_in_prod
POSTGRES_DB=leasing_db

# Redis
REDIS_HOST=redis
REDIS_PORT=6379

# JWT & Sessions (core-api)
JWT_SECRET=your-secret-key-change-in-prod
SESSION_SECRET=session-secret-change-in-prod

# Payment Methods (payment-service)
STRIPE_API_KEY=sk_test_...
BANK_API_URL=https://bank-api.example.com
BANK_API_KEY=...

# Blockchain (blockchain-service)
TON_API_URL=https://testnet.toncenter.com/api/v2
TON_WALLET_ADDRESS=0:...
TON_PRIVATE_KEY=...
```

---

## Key Files & Patterns Reference

### Core-API Structure
```
core-api/
├── src/main.ts              # Bootstrap: register plugins + routes
├── src/config/index.ts      # YAML config loader with env overrides
├── src/plugins/
│   ├── jwt.ts              # JWT plugin, decorate app.authenticate
│   ├── redis.ts            # Redis plugin, decorate app.redis
│   └── meilisearch.ts      # MeiliSearch plugin, decorate app.meili
├── src/routes/
│   ├── auth.ts             # POST /auth/login
│   └── leases.ts           # GET/POST /api/v1/leases
├── src/controllers/
│   ├── authController.ts   # Login handler, validate input
│   └── leaseController.ts  # Search via MeiliSearch, proxy to lease-service
├── src/services/
│   └── authService.ts      # JWT signing, password verification
└── src/repositories/
    └── userRepository.ts   # Query users from PostgreSQL
```

### Go Microservice Structure (lease-service example)
```
lease-service/
├── cmd/lease-service/main.go      # Bootstrap: setup DB, Redis, Fiber, routes
├── internal/
│   ├── dtos/lease_dto.go           # LeaseCreateRequest, Lease structs
│   ├── repositories/
│   │   └── lease_repository.go    # Create, GetByID, Search queries
│   ├── services/
│   │   └── lease_service.go       # Business logic: Create, GetByID, Search + MeiliSearch indexing
│   ├── adapters/
│   │   └── meili_adapter.go       # MeiliSearch client wrapper
│   └── controllers/
│       └── lease_controller.go    # HTTP handlers: Create, GetByID, Search
├── config/config.yaml              # Server, DB, Redis, MeiliSearch settings
└── go.mod                           # Module dependencies + local replace for utils
```

---

## Adding a New Payment Strategy (Step-by-Step)

1. **Create strategy file:** `payment-service/internal/strategies/newprovider_strategy.go`
   ```go
   type NewProviderStrategy struct { /* config */ }
   func (s *NewProviderStrategy) Validate(req *PaymentRequest) error { /* ... */ }
   func (s *NewProviderStrategy) Process(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
       // Call external API
       // Return response with provider_tx_id
   }
   ```

2. **Register in factory:** `payment-service/internal/factory/payment_factory.go`
   ```go
   case "new_provider":
       return NewNewProviderStrategy(...)
   ```

3. **Add env vars:** Update `.env` with provider credentials

4. **Test:** `curl -X POST http://localhost:3002/payments -d '{"provider":"new_provider",...}'`

---

## Testing Strategy
- **core-api:** Jest unit tests for routes + integration tests with mock Redis
- **Go services:** Go `testing` package + table-driven tests
- **Integration:** Docker Compose full stack test
- **Run:** `npm test` (Node.js), `go test ./...` (Go services)

---

## Build & Run Commands

```bash
# Full stack in Docker
docker-compose up -d

# Check health
curl http://localhost:3000/health
curl http://localhost:3001/health
curl http://localhost:3002/health
curl http://localhost:3003/health

# View logs
docker-compose logs -f [service-name]

# Local dev (requires Docker for DB/Redis/MeiliSearch)
cd core-api && npm run dev
cd lease-service && go run ./cmd/lease-service
```

---

## Common Mistakes to AVOID

1. ❌ **Don't call blockchain-service from core-api directly** → Go through payment-service + Redis
2. ❌ **Don't bypass MeiliSearch** → Always index leases after updates (async is fine)
3. ❌ **Don't hardcode credentials** → Use env vars everywhere
4. ❌ **Don't skip error handling in payments** → Always validate, return proper HTTP codes
5. ❌ **Don't make blocking calls in blockchain-service** → Use goroutines + channels
6. ❌ **Don't violate microservice boundaries** → Each service owns its DB tables
7. ❌ **Don't query another service's DB directly** → Use REST/Pub-Sub

---

## File Structure Overview

```
car-leasing/
├── .env                              # Environment variables (SHARED)
├── docker-compose.yml                # Full stack orchestration
├── README.md                          # Quick start + architecture
├── HOW_TO_RUN.md                      # Detailed deployment guide
├── ARCHITECTURE.md                    # Design patterns, flows, diagrams
├── EXTENDING.md                       # How to add payment strategies & TON
│
├── migrations/
│   ├── 001_init.sql                  # Users, audit log
│   ├── 002_leases.sql                # Vehicles, leases, lease_payments
│   └── 003_payments.sql              # Payments, blockchain_transactions
│
├── utils/                             # Shared Go libraries
│   ├── go.mod                         # Shared module
│   ├── logger/logger.go               # Zap wrapper
│   ├── config/config.go               # YAML → struct loader
│   └── redis/redis.go                 # Redis client wrapper
│
├── core-api/                          # Fastify API Gateway (Node.js)
│   ├── package.json                   # npm dependencies
│   ├── tsconfig.json                  # TypeScript config
│   └── src/                           # Full structure (see above)
│
├── lease-service/                     # Go Fiber Leases
│   ├── go.mod                         # Go dependencies
│   ├── cmd/lease-service/main.go      # Bootstrap
│   ├── config/config.yaml             # Service config
│   └── internal/                      # Full structure (see above)
│
├── payment-service/                   # Go Fiber Payments
│   ├── go.mod
│   ├── cmd/payment-service/main.go
│   ├── config/config.yaml
│   ├── internal/
│   │   ├── strategies/                # Strategy implementations
│   │   ├── adapters/                  # External API adapters
│   │   ├── factory/                   # Strategy factory
│   │   └── ...
│
├── blockchain-service/                # Go Fiber Blockchain
│   ├── go.mod
│   ├── cmd/blockchain-service/        # Main + factory (DI)
│   ├── config/config.yaml
│   ├── internal/
│   │   ├── adapters/ton_adapter.go    # TON blockchain integration
│   │   ├── services/blockchain_service.go  # Observer + polling
│   │   └── ...
│
└── docker/                            # Dockerfiles
    ├── Dockerfile.core-api            # Node.js build + runtime
    ├── Dockerfile.lease-service       # Go build + Alpine runtime
    ├── Dockerfile.payment-service
    └── Dockerfile.blockchain-service
```

---

## When Adding New Features

1. **Start with DTOs:** Define request/response structures first
2. **Add repository methods:** Implement DB queries
3. **Implement service logic:** Business rules, validation
4. **Create controller/handler:** HTTP endpoints
5. **Register routes:** Wire up in main.go / routes
6. **Add logging:** Use logger package (Go) or app.log (Node.js)
7. **Add error handling:** Return proper HTTP status codes + error messages
8. **Add tests:** Unit test service, integration test full flow

---

## Performance Tips

- **Database:** Connection pooling enabled, indexes on status/created_at/tx_hash
- **Redis:** Pub/Sub for async events, connection reuse
- **MeiliSearch:** Indexing runs in background goroutine (non-blocking)
- **Blockchain:** Polling uses exponential backoff, batch checks for efficiency

---

## You are an expert Senior Software Architect & Principal Engineer.

Your task is to maintain and extend this **complete production-ready monorepo** for a car leasing service, following the rules above.

─────────────────────────────
ARCHITECTURE
─────────────────────────────
- Single-machine deployment using docker-compose
- All services must be lightweight and open-source only
- Tech stack:
  * Fastify (Node.js) → API Gateway (core service)
  * Go Fiber → microservices (lease-service, payment-service, blockchain-service)
  * Redis → sessions + cache
  * PostgreSQL → relational DB
  * MeiliSearch → full text search
  * MinIO or LocalFS → file storage
  * TON blockchain smart contracts (FunC, NOT Solidity)

─────────────────────────────
MICROSERVICE STRUCTURE
─────────────────────────────
- **Go services (Fiber)** must have standard structure:
/cmd/<service_name>/main.go
/internal/
/controllers
/services
/repositories
/adapters
/strategies
/dtos
/utils
/pkg/ (shared libraries)
/config/config.yaml

- **Fastify core-api** structure:
core-api/
main.ts
routes/
controllers/
services/
repositories/
plugins/
config/config.yaml

─────────────────────────────
DESIGN PATTERNS
─────────────────────────────
- Fastify:
- Plugin architecture
- Decorators
- Dependency Injection
- Request/Response validation
- Clean layering: routes → controllers → services → repositories

- Fiber microservices:
- Strategy (payment processors)
- Adapter (bank API integration)
- Observer (payment → blockchain)
- Factory (payment providers)
- Repository (PostgreSQL)
- DTO layer
- Config loader (.yaml → struct)
- Zap logger utility

─────────────────────────────
MICROSERVICE BOUNDARIES
─────────────────────────────
- core-api:
- Auth (JWT + Redis sessions)
- Routing & aggregation
- API versioning
- MeiliSearch queries

- lease-service:
- Lease CRUD
- Payment schedule generator
- Index leases to MeiliSearch

- payment-service:
- PaymentStrategy + PaymentAdapter
- Webhooks
- Event emitter to blockchain
- State machine

- blockchain-service:
- TON contracts interaction
- Return tx hash

─────────────────────────────
DELIVERABLES
─────────────────────────────
1. Full monorepo FS tree
2. Full docker-compose with healthchecks
3. Dockerfiles for each service
4. All configs in .yaml
5. All utilities:
 - Zap logger wrapper
 - Config loader
 - Redis client wrapper
6. Full working code for:
 - Fastify API Gateway
 - Fiber microservices
 - All patterns (Strategy/Adapter/Factory/Observer)
 - Full migrations
 - TON integration code structure
7. Documentation:
 - how-to-run
 - architecture overview
 - extending payment strategies
 - TON integration flow

─────────────────────────────
CODING RULES
─────────────────────────────
- Always generate REAL code (no pseudocode)
- Always include imports
- Always use env variables
- Clear error handling
- Production-ready defaults
- Multi-step generation
- Generate one service at a time
- Wait for my approval after each step

─────────────────────────────
INSTRUCTIONS FOR YOU (AI)
─────────────────────────────
1. Start with generating **core-api**: main.ts, routes, controllers, services, repositories, plugins, config.
2. After approval, generate **lease-service** with proper Go structure (/cmd, /internal, /pkg).
3. Then **payment-service**, then **blockchain-service**.
4. Then **docker-compose.yml** and Dockerfiles for all services.
5. Then **utils** (Zap logger, config loader, Redis wrapper).
6. Then **documentation**.

- At each step, output **full file content with path**, a **brief explanation of what it does**, and a **file tree snapshot**.
- Wait for user approval before continuing to the next step.
- Include env variables for DB, Redis, MeiliSearch, MinIO, and TON SDK.

─────────────────────────────
GOAL
─────────────────────────────
Generate the full project step-by-step with explanations, file trees, and full files, strictly following this prompt.
