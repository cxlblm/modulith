package mysql

import (
	"gorm.io/gorm"

	"modular_monolith/internal/account/app/command"
	"modular_monolith/internal/platform/dbtx"
)

func NewUnitOfWork(db *gorm.DB) command.UnitOfWork {
	return dbtx.NewUnitOfWork[command.Repositories](db, "account.UnitOfWork.RunInTx", nil, func(tx *gorm.DB, _ *dbtx.PendingCollector) command.Repositories {
		return command.Repositories{
			Users: NewUserRepositoryWithTx(tx),
		}
	})
}
