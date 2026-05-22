package order

import "context"

type Repository interface {
	Save(ctx context.Context, o *Order) error
	FindByUUID(ctx context.Context, uuid OrderUUID) (*Order, error)
	MarkPaid(ctx context.Context, uuid OrderUUID, paymentUUID string) error
	MarkShipped(ctx context.Context, uuid OrderUUID, shipmentUUID string) error
}
