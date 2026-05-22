package order

import "modular_monolith/internal/shared/bizid"

type OrderUUID string

func (uuid OrderUUID) String() string { return string(uuid) }

type Status string

const (
	StatusPlaced  Status = "placed"
	StatusPaid    Status = "paid"
	StatusShipped Status = "shipped"
)

type AddressSnapshot struct {
	Receiver string
	Phone    string
	City     string
	Detail   string
}

type Item struct {
	ProductUUID    string
	ProductName    string
	UnitPriceCents int64
	Qty            int
}

type Order struct {
	uuid            OrderUUID
	userUUID        string
	addressUUID     string
	addressSnapshot AddressSnapshot
	items           []Item
	status          Status
	totalCents      int64
	paymentUUID     string
	shipmentUUID    string
	events          []DomainEvent
}

func NewOrder(userUUID string, addressUUID string, address AddressSnapshot, items []Item) (*Order, error) {
	if userUUID == "" || addressUUID == "" || len(items) == 0 {
		return nil, ErrInvalidOrder
	}
	if address.Receiver == "" || address.Phone == "" || address.City == "" || address.Detail == "" {
		return nil, ErrInvalidOrder
	}
	var total int64
	for _, item := range items {
		if item.ProductUUID == "" || item.ProductName == "" || item.UnitPriceCents <= 0 || item.Qty <= 0 {
			return nil, ErrInvalidOrder
		}
		total += item.UnitPriceCents * int64(item.Qty)
	}
	o := &Order{
		uuid:            OrderUUID(bizid.New()),
		userUUID:        userUUID,
		addressUUID:     addressUUID,
		addressSnapshot: address,
		items:           append([]Item(nil), items...),
		status:          StatusPlaced,
		totalCents:      total,
	}
	o.events = append(o.events, OrderPlaced{OrderUUID: o.uuid.String(), UserUUID: userUUID, TotalCents: total})
	return o, nil
}

func Rehydrate(uuid OrderUUID, userUUID string, addressUUID string, address AddressSnapshot, items []Item, status Status, totalCents int64, paymentUUID string, shipmentUUID string) *Order {
	return &Order{
		uuid:            uuid,
		userUUID:        userUUID,
		addressUUID:     addressUUID,
		addressSnapshot: address,
		items:           append([]Item(nil), items...),
		status:          status,
		totalCents:      totalCents,
		paymentUUID:     paymentUUID,
		shipmentUUID:    shipmentUUID,
	}
}

func (o *Order) MarkPaid(paymentUUID string) error {
	if paymentUUID == "" {
		return ErrInvalidOrder
	}
	if o.status == StatusPaid || o.status == StatusShipped {
		return nil
	}
	o.status = StatusPaid
	o.paymentUUID = paymentUUID
	o.events = append(o.events, OrderPaid{OrderUUID: o.uuid.String(), PaymentUUID: paymentUUID})
	return nil
}

func (o *Order) MarkShipped(shipmentUUID string) error {
	if shipmentUUID == "" {
		return ErrInvalidOrder
	}
	if o.status == StatusShipped {
		return nil
	}
	o.status = StatusShipped
	o.shipmentUUID = shipmentUUID
	return nil
}

func (o *Order) UUID() OrderUUID                  { return o.uuid }
func (o *Order) UserUUID() string                 { return o.userUUID }
func (o *Order) AddressUUID() string              { return o.addressUUID }
func (o *Order) AddressSnapshot() AddressSnapshot { return o.addressSnapshot }
func (o *Order) Items() []Item                    { return append([]Item(nil), o.items...) }
func (o *Order) Status() Status                   { return o.status }
func (o *Order) TotalCents() int64                { return o.totalCents }
func (o *Order) PaymentUUID() string              { return o.paymentUUID }
func (o *Order) ShipmentUUID() string             { return o.shipmentUUID }
func (o *Order) PeekEvents() []DomainEvent        { return append([]DomainEvent(nil), o.events...) }
func (o *Order) ClearEvents()                     { o.events = nil }
