package mysql

import "time"

type RevivalCardModel struct {
	ID        uint64 `gorm:"primaryKey;autoIncrement;type:bigint unsigned"`
	UserUUID  string `gorm:"type:char(36);not null;uniqueIndex"`
	Balance   int    `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (RevivalCardModel) TableName() string {
	return "revival_card_models"
}

func Models() []any {
	return []any{
		&RevivalCardModel{},
	}
}
