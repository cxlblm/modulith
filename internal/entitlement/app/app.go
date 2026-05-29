package app

import "modular_monolith/internal/entitlement/app/command"

type Application struct {
	Commands Commands
}

type Commands struct {
	GrantRevivalCards     command.GrantRevivalCardsHandler
	TryConsumeRevivalCard command.TryConsumeRevivalCardHandler
}
