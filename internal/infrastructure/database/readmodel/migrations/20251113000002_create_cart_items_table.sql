-- +goose Up
-- Create items table for read model
CREATE TABLE cart_items (
    id VARCHAR(36) PRIMARY KEY,
    cart_id VARCHAR(36) NOT NULL,
    quantity INT NOT NULL DEFAULT 1,
    price DECIMAL(8,2) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    UNIQUE KEY unique_cart_item (cart_id, id),
    FOREIGN KEY (cart_id) REFERENCES carts(id) ON DELETE CASCADE,
    INDEX idx_cart_id (cart_id),
    INDEX idx_item_id (id)
);

-- +goose Down
DROP TABLE IF EXISTS cart_items;