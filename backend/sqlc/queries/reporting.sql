-- name: GetOffeneTische :one
-- Dashboard: Anzahl Tische mit offenem Saldo > 0.
SELECT COALESCE(COUNT(*), 0)::int AS anzahl
FROM tisch_session_state WHERE saldo_cents > 0;

-- name: GetReportingStats :one
-- Reporting: Aggregierte Kennzahlen fuer eine Kassensitzung.
SELECT
    COALESCE(SUM(CASE WHEN type = 'zahlung-kassiert:v1'
        THEN (data->>'gesamtZahlungCents')::int END), 0)::int
        - COALESCE(SUM(CASE WHEN type = 'auszahlung-geleistet:v1'
        THEN (data->>'betragCents')::int END), 0)::int AS gesamt_umsatz_cents,
    COALESCE(SUM(CASE WHEN type = 'auszahlung-geleistet:v1'
        THEN (data->>'betragCents')::int END), 0)::int AS gesamt_auszahlungen_cents,
    COALESCE(SUM(CASE WHEN type = 'bestellung-aufgenommen:v1'
        THEN (data->>'gesamtPreisCents')::int END), 0)::int AS gesamt_bestellungen_cents,
    COALESCE(SUM(CASE WHEN type = 'stornierung-erteilt:v1'
        THEN (data->>'gesamtStornierungCents')::int END), 0)::int AS gesamt_stornierungen_cents,
    COALESCE(COUNT(CASE WHEN type = 'bestellung-aufgenommen:v1' THEN 1 END), 0)::int AS anzahl_bestellungen,
    COALESCE(COUNT(CASE WHEN type = 'stornierung-erteilt:v1' THEN 1 END), 0)::int AS anzahl_stornierungen
FROM kassenjournal
WHERE type IN (
    'bestellung-aufgenommen:v1',
    'zahlung-kassiert:v1',
    'stornierung-erteilt:v1',
    'auszahlung-geleistet:v1'
)
AND kassensitzung_nr = @kassensitzung_nr;

-- name: GetOffeneSaldi :one
-- Tagesabrechnung: Summe aller offenen Saldi (zeitraumunabhängig, aktueller Ist-Zustand).
SELECT COALESCE(SUM(saldo_cents), 0)::int AS offene_saldi_cents
FROM tisch_session_state WHERE saldo_cents > 0;

-- name: GetUmsatzProServicekraft :many
-- Tagesabrechnung: Zahlungen und Auszahlungen gruppiert nach Servicekraft pro Kassensitzung.
-- MAX(user_name) nimmt den lexikographisch letzten Namen bei Namensaenderungen.
SELECT
    user_id,
    MAX(user_name) AS user_name,
    COALESCE(SUM(CASE WHEN type = 'zahlung-kassiert:v1'
        THEN (data->>'gesamtZahlungCents')::int END), 0)::int AS zahlungen_cents,
    COALESCE(SUM(CASE WHEN type = 'auszahlung-geleistet:v1'
        THEN (data->>'betragCents')::int END), 0)::int AS auszahlungen_cents,
    COUNT(CASE WHEN type = 'zahlung-kassiert:v1' THEN 1 END)::int AS anzahl_zahlungen
FROM kassenjournal
WHERE type IN ('zahlung-kassiert:v1', 'auszahlung-geleistet:v1')
AND kassensitzung_nr = @kassensitzung_nr
GROUP BY user_id
ORDER BY zahlungen_cents DESC;

-- name: GetStornierungen :many
-- Reporting: Stornierungsevents mit Tischname pro Kassensitzung.
-- Events contain fat positions (produktName, varianteName, einzelpreis, menge) — parse in Go.
SELECT
    e.timestamp,
    tss.tisch_id,
    t.name AS tisch_name,
    e.user_id,
    e.user_name,
    e.data
FROM kassenjournal e
JOIN tisch_session_state tss ON tss.subject = e.subject
JOIN tische t ON t.id = tss.tisch_id
WHERE e.type = 'stornierung-erteilt:v1'
AND e.kassensitzung_nr = @kassensitzung_nr
ORDER BY e.timestamp DESC;

-- name: GetUmsatzProTisch :many
-- Tagesabrechnung: Zahlungen und Auszahlungen gruppiert nach Tisch pro Kassensitzung.
SELECT
    tss.tisch_id,
    t.name AS tisch_name,
    COALESCE(SUM(CASE WHEN e.type = 'zahlung-kassiert:v1'
        THEN (e.data->>'gesamtZahlungCents')::int END), 0)::int AS zahlungen_cents,
    COALESCE(SUM(CASE WHEN e.type = 'auszahlung-geleistet:v1'
        THEN (e.data->>'betragCents')::int END), 0)::int AS auszahlungen_cents,
    COUNT(CASE WHEN e.type = 'zahlung-kassiert:v1' THEN 1 END)::int AS anzahl_zahlungen
FROM kassenjournal e
JOIN tisch_session_state tss ON tss.subject = e.subject
JOIN tische t ON t.id = tss.tisch_id
WHERE e.type IN ('zahlung-kassiert:v1', 'auszahlung-geleistet:v1')
AND e.kassensitzung_nr = @kassensitzung_nr
GROUP BY tss.tisch_id, t.name
ORDER BY zahlungen_cents DESC;

-- name: GetAusstehendAuszahlungen :one
-- Aktuelle Schulden: Summe aller negativen Tischsaldi (zeitraumunabhaengig).
SELECT COALESCE(SUM(ABS(saldo_cents)), 0)::int AS ausstehend_auszahlungen_cents
FROM tisch_session_state
WHERE saldo_cents < 0;

-- name: GetEigeneUebersicht :one
-- Service-Dashboard: Eigene KPIs der eingeloggten Servicekraft pro Kassensitzung.
SELECT
    COALESCE(COUNT(CASE WHEN type = 'bestellung-aufgenommen:v1' THEN 1 END), 0)::int AS anzahl_bestellungen,
    COALESCE(SUM(CASE WHEN type = 'bestellung-aufgenommen:v1'
        THEN (data->>'gesamtPreisCents')::int END), 0)::int AS bestellungen_cents,
    COALESCE(COUNT(CASE WHEN type = 'zahlung-kassiert:v1' THEN 1 END), 0)::int AS anzahl_zahlungen,
    COALESCE(SUM(CASE WHEN type = 'zahlung-kassiert:v1'
        THEN (data->>'gesamtZahlungCents')::int END), 0)::int AS zahlungen_cents
FROM kassenjournal
WHERE type IN ('bestellung-aufgenommen:v1', 'zahlung-kassiert:v1')
AND user_id = @user_id
AND kassensitzung_nr = @kassensitzung_nr;
