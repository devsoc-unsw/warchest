-- name: getPR :one
SELECT * FROM purchase_requests
WHERE id = $1 LIMIT 1;

-- name: ListPR :many
SELECT * FROM purchase_requests
ORDER BY id
LIMIT $1 OFFSET $2;

-- name: CreatePR :one
INSERT INTO purchase_requests (
    created_by_user_id, event_id, portfolio_id, title, description, expected_budget, actual_budget
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: UpdatePR :one
UPDATE purchase_requests
SET expected_budget = $2,
    actual_budget = $3
WHERE id = $1
RETURNING *;


-- name: DeletePR :
DELETE FROM purchase_requests
WHERE id = $1;