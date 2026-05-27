package account

import (
	"log/slog"

	"github.com/labstack/echo/v5"
	"gorm.io/gorm"

	"modular_monolith/internal/account/adapters/mysql"
	"modular_monolith/internal/account/app"
	"modular_monolith/internal/account/app/command"
	"modular_monolith/internal/account/app/query"
	accounthttp "modular_monolith/internal/account/ports/http"
	accountmod "modular_monolith/internal/account/ports/module"
)

type Deps struct {
	DB     *gorm.DB
	Logger *slog.Logger
}

type Module struct {
	App         *app.Application
	PortsModule accountmod.AccountModule
}

func NewModule(deps Deps) (*Module, error) {
	users := mysql.NewUserRepository(deps.DB)
	readModel := mysql.NewReadModel(deps.DB)
	application := &app.Application{
		Commands: app.Commands{
			CreateUser:    command.CreateUserHandler{Users: users},
			AddAddress:    command.AddAddressHandler{Users: users},
			UpdateAddress: command.UpdateAddressHandler{Users: users},
			DeleteAddress: command.DeleteAddressHandler{Users: users},
		},
		Queries: app.Queries{
			GetUser:          query.GetUserHandler{ReadModel: readModel},
			EnsureUserActive: query.EnsureUserActiveHandler{ReadModel: readModel},
			ListAddresses:    query.ListAddressesHandler{ReadModel: readModel},
			GetAddress:       query.GetAddressHandler{ReadModel: readModel},
		},
	}
	return &Module{App: application, PortsModule: accountmod.NewAccountModule(application)}, nil
}

func (m *Module) RegisterHTTP(e *echo.Echo) {
	accounthttp.Register(e, m.App)
}

func Models() []any {
	return mysql.Models()
}
