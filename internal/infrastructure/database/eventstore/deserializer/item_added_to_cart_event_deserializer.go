package deserializer

import (
	"encoding/json"

	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/event"
)

type itemAddedToCartEventDeserializer struct{}

func NewItemAddedToCartEventDeserializer() eventDeserializer {
	return &itemAddedToCartEventDeserializer{}
}

func (d *itemAddedToCartEventDeserializer) EventType() string {
	return "ItemAddedToCartEvent"
}

func (d *itemAddedToCartEventDeserializer) Deserialize(eventData []byte) (event.Event, error) {
	var evt event.ItemAddedToCartEvent
	if err := json.Unmarshal(eventData, &evt); err != nil {
		return nil, err
	}
	return &evt, nil
}