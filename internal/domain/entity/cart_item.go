package entity

import (
	"github.com/google/uuid"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/value"
)

type CartItem struct {
	ItemID   uuid.UUID
	Quantity value.Quantity
	Price    value.Price
}

func NewCartItem(itemID uuid.UUID, quantity value.Quantity, price value.Price) *CartItem {
	return &CartItem{
		ItemID:   itemID,
		Quantity: quantity,
		Price:    price,
	}
}

func (ci *CartItem) GetItemID() uuid.UUID {
	return ci.ItemID
}

func (ci *CartItem) GetQuantity() value.Quantity {
	return ci.Quantity
}

func (ci *CartItem) GetPrice() value.Price {
	return ci.Price
}

func (ci *CartItem) GetTotal() value.Price {
	// Calculate total: price * quantity
	totalFloat := ci.Price.Float64() * float64(ci.Quantity.Int())
	// value.Price validation is handled in NewPrice
	total, _ := value.NewPrice(totalFloat) // Should never error since individual price was already validated
	return total
}
