package gateway

import "context"

type CartAbandonmentService interface {
	Start(ctx context.Context) error
}