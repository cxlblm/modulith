package mysql

import (
	"context"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestClaimRepositoryClaimIsIdempotent(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(Models()...); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	repo := NewClaimRepository(db)

	already, err := repo.Claim(context.Background(), "contest-1", "user-1")
	if err != nil {
		t.Fatalf("first Claim() error = %v", err)
	}
	if already {
		t.Fatal("first Claim() already = true, want false")
	}

	already, err = repo.Claim(context.Background(), "contest-1", "user-1")
	if err != nil {
		t.Fatalf("second Claim() error = %v", err)
	}
	if !already {
		t.Fatal("second Claim() already = false, want true")
	}
}
