package adapters

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"leaseCar/blockchain-service/internal/dtos"
	"leaseCar/utils/logger"
)

type TONAdapter struct {
	apiUrl string
	walletAddress string
	privateKey string
}

func NewTONAdapter(apiUrl, walletAddr, privKey string) *TONAdapter {
	return &TONAdapter{apiUrl: apiUrl, walletAddress: walletAddr, privateKey: privKey}
}

func (t *TONAdapter) SendTransaction(ctx context.Context, toAddr string, amount string) (*dtos.BlockchainTransaction, error) {
	logger.Info("TONAdapter.SendTransaction start")
	
	if t.privateKey == "change_me" || t.walletAddress == "0:change_me" {
		logger.Error("TON credentials not configured")
		return nil, errors.New("TON wallet not configured")
	}

	// In production: sign transaction with private key, send to TON API
	// For now, simulate successful submission
	txHash := fmt.Sprintf("ton_tx_%d", time.Now().Unix())
	
	// Simulate HTTP call to TON API
	time.Sleep(200 * time.Millisecond)
	
	result := &dtos.BlockchainTransaction{
		TxHash: txHash,
		From:   t.walletAddress,
		To:     toAddr,
		Amount: amount,
		Status: "SUBMITTED",
	}
	
	logger.Info("TONAdapter.SendTransaction done", logger.WithFields())
	return result, nil
}

// CheckStatus queries TON blockchain for transaction status (simulated)
func (t *TONAdapter) CheckStatus(ctx context.Context, txHash string) (string, error) {
	logger.Info("TONAdapter.CheckStatus")
	// In production: query TON API for tx status
	// Simulate: always return CONFIRMED after delay
	time.Sleep(100 * time.Millisecond)
	return "CONFIRMED", nil
}

// Helper for actual HTTP calls to TON API (structure only)
func (t *TONAdapter) callTONAPI(method string, params map[string]interface{}) (map[string]interface{}, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	body, _ := json.Marshal(params)
	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/%s", t.apiUrl, method), nil)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return result, nil
}
