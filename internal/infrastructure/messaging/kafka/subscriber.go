package kafka

import (
	"context"
	"encoding/json"

	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/event"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/repository"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/messaging"
)

type EventSubscriber struct {
	consumerGroup *ConsumerGroup
	deserializer  repository.EventDeserializer
}

func NewEventSubscriber(brokers []string, groupID string, topics []string, deserializer repository.EventDeserializer) (*EventSubscriber, error) {
	consumerGroup, err := NewConsumerGroup(brokers, groupID, topics, deserializer)
	if err != nil {
		return nil, err
	}

	return &EventSubscriber{
		consumerGroup: consumerGroup,
		deserializer:  deserializer,
	}, nil
}

func (es *EventSubscriber) Subscribe(handler func(context.Context, event.Event) error) {
	messageHandler := func(ctx context.Context, message *messaging.Message) error {
		// Convert Message back to domain Event
		dataBytes, err := json.Marshal(message.Data)
		if err != nil {
			return err
		}
		
		domainEvent, err := es.deserializer.Deserialize(message.Type, dataBytes)
		if err != nil {
			return err
		}

		return handler(ctx, domainEvent)
	}

	es.consumerGroup.AddHandler(messageHandler)
}

func (es *EventSubscriber) Start(ctx context.Context) error {
	return es.consumerGroup.Start(ctx)
}

func (es *EventSubscriber) Close() error {
	return es.consumerGroup.Close()
}