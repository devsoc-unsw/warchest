-- +goose Up
CREATE TYPE budget_type AS ENUM ('EVENT', 'PORTFOLIO');
CREATE TYPE event_status AS ENUM ('DRAFT', 'IN_PROGRESS', 'CLOSED');

CREATE TABLE budgets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type budget_type NOT NULL,
    society_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    money_budget_id UUID UNIQUE,
    team_id TEXT,
    event_time TIMESTAMPTZ,
    location TEXT,
    status event_status,
    created_by UUID,
    archived_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Imports of the same AMBUSH Team must resolve to the same Portfolio.
    CONSTRAINT budget_portfolio_team_unique UNIQUE (society_id, team_id)
);

CREATE INDEX budgets_society_type_idx ON budgets (society_id, type);
CREATE INDEX budgets_event_status_idx ON budgets (society_id, status)
    WHERE type = 'EVENT' AND archived_at IS NULL;

-- +goose Down
DROP TABLE budgets;
DROP TYPE event_status;
DROP TYPE budget_type;
