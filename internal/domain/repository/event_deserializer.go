package repository

import (
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/event"
)

type EventDeserializer interface {
	Deserialize(eventType string, eventData []byte) (event.Event, error)
}
