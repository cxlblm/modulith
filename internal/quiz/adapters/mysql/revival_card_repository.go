package mysql

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"modular_monolith/internal/quiz/domain/participation"
)

type RevivalCardRepository struct {
	db *gorm.DB
}

func NewRevivalCardRepository(db *gorm.DB) *RevivalCardRepository {
	return &RevivalCardRepository{db: db}
}

func (r *RevivalCardRepository) Balance(ctx context.Context, userID string) (int, error) {
	var model RevivalCardModel
	if err := r.db.WithContext(ctx).First(&model, "user_uuid = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, fmt.Errorf("find revival cards: %w", err)
	}
	return model.Balance, nil
}

func (r *RevivalCardRepository) Grant(ctx context.Context, userID string, count int) error {
	model := RevivalCardModel{UserUUID: userID, Balance: count}
	if err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_uuid"}},
		DoUpdates: clause.Assignments(map[string]any{
			"balance":    gorm.Expr("balance + ?", count),
			"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
		}),
	}).Create(&model).Error; err != nil {
		return fmt.Errorf("grant revival card: %w", err)
	}
	return nil
}

func (r *RevivalCardRepository) ConsumeOne(ctx context.Context, userID string) error {
	result := r.db.WithContext(ctx).Model(&RevivalCardModel{}).
		Where("user_uuid = ? AND balance > 0", userID).
		Update("balance", gorm.Expr("balance - 1"))
	if result.Error != nil {
		return fmt.Errorf("consume revival card: %w", result.Error)
	}
	if result.RowsAffected != 1 {
		return participation.ErrInvalidParticipation
	}
	return nil
}
