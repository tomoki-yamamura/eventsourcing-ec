package cart

import (
	"context"

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
	
	if _, ok := p.seen[eventID]; ok {
		return nil
	}
	p.seen[eventID] = struct{}{}

	switch e.(type) {
	case *event.CartCreatedEvent, *event.ItemAddedToCartEvent, *event.CartSubmittedEvent:
		aggID := e.GetAggregateID().String()

		current, err := p.viewRepo.Get(ctx, aggID)
		if err != nil {
			if errors.IsCode(err, errors.NotFound) {
				current = nil
			} else {
				return err
			}
		} else {
		}

		updated := p.applyToView(current, e)
		if updated != nil {
			err := p.viewRepo.Upsert(ctx, aggID, updated)
			if err != nil {
			}
			return err
		} else {
		}
	default:
		return nil
	}

	return nil
}

func (p *CartProjectorImpl) Start(ctx context.Context, bus gateway.EventSubscriber) error {
	bus.Subscribe(p.Handle)
	return nil
}

func (p *CartProjectorImpl) applyToView(view *dto.CartViewDTO, e event.Event) *dto.CartViewDTO {
	switch evt := e.(type) {
	case *event.CartCreatedEvent:
		return &dto.CartViewDTO{
			ID:          evt.GetAggregateID().String(),
			UserID:      evt.GetUserID().String(),
			TenantID:    evt.GetTenantID().String(),
			Status:      "OPEN",
			TotalAmount: 0.0,
			ItemCount:   0,
			Items:       []dto.CartItemViewDTO{},
			CreatedAt:   evt.GetTimestamp(),
			UpdatedAt:   evt.GetTimestamp(),
			Version:     evt.GetVersion(),
		}
	case *event.ItemAddedToCartEvent:
		if view == nil {
			return nil
		}

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
			TenantID:    view.TenantID,
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
			TenantID:    view.TenantID,
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
