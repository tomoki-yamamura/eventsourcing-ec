package command

import "github.com/google/uuid"

// AddItemToCartCommand represents a command to add an item to the cart
type AddItemToCartCommand struct {
	CartID   uuid.UUID
	ItemID   uuid.UUID
	Quantity int
	Price    float64
}