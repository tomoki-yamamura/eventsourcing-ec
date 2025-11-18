package event

import (
	"time"

	"github.com/google/uuid"
)

// ItemAddedToCartEvent represents the event when an item is added to the cart
type ItemAddedToCartEvent struct {
	AggregateID uuid.UUID `json:"aggregate_id"`
	ItemID      uuid.UUID `json:"item_id"`
	Quantity    int        `json:"quantity"`
	Price       float64    `json:"price"`
	EventID     uuid.UUID `json:"event_id"`
	Timestamp   time.Time `json:"timestamp"`
	Version     int       `json:"version"`
}

// NewItemAddedToCartEvent creates a new ItemAddedToCartEvent
func NewItemAddedToCartEvent(aggregateID uuid.UUID, version int, itemID uuid.UUID, quantity int, price float64) *ItemAddedToCartEvent {
	return &ItemAddedToCartEvent{
		AggregateID: aggregateID,
		ItemID:      itemID,
		Quantity:    quantity,
		Price:       price,
		EventID:     uuid.New(),
		Timestamp:   time.Now(),
		Version:     version,
	}
}

// GetAggregateID returns the aggregate ID
func (e ItemAddedToCartEvent) GetAggregateID() uuid.UUID {
	return e.AggregateID
}

// GetEventID returns the event ID
func (e ItemAddedToCartEvent) GetEventID() uuid.UUID {
	return e.EventID
}

// GetTimestamp returns the timestamp
func (e ItemAddedToCartEvent) GetTimestamp() time.Time {
	return e.Timestamp
}

// GetVersion returns the version
func (e ItemAddedToCartEvent) GetVersion() int {
	return e.Version
}

// GetEventType returns the event type
func (e ItemAddedToCartEvent) GetEventType() string {
	return "ItemAddedToCartEvent"
}

// GetAggregateType returns the aggregate type
func (e ItemAddedToCartEvent) GetAggregateType() string {
	return "Cart"
}

// GetItemID returns the item ID
func (e *ItemAddedToCartEvent) GetItemID() uuid.UUID {
	return e.ItemID
}

// GetQuantity returns the quantity
func (e *ItemAddedToCartEvent) GetQuantity() int {
	return e.Quantity
}

// GetPrice returns the price
func (e *ItemAddedToCartEvent) GetPrice() float64 {
	return e.Price
}