package worker

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"faucet/internal/busha"
	"faucet/internal/config"
	"faucet/internal/db"
	"faucet/internal/models"
	rds "faucet/internal/redis"
)

type Worker struct {
	cfg		*config.Config
	claimRepo	*db.ClaimRepository
	redis		*rds.Client
	busha		*busha.Client
	consumer	string
}

func NewWorker(cfg *config.Config, claimRepo *db.ClaimRepository, redis *rds.Client, busha *busha.Client) *Worker {
	hostname, _ := os.Hostname()
	return &Worker{
		cfg:		cfg,
		claimRepo:	claimRepo,
		redis:		redis,
		busha:		busha,
		consumer:	"worker-" + hostname,
	}
}

func (w *Worker) Run(ctx context.Context) error {

	if err := w.redis.CreateConsumerGroup(ctx); err != nil {
		return err
	}

	slog.Info("faucet worker started", "consumer", w.consumer)

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			slog.Info("worker shutting down")
			return nil
		default:
		}

		streams, err := w.redis.ReadJobs(ctx, w.consumer, 1, 5*time.Second)
		if err != nil {

			if ctx.Err() != nil {
				return nil
			}
			slog.Error("failed to read jobs", "error", err)
			time.Sleep(1 * time.Second)
			continue
		}

		if streams == nil {
			continue
		}

		for _, stream := range streams {
			for _, msg := range stream.Messages {
				w.processJob(ctx, msg.ID, msg.Values)
			}
		}
	}
}

func (w *Worker) processJob(ctx context.Context, messageID string, values map[string]interface{}) {
	claimID, _ := values["claim_id"].(string)
	walletAddress, _ := values["wallet_address"].(string)
	blockchain, _ := values["blockchain"].(string)

	logger := slog.With(
		"claim_id", claimID,
		"wallet", walletAddress,
		"blockchain", blockchain,
		"message_id", messageID,
	)

	logger.Info("processing faucet job")

	claim, err := w.claimRepo.FindByID(claimID)
	if err != nil {
		logger.Error("failed to find claim", "error", err)
		return
	}
	if claim == nil {
		logger.Warn("claim not found, acknowledging orphan message")
		_ = w.redis.AckJob(ctx, messageID)
		return
	}

	if claim.Status != models.StatusPending {
		logger.Info("claim already processed", "status", claim.Status)
		_ = w.redis.AckJob(ctx, messageID)
		return
	}

	amountStr := strconv.FormatFloat(claim.Amount, 'f', -1, 64)
	network := config.BlockchainNetworks[blockchain]

	logger.Info("creating Busha quote", "amount", amountStr, "network", network)

	quote, err := w.busha.CreateQuote(blockchain, amountStr, walletAddress, network)
	if err != nil {
		logger.Error("busha quote failed", "error", err)
		_ = w.claimRepo.UpdateStatus(claimID, models.StatusFailed, "", "", err.Error())
		_ = w.redis.AckJob(ctx, messageID)
		return
	}

	logger.Info("quote created", "quote_id", quote.ID)

	transfer, err := w.busha.CreateTransfer(quote.ID)
	if err != nil {
		logger.Error("busha transfer failed", "error", err)
		_ = w.claimRepo.UpdateStatus(claimID, models.StatusFailed, quote.ID, "", err.Error())
		_ = w.redis.AckJob(ctx, messageID)
		return
	}

	logger.Info("transfer created", "transfer_id", transfer.ID, "status", transfer.Status)

	if err := w.claimRepo.UpdateStatus(claimID, models.StatusCompleted, quote.ID, transfer.ID, ""); err != nil {
		logger.Error("failed to update claim to COMPLETED", "error", err)

		return
	}

	if err := w.redis.AckJob(ctx, messageID); err != nil {
		logger.Error("failed to ACK job", "error", err)
	}

	logger.Info("faucet job completed successfully")
}
