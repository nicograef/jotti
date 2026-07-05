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
    letzter_fehler = @letzter_fehler,
    status = CASE WHEN versuche + 1 >= @max_versuche THEN 'fehlgeschlagen' ELSE status END
WHERE id = @id AND status = 'offen';

-- name: GetFehlgeschlageneDruckauftraege :many
SELECT id, ziel_ip, bon_art, referenz, versuche, letzter_fehler, erstellt_am
FROM druckauftraege
WHERE status = 'fehlgeschlagen'
ORDER BY id ASC;

-- name: RetryDruckauftrag :exec
UPDATE druckauftraege
SET status = 'offen', versuche = 0, letzter_fehler = NULL
WHERE id = $1 AND status = 'fehlgeschlagen';

-- name: DiscardDruckauftrag :exec
UPDATE druckauftraege
SET status = 'verworfen'
WHERE id = $1 AND status = 'fehlgeschlagen';
