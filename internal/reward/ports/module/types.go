package module

type ClaimContestRewardRequest struct {
	ContestID string
	UserID    string
}

type ClaimContestRewardResult struct {
	Claimed        bool
	AlreadyClaimed bool
}
