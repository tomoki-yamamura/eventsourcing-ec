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

type UpdateTenantCartAbandonedPolicyCommandHandler struct {
	updateTenantCartAbandonedPolicyCommand commandUseCase.UpdateTenantCartAbandonedPolicyCommandInterface
}

func NewUpdateTenantCartAbandonedPolicyCommandHandler(updateTenantCartAbandonedPolicyCommand commandUseCase.UpdateTenantCartAbandonedPolicyCommandInterface) *UpdateTenantCartAbandonedPolicyCommandHandler {
	return &UpdateTenantCartAbandonedPolicyCommandHandler{
		updateTenantCartAbandonedPolicyCommand: updateTenantCartAbandonedPolicyCommand,
	}
}

func (h *UpdateTenantCartAbandonedPolicyCommandHandler) UpdateTenantCartAbandonedPolicy(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	aggregateID := vars["aggregate_id"]

	var requestBody input.UpdateTenantCartAbandonedPolicyInput
	if err := json.NewDecoder(req.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	requestBody.TenantID = aggregateID

	httpView := view.NewHTTPCommandResultView(w)
	commandPresenter := presenter.NewCommandResultPresenterImpl(httpView)

	if err := h.updateTenantCartAbandonedPolicyCommand.Execute(req.Context(), &requestBody, commandPresenter); err != nil {
		commandPresenter.PresentError(req.Context(), err)
	}
}