package eventsub

import (
	"context"

	"modular_monolith/internal/order/app"
	"modular_monolith/internal/order/app/command"
	paymentmod "modular_monolith/internal/payment/ports/module"
	"modular_monolith/internal/platform/eventbus"
)

type PaymentConfirmedHandler struct {
	app *app.Application
}

func NewPaymentConfirmedHandler(app *app.Application) *PaymentConfirmedHandler {
	return &PaymentConfirmedHandler{app: app}
}

func (h *PaymentConfirmedHandler) Handle(ctx context.Context, env eventbus.Envelope) error {
	var payload paymentmod.PaymentConfirmedV1
	if err := env.Decode(&payload); err != nil {
		return err
	}
	return h.app.Commands.MarkPaid.Handle(ctx, command.MarkPaid{OrderID: payload.OrderID, PaymentID: payload.PaymentID})
}
