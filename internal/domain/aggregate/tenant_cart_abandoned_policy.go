package aggregate

import (
	"time"

	"github.com/google/uuid"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/command"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/event"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/errors"
)

type TenantCartAbandonedPolicyAggregate struct {
	tenantID             uuid.UUID
	title                string
	cartAbandonedMinutes int
	quietTimeFrom        time.Time
	quietTimeTo          time.Time
	version              int
	uncommitted          []event.Event
}

func NewTenantCartAbandonedPolicyAggregate() *TenantCartAbandonedPolicyAggregate {
	return &TenantCartAbandonedPolicyAggregate{
		version:     -1,
		uncommitted: make([]event.Event, 0),
	}
}

func (a *TenantCartAbandonedPolicyAggregate) GetAggregateID() uuid.UUID { return a.tenantID }
func (a *TenantCartAbandonedPolicyAggregate) GetVersion() int           { return a.version }
func (a *TenantCartAbandonedPolicyAggregate) GetTitle() string          { return a.title }

func (a *TenantCartAbandonedPolicyAggregate) GetUncommittedEvents() []event.Event {
	return a.uncommitted
}

func (a *TenantCartAbandonedPolicyAggregate) MarkEventsAsCommitted() {
	a.uncommitted = nil
}

func (a *TenantCartAbandonedPolicyAggregate) Hydration(events []event.Event) error {
	for _, ev := range events {
		a.apply(ev)
	}
	return nil
}

func (a *TenantCartAbandonedPolicyAggregate) apply(ev event.Event) {
	switch e := ev.(type) {
	case *event.TenantCartAbandonedPolicyCreatedEvent:
		a.tenantID = e.GetAggregateID()
		a.title = e.GetTitle()
		a.cartAbandonedMinutes = e.GetAbandonedMinutes()
		a.quietTimeFrom = e.GetQuietTimeFrom()
		a.quietTimeTo = e.GetQuietTimeTo()
		a.version = e.GetVersion()
	case *event.TenantCartAbandonedPolicyUpdatedEvent:
		a.title = e.GetTitle()
		a.cartAbandonedMinutes = e.GetAbandonedMinutes()
		a.quietTimeFrom = e.GetQuietTimeFrom()
		a.quietTimeTo = e.GetQuietTimeTo()
		a.version = e.GetVersion()
	default:
	}
}

func (a *TenantCartAbandonedPolicyAggregate) CartAbandonedDelay() time.Duration {
	return time.Duration(a.cartAbandonedMinutes) * time.Minute
}

func (a *TenantCartAbandonedPolicyAggregate) IsWithinQuietTime(now time.Time) (bool, error) {
	if a.quietTimeFrom.IsZero() || a.quietTimeTo.IsZero() {
		return false, nil
	}

	nowUTC := now.UTC()
	fromUTC := a.quietTimeFrom.UTC()
	toUTC := a.quietTimeTo.UTC()

	currentMinutes := nowUTC.Hour()*60 + nowUTC.Minute()
	fromMinutes := fromUTC.Hour()*60 + fromUTC.Minute()
	toMinutes := toUTC.Hour()*60 + toUTC.Minute()

	if fromMinutes < toMinutes {
		return currentMinutes >= fromMinutes && currentMinutes < toMinutes, nil
	} else if fromMinutes > toMinutes {
		return currentMinutes >= fromMinutes || currentMinutes < toMinutes, nil
	} else {
		return false, nil
	}
}

func (a *TenantCartAbandonedPolicyAggregate) ExecuteCreateTenantCartAbandonedPolicyCommand(cmd command.CreateTenantCartAbandonedPolicyCommand) error {
	if a.version != -1 {
		return errors.UnpermittedOp.New("tenant policy already exists")
	}

	ev := event.NewTenantCartAbandonedPolicyCreatedEvent(
		cmd.TenantID,
		1,
		cmd.Title,
		cmd.AbandonedMinutes,
		cmd.QuietTimeFrom,
		cmd.QuietTimeTo,
	)
	a.apply(ev)
	a.uncommitted = append(a.uncommitted, ev)

	return nil
}

func (a *TenantCartAbandonedPolicyAggregate) ExecuteUpdateTenantCartAbandonedPolicyCommand(cmd command.UpdateTenantCartAbandonedPolicyCommand) error {
	if a.version == -1 {
		return errors.UnpermittedOp.New("tenant policy not created")
	}

	if a.title == cmd.Title &&
		a.cartAbandonedMinutes == cmd.AbandonedMinutes &&
		a.quietTimeFrom.Equal(cmd.QuietTimeFrom) &&
		a.quietTimeTo.Equal(cmd.QuietTimeTo) {
		return nil
	}

	ev := event.NewTenantCartAbandonedPolicyUpdatedEvent(
		a.tenantID,
		a.version+1,
		cmd.Title,
		cmd.AbandonedMinutes,
		cmd.QuietTimeFrom,
		cmd.QuietTimeTo,
	)
	a.apply(ev)
	a.uncommitted = append(a.uncommitted, ev)

	return nil
}
