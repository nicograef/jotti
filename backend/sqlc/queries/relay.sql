-- name: GetBestellungEventsSinceCursor :many
SELECT id, user_name, subject, data, timestamp
FROM events
WHERE type = 'tisch.bestellung-aufgenommen:v1'
  AND id > $1
ORDER BY id ASC
LIMIT 50;
