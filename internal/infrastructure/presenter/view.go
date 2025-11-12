package presenter

import (
	"context"

	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/presenter/viewmodel"
)

type CommandView interface {
	Render(ctx context.Context, vm *viewmodel.CommandResultViewModel, status int, err error) error
}

type TodoListView interface {
	Render(ctx context.Context, vm *viewmodel.TodoListVM, status int, err error) error
}
