-- name: GetKassenidentitaet :one
SELECT seriennummer, angelegt_am
FROM kassenidentitaet
LIMIT 1;
