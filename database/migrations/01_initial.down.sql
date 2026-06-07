BEGIN;

-- Drop in reverse dependency order
DROP TABLE IF EXISTS druckauftraege;
DROP TABLE IF EXISTS druckstationen;
DROP TABLE IF EXISTS tisch_favoriten;
DROP TABLE IF EXISTS tisch_sessions;

-- Drop all write-protection triggers (both tables) before the shared trigger function,
-- which the triggers depend on.
DROP TRIGGER IF EXISTS kassenidentitaet_no_truncate ON kassenidentitaet;
DROP TRIGGER IF EXISTS kassenidentitaet_no_delete ON kassenidentitaet;
DROP TRIGGER IF EXISTS kassenidentitaet_no_update ON kassenidentitaet;
DROP TRIGGER IF EXISTS kassenidentitaet_no_insert ON kassenidentitaet;
DROP TRIGGER IF EXISTS kassenjournal_no_truncate ON kassenjournal;
DROP TRIGGER IF EXISTS kassenjournal_no_delete ON kassenjournal;
DROP TRIGGER IF EXISTS kassenjournal_no_update ON kassenjournal;
DROP FUNCTION IF EXISTS prevent_table_mutation;
DROP FUNCTION IF EXISTS kj_extract_stornierung_cents(TEXT, JSONB);
DROP FUNCTION IF EXISTS kj_extract_bestellung_cents(TEXT, JSONB);
DROP FUNCTION IF EXISTS kj_extract_differenz_cents(TEXT, JSONB);
DROP FUNCTION IF EXISTS kj_extract_geldtransit_cents(TEXT, JSONB);
DROP FUNCTION IF EXISTS kj_extract_auszahlung_cents(TEXT, JSONB);
DROP FUNCTION IF EXISTS kj_extract_eroeffnung_cents(TEXT, JSONB);
DROP FUNCTION IF EXISTS kj_extract_zahlung_cents(TEXT, JSONB);

DROP TABLE IF EXISTS kassenidentitaet;
DROP TABLE IF EXISTS betreiber;
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
