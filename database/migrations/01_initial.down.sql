BEGIN;

-- Drop in reverse dependency order
DROP TABLE IF EXISTS kategorie_drucker;
DROP TABLE IF EXISTS tisch_favoriten;
DROP TABLE IF EXISTS tisch_sessions;

DROP TRIGGER IF EXISTS kassenjournal_no_truncate ON kassenjournal;
DROP TRIGGER IF EXISTS kassenjournal_no_delete ON kassenjournal;
DROP TRIGGER IF EXISTS kassenjournal_no_update ON kassenjournal;
DROP FUNCTION IF EXISTS prevent_kassenjournal_mutation;
DROP TABLE IF EXISTS kassenjournal;

DROP TABLE IF EXISTS kassensitzungen;
DROP TABLE IF EXISTS produkt_varianten;
DROP TABLE IF EXISTS produkte;
DROP TABLE IF EXISTS tische;
DROP TABLE IF EXISTS users;

-- Drop types after dropping tables that depend on them
DROP TYPE IF EXISTS EntityStatus;
DROP TYPE IF EXISTS UserRole;
DROP TYPE IF EXISTS ProduktKategorie;

COMMIT;
