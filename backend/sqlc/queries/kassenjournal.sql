-- name: WriteEvent :one
INSERT INTO kassenjournal (user_id, user_name, type, subject, version, data, timestamp, kassensitzung_nr)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id;

-- name: ReadEvent :one
SELECT id, user_id, user_name, version, type, subject, data, timestamp, kassensitzung_nr
FROM kassenjournal WHERE id = $1;

-- name: ReadEventsBySubject :many
SELECT id, user_id, user_name, version, type, subject, data, timestamp, kassensitzung_nr
FROM kassenjournal WHERE subject = $1 ORDER BY id ASC;

-- name: GetMaxVersion :one
SELECT COALESCE(MAX(version), 0)::int AS version FROM kassenjournal WHERE subject = $1;

-- name: GetDistinctSubjects :many
SELECT DISTINCT subject FROM kassenjournal ORDER BY subject ASC;
