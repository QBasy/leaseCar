module leaseCar/payment-service

go 1.21

require (
	github.com/gofiber/fiber/v2 v2.46.0
	github.com/jackc/pgx/v5 v5.10.0
	github.com/google/uuid v1.4.0
)

replace leaseCar/utils => ../utils
