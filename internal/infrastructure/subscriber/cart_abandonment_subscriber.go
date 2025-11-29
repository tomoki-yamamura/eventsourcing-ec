package subscriber

import (
	"context"
	"log"

	"github.com/google/uuid"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/aggregate"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/event"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/repository"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/errors"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/gateway"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/messaging"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/messaging/dto"
)

type CartAbandonmentSubscriber struct {
	tx         repository.Transaction
	eventStore repository.EventStore
	delayQueue messaging.DelayQueue
	seen       map[string]struct{}
}

func NewCartAbandonmentSubscriber(
	tx repository.Transaction,
	eventStore repository.EventStore,
	delayQueue messaging.DelayQueue,
) *CartAbandonmentSubscriber {
	return &CartAbandonmentSubscriber{
		tx:         tx,
		eventStore: eventStore,
		delayQueue: delayQueue,
		seen:       make(map[string]struct{}),
	}
}

func (s *CartAbandonmentSubscriber) Handle(ctx context.Context, e event.Event) error {
	eventID := e.GetEventID().String()
	if _, ok := s.seen[eventID]; ok {
		return nil
	}
	s.seen[eventID] = struct{}{}

	if itemAdded, ok := e.(*event.ItemAddedToCartEvent); ok {
		log.Printf("Processing ItemAddedToCartEvent for cart abandonment: %s", eventID)
		return s.scheduleCartAbandonmentCheck(ctx, itemAdded)
	}

	return nil
}

func (s *CartAbandonmentSubscriber) Start(ctx context.Context, bus gateway.EventSubscriber) error {
	bus.Subscribe(s.Handle)
	return nil
}

func (s *CartAbandonmentSubscriber) scheduleCartAbandonmentCheck(ctx context.Context, itemAdded *event.ItemAddedToCartEvent) error {
	cartID := itemAdded.GetAggregateID()
	tenantID := itemAdded.GetTenantID()

	policy, err := s.loadTenantPolicy(ctx, tenantID)
	if err != nil {
		if errors.IsCode(err, errors.NotFound) {
			log.Printf("No tenant policy found for tenant %s, skipping cart abandonment check", tenantID)
			return nil
		}
		return err
	}

	delay := policy.CartAbandonedDelay()

	delayedMessage := &dto.Message{
		ID:   uuid.New(),
		Type: "CheckCartAbandonmentCommand",
		Data: map[string]any{
			"cart_id":             cartID.String(),
			"tenant_id":           tenantID.String(),
			"item_added_event_id": itemAdded.GetEventID().String(),
			"item_added_at":       itemAdded.GetTimestamp().Unix(),
			"delay_minutes":       delay.Minutes(),
		},
		AggregateID: cartID,
		Version:     itemAdded.GetVersion(),
	}

	log.Printf("Scheduling cart abandonment check for cart %s in %v", cartID, delay)
	return s.delayQueue.PublishDelayedMessage("cart-abandonment-check", cartID.String(), delayedMessage, delay)
}


func (s *CartAbandonmentSubscriber) loadTenantPolicy(ctx context.Context, tenantID uuid.UUID) (*aggregate.TenantCartAbandonedPolicyAggregate, error) {
	events, err := s.eventStore.LoadEvents(ctx, tenantID)
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