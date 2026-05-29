package reward

import (
	"log/slog"

	"gorm.io/gorm"

	"modular_monolith/internal/reward/adapters/mysql"
	"modular_monolith/internal/reward/app"
	"modular_monolith/internal/reward/app/command"
	rewardmod "modular_monolith/internal/reward/ports/module"
)

type Deps struct {
	DB     *gorm.DB
	Logger *slog.Logger
}

type Module struct {
	App         *app.Application
	PortsModule rewardmod.RewardModule
}

func NewModule(deps Deps) (*Module, error) {
	claims := mysql.NewClaimRepository(deps.DB)
	application := &app.Application{
		Commands: app.Commands{
			ClaimContestReward: command.ClaimContestRewardHandler{Claims: claims},
		},
	}
	return &Module{App: application, PortsModule: rewardmod.NewRewardModule(application)}, nil
}

func Models() []any {
	return mysql.Models()
}
