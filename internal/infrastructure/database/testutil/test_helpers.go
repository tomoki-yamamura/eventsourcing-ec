package testutil

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	"github.com/tomoki-yamamura/eventsourcing-ec/internal/config"
	domainevent "github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/event"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/database/client"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/database/transaction"
)

const (
	TestTypeA = "TestEventA"
	TestTypeB = "TestEventB" 
	TestTypeC = "TestEventC"
)

type TestEvent struct {
	AggregateID uuid.UUID `json:"aggregate_id"`
	EventID     uuid.UUID `json:"event_id"`
	Type        string    `json:"event_type"`
	Version     int       `json:"version"`
	Title       string    `json:"title,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

func (e TestEvent) GetAggregateID() uuid.UUID { return e.AggregateID }
func (e TestEvent) GetEventID() uuid.UUID     { return e.EventID }
func (e TestEvent) GetEventType() string      { return e.Type }
func (e TestEvent) GetVersion() int           { return e.Version }
func (e TestEvent) GetTimestamp() time.Time   { return e.CreatedAt }
func (e TestEvent) GetAggregateType() string  { return "cart" }

type FakeDeserializer struct{}

func (f FakeDeserializer) Deserialize(eventType string, data []byte) (domainevent.Event, error) {
	var te TestEvent
	if err := json.Unmarshal(data, &te); err != nil {
		return nil, err
	}
	return te, nil
}

func NewTestDBClient(t *testing.T) *client.Client {
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

func BeginTxCtx(t *testing.T, dbClient *client.Client) (context.Context, *sqlx.Tx) {
	t.Helper()

	db := dbClient.GetDB()
	tx, err := db.Beginx()
	require.NoError(t, err)

	ctx := transaction.WithTx(context.Background(), tx)

	return ctx, tx
}
