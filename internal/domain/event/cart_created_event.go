package event

import (
	"time"

	"github.com/google/uuid"
)

type CartCreatedEvent struct {
	AggregateID uuid.UUID `json:"aggregate_id"`
	UserID      uuid.UUID `json:"user_id"`
	EventID     uuid.UUID `json:"event_id"`
	Timestamp   time.Time `json:"timestamp"`
	Version     int       `json:"version"`
}

func NewCartCreatedEvent(aggregateID uuid.UUID, version int, userID uuid.UUID) *CartCreatedEvent {
	return &CartCreatedEvent{
		AggregateID: aggregateID,
		UserID:      userID,
		EventID:     uuid.New(),
		Timestamp:   time.Now(),
		Version:     version,
	}
}

// GetAggregateID returns the aggregate ID
func (e CartCreatedEvent) GetAggregateID() uuid.UUID {
	return e.AggregateID
}

// GetEventID returns the event ID
func (e CartCreatedEvent) GetEventID() uuid.UUID {
	return e.EventID
}

// GetTimestamp returns the timestamp
func (e CartCreatedEvent) GetTimestamp() time.Time {
	return e.Timestamp
}

// GetVersion returns the version
func (e CartCreatedEvent) GetVersion() int {
	return e.Version
}

// GetEventType returns the event type
func (e CartCreatedEvent) GetEventType() string {
	return "CartCreatedEvent"
}

// GetAggregateType returns the aggregate type
func (e CartCreatedEvent) GetAggregateType() string {
	return "Cart"
}

// GetUserID returns the user ID associated with this cart
func (e *CartCreatedEvent) GetUserID() uuid.UUID {
	return e.UserID
}
