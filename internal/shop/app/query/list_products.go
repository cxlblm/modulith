package query

import "context"

type ListProducts struct{}

type ListProductsHandler struct {
	ReadModel ReadModel
}

func (h ListProductsHandler) Handle(ctx context.Context, q ListProducts) ([]ProductDTO, error) {
	return h.ReadModel.ListProducts(ctx)
}
