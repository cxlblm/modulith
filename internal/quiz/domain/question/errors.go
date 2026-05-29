package question

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
	ErrInvalidQuestion  = NewInvalidQuestion(nil)
	ErrQuestionNotFound = NewQuestionNotFound(nil)
)

func NewInvalidQuestion(cause error) error {
	return invalidError{message: "invalid question", reason: "invalid_question", cause: cause}
}

func NewQuestionNotFound(cause error) error {
	return notFoundError{message: "question not found", reason: "question_not_found", cause: cause}
}
