package eventsub

import (
	fulfillmentmod "modular_monolith/internal/fulfillment/ports/module"
	"modular_monolith/internal/order/app"
	paymentmod "modular_monolith/internal/payment/ports/module"
	"modular_monolith/internal/platform/eventbus"
)

func RegisterAll(bus eventbus.Bus, app *app.Application) {
	bus.Subscribe(paymentmod.PaymentConfirmedV1Type, NewPaymentConfirmedHandler(app))
	bus.Subscribe(fulfillmentmod.ShipmentSentV1Type, NewShipmentSentHandler(app))
}
