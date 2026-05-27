package user

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

func (e notFoundError) Unwrap() error {
	return e.cause
}

func (e notFoundError) Is(target error) bool {
	other, ok := target.(interface {
		Reason() string
	})
	return ok && other.Reason() == e.reason
}

func (e notFoundError) NotFound() {}

func (e notFoundError) Reason() string {
	return e.reason
}

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

func (e forbiddenError) Unwrap() error {
	return e.cause
}

func (e forbiddenError) Is(target error) bool {
	other, ok := target.(interface {
		Reason() string
	})
	return ok && other.Reason() == e.reason
}

func (e forbiddenError) Forbidden() {}

func (e forbiddenError) Reason() string {
	return e.reason
}

var (
	ErrUserNotFound    = NewUserNotFound(nil)
	ErrAddressNotFound = NewAddressNotFound(nil)
	ErrInvalidUser     = NewInvalidUser(nil)
	ErrInvalidAddress  = NewInvalidAddress(nil)
	ErrUserDisabled    = NewUserDisabled(nil)
)

func NewUserNotFound(cause error) error {
	return notFoundError{message: "user not found", reason: "user_not_found", cause: cause}
}

func NewAddressNotFound(cause error) error {
	return notFoundError{message: "address not found", reason: "address_not_found", cause: cause}
}

func NewInvalidUser(cause error) error {
	return invalidError{message: "invalid user", reason: "invalid_user", cause: cause}
}

func NewInvalidAddress(cause error) error {
	return invalidError{message: "invalid address", reason: "invalid_address", cause: cause}
}

func NewUserDisabled(cause error) error {
	return forbiddenError{message: "user disabled", reason: "user_disabled", cause: cause}
}
