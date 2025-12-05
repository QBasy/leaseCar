module leaseCar/lease-service

go 1.21

require (
    github.com/gofiber/fiber/v2 v2.46.0
    github.com/jackc/pgx/v5 v5.10.0
    github.com/meilisearch/meilisearch-go v0.1.1
)

replace leaseCar/utils => ../utils
