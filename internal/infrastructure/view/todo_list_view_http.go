package view

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/presenter"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/presenter/viewmodel"
)

type HTTPTodoListView struct {
	writer http.ResponseWriter
}

func NewHTTPTodoListView(w http.ResponseWriter) presenter.TodoListView {
	return &HTTPTodoListView{writer: w}
}

func (v *HTTPTodoListView) Render(ctx context.Context, vm *viewmodel.TodoListVM, status int, err error) error {
	v.writer.Header().Set("Content-Type", "application/json")
	v.writer.WriteHeader(status)

	if err != nil {
		errorResponse := map[string]any{
			"status":  "error",
			"message": err.Error(),
		}
		return json.NewEncoder(v.writer).Encode(errorResponse)
	}

	return json.NewEncoder(v.writer).Encode(vm)
}
