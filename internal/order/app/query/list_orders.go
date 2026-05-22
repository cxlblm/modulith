package query

import "context"

type ListOrders struct {
	UserID string
}

type ListOrdersHandler struct {
	ReadModel ReadModel
}

func (h ListOrdersHandler) Handle(ctx context.Context, q ListOrders) ([]OrderDTO, error) {
	return h.ReadModel.ListOrders(ctx, q.UserID)
}
