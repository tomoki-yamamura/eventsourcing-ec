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
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/value"
	appErrors "github.com/tomoki-yamamura/eventsourcing-ec/internal/errors"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/database/transaction"
)

type outboxRepositoryImpl struct{}

func NewOutboxRepository() repository.OutboxRepository {
	return &outboxRepositoryImpl{}
}

func (o *outboxRepositoryImpl) SaveEvents(ctx context.Context, aggregateID uuid.UUID, events []event.Event) error {
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
			evt.GetAggregateType(),
			evt.GetEventType(),
			eventData,
			evt.GetVersion(),
			time.Now(),
			value.OutboxStatusPending,
		)
		if err != nil {
			return appErrors.RepositoryError.Wrap(err, "failed to save event to outbox")
		}
	}

	return nil
}

func (o *outboxRepositoryImpl) GetPendingEvents(ctx context.Context, limit int) ([]event.OutboxEvent, error) {
	tx, err := transaction.GetTx(ctx)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT id, event_id, aggregate_id, aggregate_type, event_type, event_data, version, 
		       created_at, published_at, status, retry_count, error_message
		FROM outbox 
		WHERE status = ? 
		ORDER BY created_at ASC 
		LIMIT ?
	`

	rows, err := tx.QueryContext(ctx, query, value.OutboxStatusPending, limit)
	if err != nil {
		return nil, appErrors.QueryError.Wrap(err, "failed to get pending events")
	}
	defer rows.Close()

	var events []event.OutboxEvent
	for rows.Next() {
		var outboxEvent event.OutboxEvent
		var publishedAt sql.NullTime
		var errorMessage sql.NullString

		err := rows.Scan(
			&outboxEvent.ID,
			&outboxEvent.EventID,
			&outboxEvent.AggregateID,
			&outboxEvent.AggregateType,
			&outboxEvent.EventType,
			&outboxEvent.EventData,
			&outboxEvent.Version,
			&outboxEvent.CreatedAt,
			&publishedAt,
			&outboxEvent.Status,
			&outboxEvent.RetryCount,
			&errorMessage,
		)
		if err != nil {
			return nil, appErrors.QueryError.Wrap(err, "failed to scan outbox event")
		}

		if publishedAt.Valid {
			outboxEvent.PublishedAt = &publishedAt.Time
		}
		if errorMessage.Valid {
			outboxEvent.ErrorMessage = &errorMessage.String
		}

		events = append(events, outboxEvent)
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

	tx, err := transaction.GetTx(ctx)
	if err != nil {
		return err
	}

	placeholders := make([]string, len(eventIDs))
	args := make([]any, len(eventIDs)+3)
	args[0] = value.OutboxStatusPublished
	args[1] = time.Now()
	args[2] = value.OutboxStatusProcessing

	for i, eventID := range eventIDs {
		placeholders[i] = "?"
		args[i+3] = eventID
	}

	query := `
		UPDATE outbox 
		SET status = ?, published_at = ?
		WHERE status = ?
		  AND event_id IN (` + strings.Join(placeholders, ", ") + `)
	`

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return appErrors.RepositoryError.Wrap(err, "failed to mark events as published")
	}

	return nil
}

func (o *outboxRepositoryImpl) MarkAsFailed(ctx context.Context, eventID uuid.UUID, errorMessage string) error {
	tx, err := transaction.GetTx(ctx)
	if err != nil {
		return err
	}

	query := `
		UPDATE outbox 
		SET status = ?, error_message = ?
		WHERE event_id = ?
		  AND status = ?
	`

	_, err = tx.ExecContext(ctx, query, value.OutboxStatusFailed, errorMessage, eventID, value.OutboxStatusProcessing)
	if err != nil {
		return appErrors.RepositoryError.Wrap(err, "failed to mark event as failed")
	}

	return nil
}

func (o *outboxRepositoryImpl) IncrementRetryCount(ctx context.Context, eventID uuid.UUID) error {
	tx, err := transaction.GetTx(ctx)
	if err != nil {
		return err
	}

	query := `
		UPDATE outbox 
		SET retry_count = retry_count + 1
		WHERE event_id = ?
		  AND status = ?
	`

	_, err = tx.ExecContext(ctx, query, eventID, value.OutboxStatusProcessing)
	if err != nil {
		return appErrors.RepositoryError.Wrap(err, "failed to increment retry count")
	}

	return nil
}

func (o *outboxRepositoryImpl) GetAndMarkProcessing(ctx context.Context, limit int, maxRetries int) ([]event.OutboxEvent, error) {
	tx, err := transaction.GetTx(ctx)
	if err != nil {
		return nil, err
	}

	// Step 1: SELECT ... FOR UPDATE SKIP LOCKED
	selectQuery := `
		SELECT id, event_id, aggregate_id, aggregate_type, event_type, event_data, version, 
		       created_at, published_at, status, retry_count, error_message
		FROM outbox 
		WHERE status = ? 
		  AND retry_count < ?
		ORDER BY created_at ASC 
		LIMIT ?
		FOR UPDATE SKIP LOCKED
	`

	rows, err := tx.QueryContext(ctx, selectQuery, value.OutboxStatusPending, maxRetries, limit)
	if err != nil {
		return nil, appErrors.QueryError.Wrap(err, "failed to get pending events")
	}
	defer rows.Close()

	var events []event.OutboxEvent
	var eventIDs []uuid.UUID

	for rows.Next() {
		var outboxEvent event.OutboxEvent
		var publishedAt sql.NullTime
		var errorMessage sql.NullString

		err := rows.Scan(
			&outboxEvent.ID,
			&outboxEvent.EventID,
			&outboxEvent.AggregateID,
			&outboxEvent.AggregateType,
			&outboxEvent.EventType,
			&outboxEvent.EventData,
			&outboxEvent.Version,
			&outboxEvent.CreatedAt,
			&publishedAt,
			&outboxEvent.Status,
			&outboxEvent.RetryCount,
			&errorMessage,
		)
		if err != nil {
			return nil, appErrors.QueryError.Wrap(err, "failed to scan outbox event")
		}

		if publishedAt.Valid {
			outboxEvent.PublishedAt = &publishedAt.Time
		}
		if errorMessage.Valid {
			outboxEvent.ErrorMessage = &errorMessage.String
		}

		events = append(events, outboxEvent)
		eventIDs = append(eventIDs, outboxEvent.EventID)
	}

	if err := rows.Err(); err != nil {
		return nil, appErrors.QueryError.Wrap(err, "rows iteration error")
	}

	// Step 2: Update to PROCESSING status
	if len(eventIDs) > 0 {
		placeholders := make([]string, len(eventIDs))
		args := make([]any, len(eventIDs)+2)
		args[0] = value.OutboxStatusProcessing
		args[1] = value.OutboxStatusPending

		for i, eventID := range eventIDs {
			placeholders[i] = "?"
			args[i+2] = eventID
		}

		updateQuery := `
			UPDATE outbox 
			SET status = ?
			WHERE status = ?
			  AND event_id IN (` + strings.Join(placeholders, ", ") + `)
		`

		_, err = tx.ExecContext(ctx, updateQuery, args...)
		if err != nil {
			return nil, appErrors.RepositoryError.Wrap(err, "failed to mark events as processing")
		}

		// Update the events status in memory to reflect the change
		for i := range events {
			events[i].Status = value.OutboxStatusProcessing
		}
	}

	return events, nil
}

func (o *outboxRepositoryImpl) MarkMaxRetriesExceededAsFailed(ctx context.Context, maxRetries int) error {
	tx, err := transaction.GetTx(ctx)
	if err != nil {
		return err
	}

	query := `
		UPDATE outbox 
		SET status = ?, error_message = ?
		WHERE status = ?
		  AND retry_count >= ?
	`

	_, err = tx.ExecContext(ctx, query, value.OutboxStatusFailed, "max retries exceeded", value.OutboxStatusPending, maxRetries)
	if err != nil {
		return appErrors.RepositoryError.Wrap(err, "failed to mark max retries exceeded events as failed")
	}

	return nil
}
