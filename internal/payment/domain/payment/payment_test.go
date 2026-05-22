package payment

import (
	"reflect"
	"testing"
)

func TestPayment_ConfirmIsIdempotent(t *testing.T) {
	p, err := NewPayment("order-1", "user-1", 1000)
	if err != nil {
		t.Fatalf("NewPayment() error = %v", err)
	}

	if err := p.Confirm(); err != nil {
		t.Fatalf("Confirm() error = %v", err)
	}
	if err := p.Confirm(); err != nil {
		t.Fatalf("Confirm() duplicate error = %v", err)
	}
	if p.Status() != StatusConfirmed {
		t.Fatalf("Status() = %q, want %q", p.Status(), StatusConfirmed)
	}
}

func TestPaymentUsesUUIDTerminology(t *testing.T) {
	p, err := NewPayment("order-1", "user-1", 100)
	if err != nil {
		t.Fatalf("NewPayment() error = %v", err)
	}
	var uuid PaymentUUID = p.UUID()
	if uuid == "" {
		t.Fatal("UUID() is empty")
	}
	if _, ok := reflect.TypeOf(p).MethodByName("ID"); ok {
		t.Fatal("Payment exposes ID(), want UUID()")
	}
}
