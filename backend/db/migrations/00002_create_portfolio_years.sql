-- +goose Up
CREATE TABLE portfolio_years (
    society_id TEXT NOT NULL,
    team_id TEXT NOT NULL,
    year INTEGER NOT NULL,
    archived_at TIMESTAMPTZ,

    PRIMARY KEY (society_id, team_id, year),
    FOREIGN KEY (society_id, team_id) REFERENCES budgets (society_id, team_id)
);

-- +goose Down
DROP TABLE portfolio_years;
