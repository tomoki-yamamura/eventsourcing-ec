package messaging

import "context"

type OutboxPublisher interface {
	Start(ctx context.Context) error
}
