package product

import (
	"errors"
	"testing"
)

func TestNewProductNotFoundWrapsCauseAndMatchesSentinel(t *testing.T) {
	cause := errors.New("record missing")
	err := NewProductNotFound(cause)

	if !errors.Is(err, ErrProductNotFound) {
		t.Fatalf("errors.Is(err, ErrProductNotFound) = false, err = %v", err)
	}
	if !errors.Is(err, cause) {
		t.Fatalf("errors.Is(err, cause) = false, err = %v", err)
	}
	var notFound interface {
		NotFound()
	}
	if !errors.As(err, &notFound) {
		t.Fatal("wrapped product not found does not expose NotFound()")
	}
	var reasoned interface {
		Reason() string
	}
	if !errors.As(err, &reasoned) {
		t.Fatal("wrapped product not found does not expose Reason()")
	}
	if reasoned.Reason() != "product_not_found" {
		t.Fatalf("Reason() = %q, want %q", reasoned.Reason(), "product_not_found")
	}
}

func TestNewInsufficientStockWrapsCauseAndMatchesSentinel(t *testing.T) {
	cause := errors.New("stock row locked")
	err := NewInsufficientStock(cause)

	if !errors.Is(err, ErrInsufficientStock) {
		t.Fatalf("errors.Is(err, ErrInsufficientStock) = false, err = %v", err)
	}
	if !errors.Is(err, cause) {
		t.Fatalf("errors.Is(err, cause) = false, err = %v", err)
	}
	var conflict interface {
		Conflict()
	}
	if !errors.As(err, &conflict) {
		t.Fatal("wrapped insufficient stock does not expose Conflict()")
	}
}
