package deserializer

import (
	"encoding/json"

	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/event"
)

type tenantCartAbandonedPolicyUpdatedEventDeserializer struct{}

func NewTenantCartAbandonedPolicyUpdatedEventDeserializer() eventDeserializer {
	return &tenantCartAbandonedPolicyUpdatedEventDeserializer{}
}

func (d *tenantCartAbandonedPolicyUpdatedEventDeserializer) EventType() string {
	return "TenantCartAbandonedPolicyUpdatedEvent"
}

func (d *tenantCartAbandonedPolicyUpdatedEventDeserializer) Deserialize(eventData []byte) (event.Event, error) {
	var evt event.TenantCartAbandonedPolicyUpdatedEvent
	if err := json.Unmarshal(eventData, &evt); err != nil {
		return nil, err
	}
	return &evt, nil
}
