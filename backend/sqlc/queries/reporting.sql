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

-- name: GetKassiertProServicekraft :many
-- Tagesabrechnung: kassierte Zahlungen gruppiert nach Servicekraft pro Kassensitzung — die
-- Kassiert-Seite der Abrechnung pro Servicekraft (die zugeordneten Ruecknahmen kommen aus den
-- Storno-Detailzeilen und werden in der Anwendungsschicht gegengerechnet).
-- Tischservice-Umsatz (Direktverkaeufe haben keine Tischzuordnung und sind hier bewusst nicht enthalten).
-- MAX(user_name) nimmt den lexikographisch letzten eingefrorenen Username; name ist der live aus users
-- aufgeloeste Klarname (bleibt auch fuer soft-geloeschte Benutzer verfuegbar, leer wenn der Benutzer fehlt).
SELECT
    e.user_id,
    MAX(e.user_name)::text AS user_name,
    COALESCE(MAX(u.name), '')::text AS name,
    COALESCE(SUM(kj_extract_zahlung_cents(e.type, e.data)), 0)::int AS kassiert_cents,
    COUNT(CASE WHEN e.type = 'zahlung-kassiert:v1' THEN 1 END)::int AS anzahl_zahlungen
FROM kassenjournal e
LEFT JOIN users u ON u.id = e.user_id
WHERE e.type = 'zahlung-kassiert:v1'
AND e.kassensitzung_nr = @kassensitzung_nr
GROUP BY e.user_id
ORDER BY kassiert_cents DESC;

-- name: GetStornierungen :many
-- Reporting: Storno-Events pro Kassensitzung — kassenwirksame Warenrücknahme (stornierung-erteilt),
-- geldneutrale Korrektur (bestellung-korrigiert) und Direktverkauf-Storno (direktverkauf-storniert).
-- bar_rueckgabe markiert die kassenwirksamen Stornos (Bargeld-Rückgabe). Tisch-Stornos tragen einen
-- Tischbezug; Direktverkauf-Stornos haben keinen (quelle = 'direktverkauf', tisch_id = 0, tisch_name = '').
-- Der Betrag liegt je nach Event-Typ in gesamtStornierungCents oder gesamtCents; kommentar und fat
-- positions werden in Go aus data geparst.
--
-- Storno-Zuordnung: Neben dem Akteur (user_id/user_name/name — wer storniert hat) liefert die Query
-- die betroffenen Servicekräfte, deren Vorgang der Storno rückgängig macht, als JSONB-Array
-- [{userId, userName, name}]. Aufgelöst wird jeweils innerhalb derselben Kassensitzung:
--   Warenrücknahme       → zahlung-kassiert:v1 mit derselben zahlungId (Kassierer, einwertig)
--   Direktverkauf-Storno → direktverkauf-getaetigt:v1 mit derselben verkaufId (Verkäufer, einwertig)
--   Korrektur            → je Positions-ID das bestellung-aufgenommen:v1, dessen Positions-Array
--                          diese ID enthält (Besteller, mehrwertig, je Person genau einmal)
-- Findet die Auflösung nichts, fällt die Liste auf den Akteur zurück und ist damit nie leer.
--
-- Die Auflösung läuft über CTEs statt über ein korreliertes LATERAL: Das Sitzungs-Journal wird einmal
-- gelesen und die Positions-Arrays der Bestellungen einmal expandiert, statt je Storno-Zeile erneut.
-- Dedupliziert wird nach user_id (MAX(user_name) nimmt den lexikographisch letzten eingefrorenen
-- Username, wie GetKassiertProServicekraft) — ein Rename während der Sitzung darf dieselbe Person
-- nicht zweimal listen.
WITH storno AS (
    SELECT e.id, e.type, e.data
    FROM kassenjournal e
    WHERE e.type IN ('stornierung-erteilt:v1', 'bestellung-korrigiert:v1', 'direktverkauf-storniert:v1')
    AND e.kassensitzung_nr = @kassensitzung_nr
), ursprung AS (
    SELECT k.type, k.data, k.user_id, k.user_name
    FROM kassenjournal k
    WHERE k.type IN ('zahlung-kassiert:v1', 'direktverkauf-getaetigt:v1', 'bestellung-aufgenommen:v1')
    AND k.kassensitzung_nr = @kassensitzung_nr
), bestell_position AS (
    SELECT u.user_id, u.user_name, pos->>'positionId' AS position_id
    FROM ursprung u, jsonb_array_elements(u.data->'positionen') pos
    WHERE u.type = 'bestellung-aufgenommen:v1'
), zuordnung AS (
    SELECT s.id, u.user_id, u.user_name
    FROM storno s
    JOIN ursprung u ON u.type = 'zahlung-kassiert:v1' AND u.data->>'zahlungId' = s.data->>'zahlungId'
    WHERE s.type = 'stornierung-erteilt:v1'
    UNION ALL
    SELECT s.id, u.user_id, u.user_name
    FROM storno s
    JOIN ursprung u ON u.type = 'direktverkauf-getaetigt:v1' AND u.data->>'verkaufId' = s.data->>'verkaufId'
    WHERE s.type = 'direktverkauf-storniert:v1'
    UNION ALL
    SELECT s.id, b.user_id, b.user_name
    FROM storno s
    CROSS JOIN LATERAL jsonb_array_elements(s.data->'positionen') storno_pos
    JOIN bestell_position b ON b.position_id = storno_pos->>'positionId'
    WHERE s.type = 'bestellung-korrigiert:v1'
), betroffene_je_storno AS (
    SELECT je_person.id, jsonb_agg(
        jsonb_build_object('userId', je_person.user_id, 'userName', je_person.user_name, 'name', COALESCE(bu.name, ''))
        ORDER BY je_person.user_id
    ) AS betroffene
    FROM (
        SELECT z.id, z.user_id, MAX(z.user_name)::text AS user_name
        FROM zuordnung z
        GROUP BY z.id, z.user_id
    ) je_person
    LEFT JOIN users bu ON bu.id = je_person.user_id
    GROUP BY je_person.id
)
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
    e.data,
    COALESCE(
        bjs.betroffene,
        jsonb_build_array(jsonb_build_object(
            'userId', e.user_id, 'userName', e.user_name, 'name', COALESCE(u.name, '')
        ))
    )::jsonb AS betroffene
