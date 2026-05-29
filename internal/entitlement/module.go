package entitlement

import (
	"log/slog"

	"github.com/labstack/echo/v5"
	"gorm.io/gorm"

	"modular_monolith/internal/entitlement/adapters/mysql"
	"modular_monolith/internal/entitlement/app"
	"modular_monolith/internal/entitlement/app/command"
	entitlementhttp "modular_monolith/internal/entitlement/ports/http"
	entitlementmod "modular_monolith/internal/entitlement/ports/module"
)

type Deps struct {
	DB     *gorm.DB
	Logger *slog.Logger
}

type Module struct {
	App         *app.Application
	PortsModule entitlementmod.EntitlementModule
}

func NewModule(deps Deps) (*Module, error) {
	revivalCards := mysql.NewRevivalCardRepository(deps.DB)
	application := &app.Application{
		Commands: app.Commands{
			GrantRevivalCards:     command.GrantRevivalCardsHandler{RevivalCards: revivalCards},
			TryConsumeRevivalCard: command.TryConsumeRevivalCardHandler{RevivalCards: revivalCards},
		},
	}
	return &Module{App: application, PortsModule: entitlementmod.NewEntitlementModule(application)}, nil
}

func (m *Module) RegisterHTTP(e *echo.Echo) {
	entitlementhttp.Register(e, m.App)
}

func Models() []any {
	return mysql.Models()
}
