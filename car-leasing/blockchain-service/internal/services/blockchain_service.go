package services

import (
	"context"
	"encoding/json"
	"time"

	"leaseCar/blockchain-service/internal/adapters"
	"leaseCar/blockchain-service/internal/dtos"
	"leaseCar/blockchain-service/internal/repositories"
	"leaseCar/utils/logger"
)

type BlockchainService struct {
	repo *repositories.BlockchainRepository
	tonAdapter *adapters.TONAdapter
}

func NewBlockchainService(repo *repositories.BlockchainRepository, ton *adapters.TONAdapter) *BlockchainService {
	return &BlockchainService{repo: repo, tonAdapter: ton}
}

// ProcessPaymentEvent handles payment.completed event from Redis (Observer pattern)
func (s *BlockchainService) ProcessPaymentEvent(ctx context.Context, payload []byte) error {
	var evt dtos.PaymentEventPayload
	if err := json.Unmarshal(payload, &evt); err != nil {
		logger.Error("failed to parse payment event")
		return err
	}

	logger.Info("processing payment event", logger.WithFields())
	
	// TODO: in production, get recipient address from leaseID metadata
	recipientAddr := "UQA..." // placeholder
	amount := "1000000" // nano TON

	// Send to blockchain
	txn, err := s.tonAdapter.SendTransaction(ctx, recipientAddr, amount)
	if err != nil {
		logger.Error("failed to send TON transaction")
		return err
	}

	txn.PaymentID = evt.PaymentID
	txn.Status = "SUBMITTED"

	// Save blockchain tx record
	if err := s.repo.SaveTransaction(ctx, txn, evt.PaymentID); err != nil {
		logger.Error("failed to save blockchain transaction")
		return err
	}

	// Update payment record with tx hash
	if err := s.repo.UpdatePaymentTxHash(ctx, evt.PaymentID, txn.TxHash); err != nil {
		logger.Error("failed to update payment tx hash")
		return err
	}

	// Poll for confirmation (async)
	go s.pollConfirmation(txn.TxHash)

	logger.Info("payment event processed successfully")
	return nil
}

// pollConfirmation periodically checks blockchain tx status (background task)
func (s *BlockchainService) pollConfirmation(txHash string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			status, err := s.tonAdapter.CheckStatus(ctx, txHash)
			if err != nil {
				logger.Error("failed to check tx status")
				continue
			}
			if status == "CONFIRMED" {
				s.repo.UpdateTransactionStatus(ctx, txHash, status)
				logger.Info("blockchain tx confirmed")
				return
			}
		case <-ctx.Done():
			logger.Error("polling timeout")
			return
		}
	}
}
