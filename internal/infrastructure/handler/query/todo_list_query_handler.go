package query

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/presenter"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/view"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/query"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/query/input"
)

type TodoListQueryHandler struct {
	todoListQueryUsecase query.TodoListQueryInterface
}

func NewTodoListQueryHandler(todoListQueryUsecase query.TodoListQueryInterface) *TodoListQueryHandler {
	return &TodoListQueryHandler{
		todoListQueryUsecase: todoListQueryUsecase,
	}
}

func (h *TodoListQueryHandler) Query(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	aggregateID := vars["aggregate_id"]

	if aggregateID == "" {
		http.Error(w, "aggregate_id is required", http.StatusBadRequest)
		return
	}

	in := &input.GetTodoListInput{
		AggregateID: aggregateID,
	}

	v := view.NewHTTPTodoListView(w)
	p := presenter.NewHTTPTodoListPresenter(v)

	if err := h.todoListQueryUsecase.Execute(r.Context(), in, p); err != nil {
		return
	}
}
