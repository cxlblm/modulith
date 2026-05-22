package order

import (
	"errors"
	"fmt"
	"testing"
)

func TestErrOrderNotFoundIsCustomNotFoundError(t *testing.T) {
	var notFound interface {
		NotFound()
	}
	if !errors.As(ErrOrderNotFound, &notFound) {
		t.Fatal("ErrOrderNotFound does not expose NotFound()")
	}
}

func TestErrOrderNotFoundHasStableReason(t *testing.T) {
	var reasoned interface {
		Reason() string
	}
	if !errors.As(ErrOrderNotFound, &reasoned) {
		t.Fatal("ErrOrderNotFound does not expose Reason()")
	}
	if reasoned.Reason() != "order_not_found" {
		t.Fatalf("Reason() = %q, want %q", reasoned.Reason(), "order_not_found")
	}
}

func TestNewOrderNotFoundWrapsCauseAndMatchesSentinel(t *testing.T) {
	cause := errors.New("record missing")
	err := NewOrderNotFound(cause)

	if !errors.Is(err, ErrOrderNotFound) {
		t.Fatalf("errors.Is(err, ErrOrderNotFound) = false, err = %v", err)
	}
	if !errors.Is(err, cause) {
		t.Fatalf("errors.Is(err, cause) = false, err = %v", err)
	}
	var notFound interface {
		NotFound()
	}
	if !errors.As(err, &notFound) {
		t.Fatal("wrapped order not found does not expose NotFound()")
	}
}

func TestWrappedOrderNotFoundSurvivesAdditionalWrapping(t *testing.T) {
	cause := errors.New("record missing")
	err := fmt.Errorf("get order: %w", NewOrderNotFound(cause))

	if !errors.Is(err, ErrOrderNotFound) {
		t.Fatalf("errors.Is(err, ErrOrderNotFound) = false, err = %v", err)
	}
	if !errors.Is(err, cause) {
		t.Fatalf("errors.Is(err, cause) = false, err = %v", err)
	}
}
