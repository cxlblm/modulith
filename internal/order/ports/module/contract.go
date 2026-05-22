package module

import (
	"context"

	"modular_monolith/internal/order/app"
	"modular_monolith/internal/order/app/query"
)

type OrderModule interface {
	GetOrder(ctx context.Context, orderID string) (OrderDTO, error)
}

type orderModule struct {
	app *app.Application
}

func NewOrderModule(app *app.Application) OrderModule {
	return &orderModule{app: app}
}

func (m *orderModule) GetOrder(ctx context.Context, orderID string) (OrderDTO, error) {
	dto, err := m.app.Queries.GetOrder.Handle(ctx, query.GetOrder{OrderID: orderID})
	if err != nil {
		return OrderDTO{}, err
	}
	return OrderDTO{ID: dto.ID, UserID: dto.UserID, Status: dto.Status, TotalCents: dto.TotalCents}, nil
}
