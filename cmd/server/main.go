package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"faucet/internal/busha"
	"faucet/internal/config"
	"faucet/internal/db"
	"faucet/internal/handler"
	rds "faucet/internal/redis"
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

	if err := db.InitSchema(database); err != nil {
		log.Fatalf("failed to init schema: %v", err)
	}
	slog.Info("database schema migrated")

	redisClient, err := rds.NewClient(cfg.RedisAddr)
	if err != nil {
		log.Fatalf("failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()
	slog.Info("connected to Redis")

	bushaClient := busha.NewClient(cfg.BushaAPIKey, cfg.BushaBaseURL, cfg.BushaProfileID)
	_ = bushaClient

	claimRepo := db.NewClaimRepository(database)

	faucetHandler := handler.NewFaucetHandler(cfg, claimRepo, redisClient)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(corsMiddleware())

	api := r.Group("/api/v1")
	{
		api.POST("/faucet", faucetHandler.HandleClaim)
		api.GET("/claims", faucetHandler.HandleListClaims)
		api.GET("/claims/:id", faucetHandler.HandleGetClaim)
		api.GET("/config", faucetHandler.HandleGetConfig)
	}

	addr := ":" + cfg.ServerPort
	slog.Info("starting faucet server", "addr", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
