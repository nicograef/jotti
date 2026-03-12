-- name: WriteEvent :one
INSERT INTO events (user_id, user_name, type, subject, version, data, timestamp)
VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id;

-- name: ReadEvent :one
SELECT id, user_id, user_name, version, type, subject, data, timestamp
FROM events WHERE id = $1;

-- name: ReadEventsBySubject :many
SELECT id, user_id, user_name, version, type, subject, data, timestamp
FROM events WHERE subject = $1 ORDER BY id ASC;

-- name: ReadEventsSinceID :many
SELECT id, user_id, user_name, version, type, subject, data, timestamp
FROM events WHERE subject = $1 AND id >= $2 ORDER BY id ASC;

-- name: GetLastSnapshotID :one
SELECT COALESCE(MAX(id), 0)::int AS id FROM events WHERE subject = $1 AND type = $2;

-- name: GetMaxVersion :one
SELECT COALESCE(MAX(version), 0)::int AS version FROM events WHERE subject = $1;

-- name: ReadEventsWithSnapshot :many
WITH last_snapshot AS (
    SELECT COALESCE(MAX(id), 0) AS id 
    FROM events 
    WHERE events.subject = $1 AND events.type = $2
)
SELECT e.id, e.user_id, e.user_name, e.version, e.type, e.subject, e.data, e.timestamp
FROM events e, last_snapshot ls
WHERE e.subject = $1 AND e.id >= ls.id
ORDER BY e.id ASC;
