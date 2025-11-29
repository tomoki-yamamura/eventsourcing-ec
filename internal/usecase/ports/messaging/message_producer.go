package messaging

import "time"

type MessageProducer interface {
	PublishMessage(topic, key string, message *Message) error
	PublishDelayedMessage(topic, key string, message *Message, delay time.Duration) error
}
