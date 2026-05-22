package command

import (
	"context"

	"modular_monolith/internal/account/domain/user"
)

type UnitOfWork interface {
	RunInTx(ctx context.Context, fn func(ctx context.Context, repos Repositories) error) error
}

type Repositories struct {
	Users user.Repository
}
