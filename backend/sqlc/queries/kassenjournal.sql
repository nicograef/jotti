-- name: WriteEvent :one
INSERT INTO kassenjournal (user_id, user_name, type, subject, version, data, timestamp, kassensitzung_nr)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id;

-- name: ReadEventsBySubject :many
SELECT id, user_id, user_name, version, type, subject, data, timestamp
FROM kassenjournal WHERE subject = $1 ORDER BY id ASC;

-- name: ReadDirektverkaufEvents :many
SELECT id, user_id, user_name, version, type, subject, data, timestamp
FROM kassenjournal
WHERE kassensitzung_nr = $1
  AND type IN ('direktverkauf-getaetigt:v1', 'direktverkauf-storniert:v1')
ORDER BY id ASC;

-- name: ReadEventsByKassensitzung :many
-- Alle Events einer Kassensitzung (Kassensitzungs-, Tisch-Session- und
-- Direktverkauf-Streams), nach id geordnet — Grundlage des DSFinV-K-Exports.
SELECT id, user_id, user_name, version, type, subject, data, timestamp
FROM kassenjournal WHERE kassensitzung_nr = $1 ORDER BY id ASC;

-- name: GetMaxVersion :one
SELECT COALESCE(MAX(version), 0)::int AS version FROM kassenjournal WHERE subject = $1;

-- name: GetDistinctTischSessionSubjects :many
-- Nur Tisch-Session-Subjects (enthalten "/tisch-"); Filterung in SQL fuer RebuildAllProjections.
SELECT DISTINCT subject FROM kassenjournal WHERE subject LIKE '%/tisch-%' ORDER BY subject ASC;
