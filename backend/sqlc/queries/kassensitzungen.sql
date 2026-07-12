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
-- Umsatz und Abschlusszeitpunkt sind reine Projektionen aus dem
-- tagesabschluss-erstellt:v1-Event (data->umsatzGesamtCents bzw. dessen
-- timestamp); das LEFT JOIN LATERAL bleibt tolerant, falls das Event fehlt.
SELECT
    ks.z_nr,
    ks.datum,
    ks.bezeichnung,
    ks.status,
    ks.created_at,
    ks.updated_at,
    COALESCE(ta.umsatz_gesamt_cents, 0)::int AS umsatz_gesamt_cents,
    ta.abgeschlossen_am AS abgeschlossen_am
FROM kassensitzungen ks
LEFT JOIN LATERAL (
    SELECT
        COALESCE((kj.data->>'umsatzGesamtCents')::int, 0) AS umsatz_gesamt_cents,
        kj.timestamp AS abgeschlossen_am
    FROM kassenjournal kj
    WHERE kj.kassensitzung_nr = ks.z_nr
      AND kj.type = 'tagesabschluss-erstellt:v1'
    ORDER BY kj.timestamp DESC
    LIMIT 1
) ta ON true
WHERE ks.status = 'abgeschlossen'
ORDER BY ks.datum DESC, ks.created_at DESC;

-- name: GetKassenbestand :one
-- Kassenbestand (Soll) samt Aufschlüsselung. Der Soll-Bestand ist die Summe aus
-- Anfangsbestand, Zahlungen, Warenrücknahmen, Geldtransits und Differenz-Buchungen.
-- Die kassenwirksame Warenrücknahme (stornierung-erteilt) gibt Bargeld zurück und mindert den Bestand;
-- geldneutrale Vorgänge (bestellung-korrigiert, bestellung-umgebucht) berühren den Kassenbestand nicht.
-- Die Differenz ist als Soll − Ist gebucht; ihre Bargeldwirkung ist Ist − Soll und wird deshalb
-- subtrahiert: Nach der Differenzbuchung entspricht der Soll-Bestand dem gezählten Ist-Bestand.
--
-- Die vier Komponenten werten dieselben kj_extract_*-Funktionen einzeln aus:
--   - anfangsbestand = Anfangsbestand der Eröffnung
--   - bareinnahmen   = Zahlungen + Direktverkauf − geldwirksame Stornos (Warenrücknahme + Direktverkauf-Storno)
--   - einlagen       = geldtransit-gebucht:v1 mit richtung='einlage'
--   - entnahmen      = geldtransit-gebucht:v1 mit richtung='entnahme'
-- Invariante (vor Kassensturz, also ohne Differenzbuchung):
--   anfangsbestand + bareinnahmen + einlagen − entnahmen = soll_bestand_cents.
SELECT
    (
        COALESCE(SUM(kj_extract_eroeffnung_cents(type, data)), 0)::int
        + COALESCE(SUM(kj_extract_zahlung_cents(type, data)), 0)::int
        - COALESCE(SUM(kj_extract_stornierung_cents(type, data)), 0)::int
        + COALESCE(SUM(kj_extract_direktverkauf_cents(type, data)), 0)::int
        - COALESCE(SUM(kj_extract_direktverkauf_storno_cents(type, data)), 0)::int
        + COALESCE(SUM(kj_extract_geldtransit_cents(type, data)), 0)::int
        - COALESCE(SUM(kj_extract_differenz_cents(type, data)), 0)::int
    )::int AS soll_bestand_cents,
    COALESCE(SUM(kj_extract_eroeffnung_cents(type, data)), 0)::int AS anfangsbestand_cents,
    (
        COALESCE(SUM(kj_extract_zahlung_cents(type, data)), 0)::int
        + COALESCE(SUM(kj_extract_direktverkauf_cents(type, data)), 0)::int
        - COALESCE(SUM(kj_extract_stornierung_cents(type, data)), 0)::int
        - COALESCE(SUM(kj_extract_direktverkauf_storno_cents(type, data)), 0)::int
    )::int AS bareinnahmen_cents,
    COALESCE(SUM(CASE WHEN type = 'geldtransit-gebucht:v1' AND data->>'richtung' = 'einlage'
        THEN (data->>'betragCents')::int END), 0)::int AS einlagen_cents,
    COALESCE(SUM(CASE WHEN type = 'geldtransit-gebucht:v1' AND data->>'richtung' = 'entnahme'
        THEN (data->>'betragCents')::int END), 0)::int AS entnahmen_cents
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

-- name: GetGeldtransitListe :many
-- Alle Geldbewegungen (Einlagen/Entnahmen) einer Kassensitzung als reine
-- Projektion der geldtransit-gebucht:v1-Events, neueste zuerst. Der Anzeigename
-- ist der eingefrorene user_name aus dem Kassenjournal.
SELECT
    timestamp AS zeitpunkt,
    (data->>'richtung')::text AS richtung,
    (data->>'betragCents')::int AS betrag_cents,
    (data->>'kommentar')::text AS kommentar,
    user_name AS gebucht_von
FROM kassenjournal
WHERE kassensitzung_nr = $1
  AND type = 'geldtransit-gebucht:v1'
ORDER BY timestamp DESC, id DESC;
