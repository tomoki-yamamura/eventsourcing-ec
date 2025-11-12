package command

import "github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/value"

type CreateTodoListCommand struct {
	UserID value.UserID
}
