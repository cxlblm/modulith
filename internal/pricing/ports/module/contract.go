package module

import (
	"context"

	"modular_monolith/internal/pricing/app"
	"modular_monolith/internal/pricing/app/command"
)

type PricingModule interface {
	CalculateOrderPricing(ctx context.Context, req CalculateOrderPricingRequest) (PricingDTO, error)
}

type pricingModule struct {
	app *app.Application
}

func NewPricingModule(app *app.Application) PricingModule {
	return &pricingModule{app: app}
}

func (m *pricingModule) CalculateOrderPricing(ctx context.Context, req CalculateOrderPricingRequest) (PricingDTO, error) {
	items := make([]command.PricingItem, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, command.PricingItem{ProductID: item.ProductID, Qty: item.Qty})
	}
	result, err := m.app.Commands.CalculateOrderPricing.Handle(ctx, command.CalculateOrderPricing{
		UserID: req.UserID,
		Items:  items,
	})
	if err != nil {
		return PricingDTO{}, err
	}
	return pricingDTO(result), nil
}

func pricingDTO(result command.PricingResult) PricingDTO {
	items := make([]PricingItemDTO, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, PricingItemDTO{
			ProductID:              item.ProductID,
			ProductName:            item.ProductName,
			Qty:                    item.Qty,
			OriginalUnitPriceCents: item.OriginalUnitPriceCents,
			OriginalSubtotalCents:  item.OriginalSubtotalCents,
			DiscountCents:          item.DiscountCents,
			PayableCents:           item.PayableCents,
			AppliedPromotions:      promotionDTOs(item.AppliedPromotions),
		})
	}
	return PricingDTO{
		OriginalTotalCents: result.OriginalTotalCents,
		DiscountTotalCents: result.DiscountTotalCents,
		PayableTotalCents:  result.PayableTotalCents,
		Items:              items,
		AppliedPromotions:  promotionDTOs(result.AppliedPromotions),
	}
}

func promotionDTOs(promotions []command.AppliedPromotionResult) []AppliedPromotionDTO {
	out := make([]AppliedPromotionDTO, 0, len(promotions))
	for _, promo := range promotions {
		out = append(out, AppliedPromotionDTO{
			UUID:          promo.UUID,
			Name:          promo.Name,
			DiscountCents: promo.DiscountCents,
		})
	}
	return out
}
