package mysql

import (
	"gorm.io/gorm"

	"modular_monolith/internal/order/app/command"
	"modular_monolith/internal/platform/dbtx"
	"modular_monolith/internal/platform/eventbus"
)

func NewUnitOfWork(db *gorm.DB, bus eventbus.Bus) command.UnitOfWork {
	return dbtx.NewUnitOfWork[command.Repositories](db, "order.UnitOfWork.RunInTx", bus, func(tx *gorm.DB, pending *dbtx.PendingCollector) command.Repositories {
		return command.Repositories{
			Orders: NewOrderRepositoryWithTx(tx, pending, bus),
		}
	})
}
