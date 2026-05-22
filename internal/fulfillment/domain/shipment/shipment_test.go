package shipment

import (
	"reflect"
	"testing"
)

func TestShipment_SendIsIdempotent(t *testing.T) {
	s, err := NewShipment("order-1")
	if err != nil {
		t.Fatalf("NewShipment() error = %v", err)
	}

	if err := s.Send(); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if err := s.Send(); err != nil {
		t.Fatalf("Send() duplicate error = %v", err)
	}
	if s.Status() != StatusSent {
		t.Fatalf("Status() = %q, want %q", s.Status(), StatusSent)
	}
}

func TestShipmentUsesUUIDTerminology(t *testing.T) {
	s, err := NewShipment("order-1")
	if err != nil {
		t.Fatalf("NewShipment() error = %v", err)
	}
	var uuid ShipmentUUID = s.UUID()
	if uuid == "" {
		t.Fatal("UUID() is empty")
	}
	if _, ok := reflect.TypeOf(s).MethodByName("ID"); ok {
		t.Fatal("Shipment exposes ID(), want UUID()")
	}
}
