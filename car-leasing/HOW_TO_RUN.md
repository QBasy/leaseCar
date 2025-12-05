# How to Run — Complete Guide

## Local Development Setup

### Prerequisites
- **Docker & Docker Compose** (https://www.docker.com/products/docker-desktop)
- **Go 1.21** (for local Go service development)
- **Node.js 20** (for local core-api development)
- **PostgreSQL client** (optional, for manual DB access)

### Quick Start (5 minutes)

```bash
# 1. Clone repository
git clone <repo-url>
cd car-leasing

# 2. Review environment
cat .env
# Modify JWT_SECRET, POSTGRES_PASSWORD if needed

# 3. Start full stack
docker-compose up -d

# 4. Wait for services to be healthy (~30s)
docker-compose ps

# 5. Test health endpoints
curl http://localhost:3000/health
curl http://localhost:3001/health
curl http://localhost:3002/health
curl http://localhost:3003/health

# All should return: {"status":"ok"}
```

### Detailed Step-by-Step

#### Step 1: Start Infrastructure Only
```bash
docker-compose up -d postgres redis meilisearch
sleep 10  # Wait for DB to be ready
```

Check:
```bash
# PostgreSQL
docker-compose exec postgres psql -U leasing_user -d leasing_db -c "SELECT 1"
# Should return: 1

# Redis
docker-compose exec redis redis-cli ping
# Should return: PONG

# MeiliSearch
curl http://localhost:7700/health
# Should return: {"status":"available"}
```

#### Step 2: Start Microservices
```bash
docker-compose up -d core-api lease-service payment-service blockchain-service
sleep 15  # Wait for services to boot
```

Check:
```bash
docker-compose ps
# All should show "Up"

docker-compose logs core-api | tail -20
# Should see: "core-api listening on 3000"
```

#### Step 3: Verify Databases
```bash
docker-compose exec postgres psql -U leasing_user -d leasing_db << EOF
\dt+
EOF
```

Should show tables:
- `users`
- `vehicles`
- `leases`
- `lease_payments`
- `payments`
- `payment_webhooks`
- `blockchain_transactions`

---

## Local Development (Without Docker)

### Option A: Run Individual Services Locally

#### Core API (Node.js)
```bash
cd core-api

# Install dependencies
npm install

# Set environment (or use .env)
export JWT_SECRET=dev-secret
export SESSION_SECRET=dev-session
export REDIS_HOST=localhost
export POSTGRES_HOST=localhost

# Start dev server
npm run dev
# Listens on: http://localhost:3000
```

#### Lease Service (Go)
```bash
cd lease-service

# Download dependencies
go mod download
go mod tidy

# Set environment
export POSTGRES_HOST=localhost
export REDIS_HOST=localhost

# Run
go run ./cmd/lease-service
# Listens on: http://localhost:3001
```

#### Payment Service (Go)
```bash
cd payment-service

go mod download
go mod tidy

export POSTGRES_HOST=localhost
export REDIS_HOST=localhost

go run ./cmd/payment-service
# Listens on: http://localhost:3002
```

#### Blockchain Service (Go)
```bash
cd blockchain-service

go mod download
go mod tidy

export REDIS_HOST=localhost
export POSTGRES_HOST=localhost

go run ./cmd/blockchain-service
# Listens on: http://localhost:3003
# Subscribes to Redis "payments" channel
```

**Note:** Requires Docker running for PostgreSQL, Redis, MeiliSearch.

---

## Testing the APIs

### 1. Authentication

```bash
# Request JWT token
curl -X POST http://localhost:3000/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "testpass123"
  }'

# Response:
# {"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."}

# Save token
TOKEN="eyJ..."

# Use token in subsequent requests
curl -H "Authorization: Bearer $TOKEN" http://localhost:3000/api/v1/leases
```

### 2. Lease Management

#### Create Lease
```bash
curl -X POST http://localhost:3001/leases \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "vehicle_id": "550e8400-e29b-41d4-a716-446655440001",
    "start_date": "2024-01-01T00:00:00Z",
    "end_date": "2024-12-31T00:00:00Z",
    "monthly_payment": 500.00,
    "deposit_paid": 1000.00,
    "mileage_limit": 50000
  }'

# Response:
# {"id":"550e8400-e29b-41d4-a716-446655440002"}
```

#### Get Lease by ID
```bash
LEASE_ID="550e8400-e29b-41d4-a716-446655440002"

curl http://localhost:3001/leases/$LEASE_ID

# Response:
# {
#   "id": "...",
#   "user_id": "...",
#   "vehicle_id": "...",
#   "status": "DRAFT",
#   "start_date": "...",
#   ...
# }
```

#### Search Leases
```bash
# Via lease-service (direct)
curl "http://localhost:3001/leases?q=luxury"

# Via core-api (MeiliSearch aggregation)
curl "http://localhost:3000/api/v1/leases?q=luxury"
```

### 3. Payment Processing

#### Create Payment (Stripe)
```bash
PAYMENT_PAYLOAD='{
  "lease_id": "550e8400-e29b-41d4-a716-446655440002",
  "lease_payment_id": "550e8400-e29b-41d4-a716-446655440003",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "amount": 500.00,
  "currency": "USD",
  "method": "CARD",
  "provider": "stripe"
}'

curl -X POST http://localhost:3002/payments \
  -H "Content-Type: application/json" \
  -d "$PAYMENT_PAYLOAD"

# Response:
# {
#   "payment_id": "550e8400-e29b-41d4-a716-446655440004",
#   "status": "COMPLETED",
#   "provider_tx_id": "stripe_tx_20240115143025",
#   "created_at": "2024-01-15T14:30:25Z"
# }
```

#### Create Payment (Bank Transfer)
```bash
PAYMENT_PAYLOAD='{
  "lease_id": "550e8400-e29b-41d4-a716-446655440002",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "amount": 500.00,
  "currency": "USD",
  "method": "BANK_TRANSFER",
  "provider": "bank_api"
}'

curl -X POST http://localhost:3002/payments \
  -H "Content-Type: application/json" \
  -d "$PAYMENT_PAYLOAD"

# This also triggers blockchain-service via Redis!
```

#### Monitor Payment to Blockchain
```bash
# Watch blockchain-service logs
docker-compose logs -f blockchain-service

# You should see:
# blockchain-service | received payment event from Redis
# blockchain-service | TON: Sending transaction
# blockchain-service | TON: Transaction submitted
# blockchain-service | blockchain tx confirmed (after polling)
```

#### Webhook Reception
```bash
# Simulate provider webhook
curl -X POST http://localhost:3002/webhooks/stripe \
  -H "Content-Type: application/json" \
  -d '{
    "event": "charge.succeeded",
    "charge_id": "ch_1234567890",
    "amount": 50000
  }'
```

---

## Monitoring & Debugging

### View Logs

```bash
# All services
docker-compose logs -f

# Single service
docker-compose logs -f core-api
docker-compose logs -f payment-service
docker-compose logs -f blockchain-service

# Last N lines
docker-compose logs --tail=50 core-api

# Grep for errors
docker-compose logs | grep -i error
```

### Database Inspection

```bash
# Connect to PostgreSQL
docker-compose exec postgres psql -U leasing_user -d leasing_db

# Common queries:
psql> SELECT * FROM leases LIMIT 5;
psql> SELECT * FROM payments LIMIT 5;
psql> SELECT * FROM blockchain_transactions LIMIT 5;
psql> SELECT COUNT(*) FROM payments WHERE status = 'COMPLETED';
psql> \dt  # List all tables
```

### Redis Inspection

```bash
# Connect to Redis
docker-compose exec redis redis-cli

# Monitor pub/sub
redis> SUBSCRIBE payments
# Should see: {"event":"payment.completed",...}

# Check keys
redis> KEYS *
redis> TTL session:xxx

# Flush (dev only!)
redis> FLUSHDB
```

### MeiliSearch Inspection

```bash
# Web UI
open http://localhost:7700

# OR API
curl http://localhost:7700/indexes

# Check leases index
curl http://localhost:7700/indexes/leases/stats

# Search
curl "http://localhost:7700/indexes/leases/search" \
  -d '{"q":"luxury"}'
```

---

## Common Issues & Solutions

### 1. Ports Already in Use

```bash
# Find process on port
lsof -i :3000
lsof -i :3001
lsof -i :5432
lsof -i :6379

# Kill process
kill -9 <PID>

# OR change docker-compose ports
# Edit docker-compose.yml:
# ports:
#   - "3000:3000"  → "3010:3000"
```

### 2. Database Connection Failed

```bash
# Ensure DB is running
docker-compose ps postgres
# Should show "Up"

# Check logs
docker-compose logs postgres

# Verify connection string
echo "postgres://leasing_user:secure_pass@localhost:5432/leasing_db"

# Test manually
docker-compose exec postgres psql -U leasing_user -d leasing_db -c "SELECT 1"
```

### 3. Services Not Starting

```bash
# Check logs
docker-compose logs <service-name>

# Rebuild images
docker-compose build --no-cache <service-name>

# Start with foreground logging
docker-compose up <service-name>

# Check environment
env | grep -i postgres
env | grep -i redis
```

### 4. Blockchain Events Not Processing

```bash
# Check blockchain-service subscriber
docker-compose logs blockchain-service | grep -i subscribe

# Test Redis pub/sub manually
docker-compose exec redis redis-cli
redis> PUBLISH payments '{"event":"test"}'

# Check if blockchain-service received it
docker-compose logs blockchain-service
```

### 5. Build Failures

```bash
# Clean images
docker-compose down
docker system prune -a

# Rebuild
docker-compose build --no-cache

# Start
docker-compose up -d
```

---

## Performance Testing

### Load Test Lease Service

```bash
# Using ab (Apache Bench)
ab -n 1000 -c 10 http://localhost:3001/health

# Using hey (https://github.com/rakyll/hey)
hey -n 1000 -c 10 http://localhost:3001/health

# Using wrk (https://github.com/wg/wrk)
wrk -t 4 -c 100 -d 30s http://localhost:3001/health
```

### Load Test Payment Processing

```bash
# Create 100 payments sequentially
for i in {1..100}; do
  curl -X POST http://localhost:3002/payments \
    -H "Content-Type: application/json" \
    -d '{
      "lease_id": "'$i'",
      "user_id": "user_'$i'",
      "amount": 500.00,
      "provider": "stripe"
    }'
  sleep 0.5
done
```

### Monitor Database Performance

```bash
docker-compose exec postgres psql -U leasing_user -d leasing_db << EOF
-- Check active connections
SELECT datname, count(*) FROM pg_stat_activity GROUP BY datname;

-- Check slow queries
SELECT query, calls, mean_time FROM pg_stat_statements ORDER BY mean_time DESC LIMIT 5;

-- Check index usage
SELECT schemaname, tablename, indexname FROM pg_indexes WHERE schemaname NOT IN ('pg_catalog', 'information_schema');
EOF
```

---

## Deployment to Production

### Prerequisites
1. Docker & Docker Compose on production server
2. SSL certificate (Let's Encrypt recommended)
3. Persistent volumes for PostgreSQL
4. Backup strategy
5. Monitoring setup (Prometheus, DataDog, etc.)

### Production Steps

```bash
# 1. Pull code
git clone <repo-url>
cd car-leasing

# 2. Update .env for production
cat .env
# Change:
# - POSTGRES_PASSWORD=<strong-password>
# - JWT_SECRET=<random-key>
# - STRIPE_API_KEY=sk_live_...
# - TON_API_URL=https://toncenter.com/api/v2 (mainnet)
# - TON_WALLET_ADDRESS=<real-address>
# - TON_PRIVATE_KEY=<real-key>

# 3. Set resource limits
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d

# 4. Verify health
curl https://leasing.example.com/health  # Behind nginx

# 5. Monitor
docker-compose logs -f
```

### Production Docker Compose (docker-compose.prod.yml)

```yaml
version: '3.9'

services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}  # Use strong password!
    volumes:
      - postgres_data_prod:/var/lib/postgresql/data
    restart: unless-stopped
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 2G

  core-api:
    restart: unless-stopped
    deploy:
      replicas: 2  # Run 2 replicas for HA
      resources:
        limits:
          cpus: '1'
          memory: 512M

  # ... other services with similar prod settings

volumes:
  postgres_data_prod:
    driver: local
```

---

## Backup & Recovery

### PostgreSQL Backup

```bash
# Backup all databases
docker-compose exec postgres pg_dump -U leasing_user leasing_db > backup_$(date +%Y%m%d_%H%M%S).sql

# Restore from backup
docker-compose exec -T postgres psql -U leasing_user leasing_db < backup_20240115_143025.sql
```

### Redis Backup

```bash
# Redis persistence is configured in docker-compose
# Data is in redis_data volume

# Backup volume
docker run --rm -v leasing_redis_data:/data -v $(pwd):/backup \
  busybox tar czf /backup/redis_backup_$(date +%Y%m%d_%H%M%S).tar.gz /data
```

---

## Next Steps

1. Review **README.md** for architecture overview
2. Read **EXTENDING.md** for adding new payment strategies
3. Check **Docker Compose** services are healthy
4. Run integration tests
5. Deploy to staging/production

**Questions?** Check logs, consult EXTENDING.md, or contact the development team.
