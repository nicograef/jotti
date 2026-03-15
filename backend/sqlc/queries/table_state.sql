-- name: UpsertTableState :exec
INSERT INTO table_state (tisch_id, saldo_cents, unbezahlte_positionen, ausstehende_positionen, gesamt_zahlungen_cents, last_event_id, last_event_version, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
ON CONFLICT (tisch_id) DO UPDATE SET
    saldo_cents = $2,
    unbezahlte_positionen = $3,
    ausstehende_positionen = $4,
    gesamt_zahlungen_cents = $5,
    last_event_id = $6,
    last_event_version = $7,
    updated_at = NOW();

-- name: GetTableState :one
SELECT tisch_id, saldo_cents, unbezahlte_positionen, ausstehende_positionen, gesamt_zahlungen_cents, last_event_id, last_event_version, updated_at
FROM table_state WHERE tisch_id = $1;

-- name: GetTableStatesByTischIDs :many
SELECT tisch_id, saldo_cents, unbezahlte_positionen, ausstehende_positionen, gesamt_zahlungen_cents, last_event_id, last_event_version, updated_at
FROM table_state WHERE tisch_id = ANY($1::int[]);

-- name: DeleteAllTableState :exec
DELETE FROM table_state;
