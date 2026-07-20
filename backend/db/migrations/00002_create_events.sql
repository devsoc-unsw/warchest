-- +goose Up
CREATE TYPE event_status AS ENUM ('DRAFT', 'IN_PROGRESS', 'CLOSED');

CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id INT,
    CONSTRAINT fk_user
        FOREIGN KEY (user_id)
        REFERENCES users(id),
    event_name VARCHAR(255) NOT NULL,
    event_time TIMESTAMPTZ,
    budget NUMERIC(15, 2),
    status event_status NOT NULL, 
    location VARCHAR(255),
    description VARCHAR(255),
    society_id UUID NOT NULL,
    created_by UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE events;