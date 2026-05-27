package order

import (
	"log/slog"

	"github.com/labstack/echo/v5"
	"gorm.io/gorm"

	"modular_monolith/internal/order/adapters/mysql"
	"modular_monolith/internal/order/app"
	"modular_monolith/internal/order/app/command"
	"modular_monolith/internal/order/app/query"
	ordereventsub "modular_monolith/internal/order/ports/eventsub"
	orderhttp "modular_monolith/internal/order/ports/http"
	ordermod "modular_monolith/internal/order/ports/module"
	"modular_monolith/internal/platform/eventbus"
)

type Deps struct {
	DB        *gorm.DB
	Logger    *slog.Logger
	EventBus  eventbus.Bus
	Products  command.ProductsService
	Addresses command.AddressService
	Users     command.UserEligibilityService
}

type Module struct {
	App         *app.Application
	PortsModule ordermod.OrderModule
	eventBus    eventbus.Bus
}

func NewModule(deps Deps) (*Module, error) {
	orders := mysql.NewOrderRepository(deps.DB, deps.EventBus)
	readModel := mysql.NewReadModel(deps.DB)
	application := &app.Application{
		Commands: app.Commands{
			PlaceOrder:  command.PlaceOrderHandler{Orders: orders, Products: deps.Products, Addresses: deps.Addresses, Users: deps.Users},
			MarkPaid:    command.MarkPaidHandler{Orders: orders},
			MarkShipped: command.MarkShippedHandler{Orders: orders},
		},
		Queries: app.Queries{
			GetOrder:   query.GetOrderHandler{ReadModel: readModel},
			ListOrders: query.ListOrdersHandler{ReadModel: readModel},
		},
	}
	module := &Module{App: application, PortsModule: ordermod.NewOrderModule(application), eventBus: deps.EventBus}
	return module, nil
}

func (m *Module) RegisterHTTP(e *echo.Echo) {
	orderhttp.Register(e, m.App)
}

func (m *Module) RegisterEventSubs() {
	ordereventsub.RegisterAll(m.eventBus, m.App)
}

func Models() []any {
	return mysql.Models()
}
