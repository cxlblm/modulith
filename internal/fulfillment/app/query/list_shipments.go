package query

import "context"

type ListShipments struct {
	OrderID string
}

type ListShipmentsHandler struct {
	ReadModel ReadModel
}

func (h ListShipmentsHandler) Handle(ctx context.Context, q ListShipments) ([]ShipmentDTO, error) {
	return h.ReadModel.ListShipments(ctx, q.OrderID)
}
