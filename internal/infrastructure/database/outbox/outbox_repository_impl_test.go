package outbox_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	domainevent "github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/event"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/value"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/database/outbox"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/database/testutil"
)

func TestOutboxRepository_SaveEvents(t *testing.T) {
	testAggregateID := uuid.MustParse("12345678-1234-1234-1234-123456789012")

	tests := map[string]struct {
		events []domainevent.Event
	}{
		"successful save single event": {
			events: []domainevent.Event{
				testutil.TestEvent{
					AggregateID: testAggregateID,
					EventID:     uuid.New(),
					Type:        testutil.TestTypeA,
					Version:     1,
					CreatedAt:   time.Now(),
				},
			},
		},
		"successful save multiple events": {
			events: []domainevent.Event{
				testutil.TestEvent{
					AggregateID: testAggregateID,
					EventID:     uuid.New(),
					Type:        testutil.TestTypeA,
					Version:     1,
					CreatedAt:   time.Now(),
				},
				testutil.TestEvent{
					AggregateID: testAggregateID,
					EventID:     uuid.New(),
					Type:        testutil.TestTypeB,
					Version:     2,
					Title:       "Test Item",
					CreatedAt:   time.Now(),
				},
			},
		},
		"save empty events": {
			events: []domainevent.Event{},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			dbClient := testutil.NewTestDBClient(t)
			ctx, tx := testutil.BeginTxCtx(t, dbClient)
			repo := outbox.NewOutboxRepository()

			err := repo.SaveEvents(ctx, testAggregateID, tt.events)

			rollbackErr := tx.Rollback()
			require.NoError(t, rollbackErr)

			require.NoError(t, err)
		})
	}
}

func TestOutboxRepository_GetPendingEvents(t *testing.T) {
	testAggregateID := uuid.MustParse("12345678-1234-1234-1234-123456789012")

	tests := map[string]struct {
		savedEvents   []domainevent.Event
		limit         int
		expectedCount int
	}{
		"get pending events with limit": {
			savedEvents: []domainevent.Event{
				testutil.TestEvent{
					AggregateID: testAggregateID,
					EventID:     uuid.New(),
					Type:        testutil.TestTypeA,
					Version:     1,
					CreatedAt:   time.Now().Add(-2 * time.Minute),
				},
				testutil.TestEvent{
					AggregateID: testAggregateID,
					EventID:     uuid.New(),
					Type:        testutil.TestTypeB,
					Version:     2,
					Title:       "Test Item",
					CreatedAt:   time.Now().Add(-1 * time.Minute),
				},
			},
			limit:         5,
			expectedCount: 2,
		},
		"get pending events with small limit": {
			savedEvents: []domainevent.Event{
				testutil.TestEvent{
					AggregateID: testAggregateID,
					EventID:     uuid.New(),
					Type:        testutil.TestTypeA,
					Version:     1,
					CreatedAt:   time.Now().Add(-2 * time.Minute),
				},
				testutil.TestEvent{
					AggregateID: testAggregateID,
					EventID:     uuid.New(),
					Type:        testutil.TestTypeB,
					Version:     2,
					Title:       "Test Item",
					CreatedAt:   time.Now().Add(-1 * time.Minute),
				},
			},
			limit:         1,
			expectedCount: 1,
		},
		"get pending events when none exist": {
			savedEvents:   []domainevent.Event{},
			limit:         5,
			expectedCount: 0,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Arrange
			dbClient := testutil.NewTestDBClient(t)
			ctx, tx := testutil.BeginTxCtx(t, dbClient)
			repo := outbox.NewOutboxRepository()

			if len(tt.savedEvents) > 0 {
				err := repo.SaveEvents(ctx, testAggregateID, tt.savedEvents)
				require.NoError(t, err)
			}

			// Act
			events, err := repo.GetPendingEvents(ctx, tt.limit)

			rollbackErr := tx.Rollback()
			require.NoError(t, rollbackErr)

			// Assert
			require.NoError(t, err)
			require.Len(t, events, tt.expectedCount)

			if len(events) > 1 {
				for i := 1; i < len(events); i++ {
					require.True(t, events[i-1].CreatedAt.Before(events[i].CreatedAt) || 
						events[i-1].CreatedAt.Equal(events[i].CreatedAt))
				}
			}

			for _, evt := range events {
				require.Equal(t, value.OutboxStatusPending, evt.Status)
			}
		})
	}
}

