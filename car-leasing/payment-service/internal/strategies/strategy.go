package strategies

import (
	"context"
	"leaseCar/payment-service/internal/dtos"
)

type PaymentStrategy interface {
	Process(ctx context.Context, req *dtos.PaymentRequest) (*dtos.PaymentResponse, error)
	Validate(req *dtos.PaymentRequest) error
}
