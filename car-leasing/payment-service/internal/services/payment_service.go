package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"leaseCar/payment-service/internal/dtos"
	"leaseCar/payment-service/internal/factory"
	"leaseCar/payment-service/internal/repositories"
	"leaseCar/payment-service/internal/strategies"
	redisutil "leaseCar/utils/redis"
	"leaseCar/utils/logger"
)

type PaymentService struct {
	repo *repositories.PaymentRepository
	factory *factory.PaymentFactory
	redisClient *redisutil.Client
}

func NewPaymentService(repo *repositories.PaymentRepository, factory *factory.PaymentFactory, r *redisutil.Client) *PaymentService {
	return &PaymentService{repo: repo, factory: factory, redisClient: r}
}

func (s *PaymentService) CreatePayment(ctx context.Context, req *dtos.PaymentRequest) (*dtos.PaymentResponse, error) {
	// validate and create record
	id, err := s.repo.Create(ctx, req)
	if err != nil {
		return nil, err
	}

	// pick strategy
	strat := s.factory.GetStrategy(req.Provider)
	if strat == nil {
		return nil, fmt.Errorf("no strategy for provider %s", req.Provider)
	}

	if err := strat.Validate(req); err != nil {
		s.repo.UpdateStatus(ctx, id, "FAILED", "")
		return nil, err
	}

	// process
	resp, err := strat.Process(ctx, req)
	if err != nil {
		s.repo.UpdateStatus(ctx, id, "FAILED", "")
		return nil, err
	}

	// update record with provider tx id and status
	status := resp.Status
	tx := resp.ProviderTxID
	s.repo.UpdateStatus(ctx, id, status, tx)

	// emit event to redis for observer (blockchain-service)
	event := map[string]interface{}{"event": "payment.completed", "payment_id": id, "provider_tx": tx, "status": status}
	b, _ := json.Marshal(event)
	if err := s.redisClient.Publish(context.Background(), "payments", string(b)); err != nil {
		logger.Error("failed to publish payment event")
	}

	resp.PaymentID = id
	return resp, nil
}

func (s *PaymentService) HandleProviderWebhook(ctx context.Context, provider string, payload map[string]interface{}) error {
	// For simplicity, provider-specific parsing omitted
	logger.Info("received webhook", logger.WithFields())
	return nil
}
