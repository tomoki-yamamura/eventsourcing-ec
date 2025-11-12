package deserializer

import (
	"encoding/json"

	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/event"
)

type TodoListCreatedEventDeserializer struct{}

func NewTodoListCreatedEventDeserializer() eventDeserializer {
	return &TodoListCreatedEventDeserializer{}
}

func (d *TodoListCreatedEventDeserializer) EventType() string {
	return "TodoListCreatedEvent"
}

func (d *TodoListCreatedEventDeserializer) Deserialize(eventData []byte) (event.Event, error) {
	var evt event.TodoListCreatedEvent
	if err := json.Unmarshal(eventData, &evt); err != nil {
		return nil, err
	}
	return evt, nil
}
