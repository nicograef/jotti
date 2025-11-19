BEGIN;

-- Drop in reverse dependency order
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS tables;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS products;

-- Drop types after dropping tables that depend on them
DROP TYPE IF EXISTS EntityStatus;
DROP TYPE IF EXISTS UserRole;
DROP TYPE IF EXISTS ProductCategory;

COMMIT;
