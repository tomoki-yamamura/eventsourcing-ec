package deserializer

import (
	"encoding/json"

	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/event"
)

type TodoAddedEventDeserializer struct{}

func NewTodoAddedEventDeserializer() eventDeserializer {
	return &TodoAddedEventDeserializer{}
}

func (d *TodoAddedEventDeserializer) EventType() string {
	return "TodoAddedEvent"
}

func (d *TodoAddedEventDeserializer) Deserialize(eventData []byte) (event.Event, error) {
	var evt event.TodoAddedEvent
	if err := json.Unmarshal(eventData, &evt); err != nil {
		return nil, err
	}
	return evt, nil
}
