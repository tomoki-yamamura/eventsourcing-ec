package event

import (
	"time"

	"github.com/google/uuid"
)

type CartSubmittedEvent struct {
	AggregateID uuid.UUID `json:"aggregate_id"`
	TotalAmount float64   `json:"total_amount"`
	SubmittedAt time.Time `json:"submitted_at"`
	EventID     uuid.UUID `json:"event_id"`
	Timestamp   time.Time `json:"timestamp"`
	Version     int       `json:"version"`
}

func NewCartSubmittedEvent(aggregateID uuid.UUID, version int, totalAmount float64) *CartSubmittedEvent {
	return &CartSubmittedEvent{
		AggregateID: aggregateID,
		TotalAmount: totalAmount,
		SubmittedAt: time.Now(),
		EventID:     uuid.New(),
		Timestamp:   time.Now(),
		Version:     version,
	}
}

func (e CartSubmittedEvent) GetAggregateID() uuid.UUID {
	return e.AggregateID
}

func (e CartSubmittedEvent) GetEventID() uuid.UUID {
	return e.EventID
}

func (e CartSubmittedEvent) GetTimestamp() time.Time {
	return e.Timestamp
}

func (e CartSubmittedEvent) GetVersion() int {
	return e.Version
}

func (e CartSubmittedEvent) GetEventType() string {
	return "CartSubmittedEvent"
}

func (e CartSubmittedEvent) GetAggregateType() string {
	return "Cart"
}

func (e *CartSubmittedEvent) GetTotalAmount() float64 {
	return e.TotalAmount
}

func (e *CartSubmittedEvent) GetSubmittedAt() time.Time {
	return e.SubmittedAt
}