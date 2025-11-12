package eventstore_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	"github.com/tomoki-yamamura/eventsourcing-ec/internal/config"
	domainevent "github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/event"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/errors"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/database/client"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/database/eventstore"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/database/transaction"
)

type testEvent struct {
	AggregateID uuid.UUID `json:"aggregate_id"`
	EventID     uuid.UUID `json:"event_id"`
	Type        string    `json:"event_type"`
	Version     int       `json:"version"`
	Title       string    `json:"title,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

func (e testEvent) GetAggregateID() uuid.UUID { return e.AggregateID }
func (e testEvent) GetEventID() uuid.UUID     { return e.EventID }
func (e testEvent) GetEventType() string      { return e.Type }
func (e testEvent) GetVersion() int           { return e.Version }
func (e testEvent) GetTimestamp() time.Time   { return e.CreatedAt }

type fakeDeserializer struct{}

func (f fakeDeserializer) Deserialize(eventType string, data []byte) (domainevent.Event, error) {
	var te testEvent
	if err := json.Unmarshal(data, &te); err != nil {
		return nil, err
	}
	return te, nil
}

func newTestDBClient(t *testing.T) *client.Client {
	t.Helper()

	testCfg, err := config.NewTestDatabaseConfig()
	require.NoError(t, err)

	c, err := client.NewClient(config.DatabaseConfig{
		User:     testCfg.User,
		Password: testCfg.Password,
		Host:     testCfg.Host,
		Port:     testCfg.Port,
		Name:     testCfg.Name,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		err = c.Close()
		require.NoError(t, err)
	})

	return c
}

func beginTxCtx(t *testing.T, dbClient *client.Client) (context.Context, *sqlx.Tx) {
	t.Helper()

	db := dbClient.GetDB()
	tx, err := db.Beginx()
	require.NoError(t, err)

	ctx := transaction.WithTx(context.Background(), tx)

	return ctx, tx
}

func TestEventStore_SaveEvents(t *testing.T) {
	testAggregateID := uuid.MustParse("12345678-1234-1234-1234-123456789012")

	tests := map[string]struct {
		events    []domainevent.Event
		wantError error
	}{
		"successful save single event": {
			events: []domainevent.Event{
				testEvent{
					AggregateID: testAggregateID,
					EventID:     uuid.New(),
					Type:        "TodoListCreated",
					Version:     1,
					CreatedAt:   time.Now(),
				},
			},
			wantError: nil,
		},
		"successful save multiple events": {
			events: []domainevent.Event{
				testEvent{
					AggregateID: testAggregateID,
					EventID:     uuid.New(),
					Type:        "TodoListCreated",
					Version:     1,
					CreatedAt:   time.Now(),
				},
				testEvent{
					AggregateID: testAggregateID,
					EventID:     uuid.New(),
					Type:        "TodoAdded",
					Version:     2,
					Title:       "Test Todo",
					CreatedAt:   time.Now(),
				},
			},
			wantError: nil,
		},
		"no conflict with different versions": {
			events: []domainevent.Event{
				testEvent{
					AggregateID: testAggregateID,
					EventID:     uuid.New(),
					Type:        "TestEvent",
					Version:     1,
					CreatedAt:   time.Now(),
				},
				testEvent{
					AggregateID: testAggregateID,
					EventID:     uuid.New(),
					Type:        "TestEvent",
					Version:     2,
					CreatedAt:   time.Now(),
				},
			},
			wantError: nil,
		},
		"version conflict same aggregate same version": {
			events: []domainevent.Event{
				testEvent{
					AggregateID: testAggregateID,
					EventID:     uuid.New(),
					Type:        "TestEvent",
					Version:     1,
					CreatedAt:   time.Now(),
				},
				testEvent{
					AggregateID: testAggregateID,
					EventID:     uuid.New(),
					Type:        "TestEvent",
					Version:     1,
					CreatedAt:   time.Now(),
				},
			},
			wantError: errors.OptimisticLock.New(""),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			dbClient := newTestDBClient(t)
			ctx, tx := beginTxCtx(t, dbClient)
			store := eventstore.NewEventStore(fakeDeserializer{})

			err := store.SaveEvents(ctx, testAggregateID, tt.events)

			// 明示的にrollbackしてテストデータをclean up
			rollbackErr := tx.Rollback()
			require.NoError(t, rollbackErr)

			if tt.wantError != nil {
				require.Error(t, err)
				require.True(t, errors.IsCode(err, errors.OptimisticLock))
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestEventStore_LoadEvents(t *testing.T) {
	testAggregateID := uuid.MustParse("12345678-1234-1234-1234-123456789012")

	tests := map[string]struct {
		aggregateID   uuid.UUID
		savedEvents   []domainevent.Event
		expectedCount int
		wantError     error
	}{
		"load existing events": {
			aggregateID: testAggregateID,
			savedEvents: []domainevent.Event{
				testEvent{
					AggregateID: testAggregateID,
					EventID:     uuid.New(),
					Type:        "TodoAdded",
					Version:     2,
					Title:       "Second Todo",
					CreatedAt:   time.Now(),
				},
				testEvent{
					AggregateID: testAggregateID,
					EventID:     uuid.New(),
					Type:        "TodoListCreated",
					Version:     1,
					CreatedAt:   time.Now().Add(-time.Minute),
				},
				testEvent{
					AggregateID: testAggregateID,
					EventID:     uuid.New(),
					Type:        "TodoAdded",
					Version:     3,
					Title:       "Third Todo",
					CreatedAt:   time.Now().Add(time.Minute),
				},
			},
			expectedCount: 3,
			wantError:     nil,
		},
		"load non-existent aggregate": {
			aggregateID:   uuid.MustParse("99999999-9999-9999-9999-999999999999"),
			savedEvents:   []domainevent.Event{},
			expectedCount: 0,
			wantError:     errors.NotFound.New("todo list not found"),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Arrange
			dbClient := newTestDBClient(t)
			ctx, tx := beginTxCtx(t, dbClient)
			store := eventstore.NewEventStore(fakeDeserializer{})
			if len(tt.savedEvents) > 0 {
				err := store.SaveEvents(ctx, testAggregateID, tt.savedEvents)
				require.NoError(t, err)
			}

			// Act
			loadedEvents, err := store.LoadEvents(ctx, tt.aggregateID)

			// 明示的にrollbackしてテストデータをclean up
			rollbackErr := tx.Rollback()
			require.NoError(t, rollbackErr)

			// Assert
			if tt.wantError != nil {
				require.Error(t, err)
				require.True(t, errors.IsCode(err, errors.NotFound))
			} else {
				require.NoError(t, err)
				require.Len(t, loadedEvents, tt.expectedCount)
			}
		})
	}
}
