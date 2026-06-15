-- name: GetOffeneTischeDetails :many
-- Live-Dashboard: Offene Tische der offenen Kassensitzung mit Name und aktuellem Saldo.
SELECT ts.tisch_id, t.name AS tisch_name, ts.saldo_cents
FROM tisch_sessions ts
JOIN tische t ON t.id = ts.tisch_id
WHERE ts.saldo_cents > 0
  AND ts.kassensitzung_nr = @kassensitzung_nr
ORDER BY t.name;

-- name: GetReportingStats :one
-- Reporting: Aggregierte Kennzahlen fuer eine Kassensitzung.
SELECT
    (
        COALESCE(SUM(kj_extract_zahlung_cents(type, data)), 0)::int
        + COALESCE(SUM(kj_extract_direktverkauf_cents(type, data)), 0)::int
        - COALESCE(SUM(kj_extract_direktverkauf_storno_cents(type, data)), 0)::int
        - COALESCE(SUM(kj_extract_auszahlung_cents(type, data)), 0)::int
    )::int AS gesamt_umsatz_cents,
    COALESCE(SUM(kj_extract_auszahlung_cents(type, data)), 0)::int AS gesamt_auszahlungen_cents,
    COALESCE(SUM(kj_extract_bestellung_cents(type, data)), 0)::int AS gesamt_bestellungen_cents,
    (
        COALESCE(SUM(kj_extract_stornierung_cents(type, data)), 0)::int
        + COALESCE(SUM(kj_extract_direktverkauf_storno_cents(type, data)), 0)::int
    )::int AS gesamt_stornierungen_cents,
    COALESCE(SUM(kj_extract_geldtransit_cents(type, data)), 0)::int AS gesamt_geldtransit_cents,
    COALESCE(COUNT(CASE WHEN type = 'bestellung-aufgenommen:v1' THEN 1 END), 0)::int AS anzahl_bestellungen,
    COALESCE(COUNT(CASE WHEN type IN ('stornierung-erteilt:v1', 'direktverkauf-storniert:v1') THEN 1 END), 0)::int AS anzahl_stornierungen,
    COALESCE(COUNT(CASE WHEN type = 'direktverkauf-getaetigt:v1' THEN 1 END), 0)::int AS anzahl_direktverkaeufe,
    (
        COALESCE(SUM(kj_extract_direktverkauf_cents(type, data)), 0)::int
        - COALESCE(SUM(kj_extract_direktverkauf_storno_cents(type, data)), 0)::int
    )::int AS direktverkauf_umsatz_cents
FROM kassenjournal
WHERE type IN (
    'bestellung-aufgenommen:v1',
    'zahlung-kassiert:v1',
    'stornierung-erteilt:v1',
    'auszahlung-geleistet:v1',
    'direktverkauf-getaetigt:v1',
    'direktverkauf-storniert:v1',
    'geldtransit-gebucht:v1'
)
AND kassensitzung_nr = @kassensitzung_nr;

-- name: GetOffeneSaldi :one
-- Live-Dashboard: Summe der offenen Saldi der offenen Kassensitzung.
SELECT COALESCE(SUM(saldo_cents), 0)::int AS offene_saldi_cents
FROM tisch_sessions WHERE saldo_cents > 0 AND kassensitzung_nr = @kassensitzung_nr;

-- name: GetUmsatzProServicekraft :many
-- Tagesabrechnung: Zahlungen und Auszahlungen gruppiert nach Servicekraft pro Kassensitzung.
-- Tischservice-Umsatz (Direktverkaeufe haben keine Tischzuordnung und sind hier bewusst nicht enthalten).
-- MAX(user_name) nimmt den lexikographisch letzten Namen bei Namensaenderungen.
SELECT
    user_id,
    MAX(user_name)::text AS user_name,
    COALESCE(SUM(kj_extract_zahlung_cents(type, data)), 0)::int AS zahlungen_cents,
    COALESCE(SUM(kj_extract_auszahlung_cents(type, data)), 0)::int AS auszahlungen_cents,
    COUNT(CASE WHEN type = 'zahlung-kassiert:v1' THEN 1 END)::int AS anzahl_zahlungen
FROM kassenjournal
WHERE type IN ('zahlung-kassiert:v1', 'auszahlung-geleistet:v1')
AND kassensitzung_nr = @kassensitzung_nr
GROUP BY user_id
ORDER BY zahlungen_cents DESC;

