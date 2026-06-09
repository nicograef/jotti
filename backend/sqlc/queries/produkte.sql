-- name: GetProdukt :one
SELECT 
    p.id,
    p.name,
    p.kategorie,
    p.steuersatz,
    p.status,
    p.created_at,
    p.updated_at,
    COALESCE(
        (SELECT json_agg(
            json_build_object(
                'id', pv.id,
                'name', pv.name,
                'preisCents', pv.preis_cents,
                'status', pv.status,
                'createdAt', pv.created_at,
                'updatedAt', pv.updated_at
            )
        )
        FROM produkt_varianten pv
        WHERE pv.produkt_id = p.id AND pv.status != 'deleted'),
        '[]'
    )::json AS varianten
FROM produkte p
WHERE p.id = $1 AND p.status != 'deleted';

-- name: GetAlleProdukte :many
WITH varianten_json AS (
    SELECT 
        produkt_id,
        json_agg(
            json_build_object(
                'id', id,
                'name', name,
                'preisCents', preis_cents,
                'status', status,
                'createdAt', created_at,
                'updatedAt', updated_at
            )
        ) AS varianten
    FROM produkt_varianten
    WHERE status != 'deleted'
    GROUP BY produkt_id
)
SELECT 
    p.id,
    p.name,
    p.kategorie,
    p.steuersatz,
    p.status,
    p.created_at,
    p.updated_at,
    COALESCE(vj.varianten, '[]')::json AS varianten
FROM produkte p
LEFT JOIN varianten_json vj ON vj.produkt_id = p.id
WHERE p.status != 'deleted'
ORDER BY p.id ASC;

-- name: GetAktiveProdukte :many
WITH varianten_json AS (
    SELECT 
        produkt_id,
        json_agg(
            json_build_object(
                'id', id,
                'name', name,
                'preisCents', preis_cents,
                'status', status,
                'createdAt', created_at,
                'updatedAt', updated_at
            )
        ) AS varianten
    FROM produkt_varianten
    WHERE status = 'active'
    GROUP BY produkt_id
)
SELECT 
    p.id,
    p.name,
    p.kategorie,
    p.steuersatz,
    p.status,
    p.created_at,
    p.updated_at,
    vj.varianten::json AS varianten
FROM produkte p
INNER JOIN varianten_json vj ON vj.produkt_id = p.id
WHERE p.status = 'active'
ORDER BY p.id ASC;

-- name: CreateProdukt :one
INSERT INTO produkte (name, kategorie, steuersatz, status, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6) RETURNING id;

-- name: UpdateProdukt :execresult
UPDATE produkte SET name = $1, kategorie = $2, steuersatz = $3, status = $4, updated_at = $5 WHERE id = $6;

-- name: GetVariante :one
SELECT id, name, preis_cents, status, created_at, updated_at
FROM produkt_varianten WHERE id = $1 AND status != 'deleted';

-- name: CreateVariante :one
INSERT INTO produkt_varianten (produkt_id, name, preis_cents, status, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6) RETURNING id;

-- name: UpdateVariante :execresult
UPDATE produkt_varianten SET name = $1, preis_cents = $2, status = $3, updated_at = $4 WHERE id = $5;
