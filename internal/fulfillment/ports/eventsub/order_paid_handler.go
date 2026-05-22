package eventsub

import (
	"context"

	"modular_monolith/internal/fulfillment/app"
	"modular_monolith/internal/fulfillment/app/command"
	ordermod "modular_monolith/internal/order/ports/module"
	"modular_monolith/internal/platform/eventbus"
)

type OrderPaidHandler struct {
	app *app.Application
}

func NewOrderPaidHandler(app *app.Application) *OrderPaidHandler {
	return &OrderPaidHandler{app: app}
}

func (h *OrderPaidHandler) Handle(ctx context.Context, env eventbus.Envelope) error {
	var payload ordermod.OrderPaidV1
	if err := env.Decode(&payload); err != nil {
		return err
	}
	return h.app.Commands.CreateShipment.Handle(ctx, command.CreateShipment{OrderID: payload.OrderID})
}
