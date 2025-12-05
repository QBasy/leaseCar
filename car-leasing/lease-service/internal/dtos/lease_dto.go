package dtos

import "time"

type LeaseCreateRequest struct {
    UserID     string    `json:"user_id"`
    VehicleID  string    `json:"vehicle_id"`
    StartDate  time.Time `json:"start_date"`
    EndDate    time.Time `json:"end_date"`
    Monthly    float64   `json:"monthly_payment"`
    Deposit    float64   `json:"deposit_paid"`
    MileageLimit int     `json:"mileage_limit"`
}

type Lease struct {
    ID         string    `json:"id"`
    UserID     string    `json:"user_id"`
    VehicleID  string    `json:"vehicle_id"`
    Status     string    `json:"status"`
    StartDate  time.Time `json:"start_date"`
    EndDate    time.Time `json:"end_date"`
    Monthly    float64   `json:"monthly_payment"`
    Deposit    float64   `json:"deposit_paid"`
    TotalCost  float64   `json:"total_cost"`
    CreatedAt  time.Time `json:"created_at"`
}
