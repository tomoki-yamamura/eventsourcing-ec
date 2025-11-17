package command

import (
	"context"

	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/aggregate"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/command"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/event"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/repository"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/value"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/command/input"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/presenter"
)

type TodoListCreateCommandInterface interface {
	Execute(ctx context.Context, input *input.CreateTodoListInput, out presenter.CommandResultPresenter) error
}

type TodoListCreateCommand struct {
	tx         repository.Transaction
	eventStore repository.EventStore
}

func NewTodoListCreateCommand(tx repository.Transaction, eventStore repository.EventStore) TodoListCreateCommandInterface {
	return &TodoListCreateCommand{
		tx:         tx,
		eventStore: eventStore,
	}
}

func (u *TodoListCreateCommand) Execute(ctx context.Context, input *input.CreateTodoListInput, out presenter.CommandResultPresenter) error {
	var aggregateID string
	var version int
	var events []event.Event

	err := u.tx.RWTx(ctx, func(ctx context.Context) error {
		userID, err := value.NewUserID(input.UserID)
		if err != nil {
			return err
		}

		cmd := command.CreateTodoListCommand{
			UserID: userID,
		}

		todoList := aggregate.NewTodoListAggregate()
		if err := todoList.ExecuteCreateTodoListCommand(cmd); err != nil {
			return err
		}

		if err := u.eventStore.SaveEvents(ctx, todoList.GetAggregateID(), "TodoList", todoList.GetUncommittedEvents()); err != nil {
			return err
		}

		aggregateID = todoList.GetAggregateID().String()
		version = todoList.GetVersion()
		events = todoList.GetUncommittedEvents()

		todoList.MarkEventsAsCommitted()

		return nil
	})
	if err != nil {
		return out.PresentError(ctx, err)
	}

	return out.PresentSuccess(ctx, aggregateID, version, events)
}
