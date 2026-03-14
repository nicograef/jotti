-- name: GetDashboardStats :one
-- Dashboard: Gesamtumsatz (Zahlungen), Bestellungen, Stornierungen — alle Daten (kein Zeitraumfilter).
SELECT
    COALESCE(SUM(CASE WHEN type = 'tisch.zahlung-registriert:v1'
        THEN (data->>'gesamtZahlungCents')::int END), 0)::int AS gesamt_umsatz_cents,
    COALESCE(SUM(CASE WHEN type = 'tisch.bestellung-aufgegeben:v1'
        THEN (data->>'gesamtPreisCents')::int END), 0)::int AS gesamt_bestellungen_cents,
    COALESCE(SUM(CASE WHEN type = 'tisch.produkte-storniert:v1'
        THEN (data->>'gesamtStornierungCents')::int END), 0)::int AS gesamt_stornierungen_cents,
    COALESCE(COUNT(CASE WHEN type = 'tisch.bestellung-aufgegeben:v1' THEN 1 END), 0)::int AS anzahl_bestellungen,
    COALESCE(COUNT(CASE WHEN type = 'tisch.produkte-storniert:v1' THEN 1 END), 0)::int AS anzahl_stornierungen
FROM events
WHERE type IN (
    'tisch.bestellung-aufgegeben:v1',
    'tisch.zahlung-registriert:v1',
    'tisch.produkte-storniert:v1'
);

-- name: GetOffeneTische :one
-- Dashboard: Anzahl Tische mit offenem Saldo > 0.
SELECT COALESCE(COUNT(*), 0)::int AS anzahl
FROM table_state WHERE saldo_cents > 0;

-- name: GetAbrechnungStats :one
-- Tagesabrechnung: Aggregierte Kennzahlen im Abrechnungszeitraum.
SELECT
    COALESCE(SUM(CASE WHEN type = 'tisch.zahlung-registriert:v1'
        THEN (data->>'gesamtZahlungCents')::int END), 0)::int AS gesamt_umsatz_cents,
    COALESCE(SUM(CASE WHEN type = 'tisch.bestellung-aufgegeben:v1'
        THEN (data->>'gesamtPreisCents')::int END), 0)::int AS gesamt_bestellungen_cents,
    COALESCE(SUM(CASE WHEN type = 'tisch.produkte-storniert:v1'
        THEN (data->>'gesamtStornierungCents')::int END), 0)::int AS gesamt_stornierungen_cents,
    COALESCE(COUNT(CASE WHEN type = 'tisch.bestellung-aufgegeben:v1' THEN 1 END), 0)::int AS anzahl_bestellungen,
    COALESCE(COUNT(CASE WHEN type = 'tisch.produkte-storniert:v1' THEN 1 END), 0)::int AS anzahl_stornierungen
FROM events
WHERE type IN (
    'tisch.bestellung-aufgegeben:v1',
    'tisch.zahlung-registriert:v1',
    'tisch.produkte-storniert:v1'
)
AND timestamp >= @von AND timestamp < @bis;

-- name: GetOffeneSaldi :one
-- Tagesabrechnung: Summe aller offenen Saldi (zeitraumunabhängig, aktueller Ist-Zustand).
SELECT COALESCE(SUM(saldo_cents), 0)::int AS offene_saldi_cents
FROM table_state WHERE saldo_cents > 0;

-- name: GetUmsatzProServicekraft :many
-- Tagesabrechnung: Zahlungen gruppiert nach Servicekraft im Zeitraum.
SELECT
    user_id,
    user_name,
    COALESCE(SUM((data->>'gesamtZahlungCents')::int), 0)::int AS zahlungen_cents,
    COUNT(*)::int AS anzahl_zahlungen
FROM events
WHERE type = 'tisch.zahlung-registriert:v1'
AND timestamp >= @von AND timestamp < @bis
GROUP BY user_id, user_name
ORDER BY zahlungen_cents DESC;

-- name: GetStornierungen :many
-- Tagesabrechnung: Stornierungsevents mit Tischname im Zeitraum.
-- Events contain fat positions (produktName, varianteName, einzelpreis, menge) — parse in Go.
SELECT
    e.timestamp,
    t.id AS tisch_id,
    t.name AS tisch_name,
    e.user_id,
    e.user_name,
    e.data
FROM events e
JOIN tische t ON t.id = CAST(SPLIT_PART(e.subject, ':', 2) AS INTEGER)
WHERE e.type = 'tisch.produkte-storniert:v1'
AND e.timestamp >= @von AND e.timestamp < @bis
ORDER BY e.timestamp DESC;

-- name: GetUmsatzProTisch :many
-- Tagesabrechnung: Zahlungen gruppiert nach Tisch im Zeitraum.
SELECT
    t.id AS tisch_id,
    t.name AS tisch_name,
    COALESCE(SUM((e.data->>'gesamtZahlungCents')::int), 0)::int AS zahlungen_cents,
    COUNT(*)::int AS anzahl_zahlungen
FROM events e
JOIN tische t ON t.id = CAST(SPLIT_PART(e.subject, ':', 2) AS INTEGER)
WHERE e.type = 'tisch.zahlung-registriert:v1'
AND e.timestamp >= @von AND e.timestamp < @bis
GROUP BY t.id, t.name
ORDER BY zahlungen_cents DESC;
