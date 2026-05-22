package eventsub

import (
	"context"

	"modular_monolith/internal/order/ports/module"
	"modular_monolith/internal/payment/app"
	"modular_monolith/internal/payment/app/command"
	"modular_monolith/internal/platform/eventbus"
)

type OrderPlacedHandler struct {
	app *app.Application
}

func NewOrderPlacedHandler(app *app.Application) *OrderPlacedHandler {
	return &OrderPlacedHandler{app: app}
}

func (h *OrderPlacedHandler) Handle(ctx context.Context, env eventbus.Envelope) error {
	var payload module.OrderPlacedV1
	if err := env.Decode(&payload); err != nil {
		return err
	}
	return h.app.Commands.InitializePayment.Handle(ctx, command.InitializePayment{OrderID: payload.OrderID, UserID: payload.UserID, TotalCents: payload.TotalCents})
}
