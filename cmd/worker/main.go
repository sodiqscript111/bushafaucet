package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/joho/godotenv"

	"faucet/internal/busha"
	"faucet/internal/config"
	"faucet/internal/db"
	rds "faucet/internal/redis"
	"faucet/internal/worker"
)

func main() {

	_ = godotenv.Load()

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	cfg := config.Load()

	database, err := db.NewDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	slog.Info("connected to PostgreSQL")

	redisClient, err := rds.NewClient(cfg.RedisAddr)
	if err != nil {
		log.Fatalf("failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()
	slog.Info("connected to Redis")

	bushaClient := busha.NewClient(cfg.BushaAPIKey, cfg.BushaBaseURL, cfg.BushaProfileID)

	claimRepo := db.NewClaimRepository(database)

	w := worker.NewWorker(cfg, claimRepo, redisClient, bushaClient)

	slog.Info("starting faucet worker")
	if err := w.Run(context.Background()); err != nil {
		log.Fatalf("worker failed: %v", err)
	}
}
