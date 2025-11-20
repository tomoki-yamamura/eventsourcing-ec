package command

import "github.com/google/uuid"

type AddItemToCartCommand struct {
	CartID   uuid.UUID
	UserID   uuid.UUID
	ItemID   uuid.UUID
	Quantity int
	Price    float64
}
