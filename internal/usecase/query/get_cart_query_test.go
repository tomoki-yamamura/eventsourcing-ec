package query_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/database/testutil"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/database/transaction"
	cartReadModel "github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/database/readmodel/cart"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/readmodelstore/dto"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/query"
)

type queryTestPresenter struct {
	lastData  []byte
	lastError error
}

func (p *queryTestPresenter) PresentSuccess(ctx context.Context, data []byte) error {
	p.lastData = data
	return nil
}

func (p *queryTestPresenter) PresentError(ctx context.Context, err error) error {
	p.lastError = err
	return nil
}

func TestGetCartQuery_Query(t *testing.T) {
	aggregateID := uuid.New().String()
	userID := uuid.New().String()
	itemID1 := uuid.New().String()
	itemID2 := uuid.New().String()

	tests := map[string]struct {
		aggregateID       string
		expectedID        string
		expectedUserID    string
		expectedStatus    string
		expectedItemCount int
	}{
		"get cart with items": {
			aggregateID:       aggregateID,
			expectedID:        aggregateID,
			expectedUserID:    userID,
			expectedStatus:    "OPEN",
			expectedItemCount: 2,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Arrange
			dbClient := testutil.NewTestDBClient(t)
			txRepo := transaction.NewTransaction(dbClient.GetDB())
			_, err := dbClient.GetDB().Exec(`
				INSERT INTO carts (id, user_id, status, total_amount, item_count, version) VALUES (?, ?, ?, ?, ?, ?)
			`, tt.aggregateID, tt.expectedUserID, tt.expectedStatus, 150.0, 2, 1)
			require.NoError(t, err)
			_, err = dbClient.GetDB().Exec(`
				INSERT INTO cart_items (id, cart_id, name, price) VALUES (?, ?, ?, ?), (?, ?, ?, ?)
			`, itemID1, tt.aggregateID, "Test Item 1", 100.0, itemID2, tt.aggregateID, "Test Item 2", 50.0)
			require.NoError(t, err)
			cartStore := cartReadModel.NewCartReadModel(txRepo)
			getCartQuery := query.NewGetCartQuery(cartStore)
			presenter := &queryTestPresenter{}

			// Act
			err = getCartQuery.Query(context.Background(), tt.aggregateID, presenter)

			// Assert
			require.NoError(t, err)
			require.Nil(t, presenter.lastError)
			require.NotNil(t, presenter.lastData)
			var actualCart dto.CartViewDTO
			err = json.Unmarshal(presenter.lastData, &actualCart)
			require.NoError(t, err)
			require.Equal(t, tt.expectedID, actualCart.ID)
			require.Equal(t, tt.expectedUserID, actualCart.UserID)
			require.Equal(t, tt.expectedStatus, actualCart.Status)
			require.Equal(t, tt.expectedItemCount, len(actualCart.Items))

			if len(actualCart.Items) > 0 {
				require.Equal(t, "Test Item 1", actualCart.Items[0].Name)
				require.Equal(t, 100.0, actualCart.Items[0].Price)
			}
			if len(actualCart.Items) > 1 {
				require.Equal(t, "Test Item 2", actualCart.Items[1].Name)
				require.Equal(t, 50.0, actualCart.Items[1].Price)
			}

			t.Cleanup(func() {
				_, cleanupErr := dbClient.GetDB().Exec("DELETE FROM cart_items WHERE cart_id = ?", tt.aggregateID)
				require.NoError(t, cleanupErr)
				_, cleanupErr = dbClient.GetDB().Exec("DELETE FROM carts WHERE id = ?", tt.aggregateID)
				require.NoError(t, cleanupErr)
			})
		})
	}
}
