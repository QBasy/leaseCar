package strategies

import (
	"context"
	"time"

	"leaseCar/payment-service/internal/dtos"
	"leaseCar/utils/logger"
)

// A minimal Stripe strategy that simulates processing
type StripeStrategy struct {
	apiKey string
}

func NewStripeStrategy(apiKey string) *StripeStrategy { return &StripeStrategy{apiKey: apiKey} }

func (s *StripeStrategy) Validate(req *dtos.PaymentRequest) error {
	// Example: check card fields in metadata (omitted)
	return nil
}

func (s *StripeStrategy) Process(ctx context.Context, req *dtos.PaymentRequest) (*dtos.PaymentResponse, error) {
	logger.Info("StripeStrategy.Process start")
	// Simulate network call
	time.Sleep(500 * time.Millisecond)
	// Return simulated provider transaction id
	resp := &dtos.PaymentResponse{
		PaymentID:    "",
		Status:       "COMPLETED",
		ProviderTxID: "stripe_tx_" + time.Now().Format("20060102150405"),
		CreatedAt:    time.Now(),
	}
	logger.Info("StripeStrategy.Process done")
	return resp, nil
}
