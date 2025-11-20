package presenter

import "context"

type QueryResultPresenter interface {
	PresentSuccess(ctx context.Context, data []byte) error
	PresentError(ctx context.Context, err error) error
}