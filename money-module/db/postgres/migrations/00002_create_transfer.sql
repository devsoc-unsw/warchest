-- +goose Up
CREATE TYPE transfer_types AS ENUM ('VIRTUAL', 'REAL');
CREATE TABLE transfer (
    id UUID PRIMARY KEY,
    sender UUID, 
    recipient UUID,
    description TEXT,
    transfer_type transfer_types
);

-- +goose Down
DROP TABLE transfer;
DROP TYPE transfer_types;