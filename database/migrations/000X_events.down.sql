
BEGIN;

-- Drop indexes explicitly (IF EXISTS for idempotency during rollback)
DROP INDEX IF EXISTS idx_events_user_id;
DROP INDEX IF EXISTS idx_events_type_time;
DROP INDEX IF EXISTS idx_events_subject_time;

-- Drop tables (events first due to FK)
DROP TABLE IF EXISTS events;

COMMIT;
