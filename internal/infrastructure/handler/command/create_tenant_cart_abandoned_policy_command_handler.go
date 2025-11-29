package command

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/presenter"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/view"
	commandUseCase "github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/command"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/command/input"
)

type CreateTenantCartAbandonedPolicyCommandHandler struct {
	createTenantCartAbandonedPolicyCommand commandUseCase.CreateTenantCartAbandonedPolicyCommandInterface
}

func NewCreateTenantCartAbandonedPolicyCommandHandler(createTenantCartAbandonedPolicyCommand commandUseCase.CreateTenantCartAbandonedPolicyCommandInterface) *CreateTenantCartAbandonedPolicyCommandHandler {
	return &CreateTenantCartAbandonedPolicyCommandHandler{
		createTenantCartAbandonedPolicyCommand: createTenantCartAbandonedPolicyCommand,
	}
}

func (h *CreateTenantCartAbandonedPolicyCommandHandler) CreateTenantCartAbandonedPolicy(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	aggregateID := vars["aggregate_id"]

	var requestBody input.CreateTenantCartAbandonedPolicyInput
	if err := json.NewDecoder(req.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	requestBody.TenantID = aggregateID

	httpView := view.NewHTTPCommandResultView(w)
	commandPresenter := presenter.NewCommandResultPresenterImpl(httpView)

	if err := h.createTenantCartAbandonedPolicyCommand.Execute(req.Context(), &requestBody, commandPresenter); err != nil {
		commandPresenter.PresentError(req.Context(), err)
	}
}
