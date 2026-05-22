package mysql

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"modular_monolith/internal/fulfillment/domain/shipment"
	fulfillmentmod "modular_monolith/internal/fulfillment/ports/module"
	"modular_monolith/internal/platform/dbtx"
	"modular_monolith/internal/platform/eventbus"
	platformmysql "modular_monolith/internal/platform/mysql"
)

type ShipmentModel struct {
	ID        uint64 `gorm:"primaryKey;autoIncrement;type:bigint unsigned"`
	UUID      string `gorm:"type:char(36);not null;uniqueIndex"`
	OrderUUID string `gorm:"type:char(36);not null;uniqueIndex"`
	Status    string `gorm:"size:32;not null;index"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ShipmentRepository struct {
	db      *gorm.DB
	bus     eventbus.Bus
	pending *dbtx.PendingCollector
}

func NewShipmentRepository(db *gorm.DB, bus eventbus.Bus) *ShipmentRepository {
	return &ShipmentRepository{db: db, bus: bus}
}

func NewShipmentRepositoryWithTx(tx *gorm.DB, pending *dbtx.PendingCollector, bus eventbus.Bus) *ShipmentRepository {
	if pending == nil {
		pending = dbtx.NewPendingCollector()
	}
	return &ShipmentRepository{db: tx, bus: bus, pending: pending}
}

func Models() []any {
	return []any{&ShipmentModel{}}
}

func (r *ShipmentRepository) CreateForOrder(ctx context.Context, s *shipment.Shipment) error {
	return r.executeWrite(ctx, "fulfillment.Repository.CreateForOrder", func(tx *gorm.DB) ([]*shipment.Shipment, error) {
		var existing ShipmentModel
		err := tx.First(&existing, "order_uuid = ?", s.OrderUUID()).Error
		if err == nil {
			return nil, nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("find shipment by order: %w", err)
		}
		model := ShipmentModel{UUID: s.UUID().String(), OrderUUID: s.OrderUUID(), Status: string(s.Status())}
		if err := tx.Create(&model).Error; err != nil {
			if platformmysql.IsDuplicateKey(err) {
				var existing ShipmentModel
				if findErr := tx.First(&existing, "order_uuid = ?", s.OrderUUID()).Error; findErr == nil {
					return nil, nil
				}
			}
			return nil, fmt.Errorf("create shipment: %w", err)
		}
		return []*shipment.Shipment{s}, nil
	})
}

func (r *ShipmentRepository) Send(ctx context.Context, uuid shipment.ShipmentUUID) error {
	return r.executeWrite(ctx, "fulfillment.Repository.Send", func(tx *gorm.DB) ([]*shipment.Shipment, error) {
		var model ShipmentModel
		if err := tx.First(&model, "uuid = ?", uuid.String()).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, shipment.NewShipmentNotFound(err)
			}
			return nil, fmt.Errorf("find shipment: %w", err)
		}
		s := shipment.Rehydrate(shipment.ShipmentUUID(model.UUID), model.OrderUUID, shipment.Status(model.Status))
		before := s.Status()
		if err := s.Send(); err != nil {
			return nil, err
		}
		if before == s.Status() {
			return nil, nil
		}
		if err := tx.Model(&ShipmentModel{}).Where("id = ?", model.ID).Update("status", string(s.Status())).Error; err != nil {
			return nil, fmt.Errorf("update shipment status: %w", err)
		}
		return []*shipment.Shipment{s}, nil
	})
}

func (r *ShipmentRepository) executeWrite(ctx context.Context, op string, fn func(tx *gorm.DB) ([]*shipment.Shipment, error)) error {
	return dbtx.ExecuteWriteWithEvents(ctx, r.db, r.pending, r.bus, op, fn, r.collectEvents)
}

func (r *ShipmentRepository) collectEvents(collector *dbtx.PendingCollector, shipments []*shipment.Shipment) error {
	for _, s := range shipments {
		if s == nil {
			continue
		}
		publishes, err := pendingPublishes(s)
		if err != nil {
			return err
		}
		collector.Add(s.ClearEvents, publishes...)
	}
	return nil
}

func pendingPublishes(s *shipment.Shipment) ([]dbtx.PendingPublish, error) {
	var publishes []dbtx.PendingPublish
	for _, domainEvent := range s.PeekEvents() {
		eventType, payload, ok := fulfillmentmod.Translate(domainEvent)
		if !ok {
			continue
		}
		if _, err := json.Marshal(payload); err != nil {
			return nil, fmt.Errorf("encode fulfillment integration event: %w", err)
		}
		publishes = append(publishes, dbtx.PendingPublish{
			EventType: eventbus.EventType(eventType),
			Payload:   payload,
		})
	}
	return publishes, nil
}
