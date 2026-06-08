package models

import (
	"time"
)

type ClaimStatus string

const (
	StatusPending   ClaimStatus = "PENDING"
	StatusCompleted ClaimStatus = "COMPLETED"
	StatusFailed    ClaimStatus = "FAILED"
)

type FaucetClaim struct {
	ID              string      `json:"id"`
	WalletAddress   string      `json:"wallet_address"`
	Blockchain      string      `json:"blockchain"`
	Network         string      `json:"network"`
	Amount          float64     `json:"amount"`
	Status          ClaimStatus `json:"status"`
	BushaQuoteID    *string     `json:"busha_quote_id,omitempty"`
	BushaTransferID *string     `json:"busha_transfer_id,omitempty"`
	ErrorMessage    string      `json:"error_message,omitempty"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}
