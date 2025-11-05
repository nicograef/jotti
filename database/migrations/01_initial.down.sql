
BEGIN;

-- Drop tables
DROP TABLE IF EXISTS users;

-- Drop extension last (only if nothing else uses it)
DROP EXTENSION IF EXISTS citext;

COMMIT;
