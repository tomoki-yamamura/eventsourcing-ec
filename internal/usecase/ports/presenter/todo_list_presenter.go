package presenter

import (
	"context"

	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/query/output"
)

type TodoListPresenter interface {
	Present(ctx context.Context, output *output.GetTodoListOutput) error
	PresentNotFound(ctx context.Context, err error) error
	PresentError(ctx context.Context, err error) error
}
