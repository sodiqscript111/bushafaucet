package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ClaimStatus string

const (
	StatusPending	ClaimStatus	= "PENDING"
	StatusCompleted	ClaimStatus	= "COMPLETED"
	StatusFailed	ClaimStatus	= "FAILED"
)

type FaucetClaim struct {
	ID		string		`gorm:"type:uuid;primaryKey"           json:"id"`
	WalletAddress	string		`gorm:"type:varchar(255);index;not null" json:"wallet_address"`
	Blockchain	string		`gorm:"type:varchar(50);not null"      json:"blockchain"`
	Amount		float64		`gorm:"type:decimal(20,8);not null"    json:"amount"`
	Status		ClaimStatus	`gorm:"type:varchar(20);not null;default:'PENDING'" json:"status"`
	BushaQuoteID	*string		`gorm:"type:varchar(255);uniqueIndex"  json:"busha_quote_id,omitempty"`
	BushaTransferID	*string		`gorm:"type:varchar(255);uniqueIndex"  json:"busha_transfer_id,omitempty"`
	ErrorMessage	string		`gorm:"type:text"                      json:"error_message,omitempty"`
	CreatedAt	time.Time	`gorm:"autoCreateTime"                 json:"created_at"`
	UpdatedAt	time.Time	`gorm:"autoUpdateTime"                 json:"updated_at"`
}

func (c *FaucetClaim) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}

func (FaucetClaim) TableName() string {
	return "faucet_claims"
}
