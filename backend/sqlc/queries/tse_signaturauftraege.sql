-- InsertTSESignaturauftrag reiht den Signaturauftrag eines Events ein — im
-- selben Commit wie das Event (transaktionale Outbox). event_id UNIQUE sichert
-- genau einen Auftrag je Event.
-- name: InsertTSESignaturauftrag :exec
INSERT INTO tse_signaturauftraege (event_id, tx_id, process_type, process_data, status, naechster_versuch_am, erstellt_am)
VALUES ($1, $2, $3, $4, 'offen', NOW(), NOW());

-- GetOffeneTSESignaturauftraege liefert die fälligen offenen Aufträge in
-- Einreihungs-Reihenfolge (FIFO als Soll-Eigenschaft).
-- name: GetOffeneTSESignaturauftraege :many
SELECT id, tx_id, process_type, process_data
FROM tse_signaturauftraege
WHERE status = 'offen'
  AND naechster_versuch_am <= NOW()
ORDER BY id ASC
LIMIT $1;

-- QuittiereTSESignaturauftrag schreibt die Signatur als einzelnes Update an den
-- Auftrag: Signaturspalten füllen, Status erledigt. Der Status-Guard macht die
-- Quittierung idempotent (Signaturspalten werden genau einmal beschrieben).
-- name: QuittiereTSESignaturauftrag :exec
UPDATE tse_signaturauftraege
SET status = 'erledigt',
    erledigt_am = NOW(),
    transaktion_nummer = @transaktion_nummer,
    signatur_zaehler = @signatur_zaehler,
    tse_seriennummer = @tse_seriennummer,
    log_time_start = @log_time_start,
    log_time_end = @log_time_end,
    signatur = @signatur,
    qr_code_data = @qr_code_data
WHERE id = @id AND status = 'offen';

-- TSESignaturauftragFehlversuch verbucht einen auftragsspezifischen
-- Fehlversuch mit Sekunden-Backoff (5 * 3^versuche: 5, 15, 45 s). Beim
-- max_versuche-ten Fehlversuch wechselt der Auftrag auf fehlgeschlagen und
-- wird nicht mehr automatisch versucht — die Kurve endet bewusst unter der
-- Rückstands-Schwelle, TSE-weite Fehler zählen nie auf den Auftrag.
-- name: TSESignaturauftragFehlversuch :exec
UPDATE tse_signaturauftraege
SET versuche = versuche + 1,
    letzter_fehler = @letzter_fehler,
    naechster_versuch_am = NOW() + (5 * POWER(3, versuche)) * interval '1 second',
    status = CASE WHEN versuche + 1 >= @max_versuche THEN 'fehlgeschlagen' ELSE status END
WHERE id = @id AND status = 'offen';

-- MarkOffeneTSESignaturauftraegeNichtKonfiguriert markiert alle offenen
-- Aufträge endgültig als tse_nicht_konfiguriert: ohne vorhandene
-- TSE-Konfiguration gibt es keine Signatur, ein Nachsignieren ist ausgeschlossen
-- (keine Fehlversuche, keine automatische Wiederaufnahme). Der Status-Guard
-- lässt bereits endgültig markierte Aufträge unberührt. Zwei Schreiber: der
-- Signatur-Worker (Dauerzustand ohne Konfiguration) und der Einrichtungs-Sweep
-- (Übergang zu konfiguriert, in derselben Transaktion wie das Speichern).
-- name: MarkOffeneTSESignaturauftraegeNichtKonfiguriert :execrows
UPDATE tse_signaturauftraege
SET status = 'tse_nicht_konfiguriert'
WHERE status = 'offen';

