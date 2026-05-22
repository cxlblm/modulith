package httpserver

import (
	"errors"
	"net/http"
	"strings"
	"testing"
)

type testNotFoundError struct {
	cause error
}

func (e testNotFoundError) Error() string {
	if e.cause != nil {
		return "thing not found: " + e.cause.Error()
	}
	return "thing not found"
}
func (e testNotFoundError) Unwrap() error { return e.cause }
func (testNotFoundError) NotFound()       {}
func (testNotFoundError) Reason() string {
	return "thing_not_found"
}

func TestErrorResponseMapsAppErrorToHTTPStatusAndReason(t *testing.T) {
	status, body := errorResponse(testNotFoundError{cause: errors.New("database row 42 missing")})

	if status != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", status, http.StatusNotFound)
	}
	if body.Error.Code != "not_found" {
		t.Fatalf("error code = %q, want %q", body.Error.Code, "not_found")
	}
	if body.Error.Reason != "thing_not_found" {
		t.Fatalf("reason = %q, want %q", body.Error.Reason, "thing_not_found")
	}
	if strings.Contains(body.Error.Message, "database row 42 missing") {
		t.Fatalf("response message leaks internal cause: %q", body.Error.Message)
	}
}
