package mysql

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestConfig_WithDefaults(t *testing.T) {
	cfg := Config{}.WithDefaults()

	if cfg.MaxOpenConns != 25 {
		t.Fatalf("MaxOpenConns = %d, want %d", cfg.MaxOpenConns, 25)
	}
	if cfg.MaxIdleConns != 10 {
		t.Fatalf("MaxIdleConns = %d, want %d", cfg.MaxIdleConns, 10)
	}
	if cfg.ConnMaxLifetime != 5*time.Minute {
		t.Fatalf("ConnMaxLifetime = %s, want %s", cfg.ConnMaxLifetime, 5*time.Minute)
	}
	if cfg.ConnMaxIdleTime != time.Minute {
		t.Fatalf("ConnMaxIdleTime = %s, want %s", cfg.ConnMaxIdleTime, time.Minute)
	}
}

func TestOpen_MissingDSN(t *testing.T) {
	_, err := Open(context.Background(), Config{})
	if !errors.Is(err, ErrMissingDSN) {
		t.Fatalf("Open() error = %v, want ErrMissingDSN", err)
	}
}
