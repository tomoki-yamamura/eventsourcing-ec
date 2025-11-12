package command

import (
	"github.com/google/uuid"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/value"
)

type AddTodoCommand struct {
	AggregateID uuid.UUID
	UserID      value.UserID
	TodoText    value.TodoText
}
