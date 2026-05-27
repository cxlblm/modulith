package pricing

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

func (e invalidError) Unwrap() error {
	return e.cause
}

func (e invalidError) Is(target error) bool {
	other, ok := target.(interface {
		Reason() string
	})
	return ok && other.Reason() == e.reason
}

func (e invalidError) Invalid() {}

func (e invalidError) Reason() string {
	return e.reason
}

var ErrInvalidPricing = NewInvalidPricing(nil)

func NewInvalidPricing(cause error) error {
	return invalidError{message: "invalid pricing", reason: "invalid_pricing", cause: cause}
}
