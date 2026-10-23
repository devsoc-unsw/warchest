-- name: GetBudget :one
SELECT * FROM budgets
WHERE id = $1 AND society_id = $2 AND archived_at IS NULL;

-- name: ListBudgets :many
SELECT * FROM budgets
WHERE society_id = $1 AND archived_at IS NULL
ORDER BY created_at DESC, id;

-- name: CreateEvent :one
INSERT INTO budgets (type, society_id, name, description, event_time, location, status, created_by)
VALUES ('EVENT', $1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: ListEvents :many
SELECT * FROM budgets
WHERE type = 'EVENT' AND society_id = $1 AND status = $2 AND archived_at IS NULL
ORDER BY event_time NULLS LAST, id;

-- name: UpdateEvent :one
UPDATE budgets
SET name = $3, description = $4, event_time = $5, location = $6, updated_at = NOW()
WHERE id = $1 AND society_id = $2 AND type = 'EVENT'
    AND status <> 'CLOSED' AND archived_at IS NULL
RETURNING *;

-- name: StartEvent :one
UPDATE budgets SET status = 'IN_PROGRESS', updated_at = NOW()
WHERE id = $1 AND society_id = $2 AND type = 'EVENT'
    AND status = 'DRAFT' AND archived_at IS NULL
RETURNING *;

-- name: CloseEvent :one
UPDATE budgets SET status = 'CLOSED', updated_at = NOW()
WHERE id = $1 AND society_id = $2 AND type = 'EVENT'
    AND status IN ('DRAFT', 'IN_PROGRESS') AND archived_at IS NULL
RETURNING *;

-- name: ArchiveEvent :one
UPDATE budgets SET archived_at = NOW(), updated_at = NOW()
WHERE id = $1 AND society_id = $2 AND type = 'EVENT'
    AND status = 'CLOSED' AND archived_at IS NULL
RETURNING *;

-- TODO(AMBUSH): implement Portfolio import/synchronisation after its contract is confirmed.

-- name: GetPortfolioByTeam :one
SELECT * FROM budgets
WHERE type = 'PORTFOLIO' AND society_id = $1 AND team_id = $2 AND archived_at IS NULL;

-- TODO(Money Module): confirm Budget identity mapping and linking behaviour.
