-- name: InsertDruckauftrag :exec
INSERT INTO druckauftraege (ziel_ip, payload, status, bon_art, referenz, erstellt_am)
VALUES ($1, $2, 'offen', $3, $4, NOW());

-- name: GetOffeneDruckauftraege :many
-- Eine Ziel-IP wird komplett uebersprungen, solange irgendein offener Auftrag
-- dieses Druckers noch auf seine Backoff-Wartezeit wartet. Sonst wuerde ein
-- waehrend des Backoff-Fensters neu eingereihter Auftrag (naechster_versuch_ab
-- ist dann NULL) die gebremste Warteschlange ueberholen und die Bon-Reihenfolge
-- brechen.
SELECT id, ziel_ip, payload
FROM druckauftraege
WHERE status = 'offen'
  AND ziel_ip NOT IN (
    SELECT ziel_ip
    FROM druckauftraege
    WHERE status = 'offen' AND naechster_versuch_ab > NOW()
  )
ORDER BY id ASC
LIMIT 200;

-- name: MarkDruckauftragGedruckt :exec
UPDATE druckauftraege
SET status = 'gedruckt', gedruckt_am = NOW()
WHERE id = $1 AND status = 'offen';

-- name: IncrementDruckauftragFehlversuch :one
UPDATE druckauftraege
SET versuche = versuche + 1,
    letzter_fehler = @letzter_fehler,
    status = CASE WHEN versuche + 1 >= @max_versuche THEN 'fehlgeschlagen' ELSE status END
WHERE id = @id AND status = 'offen'
RETURNING versuche, status;

-- name: SetDruckauftragFaelligkeit :exec
UPDATE druckauftraege
SET naechster_versuch_ab = NOW() + (sqlc.arg(sekunden)::int * INTERVAL '1 second')
WHERE id = sqlc.arg(id) AND status = 'offen';

-- name: GetFehlgeschlageneDruckauftraege :many
SELECT id, ziel_ip, bon_art, referenz, versuche, letzter_fehler, erstellt_am
FROM druckauftraege
WHERE status = 'fehlgeschlagen'
ORDER BY id ASC;

-- name: RetryDruckauftrag :exec
UPDATE druckauftraege
SET status = 'offen', versuche = 0, letzter_fehler = NULL, naechster_versuch_ab = NULL
WHERE id = $1 AND status = 'fehlgeschlagen';

-- name: DiscardDruckauftrag :exec
UPDATE druckauftraege
SET status = 'verworfen'
WHERE id = $1 AND status = 'fehlgeschlagen';

-- name: DiscardAlleFehlgeschlagenenDruckauftraege :execrows
UPDATE druckauftraege
SET status = 'verworfen'
WHERE status = 'fehlgeschlagen';
