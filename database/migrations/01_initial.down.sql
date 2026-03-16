BEGIN;

-- Drop in reverse dependency order
DROP TRIGGER IF EXISTS events_no_truncate ON events;
DROP TRIGGER IF EXISTS events_no_delete ON events;
DROP TRIGGER IF EXISTS events_no_update ON events;
DROP FUNCTION IF EXISTS prevent_event_mutation;
DROP TABLE IF EXISTS table_state;
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS produkt_varianten;
DROP TABLE IF EXISTS produkte;
DROP TABLE IF EXISTS tisch_favoriten;
DROP TABLE IF EXISTS tische;
DROP TABLE IF EXISTS users;

-- Drop types after dropping tables that depend on them
DROP TYPE IF EXISTS EntityStatus;
DROP TYPE IF EXISTS UserRole;
DROP TYPE IF EXISTS ProductCategory;

COMMIT;
