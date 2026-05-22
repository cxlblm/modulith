package command

import (
	"context"
	"fmt"

	paymentdomain "modular_monolith/internal/payment/domain/payment"
)

type InitializePayment struct {
	OrderID    string
	UserID     string
	TotalCents int64
}

type InitializePaymentHandler struct {
	Payments paymentdomain.Repository
}

func (h InitializePaymentHandler) Handle(ctx context.Context, cmd InitializePayment) error {
	p, err := paymentdomain.NewPayment(cmd.OrderID, cmd.UserID, cmd.TotalCents)
	if err != nil {
		return err
	}
	if err := h.Payments.CreateForOrder(ctx, p); err != nil {
		return fmt.Errorf("create payment for order: %w", err)
	}
	return nil
}
