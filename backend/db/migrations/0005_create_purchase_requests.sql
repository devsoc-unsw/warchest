-- +goose Up

-- PURCHASE REQUEST STATUS
CREATE TYPE purchase_request_status AS ENUM (
    'draft',
    'pending',
    'approved',
    'rejected'
);

-- PURCHASE REQUESTS
CREATE TABLE purchase_requests (
    id BIGSERIAL PRIMARY KEY,
    
    created_by_user_id BIGINT NOT NULL,
    event_id BIGINT,
    portfolio_id BIGINT,
    title TEXT NOT NULL,
    description TEXT,
    expected_budget BIGINT,
    actual_budget BIGINT,
    
    status purchase_request_status
        NOT NULL
        DEFAULT 'draft',
    created_at TIMESTAMPTZ
        NOT NULL
        DEFAULT NOW(),
    updated_at TIMESTAMPTZ
        NOT NULL
        DEFAULT NOW(),
        
    -- exatly ONE must exist:
    -- event_id OR portfolio_id
    CHECK (
        (event_id IS NOT NULL AND portfolio_id IS NULL)
        OR
        (event_id IS NULL AND portfolio_id IS NOT NULL)
    )
    -- eventually:
    -- FOREIGN KEY (created_by_user_id) REFERENCES users(id),
    -- FOREIGN KEY (event_id) REFERENCES events(id),
    -- FOREIGN KEY (portfolio_id) REFERENCES portfolios(id)
);

-- +goose Down
DROP TABLE purchase_requests;
DROP TABLE purchase_request_status;