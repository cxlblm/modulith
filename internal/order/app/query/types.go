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
	ProductID      string    `json:"product_id"`
	ProductName    string    `json:"product_name"`
	UnitPriceCents int64     `json:"unit_price_cents"`
	Qty            int       `json:"qty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
