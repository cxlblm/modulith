package shipment

import "modular_monolith/internal/shared/bizid"

type ShipmentUUID string

func (uuid ShipmentUUID) String() string { return string(uuid) }

type Status string

const (
	StatusPending Status = "pending"
	StatusSent    Status = "sent"
)

type Shipment struct {
	uuid      ShipmentUUID
	orderUUID string
	status    Status
	events    []DomainEvent
}

func NewShipment(orderUUID string) (*Shipment, error) {
	if orderUUID == "" {
		return nil, ErrInvalidShipment
	}
	return &Shipment{uuid: ShipmentUUID(bizid.New()), orderUUID: orderUUID, status: StatusPending}, nil
}

func Rehydrate(uuid ShipmentUUID, orderUUID string, status Status) *Shipment {
	return &Shipment{uuid: uuid, orderUUID: orderUUID, status: status}
}

func (s *Shipment) Send() error {
	if s.status == StatusSent {
		return nil
	}
	s.status = StatusSent
	s.events = append(s.events, ShipmentSent{ShipmentUUID: s.uuid.String(), OrderUUID: s.orderUUID})
	return nil
}

func (s *Shipment) UUID() ShipmentUUID        { return s.uuid }
func (s *Shipment) OrderUUID() string         { return s.orderUUID }
func (s *Shipment) Status() Status            { return s.status }
func (s *Shipment) PeekEvents() []DomainEvent { return append([]DomainEvent(nil), s.events...) }
func (s *Shipment) ClearEvents()              { s.events = nil }
