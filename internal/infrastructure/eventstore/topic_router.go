package eventstore

type TopicRouter interface {
	TopicFor(eventType string, aggregateType string) string
}

type StaticMapRouter struct {
	aggregateTopicMap map[string]string
}

func NewStaticMapRouter() *StaticMapRouter {
	return &StaticMapRouter{
		aggregateTopicMap: map[string]string{
			"Cart":      "ec.cart-events",
		},
	}
}

func (r *StaticMapRouter) TopicFor(eventType, aggregateType string) string {
	if topic, exists := r.aggregateTopicMap[aggregateType]; exists {
		return topic
	}
	return "ec.misc-events"
}

func (r *StaticMapRouter) AddMapping(aggregateType, topic string) {
	r.aggregateTopicMap[aggregateType] = topic
}