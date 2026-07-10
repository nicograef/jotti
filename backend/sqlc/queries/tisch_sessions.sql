-- name: UpsertTischSession :exec
INSERT INTO tisch_sessions (subject, tisch_id, kassensitzung_nr, saldo_cents, unbezahlte_positionen, gesamt_zahlungen_cents, erste_bestellung_logtime, last_event_id, last_event_version, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
ON CONFLICT (subject) DO UPDATE SET
    tisch_id = $2,
    kassensitzung_nr = $3,
    saldo_cents = $4,
    unbezahlte_positionen = $5,
    gesamt_zahlungen_cents = $6,
    erste_bestellung_logtime = $7,
    last_event_id = $8,
    last_event_version = $9,
    updated_at = NOW();

-- name: GetTischSession :one
SELECT subject, tisch_id, kassensitzung_nr, saldo_cents, unbezahlte_positionen, gesamt_zahlungen_cents, erste_bestellung_logtime, last_event_id, last_event_version, updated_at
FROM tisch_sessions WHERE subject = $1;

-- name: GetTischSessionsByKassensitzungNr :many
SELECT subject, tisch_id, kassensitzung_nr, saldo_cents, unbezahlte_positionen, gesamt_zahlungen_cents, erste_bestellung_logtime, last_event_id, last_event_version, updated_at
FROM tisch_sessions WHERE kassensitzung_nr = $1;

-- name: DeleteAllTischSession :exec
DELETE FROM tisch_sessions;
