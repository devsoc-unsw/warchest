-- +goose Up
CREATE TYPE account_type AS ENUM ('VIRTUAL', 'REAL');
CREATE TYPE owner_type AS ENUM ('ORGANISATION', 'USER', 'PLATFORM', 'WORLD');

CREATE TABLE account (
    id UUID PRIMARY KEY,
    account_type account_type,
    account_name TEXT NOT NULL,
    owner_type owner_type NOT NULL, 
    owner_id UUID NOT NULL,
    archived_at TIMESTAMPTZ DEFAULT NULL
);

-- +goose Down
DROP TABLE account;
DROP type account_types;