package quiz

import (
	"log/slog"

	"github.com/labstack/echo/v5"
	"gorm.io/gorm"

	"modular_monolith/internal/quiz/adapters/mysql"
	"modular_monolith/internal/quiz/app"
	"modular_monolith/internal/quiz/app/command"
	"modular_monolith/internal/quiz/app/query"
	quizhttp "modular_monolith/internal/quiz/ports/http"
	quizmod "modular_monolith/internal/quiz/ports/module"
)

type Deps struct {
	DB      *gorm.DB
	Logger  *slog.Logger
	Rewards command.RewardService
}

type Module struct {
	App         *app.Application
	PortsModule quizmod.QuizModule
}

func NewModule(deps Deps) (*Module, error) {
	questions := mysql.NewQuestionRepository(deps.DB)
	contests := mysql.NewContestRepository(deps.DB)
	participations := mysql.NewParticipationRepository(deps.DB)
	revivalCards := mysql.NewRevivalCardRepository(deps.DB)
	readModel := mysql.NewReadModel(deps.DB)
	application := &app.Application{
		Commands: app.Commands{
			CreateQuestion:    command.CreateQuestionHandler{Questions: questions},
			CreateContest:     command.CreateContestHandler{Contests: contests, Questions: questions},
			PublishContest:    command.PublishContestHandler{Contests: contests},
			SubmitAnswer:      command.SubmitAnswerHandler{Contests: contests, Participations: participations, RevivalCards: revivalCards},
			ClaimReward:       command.ClaimRewardHandler{Participations: participations, Rewards: deps.Rewards},
			GrantRevivalCards: command.GrantRevivalCardsHandler{RevivalCards: revivalCards},
		},
		Queries: app.Queries{
			ListQuestions: query.ListQuestionsHandler{ReadModel: readModel},
			GetArena:      query.GetArenaHandler{ReadModel: readModel},
		},
	}
	return &Module{App: application, PortsModule: quizmod.NewQuizModule()}, nil
}

func (m *Module) RegisterHTTP(e *echo.Echo) {
	quizhttp.Register(e, m.App)
}

func Models() []any {
	return mysql.Models()
}
