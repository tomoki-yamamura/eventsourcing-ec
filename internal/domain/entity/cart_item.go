package entity

import (
	"github.com/google/uuid"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/value"
)

type CartItem struct {
	ItemID uuid.UUID
	Name   string
	Price  value.Price
}

func NewCartItem(itemID uuid.UUID, name string, price value.Price) *CartItem {
	return &CartItem{
		ItemID: itemID,
		Name:   name,
		Price:  price,
	}
}

func (ci *CartItem) GetItemID() uuid.UUID {
	return ci.ItemID
}

func (ci *CartItem) GetName() string {
	return ci.Name
}

func (ci *CartItem) GetPrice() value.Price {
	return ci.Price
}
