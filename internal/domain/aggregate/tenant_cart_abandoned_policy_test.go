package aggregate_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/aggregate"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/command"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/event"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/errors"
)

func TestTenantCartAbandonedPolicyAggregate_ExecuteCreateTenantCartAbandonedPolicyCommand(t *testing.T) {
	tenantID := uuid.New()

	tests := map[string]struct {
		existingPolicy bool
		cmd            command.CreateTenantCartAbandonedPolicyCommand
		wantErr        error
		wantEventsLen  int
		wantVersion    int
	}{
		"should create new policy": {
			existingPolicy: false,
			cmd: command.CreateTenantCartAbandonedPolicyCommand{
				TenantID:         tenantID,
				Title:            "Test Policy",
				AbandonedMinutes: 30,
				QuietTimeFrom:    time.Date(2023, 1, 1, 22, 0, 0, 0, time.UTC),
				QuietTimeTo:      time.Date(2023, 1, 1, 8, 0, 0, 0, time.UTC),
			},
			wantErr:       nil,
			wantEventsLen: 1,
			wantVersion:   1,
		},
		"should return error when policy already exists": {
			existingPolicy: true,
			cmd: command.CreateTenantCartAbandonedPolicyCommand{
				TenantID:         tenantID,
				Title:            "Another Policy",
				AbandonedMinutes: 60,
				QuietTimeFrom:    time.Date(2023, 1, 1, 23, 0, 0, 0, time.UTC),
				QuietTimeTo:      time.Date(2023, 1, 1, 7, 0, 0, 0, time.UTC),
			},
			wantErr:       errors.UnpermittedOp.New("tenant policy already exists"),
			wantEventsLen: 0,
			wantVersion:   1,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			policy := aggregate.NewTenantCartAbandonedPolicyAggregate()
			if tt.existingPolicy {
				existingCmd := command.CreateTenantCartAbandonedPolicyCommand{
					TenantID:         tenantID,
					Title:            "Existing Policy",
					AbandonedMinutes: 45,
					QuietTimeFrom:    time.Date(2023, 1, 1, 21, 0, 0, 0, time.UTC),
					QuietTimeTo:      time.Date(2023, 1, 1, 9, 0, 0, 0, time.UTC),
				}
				policy.ExecuteCreateTenantCartAbandonedPolicyCommand(existingCmd)
				policy.MarkEventsAsCommitted()
			}

			// Act
			err := policy.ExecuteCreateTenantCartAbandonedPolicyCommand(tt.cmd)

			// Assert
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "tenant policy already exists")
			} else {
				assert.NoError(t, err)
			}

			assert.Len(t, policy.GetUncommittedEvents(), tt.wantEventsLen)
			assert.Equal(t, tt.wantVersion, policy.GetVersion())
		})
	}
}

