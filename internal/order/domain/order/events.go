package order

type DomainEvent interface {
	eventName() string
}

type OrderPlaced struct {
	OrderUUID  string
	UserUUID   string
	TotalCents int64
}

func (OrderPlaced) eventName() string { return "order placed" }

type OrderPaid struct {
	OrderUUID   string
	PaymentUUID string
}

func (OrderPaid) eventName() string { return "order paid" }
