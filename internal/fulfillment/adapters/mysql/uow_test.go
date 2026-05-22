package mysql

import "modular_monolith/internal/fulfillment/app/command"

var _ command.UnitOfWork = NewUnitOfWork(nil, nil)
