-- +goose Up
CREATE TABLE users (
    id UUID PRIMARY KEY
);

-- +goose Down
DROP TABLE users;
