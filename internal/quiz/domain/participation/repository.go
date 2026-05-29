package participation

import "context"

type Repository interface {
	FindByContestAndUser(ctx context.Context, contestID string, userID string) (*Participation, error)
	Save(ctx context.Context, p *Participation) error
}
