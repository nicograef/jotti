-- name: GetOffeneTischeDetails :many
-- Live-Dashboard: Offene Tische der offenen Kassensitzung mit Name und aktuellem
-- Saldo, größte offene Beträge zuerst (Tisch-Name als stabiler Tiebreaker).
SELECT ts.tisch_id, t.name AS tisch_name, ts.saldo_cents
FROM tisch_sessions ts
JOIN tische t ON t.id = ts.tisch_id
WHERE ts.saldo_cents > 0
  AND ts.kassensitzung_nr = @kassensitzung_nr
ORDER BY ts.saldo_cents DESC, t.name;

-- name: GetReportingStats :one
-- Reporting: Aggregierte Kennzahlen fuer eine Kassensitzung.
SELECT
    (
        COALESCE(SUM(kj_extract_zahlung_cents(type, data)), 0)::int
        + COALESCE(SUM(kj_extract_direktverkauf_cents(type, data)), 0)::int
        - COALESCE(SUM(kj_extract_direktverkauf_storno_cents(type, data)), 0)::int
        - COALESCE(SUM(kj_extract_stornierung_cents(type, data)), 0)::int
    )::int AS gesamt_umsatz_cents,
    COALESCE(SUM(kj_extract_bestellung_cents(type, data)), 0)::int AS gesamt_bestellungen_cents,
    (
        COALESCE(SUM(kj_extract_stornierung_cents(type, data)), 0)::int
        + COALESCE(SUM(kj_extract_korrektur_cents(type, data)), 0)::int
        + COALESCE(SUM(kj_extract_direktverkauf_storno_cents(type, data)), 0)::int
    )::int AS gesamt_stornierungen_cents,
    COALESCE(SUM(kj_extract_geldtransit_cents(type, data)), 0)::int AS gesamt_geldtransit_cents,
    COALESCE(COUNT(CASE WHEN type = 'bestellung-aufgenommen:v1' THEN 1 END), 0)::int AS anzahl_bestellungen,
    COALESCE(COUNT(CASE WHEN type IN ('stornierung-erteilt:v1', 'bestellung-korrigiert:v1', 'direktverkauf-storniert:v1') THEN 1 END), 0)::int AS anzahl_stornierungen,
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
    'bestellung-korrigiert:v1',
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
-- Tagesabrechnung: kassierte Zahlungen gruppiert nach Servicekraft pro Kassensitzung.
-- Tischservice-Umsatz (Direktverkaeufe haben keine Tischzuordnung und sind hier bewusst nicht enthalten).
-- MAX(user_name) nimmt den lexikographisch letzten eingefrorenen Username; name ist der live aus users
-- aufgeloeste Klarname (bleibt auch fuer soft-geloeschte Benutzer verfuegbar, leer wenn der Benutzer fehlt).
SELECT
    e.user_id,
    MAX(e.user_name)::text AS user_name,
    COALESCE(MAX(u.name), '')::text AS name,
    COALESCE(SUM(kj_extract_zahlung_cents(e.type, e.data)), 0)::int AS zahlungen_cents,
    COUNT(CASE WHEN e.type = 'zahlung-kassiert:v1' THEN 1 END)::int AS anzahl_zahlungen
FROM kassenjournal e
LEFT JOIN users u ON u.id = e.user_id
WHERE e.type = 'zahlung-kassiert:v1'
AND e.kassensitzung_nr = @kassensitzung_nr
GROUP BY e.user_id
ORDER BY zahlungen_cents DESC;

