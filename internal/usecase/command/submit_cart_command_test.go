package command_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/event"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/database/eventstore"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/database/eventstore/deserializer"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/database/outbox"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/database/testutil"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/database/transaction"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/command"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/command/input"
)

type submitTestPresenter struct {
	lastAggregateID string
	lastVersion     int
	lastError       error
}

func (p *submitTestPresenter) PresentSuccess(ctx context.Context, aggregateID string, version int, events []event.Event) error {
	p.lastAggregateID = aggregateID
	p.lastVersion = version
	return nil
}

func (p *submitTestPresenter) PresentError(ctx context.Context, err error) error {
	p.lastError = err
	return nil
}

func TestSubmitCartCommand_Execute(t *testing.T) {
	cartID := uuid.New().String()
	
	tests := map[string]struct {
		input           *input.SubmitCartInput
		expectedVersion int
	}{
		"submit cart with items": {
			input: &input.SubmitCartInput{
				CartID: cartID,
			},
			expectedVersion: 3,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Arrange
			dbClient := testutil.NewTestDBClient(t)
			txRepo := transaction.NewTransaction(dbClient.GetDB())
			eventStore := eventstore.NewEventStore(deserializer.NewEventDeserializer())
			outboxRepo := outbox.NewOutboxRepository()

			// First add an item to the cart
			addItemCmd := command.NewCartAddItemCommand(txRepo, eventStore, outboxRepo)
			addItemPresenter := &submitTestPresenter{}
			err := addItemCmd.Execute(context.Background(), &input.AddItemToCartInput{
				CartID: tt.input.CartID,
				UserID: uuid.New().String(),
				ItemID: uuid.New().String(),
				Name:   "Test Item",
				Price:  100.0,
			}, addItemPresenter)
			require.NoError(t, err)

			// Then submit the cart
			submitCmd := command.NewSubmitCartCommand(txRepo, eventStore, outboxRepo)
			presenter := &submitTestPresenter{}

			// Act
			err = submitCmd.Execute(context.Background(), tt.input, presenter)

			// Assert
			require.NoError(t, err)
			require.Nil(t, presenter.lastError)
			require.NotEmpty(t, presenter.lastAggregateID)
			require.Equal(t, tt.expectedVersion, presenter.lastVersion)
			_, parseErr := uuid.Parse(presenter.lastAggregateID)
			require.NoError(t, parseErr)
			require.Equal(t, tt.input.CartID, presenter.lastAggregateID)

			t.Cleanup(func() {
				_, cleanupErr := dbClient.GetDB().Exec("DELETE FROM events WHERE aggregate_id = ?", tt.input.CartID)
				require.NoError(t, cleanupErr)
				_, cleanupErr = dbClient.GetDB().Exec("DELETE FROM outbox WHERE aggregate_id = ?", tt.input.CartID)
				require.NoError(t, cleanupErr)
			})
		})
	}
}
