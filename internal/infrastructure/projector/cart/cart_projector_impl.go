package cart

import (
	"context"
	"log"

	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/event"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/errors"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/gateway"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/readmodelstore"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/readmodelstore/dto"
)

type CartProjectorImpl struct {
	viewRepo readmodelstore.CartStore
	seen     map[string]struct{}
}

func NewCartProjector(viewRepo readmodelstore.CartStore) gateway.Projector {
	return &CartProjectorImpl{
		viewRepo: viewRepo,
		seen:     make(map[string]struct{}),
	}
}

func (p *CartProjectorImpl) Handle(ctx context.Context, e event.Event) error {
	eventID := e.GetEventID().String()
	log.Printf("[CartProjector] Handling event: %s (ID: %s, AggregateID: %s)", e.GetEventType(), eventID, e.GetAggregateID().String())
	
	if _, ok := p.seen[eventID]; ok {
		log.Printf("[CartProjector] Event %s already processed, skipping", eventID)
		return nil
	}
	p.seen[eventID] = struct{}{}

	switch e.(type) {
	case *event.CartCreatedEvent, *event.ItemAddedToCartEvent, *event.CartSubmittedEvent:
		aggID := e.GetAggregateID().String()
		log.Printf("[CartProjector] Processing %s for cart %s", e.GetEventType(), aggID)

		current, err := p.viewRepo.Get(ctx, aggID)
		if err != nil {
			if errors.IsCode(err, errors.NotFound) {
				log.Printf("[CartProjector] Cart %s not found in read model, current=nil", aggID)
				current = nil
			} else {
				log.Printf("[CartProjector] Error getting cart %s: %v", aggID, err)
				return err
			}
		} else {
			log.Printf("[CartProjector] Found existing cart %s, version=%d, items=%d", aggID, current.Version, len(current.Items))
		}

		updated := p.applyToView(current, e)
		if updated != nil {
			log.Printf("[CartProjector] Upserting cart %s, version=%d, items=%d", aggID, updated.Version, len(updated.Items))
			err := p.viewRepo.Upsert(ctx, aggID, updated)
			if err != nil {
				log.Printf("[CartProjector] Error upserting cart %s: %v", aggID, err)
			}
			return err
		} else {
			log.Printf("[CartProjector] applyToView returned nil for %s", e.GetEventType())
		}
	default:
		log.Printf("[CartProjector] Ignoring event type: %s", e.GetEventType())
		return nil
	}

	return nil
}

func (p *CartProjectorImpl) Start(ctx context.Context, bus gateway.EventSubscriber) error {
	bus.Subscribe(p.Handle)
	return nil
}

func (p *CartProjectorImpl) applyToView(view *dto.CartViewDTO, e event.Event) *dto.CartViewDTO {
	log.Printf("[CartProjector] applyToView called for event %s", e.GetEventType())
	switch evt := e.(type) {
	case *event.CartCreatedEvent:
		log.Printf("[CartProjector] Creating new cart view for %s", evt.GetAggregateID().String())
		return &dto.CartViewDTO{
			ID:          evt.GetAggregateID().String(),
			UserID:      evt.GetUserID().String(),
			Status:      "OPEN",
			TotalAmount: 0.0,
			ItemCount:   0,
			Items:       []dto.CartItemViewDTO{},
			CreatedAt:   evt.GetTimestamp(),
			UpdatedAt:   evt.GetTimestamp(),
			Version:     evt.GetVersion(),
		}
	case *event.ItemAddedToCartEvent:
		log.Printf("[CartProjector] Processing ItemAddedToCartEvent for %s", evt.GetAggregateID().String())
		if view == nil {
			log.Printf("[CartProjector] WARNING: view is nil for ItemAddedToCartEvent %s - cart should exist first!", evt.GetAggregateID().String())
			return nil
		}
		log.Printf("[CartProjector] Adding item %s to cart %s", evt.GetItemID().String(), evt.GetAggregateID().String())

		newItems := make([]dto.CartItemViewDTO, len(view.Items))
		copy(newItems, view.Items)

		newItems = append(newItems, dto.CartItemViewDTO{
			ID:     evt.GetItemID().String(),
			CartID: evt.GetAggregateID().String(),
			Name:   evt.GetName(),
			Price:  evt.GetPrice(),
		})

		totalAmount := 0.0
		itemCount := 0
		for _, item := range newItems {
			totalAmount += item.Price
			itemCount++
		}

		return &dto.CartViewDTO{
			ID:          view.ID,
			UserID:      view.UserID,
			Status:      view.Status,
			TotalAmount: totalAmount,
			ItemCount:   itemCount,
			Items:       newItems,
			CreatedAt:   view.CreatedAt,
			UpdatedAt:   evt.GetTimestamp(),
			PurchasedAt: view.PurchasedAt,
			Version:     evt.GetVersion(),
		}
	case *event.CartSubmittedEvent:
		if view == nil {
			return nil
		}

		return &dto.CartViewDTO{
			ID:          view.ID,
			UserID:      view.UserID,
			Status:      "SUBMITTED",
			TotalAmount: view.TotalAmount,
			ItemCount:   view.ItemCount,
			Items:       view.Items,
			CreatedAt:   view.CreatedAt,
			UpdatedAt:   evt.GetTimestamp(),
			PurchasedAt: view.PurchasedAt,
			Version:     evt.GetVersion(),
		}
	}

	return view
}
