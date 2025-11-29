-- +goose Up
-- +goose StatementBegin
CREATE TABLE tenant_cart_abandoned_policies (
    id VARCHAR(36) PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    abandoned_minutes INT NOT NULL,
    quiet_time_from TIMESTAMP NOT NULL,
    quiet_time_to TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    version INT NOT NULL DEFAULT 1,
    INDEX idx_tenant_cart_abandoned_policies_id (id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS tenant_cart_abandoned_policies;
-- +goose StatementEnd
