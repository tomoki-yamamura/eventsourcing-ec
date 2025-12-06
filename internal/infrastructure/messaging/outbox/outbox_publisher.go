package outbox

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/event"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/repository"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/messaging"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/messaging/dto"
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
	var pendingEvents []event.OutboxEvent
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

		if err := op.handlerSingleEvent(ctx, outboxEvent); err != nil {
			continue
		}

		publishedEventIDs = append(publishedEventIDs, outboxEvent.EventID)
	}

	if err := op.tx.RWTx(ctx, func(ctx context.Context) error {
		return op.outboxRepo.MarkAsPublished(ctx, publishedEventIDs)
	}); err != nil {
		log.Printf("Failed to mark events as published: %v", err)
		return err
	}

	return nil
}

func (op *OutboxPublisher) publishMessage(topic, key string, message *dto.Message) error {
	return op.messageProducer.PublishMessage(topic, key, message)
}

func (op *OutboxPublisher) handlerSingleEvent(ctx context.Context, outboxEvent event.OutboxEvent) error {
	message := &dto.Message{
		ID:          outboxEvent.EventID,
		Type:        outboxEvent.EventType,
		Data:        json.RawMessage(outboxEvent.EventData),
		AggregateID: outboxEvent.AggregateID,
		Version:     outboxEvent.Version,
	}

	topic := op.topicRouter.TopicFor(outboxEvent.EventType, outboxEvent.AggregateType)

	if err := op.publishMessage(topic, outboxEvent.AggregateType, message); err != nil {
		log.Printf("Failed to publish event %s: %v", outboxEvent.EventID, err)
		if err := op.handlePublishError(ctx, outboxEvent.EventID, err); err != nil {
			return err
		}
		return err
	}
	log.Printf("Successfully published event %s to topic %s", outboxEvent.EventID, topic)
	return nil
}

func (op *OutboxPublisher) handlePublishError(
	ctx context.Context,
	eventID uuid.UUID,
	pubErr error,
) error {
	return op.tx.RWTx(ctx, func(ctx context.Context) error {
		if err := op.outboxRepo.MarkAsFailed(ctx, eventID, pubErr.Error()); err != nil {
			log.Printf("Failed to mark event %s as failed: %v", eventID, err)
		}

		if err := op.outboxRepo.IncrementRetryCount(ctx, eventID); err != nil {
			log.Printf("Failed to increment retry count for event %s: %v", eventID, err)
		}

		return nil
	})
}
