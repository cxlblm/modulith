package mysql

import "testing"

func TestOrderDTOsGroupsItemsByOrder(t *testing.T) {
	models := []OrderModel{
		{UUID: "order-1", UserID: "user-1", AddressID: "addr-1", Receiver: "Ada", Phone: "138", City: "Shanghai", Detail: "Road", Status: "placed", TotalCents: 2000},
		{UUID: "order-2", UserID: "user-1", AddressID: "addr-1", Receiver: "Ada", Phone: "138", City: "Shanghai", Detail: "Road", Status: "placed", TotalCents: 3000},
	}
	items := []OrderItemModel{
		{OrderUUID: "order-2", ProductUUID: "product-2", ProductName: "Mouse", UnitPriceCents: 3000, Qty: 1},
		{OrderUUID: "order-1", ProductUUID: "product-1", ProductName: "Keyboard", UnitPriceCents: 1000, Qty: 2},
	}

	got := orderDTOs(models, items)

	if len(got) != 2 {
		t.Fatalf("len(orderDTOs()) = %d, want 2", len(got))
	}
	if got[0].ID != "order-1" || len(got[0].Items) != 1 || got[0].Items[0].ProductID != "product-1" {
		t.Fatalf("first order items = %#v", got[0])
	}
	if got[1].ID != "order-2" || len(got[1].Items) != 1 || got[1].Items[0].ProductID != "product-2" {
		t.Fatalf("second order items = %#v", got[1])
	}
}
