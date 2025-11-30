-- +goose Up
-- Create carts table for read model
CREATE TABLE carts (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    tenant_id VARCHAR(36) NOT NULL,
    status ENUM('OPEN', 'SUBMITTED', 'CLOSED', 'ABANDONED') NOT NULL DEFAULT 'OPEN',
    total_amount DECIMAL(10,2) NOT NULL DEFAULT 0.0,
    item_count INT NOT NULL DEFAULT 0,
    purchased_at TIMESTAMP NULL,
    version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_user_id (user_id),
    INDEX idx_tenant_id (tenant_id),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
);

-- +goose Down
DROP TABLE IF EXISTS carts;