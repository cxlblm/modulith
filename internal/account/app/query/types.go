package query

import "time"

type UserDTO struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AddressDTO struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Receiver  string    `json:"receiver"`
	Phone     string    `json:"phone"`
	City      string    `json:"city"`
	Detail    string    `json:"detail"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
