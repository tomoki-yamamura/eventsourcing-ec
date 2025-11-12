package command

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/handler/request"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/presenter"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/view"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/command"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/command/input"
)

type TodoAddItemCommandHandler struct {
	addCommand command.TodoAddItemCommandInterface
}

func NewTodoAddItemCommandHandler(addCommand command.TodoAddItemCommandInterface) *TodoAddItemCommandHandler {
	return &TodoAddItemCommandHandler{
		addCommand: addCommand,
	}
}

func (h *TodoAddItemCommandHandler) AddTodo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	aggregateID := vars["aggregate_id"]

	var req request.AddTodoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	usecaseInput := &input.AddTodoInput{
		AggregateID: aggregateID,
		UserID:      req.UserID,
		Todo:        req.Text,
	}

	view := view.NewHTTPCommandResultView(w)
	presenter := presenter.NewCommandResultPresenterImpl(view)

	err := h.addCommand.Execute(r.Context(), usecaseInput, presenter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
