package command

import "context"

type ProductCatalogService interface {
	GetProduct(ctx context.Context, productID string) (ProductInfo, error)
}

type ProductInfo struct {
	ID         string
	Name       string
	PriceCents int64
}

type PromotionRepository interface {
	ListActivePromotions(ctx context.Context) ([]PromotionInfo, error)
}

type PromotionInfo struct {
	UUID           string
	Name           string
	ThresholdCents int64
	DiscountCents  int64
	Active         bool
}
