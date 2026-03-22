-- name: UpsertTischSessionState :exec
INSERT INTO tisch_session_state (subject, tisch_id, kassensitzung_nr, saldo_cents, unbezahlte_positionen, ausstehende_positionen, gesamt_zahlungen_cents, last_event_id, last_event_version, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
ON CONFLICT (subject) DO UPDATE SET
    tisch_id = $2,
    kassensitzung_nr = $3,
    saldo_cents = $4,
    unbezahlte_positionen = $5,
    ausstehende_positionen = $6,
    gesamt_zahlungen_cents = $7,
    last_event_id = $8,
    last_event_version = $9,
    updated_at = NOW();

-- name: GetTischSessionState :one
SELECT subject, tisch_id, kassensitzung_nr, saldo_cents, unbezahlte_positionen, ausstehende_positionen, gesamt_zahlungen_cents, last_event_id, last_event_version, updated_at
FROM tisch_session_state WHERE subject = $1;

-- name: GetTischSessionStatesByKassensitzungNr :many
SELECT subject, tisch_id, kassensitzung_nr, saldo_cents, unbezahlte_positionen, ausstehende_positionen, gesamt_zahlungen_cents, last_event_id, last_event_version, updated_at
FROM tisch_session_state WHERE kassensitzung_nr = $1;

-- name: DeleteAllTischSessionState :exec
DELETE FROM tisch_session_state;
