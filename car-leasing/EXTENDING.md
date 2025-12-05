# Extending Payment Strategies & TON Integration Guide

## Adding a New Payment Strategy (Step-by-Step)

### Example: Apple Pay Strategy

#### 1. Create Strategy Implementation
File: `payment-service/internal/strategies/applepay_strategy.go`

```go
package strategies

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"leaseCar/payment-service/internal/dtos"
	"leaseCar/utils/logger"
)

type ApplePayStrategy struct {
	merchantID string
	domainName string
	certPath   string
	keyPath    string
}

func NewApplePayStrategy(merchantID, domainName, certPath, keyPath string) *ApplePayStrategy {
	return &ApplePayStrategy{
		merchantID: merchantID,
		domainName: domainName,
		certPath:   certPath,
		keyPath:    keyPath,
	}
}

func (s *ApplePayStrategy) Validate(req *dtos.PaymentRequest) error {
	// Validate Apple token structure
	if req.Amount <= 0 {
		return errors.New("invalid amount")
	}
	// In production: verify Apple signature
	return nil
}

func (s *ApplePayStrategy) Process(ctx context.Context, req *dtos.PaymentRequest) (*dtos.PaymentResponse, error) {
	logger.Info("ApplePayStrategy.Process start")

	// 1. Verify Apple token (decrypt, validate signature)
	// token := req.Metadata["token"] // from client
	// if err := s.verifyAppleToken(token); err != nil { return nil, err }

	// 2. Call Apple Pay server API
	txID := "applepay_" + time.Now().Format("20060102150405")
	
	// 3. Simulate network call
	time.Sleep(800 * time.Millisecond)

	resp := &dtos.PaymentResponse{
		Status:       "COMPLETED",
		ProviderTxID: txID,
		CreatedAt:    time.Now(),
	}

	logger.Info("ApplePayStrategy.Process done")
	return resp, nil
}

// Helper to verify Apple Pay token (production use)
func (s *ApplePayStrategy) verifyAppleToken(token string) error {
	// In production:
	// 1. Decrypt token using merchant cert/key
	// 2. Verify signature
	// 3. Check token timestamp (not expired)
	// 4. Extract payment data (amount, currency, PAN)
	return nil
}
```

#### 2. Register in Factory
File: `payment-service/internal/factory/payment_factory.go`

```go
package factory

import (
	"os"
	"leaseCar/payment-service/internal/strategies"
	cfg "leaseCar/utils/config"
)

func (f *PaymentFactory) GetStrategy(provider string) strategies.PaymentStrategy {
	switch provider {
	case "stripe":
		apiKey := os.Getenv("STRIPE_API_KEY")
		return strategies.NewStripeStrategy(apiKey)
	
	case "bank_api":
		url := os.Getenv("BANK_API_URL")
		apiKey := os.Getenv("BANK_API_KEY")
		adapter := adapters.NewBankAdapter(url, apiKey)
		return strategies.NewBankStrategy(adapter)
	
	case "applepay":  // NEW
		merchantID := os.Getenv("APPLEPAY_MERCHANT_ID")
		domainName := os.Getenv("APPLEPAY_DOMAIN_NAME")
		certPath := os.Getenv("APPLEPAY_CERT_PATH")
		keyPath := os.Getenv("APPLEPAY_KEY_PATH")
		return strategies.NewApplePayStrategy(merchantID, domainName, certPath, keyPath)
	
	default:
		return nil
	}
}
```

#### 3. Add Environment Variables
Update `.env`:
```bash
APPLEPAY_MERCHANT_ID=com.example.leasing
APPLEPAY_DOMAIN_NAME=leasing.example.com
APPLEPAY_CERT_PATH=/app/certs/apple_cert.pem
APPLEPAY_KEY_PATH=/app/certs/apple_key.pem
```

#### 4. Test
```bash
curl -X POST http://localhost:3002/payments \
  -H "Content-Type: application/json" \
  -d '{
    "lease_id": "550e8400-e29b-41d4-a716-446655440000",
    "user_id": "550e8400-e29b-41d4-a716-446655440001",
    "amount": 500.00,
    "currency": "USD",
    "method": "DIGITAL_WALLET",
    "provider": "applepay"
  }'
```

Response:
```json
{
  "payment_id": "generated-uuid",
  "status": "COMPLETED",
  "provider_tx_id": "applepay_20240115143025",
  "created_at": "2024-01-15T14:30:25Z"
}
```

