package command

import (
	"context"

	"modular_monolith/internal/fulfillment/domain/shipment"
)

type UnitOfWork interface {
	RunInTx(ctx context.Context, fn func(ctx context.Context, repos Repositories) error) error
}

type Repositories struct {
	Shipments shipment.Repository
}
