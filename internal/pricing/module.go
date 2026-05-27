package pricing

import (
	"log/slog"

	"gorm.io/gorm"

	"modular_monolith/internal/pricing/adapters/mysql"
	"modular_monolith/internal/pricing/app"
	"modular_monolith/internal/pricing/app/command"
	pricingmod "modular_monolith/internal/pricing/ports/module"
)

type Deps struct {
	DB       *gorm.DB
	Logger   *slog.Logger
	Products command.ProductCatalogService
}

type Module struct {
	App         *app.Application
	PortsModule pricingmod.PricingModule
}

func NewModule(deps Deps) (*Module, error) {
	promotions := mysql.NewPromotionRepository(deps.DB)
	application := &app.Application{
		Commands: app.Commands{
			CalculateOrderPricing: command.CalculateOrderPricingHandler{
				Products:   deps.Products,
				Promotions: promotions,
			},
		},
	}
	return &Module{App: application, PortsModule: pricingmod.NewPricingModule(application)}, nil
}

func Models() []any {
	return mysql.Models()
}
