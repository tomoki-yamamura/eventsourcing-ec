package readmodelstore

import (
	"context"

	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/readmodelstore/dto"
)

type TenantPolicyStore interface {
	Get(ctx context.Context, tenantID string) (*dto.TenantPolicyViewDTO, error)
	Upsert(ctx context.Context, tenantID string, view *dto.TenantPolicyViewDTO) error
}
