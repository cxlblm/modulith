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
	Pricing   PricingService
}

func (h PlaceOrderHandler) Handle(ctx context.Context, cmd PlaceOrder) (PlaceOrderResult, error) {
	if err := h.Users.EnsureCanPlaceOrder(ctx, cmd.UserID); err != nil {
		return PlaceOrderResult{}, err
	}
	address, err := h.Addresses.GetAddress(ctx, cmd.UserID, cmd.AddressID)
	if err != nil {
		return PlaceOrderResult{}, err
	}
	pricingItems := make([]PricingRequestItem, 0, len(cmd.Items))
	for _, item := range cmd.Items {
		pricingItems = append(pricingItems, PricingRequestItem{ProductID: item.ProductID, Qty: item.Qty})
	}
	pricing, err := h.Pricing.CalculateOrderPricing(ctx, PricingRequest{
		UserID: cmd.UserID,
		Items:  pricingItems,
	})
	if err != nil {
		return PlaceOrderResult{}, err
	}
	o, err := orderdomain.NewOrder(cmd.UserID, cmd.AddressID, orderdomain.AddressSnapshot{
		Receiver: address.Receiver,
		Phone:    address.Phone,
		City:     address.City,
		Detail:   address.Detail,
	}, orderItems(pricing.Items))
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

func orderItems(items []PricingItemResult) []orderdomain.Item {
	out := make([]orderdomain.Item, 0, len(items))
	for _, item := range items {
		out = append(out, orderdomain.Item{
			ProductUUID:            item.ProductID,
			ProductName:            item.ProductName,
			OriginalUnitPriceCents: item.OriginalUnitPriceCents,
			OriginalSubtotalCents:  item.OriginalSubtotalCents,
			DiscountCents:          item.DiscountCents,
			PayableCents:           item.PayableCents,
			Qty:                    item.Qty,
			AppliedPromotions:      orderPromotions(item.AppliedPromotions),
		})
	}
	return out
}

func orderPromotions(promotions []AppliedPromotionResult) []orderdomain.AppliedPromotion {
	out := make([]orderdomain.AppliedPromotion, 0, len(promotions))
	for _, promo := range promotions {
		out = append(out, orderdomain.AppliedPromotion{
			UUID:          promo.UUID,
			Name:          promo.Name,
			DiscountCents: promo.DiscountCents,
		})
	}
	return out
}
