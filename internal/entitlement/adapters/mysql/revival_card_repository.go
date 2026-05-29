package mysql

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RevivalCardRepository struct {
	db *gorm.DB
}

func NewRevivalCardRepository(db *gorm.DB) *RevivalCardRepository {
	return &RevivalCardRepository{db: db}
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

func (r *RevivalCardRepository) TryConsumeOne(ctx context.Context, userID string) (bool, error) {
	result := r.db.WithContext(ctx).Model(&RevivalCardModel{}).
		Where("user_uuid = ? AND balance > 0", userID).
		Update("balance", gorm.Expr("balance - 1"))
	if result.Error != nil {
		return false, fmt.Errorf("try consume revival card: %w", result.Error)
	}
	return result.RowsAffected == 1, nil
}
