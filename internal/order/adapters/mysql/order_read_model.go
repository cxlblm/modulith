package mysql

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"modular_monolith/internal/order/app/query"
	orderdomain "modular_monolith/internal/order/domain/order"
)

type ReadModel struct {
	db *gorm.DB
}

func NewReadModel(db *gorm.DB) *ReadModel {
	return &ReadModel{db: db}
}

func (r *ReadModel) GetOrder(ctx context.Context, orderID string) (query.OrderDTO, error) {
	var model OrderModel
	if err := r.db.WithContext(ctx).First(&model, "uuid = ?", orderID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return query.OrderDTO{}, orderdomain.NewOrderNotFound(err)
		}
		return query.OrderDTO{}, fmt.Errorf("get order: %w", err)
	}
	var items []OrderItemModel
	if err := r.db.WithContext(ctx).Where("order_uuid = ?", model.UUID).Order("id").Find(&items).Error; err != nil {
		return query.OrderDTO{}, fmt.Errorf("get order items: %w", err)
	}
	return orderDTO(model, items), nil
}

func (r *ReadModel) ListOrders(ctx context.Context, userID string) ([]query.OrderDTO, error) {
	var models []OrderModel
	db := r.db.WithContext(ctx).Order("id")
	if userID != "" {
		db = db.Where("user_id = ?", userID)
	}
	if err := db.Find(&models).Error; err != nil {
		return nil, fmt.Errorf("list orders: %w", err)
	}
	if len(models) == 0 {
		return []query.OrderDTO{}, nil
	}
	orderUUIDs := make([]string, 0, len(models))
	for _, model := range models {
		orderUUIDs = append(orderUUIDs, model.UUID)
	}
	var itemModels []OrderItemModel
	if err := r.db.WithContext(ctx).Where("order_uuid IN ?", orderUUIDs).Order("order_uuid, id").Find(&itemModels).Error; err != nil {
		return nil, fmt.Errorf("list order items: %w", err)
	}
	return orderDTOs(models, itemModels), nil
}

func orderDTOs(models []OrderModel, itemModels []OrderItemModel) []query.OrderDTO {
	itemsByOrder := make(map[string][]OrderItemModel, len(models))
	for _, item := range itemModels {
		itemsByOrder[item.OrderUUID] = append(itemsByOrder[item.OrderUUID], item)
	}
	orders := make([]query.OrderDTO, 0, len(models))
	for _, model := range models {
		orders = append(orders, orderDTO(model, itemsByOrder[model.UUID]))
	}
	return orders
}

func orderDTO(model OrderModel, itemModels []OrderItemModel) query.OrderDTO {
	items := make([]query.OrderItemDTO, 0, len(itemModels))
	for _, item := range itemModels {
		promotions, err := decodeQueryPromotions(item.AppliedPromotionsJSON)
		if err != nil {
			promotions = []query.AppliedPromotionDTO{}
		}
		unitPrice := item.OriginalUnitPriceCents
		if unitPrice == 0 {
			unitPrice = item.UnitPriceCents
		}
		items = append(items, query.OrderItemDTO{
			ProductID:              item.ProductUUID,
			ProductName:            item.ProductName,
			UnitPriceCents:         unitPrice,
			OriginalUnitPriceCents: unitPrice,
			OriginalSubtotalCents:  item.OriginalSubtotalCents,
			DiscountCents:          item.DiscountCents,
			PayableCents:           item.PayableCents,
			Qty:                    item.Qty,
			AppliedPromotions:      promotions,
			CreatedAt:              item.CreatedAt,
			UpdatedAt:              item.UpdatedAt,
		})
	}
	return query.OrderDTO{
		ID:         model.UUID,
		UserID:     model.UserID,
		AddressID:  model.AddressID,
		Status:     model.Status,
		TotalCents: model.TotalCents,
		PaymentID:  model.PaymentID,
		ShipmentID: model.ShipmentID,
		Address:    query.AddressDTO{Receiver: model.Receiver, Phone: model.Phone, City: model.City, Detail: model.Detail},
		Items:      items,
		CreatedAt:  model.CreatedAt,
		UpdatedAt:  model.UpdatedAt,
	}
}

func decodeQueryPromotions(payload string) ([]query.AppliedPromotionDTO, error) {
	promotions, err := decodePromotions(payload)
	if err != nil {
		return nil, err
	}
	out := make([]query.AppliedPromotionDTO, 0, len(promotions))
	for _, promo := range promotions {
		out = append(out, query.AppliedPromotionDTO{
			UUID:          promo.UUID,
			Name:          promo.Name,
			DiscountCents: promo.DiscountCents,
		})
	}
	return out, nil
}
