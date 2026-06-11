-- name: InsertDruckauftrag :exec
INSERT INTO druckauftraege (ziel_ip, payload, status, bon_art, referenz, erstellt_am)
VALUES ($1, $2, 'offen', $3, $4, NOW());

-- name: GetOffeneDruckauftraege :many
SELECT id, ziel_ip, payload
FROM druckauftraege
WHERE status = 'offen'
ORDER BY id ASC
LIMIT 200;

-- name: MarkDruckauftragGedruckt :exec
UPDATE druckauftraege
SET status = 'gedruckt', gedruckt_am = NOW()
WHERE id = $1 AND status = 'offen';

-- name: IncrementDruckauftragFehlversuch :exec
UPDATE druckauftraege
SET versuche = versuche + 1,
    letzter_fehler = sqlc.arg(letzter_fehler),
    status = CASE WHEN versuche + 1 >= sqlc.arg(max_versuche) THEN 'fehlgeschlagen' ELSE status END
WHERE id = sqlc.arg(id) AND status = 'offen';
