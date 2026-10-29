-- name: GetLineItemByID :one
SELECT * FROM line_items
WHERE id = $1 LIMIT 1;

-- name: GetLineItemsByPRID :many
SELECT * FROM line_items
WHERE purchase_request_ID = $1 AND is_active
ORDER BY id;

-- name: GetLineItemsByReimbID :many
SELECT * FROM line_items
WHERE reimbursement_request_ID = $1 AND is_active
ORDER BY ID;

-- name: ListLineItems :many
SELECT * FROM line_items
ORDER BY id
LIMIT $1;

-- name: CreateLineItem :one
INSERT INTO line_items (
    estimated_quantity, 
    estimated_cost_per_item,
    estimated_unit_cost,
    pr_description,
    purchase_request_ID

) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: UpdateLineItem :one
UPDATE line_items
SET 
    estimated_quantity = $2, 
    estimated_cost_per_item = $3, 
    estimated_unit_cost = $4, 
    pr_description = $5, 

    actual_quantity = $6,
    actual_cost_per_item = $7,
	actual_unit_cost = $8,
	reimb_description = $9,

    reimbursement_request_ID = $10,

    status = $11,
    is_active = $12,
    
    updated_at = NOW()
-- no row if anything changed since it was read
WHERE id = $1 AND updated_at = $13
RETURNING *;

-- name: ArchiveLineItem :exec
UPDATE line_items
SET 
    is_active = FALSE,
    updated_at = NOW()
WHERE id = $1;

-- name: TransitionLineItem :one
-- no row if status changed since it was read
UPDATE line_items
SET
    status = @to_status,
    updated_at = NOW()
WHERE id = @id AND status = @from_status AND is_active
RETURNING *;
