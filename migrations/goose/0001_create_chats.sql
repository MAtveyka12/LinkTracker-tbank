-- +goose Up
CREATE TABLE chats (
    id SERIAL PRIMARY KEY
);

-- +goose Down
DROP TABLE chats;