package command

import "context"

type ClaimRepository interface {
	Claim(ctx context.Context, contestID string, userID string) (alreadyClaimed bool, err error)
}
