package module

import "modular_monolith/internal/platform/eventbus"

const (
	OrderPlacedV1Type eventbus.EventType = "order.OrderPlacedV1"
	OrderPaidV1Type   eventbus.EventType = "order.OrderPaidV1"
)

type OrderPlacedV1 struct {
	OrderID    string `json:"order_id"`
	UserID     string `json:"user_id"`
	TotalCents int64  `json:"total_cents"`
}

type OrderPaidV1 struct {
	OrderID   string `json:"order_id"`
	PaymentID string `json:"payment_id"`
}
