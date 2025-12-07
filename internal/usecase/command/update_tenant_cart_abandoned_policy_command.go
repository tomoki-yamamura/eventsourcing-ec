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

type UpdateTenantCartAbandonedPolicyCommandInterface interface {
	Execute(ctx context.Context, input *input.UpdateTenantCartAbandonedPolicyInput, out presenter.CommandResultPresenter) error
}

type UpdateTenantCartAbandonedPolicyCommand struct {
	tx         repository.Transaction
	eventStore repository.EventStore
	outboxRepo repository.OutboxRepository
}

func NewUpdateTenantCartAbandonedPolicyCommand(tx repository.Transaction, eventStore repository.EventStore, outboxRepo repository.OutboxRepository) UpdateTenantCartAbandonedPolicyCommandInterface {
	return &UpdateTenantCartAbandonedPolicyCommand{
		tx:         tx,
		eventStore: eventStore,
		outboxRepo: outboxRepo,
	}
}

func (u *UpdateTenantCartAbandonedPolicyCommand) Execute(ctx context.Context, input *input.UpdateTenantCartAbandonedPolicyInput, out presenter.CommandResultPresenter) error {
	maxRetries := 3
	var err error
	var aggregateID string
	var version int
	var events []event.Event

	for attempt := range maxRetries {
		err = u.tx.RWTx(ctx, func(ctx context.Context) error {
			tenantUUID, err := uuid.Parse(input.TenantID)
			if err != nil {
				return err
			}

			loadedEvents, err := u.eventStore.LoadEvents(ctx, tenantUUID)
			if err != nil && !errors.IsCode(err, errors.NotFound) {
				return err
			}

			policy := aggregate.NewTenantCartAbandonedPolicyAggregate()
			if len(loadedEvents) > 0 {
				if err := policy.Hydration(loadedEvents); err != nil {
					return err
				}
			}

			cmd := command.UpdateTenantCartAbandonedPolicyCommand{
				TenantID:         tenantUUID,
				Title:            input.Title,
				AbandonedMinutes: input.AbandonedMinutes,
				QuietTimeFrom:    input.QuietTimeFrom,
				QuietTimeTo:      input.QuietTimeTo,
			}

			if err := policy.ExecuteUpdateTenantCartAbandonedPolicyCommand(cmd); err != nil {
				return err
			}

			if err := u.eventStore.SaveEvents(ctx, policy.GetAggregateID(), policy.GetUncommittedEvents()); err != nil {
				return err
			}

			events := policy.GetUncommittedEvents()
			if len(events) > 0 {
				if err := u.outboxRepo.SaveEvents(ctx, policy.GetAggregateID(), events); err != nil {
					return err
				}
			}

			aggregateID = policy.GetAggregateID().String()
			version = policy.GetVersion()
			events = policy.GetUncommittedEvents()

			policy.MarkEventsAsCommitted()

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
