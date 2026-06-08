package handler

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"faucet/internal/busha"
	"faucet/internal/config"
	"faucet/internal/models"
)

type FaucetHandler struct {
	cfg         *config.Config
	bushaClient *busha.Client

	// Simple in-memory rate limiter
	lastClaim map[string]time.Time
	mu        sync.RWMutex
}

func NewFaucetHandler(cfg *config.Config, bushaClient *busha.Client) *FaucetHandler {
	return &FaucetHandler{
		cfg:         cfg,
		bushaClient: bushaClient,
		lastClaim:   make(map[string]time.Time),
	}
}

type ClaimRequest struct {
	WalletAddress string  `json:"wallet_address" binding:"required"`
	Blockchain    string  `json:"blockchain"      binding:"required"`
	Amount        float64 `json:"amount"          binding:"required,gt=0"`
}

type ClaimResponse struct {
	Data    ClaimData `json:"data"`
	Message string    `json:"message"`
}

type ClaimData struct {
	ID            string             `json:"id"`
	WalletAddress string             `json:"wallet_address"`
	Blockchain    string             `json:"blockchain"`
	Amount        float64            `json:"amount"`
	Status        models.ClaimStatus `json:"status"`
	CreatedAt     time.Time          `json:"created_at"`
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

	// In-memory rate limiting
	h.mu.RLock()
	lastTime, exists := h.lastClaim[req.WalletAddress]
	h.mu.RUnlock()

	if exists && time.Since(lastTime) < 24*time.Hour {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "you have already claimed from the faucet in the last 24 hours",
		})
		return
	}

	claim := &models.FaucetClaim{
		ID:            uuid.New().String(),
		WalletAddress: req.WalletAddress,
		Blockchain:    req.Blockchain,
		Amount:        req.Amount,
		Status:        models.StatusPending,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	slog.Info("claim created", "claim_id", claim.ID, "wallet", req.WalletAddress, "blockchain", req.Blockchain)

	amountStr := strconv.FormatFloat(claim.Amount, 'f', -1, 64)
	network := config.BlockchainNetworks[req.Blockchain]

	slog.Info("creating Busha quote", "amount", amountStr, "network", network)
	quote, err := h.bushaClient.CreateQuote(req.Blockchain, amountStr, req.WalletAddress, network)
	if err != nil {
		slog.Error("busha quote failed", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to create quote: " + err.Error()})
		return
	}

	slog.Info("quote created", "quote_id", quote.ID)

	transfer, err := h.bushaClient.CreateTransfer(quote.ID)
	if err != nil {
		slog.Error("busha transfer failed", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to create transfer: " + err.Error()})
		return
	}

	slog.Info("transfer created", "transfer_id", transfer.ID, "status", transfer.Status)

	claim.Status = models.StatusCompleted

	// Update last claim time since it was successful
	h.mu.Lock()
	h.lastClaim[req.WalletAddress] = time.Now()
	h.mu.Unlock()

	c.JSON(http.StatusOK, ClaimResponse{
		Data: ClaimData{
			ID:            claim.ID,
			WalletAddress: claim.WalletAddress,
			Blockchain:    claim.Blockchain,
			Amount:        claim.Amount,
			Status:        claim.Status,
			CreatedAt:     claim.CreatedAt,
		},
		Message: "claim processed successfully.",
	})
}

func (h *FaucetHandler) HandleGetConfig(c *gin.Context) {
	maxAmounts := make(map[string]string)
	for _, bc := range config.SupportedBlockchains() {
		maxAmounts[bc] = h.cfg.MaxFaucetAmount(bc)
	}

	c.JSON(http.StatusOK, gin.H{
		"blockchains": config.SupportedBlockchains(),
		"max_amounts": maxAmounts,
	})
}
