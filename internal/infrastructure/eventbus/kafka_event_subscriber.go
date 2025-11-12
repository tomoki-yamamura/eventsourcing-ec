package eventbus

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/event"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/repository"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/gateway"
)

type StoredEvent struct {
	EventID     uuid.UUID `json:"event_id"`
	EventType   string    `json:"event_type"`
	EventData   any       `json:"event_data"`
	Version     int       `json:"version"`
	AggregateID uuid.UUID `json:"aggregate_id"`
	CreatedAt   time.Time `json:"created_at"`
}

type KafkaEventSubscriber struct {
	brokers      []string
	topics       []string
	consumerGroup string
	deserializer repository.EventDeserializer
	handlers     []func(context.Context, event.Event) error
	mu           sync.RWMutex
}

func NewKafkaEventSubscriber(brokers []string, topics []string, consumerGroup string, deserializer repository.EventDeserializer) gateway.EventSubscriber {
	return &KafkaEventSubscriber{
		brokers:       brokers,
		topics:        topics,
		consumerGroup: consumerGroup,
		deserializer:  deserializer,
		handlers:      make([]func(context.Context, event.Event) error, 0),
	}
}

func (k *KafkaEventSubscriber) Subscribe(handler func(context.Context, event.Event) error) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.handlers = append(k.handlers, handler)
}

func (k *KafkaEventSubscriber) Start(ctx context.Context) error {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Version = sarama.V2_8_0_0

	consumerGroup, err := sarama.NewConsumerGroup(k.brokers, k.consumerGroup, config)
	if err != nil {
		return err
	}

	consumer := &Consumer{
		subscriber: k,
	}

	go func() {
		for {
			if err := consumerGroup.Consume(ctx, k.topics, consumer); err != nil {
				log.Printf("Error from consumer: %v", err)
				return
			}

			if ctx.Err() != nil {
				return
			}
		}
	}()

	go func() {
		for err := range consumerGroup.Errors() {
			log.Printf("Error from consumer group: %v", err)
		}
	}()

	<-ctx.Done()
	return consumerGroup.Close()
}

func (k *KafkaEventSubscriber) handleEvent(ctx context.Context, message *sarama.ConsumerMessage) error {
	var storedEvent StoredEvent
	if err := json.Unmarshal(message.Value, &storedEvent); err != nil {
		log.Printf("Failed to unmarshal event: %v", err)
		return nil
	}

	domainEvent, err := k.deserializer.Deserialize(storedEvent.EventType, storedEvent.EventData)
	if err != nil {
		log.Printf("Failed to deserialize event: %v", err)
		return nil
	}

	k.mu.RLock()
	handlers := make([]func(context.Context, event.Event) error, len(k.handlers))
	copy(handlers, k.handlers)
	k.mu.RUnlock()

	for _, handler := range handlers {
		if err := handler(ctx, domainEvent); err != nil {
			log.Printf("Error handling event: %v", err)
		}
	}

	return nil
}

type Consumer struct {
	subscriber *KafkaEventSubscriber
}

func (c *Consumer) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (c *Consumer) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}
			err := c.subscriber.handleEvent(session.Context(), message)
			if err != nil {
				log.Printf("Error handling message: %v", err)
			}
			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}