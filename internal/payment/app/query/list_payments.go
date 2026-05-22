package query

import "context"

type ListPayments struct {
	OrderID string
}

type ListPaymentsHandler struct {
	ReadModel ReadModel
}

func (h ListPaymentsHandler) Handle(ctx context.Context, q ListPayments) ([]PaymentDTO, error) {
	return h.ReadModel.ListPayments(ctx, q.OrderID)
}
