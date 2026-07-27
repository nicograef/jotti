-- name: AddFavorit :exec
INSERT INTO tisch_favoriten (user_id, tisch_id, created_at)
VALUES ($1, $2, NOW())
ON CONFLICT DO NOTHING;

-- name: RemoveFavorit :exec
DELETE FROM tisch_favoriten
WHERE user_id = $1 AND tisch_id = $2;

-- name: RemoveFavoritenByTisch :exec
-- Entfernt die Markierungen aller Servicekräfte für einen Tisch. Wird beim
-- Löschen eines Tisches ausgeführt: der gelöschte Tisch erscheint nicht mehr in
-- der Tischauswahl und wäre für die betroffenen Servicekräfte nicht mehr
-- abwählbar.
DELETE FROM tisch_favoriten
WHERE tisch_id = $1;

-- name: GetFavoritenByUser :many
SELECT tisch_id
FROM tisch_favoriten
WHERE user_id = $1
ORDER BY created_at ASC;