FROM kassenjournal e
LEFT JOIN tisch_sessions tss ON tss.subject = e.subject
LEFT JOIN tische t ON t.id = tss.tisch_id
LEFT JOIN users u ON u.id = e.user_id
LEFT JOIN betroffene_je_storno bjs ON bjs.id = e.id
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
-- Anwendungsschicht gruppiert und sortiert sie zu Kategorie-Abschnitten. Beide
-- Zahlen ruhen auf derselben Ereignismenge/Gewichtung (Bestellbasis):
--   Ausgegebene Menge (Produktion) = Σ menge: bestellung-aufgenommen (+),
--     bestellung-korrigiert (−), direktverkauf-getaetigt (+).
--   Umsatz (Bestellwert) = Σ einzelpreisCents × menge über dieselbe Ereignismenge
--     und Gewichtung: bestellung-aufgenommen (+), bestellung-korrigiert (−),
--     direktverkauf-getaetigt (+). Es ist der Euro-Wert genau der Portionen, die
--     „Ausgegeben" zählt (Preise zum Bestellzeitpunkt).
-- bestellung-umgebucht zählt nicht (Positionen bereits bei der Bestellung erfasst).
-- Bewusst NICHT kassenbasiert: nachträgliche Zahlungen und Stornierungen
-- (zahlung-kassiert, stornierung-erteilt, direktverkauf-storniert) ändern den
-- Produkt-Umsatz nicht. Die fiskalische, kassierte Umsatzbasis liefert separat
-- GetUmsatzPositionszeilen (Steuer-Aufschlüsselung, DSFinV-K, „Kassierter Umsatz");
-- beide Größen weichen bei Stornos bewusst voneinander ab.
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
            WHEN kj.type IN ('bestellung-aufgenommen:v1', 'direktverkauf-getaetigt:v1') THEN 1
            WHEN kj.type = 'bestellung-korrigiert:v1' THEN -1
            ELSE 0
        END
    ), 0)::int AS umsatz_cents
FROM kassenjournal kj
CROSS JOIN LATERAL jsonb_array_elements(kj.data->'positionen') AS position
WHERE kj.type IN (
    'bestellung-aufgenommen:v1',
    'bestellung-korrigiert:v1',
    'direktverkauf-getaetigt:v1'
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
--
-- Die Rücknahmen folgen derselben Storno-Zuordnung wie GetStornierungen, hier aber für genau
-- einen Benutzer und ohne Umweg über die Detailzeilen: Eine Warenrücknahme zählt für diese
-- Servicekraft, wenn die über zahlungId referenzierte Zahlung von ihr kassiert wurde — egal,
-- wer storniert hat. Geldneutrale Korrekturen bleiben bewusst außen vor; sie ändern nichts an
-- dem, was abzugeben ist. abzugeben_cents ist nie negativ: Pro Zahlung gilt
-- Σ Rücknahmen <= Zahlbetrag (FIFO-Buchführung in ComputeStornoAufteilung), und beide Seiten
-- sind demselben Kassierer zugeordnet.
SELECT
    eigene.anzahl_bestellungen,
    eigene.bestellungen_cents,
    eigene.anzahl_zahlungen,
    eigene.zahlungen_cents,
    ruecknahmen.anzahl_ruecknahmen,
    ruecknahmen.ruecknahmen_cents,
    (eigene.zahlungen_cents - ruecknahmen.ruecknahmen_cents)::int AS abzugeben_cents
FROM (
    SELECT
        COALESCE(COUNT(CASE WHEN kj.type = 'bestellung-aufgenommen:v1' THEN 1 END), 0)::int AS anzahl_bestellungen,
        COALESCE(SUM(kj_extract_bestellung_cents(kj.type, kj.data)), 0)::int AS bestellungen_cents,
        COALESCE(COUNT(CASE WHEN kj.type = 'zahlung-kassiert:v1' THEN 1 END), 0)::int AS anzahl_zahlungen,
        COALESCE(SUM(kj_extract_zahlung_cents(kj.type, kj.data)), 0)::int AS zahlungen_cents
    FROM kassenjournal kj
    WHERE kj.type IN ('bestellung-aufgenommen:v1', 'zahlung-kassiert:v1')
    AND kj.user_id = @user_id
    AND kj.kassensitzung_nr = @kassensitzung_nr
) eigene
CROSS JOIN (
    SELECT
        COUNT(*)::int AS anzahl_ruecknahmen,
        COALESCE(SUM(kj_extract_stornierung_cents(storno.type, storno.data)), 0)::int AS ruecknahmen_cents
    FROM kassenjournal storno
    WHERE storno.type = 'stornierung-erteilt:v1'
    AND storno.kassensitzung_nr = @kassensitzung_nr
    AND EXISTS (
        SELECT 1
        FROM kassenjournal zahlung
        WHERE zahlung.type = 'zahlung-kassiert:v1'
        AND zahlung.kassensitzung_nr = @kassensitzung_nr
        AND zahlung.user_id = @user_id
        AND zahlung.data->>'zahlungId' = storno.data->>'zahlungId'
    )
) ruecknahmen;
