package payment

import (
	"log/slog"

	"github.com/labstack/echo/v5"
	"gorm.io/gorm"

	"modular_monolith/internal/payment/adapters/mysql"
	"modular_monolith/internal/payment/app"
	"modular_monolith/internal/payment/app/command"
	"modular_monolith/internal/payment/app/query"
	paymenteventsub "modular_monolith/internal/payment/ports/eventsub"
	paymenthttp "modular_monolith/internal/payment/ports/http"
	paymentmod "modular_monolith/internal/payment/ports/module"
	"modular_monolith/internal/platform/eventbus"
)

type Deps struct {
	DB       *gorm.DB
	Logger   *slog.Logger
	EventBus eventbus.Bus
}

type Module struct {
	App         *app.Application
	PortsModule paymentmod.PaymentModule
	eventBus    eventbus.Bus
}

func NewModule(deps Deps) (*Module, error) {
	payments := mysql.NewPaymentRepository(deps.DB, deps.EventBus)
	readModel := mysql.NewReadModel(deps.DB)
	application := &app.Application{
		Commands: app.Commands{
			InitializePayment: command.InitializePaymentHandler{Payments: payments},
			ConfirmPayment:    command.ConfirmPaymentHandler{Payments: payments},
		},
		Queries: app.Queries{
			ListPayments: query.ListPaymentsHandler{ReadModel: readModel},
		},
	}
	module := &Module{App: application, PortsModule: paymentmod.NewPaymentModule(), eventBus: deps.EventBus}
	return module, nil
}

func (m *Module) RegisterHTTP(e *echo.Echo) {
	paymenthttp.Register(e, m.App)
}

func (m *Module) RegisterEventSubs() {
	paymenteventsub.RegisterAll(m.eventBus, m.App)
}

func Models() []any {
	return mysql.Models()
}
