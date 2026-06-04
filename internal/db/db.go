package db

import (
	"context"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"faucet/internal/models"
)

func NewDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql db: %w", err)
	}

	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	return db, nil
}

func InitSchema(db *gorm.DB) error {
	if err := db.AutoMigrate(&models.FaucetClaim{}); err != nil {
		return fmt.Errorf("init schema: %w", err)
	}
	return nil
}

type ClaimRepository struct {
	db *gorm.DB
}

func NewClaimRepository(db *gorm.DB) *ClaimRepository {
	return &ClaimRepository{db: db}
}

func (r *ClaimRepository) Create(claim *models.FaucetClaim) error {
	if err := r.db.Create(claim).Error; err != nil {
		return fmt.Errorf("create claim: %w", err)
	}
	return nil
}

func (r *ClaimRepository) FindByID(id string) (*models.FaucetClaim, error) {
	claim := &models.FaucetClaim{}
	err := r.db.Where("id = ?", id).First(claim).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find claim: %w", err)
	}
	return claim, nil
}

func (r *ClaimRepository) UpdateStatus(id string, status models.ClaimStatus, quoteID, transferID, errMsg string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if quoteID != "" {
		updates["busha_quote_id"] = quoteID
	}
	if transferID != "" {
		updates["busha_transfer_id"] = transferID
	}
	if errMsg != "" {
		updates["error_message"] = errMsg
	}

	if err := r.db.Model(&models.FaucetClaim{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return fmt.Errorf("update claim status: %w", err)
	}
	return nil
}

func (r *ClaimRepository) CountRecentClaims(walletAddress string, since time.Time) (int64, error) {
	var count int64
	err := r.db.Model(&models.FaucetClaim{}).
		Where("wallet_address = ? AND status IN (?, ?) AND created_at > ?",
			walletAddress, models.StatusCompleted, models.StatusPending, since).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("count recent claims: %w", err)
	}
	return count, nil
}

func (r *ClaimRepository) ListRecent(limit int) ([]models.FaucetClaim, error) {
	var claims []models.FaucetClaim
	err := r.db.Order("created_at DESC").Limit(limit).Find(&claims).Error
	if err != nil {
		return nil, fmt.Errorf("list recent claims: %w", err)
	}
	return claims, nil
}
