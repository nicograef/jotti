-- name: AddFavorit :exec
INSERT INTO tisch_favoriten (user_id, tisch_id, created_at)
VALUES ($1, $2, NOW())
ON CONFLICT DO NOTHING;

-- name: RemoveFavorit :exec
DELETE FROM tisch_favoriten
WHERE user_id = $1 AND tisch_id = $2;

-- name: GetFavoritenByUser :many
SELECT tisch_id
FROM tisch_favoriten
WHERE user_id = $1
ORDER BY created_at ASC;
