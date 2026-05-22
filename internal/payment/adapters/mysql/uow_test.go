package mysql

import "modular_monolith/internal/payment/app/command"

var _ command.UnitOfWork = NewUnitOfWork(nil, nil)
