# Architecture Overview

## System Architecture Diagram

```
┌──────────────────────────────────────────────────────────────────────────────┐
│                              CLIENT APPLICATIONS                              │
│                    (Web, Mobile, Third-party Integrations)                    │
└────────────────────────────────────┬─────────────────────────────────────────┘
                                     │ HTTPS
                    ┌────────────────▼─────────────────┐
                    │                                  │
                    │  NGINX (Reverse Proxy)           │
                    │  - SSL/TLS Termination           │
                    │  - Load Balancing                │
                    │  - Rate Limiting                 │
                    │                                  │
                    └────────────────┬─────────────────┘
                                     │
                ┌────────────────────┴────────────────────┐
                │                                         │
    ┌───────────▼──────────────┐         ┌────────────────▼────────────┐
    │                          │         │                             │
    │   CORE-API (Fastify)     │         │  LEASE-SERVICE (Fiber)      │
    │   Node.js - Port 3000    │         │  Go - Port 3001             │
    │                          │         │                             │
    │ ┌──────────────────────┐ │         │ ┌─────────────────────────┐ │
    │ │ Routes               │ │         │ │ Controllers             │ │
    │ │  /auth/login         │ │         │ │  POST /leases           │ │
    │ │  /api/v1/leases      │ │         │ │  GET  /leases/:id       │ │
    │ │  /api/v1/leases/:id  │ │         │ │  GET  /leases?q=search  │ │
    │ └──────────┬───────────┘ │         │ └─────────────┬───────────┘ │
    │            │             │         │               │             │
    │ ┌──────────▼───────────┐ │         │ ┌─────────────▼───────────┐ │
    │ │ Controllers          │ │         │ │ Services                │ │
    │ │  AuthController      │ │         │ │  LeaseService           │ │
    │ │  LeaseController     │ │         │ │   - Create lease        │ │
    │ └──────────┬───────────┘ │         │ │   - Index to MeiliSearch│ │
    │            │             │         │ │   - Search              │ │
    │ ┌──────────▼───────────┐ │         │ └─────────────┬───────────┘ │
    │ │ Services             │ │         │               │             │
    │ │  AuthService         │ │         │ ┌─────────────▼───────────┐ │
    │ │  - JWT signing       │ │         │ │ Adapters                │ │
    │ │  - Redis sessions    │ │         │ │  MeiliAdapter           │ │
    │ └──────────┬───────────┘ │         │ │   - Index/Search        │ │
    │            │             │         │ └─────────────┬───────────┘ │
    │ ┌──────────▼───────────┐ │         │               │             │
    │ │ Repositories         │ │         │ ┌─────────────▼───────────┐ │
    │ │  UserRepository      │ │         │ │ Repositories            │ │
    │ │   - Query users      │ │         │ │  LeaseRepository        │ │
    │ └──────────┬───────────┘ │         │ │   - CRUD operations     │ │
    │            │             │         │ └─────────────┬───────────┘ │
    │ ┌──────────▼───────────┐ │         │               │             │
    │ │ Plugins              │ │         │ Connected to: │             │
    │ │  JWT Auth            │ │         │   ✓ PostgreSQL│             │
    │ │  Redis Sessions      │ │         │   ✓ Redis     │             │
    │ │  MeiliSearch Client  │ │         │   ✓ MeiliSearch          │ │
    │ └──────────────────────┘ │         └─────────────────────────┘ │
    │                          │                                      │
    │ Connected to:            │         ┌────────────────────────────┐
    │  ✓ PostgreSQL            │         │                            │
    │  ✓ Redis                 │         │ PAYMENT-SERVICE (Fiber)    │
    │  ✓ MeiliSearch           │         │ Go - Port 3002             │
    │                          │         │                            │
    └──────────────────────────┘         │ ┌──────────────────────────┤
                                         │ │ Controllers              │
                                         │ │  POST /payments          │
                                         │ │  POST /webhooks/:provider│
                                         │ └─────────────┬────────────┤
                                         │               │            │
                                         │ ┌─────────────▼────────────┤
                                         │ │ Services                 │
                                         │ │  PaymentService          │
                                         │ │   - Process payment      │
                                         │ │   - Emit events to Redis │
                                         │ └─────────────┬────────────┤
                                         │               │            │
                                         │ ┌─────────────▼────────────┤
                                         │ │ Factory & Strategies     │
                                         │ │  PaymentFactory          │
                                         │ │  - StripeStrategy        │
                                         │ │  - BankStrategy          │
                                         │ │  (extensible)            │
                                         │ └─────────────┬────────────┤
                                         │               │            │
                                         │ ┌─────────────▼────────────┤
                                         │ │ Adapters                 │
                                         │ │  BankAdapter             │
                                         │ │   - Call Bank API        │
                                         │ └─────────────┬────────────┤
                                         │               │            │
                                         │ ┌─────────────▼────────────┤
                                         │ │ Repositories             │
                                         │ │  PaymentRepository       │
                                         │ │   - Persist payments     │
                                         │ └──────────────────────────┤
                                         │                            │
                                         │ Connected to:              │
                                         │  ✓ PostgreSQL              │
                                         │  ✓ Redis (publisher)       │
                                         │                            │
                                         └────────────────────────────┘
                                                   │
                                  ┌────────────────┘
                                  │ Redis PUB/SUB
                                  │ Channel: "payments"
                                  │ Event: payment.completed
                                  │
                         ┌────────▼──────────────────────┐
                         │                               │
                         │ BLOCKCHAIN-SERVICE (Fiber)    │
                         │ Go - Port 3003                │
                         │                               │
                         │ ┌──────────────────────────┐  │
                         │ │ Redis Subscriber         │  │
                         │ │  - Listen payments       │  │
                         │ │  - Parse events          │  │
                         │ └──────────┬───────────────┘  │
                         │            │                  │
                         │ ┌──────────▼───────────────┐  │
                         │ │ Services                 │  │
                         │ │  BlockchainService       │  │
                         │ │   - Process events       │  │
                         │ │   - Poll confirmations   │  │
                         │ └──────────┬───────────────┘  │
                         │            │                  │
                         │ ┌──────────▼───────────────┐  │
                         │ │ Adapters                 │  │
                         │ │  TONAdapter              │  │
                         │ │   - Send to TON          │  │
                         │ │   - Check status         │  │
                         │ └──────────┬───────────────┘  │
                         │            │                  │
                         │ ┌──────────▼───────────────┐  │
                         │ │ Repositories             │  │
                         │ │  BlockchainRepository    │  │
                         │ │   - Save tx records      │  │
                         │ └──────────────────────────┘  │
                         │                               │
                         │ Connected to:                 │
                         │  ✓ Redis (subscriber)         │
                         │  ✓ PostgreSQL                 │
                         │  ✓ TON Blockchain             │
                         │  ✓ TON API                    │
                         │                               │
                         └───────────────────────────────┘
```

