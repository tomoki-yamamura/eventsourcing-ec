package deserializer

import (
	"encoding/json"

	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/event"
)

type cartPurchasedEventDeserializer struct{}

func NewCartPurchasedEventDeserializer() eventDeserializer {
	return &cartPurchasedEventDeserializer{}
}

func (d *cartPurchasedEventDeserializer) EventType() string {
	return "CartPurchasedEvent"
}

func (d *cartPurchasedEventDeserializer) Deserialize(eventData []byte) (event.Event, error) {
	var evt event.CartPurchasedEvent
	if err := json.Unmarshal(eventData, &evt); err != nil {
		return nil, err
	}
	return &evt, nil
}