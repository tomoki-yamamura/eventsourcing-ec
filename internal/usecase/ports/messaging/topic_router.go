package messaging

type TopicRouter interface {
	TopicFor(eventType, aggregateType string) string
}