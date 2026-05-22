package payment

type DomainEvent interface {
	eventName() string
}

type PaymentConfirmed struct {
	PaymentUUID string
	OrderUUID   string
}

func (PaymentConfirmed) eventName() string { return "payment confirmed" }
