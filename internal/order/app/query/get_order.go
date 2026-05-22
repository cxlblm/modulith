package query

import "context"

type GetOrder struct {
	OrderID string
}

type GetOrderHandler struct {
	ReadModel ReadModel
}

func (h GetOrderHandler) Handle(ctx context.Context, q GetOrder) (OrderDTO, error) {
	return h.ReadModel.GetOrder(ctx, q.OrderID)
}
