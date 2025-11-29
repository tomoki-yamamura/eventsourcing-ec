package messaging

import (
	"context"

	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/event"
)

type Subscriber interface {
	Handle(ctx context.Context, e event.Event) error
}