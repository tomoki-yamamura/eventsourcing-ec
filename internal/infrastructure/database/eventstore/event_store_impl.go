package eventstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/event"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/repository"
	appErrors "github.com/tomoki-yamamura/eventsourcing-ec/internal/errors"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/database/transaction"
)

type eventStoreImpl struct {
	deserializer repository.EventDeserializer
}

func NewEventStore(deserializer repository.EventDeserializer) repository.EventStore {
	return &eventStoreImpl{
		deserializer: deserializer,
	}
}

func (e *eventStoreImpl) SaveEvents(ctx context.Context, aggregateID uuid.UUID, events []event.Event) error {
	tx, err := transaction.GetTx(ctx)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO events (
			aggregate_id, 
			event_id, 
			event_type, 
			event_data, 
			version, 
			created_at
		) VALUES (?, ?, ?, ?, ?, ?)
	`

	for _, evt := range events {
		eventData, err := json.Marshal(evt)
		if err != nil {
			return err
		}

		_, err = tx.ExecContext(ctx, query,
			aggregateID,
			evt.GetEventID(),
			evt.GetEventType(),
			eventData,
			evt.GetVersion(),
			time.Now(),
		)
		if err != nil {
			if isDuplicateKeyError(err) {
				return appErrors.OptimisticLock.Wrap(err, fmt.Sprintf("version conflict for aggregate %s version %d", aggregateID, evt.GetVersion()))
			}
			return appErrors.RepositoryError.Wrap(err, "failed to save event")
		}
	}

	return nil
}

func isDuplicateKeyError(err error) bool {
	// MySQL error 1062 = ER_DUP_ENTRY
	if errors.Is(err, &mysql.MySQLError{Number: 1062}) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate") || strings.Contains(msg, "constraint failed")
}

func (e *eventStoreImpl) LoadEvents(ctx context.Context, aggregateID uuid.UUID) ([]event.Event, error) {
	tx, err := transaction.GetTx(ctx)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT event_id, event_type, event_data, version, created_at
		FROM events 
		WHERE aggregate_id = ?
		ORDER BY version ASC
	`

	rows, err := tx.QueryContext(ctx, query, aggregateID)
	if err != nil {
		return nil, appErrors.QueryError.Wrap(err, "failed to load events")
	}
	defer rows.Close()

	var events []event.Event
	for rows.Next() {
		var eventID uuid.UUID
		var eventType string
		var eventData []byte
		var version int
		var createdAt time.Time

		err := rows.Scan(&eventID, &eventType, &eventData, &version, &createdAt)
		if err != nil {
			return nil, appErrors.QueryError.Wrap(err, "failed to scan event row")
		}

		evt, err := e.deserializer.Deserialize(eventType, eventData)
		if err != nil {
			return nil, appErrors.QueryError.Wrap(err, fmt.Sprintf("failed to deserialize event %s", eventType))
		}

		events = append(events, evt)
	}

	if err := rows.Err(); err != nil {
		return nil, appErrors.QueryError.Wrap(err, "rows iteration error")
	}

	if len(events) == 0 {
		return nil, appErrors.NotFound.New("todo list not found")
	}

	return events, nil
}

func (e *eventStoreImpl) GetAllEvents(ctx context.Context) ([]event.Event, error) {
	tx, err := transaction.GetTx(ctx)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT event_id, event_type, event_data, version, created_at, aggregate_id
		FROM events 
		ORDER BY aggregate_id ASC, version ASC
	`

	rows, err := tx.QueryContext(ctx, query)
	if err != nil {
		return nil, appErrors.QueryError.Wrap(err, "failed to load all events")
	}
	defer rows.Close()

	var events []event.Event
	for rows.Next() {
		var eventID uuid.UUID
		var eventType string
		var eventData []byte
		var version int
		var createdAt time.Time
		var aggregateID uuid.UUID

		err := rows.Scan(&eventID, &eventType, &eventData, &version, &createdAt, &aggregateID)
		if err != nil {
			return nil, appErrors.QueryError.Wrap(err, "failed to scan event row")
		}

		evt, err := e.deserializer.Deserialize(eventType, eventData)
		if err != nil {
			return nil, appErrors.QueryError.Wrap(err, fmt.Sprintf("failed to deserialize event %s", eventType))
		}

		events = append(events, evt)
	}

	if err := rows.Err(); err != nil {
		return nil, appErrors.QueryError.Wrap(err, "rows iteration error")
	}

	return events, nil
}
