package mysql

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	platformmysql "modular_monolith/internal/platform/mysql"
)

type RewardClaimModel struct {
	ID          uint64 `gorm:"primaryKey;autoIncrement;type:bigint unsigned"`
	ContestUUID string `gorm:"type:char(36);not null;uniqueIndex:idx_reward_claim_contest_user"`
	UserUUID    string `gorm:"type:char(36);not null;uniqueIndex:idx_reward_claim_contest_user"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ClaimRepository struct {
	db *gorm.DB
}

func NewClaimRepository(db *gorm.DB) *ClaimRepository {
	return &ClaimRepository{db: db}
}

func Models() []any {
	return []any{&RewardClaimModel{}}
}

func (r *ClaimRepository) Claim(ctx context.Context, contestID string, userID string) (bool, error) {
	model := RewardClaimModel{ContestUUID: contestID, UserUUID: userID}
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		if platformmysql.IsDuplicateKey(err) || strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return true, nil
		}
		return false, fmt.Errorf("create reward claim: %w", err)
	}
	return false, nil
}
