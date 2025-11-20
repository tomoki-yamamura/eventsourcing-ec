package presenter

import (
	"context"

	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/presenter"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/view"
)

type QueryResultPresenterImpl struct {
	view view.QueryResultView
}

func NewQueryResultPresenterImpl(view view.QueryResultView) presenter.QueryResultPresenter {
	return &QueryResultPresenterImpl{
		view: view,
	}
}

func (p *QueryResultPresenterImpl) PresentSuccess(ctx context.Context, data []byte) error {
	return p.view.Success(data)
}

func (p *QueryResultPresenterImpl) PresentError(ctx context.Context, err error) error {
	return p.view.Error(err)
}