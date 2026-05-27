package mysql

import (
	"context"
	"errors"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"modular_monolith/internal/order/app/command"
	orderdomain "modular_monolith/internal/order/domain/order"
	"modular_monolith/internal/platform/dbtx"
	"modular_monolith/internal/platform/eventbus"
)

type publishedEvent struct {
	eventType eventbus.EventType
	payload   any
}

type recordingBus struct {
	events []publishedEvent
}

var _ command.UnitOfWork = NewUnitOfWork(nil, nil)

func (b *recordingBus) Publish(_ context.Context, eventType eventbus.EventType, payload any) error {
	b.events = append(b.events, publishedEvent{eventType: eventType, payload: payload})
	return nil
}

func (b *recordingBus) Subscribe(eventbus.EventType, eventbus.Handler) {}

func TestOrderRepositorySave_DefaultsToAutomaticTransaction(t *testing.T) {
	ctx := context.Background()
	db := openOrderRepositoryTestDB(t)
	bus := &recordingBus{}
	repo := NewOrderRepository(db, bus)
	o := newRepositoryTestOrder(t)

	if err := repo.Save(ctx, o); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	if got := len(bus.events); got != 1 {
		t.Fatalf("published events = %d, want 1", got)
	}
	if got := len(o.PeekEvents()); got != 0 {
		t.Fatalf("remaining domain events = %d, want 0", got)
	}

	var count int64
	if err := db.Model(&OrderModel{}).Where("uuid = ?", o.UUID().String()).Count(&count).Error; err != nil {
		t.Fatalf("count order: %v", err)
	}
	if count != 1 {
		t.Fatalf("persisted orders = %d, want 1", count)
	}

	var item OrderItemModel
	if err := db.First(&item, "order_uuid = ?", o.UUID().String()).Error; err != nil {
		t.Fatalf("find item: %v", err)
	}
	if item.OriginalUnitPriceCents != 1000 || item.OriginalSubtotalCents != 2000 || item.DiscountCents != 300 || item.PayableCents != 1700 {
		t.Fatalf("item pricing = unit %d original %d discount %d payable %d, want 1000/2000/300/1700",
			item.OriginalUnitPriceCents,
			item.OriginalSubtotalCents,
			item.DiscountCents,
			item.PayableCents,
		)
	}
	if item.AppliedPromotionsJSON == "" {
		t.Fatal("AppliedPromotionsJSON is empty")
	}
}

func TestOrderRepositorySave_WithTxCollectsPendingEvents(t *testing.T) {
	ctx := context.Background()
	db := openOrderRepositoryTestDB(t)
	bus := &recordingBus{}
	tx := db.Begin()
	if tx.Error != nil {
		t.Fatalf("begin tx: %v", tx.Error)
	}
	collector := dbtx.NewPendingCollector()
	repo := NewOrderRepositoryWithTx(tx, collector, bus)
	o := newRepositoryTestOrder(t)

	if err := repo.Save(ctx, o); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if got := len(bus.events); got != 0 {
		t.Fatalf("published events before outer commit = %d, want 0", got)
	}
	if got := len(o.PeekEvents()); got == 0 {
		t.Fatal("domain events were cleared before outer commit")
	}

	if err := tx.Commit().Error; err != nil {
		t.Fatalf("commit tx: %v", err)
	}
	if err := collector.PublishAndClear(ctx, bus); err != nil {
		t.Fatalf("PublishAndClear() error = %v", err)
	}

	if got := len(bus.events); got != 1 {
		t.Fatalf("published events after outer commit = %d, want 1", got)
	}
	if got := len(o.PeekEvents()); got != 0 {
		t.Fatalf("remaining domain events = %d, want 0", got)
	}
}

func TestOrderUnitOfWorkRunInTx_PublishesAfterOuterCommit(t *testing.T) {
	ctx := context.Background()
	db := openOrderRepositoryTestDB(t)
	bus := &recordingBus{}
	uow := NewUnitOfWork(db, bus)
	o := newRepositoryTestOrder(t)

	err := uow.RunInTx(ctx, func(ctx context.Context, repos command.Repositories) error {
		return repos.Orders.Save(ctx, o)
	})
	if err != nil {
		t.Fatalf("RunInTx() error = %v", err)
	}

	if got := len(bus.events); got != 1 {
		t.Fatalf("published events = %d, want 1", got)
	}
	if got := len(o.PeekEvents()); got != 0 {
		t.Fatalf("remaining domain events = %d, want 0", got)
	}
}

func TestOrderUnitOfWorkRunInTx_RollsBackAndDoesNotPublishOnError(t *testing.T) {
	ctx := context.Background()
	db := openOrderRepositoryTestDB(t)
	bus := &recordingBus{}
	uow := NewUnitOfWork(db, bus)
	o := newRepositoryTestOrder(t)
	cause := errors.New("stop transaction")

	err := uow.RunInTx(ctx, func(ctx context.Context, repos command.Repositories) error {
		if err := repos.Orders.Save(ctx, o); err != nil {
			return err
		}
		return cause
	})
	if !errors.Is(err, cause) {
		t.Fatalf("RunInTx() error = %v, want %v", err, cause)
	}

	if got := len(bus.events); got != 0 {
		t.Fatalf("published events = %d, want 0", got)
	}
	if got := len(o.PeekEvents()); got == 0 {
		t.Fatal("domain events were cleared after rollback")
	}

	var count int64
	if err := db.Model(&OrderModel{}).Where("uuid = ?", o.UUID().String()).Count(&count).Error; err != nil {
		t.Fatalf("count order: %v", err)
	}
	if count != 0 {
		t.Fatalf("persisted orders = %d, want 0", count)
	}
}

func openOrderRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(Models()...); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	return db
}

func newRepositoryTestOrder(t *testing.T) *orderdomain.Order {
	t.Helper()

	o, err := orderdomain.NewOrder(
		"user-1",
		"address-1",
		orderdomain.AddressSnapshot{
			Receiver: "Ian",
			Phone:    "123456",
			City:     "Shanghai",
			Detail:   "Road 1",
		},
		[]orderdomain.Item{{
			ProductUUID:            "product-1",
			ProductName:            "Keyboard",
			OriginalUnitPriceCents: 1000,
			OriginalSubtotalCents:  2000,
			DiscountCents:          300,
			PayableCents:           1700,
			Qty:                    2,
			AppliedPromotions: []orderdomain.AppliedPromotion{{
				UUID:          "promo-1",
				Name:          "Spend 2000 save 300",
				DiscountCents: 300,
			}},
		}},
	)
	if err != nil {
		t.Fatalf("new order: %v", err)
	}
	return o
}