-- name: GetStornierungen :many
-- Reporting: Storno-Events pro Kassensitzung — kassenwirksame Warenrücknahme (stornierung-erteilt),
-- geldneutrale Korrektur (bestellung-korrigiert) und Direktverkauf-Storno (direktverkauf-storniert).
-- bar_rueckgabe markiert die kassenwirksamen Stornos (Bargeld-Rückgabe). Tisch-Stornos tragen einen
-- Tischbezug; Direktverkauf-Stornos haben keinen (quelle = 'direktverkauf', tisch_id = 0, tisch_name = '').
-- Der Betrag liegt je nach Event-Typ in gesamtStornierungCents oder gesamtCents; kommentar und fat
-- positions werden in Go aus data geparst.
SELECT
    e.timestamp,
    CASE WHEN e.type = 'direktverkauf-storniert:v1' THEN 'direktverkauf' ELSE 'tisch' END::text AS quelle,
    (e.type IN ('stornierung-erteilt:v1', 'direktverkauf-storniert:v1'))::bool AS bar_rueckgabe,
    COALESCE(tss.tisch_id, 0)::int AS tisch_id,
    COALESCE(t.name, '')::text AS tisch_name,
    e.user_id,
    e.user_name,
    COALESCE(u.name, '')::text AS name,
    COALESCE((e.data->>'gesamtStornierungCents')::int, (e.data->>'gesamtCents')::int, 0)::int AS betrag_cents,
    e.data
FROM kassenjournal e
LEFT JOIN tisch_sessions tss ON tss.subject = e.subject
LEFT JOIN tische t ON t.id = tss.tisch_id
LEFT JOIN users u ON u.id = e.user_id
WHERE e.type IN ('stornierung-erteilt:v1', 'bestellung-korrigiert:v1', 'direktverkauf-storniert:v1')
AND e.kassensitzung_nr = @kassensitzung_nr
ORDER BY e.timestamp DESC;

-- name: GetUmsatzPositionszeilen :many
-- Tagesabrechnung: umsatzwirksame Brutto-Positionszeilen mit Steuersatz pro
-- Kassensitzung, unaggregiert (eine Zeile je Position). Die USt-Aufschlüsselung
-- rechnet die Anwendungsschicht auf Zeilenbasis (steuer.Aufteilen je Zeile,
-- danach Aggregation) — dieselbe Basis wie Beleg, TSE-processData und
-- DSFinV-K-Export. Kassenwirksame Warenrücknahmen (stornierung-erteilt) und
-- Direktverkauf-Stornos zählen als negative Zeilen (Faktor -1 in
-- kj_extract_umsatz_pro_steuersatz).
SELECT
    s.steuersatz::Steuersatz AS steuersatz,
    s.brutto_cents::int AS brutto_cents
FROM kassenjournal kj
CROSS JOIN LATERAL kj_extract_umsatz_pro_steuersatz(kj.type, kj.data) AS s(steuersatz, brutto_cents)
WHERE kj.type IN ('zahlung-kassiert:v1', 'direktverkauf-getaetigt:v1', 'direktverkauf-storniert:v1', 'stornierung-erteilt:v1')
AND kj.kassensitzung_nr = @kassensitzung_nr;

-- name: GetProduktStatistik :many
-- Tagesabrechnung/Live: Verkäufe je Produkt und Variante einer Kassensitzung,
-- aus den eingefrorenen Fat-Event-Positionen aggregiert (kein Stammdaten-Join).
-- Flache Zeilen je Variante mit zwei bewusst getrennten Zahlen — die
-- Anwendungsschicht gruppiert und sortiert sie zu Kategorie-Abschnitten:
--   Ausgegebene Menge (Produktion) = Σ menge: bestellung-aufgenommen (+),
--     bestellung-korrigiert (−), direktverkauf-getaetigt (+).
--   Umsatz (Einnahmen) = Σ einzelpreisCents × menge: zahlung-kassiert (+),
--     direktverkauf-getaetigt (+), stornierung-erteilt (−), direktverkauf-storniert (−).
-- bestellung-umgebucht zählt nicht (Positionen bereits bei der Bestellung erfasst).
-- Dieselbe Positions-/Vorzeichenbasis wie GetUmsatzPositionszeilen für den
-- Umsatzanteil, damit Σ umsatzCents dem kassierten Gesamtumsatz entspricht.
SELECT
    (position->>'kategorie')::text AS kategorie,
    (position->>'varianteId')::int AS variante_id,
    (position->>'produktName')::text AS produkt_name,
    (position->>'varianteName')::text AS variante_name,
    COALESCE(SUM(
        (position->>'menge')::int * CASE
            WHEN kj.type IN ('bestellung-aufgenommen:v1', 'direktverkauf-getaetigt:v1') THEN 1
            WHEN kj.type = 'bestellung-korrigiert:v1' THEN -1
            ELSE 0
        END
    ), 0)::int AS ausgegebene_menge,
    COALESCE(SUM(
        ((position->>'einzelpreisCents')::int * (position->>'menge')::int) * CASE
            WHEN kj.type IN ('zahlung-kassiert:v1', 'direktverkauf-getaetigt:v1') THEN 1
            WHEN kj.type IN ('stornierung-erteilt:v1', 'direktverkauf-storniert:v1') THEN -1
            ELSE 0
        END
    ), 0)::int AS umsatz_cents
