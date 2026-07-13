-- +goose Up
CREATE TYPE transfer_type AS ENUM ('VIRTUAL', 'REAL');
CREATE TABLE transfers (
    id UUID PRIMARY KEY,
    sender UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE, 
    recipient UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    description TEXT,
    transfer_type transfer_type NOT NULL
);

-- +goose Down
DROP TABLE transfers;
DROP TYPE transfer_type;