package gateway

import (
	"context"

	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/event"
)

type Projector interface {
	Handle(ctx context.Context, e event.Event) error
	Start(ctx context.Context, bus EventSubscriber) error
}
