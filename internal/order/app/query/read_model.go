package query

import "context"

type ReadModel interface {
	GetOrder(ctx context.Context, orderID string) (OrderDTO, error)
	ListOrders(ctx context.Context, userID string) ([]OrderDTO, error)
}
