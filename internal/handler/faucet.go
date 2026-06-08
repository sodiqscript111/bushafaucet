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

type NetworkInfo struct {
	Network             string `json:"network"`
	Name                string `json:"name"`
	MinWithdrawalAmount string `json:"min_withdrawal_amount"`
	WithdrawalFee       string `json:"withdrawal_fee"`
}

type AssetInfo struct {
	Code      string        `json:"code"`
	Name      string        `json:"name"`
	MaxAmount string        `json:"max_amount"`
	Networks  []NetworkInfo `json:"networks"`
}

type FaucetHandler struct {
	cfg         *config.Config
	bushaClient *busha.Client

	lastClaim map[string]time.Time
	mu        sync.RWMutex

	cachedAssets    []AssetInfo
	cacheTime       time.Time
	cacheMu         sync.RWMutex
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
	Network       string  `json:"network"         binding:"required"`
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
	Network       string             `json:"network"`
	Amount        float64            `json:"amount"`
	Status        models.ClaimStatus `json:"status"`
	CreatedAt     time.Time          `json:"created_at"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (h *FaucetHandler) loadAssets() ([]AssetInfo, error) {
	h.cacheMu.RLock()
	if h.cachedAssets != nil && time.Since(h.cacheTime) < 5*time.Minute {
		assets := h.cachedAssets
		h.cacheMu.RUnlock()
		return assets, nil
	}
	h.cacheMu.RUnlock()

	currencies, err := h.bushaClient.GetCurrencies()
	if err != nil {
		return nil, err
	}

	supported := make(map[string]bool)
	for _, bc := range config.SupportedBlockchains() {
		supported[bc] = true
	}

	var assets []AssetInfo
	for _, cur := range currencies {
		if !supported[cur.Code] || cur.Type != "crypto" {
			continue
		}
		if len(cur.SupportedNetworks) == 0 {
			continue
		}

		var networks []NetworkInfo
		for _, n := range cur.SupportedNetworks {
			if n.Status != "active" || !n.Withdrawal {
				continue
			}
			networks = append(networks, NetworkInfo{
				Network:             n.Network,
				Name:                n.Name,
				MinWithdrawalAmount: n.MinWithdrawalAmount,
				WithdrawalFee:       n.WithdrawalFee,
			})
		}

		if len(networks) == 0 {
			continue
		}

		assets = append(assets, AssetInfo{
			Code:      cur.Code,
			Name:      cur.Name,
			MaxAmount: h.cfg.MaxFaucetAmount(cur.Code),
			Networks:  networks,
		})
	}

	h.cacheMu.Lock()
	h.cachedAssets = assets
	h.cacheTime = time.Now()
	h.cacheMu.Unlock()

	return assets, nil
}

func (h *FaucetHandler) HandleClaim(c *gin.Context) {
	var req ClaimRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "wallet_address, blockchain, network, and valid amount are required"})
		return
	}

	req.WalletAddress = strings.TrimSpace(req.WalletAddress)
	req.Blockchain = strings.ToUpper(strings.TrimSpace(req.Blockchain))
	req.Network = strings.ToUpper(strings.TrimSpace(req.Network))

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
		Network:       req.Network,
		Amount:        req.Amount,
		Status:        models.StatusPending,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	slog.Info("claim created", "claim_id", claim.ID, "wallet", req.WalletAddress, "blockchain", req.Blockchain, "network", req.Network)

	amountStr := strconv.FormatFloat(claim.Amount, 'f', -1, 64)

	slog.Info("creating Busha quote", "amount", amountStr, "network", req.Network)
	quote, err := h.bushaClient.CreateQuote(req.Blockchain, amountStr, req.WalletAddress, req.Network)
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

	h.mu.Lock()
	h.lastClaim[req.WalletAddress] = time.Now()
	h.mu.Unlock()

	c.JSON(http.StatusOK, ClaimResponse{
		Data: ClaimData{
			ID:            claim.ID,
			WalletAddress: claim.WalletAddress,
			Blockchain:    claim.Blockchain,
			Network:       claim.Network,
			Amount:        claim.Amount,
			Status:        claim.Status,
			CreatedAt:     claim.CreatedAt,
		},
		Message: "claim processed successfully.",
	})
}

func (h *FaucetHandler) HandleGetConfig(c *gin.Context) {
	assets, err := h.loadAssets()
	if err != nil {
		slog.Error("failed to load currencies from Busha", "error", err)

		maxAmounts := make(map[string]string)
		for _, bc := range config.SupportedBlockchains() {
			maxAmounts[bc] = h.cfg.MaxFaucetAmount(bc)
		}
		c.JSON(http.StatusOK, gin.H{
			"assets": []gin.H{},
			"blockchains": config.SupportedBlockchains(),
			"max_amounts": maxAmounts,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"assets": assets,
	})
}
