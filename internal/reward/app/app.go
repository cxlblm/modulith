package app

import "modular_monolith/internal/reward/app/command"

type Application struct {
	Commands Commands
}

type Commands struct {
	ClaimContestReward command.ClaimContestRewardHandler
}
