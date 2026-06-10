-- name: UpsertTSESignatur :exec
INSERT INTO tse_signaturen (
    tx_id,
    transaktion_nummer,
    signatur_zaehler,
    tse_seriennummer,
    log_time_start,
    log_time_end,
    signatur,
    qr_code_data,
    erstellt_am
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
ON CONFLICT (tx_id) DO UPDATE SET
    transaktion_nummer = EXCLUDED.transaktion_nummer,
    signatur_zaehler = EXCLUDED.signatur_zaehler,
    tse_seriennummer = EXCLUDED.tse_seriennummer,
    log_time_start = EXCLUDED.log_time_start,
    log_time_end = EXCLUDED.log_time_end,
    signatur = EXCLUDED.signatur,
    qr_code_data = EXCLUDED.qr_code_data,
    erstellt_am = EXCLUDED.erstellt_am;

-- name: GetTSESignaturByTxID :one
SELECT tx_id, transaktion_nummer, signatur_zaehler, tse_seriennummer, log_time_start, log_time_end, signatur, qr_code_data
FROM tse_signaturen
WHERE tx_id = $1;
