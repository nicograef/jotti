-- name: GetTable :one
SELECT id, name, status, created_at, updated_at
FROM tables WHERE id = $1 AND status != 'deleted';

-- name: GetAllTables :many
SELECT id, name, status, created_at, updated_at
FROM tables WHERE status != 'deleted' ORDER BY id ASC;

-- name: GetActiveTables :many
SELECT id, name, status, created_at, updated_at
FROM tables WHERE status = 'active' ORDER BY id ASC;

-- name: CreateTable :one
INSERT INTO tables (name, status, created_at, updated_at)
VALUES ($1, $2, $3, $4) RETURNING id;

-- name: UpdateTable :execresult
UPDATE tables SET name = $1, status = $2, updated_at = $3 WHERE id = $4;
