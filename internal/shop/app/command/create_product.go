package command

import (
	"context"
	"fmt"

	"modular_monolith/internal/shop/domain/product"
)

type CreateProduct struct {
	Name       string
	PriceCents int64
	Stock      int
}

type CreateProductResult struct {
	ProductID string
}

type CreateProductHandler struct {
	Products product.Repository
}

func (h CreateProductHandler) Handle(ctx context.Context, cmd CreateProduct) (CreateProductResult, error) {
	p, err := product.NewProduct(cmd.Name, cmd.PriceCents, cmd.Stock)
	if err != nil {
		return CreateProductResult{}, err
	}
	if err := h.Products.Save(ctx, p); err != nil {
		return CreateProductResult{}, fmt.Errorf("save product: %w", err)
	}
	return CreateProductResult{ProductID: p.UUID().String()}, nil
}
