-- +goose Up
CREATE TYPE account_type AS ENUM ('virtual', 'real');
CREATE TYPE owner_type AS ENUM ('organisation', 'user', 'platform', 'world');

CREATE TABLE accounts (
    id UUID PRIMARY KEY,
    account_type account_type NOT NULL,
    account_name TEXT NOT NULL,
    owner_type owner_type NOT NULL, 
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    archived_at TIMESTAMPTZ DEFAULT NULL
);

-- +goose Down
DROP TABLE accounts;
DROP type account_type;
DROP type owner_type;