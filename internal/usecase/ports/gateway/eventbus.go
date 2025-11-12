package gateway

import (
	"context"

	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/event"
)

type EventPublisher interface {
	Publish(ctx context.Context, events ...event.Event) error
}

type EventSubscriber interface {
	Subscribe(handler func(context.Context, event.Event) error)
}

type EventBus interface {
	EventPublisher
	EventSubscriber
}
