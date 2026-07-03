-- name: GetUser :one
SELECT id, name, username, role, status, password_hash, onetime_password_hash, onetime_password_attempts, created_at, updated_at
FROM users WHERE id = $1 AND status != 'deleted';

-- name: GetUserByUsername :one
SELECT id, name, username, role, status, password_hash, onetime_password_hash, onetime_password_attempts, created_at, updated_at
FROM users WHERE username = $1 AND status != 'deleted';

-- name: GetAllUsers :many
SELECT id, name, username, role, status, created_at, updated_at
FROM users WHERE status != 'deleted' ORDER BY id ASC;

-- name: CreateUser :one
INSERT INTO users (name, username, role, status, password_hash, onetime_password_hash, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id;

-- name: UpdateUser :execresult
UPDATE users SET name = $1, username = $2, role = $3, status = $4, password_hash = $5, onetime_password_hash = $6, onetime_password_attempts = $7, updated_at = $8
WHERE id = $9;
