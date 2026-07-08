-- Queries für den Demo-Daten-Seeder (backend/seed). Die Insert-Queries setzen die IDs
-- explizit (Stammdaten und Kassensitzungen referenzieren sich gegenseitig), die Reset-Queries
-- ziehen die IDENTITY-Sequenzen anschließend auf den höchsten vergebenen Wert nach.

-- name: SeedCountKassenjournal :one
SELECT COUNT(*)::int AS count FROM kassenjournal;

-- SeedTruncateAll leert alle Daten-Tabellen (Kassenjournal, Projektionen,
-- Stammdaten, TSE-Zustand) und setzt die IDENTITY-Sequenzen zurueck. Nur fuer
-- den Test-Reset-Endpoint (POST /test/reset-and-seed); danach schreibt der
-- Seeder den Demo-Zustand neu. CASCADE loest die Fremdschluessel-Reihenfolge auf.
-- kassenidentitaet bleibt bewusst aussen vor: sie ist die einmalig bei der
-- DB-Migration eingebrannte Install-Identitaet (kein Demo-Datum, insert-once)
-- und wird ausserhalb der Migration nie neu geschrieben.
-- name: SeedTruncateAll :exec
TRUNCATE TABLE
    kassenjournal,
    tisch_sessions,
    kassensitzungen,
    tisch_favoriten,
    produkt_varianten,
    produkte,
    tische,
    users,
    betreiber,
    druckauftraege,
    druckstationen,
    tse_konfiguration,
    tse_stammdaten,
    tse_signaturauftraege,
    tse_stoerungen
RESTART IDENTITY CASCADE;

-- name: SeedInsertUser :exec
INSERT INTO users (id, name, username, password_hash, role, status, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: SeedInsertTisch :exec
INSERT INTO tische (id, name, status, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5);

-- name: SeedInsertProdukt :exec
INSERT INTO produkte (id, name, kategorie, steuersatz, status, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: SeedInsertVariante :exec
INSERT INTO produkt_varianten (id, produkt_id, name, preis_cents, status, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: SeedInsertKassensitzung :exec
INSERT INTO kassensitzungen (z_nr, datum, bezeichnung, status, created_at, updated_at)
OVERRIDING SYSTEM VALUE
VALUES ($1, $2, $3, $4, $5, $6);

-- name: SeedInsertEvent :exec
INSERT INTO kassenjournal (id, user_id, user_name, type, subject, version, data, timestamp, kassensitzung_nr)
OVERRIDING SYSTEM VALUE
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: SeedInsertDruckauftrag :exec
INSERT INTO druckauftraege (ziel_ip, payload, status, bon_art, referenz, versuche, letzter_fehler, erstellt_am, gedruckt_am)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: SeedResetUsersSeq :exec
SELECT setval(pg_get_serial_sequence('users', 'id'), COALESCE((SELECT MAX(id) FROM users), 1));

-- name: SeedResetTischeSeq :exec
SELECT setval(pg_get_serial_sequence('tische', 'id'), COALESCE((SELECT MAX(id) FROM tische), 1));

-- name: SeedResetProdukteSeq :exec
SELECT setval(pg_get_serial_sequence('produkte', 'id'), COALESCE((SELECT MAX(id) FROM produkte), 1));

-- name: SeedResetVariantenSeq :exec
SELECT setval(pg_get_serial_sequence('produkt_varianten', 'id'), COALESCE((SELECT MAX(id) FROM produkt_varianten), 1));

-- name: SeedResetKassensitzungenSeq :exec
SELECT setval(pg_get_serial_sequence('kassensitzungen', 'z_nr'), COALESCE((SELECT MAX(z_nr) FROM kassensitzungen), 1));

-- name: SeedResetKassenjournalSeq :exec
SELECT setval(pg_get_serial_sequence('kassenjournal', 'id'), COALESCE((SELECT MAX(id) FROM kassenjournal), 1));

-- name: SeedInsertTSESignaturauftrag :exec
INSERT INTO tse_signaturauftraege (event_id, tx_id, process_type, process_data, status, versuche, letzter_fehler, naechster_versuch_am, erstellt_am, erledigt_am,
    transaktion_nummer, signatur_zaehler, tse_seriennummer, log_time_start, log_time_end, signatur, qr_code_data)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17);

-- SeedInsertTSEStoerung schreibt einen abgeschlossenen Stoerungszeitraum des
-- Demo-Szenarios (aufgeloestes Ausfallfenster) ins Stoerungsprotokoll.
-- name: SeedInsertTSEStoerung :exec
INSERT INTO tse_stoerungen (beginn, ende, grund_art, fehlertext)
VALUES ($1, $2, $3, $4);
