package module

import "modular_monolith/internal/platform/eventbus"

const ShipmentSentV1Type eventbus.EventType = "fulfillment.ShipmentSentV1"

type ShipmentSentV1 struct {
	ShipmentID string `json:"shipment_id"`
	OrderID    string `json:"order_id"`
}
