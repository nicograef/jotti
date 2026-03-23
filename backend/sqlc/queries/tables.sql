-- name: GetTisch :one
SELECT id, name, status, created_at, updated_at
FROM tische WHERE id = $1 AND status != 'deleted';

-- name: GetAlleTische :many
SELECT id, name, status, created_at, updated_at
FROM tische WHERE status != 'deleted' ORDER BY id ASC;

-- name: GetAktiveTische :many
SELECT t.id, t.name, COALESCE(tss.saldo_cents, 0)::integer AS saldo_cents
FROM tische t
LEFT JOIN tisch_sessions tss ON tss.tisch_id = t.id AND tss.kassensitzung_nr = $1
WHERE t.status = 'active'
ORDER BY t.id ASC;

-- name: GetAktiveTischeMitFavoriten :many
SELECT t.id, t.name, COALESCE(tss.saldo_cents, 0)::integer AS saldo_cents, (f.user_id IS NOT NULL) AS ist_favorit
FROM tische t
LEFT JOIN tisch_sessions tss ON tss.tisch_id = t.id AND tss.kassensitzung_nr = $2
LEFT JOIN tisch_favoriten f ON f.tisch_id = t.id AND f.user_id = $1
WHERE t.status = 'active'
ORDER BY t.id ASC;

-- name: CreateTisch :one
INSERT INTO tische (name, status, created_at, updated_at)
VALUES ($1, $2, $3, $4) RETURNING id;

-- name: UpdateTisch :execresult
UPDATE tische SET name = $1, status = $2, updated_at = $3 WHERE id = $4;
