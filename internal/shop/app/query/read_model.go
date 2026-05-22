package query

import "context"

type ReadModel interface {
	GetProduct(ctx context.Context, productID string) (ProductDTO, error)
	ListProducts(ctx context.Context) ([]ProductDTO, error)
}
