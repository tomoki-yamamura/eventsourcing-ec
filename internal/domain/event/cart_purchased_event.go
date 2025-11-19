package event

import (
	"time"

	"github.com/google/uuid"
)

type CartPurchasedEvent struct {
	AggregateID uuid.UUID `json:"aggregate_id"`
	TotalAmount float64   `json:"total_amount"`
	PurchasedAt time.Time `json:"purchased_at"`
	EventID     uuid.UUID `json:"event_id"`
	Timestamp   time.Time `json:"timestamp"`
	Version     int       `json:"version"`
}

func NewCartPurchasedEvent(aggregateID uuid.UUID, version int, totalAmount float64) *CartPurchasedEvent {
	return &CartPurchasedEvent{
		AggregateID: aggregateID,
		TotalAmount: totalAmount,
		PurchasedAt: time.Now(),
		EventID:     uuid.New(),
		Timestamp:   time.Now(),
		Version:     version,
	}
}

func (e CartPurchasedEvent) GetAggregateID() uuid.UUID {
	return e.AggregateID
}

func (e CartPurchasedEvent) GetEventID() uuid.UUID {
	return e.EventID
}

func (e CartPurchasedEvent) GetTimestamp() time.Time {
	return e.Timestamp
}

func (e CartPurchasedEvent) GetVersion() int {
	return e.Version
}

func (e CartPurchasedEvent) GetEventType() string {
	return "CartPurchasedEvent"
}

func (e CartPurchasedEvent) GetAggregateType() string {
	return "Cart"
}

func (e *CartPurchasedEvent) GetTotalAmount() float64 {
	return e.TotalAmount
}

func (e *CartPurchasedEvent) GetPurchasedAt() time.Time {
	return e.PurchasedAt
}
