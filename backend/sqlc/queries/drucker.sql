-- name: GetKategorieDrucker :many
SELECT kategorie, drucker_ip, bonmodus
FROM kategorie_drucker;

-- name: GetKonfigurierteKategorieDrucker :many
SELECT kategorie, drucker_ip, bonmodus
FROM kategorie_drucker
WHERE drucker_ip != '';

-- name: UpsertKategorieDrucker :exec
INSERT INTO kategorie_drucker (kategorie, drucker_ip, bonmodus, updated_at)
VALUES ($1, $2, $3, NOW())
ON CONFLICT (kategorie) DO UPDATE SET
    drucker_ip = EXCLUDED.drucker_ip,
    bonmodus = EXCLUDED.bonmodus,
    updated_at = NOW();
