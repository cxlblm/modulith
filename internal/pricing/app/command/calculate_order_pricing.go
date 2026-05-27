package command

import (
	"context"
	"fmt"

	pricingdomain "modular_monolith/internal/pricing/domain/pricing"
)

type CalculateOrderPricing struct {
	UserID string
	Items  []PricingItem
}

type PricingItem struct {
	ProductID string
	Qty       int
}

type AppliedPromotionResult struct {
	UUID          string
	Name          string
	DiscountCents int64
}

type PricingItemResult struct {
	ProductID              string
	ProductName            string
	Qty                    int
	OriginalUnitPriceCents int64
	OriginalSubtotalCents  int64
	DiscountCents          int64
	PayableCents           int64
	AppliedPromotions      []AppliedPromotionResult
}

type PricingResult struct {
	OriginalTotalCents int64
	DiscountTotalCents int64
	PayableTotalCents  int64
	Items              []PricingItemResult
	AppliedPromotions  []AppliedPromotionResult
}

type CalculateOrderPricingHandler struct {
	Products   ProductCatalogService
	Promotions PromotionRepository
}

func (h CalculateOrderPricingHandler) Handle(ctx context.Context, cmd CalculateOrderPricing) (PricingResult, error) {
	lines := make([]pricingdomain.ProductLine, 0, len(cmd.Items))
	for _, item := range cmd.Items {
		product, err := h.Products.GetProduct(ctx, item.ProductID)
		if err != nil {
			return PricingResult{}, fmt.Errorf("get product for pricing: %w", err)
		}
		lines = append(lines, pricingdomain.ProductLine{
			ProductID:      product.ID,
			ProductName:    product.Name,
			UnitPriceCents: product.PriceCents,
			Qty:            item.Qty,
		})
	}

	promotions, err := h.Promotions.ListActivePromotions(ctx)
	if err != nil {
		return PricingResult{}, fmt.Errorf("list active promotions: %w", err)
	}
	quote, err := pricingdomain.CalculateQuote(cmd.UserID, lines, promotionDomainItems(promotions))
	if err != nil {
		return PricingResult{}, err
	}
	return pricingResult(quote), nil
}

func promotionDomainItems(promotions []PromotionInfo) []pricingdomain.Promotion {
	out := make([]pricingdomain.Promotion, 0, len(promotions))
	for _, promo := range promotions {
		out = append(out, pricingdomain.Promotion{
			UUID:           promo.UUID,
			Name:           promo.Name,
			ThresholdCents: promo.ThresholdCents,
			DiscountCents:  promo.DiscountCents,
			Active:         promo.Active,
		})
	}
	return out
}

func pricingResult(quote pricingdomain.Quote) PricingResult {
	items := make([]PricingItemResult, 0, len(quote.Items))
	for _, item := range quote.Items {
		items = append(items, PricingItemResult{
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
	return PricingResult{
		OriginalTotalCents: quote.OriginalTotalCents,
		DiscountTotalCents: quote.DiscountTotalCents,
		PayableTotalCents:  quote.PayableTotalCents,
		Items:              items,
		AppliedPromotions:  promotionResults(quote.AppliedPromotions),
	}
}

func promotionResults(promotions []pricingdomain.AppliedPromotion) []AppliedPromotionResult {
	out := make([]AppliedPromotionResult, 0, len(promotions))
	for _, promo := range promotions {
		out = append(out, AppliedPromotionResult{
			UUID:          promo.UUID,
			Name:          promo.Name,
			DiscountCents: promo.DiscountCents,
		})
	}
	return out
}
