package messaging

import (
	"context"

	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/messaging/dto"
)

type ConsumerGroup interface {
	AddHandler(handler MessageHandler)
	Start(ctx context.Context) error
	Close() error
}

type MessageHandler func(context.Context, *dto.Message) error
