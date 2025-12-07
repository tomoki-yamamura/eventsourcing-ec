package query

import (
	"context"
	"encoding/json"

	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/presenter"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/readmodelstore"
)

type GetTenantPolicyQueryInterface interface {
	Query(ctx context.Context, tenantID string, out presenter.QueryResultPresenter) error
}

type GetTenantPolicyQueryImpl struct {
	tenantPolicyStore readmodelstore.TenantPolicyStore
}

func NewGetTenantPolicyQuery(tenantPolicyStore readmodelstore.TenantPolicyStore) GetTenantPolicyQueryInterface {
	return &GetTenantPolicyQueryImpl{
		tenantPolicyStore: tenantPolicyStore,
	}
}

func (g *GetTenantPolicyQueryImpl) Query(ctx context.Context, tenantID string, out presenter.QueryResultPresenter) error {
	tenantPolicy, err := g.tenantPolicyStore.Get(ctx, tenantID)
	if err != nil {
		return out.PresentError(ctx, err)
	}

	jsonData, err := json.Marshal(tenantPolicy)
	if err != nil {
		return out.PresentError(ctx, err)
	}

	return out.PresentSuccess(ctx, jsonData)
}
