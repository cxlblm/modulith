package module

type CalculateOrderPricingRequest struct {
	UserID string
	Items  []PricingItemRequest
}

type PricingItemRequest struct {
	ProductID string
	Qty       int
}

type AppliedPromotionDTO struct {
	UUID          string
	Name          string
	DiscountCents int64
}

type PricingItemDTO struct {
	ProductID              string
	ProductName            string
	Qty                    int
	OriginalUnitPriceCents int64
	OriginalSubtotalCents  int64
	DiscountCents          int64
	PayableCents           int64
	AppliedPromotions      []AppliedPromotionDTO
}

type PricingDTO struct {
	OriginalTotalCents int64
	DiscountTotalCents int64
	PayableTotalCents  int64
	Items              []PricingItemDTO
	AppliedPromotions  []AppliedPromotionDTO
}
