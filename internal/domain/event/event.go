package event

import (
	"time"

	"github.com/google/uuid"
)

type Event interface {
	GetAggregateID() uuid.UUID
	GetEventID() uuid.UUID
	GetTimestamp() time.Time
	GetVersion() int
	GetEventType() string
	GetAggregateType() string
}
