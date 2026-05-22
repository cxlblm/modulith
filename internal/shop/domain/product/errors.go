package product

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

type conflictError struct {
	message string
	reason  string
	cause   error
}

func (e conflictError) Error() string {
	if e.cause != nil {
		return e.message + ": " + e.cause.Error()
	}
	return e.message
}

func (e conflictError) Unwrap() error {
	return e.cause
}

func (e conflictError) Is(target error) bool {
	other, ok := target.(interface {
		Reason() string
	})
	return ok && other.Reason() == e.reason
}

func (e conflictError) Conflict() {}

func (e conflictError) Reason() string {
	return e.reason
}

var (
	ErrProductNotFound   = NewProductNotFound(nil)
	ErrInvalidProduct    = NewInvalidProduct(nil)
	ErrInsufficientStock = NewInsufficientStock(nil)
)

func NewProductNotFound(cause error) error {
	return notFoundError{message: "product not found", reason: "product_not_found", cause: cause}
}

func NewInvalidProduct(cause error) error {
	return invalidError{message: "invalid product", reason: "invalid_product", cause: cause}
}

func NewInsufficientStock(cause error) error {
	return conflictError{message: "insufficient stock", reason: "insufficient_stock", cause: cause}
}