---

## Data Flow

### 1. Lease Creation Flow

```
Client
  ├── POST /api/v1/leases (via core-api)
  │   ├── Forward to lease-service
  │   │   ├── Validate input (DTO)
  │   │   ├── Call LeaseService.Create()
  │   │   │   ├── LeaseRepository.Create() → PostgreSQL
  │   │   │   │   └── INSERT INTO leases (user_id, vehicle_id, ...)
  │   │   │   └── Async: MeiliAdapter.IndexLease() → MeiliSearch
  │   │   │       └── Add to "leases" index (background)
  │   │   └── Return lease_id
  │   └── Core-API proxies response
  │
  └── Client receives: {"id": "uuid"}
```

### 2. Lease Search Flow

```
Client
  ├── GET /api/v1/leases?q=luxury (via core-api)
  │   ├── Core-API receives request
  │   ├── Query MeiliSearch.index("leases").search("luxury", {limit: 20})
  │   │   └── MeiliSearch returns matching leases (instant, full-text)
  │   └── Return hits to client
  │
  └── Client receives: [{id, user_id, vehicle_id, ...}]
```

### 3. Payment Processing Flow (Synchronous)

```
Client
  ├── POST /payments {lease_id, user_id, amount, provider: "stripe"}
  │   ├── Payment-Service receives
  │   ├── PaymentRepository.Create() → INSERT INTO payments (status=PENDING)
  │   ├── PaymentFactory.GetStrategy("stripe")
  │   │   └── Returns StripeStrategy instance
  │   ├── StripeStrategy.Validate() + .Process()
  │   │   └── Call Stripe API (simulated)
  │   ├── PaymentRepository.UpdateStatus(id, "COMPLETED", "stripe_tx_xxx")
  │   │   └── UPDATE payments SET status=COMPLETED, transaction_id=...
  │   └── Return PaymentResponse {payment_id, status, provider_tx_id}
  │
  └── Client receives: {"payment_id": "uuid", "status": "COMPLETED", ...}
```

