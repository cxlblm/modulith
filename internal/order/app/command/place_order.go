package command

import (
	"context"
	"fmt"

	orderdomain "modular_monolith/internal/order/domain/order"
)

type PlaceOrder struct {
	UserID    string
	AddressID string
	Items     []PlaceOrderItem
}

type PlaceOrderItem struct {
	ProductID string
	Qty       int
}

type PlaceOrderResult struct {
	OrderID string
}

type PlaceOrderHandler struct {
	Orders    orderdomain.Repository
	Products  ProductsService
	Addresses AddressService
	Users     UserEligibilityService
}

func (h PlaceOrderHandler) Handle(ctx context.Context, cmd PlaceOrder) (PlaceOrderResult, error) {
	if err := h.Users.EnsureCanPlaceOrder(ctx, cmd.UserID); err != nil {
		return PlaceOrderResult{}, err
	}
	address, err := h.Addresses.GetAddress(ctx, cmd.UserID, cmd.AddressID)
	if err != nil {
		return PlaceOrderResult{}, err
	}
	items := make([]orderdomain.Item, 0, len(cmd.Items))
	for _, item := range cmd.Items {
		product, err := h.Products.GetProduct(ctx, item.ProductID)
		if err != nil {
			return PlaceOrderResult{}, err
		}
		items = append(items, orderdomain.Item{
			ProductUUID:    product.ID,
			ProductName:    product.Name,
			UnitPriceCents: product.PriceCents,
			Qty:            item.Qty,
		})
	}
	o, err := orderdomain.NewOrder(cmd.UserID, cmd.AddressID, orderdomain.AddressSnapshot{
		Receiver: address.Receiver,
		Phone:    address.Phone,
		City:     address.City,
		Detail:   address.Detail,
	}, items)
	if err != nil {
		return PlaceOrderResult{}, err
	}
	for _, item := range cmd.Items {
		if err := h.Products.ReserveStock(ctx, item.ProductID, o.UUID().String(), item.Qty); err != nil {
			return PlaceOrderResult{}, err
		}
	}
	if err := h.Orders.Save(ctx, o); err != nil {
		return PlaceOrderResult{}, fmt.Errorf("save order: %w", err)
	}
	return PlaceOrderResult{OrderID: o.UUID().String()}, nil
}
