package question

import "context"

type Repository interface {
	Save(ctx context.Context, q *Question) error
	FindByUUID(ctx context.Context, uuid QuestionUUID) (*Question, error)
	List(ctx context.Context) ([]*Question, error)
}
