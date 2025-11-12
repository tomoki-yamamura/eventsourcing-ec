package event

import (
	"time"

	"github.com/google/uuid"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/value"
)

type TodoListCreatedEvent struct {
	AggregateID uuid.UUID
	UserID      value.UserID
	EventID     uuid.UUID
	Timestamp   time.Time
	Version     int
}

func (e TodoListCreatedEvent) GetAggregateID() uuid.UUID {
	return e.AggregateID
}

func (e TodoListCreatedEvent) GetEventID() uuid.UUID {
	return e.EventID
}

func (e TodoListCreatedEvent) GetTimestamp() time.Time {
	return e.Timestamp
}

func (e TodoListCreatedEvent) GetVersion() int {
	return e.Version
}

func (e TodoListCreatedEvent) GetEventType() string {
	return "TodoListCreatedEvent"
}
