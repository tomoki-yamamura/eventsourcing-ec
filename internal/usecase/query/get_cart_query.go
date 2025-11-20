package query

import (
	"context"
	"encoding/json"

	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/presenter"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/readmodelstore"
)

type GetCartQueryInterface interface {
	Query(ctx context.Context, aggregateID string, out presenter.QueryResultPresenter) error
}

type GetCartQuery struct {
	cartStore readmodelstore.CartStore
}

func NewGetCartQuery(cartStore readmodelstore.CartStore) GetCartQueryInterface {
	return &GetCartQuery{
		cartStore: cartStore,
	}
}

func (q *GetCartQuery) Query(ctx context.Context, aggregateID string, out presenter.QueryResultPresenter) error {
	cartView, err := q.cartStore.Get(ctx, aggregateID)
	if err != nil {
		return out.PresentError(ctx, err)
	}

	jsonData, err := json.Marshal(cartView)
	if err != nil {
		return out.PresentError(ctx, err)
	}

	return out.PresentSuccess(ctx, jsonData)
}
