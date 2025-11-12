package command

import (
	"encoding/json"
	"net/http"

	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/handler/request"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/presenter"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/view"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/command"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/command/input"
)

type TodoListCreateCommandHandler struct {
	createCommand command.TodoListCreateCommandInterface
}

func NewTodoListCreateCommandHandler(createCommand command.TodoListCreateCommandInterface) *TodoListCreateCommandHandler {
	return &TodoListCreateCommandHandler{
		createCommand: createCommand,
	}
}

func (h *TodoListCreateCommandHandler) CreateTodoList(w http.ResponseWriter, r *http.Request) {
	var req request.CreateTodoListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	usecaseInput := &input.CreateTodoListInput{
		UserID: req.UserID,
	}

	view := view.NewHTTPCommandResultView(w)
	presenter := presenter.NewCommandResultPresenterImpl(view)

	err := h.createCommand.Execute(r.Context(), usecaseInput, presenter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
