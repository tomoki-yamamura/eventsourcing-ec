package dto

import (
	"time"
)

type CartViewDTO struct {
	ID          string            `json:"id"`
	UserID      string            `json:"user_id"`
	Status      string            `json:"status"`
	TotalAmount float64           `json:"total_amount"`
	ItemCount   int               `json:"item_count"`
	Items       []CartItemViewDTO `json:"items"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	PurchasedAt *time.Time        `json:"purchased_at,omitempty"`
	Version     int               `json:"version"`
}

type CartItemViewDTO struct {
	ID     string  `json:"id"`
	CartID string  `json:"cart_id"`
	Name   string  `json:"name"`
	Price  float64 `json:"price"`
}
