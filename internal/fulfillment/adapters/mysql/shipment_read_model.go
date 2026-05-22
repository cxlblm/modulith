package mysql

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"modular_monolith/internal/fulfillment/app/query"
)

type ReadModel struct {
	db *gorm.DB
}

func NewReadModel(db *gorm.DB) *ReadModel {
	return &ReadModel{db: db}
}

func (r *ReadModel) ListShipments(ctx context.Context, orderID string) ([]query.ShipmentDTO, error) {
	var models []ShipmentModel
	db := r.db.WithContext(ctx).Order("id")
	if orderID != "" {
		db = db.Where("order_uuid = ?", orderID)
	}
	if err := db.Find(&models).Error; err != nil {
		return nil, fmt.Errorf("list shipments: %w", err)
	}
	items := make([]query.ShipmentDTO, 0, len(models))
	for _, model := range models {
		items = append(items, query.ShipmentDTO{
			ID:        model.UUID,
			OrderID:   model.OrderUUID,
			Status:    model.Status,
			CreatedAt: model.CreatedAt,
			UpdatedAt: model.UpdatedAt,
		})
	}
	return items, nil
}
