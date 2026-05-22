package shipment

import "context"

type Repository interface {
	CreateForOrder(ctx context.Context, s *Shipment) error
	Send(ctx context.Context, uuid ShipmentUUID) error
}
