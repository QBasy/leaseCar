package repositories

import (
    "context"
    "time"

    "github.com/jackc/pgx/v5/pgxpool"
    "leaseCar/lease-service/internal/dtos"
)

type LeaseRepository struct {
    pool *pgxpool.Pool
}

func NewLeaseRepository(pool *pgxpool.Pool) *LeaseRepository {
    return &LeaseRepository{pool: pool}
}

func (r *LeaseRepository) Create(ctx context.Context, in *dtos.LeaseCreateRequest) (string, error) {
    var id string
    sql := `INSERT INTO leases (user_id, vehicle_id, start_date, end_date, monthly_payment, deposit_paid, total_cost, mileage_limit)
            VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`
    total := in.Monthly // simplification; real total calc elsewhere
    err := r.pool.QueryRow(ctx, sql, in.UserID, in.VehicleID, in.StartDate, in.EndDate, in.Monthly, in.Deposit, total, in.MileageLimit).Scan(&id)
    if err != nil {
        return "", err
    }
    return id, nil
}

func (r *LeaseRepository) GetByID(ctx context.Context, id string) (*dtos.Lease, error) {
    sql := `SELECT id, user_id, vehicle_id, status, start_date, end_date, monthly_payment, deposit_paid, total_cost, created_at FROM leases WHERE id = $1 LIMIT 1`
    row := r.pool.QueryRow(ctx, sql, id)
    var l dtos.Lease
    err := row.Scan(&l.ID, &l.UserID, &l.VehicleID, &l.Status, &l.StartDate, &l.EndDate, &l.Monthly, &l.Deposit, &l.TotalCost, &l.CreatedAt)
    if err != nil {
        return nil, err
    }
    return &l, nil
}
