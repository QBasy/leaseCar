package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	cfg "leaseCar/utils/config"
	"leaseCar/utils/logger"
	redisutil "leaseCar/utils/redis"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "./config/config.yaml"
	}
	conf, err := cfg.Load(configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	logger.Info("blockchain-service config loaded")

	// DB (optional for blockchain-service, but can record tx data)
	dbUrl := fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		conf.Database.User, conf.Database.Password, conf.Database.Host, conf.Database.Port, conf.Database.DBName)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, dbUrl)
	if err != nil {
		logger.Error("failed to connect db")
		log.Fatalf("db connect error: %v", err)
	}
	defer pool.Close()

	// redis
	r, err := redisutil.New(conf.Redis.Host, conf.Redis.Port, conf.Redis.Password, conf.Redis.DB)
	if err != nil {
		log.Fatalf("redis connect error: %v", err)
	}
	defer r.Close()

	app := fiber.New()

	// health
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(200).JSON(fiber.Map{"status": "ok"})
	})

	// wire components
	repo := NewBlockchainRepository(pool)
	tonAdapter := NewTONAdapter(
		os.Getenv("TON_API_URL"),
		os.Getenv("TON_WALLET_ADDRESS"),
		os.Getenv("TON_PRIVATE_KEY"),
	)
	svc := NewBlockchainService(repo, tonAdapter)

	// Start Redis subscriber (Observer pattern)
	go subscribeToPayments(r, svc)

	port := conf.Server.Port
	logger.Info("blockchain-service starting")
	if err := app.Listen(fmt.Sprintf(":%d", port)); err != nil {
		logger.Error("fiber listen error")
		log.Fatalf("fiber error: %v", err)
	}
}

// subscribeToPayments subscribes to Redis payments channel and processes events
func subscribeToPayments(r *redisutil.Client, svc *BlockchainService) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pubsub := r.Subscribe(ctx, "payments")
	defer pubsub.Close()

	for msg := range pubsub.Channel() {
		logger.Info("received payment event from Redis")
		_ = svc.ProcessPaymentEvent(context.Background(), []byte(msg.Payload))
	}
}
