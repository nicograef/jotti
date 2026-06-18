-- name: InsertKassensitzung :one
INSERT INTO kassensitzungen (datum, bezeichnung, status, created_at, updated_at)
VALUES ($1, $2, $3, NOW(), NOW()) RETURNING z_nr;

-- name: UpdateKassensitzung :exec
UPDATE kassensitzungen SET status = $2, updated_at = NOW() WHERE z_nr = $1;

-- name: GetOffeneKassensitzung :one
SELECT z_nr, datum, bezeichnung, status, created_at, updated_at
FROM kassensitzungen WHERE status = 'offen' LIMIT 1;

-- name: GetAllKassensitzungen :many
SELECT z_nr, datum, bezeichnung, status, created_at, updated_at
FROM kassensitzungen ORDER BY datum DESC, created_at DESC;

-- name: GetKassenbestand :one
-- Kassenbestand (Soll): Summe aus Anfangsbestand, Zahlungen, Auszahlungen, Geldtransits und Differenz-Buchungen.
SELECT (
    COALESCE(SUM(kj_extract_eroeffnung_cents(type, data)), 0)::int
    + COALESCE(SUM(kj_extract_zahlung_cents(type, data)), 0)::int
    - COALESCE(SUM(kj_extract_auszahlung_cents(type, data)), 0)::int
    + COALESCE(SUM(kj_extract_direktverkauf_cents(type, data)), 0)::int
    - COALESCE(SUM(kj_extract_direktverkauf_storno_cents(type, data)), 0)::int
    + COALESCE(SUM(kj_extract_geldtransit_cents(type, data)), 0)::int
    + COALESCE(SUM(kj_extract_differenz_cents(type, data)), 0)::int
)::int AS soll_bestand_cents
FROM kassenjournal
WHERE kassensitzung_nr = $1
  AND type IN (
    'kassensitzung-eroeffnet:v1',
    'zahlung-kassiert:v1',
    'auszahlung-geleistet:v1',
    'direktverkauf-getaetigt:v1',
    'direktverkauf-storniert:v1',
    'geldtransit-gebucht:v1',
    'differenz-soll-ist-gebucht:v1'
  );
