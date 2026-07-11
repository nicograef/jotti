-- name: InsertKassensitzung :one
INSERT INTO kassensitzungen (datum, bezeichnung, status, created_at, updated_at)
VALUES ($1, $2, $3, NOW(), NOW()) RETURNING z_nr;

-- name: UpdateKassensitzung :exec
UPDATE kassensitzungen SET status = $2, updated_at = NOW() WHERE z_nr = $1;

-- name: GetKassensitzungStatusForShare :one
-- Status-Guard für Event-Writes: sperrt die Kassensitzungs-Zeile mit FOR SHARE,
-- damit der Statuswechsel auf 'wird_abgeschlossen' (UPDATE = FOR UPDATE) erst nach
-- Commit der laufenden Event-Transaktion durchkommt — und umgekehrt spätere Writes
-- den neuen Status sehen und abgelehnt werden.
SELECT status FROM kassensitzungen WHERE z_nr = $1 FOR SHARE;

-- name: GetOffeneKassensitzung :one
SELECT z_nr, datum, bezeichnung, status, created_at, updated_at
FROM kassensitzungen WHERE status = 'offen' LIMIT 1;

-- name: GetAktiveKassensitzung :one
-- Die aktive (nicht abgeschlossene) Kassensitzung, also 'offen' oder 'wird_abgeschlossen'.
-- Dank idx_kassensitzungen_eine_aktiv gibt es davon höchstens eine.
SELECT z_nr, datum, bezeichnung, status, created_at, updated_at
FROM kassensitzungen WHERE status <> 'abgeschlossen' LIMIT 1;

-- name: SetKassensitzungWirdAbgeschlossen :execrows
-- Erster Schritt des Abschlusses: setzt die Barriere. Der UPDATE (FOR UPDATE) wartet auf
-- alle noch laufenden Buchungs-Transaktionen (FOR SHARE) und ist idempotent, sodass ein
-- Wiederholungs-Aufruf im Zwischenstatus fortsetzt. Liefert 0, wenn die Sitzung bereits
-- abgeschlossen ist.
UPDATE kassensitzungen SET status = 'wird_abgeschlossen', updated_at = NOW()
WHERE z_nr = $1 AND status IN ('offen', 'wird_abgeschlossen');

-- name: SetKassensitzungOffen :execrows
-- Rücksetzen der Barriere nach einem Fehler im Abschluss (best effort), damit die Sitzung
-- nicht im Zwischenstatus hängen bleibt. Nur wirksam, solange sie noch nicht abgeschlossen ist.
UPDATE kassensitzungen SET status = 'offen', updated_at = NOW()
WHERE z_nr = $1 AND status = 'wird_abgeschlossen';

-- name: GetAllKassensitzungen :many
SELECT z_nr, datum, bezeichnung, status, created_at, updated_at
FROM kassensitzungen ORDER BY datum DESC, created_at DESC;

-- name: GetAbgeschlosseneKassensitzungen :many
-- Nur abgeschlossene Sitzungen für die Kassenberichte-Seite; der transiente
-- Status 'wird_abgeschlossen' bleibt außen vor. Sortierung wie GetAllKassensitzungen.
SELECT z_nr, datum, bezeichnung, status, created_at, updated_at
FROM kassensitzungen WHERE status = 'abgeschlossen' ORDER BY datum DESC, created_at DESC;

-- name: GetKassenbestand :one
-- Kassenbestand (Soll): Summe aus Anfangsbestand, Zahlungen, Warenrücknahmen, Geldtransits und Differenz-Buchungen.
-- Die kassenwirksame Warenrücknahme (stornierung-erteilt) gibt Bargeld zurück und mindert den Bestand;
-- geldneutrale Vorgänge (bestellung-korrigiert, bestellung-umgebucht) berühren den Kassenbestand nicht.
-- Die Differenz ist als Soll − Ist gebucht; ihre Bargeldwirkung ist Ist − Soll und wird deshalb
-- subtrahiert: Nach der Differenzbuchung entspricht der Soll-Bestand dem gezählten Ist-Bestand.
SELECT (
    COALESCE(SUM(kj_extract_eroeffnung_cents(type, data)), 0)::int
    + COALESCE(SUM(kj_extract_zahlung_cents(type, data)), 0)::int
    - COALESCE(SUM(kj_extract_stornierung_cents(type, data)), 0)::int
    + COALESCE(SUM(kj_extract_direktverkauf_cents(type, data)), 0)::int
    - COALESCE(SUM(kj_extract_direktverkauf_storno_cents(type, data)), 0)::int
    + COALESCE(SUM(kj_extract_geldtransit_cents(type, data)), 0)::int
    - COALESCE(SUM(kj_extract_differenz_cents(type, data)), 0)::int
)::int AS soll_bestand_cents
FROM kassenjournal
WHERE kassensitzung_nr = $1
  AND type IN (
    'kassensitzung-eroeffnet:v1',
    'zahlung-kassiert:v1',
    'stornierung-erteilt:v1',
    'direktverkauf-getaetigt:v1',
    'direktverkauf-storniert:v1',
    'geldtransit-gebucht:v1',
    'differenz-soll-ist-gebucht:v1'
  );
