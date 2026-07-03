-- +goose Up
CREATE TYPE account_types AS ENUM ('VIRTUAL', 'REAL');
CREATE TABLE account (
    id UUID PRIMARY KEY,
    account_type account_types,
    account_name TEXT NOT NULL,
    owner_type account_types NOT NULL, 
    owner_id UUID NOT NULL,
    archived_at TIMESTAMPTZ DEFAULT NULL
);

-- +goose Down
DROP TABLE account;
DROP type account_types;