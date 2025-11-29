package command

import "github.com/google/uuid"

type AddItemToCartCommand struct {
	CartID   uuid.UUID
	UserID   uuid.UUID
	ItemID   uuid.UUID
	Name     string
	Price    float64
	TenantID uuid.UUID
}
