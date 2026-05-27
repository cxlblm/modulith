package command

import (
	"context"
	"errors"
	"testing"

	orderdomain "modular_monolith/internal/order/domain/order"
)

type fakeUserEligibilityService struct {
	err   error
	calls int
}

func (s *fakeUserEligibilityService) EnsureCanPlaceOrder(context.Context, string) error {
	s.calls++
	return s.err
}

type fakeAddressService struct {
	calls int
}

func (s *fakeAddressService) GetAddress(context.Context, string, string) (AddressInfo, error) {
	s.calls++
	return AddressInfo{
		ID:       "address-1",
		UserID:   "user-1",
		Receiver: "Ada",
		Phone:    "13800000000",
		City:     "Shanghai",
		Detail:   "Road 1",
	}, nil
}

type fakeProductsService struct {
	getCalls     int
	reserveCalls int
}

func (s *fakeProductsService) GetProduct(context.Context, string) (ProductInfo, error) {
	s.getCalls++
	return ProductInfo{ID: "product-1", Name: "Keyboard", PriceCents: 1000}, nil
}

func (s *fakeProductsService) ReserveStock(context.Context, string, string, int) error {
	s.reserveCalls++
	return nil
}

type fakeOrderRepository struct {
	saveCalls int
	order     *orderdomain.Order
}

func (r *fakeOrderRepository) Save(_ context.Context, o *orderdomain.Order) error {
	r.saveCalls++
	r.order = o
	return nil
}

func (r *fakeOrderRepository) FindByUUID(context.Context, orderdomain.OrderUUID) (*orderdomain.Order, error) {
	return nil, orderdomain.ErrOrderNotFound
}

func (r *fakeOrderRepository) MarkPaid(context.Context, orderdomain.OrderUUID, string) error {
	return nil
}

func (r *fakeOrderRepository) MarkShipped(context.Context, orderdomain.OrderUUID, string) error {
	return nil
}

func TestPlaceOrder_DisabledUserReturnsBeforeSideEffects(t *testing.T) {
	disabledErr := errors.New("user disabled")
	users := &fakeUserEligibilityService{err: disabledErr}
	addresses := &fakeAddressService{}
	products := &fakeProductsService{}
	orders := &fakeOrderRepository{}
	h := PlaceOrderHandler{Orders: orders, Products: products, Addresses: addresses, Users: users}

	_, err := h.Handle(context.Background(), PlaceOrder{
		UserID:    "user-1",
		AddressID: "address-1",
		Items:     []PlaceOrderItem{{ProductID: "product-1", Qty: 1}},
	})

	if !errors.Is(err, disabledErr) {
		t.Fatalf("Handle() error = %v, want %v", err, disabledErr)
	}
	if users.calls != 1 {
		t.Fatalf("user eligibility calls = %d, want 1", users.calls)
	}
	if addresses.calls != 0 {
		t.Fatalf("address calls = %d, want 0", addresses.calls)
	}
	if products.getCalls != 0 {
		t.Fatalf("product get calls = %d, want 0", products.getCalls)
	}
	if products.reserveCalls != 0 {
		t.Fatalf("reserve calls = %d, want 0", products.reserveCalls)
	}
	if orders.saveCalls != 0 {
		t.Fatalf("save calls = %d, want 0", orders.saveCalls)
	}
}

func TestPlaceOrder_ActiveUserPlacesOrder(t *testing.T) {
	users := &fakeUserEligibilityService{}
	addresses := &fakeAddressService{}
	products := &fakeProductsService{}
	orders := &fakeOrderRepository{}
	h := PlaceOrderHandler{Orders: orders, Products: products, Addresses: addresses, Users: users}

	result, err := h.Handle(context.Background(), PlaceOrder{
		UserID:    "user-1",
		AddressID: "address-1",
		Items:     []PlaceOrderItem{{ProductID: "product-1", Qty: 2}},
	})

	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if result.OrderID == "" {
		t.Fatal("OrderID is empty")
	}
	if users.calls != 1 {
		t.Fatalf("user eligibility calls = %d, want 1", users.calls)
	}
	if addresses.calls != 1 {
		t.Fatalf("address calls = %d, want 1", addresses.calls)
	}
	if products.getCalls != 1 {
		t.Fatalf("product get calls = %d, want 1", products.getCalls)
	}
	if products.reserveCalls != 1 {
		t.Fatalf("reserve calls = %d, want 1", products.reserveCalls)
	}
	if orders.saveCalls != 1 {
		t.Fatalf("save calls = %d, want 1", orders.saveCalls)
	}
	if orders.order == nil {
		t.Fatal("saved order is nil")
	}
}
