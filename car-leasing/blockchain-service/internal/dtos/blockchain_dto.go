package dtos

type PaymentEventPayload struct {
	Event     string `json:"event"`
	PaymentID string `json:"payment_id"`
	ProviderTx string `json:"provider_tx"`
	Status    string `json:"status"`
}

type BlockchainTransaction struct {
	PaymentID string `json:"payment_id"`
	TxHash    string `json:"tx_hash"`
	From      string `json:"from"`
	To        string `json:"to"`
	Amount    string `json:"amount"`
	Status    string `json:"status"`
}
