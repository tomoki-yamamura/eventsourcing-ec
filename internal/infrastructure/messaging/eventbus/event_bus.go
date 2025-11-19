package eventbus

import (
	"context"

	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/event"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/messaging/kafka"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/gateway"
)

type EventBus struct {
	publisher  *kafka.EventPublisher
	subscriber *kafka.EventSubscriber
}

func NewEventBus(publisher *kafka.EventPublisher, subscriber *kafka.EventSubscriber) gateway.EventBus {
	return &EventBus{
		publisher:  publisher,
		subscriber: subscriber,
	}
}

func (eb *EventBus) Publish(ctx context.Context, events ...event.Event) error {
	return eb.publisher.Publish(ctx, events...)
}

func (eb *EventBus) Subscribe(handler func(context.Context, event.Event) error) {
	eb.subscriber.Subscribe(handler)
}

func (eb *EventBus) Start(ctx context.Context) error {
	return eb.subscriber.Start(ctx)
}
