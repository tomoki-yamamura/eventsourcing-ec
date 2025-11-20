package command

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/aggregate"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/command"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/event"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/repository"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/errors"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/command/input"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/presenter"
)

type SubmitCartCommandInterface interface {
	Execute(ctx context.Context, input *input.SubmitCartInput, out presenter.CommandResultPresenter) error
}

type SubmitCartCommand struct {
	tx         repository.Transaction
	eventStore repository.EventStore
	outboxRepo repository.OutboxRepository
}

func NewSubmitCartCommand(tx repository.Transaction, eventStore repository.EventStore, outboxRepo repository.OutboxRepository) SubmitCartCommandInterface {
	return &SubmitCartCommand{
		tx:         tx,
		eventStore: eventStore,
		outboxRepo: outboxRepo,
	}
}

func (s *SubmitCartCommand) Execute(ctx context.Context, input *input.SubmitCartInput, out presenter.CommandResultPresenter) error {
	maxRetries := 3
	var err error
	var aggregateID string
	var version int
	var events []event.Event

	for attempt := range maxRetries {
		err = s.tx.RWTx(ctx, func(ctx context.Context) error {
			cartID, err := uuid.Parse(input.CartID)
			if err != nil {
				return err
			}

			loadedEvents, err := s.eventStore.LoadEvents(ctx, cartID)
			if err != nil && !errors.IsCode(err, errors.NotFound) {
				return err
			}

			cart := aggregate.NewCartAggregate()
			if len(loadedEvents) > 0 {
				if err := cart.Hydration(loadedEvents); err != nil {
					return err
				}
			}

			cmd := command.SubmitCartCommand{
				CartID: cartID,
			}

			err = cart.ExecuteSubmitCartCommand(cmd)
			if err != nil {
				return err
			}

			if err := s.eventStore.SaveEvents(ctx, cart.GetAggregateID(), cart.GetUncommittedEvents()); err != nil {
				return err
			}
			events := cart.GetUncommittedEvents()
			if len(events) > 0 {
				if err := s.outboxRepo.SaveEvents(ctx, cart.GetAggregateID(), events); err != nil {
					return err
				}
			}

			aggregateID = cart.GetAggregateID().String()
			version = cart.GetVersion()
			events = cart.GetUncommittedEvents()

			cart.MarkEventsAsCommitted()

			return nil
		})
		if err != nil {
			if errors.IsCode(err, errors.OptimisticLock) && attempt < maxRetries-1 {
				waitTime := time.Duration(attempt+1) * 10 * time.Millisecond
				time.Sleep(waitTime)
				continue
			}
			break
		}
		break
	}
	if err != nil {
		return out.PresentError(ctx, err)
	}

	return out.PresentSuccess(ctx, aggregateID, version, events)
}
