BEGIN;

-- Drop in reverse dependency order
DROP TRIGGER IF EXISTS events_no_truncate ON events;
DROP TRIGGER IF EXISTS events_no_delete ON events;
DROP TRIGGER IF EXISTS events_no_update ON events;
DROP FUNCTION IF EXISTS prevent_event_mutation;
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS tables;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS product_variants;
DROP TABLE IF EXISTS products;

-- Drop types after dropping tables that depend on them
DROP TYPE IF EXISTS EntityStatus;
DROP TYPE IF EXISTS UserRole;
DROP TYPE IF EXISTS ProductCategory;

COMMIT;
