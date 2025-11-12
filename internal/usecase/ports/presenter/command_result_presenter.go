package presenter

import (
	"context"

	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/event"
)

type CommandResultPresenter interface {
	PresentSuccess(ctx context.Context, aggregateID string, version int, events []event.Event) error
	PresentError(ctx context.Context, err error) error
}
