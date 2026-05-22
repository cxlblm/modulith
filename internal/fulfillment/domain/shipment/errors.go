package shipment

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

var (
	ErrShipmentNotFound = NewShipmentNotFound(nil)
	ErrInvalidShipment  = NewInvalidShipment(nil)
)

func NewShipmentNotFound(cause error) error {
	return notFoundError{message: "shipment not found", reason: "shipment_not_found", cause: cause}
}

func NewInvalidShipment(cause error) error {
	return invalidError{message: "invalid shipment", reason: "invalid_shipment", cause: cause}
}
