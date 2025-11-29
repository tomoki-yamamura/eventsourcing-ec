package messaging

import (
	"time"

	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/messaging/dto"
)

type MessageProducer interface {
	PublishMessage(topic, key string, message *dto.Message) error
	PublishDelayedMessage(topic, key string, message *dto.Message, delay time.Duration) error
}
