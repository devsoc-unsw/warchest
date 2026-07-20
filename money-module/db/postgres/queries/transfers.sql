-- name: CreateTransfer :exec
INSERT INTO transfer (id, sender, recipient, description, transfer_type)
VALUES (
    $1, $2, $3, $4, $5
);

-- name: ReadTransfer :one
SELECT * FROM transfer
WHERE id = $1;