package messaging

type MessageProducer interface {
	PublishMessage(topic, key string, message *Message) error
}