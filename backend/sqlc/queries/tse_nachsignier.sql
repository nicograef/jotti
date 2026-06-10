-- name: InsertTSENachsignierAuftrag :exec
INSERT INTO tse_nachsignier_auftraege (tx_id, process_type, process_data, status, erstellt_am)
VALUES ($1, $2, $3, 'offen', NOW())
ON CONFLICT (tx_id) DO NOTHING;

-- name: GetOffeneTSENachsignierAuftraege :many
SELECT id, tx_id, process_type, process_data
FROM tse_nachsignier_auftraege
WHERE status = 'offen'
ORDER BY id ASC
LIMIT $1;

-- name: MarkTSENachsignierAuftragErledigt :exec
UPDATE tse_nachsignier_auftraege
SET status = 'erledigt',
    erledigt_am = NOW()
WHERE id = $1;

-- name: CountOffeneTSENachsignierAuftraege :one
SELECT COUNT(*)::int
FROM tse_nachsignier_auftraege
WHERE status = 'offen';