func TestTenantCartAbandonedPolicyAggregate_ExecuteUpdateTenantCartAbandonedPolicyCommand(t *testing.T) {
	tenantID := uuid.New()

	tests := map[string]struct {
		existingPolicy bool
		cmd            command.UpdateTenantCartAbandonedPolicyCommand
		wantErr        error
		wantEventsLen  int
		wantVersion    int
	}{
		"should return error when policy not created": {
			existingPolicy: false,
			cmd: command.UpdateTenantCartAbandonedPolicyCommand{
				TenantID:         tenantID,
				Title:            "Updated Policy",
				AbandonedMinutes: 60,
				QuietTimeFrom:    time.Date(2023, 1, 1, 23, 0, 0, 0, time.UTC),
				QuietTimeTo:      time.Date(2023, 1, 1, 7, 0, 0, 0, time.UTC),
			},
			wantErr:       errors.UnpermittedOp.New("tenant policy not created"),
			wantEventsLen: 0,
			wantVersion:   -1,
		},
		"should update existing policy": {
			existingPolicy: true,
			cmd: command.UpdateTenantCartAbandonedPolicyCommand{
				TenantID:         tenantID,
				Title:            "Updated Policy",
				AbandonedMinutes: 60,
				QuietTimeFrom:    time.Date(2023, 1, 1, 23, 0, 0, 0, time.UTC),
				QuietTimeTo:      time.Date(2023, 1, 1, 7, 0, 0, 0, time.UTC),
			},
			wantErr:       nil,
			wantEventsLen: 1,
			wantVersion:   2,
		},
		"should not create event when no changes": {
			existingPolicy: true,
			cmd: command.UpdateTenantCartAbandonedPolicyCommand{
				TenantID:         tenantID,
				Title:            "Existing Policy",
				AbandonedMinutes: 45,
				QuietTimeFrom:    time.Date(2023, 1, 1, 21, 0, 0, 0, time.UTC),
				QuietTimeTo:      time.Date(2023, 1, 1, 9, 0, 0, 0, time.UTC),
			},
			wantErr:       nil,
			wantEventsLen: 0,
			wantVersion:   1,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			policy := aggregate.NewTenantCartAbandonedPolicyAggregate()
			if tt.existingPolicy {
				existingCmd := command.CreateTenantCartAbandonedPolicyCommand{
					TenantID:         tenantID,
					Title:            "Existing Policy",
					AbandonedMinutes: 45,
					QuietTimeFrom:    time.Date(2023, 1, 1, 21, 0, 0, 0, time.UTC),
					QuietTimeTo:      time.Date(2023, 1, 1, 9, 0, 0, 0, time.UTC),
				}
				policy.ExecuteCreateTenantCartAbandonedPolicyCommand(existingCmd)
				policy.MarkEventsAsCommitted()
			}

			// Act
			err := policy.ExecuteUpdateTenantCartAbandonedPolicyCommand(tt.cmd)

			// Assert
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "tenant policy not created")
			} else {
				assert.NoError(t, err)
			}

			assert.Len(t, policy.GetUncommittedEvents(), tt.wantEventsLen)
			assert.Equal(t, tt.wantVersion, policy.GetVersion())
		})
	}
}

