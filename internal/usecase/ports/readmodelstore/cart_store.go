package readmodelstore

import (
	"context"

	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/readmodelstore/dto"
)

type CartStore interface {
	Get(ctx context.Context, aggregateID string) (*dto.CartViewDTO, error)
	Upsert(ctx context.Context, aggregateID string, view *dto.CartViewDTO) error
}
