package adapters

import (
	"context"
	"errors"
	"time"

	"leaseCar/payment-service/internal/dtos"
)

type BankAdapter struct {
	url string
	apiKey string
}

func NewBankAdapter(url, apiKey string) *BankAdapter { return &BankAdapter{url: url, apiKey: apiKey} }

func (b *BankAdapter) SendPayment(req *dtos.PaymentRequest) (*BankResponse, error) {
	// In real implementation do HTTP calls to bank API
	// Here we simulate
	if b.url == "" {
		return nil, errors.New("bank adapter not configured")
	}
	// simulate success
	return &BankResponse{TransactionID: "bank_tx_" + time.Now().Format("20060102150405"), Status: "OK"}, nil
}

type BankResponse struct {
	TransactionID string
	Status string
}
