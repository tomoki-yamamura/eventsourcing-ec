package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/event"
)

type OutboxRepository interface {
	SaveEvents(ctx context.Context, aggregateID uuid.UUID, events []event.Event) error
	GetPendingEvents(ctx context.Context, limit int) ([]event.OutboxEvent, error)
	MarkAsPublished(ctx context.Context, eventIDs []uuid.UUID) error
	MarkAsFailed(ctx context.Context, eventID uuid.UUID, errorMessage string) error
	IncrementRetryCount(ctx context.Context, eventID uuid.UUID) error
}