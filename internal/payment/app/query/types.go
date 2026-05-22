package query

import "time"

type PaymentDTO struct {
	ID         string    `json:"id"`
	OrderID    string    `json:"order_id"`
	UserID     string    `json:"user_id"`
	TotalCents int64     `json:"total_cents"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
