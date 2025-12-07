package event

import (
	"time"

	"github.com/google/uuid"
)

type TenantCartAbandonedPolicyCreatedEvent struct {
	AggregateID      uuid.UUID
	Title            string
	AbandonedMinutes int
	QuietTimeFrom    time.Time
	QuietTimeTo      time.Time
	EventID          uuid.UUID
	Timestamp        time.Time
	Version          int
}

func NewTenantCartAbandonedPolicyCreatedEvent(aggregateID uuid.UUID, version int, title string, abandonedMinutes int, quietTimeFrom time.Time, quietTimeTo time.Time) *TenantCartAbandonedPolicyCreatedEvent {
	return &TenantCartAbandonedPolicyCreatedEvent{
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

func (e TenantCartAbandonedPolicyCreatedEvent) GetAggregateID() uuid.UUID {
	return e.AggregateID
}

func (e TenantCartAbandonedPolicyCreatedEvent) GetEventID() uuid.UUID {
	return e.EventID
}

func (e TenantCartAbandonedPolicyCreatedEvent) GetTimestamp() time.Time {
	return e.Timestamp
}

func (e TenantCartAbandonedPolicyCreatedEvent) GetVersion() int {
	return e.Version
}

func (e TenantCartAbandonedPolicyCreatedEvent) GetEventType() string {
	return "TenantCartAbandonedPolicyCreatedEvent"
}

func (e TenantCartAbandonedPolicyCreatedEvent) GetAggregateType() string {
	return "TenantCartAbandonedPolicy"
}

func (e *TenantCartAbandonedPolicyCreatedEvent) GetTitle() string {
	return e.Title
}

func (e *TenantCartAbandonedPolicyCreatedEvent) GetAbandonedMinutes() int {
	return e.AbandonedMinutes
}

func (e *TenantCartAbandonedPolicyCreatedEvent) GetQuietTimeFrom() time.Time {
	return e.QuietTimeFrom
}

func (e *TenantCartAbandonedPolicyCreatedEvent) GetQuietTimeTo() time.Time {
	return e.QuietTimeTo
}
