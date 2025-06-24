-- +goose Up
CREATE TABLE outbox (
);

-- -goose Down
DROP TABLE outbox;