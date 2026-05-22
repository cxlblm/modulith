package module

import "modular_monolith/internal/platform/eventbus"

const PaymentConfirmedV1Type eventbus.EventType = "payment.PaymentConfirmedV1"

type PaymentConfirmedV1 struct {
	PaymentID string `json:"payment_id"`
	OrderID   string `json:"order_id"`
}
