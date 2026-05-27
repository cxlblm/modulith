package pricingclient

import (
	"context"

	ordercmd "modular_monolith/internal/order/app/command"
	pricingmod "modular_monolith/internal/pricing/ports/module"
)

type PricingService struct {
	pricing pricingmod.PricingModule
}

func NewPricingService(pricing pricingmod.PricingModule) *PricingService {
	return &PricingService{pricing: pricing}
}

func (s *PricingService) CalculateOrderPricing(ctx context.Context, req ordercmd.PricingRequest) (ordercmd.PricingResult, error) {
	items := make([]pricingmod.PricingItemRequest, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, pricingmod.PricingItemRequest{ProductID: item.ProductID, Qty: item.Qty})
	}
	dto, err := s.pricing.CalculateOrderPricing(ctx, pricingmod.CalculateOrderPricingRequest{
		UserID: req.UserID,
		Items:  items,
	})
	if err != nil {
		return ordercmd.PricingResult{}, err
	}
	return pricingResult(dto), nil
}

func pricingResult(dto pricingmod.PricingDTO) ordercmd.PricingResult {
	items := make([]ordercmd.PricingItemResult, 0, len(dto.Items))
	for _, item := range dto.Items {
		items = append(items, ordercmd.PricingItemResult{
			ProductID:              item.ProductID,
			ProductName:            item.ProductName,
			Qty:                    item.Qty,
			OriginalUnitPriceCents: item.OriginalUnitPriceCents,
			OriginalSubtotalCents:  item.OriginalSubtotalCents,
			DiscountCents:          item.DiscountCents,
			PayableCents:           item.PayableCents,
			AppliedPromotions:      promotionResults(item.AppliedPromotions),
		})
	}
	return ordercmd.PricingResult{
		OriginalTotalCents: dto.OriginalTotalCents,
		DiscountTotalCents: dto.DiscountTotalCents,
		PayableTotalCents:  dto.PayableTotalCents,
		Items:              items,
		AppliedPromotions:  promotionResults(dto.AppliedPromotions),
	}
}

func promotionResults(promotions []pricingmod.AppliedPromotionDTO) []ordercmd.AppliedPromotionResult {
	out := make([]ordercmd.AppliedPromotionResult, 0, len(promotions))
	for _, promo := range promotions {
		out = append(out, ordercmd.AppliedPromotionResult{
			UUID:          promo.UUID,
			Name:          promo.Name,
			DiscountCents: promo.DiscountCents,
		})
	}
	return out
}
