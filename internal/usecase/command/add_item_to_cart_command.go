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
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/gateway"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/presenter"
)

type CartAddItemCommandInterface interface {
	Execute(ctx context.Context, input *input.AddItemToCartInput, out presenter.CommandResultPresenter) error
}

type CartAddItemCommand struct {
	tx         repository.Transaction
	eventStore repository.EventStore
	eventBus   gateway.EventPublisher
}

func NewCartAddItemCommand(tx repository.Transaction, eventStore repository.EventStore, eventBus gateway.EventPublisher) CartAddItemCommandInterface {
	return &CartAddItemCommand{
		tx:         tx,
		eventStore: eventStore,
		eventBus:   eventBus,
	}
}

func (u *CartAddItemCommand) Execute(ctx context.Context, input *input.AddItemToCartInput, out presenter.CommandResultPresenter) error {
	maxRetries := 3
	var err error
	var aggregateID string
	var version int
	var events []event.Event

	for attempt := range maxRetries {
		err = u.tx.RWTx(ctx, func(ctx context.Context) error {
			cartUUID, err := uuid.Parse(input.CartID)
			if err != nil {
				return err
			}

			itemUUID, err := uuid.Parse(input.ItemID)
			if err != nil {
				return err
			}

			loadedEvents, err := u.eventStore.LoadEvents(ctx, cartUUID, "Cart")
			if err != nil && !errors.IsCode(err, errors.NotFound) {
				return err
			}

			cart := aggregate.NewCartAggregate()
			if len(loadedEvents) > 0 {
				if err := cart.Hydration(loadedEvents); err != nil {
					return err
				}
			}

			cmd := command.AddItemToCartCommand{
				CartID:   cartUUID,
				ItemID:   itemUUID,
				Quantity: input.Quantity,
				Price:    input.Price,
			}

			if err := cart.ExecuteAddItemToCartCommand(cmd); err != nil {
				return err
			}

			if err := u.eventStore.SaveEvents(ctx, cart.GetAggregateID(), "Cart", cart.GetUncommittedEvents()); err != nil {
				return err
			}

			aggregateID = cart.GetAggregateID().String()
			version = cart.GetVersion()
			events = cart.GetUncommittedEvents()

			evs := cart.GetUncommittedEvents()
			u.tx.AfterCommit(func() error {
				return u.eventBus.Publish(context.Background(), evs...)
			})

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