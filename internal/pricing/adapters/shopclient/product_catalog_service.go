package shopclient

import (
	"context"

	pricingcmd "modular_monolith/internal/pricing/app/command"
	shopmod "modular_monolith/internal/shop/ports/module"
)

type ProductCatalogService struct {
	shop shopmod.ShopModule
}

func NewProductCatalogService(shop shopmod.ShopModule) *ProductCatalogService {
	return &ProductCatalogService{shop: shop}
}

func (s *ProductCatalogService) GetProduct(ctx context.Context, productID string) (pricingcmd.ProductInfo, error) {
	dto, err := s.shop.GetProduct(ctx, productID)
	if err != nil {
		return pricingcmd.ProductInfo{}, err
	}
	return pricingcmd.ProductInfo{
		ID:         dto.ID,
		Name:       dto.Name,
		PriceCents: dto.PriceCents,
	}, nil
}