### 4. Blockchain Integration Flow (Asynchronous / Observer)

```
Payment-Service (publishes event after processing)
  │
  ├── After successful payment
  ├── Publish to Redis: PUBLISH payments '{"event":"payment.completed","payment_id":"uuid","provider_tx":"stripe_tx_xxx"}'
  │
  └── Redis Channel "payments"
      │
      └── Blockchain-Service (subscriber)
          ├── Receive: payment.completed event
          ├── Parse PaymentEventPayload
          ├── BlockchainService.ProcessPaymentEvent()
          │   ├── TONAdapter.SendTransaction(recipient_addr, amount)
          │   │   └── Call TON API
          │   │       └── TON blockchain returns tx_hash
          │   ├── BlockchainRepository.SaveTransaction()
          │   │   └── INSERT INTO blockchain_transactions (tx_hash, status=SUBMITTED, ...)
          │   ├── BlockchainRepository.UpdatePaymentTxHash()
          │   │   └── UPDATE payments SET blockchain_tx_hash=...
          │   └── Start background polling goroutine
          │
          └── Background: pollConfirmation()
              ├── Loop every 5 seconds
              ├── TONAdapter.CheckStatus(tx_hash)
              │   └── Query TON API for tx status
              ├── When status == CONFIRMED
              │   └── BlockchainRepository.UpdateTransactionStatus()
              │       └── UPDATE blockchain_transactions SET confirmed=true
              └── Done
```

---

## Design Patterns Used

### 1. Strategy Pattern (Payment-Service)

**Problem:** Multiple payment providers with different APIs

**Solution:** Define `PaymentStrategy` interface with `Process()` and `Validate()` methods
- Each provider (Stripe, Bank, ApplePay) implements the interface
- Strategies are interchangeable at runtime
- New providers can be added without modifying core logic

**Example:**
```go
type PaymentStrategy interface {
    Process(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error)
    Validate(req *PaymentRequest) error
}

// Implementations:
type StripeStrategy struct { ... }
type BankStrategy struct { ... }
```

### 2. Factory Pattern (Payment-Service)

**Problem:** How to select correct strategy based on provider string?

**Solution:** Use `PaymentFactory.GetStrategy(provider string)` to return concrete strategy
- Encapsulates strategy instantiation
- Centralizes provider-specific configuration
- Enables easy addition of new providers

**Example:**
```go
func (f *PaymentFactory) GetStrategy(provider string) PaymentStrategy {
    case "stripe":
        return NewStripeStrategy(apiKey)
    case "bank_api":
        return NewBankStrategy(adapter)
}
```

### 3. Adapter Pattern (Payment-Service & Blockchain-Service)

**Problem:** Integrating external systems (Bank API, TON blockchain) without coupling to them

**Solution:** Create adapters that abstract external calls
- `BankAdapter` wraps Bank API calls
- `TONAdapter` wraps TON blockchain calls
- Services call adapters, not external systems directly
- Easy to mock for testing

**Example:**
```go
type BankAdapter struct { url, apiKey }
func (b *BankAdapter) SendPayment(req *PaymentRequest) (*BankResponse, error)

type TONAdapter struct { apiUrl, walletAddress, privateKey }
func (t *TONAdapter) SendTransaction(ctx, toAddr, amount) (*BlockchainTransaction, error)
```

### 4. Observer Pattern (Blockchain-Service)

**Problem:** Blockchain-Service needs to react to payment completion events from Payment-Service

