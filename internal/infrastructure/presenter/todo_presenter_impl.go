package presenter

import (
	"context"
	"net/http"
	"time"

	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/presenter/viewmodel"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/presenter"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/query/output"
)

type HTTPTodoListPresenter struct {
	view TodoListView
}

func NewHTTPTodoListPresenter(view TodoListView) presenter.TodoListPresenter {
	return &HTTPTodoListPresenter{view: view}
}

func (p *HTTPTodoListPresenter) Present(ctx context.Context, out *output.GetTodoListOutput) error {
	var items []viewmodel.TodoItem
	for _, it := range out.Items {
		items = append(items, viewmodel.TodoItem{Text: it.Text})
	}
	vm := &viewmodel.TodoListVM{
		AggregateID: out.AggregateID,
		UserID:      out.UserID,
		Items:       items,
		UpdatedAt:   out.UpdatedAt.Format(time.RFC3339),
	}
	return p.view.Render(ctx, vm, http.StatusOK, nil)
}

func (p *HTTPTodoListPresenter) PresentNotFound(ctx context.Context, err error) error {
	return p.view.Render(ctx, nil, http.StatusNotFound, err)
}

func (p *HTTPTodoListPresenter) PresentError(ctx context.Context, err error) error {
	return p.view.Render(ctx, nil, http.StatusInternalServerError, err)
}
