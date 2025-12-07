package tenant

import (
	"context"

	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/event"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/errors"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/gateway"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/readmodelstore"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/readmodelstore/dto"
)

type TenantPolicyProjectorImpl struct {
	viewRepo readmodelstore.TenantPolicyStore
	seen     map[string]struct{}
}

func NewTenantPolicyProjector(viewRepo readmodelstore.TenantPolicyStore) gateway.Projector {
	return &TenantPolicyProjectorImpl{
		viewRepo: viewRepo,
		seen:     make(map[string]struct{}),
	}
}

func (p *TenantPolicyProjectorImpl) Handle(ctx context.Context, e event.Event) error {
	eventID := e.GetEventID().String()
	if _, ok := p.seen[eventID]; ok {
		return nil
	}
	p.seen[eventID] = struct{}{}

	switch e.(type) {
	case *event.TenantCartAbandonedPolicyCreatedEvent, *event.TenantCartAbandonedPolicyUpdatedEvent:
		aggID := e.GetAggregateID().String()

		current, err := p.viewRepo.Get(ctx, aggID)
		if err != nil {
			if errors.IsCode(err, errors.NotFound) {
				current = nil
			} else {
				return err
			}
		}

		updated := p.applyToView(current, e)
		if updated != nil {
			return p.viewRepo.Upsert(ctx, aggID, updated)
		}
	default:
		return nil
	}

	return nil
}

func (p *TenantPolicyProjectorImpl) Start(ctx context.Context, bus gateway.EventSubscriber) error {
	bus.Subscribe(p.Handle)
	return nil
}

func (p *TenantPolicyProjectorImpl) applyToView(view *dto.TenantPolicyViewDTO, e event.Event) *dto.TenantPolicyViewDTO {
	switch evt := e.(type) {
	case *event.TenantCartAbandonedPolicyCreatedEvent:
		return &dto.TenantPolicyViewDTO{
			ID:               evt.GetAggregateID().String(),
			Title:            evt.Title,
			AbandonedMinutes: evt.AbandonedMinutes,
			QuietTimeFrom:    evt.QuietTimeFrom,
			QuietTimeTo:      evt.QuietTimeTo,
			CreatedAt:        evt.GetTimestamp(),
			UpdatedAt:        evt.GetTimestamp(),
			Version:          evt.GetVersion(),
		}
	case *event.TenantCartAbandonedPolicyUpdatedEvent:
		if view == nil {
			return nil
		}

		return &dto.TenantPolicyViewDTO{
			ID:               view.ID,
			Title:            evt.Title,
			AbandonedMinutes: evt.AbandonedMinutes,
			QuietTimeFrom:    evt.QuietTimeFrom,
			QuietTimeTo:      evt.QuietTimeTo,
			CreatedAt:        view.CreatedAt,
			UpdatedAt:        evt.GetTimestamp(),
			Version:          evt.GetVersion(),
		}
	}

	return view
}
