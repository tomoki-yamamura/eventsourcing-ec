package deserializer

import (
	"encoding/json"

	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/event"
)

type cartSubmittedEventDeserializer struct{}

func NewCartSubmittedEventDeserializer() eventDeserializer {
	return &cartSubmittedEventDeserializer{}
}

func (d *cartSubmittedEventDeserializer) EventType() string {
	return "CartSubmittedEvent"
}

func (d *cartSubmittedEventDeserializer) Deserialize(eventData []byte) (event.Event, error) {
	var evt event.CartSubmittedEvent
	if err := json.Unmarshal(eventData, &evt); err != nil {
		return nil, err
	}
	return &evt, nil
}
