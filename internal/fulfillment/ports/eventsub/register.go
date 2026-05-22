package eventsub

import (
	"modular_monolith/internal/fulfillment/app"
	ordermod "modular_monolith/internal/order/ports/module"
	"modular_monolith/internal/platform/eventbus"
)

func RegisterAll(bus eventbus.Bus, app *app.Application) {
	bus.Subscribe(ordermod.OrderPaidV1Type, NewOrderPaidHandler(app))
}
