package readmodelstore

import (
	"context"

	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/readmodelstore/dto"
)

type TodoListStore interface {
	Get(ctx context.Context, aggregateID string) (*dto.TodoListViewDTO, error)
	Upsert(ctx context.Context, aggregateID string, view *dto.TodoListViewDTO) error
}
