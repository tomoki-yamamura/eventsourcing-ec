package worker

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/aggregate"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/event"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/repository"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/errors"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/messaging"
)

type CartAbandonmentWorker struct {
	tx                 repository.Transaction
	outboxRepo         repository.OutboxRepository
	eventStore         repository.EventStore
	delayQueueProducer messaging.MessageProducer
}

func NewCartAbandonmentWorker(
	tx repository.Transaction,
	outboxRepo repository.OutboxRepository,
	eventStore repository.EventStore,
	delayQueueProducer messaging.MessageProducer,
) *CartAbandonmentWorker {
	return &CartAbandonmentWorker{
		tx:                 tx,
		outboxRepo:         outboxRepo,
		eventStore:         eventStore,
		delayQueueProducer: delayQueueProducer,
	}
}

func (w *CartAbandonmentWorker) Start(ctx context.Context) error {
	log.Println("Starting cart abandonment worker")

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Cart abandonment worker stopped")
			return nil
		case <-ticker.C:
			if err := w.processItemAddedEvents(ctx); err != nil {
				log.Printf("Error processing item added events: %v", err)
			}
		}
	}
}

func (w *CartAbandonmentWorker) processItemAddedEvents(ctx context.Context) error {
	var pendingEvents []event.OutboxEvent
	err := w.tx.RWTx(ctx, func(ctx context.Context) error {
		var err error
		pendingEvents, err = w.outboxRepo.GetPendingEventsByType(ctx, "ItemAddedToCartEvent", 100)
		return err
	})
	if err != nil {
		return err
	}

	if len(pendingEvents) == 0 {
		return nil
	}

	log.Printf("Found %d item added events to schedule for cart abandonment", len(pendingEvents))

	var processedEventIDs []uuid.UUID

	for _, outboxEvent := range pendingEvents {
		if err := w.scheduleCartAbandonmentCheck(ctx, outboxEvent); err != nil {
			log.Printf("Failed to schedule cart abandonment check for event %s: %v", outboxEvent.EventID, err)
			continue
		}

		processedEventIDs = append(processedEventIDs, outboxEvent.EventID)
		log.Printf("Scheduled cart abandonment check for event %s", outboxEvent.EventID)
	}

	if len(processedEventIDs) > 0 {
		err := w.tx.RWTx(ctx, func(ctx context.Context) error {
			return w.outboxRepo.MarkAsProcessed(ctx, processedEventIDs, "cart_abandonment_worker")
		})
		if err != nil {
			log.Printf("Failed to mark events as processed: %v", err)
			return err
		}
		log.Printf("Marked %d events as processed by cart abandonment worker", len(processedEventIDs))
	}

	return nil
}

func (w *CartAbandonmentWorker) scheduleCartAbandonmentCheck(ctx context.Context, outboxEvent event.OutboxEvent) error {
	var itemAddedEvent event.ItemAddedToCartEvent
	if err := json.Unmarshal([]byte(outboxEvent.EventData), &itemAddedEvent); err != nil {
		return err
	}

	// Get tenant ID from cart
	cartID := itemAddedEvent.GetAggregateID()
	tenantID, err := w.getTenantIDFromCart(ctx, cartID)
	if err != nil {
		return err
	}

	// Load tenant policy and get delay minutes
	policy, err := w.loadTenantPolicy(ctx, tenantID)
	if err != nil {
		if errors.IsCode(err, errors.NotFound) {
			log.Printf("No tenant policy found for tenant %s, skipping cart abandonment check", tenantID)
			return nil
		}
		return err
	}

	delay := policy.CartAbandonedDelay()

	delayedMessage := &messaging.Message{
		ID:   uuid.New(),
		Type: "CheckCartAbandonmentCommand",
		Data: map[string]interface{}{
			"cart_id":             cartID.String(),
			"tenant_id":           tenantID.String(),
			"item_added_event_id": itemAddedEvent.GetEventID().String(),
			"item_added_at":       itemAddedEvent.GetTimestamp().Unix(),
			"delay_minutes":       delay.Minutes(),
		},
		AggregateID: cartID,
		Version:     itemAddedEvent.GetVersion(),
	}

	return w.delayQueueProducer.PublishDelayedMessage("cart-abandonment-check", cartID.String(), delayedMessage, delay)
}

func (w *CartAbandonmentWorker) getTenantIDFromCart(ctx context.Context, cartID uuid.UUID) (uuid.UUID, error) {
	events, err := w.eventStore.LoadEvents(ctx, cartID)
	if err != nil {
		return uuid.Nil, err
	}

	for _, ev := range events {
		if cartCreated, ok := ev.(*event.CartCreatedEvent); ok {
			return w.getTenantIDFromUserID(ctx, cartCreated.GetUserID())
		}
	}

	return uuid.Nil, errors.NotFound.New("cart creation event not found")
}

func (w *CartAbandonmentWorker) getTenantIDFromUserID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	// TODO: Implement actual user -> tenant mapping
	// For now, return mock tenant ID
	return uuid.New(), nil
}

func (w *CartAbandonmentWorker) loadTenantPolicy(ctx context.Context, tenantID uuid.UUID) (*aggregate.TenantCartAbandonedPolicyAggregate, error) {
	events, err := w.eventStore.LoadEvents(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	policy := aggregate.NewTenantCartAbandonedPolicyAggregate()
	if len(events) > 0 {
		if err := policy.Hydration(events); err != nil {
			return nil, err
		}
	}

	if policy.GetVersion() == -1 {
		return nil, errors.NotFound.New("tenant policy not found")
	}

	return policy, nil
}