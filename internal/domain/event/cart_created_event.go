package event

import (
	"time"

	"github.com/google/uuid"
)

type CartCreatedEvent struct {
	AggregateID uuid.UUID
	UserID      uuid.UUID
	TenantID    uuid.UUID
	EventID     uuid.UUID
	Timestamp   time.Time
	Version     int
}

func NewCartCreatedEvent(aggregateID uuid.UUID, version int, userID uuid.UUID, tenantID uuid.UUID) *CartCreatedEvent {
	return &CartCreatedEvent{
		AggregateID: aggregateID,
		UserID:      userID,
		TenantID:    tenantID,
		EventID:     uuid.New(),
		Timestamp:   time.Now(),
		Version:     version,
	}
}

func (e CartCreatedEvent) GetAggregateID() uuid.UUID {
	return e.AggregateID
}

func (e CartCreatedEvent) GetEventID() uuid.UUID {
	return e.EventID
}

func (e CartCreatedEvent) GetTimestamp() time.Time {
	return e.Timestamp
}

func (e CartCreatedEvent) GetVersion() int {
	return e.Version
}

func (e CartCreatedEvent) GetEventType() string {
	return "CartCreatedEvent"
}

func (e CartCreatedEvent) GetAggregateType() string {
	return "Cart"
}

func (e *CartCreatedEvent) GetUserID() uuid.UUID {
	return e.UserID
}

func (e *CartCreatedEvent) GetTenantID() uuid.UUID {
	return e.TenantID
}
