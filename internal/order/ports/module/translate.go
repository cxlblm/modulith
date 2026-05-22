package module

import orderdomain "modular_monolith/internal/order/domain/order"

func Translate(event orderdomain.DomainEvent) (eventType string, payload any, ok bool) {
	switch e := event.(type) {
	case orderdomain.OrderPlaced:
		return string(OrderPlacedV1Type), OrderPlacedV1{OrderID: e.OrderUUID, UserID: e.UserUUID, TotalCents: e.TotalCents}, true
	case orderdomain.OrderPaid:
		return string(OrderPaidV1Type), OrderPaidV1{OrderID: e.OrderUUID, PaymentID: e.PaymentUUID}, true
	default:
		return "", nil, false
	}
}
