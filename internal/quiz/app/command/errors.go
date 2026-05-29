package command

import "errors"

var ErrParticipationNotFound = errors.New("participation not found")

type invalidError struct {
	message string
	reason  string
}

func (e invalidError) Error() string { return e.message }
func (e invalidError) Invalid()      {}
func (e invalidError) Reason() string {
	return e.reason
}

type forbiddenError struct {
	message string
	reason  string
}

func (e forbiddenError) Error() string { return e.message }
func (e forbiddenError) Forbidden()    {}
func (e forbiddenError) Reason() string {
	return e.reason
}

var (
	ErrContestNotOpen   = forbiddenError{message: "contest is not open", reason: "contest_not_open"}
	ErrRewardNotAllowed = forbiddenError{message: "reward claim is not allowed", reason: "reward_not_allowed"}
	ErrInvalidCommand   = invalidError{message: "invalid quiz command", reason: "invalid_quiz_command"}
)
