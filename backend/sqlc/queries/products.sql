-- name: GetProduct :one
SELECT 
    p.id,
    p.name,
    p.category,
    p.status,
    p.created_at,
    p.updated_at,
    COALESCE(
        (SELECT json_agg(
            json_build_object(
                'id', pv.id,
                'name', pv.name,
                'price_cents', pv.price_cents,
                'status', pv.status,
                'created_at', pv.created_at,
                'updated_at', pv.updated_at
            )
        )
        FROM product_variants pv
        WHERE pv.product_id = p.id AND pv.status != 'deleted'),
        '[]'
    )::json AS variants
FROM products p
WHERE p.id = $1 AND p.status != 'deleted';

-- name: GetAllProducts :many
WITH variant_json AS (
    SELECT 
        product_id,
        json_agg(
            json_build_object(
                'id', id,
                'name', name,
                'price_cents', price_cents,
                'status', status,
                'created_at', created_at,
                'updated_at', updated_at
            )
        ) AS variants
    FROM product_variants
    WHERE status != 'deleted'
    GROUP BY product_id
)
SELECT 
    p.id,
    p.name,
    p.category,
    p.status,
    p.created_at,
    p.updated_at,
    COALESCE(vj.variants, '[]')::json AS variants
FROM products p
LEFT JOIN variant_json vj ON vj.product_id = p.id
WHERE p.status != 'deleted'
ORDER BY p.id ASC;

-- name: GetActiveProducts :many
WITH variant_json AS (
    SELECT 
        product_id,
        json_agg(
            json_build_object(
                'id', id,
                'name', name,
                'price_cents', price_cents,
                'status', status,
                'created_at', created_at,
                'updated_at', updated_at
            )
        ) AS variants
    FROM product_variants
    WHERE status = 'active'
    GROUP BY product_id
)
SELECT 
    p.id,
    p.name,
    p.category,
    p.status,
    p.created_at,
    p.updated_at,
    vj.variants::json AS variants
FROM products p
INNER JOIN variant_json vj ON vj.product_id = p.id
WHERE p.status = 'active'
ORDER BY p.id ASC;

-- name: CreateProduct :one
INSERT INTO products (name, category, status, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5) RETURNING id;

-- name: UpdateProduct :execresult
UPDATE products SET name = $1, category = $2, status = $3, updated_at = $4 WHERE id = $5;

-- name: GetVariant :one
SELECT id, name, price_cents, status, created_at, updated_at
FROM product_variants WHERE id = $1 AND status != 'deleted';

-- name: CreateVariant :one
INSERT INTO product_variants (product_id, name, price_cents, status, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6) RETURNING id;

-- name: UpdateVariant :execresult
UPDATE product_variants SET name = $1, price_cents = $2, status = $3, updated_at = $4 WHERE id = $5;
