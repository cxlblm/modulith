package mysql

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"modular_monolith/internal/shop/app/query"
	"modular_monolith/internal/shop/domain/product"
)

type ReadModel struct {
	db *gorm.DB
}

func NewReadModel(db *gorm.DB) *ReadModel {
	return &ReadModel{db: db}
}

func (r *ReadModel) GetProduct(ctx context.Context, productID string) (query.ProductDTO, error) {
	var model ProductModel
	if err := r.db.WithContext(ctx).First(&model, "uuid = ?", productID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return query.ProductDTO{}, product.NewProductNotFound(err)
		}
		return query.ProductDTO{}, fmt.Errorf("get product: %w", err)
	}
	return productDTO(model), nil
}

func (r *ReadModel) ListProducts(ctx context.Context) ([]query.ProductDTO, error) {
	var models []ProductModel
	if err := r.db.WithContext(ctx).Order("id").Find(&models).Error; err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}
	items := make([]query.ProductDTO, 0, len(models))
	for _, model := range models {
		items = append(items, productDTO(model))
	}
	return items, nil
}

func productDTO(model ProductModel) query.ProductDTO {
	return query.ProductDTO{
		ID:         model.UUID,
		Name:       model.Name,
		PriceCents: model.PriceCents,
		Stock:      model.Stock,
		CreatedAt:  model.CreatedAt,
		UpdatedAt:  model.UpdatedAt,
	}
}
