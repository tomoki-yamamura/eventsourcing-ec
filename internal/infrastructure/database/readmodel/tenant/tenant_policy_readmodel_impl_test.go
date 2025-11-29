package tenant_test

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
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/database/readmodel/tenant"
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

func TestTenantPolicyReadModel_Get(t *testing.T) {
	tests := map[string]struct {
		tenantID  string
		wantError bool
	}{
		"get non-existent policy": {
			tenantID:  "99999999-9999-9999-9999-999999999999",
			wantError: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			dbClient := newTestDBClient(t)
			ctx, tx := beginTxCtx(t, dbClient)
			store := tenant.NewTenantPolicyReadModel(transaction.NewTransaction(dbClient.GetDB()))

			_, err := store.Get(ctx, tt.tenantID)

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

func TestTenantPolicyReadModel_Upsert(t *testing.T) {
	testTenantID := "12345678-1234-1234-1234-123456789012"

	tests := map[string]struct {
		policyData *dto.TenantPolicyViewDTO
		wantError  bool
	}{
		"successful upsert": {
			policyData: &dto.TenantPolicyViewDTO{
				ID:               testTenantID,
				Title:            "Test Policy",
				AbandonedMinutes: 30,
				QuietTimeFrom:    time.Date(2023, 1, 1, 9, 0, 0, 0, time.UTC),
				QuietTimeTo:      time.Date(2023, 1, 1, 17, 0, 0, 0, time.UTC),
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
				Version:          1,
			},
			wantError: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			dbClient := newTestDBClient(t)
			ctx, tx := beginTxCtx(t, dbClient)
			store := tenant.NewTenantPolicyReadModel(transaction.NewTransaction(dbClient.GetDB()))

			err := store.Upsert(ctx, testTenantID, tt.policyData)

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