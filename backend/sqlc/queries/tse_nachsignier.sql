-- name: InsertTSENachsignierAuftrag :exec
INSERT INTO tse_nachsignier_auftraege (tx_id, process_type, process_data, status, naechster_versuch_am, erstellt_am)
VALUES ($1, $2, $3, 'offen', NOW(), NOW())
ON CONFLICT (tx_id) DO NOTHING;

-- name: GetOffeneTSENachsignierAuftraege :many
SELECT id, tx_id, process_type, process_data
FROM tse_nachsignier_auftraege
WHERE status = 'offen'
  AND naechster_versuch_am <= NOW()
ORDER BY id ASC
LIMIT $1;

-- name: MarkTSENachsignierAuftragErledigt :exec
UPDATE tse_nachsignier_auftraege
SET status = 'erledigt',
    erledigt_am = NOW()
WHERE id = $1
  AND status = 'offen';

-- TSENachsignierAuftragFehlversuch verbucht einen Fehlversuch mit exponentiellem
-- Backoff (1, 2, 4, ... Minuten, gedeckelt auf 30). Beim max_versuche-ten
-- Fehlversuch wechselt der Auftrag auf fehlgeschlagen und wird nicht mehr
-- automatisch versucht.
-- name: TSENachsignierAuftragFehlversuch :exec
UPDATE tse_nachsignier_auftraege
SET versuche = versuche + 1,
    letzter_fehler = @letzter_fehler,
    naechster_versuch_am = NOW() + LEAST(POWER(2, versuche), 30) * interval '1 minute',
    status = CASE WHEN versuche + 1 >= @max_versuche THEN 'fehlgeschlagen' ELSE status END
WHERE id = @id AND status = 'offen';

-- Zaehlt noch nicht erledigte Nachsignierungen (offen und fehlgeschlagen):
-- beide Status bedeuten unsignierte Vorgaenge.
-- name: CountOffeneTSENachsignierAuftraege :one
SELECT COUNT(*)::int
FROM tse_nachsignier_auftraege
WHERE status IN ('offen', 'fehlgeschlagen');

-- Admin-Ansicht aller Nachsignier-Auftraege; dient zugleich als
-- TSE-Ausfalldokumentation (erstellt_am = Beginn, erledigt_am = Ende).
-- name: GetTSENachsignierAuftraege :many
SELECT id, tx_id, process_type, status, versuche, letzter_fehler, erstellt_am, erledigt_am
FROM tse_nachsignier_auftraege
ORDER BY id DESC
LIMIT 200;

-- name: TSENachsignierAuftragZuruecksetzen :exec
UPDATE tse_nachsignier_auftraege
SET status = 'offen', versuche = 0, letzter_fehler = NULL, naechster_versuch_am = NOW()
WHERE id = $1 AND status = 'fehlgeschlagen';

-- name: TSENachsignierAuftragVerwerfen :exec
UPDATE tse_nachsignier_auftraege
SET status = 'verworfen'
WHERE id = $1 AND status = 'fehlgeschlagen';
