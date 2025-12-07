package messaging

import (
	"context"
	"time"

	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/messaging/dto"
)

type DelayQueue interface {
	PublishDelayedMessage(topic, key string, message *dto.Message, delay time.Duration) error
	Start(ctx context.Context) error
}
