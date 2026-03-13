-- name: WriteEvent :one
INSERT INTO events (user_id, user_name, type, subject, version, data, timestamp)
VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id;

-- name: ReadEvent :one
SELECT id, user_id, user_name, version, type, subject, data, timestamp
FROM events WHERE id = $1;

-- name: ReadEventsBySubject :many
SELECT id, user_id, user_name, version, type, subject, data, timestamp
FROM events WHERE subject = $1 ORDER BY id ASC;

-- name: GetMaxVersion :one
SELECT COALESCE(MAX(version), 0)::int AS version FROM events WHERE subject = $1;
