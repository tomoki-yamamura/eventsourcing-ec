package query

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/presenter"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/view"
	queryUseCase "github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/query"
)

type GetCartQueryHandler struct {
	getCartQuery queryUseCase.GetCartQueryInterface
}

func NewGetCartQueryHandler(getCartQuery queryUseCase.GetCartQueryInterface) *GetCartQueryHandler {
	return &GetCartQueryHandler{
		getCartQuery: getCartQuery,
	}
}

func (h *GetCartQueryHandler) GetCart(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	aggregateID := vars["aggregate_id"]

	httpView := view.NewHTTPQueryResultView(w)
	queryPresenter := presenter.NewQueryResultPresenterImpl(httpView)

	if err := h.getCartQuery.Query(req.Context(), aggregateID, queryPresenter); err != nil {
		queryPresenter.PresentError(req.Context(), err)
	}
}