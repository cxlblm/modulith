package product

import (
	"errors"
	"reflect"
	"testing"
)

func TestProduct_ReserveStockDeductsOncePerOrder(t *testing.T) {
	p, err := NewProduct("Keyboard", 1000, 5)
	if err != nil {
		t.Fatalf("NewProduct() error = %v", err)
	}

	if err := p.ReserveStock("order-1", 2); err != nil {
		t.Fatalf("ReserveStock() error = %v", err)
	}
	if err := p.ReserveStock("order-1", 2); err != nil {
		t.Fatalf("ReserveStock() duplicate error = %v", err)
	}
	if p.Stock() != 3 {
		t.Fatalf("Stock() = %d, want %d", p.Stock(), 3)
	}
}

func TestProductUsesUUIDTerminology(t *testing.T) {
	p, err := NewProduct("Keyboard", 1000, 5)
	if err != nil {
		t.Fatalf("NewProduct() error = %v", err)
	}
	var uuid ProductUUID = p.UUID()
	if uuid == "" {
		t.Fatal("UUID() is empty")
	}
	if _, ok := reflect.TypeOf(p).MethodByName("ID"); ok {
		t.Fatal("Product exposes ID(), want UUID()")
	}
}

func TestProduct_ReserveStockRejectsShortage(t *testing.T) {
	p, err := NewProduct("Keyboard", 1000, 1)
	if err != nil {
		t.Fatalf("NewProduct() error = %v", err)
	}

	err = p.ReserveStock("order-1", 2)
	if !errors.Is(err, ErrInsufficientStock) {
		t.Fatalf("ReserveStock() error = %v, want ErrInsufficientStock", err)
	}
}
