package participation

type invalidError struct {
	message string
	reason  string
	cause   error
}

func (e invalidError) Error() string {
	if e.cause != nil {
		return e.message + ": " + e.cause.Error()
	}
	return e.message
}

func (e invalidError) Unwrap() error { return e.cause }
func (e invalidError) Invalid()      {}
func (e invalidError) Reason() string {
	return e.reason
}

type forbiddenError struct {
	message string
	reason  string
	cause   error
}

func (e forbiddenError) Error() string {
	if e.cause != nil {
		return e.message + ": " + e.cause.Error()
	}
	return e.message
}

func (e forbiddenError) Unwrap() error { return e.cause }
func (e forbiddenError) Forbidden()    {}
func (e forbiddenError) Reason() string {
	return e.reason
}

var (
	ErrInvalidParticipation    = NewInvalidParticipation(nil)
	ErrParticipantEliminated   = NewParticipantEliminated(nil)
	ErrQuestionAlreadyAnswered = NewQuestionAlreadyAnswered(nil)
)

func NewInvalidParticipation(cause error) error {
	return invalidError{message: "invalid participation", reason: "invalid_participation", cause: cause}
}

func NewParticipantEliminated(cause error) error {
	return forbiddenError{message: "participant eliminated", reason: "participant_eliminated", cause: cause}
}

func NewQuestionAlreadyAnswered(cause error) error {
	return invalidError{message: "question already answered", reason: "question_already_answered", cause: cause}
}
