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
			ctx, tx := testutil.BeginTxCtx(t, dbClient)
			txRepo := transaction.NewTransaction(dbClient.GetDB())
			
			uniqueCartID := uuid.New().String()
			uniqueItemID1 := uuid.New().String()
			uniqueItemID2 := uuid.New().String()
			
			// Insert test data in transaction
			_, err := tx.ExecContext(ctx, `
				INSERT INTO carts (id, user_id, tenant_id, status, total_amount, item_count, version) 
				VALUES (?, ?, ?, ?, ?, ?, ?)
			`, uniqueCartID, tt.expectedUserID, uuid.New().String(), tt.expectedStatus, 150.0, 2, 1)
			require.NoError(t, err)
			
			_, err = tx.ExecContext(ctx, `
				INSERT INTO cart_items (id, cart_id, name, price) 
				VALUES (?, ?, ?, ?), (?, ?, ?, ?)
			`, uniqueItemID1, uniqueCartID, "Test Item 1", 100.0, uniqueItemID2, uniqueCartID, "Test Item 2", 50.0)
			require.NoError(t, err)
			
			// Commit the transaction so the data is visible to the query
			err = tx.Commit()
			require.NoError(t, err)

			// Act
			cartStore := cartReadModel.NewCartReadModel(txRepo)
			getCartQuery := query.NewGetCartQuery(cartStore)
			presenter := &queryTestPresenter{}
			err = getCartQuery.Query(context.Background(), uniqueCartID, presenter)

			// Assert
			require.NoError(t, err)
			require.Nil(t, presenter.lastError)
			require.NotNil(t, presenter.lastData)
			var actualCart dto.CartViewDTO
			err = json.Unmarshal(presenter.lastData, &actualCart)
			require.NoError(t, err)
			require.Equal(t, uniqueCartID, actualCart.ID)
			require.Equal(t, tt.expectedUserID, actualCart.UserID)
			require.Equal(t, tt.expectedStatus, actualCart.Status)
			require.Equal(t, tt.expectedItemCount, len(actualCart.Items))

			// Check items exist (order may vary)
			if len(actualCart.Items) >= 2 {
				itemNames := []string{actualCart.Items[0].Name, actualCart.Items[1].Name}
				require.Contains(t, itemNames, "Test Item 1")
				require.Contains(t, itemNames, "Test Item 2")
			}

			// Cleanup - delete test data
			t.Cleanup(func() {
				_, _ = dbClient.GetDB().Exec("DELETE FROM cart_items WHERE cart_id = ?", uniqueCartID)
				_, _ = dbClient.GetDB().Exec("DELETE FROM carts WHERE id = ?", uniqueCartID)
			})
		})
	}
}
