package eventsub

import (
	ordermod "modular_monolith/internal/order/ports/module"
	"modular_monolith/internal/payment/app"
	"modular_monolith/internal/platform/eventbus"
)

func RegisterAll(bus eventbus.Bus, app *app.Application) {
	bus.Subscribe(ordermod.OrderPlacedV1Type, NewOrderPlacedHandler(app))
}
