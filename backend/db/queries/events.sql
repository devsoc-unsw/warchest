-- name: GetEvent :one
SELECT * FROM events
WHERE id = $1 LIMIT 1;

-- name: ListEvent :many
SELECT * FROM events
WHERE id = $1 AND status = $2
ORDER BY event_time;

-- name: CreateEvent :one
INSERT INTO events (
    event_name,
    event_time,
    estimated_budget,
    location,
    description,
    created_by
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: StartEvent :one
UPDATE events
SET 
    status = 'IN_PROGRESS',
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateEvent :one
Update events
SET
    event_name = $2,
    event_time = $3,
    estimated_budget = $4,
    location = $5,
    description = $6
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: ArchiveEvent :one
UPDATE events
SET
    status = 'CLOSED'
    updated_at = NOW()
WHERE id = $1
RETURNING *;