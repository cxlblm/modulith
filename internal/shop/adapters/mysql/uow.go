package mysql

import (
	"gorm.io/gorm"

	"modular_monolith/internal/platform/dbtx"
	"modular_monolith/internal/shop/app/command"
)

func NewUnitOfWork(db *gorm.DB) command.UnitOfWork {
	return dbtx.NewUnitOfWork[command.Repositories](db, "shop.UnitOfWork.RunInTx", nil, func(tx *gorm.DB, _ *dbtx.PendingCollector) command.Repositories {
		return command.Repositories{
			Products: NewProductRepositoryWithTx(tx),
		}
	})
}
