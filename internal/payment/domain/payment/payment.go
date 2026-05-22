package payment

import "modular_monolith/internal/shared/bizid"

type PaymentUUID string

func (uuid PaymentUUID) String() string { return string(uuid) }

type Status string

const (
	StatusPending   Status = "pending"
	StatusConfirmed Status = "confirmed"
)

type Payment struct {
	uuid       PaymentUUID
	orderUUID  string
	userUUID   string
	totalCents int64
	status     Status
	events     []DomainEvent
}

func NewPayment(orderUUID string, userUUID string, totalCents int64) (*Payment, error) {
	if orderUUID == "" || userUUID == "" || totalCents <= 0 {
		return nil, ErrInvalidPayment
	}
	return &Payment{uuid: PaymentUUID(bizid.New()), orderUUID: orderUUID, userUUID: userUUID, totalCents: totalCents, status: StatusPending}, nil
}

func Rehydrate(uuid PaymentUUID, orderUUID string, userUUID string, totalCents int64, status Status) *Payment {
	return &Payment{uuid: uuid, orderUUID: orderUUID, userUUID: userUUID, totalCents: totalCents, status: status}
}

func (p *Payment) Confirm() error {
	if p.status == StatusConfirmed {
		return nil
	}
	p.status = StatusConfirmed
	p.events = append(p.events, PaymentConfirmed{PaymentUUID: p.uuid.String(), OrderUUID: p.orderUUID})
	return nil
}

func (p *Payment) UUID() PaymentUUID         { return p.uuid }
func (p *Payment) OrderUUID() string         { return p.orderUUID }
func (p *Payment) UserUUID() string          { return p.userUUID }
func (p *Payment) TotalCents() int64         { return p.totalCents }
func (p *Payment) Status() Status            { return p.status }
func (p *Payment) PeekEvents() []DomainEvent { return append([]DomainEvent(nil), p.events...) }
func (p *Payment) ClearEvents()              { p.events = nil }