-- GetTSESignaturQueueZustand berechnet den Zustand der Signatur-Queue in einem
-- Durchlauf: offene Aufträge, das Alter des ältesten offenen Auftrags
-- (Rückstand) sowie Durchsatz (Signaturen pro Minute) und Latenz (Signierdauer
-- p95, erstellt_am -> TSE-logTime) über ein gleitendes 15-Minuten-Fenster —
-- diese Kennzahlen global. Die fehlgeschlagenen Aufträge dagegen zählen nur
-- die der aktiven Kassensitzung (Status offen oder wird_abgeschlossen), und
-- letzter_fehler trägt den Fehlertext des jüngsten davon. Ohne aktive Sitzung
-- ist beides leer — der Kassenabschluss weist die Ausfall-Reste aus und
-- quittiert damit die Warnung. On demand aus den Auftrags- und Signaturzeiten,
-- kein Metrik-Subsystem und kein In-Memory-Zustand.
-- name: GetTSESignaturQueueZustand :one
WITH aktive_fehlgeschlagene AS (
    SELECT a.id, a.letzter_fehler
    FROM tse_signaturauftraege a
    JOIN kassenjournal k ON k.id = a.event_id
    JOIN kassensitzungen s ON s.z_nr = k.kassensitzung_nr
    WHERE a.status = 'fehlgeschlagen'
      AND s.status IN ('offen', 'wird_abgeschlossen')
)
SELECT
    COUNT(*) FILTER (WHERE status = 'offen')::int AS offene_auftraege,
    (SELECT COUNT(*) FROM aktive_fehlgeschlagene)::int AS fehlgeschlagene_auftraege,
    COALESCE((SELECT letzter_fehler FROM aktive_fehlgeschlagene ORDER BY id DESC LIMIT 1), '')::text AS letzter_fehler,
    COALESCE(EXTRACT(EPOCH FROM (NOW() - MIN(erstellt_am) FILTER (WHERE status = 'offen'))), 0)::int AS rueckstand_sekunden,
    (COUNT(*) FILTER (WHERE status = 'erledigt' AND erledigt_am >= NOW() - interval '15 minutes')::float8 / 15.0)::float8 AS signaturen_pro_minute,
    COALESCE(percentile_cont(0.95) WITHIN GROUP (ORDER BY EXTRACT(EPOCH FROM (log_time_end - erstellt_am))) FILTER (WHERE status = 'erledigt' AND erledigt_am >= NOW() - interval '15 minutes'), 0)::float8 AS signierdauer_p95_sekunden
FROM tse_signaturauftraege;

-- GetAeltesterOffenerTSESignaturauftrag liefert den Erstellungszeitpunkt des
-- ältesten offenen Auftrags — der Rückstands-Watchdog bemisst daran den
-- Signatur-Rückstand.
-- name: GetAeltesterOffenerTSESignaturauftrag :one
SELECT erstellt_am
FROM tse_signaturauftraege
WHERE status = 'offen'
ORDER BY erstellt_am ASC
LIMIT 1;

-- GetTSESignaturauftragZuEvent liefert den Signatur-Stand eines Events für den
-- Beleg-Abruf: Status plus Signaturspalten (gefüllt sobald quittiert).
-- Kein Treffer heißt: Das Event ist nicht signaturpflichtig.
-- name: GetTSESignaturauftragZuEvent :one
SELECT status, erstellt_am, transaktion_nummer, signatur_zaehler, tse_seriennummer, log_time_start, log_time_end, signatur, qr_code_data
FROM tse_signaturauftraege
WHERE event_id = $1;

-- GetOffeneSignaturauftragStaendeFuerKassensitzung liefert die Signatur-Stände
-- aller noch nicht erledigten Signaturaufträge einer Kassensitzung — die
-- Grundlage des Kassenabschluss-Gates. Erledigte Aufträge sind irrelevant
-- (bereits signiert); die vier nicht-erledigten Status ordnet
-- DetermineSignaturstatus in ausstehend (blockiert) bzw. Ausfall (Rest) ein.
-- name: GetOffeneSignaturauftragStaendeFuerKassensitzung :many
SELECT a.status, a.erstellt_am
FROM tse_signaturauftraege a
JOIN kassenjournal k ON k.id = a.event_id
WHERE k.kassensitzung_nr = $1 AND a.status <> 'erledigt'
ORDER BY a.erstellt_am ASC;
