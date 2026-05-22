package query

import "context"

type ReadModel interface {
	ListPayments(ctx context.Context, orderID string) ([]PaymentDTO, error)
}
