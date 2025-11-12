package aggregate

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/command"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/entity"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/event"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/value"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/errors"
)

var ErrTooManyTodos = errors.UnpermittedOp.New("cannot add more than 3 todos per day")

type TodoListAggregate struct {
	aggregateID       uuid.UUID
	userID            value.UserID
	items             []*entity.TodoItem
	version           int
	uncommittedEvents []event.Event
}

func NewTodoListAggregate() *TodoListAggregate {
	return &TodoListAggregate{
		items:             make([]*entity.TodoItem, 0),
		uncommittedEvents: make([]event.Event, 0),
	}
}

func (a *TodoListAggregate) GetAggregateID() uuid.UUID {
	return a.aggregateID
}

func (a *TodoListAggregate) GetUserID() value.UserID {
	return a.userID
}

func (a *TodoListAggregate) GetItems() []*entity.TodoItem {
	return a.items
}

func (a *TodoListAggregate) GetVersion() int {
	return a.version
}

func (a *TodoListAggregate) GetUncommittedEvents() []event.Event {
	return a.uncommittedEvents
}

func (a *TodoListAggregate) MarkEventsAsCommitted() {
	a.uncommittedEvents = make([]event.Event, 0)
}

func (a *TodoListAggregate) Hydration(events []event.Event) error {
	for _, evt := range events {
		if err := a.applyEvent(evt, false); err != nil {
			return fmt.Errorf("failed to apply event: %w", err)
		}
	}
	return nil
}

func (a *TodoListAggregate) ExecuteCreateTodoListCommand(cmd command.CreateTodoListCommand) error {
	evt := event.TodoListCreatedEvent{
		AggregateID: uuid.New(),
		UserID:      cmd.UserID,
		EventID:     uuid.New(),
		Timestamp:   time.Now(),
		Version:     a.version + 1,
	}

	return a.applyEvent(evt, true)
}

func (a *TodoListAggregate) ExecuteAddTodoCommand(cmd command.AddTodoCommand) error {
	// set a limit of only three items per day for Todo.
	if len(a.items) >= 3 {
		return ErrTooManyTodos
	}

	evt := event.TodoAddedEvent{
		AggregateID: cmd.AggregateID,
		UserID:      cmd.UserID,
		TodoText:    cmd.TodoText,
		EventID:     uuid.New(),
		Timestamp:   time.Now(),
		Version:     a.version + 1,
	}

	return a.applyEvent(evt, true)
}

func (a *TodoListAggregate) applyEvent(evt event.Event, isNew bool) error {
	switch e := evt.(type) {
	case event.TodoListCreatedEvent:
		a.onTodoListCreated(e)
	case event.TodoAddedEvent:
		a.onTodoAdded(e)
	default:
		return fmt.Errorf("unknown event type: %T", evt)
	}

	if isNew {
		a.uncommittedEvents = append(a.uncommittedEvents, evt)
	}

	a.version = evt.GetVersion()
	return nil
}

func (a *TodoListAggregate) onTodoListCreated(evt event.TodoListCreatedEvent) {
	a.aggregateID = evt.AggregateID
	a.userID = evt.UserID
}

func (a *TodoListAggregate) onTodoAdded(evt event.TodoAddedEvent) {
	todoItem := entity.NewTodoItem(evt.TodoText)
	a.items = append(a.items, todoItem)
}
