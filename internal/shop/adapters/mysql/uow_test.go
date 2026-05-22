package mysql

import "modular_monolith/internal/shop/app/command"

var _ command.UnitOfWork = NewUnitOfWork(nil)
