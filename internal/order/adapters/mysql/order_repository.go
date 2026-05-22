package mysql

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	orderdomain "modular_monolith/internal/order/domain/order"
	ordermod "modular_monolith/internal/order/ports/module"
	"modular_monolith/internal/platform/dbtx"
	"modular_monolith/internal/platform/eventbus"
)

type OrderModel struct {
	ID         uint64 `gorm:"primaryKey;autoIncrement;type:bigint unsigned"`
	UUID       string `gorm:"type:char(36);not null;uniqueIndex"`
	UserID     string `gorm:"type:char(36);not null;index"`
	AddressID  string `gorm:"type:char(36);not null"`
	Receiver   string `gorm:"size:255;not null"`
	Phone      string `gorm:"size:64;not null"`
	City       string `gorm:"size:255;not null"`
	Detail     string `gorm:"size:512;not null"`
	Status     string `gorm:"size:32;not null;index"`
	TotalCents int64  `gorm:"not null"`
	PaymentID  string `gorm:"size:64"`
	ShipmentID string `gorm:"size:64"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type OrderItemModel struct {
	ID             uint64 `gorm:"primaryKey;autoIncrement;type:bigint unsigned"`
	OrderUUID      string `gorm:"type:char(36);not null;index"`
	ProductUUID    string `gorm:"type:char(36);not null"`
	ProductName    string `gorm:"size:255;not null"`
	UnitPriceCents int64  `gorm:"not null"`
	Qty            int    `gorm:"not null"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type OrderRepository struct {
	db      *gorm.DB
	bus     eventbus.Bus
	pending *dbtx.PendingCollector
}

func NewOrderRepository(db *gorm.DB, bus eventbus.Bus) *OrderRepository {
	return &OrderRepository{db: db, bus: bus}
}

func NewOrderRepositoryWithTx(tx *gorm.DB, pending *dbtx.PendingCollector, bus eventbus.Bus) *OrderRepository {
	if pending == nil {
		pending = dbtx.NewPendingCollector()
	}
	return &OrderRepository{db: tx, bus: bus, pending: pending}
}

func Models() []any {
	return []any{&OrderModel{}, &OrderItemModel{}}
}

func (r *OrderRepository) Save(ctx context.Context, o *orderdomain.Order) error {
	return r.executeWrite(ctx, "order.Repository.Save", func(tx *gorm.DB) ([]*orderdomain.Order, error) {
		model := orderModel(o)
		if err := tx.Create(&model).Error; err != nil {
			return nil, fmt.Errorf("create order: %w", err)
		}
		for _, item := range o.Items() {
			itemModel := OrderItemModel{
				OrderUUID:      o.UUID().String(),
				ProductUUID:    item.ProductUUID,
				ProductName:    item.ProductName,
				UnitPriceCents: item.UnitPriceCents,
				Qty:            item.Qty,
			}
			if err := tx.Create(&itemModel).Error; err != nil {
				return nil, fmt.Errorf("create order item: %w", err)
			}
		}
		return []*orderdomain.Order{o}, nil
	})
}

func (r *OrderRepository) FindByUUID(ctx context.Context, uuid orderdomain.OrderUUID) (*orderdomain.Order, error) {
	var model OrderModel
	if err := r.db.WithContext(ctx).First(&model, "uuid = ?", uuid.String()).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, orderdomain.NewOrderNotFound(err)
		}
		return nil, fmt.Errorf("find order: %w", err)
	}
	var itemModels []OrderItemModel
	if err := r.db.WithContext(ctx).Where("order_uuid = ?", model.UUID).Order("id").Find(&itemModels).Error; err != nil {
		return nil, fmt.Errorf("find order items: %w", err)
	}
	return toOrder(model, itemModels), nil
}

func (r *OrderRepository) MarkPaid(ctx context.Context, uuid orderdomain.OrderUUID, paymentUUID string) error {
	return r.executeWrite(ctx, "order.Repository.MarkPaid", func(tx *gorm.DB) ([]*orderdomain.Order, error) {
		o, err := findOrder(ctx, tx, uuid)
		if err != nil {
			return nil, err
		}
		before := o.Status()
		if err := o.MarkPaid(paymentUUID); err != nil {
			return nil, err
		}
		if before == o.Status() {
			return nil, nil
		}
		if err := tx.Model(&OrderModel{}).Where("uuid = ?", uuid.String()).Updates(map[string]any{
			"status":     string(o.Status()),
			"payment_id": o.PaymentUUID(),
		}).Error; err != nil {
			return nil, fmt.Errorf("update order paid: %w", err)
		}
		return []*orderdomain.Order{o}, nil
	})
}

