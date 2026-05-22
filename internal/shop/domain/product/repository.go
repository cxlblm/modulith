package product

import "context"

type Repository interface {
	Save(ctx context.Context, p *Product) error
	FindByUUID(ctx context.Context, uuid ProductUUID) (*Product, error)
	ReserveStock(ctx context.Context, productUUID ProductUUID, orderUUID string, qty int) error
}
