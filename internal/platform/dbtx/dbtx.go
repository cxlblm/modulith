package dbtx

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"modular_monolith/internal/platform/eventbus"
	"modular_monolith/internal/platform/txerr"
)

var ErrNestedTransaction = errors.New("nested transaction is not supported")

type txContextKey struct{}

type PendingPublish struct {
	EventType eventbus.EventType
	Payload   any
}

type PendingCollector struct {
	entries []pendingEntry
}

type RepoFactory[R any] func(tx *gorm.DB, pending *PendingCollector) R

type UnitOfWork[R any] struct {
	db         *gorm.DB
	op         string
	bus        eventbus.Bus
	buildRepos RepoFactory[R]
}

type pendingEntry struct {
	publishes []PendingPublish
	clear     func()
}

func NewUnitOfWork[R any](db *gorm.DB, op string, bus eventbus.Bus, buildRepos RepoFactory[R]) *UnitOfWork[R] {
	return &UnitOfWork[R]{
		db:         db,
		op:         op,
		bus:        bus,
		buildRepos: buildRepos,
	}
}

func NewPendingCollector() *PendingCollector {
	return &PendingCollector{}
}

func ExecuteWrite(ctx context.Context, db *gorm.DB, txBound bool, op string, fn func(tx *gorm.DB) error) error {
	if txBound {
		return fn(db.WithContext(ctx))
	}

	tx := db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("%s begin transaction: %w", op, tx.Error)
	}
	done := false
	defer func() {
		if !done {
			_ = tx.Rollback().Error
		}
	}()

	if err := fn(tx); err != nil {
		return err
	}
	if err := tx.Commit().Error; err != nil {
		done = true
		return &txerr.CommitOutcomeUnknownError{Op: op, Err: err}
	}
	done = true
	return nil
}

func ExecuteWriteWithEvents[T any](
	ctx context.Context,
	db *gorm.DB,
	pending *PendingCollector,
	bus eventbus.Bus,
	op string,
	fn func(tx *gorm.DB) (T, error),
	collect func(collector *PendingCollector, changed T) error,
) error {
	if pending != nil {
		changed, err := fn(db.WithContext(ctx))
		if err != nil {
			return err
		}
		return collect(pending, changed)
	}

	tx := db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("%s begin transaction: %w", op, tx.Error)
	}
	done := false
	defer func() {
		if !done {
			_ = tx.Rollback().Error
		}
	}()

	collector := NewPendingCollector()
	changed, err := fn(tx)
	if err != nil {
		return err
	}
	if err := collect(collector, changed); err != nil {
		return err
	}
	if err := tx.Commit().Error; err != nil {
		done = true
		return &txerr.CommitOutcomeUnknownError{Op: op, Err: err}
	}
	done = true
	if err := collector.PublishAndClear(ctx, bus); err != nil {
		return &txerr.PostCommitPublishError{Op: op, Err: err, Committed: true}
	}
	return nil
}

func (c *PendingCollector) Add(clear func(), publishes ...PendingPublish) {
	if len(publishes) == 0 {
		return
	}
	c.entries = append(c.entries, pendingEntry{
		publishes: append([]PendingPublish(nil), publishes...),
		clear:     clear,
	})
}

func (c *PendingCollector) PublishAndClear(ctx context.Context, bus eventbus.Bus) error {
	for _, entry := range c.entries {
		for _, publish := range entry.publishes {
			if err := bus.Publish(ctx, publish.EventType, publish.Payload); err != nil {
				return err
			}
		}
	}
	for _, entry := range c.entries {
		if entry.clear != nil {
			entry.clear()
		}
	}
	return nil
}

func (u *UnitOfWork[R]) RunInTx(ctx context.Context, fn func(ctx context.Context, repos R) error) error {
	if inTransaction(ctx) {
		return ErrNestedTransaction
	}

	tx := u.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("%s begin transaction: %w", u.op, tx.Error)
	}
	done := false
	defer func() {
		if !done {
			_ = tx.Rollback().Error
		}
	}()

	collector := NewPendingCollector()
	txCtx := context.WithValue(ctx, txContextKey{}, true)
	if err := fn(txCtx, u.buildRepos(tx, collector)); err != nil {
		return err
	}
	if err := tx.Commit().Error; err != nil {
		done = true
		return &txerr.CommitOutcomeUnknownError{Op: u.op, Err: err}
	}
	done = true
	if err := collector.PublishAndClear(ctx, u.bus); err != nil {
		return &txerr.PostCommitPublishError{Op: u.op, Err: err, Committed: true}
	}
	return nil
}

func inTransaction(ctx context.Context) bool {
	inTx, _ := ctx.Value(txContextKey{}).(bool)
	return inTx
}
