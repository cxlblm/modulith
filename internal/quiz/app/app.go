package app

import (
	"modular_monolith/internal/quiz/app/command"
	"modular_monolith/internal/quiz/app/query"
)

type Application struct {
	Commands Commands
	Queries  Queries
}

type Commands struct {
	CreateQuestion    command.CreateQuestionHandler
	CreateContest     command.CreateContestHandler
	PublishContest    command.PublishContestHandler
	SubmitAnswer      command.SubmitAnswerHandler
	ClaimReward       command.ClaimRewardHandler
	GrantRevivalCards command.GrantRevivalCardsHandler
}

type Queries struct {
	ListQuestions query.ListQuestionsHandler
	GetArena      query.GetArenaHandler
}
