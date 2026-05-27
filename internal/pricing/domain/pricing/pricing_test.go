package pricing

import (
	"errors"
	"testing"
)

func TestCalculateQuote_AppliesBestThresholdDiscountAndAllocatesByLine(t *testing.T) {
	quote, err := CalculateQuote("user-1", []ProductLine{
		{ProductID: "product-1", ProductName: "Keyboard", UnitPriceCents: 1000, Qty: 2},
		{ProductID: "product-2", ProductName: "Mouse", UnitPriceCents: 3000, Qty: 1},
	}, []Promotion{
		{UUID: "promo-small", Name: "Spend 1000 save 100", ThresholdCents: 1000, DiscountCents: 100, Active: true},
		{UUID: "promo-big", Name: "Spend 5000 save 999", ThresholdCents: 5000, DiscountCents: 999, Active: true},
	})
	if err != nil {
		t.Fatalf("CalculateQuote() error = %v", err)
	}

	if quote.OriginalTotalCents != 5000 {
		t.Fatalf("OriginalTotalCents = %d, want 5000", quote.OriginalTotalCents)
	}
	if quote.DiscountTotalCents != 999 {
		t.Fatalf("DiscountTotalCents = %d, want 999", quote.DiscountTotalCents)
	}
	if quote.PayableTotalCents != 4001 {
		t.Fatalf("PayableTotalCents = %d, want 4001", quote.PayableTotalCents)
	}
	if len(quote.AppliedPromotions) != 1 || quote.AppliedPromotions[0].UUID != "promo-big" {
		t.Fatalf("AppliedPromotions = %#v, want promo-big", quote.AppliedPromotions)
	}
	if got := quote.Items[0].DiscountCents; got != 400 {
		t.Fatalf("first item discount = %d, want 400", got)
	}
	if got := quote.Items[1].DiscountCents; got != 599 {
		t.Fatalf("second item discount = %d, want 599", got)
	}
}

func TestCalculateQuote_NoEligiblePromotionKeepsPayableAtOriginal(t *testing.T) {
	quote, err := CalculateQuote("user-1", []ProductLine{
		{ProductID: "product-1", ProductName: "Keyboard", UnitPriceCents: 1000, Qty: 1},
	}, []Promotion{
		{UUID: "promo-1", Name: "Spend 5000 save 500", ThresholdCents: 5000, DiscountCents: 500, Active: true},
	})
	if err != nil {
		t.Fatalf("CalculateQuote() error = %v", err)
	}

	if quote.OriginalTotalCents != 1000 || quote.DiscountTotalCents != 0 || quote.PayableTotalCents != 1000 {
		t.Fatalf("quote totals = original %d discount %d payable %d, want 1000/0/1000",
			quote.OriginalTotalCents,
			quote.DiscountTotalCents,
			quote.PayableTotalCents,
		)
	}
	if len(quote.AppliedPromotions) != 0 {
		t.Fatalf("AppliedPromotions length = %d, want 0", len(quote.AppliedPromotions))
	}
}

func TestCalculateQuote_CapsDiscountAtOriginalTotal(t *testing.T) {
	quote, err := CalculateQuote("user-1", []ProductLine{
		{ProductID: "product-1", ProductName: "Keyboard", UnitPriceCents: 1000, Qty: 1},
	}, []Promotion{
		{UUID: "promo-1", Name: "Spend 1000 save 5000", ThresholdCents: 1000, DiscountCents: 5000, Active: true},
	})
	if err != nil {
		t.Fatalf("CalculateQuote() error = %v", err)
	}

	if quote.DiscountTotalCents != 1000 {
		t.Fatalf("DiscountTotalCents = %d, want 1000", quote.DiscountTotalCents)
	}
	if quote.Items[0].PayableCents != 0 {
		t.Fatalf("item payable = %d, want 0", quote.Items[0].PayableCents)
	}
}

func TestCalculateQuote_RejectsInvalidProductLine(t *testing.T) {
	_, err := CalculateQuote("user-1", []ProductLine{
		{ProductID: "product-1", ProductName: "Keyboard", UnitPriceCents: 0, Qty: 1},
	}, nil)

	if !errors.Is(err, ErrInvalidPricing) {
		t.Fatalf("CalculateQuote() error = %v, want ErrInvalidPricing", err)
	}
}
