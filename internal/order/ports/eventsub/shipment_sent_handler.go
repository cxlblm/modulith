package eventsub

import (
	"context"

	fulfillmentmod "modular_monolith/internal/fulfillment/ports/module"
	"modular_monolith/internal/order/app"
	"modular_monolith/internal/order/app/command"
	"modular_monolith/internal/platform/eventbus"
)

type ShipmentSentHandler struct {
	app *app.Application
}

func NewShipmentSentHandler(app *app.Application) *ShipmentSentHandler {
	return &ShipmentSentHandler{app: app}
}

func (h *ShipmentSentHandler) Handle(ctx context.Context, env eventbus.Envelope) error {
	var payload fulfillmentmod.ShipmentSentV1
	if err := env.Decode(&payload); err != nil {
		return err
	}
	return h.app.Commands.MarkShipped.Handle(ctx, command.MarkShipped{OrderID: payload.OrderID, ShipmentID: payload.ShipmentID})
}
