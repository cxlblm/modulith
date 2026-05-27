package command

import (
	"context"
	"testing"
)

type fakeProductCatalog struct {
	products map[string]ProductInfo
}

func (c fakeProductCatalog) GetProduct(_ context.Context, productID string) (ProductInfo, error) {
	return c.products[productID], nil
}

type fakePromotionRepository struct {
	promotions []PromotionInfo
}

func (r fakePromotionRepository) ListActivePromotions(context.Context) ([]PromotionInfo, error) {
	return r.promotions, nil
}

func TestCalculateOrderPricingHandler_ReturnsLineAllocationsAndPromotions(t *testing.T) {
	h := CalculateOrderPricingHandler{
		Products: fakeProductCatalog{products: map[string]ProductInfo{
			"product-1": {ID: "product-1", Name: "Keyboard", PriceCents: 1000},
			"product-2": {ID: "product-2", Name: "Mouse", PriceCents: 3000},
		}},
		Promotions: fakePromotionRepository{promotions: []PromotionInfo{
			{UUID: "promo-1", Name: "Spend 5000 save 999", ThresholdCents: 5000, DiscountCents: 999, Active: true},
		}},
	}

	result, err := h.Handle(context.Background(), CalculateOrderPricing{
		UserID: "user-1",
		Items:  []PricingItem{{ProductID: "product-1", Qty: 2}, {ProductID: "product-2", Qty: 1}},
	})
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	if result.OriginalTotalCents != 5000 || result.DiscountTotalCents != 999 || result.PayableTotalCents != 4001 {
		t.Fatalf("totals = %d/%d/%d, want 5000/999/4001",
			result.OriginalTotalCents,
			result.DiscountTotalCents,
			result.PayableTotalCents,
		)
	}
	if len(result.Items) != 2 {
		t.Fatalf("items length = %d, want 2", len(result.Items))
	}
	if result.Items[0].DiscountCents != 400 || result.Items[1].DiscountCents != 599 {
		t.Fatalf("item discounts = %d/%d, want 400/599", result.Items[0].DiscountCents, result.Items[1].DiscountCents)
	}
	if len(result.AppliedPromotions) != 1 || result.AppliedPromotions[0].UUID != "promo-1" {
		t.Fatalf("AppliedPromotions = %#v, want promo-1", result.AppliedPromotions)
	}
}
