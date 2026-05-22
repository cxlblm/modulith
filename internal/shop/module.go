package shop

import (
	"log/slog"

	"github.com/labstack/echo/v5"
	"gorm.io/gorm"

	"modular_monolith/internal/shop/adapters/mysql"
	"modular_monolith/internal/shop/app"
	"modular_monolith/internal/shop/app/command"
	"modular_monolith/internal/shop/app/query"
	shophttp "modular_monolith/internal/shop/ports/http"
	shopmod "modular_monolith/internal/shop/ports/module"
)

type Deps struct {
	DB     *gorm.DB
	Logger *slog.Logger
}

type Module struct {
	App         *app.Application
	PortsModule shopmod.ShopModule
}

func NewModule(deps Deps) (*Module, error) {
	products := mysql.NewProductRepository(deps.DB)
	readModel := mysql.NewReadModel(deps.DB)
	application := &app.Application{
		Commands: app.Commands{
			CreateProduct: command.CreateProductHandler{Products: products},
			ReserveStock:  command.ReserveStockHandler{Products: products},
		},
		Queries: app.Queries{
			GetProduct:  query.GetProductHandler{ReadModel: readModel},
			ListProduct: query.ListProductsHandler{ReadModel: readModel},
		},
	}
	return &Module{App: application, PortsModule: shopmod.NewShopModule(application)}, nil
}

func (m *Module) RegisterHTTP(e *echo.Echo) {
	shophttp.Register(e, m.App)
}

func Models() []any {
	return mysql.Models()
}
