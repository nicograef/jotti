-- name: GetBetreiber :one
SELECT vereinsname, strasse, plz, ort, steuernummer, ust_id, updated_at
FROM betreiber
LIMIT 1;

-- name: UpsertBetreiber :exec
INSERT INTO betreiber (id, vereinsname, strasse, plz, ort, steuernummer, ust_id, updated_at)
VALUES (1, $1, $2, $3, $4, $5, $6, NOW())
ON CONFLICT (id) DO UPDATE SET
    vereinsname  = EXCLUDED.vereinsname,
    strasse      = EXCLUDED.strasse,
    plz          = EXCLUDED.plz,
    ort          = EXCLUDED.ort,
    steuernummer = EXCLUDED.steuernummer,
    ust_id       = EXCLUDED.ust_id,
    updated_at   = EXCLUDED.updated_at;
