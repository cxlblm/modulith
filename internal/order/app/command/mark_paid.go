package command

import (
	"context"
	"fmt"

	orderdomain "modular_monolith/internal/order/domain/order"
)

type MarkPaid struct {
	OrderID   string
	PaymentID string
}

type MarkPaidHandler struct {
	Orders orderdomain.Repository
}

func (h MarkPaidHandler) Handle(ctx context.Context, cmd MarkPaid) error {
	if err := h.Orders.MarkPaid(ctx, orderdomain.OrderUUID(cmd.OrderID), cmd.PaymentID); err != nil {
		return fmt.Errorf("mark order paid: %w", err)
	}
	return nil
}
