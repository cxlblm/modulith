package shopclient

import (
	"context"

	shopmod "modular_monolith/internal/shop/ports/module"
)

type ProductsService struct {
	shop shopmod.ShopModule
}

func NewProductsService(shop shopmod.ShopModule) *ProductsService {
	return &ProductsService{shop: shop}
}

func (s *ProductsService) ReserveStock(ctx context.Context, productID string, orderID string, qty int) error {
	return s.shop.ReserveStock(ctx, productID, orderID, qty)
}
