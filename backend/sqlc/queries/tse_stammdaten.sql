-- name: GetTSEStammdaten :one
SELECT signatur_algorithmus, public_key, zertifikat, log_time_format, updated_at
FROM tse_stammdaten WHERE id = 1;

-- name: UpsertTSEStammdaten :exec
INSERT INTO tse_stammdaten (id, signatur_algorithmus, public_key, zertifikat, log_time_format, updated_at)
VALUES (1, $1, $2, $3, $4, NOW())
ON CONFLICT (id) DO UPDATE SET
    signatur_algorithmus = EXCLUDED.signatur_algorithmus,
    public_key = EXCLUDED.public_key,
    zertifikat = EXCLUDED.zertifikat,
    log_time_format = EXCLUDED.log_time_format,
    updated_at = EXCLUDED.updated_at;
