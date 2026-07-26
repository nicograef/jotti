-- name: GetTisch :one
SELECT id, name, status, created_at, updated_at
FROM tische WHERE id = $1 AND status != 'deleted';

-- name: GetAlleTische :many
SELECT id, name, status, created_at, updated_at
FROM tische WHERE status != 'deleted' ORDER BY id ASC;

-- name: GetAlleTischNamen :many
-- Historische Namensauflösung: ALLE Tische inklusive gelöschter. Der DSFinV-K-
-- Export benennt die Abrechnungskreise vergangener Kassensitzungen; ein nach
-- dem Tagesabschluss gelöschter Tisch muss dort weiterhin unter seinem Namen
-- erscheinen. GetAlleTische filtert 'deleted' bewusst weg und taugt dafür nicht.
SELECT id, name FROM tische ORDER BY id ASC;

-- name: GetTischSaldiOffeneSitzung :many
-- Liefert je Tisch mit offenem Saldo (> 0) den Betrag aus der tisch_sessions-
-- Projektion der aktuell offenen Kassensitzung. Ohne offene Sitzung ist das
-- Ergebnis leer. Reine Projektion; das Kassenjournal bleibt unberührt.
SELECT tss.tisch_id, tss.saldo_cents
FROM tisch_sessions tss
JOIN kassensitzungen k ON k.z_nr = tss.kassensitzung_nr AND k.status = 'offen'
WHERE tss.saldo_cents > 0;

-- name: TischHatOffenenSaldo :one
-- Schutz-Guard: true, wenn der Tisch in der offenen Kassensitzung einen
-- offenen Saldo (> 0) trägt. Ohne offene Sitzung immer false.
SELECT EXISTS (
    SELECT 1
    FROM tisch_sessions tss
    JOIN kassensitzungen k ON k.z_nr = tss.kassensitzung_nr AND k.status = 'offen'
    WHERE tss.tisch_id = $1
      AND tss.saldo_cents > 0
)::bool AS hat_offenen_saldo;

-- name: GetAktiveTische :many
SELECT t.id, t.name, COALESCE(tss.saldo_cents, 0)::integer AS saldo_cents
FROM tische t
LEFT JOIN tisch_sessions tss ON tss.tisch_id = t.id AND tss.kassensitzung_nr = $1
WHERE t.status = 'active'
ORDER BY t.id ASC;

-- name: GetAktiveTischeMitFavoriten :many
SELECT t.id, t.name, COALESCE(tss.saldo_cents, 0)::integer AS saldo_cents, (f.user_id IS NOT NULL)::boolean AS ist_favorit
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
