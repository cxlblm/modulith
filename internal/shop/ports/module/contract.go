package module

import (
	"context"

	"modular_monolith/internal/shop/app"
	"modular_monolith/internal/shop/app/command"
	"modular_monolith/internal/shop/app/query"
)

type ShopModule interface {
	GetProduct(ctx context.Context, productID string) (ProductDTO, error)
	ReserveStock(ctx context.Context, productID string, orderID string, qty int) error
}

type shopModule struct {
	app *app.Application
}

func NewShopModule(app *app.Application) ShopModule {
	return &shopModule{app: app}
}

func (m *shopModule) GetProduct(ctx context.Context, productID string) (ProductDTO, error) {
	dto, err := m.app.Queries.GetProduct.Handle(ctx, query.GetProduct{ProductID: productID})
	if err != nil {
		return ProductDTO{}, err
	}
	return ProductDTO{
		ID:         dto.ID,
		Name:       dto.Name,
		PriceCents: dto.PriceCents,
		Stock:      dto.Stock,
	}, nil
}

func (m *shopModule) ReserveStock(ctx context.Context, productID string, orderID string, qty int) error {
	return m.app.Commands.ReserveStock.Handle(ctx, command.ReserveStock{ProductID: productID, OrderID: orderID, Qty: qty})
}
