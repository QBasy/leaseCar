package strategies

import (
	"context"
	"time"

	"leaseCar/payment-service/internal/dtos"
	"leaseCar/payment-service/internal/adapters"
	"leaseCar/utils/logger"
)

type BankStrategy struct {
	adapter *adapters.BankAdapter
}

func NewBankStrategy(a *adapters.BankAdapter) *BankStrategy { return &BankStrategy{adapter: a} }

func (s *BankStrategy) Validate(req *dtos.PaymentRequest) error {
	// Basic validation for bank transfer
	return nil
}

func (s *BankStrategy) Process(ctx context.Context, req *dtos.PaymentRequest) (*dtos.PaymentResponse, error) {
	logger.Info("BankStrategy.Process start")
	// Use adapter to call bank API
	res, err := s.adapter.SendPayment(req)
	if err != nil {
		return nil, err
	}
	// simulate processing time
	time.Sleep(200 * time.Millisecond)
	return &dtos.PaymentResponse{
		PaymentID:    "",
		Status:       "PROCESSING",
		ProviderTxID: res.TransactionID,
		CreatedAt:    time.Now(),
	}, nil
}
