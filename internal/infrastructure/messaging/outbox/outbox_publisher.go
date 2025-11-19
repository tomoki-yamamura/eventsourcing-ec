package outbox

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/repository"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/messaging"
)

const (
	DefaultPollingInterval = 200 * time.Millisecond
	DefaultBatchSize       = 100
	DefaultMaxRetries      = 3
)

type OutboxPublisher struct {
	tx              repository.Transaction
	outboxRepo      repository.OutboxRepository
	messageProducer messaging.MessageProducer
	topicRouter     messaging.TopicRouter
}

func NewOutboxPublisher(
	tx repository.Transaction,
	outboxRepo repository.OutboxRepository,
	messageProducer messaging.MessageProducer,
	topicRouter messaging.TopicRouter,
) messaging.OutboxPublisher {
	return &OutboxPublisher{
		tx:              tx,
		outboxRepo:      outboxRepo,
		messageProducer: messageProducer,
		topicRouter:     topicRouter,
	}
}

func (op *OutboxPublisher) Start(ctx context.Context) error {
	log.Printf("Starting outbox publisher with polling interval: %v", DefaultPollingInterval)

	ticker := time.NewTicker(DefaultPollingInterval)
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
		pendingEvents, err = op.outboxRepo.GetPendingEvents(ctx, DefaultBatchSize)
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
		if outboxEvent.RetryCount >= DefaultMaxRetries {
			log.Printf("Skipping event %s - max retries (%d) exceeded", outboxEvent.EventID, DefaultMaxRetries)
			continue
		}

		message := &messaging.Message{
			ID:          outboxEvent.EventID,
			Type:        outboxEvent.EventType,
			Data:        json.RawMessage(outboxEvent.EventData),
			AggregateID: outboxEvent.AggregateID,
			Version:     outboxEvent.Version,
		}

		topic := op.topicRouter.TopicFor(outboxEvent.EventType, outboxEvent.AggregateType)
		if err := op.publishMessage(topic, outboxEvent.AggregateID.String(), message); err != nil {
			log.Printf("Failed to publish event %s: %v", outboxEvent.EventID, err)

			if txErr := op.tx.RWTx(ctx, func(ctx context.Context) error {
				if markErr := op.outboxRepo.MarkAsFailed(ctx, outboxEvent.EventID, err.Error()); markErr != nil {
					log.Printf("Failed to mark event as failed: %v", markErr)
				}
				if retryErr := op.outboxRepo.IncrementRetryCount(ctx, outboxEvent.EventID); retryErr != nil {
					log.Printf("Failed to increment retry count: %v", retryErr)
				}
				return nil
			}); txErr != nil {
				log.Printf("Transaction failed while handling publish error for event %s: %v", outboxEvent.EventID, txErr)
			}
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

func (op *OutboxPublisher) publishMessage(topic, key string, message *messaging.Message) error {
	return op.messageProducer.PublishMessage(topic, key, message)
}