func TestOutboxRepository_MarkAsPublished(t *testing.T) {
	testAggregateID := uuid.MustParse("12345678-1234-1234-1234-123456789012")

	tests := map[string]struct {
		savedEvents []domainevent.Event
		eventIDs    []uuid.UUID
		description string
	}{
		"mark single event as published": {
			savedEvents: []domainevent.Event{
				testutil.TestEvent{
					AggregateID: testAggregateID,
					EventID:     uuid.MustParse("11111111-1111-1111-1111-111111111111"),
					Type:        testutil.TestTypeA,
					Version:     1,
					CreatedAt:   time.Now(),
				},
			},
			eventIDs:    []uuid.UUID{uuid.MustParse("11111111-1111-1111-1111-111111111111")},
			description: "should mark one event as published",
		},
		"mark multiple events as published": {
			savedEvents: []domainevent.Event{
				testutil.TestEvent{
					AggregateID: testAggregateID,
					EventID:     uuid.MustParse("11111111-1111-1111-1111-111111111111"),
					Type:        testutil.TestTypeA,
					Version:     1,
					CreatedAt:   time.Now(),
				},
				testutil.TestEvent{
					AggregateID: testAggregateID,
					EventID:     uuid.MustParse("22222222-2222-2222-2222-222222222222"),
					Type:        testutil.TestTypeB,
					Version:     2,
					Title:       "Test Item",
					CreatedAt:   time.Now(),
				},
			},
			eventIDs: []uuid.UUID{
				uuid.MustParse("11111111-1111-1111-1111-111111111111"),
				uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			},
			description: "should mark multiple events as published",
		},
		"mark no events as published": {
			savedEvents: []domainevent.Event{},
			eventIDs:    []uuid.UUID{},
			description: "should handle empty event IDs gracefully",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Arrange
			dbClient := testutil.NewTestDBClient(t)
			ctx, tx := testutil.BeginTxCtx(t, dbClient)
			repo := outbox.NewOutboxRepository()

			if len(tt.savedEvents) > 0 {
				err := repo.SaveEvents(ctx, testAggregateID, tt.savedEvents)
				require.NoError(t, err)
			}

			// Act
			err := repo.MarkAsPublished(ctx, tt.eventIDs)

			rollbackErr := tx.Rollback()
			require.NoError(t, rollbackErr)

			// Assert
			require.NoError(t, err)
		})
	}
}

func TestOutboxRepository_MarkAsFailed(t *testing.T) {
	testAggregateID := uuid.MustParse("12345678-1234-1234-1234-123456789012")
	testEventID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	tests := map[string]struct {
		savedEvents  []domainevent.Event
		eventID      uuid.UUID
		errorMessage string
	}{
		"mark event as failed with error message": {
			savedEvents: []domainevent.Event{
				testutil.TestEvent{
					AggregateID: testAggregateID,
					EventID:     testEventID,
					Type:        testutil.TestTypeA,
					Version:     1,
					CreatedAt:   time.Now(),
				},
			},
			eventID:      testEventID,
			errorMessage: "failed to publish event",
		},
		"mark event as failed with empty error message": {
			savedEvents: []domainevent.Event{
				testutil.TestEvent{
					AggregateID: testAggregateID,
					EventID:     testEventID,
					Type:        testutil.TestTypeA,
					Version:     1,
					CreatedAt:   time.Now(),
				},
			},
			eventID:      testEventID,
			errorMessage: "",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Arrange
			dbClient := testutil.NewTestDBClient(t)
			ctx, tx := testutil.BeginTxCtx(t, dbClient)
			repo := outbox.NewOutboxRepository()

			if len(tt.savedEvents) > 0 {
				err := repo.SaveEvents(ctx, testAggregateID, tt.savedEvents)
				require.NoError(t, err)
			}

			// Act
			err := repo.MarkAsFailed(ctx, tt.eventID, tt.errorMessage)

			rollbackErr := tx.Rollback()
			require.NoError(t, rollbackErr)

			// Assert
			require.NoError(t, err)
		})
	}
}

func TestOutboxRepository_IncrementRetryCount(t *testing.T) {
	testAggregateID := uuid.MustParse("12345678-1234-1234-1234-123456789012")
	testEventID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	tests := map[string]struct {
		savedEvents []domainevent.Event
		eventID     uuid.UUID
	}{
		"increment retry count for existing event": {
			savedEvents: []domainevent.Event{
				testutil.TestEvent{
					AggregateID: testAggregateID,
					EventID:     testEventID,
					Type:        testutil.TestTypeA,
					Version:     1,
					CreatedAt:   time.Now(),
				},
			},
			eventID: testEventID,
		},
		"increment retry count for non-existent event": {
			savedEvents: []domainevent.Event{},
			eventID:     uuid.New(),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Arrange
			dbClient := testutil.NewTestDBClient(t)
			ctx, tx := testutil.BeginTxCtx(t, dbClient)
			repo := outbox.NewOutboxRepository()

			if len(tt.savedEvents) > 0 {
				err := repo.SaveEvents(ctx, testAggregateID, tt.savedEvents)
				require.NoError(t, err)
			}

			// Act
			err := repo.IncrementRetryCount(ctx, tt.eventID)

			rollbackErr := tx.Rollback()
			require.NoError(t, rollbackErr)

			// Assert
			require.NoError(t, err)
		})
	}
}