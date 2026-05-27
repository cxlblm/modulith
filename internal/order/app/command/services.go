package command

import (
	"context"

	orderdomain "modular_monolith/internal/order/domain/order"
)

type UnitOfWork interface {
	RunInTx(ctx context.Context, fn func(ctx context.Context, repos Repositories) error) error
}

type Repositories struct {
	Orders orderdomain.Repository
}

type ProductsService interface {
	ReserveStock(ctx context.Context, productID string, orderID string, qty int) error
}

type AddressService interface {
	GetAddress(ctx context.Context, userID string, addressID string) (AddressInfo, error)
}

type AddressInfo struct {
	ID       string
	UserID   string
	Receiver string
	Phone    string
	City     string
	Detail   string
}

type UserEligibilityService interface {
	EnsureCanPlaceOrder(ctx context.Context, userID string) error
}

type PricingService interface {
	CalculateOrderPricing(ctx context.Context, req PricingRequest) (PricingResult, error)
}

type PricingRequest struct {
	UserID string
	Items  []PricingRequestItem
}

type PricingRequestItem struct {
	ProductID string
	Qty       int
}

type AppliedPromotionResult struct {
	UUID          string
	Name          string
	DiscountCents int64
}

type PricingItemResult struct {
	ProductID              string
	ProductName            string
	Qty                    int
	OriginalUnitPriceCents int64
	OriginalSubtotalCents  int64
	DiscountCents          int64
	PayableCents           int64
	AppliedPromotions      []AppliedPromotionResult
}

type PricingResult struct {
	OriginalTotalCents int64
	DiscountTotalCents int64
	PayableTotalCents  int64
	Items              []PricingItemResult
	AppliedPromotions  []AppliedPromotionResult
}
