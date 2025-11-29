package event

import (
	"time"

	"github.com/google/uuid"
)

type TenantCartAbandonedPolicyUpdatedEvent struct {
	AggregateID      uuid.UUID
	Title            string
	AbandonedMinutes int
	QuietTimeFrom    time.Time
	QuietTimeTo      time.Time
	EventID          uuid.UUID
	Timestamp        time.Time
	Version          int
}

func NewTenantCartAbandonedPolicyUpdatedEvent(aggregateID uuid.UUID, version int, title string, abandonedMinutes int, quietTimeFrom time.Time, quietTimeTo time.Time) *TenantCartAbandonedPolicyUpdatedEvent {
	return &TenantCartAbandonedPolicyUpdatedEvent{
		AggregateID:      aggregateID,
		Title:            title,
		AbandonedMinutes: abandonedMinutes,
		QuietTimeFrom:    quietTimeFrom,
		QuietTimeTo:      quietTimeTo,
		EventID:          uuid.New(),
		Timestamp:        time.Now(),
		Version:          version,
	}
}

func (e TenantCartAbandonedPolicyUpdatedEvent) GetAggregateID() uuid.UUID {
	return e.AggregateID
}

func (e TenantCartAbandonedPolicyUpdatedEvent) GetEventID() uuid.UUID {
	return e.EventID
}

func (e TenantCartAbandonedPolicyUpdatedEvent) GetTimestamp() time.Time {
	return e.Timestamp
}

func (e TenantCartAbandonedPolicyUpdatedEvent) GetVersion() int {
	return e.Version
}

func (e TenantCartAbandonedPolicyUpdatedEvent) GetEventType() string {
	return "TenantCartAbandonedPolicyUpdatedEvent"
}

func (e TenantCartAbandonedPolicyUpdatedEvent) GetAggregateType() string {
	return "TenantCartAbandonedPolicy"
}

func (e *TenantCartAbandonedPolicyUpdatedEvent) GetTitle() string {
	return e.Title
}

func (e *TenantCartAbandonedPolicyUpdatedEvent) GetAbandonedMinutes() int {
	return e.AbandonedMinutes
}

func (e *TenantCartAbandonedPolicyUpdatedEvent) GetQuietTimeFrom() time.Time {
	return e.QuietTimeFrom
}

func (e *TenantCartAbandonedPolicyUpdatedEvent) GetQuietTimeTo() time.Time {
	return e.QuietTimeTo
}