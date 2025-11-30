package command_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/event"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/database/eventstore"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/database/outbox"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/database/testutil"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/database/transaction"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/command"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/command/input"
)

type testPresenter struct {
	lastAggregateID string
	lastVersion     int
	lastEvents      []event.Event
	lastError       error
}

func (p *testPresenter) PresentSuccess(ctx context.Context, aggregateID string, version int, events []event.Event) error {
	p.lastAggregateID = aggregateID
	p.lastVersion = version
	p.lastEvents = events
	return nil
}

func (p *testPresenter) PresentError(ctx context.Context, err error) error {
	p.lastError = err
	return nil
}

func TestCartAddItemCommand_Execute(t *testing.T) {
	tests := map[string]struct {
		input              *input.AddItemToCartInput
		expectedEventCount int
		expectedVersion    int
	}{
		"add item to new cart": {
			input: &input.AddItemToCartInput{
				CartID:   uuid.New().String(),
				UserID:   uuid.New().String(),
				ItemID:   uuid.New().String(),
				Name:     "Test Item",
				Price:    100.0,
				TenantID: uuid.New().String(),
			},
			expectedEventCount: 0,
			expectedVersion:    2,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Arrange
			dbClient := testutil.NewTestDBClient(t)
			ctx, tx := testutil.BeginTxCtx(t, dbClient)
			txRepo := transaction.NewTransaction(dbClient.GetDB())
			eventStore := eventstore.NewEventStore(testutil.FakeDeserializer{})
			outboxRepo := outbox.NewOutboxRepository()
			presenter := &testPresenter{}

			// Create command
			addItemCmd := command.NewCartAddItemCommand(txRepo, eventStore, outboxRepo)

			// Act
			err := addItemCmd.Execute(ctx, tt.input, presenter)

			// Assert
			require.NoError(t, err)
			require.Nil(t, presenter.lastError)
			require.NotEmpty(t, presenter.lastAggregateID)
			require.Equal(t, tt.expectedVersion, presenter.lastVersion)
			_, parseErr := uuid.Parse(presenter.lastAggregateID)
			require.NoError(t, parseErr)
			require.Equal(t, tt.input.CartID, presenter.lastAggregateID)

			rollbackErr := tx.Rollback()
			require.NoError(t, rollbackErr)

			t.Cleanup(func() {
				_, cleanupErr := dbClient.GetDB().Exec("DELETE FROM events WHERE aggregate_id = ?", tt.input.CartID)
				require.NoError(t, cleanupErr)
				_, cleanupErr = dbClient.GetDB().Exec("DELETE FROM outbox WHERE aggregate_id = ?", tt.input.CartID)
				require.NoError(t, cleanupErr)
			})
		})
	}
}
