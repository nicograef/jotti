-- name: InsertKassensitzung :one
INSERT INTO kassensitzungen (datum, bezeichnung, status, created_at, updated_at)
VALUES ($1, $2, $3, NOW(), NOW()) RETURNING z_nr;

-- name: UpdateKassensitzungStatus :exec
UPDATE kassensitzungen SET status = $2, updated_at = NOW() WHERE z_nr = $1;

-- name: GetOffeneKassensitzung :one
SELECT z_nr, datum, bezeichnung, status, created_at, updated_at
FROM kassensitzungen WHERE status = 'offen' LIMIT 1;

-- name: GetKassenbestand :one
-- Kassenbestand (Soll): Summe aus Anfangsbestand, Zahlungen, Auszahlungen, Kassenbewegungen und Differenz-Buchungen.
SELECT COALESCE(SUM(CASE
    WHEN type = 'anfangsbestand-gesetzt:v1'
        THEN (data->>'betragCents')::INT
    WHEN type = 'zahlung-kassiert:v1'
        THEN (data->>'gesamtZahlungCents')::INT
    WHEN type = 'auszahlung-geleistet:v1'
        THEN -(data->>'betragCents')::INT
    WHEN type = 'kassenbewegung-gebucht:v1' AND data->>'art' = 'privateinlage'
        THEN (data->>'betragCents')::INT
    WHEN type = 'kassenbewegung-gebucht:v1' AND data->>'art' IN ('privatentnahme', 'geldtransit')
        THEN -(data->>'betragCents')::INT
    WHEN type = 'differenz-soll-ist-gebucht:v1'
        THEN (data->>'betragCents')::INT
    ELSE 0
END), 0)::int AS soll_bestand_cents
FROM kassenjournal
WHERE kassensitzung_nr = $1
  AND type IN (
    'anfangsbestand-gesetzt:v1',
    'zahlung-kassiert:v1',
    'auszahlung-geleistet:v1',
    'kassenbewegung-gebucht:v1',
    'differenz-soll-ist-gebucht:v1'
  );
