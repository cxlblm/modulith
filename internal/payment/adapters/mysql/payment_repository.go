package mysql

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	paymentdomain "modular_monolith/internal/payment/domain/payment"
	paymentmod "modular_monolith/internal/payment/ports/module"
	"modular_monolith/internal/platform/dbtx"
	"modular_monolith/internal/platform/eventbus"
	platformmysql "modular_monolith/internal/platform/mysql"
)

type PaymentModel struct {
	ID         uint64 `gorm:"primaryKey;autoIncrement;type:bigint unsigned"`
	UUID       string `gorm:"type:char(36);not null;uniqueIndex"`
	OrderUUID  string `gorm:"type:char(36);not null;uniqueIndex"`
	UserID     string `gorm:"type:char(36);not null;index"`
	TotalCents int64  `gorm:"not null"`
	Status     string `gorm:"size:32;not null;index"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type PaymentRepository struct {
	db      *gorm.DB
	bus     eventbus.Bus
	pending *dbtx.PendingCollector
}

func NewPaymentRepository(db *gorm.DB, bus eventbus.Bus) *PaymentRepository {
	return &PaymentRepository{db: db, bus: bus}
}

func NewPaymentRepositoryWithTx(tx *gorm.DB, pending *dbtx.PendingCollector, bus eventbus.Bus) *PaymentRepository {
	if pending == nil {
		pending = dbtx.NewPendingCollector()
	}
	return &PaymentRepository{db: tx, bus: bus, pending: pending}
}

func Models() []any {
	return []any{&PaymentModel{}}
}

func (r *PaymentRepository) CreateForOrder(ctx context.Context, p *paymentdomain.Payment) error {
	return r.executeWrite(ctx, "payment.Repository.CreateForOrder", func(tx *gorm.DB) ([]*paymentdomain.Payment, error) {
		var existing PaymentModel
		err := tx.First(&existing, "order_uuid = ?", p.OrderUUID()).Error
		if err == nil {
			return nil, nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("find payment by order: %w", err)
		}
		model := PaymentModel{
			UUID:       p.UUID().String(),
			OrderUUID:  p.OrderUUID(),
			UserID:     p.UserUUID(),
			TotalCents: p.TotalCents(),
			Status:     string(p.Status()),
		}
		if err := tx.Create(&model).Error; err != nil {
			if platformmysql.IsDuplicateKey(err) {
				var existing PaymentModel
				if findErr := tx.First(&existing, "order_uuid = ?", p.OrderUUID()).Error; findErr == nil {
					return nil, nil
				}
			}
			return nil, fmt.Errorf("create payment: %w", err)
		}
		return []*paymentdomain.Payment{p}, nil
	})
}

func (r *PaymentRepository) Confirm(ctx context.Context, uuid paymentdomain.PaymentUUID) error {
	return r.executeWrite(ctx, "payment.Repository.Confirm", func(tx *gorm.DB) ([]*paymentdomain.Payment, error) {
		var model PaymentModel
		if err := tx.First(&model, "uuid = ?", uuid.String()).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, paymentdomain.NewPaymentNotFound(err)
			}
			return nil, fmt.Errorf("find payment: %w", err)
		}
		p := paymentdomain.Rehydrate(
			paymentdomain.PaymentUUID(model.UUID),
			model.OrderUUID,
			model.UserID,
			model.TotalCents,
			paymentdomain.Status(model.Status),
		)
		before := p.Status()
		if err := p.Confirm(); err != nil {
			return nil, err
		}
		if before == p.Status() {
			return nil, nil
		}
		if err := tx.Model(&PaymentModel{}).Where("id = ?", model.ID).Update("status", string(p.Status())).Error; err != nil {
			return nil, fmt.Errorf("update payment status: %w", err)
		}
		return []*paymentdomain.Payment{p}, nil
	})
}

func (r *PaymentRepository) executeWrite(ctx context.Context, op string, fn func(tx *gorm.DB) ([]*paymentdomain.Payment, error)) error {
	return dbtx.ExecuteWriteWithEvents(ctx, r.db, r.pending, r.bus, op, fn, r.collectEvents)
}

func (r *PaymentRepository) collectEvents(collector *dbtx.PendingCollector, payments []*paymentdomain.Payment) error {
	for _, p := range payments {
		if p == nil {
			continue
		}
		publishes, err := pendingPublishes(p)
		if err != nil {
			return err
		}
		collector.Add(p.ClearEvents, publishes...)
	}
	return nil
}

func pendingPublishes(p *paymentdomain.Payment) ([]dbtx.PendingPublish, error) {
	var publishes []dbtx.PendingPublish
	for _, domainEvent := range p.PeekEvents() {
		eventType, payload, ok := paymentmod.Translate(domainEvent)
		if !ok {
			continue
		}
		if _, err := json.Marshal(payload); err != nil {
			return nil, fmt.Errorf("encode payment integration event: %w", err)
		}
		publishes = append(publishes, dbtx.PendingPublish{
			EventType: eventbus.EventType(eventType),
			Payload:   payload,
		})
	}
	return publishes, nil
}
