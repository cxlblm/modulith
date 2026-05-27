package mysql

import (
	"context"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestPromotionRepository_ListActivePromotionsReturnsOnlyActive(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(Models()...); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	if err := db.Create(&PromotionModel{
		UUID:           "promo-active",
		Name:           "Spend 5000 save 999",
		ThresholdCents: 5000,
		DiscountCents:  999,
		Active:         true,
	}).Error; err != nil {
		t.Fatalf("create active promotion: %v", err)
	}
	if err := db.Create(&PromotionModel{
		UUID:           "promo-inactive",
		Name:           "Inactive",
		ThresholdCents: 1000,
		DiscountCents:  100,
		Active:         false,
	}).Error; err != nil {
		t.Fatalf("create inactive promotion: %v", err)
	}

	promotions, err := NewPromotionRepository(db).ListActivePromotions(context.Background())
	if err != nil {
		t.Fatalf("ListActivePromotions() error = %v", err)
	}

	if len(promotions) != 1 {
		t.Fatalf("promotions length = %d, want 1", len(promotions))
	}
	if promotions[0].UUID != "promo-active" {
		t.Fatalf("promotion UUID = %q, want promo-active", promotions[0].UUID)
	}
}
