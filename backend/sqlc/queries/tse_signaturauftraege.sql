-- InsertTSESignaturauftrag reiht den Signaturauftrag eines Events ein — im
-- selben Commit wie das Event (transaktionale Outbox). event_id UNIQUE sichert
-- genau einen Auftrag je Event.
-- name: InsertTSESignaturauftrag :exec
INSERT INTO tse_signaturauftraege (event_id, tx_id, process_type, process_data, status, naechster_versuch_am, erstellt_am)
VALUES ($1, $2, $3, $4, 'offen', NOW(), NOW());

-- GetOffeneTSESignaturauftraege liefert die faelligen offenen Auftraege in
-- Einreihungs-Reihenfolge (FIFO als Soll-Eigenschaft).
-- name: GetOffeneTSESignaturauftraege :many
SELECT id, tx_id, process_type, process_data
FROM tse_signaturauftraege
WHERE status = 'offen'
  AND naechster_versuch_am <= NOW()
ORDER BY id ASC
LIMIT $1;

-- QuittiereTSESignaturauftrag schreibt die Signatur als einzelnes Update an den
-- Auftrag: Signaturspalten fuellen, Status erledigt. Der Status-Guard macht die
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
-- Rueckstands-Schwelle, TSE-weite Fehler zaehlen nie auf den Auftrag.
-- name: TSESignaturauftragFehlversuch :exec
UPDATE tse_signaturauftraege
SET versuche = versuche + 1,
    letzter_fehler = @letzter_fehler,
    naechster_versuch_am = NOW() + (5 * POWER(3, versuche)) * interval '1 second',
    status = CASE WHEN versuche + 1 >= @max_versuche THEN 'fehlgeschlagen' ELSE status END
WHERE id = @id AND status = 'offen';

-- MarkiereOffeneTSESignaturauftraegeNichtKonfiguriert markiert alle offenen
-- Auftraege endgueltig als tse_nicht_konfiguriert: ohne vorhandene
-- TSE-Konfiguration gibt es keine Signatur, ein Nachsignieren ist ausgeschlossen
-- (keine Fehlversuche, keine automatische Wiederaufnahme). Der Status-Guard
-- laesst bereits endgueltig markierte Auftraege unberuehrt. Zwei Schreiber: der
-- Signatur-Worker (Dauerzustand ohne Konfiguration) und der Einrichtungs-Sweep
-- (Uebergang zu konfiguriert, in derselben Transaktion wie das Speichern).
-- name: MarkiereOffeneTSESignaturauftraegeNichtKonfiguriert :execrows
UPDATE tse_signaturauftraege
SET status = 'tse_nicht_konfiguriert'
WHERE status = 'offen';

-- GetTSESignaturQueueZustand berechnet den Zustand der Signatur-Queue in einem
-- Durchlauf: offene und fehlgeschlagene Auftraege, das Alter des aeltesten
-- offenen Auftrags (Rueckstand) sowie Durchsatz (Signaturen pro Minute) und
-- Latenz (Signierdauer p95, erstellt_am -> TSE-logTime) ueber ein gleitendes
-- 15-Minuten-Fenster. On demand aus den Auftrags- und Signaturzeiten, kein
-- Metrik-Subsystem und kein In-Memory-Zustand.
-- name: GetTSESignaturQueueZustand :one
SELECT
    COUNT(*) FILTER (WHERE status = 'offen')::int AS offene_auftraege,
    COUNT(*) FILTER (WHERE status = 'fehlgeschlagen')::int AS fehlgeschlagene_auftraege,
    COALESCE(EXTRACT(EPOCH FROM (NOW() - MIN(erstellt_am) FILTER (WHERE status = 'offen'))), 0)::int AS rueckstand_sekunden,
    (COUNT(*) FILTER (WHERE status = 'erledigt' AND erledigt_am >= NOW() - interval '15 minutes')::float8 / 15.0)::float8 AS signaturen_pro_minute,
    COALESCE(percentile_cont(0.95) WITHIN GROUP (ORDER BY EXTRACT(EPOCH FROM (log_time_end - erstellt_am))) FILTER (WHERE status = 'erledigt' AND erledigt_am >= NOW() - interval '15 minutes'), 0)::float8 AS signierdauer_p95_sekunden
FROM tse_signaturauftraege;

-- GetAeltesterOffenerTSESignaturauftrag liefert den Erstellungszeitpunkt des
-- aeltesten offenen Auftrags — der Rueckstands-Watchdog bemisst daran den
-- Signatur-Rueckstand.
-- name: GetAeltesterOffenerTSESignaturauftrag :one
SELECT erstellt_am
FROM tse_signaturauftraege
WHERE status = 'offen'
ORDER BY erstellt_am ASC
LIMIT 1;

-- GetTSESignaturauftragZuEvent liefert den Signatur-Stand eines Events fuer den
-- Beleg-Abruf: Status plus Signaturspalten (gefuellt sobald quittiert).
-- Kein Treffer heisst: Das Event ist nicht signaturpflichtig.
-- name: GetTSESignaturauftragZuEvent :one
SELECT status, erstellt_am, transaktion_nummer, signatur_zaehler, tse_seriennummer, log_time_start, log_time_end, signatur, qr_code_data
FROM tse_signaturauftraege
WHERE event_id = $1;

-- GetOffeneSignaturauftragStaendeFuerKassensitzung liefert die Signatur-Stände
-- aller noch nicht erledigten Signaturauftraege einer Kassensitzung — die
-- Grundlage des Kassenabschluss-Gates. Erledigte Auftraege sind irrelevant
-- (bereits signiert); die vier nicht-erledigten Status ordnet
-- BestimmeSignaturstatus in ausstehend (blockiert) bzw. Ausfall (Rest) ein.
-- name: GetOffeneSignaturauftragStaendeFuerKassensitzung :many
SELECT a.status, a.erstellt_am
FROM tse_signaturauftraege a
JOIN kassenjournal k ON k.id = a.event_id
WHERE k.kassensitzung_nr = $1 AND a.status <> 'erledigt'
ORDER BY a.erstellt_am ASC;
