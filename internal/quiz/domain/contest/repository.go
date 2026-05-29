package contest

import "context"

type Repository interface {
	Save(ctx context.Context, c *Contest) error
	FindByUUID(ctx context.Context, uuid ContestUUID) (*Contest, error)
}