func TestTenantCartAbandonedPolicyAggregate_CartAbandonedDelay(t *testing.T) {
	tenantID := uuid.New()

	tests := map[string]struct {
		abandonedMinutes int
		want             time.Duration
	}{
		"should return 30 minutes": {
			abandonedMinutes: 30,
			want:             30 * time.Minute,
		},
		"should return 60 minutes": {
			abandonedMinutes: 60,
			want:             60 * time.Minute,
		},
		"should return 120 minutes": {
			abandonedMinutes: 120,
			want:             120 * time.Minute,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			policy := aggregate.NewTenantCartAbandonedPolicyAggregate()
			cmd := command.CreateTenantCartAbandonedPolicyCommand{
				TenantID:         tenantID,
				Title:            "Test Policy",
				AbandonedMinutes: tt.abandonedMinutes,
				QuietTimeFrom:    time.Date(2023, 1, 1, 22, 0, 0, 0, time.UTC),
				QuietTimeTo:      time.Date(2023, 1, 1, 8, 0, 0, 0, time.UTC),
			}
			policy.ExecuteCreateTenantCartAbandonedPolicyCommand(cmd)

			// Act & Assert
			got := policy.CartAbandonedDelay()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTenantCartAbandonedPolicyAggregate_IsWithinQuietTime(t *testing.T) {
	tenantID := uuid.New()

	tests := map[string]struct {
		quietTimeFrom time.Time
		quietTimeTo   time.Time
		now           time.Time
		want          bool
	}{
		"should return true when within normal quiet time": {
			quietTimeFrom: time.Date(2023, 1, 1, 22, 0, 0, 0, time.UTC), // 22:00
			quietTimeTo:   time.Date(2023, 1, 1, 8, 0, 0, 0, time.UTC),  // 08:00
			now:           time.Date(2023, 1, 1, 23, 30, 0, 0, time.UTC), // 23:30
			want:          true,
		},
		"should return true when within overnight quiet time": {
			quietTimeFrom: time.Date(2023, 1, 1, 22, 0, 0, 0, time.UTC), // 22:00
			quietTimeTo:   time.Date(2023, 1, 1, 8, 0, 0, 0, time.UTC),  // 08:00
			now:           time.Date(2023, 1, 1, 2, 0, 0, 0, time.UTC),   // 02:00
			want:          true,
		},
		"should return false when outside quiet time": {
			quietTimeFrom: time.Date(2023, 1, 1, 22, 0, 0, 0, time.UTC), // 22:00
			quietTimeTo:   time.Date(2023, 1, 1, 8, 0, 0, 0, time.UTC),  // 08:00
			now:           time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC),  // 10:00
			want:          false,
		},
		"should return true when within same day quiet time": {
			quietTimeFrom: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC), // 12:00
			quietTimeTo:   time.Date(2023, 1, 1, 14, 0, 0, 0, time.UTC), // 14:00
			now:           time.Date(2023, 1, 1, 13, 0, 0, 0, time.UTC), // 13:00
			want:          true,
		},
		"should return false when quiet times are zero": {
			quietTimeFrom: time.Time{},
			quietTimeTo:   time.Time{},
			now:           time.Date(2023, 1, 1, 13, 0, 0, 0, time.UTC),
			want:          false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			policy := aggregate.NewTenantCartAbandonedPolicyAggregate()
			cmd := command.CreateTenantCartAbandonedPolicyCommand{
				TenantID:         tenantID,
				Title:            "Test Policy",
				AbandonedMinutes: 30,
				QuietTimeFrom:    tt.quietTimeFrom,
				QuietTimeTo:      tt.quietTimeTo,
			}
			policy.ExecuteCreateTenantCartAbandonedPolicyCommand(cmd)

			// Act
			got, err := policy.IsWithinQuietTime(tt.now)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTenantCartAbandonedPolicyAggregate_Hydration(t *testing.T) {
	tenantID := uuid.New()

	tests := map[string]struct {
		events      []event.Event
		wantVersion int
		wantTitle   string
	}{
		"should hydrate empty policy with no events": {
			events:      []event.Event{},
			wantVersion: -1,
			wantTitle:   "",
		},
		"should hydrate policy with created event": {
			events: []event.Event{
				event.NewTenantCartAbandonedPolicyCreatedEvent(
					tenantID,
					1,
					"Test Policy",
					30,
					time.Date(2023, 1, 1, 22, 0, 0, 0, time.UTC),
					time.Date(2023, 1, 1, 8, 0, 0, 0, time.UTC),
				),
			},
			wantVersion: 1,
			wantTitle:   "Test Policy",
		},
		"should hydrate policy with created and updated events": {
			events: []event.Event{
				event.NewTenantCartAbandonedPolicyCreatedEvent(
					tenantID,
					1,
					"Test Policy",
					30,
					time.Date(2023, 1, 1, 22, 0, 0, 0, time.UTC),
					time.Date(2023, 1, 1, 8, 0, 0, 0, time.UTC),
				),
				event.NewTenantCartAbandonedPolicyUpdatedEvent(
					tenantID,
					2,
					"Updated Policy",
					60,
					time.Date(2023, 1, 1, 23, 0, 0, 0, time.UTC),
					time.Date(2023, 1, 1, 7, 0, 0, 0, time.UTC),
				),
			},
			wantVersion: 2,
			wantTitle:   "Updated Policy",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			policy := aggregate.NewTenantCartAbandonedPolicyAggregate()

			// Act & Assert
			err := policy.Hydration(tt.events)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantVersion, policy.GetVersion())
			assert.Equal(t, tt.wantTitle, policy.GetTitle())
			assert.Len(t, policy.GetUncommittedEvents(), 0)
		})
	}
}

func TestTenantCartAbandonedPolicyAggregate_MarkEventsAsCommitted(t *testing.T) {
	tenantID := uuid.New()

	tests := map[string]struct {
		executeCommand bool
	}{
		"should clear uncommitted events after creating policy": {
			executeCommand: true,
		},
		"should handle empty events": {
			executeCommand: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			policy := aggregate.NewTenantCartAbandonedPolicyAggregate()
			if tt.executeCommand {
				cmd := command.CreateTenantCartAbandonedPolicyCommand{
					TenantID:         tenantID,
					Title:            "Test Policy",
					AbandonedMinutes: 30,
					QuietTimeFrom:    time.Date(2023, 1, 1, 22, 0, 0, 0, time.UTC),
					QuietTimeTo:      time.Date(2023, 1, 1, 8, 0, 0, 0, time.UTC),
				}
				policy.ExecuteCreateTenantCartAbandonedPolicyCommand(cmd)
			}

			// Act
			beforeEventCount := len(policy.GetUncommittedEvents())
			policy.MarkEventsAsCommitted()

			// Assert
			assert.Len(t, policy.GetUncommittedEvents(), 0)

			if beforeEventCount > 0 {
				assert.Greater(t, beforeEventCount, 0, "Should have had events before committing")
			}
		})
	}
}