**Solution:** Use Redis Pub/Sub for event-driven communication
- Payment-Service publishes to Redis channel "payments"
- Blockchain-Service subscribes (Observer) and listens for events
- Completely decoupled services via async messaging

**Example:**
```go
// Payment-Service publishes:
redis.Publish("payments", '{"event":"payment.completed","payment_id":"uuid"}')

// Blockchain-Service subscribes:
pubsub := redis.Subscribe("payments")
for msg := range pubsub.Channel() {
    svc.ProcessPaymentEvent(msg.Payload)
}
```

### 5. Repository Pattern (All Services)

**Problem:** Separate data access logic from business logic

**Solution:** Each entity has a Repository (UserRepository, LeaseRepository, etc.)
- Repositories handle all DB operations
- Services depend on repositories (abstraction)
- Easy to mock for testing
- DB queries are centralized and testable

**Example:**
```go
type LeaseRepository struct { pool *pgxpool.Pool }
func (r *LeaseRepository) Create(ctx, lease) (string, error)
func (r *LeaseRepository) GetByID(ctx, id) (*Lease, error)
```

### 6. DTO Layer (All Services)

**Problem:** Decoupling request/response formats from internal models

**Solution:** Define DTOs for each API endpoint
- Request validation (Zod in Node.js, structs in Go)
- Type safety
- API contract clarity
- Easier to evolve APIs

**Example:**
```go
type PaymentRequest struct {
    LeaseID string `json:"lease_id"`
    Amount float64 `json:"amount"`
    Provider string `json:"provider"`
}
```

### 7. Dependency Injection (All Services)

**Problem:** Services depend on many other services/repositories

**Solution:** Pass dependencies via constructors
- Loose coupling
- Easy to test (swap real with mock)
- Clear service dependencies

**Example:**
```go
type PaymentService struct {
    repo *PaymentRepository
    factory *PaymentFactory
}
func NewPaymentService(repo, factory) *PaymentService { ... }
```

### 8. Plugin Architecture (Core-API / Fastify)

**Problem:** Decoupling Fastify setup (JWT, Redis, MeiliSearch) from routes

**Solution:** Fastify plugins for each concern
- JWT plugin registers auth decorator
- Redis plugin decorates Fastify with redis client
- MeiliSearch plugin decorates Fastify with meili client
- Reusable and maintainable

**Example:**
```typescript
// Plugins registered in main.ts:
app.register(jwtPlugin)
app.register(redisPlugin)
app.register(meiliPlugin)

// Available in routes/controllers:
app.meili.index("leases").search(q)
```

---

## Microservice Boundaries

### Core-API Responsibilities
✓ User authentication (JWT)
✓ Request routing and aggregation
✓ API versioning
✓ MeiliSearch querying
✗ Payment processing (delegates to payment-service)
✗ Lease CRUD (proxies to lease-service)

### Lease-Service Responsibilities
✓ Lease CRUD operations
✓ Payment schedule generation (future)
✓ MeiliSearch indexing
✗ Payment processing
✗ Blockchain operations

### Payment-Service Responsibilities
✓ Payment processing via multiple strategies
✓ Webhook handling from providers
✓ Publishing events to Redis (Observer)
✗ Blockchain operations (handled by blockchain-service)
✗ User management

### Blockchain-Service Responsibilities
✓ Listening to payment events (Redis subscriber)
✓ TON blockchain interactions
✓ Transaction confirmation polling
✗ Payment processing
✗ Lease management

---

## Scalability Considerations

### Horizontal Scaling
- **Core-API:** Stateless (except Redis sessions), can run multiple replicas behind load balancer
- **Lease-Service:** Stateless, can run multiple replicas
- **Payment-Service:** Stateless, can run multiple replicas (careful with webhook idempotency)
- **Blockchain-Service:** Single instance (single Redis subscriber group, or use Redis Streams for fan-out)

### Database Optimization
- **Connection pooling:** PgBouncer recommended for production (100+ connections)
- **Indexing:** Already created on status, created_at, tx_hash columns
- **Partitioning:** Consider monthly partitions for large `payments` table (billions of records)
- **Read replicas:** PostgreSQL streaming replication for read-only analytics

