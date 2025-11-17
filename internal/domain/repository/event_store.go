package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/event"
)

type EventStore interface {
	SaveEvents(ctx context.Context, aggregateID uuid.UUID, aggregateType string, events []event.Event) error
	LoadEvents(ctx context.Context, aggregateID uuid.UUID, aggregateType string) ([]event.Event, error)
}
