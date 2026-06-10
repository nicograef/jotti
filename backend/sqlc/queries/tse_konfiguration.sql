-- name: GetTSEKonfiguration :one
SELECT api_key, api_secret, tss_id, client_id, updated_at
FROM tse_konfiguration
WHERE id = 1;

-- name: UpsertTSEKonfiguration :exec
INSERT INTO tse_konfiguration (id, api_key, api_secret, tss_id, client_id, updated_at)
VALUES (1, $1, $2, $3, $4, NOW())
ON CONFLICT (id) DO UPDATE SET
    api_key = EXCLUDED.api_key,
    api_secret = EXCLUDED.api_secret,
    tss_id = EXCLUDED.tss_id,
    client_id = EXCLUDED.client_id,
    updated_at = EXCLUDED.updated_at;
