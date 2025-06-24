-- +goose Up
CREATE TABLE links (
    id SERIAL PRIMARY KEY,
    chat_id INTEGER REFERENCES chats(id) ON DELETE CASCADE,
    url TEXT UNIQUE NOT NULL
);

-- +goose Down
DROP TABLE links;