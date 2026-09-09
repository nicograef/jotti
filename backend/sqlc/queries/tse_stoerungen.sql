-- OpenTSEStoerung öffnet einen Störungszeitraum im Störungsprotokoll.
-- Der partielle Unique-Index (höchstens eine Zeile mit ende IS NULL) macht
-- das Öffnen idempotent: Bei aktivem Zeitraum ist es ein No-Op.
-- name: OpenTSEStoerung :exec
INSERT INTO tse_stoerungen (beginn, grund_art, fehlertext)
VALUES (NOW(), $1, $2)
ON CONFLICT DO NOTHING;

-- CloseTSEStoerung beendet den aktiven Störungszeitraum, falls er die
-- Grund-Art des Schreibers trägt (jeder Schreiber schließt nur Zeiträume
-- seiner Grund-Art); sonst ein No-Op.
-- name: CloseTSEStoerung :exec
UPDATE tse_stoerungen
SET ende = NOW()
WHERE ende IS NULL AND grund_art = $1;

-- GetAktiveTSEStoerung liefert den aktiven Störungszeitraum (höchstens
-- einer, per partiellem Unique-Index).
-- name: GetAktiveTSEStoerung :one
SELECT beginn, grund_art, fehlertext
FROM tse_stoerungen
WHERE ende IS NULL;

-- GetAlleTSEStoerungen liefert das Störungsprotokoll (Ausfalldokumentation):
-- alle Störungszeiträume mit Beginn, Ende und Grund, neueste zuerst.
-- name: GetAlleTSEStoerungen :many
SELECT id, beginn, ende, grund_art, fehlertext
FROM tse_stoerungen
ORDER BY beginn DESC
LIMIT 200;
