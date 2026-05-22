package shopclient

import (
	"context"

	ordercmd "modular_monolith/internal/order/app/command"
	shopmod "modular_monolith/internal/shop/ports/module"
)

type ProductsService struct {
	shop shopmod.ShopModule
}

func NewProductsService(shop shopmod.ShopModule) *ProductsService {
	return &ProductsService{shop: shop}
}

func (s *ProductsService) GetProduct(ctx context.Context, productID string) (ordercmd.ProductInfo, error) {
	dto, err := s.shop.GetProduct(ctx, productID)
	if err != nil {
		return ordercmd.ProductInfo{}, err
	}
	return ordercmd.ProductInfo{ID: dto.ID, Name: dto.Name, PriceCents: dto.PriceCents}, nil
}

func (s *ProductsService) ReserveStock(ctx context.Context, productID string, orderID string, qty int) error {
	return s.shop.ReserveStock(ctx, productID, orderID, qty)
}
