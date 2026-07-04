-- OeffneTSEStoerung oeffnet einen Stoerungszeitraum im Stoerungsprotokoll.
-- Der partielle Unique-Index (hoechstens eine Zeile mit ende IS NULL) macht
-- das Oeffnen idempotent: Bei aktivem Zeitraum ist es ein No-Op.
-- name: OeffneTSEStoerung :exec
INSERT INTO tse_stoerungen (beginn, grund_art, fehlertext)
VALUES (NOW(), $1, $2)
ON CONFLICT DO NOTHING;

-- SchliesseTSEStoerung beendet den aktiven Stoerungszeitraum, falls er die
-- Grund-Art des Schreibers traegt (jeder Schreiber schliesst nur Zeitraeume
-- seiner Grund-Art); sonst ein No-Op.
-- name: SchliesseTSEStoerung :exec
UPDATE tse_stoerungen
SET ende = NOW()
WHERE ende IS NULL AND grund_art = $1;

-- GetAktiveTSEStoerung liefert den aktiven Stoerungszeitraum (hoechstens
-- einer, per partiellem Unique-Index).
-- name: GetAktiveTSEStoerung :one
SELECT beginn, grund_art, fehlertext
FROM tse_stoerungen
WHERE ende IS NULL;

-- GetAlleTSEStoerungen liefert das Stoerungsprotokoll (Ausfalldokumentation):
-- alle Stoerungszeitraeme mit Beginn, Ende und Grund, neueste zuerst.
-- name: GetAlleTSEStoerungen :many
SELECT id, beginn, ende, grund_art, fehlertext
FROM tse_stoerungen
ORDER BY beginn DESC
LIMIT 200;
