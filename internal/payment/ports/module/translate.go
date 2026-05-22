package module

import paymentdomain "modular_monolith/internal/payment/domain/payment"

func Translate(event paymentdomain.DomainEvent) (eventType string, payload any, ok bool) {
	switch e := event.(type) {
	case paymentdomain.PaymentConfirmed:
		return string(PaymentConfirmedV1Type), PaymentConfirmedV1{PaymentID: e.PaymentUUID, OrderID: e.OrderUUID}, true
	default:
		return "", nil, false
	}
}
