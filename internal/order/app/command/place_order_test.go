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
	reserveCalls int
}

func (s *fakeProductsService) ReserveStock(context.Context, string, string, int) error {
	s.reserveCalls++
	return nil
}

type fakePricingService struct {
	err    error
	calls  int
	result PricingResult
}

func (s *fakePricingService) CalculateOrderPricing(context.Context, PricingRequest) (PricingResult, error) {
	s.calls++
	if s.err != nil {
		return PricingResult{}, s.err
	}
	if len(s.result.Items) != 0 {
		return s.result, nil
	}
	return PricingResult{
		OriginalTotalCents: 2000,
		DiscountTotalCents: 300,
		PayableTotalCents:  1700,
		Items: []PricingItemResult{{
			ProductID:              "product-1",
			ProductName:            "Keyboard",
			Qty:                    2,
			OriginalUnitPriceCents: 1000,
			OriginalSubtotalCents:  2000,
			DiscountCents:          300,
			PayableCents:           1700,
			AppliedPromotions: []AppliedPromotionResult{{
				UUID:          "promo-1",
				Name:          "Spend 2000 save 300",
				DiscountCents: 300,
			}},
		}},
		AppliedPromotions: []AppliedPromotionResult{{
			UUID:          "promo-1",
			Name:          "Spend 2000 save 300",
			DiscountCents: 300,
		}},
	}, nil
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
	pricing := &fakePricingService{}
	orders := &fakeOrderRepository{}
	h := PlaceOrderHandler{Orders: orders, Products: products, Addresses: addresses, Users: users, Pricing: pricing}

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
	if pricing.calls != 0 {
		t.Fatalf("pricing calls = %d, want 0", pricing.calls)
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
	pricing := &fakePricingService{}
	orders := &fakeOrderRepository{}
	h := PlaceOrderHandler{Orders: orders, Products: products, Addresses: addresses, Users: users, Pricing: pricing}

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
	if pricing.calls != 1 {
		t.Fatalf("pricing calls = %d, want 1", pricing.calls)
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
	if orders.order.TotalCents() != 1700 {
		t.Fatalf("saved order total = %d, want 1700", orders.order.TotalCents())
	}
	if got := orders.order.Items()[0].AppliedPromotions[0].UUID; got != "promo-1" {
		t.Fatalf("saved order promotion = %q, want promo-1", got)
	}
}

func TestPlaceOrder_PricingErrorReturnsBeforeStockReservationAndSave(t *testing.T) {
	pricingErr := errors.New("pricing unavailable")
	users := &fakeUserEligibilityService{}
	addresses := &fakeAddressService{}
	products := &fakeProductsService{}
	pricing := &fakePricingService{err: pricingErr}
	orders := &fakeOrderRepository{}
	h := PlaceOrderHandler{Orders: orders, Products: products, Addresses: addresses, Users: users, Pricing: pricing}

	_, err := h.Handle(context.Background(), PlaceOrder{
		UserID:    "user-1",
		AddressID: "address-1",
		Items:     []PlaceOrderItem{{ProductID: "product-1", Qty: 2}},
	})

	if !errors.Is(err, pricingErr) {
		t.Fatalf("Handle() error = %v, want %v", err, pricingErr)
	}
	if products.reserveCalls != 0 {
		t.Fatalf("reserve calls = %d, want 0", products.reserveCalls)
	}
	if orders.saveCalls != 0 {
		t.Fatalf("save calls = %d, want 0", orders.saveCalls)
	}
}
