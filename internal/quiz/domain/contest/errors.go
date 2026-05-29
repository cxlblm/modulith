package contest

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

type notFoundError struct {
	message string
	reason  string
	cause   error
}

func (e notFoundError) Error() string {
	if e.cause != nil {
		return e.message + ": " + e.cause.Error()
	}
	return e.message
}

func (e notFoundError) Unwrap() error { return e.cause }
func (e notFoundError) NotFound()     {}
func (e notFoundError) Reason() string {
	return e.reason
}

var (
	ErrInvalidContest  = NewInvalidContest(nil)
	ErrContestNotFound = NewContestNotFound(nil)
)

func NewInvalidContest(cause error) error {
	return invalidError{message: "invalid contest", reason: "invalid_contest", cause: cause}
}

func NewContestNotFound(cause error) error {
	return notFoundError{message: "contest not found", reason: "contest_not_found", cause: cause}
}
