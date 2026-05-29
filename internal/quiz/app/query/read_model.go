package query

import "context"

type ReadModel interface {
	ListQuestions(ctx context.Context) ([]QuestionDTO, error)
	GetArena(ctx context.Context, contestID string) (ArenaDTO, error)
}
