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

var ErrInvalidCommand error = invalidError{message: "invalid entitlement command", reason: "invalid_entitlement_command"}
