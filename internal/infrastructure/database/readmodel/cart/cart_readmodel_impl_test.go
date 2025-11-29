package cart_test

import (
	"context"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	"github.com/tomoki-yamamura/eventsourcing-ec/internal/config"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/errors"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/database/client"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/database/readmodel/cart"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/database/transaction"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/readmodelstore/dto"
)

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

func TestCartReadModel_Get(t *testing.T) {

	tests := map[string]struct {
		cartID    string
		wantError bool
	}{
		"get non-existent cart": {
			cartID:    "99999999-9999-9999-9999-999999999999",
			wantError: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			dbClient := newTestDBClient(t)
			ctx, tx := beginTxCtx(t, dbClient)
			store := cart.NewCartReadModel(transaction.NewTransaction(dbClient.GetDB()))

			_, err := store.Get(ctx, tt.cartID)

			rollbackErr := tx.Rollback()
			require.NoError(t, rollbackErr)

			if tt.wantError {
				require.Error(t, err)
				require.True(t, errors.IsCode(err, errors.NotFound))
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCartReadModel_Upsert(t *testing.T) {
	testCartID := "12345678-1234-1234-1234-123456789012"

	tests := map[string]struct {
		cartData  *dto.CartViewDTO
		wantError bool
	}{
		"successful upsert": {
			cartData: &dto.CartViewDTO{
				ID:          testCartID,
				UserID:      "user123",
				Status:      "active",
				TotalAmount: 100.0,
				ItemCount:   2,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
				Version:     1,
				Items: []dto.CartItemViewDTO{
					{
						ID:     "item1",
						CartID: testCartID,
						Name:   "Test Item",
						Price:  50.0,
					},
				},
			},
			wantError: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			dbClient := newTestDBClient(t)
			ctx, tx := beginTxCtx(t, dbClient)
			store := cart.NewCartReadModel(transaction.NewTransaction(dbClient.GetDB()))

			err := store.Upsert(ctx, testCartID, tt.cartData)

			rollbackErr := tx.Rollback()
			require.NoError(t, rollbackErr)

			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
