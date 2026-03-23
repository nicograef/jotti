-- name: InsertKassensitzung :one
INSERT INTO kassensitzungen (datum, bezeichnung, status, created_at, updated_at)
VALUES ($1, $2, $3, NOW(), NOW()) RETURNING z_nr;

-- name: UpdateKassensitzungStatus :exec
UPDATE kassensitzungen SET status = $2, updated_at = NOW() WHERE z_nr = $1;

-- name: GetOffeneKassensitzung :one
SELECT z_nr, datum, bezeichnung, status, created_at, updated_at
FROM kassensitzungen WHERE status = 'offen' LIMIT 1;

-- name: GetKassensitzungByZNr :one
SELECT z_nr, datum, bezeichnung, status, created_at, updated_at
FROM kassensitzungen WHERE z_nr = $1;

-- name: GetNextZNr :one
SELECT COALESCE(MAX(z_nr), 0) + 1 AS next_z_nr FROM kassensitzungen;

-- name: DeleteAllKassensitzungen :exec
DELETE FROM kassensitzungen;
