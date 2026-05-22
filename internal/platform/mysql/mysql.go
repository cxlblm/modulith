package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var ErrMissingDSN = errors.New("mysql dsn is required")

type Config struct {
	DSN             string
	AutoMigrate     bool
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

func (c Config) WithDefaults() Config {
	if c.MaxOpenConns == 0 {
		c.MaxOpenConns = 25
	}
	if c.MaxIdleConns == 0 {
		c.MaxIdleConns = 10
	}
	if c.ConnMaxLifetime == 0 {
		c.ConnMaxLifetime = 5 * time.Minute
	}
	if c.ConnMaxIdleTime == 0 {
		c.ConnMaxIdleTime = time.Minute
	}
	return c
}

func Open(ctx context.Context, cfg Config) (*gorm.DB, error) {
	cfg = cfg.WithDefaults()
	if strings.TrimSpace(cfg.DSN) == "" {
		return nil, ErrMissingDSN
	}

	db, err := gorm.Open(gormmysql.Open(cfg.DSN), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql db: %w", err)
	}

	applyPoolSettings(sqlDB, cfg)
	if err := sqlDB.PingContext(ctx); err != nil {
		closeErr := sqlDB.Close()
		if closeErr != nil {
			return nil, errors.Join(
				fmt.Errorf("ping mysql: %w", err),
				fmt.Errorf("close mysql after ping failure: %w", closeErr),
			)
		}
		return nil, fmt.Errorf("ping mysql: %w", err)
	}

	return db, nil
}

func Close(db *gorm.DB) error {
	if db == nil {
		return nil
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get sql db: %w", err)
	}
	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("close mysql: %w", err)
	}
	return nil
}

func applyPoolSettings(db *sql.DB, cfg Config) {
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
}
