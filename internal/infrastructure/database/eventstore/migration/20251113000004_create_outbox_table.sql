-- +goose Up
-- +goose StatementBegin
CREATE TABLE
    outbox (
        id BIGINT PRIMARY KEY AUTO_INCREMENT,
        event_id CHAR(36) NOT NULL,
        aggregate_id CHAR(36) NOT NULL,
        aggregate_type VARCHAR(255) NOT NULL,
        event_type VARCHAR(255) NOT NULL,
        event_data JSON NOT NULL,
        version INT NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        published_at TIMESTAMP NULL,
        status ENUM ('PENDING', 'PUBLISHED', 'FAILED') NOT NULL DEFAULT 'PENDING',
        retry_count INT NOT NULL DEFAULT 0,
        error_message TEXT NULL,
        INDEX idx_status_created_at_id (status, created_at, id),
        INDEX idx_aggregate_id (aggregate_id),
        UNIQUE INDEX unique_event_id (event_id)
    );

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE outbox;

-- +goose StatementEnd