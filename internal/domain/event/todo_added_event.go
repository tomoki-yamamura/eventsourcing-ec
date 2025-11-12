package event

import (
	"time"

	"github.com/google/uuid"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/value"
)

type TodoAddedEvent struct {
	AggregateID uuid.UUID
	UserID      value.UserID
	TodoText    value.TodoText
	EventID     uuid.UUID
	Timestamp   time.Time
	Version     int
}

func (e TodoAddedEvent) GetAggregateID() uuid.UUID {
	return e.AggregateID
}

func (e TodoAddedEvent) GetEventID() uuid.UUID {
	return e.EventID
}

func (e TodoAddedEvent) GetTimestamp() time.Time {
	return e.Timestamp
}

func (e TodoAddedEvent) GetVersion() int {
	return e.Version
}

func (e TodoAddedEvent) GetEventType() string {
	return "TodoAddedEvent"
}
