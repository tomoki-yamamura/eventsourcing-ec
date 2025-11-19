package event

import (
	"time"

	"github.com/google/uuid"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/value"
)

type OutboxEvent struct {
	ID            int64
	EventID       uuid.UUID
	AggregateID   uuid.UUID
	AggregateType string
	EventType     string
	EventData     []byte
	Version       int
	CreatedAt     time.Time
	PublishedAt   *time.Time
	Status        value.OutboxStatus
	RetryCount    int
	ErrorMessage  *string
}
