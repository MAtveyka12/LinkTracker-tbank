-- links.sql
--liquibase formatted sql


--changeset matvey:create-links
CREATE TABLE links (
    id SERIAL PRIMARY KEY,
    chat_id INTEGER REFERENCES chats(id) ON DELETE CASCADE,
    url TEXT UNIQUE NOT NULL
);

