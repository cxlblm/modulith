package command

import (
	"context"

	paymentdomain "modular_monolith/internal/payment/domain/payment"
)

type UnitOfWork interface {
	RunInTx(ctx context.Context, fn func(ctx context.Context, repos Repositories) error) error
}

type Repositories struct {
	Payments paymentdomain.Repository
}
