package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"faucet/internal/busha"
	"faucet/internal/config"
	"faucet/internal/handler"
)

func main() {

	_ = godotenv.Load()

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	cfg := config.Load()

	bushaClient := busha.NewClient(cfg.BushaAPIKey, cfg.BushaBaseURL, cfg.BushaProfileID)

	faucetHandler := handler.NewFaucetHandler(cfg, bushaClient)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(corsMiddleware())

	api := r.Group("/api/v1")
	{
		api.POST("/faucet", faucetHandler.HandleClaim)
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
