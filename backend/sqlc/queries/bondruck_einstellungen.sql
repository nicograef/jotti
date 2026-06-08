-- name: GetBondruckEinstellungen :one
SELECT kassenbeleg_drucker_ip, direktverkauf_modus, abholbon_drucker_ip, updated_at
FROM bondruck_einstellungen
WHERE id = 1;

-- name: UpsertBondruckEinstellungen :exec
INSERT INTO bondruck_einstellungen (
    id,
    kassenbeleg_drucker_ip,
    direktverkauf_modus,
    abholbon_drucker_ip,
    updated_at
)
VALUES (1, $1, $2, $3, NOW())
ON CONFLICT (id) DO UPDATE SET
    kassenbeleg_drucker_ip = EXCLUDED.kassenbeleg_drucker_ip,
    direktverkauf_modus = EXCLUDED.direktverkauf_modus,
    abholbon_drucker_ip = EXCLUDED.abholbon_drucker_ip,
    updated_at = EXCLUDED.updated_at;
