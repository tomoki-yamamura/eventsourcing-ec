package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/event"
)

type OutboxStatus string

const (
	OutboxStatusPending   OutboxStatus = "PENDING"
	OutboxStatusPublished OutboxStatus = "PUBLISHED"
	OutboxStatusFailed    OutboxStatus = "FAILED"
)

type OutboxEvent struct {
	ID            int64        `json:"id"`
	EventID       uuid.UUID    `json:"event_id"`
	AggregateID   uuid.UUID    `json:"aggregate_id"`
	AggregateType string       `json:"aggregate_type"`
	EventType     string       `json:"event_type"`
	EventData     []byte       `json:"event_data"`
	Version       int          `json:"version"`
	CreatedAt     time.Time    `json:"created_at"`
	PublishedAt   *time.Time   `json:"published_at,omitempty"`
	Status        OutboxStatus `json:"status"`
	RetryCount    int          `json:"retry_count"`
	ErrorMessage  *string      `json:"error_message,omitempty"`
}

type OutboxRepository interface {
	SaveEvents(ctx context.Context, aggregateID uuid.UUID, events []event.Event) error
	GetPendingEvents(ctx context.Context, limit int) ([]OutboxEvent, error)
	MarkAsPublished(ctx context.Context, eventIDs []uuid.UUID) error
	MarkAsFailed(ctx context.Context, eventID uuid.UUID, errorMessage string) error
	IncrementRetryCount(ctx context.Context, eventID uuid.UUID) error
}