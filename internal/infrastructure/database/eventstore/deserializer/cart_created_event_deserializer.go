package deserializer

import (
	"encoding/json"

	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/event"
)

type cartCreatedEventDeserializer struct{}

func NewCartCreatedEventDeserializer() eventDeserializer {
	return &cartCreatedEventDeserializer{}
}

func (d *cartCreatedEventDeserializer) EventType() string {
	return "CartCreatedEvent"
}

func (d *cartCreatedEventDeserializer) Deserialize(eventData []byte) (event.Event, error) {
	var evt event.CartCreatedEvent
	if err := json.Unmarshal(eventData, &evt); err != nil {
		return nil, err
	}
	return &evt, nil
}