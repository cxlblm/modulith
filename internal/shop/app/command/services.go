package command

import (
	"context"

	"modular_monolith/internal/shop/domain/product"
)

type UnitOfWork interface {
	RunInTx(ctx context.Context, fn func(ctx context.Context, repos Repositories) error) error
}

type Repositories struct {
	Products product.Repository
}
