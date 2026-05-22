package module

import "modular_monolith/internal/fulfillment/domain/shipment"

func Translate(event shipment.DomainEvent) (eventType string, payload any, ok bool) {
	switch e := event.(type) {
	case shipment.ShipmentSent:
		return string(ShipmentSentV1Type), ShipmentSentV1{ShipmentID: e.ShipmentUUID, OrderID: e.OrderUUID}, true
	default:
		return "", nil, false
	}
}
