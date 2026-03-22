-- name: UpsertKassensitzungState :exec
INSERT INTO kassensitzung_state (subject, z_nr, datum, status, last_event_id, last_event_version)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (subject) DO UPDATE SET
    z_nr = $2,
    datum = $3,
    status = $4,
    last_event_id = $5,
    last_event_version = $6;

-- name: GetOffeneKassensitzung :one
SELECT subject, z_nr, datum, status, last_event_id, last_event_version
FROM kassensitzung_state WHERE status = 'offen' LIMIT 1;

-- name: GetKassensitzungBySubject :one
SELECT subject, z_nr, datum, status, last_event_id, last_event_version
FROM kassensitzung_state WHERE subject = $1;

-- name: GetNextZNr :one
SELECT COALESCE(MAX(z_nr), 0) + 1 AS next_z_nr FROM kassensitzung_state;

-- name: DeleteAllKassensitzungState :exec
DELETE FROM kassensitzung_state;
