package main

import (
	"leaseCar/blockchain-service/internal/adapters"
	"leaseCar/blockchain-service/internal/repositories"
	"leaseCar/blockchain-service/internal/services"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Factory functions for dependency injection
func NewBlockchainRepository(pool *pgxpool.Pool) *repositories.BlockchainRepository {
	return repositories.NewBlockchainRepository(pool)
}

func NewTONAdapter(apiUrl, walletAddr, privKey string) *adapters.TONAdapter {
	return adapters.NewTONAdapter(apiUrl, walletAddr, privKey)
}

func NewBlockchainService(repo *repositories.BlockchainRepository, ton *adapters.TONAdapter) *services.BlockchainService {
	return services.NewBlockchainService(repo, ton)
}
