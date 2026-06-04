package handler

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"faucet/internal/config"
	"faucet/internal/db"
	"faucet/internal/models"
	rds "faucet/internal/redis"
)

type FaucetHandler struct {
	cfg		*config.Config
	claimRepo	*db.ClaimRepository
	redis		*rds.Client
}

func NewFaucetHandler(cfg *config.Config, claimRepo *db.ClaimRepository, redis *rds.Client) *FaucetHandler {
	return &FaucetHandler{
		cfg:		cfg,
		claimRepo:	claimRepo,
		redis:		redis,
	}
}

type ClaimRequest struct {
	WalletAddress	string	`json:"wallet_address" binding:"required"`
	Blockchain	string	`json:"blockchain"      binding:"required"`
	Amount		float64	`json:"amount"          binding:"required,gt=0"`
}

type ClaimResponse struct {
	Data	ClaimData	`json:"data"`
	Message	string		`json:"message"`
}

type ClaimData struct {
	ID		string			`json:"id"`
	WalletAddress	string			`json:"wallet_address"`
	Blockchain	string			`json:"blockchain"`
	Amount		float64			`json:"amount"`
	Status		models.ClaimStatus	`json:"status"`
	CreatedAt	time.Time		`json:"created_at"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (h *FaucetHandler) HandleClaim(c *gin.Context) {
	var req ClaimRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "wallet_address, blockchain, and valid amount are required"})
		return
	}

	req.WalletAddress = strings.TrimSpace(req.WalletAddress)
	req.Blockchain = strings.ToUpper(strings.TrimSpace(req.Blockchain))

	if req.WalletAddress == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "wallet_address cannot be empty"})
		return
	}

	if !config.IsSupportedBlockchain(req.Blockchain) {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "unsupported blockchain. Supported: BTC, ETH, USDT, USDC, BNB",
		})
		return
	}

	maxAmountStr := h.cfg.MaxFaucetAmount(req.Blockchain)
	maxAmount, _ := strconv.ParseFloat(maxAmountStr, 64)
	if req.Amount > maxAmount {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "requested amount exceeds the maximum allowed for " + req.Blockchain + " (" + maxAmountStr + ")",
		})
		return
	}

	ctx := c.Request.Context()

	locked, err := h.redis.AcquireLock(ctx, req.WalletAddress)
	if err != nil {
		slog.Error("failed to acquire lock", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}
	if !locked {
		c.JSON(http.StatusTooManyRequests, ErrorResponse{
			Error: "request already in progress for this wallet. Please wait.",
		})
		return
	}

	defer func() {
		if err := h.redis.ReleaseLock(ctx, req.WalletAddress); err != nil {
			slog.Error("failed to release lock", "error", err, "wallet", req.WalletAddress)
		}
	}()

	since := time.Now().Add(-24 * time.Hour)
	count, err := h.claimRepo.CountRecentClaims(req.WalletAddress, since)
	if err != nil {
		slog.Error("failed to count recent claims", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}
	if count >= 1 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "you have already claimed from the faucet in the last 24 hours",
		})
		return
	}

	claim := &models.FaucetClaim{
		WalletAddress:	req.WalletAddress,
		Blockchain:	req.Blockchain,
		Amount:		req.Amount,
		Status:		models.StatusPending,
	}

	if err := h.claimRepo.Create(claim); err != nil {
		slog.Error("failed to create claim", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}

	slog.Info("claim created", "claim_id", claim.ID, "wallet", req.WalletAddress, "blockchain", req.Blockchain)

	if err := h.redis.PublishJob(ctx, claim.ID, req.WalletAddress, req.Blockchain); err != nil {
		slog.Error("failed to publish job", "error", err, "claim_id", claim.ID)

		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}

	c.JSON(http.StatusAccepted, ClaimResponse{
		Data: ClaimData{
			ID:		claim.ID,
			WalletAddress:	claim.WalletAddress,
			Blockchain:	claim.Blockchain,
			Amount:		claim.Amount,
			Status:		claim.Status,
			CreatedAt:	claim.CreatedAt,
		},
		Message:	"claim submitted successfully. Processing will begin shortly.",
	})
}

func (h *FaucetHandler) HandleGetClaim(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "claim id is required"})
		return
	}

	claim, err := h.claimRepo.FindByID(id)
	if err != nil {
		slog.Error("failed to find claim", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}
	if claim == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "claim not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": claim})
}

func (h *FaucetHandler) HandleListClaims(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 50
	}

	claims, err := h.claimRepo.ListRecent(limit)
	if err != nil {
		slog.Error("failed to list claims", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": claims})
}

func (h *FaucetHandler) HandleGetConfig(c *gin.Context) {
	maxAmounts := make(map[string]string)
	for _, bc := range config.SupportedBlockchains() {
		maxAmounts[bc] = h.cfg.MaxFaucetAmount(bc)
	}

	c.JSON(http.StatusOK, gin.H{
		"blockchains":	config.SupportedBlockchains(),
		"max_amounts":	maxAmounts,
	})
}
