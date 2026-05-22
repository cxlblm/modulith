package shipment

type DomainEvent interface {
	eventName() string
}

type ShipmentSent struct {
	ShipmentUUID string
	OrderUUID    string
}

func (ShipmentSent) eventName() string { return "shipment sent" }
