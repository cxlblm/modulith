package dbtx

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"modular_monolith/internal/platform/eventbus"
)

type testBus struct{}

func (testBus) Publish(context.Context, eventbus.EventType, any) error {
	return nil
}

func (testBus) Subscribe(eventbus.EventType, eventbus.Handler) {}

type recordingBus struct {
	published []eventbus.EventType
}

func (b *recordingBus) Publish(_ context.Context, eventType eventbus.EventType, _ any) error {
	b.published = append(b.published, eventType)
	return nil
}

func (*recordingBus) Subscribe(eventbus.EventType, eventbus.Handler) {}

type txRecord struct {
	ID   uint64 `gorm:"primaryKey"`
	Name string
}

func TestUnitOfWorkRunInTxBuildsRepositories(t *testing.T) {
	ctx := context.Background()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	uow := NewUnitOfWork[string](db, "test.UnitOfWork.RunInTx", testBus{}, func(tx *gorm.DB, pending *PendingCollector) string {
		if tx == nil {
			t.Fatal("tx is nil")
		}
		if pending == nil {
			t.Fatal("pending collector is nil")
		}
		return "repos"
	})

	err = uow.RunInTx(ctx, func(_ context.Context, repos string) error {
		if repos != "repos" {
			t.Fatalf("repos = %q, want repos", repos)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("RunInTx() error = %v", err)
	}
}

func TestExecuteWriteCommitsWrite(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	err := ExecuteWrite(ctx, db, false, "test.ExecuteWrite", func(tx *gorm.DB) error {
		return tx.Create(&txRecord{Name: "committed"}).Error
	})
	if err != nil {
		t.Fatalf("ExecuteWrite() error = %v", err)
	}

	var count int64
	if err := db.Model(&txRecord{}).Where("name = ?", "committed").Count(&count).Error; err != nil {
		t.Fatalf("count records: %v", err)
	}
	if count != 1 {
		t.Fatalf("count = %d, want 1", count)
	}
}

func TestExecuteWriteRollsBackWhenFunctionFails(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)
	wantErr := errors.New("write failed")

	err := ExecuteWrite(ctx, db, false, "test.ExecuteWrite", func(tx *gorm.DB) error {
		if err := tx.Create(&txRecord{Name: "rolled-back"}).Error; err != nil {
			return err
		}
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("ExecuteWrite() error = %v, want %v", err, wantErr)
	}

	var count int64
	if err := db.Model(&txRecord{}).Where("name = ?", "rolled-back").Count(&count).Error; err != nil {
		t.Fatalf("count records: %v", err)
	}
	if count != 0 {
		t.Fatalf("count = %d, want 0", count)
	}
}

func TestExecuteWriteUsesExistingTransactionWhenTxBound(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)
	tx := db.Begin()
	if tx.Error != nil {
		t.Fatalf("begin tx: %v", tx.Error)
	}

	err := ExecuteWrite(ctx, tx, true, "test.ExecuteWrite", func(tx *gorm.DB) error {
		return tx.Create(&txRecord{Name: "outer-tx"}).Error
	})
	if err != nil {
		t.Fatalf("ExecuteWrite() error = %v", err)
	}

	if err := tx.Rollback().Error; err != nil {
		t.Fatalf("rollback outer tx: %v", err)
	}

	var countAfterRollback int64
	if err := db.Model(&txRecord{}).Where("name = ?", "outer-tx").Count(&countAfterRollback).Error; err != nil {
		t.Fatalf("count after rollback: %v", err)
	}
	if countAfterRollback != 0 {
		t.Fatalf("count after outer rollback = %d, want 0", countAfterRollback)
	}
}

func TestExecuteWriteWithEventsPublishesAndClearsAfterCommit(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)
	bus := &recordingBus{}
	cleared := false

	err := ExecuteWriteWithEvents(ctx, db, nil, bus, "test.ExecuteWriteWithEvents",
		func(tx *gorm.DB) ([]string, error) {
			if err := tx.Create(&txRecord{Name: "with-events"}).Error; err != nil {
				return nil, err
			}
			return []string{"changed"}, nil
		},
		func(collector *PendingCollector, changed []string) error {
			if len(changed) != 1 {
				t.Fatalf("changed length = %d, want 1", len(changed))
			}
			collector.Add(func() { cleared = true }, PendingPublish{EventType: "record.created", Payload: "payload"})
			return nil
		},
	)
	if err != nil {
		t.Fatalf("ExecuteWriteWithEvents() error = %v", err)
	}
	if len(bus.published) != 1 || bus.published[0] != "record.created" {
		t.Fatalf("published = %v, want [record.created]", bus.published)
	}
	if !cleared {
		t.Fatal("event clear callback was not called")
	}
}

func TestExecuteWriteWithEventsRollsBackWhenCollectFails(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)
	bus := &recordingBus{}
	wantErr := errors.New("collect failed")

	err := ExecuteWriteWithEvents(ctx, db, nil, bus, "test.ExecuteWriteWithEvents",
		func(tx *gorm.DB) ([]string, error) {
			if err := tx.Create(&txRecord{Name: "collect-failed"}).Error; err != nil {
				return nil, err
			}
			return []string{"changed"}, nil
		},
		func(*PendingCollector, []string) error {
			return wantErr
		},
	)
	if !errors.Is(err, wantErr) {
		t.Fatalf("ExecuteWriteWithEvents() error = %v, want %v", err, wantErr)
	}
	if len(bus.published) != 0 {
		t.Fatalf("published = %v, want none", bus.published)
	}

	var count int64
	if err := db.Model(&txRecord{}).Where("name = ?", "collect-failed").Count(&count).Error; err != nil {
		t.Fatalf("count records: %v", err)
	}
	if count != 0 {
		t.Fatalf("count = %d, want 0", count)
	}
}

func TestExecuteWriteWithEventsCollectsPendingWithoutPublishing(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)
	bus := &recordingBus{}
	pending := NewPendingCollector()
	cleared := false

	err := ExecuteWriteWithEvents(ctx, db, pending, bus, "test.ExecuteWriteWithEvents",
		func(tx *gorm.DB) ([]string, error) {
			return []string{"changed"}, tx.Create(&txRecord{Name: "pending"}).Error
		},
		func(collector *PendingCollector, changed []string) error {
			if len(changed) != 1 {
				t.Fatalf("changed length = %d, want 1", len(changed))
			}
			collector.Add(func() { cleared = true }, PendingPublish{EventType: "record.pending", Payload: "payload"})
			return nil
		},
	)
	if err != nil {
		t.Fatalf("ExecuteWriteWithEvents() error = %v", err)
	}
	if len(bus.published) != 0 {
		t.Fatalf("published = %v, want none", bus.published)
	}
	if cleared {
		t.Fatal("event clear callback was called before pending publish")
	}

	if err := pending.PublishAndClear(ctx, bus); err != nil {
		t.Fatalf("PublishAndClear() error = %v", err)
	}
	if len(bus.published) != 1 || bus.published[0] != "record.pending" {
		t.Fatalf("published = %v, want [record.pending]", bus.published)
	}
	if !cleared {
		t.Fatal("event clear callback was not called")
	}
}

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	name := strings.ReplaceAll(t.Name(), "/", "_")
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", name)), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&txRecord{}); err != nil {
		t.Fatalf("migrate txRecord: %v", err)
	}
	return db
}
