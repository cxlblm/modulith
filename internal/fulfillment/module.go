package fulfillment

import (
	"log/slog"

	"github.com/labstack/echo/v5"
	"gorm.io/gorm"

	"modular_monolith/internal/fulfillment/adapters/mysql"
	"modular_monolith/internal/fulfillment/app"
	"modular_monolith/internal/fulfillment/app/command"
	"modular_monolith/internal/fulfillment/app/query"
	fulfillmenteventsub "modular_monolith/internal/fulfillment/ports/eventsub"
	fulfillmenthttp "modular_monolith/internal/fulfillment/ports/http"
	fulfillmentmod "modular_monolith/internal/fulfillment/ports/module"
	"modular_monolith/internal/platform/eventbus"
)

type Deps struct {
	DB       *gorm.DB
	Logger   *slog.Logger
	EventBus eventbus.Bus
}

type Module struct {
	App         *app.Application
	PortsModule fulfillmentmod.FulfillmentModule
	eventBus    eventbus.Bus
}

func NewModule(deps Deps) (*Module, error) {
	shipments := mysql.NewShipmentRepository(deps.DB, deps.EventBus)
	readModel := mysql.NewReadModel(deps.DB)
	application := &app.Application{
		Commands: app.Commands{
			CreateShipment: command.CreateShipmentHandler{Shipments: shipments},
			SendShipment:   command.SendShipmentHandler{Shipments: shipments},
		},
		Queries: app.Queries{
			ListShipments: query.ListShipmentsHandler{ReadModel: readModel},
		},
	}
	module := &Module{App: application, PortsModule: fulfillmentmod.NewFulfillmentModule(), eventBus: deps.EventBus}
	return module, nil
}

func (m *Module) RegisterHTTP(e *echo.Echo) {
	fulfillmenthttp.Register(e, m.App)
}

func (m *Module) RegisterEventSubs() {
	fulfillmenteventsub.RegisterAll(m.eventBus, m.App)
}

func Models() []any {
	return mysql.Models()
}
