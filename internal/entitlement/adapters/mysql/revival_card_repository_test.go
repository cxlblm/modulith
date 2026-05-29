package mysql

import (
	"context"
	"net/url"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRevivalCardRepository_GrantAccumulatesBalance(t *testing.T) {
	repo := NewRevivalCardRepository(openEntitlementRepositoryTestDB(t))
	ctx := context.Background()

	if err := repo.Grant(ctx, "user-1", 2); err != nil {
		t.Fatalf("first Grant() error = %v", err)
	}
	if err := repo.Grant(ctx, "user-1", 3); err != nil {
		t.Fatalf("second Grant() error = %v", err)
	}

	var model RevivalCardModel
	if err := repo.db.WithContext(ctx).First(&model, "user_uuid = ?", "user-1").Error; err != nil {
		t.Fatalf("find revival card model: %v", err)
	}
	if model.Balance != 5 {
		t.Fatalf("Balance = %d, want 5", model.Balance)
	}
}

func TestRevivalCardRepository_TryConsumeOne(t *testing.T) {
	tests := []struct {
		name           string
		initialBalance int
		grantFirst     bool
		wantConsumed   bool
		wantBalance    int
	}{
		{name: "decrements positive balance", initialBalance: 2, grantFirst: true, wantConsumed: true, wantBalance: 1},
		{name: "zero balance", initialBalance: 0, grantFirst: true, wantConsumed: false, wantBalance: 0},
		{name: "missing row", wantConsumed: false, wantBalance: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewRevivalCardRepository(openEntitlementRepositoryTestDB(t))
			ctx := context.Background()
			if tt.grantFirst {
				if err := repo.Grant(ctx, "user-1", tt.initialBalance); err != nil {
					t.Fatalf("Grant() error = %v", err)
				}
			}

			consumed, err := repo.TryConsumeOne(ctx, "user-1")
			if err != nil {
				t.Fatalf("TryConsumeOne() error = %v", err)
			}
			if consumed != tt.wantConsumed {
				t.Fatalf("consumed = %v, want %v", consumed, tt.wantConsumed)
			}

			var model RevivalCardModel
			err = repo.db.WithContext(ctx).First(&model, "user_uuid = ?", "user-1").Error
			if !tt.grantFirst {
				if err == nil {
					t.Fatal("missing row was created")
				}
				return
			}
			if err != nil {
				t.Fatalf("find revival card model: %v", err)
			}
			if model.Balance != tt.wantBalance {
				t.Fatalf("Balance = %d, want %d", model.Balance, tt.wantBalance)
			}
		})
	}
}

func openEntitlementRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := "file:" + url.QueryEscape(t.Name()) + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(Models()...); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	return db
}
