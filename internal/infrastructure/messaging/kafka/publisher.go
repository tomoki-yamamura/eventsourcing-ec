package kafka

import (
	"context"

	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/event"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/gateway"
)

type EventPublisher struct {
	producer    *Producer
	topicRouter TopicRouter
}


func NewEventPublisher(producer *Producer, router TopicRouter) gateway.EventPublisher {
	return &EventPublisher{
		producer:    producer,
		topicRouter: router,
	}
}

func (ep *EventPublisher) Publish(ctx context.Context, events ...event.Event) error {
	for _, evt := range events {
		// Determine topic based on aggregate type
		// For now, we'll use "Cart" as default aggregate type
		topic := ep.topicRouter.TopicFor(evt.GetEventType(), "Cart")
		
		message := &Message{
			ID:          evt.GetEventID(),
			Type:        evt.GetEventType(),
			Data:        evt,
			AggregateID: evt.GetAggregateID(),
			Version:     evt.GetVersion(),
		}

		if err := ep.producer.PublishMessage(topic, evt.GetAggregateID().String(), message); err != nil {
			return err
		}
	}

	return nil
}