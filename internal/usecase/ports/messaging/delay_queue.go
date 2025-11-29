package messaging

import (
	"context"
	"time"
)

type DelayQueue interface {
	PublishDelayedMessage(topic, key string, message *Message, delay time.Duration) error
	Start(ctx context.Context) error
}