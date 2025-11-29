package service

import (
	"context"
	"encoding/json"
	"log"

	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/repository"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/messaging/kafka"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/subscriber"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/messaging"
)

type CartAbandonmentService struct {
	deserializer           repository.EventDeserializer
	cartAbandonmentSubscriber *subscriber.CartAbandonmentSubscriber
	kafkaConsumer          *kafka.ConsumerGroup
	delayQueue             messaging.DelayQueue
}

func NewCartAbandonmentService(
	deserializer repository.EventDeserializer,
	cartAbandonmentSubscriber *subscriber.CartAbandonmentSubscriber,
	kafkaConsumer *kafka.ConsumerGroup,
	delayQueue messaging.DelayQueue,
) *CartAbandonmentService {
	service := &CartAbandonmentService{
		deserializer:              deserializer,
		cartAbandonmentSubscriber: cartAbandonmentSubscriber,
		kafkaConsumer:             kafkaConsumer,
		delayQueue:                delayQueue,
	}

	// Add abandonment handler to consumer
	kafkaConsumer.AddHandler(service.handleMessage)

	return service
}

func (s *CartAbandonmentService) handleMessage(ctx context.Context, msg *messaging.Message) error {
	// Marshal message data for deserializer
	eventData, err := json.Marshal(msg.Data)
	if err != nil {
		return err
	}

	// Deserialize event
	event, err := s.deserializer.Deserialize(msg.Type, eventData)
	if err != nil {
		return err
	}

	// Handle with cart abandonment subscriber
	return s.cartAbandonmentSubscriber.Handle(ctx, event)
}

func (s *CartAbandonmentService) Start(ctx context.Context) error {
	log.Println("Starting Cart Abandonment Service...")
	
	// Start delay queue processing
	go func() {
		if err := s.delayQueue.Start(ctx); err != nil {
			log.Printf("Delay queue stopped: %v", err)
		}
	}()
	
	// Start Kafka consumer
	return s.kafkaConsumer.Start(ctx)
}
