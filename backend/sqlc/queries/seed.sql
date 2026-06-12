-- Queries für den Demo-Daten-Seeder (backend/seed). Die Insert-Queries setzen die IDs
-- explizit (Stammdaten und Kassensitzungen referenzieren sich gegenseitig), die Reset-Queries
-- ziehen die IDENTITY-Sequenzen anschließend auf den höchsten vergebenen Wert nach.

-- name: SeedCountKassenjournal :one
SELECT COUNT(*)::int AS count FROM kassenjournal;

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
VALUES ($1, $2, $3, $4, $5, $6);

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
