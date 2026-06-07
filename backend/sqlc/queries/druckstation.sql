-- name: GetDruckstationen :many
SELECT kategorie, drucker_ip, bonmodus
FROM druckstationen;

-- name: GetKonfigurierteDruckstationen :many
SELECT kategorie, drucker_ip, bonmodus
FROM druckstationen
WHERE drucker_ip != '';

-- name: UpsertDruckstation :exec
INSERT INTO druckstationen (kategorie, drucker_ip, bonmodus, updated_at)
VALUES ($1, $2, $3, NOW())
ON CONFLICT (kategorie) DO UPDATE SET
    drucker_ip = EXCLUDED.drucker_ip,
    bonmodus = EXCLUDED.bonmodus,
    updated_at = NOW();
