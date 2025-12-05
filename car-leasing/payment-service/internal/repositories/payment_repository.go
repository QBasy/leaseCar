package repositories

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/google/uuid"
	"leaseCar/payment-service/internal/dtos"
)

type PaymentRepository struct {
	pool *pgxpool.Pool
}

func NewPaymentRepository(pool *pgxpool.Pool) *PaymentRepository { return &PaymentRepository{pool: pool} }

func (r *PaymentRepository) Create(ctx context.Context, req *dtos.PaymentRequest) (string, error) {
	id := uuid.New().String()
	sql := `INSERT INTO payments (id, lease_id, lease_payment_id, user_id, amount, currency, status, method, provider, created_at)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`
	_, err := r.pool.Exec(ctx, sql, id, req.LeaseID, req.LeasePaymentID, req.UserID, req.Amount, req.Currency, "PENDING", req.Method, req.Provider, time.Now())
	if err != nil {
		return "", err
	}
	return id, nil
}

func (r *PaymentRepository) UpdateStatus(ctx context.Context, id, status, txHash string) error {
	sql := `UPDATE payments SET status=$1, transaction_id=$2, updated_at=$3 WHERE id=$4`
	_, err := r.pool.Exec(ctx, sql, status, txHash, time.Now(), id)
	return err
}
