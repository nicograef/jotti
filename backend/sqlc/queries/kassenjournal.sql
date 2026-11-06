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

-- name: GetDirektverkauf :one
-- Der Verkauf zu einer verkaufId über alle Kassensitzungen; idx_kassenjournal_verkauf_id
-- hält ihn eindeutig.
SELECT id, user_id, user_name, version, type, subject, data, timestamp
FROM kassenjournal
WHERE type = 'direktverkauf-getaetigt:v1'
  AND data->>'verkaufId' = sqlc.arg(verkauf_id)::text;

-- name: ReadPfandrueckgabeEvents :many
-- Rückgaben und ihre Stornierungen eines Tisches; tischId 0 ist der Direktverkauf.
-- Beide Event-Typen tragen tischId selbst.
SELECT id, user_id, user_name, version, type, subject, data, timestamp
FROM kassenjournal
WHERE kassensitzung_nr = sqlc.arg(kassensitzung_nr)
  AND type IN ('pfand-zurueckgegeben:v1', 'pfandrueckgabe-storniert:v1')
  AND (data->>'tischId')::int = sqlc.arg(tisch_id)::int
ORDER BY id ASC;

-- name: GetPfandrueckgabe :one
-- Die Rückgabe zu einer rueckgabeId über alle Kassensitzungen; idx_kassenjournal_rueckgabe_id
-- hält sie eindeutig.
SELECT id, user_id, user_name, version, type, subject, data, timestamp
FROM kassenjournal
WHERE type = 'pfand-zurueckgegeben:v1'
  AND data->>'rueckgabeId' = sqlc.arg(rueckgabe_id)::text;

-- name: GetPfandrueckgabeStornierung :one
-- Die Pfandrückgabe-Stornierung zu einer stornierungId über alle Kassensitzungen;
-- idx_kassenjournal_pfand_stornierung_id hält sie eindeutig.
SELECT id, user_id, user_name, version, type, subject, data, timestamp
FROM kassenjournal
WHERE type = 'pfandrueckgabe-storniert:v1'
  AND data->>'stornierungId' = sqlc.arg(stornierung_id)::text;

-- name: ReadEventsByKassensitzung :many
-- Alle Events einer Kassensitzung (Kassensitzungs-, Tisch-Session-, Direktverkauf-
-- und Pfandrückgabe-Streams), nach id geordnet — Grundlage des DSFinV-K-Exports.
-- Der LEFT JOIN auf die Signaturaufträge liefert je Event den Signatur-Stand:
-- kein Auftrag = nicht signaturpflichtig (keine Projektion zur Lesezeit).
SELECT k.id, k.user_id, k.user_name, k.version, k.type, k.subject, k.data, k.timestamp,
       a.process_type, a.transaktion_nummer, a.signatur_zaehler, a.tse_seriennummer,
       a.log_time_start, a.log_time_end, a.signatur, a.qr_code_data
FROM kassenjournal k
LEFT JOIN tse_signaturauftraege a ON a.event_id = k.id
WHERE k.kassensitzung_nr = $1 ORDER BY k.id ASC;

-- name: ReadKassensitzungEvents :many
-- Alle Events einer Kassensitzung ohne Signatur-JOIN: der events-only-Leseweg
-- für die Tagesabschluss-Aggregation (Signaturen braucht nur der Export).
SELECT id, user_id, user_name, version, type, subject, data, timestamp
FROM kassenjournal
WHERE kassensitzung_nr = $1
ORDER BY id ASC;

-- name: GetMaxVersion :one
SELECT COALESCE(MAX(version), 0)::int AS version FROM kassenjournal WHERE subject = $1;

-- name: GetDistinctTischSessionSubjects :many
-- Nur Tisch-Session-Subjects (enthalten "/tisch-"); Filterung in SQL für RebuildAllProjections.
SELECT DISTINCT subject FROM kassenjournal WHERE subject LIKE '%/tisch-%' ORDER BY subject ASC;