FROM kassenjournal kj
CROSS JOIN LATERAL jsonb_array_elements(kj.data->'positionen') AS position
WHERE kj.type IN (
    'bestellung-aufgenommen:v1',
    'bestellung-korrigiert:v1',
    'direktverkauf-getaetigt:v1',
    'zahlung-kassiert:v1',
    'stornierung-erteilt:v1',
    'direktverkauf-storniert:v1'
)
AND kj.kassensitzung_nr = @kassensitzung_nr
GROUP BY kategorie, variante_id, produkt_name, variante_name;

-- name: GetKassensitzungMetadaten :one
-- Tagesabrechnung: Sitzungs-Metadaten für den formalen Berichtskopf, reine
-- Projektion über die vorhandenen Journal-Events. Eröffnungs- und
-- Abschlusszeitpunkt sind die Event-timestamps von kassensitzung-eroeffnet:v1
-- bzw. tagesabschluss-erstellt:v1; der abschließende Benutzer ist der
-- eingefrorene user_name des Tagesabschluss-Events; die Kassensturz-Differenz
-- kommt aus data->differenzCents des kassensturz-durchgefuehrt:v1-Events. Alle
-- Felder sind nullable, solange die zugehörigen Events noch nicht existieren.
SELECT
    eroeffnet.eroeffnet_am,
    abschluss.abgeschlossen_am,
    abschluss.abgeschlossen_von,
    -- kassensturz_data ist das JSONB-Event des letzten Kassensturzes; ohne
    -- Kassensturz liefert die COALESCE das JSON-Literal 'null' (kein SQL-NULL,
    -- damit json.RawMessage sauber scannt). Die Anwendungsschicht parst
    -- differenzCents daraus — wie das Reporting-Repo die Storno-Positionen.
    COALESCE(kassensturz.kassensturz_data, 'null'::jsonb)::jsonb AS kassensturz_data
FROM (SELECT @kassensitzung_nr::int AS nr) params
LEFT JOIN LATERAL (
    SELECT kj.timestamp AS eroeffnet_am FROM kassenjournal kj
    WHERE kj.kassensitzung_nr = params.nr
      AND kj.type = 'kassensitzung-eroeffnet:v1'
    ORDER BY kj.timestamp ASC LIMIT 1
) eroeffnet ON true
LEFT JOIN LATERAL (
    SELECT kj.timestamp AS abgeschlossen_am, kj.user_name AS abgeschlossen_von FROM kassenjournal kj
    WHERE kj.kassensitzung_nr = params.nr
      AND kj.type = 'tagesabschluss-erstellt:v1'
    ORDER BY kj.timestamp DESC LIMIT 1
) abschluss ON true
LEFT JOIN LATERAL (
    SELECT kj.data AS kassensturz_data FROM kassenjournal kj
    WHERE kj.kassensitzung_nr = params.nr
      AND kj.type = 'kassensturz-durchgefuehrt:v1'
    ORDER BY kj.timestamp DESC LIMIT 1
) kassensturz ON true;

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
