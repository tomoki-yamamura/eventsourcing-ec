package gateway

import "context"

type ProjectorService interface {
	Start(ctx context.Context) error
	Close() error
}
