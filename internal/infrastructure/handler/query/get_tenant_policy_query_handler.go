package query

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/presenter"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/view"
	queryUseCase "github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/query"
)

type GetTenantPolicyQueryHandler struct {
	getTenantPolicyQuery queryUseCase.GetTenantPolicyQueryInterface
}

func NewGetTenantPolicyQueryHandler(getTenantPolicyQuery queryUseCase.GetTenantPolicyQueryInterface) *GetTenantPolicyQueryHandler {
	return &GetTenantPolicyQueryHandler{
		getTenantPolicyQuery: getTenantPolicyQuery,
	}
}

func (h *GetTenantPolicyQueryHandler) GetTenantPolicy(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	tenantID := vars["aggregate_id"]

	httpView := view.NewHTTPQueryResultView(w)
	queryPresenter := presenter.NewQueryResultPresenterImpl(httpView)

	if err := h.getTenantPolicyQuery.Query(req.Context(), tenantID, queryPresenter); err != nil {
		queryPresenter.PresentError(req.Context(), err)
	}
}