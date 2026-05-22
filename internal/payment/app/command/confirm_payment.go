package command

import (
	"context"
	"fmt"

	paymentdomain "modular_monolith/internal/payment/domain/payment"
)

type ConfirmPayment struct {
	PaymentID string
}

type ConfirmPaymentHandler struct {
	Payments paymentdomain.Repository
}

func (h ConfirmPaymentHandler) Handle(ctx context.Context, cmd ConfirmPayment) error {
	if err := h.Payments.Confirm(ctx, paymentdomain.PaymentUUID(cmd.PaymentID)); err != nil {
		return fmt.Errorf("confirm payment: %w", err)
	}
	return nil
}
