package kafka

import "github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/messaging"

type StaticMapRouter struct {
	aggregateTopicMap map[string]string
}

func NewStaticTopicRouter() messaging.TopicRouter {
	return &StaticMapRouter{
		aggregateTopicMap: map[string]string{
			"Cart":                      "ec.cart-events",
			"TenantCartAbandonedPolicy": "ec.cart-events",
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
