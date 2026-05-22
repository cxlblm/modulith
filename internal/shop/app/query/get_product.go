package query

import "context"

type GetProduct struct {
	ProductID string
}

type GetProductHandler struct {
	ReadModel ReadModel
}

func (h GetProductHandler) Handle(ctx context.Context, q GetProduct) (ProductDTO, error) {
	return h.ReadModel.GetProduct(ctx, q.ProductID)
}
