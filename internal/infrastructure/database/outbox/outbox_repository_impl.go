package outbox

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/event"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/repository"
	appErrors "github.com/tomoki-yamamura/eventsourcing-ec/internal/errors"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/database/transaction"
)

type outboxRepositoryImpl struct {
	db *sql.DB
}

func NewOutboxRepository(db *sql.DB) repository.OutboxRepository {
	return &outboxRepositoryImpl{
		db: db,
	}
}

func (o *outboxRepositoryImpl) SaveEvents(ctx context.Context, aggregateID uuid.UUID, aggregateType string, events []event.Event) error {
	tx, err := transaction.GetTx(ctx)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO outbox (
			event_id,
			aggregate_id,
			aggregate_type,
			event_type,
			event_data,
			version,
			created_at,
			status
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	for _, evt := range events {
		eventData, err := json.Marshal(evt)
		if err != nil {
			return appErrors.RepositoryError.Wrap(err, "failed to marshal event data")
		}

		_, err = tx.ExecContext(ctx, query,
			evt.GetEventID(),
			aggregateID,
			aggregateType,
			evt.GetEventType(),
			eventData,
			evt.GetVersion(),
			time.Now(),
			repository.OutboxStatusPending,
		)
		if err != nil {
			return appErrors.RepositoryError.Wrap(err, "failed to save event to outbox")
		}
	}

	return nil
}

func (o *outboxRepositoryImpl) GetPendingEvents(ctx context.Context, limit int) ([]repository.OutboxEvent, error) {
	var executor interface {
		QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	}
	
	if tx, err := transaction.GetTx(ctx); err == nil {
		executor = tx
	} else {
		executor = o.db
	}

	query := `
		SELECT id, event_id, aggregate_id, aggregate_type, event_type, event_data, version, 
		       created_at, published_at, status, retry_count, error_message
		FROM outbox 
		WHERE status = ? 
		ORDER BY created_at ASC 
		LIMIT ?
	`

	rows, err := executor.QueryContext(ctx, query, repository.OutboxStatusPending, limit)
	if err != nil {
		return nil, appErrors.QueryError.Wrap(err, "failed to get pending events")
	}
	defer rows.Close()

	var events []repository.OutboxEvent
	for rows.Next() {
		var event repository.OutboxEvent
		var publishedAt sql.NullTime
		var errorMessage sql.NullString

		err := rows.Scan(
			&event.ID,
			&event.EventID,
			&event.AggregateID,
			&event.AggregateType,
			&event.EventType,
			&event.EventData,
			&event.Version,
			&event.CreatedAt,
			&publishedAt,
			&event.Status,
			&event.RetryCount,
			&errorMessage,
		)
		if err != nil {
			return nil, appErrors.QueryError.Wrap(err, "failed to scan outbox event")
		}

		if publishedAt.Valid {
			event.PublishedAt = &publishedAt.Time
		}
		if errorMessage.Valid {
			event.ErrorMessage = &errorMessage.String
		}

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, appErrors.QueryError.Wrap(err, "rows iteration error")
	}

	return events, nil
}

func (o *outboxRepositoryImpl) MarkAsPublished(ctx context.Context, eventIDs []uuid.UUID) error {
	if len(eventIDs) == 0 {
		return nil
	}

	var executor interface {
		ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	}
	
	if tx, err := transaction.GetTx(ctx); err == nil {
		executor = tx
	} else {
		executor = o.db
	}

	placeholders := make([]string, len(eventIDs))
	args := make([]interface{}, len(eventIDs)+2)
	args[0] = repository.OutboxStatusPublished
	args[1] = time.Now()

	for i, eventID := range eventIDs {
		placeholders[i] = "?"
		args[i+2] = eventID
	}

	query := `
		UPDATE outbox 
		SET status = ?, published_at = ?
		WHERE event_id IN (` + strings.Join(placeholders, ", ") + `)
	`

	_, err := executor.ExecContext(ctx, query, args...)
	if err != nil {
		return appErrors.RepositoryError.Wrap(err, "failed to mark events as published")
	}

	return nil
}

func (o *outboxRepositoryImpl) MarkAsFailed(ctx context.Context, eventID uuid.UUID, errorMessage string) error {
	var executor interface {
		ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	}
	
	if tx, err := transaction.GetTx(ctx); err == nil {
		executor = tx
	} else {
		executor = o.db
	}

	query := `
		UPDATE outbox 
		SET status = ?, error_message = ?
		WHERE event_id = ?
	`

	_, err := executor.ExecContext(ctx, query, repository.OutboxStatusFailed, errorMessage, eventID)
	if err != nil {
		return appErrors.RepositoryError.Wrap(err, "failed to mark event as failed")
	}

	return nil
}

func (o *outboxRepositoryImpl) IncrementRetryCount(ctx context.Context, eventID uuid.UUID) error {
	var executor interface {
		ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	}
	
	if tx, err := transaction.GetTx(ctx); err == nil {
		executor = tx
	} else {
		executor = o.db
	}

	query := `
		UPDATE outbox 
		SET retry_count = retry_count + 1
		WHERE event_id = ?
	`

	_, err := executor.ExecContext(ctx, query, eventID)
	if err != nil {
		return appErrors.RepositoryError.Wrap(err, "failed to increment retry count")
	}

	return nil
}

