package mysql

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"modular_monolith/internal/pricing/app/command"
)

type PromotionModel struct {
	ID             uint64 `gorm:"primaryKey;autoIncrement;type:bigint unsigned"`
	UUID           string `gorm:"type:char(36);not null;uniqueIndex"`
	Name           string `gorm:"size:255;not null"`
	ThresholdCents int64  `gorm:"not null;index"`
	DiscountCents  int64  `gorm:"not null"`
	Active         bool   `gorm:"not null;index"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type PromotionRepository struct {
	db *gorm.DB
}

func NewPromotionRepository(db *gorm.DB) *PromotionRepository {
	return &PromotionRepository{db: db}
}

func Models() []any {
	return []any{&PromotionModel{}}
}

func (r *PromotionRepository) ListActivePromotions(ctx context.Context) ([]command.PromotionInfo, error) {
	var models []PromotionModel
	if err := r.db.WithContext(ctx).Where("active = ?", true).Order("id").Find(&models).Error; err != nil {
		return nil, fmt.Errorf("list active promotions: %w", err)
	}
	promotions := make([]command.PromotionInfo, 0, len(models))
	for _, model := range models {
		promotions = append(promotions, command.PromotionInfo{
			UUID:           model.UUID,
			Name:           model.Name,
			ThresholdCents: model.ThresholdCents,
			DiscountCents:  model.DiscountCents,
			Active:         model.Active,
		})
	}
	return promotions, nil
}
