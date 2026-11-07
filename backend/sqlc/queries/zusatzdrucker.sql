-- Zusatzdrucker und ihre Produkt-Zuordnung (Migration 19). Die Zuordnung je
-- Variante (GetZusatzdruckerJeVariante) steht handgeschrieben im Repository:
-- Ein int[]-Parameter bräuchte in sqlcs database/sql-Code lib/pq.

-- name: GetAlleZusatzdrucker :many
SELECT id, name, drucker_ip, bonmodus
FROM zusatzdrucker
ORDER BY id;

-- name: GetZusatzdrucker :one
SELECT id, name, drucker_ip, bonmodus
FROM zusatzdrucker
WHERE id = $1;

-- name: CreateZusatzdrucker :one
INSERT INTO zusatzdrucker (name, drucker_ip, bonmodus, updated_at)
VALUES ($1, $2, $3, NOW())
RETURNING id;

-- name: UpdateZusatzdrucker :execresult
UPDATE zusatzdrucker
SET name = $2, drucker_ip = $3, bonmodus = $4, updated_at = NOW()
WHERE id = $1;

-- name: DeleteZusatzdrucker :execresult
DELETE FROM zusatzdrucker
WHERE id = $1;

-- AssignProduktZusatzdrucker fügt nichts ein, wenn das Produkt gelöscht ist
-- oder der Zusatzdrucker fehlt; eine bestehende Zuordnung wandert.
-- name: AssignProduktZusatzdrucker :execresult
INSERT INTO zusatzdrucker_produkte (produkt_id, zusatzdrucker_id)
SELECT p.id, z.id
FROM produkte p, zusatzdrucker z
WHERE p.id = sqlc.arg(produkt_id) AND p.status != 'deleted' AND z.id = sqlc.arg(zusatzdrucker_id)
ON CONFLICT (produkt_id) DO UPDATE SET zusatzdrucker_id = EXCLUDED.zusatzdrucker_id;

-- name: UnassignProduktZusatzdrucker :exec
DELETE FROM zusatzdrucker_produkte
WHERE produkt_id = $1;

-- zusatzdrucker_id 0 und leerer zusatzdrucker_name heißen: keine Zuordnung.
-- name: GetProduktZuordnungen :many
SELECT
    p.id AS produkt_id,
    p.name AS produkt_name,
    p.kategorie,
    COALESCE(z.id, 0)::int AS zusatzdrucker_id,
    COALESCE(z.name, '')::text AS zusatzdrucker_name
FROM produkte p
LEFT JOIN zusatzdrucker_produkte zp ON zp.produkt_id = p.id
LEFT JOIN zusatzdrucker z ON z.id = zp.zusatzdrucker_id
WHERE p.status != 'deleted'
ORDER BY p.kategorie, p.reihenfolge, p.id;
