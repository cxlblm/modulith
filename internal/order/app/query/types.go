package query

import "time"

type OrderDTO struct {
	ID         string         `json:"id"`
	UserID     string         `json:"user_id"`
	AddressID  string         `json:"address_id"`
	Status     string         `json:"status"`
	TotalCents int64          `json:"total_cents"`
	PaymentID  string         `json:"payment_id,omitempty"`
	ShipmentID string         `json:"shipment_id,omitempty"`
	Address    AddressDTO     `json:"address"`
	Items      []OrderItemDTO `json:"items"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

type AddressDTO struct {
	Receiver string `json:"receiver"`
	Phone    string `json:"phone"`
	City     string `json:"city"`
	Detail   string `json:"detail"`
}

type OrderItemDTO struct {
	ProductID              string                `json:"product_id"`
	ProductName            string                `json:"product_name"`
	UnitPriceCents         int64                 `json:"unit_price_cents"`
	OriginalUnitPriceCents int64                 `json:"original_unit_price_cents"`
	OriginalSubtotalCents  int64                 `json:"original_subtotal_cents"`
	DiscountCents          int64                 `json:"discount_cents"`
	PayableCents           int64                 `json:"payable_cents"`
	Qty                    int                   `json:"qty"`
	AppliedPromotions      []AppliedPromotionDTO `json:"applied_promotions"`
	CreatedAt              time.Time             `json:"created_at"`
	UpdatedAt              time.Time             `json:"updated_at"`
}

type AppliedPromotionDTO struct {
	UUID          string `json:"uuid"`
	Name          string `json:"name"`
	DiscountCents int64  `json:"discount_cents"`
}
