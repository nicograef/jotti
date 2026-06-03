-- name: GetSystemConfig :one
SELECT seriennummer, angelegt_am
FROM system_config
LIMIT 1;
