package command

type invalidError struct {
	message string
	reason  string
}

func (e invalidError) Error() string { return e.message }
func (e invalidError) Invalid()      {}
func (e invalidError) Reason() string {
	return e.reason
}

var ErrInvalidRewardClaim error = invalidError{message: "invalid reward claim", reason: "invalid_reward_claim"}
