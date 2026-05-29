package command

import (
	"context"
	"testing"
)

func TestClaimContestRewardIsIdempotent(t *testing.T) {
	claims := &fakeClaims{claimed: map[string]bool{}}
	handler := ClaimContestRewardHandler{Claims: claims}
	ctx := context.Background()
	cmd := ClaimContestReward{ContestID: "contest-1", UserID: "user-1"}

	first, err := handler.Handle(ctx, cmd)
	if err != nil {
		t.Fatalf("first Handle() error = %v", err)
	}
	if !first.Claimed || first.AlreadyClaimed {
		t.Fatalf("first result = %+v, want claimed and not already claimed", first)
	}

	second, err := handler.Handle(ctx, cmd)
	if err != nil {
		t.Fatalf("second Handle() error = %v", err)
	}
	if !second.Claimed || !second.AlreadyClaimed {
		t.Fatalf("second result = %+v, want claimed and already claimed", second)
	}
}

type fakeClaims struct {
	claimed map[string]bool
}

func (f *fakeClaims) Claim(ctx context.Context, contestID string, userID string) (bool, error) {
	key := contestID + ":" + userID
	already := f.claimed[key]
	f.claimed[key] = true
	return already, nil
}
