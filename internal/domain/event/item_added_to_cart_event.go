package event

import (
	"time"

	"github.com/google/uuid"
)

type ItemAddedToCartEvent struct {
	AggregateID uuid.UUID
	ItemID      uuid.UUID
	Name        string
	Price       float64
	TenantID    uuid.UUID
	EventID     uuid.UUID
	Timestamp   time.Time
	Version     int
}

func NewItemAddedToCartEvent(aggregateID uuid.UUID, version int, itemID uuid.UUID, name string, price float64, tenantID uuid.UUID) *ItemAddedToCartEvent {
	return &ItemAddedToCartEvent{
		AggregateID: aggregateID,
		ItemID:      itemID,
		Name:        name,
		Price:       price,
		TenantID:    tenantID,
		EventID:     uuid.New(),
		Timestamp:   time.Now(),
		Version:     version,
	}
}

func (e ItemAddedToCartEvent) GetAggregateID() uuid.UUID {
	return e.AggregateID
}

func (e ItemAddedToCartEvent) GetEventID() uuid.UUID {
	return e.EventID
}

func (e ItemAddedToCartEvent) GetTimestamp() time.Time {
	return e.Timestamp
}

func (e ItemAddedToCartEvent) GetVersion() int {
	return e.Version
}

func (e ItemAddedToCartEvent) GetEventType() string {
	return "ItemAddedToCartEvent"
}

func (e ItemAddedToCartEvent) GetAggregateType() string {
	return "Cart"
}

func (e *ItemAddedToCartEvent) GetItemID() uuid.UUID {
	return e.ItemID
}

func (e *ItemAddedToCartEvent) GetName() string {
	return e.Name
}

func (e *ItemAddedToCartEvent) GetPrice() float64 {
	return e.Price
}

func (e *ItemAddedToCartEvent) GetTenantID() uuid.UUID {
	return e.TenantID
}
