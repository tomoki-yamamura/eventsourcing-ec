package testutil

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/config"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/database/client"

	_ "github.com/go-sql-driver/mysql"
)

func SetupTestDB(t *testing.T) (*sqlx.DB, func()) {
	cfg := config.DatabaseConfig{
		Host:     "localhost",
		Port:     "23306",
		User:     "root",
		Password: "password",
		Name:     "event_test",
	}

	dbClient, err := client.NewClient(cfg)
	require.NoError(t, err)

	cleanup := func() {
		_ = dbClient.Close()
	}

	return dbClient.GetDB(), cleanup
}
