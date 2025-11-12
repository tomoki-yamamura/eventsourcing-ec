package presenter

import (
	"context"
	"time"

	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/event"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/errors"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/presenter/viewmodel"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/presenter"
)

type CommandResultPresenterImpl struct {
	view CommandView
}

func NewCommandResultPresenterImpl(view CommandView) presenter.CommandResultPresenter {
	return &CommandResultPresenterImpl{
		view: view,
	}
}

func (p *CommandResultPresenterImpl) PresentSuccess(ctx context.Context, aggregateID string, version int, events []event.Event) error {
	eventVMs := make([]viewmodel.EventViewModel, 0, len(events))
	for _, ev := range events {
		eventVMs = append(eventVMs, viewmodel.EventViewModel{
			Type:       ev.GetEventType(),
			Version:    ev.GetVersion(),
			Data:       ev,
			OccurredAt: ev.GetTimestamp().Format(time.RFC3339),
		})
	}

	vm := &viewmodel.CommandResultViewModel{
		AggregateID: aggregateID,
		Version:     version,
		Events:      eventVMs,
		Status:      "success",
		ExecutedAt:  time.Now().Format(time.RFC3339),
	}

	return p.view.Render(ctx, vm, 200, nil)
}

func (p *CommandResultPresenterImpl) PresentError(ctx context.Context, err error) error {
	vm := &viewmodel.CommandResultViewModel{
		Status:     "error",
		ExecutedAt: time.Now().Format(time.RFC3339),
	}

	statusCode := p.determineStatusCode(err)
	return p.view.Render(ctx, vm, statusCode, err)
}

func (p *CommandResultPresenterImpl) determineStatusCode(err error) int {
	if errors.IsCode(err, errors.InvalidParameter) {
		return 422
	}
	if errors.IsCode(err, errors.NotFound) {
		return 404
	}
	if errors.IsCode(err, errors.OptimisticLock) {
		return 409
	}
	return 500
}
