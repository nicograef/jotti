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

-- TSESignaturauftragFehlversuch verbucht einen Fehlversuch mit exponentiellem
-- Backoff (1, 2, 4, ... Minuten, gedeckelt auf 30). Beim max_versuche-ten
-- Fehlversuch wechselt der Auftrag auf fehlgeschlagen und wird nicht mehr
-- automatisch versucht.
-- name: TSESignaturauftragFehlversuch :exec
UPDATE tse_signaturauftraege
SET versuche = versuche + 1,
    letzter_fehler = @letzter_fehler,
    naechster_versuch_am = NOW() + LEAST(POWER(2, versuche), 30) * interval '1 minute',
    status = CASE WHEN versuche + 1 >= @max_versuche THEN 'fehlgeschlagen' ELSE status END
WHERE id = @id AND status = 'offen';

-- Zaehlt noch nicht erledigte Signaturauftraege (offen und fehlgeschlagen):
-- beide Status bedeuten unsignierte Vorgaenge.
-- name: CountOffeneTSESignaturauftraege :one
SELECT COUNT(*)::int
FROM tse_signaturauftraege
WHERE status IN ('offen', 'fehlgeschlagen');

-- Admin-Ansicht aller Signaturauftraege; dient zugleich als
-- TSE-Ausfalldokumentation (erstellt_am = Beginn, erledigt_am = Ende).
-- name: GetTSESignaturauftraege :many
SELECT id, tx_id, process_type, status, versuche, letzter_fehler, erstellt_am, erledigt_am
FROM tse_signaturauftraege
ORDER BY id DESC
LIMIT 200;

-- name: TSESignaturauftragZuruecksetzen :exec
UPDATE tse_signaturauftraege
SET status = 'offen', versuche = 0, letzter_fehler = NULL, naechster_versuch_am = NOW()
WHERE id = $1 AND status = 'fehlgeschlagen';

-- name: TSESignaturauftragVerwerfen :exec
UPDATE tse_signaturauftraege
SET status = 'verworfen'
WHERE id = $1 AND status = 'fehlgeschlagen';

-- GetTSESignaturauftragZuEvent liefert den Signatur-Stand eines Events fuer den
-- Beleg-Abruf: Status plus Signaturspalten (gefuellt sobald quittiert).
-- Kein Treffer heisst: Das Event ist nicht signaturpflichtig.
-- name: GetTSESignaturauftragZuEvent :one
SELECT status, erstellt_am, transaktion_nummer, signatur_zaehler, tse_seriennummer, log_time_start, log_time_end, signatur, qr_code_data
FROM tse_signaturauftraege
WHERE event_id = $1;
