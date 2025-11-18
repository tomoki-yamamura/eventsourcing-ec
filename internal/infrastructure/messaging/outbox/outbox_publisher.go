package outbox

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/repository"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/messaging/kafka"
)

type OutboxPublisher struct {
	tx              repository.Transaction
	outboxRepo      repository.OutboxRepository
	kafkaProducer   *kafka.Producer
	topicRouter     kafka.TopicRouter
	pollingInterval time.Duration
	batchSize       int
	maxRetries      int
}

func NewOutboxPublisher(
	tx repository.Transaction,
	outboxRepo repository.OutboxRepository,
	kafkaProducer *kafka.Producer,
	topicRouter kafka.TopicRouter,
	pollingInterval time.Duration,
	batchSize int,
	maxRetries int,
) *OutboxPublisher {
	return &OutboxPublisher{
		tx:              tx,
		outboxRepo:      outboxRepo,
		kafkaProducer:   kafkaProducer,
		topicRouter:     topicRouter,
		pollingInterval: pollingInterval,
		batchSize:       batchSize,
		maxRetries:      maxRetries,
	}
}

func (op *OutboxPublisher) Start(ctx context.Context) error {
	log.Printf("Starting outbox publisher with polling interval: %v", op.pollingInterval)
	
	ticker := time.NewTicker(op.pollingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Outbox publisher stopped")
			return nil
		case <-ticker.C:
			if err := op.publishPendingEvents(ctx); err != nil {
				log.Printf("Error publishing pending events: %v", err)
			}
		}
	}
}

func (op *OutboxPublisher) publishPendingEvents(ctx context.Context) error {
	var pendingEvents []repository.OutboxEvent
	err := op.tx.RWTx(ctx, func(ctx context.Context) error {
		var err error
		pendingEvents, err = op.outboxRepo.GetPendingEvents(ctx, op.batchSize)
		return err
	})
	if err != nil {
		return err
	}

	if len(pendingEvents) == 0 {
		return nil
	}

	log.Printf("Found %d pending events to publish", len(pendingEvents))

	var publishedEventIDs []uuid.UUID
	
	for _, outboxEvent := range pendingEvents {
		// Skip events that exceeded max retries
		if outboxEvent.RetryCount >= op.maxRetries {
			log.Printf("Skipping event %s - max retries (%d) exceeded", outboxEvent.EventID, op.maxRetries)
			continue
		}

		message := &kafka.Message{
			ID:          outboxEvent.EventID,
			Type:        outboxEvent.EventType,
			Data:        json.RawMessage(outboxEvent.EventData),
			AggregateID: outboxEvent.AggregateID,
			Version:     outboxEvent.Version,
		}

		topic := op.topicRouter.TopicFor(outboxEvent.EventType, outboxEvent.AggregateType)
		if err := op.publishToKafka(topic, outboxEvent.AggregateID.String(), message); err != nil {
			log.Printf("Failed to publish event %s: %v", outboxEvent.EventID, err)
			
			// Mark failed and increment retry count in separate transactions
			_ = op.tx.RWTx(ctx, func(ctx context.Context) error {
				if err := op.outboxRepo.MarkAsFailed(ctx, outboxEvent.EventID, err.Error()); err != nil {
					log.Printf("Failed to mark event as failed: %v", err)
				}
				if err := op.outboxRepo.IncrementRetryCount(ctx, outboxEvent.EventID); err != nil {
					log.Printf("Failed to increment retry count: %v", err)
				}
				return nil
			})
			continue
		}

		publishedEventIDs = append(publishedEventIDs, outboxEvent.EventID)
		log.Printf("Successfully published event %s to topic %s", outboxEvent.EventID, topic)
	}

	// Mark all successfully published events in a single transaction
	if len(publishedEventIDs) > 0 {
		err := op.tx.RWTx(ctx, func(ctx context.Context) error {
			return op.outboxRepo.MarkAsPublished(ctx, publishedEventIDs)
		})
		if err != nil {
			log.Printf("Failed to mark events as published: %v", err)
			return err
		}
		log.Printf("Marked %d events as published", len(publishedEventIDs))
	}

	return nil
}

func (op *OutboxPublisher) publishToKafka(topic, key string, message *kafka.Message) error {
	return op.kafkaProducer.PublishMessage(topic, key, message)
}