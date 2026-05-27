package order

import (
	"errors"
	"reflect"
	"testing"
)

func TestNewOrderPlacesOrderAndRecordsEvent(t *testing.T) {
	o, err := NewOrder("user-1", "addr-1", AddressSnapshot{Receiver: "Ada", Phone: "138", City: "Shanghai", Detail: "Road"}, []Item{
		{
			ProductUUID:            "product-1",
			ProductName:            "Keyboard",
			OriginalUnitPriceCents: 1000,
			OriginalSubtotalCents:  2000,
			DiscountCents:          300,
			PayableCents:           1700,
			Qty:                    2,
		},
	})
	if err != nil {
		t.Fatalf("NewOrder() error = %v", err)
	}
	if o.Status() != StatusPlaced {
		t.Fatalf("Status() = %q, want %q", o.Status(), StatusPlaced)
	}
	if o.TotalCents() != 1700 {
		t.Fatalf("TotalCents() = %d, want %d", o.TotalCents(), 1700)
	}
	if len(o.PeekEvents()) != 1 {
		t.Fatalf("len(PeekEvents()) = %d, want 1", len(o.PeekEvents()))
	}
}

func TestOrderUsesUUIDTerminology(t *testing.T) {
	o, err := NewOrder("user-1", "addr-1", AddressSnapshot{Receiver: "Ada", Phone: "138", City: "Shanghai", Detail: "Road"}, []Item{
		validOrderItem(),
	})
	if err != nil {
		t.Fatalf("NewOrder() error = %v", err)
	}
	var uuid OrderUUID = o.UUID()
	if uuid == "" {
		t.Fatal("UUID() is empty")
	}
	if _, ok := reflect.TypeOf(o).MethodByName("ID"); ok {
		t.Fatal("Order exposes ID(), want UUID()")
	}
}

func TestNewOrderRejectsEmptyAddressSnapshot(t *testing.T) {
	_, err := NewOrder("user-1", "addr-1", AddressSnapshot{}, []Item{
		validOrderItem(),
	})
	if !errors.Is(err, ErrInvalidOrder) {
		t.Fatalf("NewOrder() error = %v, want ErrInvalidOrder", err)
	}
}

func TestOrder_MarkPaidIsIdempotent(t *testing.T) {
	o, err := NewOrder("user-1", "addr-1", AddressSnapshot{Receiver: "Ada", Phone: "138", City: "Shanghai", Detail: "Road"}, []Item{
		{
			ProductUUID:            "product-1",
			ProductName:            "Keyboard",
			OriginalUnitPriceCents: 1000,
			OriginalSubtotalCents:  1000,
			DiscountCents:          0,
			PayableCents:           1000,
			Qty:                    1,
		},
	})
	if err != nil {
		t.Fatalf("NewOrder() error = %v", err)
	}

	if err := o.MarkPaid("payment-1"); err != nil {
		t.Fatalf("MarkPaid() error = %v", err)
	}
	if err := o.MarkPaid("payment-1"); err != nil {
		t.Fatalf("MarkPaid() duplicate error = %v", err)
	}
	if o.Status() != StatusPaid {
		t.Fatalf("Status() = %q, want %q", o.Status(), StatusPaid)
	}
}

func TestNewOrderRejectsInconsistentPricingSnapshot(t *testing.T) {
	_, err := NewOrder("user-1", "addr-1", AddressSnapshot{Receiver: "Ada", Phone: "138", City: "Shanghai", Detail: "Road"}, []Item{
		{
			ProductUUID:            "product-1",
			ProductName:            "Keyboard",
			OriginalUnitPriceCents: 1000,
			OriginalSubtotalCents:  2000,
			DiscountCents:          300,
			PayableCents:           1800,
			Qty:                    2,
		},
	})

	if !errors.Is(err, ErrInvalidOrder) {
		t.Fatalf("NewOrder() error = %v, want ErrInvalidOrder", err)
	}
}

func TestOrderItemsDefensivelyCopyAppliedPromotions(t *testing.T) {
	item := validOrderItem()
	item.AppliedPromotions = []AppliedPromotion{{UUID: "promo-1", Name: "Spend 100 save 10", DiscountCents: 10}}
	o, err := NewOrder(
		"user-1",
		"addr-1",
		AddressSnapshot{Receiver: "Ada", Phone: "138", City: "Shanghai", Detail: "Road"},
		[]Item{item},
	)
	if err != nil {
		t.Fatalf("NewOrder() error = %v", err)
	}

	item.AppliedPromotions[0].Name = "mutated"
	got := o.Items()
	got[0].AppliedPromotions[0].Name = "also-mutated"

	again := o.Items()
	if again[0].AppliedPromotions[0].Name != "Spend 100 save 10" {
		t.Fatalf("promotion name = %q, want original snapshot", again[0].AppliedPromotions[0].Name)
	}
}

func validOrderItem() Item {
	return Item{
		ProductUUID:            "product-1",
		ProductName:            "Keyboard",
		OriginalUnitPriceCents: 1000,
		OriginalSubtotalCents:  2000,
		DiscountCents:          300,
		PayableCents:           1700,
		Qty:                    2,
	}
}
