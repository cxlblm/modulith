package command

import (
	"context"
	"fmt"

	orderdomain "modular_monolith/internal/order/domain/order"
)

type MarkShipped struct {
	OrderID    string
	ShipmentID string
}

type MarkShippedHandler struct {
	Orders orderdomain.Repository
}

func (h MarkShippedHandler) Handle(ctx context.Context, cmd MarkShipped) error {
	if err := h.Orders.MarkShipped(ctx, orderdomain.OrderUUID(cmd.OrderID), cmd.ShipmentID); err != nil {
		return fmt.Errorf("mark order shipped: %w", err)
	}
	return nil
}
