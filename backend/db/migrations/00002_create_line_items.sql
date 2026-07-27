-- +goose Up
CREATE TYPE line_item_status AS ENUM (
    -- purchase request phase
    'pr_draft',         -- created, not yet submitted
    'pr_pending',       -- submitted, awaiting decision
    'pr_approved',      -- moves on to reimbursement request
    'pr_rejected',      -- terminal (dead)
    -- reimbursement phase (entered after pr_approved)
    'reimb_draft',      -- can only get here from purchase request
    'reimb_pending',    -- submitted, awaiting decision
    'reimb_approved',   -- terminal (success)
    'reimb_rejected'    -- terminal (dead)
);

CREATE TABLE line_items (
    id BIGSERIAL PRIMARY KEY,

    estimated_quantity BIGINT NOT NULL,
    estimated_cost_per_item BIGINT NOT NULL,
	estimated_unit_cost BIGINT NOT NULL,
	pr_description TEXT,
	
	actual_quantity BIGINT NOT NULL,
    actual_cost_per_item BIGINT NOT NULL,
	actual_unit_cost BIGINT NOT NULL,
	reimb_description TEXT,

	-- IDs are not explicitly entered by the user
	reimbursement_request_ID BIGSERIAL, -- should this be BIGSERIAL?
	purchase_request_ID BIGSERIAL NOT NULL, -- should this be BIGSERIAL?

	status line_item_status,
    is_active BOOLEAN,

	-- logging
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE line_items;
