package dtos

import "time"

type PaymentRequest struct {
	LeaseID       string  `json:"lease_id"`
	LeasePaymentID string `json:"lease_payment_id"`
	UserID        string  `json:"user_id"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	Method        string  `json:"method"`
	Provider      string  `json:"provider"`
}

type PaymentResponse struct {
	PaymentID     string    `json:"payment_id"`
	Status        string    `json:"status"`
	ProviderTxID  string    `json:"provider_tx_id,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

*** End Patch