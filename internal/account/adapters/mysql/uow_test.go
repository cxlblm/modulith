package mysql

import "modular_monolith/internal/account/app/command"

var _ command.UnitOfWork = NewUnitOfWork(nil)
