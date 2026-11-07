-- name: GetTSEStammdaten :one
SELECT seriennummer, signatur_algorithmus, public_key, zertifikat, log_time_format, updated_at
FROM tse_stammdaten WHERE id = 1;

-- name: UpsertTSEStammdaten :exec
INSERT INTO tse_stammdaten (id, seriennummer, signatur_algorithmus, public_key, zertifikat, log_time_format, updated_at)
VALUES (1, $1, $2, $3, $4, $5, NOW())
ON CONFLICT (id) DO UPDATE SET
    seriennummer = EXCLUDED.seriennummer,
    signatur_algorithmus = EXCLUDED.signatur_algorithmus,
    public_key = EXCLUDED.public_key,
    zertifikat = EXCLUDED.zertifikat,
    log_time_format = EXCLUDED.log_time_format,
    updated_at = EXCLUDED.updated_at;

-- name: UpsertTSEStammdatenJeTSS :exec
-- The serial is the SHA-256 of the public key, so a repeated setup of one TSS only refreshes its row.
INSERT INTO tse_stammdaten_je_tss (seriennummer, signatur_algorithmus, public_key, zertifikat, log_time_format, updated_at)
VALUES ($1, $2, $3, $4, $5, NOW())
ON CONFLICT (seriennummer) DO UPDATE SET
    signatur_algorithmus = EXCLUDED.signatur_algorithmus,
    public_key = EXCLUDED.public_key,
    zertifikat = EXCLUDED.zertifikat,
    log_time_format = EXCLUDED.log_time_format,
    updated_at = EXCLUDED.updated_at;

-- name: GetTSEStammdatenJeTSS :many
SELECT seriennummer, signatur_algorithmus, public_key, zertifikat, log_time_format, updated_at
FROM tse_stammdaten_je_tss ORDER BY seriennummer;
