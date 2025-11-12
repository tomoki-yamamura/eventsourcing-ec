package todo

import (
	"context"

	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/event"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/errors"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/gateway"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/readmodelstore"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/readmodelstore/dto"
)

type TodoProjectorImpl struct {
	viewRepo readmodelstore.TodoListStore
	seen     map[string]struct{}
}

func NewTodoProjector(viewRepo readmodelstore.TodoListStore) gateway.Projector {
	return &TodoProjectorImpl{
		viewRepo: viewRepo,
		seen:     make(map[string]struct{}),
	}
}

func (p *TodoProjectorImpl) Handle(ctx context.Context, e event.Event) error {
	eventID := e.GetEventID().String()
	if _, ok := p.seen[eventID]; ok {
		return nil
	}
	p.seen[eventID] = struct{}{}

	switch e.(type) {
	case event.TodoListCreatedEvent, event.TodoAddedEvent:
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

func (p *TodoProjectorImpl) Start(ctx context.Context, bus gateway.EventSubscriber) error {
	bus.Subscribe(p.Handle)
	return nil
}

func (p *TodoProjectorImpl) applyToView(view *dto.TodoListViewDTO, e event.Event) *dto.TodoListViewDTO {
	switch evt := e.(type) {
	case event.TodoListCreatedEvent:
		return &dto.TodoListViewDTO{
			AggregateID: evt.GetAggregateID().String(),
			UserID:      evt.UserID.String(),
			Items:       []dto.TodoItemViewDTO{},
			Version:     evt.GetVersion(),
			UpdatedAt:   evt.GetTimestamp(),
		}
	case event.TodoAddedEvent:
		newItems := make([]dto.TodoItemViewDTO, len(view.Items))
		copy(newItems, view.Items)
		newItems = append(newItems, dto.TodoItemViewDTO{
			Text: evt.TodoText.String(),
		})

		return &dto.TodoListViewDTO{
			AggregateID: view.AggregateID,
			UserID:      view.UserID,
			Items:       newItems,
			Version:     evt.GetVersion(),
			UpdatedAt:   evt.GetTimestamp(),
		}
	}

	return view
}
