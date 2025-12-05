package repositories

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"leaseCar/blockchain-service/internal/dtos"
)

type BlockchainRepository struct {
	pool *pgxpool.Pool
}

func NewBlockchainRepository(pool *pgxpool.Pool) *BlockchainRepository {
	return &BlockchainRepository{pool: pool}
}

func (r *BlockchainRepository) SaveTransaction(ctx context.Context, txn *dtos.BlockchainTransaction, paymentID string) error {
	sql := `INSERT INTO blockchain_transactions (payment_id, tx_hash, from_address, to_address, amount, status, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.pool.Exec(ctx, sql, paymentID, txn.TxHash, txn.From, txn.To, txn.Amount, txn.Status, time.Now())
	return err
}

func (r *BlockchainRepository) UpdateTransactionStatus(ctx context.Context, txHash, status string) error {
	sql := `UPDATE blockchain_transactions SET status=$1, confirmed=$2, confirmed_at=$3, updated_at=$4 WHERE tx_hash=$5`
	var confirmed bool
	if status == "CONFIRMED" {
		confirmed = true
	}
	_, err := r.pool.Exec(ctx, sql, status, confirmed, time.Now(), time.Now(), txHash)
	return err
}

func (r *BlockchainRepository) UpdatePaymentTxHash(ctx context.Context, paymentID, txHash string) error {
	sql := `UPDATE payments SET blockchain_tx_hash=$1, updated_at=$2 WHERE id=$3`
	_, err := r.pool.Exec(ctx, sql, txHash, time.Now(), paymentID)
	return err
}