-- name: GetStornierungen :many
-- Reporting: Storno-Events (Tisch und Direktverkauf) pro Kassensitzung.
-- Tisch-Stornos tragen einen Tischbezug; Direktverkauf-Stornos haben keinen (quelle = 'direktverkauf',
-- tisch_id = 0, tisch_name = ''). Beide Event-Typen teilen dieselbe JSONB-Form (gesamtStornierungCents,
-- kommentar, fat positions) — parse in Go.
SELECT
    e.timestamp,
    CASE WHEN e.type = 'direktverkauf-storniert:v1' THEN 'direktverkauf' ELSE 'tisch' END::text AS quelle,
    COALESCE(tss.tisch_id, 0)::int AS tisch_id,
    COALESCE(t.name, '')::text AS tisch_name,
    e.user_id,
    e.user_name,
    e.data
FROM kassenjournal e
LEFT JOIN tisch_sessions tss ON tss.subject = e.subject
LEFT JOIN tische t ON t.id = tss.tisch_id
WHERE e.type IN ('stornierung-erteilt:v1', 'direktverkauf-storniert:v1')
AND e.kassensitzung_nr = @kassensitzung_nr
ORDER BY e.timestamp DESC;

-- name: GetUmsatzProTisch :many
-- Tagesabrechnung: Zahlungen und Auszahlungen gruppiert nach Tisch pro Kassensitzung.
-- Tischservice-Umsatz (Direktverkaeufe haben keine Tischzuordnung und sind hier bewusst nicht enthalten).
SELECT
    tss.tisch_id,
    t.name AS tisch_name,
    COALESCE(SUM(kj_extract_zahlung_cents(e.type, e.data)), 0)::int AS zahlungen_cents,
    COALESCE(SUM(kj_extract_auszahlung_cents(e.type, e.data)), 0)::int AS auszahlungen_cents,
    COUNT(CASE WHEN e.type = 'zahlung-kassiert:v1' THEN 1 END)::int AS anzahl_zahlungen
FROM kassenjournal e
JOIN tisch_sessions tss ON tss.subject = e.subject
JOIN tische t ON t.id = tss.tisch_id
WHERE e.type IN ('zahlung-kassiert:v1', 'auszahlung-geleistet:v1')
AND e.kassensitzung_nr = @kassensitzung_nr
GROUP BY tss.tisch_id, t.name
ORDER BY zahlungen_cents DESC;

-- name: GetUmsatzProSteuersatz :many
-- Tagesabrechnung: Bruttoumsatz gruppiert nach Steuersatz pro Kassensitzung.
SELECT
    s.steuersatz::Steuersatz AS steuersatz,
    COALESCE(SUM(s.brutto_cents), 0)::int AS brutto_cents
FROM kassenjournal kj
CROSS JOIN LATERAL kj_extract_umsatz_pro_steuersatz(kj.type, kj.data) AS s(steuersatz, brutto_cents)
WHERE kj.type IN ('zahlung-kassiert:v1', 'direktverkauf-getaetigt:v1', 'direktverkauf-storniert:v1')
AND kj.kassensitzung_nr = @kassensitzung_nr
GROUP BY s.steuersatz
ORDER BY CASE s.steuersatz
    WHEN 'regel' THEN 1
    WHEN 'ermaessigt' THEN 2
    WHEN 'befreit' THEN 3
    WHEN 'kombi' THEN 4
    ELSE 5
END;

-- name: GetAusstehendAuszahlungen :one
-- Live-Dashboard: Summe der negativen Tischsaldi der offenen Kassensitzung.
SELECT COALESCE(SUM(ABS(saldo_cents)), 0)::int AS ausstehend_auszahlungen_cents
FROM tisch_sessions
WHERE saldo_cents < 0 AND kassensitzung_nr = @kassensitzung_nr;

-- name: GetEigeneUebersicht :one
-- Service-Dashboard: Eigene KPIs der eingeloggten Servicekraft pro Kassensitzung.
SELECT
    COALESCE(COUNT(CASE WHEN type = 'bestellung-aufgenommen:v1' THEN 1 END), 0)::int AS anzahl_bestellungen,
    COALESCE(SUM(kj_extract_bestellung_cents(type, data)), 0)::int AS bestellungen_cents,
    COALESCE(COUNT(CASE WHEN type = 'zahlung-kassiert:v1' THEN 1 END), 0)::int AS anzahl_zahlungen,
    COALESCE(SUM(kj_extract_zahlung_cents(type, data)), 0)::int AS zahlungen_cents
FROM kassenjournal
WHERE type IN ('bestellung-aufgenommen:v1', 'zahlung-kassiert:v1')
AND user_id = @user_id
AND kassensitzung_nr = @kassensitzung_nr;
