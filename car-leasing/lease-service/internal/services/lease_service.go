package services

import (
    "context"
    "time"

    "leaseCar/lease-service/internal/dtos"
    "leaseCar/lease-service/internal/repositories"
    "leaseCar/lease-service/internal/adapters"
)

type LeaseService struct {
    repo *repositories.LeaseRepository
    meili *adapters.MeiliAdapter
}

func NewLeaseService(r *repositories.LeaseRepository, m *adapters.MeiliAdapter) *LeaseService {
    return &LeaseService{repo: r, meili: m}
}

func (s *LeaseService) Create(ctx context.Context, in *dtos.LeaseCreateRequest) (string, error) {
    id, err := s.repo.Create(ctx, in)
    if err != nil {
        return "", err
    }

    // create simple lease payload for indexing
    lease := map[string]interface{}{
        "id": id,
        "user_id": in.UserID,
        "vehicle_id": in.VehicleID,
        "start_date": in.StartDate,
        "end_date": in.EndDate,
        "monthly_payment": in.Monthly,
    }
    // index to meili (best-effort)
    go func() {
        ctx2, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        _ = s.meili.IndexLease(ctx2, "leases", lease)
    }()

    return id, nil
}

func (s *LeaseService) GetByID(ctx context.Context, id string) (*dtos.Lease, error) {
    return s.repo.GetByID(ctx, id)
}

func (s *LeaseService) Search(ctx context.Context, q string, limit int) (interface{}, error) {
    res, err := s.meili.Search(ctx, "leases", q, limit)
    if err != nil {
        return nil, err
    }
    return res.Hits, nil
}