### Caching Strategy
- **Redis:** Sessions, cached lease data, payment status
- **MeiliSearch:** Full-text search results (already cached in index)
- **HTTP cache headers:** For stable data (vehicles, lease templates)

### Asynchronous Processing
- **Payment webhooks:** Queue (Bull/BullMQ) for reliable webhook processing
- **MeiliSearch indexing:** Already async (background goroutine in lease-service)
- **Blockchain polling:** Async goroutines with exponential backoff

---

## Security Considerations

### Authentication & Authorization
✓ JWT tokens with expiration (3600 seconds default)
✓ Redis session storage
✗ TODO: Role-based access control (RBAC)
✗ TODO: OAuth2 integration

### Data Protection
✓ Database passwords in environment variables
✓ JWT secret in environment variables
✓ TON private keys in environment variables (rotate regularly)
✗ TODO: Encryption at rest (PostgreSQL pgcrypto extension)
✗ TODO: TLS for Redis (production)

### API Security
✓ HTTPS (behind nginx reverse proxy in production)
✓ Request validation (Zod, Go struct tags)
✓ Rate limiting (TODO: implement in middleware)
✗ TODO: CORS configuration
✗ TODO: API key for service-to-service calls

### Payment Security
✓ PCI DSS considerations (never store card data)
✓ Token-based payment (Stripe tokens, Apple tokens)
✓ Webhook signature verification (TODO: implement per-provider)

---

## Monitoring & Observability

### Logging
- **Zap logger (Go):** Structured logs with fields
- **Fastify logger (Node.js):** Request/response logging
- **Log aggregation:** Ship to ELK or CloudWatch (TODO)

### Metrics
- **Prometheus:** Expose metrics on `/metrics` endpoints (TODO)
- **Key metrics:**
  - Payment success rate
  - Transaction confirmation time
  - Database query latency
  - Redis pub/sub lag

### Health Checks
- `/health` endpoints on all services
- Liveness: Service is running
- Readiness: Service can process requests (DB connected, Redis reachable)

### Tracing (TODO)
- OpenTelemetry for distributed tracing
- Trace payment flow across services
- Identify bottlenecks

---

## Deployment Architecture

### Development
```
Localhost
├── Docker Compose (all services + infra)
├── Port 3000: core-api
├── Port 3001: lease-service
├── Port 3002: payment-service
├── Port 3003: blockchain-service
├── Port 5432: PostgreSQL
├── Port 6379: Redis
└── Port 7700: MeiliSearch
```

### Staging
```
AWS/GCP/Azure
├── ECS/GKE Cluster
├── RDS PostgreSQL (multi-AZ)
├── ElastiCache Redis (replicated)
├── Managed Elasticsearch (logs)
├── CloudFront (CDN)
└── ACM SSL Certificate
```

### Production
```
Multi-region AWS/GCP/Azure
├── Primary Region:
│   ├── ECS/K8s with auto-scaling
│   ├── RDS PostgreSQL (multi-AZ with read replicas)
│   ├── ElastiCache Redis (clustered, high availability)
│   ├── MeiliSearch (self-hosted or managed)
│   ├── S3/GCS (file storage for leases/invoices)
│   └── CloudWatch/Stackdriver (monitoring)
├── Secondary Region (warm standby):
│   └── Similar setup for disaster recovery
└── Global:
    ├── Route 53 / Google Cloud DNS (failover)
    ├── CloudFlare (DDoS protection, WAF)
    └── PagerDuty (alerting)
```

---

## Future Enhancements

1. **API Versioning:** Implement `/v2` endpoints alongside `/v1` for backward compatibility
2. **Webhooks:** Implement reliable webhook delivery (signed, retries, exponential backoff)
3. **Reporting:** Add analytics dashboard (leases, revenue, payment success rates)
4. **Notifications:** Email/SMS for lease expiration, payment reminders
5. **Multi-tenancy:** Support multiple car leasing companies in single deployment
6. **Smart Contracts:** Move lease logic to TON smart contract (fully on-chain)
7. **Mobile App:** React Native client with offline support

---

**Architecture designed for production with scalability, reliability, and maintainability in mind.**
