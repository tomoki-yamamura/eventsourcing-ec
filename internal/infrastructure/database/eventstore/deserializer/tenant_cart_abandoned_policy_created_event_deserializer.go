package deserializer

import (
	"encoding/json"

	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/event"
)

type tenantCartAbandonedPolicyCreatedEventDeserializer struct{}

func NewTenantCartAbandonedPolicyCreatedEventDeserializer() eventDeserializer {
	return &tenantCartAbandonedPolicyCreatedEventDeserializer{}
}

func (d *tenantCartAbandonedPolicyCreatedEventDeserializer) EventType() string {
	return "TenantCartAbandonedPolicyCreatedEvent"
}

func (d *tenantCartAbandonedPolicyCreatedEventDeserializer) Deserialize(eventData []byte) (event.Event, error) {
	var evt event.TenantCartAbandonedPolicyCreatedEvent
	if err := json.Unmarshal(eventData, &evt); err != nil {
		return nil, err
	}
	return &evt, nil
}
