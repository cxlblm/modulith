package command

import (
	"context"

	orderdomain "modular_monolith/internal/order/domain/order"
)

type UnitOfWork interface {
	RunInTx(ctx context.Context, fn func(ctx context.Context, repos Repositories) error) error
}

type Repositories struct {
	Orders orderdomain.Repository
}

type ProductsService interface {
	GetProduct(ctx context.Context, productID string) (ProductInfo, error)
	ReserveStock(ctx context.Context, productID string, orderID string, qty int) error
}

type ProductInfo struct {
	ID         string
	Name       string
	PriceCents int64
}

type AddressService interface {
	GetAddress(ctx context.Context, userID string, addressID string) (AddressInfo, error)
}

type AddressInfo struct {
	ID       string
	UserID   string
	Receiver string
	Phone    string
	City     string
	Detail   string
}
