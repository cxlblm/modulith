package payment

import "context"

type Repository interface {
	CreateForOrder(ctx context.Context, p *Payment) error
	Confirm(ctx context.Context, uuid PaymentUUID) error
}