func (r *OrderRepository) MarkShipped(ctx context.Context, uuid orderdomain.OrderUUID, shipmentUUID string) error {
	return r.executeWrite(ctx, "order.Repository.MarkShipped", func(tx *gorm.DB) ([]*orderdomain.Order, error) {
		o, err := findOrder(ctx, tx, uuid)
		if err != nil {
			return nil, err
		}
		before := o.Status()
		if err := o.MarkShipped(shipmentUUID); err != nil {
			return nil, err
		}
		if before == o.Status() {
			return nil, nil
		}
		if err := tx.Model(&OrderModel{}).Where("uuid = ?", uuid.String()).Updates(map[string]any{
			"status":      string(o.Status()),
			"shipment_id": o.ShipmentUUID(),
		}).Error; err != nil {
			return nil, fmt.Errorf("update order shipped: %w", err)
		}
		return []*orderdomain.Order{o}, nil
	})
}

func (r *OrderRepository) executeWrite(ctx context.Context, op string, fn func(tx *gorm.DB) ([]*orderdomain.Order, error)) error {
	return dbtx.ExecuteWriteWithEvents(ctx, r.db, r.pending, r.bus, op, fn, r.collectEvents)
}

func (r *OrderRepository) collectEvents(collector *dbtx.PendingCollector, orders []*orderdomain.Order) error {
	for _, o := range orders {
		if o == nil {
			continue
		}
		publishes, err := pendingPublishes(o)
		if err != nil {
			return err
		}
		collector.Add(o.ClearEvents, publishes...)
	}
	return nil
}

func pendingPublishes(o *orderdomain.Order) ([]dbtx.PendingPublish, error) {
	var publishes []dbtx.PendingPublish
	for _, domainEvent := range o.PeekEvents() {
		eventType, payload, ok := ordermod.Translate(domainEvent)
		if !ok {
			continue
		}
		if _, err := json.Marshal(payload); err != nil {
			return nil, fmt.Errorf("encode order integration event: %w", err)
		}
		publishes = append(publishes, dbtx.PendingPublish{
			EventType: eventbus.EventType(eventType),
			Payload:   payload,
		})
	}
	return publishes, nil
}

func findOrder(ctx context.Context, tx *gorm.DB, uuid orderdomain.OrderUUID) (*orderdomain.Order, error) {
	var model OrderModel
	if err := tx.WithContext(ctx).First(&model, "uuid = ?", uuid.String()).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, orderdomain.NewOrderNotFound(err)
		}
		return nil, fmt.Errorf("find order: %w", err)
	}
	var items []OrderItemModel
	if err := tx.WithContext(ctx).Where("order_uuid = ?", model.UUID).Order("id").Find(&items).Error; err != nil {
		return nil, fmt.Errorf("find order items: %w", err)
	}
	return toOrder(model, items), nil
}

func orderModel(o *orderdomain.Order) OrderModel {
	address := o.AddressSnapshot()
	return OrderModel{
		UUID:       o.UUID().String(),
		UserID:     o.UserUUID(),
		AddressID:  o.AddressUUID(),
		Receiver:   address.Receiver,
		Phone:      address.Phone,
		City:       address.City,
		Detail:     address.Detail,
		Status:     string(o.Status()),
		TotalCents: o.TotalCents(),
		PaymentID:  o.PaymentUUID(),
		ShipmentID: o.ShipmentUUID(),
	}
}

func toOrder(model OrderModel, itemModels []OrderItemModel) *orderdomain.Order {
	items := make([]orderdomain.Item, 0, len(itemModels))
	for _, item := range itemModels {
		items = append(items, orderdomain.Item{
			ProductUUID:    item.ProductUUID,
			ProductName:    item.ProductName,
			UnitPriceCents: item.UnitPriceCents,
			Qty:            item.Qty,
		})
	}
	return orderdomain.Rehydrate(
		orderdomain.OrderUUID(model.UUID),
		model.UserID,
		model.AddressID,
		orderdomain.AddressSnapshot{Receiver: model.Receiver, Phone: model.Phone, City: model.City, Detail: model.Detail},
		items,
		orderdomain.Status(model.Status),
		model.TotalCents,
		model.PaymentID,
		model.ShipmentID,
	)
}
