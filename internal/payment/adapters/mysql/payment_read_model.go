package mysql

import (
	"context"
	"fmt"

	"modular_monolith/internal/payment/app/query"

	"gorm.io/gorm"
)

type ReadModel struct {
	db *gorm.DB
}

func NewReadModel(db *gorm.DB) *ReadModel {
	return &ReadModel{db: db}
}

func (r *ReadModel) ListPayments(ctx context.Context, orderID string) ([]query.PaymentDTO, error) {
	var models []PaymentModel
	db := r.db.WithContext(ctx).Order("id")
	if orderID != "" {
		db = db.Where("order_uuid = ?", orderID)
	}
	if err := db.Find(&models).Error; err != nil {
		return nil, fmt.Errorf("list payments: %w", err)
	}
	items := make([]query.PaymentDTO, 0, len(models))
	for _, model := range models {
		items = append(items, query.PaymentDTO{
			ID:         model.UUID,
			OrderID:    model.OrderUUID,
			UserID:     model.UserID,
			TotalCents: model.TotalCents,
			Status:     model.Status,
			CreatedAt:  model.CreatedAt,
			UpdatedAt:  model.UpdatedAt,
		})
	}
	return items, nil
}
