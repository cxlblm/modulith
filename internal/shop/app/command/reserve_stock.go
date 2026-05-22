package command

import (
	"context"
	"fmt"

	"modular_monolith/internal/shop/domain/product"
)

type ReserveStock struct {
	ProductID string
	OrderID   string
	Qty       int
}

type ReserveStockHandler struct {
	Products product.Repository
}

func (h ReserveStockHandler) Handle(ctx context.Context, cmd ReserveStock) error {
	if err := h.Products.ReserveStock(ctx, product.ProductUUID(cmd.ProductID), cmd.OrderID, cmd.Qty); err != nil {
		return fmt.Errorf("reserve stock: %w", err)
	}
	return nil
}
