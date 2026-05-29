package command

import "context"

type RevivalCardRepository interface {
	Grant(ctx context.Context, userID string, count int) error
	TryConsumeOne(ctx context.Context, userID string) (bool, error)
}
