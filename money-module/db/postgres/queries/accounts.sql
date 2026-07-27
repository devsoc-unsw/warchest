-- name: CreateAccount :exec
INSERT INTO account (id, account_type, account_name, owner_type, owner_id)
VALUES (
    $1, $2, $3, $4, $5
);

-- name: ReadAccount :one
SELECT * FROM account
WHERE id = $1;

-- name: ArchiveAccount :exec
UPDATE account SET archived_at = NOW() WHERE id = $1;