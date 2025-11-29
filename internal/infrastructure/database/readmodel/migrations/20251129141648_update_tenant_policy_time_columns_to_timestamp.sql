-- +goose Up
-- +goose StatementBegin
ALTER TABLE tenant_cart_abandoned_policies 
MODIFY COLUMN quiet_time_from TIMESTAMP NOT NULL,
MODIFY COLUMN quiet_time_to TIMESTAMP NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE tenant_cart_abandoned_policies 
MODIFY COLUMN quiet_time_from TIME NOT NULL,
MODIFY COLUMN quiet_time_to TIME NOT NULL;
-- +goose StatementEnd