---

## Strategy Pattern Reference

All strategies must implement this interface:

```go
type PaymentStrategy interface {
	// Validate checks input data before processing
	Validate(req *PaymentRequest) error
	
	// Process executes the payment with the provider
	Process(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error)
}
```

**Key principles:**
- Each strategy is isolated and independently testable
- Strategies can be added/removed without modifying core logic
- Factory pattern selects strategy at runtime based on provider name
- Service orchestrates: validate → process → publish event

---

## TON Blockchain Integration (Full Production Setup)

### Prerequisites
1. TON wallet with testnet coins (https://faucet.dev)
2. TON wallet keypair (private key + address)
3. TON API credentials (https://toncenter.com)

### Step 1: Get TON Credentials

```bash
# Generate wallet keypair (or import existing)
# Using tonutils-go CLI or TON documentation

# Set in .env
TON_WALLET_ADDRESS=0:1234567890abcdef...
TON_PRIVATE_KEY=abcdef1234567890...
TON_API_URL=https://testnet.toncenter.com/api/v2  # testnet
# TON_API_URL=https://toncenter.com/api/v2        # mainnet
```

### Step 2: Implement TON Adapter

Update `blockchain-service/internal/adapters/ton_adapter.go`:

```go
package adapters

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"leaseCar/blockchain-service/internal/dtos"
	"leaseCar/utils/logger"
)

type TONAdapter struct {
	apiUrl string
	walletAddress string
	privateKey string
}

// SendTransaction sends a payment transaction to TON blockchain
func (t *TONAdapter) SendTransaction(ctx context.Context, toAddr, amount string) (*dtos.BlockchainTransaction, error) {
	logger.Info("TON: Sending transaction")

	// Step 1: Build message
	// In real impl: use tonutils-go to create Cell with message
	msgBody := map[string]interface{}{
		"destination": toAddr,
		"amount":      amount, // in nano-TON
	}

	// Step 2: Sign with private key
	// (requires tonutils-go crypto package)
	// signature := signTransaction(msgBody, t.privateKey)

	// Step 3: Send to TON API
	txHash := fmt.Sprintf("ton_%d", time.Now().Unix())
	err := t.submitToTON(ctx, msgBody)
	if err != nil {
		logger.Error("TON submission failed")
		return nil, err
	}

	result := &dtos.BlockchainTransaction{
		TxHash: txHash,
		From:   t.walletAddress,
		To:     toAddr,
		Amount: amount,
		Status: "SUBMITTED",
	}

	logger.Info("TON: Transaction submitted", logger.WithFields())
	return result, nil
}

// CheckStatus polls TON blockchain for transaction confirmation
func (t *TONAdapter) CheckStatus(ctx context.Context, txHash string) (string, error) {
	// Call TON API: /getTransactionData
	// Check: confirmations > 0 → CONFIRMED
	// Return: status string
	
	// Simulated query:
	resp, err := t.callTONAPI(ctx, "/getTransactionData", map[string]interface{}{
		"tx_hash": txHash,
	})
	if err != nil {
		return "", err
	}

	confirmations := resp["confirmations"].(float64)
	if confirmations > 0 {
		return "CONFIRMED", nil
	}
	return "PENDING", nil
}

func (t *TONAdapter) submitToTON(ctx context.Context, msgBody map[string]interface{}) error {
	// POST to TON API /sendBoc
	client := &http.Client{Timeout: 10 * time.Second}
	
	bodyStr := fmt.Sprintf(`{
		"boc": "%s",
		"mode": 3
	}`, "base64_encoded_message_here")
	
	req, _ := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf("%s/sendBoc", t.apiUrl), strings.NewReader(bodyStr))
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("TON API error: %d - %s", resp.StatusCode, string(body))
	}
	
	return nil
}

func (t *TONAdapter) callTONAPI(ctx context.Context, method string, params map[string]interface{}) (map[string]interface{}, error) {
	// Generic TON API call with signature (if required)
	// Returns JSON response
	return make(map[string]interface{}), nil
}

// signTransaction signs message with private key (pseudo-code)
func signTransaction(msg map[string]interface{}, privateKey string) string {
	data := fmt.Sprintf("%v", msg)
	hash := sha256.Sum256([]byte(data))
	// Sign hash with ECDSA private key (tonutils-go crypto)
	// Return base64-encoded signature
	return base64.StdEncoding.EncodeToString(hash[:])
}
```

### Step 3: Handle Webhooks (Optional)

If TON network sends transaction confirmations via webhook:

```go
// blockchain-service/internal/controllers/webhook_controller.go
func (w *WebhookController) HandleTONWebhook(c *fiber.Ctx) error {
	var payload map[string]interface{}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid payload"})
	}

	// Verify webhook signature
	// Update blockchain_transactions.confirmed = true
	
	return c.Status(200).JSON(fiber.Map{"status": "ok"})
}
```

### Step 4: Write Smart Contract (FunC)

Optional: Create an automated lease payment smart contract on TON.

File: `blockchain-service/contracts/LeasePayment.fc`

```func
// Smart contract for automated lease payments

() recv_internal(int my_balance, int msg_value, cell in_msg_full, slice in_msg_body) impure {
    ;; Parse incoming payment
    
    ;; Verify lease exists
    
    ;; Update payment status in contract storage
    
    ;; Emit completion event
    
    send_simple_message(lease_owner, amount);
}

() send_payment(slice recipient, int amount) {
    send_simple_message(recipient, amount);
}
```

Deploy with:
```bash
# Using tontools or ton-cli
ton-cli deploy LeasePayment.fc
```

### Step 5: Test Full Flow

1. Start stack:
   ```bash
   docker-compose up -d
   ```

2. Create payment:
   ```bash
   curl -X POST http://localhost:3002/payments \
     -H "Content-Type: application/json" \
     -d '{
       "lease_id": "test-lease",
       "user_id": "test-user",
       "amount": 1000000,
       "currency": "USD",
       "method": "CRYPTO",
       "provider": "stripe"
     }'
   ```

3. Observe blockchain-service logs:
   ```bash
   docker-compose logs -f blockchain-service
   ```
   
   Expected:
   ```
   blockchain-service | payment event processed successfully
   blockchain-service | TON: Sending transaction
   blockchain-service | TON: Transaction submitted
   blockchain-service | blockchain tx confirmed
   ```

4. Verify in database:
   ```bash
   docker-compose exec postgres psql -U leasing_user -d leasing_db << EOF
   SELECT id, status, blockchain_tx_hash FROM payments LIMIT 5;
   SELECT tx_hash, status, confirmed FROM blockchain_transactions LIMIT 5;
   EOF
   ```

---

## Performance Tuning

### Payment Processing
- **Async indexing:** MeiliSearch indexing runs in background (doesn't block response)
- **Batch webhooks:** Collect multiple payment updates before DB write (if high volume)
- **Caching:** Redis caches lease data to reduce DB queries

### TON Integration
- **Connection pooling:** Reuse HTTP connections to TON API
- **Exponential backoff:** Retry failed submissions with exponential backoff
- **Batch confirmations:** Batch-check 100 pending transactions per poll cycle

### Database
- **Connection pooling:** PgBouncer recommended (production)
- **Indexes:** Already created on `payments.status`, `payments.created_at`, `blockchain_transactions.tx_hash`
- **Partitioning:** Consider partitioning `payments` and `blockchain_transactions` by month (large scale)

---

## Debugging

### Payment not reaching blockchain-service?
```bash
# Monitor Redis pub/sub
docker-compose exec redis redis-cli
> SUBSCRIBE payments
# Should see: {"event":"payment.completed",...}

# Check payment-service logs
docker-compose logs payment-service | grep -i redis
```

### TON transaction fails?
```bash
# Check TON adapter logs
docker-compose logs blockchain-service | grep -i ton

# Verify wallet balance
# Visit: https://testnet.tonviewer.com/ and search wallet address

# Check TON API availability
curl https://testnet.toncenter.com/api/v2/getMasterchainInfo
# Should return: {"ok": true, ...}
```

### Confirmation polling not working?
```bash
# Check blockchain-service goroutine
docker-compose logs blockchain-service | grep "pollConfirmation"

# Manually check TON tx status
curl https://testnet.toncenter.com/api/v2/getTransactionData \
  -d '{"tx_hash": "your_hash"}'
```

---

## References

- **TON Documentation:** https://ton.org/docs
- **TON SDK (Go):** https://github.com/xssnick/tonutils-go
- **FunC Language:** https://ton.org/docs/#/func
- **TonCenter API:** https://toncenter.com/api/v2/
- **Testnet Faucet:** https://faucet.dev

---

**Happy building! 🚀**
