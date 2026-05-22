package query

import "context"

type ReadModel interface {
	ListShipments(ctx context.Context, orderID string) ([]ShipmentDTO, error)
}
