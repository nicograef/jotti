-- =============================================================================
-- jotti Seed-Daten: "3-Tage Sommerfest TSV Musterstadt"
-- =============================================================================
-- Szenario: Ein dreitägiges Vereinsfest (Freitag bis Sonntag). 20 aktive Tische
-- in verschiedenen Zuständen — komplett abgeschlossen, teilgeliefert,
-- teilbezahlt, frisch bestellt, storniert, leer. ~1000 Events über 3 Tage
-- verteilt, mit realistischen Produkten und verschiedenen Servicekräften.
--
-- Voraussetzung: Frische DB nach Schema-Migration (01_initial.up.sql).
-- Der Admin-User "nico" (id=1) existiert bereits aus der Migration.
--
-- Passwort aller Seed-User: jotti123
-- Einmalpasswort von "nico" (aus Migration): 123456
-- =============================================================================

BEGIN;

-- =============================================================================
-- 1. BENUTZER (10 neue, id=1 existiert aus Migration)
-- =============================================================================
-- Passwort-Hash für "jotti123" (Argon2id, m=65536, t=2, p=2)

INSERT INTO users (id, name, username, password_hash, role, status, created_at) VALUES
  (2, 'Thomas Müller', 'thomas', '$argon2id$v=19$m=65536,t=2,p=2$OSImYG1ms0Phs26KwwMwkQ$rkoWKOIjsPz7y6ps/W2pVEhn5vTc0N95SyiveQCn404', 'admin', 'active', now()),
  (3, 'Felix Weber', 'felix', '$argon2id$v=19$m=65536,t=2,p=2$OSImYG1ms0Phs26KwwMwkQ$rkoWKOIjsPz7y6ps/W2pVEhn5vTc0N95SyiveQCn404', 'serviceleitung', 'active', now()),
  (4, 'Maria Schmidt', 'maria', '$argon2id$v=19$m=65536,t=2,p=2$OSImYG1ms0Phs26KwwMwkQ$rkoWKOIjsPz7y6ps/W2pVEhn5vTc0N95SyiveQCn404', 'service', 'active', now()),
  (5, 'Lisa Braun', 'lisa', '$argon2id$v=19$m=65536,t=2,p=2$OSImYG1ms0Phs26KwwMwkQ$rkoWKOIjsPz7y6ps/W2pVEhn5vTc0N95SyiveQCn404', 'service', 'active', now()),
  (6, 'Jan Hoffmann', 'jan', '$argon2id$v=19$m=65536,t=2,p=2$OSImYG1ms0Phs26KwwMwkQ$rkoWKOIjsPz7y6ps/W2pVEhn5vTc0N95SyiveQCn404', 'service', 'active', now()),
  (7, 'Sophie Becker', 'sophie', '$argon2id$v=19$m=65536,t=2,p=2$OSImYG1ms0Phs26KwwMwkQ$rkoWKOIjsPz7y6ps/W2pVEhn5vTc0N95SyiveQCn404', 'serviceleitung', 'active', now()),
  (8, 'Markus Lehmann', 'markus', '$argon2id$v=19$m=65536,t=2,p=2$OSImYG1ms0Phs26KwwMwkQ$rkoWKOIjsPz7y6ps/W2pVEhn5vTc0N95SyiveQCn404', 'service', 'active', now()),
  (9, 'Anna Krause', 'anna', '$argon2id$v=19$m=65536,t=2,p=2$OSImYG1ms0Phs26KwwMwkQ$rkoWKOIjsPz7y6ps/W2pVEhn5vTc0N95SyiveQCn404', 'service', 'active', now()),
  (10, 'Paul Fischer', 'paul', '$argon2id$v=19$m=65536,t=2,p=2$OSImYG1ms0Phs26KwwMwkQ$rkoWKOIjsPz7y6ps/W2pVEhn5vTc0N95SyiveQCn404', 'service', 'inactive', now()),
  (11, 'Sabine Wolf', 'sabine', '$argon2id$v=19$m=65536,t=2,p=2$OSImYG1ms0Phs26KwwMwkQ$rkoWKOIjsPz7y6ps/W2pVEhn5vTc0N95SyiveQCn404', 'service', 'deleted', now());

SELECT setval(pg_get_serial_sequence('users', 'id'), (SELECT MAX(id) FROM users));


-- =============================================================================
-- 2. TISCHE (22: 20 aktiv, 1 inaktiv, 1 gelöscht)
-- =============================================================================

INSERT INTO tables (id, name, status, created_at) VALUES
  (1, 'Tisch 1', 'active', now()),
  (2, 'Tisch 2', 'active', now()),
  (3, 'Tisch 3', 'active', now()),
  (4, 'Tisch 4', 'active', now()),
  (5, 'Tisch 5', 'active', now()),
  (6, 'Tisch 6', 'active', now()),
  (7, 'Tisch 7', 'active', now()),
  (8, 'Tisch 8', 'active', now()),
  (9, 'Tisch 9', 'active', now()),
  (10, 'Tisch 10', 'active', now()),
  (11, 'Tisch 11', 'active', now()),
  (12, 'Tisch 12', 'active', now()),
  (13, 'Tisch 13', 'active', now()),
  (14, 'Tisch 14', 'active', now()),
  (15, 'Tisch 15', 'active', now()),
  (16, 'Zelt A1', 'active', now()),
  (17, 'Zelt A2', 'active', now()),
  (18, 'Stehtisch Bar', 'active', now()),
  (19, 'Stehtisch Eingang', 'active', now()),
  (20, 'Stehtisch Terrasse', 'active', now()),
  (21, 'Reserviert', 'inactive', now()),
  (22, 'Alter Tisch', 'deleted', now());

SELECT setval(pg_get_serial_sequence('tables', 'id'), (SELECT MAX(id) FROM tables));


-- =============================================================================
-- 3. PRODUKTE & VARIANTEN
-- =============================================================================
-- 10 Essen, 10 Getränke, 2 Sonstiges = 22 Produkte
-- 54 Varianten insgesamt

INSERT INTO products (id, name, category, status, created_at) VALUES
  (1, 'Bratwurst', 'food', 'active', now()),
  (2, 'Pommes', 'food', 'active', now()),
  (3, 'Flammkuchen', 'food', 'active', now()),
  (4, 'Tagesgericht', 'food', 'active', now()),
  (5, 'Grillplatte', 'food', 'active', now()),
  (6, 'Salat', 'food', 'active', now()),
  (7, 'Kuchen', 'food', 'active', now()),
  (8, 'Waffeln', 'food', 'active', now()),
  (9, 'Brezel', 'food', 'active', now()),
  (10, 'Suppe', 'food', 'inactive', now()),
  (11, 'Bier', 'beverage', 'active', now()),
  (12, 'Weizen', 'beverage', 'active', now()),
  (13, 'Softdrinks', 'beverage', 'active', now()),
  (14, 'Wasser', 'beverage', 'active', now()),
  (15, 'Saftschorle', 'beverage', 'active', now()),
  (16, 'Wein', 'beverage', 'active', now()),
  (17, 'Kaffee', 'beverage', 'active', now()),
  (18, 'Tee', 'beverage', 'active', now()),
  (19, 'Hugo/Aperol', 'beverage', 'active', now()),
  (20, 'Glühwein', 'beverage', 'inactive', now()),
  (21, 'Festbändchen', 'other', 'active', now()),
  (22, 'Langos', 'food', 'deleted', now());

INSERT INTO product_variants (id, product_id, name, price_cents, status, created_at) VALUES
  (1, 1, 'Normal', 350, 'active', now()),
  (2, 1, 'XXL', 500, 'active', now()),
  (3, 1, 'Currywurst', 450, 'active', now()),
  (4, 2, 'Klein', 250, 'active', now()),
  (5, 2, 'Groß', 350, 'active', now()),
  (6, 3, 'Classic', 600, 'active', now()),
  (7, 3, 'Speck & Zwiebel', 700, 'active', now()),
  (8, 3, 'Mediterran', 750, 'active', now()),
  (9, 4, 'Fr: Schnitzel mit Pommes', 1250, 'active', now()),
  (10, 4, 'Sa: Gulasch mit Spätzle', 1150, 'active', now()),
  (11, 4, 'So: Hähnchen mit Reis', 1050, 'active', now()),
  (12, 5, 'Klein', 800, 'active', now()),
  (13, 5, 'Groß', 1400, 'active', now()),
  (14, 6, 'Gemischter Salat', 550, 'active', now()),
  (15, 6, 'Caesar Salat', 650, 'active', now()),
  (16, 7, 'Stück', 250, 'active', now()),
  (17, 8, 'mit Puderzucker', 300, 'active', now()),
  (18, 8, 'mit Sahne', 350, 'active', now()),
  (19, 8, 'mit Nutella', 400, 'active', now()),
  (20, 9, 'Normal', 200, 'active', now()),
  (21, 9, 'mit Butter', 300, 'active', now()),
  (22, 10, 'Tagessuppe', 400, 'active', now()),
  (23, 11, '0,3l', 300, 'active', now()),
  (24, 11, '0,5l', 450, 'active', now()),
  (25, 11, 'Maß 1,0l', 850, 'active', now()),
  (26, 12, 'Klein 0,3l', 300, 'active', now()),
  (27, 12, 'Groß 0,5l', 400, 'active', now()),
  (28, 12, 'Colaweizen Klein', 300, 'active', now()),
  (29, 12, 'Colaweizen Groß', 400, 'active', now()),
  (30, 12, 'Russ', 300, 'active', now()),
  (31, 13, 'Cola', 280, 'active', now()),
  (32, 13, 'Fanta', 280, 'active', now()),
  (33, 13, 'Spezi', 280, 'active', now()),
  (34, 13, 'Sprite', 280, 'active', now()),
  (35, 13, 'Mezzo Mix', 280, 'active', now()),
  (36, 14, 'Still 0,5l', 200, 'active', now()),
  (37, 14, 'Medium 0,5l', 200, 'active', now()),
  (38, 14, 'Sprudel 0,5l', 200, 'active', now()),
  (39, 15, 'Apfelschorle 0,5l', 300, 'active', now()),
  (40, 15, 'Johannisbeerschorle 0,5l', 300, 'active', now()),
  (41, 15, 'Rhabarberschorle 0,5l', 350, 'active', now()),
  (42, 16, 'Weißwein 0,2l', 400, 'active', now()),
  (43, 16, 'Rotwein 0,2l', 400, 'active', now()),
  (44, 16, 'Rosé 0,2l', 400, 'active', now()),
  (45, 17, 'Tasse', 200, 'active', now()),
  (46, 17, 'Espresso', 180, 'active', now()),
  (47, 18, 'Verschiedene Sorten', 200, 'active', now()),
  (48, 19, 'Hugo', 550, 'active', now()),
  (49, 19, 'Aperol Spritz', 550, 'active', now()),
  (50, 20, 'Tasse', 350, 'active', now()),
  (51, 21, 'Erwachsene', 500, 'active', now()),
  (52, 21, 'Kinder', 300, 'active', now()),
  (53, 22, 'mit Knoblauch', 400, 'deleted', now()),
  (54, 22, 'mit Käse', 500, 'deleted', now());

SELECT setval(pg_get_serial_sequence('products', 'id'), (SELECT MAX(id) FROM products));
SELECT setval(pg_get_serial_sequence('product_variants', 'id'), (SELECT MAX(id) FROM product_variants));


-- =============================================================================
-- 4. EVENTS
-- =============================================================================
-- Events über 3 Tage verteilt:
--   Tag 1 (Freitag):  Eröffnungsabend, gemütlicher Start
--   Tag 2 (Samstag):  Haupttag, voller Betrieb
--   Tag 3 (Sonntag):  Aktueller Tag, Nachmittag, einige Tische noch offen
--
-- Zeitstempel: now() - interval 'X' (X abnehmend = zeitlich fortschreitend)
-- Tagesgericht je Tag: Fr=Schnitzel, Sa=Gulasch, So=Hähnchen
-- =============================================================================

-- ─────────────────────────────────────────────────────────────────────────────
-- TISCH 1: Stammtisch — Dauergäste, alle 3 Tage (99 Events)
-- ─────────────────────────────────────────────────────────────────────────────

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:1', 1, now() - interval '48 hours 20 minutes', '{"bestellungId": "a0001000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000001","varianteId": 1,"produktName": "Bratwurst","varianteName": "Normal","kategorie": "food","einzelpreis": 350,"menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000002","varianteId": 5,"produktName": "Pommes","varianteName": "Groß","kategorie": "food","einzelpreis": 350,"menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000003","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 4}],"gesamtPreisCents": 3200,"kommentar": "Stammtisch Freitagabend"}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:1', 2, now() - interval '48 hours 5 minutes', '{"lieferungId": "c0001000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000001","menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000002","menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000003","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:1', 3, now() - interval '47 hours 51 minutes', '{"zahlungId": "d0001000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000001","menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000002","menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000003","menge": 4}],"gesamtZahlungCents": 3200,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:1', 4, now() - interval '47 hours 22 minutes', '{"bestellungId": "a0001000-0000-0000-0000-000000000002","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000004","varianteId": 7,"produktName": "Flammkuchen","varianteName": "Speck & Zwiebel","kategorie": "food","einzelpreis": 700,"menge": 1},{"positionId": "b0001000-0000-0000-0000-000000000005","varianteId": 33,"produktName": "Softdrinks","varianteName": "Spezi","kategorie": "beverage","einzelpreis": 280,"menge": 2}],"gesamtPreisCents": 1260,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:1', 5, now() - interval '47 hours 11 minutes', '{"lieferungId": "c0001000-0000-0000-0000-000000000002","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000004","menge": 1},{"positionId": "b0001000-0000-0000-0000-000000000005","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:1', 6, now() - interval '46 hours 48 minutes', '{"zahlungId": "d0001000-0000-0000-0000-000000000002","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000004","menge": 1},{"positionId": "b0001000-0000-0000-0000-000000000005","menge": 2}],"gesamtZahlungCents": 1260,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:1', 7, now() - interval '46 hours 22 minutes', '{"bestellungId": "a0001000-0000-0000-0000-000000000003","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000006","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 4}],"gesamtPreisCents": 1800,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:1', 8, now() - interval '46 hours 10 minutes', '{"lieferungId": "c0001000-0000-0000-0000-000000000003","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000006","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:1', 9, now() - interval '45 hours 59 minutes', '{"zahlungId": "d0001000-0000-0000-0000-000000000003","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000006","menge": 4}],"gesamtZahlungCents": 1800,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.bestellung-aufgegeben:v1', 'tisch:1', 10, now() - interval '45 hours 29 minutes', '{"bestellungId": "a0001000-0000-0000-0000-000000000004","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000007","varianteId": 16,"produktName": "Kuchen","varianteName": "Stück","kategorie": "food","einzelpreis": 250,"menge": 3},{"positionId": "b0001000-0000-0000-0000-000000000008","varianteId": 45,"produktName": "Kaffee","varianteName": "Tasse","kategorie": "beverage","einzelpreis": 200,"menge": 3}],"gesamtPreisCents": 1350,"kommentar": "Nachtisch"}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.produkte-geliefert:v1', 'tisch:1', 11, now() - interval '45 hours 14 minutes', '{"lieferungId": "c0001000-0000-0000-0000-000000000004","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000007","menge": 3},{"positionId": "b0001000-0000-0000-0000-000000000008","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.zahlung-registriert:v1', 'tisch:1', 12, now() - interval '44 hours 57 minutes', '{"zahlungId": "d0001000-0000-0000-0000-000000000004","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000007","menge": 3},{"positionId": "b0001000-0000-0000-0000-000000000008","menge": 3}],"gesamtZahlungCents": 1350,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.bestellung-aufgegeben:v1', 'tisch:1', 13, now() - interval '44 hours 31 minutes', '{"bestellungId": "a0001000-0000-0000-0000-000000000005","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000009","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000010","varianteId": 23,"produktName": "Bier","varianteName": "0,3l","kategorie": "beverage","einzelpreis": 300,"menge": 2}],"gesamtPreisCents": 1500,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.produkte-geliefert:v1', 'tisch:1', 14, now() - interval '44 hours 20 minutes', '{"lieferungId": "c0001000-0000-0000-0000-000000000005","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000009","menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000010","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.zahlung-registriert:v1', 'tisch:1', 15, now() - interval '44 hours 4 minutes', '{"zahlungId": "d0001000-0000-0000-0000-000000000005","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000009","menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000010","menge": 2}],"gesamtZahlungCents": 1500,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:1', 16, now() - interval '27 hours', '{"bestellungId": "a0001000-0000-0000-0000-000000000006","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000011","varianteId": 13,"produktName": "Grillplatte","varianteName": "Groß","kategorie": "food","einzelpreis": 1400,"menge": 1},{"positionId": "b0001000-0000-0000-0000-000000000012","varianteId": 6,"produktName": "Flammkuchen","varianteName": "Classic","kategorie": "food","einzelpreis": 600,"menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000013","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 6},{"positionId": "b0001000-0000-0000-0000-000000000014","varianteId": 48,"produktName": "Hugo/Aperol","varianteName": "Hugo","kategorie": "beverage","einzelpreis": 550,"menge": 2}],"gesamtPreisCents": 6400,"kommentar": "Geburtstagsfeier!"}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.produkte-geliefert:v1', 'tisch:1', 17, now() - interval '26 hours 51 minutes', '{"lieferungId": "c0001000-0000-0000-0000-000000000006","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000011","menge": 1},{"positionId": "b0001000-0000-0000-0000-000000000012","menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000013","menge": 6},{"positionId": "b0001000-0000-0000-0000-000000000014","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.zahlung-registriert:v1', 'tisch:1', 18, now() - interval '26 hours 28 minutes', '{"zahlungId": "d0001000-0000-0000-0000-000000000006","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000011","menge": 1},{"positionId": "b0001000-0000-0000-0000-000000000012","menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000013","menge": 6},{"positionId": "b0001000-0000-0000-0000-000000000014","menge": 2}],"gesamtZahlungCents": 6400,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:1', 19, now() - interval '26 hours 10 minutes', '{"bestellungId": "a0001000-0000-0000-0000-000000000007","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000015","varianteId": 1,"produktName": "Bratwurst","varianteName": "Normal","kategorie": "food","einzelpreis": 350,"menge": 4},{"positionId": "b0001000-0000-0000-0000-000000000016","varianteId": 5,"produktName": "Pommes","varianteName": "Groß","kategorie": "food","einzelpreis": 350,"menge": 4},{"positionId": "b0001000-0000-0000-0000-000000000017","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 4},{"positionId": "b0001000-0000-0000-0000-000000000018","varianteId": 33,"produktName": "Softdrinks","varianteName": "Spezi","kategorie": "beverage","einzelpreis": 280,"menge": 4}],"gesamtPreisCents": 5720,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.produkte-geliefert:v1', 'tisch:1', 20, now() - interval '25 hours 55 minutes', '{"lieferungId": "c0001000-0000-0000-0000-000000000007","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000015","menge": 4},{"positionId": "b0001000-0000-0000-0000-000000000016","menge": 4},{"positionId": "b0001000-0000-0000-0000-000000000017","menge": 4},{"positionId": "b0001000-0000-0000-0000-000000000018","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.bestellung-aufgegeben:v1', 'tisch:1', 21, now() - interval '25 hours 35 minutes', '{"bestellungId": "a0001000-0000-0000-0000-000000000008","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000019","varianteId": 7,"produktName": "Flammkuchen","varianteName": "Speck & Zwiebel","kategorie": "food","einzelpreis": 700,"menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000020","varianteId": 42,"produktName": "Wein","varianteName": "Weißwein 0,2l","kategorie": "beverage","einzelpreis": 400,"menge": 3},{"positionId": "b0001000-0000-0000-0000-000000000021","varianteId": 49,"produktName": "Hugo/Aperol","varianteName": "Aperol Spritz","kategorie": "beverage","einzelpreis": 550,"menge": 2}],"gesamtPreisCents": 3700,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.produkte-geliefert:v1', 'tisch:1', 22, now() - interval '25 hours 24 minutes', '{"lieferungId": "c0001000-0000-0000-0000-000000000008","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000019","menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000020","menge": 3},{"positionId": "b0001000-0000-0000-0000-000000000021","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.zahlung-registriert:v1', 'tisch:1', 23, now() - interval '24 hours 55 minutes', '{"zahlungId": "d0001000-0000-0000-0000-000000000007","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000015","menge": 4},{"positionId": "b0001000-0000-0000-0000-000000000016","menge": 4},{"positionId": "b0001000-0000-0000-0000-000000000017","menge": 4},{"positionId": "b0001000-0000-0000-0000-000000000018","menge": 4},{"positionId": "b0001000-0000-0000-0000-000000000019","menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000020","menge": 3},{"positionId": "b0001000-0000-0000-0000-000000000021","menge": 2}],"gesamtZahlungCents": 9420,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:1', 24, now() - interval '24 hours 35 minutes', '{"bestellungId": "a0001000-0000-0000-0000-000000000009","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000022","varianteId": 16,"produktName": "Kuchen","varianteName": "Stück","kategorie": "food","einzelpreis": 250,"menge": 4},{"positionId": "b0001000-0000-0000-0000-000000000023","varianteId": 45,"produktName": "Kaffee","varianteName": "Tasse","kategorie": "beverage","einzelpreis": 200,"menge": 4}],"gesamtPreisCents": 1800,"kommentar": "Kuchen zum Geburtstag"}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.produkte-geliefert:v1', 'tisch:1', 25, now() - interval '24 hours 27 minutes', '{"lieferungId": "c0001000-0000-0000-0000-000000000009","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000022","menge": 4},{"positionId": "b0001000-0000-0000-0000-000000000023","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.zahlung-registriert:v1', 'tisch:1', 26, now() - interval '24 hours 3 minutes', '{"zahlungId": "d0001000-0000-0000-0000-000000000008","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000022","menge": 4},{"positionId": "b0001000-0000-0000-0000-000000000023","menge": 4}],"gesamtZahlungCents": 1800,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.bestellung-aufgegeben:v1', 'tisch:1', 27, now() - interval '23 hours 43 minutes', '{"bestellungId": "a0001000-0000-0000-0000-000000000010","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000024","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 6},{"positionId": "b0001000-0000-0000-0000-000000000025","varianteId": 25,"produktName": "Bier","varianteName": "Maß 1,0l","kategorie": "beverage","einzelpreis": 850,"menge": 2}],"gesamtPreisCents": 4400,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.produkte-geliefert:v1', 'tisch:1', 28, now() - interval '23 hours 29 minutes', '{"lieferungId": "c0001000-0000-0000-0000-000000000010","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000024","menge": 6},{"positionId": "b0001000-0000-0000-0000-000000000025","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.zahlung-registriert:v1', 'tisch:1', 29, now() - interval '23 hours 10 minutes', '{"zahlungId": "d0001000-0000-0000-0000-000000000009","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000024","menge": 6},{"positionId": "b0001000-0000-0000-0000-000000000025","menge": 2}],"gesamtZahlungCents": 4400,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:1', 30, now() - interval '22 hours 54 minutes', '{"bestellungId": "a0001000-0000-0000-0000-000000000011","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000026","varianteId": 8,"produktName": "Flammkuchen","varianteName": "Mediterran","kategorie": "food","einzelpreis": 750,"menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000027","varianteId": 43,"produktName": "Wein","varianteName": "Rotwein 0,2l","kategorie": "beverage","einzelpreis": 400,"menge": 2}],"gesamtPreisCents": 2300,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.produkte-geliefert:v1', 'tisch:1', 31, now() - interval '22 hours 44 minutes', '{"lieferungId": "c0001000-0000-0000-0000-000000000011","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000026","menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000027","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.zahlung-registriert:v1', 'tisch:1', 32, now() - interval '22 hours 20 minutes', '{"zahlungId": "d0001000-0000-0000-0000-000000000010","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000026","menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000027","menge": 2}],"gesamtZahlungCents": 2300,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:1', 33, now() - interval '21 hours 55 minutes', '{"bestellungId": "a0001000-0000-0000-0000-000000000012","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000028","varianteId": 17,"produktName": "Waffeln","varianteName": "mit Puderzucker","kategorie": "food","einzelpreis": 300,"menge": 3},{"positionId": "b0001000-0000-0000-0000-000000000029","varianteId": 18,"produktName": "Waffeln","varianteName": "mit Sahne","kategorie": "food","einzelpreis": 350,"menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000030","varianteId": 47,"produktName": "Tee","varianteName": "Verschiedene Sorten","kategorie": "beverage","einzelpreis": 200,"menge": 3}],"gesamtPreisCents": 2200,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:1', 34, now() - interval '21 hours 41 minutes', '{"lieferungId": "c0001000-0000-0000-0000-000000000012","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000028","menge": 3},{"positionId": "b0001000-0000-0000-0000-000000000029","menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000030","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:1', 35, now() - interval '21 hours 24 minutes', '{"zahlungId": "d0001000-0000-0000-0000-000000000011","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000028","menge": 3},{"positionId": "b0001000-0000-0000-0000-000000000029","menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000030","menge": 3}],"gesamtZahlungCents": 2200,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:1', 36, now() - interval '21 hours 15 minutes', '{"bestellungId": "a0001000-0000-0000-0000-000000000028","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000075","varianteId": 13,"produktName": "Grillplatte","varianteName": "Groß","kategorie": "food","einzelpreis": 1400,"menge": 1},{"positionId": "b0001000-0000-0000-0000-000000000076","varianteId": 33,"produktName": "Softdrinks","varianteName": "Spezi","kategorie": "beverage","einzelpreis": 280,"menge": 4}],"gesamtPreisCents": 2520,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:1', 37, now() - interval '21 hours 12 minutes', '{"bestellungId": "a0001000-0000-0000-0000-000000000033","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000089","varianteId": 18,"produktName": "Waffeln","varianteName": "mit Sahne","kategorie": "food","einzelpreis": 350,"menge": 3},{"positionId": "b0001000-0000-0000-0000-000000000090","varianteId": 30,"produktName": "Weizen","varianteName": "Russ","kategorie": "beverage","einzelpreis": 300,"menge": 1},{"positionId": "b0001000-0000-0000-0000-000000000091","varianteId": 27,"produktName": "Weizen","varianteName": "Groß 0,5l","kategorie": "beverage","einzelpreis": 400,"menge": 2}],"gesamtPreisCents": 2150,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.bestellung-aufgegeben:v1', 'tisch:1', 38, now() - interval '21 hours 7 minutes', '{"bestellungId": "a0001000-0000-0000-0000-000000000031","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000083","varianteId": 6,"produktName": "Flammkuchen","varianteName": "Classic","kategorie": "food","einzelpreis": 600,"menge": 1},{"positionId": "b0001000-0000-0000-0000-000000000084","varianteId": 1,"produktName": "Bratwurst","varianteName": "Normal","kategorie": "food","einzelpreis": 350,"menge": 1},{"positionId": "b0001000-0000-0000-0000-000000000085","varianteId": 42,"produktName": "Wein","varianteName": "Weißwein 0,2l","kategorie": "beverage","einzelpreis": 400,"menge": 2}],"gesamtPreisCents": 1750,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:1', 39, now() - interval '21 hours 3 minutes', '{"lieferungId": "c0001000-0000-0000-0000-000000000028","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000075","menge": 1},{"positionId": "b0001000-0000-0000-0000-000000000076","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.produkte-geliefert:v1', 'tisch:1', 40, now() - interval '21 hours 3 minutes', '{"lieferungId": "c0001000-0000-0000-0000-000000000033","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000089","menge": 3},{"positionId": "b0001000-0000-0000-0000-000000000090","menge": 1},{"positionId": "b0001000-0000-0000-0000-000000000091","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:1', 41, now() - interval '21 hours', '{"bestellungId": "a0001000-0000-0000-0000-000000000019","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000048","varianteId": 1,"produktName": "Bratwurst","varianteName": "Normal","kategorie": "food","einzelpreis": 350,"menge": 3},{"positionId": "b0001000-0000-0000-0000-000000000049","varianteId": 36,"produktName": "Wasser","varianteName": "Still 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 1}],"gesamtPreisCents": 1250,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.produkte-geliefert:v1', 'tisch:1', 42, now() - interval '20 hours 59 minutes', '{"lieferungId": "c0001000-0000-0000-0000-000000000031","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000083","menge": 1},{"positionId": "b0001000-0000-0000-0000-000000000084","menge": 1},{"positionId": "b0001000-0000-0000-0000-000000000085","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:1', 43, now() - interval '20 hours 50 minutes', '{"lieferungId": "c0001000-0000-0000-0000-000000000019","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000048","menge": 3},{"positionId": "b0001000-0000-0000-0000-000000000049","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:1', 44, now() - interval '20 hours 48 minutes', '{"zahlungId": "d0001000-0000-0000-0000-000000000025","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000075","menge": 1},{"positionId": "b0001000-0000-0000-0000-000000000076","menge": 4}],"gesamtZahlungCents": 2520,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.zahlung-registriert:v1', 'tisch:1', 45, now() - interval '20 hours 47 minutes', '{"zahlungId": "d0001000-0000-0000-0000-000000000028","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000083","menge": 1},{"positionId": "b0001000-0000-0000-0000-000000000084","menge": 1},{"positionId": "b0001000-0000-0000-0000-000000000085","menge": 2}],"gesamtZahlungCents": 1750,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:1', 46, now() - interval '20 hours 45 minutes', '{"bestellungId": "a0001000-0000-0000-0000-000000000022","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000056","varianteId": 17,"produktName": "Waffeln","varianteName": "mit Puderzucker","kategorie": "food","einzelpreis": 300,"menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000057","varianteId": 45,"produktName": "Kaffee","varianteName": "Tasse","kategorie": "beverage","einzelpreis": 200,"menge": 1}],"gesamtPreisCents": 800,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.zahlung-registriert:v1', 'tisch:1', 47, now() - interval '20 hours 41 minutes', '{"zahlungId": "d0001000-0000-0000-0000-000000000030","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000089","menge": 3},{"positionId": "b0001000-0000-0000-0000-000000000090","menge": 1},{"positionId": "b0001000-0000-0000-0000-000000000091","menge": 2}],"gesamtZahlungCents": 2150,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.bestellung-aufgegeben:v1', 'tisch:1', 48, now() - interval '20 hours 39 minutes', '{"bestellungId": "a0001000-0000-0000-0000-000000000018","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000046","varianteId": 13,"produktName": "Grillplatte","varianteName": "Groß","kategorie": "food","einzelpreis": 1400,"menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000047","varianteId": 27,"produktName": "Weizen","varianteName": "Groß 0,5l","kategorie": "beverage","einzelpreis": 400,"menge": 4}],"gesamtPreisCents": 4400,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.bestellung-aufgegeben:v1', 'tisch:1', 49, now() - interval '20 hours 34 minutes', '{"bestellungId": "a0001000-0000-0000-0000-000000000016","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000039","varianteId": 17,"produktName": "Waffeln","varianteName": "mit Puderzucker","kategorie": "food","einzelpreis": 300,"menge": 3},{"positionId": "b0001000-0000-0000-0000-000000000040","varianteId": 17,"produktName": "Waffeln","varianteName": "mit Puderzucker","kategorie": "food","einzelpreis": 300,"menge": 1},{"positionId": "b0001000-0000-0000-0000-000000000041","varianteId": 32,"produktName": "Softdrinks","varianteName": "Fanta","kategorie": "beverage","einzelpreis": 280,"menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000042","varianteId": 36,"produktName": "Wasser","varianteName": "Still 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 4}],"gesamtPreisCents": 2560,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.produkte-geliefert:v1', 'tisch:1', 50, now() - interval '20 hours 32 minutes', '{"lieferungId": "c0001000-0000-0000-0000-000000000022","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000056","menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000057","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.produkte-geliefert:v1', 'tisch:1', 51, now() - interval '20 hours 31 minutes', '{"lieferungId": "c0001000-0000-0000-0000-000000000018","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000046","menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000047","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:1', 52, now() - interval '20 hours 29 minutes', '{"bestellungId": "a0001000-0000-0000-0000-000000000027","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000071","varianteId": 7,"produktName": "Flammkuchen","varianteName": "Speck & Zwiebel","kategorie": "food","einzelpreis": 700,"menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000072","varianteId": 1,"produktName": "Bratwurst","varianteName": "Normal","kategorie": "food","einzelpreis": 350,"menge": 3},{"positionId": "b0001000-0000-0000-0000-000000000073","varianteId": 35,"produktName": "Softdrinks","varianteName": "Mezzo Mix","kategorie": "beverage","einzelpreis": 280,"menge": 1},{"positionId": "b0001000-0000-0000-0000-000000000074","varianteId": 37,"produktName": "Wasser","varianteName": "Medium 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 2}],"gesamtPreisCents": 3130,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:1', 53, now() - interval '20 hours 27 minutes', '{"zahlungId": "d0001000-0000-0000-0000-000000000016","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000048","menge": 3},{"positionId": "b0001000-0000-0000-0000-000000000049","menge": 1}],"gesamtZahlungCents": 1250,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.produkte-geliefert:v1', 'tisch:1', 54, now() - interval '20 hours 26 minutes', '{"lieferungId": "c0001000-0000-0000-0000-000000000016","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000039","menge": 3},{"positionId": "b0001000-0000-0000-0000-000000000040","menge": 1},{"positionId": "b0001000-0000-0000-0000-000000000041","menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000042","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.bestellung-aufgegeben:v1', 'tisch:1', 55, now() - interval '20 hours 26 minutes', '{"bestellungId": "a0001000-0000-0000-0000-000000000021","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000053","varianteId": 5,"produktName": "Pommes","varianteName": "Groß","kategorie": "food","einzelpreis": 350,"menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000054","varianteId": 48,"produktName": "Hugo/Aperol","varianteName": "Hugo","kategorie": "beverage","einzelpreis": 550,"menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000055","varianteId": 48,"produktName": "Hugo/Aperol","varianteName": "Hugo","kategorie": "beverage","einzelpreis": 550,"menge": 3}],"gesamtPreisCents": 3450,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.bestellung-aufgegeben:v1', 'tisch:1', 56, now() - interval '20 hours 24 minutes', '{"bestellungId": "a0001000-0000-0000-0000-000000000029","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000077","varianteId": 5,"produktName": "Pommes","varianteName": "Groß","kategorie": "food","einzelpreis": 350,"menge": 3},{"positionId": "b0001000-0000-0000-0000-000000000078","varianteId": 21,"produktName": "Brezel","varianteName": "mit Butter","kategorie": "food","einzelpreis": 300,"menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000079","varianteId": 35,"produktName": "Softdrinks","varianteName": "Mezzo Mix","kategorie": "beverage","einzelpreis": 280,"menge": 3}],"gesamtPreisCents": 2490,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:1', 57, now() - interval '20 hours 20 minutes', '{"bestellungId": "a0001000-0000-0000-0000-000000000017","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000043","varianteId": 10,"produktName": "Tagesgericht","varianteName": "Sa: Gulasch mit Spätzle","kategorie": "food","einzelpreis": 1150,"menge": 1},{"positionId": "b0001000-0000-0000-0000-000000000044","varianteId": 13,"produktName": "Grillplatte","varianteName": "Groß","kategorie": "food","einzelpreis": 1400,"menge": 1},{"positionId": "b0001000-0000-0000-0000-000000000045","varianteId": 38,"produktName": "Wasser","varianteName": "Sprudel 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 4}],"gesamtPreisCents": 3350,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.produkte-geliefert:v1', 'tisch:1', 58, now() - interval '20 hours 17 minutes', '{"lieferungId": "c0001000-0000-0000-0000-000000000021","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000053","menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000054","menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000055","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:1', 59, now() - interval '20 hours 17 minutes', '{"bestellungId": "a0001000-0000-0000-0000-000000000026","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000068","varianteId": 5,"produktName": "Pommes","varianteName": "Groß","kategorie": "food","einzelpreis": 350,"menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000069","varianteId": 1,"produktName": "Bratwurst","varianteName": "Normal","kategorie": "food","einzelpreis": 350,"menge": 3},{"positionId": "b0001000-0000-0000-0000-000000000070","varianteId": 27,"produktName": "Weizen","varianteName": "Groß 0,5l","kategorie": "beverage","einzelpreis": 400,"menge": 3}],"gesamtPreisCents": 2950,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.produkte-geliefert:v1', 'tisch:1', 60, now() - interval '20 hours 15 minutes', '{"lieferungId": "c0001000-0000-0000-0000-000000000027","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000071","menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000072","menge": 3},{"positionId": "b0001000-0000-0000-0000-000000000073","menge": 1},{"positionId": "b0001000-0000-0000-0000-000000000074","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.zahlung-registriert:v1', 'tisch:1', 61, now() - interval '20 hours 14 minutes', '{"zahlungId": "d0001000-0000-0000-0000-000000000015","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000046","menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000047","menge": 4}],"gesamtZahlungCents": 4400,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.zahlung-registriert:v1', 'tisch:1', 62, now() - interval '20 hours 14 minutes', '{"zahlungId": "d0001000-0000-0000-0000-000000000019","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000056","menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000057","menge": 1}],"gesamtZahlungCents": 800,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:1', 63, now() - interval '20 hours 14 minutes', '{"bestellungId": "a0001000-0000-0000-0000-000000000030","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000080","varianteId": 3,"produktName": "Bratwurst","varianteName": "Currywurst","kategorie": "food","einzelpreis": 450,"menge": 1},{"positionId": "b0001000-0000-0000-0000-000000000081","varianteId": 23,"produktName": "Bier","varianteName": "0,3l","kategorie": "beverage","einzelpreis": 300,"menge": 3},{"positionId": "b0001000-0000-0000-0000-000000000082","varianteId": 39,"produktName": "Saftschorle","varianteName": "Apfelschorle 0,5l","kategorie": "beverage","einzelpreis": 300,"menge": 3}],"gesamtPreisCents": 2250,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.zahlung-registriert:v1', 'tisch:1', 64, now() - interval '20 hours 12 minutes', '{"zahlungId": "d0001000-0000-0000-0000-000000000013","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000039","menge": 3},{"positionId": "b0001000-0000-0000-0000-000000000040","menge": 1},{"positionId": "b0001000-0000-0000-0000-000000000041","menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000042","menge": 4}],"gesamtZahlungCents": 2560,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.bestellung-aufgegeben:v1', 'tisch:1', 65, now() - interval '20 hours 11 minutes', '{"bestellungId": "a0001000-0000-0000-0000-000000000025","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000065","varianteId": 6,"produktName": "Flammkuchen","varianteName": "Classic","kategorie": "food","einzelpreis": 600,"menge": 3},{"positionId": "b0001000-0000-0000-0000-000000000066","varianteId": 48,"produktName": "Hugo/Aperol","varianteName": "Hugo","kategorie": "beverage","einzelpreis": 550,"menge": 4},{"positionId": "b0001000-0000-0000-0000-000000000067","varianteId": 36,"produktName": "Wasser","varianteName": "Still 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 2}],"gesamtPreisCents": 4400,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.produkte-geliefert:v1', 'tisch:1', 66, now() - interval '20 hours 10 minutes', '{"lieferungId": "c0001000-0000-0000-0000-000000000029","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000077","menge": 3},{"positionId": "b0001000-0000-0000-0000-000000000078","menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000079","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.produkte-geliefert:v1', 'tisch:1', 67, now() - interval '20 hours 9 minutes', '{"lieferungId": "c0001000-0000-0000-0000-000000000017","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000043","menge": 1},{"positionId": "b0001000-0000-0000-0000-000000000044","menge": 1},{"positionId": "b0001000-0000-0000-0000-000000000045","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.bestellung-aufgegeben:v1', 'tisch:1', 68, now() - interval '20 hours 7 minutes', '{"bestellungId": "a0001000-0000-0000-0000-000000000034","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000092","varianteId": 13,"produktName": "Grillplatte","varianteName": "Groß","kategorie": "food","einzelpreis": 1400,"menge": 3},{"positionId": "b0001000-0000-0000-0000-000000000093","varianteId": 39,"produktName": "Saftschorle","varianteName": "Apfelschorle 0,5l","kategorie": "beverage","einzelpreis": 300,"menge": 1}],"gesamtPreisCents": 4500,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.produkte-geliefert:v1', 'tisch:1', 69, now() - interval '20 hours 6 minutes', '{"lieferungId": "c0001000-0000-0000-0000-000000000026","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000068","menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000069","menge": 3},{"positionId": "b0001000-0000-0000-0000-000000000070","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:1', 70, now() - interval '20 hours 4 minutes', '{"lieferungId": "c0001000-0000-0000-0000-000000000030","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000080","menge": 1},{"positionId": "b0001000-0000-0000-0000-000000000081","menge": 3},{"positionId": "b0001000-0000-0000-0000-000000000082","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:1', 71, now() - interval '20 hours 3 minutes', '{"bestellungId": "a0001000-0000-0000-0000-000000000023","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000058","varianteId": 12,"produktName": "Grillplatte","varianteName": "Klein","kategorie": "food","einzelpreis": 800,"menge": 1},{"positionId": "b0001000-0000-0000-0000-000000000059","varianteId": 19,"produktName": "Waffeln","varianteName": "mit Nutella","kategorie": "food","einzelpreis": 400,"menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000060","varianteId": 39,"produktName": "Saftschorle","varianteName": "Apfelschorle 0,5l","kategorie": "beverage","einzelpreis": 300,"menge": 3},{"positionId": "b0001000-0000-0000-0000-000000000061","varianteId": 48,"produktName": "Hugo/Aperol","varianteName": "Hugo","kategorie": "beverage","einzelpreis": 550,"menge": 1}],"gesamtPreisCents": 3050,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.produkte-geliefert:v1', 'tisch:1', 72, now() - interval '19 hours 59 minutes', '{"lieferungId": "c0001000-0000-0000-0000-000000000034","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000092","menge": 3},{"positionId": "b0001000-0000-0000-0000-000000000093","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:1', 73, now() - interval '19 hours 58 minutes', '{"bestellungId": "a0001000-0000-0000-0000-000000000024","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000062","varianteId": 19,"produktName": "Waffeln","varianteName": "mit Nutella","kategorie": "food","einzelpreis": 400,"menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000063","varianteId": 10,"produktName": "Tagesgericht","varianteName": "Sa: Gulasch mit Spätzle","kategorie": "food","einzelpreis": 1150,"menge": 3},{"positionId": "b0001000-0000-0000-0000-000000000064","varianteId": 46,"produktName": "Kaffee","varianteName": "Espresso","kategorie": "beverage","einzelpreis": 180,"menge": 3}],"gesamtPreisCents": 4790,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.bestellung-aufgegeben:v1', 'tisch:1', 74, now() - interval '19 hours 58 minutes', '{"bestellungId": "a0001000-0000-0000-0000-000000000032","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000086","varianteId": 14,"produktName": "Salat","varianteName": "Gemischter Salat","kategorie": "food","einzelpreis": 550,"menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000087","varianteId": 30,"produktName": "Weizen","varianteName": "Russ","kategorie": "beverage","einzelpreis": 300,"menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000088","varianteId": 28,"produktName": "Weizen","varianteName": "Colaweizen Klein","kategorie": "beverage","einzelpreis": 300,"menge": 4}],"gesamtPreisCents": 2900,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.zahlung-registriert:v1', 'tisch:1', 75, now() - interval '19 hours 57 minutes', '{"zahlungId": "d0001000-0000-0000-0000-000000000018","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000053","menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000054","menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000055","menge": 3}],"gesamtZahlungCents": 3450,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.produkte-geliefert:v1', 'tisch:1', 76, now() - interval '19 hours 56 minutes', '{"lieferungId": "c0001000-0000-0000-0000-000000000025","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000065","menge": 3},{"positionId": "b0001000-0000-0000-0000-000000000066","menge": 4},{"positionId": "b0001000-0000-0000-0000-000000000067","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.zahlung-registriert:v1', 'tisch:1', 77, now() - interval '19 hours 54 minutes', '{"zahlungId": "d0001000-0000-0000-0000-000000000014","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000043","menge": 1},{"positionId": "b0001000-0000-0000-0000-000000000044","menge": 1},{"positionId": "b0001000-0000-0000-0000-000000000045","menge": 4}],"gesamtZahlungCents": 3350,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.bestellung-aufgegeben:v1', 'tisch:1', 78, now() - interval '19 hours 53 minutes', '{"bestellungId": "a0001000-0000-0000-0000-000000000020","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000050","varianteId": 14,"produktName": "Salat","varianteName": "Gemischter Salat","kategorie": "food","einzelpreis": 550,"menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000051","varianteId": 19,"produktName": "Waffeln","varianteName": "mit Nutella","kategorie": "food","einzelpreis": 400,"menge": 3},{"positionId": "b0001000-0000-0000-0000-000000000052","varianteId": 41,"produktName": "Saftschorle","varianteName": "Rhabarberschorle 0,5l","kategorie": "beverage","einzelpreis": 350,"menge": 1}],"gesamtPreisCents": 2650,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.zahlung-registriert:v1', 'tisch:1', 79, now() - interval '19 hours 53 minutes', '{"zahlungId": "d0001000-0000-0000-0000-000000000024","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000071","menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000072","menge": 3},{"positionId": "b0001000-0000-0000-0000-000000000073","menge": 1},{"positionId": "b0001000-0000-0000-0000-000000000074","menge": 2}],"gesamtZahlungCents": 3130,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.produkte-geliefert:v1', 'tisch:1', 80, now() - interval '19 hours 49 minutes', '{"lieferungId": "c0001000-0000-0000-0000-000000000023","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000058","menge": 1},{"positionId": "b0001000-0000-0000-0000-000000000059","menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000060","menge": 3},{"positionId": "b0001000-0000-0000-0000-000000000061","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:1', 81, now() - interval '19 hours 49 minutes', '{"lieferungId": "c0001000-0000-0000-0000-000000000024","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000062","menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000063","menge": 3},{"positionId": "b0001000-0000-0000-0000-000000000064","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.zahlung-registriert:v1', 'tisch:1', 82, now() - interval '19 hours 46 minutes', '{"zahlungId": "d0001000-0000-0000-0000-000000000023","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000068","menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000069","menge": 3},{"positionId": "b0001000-0000-0000-0000-000000000070","menge": 3}],"gesamtZahlungCents": 2950,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.zahlung-registriert:v1', 'tisch:1', 83, now() - interval '19 hours 45 minutes', '{"zahlungId": "d0001000-0000-0000-0000-000000000026","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000077","menge": 3},{"positionId": "b0001000-0000-0000-0000-000000000078","menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000079","menge": 3}],"gesamtZahlungCents": 2490,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.produkte-geliefert:v1', 'tisch:1', 84, now() - interval '19 hours 43 minutes', '{"lieferungId": "c0001000-0000-0000-0000-000000000032","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000086","menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000087","menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000088","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.produkte-geliefert:v1', 'tisch:1', 85, now() - interval '19 hours 41 minutes', '{"lieferungId": "c0001000-0000-0000-0000-000000000020","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000050","menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000051","menge": 3},{"positionId": "b0001000-0000-0000-0000-000000000052","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:1', 86, now() - interval '19 hours 41 minutes', '{"zahlungId": "d0001000-0000-0000-0000-000000000027","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000080","menge": 1},{"positionId": "b0001000-0000-0000-0000-000000000081","menge": 3},{"positionId": "b0001000-0000-0000-0000-000000000082","menge": 3}],"gesamtZahlungCents": 2250,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.zahlung-registriert:v1', 'tisch:1', 87, now() - interval '19 hours 40 minutes', '{"zahlungId": "d0001000-0000-0000-0000-000000000031","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000092","menge": 3},{"positionId": "b0001000-0000-0000-0000-000000000093","menge": 1}],"gesamtZahlungCents": 4500,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:1', 88, now() - interval '19 hours 35 minutes', '{"zahlungId": "d0001000-0000-0000-0000-000000000021","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000062","menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000063","menge": 3},{"positionId": "b0001000-0000-0000-0000-000000000064","menge": 3}],"gesamtZahlungCents": 4790,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.zahlung-registriert:v1', 'tisch:1', 89, now() - interval '19 hours 34 minutes', '{"zahlungId": "d0001000-0000-0000-0000-000000000020","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000058","menge": 1},{"positionId": "b0001000-0000-0000-0000-000000000059","menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000060","menge": 3},{"positionId": "b0001000-0000-0000-0000-000000000061","menge": 1}],"gesamtZahlungCents": 3050,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.zahlung-registriert:v1', 'tisch:1', 90, now() - interval '19 hours 33 minutes', '{"zahlungId": "d0001000-0000-0000-0000-000000000022","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000065","menge": 3},{"positionId": "b0001000-0000-0000-0000-000000000066","menge": 4},{"positionId": "b0001000-0000-0000-0000-000000000067","menge": 2}],"gesamtZahlungCents": 4400,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.zahlung-registriert:v1', 'tisch:1', 91, now() - interval '19 hours 24 minutes', '{"zahlungId": "d0001000-0000-0000-0000-000000000029","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000086","menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000087","menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000088","menge": 4}],"gesamtZahlungCents": 2900,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.zahlung-registriert:v1', 'tisch:1', 92, now() - interval '19 hours 19 minutes', '{"zahlungId": "d0001000-0000-0000-0000-000000000017","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000050","menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000051","menge": 3},{"positionId": "b0001000-0000-0000-0000-000000000052","menge": 1}],"gesamtZahlungCents": 2650,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:1', 93, now() - interval '4 hours 40 minutes', '{"bestellungId": "a0001000-0000-0000-0000-000000000013","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000031","varianteId": 11,"produktName": "Tagesgericht","varianteName": "So: Hähnchen mit Reis","kategorie": "food","einzelpreis": 1050,"menge": 3},{"positionId": "b0001000-0000-0000-0000-000000000032","varianteId": 14,"produktName": "Salat","varianteName": "Gemischter Salat","kategorie": "food","einzelpreis": 550,"menge": 1},{"positionId": "b0001000-0000-0000-0000-000000000033","varianteId": 36,"produktName": "Wasser","varianteName": "Still 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 3},{"positionId": "b0001000-0000-0000-0000-000000000034","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 2}],"gesamtPreisCents": 5200,"kommentar": "Sonntagsessen"}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:1', 94, now() - interval '4 hours 29 minutes', '{"lieferungId": "c0001000-0000-0000-0000-000000000013","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000031","menge": 3},{"positionId": "b0001000-0000-0000-0000-000000000032","menge": 1},{"positionId": "b0001000-0000-0000-0000-000000000033","menge": 3},{"positionId": "b0001000-0000-0000-0000-000000000034","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:1', 95, now() - interval '4 hours 6 minutes', '{"bestellungId": "a0001000-0000-0000-0000-000000000014","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000035","varianteId": 16,"produktName": "Kuchen","varianteName": "Stück","kategorie": "food","einzelpreis": 250,"menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000036","varianteId": 45,"produktName": "Kaffee","varianteName": "Tasse","kategorie": "beverage","einzelpreis": 200,"menge": 2}],"gesamtPreisCents": 900,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:1', 96, now() - interval '3 hours 52 minutes', '{"lieferungId": "c0001000-0000-0000-0000-000000000014","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000035","menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000036","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.zahlung-registriert:v1', 'tisch:1', 97, now() - interval '3 hours 23 minutes', '{"zahlungId": "d0001000-0000-0000-0000-000000000012","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000031","menge": 3},{"positionId": "b0001000-0000-0000-0000-000000000032","menge": 1},{"positionId": "b0001000-0000-0000-0000-000000000033","menge": 3},{"positionId": "b0001000-0000-0000-0000-000000000034","menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000035","menge": 2},{"positionId": "b0001000-0000-0000-0000-000000000036","menge": 2}],"gesamtZahlungCents": 6100,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.bestellung-aufgegeben:v1', 'tisch:1', 98, now() - interval '3 hours 13 minutes', '{"bestellungId": "a0001000-0000-0000-0000-000000000015","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000037","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 3},{"positionId": "b0001000-0000-0000-0000-000000000038","varianteId": 20,"produktName": "Brezel","varianteName": "Normal","kategorie": "food","einzelpreis": 200,"menge": 2}],"gesamtPreisCents": 1750,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.produkte-geliefert:v1', 'tisch:1', 99, now() - interval '3 hours', '{"lieferungId": "c0001000-0000-0000-0000-000000000015","positionen": [{"positionId": "b0001000-0000-0000-0000-000000000037","menge": 3},{"positionId": "b0001000-0000-0000-0000-000000000038","menge": 2}],"kommentar": ""}');

-- ─────────────────────────────────────────────────────────────────────────────
-- TISCH 2: Junge Leute — Tag 2+3, viele Getränke (73 Events)
-- ─────────────────────────────────────────────────────────────────────────────

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.bestellung-aufgegeben:v1', 'tisch:2', 1, now() - interval '23 hours 20 minutes', '{"bestellungId": "a0002000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000001","varianteId": 1,"produktName": "Bratwurst","varianteName": "Normal","kategorie": "food","einzelpreis": 350,"menge": 3},{"positionId": "b0002000-0000-0000-0000-000000000002","varianteId": 5,"produktName": "Pommes","varianteName": "Groß","kategorie": "food","einzelpreis": 350,"menge": 3},{"positionId": "b0002000-0000-0000-0000-000000000003","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 6},{"positionId": "b0002000-0000-0000-0000-000000000004","varianteId": 25,"produktName": "Bier","varianteName": "Maß 1,0l","kategorie": "beverage","einzelpreis": 850,"menge": 2}],"gesamtPreisCents": 6500,"kommentar": "Großer Hunger"}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.produkte-storniert:v1', 'tisch:2', 2, now() - interval '23 hours 10 minutes', '{"stornierungId": "e0002000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000001","menge": 1}],"gesamtStornierungCents": 350,"kommentar": "Eine Bratwurst zu viel"}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.produkte-geliefert:v1', 'tisch:2', 3, now() - interval '23 hours 5 minutes', '{"lieferungId": "c0002000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000001","menge": 2},{"positionId": "b0002000-0000-0000-0000-000000000002","menge": 3},{"positionId": "b0002000-0000-0000-0000-000000000003","menge": 6},{"positionId": "b0002000-0000-0000-0000-000000000004","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.bestellung-aufgegeben:v1', 'tisch:2', 4, now() - interval '22 hours 50 minutes', '{"bestellungId": "a0002000-0000-0000-0000-000000000002","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000005","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 4},{"positionId": "b0002000-0000-0000-0000-000000000006","varianteId": 28,"produktName": "Weizen","varianteName": "Colaweizen Klein","kategorie": "beverage","einzelpreis": 300,"menge": 2}],"gesamtPreisCents": 2400,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.produkte-geliefert:v1', 'tisch:2', 5, now() - interval '22 hours 36 minutes', '{"lieferungId": "c0002000-0000-0000-0000-000000000002","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000005","menge": 4},{"positionId": "b0002000-0000-0000-0000-000000000006","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:2', 6, now() - interval '22 hours 15 minutes', '{"bestellungId": "a0002000-0000-0000-0000-000000000003","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000007","varianteId": 3,"produktName": "Bratwurst","varianteName": "Currywurst","kategorie": "food","einzelpreis": 450,"menge": 2},{"positionId": "b0002000-0000-0000-0000-000000000008","varianteId": 5,"produktName": "Pommes","varianteName": "Groß","kategorie": "food","einzelpreis": 350,"menge": 2},{"positionId": "b0002000-0000-0000-0000-000000000009","varianteId": 31,"produktName": "Softdrinks","varianteName": "Cola","kategorie": "beverage","einzelpreis": 280,"menge": 3}],"gesamtPreisCents": 2440,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.produkte-geliefert:v1', 'tisch:2', 7, now() - interval '22 hours 4 minutes', '{"lieferungId": "c0002000-0000-0000-0000-000000000003","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000007","menge": 2},{"positionId": "b0002000-0000-0000-0000-000000000008","menge": 2},{"positionId": "b0002000-0000-0000-0000-000000000009","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.zahlung-registriert:v1', 'tisch:2', 8, now() - interval '21 hours 36 minutes', '{"zahlungId": "d0002000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000001","menge": 2},{"positionId": "b0002000-0000-0000-0000-000000000002","menge": 3},{"positionId": "b0002000-0000-0000-0000-000000000003","menge": 6},{"positionId": "b0002000-0000-0000-0000-000000000004","menge": 2},{"positionId": "b0002000-0000-0000-0000-000000000005","menge": 4},{"positionId": "b0002000-0000-0000-0000-000000000006","menge": 2}],"gesamtZahlungCents": 8550,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:2', 9, now() - interval '21 hours 26 minutes', '{"bestellungId": "a0002000-0000-0000-0000-000000000004","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000010","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 4},{"positionId": "b0002000-0000-0000-0000-000000000011","varianteId": 30,"produktName": "Weizen","varianteName": "Russ","kategorie": "beverage","einzelpreis": 300,"menge": 2}],"gesamtPreisCents": 2400,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.produkte-geliefert:v1', 'tisch:2', 10, now() - interval '21 hours 15 minutes', '{"lieferungId": "c0002000-0000-0000-0000-000000000004","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000010","menge": 4},{"positionId": "b0002000-0000-0000-0000-000000000011","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.zahlung-registriert:v1', 'tisch:2', 11, now() - interval '20 hours 51 minutes', '{"zahlungId": "d0002000-0000-0000-0000-000000000002","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000010","menge": 4},{"positionId": "b0002000-0000-0000-0000-000000000011","menge": 2}],"gesamtZahlungCents": 2400,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.bestellung-aufgegeben:v1', 'tisch:2', 12, now() - interval '20 hours 36 minutes', '{"bestellungId": "a0002000-0000-0000-0000-000000000005","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000012","varianteId": 12,"produktName": "Grillplatte","varianteName": "Klein","kategorie": "food","einzelpreis": 800,"menge": 1},{"positionId": "b0002000-0000-0000-0000-000000000013","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 2}],"gesamtPreisCents": 1700,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.produkte-geliefert:v1', 'tisch:2', 13, now() - interval '20 hours 24 minutes', '{"lieferungId": "c0002000-0000-0000-0000-000000000005","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000012","menge": 1},{"positionId": "b0002000-0000-0000-0000-000000000013","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.zahlung-registriert:v1', 'tisch:2', 14, now() - interval '19 hours 59 minutes', '{"zahlungId": "d0002000-0000-0000-0000-000000000003","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000012","menge": 1},{"positionId": "b0002000-0000-0000-0000-000000000013","menge": 2}],"gesamtZahlungCents": 1700,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.zahlung-registriert:v1', 'tisch:2', 15, now() - interval '19 hours 34 minutes', '{"zahlungId": "d0002000-0000-0000-0000-000000000004","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000007","menge": 2},{"positionId": "b0002000-0000-0000-0000-000000000008","menge": 2},{"positionId": "b0002000-0000-0000-0000-000000000009","menge": 3}],"gesamtZahlungCents": 2440,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.bestellung-aufgegeben:v1', 'tisch:2', 16, now() - interval '19 hours 24 minutes', '{"bestellungId": "a0002000-0000-0000-0000-000000000006","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000014","varianteId": 48,"produktName": "Hugo/Aperol","varianteName": "Hugo","kategorie": "beverage","einzelpreis": 550,"menge": 3},{"positionId": "b0002000-0000-0000-0000-000000000015","varianteId": 49,"produktName": "Hugo/Aperol","varianteName": "Aperol Spritz","kategorie": "beverage","einzelpreis": 550,"menge": 3}],"gesamtPreisCents": 3300,"kommentar": "Digestif"}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.produkte-geliefert:v1', 'tisch:2', 17, now() - interval '19 hours 10 minutes', '{"lieferungId": "c0002000-0000-0000-0000-000000000006","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000014","menge": 3},{"positionId": "b0002000-0000-0000-0000-000000000015","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.zahlung-registriert:v1', 'tisch:2', 18, now() - interval '18 hours 49 minutes', '{"zahlungId": "d0002000-0000-0000-0000-000000000005","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000014","menge": 3},{"positionId": "b0002000-0000-0000-0000-000000000015","menge": 3}],"gesamtZahlungCents": 3300,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:2', 19, now() - interval '18 hours 18 minutes', '{"bestellungId": "a0002000-0000-0000-0000-000000000019","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000049","varianteId": 19,"produktName": "Waffeln","varianteName": "mit Nutella","kategorie": "food","einzelpreis": 400,"menge": 1},{"positionId": "b0002000-0000-0000-0000-000000000050","varianteId": 34,"produktName": "Softdrinks","varianteName": "Sprite","kategorie": "beverage","einzelpreis": 280,"menge": 2}],"gesamtPreisCents": 960,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.bestellung-aufgegeben:v1', 'tisch:2', 20, now() - interval '18 hours 15 minutes', '{"bestellungId": "a0002000-0000-0000-0000-000000000023","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000059","varianteId": 13,"produktName": "Grillplatte","varianteName": "Groß","kategorie": "food","einzelpreis": 1400,"menge": 1},{"positionId": "b0002000-0000-0000-0000-000000000060","varianteId": 20,"produktName": "Brezel","varianteName": "Normal","kategorie": "food","einzelpreis": 200,"menge": 2},{"positionId": "b0002000-0000-0000-0000-000000000061","varianteId": 27,"produktName": "Weizen","varianteName": "Groß 0,5l","kategorie": "beverage","einzelpreis": 400,"menge": 4}],"gesamtPreisCents": 3400,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.bestellung-aufgegeben:v1', 'tisch:2', 21, now() - interval '18 hours 8 minutes', '{"bestellungId": "a0002000-0000-0000-0000-000000000011","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000025","varianteId": 1,"produktName": "Bratwurst","varianteName": "Normal","kategorie": "food","einzelpreis": 350,"menge": 1},{"positionId": "b0002000-0000-0000-0000-000000000026","varianteId": 5,"produktName": "Pommes","varianteName": "Groß","kategorie": "food","einzelpreis": 350,"menge": 2},{"positionId": "b0002000-0000-0000-0000-000000000027","varianteId": 26,"produktName": "Weizen","varianteName": "Klein 0,3l","kategorie": "beverage","einzelpreis": 300,"menge": 3}],"gesamtPreisCents": 1950,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:2', 22, now() - interval '18 hours 8 minutes', '{"bestellungId": "a0002000-0000-0000-0000-000000000020","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000051","varianteId": 16,"produktName": "Kuchen","varianteName": "Stück","kategorie": "food","einzelpreis": 250,"menge": 3},{"positionId": "b0002000-0000-0000-0000-000000000052","varianteId": 40,"produktName": "Saftschorle","varianteName": "Johannisbeerschorle 0,5l","kategorie": "beverage","einzelpreis": 300,"menge": 4}],"gesamtPreisCents": 1950,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:2', 23, now() - interval '18 hours 5 minutes', '{"lieferungId": "c0002000-0000-0000-0000-000000000019","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000049","menge": 1},{"positionId": "b0002000-0000-0000-0000-000000000050","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.produkte-geliefert:v1', 'tisch:2', 24, now() - interval '18 hours 5 minutes', '{"lieferungId": "c0002000-0000-0000-0000-000000000023","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000059","menge": 1},{"positionId": "b0002000-0000-0000-0000-000000000060","menge": 2},{"positionId": "b0002000-0000-0000-0000-000000000061","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.bestellung-aufgegeben:v1', 'tisch:2', 25, now() - interval '18 hours 2 minutes', '{"bestellungId": "a0002000-0000-0000-0000-000000000017","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000043","varianteId": 21,"produktName": "Brezel","varianteName": "mit Butter","kategorie": "food","einzelpreis": 300,"menge": 2},{"positionId": "b0002000-0000-0000-0000-000000000044","varianteId": 4,"produktName": "Pommes","varianteName": "Klein","kategorie": "food","einzelpreis": 250,"menge": 2},{"positionId": "b0002000-0000-0000-0000-000000000045","varianteId": 30,"produktName": "Weizen","varianteName": "Russ","kategorie": "beverage","einzelpreis": 300,"menge": 2}],"gesamtPreisCents": 1700,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.bestellung-aufgegeben:v1', 'tisch:2', 26, now() - interval '18 hours 2 minutes', '{"bestellungId": "a0002000-0000-0000-0000-000000000022","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000056","varianteId": 14,"produktName": "Salat","varianteName": "Gemischter Salat","kategorie": "food","einzelpreis": 550,"menge": 1},{"positionId": "b0002000-0000-0000-0000-000000000057","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 4},{"positionId": "b0002000-0000-0000-0000-000000000058","varianteId": 49,"produktName": "Hugo/Aperol","varianteName": "Aperol Spritz","kategorie": "beverage","einzelpreis": 550,"menge": 4}],"gesamtPreisCents": 4550,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.produkte-geliefert:v1', 'tisch:2', 27, now() - interval '17 hours 56 minutes', '{"lieferungId": "c0002000-0000-0000-0000-000000000011","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000025","menge": 1},{"positionId": "b0002000-0000-0000-0000-000000000026","menge": 2},{"positionId": "b0002000-0000-0000-0000-000000000027","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.produkte-geliefert:v1', 'tisch:2', 28, now() - interval '17 hours 56 minutes', '{"lieferungId": "c0002000-0000-0000-0000-000000000020","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000051","menge": 3},{"positionId": "b0002000-0000-0000-0000-000000000052","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.produkte-geliefert:v1', 'tisch:2', 29, now() - interval '17 hours 54 minutes', '{"lieferungId": "c0002000-0000-0000-0000-000000000017","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000043","menge": 2},{"positionId": "b0002000-0000-0000-0000-000000000044","menge": 2},{"positionId": "b0002000-0000-0000-0000-000000000045","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.zahlung-registriert:v1', 'tisch:2', 30, now() - interval '17 hours 51 minutes', '{"zahlungId": "d0002000-0000-0000-0000-000000000020","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000059","menge": 1},{"positionId": "b0002000-0000-0000-0000-000000000060","menge": 2},{"positionId": "b0002000-0000-0000-0000-000000000061","menge": 4}],"gesamtZahlungCents": 3400,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.produkte-geliefert:v1', 'tisch:2', 31, now() - interval '17 hours 47 minutes', '{"lieferungId": "c0002000-0000-0000-0000-000000000022","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000056","menge": 1},{"positionId": "b0002000-0000-0000-0000-000000000057","menge": 4},{"positionId": "b0002000-0000-0000-0000-000000000058","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.bestellung-aufgegeben:v1', 'tisch:2', 32, now() - interval '17 hours 43 minutes', '{"bestellungId": "a0002000-0000-0000-0000-000000000016","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000039","varianteId": 14,"produktName": "Salat","varianteName": "Gemischter Salat","kategorie": "food","einzelpreis": 550,"menge": 2},{"positionId": "b0002000-0000-0000-0000-000000000040","varianteId": 2,"produktName": "Bratwurst","varianteName": "XXL","kategorie": "food","einzelpreis": 500,"menge": 3},{"positionId": "b0002000-0000-0000-0000-000000000041","varianteId": 49,"produktName": "Hugo/Aperol","varianteName": "Aperol Spritz","kategorie": "beverage","einzelpreis": 550,"menge": 4},{"positionId": "b0002000-0000-0000-0000-000000000042","varianteId": 26,"produktName": "Weizen","varianteName": "Klein 0,3l","kategorie": "beverage","einzelpreis": 300,"menge": 2}],"gesamtPreisCents": 5400,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:2', 33, now() - interval '17 hours 40 minutes', '{"zahlungId": "d0002000-0000-0000-0000-000000000016","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000049","menge": 1},{"positionId": "b0002000-0000-0000-0000-000000000050","menge": 2}],"gesamtZahlungCents": 960,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.zahlung-registriert:v1', 'tisch:2', 34, now() - interval '17 hours 40 minutes', '{"zahlungId": "d0002000-0000-0000-0000-000000000017","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000051","menge": 3},{"positionId": "b0002000-0000-0000-0000-000000000052","menge": 4}],"gesamtZahlungCents": 1950,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.zahlung-registriert:v1', 'tisch:2', 35, now() - interval '17 hours 38 minutes', '{"zahlungId": "d0002000-0000-0000-0000-000000000014","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000043","menge": 2},{"positionId": "b0002000-0000-0000-0000-000000000044","menge": 2},{"positionId": "b0002000-0000-0000-0000-000000000045","menge": 2}],"gesamtZahlungCents": 1700,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.zahlung-registriert:v1', 'tisch:2', 36, now() - interval '17 hours 34 minutes', '{"zahlungId": "d0002000-0000-0000-0000-000000000008","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000025","menge": 1},{"positionId": "b0002000-0000-0000-0000-000000000026","menge": 2},{"positionId": "b0002000-0000-0000-0000-000000000027","menge": 3}],"gesamtZahlungCents": 1950,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.produkte-geliefert:v1', 'tisch:2', 37, now() - interval '17 hours 33 minutes', '{"lieferungId": "c0002000-0000-0000-0000-000000000016","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000039","menge": 2},{"positionId": "b0002000-0000-0000-0000-000000000040","menge": 3},{"positionId": "b0002000-0000-0000-0000-000000000041","menge": 4},{"positionId": "b0002000-0000-0000-0000-000000000042","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.zahlung-registriert:v1', 'tisch:2', 38, now() - interval '17 hours 28 minutes', '{"zahlungId": "d0002000-0000-0000-0000-000000000019","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000056","menge": 1},{"positionId": "b0002000-0000-0000-0000-000000000057","menge": 4},{"positionId": "b0002000-0000-0000-0000-000000000058","menge": 4}],"gesamtZahlungCents": 4550,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.zahlung-registriert:v1', 'tisch:2', 39, now() - interval '17 hours 17 minutes', '{"zahlungId": "d0002000-0000-0000-0000-000000000013","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000039","menge": 2},{"positionId": "b0002000-0000-0000-0000-000000000040","menge": 3},{"positionId": "b0002000-0000-0000-0000-000000000041","menge": 4},{"positionId": "b0002000-0000-0000-0000-000000000042","menge": 2}],"gesamtZahlungCents": 5400,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.bestellung-aufgegeben:v1', 'tisch:2', 40, now() - interval '17 hours 13 minutes', '{"bestellungId": "a0002000-0000-0000-0000-000000000025","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000064","varianteId": 14,"produktName": "Salat","varianteName": "Gemischter Salat","kategorie": "food","einzelpreis": 550,"menge": 1},{"positionId": "b0002000-0000-0000-0000-000000000065","varianteId": 40,"produktName": "Saftschorle","varianteName": "Johannisbeerschorle 0,5l","kategorie": "beverage","einzelpreis": 300,"menge": 4}],"gesamtPreisCents": 1750,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.bestellung-aufgegeben:v1', 'tisch:2', 41, now() - interval '17 hours 9 minutes', '{"bestellungId": "a0002000-0000-0000-0000-000000000024","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000062","varianteId": 21,"produktName": "Brezel","varianteName": "mit Butter","kategorie": "food","einzelpreis": 300,"menge": 3},{"positionId": "b0002000-0000-0000-0000-000000000063","varianteId": 33,"produktName": "Softdrinks","varianteName": "Spezi","kategorie": "beverage","einzelpreis": 280,"menge": 3}],"gesamtPreisCents": 1740,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:2', 42, now() - interval '17 hours 6 minutes', '{"bestellungId": "a0002000-0000-0000-0000-000000000015","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000037","varianteId": 21,"produktName": "Brezel","varianteName": "mit Butter","kategorie": "food","einzelpreis": 300,"menge": 2},{"positionId": "b0002000-0000-0000-0000-000000000038","varianteId": 41,"produktName": "Saftschorle","varianteName": "Rhabarberschorle 0,5l","kategorie": "beverage","einzelpreis": 350,"menge": 2}],"gesamtPreisCents": 1300,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:2', 43, now() - interval '17 hours 5 minutes', '{"bestellungId": "a0002000-0000-0000-0000-000000000010","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000023","varianteId": 18,"produktName": "Waffeln","varianteName": "mit Sahne","kategorie": "food","einzelpreis": 350,"menge": 3},{"positionId": "b0002000-0000-0000-0000-000000000024","varianteId": 34,"produktName": "Softdrinks","varianteName": "Sprite","kategorie": "beverage","einzelpreis": 280,"menge": 3}],"gesamtPreisCents": 1890,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.bestellung-aufgegeben:v1', 'tisch:2', 44, now() - interval '17 hours 2 minutes', '{"bestellungId": "a0002000-0000-0000-0000-000000000014","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000034","varianteId": 1,"produktName": "Bratwurst","varianteName": "Normal","kategorie": "food","einzelpreis": 350,"menge": 3},{"positionId": "b0002000-0000-0000-0000-000000000035","varianteId": 35,"produktName": "Softdrinks","varianteName": "Mezzo Mix","kategorie": "beverage","einzelpreis": 280,"menge": 1},{"positionId": "b0002000-0000-0000-0000-000000000036","varianteId": 44,"produktName": "Wein","varianteName": "Rosé 0,2l","kategorie": "beverage","einzelpreis": 400,"menge": 4}],"gesamtPreisCents": 2930,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.produkte-geliefert:v1', 'tisch:2', 45, now() - interval '17 hours', '{"lieferungId": "c0002000-0000-0000-0000-000000000024","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000062","menge": 3},{"positionId": "b0002000-0000-0000-0000-000000000063","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.produkte-geliefert:v1', 'tisch:2', 46, now() - interval '16 hours 58 minutes', '{"lieferungId": "c0002000-0000-0000-0000-000000000015","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000037","menge": 2},{"positionId": "b0002000-0000-0000-0000-000000000038","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.produkte-geliefert:v1', 'tisch:2', 47, now() - interval '16 hours 58 minutes', '{"lieferungId": "c0002000-0000-0000-0000-000000000025","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000064","menge": 1},{"positionId": "b0002000-0000-0000-0000-000000000065","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.bestellung-aufgegeben:v1', 'tisch:2', 48, now() - interval '16 hours 57 minutes', '{"bestellungId": "a0002000-0000-0000-0000-000000000013","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000031","varianteId": 18,"produktName": "Waffeln","varianteName": "mit Sahne","kategorie": "food","einzelpreis": 350,"menge": 1},{"positionId": "b0002000-0000-0000-0000-000000000032","varianteId": 26,"produktName": "Weizen","varianteName": "Klein 0,3l","kategorie": "beverage","einzelpreis": 300,"menge": 2},{"positionId": "b0002000-0000-0000-0000-000000000033","varianteId": 40,"produktName": "Saftschorle","varianteName": "Johannisbeerschorle 0,5l","kategorie": "beverage","einzelpreis": 300,"menge": 2}],"gesamtPreisCents": 1550,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.produkte-geliefert:v1', 'tisch:2', 49, now() - interval '16 hours 55 minutes', '{"lieferungId": "c0002000-0000-0000-0000-000000000010","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000023","menge": 3},{"positionId": "b0002000-0000-0000-0000-000000000024","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.produkte-geliefert:v1', 'tisch:2', 50, now() - interval '16 hours 50 minutes', '{"lieferungId": "c0002000-0000-0000-0000-000000000014","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000034","menge": 3},{"positionId": "b0002000-0000-0000-0000-000000000035","menge": 1},{"positionId": "b0002000-0000-0000-0000-000000000036","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.bestellung-aufgegeben:v1', 'tisch:2', 51, now() - interval '16 hours 49 minutes', '{"bestellungId": "a0002000-0000-0000-0000-000000000018","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000046","varianteId": 8,"produktName": "Flammkuchen","varianteName": "Mediterran","kategorie": "food","einzelpreis": 750,"menge": 1},{"positionId": "b0002000-0000-0000-0000-000000000047","varianteId": 29,"produktName": "Weizen","varianteName": "Colaweizen Groß","kategorie": "beverage","einzelpreis": 400,"menge": 3},{"positionId": "b0002000-0000-0000-0000-000000000048","varianteId": 27,"produktName": "Weizen","varianteName": "Groß 0,5l","kategorie": "beverage","einzelpreis": 400,"menge": 1}],"gesamtPreisCents": 2350,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.produkte-geliefert:v1', 'tisch:2', 52, now() - interval '16 hours 47 minutes', '{"lieferungId": "c0002000-0000-0000-0000-000000000013","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000031","menge": 1},{"positionId": "b0002000-0000-0000-0000-000000000032","menge": 2},{"positionId": "b0002000-0000-0000-0000-000000000033","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.bestellung-aufgegeben:v1', 'tisch:2', 53, now() - interval '16 hours 46 minutes', '{"bestellungId": "a0002000-0000-0000-0000-000000000012","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000028","varianteId": 14,"produktName": "Salat","varianteName": "Gemischter Salat","kategorie": "food","einzelpreis": 550,"menge": 3},{"positionId": "b0002000-0000-0000-0000-000000000029","varianteId": 15,"produktName": "Salat","varianteName": "Caesar Salat","kategorie": "food","einzelpreis": 650,"menge": 3},{"positionId": "b0002000-0000-0000-0000-000000000030","varianteId": 23,"produktName": "Bier","varianteName": "0,3l","kategorie": "beverage","einzelpreis": 300,"menge": 1}],"gesamtPreisCents": 3900,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.bestellung-aufgegeben:v1', 'tisch:2', 54, now() - interval '16 hours 46 minutes', '{"bestellungId": "a0002000-0000-0000-0000-000000000021","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000053","varianteId": 13,"produktName": "Grillplatte","varianteName": "Groß","kategorie": "food","einzelpreis": 1400,"menge": 2},{"positionId": "b0002000-0000-0000-0000-000000000054","varianteId": 3,"produktName": "Bratwurst","varianteName": "Currywurst","kategorie": "food","einzelpreis": 450,"menge": 3},{"positionId": "b0002000-0000-0000-0000-000000000055","varianteId": 39,"produktName": "Saftschorle","varianteName": "Apfelschorle 0,5l","kategorie": "beverage","einzelpreis": 300,"menge": 3}],"gesamtPreisCents": 5050,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.zahlung-registriert:v1', 'tisch:2', 55, now() - interval '16 hours 41 minutes', '{"zahlungId": "d0002000-0000-0000-0000-000000000022","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000064","menge": 1},{"positionId": "b0002000-0000-0000-0000-000000000065","menge": 4}],"gesamtZahlungCents": 1750,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.zahlung-registriert:v1', 'tisch:2', 56, now() - interval '16 hours 39 minutes', '{"zahlungId": "d0002000-0000-0000-0000-000000000021","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000062","menge": 3},{"positionId": "b0002000-0000-0000-0000-000000000063","menge": 3}],"gesamtZahlungCents": 1740,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.produkte-geliefert:v1', 'tisch:2', 57, now() - interval '16 hours 35 minutes', '{"lieferungId": "c0002000-0000-0000-0000-000000000021","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000053","menge": 2},{"positionId": "b0002000-0000-0000-0000-000000000054","menge": 3},{"positionId": "b0002000-0000-0000-0000-000000000055","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.zahlung-registriert:v1', 'tisch:2', 58, now() - interval '16 hours 34 minutes', '{"zahlungId": "d0002000-0000-0000-0000-000000000012","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000037","menge": 2},{"positionId": "b0002000-0000-0000-0000-000000000038","menge": 2}],"gesamtZahlungCents": 1300,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.produkte-geliefert:v1', 'tisch:2', 59, now() - interval '16 hours 34 minutes', '{"lieferungId": "c0002000-0000-0000-0000-000000000018","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000046","menge": 1},{"positionId": "b0002000-0000-0000-0000-000000000047","menge": 3},{"positionId": "b0002000-0000-0000-0000-000000000048","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.zahlung-registriert:v1', 'tisch:2', 60, now() - interval '16 hours 33 minutes', '{"zahlungId": "d0002000-0000-0000-0000-000000000007","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000023","menge": 3},{"positionId": "b0002000-0000-0000-0000-000000000024","menge": 3}],"gesamtZahlungCents": 1890,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.produkte-geliefert:v1', 'tisch:2', 61, now() - interval '16 hours 32 minutes', '{"lieferungId": "c0002000-0000-0000-0000-000000000012","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000028","menge": 3},{"positionId": "b0002000-0000-0000-0000-000000000029","menge": 3},{"positionId": "b0002000-0000-0000-0000-000000000030","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.zahlung-registriert:v1', 'tisch:2', 62, now() - interval '16 hours 27 minutes', '{"zahlungId": "d0002000-0000-0000-0000-000000000011","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000034","menge": 3},{"positionId": "b0002000-0000-0000-0000-000000000035","menge": 1},{"positionId": "b0002000-0000-0000-0000-000000000036","menge": 4}],"gesamtZahlungCents": 2930,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.zahlung-registriert:v1', 'tisch:2', 63, now() - interval '16 hours 26 minutes', '{"zahlungId": "d0002000-0000-0000-0000-000000000010","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000031","menge": 1},{"positionId": "b0002000-0000-0000-0000-000000000032","menge": 2},{"positionId": "b0002000-0000-0000-0000-000000000033","menge": 2}],"gesamtZahlungCents": 1550,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.zahlung-registriert:v1', 'tisch:2', 64, now() - interval '16 hours 16 minutes', '{"zahlungId": "d0002000-0000-0000-0000-000000000015","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000046","menge": 1},{"positionId": "b0002000-0000-0000-0000-000000000047","menge": 3},{"positionId": "b0002000-0000-0000-0000-000000000048","menge": 1}],"gesamtZahlungCents": 2350,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.zahlung-registriert:v1', 'tisch:2', 65, now() - interval '16 hours 15 minutes', '{"zahlungId": "d0002000-0000-0000-0000-000000000018","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000053","menge": 2},{"positionId": "b0002000-0000-0000-0000-000000000054","menge": 3},{"positionId": "b0002000-0000-0000-0000-000000000055","menge": 3}],"gesamtZahlungCents": 5050,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.zahlung-registriert:v1', 'tisch:2', 66, now() - interval '16 hours 7 minutes', '{"zahlungId": "d0002000-0000-0000-0000-000000000009","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000028","menge": 3},{"positionId": "b0002000-0000-0000-0000-000000000029","menge": 3},{"positionId": "b0002000-0000-0000-0000-000000000030","menge": 1}],"gesamtZahlungCents": 3900,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:2', 67, now() - interval '4 hours 10 minutes', '{"bestellungId": "a0002000-0000-0000-0000-000000000007","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000016","varianteId": 11,"produktName": "Tagesgericht","varianteName": "So: Hähnchen mit Reis","kategorie": "food","einzelpreis": 1050,"menge": 2},{"positionId": "b0002000-0000-0000-0000-000000000017","varianteId": 14,"produktName": "Salat","varianteName": "Gemischter Salat","kategorie": "food","einzelpreis": 550,"menge": 2},{"positionId": "b0002000-0000-0000-0000-000000000018","varianteId": 37,"produktName": "Wasser","varianteName": "Medium 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 4}],"gesamtPreisCents": 4000,"kommentar": "Katerfrühstück"}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.produkte-geliefert:v1', 'tisch:2', 68, now() - interval '3 hours 57 minutes', '{"lieferungId": "c0002000-0000-0000-0000-000000000007","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000016","menge": 2},{"positionId": "b0002000-0000-0000-0000-000000000017","menge": 2},{"positionId": "b0002000-0000-0000-0000-000000000018","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:2', 69, now() - interval '3 hours 36 minutes', '{"bestellungId": "a0002000-0000-0000-0000-000000000008","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000019","varianteId": 45,"produktName": "Kaffee","varianteName": "Tasse","kategorie": "beverage","einzelpreis": 200,"menge": 4},{"positionId": "b0002000-0000-0000-0000-000000000020","varianteId": 16,"produktName": "Kuchen","varianteName": "Stück","kategorie": "food","einzelpreis": 250,"menge": 2}],"gesamtPreisCents": 1300,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.produkte-geliefert:v1', 'tisch:2', 70, now() - interval '3 hours 25 minutes', '{"lieferungId": "c0002000-0000-0000-0000-000000000008","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000019","menge": 4},{"positionId": "b0002000-0000-0000-0000-000000000020","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.zahlung-registriert:v1', 'tisch:2', 71, now() - interval '3 hours 10 minutes', '{"zahlungId": "d0002000-0000-0000-0000-000000000006","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000016","menge": 2},{"positionId": "b0002000-0000-0000-0000-000000000017","menge": 2},{"positionId": "b0002000-0000-0000-0000-000000000018","menge": 4},{"positionId": "b0002000-0000-0000-0000-000000000019","menge": 4},{"positionId": "b0002000-0000-0000-0000-000000000020","menge": 2}],"gesamtZahlungCents": 5300,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.bestellung-aufgegeben:v1', 'tisch:2', 72, now() - interval '3 hours', '{"bestellungId": "a0002000-0000-0000-0000-000000000009","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000021","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 3},{"positionId": "b0002000-0000-0000-0000-000000000022","varianteId": 39,"produktName": "Saftschorle","varianteName": "Apfelschorle 0,5l","kategorie": "beverage","einzelpreis": 300,"menge": 2}],"gesamtPreisCents": 1950,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.produkte-geliefert:v1', 'tisch:2', 73, now() - interval '2 hours 47 minutes', '{"lieferungId": "c0002000-0000-0000-0000-000000000009","positionen": [{"positionId": "b0002000-0000-0000-0000-000000000021","menge": 3},{"positionId": "b0002000-0000-0000-0000-000000000022","menge": 2}],"kommentar": ""}');

-- ─────────────────────────────────────────────────────────────────────────────
-- TISCH 3: Älteres Ehepaar — Tag 1+2, ruhig (18 Events)
-- ─────────────────────────────────────────────────────────────────────────────

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:3', 1, now() - interval '49 hours 10 minutes', '{"bestellungId": "a0003000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0003000-0000-0000-0000-000000000001","varianteId": 9,"produktName": "Tagesgericht","varianteName": "Fr: Schnitzel mit Pommes","kategorie": "food","einzelpreis": 1250,"menge": 2},{"positionId": "b0003000-0000-0000-0000-000000000002","varianteId": 42,"produktName": "Wein","varianteName": "Weißwein 0,2l","kategorie": "beverage","einzelpreis": 400,"menge": 2},{"positionId": "b0003000-0000-0000-0000-000000000003","varianteId": 36,"produktName": "Wasser","varianteName": "Still 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 1}],"gesamtPreisCents": 3500,"kommentar": "Abendessen zu zweit"}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:3', 2, now() - interval '48 hours 59 minutes', '{"lieferungId": "c0003000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0003000-0000-0000-0000-000000000001","menge": 2},{"positionId": "b0003000-0000-0000-0000-000000000002","menge": 2},{"positionId": "b0003000-0000-0000-0000-000000000003","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:3', 3, now() - interval '48 hours 42 minutes', '{"zahlungId": "d0003000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0003000-0000-0000-0000-000000000001","menge": 2},{"positionId": "b0003000-0000-0000-0000-000000000002","menge": 2},{"positionId": "b0003000-0000-0000-0000-000000000003","menge": 1}],"gesamtZahlungCents": 3500,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:3', 4, now() - interval '48 hours 15 minutes', '{"bestellungId": "a0003000-0000-0000-0000-000000000002","positionen": [{"positionId": "b0003000-0000-0000-0000-000000000004","varianteId": 16,"produktName": "Kuchen","varianteName": "Stück","kategorie": "food","einzelpreis": 250,"menge": 2},{"positionId": "b0003000-0000-0000-0000-000000000005","varianteId": 45,"produktName": "Kaffee","varianteName": "Tasse","kategorie": "beverage","einzelpreis": 200,"menge": 2}],"gesamtPreisCents": 900,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:3', 5, now() - interval '48 hours 5 minutes', '{"lieferungId": "c0003000-0000-0000-0000-000000000002","positionen": [{"positionId": "b0003000-0000-0000-0000-000000000004","menge": 2},{"positionId": "b0003000-0000-0000-0000-000000000005","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:3', 6, now() - interval '47 hours 49 minutes', '{"zahlungId": "d0003000-0000-0000-0000-000000000002","positionen": [{"positionId": "b0003000-0000-0000-0000-000000000004","menge": 2},{"positionId": "b0003000-0000-0000-0000-000000000005","menge": 2}],"gesamtZahlungCents": 900,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.bestellung-aufgegeben:v1', 'tisch:3', 7, now() - interval '25 hours 50 minutes', '{"bestellungId": "a0003000-0000-0000-0000-000000000003","positionen": [{"positionId": "b0003000-0000-0000-0000-000000000006","varianteId": 10,"produktName": "Tagesgericht","varianteName": "Sa: Gulasch mit Spätzle","kategorie": "food","einzelpreis": 1150,"menge": 2},{"positionId": "b0003000-0000-0000-0000-000000000007","varianteId": 44,"produktName": "Wein","varianteName": "Rosé 0,2l","kategorie": "beverage","einzelpreis": 400,"menge": 2},{"positionId": "b0003000-0000-0000-0000-000000000008","varianteId": 37,"produktName": "Wasser","varianteName": "Medium 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 1}],"gesamtPreisCents": 3300,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.produkte-geliefert:v1', 'tisch:3', 8, now() - interval '25 hours 37 minutes', '{"lieferungId": "c0003000-0000-0000-0000-000000000003","positionen": [{"positionId": "b0003000-0000-0000-0000-000000000006","menge": 2},{"positionId": "b0003000-0000-0000-0000-000000000007","menge": 2},{"positionId": "b0003000-0000-0000-0000-000000000008","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.zahlung-registriert:v1', 'tisch:3', 9, now() - interval '25 hours 20 minutes', '{"zahlungId": "d0003000-0000-0000-0000-000000000003","positionen": [{"positionId": "b0003000-0000-0000-0000-000000000006","menge": 2},{"positionId": "b0003000-0000-0000-0000-000000000007","menge": 2},{"positionId": "b0003000-0000-0000-0000-000000000008","menge": 1}],"gesamtZahlungCents": 3300,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.bestellung-aufgegeben:v1', 'tisch:3', 10, now() - interval '24 hours 48 minutes', '{"bestellungId": "a0003000-0000-0000-0000-000000000004","positionen": [{"positionId": "b0003000-0000-0000-0000-000000000009","varianteId": 17,"produktName": "Waffeln","varianteName": "mit Puderzucker","kategorie": "food","einzelpreis": 300,"menge": 2},{"positionId": "b0003000-0000-0000-0000-000000000010","varianteId": 47,"produktName": "Tee","varianteName": "Verschiedene Sorten","kategorie": "beverage","einzelpreis": 200,"menge": 2}],"gesamtPreisCents": 1000,"kommentar": "Waffeln zum Nachtisch"}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.produkte-geliefert:v1', 'tisch:3', 11, now() - interval '24 hours 38 minutes', '{"lieferungId": "c0003000-0000-0000-0000-000000000004","positionen": [{"positionId": "b0003000-0000-0000-0000-000000000009","menge": 2},{"positionId": "b0003000-0000-0000-0000-000000000010","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.zahlung-registriert:v1', 'tisch:3', 12, now() - interval '24 hours 16 minutes', '{"zahlungId": "d0003000-0000-0000-0000-000000000004","positionen": [{"positionId": "b0003000-0000-0000-0000-000000000009","menge": 2},{"positionId": "b0003000-0000-0000-0000-000000000010","menge": 2}],"gesamtZahlungCents": 1000,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.bestellung-aufgegeben:v1', 'tisch:3', 13, now() - interval '20 hours', '{"bestellungId": "a0003000-0000-0000-0000-000000000005","positionen": [{"positionId": "b0003000-0000-0000-0000-000000000011","varianteId": 6,"produktName": "Flammkuchen","varianteName": "Classic","kategorie": "food","einzelpreis": 600,"menge": 1},{"positionId": "b0003000-0000-0000-0000-000000000012","varianteId": 43,"produktName": "Wein","varianteName": "Rotwein 0,2l","kategorie": "beverage","einzelpreis": 400,"menge": 1},{"positionId": "b0003000-0000-0000-0000-000000000013","varianteId": 38,"produktName": "Wasser","varianteName": "Sprudel 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 1}],"gesamtPreisCents": 1200,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.produkte-geliefert:v1', 'tisch:3', 14, now() - interval '19 hours 47 minutes', '{"lieferungId": "c0003000-0000-0000-0000-000000000005","positionen": [{"positionId": "b0003000-0000-0000-0000-000000000011","menge": 1},{"positionId": "b0003000-0000-0000-0000-000000000012","menge": 1},{"positionId": "b0003000-0000-0000-0000-000000000013","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.zahlung-registriert:v1', 'tisch:3', 15, now() - interval '19 hours 32 minutes', '{"zahlungId": "d0003000-0000-0000-0000-000000000005","positionen": [{"positionId": "b0003000-0000-0000-0000-000000000011","menge": 1},{"positionId": "b0003000-0000-0000-0000-000000000012","menge": 1},{"positionId": "b0003000-0000-0000-0000-000000000013","menge": 1}],"gesamtZahlungCents": 1200,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.bestellung-aufgegeben:v1', 'tisch:3', 16, now() - interval '19 hours 11 minutes', '{"bestellungId": "a0003000-0000-0000-0000-000000000006","positionen": [{"positionId": "b0003000-0000-0000-0000-000000000014","varianteId": 46,"produktName": "Kaffee","varianteName": "Espresso","kategorie": "beverage","einzelpreis": 180,"menge": 2}],"gesamtPreisCents": 360,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.produkte-geliefert:v1', 'tisch:3', 17, now() - interval '19 hours 2 minutes', '{"lieferungId": "c0003000-0000-0000-0000-000000000006","positionen": [{"positionId": "b0003000-0000-0000-0000-000000000014","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.zahlung-registriert:v1', 'tisch:3', 18, now() - interval '18 hours 37 minutes', '{"zahlungId": "d0003000-0000-0000-0000-000000000006","positionen": [{"positionId": "b0003000-0000-0000-0000-000000000014","menge": 2}],"gesamtZahlungCents": 360,"kommentar": ""}');

-- ─────────────────────────────────────────────────────────────────────────────
-- TISCH 4: Großfamilie — Tag 2, große Bestellungen (57 Events)
-- ─────────────────────────────────────────────────────────────────────────────

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:4', 1, now() - interval '26 hours 40 minutes', '{"bestellungId": "a0004000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0004000-0000-0000-0000-000000000001","varianteId": 1,"produktName": "Bratwurst","varianteName": "Normal","kategorie": "food","einzelpreis": 350,"menge": 3},{"positionId": "b0004000-0000-0000-0000-000000000002","varianteId": 3,"produktName": "Bratwurst","varianteName": "Currywurst","kategorie": "food","einzelpreis": 450,"menge": 2},{"positionId": "b0004000-0000-0000-0000-000000000003","varianteId": 5,"produktName": "Pommes","varianteName": "Groß","kategorie": "food","einzelpreis": 350,"menge": 4},{"positionId": "b0004000-0000-0000-0000-000000000004","varianteId": 10,"produktName": "Tagesgericht","varianteName": "Sa: Gulasch mit Spätzle","kategorie": "food","einzelpreis": 1150,"menge": 2},{"positionId": "b0004000-0000-0000-0000-000000000005","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 4},{"positionId": "b0004000-0000-0000-0000-000000000006","varianteId": 33,"produktName": "Softdrinks","varianteName": "Spezi","kategorie": "beverage","einzelpreis": 280,"menge": 4},{"positionId": "b0004000-0000-0000-0000-000000000007","varianteId": 39,"produktName": "Saftschorle","varianteName": "Apfelschorle 0,5l","kategorie": "beverage","einzelpreis": 300,"menge": 2}],"gesamtPreisCents": 9170,"kommentar": "Großfamilie mit Kindern"}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.produkte-geliefert:v1', 'tisch:4', 2, now() - interval '26 hours 28 minutes', '{"lieferungId": "c0004000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0004000-0000-0000-0000-000000000001","menge": 3},{"positionId": "b0004000-0000-0000-0000-000000000002","menge": 2},{"positionId": "b0004000-0000-0000-0000-000000000003","menge": 4},{"positionId": "b0004000-0000-0000-0000-000000000004","menge": 2},{"positionId": "b0004000-0000-0000-0000-000000000005","menge": 4},{"positionId": "b0004000-0000-0000-0000-000000000006","menge": 4},{"positionId": "b0004000-0000-0000-0000-000000000007","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.zahlung-registriert:v1', 'tisch:4', 3, now() - interval '26 hours 8 minutes', '{"zahlungId": "d0004000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0004000-0000-0000-0000-000000000005","menge": 4},{"positionId": "b0004000-0000-0000-0000-000000000006","menge": 4},{"positionId": "b0004000-0000-0000-0000-000000000007","menge": 2}],"gesamtZahlungCents": 3520,"kommentar": "Erst mal nur Getränke"}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.produkte-storniert:v1', 'tisch:4', 4, now() - interval '25 hours 53 minutes', '{"stornierungId": "e0004000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0004000-0000-0000-0000-000000000004","menge": 1}],"gesamtStornierungCents": 1150,"kommentar": "Kind verträgt Gulasch nicht"}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:4', 5, now() - interval '25 hours 48 minutes', '{"bestellungId": "a0004000-0000-0000-0000-000000000002","positionen": [{"positionId": "b0004000-0000-0000-0000-000000000008","varianteId": 16,"produktName": "Kuchen","varianteName": "Stück","kategorie": "food","einzelpreis": 250,"menge": 4},{"positionId": "b0004000-0000-0000-0000-000000000009","varianteId": 19,"produktName": "Waffeln","varianteName": "mit Nutella","kategorie": "food","einzelpreis": 400,"menge": 3},{"positionId": "b0004000-0000-0000-0000-000000000010","varianteId": 31,"produktName": "Softdrinks","varianteName": "Cola","kategorie": "beverage","einzelpreis": 280,"menge": 4},{"positionId": "b0004000-0000-0000-0000-000000000011","varianteId": 34,"produktName": "Softdrinks","varianteName": "Sprite","kategorie": "beverage","einzelpreis": 280,"menge": 2}],"gesamtPreisCents": 3880,"kommentar": "Nachtisch für die Kinder"}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.produkte-geliefert:v1', 'tisch:4', 6, now() - interval '25 hours 40 minutes', '{"lieferungId": "c0004000-0000-0000-0000-000000000002","positionen": [{"positionId": "b0004000-0000-0000-0000-000000000008","menge": 4},{"positionId": "b0004000-0000-0000-0000-000000000009","menge": 3},{"positionId": "b0004000-0000-0000-0000-000000000010","menge": 4},{"positionId": "b0004000-0000-0000-0000-000000000011","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:4', 7, now() - interval '25 hours 20 minutes', '{"bestellungId": "a0004000-0000-0000-0000-000000000003","positionen": [{"positionId": "b0004000-0000-0000-0000-000000000012","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 3},{"positionId": "b0004000-0000-0000-0000-000000000013","varianteId": 27,"produktName": "Weizen","varianteName": "Groß 0,5l","kategorie": "beverage","einzelpreis": 400,"menge": 2}],"gesamtPreisCents": 2150,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.produkte-geliefert:v1', 'tisch:4', 8, now() - interval '25 hours 5 minutes', '{"lieferungId": "c0004000-0000-0000-0000-000000000003","positionen": [{"positionId": "b0004000-0000-0000-0000-000000000012","menge": 3},{"positionId": "b0004000-0000-0000-0000-000000000013","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.zahlung-registriert:v1', 'tisch:4', 9, now() - interval '24 hours 34 minutes', '{"zahlungId": "d0004000-0000-0000-0000-000000000002","positionen": [{"positionId": "b0004000-0000-0000-0000-000000000001","menge": 3},{"positionId": "b0004000-0000-0000-0000-000000000002","menge": 2},{"positionId": "b0004000-0000-0000-0000-000000000003","menge": 4},{"positionId": "b0004000-0000-0000-0000-000000000004","menge": 2},{"positionId": "b0004000-0000-0000-0000-000000000008","menge": 4},{"positionId": "b0004000-0000-0000-0000-000000000009","menge": 3},{"positionId": "b0004000-0000-0000-0000-000000000010","menge": 4},{"positionId": "b0004000-0000-0000-0000-000000000011","menge": 2}],"gesamtZahlungCents": 9530,"kommentar": "Essen + Nachtisch"}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:4', 10, now() - interval '24 hours 9 minutes', '{"bestellungId": "a0004000-0000-0000-0000-000000000004","positionen": [{"positionId": "b0004000-0000-0000-0000-000000000014","varianteId": 20,"produktName": "Brezel","varianteName": "Normal","kategorie": "food","einzelpreis": 200,"menge": 4},{"positionId": "b0004000-0000-0000-0000-000000000015","varianteId": 21,"produktName": "Brezel","varianteName": "mit Butter","kategorie": "food","einzelpreis": 300,"menge": 2},{"positionId": "b0004000-0000-0000-0000-000000000016","varianteId": 36,"produktName": "Wasser","varianteName": "Still 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 4}],"gesamtPreisCents": 2200,"kommentar": "Brezeln als Snack"}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.produkte-geliefert:v1', 'tisch:4', 11, now() - interval '23 hours 56 minutes', '{"lieferungId": "c0004000-0000-0000-0000-000000000004","positionen": [{"positionId": "b0004000-0000-0000-0000-000000000014","menge": 4},{"positionId": "b0004000-0000-0000-0000-000000000015","menge": 2},{"positionId": "b0004000-0000-0000-0000-000000000016","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.zahlung-registriert:v1', 'tisch:4', 12, now() - interval '23 hours 41 minutes', '{"zahlungId": "d0004000-0000-0000-0000-000000000003","positionen": [{"positionId": "b0004000-0000-0000-0000-000000000014","menge": 4},{"positionId": "b0004000-0000-0000-0000-000000000015","menge": 2},{"positionId": "b0004000-0000-0000-0000-000000000016","menge": 4}],"gesamtZahlungCents": 2200,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:4', 13, now() - interval '23 hours 25 minutes', '{"bestellungId": "a0004000-0000-0000-0000-000000000005","positionen": [{"positionId": "b0004000-0000-0000-0000-000000000017","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 2},{"positionId": "b0004000-0000-0000-0000-000000000018","varianteId": 29,"produktName": "Weizen","varianteName": "Colaweizen Groß","kategorie": "beverage","einzelpreis": 400,"menge": 2}],"gesamtPreisCents": 1700,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.produkte-geliefert:v1', 'tisch:4', 14, now() - interval '23 hours 14 minutes', '{"lieferungId": "c0004000-0000-0000-0000-000000000005","positionen": [{"positionId": "b0004000-0000-0000-0000-000000000017","menge": 2},{"positionId": "b0004000-0000-0000-0000-000000000018","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.zahlung-registriert:v1', 'tisch:4', 15, now() - interval '23 hours 2 minutes', '{"zahlungId": "d0004000-0000-0000-0000-000000000004","positionen": [{"positionId": "b0004000-0000-0000-0000-000000000012","menge": 3},{"positionId": "b0004000-0000-0000-0000-000000000013","menge": 2},{"positionId": "b0004000-0000-0000-0000-000000000017","menge": 2},{"positionId": "b0004000-0000-0000-0000-000000000018","menge": 2}],"gesamtZahlungCents": 3850,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:4', 16, now() - interval '22 hours 32 minutes', '{"bestellungId": "a0004000-0000-0000-0000-000000000006","positionen": [{"positionId": "b0004000-0000-0000-0000-000000000019","varianteId": 45,"produktName": "Kaffee","varianteName": "Tasse","kategorie": "beverage","einzelpreis": 200,"menge": 3},{"positionId": "b0004000-0000-0000-0000-000000000020","varianteId": 46,"produktName": "Kaffee","varianteName": "Espresso","kategorie": "beverage","einzelpreis": 180,"menge": 1},{"positionId": "b0004000-0000-0000-0000-000000000021","varianteId": 16,"produktName": "Kuchen","varianteName": "Stück","kategorie": "food","einzelpreis": 250,"menge": 3},{"positionId": "b0004000-0000-0000-0000-000000000022","varianteId": 18,"produktName": "Waffeln","varianteName": "mit Sahne","kategorie": "food","einzelpreis": 350,"menge": 2}],"gesamtPreisCents": 2230,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.produkte-geliefert:v1', 'tisch:4', 17, now() - interval '22 hours 21 minutes', '{"lieferungId": "c0004000-0000-0000-0000-000000000006","positionen": [{"positionId": "b0004000-0000-0000-0000-000000000019","menge": 3},{"positionId": "b0004000-0000-0000-0000-000000000020","menge": 1},{"positionId": "b0004000-0000-0000-0000-000000000021","menge": 3},{"positionId": "b0004000-0000-0000-0000-000000000022","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.zahlung-registriert:v1', 'tisch:4', 18, now() - interval '21 hours 58 minutes', '{"zahlungId": "d0004000-0000-0000-0000-000000000005","positionen": [{"positionId": "b0004000-0000-0000-0000-000000000019","menge": 3},{"positionId": "b0004000-0000-0000-0000-000000000020","menge": 1},{"positionId": "b0004000-0000-0000-0000-000000000021","menge": 3},{"positionId": "b0004000-0000-0000-0000-000000000022","menge": 2}],"gesamtZahlungCents": 2230,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.bestellung-aufgegeben:v1', 'tisch:4', 19, now() - interval '19 hours 6 minutes', '{"bestellungId": "a0004000-0000-0000-0000-000000000010","positionen": [{"positionId": "b0004000-0000-0000-0000-000000000032","varianteId": 10,"produktName": "Tagesgericht","varianteName": "Sa: Gulasch mit Spätzle","kategorie": "food","einzelpreis": 1150,"menge": 1},{"positionId": "b0004000-0000-0000-0000-000000000033","varianteId": 12,"produktName": "Grillplatte","varianteName": "Klein","kategorie": "food","einzelpreis": 800,"menge": 1},{"positionId": "b0004000-0000-0000-0000-000000000034","varianteId": 40,"produktName": "Saftschorle","varianteName": "Johannisbeerschorle 0,5l","kategorie": "beverage","einzelpreis": 300,"menge": 4}],"gesamtPreisCents": 3150,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:4', 20, now() - interval '19 hours 4 minutes', '{"bestellungId": "a0004000-0000-0000-0000-000000000012","positionen": [{"positionId": "b0004000-0000-0000-0000-000000000038","varianteId": 4,"produktName": "Pommes","varianteName": "Klein","kategorie": "food","einzelpreis": 250,"menge": 2},{"positionId": "b0004000-0000-0000-0000-000000000039","varianteId": 18,"produktName": "Waffeln","varianteName": "mit Sahne","kategorie": "food","einzelpreis": 350,"menge": 2},{"positionId": "b0004000-0000-0000-0000-000000000040","varianteId": 37,"produktName": "Wasser","varianteName": "Medium 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 2},{"positionId": "b0004000-0000-0000-0000-000000000041","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 2}],"gesamtPreisCents": 2500,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.bestellung-aufgegeben:v1', 'tisch:4', 21, now() - interval '18 hours 59 minutes', '{"bestellungId": "a0004000-0000-0000-0000-000000000016","positionen": [{"positionId": "b0004000-0000-0000-0000-000000000051","varianteId": 4,"produktName": "Pommes","varianteName": "Klein","kategorie": "food","einzelpreis": 250,"menge": 2},{"positionId": "b0004000-0000-0000-0000-000000000052","varianteId": 15,"produktName": "Salat","varianteName": "Caesar Salat","kategorie": "food","einzelpreis": 650,"menge": 3},{"positionId": "b0004000-0000-0000-0000-000000000053","varianteId": 27,"produktName": "Weizen","varianteName": "Groß 0,5l","kategorie": "beverage","einzelpreis": 400,"menge": 3}],"gesamtPreisCents": 3650,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.produkte-geliefert:v1', 'tisch:4', 22, now() - interval '18 hours 55 minutes', '{"lieferungId": "c0004000-0000-0000-0000-000000000010","positionen": [{"positionId": "b0004000-0000-0000-0000-000000000032","menge": 1},{"positionId": "b0004000-0000-0000-0000-000000000033","menge": 1},{"positionId": "b0004000-0000-0000-0000-000000000034","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.produkte-geliefert:v1', 'tisch:4', 23, now() - interval '18 hours 52 minutes', '{"lieferungId": "c0004000-0000-0000-0000-000000000012","positionen": [{"positionId": "b0004000-0000-0000-0000-000000000038","menge": 2},{"positionId": "b0004000-0000-0000-0000-000000000039","menge": 2},{"positionId": "b0004000-0000-0000-0000-000000000040","menge": 2},{"positionId": "b0004000-0000-0000-0000-000000000041","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:4', 24, now() - interval '18 hours 51 minutes', '{"bestellungId": "a0004000-0000-0000-0000-000000000009","positionen": [{"positionId": "b0004000-0000-0000-0000-000000000030","varianteId": 20,"produktName": "Brezel","varianteName": "Normal","kategorie": "food","einzelpreis": 200,"menge": 2},{"positionId": "b0004000-0000-0000-0000-000000000031","varianteId": 42,"produktName": "Wein","varianteName": "Weißwein 0,2l","kategorie": "beverage","einzelpreis": 400,"menge": 4}],"gesamtPreisCents": 2000,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:4', 25, now() - interval '18 hours 47 minutes', '{"bestellungId": "a0004000-0000-0000-0000-000000000008","positionen": [{"positionId": "b0004000-0000-0000-0000-000000000027","varianteId": 14,"produktName": "Salat","varianteName": "Gemischter Salat","kategorie": "food","einzelpreis": 550,"menge": 3},{"positionId": "b0004000-0000-0000-0000-000000000028","varianteId": 7,"produktName": "Flammkuchen","varianteName": "Speck & Zwiebel","kategorie": "food","einzelpreis": 700,"menge": 3},{"positionId": "b0004000-0000-0000-0000-000000000029","varianteId": 38,"produktName": "Wasser","varianteName": "Sprudel 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 1}],"gesamtPreisCents": 3950,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.produkte-geliefert:v1', 'tisch:4', 26, now() - interval '18 hours 46 minutes', '{"lieferungId": "c0004000-0000-0000-0000-000000000016","positionen": [{"positionId": "b0004000-0000-0000-0000-000000000051","menge": 2},{"positionId": "b0004000-0000-0000-0000-000000000052","menge": 3},{"positionId": "b0004000-0000-0000-0000-000000000053","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:4', 27, now() - interval '18 hours 46 minutes', '{"bestellungId": "a0004000-0000-0000-0000-000000000017","positionen": [{"positionId": "b0004000-0000-0000-0000-000000000054","varianteId": 8,"produktName": "Flammkuchen","varianteName": "Mediterran","kategorie": "food","einzelpreis": 750,"menge": 3},{"positionId": "b0004000-0000-0000-0000-000000000055","varianteId": 26,"produktName": "Weizen","varianteName": "Klein 0,3l","kategorie": "beverage","einzelpreis": 300,"menge": 3}],"gesamtPreisCents": 3150,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.produkte-geliefert:v1', 'tisch:4', 28, now() - interval '18 hours 43 minutes', '{"lieferungId": "c0004000-0000-0000-0000-000000000009","positionen": [{"positionId": "b0004000-0000-0000-0000-000000000030","menge": 2},{"positionId": "b0004000-0000-0000-0000-000000000031","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.bestellung-aufgegeben:v1', 'tisch:4', 29, now() - interval '18 hours 43 minutes', '{"bestellungId": "a0004000-0000-0000-0000-000000000011","positionen": [{"positionId": "b0004000-0000-0000-0000-000000000035","varianteId": 18,"produktName": "Waffeln","varianteName": "mit Sahne","kategorie": "food","einzelpreis": 350,"menge": 1},{"positionId": "b0004000-0000-0000-0000-000000000036","varianteId": 30,"produktName": "Weizen","varianteName": "Russ","kategorie": "beverage","einzelpreis": 300,"menge": 1},{"positionId": "b0004000-0000-0000-0000-000000000037","varianteId": 30,"produktName": "Weizen","varianteName": "Russ","kategorie": "beverage","einzelpreis": 300,"menge": 4}],"gesamtPreisCents": 1850,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.bestellung-aufgegeben:v1', 'tisch:4', 30, now() - interval '18 hours 40 minutes', '{"bestellungId": "a0004000-0000-0000-0000-000000000018","positionen": [{"positionId": "b0004000-0000-0000-0000-000000000056","varianteId": 5,"produktName": "Pommes","varianteName": "Groß","kategorie": "food","einzelpreis": 350,"menge": 1},{"positionId": "b0004000-0000-0000-0000-000000000057","varianteId": 31,"produktName": "Softdrinks","varianteName": "Cola","kategorie": "beverage","einzelpreis": 280,"menge": 2},{"positionId": "b0004000-0000-0000-0000-000000000058","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 3}],"gesamtPreisCents": 2260,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:4', 31, now() - interval '18 hours 38 minutes', '{"lieferungId": "c0004000-0000-0000-0000-000000000008","positionen": [{"positionId": "b0004000-0000-0000-0000-000000000027","menge": 3},{"positionId": "b0004000-0000-0000-0000-000000000028","menge": 3},{"positionId": "b0004000-0000-0000-0000-000000000029","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:4', 32, now() - interval '18 hours 38 minutes', '{"bestellungId": "a0004000-0000-0000-0000-000000000013","positionen": [{"positionId": "b0004000-0000-0000-0000-000000000042","varianteId": 19,"produktName": "Waffeln","varianteName": "mit Nutella","kategorie": "food","einzelpreis": 400,"menge": 1},{"positionId": "b0004000-0000-0000-0000-000000000043","varianteId": 14,"produktName": "Salat","varianteName": "Gemischter Salat","kategorie": "food","einzelpreis": 550,"menge": 3},{"positionId": "b0004000-0000-0000-0000-000000000044","varianteId": 41,"produktName": "Saftschorle","varianteName": "Rhabarberschorle 0,5l","kategorie": "beverage","einzelpreis": 350,"menge": 4},{"positionId": "b0004000-0000-0000-0000-000000000045","varianteId": 28,"produktName": "Weizen","varianteName": "Colaweizen Klein","kategorie": "beverage","einzelpreis": 300,"menge": 1}],"gesamtPreisCents": 3750,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.produkte-geliefert:v1', 'tisch:4', 33, now() - interval '18 hours 35 minutes', '{"lieferungId": "c0004000-0000-0000-0000-000000000011","positionen": [{"positionId": "b0004000-0000-0000-0000-000000000035","menge": 1},{"positionId": "b0004000-0000-0000-0000-000000000036","menge": 1},{"positionId": "b0004000-0000-0000-0000-000000000037","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.zahlung-registriert:v1', 'tisch:4', 34, now() - interval '18 hours 33 minutes', '{"zahlungId": "d0004000-0000-0000-0000-000000000009","positionen": [{"positionId": "b0004000-0000-0000-0000-000000000032","menge": 1},{"positionId": "b0004000-0000-0000-0000-000000000033","menge": 1},{"positionId": "b0004000-0000-0000-0000-000000000034","menge": 4}],"gesamtZahlungCents": 3150,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.zahlung-registriert:v1', 'tisch:4', 35, now() - interval '18 hours 32 minutes', '{"zahlungId": "d0004000-0000-0000-0000-000000000015","positionen": [{"positionId": "b0004000-0000-0000-0000-000000000051","menge": 2},{"positionId": "b0004000-0000-0000-0000-000000000052","menge": 3},{"positionId": "b0004000-0000-0000-0000-000000000053","menge": 3}],"gesamtZahlungCents": 3650,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.produkte-geliefert:v1', 'tisch:4', 36, now() - interval '18 hours 32 minutes', '{"lieferungId": "c0004000-0000-0000-0000-000000000018","positionen": [{"positionId": "b0004000-0000-0000-0000-000000000056","menge": 1},{"positionId": "b0004000-0000-0000-0000-000000000057","menge": 2},{"positionId": "b0004000-0000-0000-0000-000000000058","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:4', 37, now() - interval '18 hours 31 minutes', '{"lieferungId": "c0004000-0000-0000-0000-000000000017","positionen": [{"positionId": "b0004000-0000-0000-0000-000000000054","menge": 3},{"positionId": "b0004000-0000-0000-0000-000000000055","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.zahlung-registriert:v1', 'tisch:4', 38, now() - interval '18 hours 29 minutes', '{"zahlungId": "d0004000-0000-0000-0000-000000000011","positionen": [{"positionId": "b0004000-0000-0000-0000-000000000038","menge": 2},{"positionId": "b0004000-0000-0000-0000-000000000039","menge": 2},{"positionId": "b0004000-0000-0000-0000-000000000040","menge": 2},{"positionId": "b0004000-0000-0000-0000-000000000041","menge": 2}],"gesamtZahlungCents": 2500,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.zahlung-registriert:v1', 'tisch:4', 39, now() - interval '18 hours 25 minutes', '{"zahlungId": "d0004000-0000-0000-0000-000000000008","positionen": [{"positionId": "b0004000-0000-0000-0000-000000000030","menge": 2},{"positionId": "b0004000-0000-0000-0000-000000000031","menge": 4}],"gesamtZahlungCents": 2000,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.produkte-geliefert:v1', 'tisch:4', 40, now() - interval '18 hours 23 minutes', '{"lieferungId": "c0004000-0000-0000-0000-000000000013","positionen": [{"positionId": "b0004000-0000-0000-0000-000000000042","menge": 1},{"positionId": "b0004000-0000-0000-0000-000000000043","menge": 3},{"positionId": "b0004000-0000-0000-0000-000000000044","menge": 4},{"positionId": "b0004000-0000-0000-0000-000000000045","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.zahlung-registriert:v1', 'tisch:4', 41, now() - interval '18 hours 18 minutes', '{"zahlungId": "d0004000-0000-0000-0000-000000000017","positionen": [{"positionId": "b0004000-0000-0000-0000-000000000056","menge": 1},{"positionId": "b0004000-0000-0000-0000-000000000057","menge": 2},{"positionId": "b0004000-0000-0000-0000-000000000058","menge": 3}],"gesamtZahlungCents": 2260,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:4', 42, now() - interval '18 hours 17 minutes', '{"zahlungId": "d0004000-0000-0000-0000-000000000007","positionen": [{"positionId": "b0004000-0000-0000-0000-000000000027","menge": 3},{"positionId": "b0004000-0000-0000-0000-000000000028","menge": 3},{"positionId": "b0004000-0000-0000-0000-000000000029","menge": 1}],"gesamtZahlungCents": 3950,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:4', 43, now() - interval '18 hours 17 minutes', '{"zahlungId": "d0004000-0000-0000-0000-000000000016","positionen": [{"positionId": "b0004000-0000-0000-0000-000000000054","menge": 3},{"positionId": "b0004000-0000-0000-0000-000000000055","menge": 3}],"gesamtZahlungCents": 3150,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.zahlung-registriert:v1', 'tisch:4', 44, now() - interval '18 hours 15 minutes', '{"zahlungId": "d0004000-0000-0000-0000-000000000010","positionen": [{"positionId": "b0004000-0000-0000-0000-000000000035","menge": 1},{"positionId": "b0004000-0000-0000-0000-000000000036","menge": 1},{"positionId": "b0004000-0000-0000-0000-000000000037","menge": 4}],"gesamtZahlungCents": 1850,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.bestellung-aufgegeben:v1', 'tisch:4', 45, now() - interval '18 hours 12 minutes', '{"bestellungId": "a0004000-0000-0000-0000-000000000014","positionen": [{"positionId": "b0004000-0000-0000-0000-000000000046","varianteId": 5,"produktName": "Pommes","varianteName": "Groß","kategorie": "food","einzelpreis": 350,"menge": 1},{"positionId": "b0004000-0000-0000-0000-000000000047","varianteId": 47,"produktName": "Tee","varianteName": "Verschiedene Sorten","kategorie": "beverage","einzelpreis": 200,"menge": 2}],"gesamtPreisCents": 750,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.zahlung-registriert:v1', 'tisch:4', 46, now() - interval '18 hours 9 minutes', '{"zahlungId": "d0004000-0000-0000-0000-000000000012","positionen": [{"positionId": "b0004000-0000-0000-0000-000000000042","menge": 1},{"positionId": "b0004000-0000-0000-0000-000000000043","menge": 3},{"positionId": "b0004000-0000-0000-0000-000000000044","menge": 4},{"positionId": "b0004000-0000-0000-0000-000000000045","menge": 1}],"gesamtZahlungCents": 3750,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.produkte-geliefert:v1', 'tisch:4', 47, now() - interval '18 hours 1 minutes', '{"lieferungId": "c0004000-0000-0000-0000-000000000014","positionen": [{"positionId": "b0004000-0000-0000-0000-000000000046","menge": 1},{"positionId": "b0004000-0000-0000-0000-000000000047","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:4', 48, now() - interval '18 hours', '{"bestellungId": "a0004000-0000-0000-0000-000000000007","positionen": [{"positionId": "b0004000-0000-0000-0000-000000000023","varianteId": 17,"produktName": "Waffeln","varianteName": "mit Puderzucker","kategorie": "food","einzelpreis": 300,"menge": 2},{"positionId": "b0004000-0000-0000-0000-000000000024","varianteId": 5,"produktName": "Pommes","varianteName": "Groß","kategorie": "food","einzelpreis": 350,"menge": 2},{"positionId": "b0004000-0000-0000-0000-000000000025","varianteId": 42,"produktName": "Wein","varianteName": "Weißwein 0,2l","kategorie": "beverage","einzelpreis": 400,"menge": 4},{"positionId": "b0004000-0000-0000-0000-000000000026","varianteId": 35,"produktName": "Softdrinks","varianteName": "Mezzo Mix","kategorie": "beverage","einzelpreis": 280,"menge": 4}],"gesamtPreisCents": 4020,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:4', 49, now() - interval '17 hours 59 minutes', '{"bestellungId": "a0004000-0000-0000-0000-000000000019","positionen": [{"positionId": "b0004000-0000-0000-0000-000000000059","varianteId": 17,"produktName": "Waffeln","varianteName": "mit Puderzucker","kategorie": "food","einzelpreis": 300,"menge": 2},{"positionId": "b0004000-0000-0000-0000-000000000060","varianteId": 46,"produktName": "Kaffee","varianteName": "Espresso","kategorie": "beverage","einzelpreis": 180,"menge": 1},{"positionId": "b0004000-0000-0000-0000-000000000061","varianteId": 44,"produktName": "Wein","varianteName": "Rosé 0,2l","kategorie": "beverage","einzelpreis": 400,"menge": 2}],"gesamtPreisCents": 1580,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.bestellung-aufgegeben:v1', 'tisch:4', 50, now() - interval '17 hours 55 minutes', '{"bestellungId": "a0004000-0000-0000-0000-000000000015","positionen": [{"positionId": "b0004000-0000-0000-0000-000000000048","varianteId": 10,"produktName": "Tagesgericht","varianteName": "Sa: Gulasch mit Spätzle","kategorie": "food","einzelpreis": 1150,"menge": 2},{"positionId": "b0004000-0000-0000-0000-000000000049","varianteId": 34,"produktName": "Softdrinks","varianteName": "Sprite","kategorie": "beverage","einzelpreis": 280,"menge": 1},{"positionId": "b0004000-0000-0000-0000-000000000050","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 3}],"gesamtPreisCents": 3930,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.produkte-geliefert:v1', 'tisch:4', 51, now() - interval '17 hours 51 minutes', '{"lieferungId": "c0004000-0000-0000-0000-000000000019","positionen": [{"positionId": "b0004000-0000-0000-0000-000000000059","menge": 2},{"positionId": "b0004000-0000-0000-0000-000000000060","menge": 1},{"positionId": "b0004000-0000-0000-0000-000000000061","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.zahlung-registriert:v1', 'tisch:4', 52, now() - interval '17 hours 49 minutes', '{"zahlungId": "d0004000-0000-0000-0000-000000000013","positionen": [{"positionId": "b0004000-0000-0000-0000-000000000046","menge": 1},{"positionId": "b0004000-0000-0000-0000-000000000047","menge": 2}],"gesamtZahlungCents": 750,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.produkte-geliefert:v1', 'tisch:4', 53, now() - interval '17 hours 47 minutes', '{"lieferungId": "c0004000-0000-0000-0000-000000000007","positionen": [{"positionId": "b0004000-0000-0000-0000-000000000023","menge": 2},{"positionId": "b0004000-0000-0000-0000-000000000024","menge": 2},{"positionId": "b0004000-0000-0000-0000-000000000025","menge": 4},{"positionId": "b0004000-0000-0000-0000-000000000026","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.produkte-geliefert:v1', 'tisch:4', 54, now() - interval '17 hours 44 minutes', '{"lieferungId": "c0004000-0000-0000-0000-000000000015","positionen": [{"positionId": "b0004000-0000-0000-0000-000000000048","menge": 2},{"positionId": "b0004000-0000-0000-0000-000000000049","menge": 1},{"positionId": "b0004000-0000-0000-0000-000000000050","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.zahlung-registriert:v1', 'tisch:4', 55, now() - interval '17 hours 29 minutes', '{"zahlungId": "d0004000-0000-0000-0000-000000000018","positionen": [{"positionId": "b0004000-0000-0000-0000-000000000059","menge": 2},{"positionId": "b0004000-0000-0000-0000-000000000060","menge": 1},{"positionId": "b0004000-0000-0000-0000-000000000061","menge": 2}],"gesamtZahlungCents": 1580,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.zahlung-registriert:v1', 'tisch:4', 56, now() - interval '17 hours 25 minutes', '{"zahlungId": "d0004000-0000-0000-0000-000000000014","positionen": [{"positionId": "b0004000-0000-0000-0000-000000000048","menge": 2},{"positionId": "b0004000-0000-0000-0000-000000000049","menge": 1},{"positionId": "b0004000-0000-0000-0000-000000000050","menge": 3}],"gesamtZahlungCents": 3930,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.zahlung-registriert:v1', 'tisch:4', 57, now() - interval '17 hours 22 minutes', '{"zahlungId": "d0004000-0000-0000-0000-000000000006","positionen": [{"positionId": "b0004000-0000-0000-0000-000000000023","menge": 2},{"positionId": "b0004000-0000-0000-0000-000000000024","menge": 2},{"positionId": "b0004000-0000-0000-0000-000000000025","menge": 4},{"positionId": "b0004000-0000-0000-0000-000000000026","menge": 4}],"gesamtZahlungCents": 4020,"kommentar": ""}');

-- ─────────────────────────────────────────────────────────────────────────────
-- TISCH 5: VIP/Vorstand — alle 3 Tage, gehobene Bestellungen (95 Events)
-- ─────────────────────────────────────────────────────────────────────────────

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.bestellung-aufgegeben:v1', 'tisch:5', 1, now() - interval '47 hours 30 minutes', '{"bestellungId": "a0005000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000001","varianteId": 13,"produktName": "Grillplatte","varianteName": "Groß","kategorie": "food","einzelpreis": 1400,"menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000002","varianteId": 8,"produktName": "Flammkuchen","varianteName": "Mediterran","kategorie": "food","einzelpreis": 750,"menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000003","varianteId": 42,"produktName": "Wein","varianteName": "Weißwein 0,2l","kategorie": "beverage","einzelpreis": 400,"menge": 4},{"positionId": "b0005000-0000-0000-0000-000000000004","varianteId": 48,"produktName": "Hugo/Aperol","varianteName": "Hugo","kategorie": "beverage","einzelpreis": 550,"menge": 2}],"gesamtPreisCents": 7000,"kommentar": "Vorstandsessen Freitag"}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.produkte-geliefert:v1', 'tisch:5', 2, now() - interval '47 hours 18 minutes', '{"lieferungId": "c0005000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000001","menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000002","menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000003","menge": 4},{"positionId": "b0005000-0000-0000-0000-000000000004","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.zahlung-registriert:v1', 'tisch:5', 3, now() - interval '46 hours 59 minutes', '{"zahlungId": "d0005000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000001","menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000002","menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000003","menge": 4},{"positionId": "b0005000-0000-0000-0000-000000000004","menge": 2}],"gesamtZahlungCents": 7000,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.bestellung-aufgegeben:v1', 'tisch:5', 4, now() - interval '46 hours 38 minutes', '{"bestellungId": "a0005000-0000-0000-0000-000000000002","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000005","varianteId": 9,"produktName": "Tagesgericht","varianteName": "Fr: Schnitzel mit Pommes","kategorie": "food","einzelpreis": 1250,"menge": 3},{"positionId": "b0005000-0000-0000-0000-000000000006","varianteId": 15,"produktName": "Salat","varianteName": "Caesar Salat","kategorie": "food","einzelpreis": 650,"menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000007","varianteId": 43,"produktName": "Wein","varianteName": "Rotwein 0,2l","kategorie": "beverage","einzelpreis": 400,"menge": 3}],"gesamtPreisCents": 6250,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.produkte-geliefert:v1', 'tisch:5', 5, now() - interval '46 hours 24 minutes', '{"lieferungId": "c0005000-0000-0000-0000-000000000002","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000005","menge": 3},{"positionId": "b0005000-0000-0000-0000-000000000006","menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000007","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.zahlung-registriert:v1', 'tisch:5', 6, now() - interval '46 hours 9 minutes', '{"zahlungId": "d0005000-0000-0000-0000-000000000002","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000005","menge": 3},{"positionId": "b0005000-0000-0000-0000-000000000006","menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000007","menge": 3}],"gesamtZahlungCents": 6250,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.bestellung-aufgegeben:v1', 'tisch:5', 7, now() - interval '45 hours 53 minutes', '{"bestellungId": "a0005000-0000-0000-0000-000000000003","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000008","varianteId": 16,"produktName": "Kuchen","varianteName": "Stück","kategorie": "food","einzelpreis": 250,"menge": 3},{"positionId": "b0005000-0000-0000-0000-000000000009","varianteId": 46,"produktName": "Kaffee","varianteName": "Espresso","kategorie": "beverage","einzelpreis": 180,"menge": 3},{"positionId": "b0005000-0000-0000-0000-000000000010","varianteId": 47,"produktName": "Tee","varianteName": "Verschiedene Sorten","kategorie": "beverage","einzelpreis": 200,"menge": 2}],"gesamtPreisCents": 1690,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.produkte-geliefert:v1', 'tisch:5', 8, now() - interval '45 hours 41 minutes', '{"lieferungId": "c0005000-0000-0000-0000-000000000003","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000008","menge": 3},{"positionId": "b0005000-0000-0000-0000-000000000009","menge": 3},{"positionId": "b0005000-0000-0000-0000-000000000010","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.zahlung-registriert:v1', 'tisch:5', 9, now() - interval '45 hours 25 minutes', '{"zahlungId": "d0005000-0000-0000-0000-000000000003","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000008","menge": 3},{"positionId": "b0005000-0000-0000-0000-000000000009","menge": 3},{"positionId": "b0005000-0000-0000-0000-000000000010","menge": 2}],"gesamtZahlungCents": 1690,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.bestellung-aufgegeben:v1', 'tisch:5', 10, now() - interval '45 hours 6 minutes', '{"bestellungId": "a0005000-0000-0000-0000-000000000004","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000011","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 4},{"positionId": "b0005000-0000-0000-0000-000000000012","varianteId": 25,"produktName": "Bier","varianteName": "Maß 1,0l","kategorie": "beverage","einzelpreis": 850,"menge": 2}],"gesamtPreisCents": 3500,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.produkte-geliefert:v1', 'tisch:5', 11, now() - interval '44 hours 56 minutes', '{"lieferungId": "c0005000-0000-0000-0000-000000000004","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000011","menge": 4},{"positionId": "b0005000-0000-0000-0000-000000000012","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.zahlung-registriert:v1', 'tisch:5', 12, now() - interval '44 hours 35 minutes', '{"zahlungId": "d0005000-0000-0000-0000-000000000004","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000011","menge": 4},{"positionId": "b0005000-0000-0000-0000-000000000012","menge": 2}],"gesamtZahlungCents": 3500,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.produkte-storniert:v1', 'tisch:5', 13, now() - interval '44 hours 26 minutes', '{"stornierungId": "e0005000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000002","menge": 1}],"gesamtStornierungCents": 750,"kommentar": "Flammkuchen war kalt — Geld zurück"}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.bestellung-aufgegeben:v1', 'tisch:5', 14, now() - interval '26 hours 20 minutes', '{"bestellungId": "a0005000-0000-0000-0000-000000000005","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000013","varianteId": 13,"produktName": "Grillplatte","varianteName": "Groß","kategorie": "food","einzelpreis": 1400,"menge": 3},{"positionId": "b0005000-0000-0000-0000-000000000014","varianteId": 7,"produktName": "Flammkuchen","varianteName": "Speck & Zwiebel","kategorie": "food","einzelpreis": 700,"menge": 3},{"positionId": "b0005000-0000-0000-0000-000000000015","varianteId": 10,"produktName": "Tagesgericht","varianteName": "Sa: Gulasch mit Spätzle","kategorie": "food","einzelpreis": 1150,"menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000016","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 6},{"positionId": "b0005000-0000-0000-0000-000000000017","varianteId": 42,"produktName": "Wein","varianteName": "Weißwein 0,2l","kategorie": "beverage","einzelpreis": 400,"menge": 4}],"gesamtPreisCents": 12900,"kommentar": "Samstag: VIP-Gäste"}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.produkte-geliefert:v1', 'tisch:5', 15, now() - interval '26 hours 11 minutes', '{"lieferungId": "c0005000-0000-0000-0000-000000000005","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000013","menge": 3},{"positionId": "b0005000-0000-0000-0000-000000000014","menge": 3},{"positionId": "b0005000-0000-0000-0000-000000000015","menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000016","menge": 6},{"positionId": "b0005000-0000-0000-0000-000000000017","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.zahlung-registriert:v1', 'tisch:5', 16, now() - interval '25 hours 52 minutes', '{"zahlungId": "d0005000-0000-0000-0000-000000000005","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000013","menge": 3},{"positionId": "b0005000-0000-0000-0000-000000000014","menge": 3},{"positionId": "b0005000-0000-0000-0000-000000000015","menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000016","menge": 6},{"positionId": "b0005000-0000-0000-0000-000000000017","menge": 4}],"gesamtZahlungCents": 12900,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.bestellung-aufgegeben:v1', 'tisch:5', 17, now() - interval '25 hours 34 minutes', '{"bestellungId": "a0005000-0000-0000-0000-000000000006","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000018","varianteId": 49,"produktName": "Hugo/Aperol","varianteName": "Aperol Spritz","kategorie": "beverage","einzelpreis": 550,"menge": 4},{"positionId": "b0005000-0000-0000-0000-000000000019","varianteId": 48,"produktName": "Hugo/Aperol","varianteName": "Hugo","kategorie": "beverage","einzelpreis": 550,"menge": 2}],"gesamtPreisCents": 3300,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.produkte-geliefert:v1', 'tisch:5', 18, now() - interval '25 hours 23 minutes', '{"lieferungId": "c0005000-0000-0000-0000-000000000006","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000018","menge": 4},{"positionId": "b0005000-0000-0000-0000-000000000019","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.zahlung-registriert:v1', 'tisch:5', 19, now() - interval '24 hours 59 minutes', '{"zahlungId": "d0005000-0000-0000-0000-000000000006","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000018","menge": 4},{"positionId": "b0005000-0000-0000-0000-000000000019","menge": 2}],"gesamtZahlungCents": 3300,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.bestellung-aufgegeben:v1', 'tisch:5', 20, now() - interval '24 hours 38 minutes', '{"bestellungId": "a0005000-0000-0000-0000-000000000007","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000020","varianteId": 12,"produktName": "Grillplatte","varianteName": "Klein","kategorie": "food","einzelpreis": 800,"menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000021","varianteId": 1,"produktName": "Bratwurst","varianteName": "Normal","kategorie": "food","einzelpreis": 350,"menge": 4},{"positionId": "b0005000-0000-0000-0000-000000000022","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 4},{"positionId": "b0005000-0000-0000-0000-000000000023","varianteId": 27,"produktName": "Weizen","varianteName": "Groß 0,5l","kategorie": "beverage","einzelpreis": 400,"menge": 3}],"gesamtPreisCents": 6000,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.produkte-geliefert:v1', 'tisch:5', 21, now() - interval '24 hours 28 minutes', '{"lieferungId": "c0005000-0000-0000-0000-000000000007","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000020","menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000021","menge": 4},{"positionId": "b0005000-0000-0000-0000-000000000022","menge": 4},{"positionId": "b0005000-0000-0000-0000-000000000023","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.bestellung-aufgegeben:v1', 'tisch:5', 22, now() - interval '24 hours 7 minutes', '{"bestellungId": "a0005000-0000-0000-0000-000000000008","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000024","varianteId": 14,"produktName": "Salat","varianteName": "Gemischter Salat","kategorie": "food","einzelpreis": 550,"menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000025","varianteId": 6,"produktName": "Flammkuchen","varianteName": "Classic","kategorie": "food","einzelpreis": 600,"menge": 1},{"positionId": "b0005000-0000-0000-0000-000000000026","varianteId": 36,"produktName": "Wasser","varianteName": "Still 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 3}],"gesamtPreisCents": 2300,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.produkte-geliefert:v1', 'tisch:5', 23, now() - interval '23 hours 54 minutes', '{"lieferungId": "c0005000-0000-0000-0000-000000000008","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000024","menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000025","menge": 1},{"positionId": "b0005000-0000-0000-0000-000000000026","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.zahlung-registriert:v1', 'tisch:5', 24, now() - interval '23 hours 26 minutes', '{"zahlungId": "d0005000-0000-0000-0000-000000000007","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000020","menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000021","menge": 4},{"positionId": "b0005000-0000-0000-0000-000000000022","menge": 4},{"positionId": "b0005000-0000-0000-0000-000000000023","menge": 3},{"positionId": "b0005000-0000-0000-0000-000000000024","menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000025","menge": 1},{"positionId": "b0005000-0000-0000-0000-000000000026","menge": 3}],"gesamtZahlungCents": 8300,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.bestellung-aufgegeben:v1', 'tisch:5', 25, now() - interval '23 hours 16 minutes', '{"bestellungId": "a0005000-0000-0000-0000-000000000009","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000027","varianteId": 18,"produktName": "Waffeln","varianteName": "mit Sahne","kategorie": "food","einzelpreis": 350,"menge": 3},{"positionId": "b0005000-0000-0000-0000-000000000028","varianteId": 19,"produktName": "Waffeln","varianteName": "mit Nutella","kategorie": "food","einzelpreis": 400,"menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000029","varianteId": 45,"produktName": "Kaffee","varianteName": "Tasse","kategorie": "beverage","einzelpreis": 200,"menge": 4}],"gesamtPreisCents": 2650,"kommentar": "Nachtisch VIP"}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.produkte-geliefert:v1', 'tisch:5', 26, now() - interval '23 hours 8 minutes', '{"lieferungId": "c0005000-0000-0000-0000-000000000009","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000027","menge": 3},{"positionId": "b0005000-0000-0000-0000-000000000028","menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000029","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.zahlung-registriert:v1', 'tisch:5', 27, now() - interval '22 hours 47 minutes', '{"zahlungId": "d0005000-0000-0000-0000-000000000008","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000027","menge": 3},{"positionId": "b0005000-0000-0000-0000-000000000028","menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000029","menge": 4}],"gesamtZahlungCents": 2650,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.bestellung-aufgegeben:v1', 'tisch:5', 28, now() - interval '22 hours 26 minutes', '{"bestellungId": "a0005000-0000-0000-0000-000000000010","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000030","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 4},{"positionId": "b0005000-0000-0000-0000-000000000031","varianteId": 26,"produktName": "Weizen","varianteName": "Klein 0,3l","kategorie": "beverage","einzelpreis": 300,"menge": 2}],"gesamtPreisCents": 2400,"kommentar": "Letzte Runde Samstag"}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.produkte-geliefert:v1', 'tisch:5', 29, now() - interval '22 hours 17 minutes', '{"lieferungId": "c0005000-0000-0000-0000-000000000010","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000030","menge": 4},{"positionId": "b0005000-0000-0000-0000-000000000031","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.zahlung-registriert:v1', 'tisch:5', 30, now() - interval '22 hours 4 minutes', '{"zahlungId": "d0005000-0000-0000-0000-000000000009","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000030","menge": 4},{"positionId": "b0005000-0000-0000-0000-000000000031","menge": 2}],"gesamtZahlungCents": 2400,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:5', 31, now() - interval '18 hours 19 minutes', '{"bestellungId": "a0005000-0000-0000-0000-000000000013","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000038","varianteId": 7,"produktName": "Flammkuchen","varianteName": "Speck & Zwiebel","kategorie": "food","einzelpreis": 700,"menge": 1},{"positionId": "b0005000-0000-0000-0000-000000000039","varianteId": 16,"produktName": "Kuchen","varianteName": "Stück","kategorie": "food","einzelpreis": 250,"menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000040","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 4},{"positionId": "b0005000-0000-0000-0000-000000000041","varianteId": 38,"produktName": "Wasser","varianteName": "Sprudel 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 4}],"gesamtPreisCents": 3800,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.bestellung-aufgegeben:v1', 'tisch:5', 32, now() - interval '18 hours 16 minutes', '{"bestellungId": "a0005000-0000-0000-0000-000000000019","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000055","varianteId": 17,"produktName": "Waffeln","varianteName": "mit Puderzucker","kategorie": "food","einzelpreis": 300,"menge": 1},{"positionId": "b0005000-0000-0000-0000-000000000056","varianteId": 13,"produktName": "Grillplatte","varianteName": "Groß","kategorie": "food","einzelpreis": 1400,"menge": 1},{"positionId": "b0005000-0000-0000-0000-000000000057","varianteId": 39,"produktName": "Saftschorle","varianteName": "Apfelschorle 0,5l","kategorie": "beverage","einzelpreis": 300,"menge": 2}],"gesamtPreisCents": 2300,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:5', 33, now() - interval '18 hours 15 minutes', '{"bestellungId": "a0005000-0000-0000-0000-000000000021","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000061","varianteId": 4,"produktName": "Pommes","varianteName": "Klein","kategorie": "food","einzelpreis": 250,"menge": 3},{"positionId": "b0005000-0000-0000-0000-000000000062","varianteId": 19,"produktName": "Waffeln","varianteName": "mit Nutella","kategorie": "food","einzelpreis": 400,"menge": 3},{"positionId": "b0005000-0000-0000-0000-000000000063","varianteId": 33,"produktName": "Softdrinks","varianteName": "Spezi","kategorie": "beverage","einzelpreis": 280,"menge": 3},{"positionId": "b0005000-0000-0000-0000-000000000064","varianteId": 45,"produktName": "Kaffee","varianteName": "Tasse","kategorie": "beverage","einzelpreis": 200,"menge": 3}],"gesamtPreisCents": 3390,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.bestellung-aufgegeben:v1', 'tisch:5', 34, now() - interval '18 hours 12 minutes', '{"bestellungId": "a0005000-0000-0000-0000-000000000023","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000068","varianteId": 10,"produktName": "Tagesgericht","varianteName": "Sa: Gulasch mit Spätzle","kategorie": "food","einzelpreis": 1150,"menge": 3},{"positionId": "b0005000-0000-0000-0000-000000000069","varianteId": 43,"produktName": "Wein","varianteName": "Rotwein 0,2l","kategorie": "beverage","einzelpreis": 400,"menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000070","varianteId": 44,"produktName": "Wein","varianteName": "Rosé 0,2l","kategorie": "beverage","einzelpreis": 400,"menge": 2}],"gesamtPreisCents": 5050,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.bestellung-aufgegeben:v1', 'tisch:5', 35, now() - interval '18 hours 8 minutes', '{"bestellungId": "a0005000-0000-0000-0000-000000000022","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000065","varianteId": 3,"produktName": "Bratwurst","varianteName": "Currywurst","kategorie": "food","einzelpreis": 450,"menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000066","varianteId": 8,"produktName": "Flammkuchen","varianteName": "Mediterran","kategorie": "food","einzelpreis": 750,"menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000067","varianteId": 38,"produktName": "Wasser","varianteName": "Sprudel 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 1}],"gesamtPreisCents": 2600,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.produkte-geliefert:v1', 'tisch:5', 36, now() - interval '18 hours 7 minutes', '{"lieferungId": "c0005000-0000-0000-0000-000000000019","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000055","menge": 1},{"positionId": "b0005000-0000-0000-0000-000000000056","menge": 1},{"positionId": "b0005000-0000-0000-0000-000000000057","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.bestellung-aufgegeben:v1', 'tisch:5', 37, now() - interval '18 hours 7 minutes', '{"bestellungId": "a0005000-0000-0000-0000-000000000032","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000093","varianteId": 13,"produktName": "Grillplatte","varianteName": "Groß","kategorie": "food","einzelpreis": 1400,"menge": 3},{"positionId": "b0005000-0000-0000-0000-000000000094","varianteId": 40,"produktName": "Saftschorle","varianteName": "Johannisbeerschorle 0,5l","kategorie": "beverage","einzelpreis": 300,"menge": 4},{"positionId": "b0005000-0000-0000-0000-000000000095","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 3}],"gesamtPreisCents": 6750,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:5', 38, now() - interval '18 hours 6 minutes', '{"lieferungId": "c0005000-0000-0000-0000-000000000021","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000061","menge": 3},{"positionId": "b0005000-0000-0000-0000-000000000062","menge": 3},{"positionId": "b0005000-0000-0000-0000-000000000063","menge": 3},{"positionId": "b0005000-0000-0000-0000-000000000064","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:5', 39, now() - interval '18 hours 5 minutes', '{"lieferungId": "c0005000-0000-0000-0000-000000000013","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000038","menge": 1},{"positionId": "b0005000-0000-0000-0000-000000000039","menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000040","menge": 4},{"positionId": "b0005000-0000-0000-0000-000000000041","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.produkte-geliefert:v1', 'tisch:5', 40, now() - interval '18 hours 3 minutes', '{"lieferungId": "c0005000-0000-0000-0000-000000000023","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000068","menge": 3},{"positionId": "b0005000-0000-0000-0000-000000000069","menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000070","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.produkte-geliefert:v1', 'tisch:5', 41, now() - interval '17 hours 59 minutes', '{"lieferungId": "c0005000-0000-0000-0000-000000000022","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000065","menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000066","menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000067","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.bestellung-aufgegeben:v1', 'tisch:5', 42, now() - interval '17 hours 58 minutes', '{"bestellungId": "a0005000-0000-0000-0000-000000000030","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000088","varianteId": 5,"produktName": "Pommes","varianteName": "Groß","kategorie": "food","einzelpreis": 350,"menge": 1},{"positionId": "b0005000-0000-0000-0000-000000000089","varianteId": 25,"produktName": "Bier","varianteName": "Maß 1,0l","kategorie": "beverage","einzelpreis": 850,"menge": 1},{"positionId": "b0005000-0000-0000-0000-000000000090","varianteId": 30,"produktName": "Weizen","varianteName": "Russ","kategorie": "beverage","einzelpreis": 300,"menge": 1}],"gesamtPreisCents": 1500,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.produkte-geliefert:v1', 'tisch:5', 43, now() - interval '17 hours 58 minutes', '{"lieferungId": "c0005000-0000-0000-0000-000000000032","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000093","menge": 3},{"positionId": "b0005000-0000-0000-0000-000000000094","menge": 4},{"positionId": "b0005000-0000-0000-0000-000000000095","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.zahlung-registriert:v1', 'tisch:5', 44, now() - interval '17 hours 54 minutes', '{"zahlungId": "d0005000-0000-0000-0000-000000000017","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000055","menge": 1},{"positionId": "b0005000-0000-0000-0000-000000000056","menge": 1},{"positionId": "b0005000-0000-0000-0000-000000000057","menge": 2}],"gesamtZahlungCents": 2300,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:5', 45, now() - interval '17 hours 53 minutes', '{"zahlungId": "d0005000-0000-0000-0000-000000000011","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000038","menge": 1},{"positionId": "b0005000-0000-0000-0000-000000000039","menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000040","menge": 4},{"positionId": "b0005000-0000-0000-0000-000000000041","menge": 4}],"gesamtZahlungCents": 3800,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:5', 46, now() - interval '17 hours 51 minutes', '{"zahlungId": "d0005000-0000-0000-0000-000000000019","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000061","menge": 3},{"positionId": "b0005000-0000-0000-0000-000000000062","menge": 3},{"positionId": "b0005000-0000-0000-0000-000000000063","menge": 3},{"positionId": "b0005000-0000-0000-0000-000000000064","menge": 3}],"gesamtZahlungCents": 3390,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.produkte-geliefert:v1', 'tisch:5', 47, now() - interval '17 hours 43 minutes', '{"lieferungId": "c0005000-0000-0000-0000-000000000030","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000088","menge": 1},{"positionId": "b0005000-0000-0000-0000-000000000089","menge": 1},{"positionId": "b0005000-0000-0000-0000-000000000090","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.zahlung-registriert:v1', 'tisch:5', 48, now() - interval '17 hours 42 minutes', '{"zahlungId": "d0005000-0000-0000-0000-000000000030","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000093","menge": 3},{"positionId": "b0005000-0000-0000-0000-000000000094","menge": 4},{"positionId": "b0005000-0000-0000-0000-000000000095","menge": 3}],"gesamtZahlungCents": 6750,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.zahlung-registriert:v1', 'tisch:5', 49, now() - interval '17 hours 40 minutes', '{"zahlungId": "d0005000-0000-0000-0000-000000000021","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000068","menge": 3},{"positionId": "b0005000-0000-0000-0000-000000000069","menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000070","menge": 2}],"gesamtZahlungCents": 5050,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.bestellung-aufgegeben:v1', 'tisch:5', 50, now() - interval '17 hours 38 minutes', '{"bestellungId": "a0005000-0000-0000-0000-000000000026","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000077","varianteId": 17,"produktName": "Waffeln","varianteName": "mit Puderzucker","kategorie": "food","einzelpreis": 300,"menge": 3},{"positionId": "b0005000-0000-0000-0000-000000000078","varianteId": 49,"produktName": "Hugo/Aperol","varianteName": "Aperol Spritz","kategorie": "beverage","einzelpreis": 550,"menge": 2}],"gesamtPreisCents": 2000,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.zahlung-registriert:v1', 'tisch:5', 51, now() - interval '17 hours 36 minutes', '{"zahlungId": "d0005000-0000-0000-0000-000000000020","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000065","menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000066","menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000067","menge": 1}],"gesamtZahlungCents": 2600,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.produkte-geliefert:v1', 'tisch:5', 52, now() - interval '17 hours 27 minutes', '{"lieferungId": "c0005000-0000-0000-0000-000000000026","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000077","menge": 3},{"positionId": "b0005000-0000-0000-0000-000000000078","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.bestellung-aufgegeben:v1', 'tisch:5', 53, now() - interval '17 hours 26 minutes', '{"bestellungId": "a0005000-0000-0000-0000-000000000027","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000079","varianteId": 10,"produktName": "Tagesgericht","varianteName": "Sa: Gulasch mit Spätzle","kategorie": "food","einzelpreis": 1150,"menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000080","varianteId": 45,"produktName": "Kaffee","varianteName": "Tasse","kategorie": "beverage","einzelpreis": 200,"menge": 1}],"gesamtPreisCents": 2500,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.zahlung-registriert:v1', 'tisch:5', 54, now() - interval '17 hours 26 minutes', '{"zahlungId": "d0005000-0000-0000-0000-000000000028","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000088","menge": 1},{"positionId": "b0005000-0000-0000-0000-000000000089","menge": 1},{"positionId": "b0005000-0000-0000-0000-000000000090","menge": 1}],"gesamtZahlungCents": 1500,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:5', 55, now() - interval '17 hours 21 minutes', '{"bestellungId": "a0005000-0000-0000-0000-000000000031","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000091","varianteId": 4,"produktName": "Pommes","varianteName": "Klein","kategorie": "food","einzelpreis": 250,"menge": 1},{"positionId": "b0005000-0000-0000-0000-000000000092","varianteId": 25,"produktName": "Bier","varianteName": "Maß 1,0l","kategorie": "beverage","einzelpreis": 850,"menge": 2}],"gesamtPreisCents": 1950,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.bestellung-aufgegeben:v1', 'tisch:5', 56, now() - interval '17 hours 19 minutes', '{"bestellungId": "a0005000-0000-0000-0000-000000000015","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000045","varianteId": 20,"produktName": "Brezel","varianteName": "Normal","kategorie": "food","einzelpreis": 200,"menge": 3},{"positionId": "b0005000-0000-0000-0000-000000000046","varianteId": 18,"produktName": "Waffeln","varianteName": "mit Sahne","kategorie": "food","einzelpreis": 350,"menge": 1},{"positionId": "b0005000-0000-0000-0000-000000000047","varianteId": 33,"produktName": "Softdrinks","varianteName": "Spezi","kategorie": "beverage","einzelpreis": 280,"menge": 3}],"gesamtPreisCents": 1790,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.bestellung-aufgegeben:v1', 'tisch:5', 57, now() - interval '17 hours 17 minutes', '{"bestellungId": "a0005000-0000-0000-0000-000000000017","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000050","varianteId": 16,"produktName": "Kuchen","varianteName": "Stück","kategorie": "food","einzelpreis": 250,"menge": 3},{"positionId": "b0005000-0000-0000-0000-000000000051","varianteId": 26,"produktName": "Weizen","varianteName": "Klein 0,3l","kategorie": "beverage","einzelpreis": 300,"menge": 4},{"positionId": "b0005000-0000-0000-0000-000000000052","varianteId": 28,"produktName": "Weizen","varianteName": "Colaweizen Klein","kategorie": "beverage","einzelpreis": 300,"menge": 4}],"gesamtPreisCents": 3150,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:5', 58, now() - interval '17 hours 17 minutes', '{"bestellungId": "a0005000-0000-0000-0000-000000000024","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000071","varianteId": 14,"produktName": "Salat","varianteName": "Gemischter Salat","kategorie": "food","einzelpreis": 550,"menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000072","varianteId": 13,"produktName": "Grillplatte","varianteName": "Groß","kategorie": "food","einzelpreis": 1400,"menge": 1},{"positionId": "b0005000-0000-0000-0000-000000000073","varianteId": 43,"produktName": "Wein","varianteName": "Rotwein 0,2l","kategorie": "beverage","einzelpreis": 400,"menge": 2}],"gesamtPreisCents": 3300,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.bestellung-aufgegeben:v1', 'tisch:5', 59, now() - interval '17 hours 16 minutes', '{"bestellungId": "a0005000-0000-0000-0000-000000000014","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000042","varianteId": 12,"produktName": "Grillplatte","varianteName": "Klein","kategorie": "food","einzelpreis": 800,"menge": 3},{"positionId": "b0005000-0000-0000-0000-000000000043","varianteId": 5,"produktName": "Pommes","varianteName": "Groß","kategorie": "food","einzelpreis": 350,"menge": 3},{"positionId": "b0005000-0000-0000-0000-000000000044","varianteId": 35,"produktName": "Softdrinks","varianteName": "Mezzo Mix","kategorie": "beverage","einzelpreis": 280,"menge": 1}],"gesamtPreisCents": 3730,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.produkte-geliefert:v1', 'tisch:5', 60, now() - interval '17 hours 15 minutes', '{"lieferungId": "c0005000-0000-0000-0000-000000000027","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000079","menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000080","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:5', 61, now() - interval '17 hours 14 minutes', '{"bestellungId": "a0005000-0000-0000-0000-000000000029","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000084","varianteId": 10,"produktName": "Tagesgericht","varianteName": "Sa: Gulasch mit Spätzle","kategorie": "food","einzelpreis": 1150,"menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000085","varianteId": 1,"produktName": "Bratwurst","varianteName": "Normal","kategorie": "food","einzelpreis": 350,"menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000086","varianteId": 28,"produktName": "Weizen","varianteName": "Colaweizen Klein","kategorie": "beverage","einzelpreis": 300,"menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000087","varianteId": 35,"produktName": "Softdrinks","varianteName": "Mezzo Mix","kategorie": "beverage","einzelpreis": 280,"menge": 1}],"gesamtPreisCents": 3880,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.produkte-geliefert:v1', 'tisch:5', 62, now() - interval '17 hours 11 minutes', '{"lieferungId": "c0005000-0000-0000-0000-000000000015","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000045","menge": 3},{"positionId": "b0005000-0000-0000-0000-000000000046","menge": 1},{"positionId": "b0005000-0000-0000-0000-000000000047","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:5', 63, now() - interval '17 hours 10 minutes', '{"bestellungId": "a0005000-0000-0000-0000-000000000020","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000058","varianteId": 16,"produktName": "Kuchen","varianteName": "Stück","kategorie": "food","einzelpreis": 250,"menge": 1},{"positionId": "b0005000-0000-0000-0000-000000000059","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 3},{"positionId": "b0005000-0000-0000-0000-000000000060","varianteId": 35,"produktName": "Softdrinks","varianteName": "Mezzo Mix","kategorie": "beverage","einzelpreis": 280,"menge": 2}],"gesamtPreisCents": 2160,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.zahlung-registriert:v1', 'tisch:5', 64, now() - interval '17 hours 8 minutes', '{"zahlungId": "d0005000-0000-0000-0000-000000000024","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000077","menge": 3},{"positionId": "b0005000-0000-0000-0000-000000000078","menge": 2}],"gesamtZahlungCents": 2000,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:5', 65, now() - interval '17 hours 7 minutes', '{"lieferungId": "c0005000-0000-0000-0000-000000000031","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000091","menge": 1},{"positionId": "b0005000-0000-0000-0000-000000000092","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:5', 66, now() - interval '17 hours 6 minutes', '{"lieferungId": "c0005000-0000-0000-0000-000000000024","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000071","menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000072","menge": 1},{"positionId": "b0005000-0000-0000-0000-000000000073","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.bestellung-aufgegeben:v1', 'tisch:5', 67, now() - interval '17 hours 6 minutes', '{"bestellungId": "a0005000-0000-0000-0000-000000000028","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000081","varianteId": 5,"produktName": "Pommes","varianteName": "Groß","kategorie": "food","einzelpreis": 350,"menge": 3},{"positionId": "b0005000-0000-0000-0000-000000000082","varianteId": 42,"produktName": "Wein","varianteName": "Weißwein 0,2l","kategorie": "beverage","einzelpreis": 400,"menge": 1},{"positionId": "b0005000-0000-0000-0000-000000000083","varianteId": 31,"produktName": "Softdrinks","varianteName": "Cola","kategorie": "beverage","einzelpreis": 280,"menge": 4}],"gesamtPreisCents": 2570,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:5', 68, now() - interval '17 hours 5 minutes', '{"lieferungId": "c0005000-0000-0000-0000-000000000029","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000084","menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000085","menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000086","menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000087","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.produkte-geliefert:v1', 'tisch:5', 69, now() - interval '17 hours 2 minutes', '{"lieferungId": "c0005000-0000-0000-0000-000000000017","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000050","menge": 3},{"positionId": "b0005000-0000-0000-0000-000000000051","menge": 4},{"positionId": "b0005000-0000-0000-0000-000000000052","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.produkte-geliefert:v1', 'tisch:5', 70, now() - interval '17 hours 1 minutes', '{"lieferungId": "c0005000-0000-0000-0000-000000000014","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000042","menge": 3},{"positionId": "b0005000-0000-0000-0000-000000000043","menge": 3},{"positionId": "b0005000-0000-0000-0000-000000000044","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.zahlung-registriert:v1', 'tisch:5', 71, now() - interval '16 hours 56 minutes', '{"zahlungId": "d0005000-0000-0000-0000-000000000013","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000045","menge": 3},{"positionId": "b0005000-0000-0000-0000-000000000046","menge": 1},{"positionId": "b0005000-0000-0000-0000-000000000047","menge": 3}],"gesamtZahlungCents": 1790,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:5', 72, now() - interval '16 hours 55 minutes', '{"lieferungId": "c0005000-0000-0000-0000-000000000020","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000058","menge": 1},{"positionId": "b0005000-0000-0000-0000-000000000059","menge": 3},{"positionId": "b0005000-0000-0000-0000-000000000060","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:5', 73, now() - interval '16 hours 54 minutes', '{"bestellungId": "a0005000-0000-0000-0000-000000000016","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000048","varianteId": 21,"produktName": "Brezel","varianteName": "mit Butter","kategorie": "food","einzelpreis": 300,"menge": 1},{"positionId": "b0005000-0000-0000-0000-000000000049","varianteId": 36,"produktName": "Wasser","varianteName": "Still 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 3}],"gesamtPreisCents": 900,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.bestellung-aufgegeben:v1', 'tisch:5', 74, now() - interval '16 hours 53 minutes', '{"bestellungId": "a0005000-0000-0000-0000-000000000018","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000053","varianteId": 14,"produktName": "Salat","varianteName": "Gemischter Salat","kategorie": "food","einzelpreis": 550,"menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000054","varianteId": 42,"produktName": "Wein","varianteName": "Weißwein 0,2l","kategorie": "beverage","einzelpreis": 400,"menge": 4}],"gesamtPreisCents": 2700,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:5', 75, now() - interval '16 hours 53 minutes', '{"zahlungId": "d0005000-0000-0000-0000-000000000022","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000071","menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000072","menge": 1},{"positionId": "b0005000-0000-0000-0000-000000000073","menge": 2}],"gesamtZahlungCents": 3300,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:5', 76, now() - interval '16 hours 53 minutes', '{"zahlungId": "d0005000-0000-0000-0000-000000000029","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000091","menge": 1},{"positionId": "b0005000-0000-0000-0000-000000000092","menge": 2}],"gesamtZahlungCents": 1950,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.zahlung-registriert:v1', 'tisch:5', 77, now() - interval '16 hours 52 minutes', '{"zahlungId": "d0005000-0000-0000-0000-000000000025","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000079","menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000080","menge": 1}],"gesamtZahlungCents": 2500,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.produkte-geliefert:v1', 'tisch:5', 78, now() - interval '16 hours 51 minutes', '{"lieferungId": "c0005000-0000-0000-0000-000000000028","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000081","menge": 3},{"positionId": "b0005000-0000-0000-0000-000000000082","menge": 1},{"positionId": "b0005000-0000-0000-0000-000000000083","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:5', 79, now() - interval '16 hours 44 minutes', '{"bestellungId": "a0005000-0000-0000-0000-000000000025","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000074","varianteId": 13,"produktName": "Grillplatte","varianteName": "Groß","kategorie": "food","einzelpreis": 1400,"menge": 1},{"positionId": "b0005000-0000-0000-0000-000000000075","varianteId": 8,"produktName": "Flammkuchen","varianteName": "Mediterran","kategorie": "food","einzelpreis": 750,"menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000076","varianteId": 48,"produktName": "Hugo/Aperol","varianteName": "Hugo","kategorie": "beverage","einzelpreis": 550,"menge": 2}],"gesamtPreisCents": 4000,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.produkte-geliefert:v1', 'tisch:5', 80, now() - interval '16 hours 43 minutes', '{"lieferungId": "c0005000-0000-0000-0000-000000000016","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000048","menge": 1},{"positionId": "b0005000-0000-0000-0000-000000000049","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.zahlung-registriert:v1', 'tisch:5', 81, now() - interval '16 hours 43 minutes', '{"zahlungId": "d0005000-0000-0000-0000-000000000015","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000050","menge": 3},{"positionId": "b0005000-0000-0000-0000-000000000051","menge": 4},{"positionId": "b0005000-0000-0000-0000-000000000052","menge": 4}],"gesamtZahlungCents": 3150,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.produkte-geliefert:v1', 'tisch:5', 82, now() - interval '16 hours 43 minutes', '{"lieferungId": "c0005000-0000-0000-0000-000000000018","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000053","menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000054","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:5', 83, now() - interval '16 hours 43 minutes', '{"zahlungId": "d0005000-0000-0000-0000-000000000027","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000084","menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000085","menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000086","menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000087","menge": 1}],"gesamtZahlungCents": 3880,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.zahlung-registriert:v1', 'tisch:5', 84, now() - interval '16 hours 36 minutes', '{"zahlungId": "d0005000-0000-0000-0000-000000000012","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000042","menge": 3},{"positionId": "b0005000-0000-0000-0000-000000000043","menge": 3},{"positionId": "b0005000-0000-0000-0000-000000000044","menge": 1}],"gesamtZahlungCents": 3730,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:5', 85, now() - interval '16 hours 31 minutes', '{"zahlungId": "d0005000-0000-0000-0000-000000000018","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000058","menge": 1},{"positionId": "b0005000-0000-0000-0000-000000000059","menge": 3},{"positionId": "b0005000-0000-0000-0000-000000000060","menge": 2}],"gesamtZahlungCents": 2160,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.produkte-geliefert:v1', 'tisch:5', 86, now() - interval '16 hours 31 minutes', '{"lieferungId": "c0005000-0000-0000-0000-000000000025","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000074","menge": 1},{"positionId": "b0005000-0000-0000-0000-000000000075","menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000076","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.zahlung-registriert:v1', 'tisch:5', 87, now() - interval '16 hours 29 minutes', '{"zahlungId": "d0005000-0000-0000-0000-000000000016","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000053","menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000054","menge": 4}],"gesamtZahlungCents": 2700,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.zahlung-registriert:v1', 'tisch:5', 88, now() - interval '16 hours 27 minutes', '{"zahlungId": "d0005000-0000-0000-0000-000000000026","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000081","menge": 3},{"positionId": "b0005000-0000-0000-0000-000000000082","menge": 1},{"positionId": "b0005000-0000-0000-0000-000000000083","menge": 4}],"gesamtZahlungCents": 2570,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.zahlung-registriert:v1', 'tisch:5', 89, now() - interval '16 hours 26 minutes', '{"zahlungId": "d0005000-0000-0000-0000-000000000014","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000048","menge": 1},{"positionId": "b0005000-0000-0000-0000-000000000049","menge": 3}],"gesamtZahlungCents": 900,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.zahlung-registriert:v1', 'tisch:5', 90, now() - interval '16 hours 12 minutes', '{"zahlungId": "d0005000-0000-0000-0000-000000000023","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000074","menge": 1},{"positionId": "b0005000-0000-0000-0000-000000000075","menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000076","menge": 2}],"gesamtZahlungCents": 4000,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.bestellung-aufgegeben:v1', 'tisch:5', 91, now() - interval '4 hours 30 minutes', '{"bestellungId": "a0005000-0000-0000-0000-000000000011","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000032","varianteId": 11,"produktName": "Tagesgericht","varianteName": "So: Hähnchen mit Reis","kategorie": "food","einzelpreis": 1050,"menge": 4},{"positionId": "b0005000-0000-0000-0000-000000000033","varianteId": 15,"produktName": "Salat","varianteName": "Caesar Salat","kategorie": "food","einzelpreis": 650,"menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000034","varianteId": 37,"produktName": "Wasser","varianteName": "Medium 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 4},{"positionId": "b0005000-0000-0000-0000-000000000035","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 2}],"gesamtPreisCents": 7200,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.produkte-geliefert:v1', 'tisch:5', 92, now() - interval '4 hours 22 minutes', '{"lieferungId": "c0005000-0000-0000-0000-000000000011","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000032","menge": 4},{"positionId": "b0005000-0000-0000-0000-000000000033","menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000034","menge": 4},{"positionId": "b0005000-0000-0000-0000-000000000035","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.bestellung-aufgegeben:v1', 'tisch:5', 93, now() - interval '4 hours 2 minutes', '{"bestellungId": "a0005000-0000-0000-0000-000000000012","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000036","varianteId": 45,"produktName": "Kaffee","varianteName": "Tasse","kategorie": "beverage","einzelpreis": 200,"menge": 4},{"positionId": "b0005000-0000-0000-0000-000000000037","varianteId": 16,"produktName": "Kuchen","varianteName": "Stück","kategorie": "food","einzelpreis": 250,"menge": 3}],"gesamtPreisCents": 1550,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.produkte-geliefert:v1', 'tisch:5', 94, now() - interval '3 hours 51 minutes', '{"lieferungId": "c0005000-0000-0000-0000-000000000012","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000036","menge": 4},{"positionId": "b0005000-0000-0000-0000-000000000037","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.zahlung-registriert:v1', 'tisch:5', 95, now() - interval '3 hours 32 minutes', '{"zahlungId": "d0005000-0000-0000-0000-000000000010","positionen": [{"positionId": "b0005000-0000-0000-0000-000000000032","menge": 4},{"positionId": "b0005000-0000-0000-0000-000000000033","menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000034","menge": 4},{"positionId": "b0005000-0000-0000-0000-000000000035","menge": 2},{"positionId": "b0005000-0000-0000-0000-000000000036","menge": 4},{"positionId": "b0005000-0000-0000-0000-000000000037","menge": 3}],"gesamtZahlungCents": 8750,"kommentar": ""}');

-- ─────────────────────────────────────────────────────────────────────────────
-- TISCH 6: Frische Gäste — Tag 3, gerade angekommen (9 Events)
-- ─────────────────────────────────────────────────────────────────────────────

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.bestellung-aufgegeben:v1', 'tisch:6', 1, now() - interval '3 hours', '{"bestellungId": "a0006000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0006000-0000-0000-0000-000000000001","varianteId": 11,"produktName": "Tagesgericht","varianteName": "So: Hähnchen mit Reis","kategorie": "food","einzelpreis": 1050,"menge": 3},{"positionId": "b0006000-0000-0000-0000-000000000002","varianteId": 5,"produktName": "Pommes","varianteName": "Groß","kategorie": "food","einzelpreis": 350,"menge": 3},{"positionId": "b0006000-0000-0000-0000-000000000003","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 4},{"positionId": "b0006000-0000-0000-0000-000000000004","varianteId": 33,"produktName": "Softdrinks","varianteName": "Spezi","kategorie": "beverage","einzelpreis": 280,"menge": 2}],"gesamtPreisCents": 6560,"kommentar": "Familie mit 3 Kindern"}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.produkte-geliefert:v1', 'tisch:6', 2, now() - interval '2 hours 50 minutes', '{"lieferungId": "c0006000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0006000-0000-0000-0000-000000000003","menge": 4},{"positionId": "b0006000-0000-0000-0000-000000000004","menge": 2}],"kommentar": "Getränke zuerst"}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.produkte-geliefert:v1', 'tisch:6', 3, now() - interval '2 hours 38 minutes', '{"lieferungId": "c0006000-0000-0000-0000-000000000002","positionen": [{"positionId": "b0006000-0000-0000-0000-000000000001","menge": 3},{"positionId": "b0006000-0000-0000-0000-000000000002","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.produkte-storniert:v1', 'tisch:6', 4, now() - interval '2 hours 18 minutes', '{"stornierungId": "e0006000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0006000-0000-0000-0000-000000000001","menge": 1}],"gesamtStornierungCents": 1050,"kommentar": "Kind mag kein Hähnchen"}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.bestellung-aufgegeben:v1', 'tisch:6', 5, now() - interval '2 hours 13 minutes', '{"bestellungId": "a0006000-0000-0000-0000-000000000002","positionen": [{"positionId": "b0006000-0000-0000-0000-000000000005","varianteId": 33,"produktName": "Softdrinks","varianteName": "Spezi","kategorie": "beverage","einzelpreis": 280,"menge": 3},{"positionId": "b0006000-0000-0000-0000-000000000006","varianteId": 37,"produktName": "Wasser","varianteName": "Medium 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 2},{"positionId": "b0006000-0000-0000-0000-000000000007","varianteId": 17,"produktName": "Waffeln","varianteName": "mit Puderzucker","kategorie": "food","einzelpreis": 300,"menge": 2},{"positionId": "b0006000-0000-0000-0000-000000000008","varianteId": 19,"produktName": "Waffeln","varianteName": "mit Nutella","kategorie": "food","einzelpreis": 400,"menge": 1}],"gesamtPreisCents": 2240,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.produkte-geliefert:v1', 'tisch:6', 6, now() - interval '2 hours 3 minutes', '{"lieferungId": "c0006000-0000-0000-0000-000000000003","positionen": [{"positionId": "b0006000-0000-0000-0000-000000000005","menge": 3},{"positionId": "b0006000-0000-0000-0000-000000000006","menge": 2},{"positionId": "b0006000-0000-0000-0000-000000000007","menge": 2},{"positionId": "b0006000-0000-0000-0000-000000000008","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.zahlung-registriert:v1', 'tisch:6', 7, now() - interval '1 hours 52 minutes', '{"zahlungId": "d0006000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0006000-0000-0000-0000-000000000001","menge": 2},{"positionId": "b0006000-0000-0000-0000-000000000002","menge": 3},{"positionId": "b0006000-0000-0000-0000-000000000003","menge": 4},{"positionId": "b0006000-0000-0000-0000-000000000004","menge": 2}],"gesamtZahlungCents": 5510,"kommentar": "Erste Runde bezahlt"}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.bestellung-aufgegeben:v1', 'tisch:6', 8, now() - interval '1 hours 42 minutes', '{"bestellungId": "a0006000-0000-0000-0000-000000000003","positionen": [{"positionId": "b0006000-0000-0000-0000-000000000009","varianteId": 16,"produktName": "Kuchen","varianteName": "Stück","kategorie": "food","einzelpreis": 250,"menge": 3},{"positionId": "b0006000-0000-0000-0000-000000000010","varianteId": 45,"produktName": "Kaffee","varianteName": "Tasse","kategorie": "beverage","einzelpreis": 200,"menge": 2}],"gesamtPreisCents": 1150,"kommentar": "Nachtisch kommt noch"}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.produkte-geliefert:v1', 'tisch:6', 9, now() - interval '1 hours 37 minutes', '{"lieferungId": "c0006000-0000-0000-0000-000000000004","positionen": [{"positionId": "b0006000-0000-0000-0000-000000000009","menge": 3},{"positionId": "b0006000-0000-0000-0000-000000000010","menge": 2}],"kommentar": ""}');

-- ─────────────────────────────────────────────────────────────────────────────
-- TISCH 7: Kollegenrunde — Tag 2+3, teilweise bezahlt (23 Events)
-- ─────────────────────────────────────────────────────────────────────────────

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.bestellung-aufgegeben:v1', 'tisch:7', 1, now() - interval '22 hours 30 minutes', '{"bestellungId": "a0007000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0007000-0000-0000-0000-000000000001","varianteId": 7,"produktName": "Flammkuchen","varianteName": "Speck & Zwiebel","kategorie": "food","einzelpreis": 700,"menge": 3},{"positionId": "b0007000-0000-0000-0000-000000000002","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 6},{"positionId": "b0007000-0000-0000-0000-000000000003","varianteId": 26,"produktName": "Weizen","varianteName": "Klein 0,3l","kategorie": "beverage","einzelpreis": 300,"menge": 2}],"gesamtPreisCents": 5400,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.produkte-geliefert:v1', 'tisch:7', 2, now() - interval '22 hours 15 minutes', '{"lieferungId": "c0007000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0007000-0000-0000-0000-000000000001","menge": 3},{"positionId": "b0007000-0000-0000-0000-000000000002","menge": 6},{"positionId": "b0007000-0000-0000-0000-000000000003","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:7', 3, now() - interval '21 hours 55 minutes', '{"bestellungId": "a0007000-0000-0000-0000-000000000002","positionen": [{"positionId": "b0007000-0000-0000-0000-000000000004","varianteId": 1,"produktName": "Bratwurst","varianteName": "Normal","kategorie": "food","einzelpreis": 350,"menge": 3},{"positionId": "b0007000-0000-0000-0000-000000000005","varianteId": 5,"produktName": "Pommes","varianteName": "Groß","kategorie": "food","einzelpreis": 350,"menge": 3},{"positionId": "b0007000-0000-0000-0000-000000000006","varianteId": 31,"produktName": "Softdrinks","varianteName": "Cola","kategorie": "beverage","einzelpreis": 280,"menge": 3},{"positionId": "b0007000-0000-0000-0000-000000000007","varianteId": 33,"produktName": "Softdrinks","varianteName": "Spezi","kategorie": "beverage","einzelpreis": 280,"menge": 2}],"gesamtPreisCents": 3500,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.produkte-geliefert:v1', 'tisch:7', 4, now() - interval '21 hours 45 minutes', '{"lieferungId": "c0007000-0000-0000-0000-000000000002","positionen": [{"positionId": "b0007000-0000-0000-0000-000000000004","menge": 3},{"positionId": "b0007000-0000-0000-0000-000000000005","menge": 3},{"positionId": "b0007000-0000-0000-0000-000000000006","menge": 3},{"positionId": "b0007000-0000-0000-0000-000000000007","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.zahlung-registriert:v1', 'tisch:7', 5, now() - interval '21 hours 36 minutes', '{"zahlungId": "d0007000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0007000-0000-0000-0000-000000000002","menge": 6},{"positionId": "b0007000-0000-0000-0000-000000000003","menge": 2},{"positionId": "b0007000-0000-0000-0000-000000000006","menge": 3},{"positionId": "b0007000-0000-0000-0000-000000000007","menge": 2}],"gesamtZahlungCents": 4700,"kommentar": "Getränke sofort bezahlen"}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.bestellung-aufgegeben:v1', 'tisch:7', 6, now() - interval '21 hours 21 minutes', '{"bestellungId": "a0007000-0000-0000-0000-000000000003","positionen": [{"positionId": "b0007000-0000-0000-0000-000000000008","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 4},{"positionId": "b0007000-0000-0000-0000-000000000009","varianteId": 30,"produktName": "Weizen","varianteName": "Russ","kategorie": "beverage","einzelpreis": 300,"menge": 3}],"gesamtPreisCents": 2700,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.produkte-geliefert:v1', 'tisch:7', 7, now() - interval '21 hours 11 minutes', '{"lieferungId": "c0007000-0000-0000-0000-000000000003","positionen": [{"positionId": "b0007000-0000-0000-0000-000000000008","menge": 4},{"positionId": "b0007000-0000-0000-0000-000000000009","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:7', 8, now() - interval '21 hours 2 minutes', '{"bestellungId": "a0007000-0000-0000-0000-000000000004","positionen": [{"positionId": "b0007000-0000-0000-0000-000000000010","varianteId": 12,"produktName": "Grillplatte","varianteName": "Klein","kategorie": "food","einzelpreis": 800,"menge": 1},{"positionId": "b0007000-0000-0000-0000-000000000011","varianteId": 3,"produktName": "Bratwurst","varianteName": "Currywurst","kategorie": "food","einzelpreis": 450,"menge": 2},{"positionId": "b0007000-0000-0000-0000-000000000012","varianteId": 34,"produktName": "Softdrinks","varianteName": "Sprite","kategorie": "beverage","einzelpreis": 280,"menge": 2}],"gesamtPreisCents": 2260,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.produkte-geliefert:v1', 'tisch:7', 9, now() - interval '20 hours 54 minutes', '{"lieferungId": "c0007000-0000-0000-0000-000000000004","positionen": [{"positionId": "b0007000-0000-0000-0000-000000000010","menge": 1},{"positionId": "b0007000-0000-0000-0000-000000000011","menge": 2},{"positionId": "b0007000-0000-0000-0000-000000000012","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.zahlung-registriert:v1', 'tisch:7', 10, now() - interval '20 hours 45 minutes', '{"zahlungId": "d0007000-0000-0000-0000-000000000002","positionen": [{"positionId": "b0007000-0000-0000-0000-000000000001","menge": 3},{"positionId": "b0007000-0000-0000-0000-000000000004","menge": 3},{"positionId": "b0007000-0000-0000-0000-000000000005","menge": 3}],"gesamtZahlungCents": 4200,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.bestellung-aufgegeben:v1', 'tisch:7', 11, now() - interval '20 hours 25 minutes', '{"bestellungId": "a0007000-0000-0000-0000-000000000005","positionen": [{"positionId": "b0007000-0000-0000-0000-000000000013","varianteId": 49,"produktName": "Hugo/Aperol","varianteName": "Aperol Spritz","kategorie": "beverage","einzelpreis": 550,"menge": 4},{"positionId": "b0007000-0000-0000-0000-000000000014","varianteId": 48,"produktName": "Hugo/Aperol","varianteName": "Hugo","kategorie": "beverage","einzelpreis": 550,"menge": 2}],"gesamtPreisCents": 3300,"kommentar": "Cocktailrunde"}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.produkte-geliefert:v1', 'tisch:7', 12, now() - interval '20 hours 15 minutes', '{"lieferungId": "c0007000-0000-0000-0000-000000000005","positionen": [{"positionId": "b0007000-0000-0000-0000-000000000013","menge": 4},{"positionId": "b0007000-0000-0000-0000-000000000014","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.zahlung-registriert:v1', 'tisch:7', 13, now() - interval '19 hours 55 minutes', '{"zahlungId": "d0007000-0000-0000-0000-000000000003","positionen": [{"positionId": "b0007000-0000-0000-0000-000000000013","menge": 4},{"positionId": "b0007000-0000-0000-0000-000000000014","menge": 2}],"gesamtZahlungCents": 3300,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.zahlung-registriert:v1', 'tisch:7', 14, now() - interval '19 hours 35 minutes', '{"zahlungId": "d0007000-0000-0000-0000-000000000004","positionen": [{"positionId": "b0007000-0000-0000-0000-000000000008","menge": 4},{"positionId": "b0007000-0000-0000-0000-000000000009","menge": 3}],"gesamtZahlungCents": 2700,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:7', 15, now() - interval '19 hours 25 minutes', '{"bestellungId": "a0007000-0000-0000-0000-000000000006","positionen": [{"positionId": "b0007000-0000-0000-0000-000000000015","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 3},{"positionId": "b0007000-0000-0000-0000-000000000016","varianteId": 20,"produktName": "Brezel","varianteName": "Normal","kategorie": "food","einzelpreis": 200,"menge": 4}],"gesamtPreisCents": 2150,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.produkte-geliefert:v1', 'tisch:7', 16, now() - interval '19 hours 17 minutes', '{"lieferungId": "c0007000-0000-0000-0000-000000000006","positionen": [{"positionId": "b0007000-0000-0000-0000-000000000015","menge": 3},{"positionId": "b0007000-0000-0000-0000-000000000016","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.zahlung-registriert:v1', 'tisch:7', 17, now() - interval '19 hours 4 minutes', '{"zahlungId": "d0007000-0000-0000-0000-000000000005","positionen": [{"positionId": "b0007000-0000-0000-0000-000000000015","menge": 3},{"positionId": "b0007000-0000-0000-0000-000000000016","menge": 4}],"gesamtZahlungCents": 2150,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:7', 18, now() - interval '4 hours', '{"bestellungId": "a0007000-0000-0000-0000-000000000007","positionen": [{"positionId": "b0007000-0000-0000-0000-000000000017","varianteId": 11,"produktName": "Tagesgericht","varianteName": "So: Hähnchen mit Reis","kategorie": "food","einzelpreis": 1050,"menge": 3},{"positionId": "b0007000-0000-0000-0000-000000000018","varianteId": 14,"produktName": "Salat","varianteName": "Gemischter Salat","kategorie": "food","einzelpreis": 550,"menge": 2},{"positionId": "b0007000-0000-0000-0000-000000000019","varianteId": 37,"produktName": "Wasser","varianteName": "Medium 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 3},{"positionId": "b0007000-0000-0000-0000-000000000020","varianteId": 40,"produktName": "Saftschorle","varianteName": "Johannisbeerschorle 0,5l","kategorie": "beverage","einzelpreis": 300,"menge": 2}],"gesamtPreisCents": 5450,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.produkte-geliefert:v1', 'tisch:7', 19, now() - interval '3 hours 52 minutes', '{"lieferungId": "c0007000-0000-0000-0000-000000000007","positionen": [{"positionId": "b0007000-0000-0000-0000-000000000017","menge": 3},{"positionId": "b0007000-0000-0000-0000-000000000018","menge": 2},{"positionId": "b0007000-0000-0000-0000-000000000019","menge": 3},{"positionId": "b0007000-0000-0000-0000-000000000020","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.bestellung-aufgegeben:v1', 'tisch:7', 20, now() - interval '3 hours 32 minutes', '{"bestellungId": "a0007000-0000-0000-0000-000000000008","positionen": [{"positionId": "b0007000-0000-0000-0000-000000000021","varianteId": 16,"produktName": "Kuchen","varianteName": "Stück","kategorie": "food","einzelpreis": 250,"menge": 3},{"positionId": "b0007000-0000-0000-0000-000000000022","varianteId": 45,"produktName": "Kaffee","varianteName": "Tasse","kategorie": "beverage","einzelpreis": 200,"menge": 3}],"gesamtPreisCents": 1350,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.produkte-geliefert:v1', 'tisch:7', 21, now() - interval '3 hours 24 minutes', '{"lieferungId": "c0007000-0000-0000-0000-000000000008","positionen": [{"positionId": "b0007000-0000-0000-0000-000000000021","menge": 3},{"positionId": "b0007000-0000-0000-0000-000000000022","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.zahlung-registriert:v1', 'tisch:7', 22, now() - interval '3 hours 15 minutes', '{"zahlungId": "d0007000-0000-0000-0000-000000000006","positionen": [{"positionId": "b0007000-0000-0000-0000-000000000010","menge": 1},{"positionId": "b0007000-0000-0000-0000-000000000011","menge": 2},{"positionId": "b0007000-0000-0000-0000-000000000012","menge": 2}],"gesamtZahlungCents": 2260,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.zahlung-registriert:v1', 'tisch:7', 23, now() - interval '3 hours 5 minutes', '{"zahlungId": "d0007000-0000-0000-0000-000000000007","positionen": [{"positionId": "b0007000-0000-0000-0000-000000000017","menge": 3},{"positionId": "b0007000-0000-0000-0000-000000000018","menge": 2},{"positionId": "b0007000-0000-0000-0000-000000000019","menge": 3},{"positionId": "b0007000-0000-0000-0000-000000000020","menge": 2}],"gesamtZahlungCents": 5450,"kommentar": ""}');

-- ─────────────────────────────────────────────────────────────────────────────
-- TISCH 8: Familientisch — alle 3 Tage, hoher Umsatz (102 Events)
-- ─────────────────────────────────────────────────────────────────────────────

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:8', 1, now() - interval '50 hours', '{"bestellungId": "a0008000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000001","varianteId": 9,"produktName": "Tagesgericht","varianteName": "Fr: Schnitzel mit Pommes","kategorie": "food","einzelpreis": 1250,"menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000002","varianteId": 5,"produktName": "Pommes","varianteName": "Groß","kategorie": "food","einzelpreis": 350,"menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000003","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 2},{"positionId": "b0008000-0000-0000-0000-000000000004","varianteId": 33,"produktName": "Softdrinks","varianteName": "Spezi","kategorie": "beverage","einzelpreis": 280,"menge": 3}],"gesamtPreisCents": 6540,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:8', 2, now() - interval '49 hours 49 minutes', '{"lieferungId": "c0008000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000001","menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000002","menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000003","menge": 2},{"positionId": "b0008000-0000-0000-0000-000000000004","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:8', 3, now() - interval '49 hours 28 minutes', '{"zahlungId": "d0008000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000001","menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000002","menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000003","menge": 2},{"positionId": "b0008000-0000-0000-0000-000000000004","menge": 3}],"gesamtZahlungCents": 6540,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:8', 4, now() - interval '49 hours 10 minutes', '{"bestellungId": "a0008000-0000-0000-0000-000000000002","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000005","varianteId": 1,"produktName": "Bratwurst","varianteName": "Normal","kategorie": "food","einzelpreis": 350,"menge": 2},{"positionId": "b0008000-0000-0000-0000-000000000006","varianteId": 4,"produktName": "Pommes","varianteName": "Klein","kategorie": "food","einzelpreis": 250,"menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000007","varianteId": 23,"produktName": "Bier","varianteName": "0,3l","kategorie": "beverage","einzelpreis": 300,"menge": 4}],"gesamtPreisCents": 2650,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:8', 5, now() - interval '48 hours 59 minutes', '{"lieferungId": "c0008000-0000-0000-0000-000000000002","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000005","menge": 2},{"positionId": "b0008000-0000-0000-0000-000000000006","menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000007","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:8', 6, now() - interval '48 hours 40 minutes', '{"zahlungId": "d0008000-0000-0000-0000-000000000002","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000005","menge": 2},{"positionId": "b0008000-0000-0000-0000-000000000006","menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000007","menge": 4}],"gesamtZahlungCents": 2650,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:8', 7, now() - interval '48 hours 23 minutes', '{"bestellungId": "a0008000-0000-0000-0000-000000000003","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000008","varianteId": 7,"produktName": "Flammkuchen","varianteName": "Speck & Zwiebel","kategorie": "food","einzelpreis": 700,"menge": 2},{"positionId": "b0008000-0000-0000-0000-000000000009","varianteId": 36,"produktName": "Wasser","varianteName": "Still 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 3}],"gesamtPreisCents": 2000,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.produkte-geliefert:v1', 'tisch:8', 8, now() - interval '48 hours 8 minutes', '{"lieferungId": "c0008000-0000-0000-0000-000000000003","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000008","menge": 2},{"positionId": "b0008000-0000-0000-0000-000000000009","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.zahlung-registriert:v1', 'tisch:8', 9, now() - interval '47 hours 51 minutes', '{"zahlungId": "d0008000-0000-0000-0000-000000000003","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000008","menge": 2},{"positionId": "b0008000-0000-0000-0000-000000000009","menge": 3}],"gesamtZahlungCents": 2000,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:8', 10, now() - interval '47 hours 32 minutes', '{"bestellungId": "a0008000-0000-0000-0000-000000000004","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000010","varianteId": 17,"produktName": "Waffeln","varianteName": "mit Puderzucker","kategorie": "food","einzelpreis": 300,"menge": 4},{"positionId": "b0008000-0000-0000-0000-000000000011","varianteId": 45,"produktName": "Kaffee","varianteName": "Tasse","kategorie": "beverage","einzelpreis": 200,"menge": 2},{"positionId": "b0008000-0000-0000-0000-000000000012","varianteId": 47,"produktName": "Tee","varianteName": "Verschiedene Sorten","kategorie": "beverage","einzelpreis": 200,"menge": 1}],"gesamtPreisCents": 1800,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:8', 11, now() - interval '47 hours 17 minutes', '{"lieferungId": "c0008000-0000-0000-0000-000000000004","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000010","menge": 4},{"positionId": "b0008000-0000-0000-0000-000000000011","menge": 2},{"positionId": "b0008000-0000-0000-0000-000000000012","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:8', 12, now() - interval '46 hours 55 minutes', '{"zahlungId": "d0008000-0000-0000-0000-000000000004","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000010","menge": 4},{"positionId": "b0008000-0000-0000-0000-000000000011","menge": 2},{"positionId": "b0008000-0000-0000-0000-000000000012","menge": 1}],"gesamtZahlungCents": 1800,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:8', 13, now() - interval '46 hours 33 minutes', '{"bestellungId": "a0008000-0000-0000-0000-000000000005","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000013","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000014","varianteId": 26,"produktName": "Weizen","varianteName": "Klein 0,3l","kategorie": "beverage","einzelpreis": 300,"menge": 2}],"gesamtPreisCents": 1950,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.produkte-geliefert:v1', 'tisch:8', 14, now() - interval '46 hours 22 minutes', '{"lieferungId": "c0008000-0000-0000-0000-000000000005","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000013","menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000014","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.zahlung-registriert:v1', 'tisch:8', 15, now() - interval '46 hours 4 minutes', '{"zahlungId": "d0008000-0000-0000-0000-000000000005","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000013","menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000014","menge": 2}],"gesamtZahlungCents": 1950,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:8', 16, now() - interval '26 hours 50 minutes', '{"bestellungId": "a0008000-0000-0000-0000-000000000006","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000015","varianteId": 10,"produktName": "Tagesgericht","varianteName": "Sa: Gulasch mit Spätzle","kategorie": "food","einzelpreis": 1150,"menge": 4},{"positionId": "b0008000-0000-0000-0000-000000000016","varianteId": 13,"produktName": "Grillplatte","varianteName": "Groß","kategorie": "food","einzelpreis": 1400,"menge": 1},{"positionId": "b0008000-0000-0000-0000-000000000017","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 4},{"positionId": "b0008000-0000-0000-0000-000000000018","varianteId": 27,"produktName": "Weizen","varianteName": "Groß 0,5l","kategorie": "beverage","einzelpreis": 400,"menge": 2}],"gesamtPreisCents": 8600,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.produkte-geliefert:v1', 'tisch:8', 17, now() - interval '26 hours 36 minutes', '{"lieferungId": "c0008000-0000-0000-0000-000000000006","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000015","menge": 4},{"positionId": "b0008000-0000-0000-0000-000000000016","menge": 1},{"positionId": "b0008000-0000-0000-0000-000000000017","menge": 4},{"positionId": "b0008000-0000-0000-0000-000000000018","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.zahlung-registriert:v1', 'tisch:8', 18, now() - interval '26 hours 15 minutes', '{"zahlungId": "d0008000-0000-0000-0000-000000000006","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000015","menge": 4},{"positionId": "b0008000-0000-0000-0000-000000000016","menge": 1},{"positionId": "b0008000-0000-0000-0000-000000000017","menge": 4},{"positionId": "b0008000-0000-0000-0000-000000000018","menge": 2}],"gesamtZahlungCents": 8600,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:8', 19, now() - interval '25 hours 58 minutes', '{"bestellungId": "a0008000-0000-0000-0000-000000000007","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000019","varianteId": 3,"produktName": "Bratwurst","varianteName": "Currywurst","kategorie": "food","einzelpreis": 450,"menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000020","varianteId": 5,"produktName": "Pommes","varianteName": "Groß","kategorie": "food","einzelpreis": 350,"menge": 2},{"positionId": "b0008000-0000-0000-0000-000000000021","varianteId": 31,"produktName": "Softdrinks","varianteName": "Cola","kategorie": "beverage","einzelpreis": 280,"menge": 4}],"gesamtPreisCents": 3170,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.produkte-geliefert:v1', 'tisch:8', 20, now() - interval '25 hours 47 minutes', '{"lieferungId": "c0008000-0000-0000-0000-000000000007","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000019","menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000020","menge": 2},{"positionId": "b0008000-0000-0000-0000-000000000021","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.zahlung-registriert:v1', 'tisch:8', 21, now() - interval '25 hours 33 minutes', '{"zahlungId": "d0008000-0000-0000-0000-000000000007","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000019","menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000020","menge": 2},{"positionId": "b0008000-0000-0000-0000-000000000021","menge": 4}],"gesamtZahlungCents": 3170,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:8', 22, now() - interval '25 hours 14 minutes', '{"bestellungId": "a0008000-0000-0000-0000-000000000008","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000022","varianteId": 8,"produktName": "Flammkuchen","varianteName": "Mediterran","kategorie": "food","einzelpreis": 750,"menge": 2},{"positionId": "b0008000-0000-0000-0000-000000000023","varianteId": 44,"produktName": "Wein","varianteName": "Rosé 0,2l","kategorie": "beverage","einzelpreis": 400,"menge": 2},{"positionId": "b0008000-0000-0000-0000-000000000024","varianteId": 48,"produktName": "Hugo/Aperol","varianteName": "Hugo","kategorie": "beverage","einzelpreis": 550,"menge": 2}],"gesamtPreisCents": 3400,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.produkte-geliefert:v1', 'tisch:8', 23, now() - interval '25 hours', '{"lieferungId": "c0008000-0000-0000-0000-000000000008","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000022","menge": 2},{"positionId": "b0008000-0000-0000-0000-000000000023","menge": 2},{"positionId": "b0008000-0000-0000-0000-000000000024","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.zahlung-registriert:v1', 'tisch:8', 24, now() - interval '24 hours 45 minutes', '{"zahlungId": "d0008000-0000-0000-0000-000000000008","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000022","menge": 2},{"positionId": "b0008000-0000-0000-0000-000000000023","menge": 2},{"positionId": "b0008000-0000-0000-0000-000000000024","menge": 2}],"gesamtZahlungCents": 3400,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:8', 25, now() - interval '24 hours 28 minutes', '{"bestellungId": "a0008000-0000-0000-0000-000000000009","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000025","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 4},{"positionId": "b0008000-0000-0000-0000-000000000026","varianteId": 29,"produktName": "Weizen","varianteName": "Colaweizen Groß","kategorie": "beverage","einzelpreis": 400,"menge": 2}],"gesamtPreisCents": 2600,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.produkte-geliefert:v1', 'tisch:8', 26, now() - interval '24 hours 15 minutes', '{"lieferungId": "c0008000-0000-0000-0000-000000000009","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000025","menge": 4},{"positionId": "b0008000-0000-0000-0000-000000000026","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:8', 27, now() - interval '24 hours', '{"bestellungId": "a0008000-0000-0000-0000-000000000010","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000027","varianteId": 16,"produktName": "Kuchen","varianteName": "Stück","kategorie": "food","einzelpreis": 250,"menge": 4},{"positionId": "b0008000-0000-0000-0000-000000000028","varianteId": 19,"produktName": "Waffeln","varianteName": "mit Nutella","kategorie": "food","einzelpreis": 400,"menge": 2},{"positionId": "b0008000-0000-0000-0000-000000000029","varianteId": 45,"produktName": "Kaffee","varianteName": "Tasse","kategorie": "beverage","einzelpreis": 200,"menge": 3}],"gesamtPreisCents": 2400,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.produkte-geliefert:v1', 'tisch:8', 28, now() - interval '23 hours 52 minutes', '{"lieferungId": "c0008000-0000-0000-0000-000000000010","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000027","menge": 4},{"positionId": "b0008000-0000-0000-0000-000000000028","menge": 2},{"positionId": "b0008000-0000-0000-0000-000000000029","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.zahlung-registriert:v1', 'tisch:8', 29, now() - interval '23 hours 33 minutes', '{"zahlungId": "d0008000-0000-0000-0000-000000000009","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000025","menge": 4},{"positionId": "b0008000-0000-0000-0000-000000000026","menge": 2},{"positionId": "b0008000-0000-0000-0000-000000000027","menge": 4},{"positionId": "b0008000-0000-0000-0000-000000000028","menge": 2},{"positionId": "b0008000-0000-0000-0000-000000000029","menge": 3}],"gesamtZahlungCents": 5000,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:8', 30, now() - interval '23 hours 23 minutes', '{"bestellungId": "a0008000-0000-0000-0000-000000000011","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000030","varianteId": 1,"produktName": "Bratwurst","varianteName": "Normal","kategorie": "food","einzelpreis": 350,"menge": 2},{"positionId": "b0008000-0000-0000-0000-000000000031","varianteId": 23,"produktName": "Bier","varianteName": "0,3l","kategorie": "beverage","einzelpreis": 300,"menge": 4}],"gesamtPreisCents": 1900,"kommentar": "Noch eine Runde"}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.produkte-geliefert:v1', 'tisch:8', 31, now() - interval '23 hours 10 minutes', '{"lieferungId": "c0008000-0000-0000-0000-000000000011","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000030","menge": 2},{"positionId": "b0008000-0000-0000-0000-000000000031","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.zahlung-registriert:v1', 'tisch:8', 32, now() - interval '22 hours 48 minutes', '{"zahlungId": "d0008000-0000-0000-0000-000000000010","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000030","menge": 2},{"positionId": "b0008000-0000-0000-0000-000000000031","menge": 4}],"gesamtZahlungCents": 1900,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:8', 33, now() - interval '22 hours 33 minutes', '{"bestellungId": "a0008000-0000-0000-0000-000000000012","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000032","varianteId": 17,"produktName": "Waffeln","varianteName": "mit Puderzucker","kategorie": "food","einzelpreis": 300,"menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000033","varianteId": 18,"produktName": "Waffeln","varianteName": "mit Sahne","kategorie": "food","einzelpreis": 350,"menge": 2},{"positionId": "b0008000-0000-0000-0000-000000000034","varianteId": 46,"produktName": "Kaffee","varianteName": "Espresso","kategorie": "beverage","einzelpreis": 180,"menge": 2}],"gesamtPreisCents": 1960,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.produkte-geliefert:v1', 'tisch:8', 34, now() - interval '22 hours 22 minutes', '{"lieferungId": "c0008000-0000-0000-0000-000000000012","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000032","menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000033","menge": 2},{"positionId": "b0008000-0000-0000-0000-000000000034","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.zahlung-registriert:v1', 'tisch:8', 35, now() - interval '21 hours 58 minutes', '{"zahlungId": "d0008000-0000-0000-0000-000000000011","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000032","menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000033","menge": 2},{"positionId": "b0008000-0000-0000-0000-000000000034","menge": 2}],"gesamtZahlungCents": 1960,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:8', 36, now() - interval '19 hours 7 minutes', '{"bestellungId": "a0008000-0000-0000-0000-000000000023","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000060","varianteId": 6,"produktName": "Flammkuchen","varianteName": "Classic","kategorie": "food","einzelpreis": 600,"menge": 2},{"positionId": "b0008000-0000-0000-0000-000000000061","varianteId": 36,"produktName": "Wasser","varianteName": "Still 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 2},{"positionId": "b0008000-0000-0000-0000-000000000062","varianteId": 32,"produktName": "Softdrinks","varianteName": "Fanta","kategorie": "beverage","einzelpreis": 280,"menge": 3}],"gesamtPreisCents": 2440,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:8', 37, now() - interval '19 hours 3 minutes', '{"bestellungId": "a0008000-0000-0000-0000-000000000035","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000095","varianteId": 3,"produktName": "Bratwurst","varianteName": "Currywurst","kategorie": "food","einzelpreis": 450,"menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000096","varianteId": 28,"produktName": "Weizen","varianteName": "Colaweizen Klein","kategorie": "beverage","einzelpreis": 300,"menge": 2}],"gesamtPreisCents": 1950,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:8', 38, now() - interval '18 hours 56 minutes', '{"lieferungId": "c0008000-0000-0000-0000-000000000023","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000060","menge": 2},{"positionId": "b0008000-0000-0000-0000-000000000061","menge": 2},{"positionId": "b0008000-0000-0000-0000-000000000062","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.bestellung-aufgegeben:v1', 'tisch:8', 39, now() - interval '18 hours 56 minutes', '{"bestellungId": "a0008000-0000-0000-0000-000000000033","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000090","varianteId": 14,"produktName": "Salat","varianteName": "Gemischter Salat","kategorie": "food","einzelpreis": 550,"menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000091","varianteId": 34,"produktName": "Softdrinks","varianteName": "Sprite","kategorie": "beverage","einzelpreis": 280,"menge": 1}],"gesamtPreisCents": 1930,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.produkte-geliefert:v1', 'tisch:8', 40, now() - interval '18 hours 51 minutes', '{"lieferungId": "c0008000-0000-0000-0000-000000000035","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000095","menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000096","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.produkte-geliefert:v1', 'tisch:8', 41, now() - interval '18 hours 48 minutes', '{"lieferungId": "c0008000-0000-0000-0000-000000000033","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000090","menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000091","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:8', 42, now() - interval '18 hours 46 minutes', '{"bestellungId": "a0008000-0000-0000-0000-000000000018","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000049","varianteId": 16,"produktName": "Kuchen","varianteName": "Stück","kategorie": "food","einzelpreis": 250,"menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000050","varianteId": 2,"produktName": "Bratwurst","varianteName": "XXL","kategorie": "food","einzelpreis": 500,"menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000051","varianteId": 32,"produktName": "Softdrinks","varianteName": "Fanta","kategorie": "beverage","einzelpreis": 280,"menge": 1}],"gesamtPreisCents": 2530,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.bestellung-aufgegeben:v1', 'tisch:8', 43, now() - interval '18 hours 43 minutes', '{"bestellungId": "a0008000-0000-0000-0000-000000000020","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000054","varianteId": 8,"produktName": "Flammkuchen","varianteName": "Mediterran","kategorie": "food","einzelpreis": 750,"menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000055","varianteId": 46,"produktName": "Kaffee","varianteName": "Espresso","kategorie": "beverage","einzelpreis": 180,"menge": 1}],"gesamtPreisCents": 2430,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.bestellung-aufgegeben:v1', 'tisch:8', 44, now() - interval '18 hours 43 minutes', '{"bestellungId": "a0008000-0000-0000-0000-000000000021","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000056","varianteId": 2,"produktName": "Bratwurst","varianteName": "XXL","kategorie": "food","einzelpreis": 500,"menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000057","varianteId": 49,"produktName": "Hugo/Aperol","varianteName": "Aperol Spritz","kategorie": "beverage","einzelpreis": 550,"menge": 4}],"gesamtPreisCents": 3700,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.bestellung-aufgegeben:v1', 'tisch:8', 45, now() - interval '18 hours 43 minutes', '{"bestellungId": "a0008000-0000-0000-0000-000000000030","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000081","varianteId": 15,"produktName": "Salat","varianteName": "Caesar Salat","kategorie": "food","einzelpreis": 650,"menge": 2},{"positionId": "b0008000-0000-0000-0000-000000000082","varianteId": 16,"produktName": "Kuchen","varianteName": "Stück","kategorie": "food","einzelpreis": 250,"menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000083","varianteId": 36,"produktName": "Wasser","varianteName": "Still 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 4},{"positionId": "b0008000-0000-0000-0000-000000000084","varianteId": 25,"produktName": "Bier","varianteName": "Maß 1,0l","kategorie": "beverage","einzelpreis": 850,"menge": 4}],"gesamtPreisCents": 6250,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.zahlung-registriert:v1', 'tisch:8', 46, now() - interval '18 hours 36 minutes', '{"zahlungId": "d0008000-0000-0000-0000-000000000032","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000095","menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000096","menge": 2}],"gesamtZahlungCents": 1950,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.produkte-geliefert:v1', 'tisch:8', 47, now() - interval '18 hours 35 minutes', '{"lieferungId": "c0008000-0000-0000-0000-000000000030","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000081","menge": 2},{"positionId": "b0008000-0000-0000-0000-000000000082","menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000083","menge": 4},{"positionId": "b0008000-0000-0000-0000-000000000084","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:8', 48, now() - interval '18 hours 32 minutes', '{"zahlungId": "d0008000-0000-0000-0000-000000000020","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000060","menge": 2},{"positionId": "b0008000-0000-0000-0000-000000000061","menge": 2},{"positionId": "b0008000-0000-0000-0000-000000000062","menge": 3}],"gesamtZahlungCents": 2440,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.produkte-geliefert:v1', 'tisch:8', 49, now() - interval '18 hours 31 minutes', '{"lieferungId": "c0008000-0000-0000-0000-000000000018","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000049","menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000050","menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000051","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:8', 50, now() - interval '18 hours 29 minutes', '{"bestellungId": "a0008000-0000-0000-0000-000000000032","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000087","varianteId": 4,"produktName": "Pommes","varianteName": "Klein","kategorie": "food","einzelpreis": 250,"menge": 2},{"positionId": "b0008000-0000-0000-0000-000000000088","varianteId": 14,"produktName": "Salat","varianteName": "Gemischter Salat","kategorie": "food","einzelpreis": 550,"menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000089","varianteId": 42,"produktName": "Wein","varianteName": "Weißwein 0,2l","kategorie": "beverage","einzelpreis": 400,"menge": 2}],"gesamtPreisCents": 2950,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.produkte-geliefert:v1', 'tisch:8', 51, now() - interval '18 hours 28 minutes', '{"lieferungId": "c0008000-0000-0000-0000-000000000020","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000054","menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000055","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.produkte-geliefert:v1', 'tisch:8', 52, now() - interval '18 hours 28 minutes', '{"lieferungId": "c0008000-0000-0000-0000-000000000021","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000056","menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000057","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.zahlung-registriert:v1', 'tisch:8', 53, now() - interval '18 hours 28 minutes', '{"zahlungId": "d0008000-0000-0000-0000-000000000030","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000090","menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000091","menge": 1}],"gesamtZahlungCents": 1930,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:8', 54, now() - interval '18 hours 27 minutes', '{"bestellungId": "a0008000-0000-0000-0000-000000000016","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000043","varianteId": 1,"produktName": "Bratwurst","varianteName": "Normal","kategorie": "food","einzelpreis": 350,"menge": 1},{"positionId": "b0008000-0000-0000-0000-000000000044","varianteId": 5,"produktName": "Pommes","varianteName": "Groß","kategorie": "food","einzelpreis": 350,"menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000045","varianteId": 25,"produktName": "Bier","varianteName": "Maß 1,0l","kategorie": "beverage","einzelpreis": 850,"menge": 2}],"gesamtPreisCents": 3100,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:8', 55, now() - interval '18 hours 20 minutes', '{"bestellungId": "a0008000-0000-0000-0000-000000000028","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000075","varianteId": 21,"produktName": "Brezel","varianteName": "mit Butter","kategorie": "food","einzelpreis": 300,"menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000076","varianteId": 8,"produktName": "Flammkuchen","varianteName": "Mediterran","kategorie": "food","einzelpreis": 750,"menge": 2},{"positionId": "b0008000-0000-0000-0000-000000000077","varianteId": 40,"produktName": "Saftschorle","varianteName": "Johannisbeerschorle 0,5l","kategorie": "beverage","einzelpreis": 300,"menge": 2}],"gesamtPreisCents": 3000,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.zahlung-registriert:v1', 'tisch:8', 56, now() - interval '18 hours 18 minutes', '{"zahlungId": "d0008000-0000-0000-0000-000000000027","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000081","menge": 2},{"positionId": "b0008000-0000-0000-0000-000000000082","menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000083","menge": 4},{"positionId": "b0008000-0000-0000-0000-000000000084","menge": 4}],"gesamtZahlungCents": 6250,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:8', 57, now() - interval '18 hours 16 minutes', '{"lieferungId": "c0008000-0000-0000-0000-000000000032","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000087","menge": 2},{"positionId": "b0008000-0000-0000-0000-000000000088","menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000089","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.zahlung-registriert:v1', 'tisch:8', 58, now() - interval '18 hours 15 minutes', '{"zahlungId": "d0008000-0000-0000-0000-000000000017","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000054","menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000055","menge": 1}],"gesamtZahlungCents": 2430,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.produkte-geliefert:v1', 'tisch:8', 59, now() - interval '18 hours 13 minutes', '{"lieferungId": "c0008000-0000-0000-0000-000000000016","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000043","menge": 1},{"positionId": "b0008000-0000-0000-0000-000000000044","menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000045","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.zahlung-registriert:v1', 'tisch:8', 60, now() - interval '18 hours 12 minutes', '{"zahlungId": "d0008000-0000-0000-0000-000000000018","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000056","menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000057","menge": 4}],"gesamtZahlungCents": 3700,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:8', 61, now() - interval '18 hours 9 minutes', '{"bestellungId": "a0008000-0000-0000-0000-000000000031","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000085","varianteId": 2,"produktName": "Bratwurst","varianteName": "XXL","kategorie": "food","einzelpreis": 500,"menge": 1},{"positionId": "b0008000-0000-0000-0000-000000000086","varianteId": 44,"produktName": "Wein","varianteName": "Rosé 0,2l","kategorie": "beverage","einzelpreis": 400,"menge": 2}],"gesamtPreisCents": 1300,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.zahlung-registriert:v1', 'tisch:8', 62, now() - interval '18 hours 8 minutes', '{"zahlungId": "d0008000-0000-0000-0000-000000000015","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000049","menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000050","menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000051","menge": 1}],"gesamtZahlungCents": 2530,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:8', 63, now() - interval '18 hours 6 minutes', '{"bestellungId": "a0008000-0000-0000-0000-000000000026","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000070","varianteId": 15,"produktName": "Salat","varianteName": "Caesar Salat","kategorie": "food","einzelpreis": 650,"menge": 1},{"positionId": "b0008000-0000-0000-0000-000000000071","varianteId": 14,"produktName": "Salat","varianteName": "Gemischter Salat","kategorie": "food","einzelpreis": 550,"menge": 1},{"positionId": "b0008000-0000-0000-0000-000000000072","varianteId": 37,"produktName": "Wasser","varianteName": "Medium 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 2}],"gesamtPreisCents": 1600,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.produkte-geliefert:v1', 'tisch:8', 64, now() - interval '18 hours 5 minutes', '{"lieferungId": "c0008000-0000-0000-0000-000000000028","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000075","menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000076","menge": 2},{"positionId": "b0008000-0000-0000-0000-000000000077","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.bestellung-aufgegeben:v1', 'tisch:8', 65, now() - interval '18 hours 1 minutes', '{"bestellungId": "a0008000-0000-0000-0000-000000000022","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000058","varianteId": 7,"produktName": "Flammkuchen","varianteName": "Speck & Zwiebel","kategorie": "food","einzelpreis": 700,"menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000059","varianteId": 37,"produktName": "Wasser","varianteName": "Medium 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 1}],"gesamtPreisCents": 2300,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.produkte-geliefert:v1', 'tisch:8', 66, now() - interval '18 hours 1 minutes', '{"lieferungId": "c0008000-0000-0000-0000-000000000031","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000085","menge": 1},{"positionId": "b0008000-0000-0000-0000-000000000086","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.zahlung-registriert:v1', 'tisch:8', 67, now() - interval '18 hours', '{"zahlungId": "d0008000-0000-0000-0000-000000000013","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000043","menge": 1},{"positionId": "b0008000-0000-0000-0000-000000000044","menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000045","menge": 2}],"gesamtZahlungCents": 3100,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:8', 68, now() - interval '17 hours 56 minutes', '{"bestellungId": "a0008000-0000-0000-0000-000000000017","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000046","varianteId": 3,"produktName": "Bratwurst","varianteName": "Currywurst","kategorie": "food","einzelpreis": 450,"menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000047","varianteId": 1,"produktName": "Bratwurst","varianteName": "Normal","kategorie": "food","einzelpreis": 350,"menge": 2},{"positionId": "b0008000-0000-0000-0000-000000000048","varianteId": 44,"produktName": "Wein","varianteName": "Rosé 0,2l","kategorie": "beverage","einzelpreis": 400,"menge": 1}],"gesamtPreisCents": 2450,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.produkte-geliefert:v1', 'tisch:8', 69, now() - interval '17 hours 56 minutes', '{"lieferungId": "c0008000-0000-0000-0000-000000000026","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000070","menge": 1},{"positionId": "b0008000-0000-0000-0000-000000000071","menge": 1},{"positionId": "b0008000-0000-0000-0000-000000000072","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:8', 70, now() - interval '17 hours 56 minutes', '{"zahlungId": "d0008000-0000-0000-0000-000000000029","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000087","menge": 2},{"positionId": "b0008000-0000-0000-0000-000000000088","menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000089","menge": 2}],"gesamtZahlungCents": 2950,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.bestellung-aufgegeben:v1', 'tisch:8', 71, now() - interval '17 hours 54 minutes', '{"bestellungId": "a0008000-0000-0000-0000-000000000027","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000073","varianteId": 1,"produktName": "Bratwurst","varianteName": "Normal","kategorie": "food","einzelpreis": 350,"menge": 1},{"positionId": "b0008000-0000-0000-0000-000000000074","varianteId": 37,"produktName": "Wasser","varianteName": "Medium 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 3}],"gesamtPreisCents": 950,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.produkte-geliefert:v1', 'tisch:8', 72, now() - interval '17 hours 52 minutes', '{"lieferungId": "c0008000-0000-0000-0000-000000000022","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000058","menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000059","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.bestellung-aufgegeben:v1', 'tisch:8', 73, now() - interval '17 hours 50 minutes', '{"bestellungId": "a0008000-0000-0000-0000-000000000024","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000063","varianteId": 16,"produktName": "Kuchen","varianteName": "Stück","kategorie": "food","einzelpreis": 250,"menge": 1},{"positionId": "b0008000-0000-0000-0000-000000000064","varianteId": 5,"produktName": "Pommes","varianteName": "Groß","kategorie": "food","einzelpreis": 350,"menge": 1},{"positionId": "b0008000-0000-0000-0000-000000000065","varianteId": 41,"produktName": "Saftschorle","varianteName": "Rhabarberschorle 0,5l","kategorie": "beverage","einzelpreis": 350,"menge": 2},{"positionId": "b0008000-0000-0000-0000-000000000066","varianteId": 32,"produktName": "Softdrinks","varianteName": "Fanta","kategorie": "beverage","einzelpreis": 280,"menge": 4}],"gesamtPreisCents": 2420,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.bestellung-aufgegeben:v1', 'tisch:8', 74, now() - interval '17 hours 48 minutes', '{"bestellungId": "a0008000-0000-0000-0000-000000000025","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000067","varianteId": 2,"produktName": "Bratwurst","varianteName": "XXL","kategorie": "food","einzelpreis": 500,"menge": 1},{"positionId": "b0008000-0000-0000-0000-000000000068","varianteId": 19,"produktName": "Waffeln","varianteName": "mit Nutella","kategorie": "food","einzelpreis": 400,"menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000069","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 1}],"gesamtPreisCents": 2150,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.zahlung-registriert:v1', 'tisch:8', 75, now() - interval '17 hours 47 minutes', '{"zahlungId": "d0008000-0000-0000-0000-000000000028","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000085","menge": 1},{"positionId": "b0008000-0000-0000-0000-000000000086","menge": 2}],"gesamtZahlungCents": 1300,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:8', 76, now() - interval '17 hours 46 minutes', '{"bestellungId": "a0008000-0000-0000-0000-000000000029","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000078","varianteId": 3,"produktName": "Bratwurst","varianteName": "Currywurst","kategorie": "food","einzelpreis": 450,"menge": 1},{"positionId": "b0008000-0000-0000-0000-000000000079","varianteId": 16,"produktName": "Kuchen","varianteName": "Stück","kategorie": "food","einzelpreis": 250,"menge": 1},{"positionId": "b0008000-0000-0000-0000-000000000080","varianteId": 36,"produktName": "Wasser","varianteName": "Still 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 1}],"gesamtPreisCents": 900,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.produkte-geliefert:v1', 'tisch:8', 77, now() - interval '17 hours 44 minutes', '{"lieferungId": "c0008000-0000-0000-0000-000000000017","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000046","menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000047","menge": 2},{"positionId": "b0008000-0000-0000-0000-000000000048","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.zahlung-registriert:v1', 'tisch:8', 78, now() - interval '17 hours 44 minutes', '{"zahlungId": "d0008000-0000-0000-0000-000000000025","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000075","menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000076","menge": 2},{"positionId": "b0008000-0000-0000-0000-000000000077","menge": 2}],"gesamtZahlungCents": 3000,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.bestellung-aufgegeben:v1', 'tisch:8', 79, now() - interval '17 hours 43 minutes', '{"bestellungId": "a0008000-0000-0000-0000-000000000019","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000052","varianteId": 2,"produktName": "Bratwurst","varianteName": "XXL","kategorie": "food","einzelpreis": 500,"menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000053","varianteId": 30,"produktName": "Weizen","varianteName": "Russ","kategorie": "beverage","einzelpreis": 300,"menge": 4}],"gesamtPreisCents": 2700,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.produkte-geliefert:v1', 'tisch:8', 80, now() - interval '17 hours 42 minutes', '{"lieferungId": "c0008000-0000-0000-0000-000000000027","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000073","menge": 1},{"positionId": "b0008000-0000-0000-0000-000000000074","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:8', 81, now() - interval '17 hours 39 minutes', '{"bestellungId": "a0008000-0000-0000-0000-000000000034","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000092","varianteId": 7,"produktName": "Flammkuchen","varianteName": "Speck & Zwiebel","kategorie": "food","einzelpreis": 700,"menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000093","varianteId": 27,"produktName": "Weizen","varianteName": "Groß 0,5l","kategorie": "beverage","einzelpreis": 400,"menge": 4},{"positionId": "b0008000-0000-0000-0000-000000000094","varianteId": 46,"produktName": "Kaffee","varianteName": "Espresso","kategorie": "beverage","einzelpreis": 180,"menge": 1}],"gesamtPreisCents": 3880,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.produkte-geliefert:v1', 'tisch:8', 82, now() - interval '17 hours 37 minutes', '{"lieferungId": "c0008000-0000-0000-0000-000000000024","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000063","menge": 1},{"positionId": "b0008000-0000-0000-0000-000000000064","menge": 1},{"positionId": "b0008000-0000-0000-0000-000000000065","menge": 2},{"positionId": "b0008000-0000-0000-0000-000000000066","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.produkte-geliefert:v1', 'tisch:8', 83, now() - interval '17 hours 36 minutes', '{"lieferungId": "c0008000-0000-0000-0000-000000000025","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000067","menge": 1},{"positionId": "b0008000-0000-0000-0000-000000000068","menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000069","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.produkte-geliefert:v1', 'tisch:8', 84, now() - interval '17 hours 33 minutes', '{"lieferungId": "c0008000-0000-0000-0000-000000000019","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000052","menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000053","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.zahlung-registriert:v1', 'tisch:8', 85, now() - interval '17 hours 31 minutes', '{"zahlungId": "d0008000-0000-0000-0000-000000000023","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000070","menge": 1},{"positionId": "b0008000-0000-0000-0000-000000000071","menge": 1},{"positionId": "b0008000-0000-0000-0000-000000000072","menge": 2}],"gesamtZahlungCents": 1600,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:8', 86, now() - interval '17 hours 31 minutes', '{"lieferungId": "c0008000-0000-0000-0000-000000000029","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000078","menge": 1},{"positionId": "b0008000-0000-0000-0000-000000000079","menge": 1},{"positionId": "b0008000-0000-0000-0000-000000000080","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.zahlung-registriert:v1', 'tisch:8', 87, now() - interval '17 hours 30 minutes', '{"zahlungId": "d0008000-0000-0000-0000-000000000024","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000073","menge": 1},{"positionId": "b0008000-0000-0000-0000-000000000074","menge": 3}],"gesamtZahlungCents": 950,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.produkte-geliefert:v1', 'tisch:8', 88, now() - interval '17 hours 30 minutes', '{"lieferungId": "c0008000-0000-0000-0000-000000000034","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000092","menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000093","menge": 4},{"positionId": "b0008000-0000-0000-0000-000000000094","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.zahlung-registriert:v1', 'tisch:8', 89, now() - interval '17 hours 27 minutes', '{"zahlungId": "d0008000-0000-0000-0000-000000000019","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000058","menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000059","menge": 1}],"gesamtZahlungCents": 2300,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.zahlung-registriert:v1', 'tisch:8', 90, now() - interval '17 hours 24 minutes', '{"zahlungId": "d0008000-0000-0000-0000-000000000021","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000063","menge": 1},{"positionId": "b0008000-0000-0000-0000-000000000064","menge": 1},{"positionId": "b0008000-0000-0000-0000-000000000065","menge": 2},{"positionId": "b0008000-0000-0000-0000-000000000066","menge": 4}],"gesamtZahlungCents": 2420,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.zahlung-registriert:v1', 'tisch:8', 91, now() - interval '17 hours 19 minutes', '{"zahlungId": "d0008000-0000-0000-0000-000000000014","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000046","menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000047","menge": 2},{"positionId": "b0008000-0000-0000-0000-000000000048","menge": 1}],"gesamtZahlungCents": 2450,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:8', 92, now() - interval '17 hours 18 minutes', '{"zahlungId": "d0008000-0000-0000-0000-000000000026","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000078","menge": 1},{"positionId": "b0008000-0000-0000-0000-000000000079","menge": 1},{"positionId": "b0008000-0000-0000-0000-000000000080","menge": 1}],"gesamtZahlungCents": 900,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.zahlung-registriert:v1', 'tisch:8', 93, now() - interval '17 hours 15 minutes', '{"zahlungId": "d0008000-0000-0000-0000-000000000022","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000067","menge": 1},{"positionId": "b0008000-0000-0000-0000-000000000068","menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000069","menge": 1}],"gesamtZahlungCents": 2150,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.zahlung-registriert:v1', 'tisch:8', 94, now() - interval '17 hours 11 minutes', '{"zahlungId": "d0008000-0000-0000-0000-000000000031","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000092","menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000093","menge": 4},{"positionId": "b0008000-0000-0000-0000-000000000094","menge": 1}],"gesamtZahlungCents": 3880,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.zahlung-registriert:v1', 'tisch:8', 95, now() - interval '17 hours 10 minutes', '{"zahlungId": "d0008000-0000-0000-0000-000000000016","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000052","menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000053","menge": 4}],"gesamtZahlungCents": 2700,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:8', 96, now() - interval '4 hours 30 minutes', '{"bestellungId": "a0008000-0000-0000-0000-000000000013","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000035","varianteId": 11,"produktName": "Tagesgericht","varianteName": "So: Hähnchen mit Reis","kategorie": "food","einzelpreis": 1050,"menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000036","varianteId": 15,"produktName": "Salat","varianteName": "Caesar Salat","kategorie": "food","einzelpreis": 650,"menge": 1},{"positionId": "b0008000-0000-0000-0000-000000000037","varianteId": 37,"produktName": "Wasser","varianteName": "Medium 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000038","varianteId": 39,"produktName": "Saftschorle","varianteName": "Apfelschorle 0,5l","kategorie": "beverage","einzelpreis": 300,"menge": 2}],"gesamtPreisCents": 5000,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.produkte-geliefert:v1', 'tisch:8', 97, now() - interval '4 hours 22 minutes', '{"lieferungId": "c0008000-0000-0000-0000-000000000013","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000035","menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000036","menge": 1},{"positionId": "b0008000-0000-0000-0000-000000000037","menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000038","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:8', 98, now() - interval '4 hours 1 minutes', '{"bestellungId": "a0008000-0000-0000-0000-000000000014","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000039","varianteId": 21,"produktName": "Brezel","varianteName": "mit Butter","kategorie": "food","einzelpreis": 300,"menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000040","varianteId": 45,"produktName": "Kaffee","varianteName": "Tasse","kategorie": "beverage","einzelpreis": 200,"menge": 3}],"gesamtPreisCents": 1500,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:8', 99, now() - interval '3 hours 46 minutes', '{"lieferungId": "c0008000-0000-0000-0000-000000000014","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000039","menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000040","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.zahlung-registriert:v1', 'tisch:8', 100, now() - interval '3 hours 21 minutes', '{"zahlungId": "d0008000-0000-0000-0000-000000000012","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000035","menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000036","menge": 1},{"positionId": "b0008000-0000-0000-0000-000000000037","menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000038","menge": 2},{"positionId": "b0008000-0000-0000-0000-000000000039","menge": 3},{"positionId": "b0008000-0000-0000-0000-000000000040","menge": 3}],"gesamtZahlungCents": 6500,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:8', 101, now() - interval '3 hours 11 minutes', '{"bestellungId": "a0008000-0000-0000-0000-000000000015","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000041","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 2},{"positionId": "b0008000-0000-0000-0000-000000000042","varianteId": 40,"produktName": "Saftschorle","varianteName": "Johannisbeerschorle 0,5l","kategorie": "beverage","einzelpreis": 300,"menge": 2}],"gesamtPreisCents": 1500,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.produkte-geliefert:v1', 'tisch:8', 102, now() - interval '3 hours 3 minutes', '{"lieferungId": "c0008000-0000-0000-0000-000000000015","positionen": [{"positionId": "b0008000-0000-0000-0000-000000000041","menge": 2},{"positionId": "b0008000-0000-0000-0000-000000000042","menge": 2}],"kommentar": ""}');

-- ─────────────────────────────────────────────────────────────────────────────
-- TISCH 9: Kegelclub — Tag 1+2 bezahlt, Tag 3 offen (83 Events)
-- ─────────────────────────────────────────────────────────────────────────────

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:9', 1, now() - interval '48 hours 20 minutes', '{"bestellungId": "a0009000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000001","varianteId": 1,"produktName": "Bratwurst","varianteName": "Normal","kategorie": "food","einzelpreis": 350,"menge": 4},{"positionId": "b0009000-0000-0000-0000-000000000002","varianteId": 5,"produktName": "Pommes","varianteName": "Groß","kategorie": "food","einzelpreis": 350,"menge": 2},{"positionId": "b0009000-0000-0000-0000-000000000003","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 6}],"gesamtPreisCents": 4800,"kommentar": "Kegelclub Freitagabend"}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.produkte-geliefert:v1', 'tisch:9', 2, now() - interval '48 hours 7 minutes', '{"lieferungId": "c0009000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000001","menge": 4},{"positionId": "b0009000-0000-0000-0000-000000000002","menge": 2},{"positionId": "b0009000-0000-0000-0000-000000000003","menge": 6}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.zahlung-registriert:v1', 'tisch:9', 3, now() - interval '47 hours 55 minutes', '{"zahlungId": "d0009000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000001","menge": 4},{"positionId": "b0009000-0000-0000-0000-000000000002","menge": 2},{"positionId": "b0009000-0000-0000-0000-000000000003","menge": 6}],"gesamtZahlungCents": 4800,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:9', 4, now() - interval '47 hours 35 minutes', '{"bestellungId": "a0009000-0000-0000-0000-000000000002","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000004","varianteId": 7,"produktName": "Flammkuchen","varianteName": "Speck & Zwiebel","kategorie": "food","einzelpreis": 700,"menge": 2},{"positionId": "b0009000-0000-0000-0000-000000000005","varianteId": 25,"produktName": "Bier","varianteName": "Maß 1,0l","kategorie": "beverage","einzelpreis": 850,"menge": 3}],"gesamtPreisCents": 3950,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.produkte-geliefert:v1', 'tisch:9', 5, now() - interval '47 hours 23 minutes', '{"lieferungId": "c0009000-0000-0000-0000-000000000002","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000004","menge": 2},{"positionId": "b0009000-0000-0000-0000-000000000005","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.zahlung-registriert:v1', 'tisch:9', 6, now() - interval '47 hours 1 minutes', '{"zahlungId": "d0009000-0000-0000-0000-000000000002","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000004","menge": 2},{"positionId": "b0009000-0000-0000-0000-000000000005","menge": 3}],"gesamtZahlungCents": 3950,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:9', 7, now() - interval '46 hours 44 minutes', '{"bestellungId": "a0009000-0000-0000-0000-000000000003","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000006","varianteId": 3,"produktName": "Bratwurst","varianteName": "Currywurst","kategorie": "food","einzelpreis": 450,"menge": 3},{"positionId": "b0009000-0000-0000-0000-000000000007","varianteId": 23,"produktName": "Bier","varianteName": "0,3l","kategorie": "beverage","einzelpreis": 300,"menge": 6}],"gesamtPreisCents": 3150,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.produkte-geliefert:v1', 'tisch:9', 8, now() - interval '46 hours 33 minutes', '{"lieferungId": "c0009000-0000-0000-0000-000000000003","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000006","menge": 3},{"positionId": "b0009000-0000-0000-0000-000000000007","menge": 6}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.zahlung-registriert:v1', 'tisch:9', 9, now() - interval '46 hours 12 minutes', '{"zahlungId": "d0009000-0000-0000-0000-000000000003","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000006","menge": 3},{"positionId": "b0009000-0000-0000-0000-000000000007","menge": 6}],"gesamtZahlungCents": 3150,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:9', 10, now() - interval '45 hours 56 minutes', '{"bestellungId": "a0009000-0000-0000-0000-000000000004","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000008","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 6},{"positionId": "b0009000-0000-0000-0000-000000000009","varianteId": 26,"produktName": "Weizen","varianteName": "Klein 0,3l","kategorie": "beverage","einzelpreis": 300,"menge": 3}],"gesamtPreisCents": 3600,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.produkte-geliefert:v1', 'tisch:9', 11, now() - interval '45 hours 42 minutes', '{"lieferungId": "c0009000-0000-0000-0000-000000000004","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000008","menge": 6},{"positionId": "b0009000-0000-0000-0000-000000000009","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.zahlung-registriert:v1', 'tisch:9', 12, now() - interval '45 hours 29 minutes', '{"zahlungId": "d0009000-0000-0000-0000-000000000004","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000008","menge": 6},{"positionId": "b0009000-0000-0000-0000-000000000009","menge": 3}],"gesamtZahlungCents": 3600,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:9', 13, now() - interval '45 hours 12 minutes', '{"bestellungId": "a0009000-0000-0000-0000-000000000005","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000010","varianteId": 16,"produktName": "Kuchen","varianteName": "Stück","kategorie": "food","einzelpreis": 250,"menge": 4},{"positionId": "b0009000-0000-0000-0000-000000000011","varianteId": 45,"produktName": "Kaffee","varianteName": "Tasse","kategorie": "beverage","einzelpreis": 200,"menge": 4}],"gesamtPreisCents": 1800,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.produkte-geliefert:v1', 'tisch:9', 14, now() - interval '45 hours 2 minutes', '{"lieferungId": "c0009000-0000-0000-0000-000000000005","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000010","menge": 4},{"positionId": "b0009000-0000-0000-0000-000000000011","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.zahlung-registriert:v1', 'tisch:9', 15, now() - interval '44 hours 50 minutes', '{"zahlungId": "d0009000-0000-0000-0000-000000000005","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000010","menge": 4},{"positionId": "b0009000-0000-0000-0000-000000000011","menge": 4}],"gesamtZahlungCents": 1800,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:9', 16, now() - interval '25 hours', '{"bestellungId": "a0009000-0000-0000-0000-000000000006","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000012","varianteId": 10,"produktName": "Tagesgericht","varianteName": "Sa: Gulasch mit Spätzle","kategorie": "food","einzelpreis": 1150,"menge": 4},{"positionId": "b0009000-0000-0000-0000-000000000013","varianteId": 12,"produktName": "Grillplatte","varianteName": "Klein","kategorie": "food","einzelpreis": 800,"menge": 2},{"positionId": "b0009000-0000-0000-0000-000000000014","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 6},{"positionId": "b0009000-0000-0000-0000-000000000015","varianteId": 27,"produktName": "Weizen","varianteName": "Groß 0,5l","kategorie": "beverage","einzelpreis": 400,"menge": 4}],"gesamtPreisCents": 10500,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.produkte-geliefert:v1', 'tisch:9', 17, now() - interval '24 hours 48 minutes', '{"lieferungId": "c0009000-0000-0000-0000-000000000006","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000012","menge": 4},{"positionId": "b0009000-0000-0000-0000-000000000013","menge": 2},{"positionId": "b0009000-0000-0000-0000-000000000014","menge": 6},{"positionId": "b0009000-0000-0000-0000-000000000015","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.zahlung-registriert:v1', 'tisch:9', 18, now() - interval '24 hours 32 minutes', '{"zahlungId": "d0009000-0000-0000-0000-000000000006","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000012","menge": 4},{"positionId": "b0009000-0000-0000-0000-000000000013","menge": 2},{"positionId": "b0009000-0000-0000-0000-000000000014","menge": 6},{"positionId": "b0009000-0000-0000-0000-000000000015","menge": 4}],"gesamtZahlungCents": 10500,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:9', 19, now() - interval '24 hours 10 minutes', '{"bestellungId": "a0009000-0000-0000-0000-000000000007","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000016","varianteId": 1,"produktName": "Bratwurst","varianteName": "Normal","kategorie": "food","einzelpreis": 350,"menge": 3},{"positionId": "b0009000-0000-0000-0000-000000000017","varianteId": 5,"produktName": "Pommes","varianteName": "Groß","kategorie": "food","einzelpreis": 350,"menge": 4},{"positionId": "b0009000-0000-0000-0000-000000000018","varianteId": 23,"produktName": "Bier","varianteName": "0,3l","kategorie": "beverage","einzelpreis": 300,"menge": 4},{"positionId": "b0009000-0000-0000-0000-000000000019","varianteId": 28,"produktName": "Weizen","varianteName": "Colaweizen Klein","kategorie": "beverage","einzelpreis": 300,"menge": 2}],"gesamtPreisCents": 4250,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.produkte-geliefert:v1', 'tisch:9', 20, now() - interval '24 hours', '{"lieferungId": "c0009000-0000-0000-0000-000000000007","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000016","menge": 3},{"positionId": "b0009000-0000-0000-0000-000000000017","menge": 4},{"positionId": "b0009000-0000-0000-0000-000000000018","menge": 4},{"positionId": "b0009000-0000-0000-0000-000000000019","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.zahlung-registriert:v1', 'tisch:9', 21, now() - interval '23 hours 43 minutes', '{"zahlungId": "d0009000-0000-0000-0000-000000000007","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000016","menge": 3},{"positionId": "b0009000-0000-0000-0000-000000000017","menge": 4},{"positionId": "b0009000-0000-0000-0000-000000000018","menge": 4},{"positionId": "b0009000-0000-0000-0000-000000000019","menge": 2}],"gesamtZahlungCents": 4250,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:9', 22, now() - interval '23 hours 21 minutes', '{"bestellungId": "a0009000-0000-0000-0000-000000000008","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000020","varianteId": 8,"produktName": "Flammkuchen","varianteName": "Mediterran","kategorie": "food","einzelpreis": 750,"menge": 2},{"positionId": "b0009000-0000-0000-0000-000000000021","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 4}],"gesamtPreisCents": 3300,"kommentar": "Nachtisch + Bier"}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.produkte-geliefert:v1', 'tisch:9', 23, now() - interval '23 hours 7 minutes', '{"lieferungId": "c0009000-0000-0000-0000-000000000008","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000020","menge": 2},{"positionId": "b0009000-0000-0000-0000-000000000021","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.zahlung-registriert:v1', 'tisch:9', 24, now() - interval '22 hours 48 minutes', '{"zahlungId": "d0009000-0000-0000-0000-000000000008","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000020","menge": 2},{"positionId": "b0009000-0000-0000-0000-000000000021","menge": 4}],"gesamtZahlungCents": 3300,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:9', 25, now() - interval '22 hours 28 minutes', '{"bestellungId": "a0009000-0000-0000-0000-000000000009","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000022","varianteId": 25,"produktName": "Bier","varianteName": "Maß 1,0l","kategorie": "beverage","einzelpreis": 850,"menge": 4},{"positionId": "b0009000-0000-0000-0000-000000000023","varianteId": 30,"produktName": "Weizen","varianteName": "Russ","kategorie": "beverage","einzelpreis": 300,"menge": 2}],"gesamtPreisCents": 4000,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.produkte-geliefert:v1', 'tisch:9', 26, now() - interval '22 hours 18 minutes', '{"lieferungId": "c0009000-0000-0000-0000-000000000009","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000022","menge": 4},{"positionId": "b0009000-0000-0000-0000-000000000023","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.zahlung-registriert:v1', 'tisch:9', 27, now() - interval '22 hours 6 minutes', '{"zahlungId": "d0009000-0000-0000-0000-000000000009","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000022","menge": 4},{"positionId": "b0009000-0000-0000-0000-000000000023","menge": 2}],"gesamtZahlungCents": 4000,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:9', 28, now() - interval '18 hours 17 minutes', '{"bestellungId": "a0009000-0000-0000-0000-000000000019","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000050","varianteId": 4,"produktName": "Pommes","varianteName": "Klein","kategorie": "food","einzelpreis": 250,"menge": 2},{"positionId": "b0009000-0000-0000-0000-000000000051","varianteId": 17,"produktName": "Waffeln","varianteName": "mit Puderzucker","kategorie": "food","einzelpreis": 300,"menge": 2},{"positionId": "b0009000-0000-0000-0000-000000000052","varianteId": 42,"produktName": "Wein","varianteName": "Weißwein 0,2l","kategorie": "beverage","einzelpreis": 400,"menge": 4},{"positionId": "b0009000-0000-0000-0000-000000000053","varianteId": 37,"produktName": "Wasser","varianteName": "Medium 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 3}],"gesamtPreisCents": 3300,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:9', 29, now() - interval '18 hours 11 minutes', '{"bestellungId": "a0009000-0000-0000-0000-000000000018","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000046","varianteId": 20,"produktName": "Brezel","varianteName": "Normal","kategorie": "food","einzelpreis": 200,"menge": 3},{"positionId": "b0009000-0000-0000-0000-000000000047","varianteId": 13,"produktName": "Grillplatte","varianteName": "Groß","kategorie": "food","einzelpreis": 1400,"menge": 3},{"positionId": "b0009000-0000-0000-0000-000000000048","varianteId": 37,"produktName": "Wasser","varianteName": "Medium 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 4},{"positionId": "b0009000-0000-0000-0000-000000000049","varianteId": 47,"produktName": "Tee","varianteName": "Verschiedene Sorten","kategorie": "beverage","einzelpreis": 200,"menge": 1}],"gesamtPreisCents": 5800,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.bestellung-aufgegeben:v1', 'tisch:9', 30, now() - interval '18 hours 10 minutes', '{"bestellungId": "a0009000-0000-0000-0000-000000000013","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000033","varianteId": 2,"produktName": "Bratwurst","varianteName": "XXL","kategorie": "food","einzelpreis": 500,"menge": 2},{"positionId": "b0009000-0000-0000-0000-000000000034","varianteId": 27,"produktName": "Weizen","varianteName": "Groß 0,5l","kategorie": "beverage","einzelpreis": 400,"menge": 1}],"gesamtPreisCents": 1400,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.bestellung-aufgegeben:v1', 'tisch:9', 31, now() - interval '18 hours 10 minutes', '{"bestellungId": "a0009000-0000-0000-0000-000000000014","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000035","varianteId": 21,"produktName": "Brezel","varianteName": "mit Butter","kategorie": "food","einzelpreis": 300,"menge": 3},{"positionId": "b0009000-0000-0000-0000-000000000036","varianteId": 20,"produktName": "Brezel","varianteName": "Normal","kategorie": "food","einzelpreis": 200,"menge": 2},{"positionId": "b0009000-0000-0000-0000-000000000037","varianteId": 30,"produktName": "Weizen","varianteName": "Russ","kategorie": "beverage","einzelpreis": 300,"menge": 3}],"gesamtPreisCents": 2200,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.produkte-geliefert:v1', 'tisch:9', 32, now() - interval '18 hours 9 minutes', '{"lieferungId": "c0009000-0000-0000-0000-000000000019","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000050","menge": 2},{"positionId": "b0009000-0000-0000-0000-000000000051","menge": 2},{"positionId": "b0009000-0000-0000-0000-000000000052","menge": 4},{"positionId": "b0009000-0000-0000-0000-000000000053","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:9', 33, now() - interval '18 hours 2 minutes', '{"lieferungId": "c0009000-0000-0000-0000-000000000018","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000046","menge": 3},{"positionId": "b0009000-0000-0000-0000-000000000047","menge": 3},{"positionId": "b0009000-0000-0000-0000-000000000048","menge": 4},{"positionId": "b0009000-0000-0000-0000-000000000049","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.bestellung-aufgegeben:v1', 'tisch:9', 34, now() - interval '18 hours 1 minutes', '{"bestellungId": "a0009000-0000-0000-0000-000000000027","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000069","varianteId": 5,"produktName": "Pommes","varianteName": "Groß","kategorie": "food","einzelpreis": 350,"menge": 1},{"positionId": "b0009000-0000-0000-0000-000000000070","varianteId": 5,"produktName": "Pommes","varianteName": "Groß","kategorie": "food","einzelpreis": 350,"menge": 2},{"positionId": "b0009000-0000-0000-0000-000000000071","varianteId": 26,"produktName": "Weizen","varianteName": "Klein 0,3l","kategorie": "beverage","einzelpreis": 300,"menge": 4},{"positionId": "b0009000-0000-0000-0000-000000000072","varianteId": 23,"produktName": "Bier","varianteName": "0,3l","kategorie": "beverage","einzelpreis": 300,"menge": 2}],"gesamtPreisCents": 2850,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.produkte-geliefert:v1', 'tisch:9', 35, now() - interval '18 hours', '{"lieferungId": "c0009000-0000-0000-0000-000000000014","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000035","menge": 3},{"positionId": "b0009000-0000-0000-0000-000000000036","menge": 2},{"positionId": "b0009000-0000-0000-0000-000000000037","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.bestellung-aufgegeben:v1', 'tisch:9', 36, now() - interval '17 hours 59 minutes', '{"bestellungId": "a0009000-0000-0000-0000-000000000021","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000056","varianteId": 12,"produktName": "Grillplatte","varianteName": "Klein","kategorie": "food","einzelpreis": 800,"menge": 3},{"positionId": "b0009000-0000-0000-0000-000000000057","varianteId": 46,"produktName": "Kaffee","varianteName": "Espresso","kategorie": "beverage","einzelpreis": 180,"menge": 4}],"gesamtPreisCents": 3120,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.bestellung-aufgegeben:v1', 'tisch:9', 37, now() - interval '17 hours 58 minutes', '{"bestellungId": "a0009000-0000-0000-0000-000000000022","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000058","varianteId": 5,"produktName": "Pommes","varianteName": "Groß","kategorie": "food","einzelpreis": 350,"menge": 1},{"positionId": "b0009000-0000-0000-0000-000000000059","varianteId": 25,"produktName": "Bier","varianteName": "Maß 1,0l","kategorie": "beverage","einzelpreis": 850,"menge": 1}],"gesamtPreisCents": 1200,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.produkte-geliefert:v1', 'tisch:9', 38, now() - interval '17 hours 55 minutes', '{"lieferungId": "c0009000-0000-0000-0000-000000000013","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000033","menge": 2},{"positionId": "b0009000-0000-0000-0000-000000000034","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.zahlung-registriert:v1', 'tisch:9', 39, now() - interval '17 hours 48 minutes', '{"zahlungId": "d0009000-0000-0000-0000-000000000018","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000050","menge": 2},{"positionId": "b0009000-0000-0000-0000-000000000051","menge": 2},{"positionId": "b0009000-0000-0000-0000-000000000052","menge": 4},{"positionId": "b0009000-0000-0000-0000-000000000053","menge": 3}],"gesamtZahlungCents": 3300,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.produkte-geliefert:v1', 'tisch:9', 40, now() - interval '17 hours 48 minutes', '{"lieferungId": "c0009000-0000-0000-0000-000000000027","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000069","menge": 1},{"positionId": "b0009000-0000-0000-0000-000000000070","menge": 2},{"positionId": "b0009000-0000-0000-0000-000000000071","menge": 4},{"positionId": "b0009000-0000-0000-0000-000000000072","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:9', 41, now() - interval '17 hours 46 minutes', '{"zahlungId": "d0009000-0000-0000-0000-000000000017","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000046","menge": 3},{"positionId": "b0009000-0000-0000-0000-000000000047","menge": 3},{"positionId": "b0009000-0000-0000-0000-000000000048","menge": 4},{"positionId": "b0009000-0000-0000-0000-000000000049","menge": 1}],"gesamtZahlungCents": 5800,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.produkte-geliefert:v1', 'tisch:9', 42, now() - interval '17 hours 45 minutes', '{"lieferungId": "c0009000-0000-0000-0000-000000000022","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000058","menge": 1},{"positionId": "b0009000-0000-0000-0000-000000000059","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.produkte-geliefert:v1', 'tisch:9', 43, now() - interval '17 hours 44 minutes', '{"lieferungId": "c0009000-0000-0000-0000-000000000021","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000056","menge": 3},{"positionId": "b0009000-0000-0000-0000-000000000057","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.zahlung-registriert:v1', 'tisch:9', 44, now() - interval '17 hours 42 minutes', '{"zahlungId": "d0009000-0000-0000-0000-000000000013","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000035","menge": 3},{"positionId": "b0009000-0000-0000-0000-000000000036","menge": 2},{"positionId": "b0009000-0000-0000-0000-000000000037","menge": 3}],"gesamtZahlungCents": 2200,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:9', 45, now() - interval '17 hours 37 minutes', '{"bestellungId": "a0009000-0000-0000-0000-000000000016","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000040","varianteId": 16,"produktName": "Kuchen","varianteName": "Stück","kategorie": "food","einzelpreis": 250,"menge": 1},{"positionId": "b0009000-0000-0000-0000-000000000041","varianteId": 32,"produktName": "Softdrinks","varianteName": "Fanta","kategorie": "beverage","einzelpreis": 280,"menge": 1},{"positionId": "b0009000-0000-0000-0000-000000000042","varianteId": 36,"produktName": "Wasser","varianteName": "Still 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 3}],"gesamtPreisCents": 1130,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.zahlung-registriert:v1', 'tisch:9', 46, now() - interval '17 hours 34 minutes', '{"zahlungId": "d0009000-0000-0000-0000-000000000012","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000033","menge": 2},{"positionId": "b0009000-0000-0000-0000-000000000034","menge": 1}],"gesamtZahlungCents": 1400,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.zahlung-registriert:v1', 'tisch:9', 47, now() - interval '17 hours 32 minutes', '{"zahlungId": "d0009000-0000-0000-0000-000000000020","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000056","menge": 3},{"positionId": "b0009000-0000-0000-0000-000000000057","menge": 4}],"gesamtZahlungCents": 3120,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.produkte-geliefert:v1', 'tisch:9', 48, now() - interval '17 hours 28 minutes', '{"lieferungId": "c0009000-0000-0000-0000-000000000016","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000040","menge": 1},{"positionId": "b0009000-0000-0000-0000-000000000041","menge": 1},{"positionId": "b0009000-0000-0000-0000-000000000042","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:9', 49, now() - interval '17 hours 26 minutes', '{"bestellungId": "a0009000-0000-0000-0000-000000000012","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000031","varianteId": 14,"produktName": "Salat","varianteName": "Gemischter Salat","kategorie": "food","einzelpreis": 550,"menge": 1},{"positionId": "b0009000-0000-0000-0000-000000000032","varianteId": 49,"produktName": "Hugo/Aperol","varianteName": "Aperol Spritz","kategorie": "beverage","einzelpreis": 550,"menge": 3}],"gesamtPreisCents": 2200,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.bestellung-aufgegeben:v1', 'tisch:9', 50, now() - interval '17 hours 26 minutes', '{"bestellungId": "a0009000-0000-0000-0000-000000000024","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000063","varianteId": 16,"produktName": "Kuchen","varianteName": "Stück","kategorie": "food","einzelpreis": 250,"menge": 3},{"positionId": "b0009000-0000-0000-0000-000000000064","varianteId": 34,"produktName": "Softdrinks","varianteName": "Sprite","kategorie": "beverage","einzelpreis": 280,"menge": 1}],"gesamtPreisCents": 1030,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.zahlung-registriert:v1', 'tisch:9', 51, now() - interval '17 hours 24 minutes', '{"zahlungId": "d0009000-0000-0000-0000-000000000021","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000058","menge": 1},{"positionId": "b0009000-0000-0000-0000-000000000059","menge": 1}],"gesamtZahlungCents": 1200,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.zahlung-registriert:v1', 'tisch:9', 52, now() - interval '17 hours 23 minutes', '{"zahlungId": "d0009000-0000-0000-0000-000000000026","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000069","menge": 1},{"positionId": "b0009000-0000-0000-0000-000000000070","menge": 2},{"positionId": "b0009000-0000-0000-0000-000000000071","menge": 4},{"positionId": "b0009000-0000-0000-0000-000000000072","menge": 2}],"gesamtZahlungCents": 2850,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.bestellung-aufgegeben:v1', 'tisch:9', 53, now() - interval '17 hours 19 minutes', '{"bestellungId": "a0009000-0000-0000-0000-000000000017","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000043","varianteId": 6,"produktName": "Flammkuchen","varianteName": "Classic","kategorie": "food","einzelpreis": 600,"menge": 3},{"positionId": "b0009000-0000-0000-0000-000000000044","varianteId": 10,"produktName": "Tagesgericht","varianteName": "Sa: Gulasch mit Spätzle","kategorie": "food","einzelpreis": 1150,"menge": 2},{"positionId": "b0009000-0000-0000-0000-000000000045","varianteId": 45,"produktName": "Kaffee","varianteName": "Tasse","kategorie": "beverage","einzelpreis": 200,"menge": 4}],"gesamtPreisCents": 4900,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.produkte-geliefert:v1', 'tisch:9', 54, now() - interval '17 hours 18 minutes', '{"lieferungId": "c0009000-0000-0000-0000-000000000024","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000063","menge": 3},{"positionId": "b0009000-0000-0000-0000-000000000064","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.produkte-geliefert:v1', 'tisch:9', 55, now() - interval '17 hours 13 minutes', '{"lieferungId": "c0009000-0000-0000-0000-000000000012","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000031","menge": 1},{"positionId": "b0009000-0000-0000-0000-000000000032","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.zahlung-registriert:v1', 'tisch:9', 56, now() - interval '17 hours 5 minutes', '{"zahlungId": "d0009000-0000-0000-0000-000000000015","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000040","menge": 1},{"positionId": "b0009000-0000-0000-0000-000000000041","menge": 1},{"positionId": "b0009000-0000-0000-0000-000000000042","menge": 3}],"gesamtZahlungCents": 1130,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.produkte-geliefert:v1', 'tisch:9', 57, now() - interval '17 hours 5 minutes', '{"lieferungId": "c0009000-0000-0000-0000-000000000017","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000043","menge": 3},{"positionId": "b0009000-0000-0000-0000-000000000044","menge": 2},{"positionId": "b0009000-0000-0000-0000-000000000045","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.zahlung-registriert:v1', 'tisch:9', 58, now() - interval '17 hours 1 minutes', '{"zahlungId": "d0009000-0000-0000-0000-000000000023","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000063","menge": 3},{"positionId": "b0009000-0000-0000-0000-000000000064","menge": 1}],"gesamtZahlungCents": 1030,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.bestellung-aufgegeben:v1', 'tisch:9', 59, now() - interval '16 hours 59 minutes', '{"bestellungId": "a0009000-0000-0000-0000-000000000028","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000073","varianteId": 19,"produktName": "Waffeln","varianteName": "mit Nutella","kategorie": "food","einzelpreis": 400,"menge": 2},{"positionId": "b0009000-0000-0000-0000-000000000074","varianteId": 46,"produktName": "Kaffee","varianteName": "Espresso","kategorie": "beverage","einzelpreis": 180,"menge": 4},{"positionId": "b0009000-0000-0000-0000-000000000075","varianteId": 33,"produktName": "Softdrinks","varianteName": "Spezi","kategorie": "beverage","einzelpreis": 280,"menge": 2}],"gesamtPreisCents": 2080,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.bestellung-aufgegeben:v1', 'tisch:9', 60, now() - interval '16 hours 56 minutes', '{"bestellungId": "a0009000-0000-0000-0000-000000000015","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000038","varianteId": 16,"produktName": "Kuchen","varianteName": "Stück","kategorie": "food","einzelpreis": 250,"menge": 1},{"positionId": "b0009000-0000-0000-0000-000000000039","varianteId": 46,"produktName": "Kaffee","varianteName": "Espresso","kategorie": "beverage","einzelpreis": 180,"menge": 4}],"gesamtPreisCents": 970,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.zahlung-registriert:v1', 'tisch:9', 61, now() - interval '16 hours 49 minutes', '{"zahlungId": "d0009000-0000-0000-0000-000000000011","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000031","menge": 1},{"positionId": "b0009000-0000-0000-0000-000000000032","menge": 3}],"gesamtZahlungCents": 2200,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.produkte-geliefert:v1', 'tisch:9', 62, now() - interval '16 hours 48 minutes', '{"lieferungId": "c0009000-0000-0000-0000-000000000028","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000073","menge": 2},{"positionId": "b0009000-0000-0000-0000-000000000074","menge": 4},{"positionId": "b0009000-0000-0000-0000-000000000075","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.bestellung-aufgegeben:v1', 'tisch:9', 63, now() - interval '16 hours 47 minutes', '{"bestellungId": "a0009000-0000-0000-0000-000000000025","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000065","varianteId": 21,"produktName": "Brezel","varianteName": "mit Butter","kategorie": "food","einzelpreis": 300,"menge": 3},{"positionId": "b0009000-0000-0000-0000-000000000066","varianteId": 26,"produktName": "Weizen","varianteName": "Klein 0,3l","kategorie": "beverage","einzelpreis": 300,"menge": 2}],"gesamtPreisCents": 1500,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.zahlung-registriert:v1', 'tisch:9', 64, now() - interval '16 hours 44 minutes', '{"zahlungId": "d0009000-0000-0000-0000-000000000016","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000043","menge": 3},{"positionId": "b0009000-0000-0000-0000-000000000044","menge": 2},{"positionId": "b0009000-0000-0000-0000-000000000045","menge": 4}],"gesamtZahlungCents": 4900,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:9', 65, now() - interval '16 hours 44 minutes', '{"bestellungId": "a0009000-0000-0000-0000-000000000026","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000067","varianteId": 21,"produktName": "Brezel","varianteName": "mit Butter","kategorie": "food","einzelpreis": 300,"menge": 1},{"positionId": "b0009000-0000-0000-0000-000000000068","varianteId": 26,"produktName": "Weizen","varianteName": "Klein 0,3l","kategorie": "beverage","einzelpreis": 300,"menge": 4}],"gesamtPreisCents": 1500,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.produkte-geliefert:v1', 'tisch:9', 66, now() - interval '16 hours 42 minutes', '{"lieferungId": "c0009000-0000-0000-0000-000000000015","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000038","menge": 1},{"positionId": "b0009000-0000-0000-0000-000000000039","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:9', 67, now() - interval '16 hours 42 minutes', '{"bestellungId": "a0009000-0000-0000-0000-000000000023","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000060","varianteId": 20,"produktName": "Brezel","varianteName": "Normal","kategorie": "food","einzelpreis": 200,"menge": 3},{"positionId": "b0009000-0000-0000-0000-000000000061","varianteId": 1,"produktName": "Bratwurst","varianteName": "Normal","kategorie": "food","einzelpreis": 350,"menge": 1},{"positionId": "b0009000-0000-0000-0000-000000000062","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 1}],"gesamtPreisCents": 1400,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:9', 68, now() - interval '16 hours 41 minutes', '{"bestellungId": "a0009000-0000-0000-0000-000000000020","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000054","varianteId": 10,"produktName": "Tagesgericht","varianteName": "Sa: Gulasch mit Spätzle","kategorie": "food","einzelpreis": 1150,"menge": 1},{"positionId": "b0009000-0000-0000-0000-000000000055","varianteId": 47,"produktName": "Tee","varianteName": "Verschiedene Sorten","kategorie": "beverage","einzelpreis": 200,"menge": 2}],"gesamtPreisCents": 1550,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.produkte-geliefert:v1', 'tisch:9', 69, now() - interval '16 hours 36 minutes', '{"lieferungId": "c0009000-0000-0000-0000-000000000026","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000067","menge": 1},{"positionId": "b0009000-0000-0000-0000-000000000068","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.produkte-geliefert:v1', 'tisch:9', 70, now() - interval '16 hours 34 minutes', '{"lieferungId": "c0009000-0000-0000-0000-000000000025","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000065","menge": 3},{"positionId": "b0009000-0000-0000-0000-000000000066","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:9', 71, now() - interval '16 hours 30 minutes', '{"lieferungId": "c0009000-0000-0000-0000-000000000023","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000060","menge": 3},{"positionId": "b0009000-0000-0000-0000-000000000061","menge": 1},{"positionId": "b0009000-0000-0000-0000-000000000062","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.zahlung-registriert:v1', 'tisch:9', 72, now() - interval '16 hours 27 minutes', '{"zahlungId": "d0009000-0000-0000-0000-000000000014","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000038","menge": 1},{"positionId": "b0009000-0000-0000-0000-000000000039","menge": 4}],"gesamtZahlungCents": 970,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:9', 73, now() - interval '16 hours 27 minutes', '{"lieferungId": "c0009000-0000-0000-0000-000000000020","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000054","menge": 1},{"positionId": "b0009000-0000-0000-0000-000000000055","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.zahlung-registriert:v1', 'tisch:9', 74, now() - interval '16 hours 27 minutes', '{"zahlungId": "d0009000-0000-0000-0000-000000000027","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000073","menge": 2},{"positionId": "b0009000-0000-0000-0000-000000000074","menge": 4},{"positionId": "b0009000-0000-0000-0000-000000000075","menge": 2}],"gesamtZahlungCents": 2080,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.zahlung-registriert:v1', 'tisch:9', 75, now() - interval '16 hours 19 minutes', '{"zahlungId": "d0009000-0000-0000-0000-000000000025","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000067","menge": 1},{"positionId": "b0009000-0000-0000-0000-000000000068","menge": 4}],"gesamtZahlungCents": 1500,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.zahlung-registriert:v1', 'tisch:9', 76, now() - interval '16 hours 18 minutes', '{"zahlungId": "d0009000-0000-0000-0000-000000000024","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000065","menge": 3},{"positionId": "b0009000-0000-0000-0000-000000000066","menge": 2}],"gesamtZahlungCents": 1500,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:9', 77, now() - interval '16 hours 13 minutes', '{"zahlungId": "d0009000-0000-0000-0000-000000000019","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000054","menge": 1},{"positionId": "b0009000-0000-0000-0000-000000000055","menge": 2}],"gesamtZahlungCents": 1550,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:9', 78, now() - interval '16 hours 11 minutes', '{"zahlungId": "d0009000-0000-0000-0000-000000000022","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000060","menge": 3},{"positionId": "b0009000-0000-0000-0000-000000000061","menge": 1},{"positionId": "b0009000-0000-0000-0000-000000000062","menge": 1}],"gesamtZahlungCents": 1400,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:9', 79, now() - interval '4 hours 20 minutes', '{"bestellungId": "a0009000-0000-0000-0000-000000000010","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000024","varianteId": 11,"produktName": "Tagesgericht","varianteName": "So: Hähnchen mit Reis","kategorie": "food","einzelpreis": 1050,"menge": 3},{"positionId": "b0009000-0000-0000-0000-000000000025","varianteId": 14,"produktName": "Salat","varianteName": "Gemischter Salat","kategorie": "food","einzelpreis": 550,"menge": 2},{"positionId": "b0009000-0000-0000-0000-000000000026","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 4},{"positionId": "b0009000-0000-0000-0000-000000000027","varianteId": 40,"produktName": "Saftschorle","varianteName": "Johannisbeerschorle 0,5l","kategorie": "beverage","einzelpreis": 300,"menge": 2}],"gesamtPreisCents": 6650,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.produkte-geliefert:v1', 'tisch:9', 80, now() - interval '4 hours 11 minutes', '{"lieferungId": "c0009000-0000-0000-0000-000000000010","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000024","menge": 3},{"positionId": "b0009000-0000-0000-0000-000000000025","menge": 2},{"positionId": "b0009000-0000-0000-0000-000000000026","menge": 4},{"positionId": "b0009000-0000-0000-0000-000000000027","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:9', 81, now() - interval '3 hours 48 minutes', '{"bestellungId": "a0009000-0000-0000-0000-000000000011","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000028","varianteId": 20,"produktName": "Brezel","varianteName": "Normal","kategorie": "food","einzelpreis": 200,"menge": 4},{"positionId": "b0009000-0000-0000-0000-000000000029","varianteId": 21,"produktName": "Brezel","varianteName": "mit Butter","kategorie": "food","einzelpreis": 300,"menge": 2},{"positionId": "b0009000-0000-0000-0000-000000000030","varianteId": 45,"produktName": "Kaffee","varianteName": "Tasse","kategorie": "beverage","einzelpreis": 200,"menge": 3}],"gesamtPreisCents": 2000,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.produkte-geliefert:v1', 'tisch:9', 82, now() - interval '3 hours 34 minutes', '{"lieferungId": "c0009000-0000-0000-0000-000000000011","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000028","menge": 4},{"positionId": "b0009000-0000-0000-0000-000000000029","menge": 2},{"positionId": "b0009000-0000-0000-0000-000000000030","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.zahlung-registriert:v1', 'tisch:9', 83, now() - interval '3 hours 24 minutes', '{"zahlungId": "d0009000-0000-0000-0000-000000000010","positionen": [{"positionId": "b0009000-0000-0000-0000-000000000024","menge": 3},{"positionId": "b0009000-0000-0000-0000-000000000025","menge": 2},{"positionId": "b0009000-0000-0000-0000-000000000026","menge": 4},{"positionId": "b0009000-0000-0000-0000-000000000027","menge": 2}],"gesamtZahlungCents": 6650,"kommentar": "Essen bezahlt"}');

-- ─────────────────────────────────────────────────────────────────────────────
-- TISCH 10: Stornierung nach Bezahlung — Tag 2, negativer Saldo (12 Events)
-- ─────────────────────────────────────────────────────────────────────────────

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.bestellung-aufgegeben:v1', 'tisch:10', 1, now() - interval '23 hours 20 minutes', '{"bestellungId": "a0010000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0010000-0000-0000-0000-000000000001","varianteId": 1,"produktName": "Bratwurst","varianteName": "Normal","kategorie": "food","einzelpreis": 350,"menge": 3},{"positionId": "b0010000-0000-0000-0000-000000000002","varianteId": 5,"produktName": "Pommes","varianteName": "Groß","kategorie": "food","einzelpreis": 350,"menge": 2},{"positionId": "b0010000-0000-0000-0000-000000000003","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 4},{"positionId": "b0010000-0000-0000-0000-000000000004","varianteId": 33,"produktName": "Softdrinks","varianteName": "Spezi","kategorie": "beverage","einzelpreis": 280,"menge": 2}],"gesamtPreisCents": 4110,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.produkte-geliefert:v1', 'tisch:10', 2, now() - interval '23 hours 8 minutes', '{"lieferungId": "c0010000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0010000-0000-0000-0000-000000000001","menge": 3},{"positionId": "b0010000-0000-0000-0000-000000000002","menge": 2},{"positionId": "b0010000-0000-0000-0000-000000000003","menge": 4},{"positionId": "b0010000-0000-0000-0000-000000000004","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.zahlung-registriert:v1', 'tisch:10', 3, now() - interval '22 hours 48 minutes', '{"zahlungId": "d0010000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0010000-0000-0000-0000-000000000001","menge": 3},{"positionId": "b0010000-0000-0000-0000-000000000002","menge": 2},{"positionId": "b0010000-0000-0000-0000-000000000003","menge": 4},{"positionId": "b0010000-0000-0000-0000-000000000004","menge": 2}],"gesamtZahlungCents": 4110,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(2, 'Thomas Müller', 'tisch.produkte-storniert:v1', 'tisch:10', 4, now() - interval '22 hours 33 minutes', '{"stornierungId": "e0010000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0010000-0000-0000-0000-000000000001","menge": 2}],"gesamtStornierungCents": 700,"kommentar": "Bratwurst war nicht in Ordnung, Erstattung"}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.produkte-storniert:v1', 'tisch:10', 5, now() - interval '22 hours 23 minutes', '{"stornierungId": "e0010000-0000-0000-0000-000000000002","positionen": [{"positionId": "b0010000-0000-0000-0000-000000000002","menge": 1}],"gesamtStornierungCents": 350,"kommentar": "Pommes auch kalt — Reklamation"}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.bestellung-aufgegeben:v1', 'tisch:10', 6, now() - interval '22 hours 15 minutes', '{"bestellungId": "a0010000-0000-0000-0000-000000000002","positionen": [{"positionId": "b0010000-0000-0000-0000-000000000005","varianteId": 6,"produktName": "Flammkuchen","varianteName": "Classic","kategorie": "food","einzelpreis": 600,"menge": 1},{"positionId": "b0010000-0000-0000-0000-000000000006","varianteId": 36,"produktName": "Wasser","varianteName": "Still 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 2}],"gesamtPreisCents": 1000,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.produkte-geliefert:v1', 'tisch:10', 7, now() - interval '22 hours 6 minutes', '{"lieferungId": "c0010000-0000-0000-0000-000000000002","positionen": [{"positionId": "b0010000-0000-0000-0000-000000000005","menge": 1},{"positionId": "b0010000-0000-0000-0000-000000000006","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.zahlung-registriert:v1', 'tisch:10', 8, now() - interval '21 hours 54 minutes', '{"zahlungId": "d0010000-0000-0000-0000-000000000002","positionen": [{"positionId": "b0010000-0000-0000-0000-000000000005","menge": 1},{"positionId": "b0010000-0000-0000-0000-000000000006","menge": 2}],"gesamtZahlungCents": 1000,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.bestellung-aufgegeben:v1', 'tisch:10', 9, now() - interval '21 hours 33 minutes', '{"bestellungId": "a0010000-0000-0000-0000-000000000003","positionen": [{"positionId": "b0010000-0000-0000-0000-000000000007","varianteId": 23,"produktName": "Bier","varianteName": "0,3l","kategorie": "beverage","einzelpreis": 300,"menge": 3},{"positionId": "b0010000-0000-0000-0000-000000000008","varianteId": 20,"produktName": "Brezel","varianteName": "Normal","kategorie": "food","einzelpreis": 200,"menge": 2}],"gesamtPreisCents": 1300,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.produkte-geliefert:v1', 'tisch:10', 10, now() - interval '21 hours 25 minutes', '{"lieferungId": "c0010000-0000-0000-0000-000000000003","positionen": [{"positionId": "b0010000-0000-0000-0000-000000000007","menge": 3},{"positionId": "b0010000-0000-0000-0000-000000000008","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.produkte-storniert:v1', 'tisch:10', 11, now() - interval '21 hours 13 minutes', '{"stornierungId": "e0010000-0000-0000-0000-000000000003","positionen": [{"positionId": "b0010000-0000-0000-0000-000000000007","menge": 1}],"gesamtStornierungCents": 300,"kommentar": "Falsches Getränk gebracht"}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.zahlung-registriert:v1', 'tisch:10', 12, now() - interval '21 hours 8 minutes', '{"zahlungId": "d0010000-0000-0000-0000-000000000003","positionen": [{"positionId": "b0010000-0000-0000-0000-000000000007","menge": 2},{"positionId": "b0010000-0000-0000-0000-000000000008","menge": 2}],"gesamtZahlungCents": 1000,"kommentar": ""}');

-- ─────────────────────────────────────────────────────────────────────────────
-- TISCH 11: Gerade dazugekommen — Tag 3, minimal (3 Events)
-- ─────────────────────────────────────────────────────────────────────────────

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.bestellung-aufgegeben:v1', 'tisch:11', 1, now() - interval '1 hours 30 minutes', '{"bestellungId": "a0011000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0011000-0000-0000-0000-000000000001","varianteId": 11,"produktName": "Tagesgericht","varianteName": "So: Hähnchen mit Reis","kategorie": "food","einzelpreis": 1050,"menge": 2},{"positionId": "b0011000-0000-0000-0000-000000000002","varianteId": 5,"produktName": "Pommes","varianteName": "Groß","kategorie": "food","einzelpreis": 350,"menge": 2},{"positionId": "b0011000-0000-0000-0000-000000000003","varianteId": 31,"produktName": "Softdrinks","varianteName": "Cola","kategorie": "beverage","einzelpreis": 280,"menge": 2},{"positionId": "b0011000-0000-0000-0000-000000000004","varianteId": 37,"produktName": "Wasser","varianteName": "Medium 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 1}],"gesamtPreisCents": 3560,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.produkte-geliefert:v1', 'tisch:11', 2, now() - interval '1 hours 19 minutes', '{"lieferungId": "c0011000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0011000-0000-0000-0000-000000000001","menge": 2},{"positionId": "b0011000-0000-0000-0000-000000000002","menge": 2},{"positionId": "b0011000-0000-0000-0000-000000000003","menge": 2},{"positionId": "b0011000-0000-0000-0000-000000000004","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.bestellung-aufgegeben:v1', 'tisch:11', 3, now() - interval '1 hours 10 minutes', '{"bestellungId": "a0011000-0000-0000-0000-000000000002","positionen": [{"positionId": "b0011000-0000-0000-0000-000000000005","varianteId": 16,"produktName": "Kuchen","varianteName": "Stück","kategorie": "food","einzelpreis": 250,"menge": 2},{"positionId": "b0011000-0000-0000-0000-000000000006","varianteId": 45,"produktName": "Kaffee","varianteName": "Tasse","kategorie": "beverage","einzelpreis": 200,"menge": 2}],"gesamtPreisCents": 900,"kommentar": ""}');

-- ─────────────────────────────────────────────────────────────────────────────
-- TISCH 12: Jugendgruppe — Tag 1+2, viel günstiges Essen (63 Events)
-- ─────────────────────────────────────────────────────────────────────────────

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.bestellung-aufgegeben:v1', 'tisch:12', 1, now() - interval '46 hours 40 minutes', '{"bestellungId": "a0012000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000001","varianteId": 1,"produktName": "Bratwurst","varianteName": "Normal","kategorie": "food","einzelpreis": 350,"menge": 5},{"positionId": "b0012000-0000-0000-0000-000000000002","varianteId": 4,"produktName": "Pommes","varianteName": "Klein","kategorie": "food","einzelpreis": 250,"menge": 5},{"positionId": "b0012000-0000-0000-0000-000000000003","varianteId": 23,"produktName": "Bier","varianteName": "0,3l","kategorie": "beverage","einzelpreis": 300,"menge": 6},{"positionId": "b0012000-0000-0000-0000-000000000004","varianteId": 33,"produktName": "Softdrinks","varianteName": "Spezi","kategorie": "beverage","einzelpreis": 280,"menge": 4}],"gesamtPreisCents": 5920,"kommentar": "Jugendgruppe TSV"}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.produkte-geliefert:v1', 'tisch:12', 2, now() - interval '46 hours 27 minutes', '{"lieferungId": "c0012000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000001","menge": 5},{"positionId": "b0012000-0000-0000-0000-000000000002","menge": 5},{"positionId": "b0012000-0000-0000-0000-000000000003","menge": 6},{"positionId": "b0012000-0000-0000-0000-000000000004","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.zahlung-registriert:v1', 'tisch:12', 3, now() - interval '46 hours 7 minutes', '{"zahlungId": "d0012000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000001","menge": 5},{"positionId": "b0012000-0000-0000-0000-000000000002","menge": 5},{"positionId": "b0012000-0000-0000-0000-000000000003","menge": 6},{"positionId": "b0012000-0000-0000-0000-000000000004","menge": 4}],"gesamtZahlungCents": 5920,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.bestellung-aufgegeben:v1', 'tisch:12', 4, now() - interval '45 hours 50 minutes', '{"bestellungId": "a0012000-0000-0000-0000-000000000002","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000005","varianteId": 3,"produktName": "Bratwurst","varianteName": "Currywurst","kategorie": "food","einzelpreis": 450,"menge": 4},{"positionId": "b0012000-0000-0000-0000-000000000006","varianteId": 5,"produktName": "Pommes","varianteName": "Groß","kategorie": "food","einzelpreis": 350,"menge": 3},{"positionId": "b0012000-0000-0000-0000-000000000007","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 4},{"positionId": "b0012000-0000-0000-0000-000000000008","varianteId": 34,"produktName": "Softdrinks","varianteName": "Sprite","kategorie": "beverage","einzelpreis": 280,"menge": 2}],"gesamtPreisCents": 5210,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.produkte-geliefert:v1', 'tisch:12', 5, now() - interval '45 hours 39 minutes', '{"lieferungId": "c0012000-0000-0000-0000-000000000002","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000005","menge": 4},{"positionId": "b0012000-0000-0000-0000-000000000006","menge": 3},{"positionId": "b0012000-0000-0000-0000-000000000007","menge": 4},{"positionId": "b0012000-0000-0000-0000-000000000008","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.zahlung-registriert:v1', 'tisch:12', 6, now() - interval '45 hours 18 minutes', '{"zahlungId": "d0012000-0000-0000-0000-000000000002","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000005","menge": 4},{"positionId": "b0012000-0000-0000-0000-000000000006","menge": 3},{"positionId": "b0012000-0000-0000-0000-000000000007","menge": 4},{"positionId": "b0012000-0000-0000-0000-000000000008","menge": 2}],"gesamtZahlungCents": 5210,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:12', 7, now() - interval '45 hours 2 minutes', '{"bestellungId": "a0012000-0000-0000-0000-000000000003","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000009","varianteId": 20,"produktName": "Brezel","varianteName": "Normal","kategorie": "food","einzelpreis": 200,"menge": 6},{"positionId": "b0012000-0000-0000-0000-000000000010","varianteId": 21,"produktName": "Brezel","varianteName": "mit Butter","kategorie": "food","einzelpreis": 300,"menge": 3},{"positionId": "b0012000-0000-0000-0000-000000000011","varianteId": 31,"produktName": "Softdrinks","varianteName": "Cola","kategorie": "beverage","einzelpreis": 280,"menge": 6}],"gesamtPreisCents": 3780,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:12', 8, now() - interval '44 hours 48 minutes', '{"lieferungId": "c0012000-0000-0000-0000-000000000003","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000009","menge": 6},{"positionId": "b0012000-0000-0000-0000-000000000010","menge": 3},{"positionId": "b0012000-0000-0000-0000-000000000011","menge": 6}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:12', 9, now() - interval '44 hours 35 minutes', '{"zahlungId": "d0012000-0000-0000-0000-000000000003","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000009","menge": 6},{"positionId": "b0012000-0000-0000-0000-000000000010","menge": 3},{"positionId": "b0012000-0000-0000-0000-000000000011","menge": 6}],"gesamtZahlungCents": 3780,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.bestellung-aufgegeben:v1', 'tisch:12', 10, now() - interval '44 hours 16 minutes', '{"bestellungId": "a0012000-0000-0000-0000-000000000004","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000012","varianteId": 17,"produktName": "Waffeln","varianteName": "mit Puderzucker","kategorie": "food","einzelpreis": 300,"menge": 4},{"positionId": "b0012000-0000-0000-0000-000000000013","varianteId": 18,"produktName": "Waffeln","varianteName": "mit Sahne","kategorie": "food","einzelpreis": 350,"menge": 3},{"positionId": "b0012000-0000-0000-0000-000000000014","varianteId": 32,"produktName": "Softdrinks","varianteName": "Fanta","kategorie": "beverage","einzelpreis": 280,"menge": 4}],"gesamtPreisCents": 3370,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.produkte-geliefert:v1', 'tisch:12', 11, now() - interval '44 hours 2 minutes', '{"lieferungId": "c0012000-0000-0000-0000-000000000004","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000012","menge": 4},{"positionId": "b0012000-0000-0000-0000-000000000013","menge": 3},{"positionId": "b0012000-0000-0000-0000-000000000014","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.zahlung-registriert:v1', 'tisch:12', 12, now() - interval '43 hours 42 minutes', '{"zahlungId": "d0012000-0000-0000-0000-000000000004","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000012","menge": 4},{"positionId": "b0012000-0000-0000-0000-000000000013","menge": 3},{"positionId": "b0012000-0000-0000-0000-000000000014","menge": 4}],"gesamtZahlungCents": 3370,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:12', 13, now() - interval '43 hours 25 minutes', '{"bestellungId": "a0012000-0000-0000-0000-000000000005","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000015","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 6},{"positionId": "b0012000-0000-0000-0000-000000000016","varianteId": 23,"produktName": "Bier","varianteName": "0,3l","kategorie": "beverage","einzelpreis": 300,"menge": 4}],"gesamtPreisCents": 3900,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:12', 14, now() - interval '43 hours 14 minutes', '{"lieferungId": "c0012000-0000-0000-0000-000000000005","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000015","menge": 6},{"positionId": "b0012000-0000-0000-0000-000000000016","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:12', 15, now() - interval '43 hours 2 minutes', '{"zahlungId": "d0012000-0000-0000-0000-000000000005","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000015","menge": 6},{"positionId": "b0012000-0000-0000-0000-000000000016","menge": 4}],"gesamtZahlungCents": 3900,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:12', 16, now() - interval '25 hours 50 minutes', '{"bestellungId": "a0012000-0000-0000-0000-000000000006","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000017","varianteId": 10,"produktName": "Tagesgericht","varianteName": "Sa: Gulasch mit Spätzle","kategorie": "food","einzelpreis": 1150,"menge": 5},{"positionId": "b0012000-0000-0000-0000-000000000018","varianteId": 5,"produktName": "Pommes","varianteName": "Groß","kategorie": "food","einzelpreis": 350,"menge": 5},{"positionId": "b0012000-0000-0000-0000-000000000019","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 6},{"positionId": "b0012000-0000-0000-0000-000000000020","varianteId": 35,"produktName": "Softdrinks","varianteName": "Mezzo Mix","kategorie": "beverage","einzelpreis": 280,"menge": 4}],"gesamtPreisCents": 11320,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:12', 17, now() - interval '25 hours 39 minutes', '{"lieferungId": "c0012000-0000-0000-0000-000000000006","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000017","menge": 5},{"positionId": "b0012000-0000-0000-0000-000000000018","menge": 5},{"positionId": "b0012000-0000-0000-0000-000000000019","menge": 6},{"positionId": "b0012000-0000-0000-0000-000000000020","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:12', 18, now() - interval '25 hours 21 minutes', '{"zahlungId": "d0012000-0000-0000-0000-000000000006","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000017","menge": 5},{"positionId": "b0012000-0000-0000-0000-000000000018","menge": 5},{"positionId": "b0012000-0000-0000-0000-000000000019","menge": 6},{"positionId": "b0012000-0000-0000-0000-000000000020","menge": 4}],"gesamtZahlungCents": 11320,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.bestellung-aufgegeben:v1', 'tisch:12', 19, now() - interval '25 hours', '{"bestellungId": "a0012000-0000-0000-0000-000000000007","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000021","varianteId": 1,"produktName": "Bratwurst","varianteName": "Normal","kategorie": "food","einzelpreis": 350,"menge": 4},{"positionId": "b0012000-0000-0000-0000-000000000022","varianteId": 4,"produktName": "Pommes","varianteName": "Klein","kategorie": "food","einzelpreis": 250,"menge": 4},{"positionId": "b0012000-0000-0000-0000-000000000023","varianteId": 23,"produktName": "Bier","varianteName": "0,3l","kategorie": "beverage","einzelpreis": 300,"menge": 6},{"positionId": "b0012000-0000-0000-0000-000000000024","varianteId": 33,"produktName": "Softdrinks","varianteName": "Spezi","kategorie": "beverage","einzelpreis": 280,"menge": 3}],"gesamtPreisCents": 5040,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.produkte-geliefert:v1', 'tisch:12', 20, now() - interval '24 hours 45 minutes', '{"lieferungId": "c0012000-0000-0000-0000-000000000007","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000021","menge": 4},{"positionId": "b0012000-0000-0000-0000-000000000022","menge": 4},{"positionId": "b0012000-0000-0000-0000-000000000023","menge": 6},{"positionId": "b0012000-0000-0000-0000-000000000024","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.zahlung-registriert:v1', 'tisch:12', 21, now() - interval '24 hours 24 minutes', '{"zahlungId": "d0012000-0000-0000-0000-000000000007","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000021","menge": 4},{"positionId": "b0012000-0000-0000-0000-000000000022","menge": 4},{"positionId": "b0012000-0000-0000-0000-000000000023","menge": 6},{"positionId": "b0012000-0000-0000-0000-000000000024","menge": 3}],"gesamtZahlungCents": 5040,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:12', 22, now() - interval '24 hours 7 minutes', '{"bestellungId": "a0012000-0000-0000-0000-000000000008","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000025","varianteId": 19,"produktName": "Waffeln","varianteName": "mit Nutella","kategorie": "food","einzelpreis": 400,"menge": 5},{"positionId": "b0012000-0000-0000-0000-000000000026","varianteId": 16,"produktName": "Kuchen","varianteName": "Stück","kategorie": "food","einzelpreis": 250,"menge": 4},{"positionId": "b0012000-0000-0000-0000-000000000027","varianteId": 31,"produktName": "Softdrinks","varianteName": "Cola","kategorie": "beverage","einzelpreis": 280,"menge": 5}],"gesamtPreisCents": 4400,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:12', 23, now() - interval '23 hours 56 minutes', '{"lieferungId": "c0012000-0000-0000-0000-000000000008","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000025","menge": 5},{"positionId": "b0012000-0000-0000-0000-000000000026","menge": 4},{"positionId": "b0012000-0000-0000-0000-000000000027","menge": 5}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:12', 24, now() - interval '23 hours 36 minutes', '{"zahlungId": "d0012000-0000-0000-0000-000000000008","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000025","menge": 5},{"positionId": "b0012000-0000-0000-0000-000000000026","menge": 4},{"positionId": "b0012000-0000-0000-0000-000000000027","menge": 5}],"gesamtZahlungCents": 4400,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.bestellung-aufgegeben:v1', 'tisch:12', 25, now() - interval '23 hours 17 minutes', '{"bestellungId": "a0012000-0000-0000-0000-000000000009","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000028","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 6},{"positionId": "b0012000-0000-0000-0000-000000000029","varianteId": 28,"produktName": "Weizen","varianteName": "Colaweizen Klein","kategorie": "beverage","einzelpreis": 300,"menge": 4}],"gesamtPreisCents": 3900,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.produkte-geliefert:v1', 'tisch:12', 26, now() - interval '23 hours 5 minutes', '{"lieferungId": "c0012000-0000-0000-0000-000000000009","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000028","menge": 6},{"positionId": "b0012000-0000-0000-0000-000000000029","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.zahlung-registriert:v1', 'tisch:12', 27, now() - interval '22 hours 44 minutes', '{"zahlungId": "d0012000-0000-0000-0000-000000000009","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000028","menge": 6},{"positionId": "b0012000-0000-0000-0000-000000000029","menge": 4}],"gesamtZahlungCents": 3900,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.bestellung-aufgegeben:v1', 'tisch:12', 28, now() - interval '18 hours 11 minutes', '{"bestellungId": "a0012000-0000-0000-0000-000000000020","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000059","varianteId": 10,"produktName": "Tagesgericht","varianteName": "Sa: Gulasch mit Spätzle","kategorie": "food","einzelpreis": 1150,"menge": 1},{"positionId": "b0012000-0000-0000-0000-000000000060","varianteId": 49,"produktName": "Hugo/Aperol","varianteName": "Aperol Spritz","kategorie": "beverage","einzelpreis": 550,"menge": 3},{"positionId": "b0012000-0000-0000-0000-000000000061","varianteId": 31,"produktName": "Softdrinks","varianteName": "Cola","kategorie": "beverage","einzelpreis": 280,"menge": 3}],"gesamtPreisCents": 3640,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.produkte-geliefert:v1', 'tisch:12', 29, now() - interval '18 hours 2 minutes', '{"lieferungId": "c0012000-0000-0000-0000-000000000020","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000059","menge": 1},{"positionId": "b0012000-0000-0000-0000-000000000060","menge": 3},{"positionId": "b0012000-0000-0000-0000-000000000061","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:12', 30, now() - interval '18 hours', '{"bestellungId": "a0012000-0000-0000-0000-000000000013","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000038","varianteId": 16,"produktName": "Kuchen","varianteName": "Stück","kategorie": "food","einzelpreis": 250,"menge": 3},{"positionId": "b0012000-0000-0000-0000-000000000039","varianteId": 41,"produktName": "Saftschorle","varianteName": "Rhabarberschorle 0,5l","kategorie": "beverage","einzelpreis": 350,"menge": 1},{"positionId": "b0012000-0000-0000-0000-000000000040","varianteId": 27,"produktName": "Weizen","varianteName": "Groß 0,5l","kategorie": "beverage","einzelpreis": 400,"menge": 1}],"gesamtPreisCents": 1500,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.bestellung-aufgegeben:v1', 'tisch:12', 31, now() - interval '17 hours 57 minutes', '{"bestellungId": "a0012000-0000-0000-0000-000000000018","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000053","varianteId": 13,"produktName": "Grillplatte","varianteName": "Groß","kategorie": "food","einzelpreis": 1400,"menge": 2},{"positionId": "b0012000-0000-0000-0000-000000000054","varianteId": 15,"produktName": "Salat","varianteName": "Caesar Salat","kategorie": "food","einzelpreis": 650,"menge": 1},{"positionId": "b0012000-0000-0000-0000-000000000055","varianteId": 48,"produktName": "Hugo/Aperol","varianteName": "Hugo","kategorie": "beverage","einzelpreis": 550,"menge": 4},{"positionId": "b0012000-0000-0000-0000-000000000056","varianteId": 42,"produktName": "Wein","varianteName": "Weißwein 0,2l","kategorie": "beverage","einzelpreis": 400,"menge": 4}],"gesamtPreisCents": 7250,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.bestellung-aufgegeben:v1', 'tisch:12', 32, now() - interval '17 hours 53 minutes', '{"bestellungId": "a0012000-0000-0000-0000-000000000021","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000062","varianteId": 3,"produktName": "Bratwurst","varianteName": "Currywurst","kategorie": "food","einzelpreis": 450,"menge": 1},{"positionId": "b0012000-0000-0000-0000-000000000063","varianteId": 14,"produktName": "Salat","varianteName": "Gemischter Salat","kategorie": "food","einzelpreis": 550,"menge": 1},{"positionId": "b0012000-0000-0000-0000-000000000064","varianteId": 35,"produktName": "Softdrinks","varianteName": "Mezzo Mix","kategorie": "beverage","einzelpreis": 280,"menge": 3}],"gesamtPreisCents": 1840,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.produkte-geliefert:v1', 'tisch:12', 33, now() - interval '17 hours 50 minutes', '{"lieferungId": "c0012000-0000-0000-0000-000000000013","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000038","menge": 3},{"positionId": "b0012000-0000-0000-0000-000000000039","menge": 1},{"positionId": "b0012000-0000-0000-0000-000000000040","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.bestellung-aufgegeben:v1', 'tisch:12', 34, now() - interval '17 hours 49 minutes', '{"bestellungId": "a0012000-0000-0000-0000-000000000010","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000030","varianteId": 18,"produktName": "Waffeln","varianteName": "mit Sahne","kategorie": "food","einzelpreis": 350,"menge": 2},{"positionId": "b0012000-0000-0000-0000-000000000031","varianteId": 32,"produktName": "Softdrinks","varianteName": "Fanta","kategorie": "beverage","einzelpreis": 280,"menge": 1}],"gesamtPreisCents": 980,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:12', 35, now() - interval '17 hours 47 minutes', '{"bestellungId": "a0012000-0000-0000-0000-000000000012","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000035","varianteId": 3,"produktName": "Bratwurst","varianteName": "Currywurst","kategorie": "food","einzelpreis": 450,"menge": 1},{"positionId": "b0012000-0000-0000-0000-000000000036","varianteId": 25,"produktName": "Bier","varianteName": "Maß 1,0l","kategorie": "beverage","einzelpreis": 850,"menge": 4},{"positionId": "b0012000-0000-0000-0000-000000000037","varianteId": 49,"produktName": "Hugo/Aperol","varianteName": "Aperol Spritz","kategorie": "beverage","einzelpreis": 550,"menge": 3}],"gesamtPreisCents": 5500,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.bestellung-aufgegeben:v1', 'tisch:12', 36, now() - interval '17 hours 47 minutes', '{"bestellungId": "a0012000-0000-0000-0000-000000000017","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000050","varianteId": 3,"produktName": "Bratwurst","varianteName": "Currywurst","kategorie": "food","einzelpreis": 450,"menge": 1},{"positionId": "b0012000-0000-0000-0000-000000000051","varianteId": 30,"produktName": "Weizen","varianteName": "Russ","kategorie": "beverage","einzelpreis": 300,"menge": 4},{"positionId": "b0012000-0000-0000-0000-000000000052","varianteId": 44,"produktName": "Wein","varianteName": "Rosé 0,2l","kategorie": "beverage","einzelpreis": 400,"menge": 3}],"gesamtPreisCents": 2850,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.produkte-geliefert:v1', 'tisch:12', 37, now() - interval '17 hours 43 minutes', '{"lieferungId": "c0012000-0000-0000-0000-000000000021","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000062","menge": 1},{"positionId": "b0012000-0000-0000-0000-000000000063","menge": 1},{"positionId": "b0012000-0000-0000-0000-000000000064","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.produkte-geliefert:v1', 'tisch:12', 38, now() - interval '17 hours 42 minutes', '{"lieferungId": "c0012000-0000-0000-0000-000000000018","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000053","menge": 2},{"positionId": "b0012000-0000-0000-0000-000000000054","menge": 1},{"positionId": "b0012000-0000-0000-0000-000000000055","menge": 4},{"positionId": "b0012000-0000-0000-0000-000000000056","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.zahlung-registriert:v1', 'tisch:12', 39, now() - interval '17 hours 40 minutes', '{"zahlungId": "d0012000-0000-0000-0000-000000000020","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000059","menge": 1},{"positionId": "b0012000-0000-0000-0000-000000000060","menge": 3},{"positionId": "b0012000-0000-0000-0000-000000000061","menge": 3}],"gesamtZahlungCents": 3640,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:12', 40, now() - interval '17 hours 39 minutes', '{"bestellungId": "a0012000-0000-0000-0000-000000000015","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000043","varianteId": 19,"produktName": "Waffeln","varianteName": "mit Nutella","kategorie": "food","einzelpreis": 400,"menge": 1},{"positionId": "b0012000-0000-0000-0000-000000000044","varianteId": 6,"produktName": "Flammkuchen","varianteName": "Classic","kategorie": "food","einzelpreis": 600,"menge": 1},{"positionId": "b0012000-0000-0000-0000-000000000045","varianteId": 35,"produktName": "Softdrinks","varianteName": "Mezzo Mix","kategorie": "beverage","einzelpreis": 280,"menge": 1},{"positionId": "b0012000-0000-0000-0000-000000000046","varianteId": 41,"produktName": "Saftschorle","varianteName": "Rhabarberschorle 0,5l","kategorie": "beverage","einzelpreis": 350,"menge": 3}],"gesamtPreisCents": 2330,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.produkte-geliefert:v1', 'tisch:12', 41, now() - interval '17 hours 39 minutes', '{"lieferungId": "c0012000-0000-0000-0000-000000000017","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000050","menge": 1},{"positionId": "b0012000-0000-0000-0000-000000000051","menge": 4},{"positionId": "b0012000-0000-0000-0000-000000000052","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.bestellung-aufgegeben:v1', 'tisch:12', 42, now() - interval '17 hours 38 minutes', '{"bestellungId": "a0012000-0000-0000-0000-000000000014","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000041","varianteId": 13,"produktName": "Grillplatte","varianteName": "Groß","kategorie": "food","einzelpreis": 1400,"menge": 3},{"positionId": "b0012000-0000-0000-0000-000000000042","varianteId": 47,"produktName": "Tee","varianteName": "Verschiedene Sorten","kategorie": "beverage","einzelpreis": 200,"menge": 3}],"gesamtPreisCents": 4800,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.produkte-geliefert:v1', 'tisch:12', 43, now() - interval '17 hours 37 minutes', '{"lieferungId": "c0012000-0000-0000-0000-000000000012","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000035","menge": 1},{"positionId": "b0012000-0000-0000-0000-000000000036","menge": 4},{"positionId": "b0012000-0000-0000-0000-000000000037","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.produkte-geliefert:v1', 'tisch:12', 44, now() - interval '17 hours 36 minutes', '{"lieferungId": "c0012000-0000-0000-0000-000000000010","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000030","menge": 2},{"positionId": "b0012000-0000-0000-0000-000000000031","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.zahlung-registriert:v1', 'tisch:12', 45, now() - interval '17 hours 33 minutes', '{"zahlungId": "d0012000-0000-0000-0000-000000000013","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000038","menge": 3},{"positionId": "b0012000-0000-0000-0000-000000000039","menge": 1},{"positionId": "b0012000-0000-0000-0000-000000000040","menge": 1}],"gesamtZahlungCents": 1500,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.produkte-geliefert:v1', 'tisch:12', 46, now() - interval '17 hours 29 minutes', '{"lieferungId": "c0012000-0000-0000-0000-000000000015","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000043","menge": 1},{"positionId": "b0012000-0000-0000-0000-000000000044","menge": 1},{"positionId": "b0012000-0000-0000-0000-000000000045","menge": 1},{"positionId": "b0012000-0000-0000-0000-000000000046","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.produkte-geliefert:v1', 'tisch:12', 47, now() - interval '17 hours 28 minutes', '{"lieferungId": "c0012000-0000-0000-0000-000000000014","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000041","menge": 3},{"positionId": "b0012000-0000-0000-0000-000000000042","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.zahlung-registriert:v1', 'tisch:12', 48, now() - interval '17 hours 28 minutes', '{"zahlungId": "d0012000-0000-0000-0000-000000000018","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000053","menge": 2},{"positionId": "b0012000-0000-0000-0000-000000000054","menge": 1},{"positionId": "b0012000-0000-0000-0000-000000000055","menge": 4},{"positionId": "b0012000-0000-0000-0000-000000000056","menge": 4}],"gesamtZahlungCents": 7250,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.zahlung-registriert:v1', 'tisch:12', 49, now() - interval '17 hours 25 minutes', '{"zahlungId": "d0012000-0000-0000-0000-000000000021","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000062","menge": 1},{"positionId": "b0012000-0000-0000-0000-000000000063","menge": 1},{"positionId": "b0012000-0000-0000-0000-000000000064","menge": 3}],"gesamtZahlungCents": 1840,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.zahlung-registriert:v1', 'tisch:12', 50, now() - interval '17 hours 24 minutes', '{"zahlungId": "d0012000-0000-0000-0000-000000000012","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000035","menge": 1},{"positionId": "b0012000-0000-0000-0000-000000000036","menge": 4},{"positionId": "b0012000-0000-0000-0000-000000000037","menge": 3}],"gesamtZahlungCents": 5500,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.zahlung-registriert:v1', 'tisch:12', 51, now() - interval '17 hours 23 minutes', '{"zahlungId": "d0012000-0000-0000-0000-000000000017","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000050","menge": 1},{"positionId": "b0012000-0000-0000-0000-000000000051","menge": 4},{"positionId": "b0012000-0000-0000-0000-000000000052","menge": 3}],"gesamtZahlungCents": 2850,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:12', 52, now() - interval '17 hours 19 minutes', '{"bestellungId": "a0012000-0000-0000-0000-000000000016","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000047","varianteId": 15,"produktName": "Salat","varianteName": "Caesar Salat","kategorie": "food","einzelpreis": 650,"menge": 3},{"positionId": "b0012000-0000-0000-0000-000000000048","varianteId": 37,"produktName": "Wasser","varianteName": "Medium 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 2},{"positionId": "b0012000-0000-0000-0000-000000000049","varianteId": 44,"produktName": "Wein","varianteName": "Rosé 0,2l","kategorie": "beverage","einzelpreis": 400,"menge": 4}],"gesamtPreisCents": 3950,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.zahlung-registriert:v1', 'tisch:12', 53, now() - interval '17 hours 16 minutes', '{"zahlungId": "d0012000-0000-0000-0000-000000000014","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000041","menge": 3},{"positionId": "b0012000-0000-0000-0000-000000000042","menge": 3}],"gesamtZahlungCents": 4800,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.zahlung-registriert:v1', 'tisch:12', 54, now() - interval '17 hours 11 minutes', '{"zahlungId": "d0012000-0000-0000-0000-000000000010","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000030","menge": 2},{"positionId": "b0012000-0000-0000-0000-000000000031","menge": 1}],"gesamtZahlungCents": 980,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.produkte-geliefert:v1', 'tisch:12', 55, now() - interval '17 hours 10 minutes', '{"lieferungId": "c0012000-0000-0000-0000-000000000016","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000047","menge": 3},{"positionId": "b0012000-0000-0000-0000-000000000048","menge": 2},{"positionId": "b0012000-0000-0000-0000-000000000049","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.zahlung-registriert:v1', 'tisch:12', 56, now() - interval '17 hours 5 minutes', '{"zahlungId": "d0012000-0000-0000-0000-000000000015","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000043","menge": 1},{"positionId": "b0012000-0000-0000-0000-000000000044","menge": 1},{"positionId": "b0012000-0000-0000-0000-000000000045","menge": 1},{"positionId": "b0012000-0000-0000-0000-000000000046","menge": 3}],"gesamtZahlungCents": 2330,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:12', 57, now() - interval '17 hours 5 minutes', '{"bestellungId": "a0012000-0000-0000-0000-000000000019","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000057","varianteId": 2,"produktName": "Bratwurst","varianteName": "XXL","kategorie": "food","einzelpreis": 500,"menge": 2},{"positionId": "b0012000-0000-0000-0000-000000000058","varianteId": 30,"produktName": "Weizen","varianteName": "Russ","kategorie": "beverage","einzelpreis": 300,"menge": 1}],"gesamtPreisCents": 1300,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:12', 58, now() - interval '16 hours 56 minutes', '{"lieferungId": "c0012000-0000-0000-0000-000000000019","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000057","menge": 2},{"positionId": "b0012000-0000-0000-0000-000000000058","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.zahlung-registriert:v1', 'tisch:12', 59, now() - interval '16 hours 53 minutes', '{"zahlungId": "d0012000-0000-0000-0000-000000000016","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000047","menge": 3},{"positionId": "b0012000-0000-0000-0000-000000000048","menge": 2},{"positionId": "b0012000-0000-0000-0000-000000000049","menge": 4}],"gesamtZahlungCents": 3950,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:12', 60, now() - interval '16 hours 40 minutes', '{"bestellungId": "a0012000-0000-0000-0000-000000000011","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000032","varianteId": 2,"produktName": "Bratwurst","varianteName": "XXL","kategorie": "food","einzelpreis": 500,"menge": 1},{"positionId": "b0012000-0000-0000-0000-000000000033","varianteId": 12,"produktName": "Grillplatte","varianteName": "Klein","kategorie": "food","einzelpreis": 800,"menge": 1},{"positionId": "b0012000-0000-0000-0000-000000000034","varianteId": 43,"produktName": "Wein","varianteName": "Rotwein 0,2l","kategorie": "beverage","einzelpreis": 400,"menge": 3}],"gesamtPreisCents": 2500,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:12', 61, now() - interval '16 hours 31 minutes', '{"zahlungId": "d0012000-0000-0000-0000-000000000019","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000057","menge": 2},{"positionId": "b0012000-0000-0000-0000-000000000058","menge": 1}],"gesamtZahlungCents": 1300,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.produkte-geliefert:v1', 'tisch:12', 62, now() - interval '16 hours 28 minutes', '{"lieferungId": "c0012000-0000-0000-0000-000000000011","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000032","menge": 1},{"positionId": "b0012000-0000-0000-0000-000000000033","menge": 1},{"positionId": "b0012000-0000-0000-0000-000000000034","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.zahlung-registriert:v1', 'tisch:12', 63, now() - interval '16 hours 6 minutes', '{"zahlungId": "d0012000-0000-0000-0000-000000000011","positionen": [{"positionId": "b0012000-0000-0000-0000-000000000032","menge": 1},{"positionId": "b0012000-0000-0000-0000-000000000033","menge": 1},{"positionId": "b0012000-0000-0000-0000-000000000034","menge": 3}],"gesamtZahlungCents": 2500,"kommentar": ""}');

-- ─────────────────────────────────────────────────────────────────────────────
-- TISCH 13: Musikverein — Tag 2, Mittag + Abend (18 Events)
-- ─────────────────────────────────────────────────────────────────────────────

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:13', 1, now() - interval '26 hours 40 minutes', '{"bestellungId": "a0013000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0013000-0000-0000-0000-000000000001","varianteId": 10,"produktName": "Tagesgericht","varianteName": "Sa: Gulasch mit Spätzle","kategorie": "food","einzelpreis": 1150,"menge": 3},{"positionId": "b0013000-0000-0000-0000-000000000002","varianteId": 14,"produktName": "Salat","varianteName": "Gemischter Salat","kategorie": "food","einzelpreis": 550,"menge": 2},{"positionId": "b0013000-0000-0000-0000-000000000003","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 3},{"positionId": "b0013000-0000-0000-0000-000000000004","varianteId": 36,"produktName": "Wasser","varianteName": "Still 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 2}],"gesamtPreisCents": 6300,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.produkte-geliefert:v1', 'tisch:13', 2, now() - interval '26 hours 25 minutes', '{"lieferungId": "c0013000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0013000-0000-0000-0000-000000000001","menge": 3},{"positionId": "b0013000-0000-0000-0000-000000000002","menge": 2},{"positionId": "b0013000-0000-0000-0000-000000000003","menge": 3},{"positionId": "b0013000-0000-0000-0000-000000000004","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.zahlung-registriert:v1', 'tisch:13', 3, now() - interval '26 hours 3 minutes', '{"zahlungId": "d0013000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0013000-0000-0000-0000-000000000001","menge": 3},{"positionId": "b0013000-0000-0000-0000-000000000002","menge": 2},{"positionId": "b0013000-0000-0000-0000-000000000003","menge": 3},{"positionId": "b0013000-0000-0000-0000-000000000004","menge": 2}],"gesamtZahlungCents": 6300,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:13', 4, now() - interval '25 hours 47 minutes', '{"bestellungId": "a0013000-0000-0000-0000-000000000002","positionen": [{"positionId": "b0013000-0000-0000-0000-000000000005","varianteId": 6,"produktName": "Flammkuchen","varianteName": "Classic","kategorie": "food","einzelpreis": 600,"menge": 2},{"positionId": "b0013000-0000-0000-0000-000000000006","varianteId": 43,"produktName": "Wein","varianteName": "Rotwein 0,2l","kategorie": "beverage","einzelpreis": 400,"menge": 3}],"gesamtPreisCents": 2400,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.produkte-geliefert:v1', 'tisch:13', 5, now() - interval '25 hours 33 minutes', '{"lieferungId": "c0013000-0000-0000-0000-000000000002","positionen": [{"positionId": "b0013000-0000-0000-0000-000000000005","menge": 2},{"positionId": "b0013000-0000-0000-0000-000000000006","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.zahlung-registriert:v1', 'tisch:13', 6, now() - interval '25 hours 15 minutes', '{"zahlungId": "d0013000-0000-0000-0000-000000000002","positionen": [{"positionId": "b0013000-0000-0000-0000-000000000005","menge": 2},{"positionId": "b0013000-0000-0000-0000-000000000006","menge": 3}],"gesamtZahlungCents": 2400,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:13', 7, now() - interval '24 hours 56 minutes', '{"bestellungId": "a0013000-0000-0000-0000-000000000003","positionen": [{"positionId": "b0013000-0000-0000-0000-000000000007","varianteId": 16,"produktName": "Kuchen","varianteName": "Stück","kategorie": "food","einzelpreis": 250,"menge": 3},{"positionId": "b0013000-0000-0000-0000-000000000008","varianteId": 45,"produktName": "Kaffee","varianteName": "Tasse","kategorie": "beverage","einzelpreis": 200,"menge": 3}],"gesamtPreisCents": 1350,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:13', 8, now() - interval '24 hours 43 minutes', '{"lieferungId": "c0013000-0000-0000-0000-000000000003","positionen": [{"positionId": "b0013000-0000-0000-0000-000000000007","menge": 3},{"positionId": "b0013000-0000-0000-0000-000000000008","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:13', 9, now() - interval '24 hours 23 minutes', '{"zahlungId": "d0013000-0000-0000-0000-000000000003","positionen": [{"positionId": "b0013000-0000-0000-0000-000000000007","menge": 3},{"positionId": "b0013000-0000-0000-0000-000000000008","menge": 3}],"gesamtZahlungCents": 1350,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:13', 10, now() - interval '20 hours', '{"bestellungId": "a0013000-0000-0000-0000-000000000004","positionen": [{"positionId": "b0013000-0000-0000-0000-000000000009","varianteId": 7,"produktName": "Flammkuchen","varianteName": "Speck & Zwiebel","kategorie": "food","einzelpreis": 700,"menge": 2},{"positionId": "b0013000-0000-0000-0000-000000000010","varianteId": 12,"produktName": "Grillplatte","varianteName": "Klein","kategorie": "food","einzelpreis": 800,"menge": 1},{"positionId": "b0013000-0000-0000-0000-000000000011","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 4},{"positionId": "b0013000-0000-0000-0000-000000000012","varianteId": 27,"produktName": "Weizen","varianteName": "Groß 0,5l","kategorie": "beverage","einzelpreis": 400,"menge": 2}],"gesamtPreisCents": 4800,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:13', 11, now() - interval '19 hours 49 minutes', '{"lieferungId": "c0013000-0000-0000-0000-000000000004","positionen": [{"positionId": "b0013000-0000-0000-0000-000000000009","menge": 2},{"positionId": "b0013000-0000-0000-0000-000000000010","menge": 1},{"positionId": "b0013000-0000-0000-0000-000000000011","menge": 4},{"positionId": "b0013000-0000-0000-0000-000000000012","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:13', 12, now() - interval '19 hours 24 minutes', '{"zahlungId": "d0013000-0000-0000-0000-000000000004","positionen": [{"positionId": "b0013000-0000-0000-0000-000000000009","menge": 2},{"positionId": "b0013000-0000-0000-0000-000000000010","menge": 1},{"positionId": "b0013000-0000-0000-0000-000000000011","menge": 4},{"positionId": "b0013000-0000-0000-0000-000000000012","menge": 2}],"gesamtZahlungCents": 4800,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:13', 13, now() - interval '19 hours 4 minutes', '{"bestellungId": "a0013000-0000-0000-0000-000000000005","positionen": [{"positionId": "b0013000-0000-0000-0000-000000000013","varianteId": 1,"produktName": "Bratwurst","varianteName": "Normal","kategorie": "food","einzelpreis": 350,"menge": 3},{"positionId": "b0013000-0000-0000-0000-000000000014","varianteId": 5,"produktName": "Pommes","varianteName": "Groß","kategorie": "food","einzelpreis": 350,"menge": 2},{"positionId": "b0013000-0000-0000-0000-000000000015","varianteId": 23,"produktName": "Bier","varianteName": "0,3l","kategorie": "beverage","einzelpreis": 300,"menge": 4}],"gesamtPreisCents": 2950,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.produkte-geliefert:v1', 'tisch:13', 14, now() - interval '18 hours 53 minutes', '{"lieferungId": "c0013000-0000-0000-0000-000000000005","positionen": [{"positionId": "b0013000-0000-0000-0000-000000000013","menge": 3},{"positionId": "b0013000-0000-0000-0000-000000000014","menge": 2},{"positionId": "b0013000-0000-0000-0000-000000000015","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.zahlung-registriert:v1', 'tisch:13', 15, now() - interval '18 hours 35 minutes', '{"zahlungId": "d0013000-0000-0000-0000-000000000005","positionen": [{"positionId": "b0013000-0000-0000-0000-000000000013","menge": 3},{"positionId": "b0013000-0000-0000-0000-000000000014","menge": 2},{"positionId": "b0013000-0000-0000-0000-000000000015","menge": 4}],"gesamtZahlungCents": 2950,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:13', 16, now() - interval '18 hours 13 minutes', '{"bestellungId": "a0013000-0000-0000-0000-000000000006","positionen": [{"positionId": "b0013000-0000-0000-0000-000000000016","varianteId": 48,"produktName": "Hugo/Aperol","varianteName": "Hugo","kategorie": "beverage","einzelpreis": 550,"menge": 3},{"positionId": "b0013000-0000-0000-0000-000000000017","varianteId": 49,"produktName": "Hugo/Aperol","varianteName": "Aperol Spritz","kategorie": "beverage","einzelpreis": 550,"menge": 2}],"gesamtPreisCents": 2750,"kommentar": "Absacker"}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:13', 17, now() - interval '18 hours 1 minutes', '{"lieferungId": "c0013000-0000-0000-0000-000000000006","positionen": [{"positionId": "b0013000-0000-0000-0000-000000000016","menge": 3},{"positionId": "b0013000-0000-0000-0000-000000000017","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:13', 18, now() - interval '17 hours 45 minutes', '{"zahlungId": "d0013000-0000-0000-0000-000000000006","positionen": [{"positionId": "b0013000-0000-0000-0000-000000000016","menge": 3},{"positionId": "b0013000-0000-0000-0000-000000000017","menge": 2}],"gesamtZahlungCents": 2750,"kommentar": ""}');

-- ─────────────────────────────────────────────────────────────────────────────
-- TISCH 14: Gerade bestellt — Tag 3, ganz frisch (1 Events)
-- ─────────────────────────────────────────────────────────────────────────────

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:14', 1, now() - interval '15 minutes', '{"bestellungId": "a0014000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0014000-0000-0000-0000-000000000001","varianteId": 11,"produktName": "Tagesgericht","varianteName": "So: Hähnchen mit Reis","kategorie": "food","einzelpreis": 1050,"menge": 2},{"positionId": "b0014000-0000-0000-0000-000000000002","varianteId": 15,"produktName": "Salat","varianteName": "Caesar Salat","kategorie": "food","einzelpreis": 650,"menge": 1},{"positionId": "b0014000-0000-0000-0000-000000000003","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 2},{"positionId": "b0014000-0000-0000-0000-000000000004","varianteId": 33,"produktName": "Softdrinks","varianteName": "Spezi","kategorie": "beverage","einzelpreis": 280,"menge": 2}],"gesamtPreisCents": 4210,"kommentar": "Gerade bestellt"}');

-- TISCH 15: Leer — keine Events
-- (Keine Events)

-- ─────────────────────────────────────────────────────────────────────────────
-- TISCH 16: Zelt A1 — Großes Vereinsfest, SEHR viel Betrieb (75 Events)
-- ─────────────────────────────────────────────────────────────────────────────

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:16', 1, now() - interval '50 hours 50 minutes', '{"bestellungId": "a0016000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000001","varianteId": 51,"produktName": "Festbändchen","varianteName": "Erwachsene","kategorie": "other","einzelpreis": 500,"menge": 8},{"positionId": "b0016000-0000-0000-0000-000000000002","varianteId": 52,"produktName": "Festbändchen","varianteName": "Kinder","kategorie": "other","einzelpreis": 300,"menge": 4}],"gesamtPreisCents": 5200,"kommentar": "Festbändchen Zelt A1"}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:16', 2, now() - interval '50 hours 35 minutes', '{"lieferungId": "c0016000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000001","menge": 8},{"positionId": "b0016000-0000-0000-0000-000000000002","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:16', 3, now() - interval '50 hours 13 minutes', '{"zahlungId": "d0016000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000001","menge": 8},{"positionId": "b0016000-0000-0000-0000-000000000002","menge": 4}],"gesamtZahlungCents": 5200,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:16', 4, now() - interval '49 hours 58 minutes', '{"bestellungId": "a0016000-0000-0000-0000-000000000002","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000003","varianteId": 13,"produktName": "Grillplatte","varianteName": "Groß","kategorie": "food","einzelpreis": 1400,"menge": 2},{"positionId": "b0016000-0000-0000-0000-000000000004","varianteId": 7,"produktName": "Flammkuchen","varianteName": "Speck & Zwiebel","kategorie": "food","einzelpreis": 700,"menge": 3},{"positionId": "b0016000-0000-0000-0000-000000000005","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 8},{"positionId": "b0016000-0000-0000-0000-000000000006","varianteId": 25,"produktName": "Bier","varianteName": "Maß 1,0l","kategorie": "beverage","einzelpreis": 850,"menge": 4}],"gesamtPreisCents": 11900,"kommentar": "Eröffnung Zelt A1"}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:16', 5, now() - interval '49 hours 46 minutes', '{"lieferungId": "c0016000-0000-0000-0000-000000000002","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000003","menge": 2},{"positionId": "b0016000-0000-0000-0000-000000000004","menge": 3},{"positionId": "b0016000-0000-0000-0000-000000000005","menge": 8},{"positionId": "b0016000-0000-0000-0000-000000000006","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:16', 6, now() - interval '49 hours 28 minutes', '{"zahlungId": "d0016000-0000-0000-0000-000000000002","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000003","menge": 2},{"positionId": "b0016000-0000-0000-0000-000000000004","menge": 3},{"positionId": "b0016000-0000-0000-0000-000000000005","menge": 8},{"positionId": "b0016000-0000-0000-0000-000000000006","menge": 4}],"gesamtZahlungCents": 11900,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.bestellung-aufgegeben:v1', 'tisch:16', 7, now() - interval '49 hours 10 minutes', '{"bestellungId": "a0016000-0000-0000-0000-000000000003","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000007","varianteId": 1,"produktName": "Bratwurst","varianteName": "Normal","kategorie": "food","einzelpreis": 350,"menge": 6},{"positionId": "b0016000-0000-0000-0000-000000000008","varianteId": 5,"produktName": "Pommes","varianteName": "Groß","kategorie": "food","einzelpreis": 350,"menge": 6},{"positionId": "b0016000-0000-0000-0000-000000000009","varianteId": 23,"produktName": "Bier","varianteName": "0,3l","kategorie": "beverage","einzelpreis": 300,"menge": 8},{"positionId": "b0016000-0000-0000-0000-000000000010","varianteId": 33,"produktName": "Softdrinks","varianteName": "Spezi","kategorie": "beverage","einzelpreis": 280,"menge": 4}],"gesamtPreisCents": 7720,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.produkte-geliefert:v1', 'tisch:16', 8, now() - interval '48 hours 58 minutes', '{"lieferungId": "c0016000-0000-0000-0000-000000000003","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000007","menge": 6},{"positionId": "b0016000-0000-0000-0000-000000000008","menge": 6},{"positionId": "b0016000-0000-0000-0000-000000000009","menge": 8},{"positionId": "b0016000-0000-0000-0000-000000000010","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.zahlung-registriert:v1', 'tisch:16', 9, now() - interval '48 hours 40 minutes', '{"zahlungId": "d0016000-0000-0000-0000-000000000003","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000007","menge": 6},{"positionId": "b0016000-0000-0000-0000-000000000008","menge": 6},{"positionId": "b0016000-0000-0000-0000-000000000009","menge": 8},{"positionId": "b0016000-0000-0000-0000-000000000010","menge": 4}],"gesamtZahlungCents": 7720,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.produkte-storniert:v1', 'tisch:16', 10, now() - interval '48 hours 33 minutes', '{"stornierungId": "e0016000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000007","menge": 2}],"gesamtStornierungCents": 700,"kommentar": "Verwechslung: sollte Currywurst sein"}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:16', 11, now() - interval '48 hours 18 minutes', '{"bestellungId": "a0016000-0000-0000-0000-000000000004","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000011","varianteId": 9,"produktName": "Tagesgericht","varianteName": "Fr: Schnitzel mit Pommes","kategorie": "food","einzelpreis": 1250,"menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000012","varianteId": 14,"produktName": "Salat","varianteName": "Gemischter Salat","kategorie": "food","einzelpreis": 550,"menge": 3},{"positionId": "b0016000-0000-0000-0000-000000000013","varianteId": 36,"produktName": "Wasser","varianteName": "Still 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000014","varianteId": 42,"produktName": "Wein","varianteName": "Weißwein 0,2l","kategorie": "beverage","einzelpreis": 400,"menge": 3}],"gesamtPreisCents": 8650,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.produkte-geliefert:v1', 'tisch:16', 12, now() - interval '48 hours 5 minutes', '{"lieferungId": "c0016000-0000-0000-0000-000000000004","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000011","menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000012","menge": 3},{"positionId": "b0016000-0000-0000-0000-000000000013","menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000014","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.zahlung-registriert:v1', 'tisch:16', 13, now() - interval '47 hours 49 minutes', '{"zahlungId": "d0016000-0000-0000-0000-000000000004","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000011","menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000012","menge": 3},{"positionId": "b0016000-0000-0000-0000-000000000013","menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000014","menge": 3}],"gesamtZahlungCents": 8650,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:16', 14, now() - interval '47 hours 32 minutes', '{"bestellungId": "a0016000-0000-0000-0000-000000000005","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000015","varianteId": 6,"produktName": "Flammkuchen","varianteName": "Classic","kategorie": "food","einzelpreis": 600,"menge": 3},{"positionId": "b0016000-0000-0000-0000-000000000016","varianteId": 12,"produktName": "Grillplatte","varianteName": "Klein","kategorie": "food","einzelpreis": 800,"menge": 2},{"positionId": "b0016000-0000-0000-0000-000000000017","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 6},{"positionId": "b0016000-0000-0000-0000-000000000018","varianteId": 27,"produktName": "Weizen","varianteName": "Groß 0,5l","kategorie": "beverage","einzelpreis": 400,"menge": 4}],"gesamtPreisCents": 7700,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:16', 15, now() - interval '47 hours 23 minutes', '{"lieferungId": "c0016000-0000-0000-0000-000000000005","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000015","menge": 3},{"positionId": "b0016000-0000-0000-0000-000000000016","menge": 2},{"positionId": "b0016000-0000-0000-0000-000000000017","menge": 6},{"positionId": "b0016000-0000-0000-0000-000000000018","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:16', 16, now() - interval '47 hours 3 minutes', '{"zahlungId": "d0016000-0000-0000-0000-000000000005","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000015","menge": 3},{"positionId": "b0016000-0000-0000-0000-000000000016","menge": 2},{"positionId": "b0016000-0000-0000-0000-000000000017","menge": 6},{"positionId": "b0016000-0000-0000-0000-000000000018","menge": 4}],"gesamtZahlungCents": 7700,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.bestellung-aufgegeben:v1', 'tisch:16', 17, now() - interval '46 hours 45 minutes', '{"bestellungId": "a0016000-0000-0000-0000-000000000006","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000019","varianteId": 3,"produktName": "Bratwurst","varianteName": "Currywurst","kategorie": "food","einzelpreis": 450,"menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000020","varianteId": 4,"produktName": "Pommes","varianteName": "Klein","kategorie": "food","einzelpreis": 250,"menge": 6},{"positionId": "b0016000-0000-0000-0000-000000000021","varianteId": 23,"produktName": "Bier","varianteName": "0,3l","kategorie": "beverage","einzelpreis": 300,"menge": 8},{"positionId": "b0016000-0000-0000-0000-000000000022","varianteId": 31,"produktName": "Softdrinks","varianteName": "Cola","kategorie": "beverage","einzelpreis": 280,"menge": 6}],"gesamtPreisCents": 7380,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.produkte-geliefert:v1', 'tisch:16', 18, now() - interval '46 hours 30 minutes', '{"lieferungId": "c0016000-0000-0000-0000-000000000006","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000019","menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000020","menge": 6},{"positionId": "b0016000-0000-0000-0000-000000000021","menge": 8},{"positionId": "b0016000-0000-0000-0000-000000000022","menge": 6}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.zahlung-registriert:v1', 'tisch:16', 19, now() - interval '46 hours 3 minutes', '{"zahlungId": "d0016000-0000-0000-0000-000000000006","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000019","menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000020","menge": 6},{"positionId": "b0016000-0000-0000-0000-000000000021","menge": 8},{"positionId": "b0016000-0000-0000-0000-000000000022","menge": 6}],"gesamtZahlungCents": 7380,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:16', 20, now() - interval '45 hours 53 minutes', '{"bestellungId": "a0016000-0000-0000-0000-000000000007","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000023","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 8},{"positionId": "b0016000-0000-0000-0000-000000000024","varianteId": 26,"produktName": "Weizen","varianteName": "Klein 0,3l","kategorie": "beverage","einzelpreis": 300,"menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000025","varianteId": 30,"produktName": "Weizen","varianteName": "Russ","kategorie": "beverage","einzelpreis": 300,"menge": 3}],"gesamtPreisCents": 5700,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.produkte-geliefert:v1', 'tisch:16', 21, now() - interval '45 hours 41 minutes', '{"lieferungId": "c0016000-0000-0000-0000-000000000007","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000023","menge": 8},{"positionId": "b0016000-0000-0000-0000-000000000024","menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000025","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.zahlung-registriert:v1', 'tisch:16', 22, now() - interval '45 hours 26 minutes', '{"zahlungId": "d0016000-0000-0000-0000-000000000007","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000023","menge": 8},{"positionId": "b0016000-0000-0000-0000-000000000024","menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000025","menge": 3}],"gesamtZahlungCents": 5700,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:16', 23, now() - interval '45 hours 5 minutes', '{"bestellungId": "a0016000-0000-0000-0000-000000000008","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000026","varianteId": 17,"produktName": "Waffeln","varianteName": "mit Puderzucker","kategorie": "food","einzelpreis": 300,"menge": 6},{"positionId": "b0016000-0000-0000-0000-000000000027","varianteId": 18,"produktName": "Waffeln","varianteName": "mit Sahne","kategorie": "food","einzelpreis": 350,"menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000028","varianteId": 45,"produktName": "Kaffee","varianteName": "Tasse","kategorie": "beverage","einzelpreis": 200,"menge": 6}],"gesamtPreisCents": 4400,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:16', 24, now() - interval '44 hours 53 minutes', '{"lieferungId": "c0016000-0000-0000-0000-000000000008","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000026","menge": 6},{"positionId": "b0016000-0000-0000-0000-000000000027","menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000028","menge": 6}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:16', 25, now() - interval '44 hours 39 minutes', '{"zahlungId": "d0016000-0000-0000-0000-000000000008","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000026","menge": 6},{"positionId": "b0016000-0000-0000-0000-000000000027","menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000028","menge": 6}],"gesamtZahlungCents": 4400,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.bestellung-aufgegeben:v1', 'tisch:16', 26, now() - interval '44 hours 17 minutes', '{"bestellungId": "a0016000-0000-0000-0000-000000000009","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000029","varianteId": 25,"produktName": "Bier","varianteName": "Maß 1,0l","kategorie": "beverage","einzelpreis": 850,"menge": 6},{"positionId": "b0016000-0000-0000-0000-000000000030","varianteId": 48,"produktName": "Hugo/Aperol","varianteName": "Hugo","kategorie": "beverage","einzelpreis": 550,"menge": 4}],"gesamtPreisCents": 7300,"kommentar": "Letzte Runde Freitag"}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.produkte-geliefert:v1', 'tisch:16', 27, now() - interval '44 hours 4 minutes', '{"lieferungId": "c0016000-0000-0000-0000-000000000009","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000029","menge": 6},{"positionId": "b0016000-0000-0000-0000-000000000030","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.zahlung-registriert:v1', 'tisch:16', 28, now() - interval '43 hours 52 minutes', '{"zahlungId": "d0016000-0000-0000-0000-000000000009","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000029","menge": 6},{"positionId": "b0016000-0000-0000-0000-000000000030","menge": 4}],"gesamtZahlungCents": 7300,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:16', 29, now() - interval '28 hours 20 minutes', '{"bestellungId": "a0016000-0000-0000-0000-000000000010","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000031","varianteId": 10,"produktName": "Tagesgericht","varianteName": "Sa: Gulasch mit Spätzle","kategorie": "food","einzelpreis": 1150,"menge": 6},{"positionId": "b0016000-0000-0000-0000-000000000032","varianteId": 13,"produktName": "Grillplatte","varianteName": "Groß","kategorie": "food","einzelpreis": 1400,"menge": 3},{"positionId": "b0016000-0000-0000-0000-000000000033","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 8},{"positionId": "b0016000-0000-0000-0000-000000000034","varianteId": 27,"produktName": "Weizen","varianteName": "Groß 0,5l","kategorie": "beverage","einzelpreis": 400,"menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000035","varianteId": 36,"produktName": "Wasser","varianteName": "Still 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 4}],"gesamtPreisCents": 17100,"kommentar": "Mittagessen Samstag"}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.produkte-geliefert:v1', 'tisch:16', 30, now() - interval '28 hours 11 minutes', '{"lieferungId": "c0016000-0000-0000-0000-000000000010","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000031","menge": 6},{"positionId": "b0016000-0000-0000-0000-000000000032","menge": 3},{"positionId": "b0016000-0000-0000-0000-000000000033","menge": 8},{"positionId": "b0016000-0000-0000-0000-000000000034","menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000035","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.zahlung-registriert:v1', 'tisch:16', 31, now() - interval '27 hours 47 minutes', '{"zahlungId": "d0016000-0000-0000-0000-000000000010","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000031","menge": 6},{"positionId": "b0016000-0000-0000-0000-000000000032","menge": 3},{"positionId": "b0016000-0000-0000-0000-000000000033","menge": 8},{"positionId": "b0016000-0000-0000-0000-000000000034","menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000035","menge": 4}],"gesamtZahlungCents": 17100,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:16', 32, now() - interval '27 hours 35 minutes', '{"bestellungId": "a0016000-0000-0000-0000-000000000011","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000036","varianteId": 1,"produktName": "Bratwurst","varianteName": "Normal","kategorie": "food","einzelpreis": 350,"menge": 6},{"positionId": "b0016000-0000-0000-0000-000000000037","varianteId": 5,"produktName": "Pommes","varianteName": "Groß","kategorie": "food","einzelpreis": 350,"menge": 6},{"positionId": "b0016000-0000-0000-0000-000000000038","varianteId": 23,"produktName": "Bier","varianteName": "0,3l","kategorie": "beverage","einzelpreis": 300,"menge": 8},{"positionId": "b0016000-0000-0000-0000-000000000039","varianteId": 33,"produktName": "Softdrinks","varianteName": "Spezi","kategorie": "beverage","einzelpreis": 280,"menge": 6}],"gesamtPreisCents": 8280,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.produkte-geliefert:v1', 'tisch:16', 33, now() - interval '27 hours 26 minutes', '{"lieferungId": "c0016000-0000-0000-0000-000000000011","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000036","menge": 6},{"positionId": "b0016000-0000-0000-0000-000000000037","menge": 6},{"positionId": "b0016000-0000-0000-0000-000000000038","menge": 8},{"positionId": "b0016000-0000-0000-0000-000000000039","menge": 6}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.zahlung-registriert:v1', 'tisch:16', 34, now() - interval '27 hours 8 minutes', '{"zahlungId": "d0016000-0000-0000-0000-000000000011","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000036","menge": 6},{"positionId": "b0016000-0000-0000-0000-000000000037","menge": 6},{"positionId": "b0016000-0000-0000-0000-000000000038","menge": 8},{"positionId": "b0016000-0000-0000-0000-000000000039","menge": 6}],"gesamtZahlungCents": 8280,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:16', 35, now() - interval '26 hours 53 minutes', '{"bestellungId": "a0016000-0000-0000-0000-000000000012","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000040","varianteId": 7,"produktName": "Flammkuchen","varianteName": "Speck & Zwiebel","kategorie": "food","einzelpreis": 700,"menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000041","varianteId": 8,"produktName": "Flammkuchen","varianteName": "Mediterran","kategorie": "food","einzelpreis": 750,"menge": 2},{"positionId": "b0016000-0000-0000-0000-000000000042","varianteId": 42,"produktName": "Wein","varianteName": "Weißwein 0,2l","kategorie": "beverage","einzelpreis": 400,"menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000043","varianteId": 44,"produktName": "Wein","varianteName": "Rosé 0,2l","kategorie": "beverage","einzelpreis": 400,"menge": 2}],"gesamtPreisCents": 6700,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:16', 36, now() - interval '26 hours 42 minutes', '{"lieferungId": "c0016000-0000-0000-0000-000000000012","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000040","menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000041","menge": 2},{"positionId": "b0016000-0000-0000-0000-000000000042","menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000043","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:16', 37, now() - interval '26 hours 18 minutes', '{"zahlungId": "d0016000-0000-0000-0000-000000000012","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000040","menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000041","menge": 2},{"positionId": "b0016000-0000-0000-0000-000000000042","menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000043","menge": 2}],"gesamtZahlungCents": 6700,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:16', 38, now() - interval '26 hours 7 minutes', '{"bestellungId": "a0016000-0000-0000-0000-000000000013","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000044","varianteId": 12,"produktName": "Grillplatte","varianteName": "Klein","kategorie": "food","einzelpreis": 800,"menge": 3},{"positionId": "b0016000-0000-0000-0000-000000000045","varianteId": 14,"produktName": "Salat","varianteName": "Gemischter Salat","kategorie": "food","einzelpreis": 550,"menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000046","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 6},{"positionId": "b0016000-0000-0000-0000-000000000047","varianteId": 39,"produktName": "Saftschorle","varianteName": "Apfelschorle 0,5l","kategorie": "beverage","einzelpreis": 300,"menge": 4}],"gesamtPreisCents": 8500,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.produkte-geliefert:v1', 'tisch:16', 39, now() - interval '25 hours 55 minutes', '{"lieferungId": "c0016000-0000-0000-0000-000000000013","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000044","menge": 3},{"positionId": "b0016000-0000-0000-0000-000000000045","menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000046","menge": 6},{"positionId": "b0016000-0000-0000-0000-000000000047","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.zahlung-registriert:v1', 'tisch:16', 40, now() - interval '25 hours 43 minutes', '{"zahlungId": "d0016000-0000-0000-0000-000000000013","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000044","menge": 3},{"positionId": "b0016000-0000-0000-0000-000000000045","menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000046","menge": 6},{"positionId": "b0016000-0000-0000-0000-000000000047","menge": 4}],"gesamtZahlungCents": 8500,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:16', 41, now() - interval '25 hours 29 minutes', '{"bestellungId": "a0016000-0000-0000-0000-000000000014","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000048","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 8},{"positionId": "b0016000-0000-0000-0000-000000000049","varianteId": 25,"produktName": "Bier","varianteName": "Maß 1,0l","kategorie": "beverage","einzelpreis": 850,"menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000050","varianteId": 26,"produktName": "Weizen","varianteName": "Klein 0,3l","kategorie": "beverage","einzelpreis": 300,"menge": 4}],"gesamtPreisCents": 8200,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.produkte-geliefert:v1', 'tisch:16', 42, now() - interval '25 hours 18 minutes', '{"lieferungId": "c0016000-0000-0000-0000-000000000014","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000048","menge": 8},{"positionId": "b0016000-0000-0000-0000-000000000049","menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000050","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.zahlung-registriert:v1', 'tisch:16', 43, now() - interval '24 hours 53 minutes', '{"zahlungId": "d0016000-0000-0000-0000-000000000014","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000048","menge": 8},{"positionId": "b0016000-0000-0000-0000-000000000049","menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000050","menge": 4}],"gesamtZahlungCents": 8200,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:16', 44, now() - interval '24 hours 37 minutes', '{"bestellungId": "a0016000-0000-0000-0000-000000000015","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000051","varianteId": 49,"produktName": "Hugo/Aperol","varianteName": "Aperol Spritz","kategorie": "beverage","einzelpreis": 550,"menge": 6},{"positionId": "b0016000-0000-0000-0000-000000000052","varianteId": 48,"produktName": "Hugo/Aperol","varianteName": "Hugo","kategorie": "beverage","einzelpreis": 550,"menge": 4}],"gesamtPreisCents": 5500,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:16', 45, now() - interval '24 hours 27 minutes', '{"lieferungId": "c0016000-0000-0000-0000-000000000015","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000051","menge": 6},{"positionId": "b0016000-0000-0000-0000-000000000052","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:16', 46, now() - interval '24 hours 12 minutes', '{"zahlungId": "d0016000-0000-0000-0000-000000000015","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000051","menge": 6},{"positionId": "b0016000-0000-0000-0000-000000000052","menge": 4}],"gesamtZahlungCents": 5500,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:16', 47, now() - interval '23 hours 50 minutes', '{"bestellungId": "a0016000-0000-0000-0000-000000000016","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000053","varianteId": 10,"produktName": "Tagesgericht","varianteName": "Sa: Gulasch mit Spätzle","kategorie": "food","einzelpreis": 1150,"menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000054","varianteId": 7,"produktName": "Flammkuchen","varianteName": "Speck & Zwiebel","kategorie": "food","einzelpreis": 700,"menge": 3},{"positionId": "b0016000-0000-0000-0000-000000000055","varianteId": 6,"produktName": "Flammkuchen","varianteName": "Classic","kategorie": "food","einzelpreis": 600,"menge": 2},{"positionId": "b0016000-0000-0000-0000-000000000056","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 8},{"positionId": "b0016000-0000-0000-0000-000000000057","varianteId": 43,"produktName": "Wein","varianteName": "Rotwein 0,2l","kategorie": "beverage","einzelpreis": 400,"menge": 4}],"gesamtPreisCents": 13100,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.produkte-geliefert:v1', 'tisch:16', 48, now() - interval '23 hours 38 minutes', '{"lieferungId": "c0016000-0000-0000-0000-000000000016","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000053","menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000054","menge": 3},{"positionId": "b0016000-0000-0000-0000-000000000055","menge": 2},{"positionId": "b0016000-0000-0000-0000-000000000056","menge": 8},{"positionId": "b0016000-0000-0000-0000-000000000057","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.zahlung-registriert:v1', 'tisch:16', 49, now() - interval '23 hours 25 minutes', '{"zahlungId": "d0016000-0000-0000-0000-000000000016","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000053","menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000054","menge": 3},{"positionId": "b0016000-0000-0000-0000-000000000055","menge": 2},{"positionId": "b0016000-0000-0000-0000-000000000056","menge": 8},{"positionId": "b0016000-0000-0000-0000-000000000057","menge": 4}],"gesamtZahlungCents": 13100,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:16', 50, now() - interval '23 hours 8 minutes', '{"bestellungId": "a0016000-0000-0000-0000-000000000017","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000058","varianteId": 1,"produktName": "Bratwurst","varianteName": "Normal","kategorie": "food","einzelpreis": 350,"menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000059","varianteId": 3,"produktName": "Bratwurst","varianteName": "Currywurst","kategorie": "food","einzelpreis": 450,"menge": 3},{"positionId": "b0016000-0000-0000-0000-000000000060","varianteId": 5,"produktName": "Pommes","varianteName": "Groß","kategorie": "food","einzelpreis": 350,"menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000061","varianteId": 23,"produktName": "Bier","varianteName": "0,3l","kategorie": "beverage","einzelpreis": 300,"menge": 6},{"positionId": "b0016000-0000-0000-0000-000000000062","varianteId": 31,"produktName": "Softdrinks","varianteName": "Cola","kategorie": "beverage","einzelpreis": 280,"menge": 4}],"gesamtPreisCents": 7070,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.produkte-geliefert:v1', 'tisch:16', 51, now() - interval '22 hours 59 minutes', '{"lieferungId": "c0016000-0000-0000-0000-000000000017","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000058","menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000059","menge": 3},{"positionId": "b0016000-0000-0000-0000-000000000060","menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000061","menge": 6},{"positionId": "b0016000-0000-0000-0000-000000000062","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.zahlung-registriert:v1', 'tisch:16', 52, now() - interval '22 hours 46 minutes', '{"zahlungId": "d0016000-0000-0000-0000-000000000017","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000058","menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000059","menge": 3},{"positionId": "b0016000-0000-0000-0000-000000000060","menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000061","menge": 6},{"positionId": "b0016000-0000-0000-0000-000000000062","menge": 4}],"gesamtZahlungCents": 7070,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:16', 53, now() - interval '22 hours 32 minutes', '{"bestellungId": "a0016000-0000-0000-0000-000000000018","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000063","varianteId": 13,"produktName": "Grillplatte","varianteName": "Groß","kategorie": "food","einzelpreis": 1400,"menge": 2},{"positionId": "b0016000-0000-0000-0000-000000000064","varianteId": 15,"produktName": "Salat","varianteName": "Caesar Salat","kategorie": "food","einzelpreis": 650,"menge": 3},{"positionId": "b0016000-0000-0000-0000-000000000065","varianteId": 42,"produktName": "Wein","varianteName": "Weißwein 0,2l","kategorie": "beverage","einzelpreis": 400,"menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000066","varianteId": 36,"produktName": "Wasser","varianteName": "Still 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 4}],"gesamtPreisCents": 7150,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:16', 54, now() - interval '22 hours 22 minutes', '{"lieferungId": "c0016000-0000-0000-0000-000000000018","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000063","menge": 2},{"positionId": "b0016000-0000-0000-0000-000000000064","menge": 3},{"positionId": "b0016000-0000-0000-0000-000000000065","menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000066","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:16', 55, now() - interval '21 hours 59 minutes', '{"zahlungId": "d0016000-0000-0000-0000-000000000018","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000063","menge": 2},{"positionId": "b0016000-0000-0000-0000-000000000064","menge": 3},{"positionId": "b0016000-0000-0000-0000-000000000065","menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000066","menge": 4}],"gesamtZahlungCents": 7150,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:16', 56, now() - interval '21 hours 38 minutes', '{"bestellungId": "a0016000-0000-0000-0000-000000000019","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000067","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 8},{"positionId": "b0016000-0000-0000-0000-000000000068","varianteId": 25,"produktName": "Bier","varianteName": "Maß 1,0l","kategorie": "beverage","einzelpreis": 850,"menge": 6},{"positionId": "b0016000-0000-0000-0000-000000000069","varianteId": 27,"produktName": "Weizen","varianteName": "Groß 0,5l","kategorie": "beverage","einzelpreis": 400,"menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000070","varianteId": 30,"produktName": "Weizen","varianteName": "Russ","kategorie": "beverage","einzelpreis": 300,"menge": 4}],"gesamtPreisCents": 11500,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.produkte-geliefert:v1', 'tisch:16', 57, now() - interval '21 hours 29 minutes', '{"lieferungId": "c0016000-0000-0000-0000-000000000019","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000067","menge": 8},{"positionId": "b0016000-0000-0000-0000-000000000068","menge": 6},{"positionId": "b0016000-0000-0000-0000-000000000069","menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000070","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.zahlung-registriert:v1', 'tisch:16', 58, now() - interval '21 hours 14 minutes', '{"zahlungId": "d0016000-0000-0000-0000-000000000019","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000067","menge": 8},{"positionId": "b0016000-0000-0000-0000-000000000068","menge": 6},{"positionId": "b0016000-0000-0000-0000-000000000069","menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000070","menge": 4}],"gesamtZahlungCents": 11500,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:16', 59, now() - interval '21 hours 1 minutes', '{"bestellungId": "a0016000-0000-0000-0000-000000000020","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000071","varianteId": 20,"produktName": "Brezel","varianteName": "Normal","kategorie": "food","einzelpreis": 200,"menge": 8},{"positionId": "b0016000-0000-0000-0000-000000000072","varianteId": 21,"produktName": "Brezel","varianteName": "mit Butter","kategorie": "food","einzelpreis": 300,"menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000073","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 6}],"gesamtPreisCents": 5500,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.produkte-geliefert:v1', 'tisch:16', 60, now() - interval '20 hours 51 minutes', '{"lieferungId": "c0016000-0000-0000-0000-000000000020","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000071","menge": 8},{"positionId": "b0016000-0000-0000-0000-000000000072","menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000073","menge": 6}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.zahlung-registriert:v1', 'tisch:16', 61, now() - interval '20 hours 35 minutes', '{"zahlungId": "d0016000-0000-0000-0000-000000000020","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000071","menge": 8},{"positionId": "b0016000-0000-0000-0000-000000000072","menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000073","menge": 6}],"gesamtZahlungCents": 5500,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:16', 62, now() - interval '20 hours 23 minutes', '{"bestellungId": "a0016000-0000-0000-0000-000000000021","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000074","varianteId": 16,"produktName": "Kuchen","varianteName": "Stück","kategorie": "food","einzelpreis": 250,"menge": 6},{"positionId": "b0016000-0000-0000-0000-000000000075","varianteId": 19,"produktName": "Waffeln","varianteName": "mit Nutella","kategorie": "food","einzelpreis": 400,"menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000076","varianteId": 45,"produktName": "Kaffee","varianteName": "Tasse","kategorie": "beverage","einzelpreis": 200,"menge": 6}],"gesamtPreisCents": 4300,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:16', 63, now() - interval '20 hours 9 minutes', '{"lieferungId": "c0016000-0000-0000-0000-000000000021","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000074","menge": 6},{"positionId": "b0016000-0000-0000-0000-000000000075","menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000076","menge": 6}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:16', 64, now() - interval '19 hours 47 minutes', '{"zahlungId": "d0016000-0000-0000-0000-000000000021","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000074","menge": 6},{"positionId": "b0016000-0000-0000-0000-000000000075","menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000076","menge": 6}],"gesamtZahlungCents": 4300,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:16', 65, now() - interval '19 hours 36 minutes', '{"bestellungId": "a0016000-0000-0000-0000-000000000022","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000077","varianteId": 25,"produktName": "Bier","varianteName": "Maß 1,0l","kategorie": "beverage","einzelpreis": 850,"menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000078","varianteId": 48,"produktName": "Hugo/Aperol","varianteName": "Hugo","kategorie": "beverage","einzelpreis": 550,"menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000079","varianteId": 49,"produktName": "Hugo/Aperol","varianteName": "Aperol Spritz","kategorie": "beverage","einzelpreis": 550,"menge": 4}],"gesamtPreisCents": 7800,"kommentar": "Letzte Runde Samstag"}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.produkte-geliefert:v1', 'tisch:16', 66, now() - interval '19 hours 26 minutes', '{"lieferungId": "c0016000-0000-0000-0000-000000000022","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000077","menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000078","menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000079","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.zahlung-registriert:v1', 'tisch:16', 67, now() - interval '19 hours 9 minutes', '{"zahlungId": "d0016000-0000-0000-0000-000000000022","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000077","menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000078","menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000079","menge": 4}],"gesamtZahlungCents": 7800,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.bestellung-aufgegeben:v1', 'tisch:16', 68, now() - interval '4 hours 50 minutes', '{"bestellungId": "a0016000-0000-0000-0000-000000000023","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000080","varianteId": 11,"produktName": "Tagesgericht","varianteName": "So: Hähnchen mit Reis","kategorie": "food","einzelpreis": 1050,"menge": 5},{"positionId": "b0016000-0000-0000-0000-000000000081","varianteId": 14,"produktName": "Salat","varianteName": "Gemischter Salat","kategorie": "food","einzelpreis": 550,"menge": 3},{"positionId": "b0016000-0000-0000-0000-000000000082","varianteId": 37,"produktName": "Wasser","varianteName": "Medium 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 6},{"positionId": "b0016000-0000-0000-0000-000000000083","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 4}],"gesamtPreisCents": 9900,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.produkte-geliefert:v1', 'tisch:16', 69, now() - interval '4 hours 37 minutes', '{"lieferungId": "c0016000-0000-0000-0000-000000000023","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000080","menge": 5},{"positionId": "b0016000-0000-0000-0000-000000000081","menge": 3},{"positionId": "b0016000-0000-0000-0000-000000000082","menge": 6},{"positionId": "b0016000-0000-0000-0000-000000000083","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:16', 70, now() - interval '4 hours 19 minutes', '{"bestellungId": "a0016000-0000-0000-0000-000000000024","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000084","varianteId": 16,"produktName": "Kuchen","varianteName": "Stück","kategorie": "food","einzelpreis": 250,"menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000085","varianteId": 17,"produktName": "Waffeln","varianteName": "mit Puderzucker","kategorie": "food","einzelpreis": 300,"menge": 3},{"positionId": "b0016000-0000-0000-0000-000000000086","varianteId": 45,"produktName": "Kaffee","varianteName": "Tasse","kategorie": "beverage","einzelpreis": 200,"menge": 4}],"gesamtPreisCents": 2700,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:16', 71, now() - interval '4 hours 8 minutes', '{"lieferungId": "c0016000-0000-0000-0000-000000000024","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000084","menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000085","menge": 3},{"positionId": "b0016000-0000-0000-0000-000000000086","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.zahlung-registriert:v1', 'tisch:16', 72, now() - interval '3 hours 50 minutes', '{"zahlungId": "d0016000-0000-0000-0000-000000000023","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000080","menge": 5},{"positionId": "b0016000-0000-0000-0000-000000000081","menge": 3},{"positionId": "b0016000-0000-0000-0000-000000000082","menge": 6},{"positionId": "b0016000-0000-0000-0000-000000000083","menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000084","menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000085","menge": 3},{"positionId": "b0016000-0000-0000-0000-000000000086","menge": 4}],"gesamtZahlungCents": 12600,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:16', 73, now() - interval '3 hours 40 minutes', '{"bestellungId": "a0016000-0000-0000-0000-000000000025","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000087","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 6},{"positionId": "b0016000-0000-0000-0000-000000000088","varianteId": 26,"produktName": "Weizen","varianteName": "Klein 0,3l","kategorie": "beverage","einzelpreis": 300,"menge": 3},{"positionId": "b0016000-0000-0000-0000-000000000089","varianteId": 40,"produktName": "Saftschorle","varianteName": "Johannisbeerschorle 0,5l","kategorie": "beverage","einzelpreis": 300,"menge": 3}],"gesamtPreisCents": 4500,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.produkte-geliefert:v1', 'tisch:16', 74, now() - interval '3 hours 28 minutes', '{"lieferungId": "c0016000-0000-0000-0000-000000000025","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000087","menge": 6},{"positionId": "b0016000-0000-0000-0000-000000000088","menge": 3},{"positionId": "b0016000-0000-0000-0000-000000000089","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:16', 75, now() - interval '3 hours 14 minutes', '{"bestellungId": "a0016000-0000-0000-0000-000000000026","positionen": [{"positionId": "b0016000-0000-0000-0000-000000000090","varianteId": 1,"produktName": "Bratwurst","varianteName": "Normal","kategorie": "food","einzelpreis": 350,"menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000091","varianteId": 5,"produktName": "Pommes","varianteName": "Groß","kategorie": "food","einzelpreis": 350,"menge": 4},{"positionId": "b0016000-0000-0000-0000-000000000092","varianteId": 33,"produktName": "Softdrinks","varianteName": "Spezi","kategorie": "beverage","einzelpreis": 280,"menge": 4}],"gesamtPreisCents": 3920,"kommentar": "Nachbestellung"}');

-- ─────────────────────────────────────────────────────────────────────────────
-- TISCH 17: Zelt A2 — Tag 2+3, mittel (31 Events)
-- ─────────────────────────────────────────────────────────────────────────────

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.bestellung-aufgegeben:v1', 'tisch:17', 1, now() - interval '25 hours 50 minutes', '{"bestellungId": "a0017000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0017000-0000-0000-0000-000000000001","varianteId": 10,"produktName": "Tagesgericht","varianteName": "Sa: Gulasch mit Spätzle","kategorie": "food","einzelpreis": 1150,"menge": 3},{"positionId": "b0017000-0000-0000-0000-000000000002","varianteId": 7,"produktName": "Flammkuchen","varianteName": "Speck & Zwiebel","kategorie": "food","einzelpreis": 700,"menge": 2},{"positionId": "b0017000-0000-0000-0000-000000000003","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 4},{"positionId": "b0017000-0000-0000-0000-000000000004","varianteId": 27,"produktName": "Weizen","varianteName": "Groß 0,5l","kategorie": "beverage","einzelpreis": 400,"menge": 2}],"gesamtPreisCents": 7450,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.produkte-geliefert:v1', 'tisch:17', 2, now() - interval '25 hours 40 minutes', '{"lieferungId": "c0017000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0017000-0000-0000-0000-000000000001","menge": 3},{"positionId": "b0017000-0000-0000-0000-000000000002","menge": 2},{"positionId": "b0017000-0000-0000-0000-000000000003","menge": 4},{"positionId": "b0017000-0000-0000-0000-000000000004","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.zahlung-registriert:v1', 'tisch:17', 3, now() - interval '25 hours 19 minutes', '{"zahlungId": "d0017000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0017000-0000-0000-0000-000000000001","menge": 3},{"positionId": "b0017000-0000-0000-0000-000000000002","menge": 2},{"positionId": "b0017000-0000-0000-0000-000000000003","menge": 4},{"positionId": "b0017000-0000-0000-0000-000000000004","menge": 2}],"gesamtZahlungCents": 7450,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:17', 4, now() - interval '25 hours', '{"bestellungId": "a0017000-0000-0000-0000-000000000002","positionen": [{"positionId": "b0017000-0000-0000-0000-000000000005","varianteId": 1,"produktName": "Bratwurst","varianteName": "Normal","kategorie": "food","einzelpreis": 350,"menge": 4},{"positionId": "b0017000-0000-0000-0000-000000000006","varianteId": 5,"produktName": "Pommes","varianteName": "Groß","kategorie": "food","einzelpreis": 350,"menge": 3},{"positionId": "b0017000-0000-0000-0000-000000000007","varianteId": 23,"produktName": "Bier","varianteName": "0,3l","kategorie": "beverage","einzelpreis": 300,"menge": 6},{"positionId": "b0017000-0000-0000-0000-000000000008","varianteId": 33,"produktName": "Softdrinks","varianteName": "Spezi","kategorie": "beverage","einzelpreis": 280,"menge": 3}],"gesamtPreisCents": 5090,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:17', 5, now() - interval '24 hours 50 minutes', '{"lieferungId": "c0017000-0000-0000-0000-000000000002","positionen": [{"positionId": "b0017000-0000-0000-0000-000000000005","menge": 4},{"positionId": "b0017000-0000-0000-0000-000000000006","menge": 3},{"positionId": "b0017000-0000-0000-0000-000000000007","menge": 6},{"positionId": "b0017000-0000-0000-0000-000000000008","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:17', 6, now() - interval '24 hours 30 minutes', '{"zahlungId": "d0017000-0000-0000-0000-000000000002","positionen": [{"positionId": "b0017000-0000-0000-0000-000000000005","menge": 4},{"positionId": "b0017000-0000-0000-0000-000000000006","menge": 3},{"positionId": "b0017000-0000-0000-0000-000000000007","menge": 6},{"positionId": "b0017000-0000-0000-0000-000000000008","menge": 3}],"gesamtZahlungCents": 5090,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.bestellung-aufgegeben:v1', 'tisch:17', 7, now() - interval '24 hours 11 minutes', '{"bestellungId": "a0017000-0000-0000-0000-000000000003","positionen": [{"positionId": "b0017000-0000-0000-0000-000000000009","varianteId": 12,"produktName": "Grillplatte","varianteName": "Klein","kategorie": "food","einzelpreis": 800,"menge": 2},{"positionId": "b0017000-0000-0000-0000-000000000010","varianteId": 14,"produktName": "Salat","varianteName": "Gemischter Salat","kategorie": "food","einzelpreis": 550,"menge": 2},{"positionId": "b0017000-0000-0000-0000-000000000011","varianteId": 36,"produktName": "Wasser","varianteName": "Still 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 3},{"positionId": "b0017000-0000-0000-0000-000000000012","varianteId": 42,"produktName": "Wein","varianteName": "Weißwein 0,2l","kategorie": "beverage","einzelpreis": 400,"menge": 2}],"gesamtPreisCents": 4100,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.produkte-geliefert:v1', 'tisch:17', 8, now() - interval '24 hours', '{"lieferungId": "c0017000-0000-0000-0000-000000000003","positionen": [{"positionId": "b0017000-0000-0000-0000-000000000009","menge": 2},{"positionId": "b0017000-0000-0000-0000-000000000010","menge": 2},{"positionId": "b0017000-0000-0000-0000-000000000011","menge": 3},{"positionId": "b0017000-0000-0000-0000-000000000012","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.zahlung-registriert:v1', 'tisch:17', 9, now() - interval '23 hours 40 minutes', '{"zahlungId": "d0017000-0000-0000-0000-000000000003","positionen": [{"positionId": "b0017000-0000-0000-0000-000000000009","menge": 2},{"positionId": "b0017000-0000-0000-0000-000000000010","menge": 2},{"positionId": "b0017000-0000-0000-0000-000000000011","menge": 3},{"positionId": "b0017000-0000-0000-0000-000000000012","menge": 2}],"gesamtZahlungCents": 4100,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:17', 10, now() - interval '23 hours 23 minutes', '{"bestellungId": "a0017000-0000-0000-0000-000000000004","positionen": [{"positionId": "b0017000-0000-0000-0000-000000000013","varianteId": 8,"produktName": "Flammkuchen","varianteName": "Mediterran","kategorie": "food","einzelpreis": 750,"menge": 2},{"positionId": "b0017000-0000-0000-0000-000000000014","varianteId": 48,"produktName": "Hugo/Aperol","varianteName": "Hugo","kategorie": "beverage","einzelpreis": 550,"menge": 3},{"positionId": "b0017000-0000-0000-0000-000000000015","varianteId": 49,"produktName": "Hugo/Aperol","varianteName": "Aperol Spritz","kategorie": "beverage","einzelpreis": 550,"menge": 2}],"gesamtPreisCents": 4250,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:17', 11, now() - interval '23 hours 9 minutes', '{"lieferungId": "c0017000-0000-0000-0000-000000000004","positionen": [{"positionId": "b0017000-0000-0000-0000-000000000013","menge": 2},{"positionId": "b0017000-0000-0000-0000-000000000014","menge": 3},{"positionId": "b0017000-0000-0000-0000-000000000015","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:17', 12, now() - interval '22 hours 46 minutes', '{"zahlungId": "d0017000-0000-0000-0000-000000000004","positionen": [{"positionId": "b0017000-0000-0000-0000-000000000013","menge": 2},{"positionId": "b0017000-0000-0000-0000-000000000014","menge": 3},{"positionId": "b0017000-0000-0000-0000-000000000015","menge": 2}],"gesamtZahlungCents": 4250,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.bestellung-aufgegeben:v1', 'tisch:17', 13, now() - interval '22 hours 28 minutes', '{"bestellungId": "a0017000-0000-0000-0000-000000000005","positionen": [{"positionId": "b0017000-0000-0000-0000-000000000016","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 6},{"positionId": "b0017000-0000-0000-0000-000000000017","varianteId": 25,"produktName": "Bier","varianteName": "Maß 1,0l","kategorie": "beverage","einzelpreis": 850,"menge": 2},{"positionId": "b0017000-0000-0000-0000-000000000018","varianteId": 30,"produktName": "Weizen","varianteName": "Russ","kategorie": "beverage","einzelpreis": 300,"menge": 3}],"gesamtPreisCents": 5300,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.produkte-geliefert:v1', 'tisch:17', 14, now() - interval '22 hours 18 minutes', '{"lieferungId": "c0017000-0000-0000-0000-000000000005","positionen": [{"positionId": "b0017000-0000-0000-0000-000000000016","menge": 6},{"positionId": "b0017000-0000-0000-0000-000000000017","menge": 2},{"positionId": "b0017000-0000-0000-0000-000000000018","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.zahlung-registriert:v1', 'tisch:17', 15, now() - interval '21 hours 53 minutes', '{"zahlungId": "d0017000-0000-0000-0000-000000000005","positionen": [{"positionId": "b0017000-0000-0000-0000-000000000016","menge": 6},{"positionId": "b0017000-0000-0000-0000-000000000017","menge": 2},{"positionId": "b0017000-0000-0000-0000-000000000018","menge": 3}],"gesamtZahlungCents": 5300,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:17', 16, now() - interval '21 hours 36 minutes', '{"bestellungId": "a0017000-0000-0000-0000-000000000006","positionen": [{"positionId": "b0017000-0000-0000-0000-000000000019","varianteId": 16,"produktName": "Kuchen","varianteName": "Stück","kategorie": "food","einzelpreis": 250,"menge": 4},{"positionId": "b0017000-0000-0000-0000-000000000020","varianteId": 18,"produktName": "Waffeln","varianteName": "mit Sahne","kategorie": "food","einzelpreis": 350,"menge": 3},{"positionId": "b0017000-0000-0000-0000-000000000021","varianteId": 45,"produktName": "Kaffee","varianteName": "Tasse","kategorie": "beverage","einzelpreis": 200,"menge": 4}],"gesamtPreisCents": 2850,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:17', 17, now() - interval '21 hours 25 minutes', '{"lieferungId": "c0017000-0000-0000-0000-000000000006","positionen": [{"positionId": "b0017000-0000-0000-0000-000000000019","menge": 4},{"positionId": "b0017000-0000-0000-0000-000000000020","menge": 3},{"positionId": "b0017000-0000-0000-0000-000000000021","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:17', 18, now() - interval '21 hours 9 minutes', '{"zahlungId": "d0017000-0000-0000-0000-000000000006","positionen": [{"positionId": "b0017000-0000-0000-0000-000000000019","menge": 4},{"positionId": "b0017000-0000-0000-0000-000000000020","menge": 3},{"positionId": "b0017000-0000-0000-0000-000000000021","menge": 4}],"gesamtZahlungCents": 2850,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.bestellung-aufgegeben:v1', 'tisch:17', 19, now() - interval '20 hours 54 minutes', '{"bestellungId": "a0017000-0000-0000-0000-000000000007","positionen": [{"positionId": "b0017000-0000-0000-0000-000000000022","varianteId": 3,"produktName": "Bratwurst","varianteName": "Currywurst","kategorie": "food","einzelpreis": 450,"menge": 3},{"positionId": "b0017000-0000-0000-0000-000000000023","varianteId": 23,"produktName": "Bier","varianteName": "0,3l","kategorie": "beverage","einzelpreis": 300,"menge": 6}],"gesamtPreisCents": 3150,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.produkte-geliefert:v1', 'tisch:17', 20, now() - interval '20 hours 45 minutes', '{"lieferungId": "c0017000-0000-0000-0000-000000000007","positionen": [{"positionId": "b0017000-0000-0000-0000-000000000022","menge": 3},{"positionId": "b0017000-0000-0000-0000-000000000023","menge": 6}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.zahlung-registriert:v1', 'tisch:17', 21, now() - interval '20 hours 27 minutes', '{"zahlungId": "d0017000-0000-0000-0000-000000000007","positionen": [{"positionId": "b0017000-0000-0000-0000-000000000022","menge": 3},{"positionId": "b0017000-0000-0000-0000-000000000023","menge": 6}],"gesamtZahlungCents": 3150,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:17', 22, now() - interval '20 hours 6 minutes', '{"bestellungId": "a0017000-0000-0000-0000-000000000008","positionen": [{"positionId": "b0017000-0000-0000-0000-000000000024","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 4},{"positionId": "b0017000-0000-0000-0000-000000000025","varianteId": 26,"produktName": "Weizen","varianteName": "Klein 0,3l","kategorie": "beverage","einzelpreis": 300,"menge": 2}],"gesamtPreisCents": 2400,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:17', 23, now() - interval '19 hours 58 minutes', '{"lieferungId": "c0017000-0000-0000-0000-000000000008","positionen": [{"positionId": "b0017000-0000-0000-0000-000000000024","menge": 4},{"positionId": "b0017000-0000-0000-0000-000000000025","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:17', 24, now() - interval '19 hours 39 minutes', '{"zahlungId": "d0017000-0000-0000-0000-000000000008","positionen": [{"positionId": "b0017000-0000-0000-0000-000000000024","menge": 4},{"positionId": "b0017000-0000-0000-0000-000000000025","menge": 2}],"gesamtZahlungCents": 2400,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:17', 25, now() - interval '4 hours 25 minutes', '{"bestellungId": "a0017000-0000-0000-0000-000000000009","positionen": [{"positionId": "b0017000-0000-0000-0000-000000000026","varianteId": 11,"produktName": "Tagesgericht","varianteName": "So: Hähnchen mit Reis","kategorie": "food","einzelpreis": 1050,"menge": 3},{"positionId": "b0017000-0000-0000-0000-000000000027","varianteId": 15,"produktName": "Salat","varianteName": "Caesar Salat","kategorie": "food","einzelpreis": 650,"menge": 2},{"positionId": "b0017000-0000-0000-0000-000000000028","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 3},{"positionId": "b0017000-0000-0000-0000-000000000029","varianteId": 37,"produktName": "Wasser","varianteName": "Medium 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 3}],"gesamtPreisCents": 6400,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:17', 26, now() - interval '4 hours 13 minutes', '{"lieferungId": "c0017000-0000-0000-0000-000000000009","positionen": [{"positionId": "b0017000-0000-0000-0000-000000000026","menge": 3},{"positionId": "b0017000-0000-0000-0000-000000000027","menge": 2},{"positionId": "b0017000-0000-0000-0000-000000000028","menge": 3},{"positionId": "b0017000-0000-0000-0000-000000000029","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:17', 27, now() - interval '3 hours 50 minutes', '{"zahlungId": "d0017000-0000-0000-0000-000000000009","positionen": [{"positionId": "b0017000-0000-0000-0000-000000000026","menge": 3},{"positionId": "b0017000-0000-0000-0000-000000000027","menge": 2},{"positionId": "b0017000-0000-0000-0000-000000000028","menge": 3},{"positionId": "b0017000-0000-0000-0000-000000000029","menge": 3}],"gesamtZahlungCents": 6400,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.bestellung-aufgegeben:v1', 'tisch:17', 28, now() - interval '3 hours 30 minutes', '{"bestellungId": "a0017000-0000-0000-0000-000000000010","positionen": [{"positionId": "b0017000-0000-0000-0000-000000000030","varianteId": 17,"produktName": "Waffeln","varianteName": "mit Puderzucker","kategorie": "food","einzelpreis": 300,"menge": 3},{"positionId": "b0017000-0000-0000-0000-000000000031","varianteId": 45,"produktName": "Kaffee","varianteName": "Tasse","kategorie": "beverage","einzelpreis": 200,"menge": 3}],"gesamtPreisCents": 1500,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.produkte-geliefert:v1', 'tisch:17', 29, now() - interval '3 hours 15 minutes', '{"lieferungId": "c0017000-0000-0000-0000-000000000010","positionen": [{"positionId": "b0017000-0000-0000-0000-000000000030","menge": 3},{"positionId": "b0017000-0000-0000-0000-000000000031","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:17', 30, now() - interval '3 hours 2 minutes', '{"bestellungId": "a0017000-0000-0000-0000-000000000011","positionen": [{"positionId": "b0017000-0000-0000-0000-000000000032","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 4},{"positionId": "b0017000-0000-0000-0000-000000000033","varianteId": 41,"produktName": "Saftschorle","varianteName": "Rhabarberschorle 0,5l","kategorie": "beverage","einzelpreis": 350,"menge": 2}],"gesamtPreisCents": 2500,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:17', 31, now() - interval '2 hours 50 minutes', '{"lieferungId": "c0017000-0000-0000-0000-000000000011","positionen": [{"positionId": "b0017000-0000-0000-0000-000000000032","menge": 4},{"positionId": "b0017000-0000-0000-0000-000000000033","menge": 2}],"kommentar": ""}');

-- ─────────────────────────────────────────────────────────────────────────────
-- TISCH 18: Stehtisch Bar — schneller Umsatz, alle 3 Tage (90 Events)
-- ─────────────────────────────────────────────────────────────────────────────

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:18', 1, now() - interval '50 hours 50 minutes', '{"bestellungId": "a0018000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000001","varianteId": 37,"produktName": "Wasser","varianteName": "Medium 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 1},{"positionId": "b0018000-0000-0000-0000-000000000002","varianteId": 44,"produktName": "Wein","varianteName": "Rosé 0,2l","kategorie": "beverage","einzelpreis": 400,"menge": 3},{"positionId": "b0018000-0000-0000-0000-000000000003","varianteId": 28,"produktName": "Weizen","varianteName": "Colaweizen Klein","kategorie": "beverage","einzelpreis": 300,"menge": 4},{"positionId": "b0018000-0000-0000-0000-000000000004","varianteId": 18,"produktName": "Waffeln","varianteName": "mit Sahne","kategorie": "food","einzelpreis": 350,"menge": 1}],"gesamtPreisCents": 2950,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.produkte-geliefert:v1', 'tisch:18', 2, now() - interval '50 hours 45 minutes', '{"lieferungId": "c0018000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000001","menge": 1},{"positionId": "b0018000-0000-0000-0000-000000000002","menge": 3},{"positionId": "b0018000-0000-0000-0000-000000000003","menge": 4},{"positionId": "b0018000-0000-0000-0000-000000000004","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.zahlung-registriert:v1', 'tisch:18', 3, now() - interval '50 hours 40 minutes', '{"zahlungId": "d0018000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000001","menge": 1},{"positionId": "b0018000-0000-0000-0000-000000000002","menge": 3},{"positionId": "b0018000-0000-0000-0000-000000000003","menge": 4},{"positionId": "b0018000-0000-0000-0000-000000000004","menge": 1}],"gesamtZahlungCents": 2950,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.bestellung-aufgegeben:v1', 'tisch:18', 4, now() - interval '50 hours 15 minutes', '{"bestellungId": "a0018000-0000-0000-0000-000000000002","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000005","varianteId": 46,"produktName": "Kaffee","varianteName": "Espresso","kategorie": "beverage","einzelpreis": 180,"menge": 3},{"positionId": "b0018000-0000-0000-0000-000000000006","varianteId": 29,"produktName": "Weizen","varianteName": "Colaweizen Groß","kategorie": "beverage","einzelpreis": 400,"menge": 2},{"positionId": "b0018000-0000-0000-0000-000000000007","varianteId": 37,"produktName": "Wasser","varianteName": "Medium 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 4}],"gesamtPreisCents": 2140,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.produkte-geliefert:v1', 'tisch:18', 5, now() - interval '50 hours 12 minutes', '{"lieferungId": "c0018000-0000-0000-0000-000000000002","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000005","menge": 3},{"positionId": "b0018000-0000-0000-0000-000000000006","menge": 2},{"positionId": "b0018000-0000-0000-0000-000000000007","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.zahlung-registriert:v1', 'tisch:18', 6, now() - interval '50 hours 9 minutes', '{"zahlungId": "d0018000-0000-0000-0000-000000000002","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000005","menge": 3},{"positionId": "b0018000-0000-0000-0000-000000000006","menge": 2},{"positionId": "b0018000-0000-0000-0000-000000000007","menge": 4}],"gesamtZahlungCents": 2140,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:18', 7, now() - interval '49 hours 43 minutes', '{"bestellungId": "a0018000-0000-0000-0000-000000000003","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000008","varianteId": 32,"produktName": "Softdrinks","varianteName": "Fanta","kategorie": "beverage","einzelpreis": 280,"menge": 1},{"positionId": "b0018000-0000-0000-0000-000000000009","varianteId": 3,"produktName": "Bratwurst","varianteName": "Currywurst","kategorie": "food","einzelpreis": 450,"menge": 1}],"gesamtPreisCents": 730,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.produkte-geliefert:v1', 'tisch:18', 8, now() - interval '49 hours 37 minutes', '{"lieferungId": "c0018000-0000-0000-0000-000000000003","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000008","menge": 1},{"positionId": "b0018000-0000-0000-0000-000000000009","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.zahlung-registriert:v1', 'tisch:18', 9, now() - interval '49 hours 32 minutes', '{"zahlungId": "d0018000-0000-0000-0000-000000000003","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000008","menge": 1},{"positionId": "b0018000-0000-0000-0000-000000000009","menge": 1}],"gesamtZahlungCents": 730,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:18', 10, now() - interval '49 hours 6 minutes', '{"bestellungId": "a0018000-0000-0000-0000-000000000004","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000010","varianteId": 44,"produktName": "Wein","varianteName": "Rosé 0,2l","kategorie": "beverage","einzelpreis": 400,"menge": 3},{"positionId": "b0018000-0000-0000-0000-000000000011","varianteId": 43,"produktName": "Wein","varianteName": "Rotwein 0,2l","kategorie": "beverage","einzelpreis": 400,"menge": 3},{"positionId": "b0018000-0000-0000-0000-000000000012","varianteId": 38,"produktName": "Wasser","varianteName": "Sprudel 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 3},{"positionId": "b0018000-0000-0000-0000-000000000013","varianteId": 7,"produktName": "Flammkuchen","varianteName": "Speck & Zwiebel","kategorie": "food","einzelpreis": 700,"menge": 1}],"gesamtPreisCents": 3700,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.produkte-geliefert:v1', 'tisch:18', 11, now() - interval '49 hours', '{"lieferungId": "c0018000-0000-0000-0000-000000000004","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000010","menge": 3},{"positionId": "b0018000-0000-0000-0000-000000000011","menge": 3},{"positionId": "b0018000-0000-0000-0000-000000000012","menge": 3},{"positionId": "b0018000-0000-0000-0000-000000000013","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.zahlung-registriert:v1', 'tisch:18', 12, now() - interval '48 hours 57 minutes', '{"zahlungId": "d0018000-0000-0000-0000-000000000004","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000010","menge": 3},{"positionId": "b0018000-0000-0000-0000-000000000011","menge": 3},{"positionId": "b0018000-0000-0000-0000-000000000012","menge": 3},{"positionId": "b0018000-0000-0000-0000-000000000013","menge": 1}],"gesamtZahlungCents": 3700,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.bestellung-aufgegeben:v1', 'tisch:18', 13, now() - interval '48 hours 27 minutes', '{"bestellungId": "a0018000-0000-0000-0000-000000000005","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000014","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 2},{"positionId": "b0018000-0000-0000-0000-000000000015","varianteId": 42,"produktName": "Wein","varianteName": "Weißwein 0,2l","kategorie": "beverage","einzelpreis": 400,"menge": 3}],"gesamtPreisCents": 2100,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.produkte-geliefert:v1', 'tisch:18', 14, now() - interval '48 hours 23 minutes', '{"lieferungId": "c0018000-0000-0000-0000-000000000005","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000014","menge": 2},{"positionId": "b0018000-0000-0000-0000-000000000015","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.zahlung-registriert:v1', 'tisch:18', 15, now() - interval '48 hours 21 minutes', '{"zahlungId": "d0018000-0000-0000-0000-000000000005","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000014","menge": 2},{"positionId": "b0018000-0000-0000-0000-000000000015","menge": 3}],"gesamtZahlungCents": 2100,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.bestellung-aufgegeben:v1', 'tisch:18', 16, now() - interval '48 hours 1 minutes', '{"bestellungId": "a0018000-0000-0000-0000-000000000006","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000016","varianteId": 47,"produktName": "Tee","varianteName": "Verschiedene Sorten","kategorie": "beverage","einzelpreis": 200,"menge": 3},{"positionId": "b0018000-0000-0000-0000-000000000017","varianteId": 29,"produktName": "Weizen","varianteName": "Colaweizen Groß","kategorie": "beverage","einzelpreis": 400,"menge": 2},{"positionId": "b0018000-0000-0000-0000-000000000018","varianteId": 17,"produktName": "Waffeln","varianteName": "mit Puderzucker","kategorie": "food","einzelpreis": 300,"menge": 1}],"gesamtPreisCents": 1700,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.produkte-geliefert:v1', 'tisch:18', 17, now() - interval '47 hours 58 minutes', '{"lieferungId": "c0018000-0000-0000-0000-000000000006","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000016","menge": 3},{"positionId": "b0018000-0000-0000-0000-000000000017","menge": 2},{"positionId": "b0018000-0000-0000-0000-000000000018","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.zahlung-registriert:v1', 'tisch:18', 18, now() - interval '47 hours 55 minutes', '{"zahlungId": "d0018000-0000-0000-0000-000000000006","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000016","menge": 3},{"positionId": "b0018000-0000-0000-0000-000000000017","menge": 2},{"positionId": "b0018000-0000-0000-0000-000000000018","menge": 1}],"gesamtZahlungCents": 1700,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:18', 19, now() - interval '47 hours 22 minutes', '{"bestellungId": "a0018000-0000-0000-0000-000000000007","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000019","varianteId": 35,"produktName": "Softdrinks","varianteName": "Mezzo Mix","kategorie": "beverage","einzelpreis": 280,"menge": 1},{"positionId": "b0018000-0000-0000-0000-000000000020","varianteId": 36,"produktName": "Wasser","varianteName": "Still 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 2},{"positionId": "b0018000-0000-0000-0000-000000000021","varianteId": 5,"produktName": "Pommes","varianteName": "Groß","kategorie": "food","einzelpreis": 350,"menge": 1}],"gesamtPreisCents": 1030,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.produkte-geliefert:v1', 'tisch:18', 20, now() - interval '47 hours 17 minutes', '{"lieferungId": "c0018000-0000-0000-0000-000000000007","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000019","menge": 1},{"positionId": "b0018000-0000-0000-0000-000000000020","menge": 2},{"positionId": "b0018000-0000-0000-0000-000000000021","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.zahlung-registriert:v1', 'tisch:18', 21, now() - interval '47 hours 15 minutes', '{"zahlungId": "d0018000-0000-0000-0000-000000000007","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000019","menge": 1},{"positionId": "b0018000-0000-0000-0000-000000000020","menge": 2},{"positionId": "b0018000-0000-0000-0000-000000000021","menge": 1}],"gesamtZahlungCents": 1030,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.bestellung-aufgegeben:v1', 'tisch:18', 22, now() - interval '46 hours 43 minutes', '{"bestellungId": "a0018000-0000-0000-0000-000000000008","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000022","varianteId": 30,"produktName": "Weizen","varianteName": "Russ","kategorie": "beverage","einzelpreis": 300,"menge": 1},{"positionId": "b0018000-0000-0000-0000-000000000023","varianteId": 33,"produktName": "Softdrinks","varianteName": "Spezi","kategorie": "beverage","einzelpreis": 280,"menge": 4}],"gesamtPreisCents": 1420,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.produkte-geliefert:v1', 'tisch:18', 23, now() - interval '46 hours 37 minutes', '{"lieferungId": "c0018000-0000-0000-0000-000000000008","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000022","menge": 1},{"positionId": "b0018000-0000-0000-0000-000000000023","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.zahlung-registriert:v1', 'tisch:18', 24, now() - interval '46 hours 32 minutes', '{"zahlungId": "d0018000-0000-0000-0000-000000000008","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000022","menge": 1},{"positionId": "b0018000-0000-0000-0000-000000000023","menge": 4}],"gesamtZahlungCents": 1420,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.bestellung-aufgegeben:v1', 'tisch:18', 25, now() - interval '28 hours 17 minutes', '{"bestellungId": "a0018000-0000-0000-0000-000000000009","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000024","varianteId": 39,"produktName": "Saftschorle","varianteName": "Apfelschorle 0,5l","kategorie": "beverage","einzelpreis": 300,"menge": 4},{"positionId": "b0018000-0000-0000-0000-000000000025","varianteId": 15,"produktName": "Salat","varianteName": "Caesar Salat","kategorie": "food","einzelpreis": 650,"menge": 1}],"gesamtPreisCents": 1850,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.produkte-geliefert:v1', 'tisch:18', 26, now() - interval '28 hours 11 minutes', '{"lieferungId": "c0018000-0000-0000-0000-000000000009","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000024","menge": 4},{"positionId": "b0018000-0000-0000-0000-000000000025","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.zahlung-registriert:v1', 'tisch:18', 27, now() - interval '28 hours 8 minutes', '{"zahlungId": "d0018000-0000-0000-0000-000000000009","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000024","menge": 4},{"positionId": "b0018000-0000-0000-0000-000000000025","menge": 1}],"gesamtZahlungCents": 1850,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.bestellung-aufgegeben:v1', 'tisch:18', 28, now() - interval '27 hours 38 minutes', '{"bestellungId": "a0018000-0000-0000-0000-000000000010","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000026","varianteId": 43,"produktName": "Wein","varianteName": "Rotwein 0,2l","kategorie": "beverage","einzelpreis": 400,"menge": 1},{"positionId": "b0018000-0000-0000-0000-000000000027","varianteId": 31,"produktName": "Softdrinks","varianteName": "Cola","kategorie": "beverage","einzelpreis": 280,"menge": 1},{"positionId": "b0018000-0000-0000-0000-000000000028","varianteId": 34,"produktName": "Softdrinks","varianteName": "Sprite","kategorie": "beverage","einzelpreis": 280,"menge": 3}],"gesamtPreisCents": 1520,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.produkte-geliefert:v1', 'tisch:18', 29, now() - interval '27 hours 32 minutes', '{"lieferungId": "c0018000-0000-0000-0000-000000000010","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000026","menge": 1},{"positionId": "b0018000-0000-0000-0000-000000000027","menge": 1},{"positionId": "b0018000-0000-0000-0000-000000000028","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.zahlung-registriert:v1', 'tisch:18', 30, now() - interval '27 hours 28 minutes', '{"zahlungId": "d0018000-0000-0000-0000-000000000010","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000026","menge": 1},{"positionId": "b0018000-0000-0000-0000-000000000027","menge": 1},{"positionId": "b0018000-0000-0000-0000-000000000028","menge": 3}],"gesamtZahlungCents": 1520,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:18', 31, now() - interval '26 hours 53 minutes', '{"bestellungId": "a0018000-0000-0000-0000-000000000011","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000029","varianteId": 39,"produktName": "Saftschorle","varianteName": "Apfelschorle 0,5l","kategorie": "beverage","einzelpreis": 300,"menge": 3},{"positionId": "b0018000-0000-0000-0000-000000000030","varianteId": 26,"produktName": "Weizen","varianteName": "Klein 0,3l","kategorie": "beverage","einzelpreis": 300,"menge": 1}],"gesamtPreisCents": 1200,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.produkte-geliefert:v1', 'tisch:18', 32, now() - interval '26 hours 47 minutes', '{"lieferungId": "c0018000-0000-0000-0000-000000000011","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000029","menge": 3},{"positionId": "b0018000-0000-0000-0000-000000000030","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.zahlung-registriert:v1', 'tisch:18', 33, now() - interval '26 hours 44 minutes', '{"zahlungId": "d0018000-0000-0000-0000-000000000011","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000029","menge": 3},{"positionId": "b0018000-0000-0000-0000-000000000030","menge": 1}],"gesamtZahlungCents": 1200,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.bestellung-aufgegeben:v1', 'tisch:18', 34, now() - interval '26 hours 23 minutes', '{"bestellungId": "a0018000-0000-0000-0000-000000000012","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000031","varianteId": 31,"produktName": "Softdrinks","varianteName": "Cola","kategorie": "beverage","einzelpreis": 280,"menge": 1}],"gesamtPreisCents": 280,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.produkte-geliefert:v1', 'tisch:18', 35, now() - interval '26 hours 19 minutes', '{"lieferungId": "c0018000-0000-0000-0000-000000000012","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000031","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.zahlung-registriert:v1', 'tisch:18', 36, now() - interval '26 hours 16 minutes', '{"zahlungId": "d0018000-0000-0000-0000-000000000012","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000031","menge": 1}],"gesamtZahlungCents": 280,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.bestellung-aufgegeben:v1', 'tisch:18', 37, now() - interval '25 hours 49 minutes', '{"bestellungId": "a0018000-0000-0000-0000-000000000013","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000032","varianteId": 27,"produktName": "Weizen","varianteName": "Groß 0,5l","kategorie": "beverage","einzelpreis": 400,"menge": 1}],"gesamtPreisCents": 400,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.produkte-geliefert:v1', 'tisch:18', 38, now() - interval '25 hours 46 minutes', '{"lieferungId": "c0018000-0000-0000-0000-000000000013","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000032","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.zahlung-registriert:v1', 'tisch:18', 39, now() - interval '25 hours 42 minutes', '{"zahlungId": "d0018000-0000-0000-0000-000000000013","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000032","menge": 1}],"gesamtZahlungCents": 400,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.bestellung-aufgegeben:v1', 'tisch:18', 40, now() - interval '25 hours 20 minutes', '{"bestellungId": "a0018000-0000-0000-0000-000000000014","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000033","varianteId": 31,"produktName": "Softdrinks","varianteName": "Cola","kategorie": "beverage","einzelpreis": 280,"menge": 1},{"positionId": "b0018000-0000-0000-0000-000000000034","varianteId": 48,"produktName": "Hugo/Aperol","varianteName": "Hugo","kategorie": "beverage","einzelpreis": 550,"menge": 4},{"positionId": "b0018000-0000-0000-0000-000000000035","varianteId": 8,"produktName": "Flammkuchen","varianteName": "Mediterran","kategorie": "food","einzelpreis": 750,"menge": 1}],"gesamtPreisCents": 3230,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.produkte-geliefert:v1', 'tisch:18', 41, now() - interval '25 hours 16 minutes', '{"lieferungId": "c0018000-0000-0000-0000-000000000014","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000033","menge": 1},{"positionId": "b0018000-0000-0000-0000-000000000034","menge": 4},{"positionId": "b0018000-0000-0000-0000-000000000035","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.zahlung-registriert:v1', 'tisch:18', 42, now() - interval '25 hours 12 minutes', '{"zahlungId": "d0018000-0000-0000-0000-000000000014","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000033","menge": 1},{"positionId": "b0018000-0000-0000-0000-000000000034","menge": 4},{"positionId": "b0018000-0000-0000-0000-000000000035","menge": 1}],"gesamtZahlungCents": 3230,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:18', 43, now() - interval '24 hours 56 minutes', '{"bestellungId": "a0018000-0000-0000-0000-000000000015","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000036","varianteId": 34,"produktName": "Softdrinks","varianteName": "Sprite","kategorie": "beverage","einzelpreis": 280,"menge": 4}],"gesamtPreisCents": 1120,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.produkte-geliefert:v1', 'tisch:18', 44, now() - interval '24 hours 51 minutes', '{"lieferungId": "c0018000-0000-0000-0000-000000000015","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000036","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.zahlung-registriert:v1', 'tisch:18', 45, now() - interval '24 hours 49 minutes', '{"zahlungId": "d0018000-0000-0000-0000-000000000015","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000036","menge": 4}],"gesamtZahlungCents": 1120,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:18', 46, now() - interval '24 hours 29 minutes', '{"bestellungId": "a0018000-0000-0000-0000-000000000016","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000037","varianteId": 48,"produktName": "Hugo/Aperol","varianteName": "Hugo","kategorie": "beverage","einzelpreis": 550,"menge": 1}],"gesamtPreisCents": 550,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.produkte-geliefert:v1', 'tisch:18', 47, now() - interval '24 hours 25 minutes', '{"lieferungId": "c0018000-0000-0000-0000-000000000016","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000037","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.zahlung-registriert:v1', 'tisch:18', 48, now() - interval '24 hours 20 minutes', '{"zahlungId": "d0018000-0000-0000-0000-000000000016","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000037","menge": 1}],"gesamtZahlungCents": 550,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:18', 49, now() - interval '23 hours 56 minutes', '{"bestellungId": "a0018000-0000-0000-0000-000000000017","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000038","varianteId": 49,"produktName": "Hugo/Aperol","varianteName": "Aperol Spritz","kategorie": "beverage","einzelpreis": 550,"menge": 4},{"positionId": "b0018000-0000-0000-0000-000000000039","varianteId": 42,"produktName": "Wein","varianteName": "Weißwein 0,2l","kategorie": "beverage","einzelpreis": 400,"menge": 1},{"positionId": "b0018000-0000-0000-0000-000000000040","varianteId": 16,"produktName": "Kuchen","varianteName": "Stück","kategorie": "food","einzelpreis": 250,"menge": 1}],"gesamtPreisCents": 2850,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:18', 50, now() - interval '23 hours 50 minutes', '{"lieferungId": "c0018000-0000-0000-0000-000000000017","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000038","menge": 4},{"positionId": "b0018000-0000-0000-0000-000000000039","menge": 1},{"positionId": "b0018000-0000-0000-0000-000000000040","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:18', 51, now() - interval '23 hours 46 minutes', '{"zahlungId": "d0018000-0000-0000-0000-000000000017","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000038","menge": 4},{"positionId": "b0018000-0000-0000-0000-000000000039","menge": 1},{"positionId": "b0018000-0000-0000-0000-000000000040","menge": 1}],"gesamtZahlungCents": 2850,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:18', 52, now() - interval '23 hours 19 minutes', '{"bestellungId": "a0018000-0000-0000-0000-000000000018","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000041","varianteId": 46,"produktName": "Kaffee","varianteName": "Espresso","kategorie": "beverage","einzelpreis": 180,"menge": 2},{"positionId": "b0018000-0000-0000-0000-000000000042","varianteId": 27,"produktName": "Weizen","varianteName": "Groß 0,5l","kategorie": "beverage","einzelpreis": 400,"menge": 4}],"gesamtPreisCents": 1960,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:18', 53, now() - interval '23 hours 15 minutes', '{"lieferungId": "c0018000-0000-0000-0000-000000000018","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000041","menge": 2},{"positionId": "b0018000-0000-0000-0000-000000000042","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:18', 54, now() - interval '23 hours 11 minutes', '{"zahlungId": "d0018000-0000-0000-0000-000000000018","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000041","menge": 2},{"positionId": "b0018000-0000-0000-0000-000000000042","menge": 4}],"gesamtZahlungCents": 1960,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:18', 55, now() - interval '22 hours 43 minutes', '{"bestellungId": "a0018000-0000-0000-0000-000000000019","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000043","varianteId": 35,"produktName": "Softdrinks","varianteName": "Mezzo Mix","kategorie": "beverage","einzelpreis": 280,"menge": 2},{"positionId": "b0018000-0000-0000-0000-000000000044","varianteId": 35,"produktName": "Softdrinks","varianteName": "Mezzo Mix","kategorie": "beverage","einzelpreis": 280,"menge": 4},{"positionId": "b0018000-0000-0000-0000-000000000045","varianteId": 21,"produktName": "Brezel","varianteName": "mit Butter","kategorie": "food","einzelpreis": 300,"menge": 1}],"gesamtPreisCents": 1980,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.produkte-geliefert:v1', 'tisch:18', 56, now() - interval '22 hours 40 minutes', '{"lieferungId": "c0018000-0000-0000-0000-000000000019","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000043","menge": 2},{"positionId": "b0018000-0000-0000-0000-000000000044","menge": 4},{"positionId": "b0018000-0000-0000-0000-000000000045","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.zahlung-registriert:v1', 'tisch:18', 57, now() - interval '22 hours 36 minutes', '{"zahlungId": "d0018000-0000-0000-0000-000000000019","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000043","menge": 2},{"positionId": "b0018000-0000-0000-0000-000000000044","menge": 4},{"positionId": "b0018000-0000-0000-0000-000000000045","menge": 1}],"gesamtZahlungCents": 1980,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.bestellung-aufgegeben:v1', 'tisch:18', 58, now() - interval '22 hours 11 minutes', '{"bestellungId": "a0018000-0000-0000-0000-000000000020","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000046","varianteId": 33,"produktName": "Softdrinks","varianteName": "Spezi","kategorie": "beverage","einzelpreis": 280,"menge": 1},{"positionId": "b0018000-0000-0000-0000-000000000047","varianteId": 14,"produktName": "Salat","varianteName": "Gemischter Salat","kategorie": "food","einzelpreis": 550,"menge": 1}],"gesamtPreisCents": 830,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.produkte-geliefert:v1', 'tisch:18', 59, now() - interval '22 hours 6 minutes', '{"lieferungId": "c0018000-0000-0000-0000-000000000020","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000046","menge": 1},{"positionId": "b0018000-0000-0000-0000-000000000047","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.zahlung-registriert:v1', 'tisch:18', 60, now() - interval '22 hours 1 minutes', '{"zahlungId": "d0018000-0000-0000-0000-000000000020","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000046","menge": 1},{"positionId": "b0018000-0000-0000-0000-000000000047","menge": 1}],"gesamtZahlungCents": 830,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:18', 61, now() - interval '21 hours 40 minutes', '{"bestellungId": "a0018000-0000-0000-0000-000000000021","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000048","varianteId": 27,"produktName": "Weizen","varianteName": "Groß 0,5l","kategorie": "beverage","einzelpreis": 400,"menge": 4},{"positionId": "b0018000-0000-0000-0000-000000000049","varianteId": 26,"produktName": "Weizen","varianteName": "Klein 0,3l","kategorie": "beverage","einzelpreis": 300,"menge": 4},{"positionId": "b0018000-0000-0000-0000-000000000050","varianteId": 37,"produktName": "Wasser","varianteName": "Medium 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 4},{"positionId": "b0018000-0000-0000-0000-000000000051","varianteId": 18,"produktName": "Waffeln","varianteName": "mit Sahne","kategorie": "food","einzelpreis": 350,"menge": 1}],"gesamtPreisCents": 3950,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:18', 62, now() - interval '21 hours 35 minutes', '{"lieferungId": "c0018000-0000-0000-0000-000000000021","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000048","menge": 4},{"positionId": "b0018000-0000-0000-0000-000000000049","menge": 4},{"positionId": "b0018000-0000-0000-0000-000000000050","menge": 4},{"positionId": "b0018000-0000-0000-0000-000000000051","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:18', 63, now() - interval '21 hours 32 minutes', '{"zahlungId": "d0018000-0000-0000-0000-000000000021","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000048","menge": 4},{"positionId": "b0018000-0000-0000-0000-000000000049","menge": 4},{"positionId": "b0018000-0000-0000-0000-000000000050","menge": 4},{"positionId": "b0018000-0000-0000-0000-000000000051","menge": 1}],"gesamtZahlungCents": 3950,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.bestellung-aufgegeben:v1', 'tisch:18', 64, now() - interval '21 hours 6 minutes', '{"bestellungId": "a0018000-0000-0000-0000-000000000022","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000052","varianteId": 23,"produktName": "Bier","varianteName": "0,3l","kategorie": "beverage","einzelpreis": 300,"menge": 4},{"positionId": "b0018000-0000-0000-0000-000000000053","varianteId": 46,"produktName": "Kaffee","varianteName": "Espresso","kategorie": "beverage","einzelpreis": 180,"menge": 3}],"gesamtPreisCents": 1740,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.produkte-geliefert:v1', 'tisch:18', 65, now() - interval '21 hours 2 minutes', '{"lieferungId": "c0018000-0000-0000-0000-000000000022","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000052","menge": 4},{"positionId": "b0018000-0000-0000-0000-000000000053","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.zahlung-registriert:v1', 'tisch:18', 66, now() - interval '20 hours 58 minutes', '{"zahlungId": "d0018000-0000-0000-0000-000000000022","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000052","menge": 4},{"positionId": "b0018000-0000-0000-0000-000000000053","menge": 3}],"gesamtZahlungCents": 1740,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.bestellung-aufgegeben:v1', 'tisch:18', 67, now() - interval '4 hours 40 minutes', '{"bestellungId": "a0018000-0000-0000-0000-000000000023","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000054","varianteId": 44,"produktName": "Wein","varianteName": "Rosé 0,2l","kategorie": "beverage","einzelpreis": 400,"menge": 4},{"positionId": "b0018000-0000-0000-0000-000000000055","varianteId": 43,"produktName": "Wein","varianteName": "Rotwein 0,2l","kategorie": "beverage","einzelpreis": 400,"menge": 4},{"positionId": "b0018000-0000-0000-0000-000000000056","varianteId": 34,"produktName": "Softdrinks","varianteName": "Sprite","kategorie": "beverage","einzelpreis": 280,"menge": 1}],"gesamtPreisCents": 3480,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.produkte-geliefert:v1', 'tisch:18', 68, now() - interval '4 hours 35 minutes', '{"lieferungId": "c0018000-0000-0000-0000-000000000023","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000054","menge": 4},{"positionId": "b0018000-0000-0000-0000-000000000055","menge": 4},{"positionId": "b0018000-0000-0000-0000-000000000056","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.zahlung-registriert:v1', 'tisch:18', 69, now() - interval '4 hours 31 minutes', '{"zahlungId": "d0018000-0000-0000-0000-000000000023","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000054","menge": 4},{"positionId": "b0018000-0000-0000-0000-000000000055","menge": 4},{"positionId": "b0018000-0000-0000-0000-000000000056","menge": 1}],"gesamtZahlungCents": 3480,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:18', 70, now() - interval '4 hours 9 minutes', '{"bestellungId": "a0018000-0000-0000-0000-000000000024","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000057","varianteId": 38,"produktName": "Wasser","varianteName": "Sprudel 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 2},{"positionId": "b0018000-0000-0000-0000-000000000058","varianteId": 43,"produktName": "Wein","varianteName": "Rotwein 0,2l","kategorie": "beverage","einzelpreis": 400,"menge": 3},{"positionId": "b0018000-0000-0000-0000-000000000059","varianteId": 37,"produktName": "Wasser","varianteName": "Medium 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 4},{"positionId": "b0018000-0000-0000-0000-000000000060","varianteId": 3,"produktName": "Bratwurst","varianteName": "Currywurst","kategorie": "food","einzelpreis": 450,"menge": 1}],"gesamtPreisCents": 2850,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:18', 71, now() - interval '4 hours 5 minutes', '{"lieferungId": "c0018000-0000-0000-0000-000000000024","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000057","menge": 2},{"positionId": "b0018000-0000-0000-0000-000000000058","menge": 3},{"positionId": "b0018000-0000-0000-0000-000000000059","menge": 4},{"positionId": "b0018000-0000-0000-0000-000000000060","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:18', 72, now() - interval '4 hours', '{"zahlungId": "d0018000-0000-0000-0000-000000000024","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000057","menge": 2},{"positionId": "b0018000-0000-0000-0000-000000000058","menge": 3},{"positionId": "b0018000-0000-0000-0000-000000000059","menge": 4},{"positionId": "b0018000-0000-0000-0000-000000000060","menge": 1}],"gesamtZahlungCents": 2850,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:18', 73, now() - interval '3 hours 34 minutes', '{"bestellungId": "a0018000-0000-0000-0000-000000000025","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000061","varianteId": 31,"produktName": "Softdrinks","varianteName": "Cola","kategorie": "beverage","einzelpreis": 280,"menge": 3},{"positionId": "b0018000-0000-0000-0000-000000000062","varianteId": 48,"produktName": "Hugo/Aperol","varianteName": "Hugo","kategorie": "beverage","einzelpreis": 550,"menge": 1}],"gesamtPreisCents": 1390,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.produkte-geliefert:v1', 'tisch:18', 74, now() - interval '3 hours 31 minutes', '{"lieferungId": "c0018000-0000-0000-0000-000000000025","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000061","menge": 3},{"positionId": "b0018000-0000-0000-0000-000000000062","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.zahlung-registriert:v1', 'tisch:18', 75, now() - interval '3 hours 29 minutes', '{"zahlungId": "d0018000-0000-0000-0000-000000000025","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000061","menge": 3},{"positionId": "b0018000-0000-0000-0000-000000000062","menge": 1}],"gesamtZahlungCents": 1390,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:18', 76, now() - interval '3 hours 2 minutes', '{"bestellungId": "a0018000-0000-0000-0000-000000000026","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000063","varianteId": 29,"produktName": "Weizen","varianteName": "Colaweizen Groß","kategorie": "beverage","einzelpreis": 400,"menge": 2},{"positionId": "b0018000-0000-0000-0000-000000000064","varianteId": 28,"produktName": "Weizen","varianteName": "Colaweizen Klein","kategorie": "beverage","einzelpreis": 300,"menge": 1},{"positionId": "b0018000-0000-0000-0000-000000000065","varianteId": 31,"produktName": "Softdrinks","varianteName": "Cola","kategorie": "beverage","einzelpreis": 280,"menge": 4}],"gesamtPreisCents": 2220,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.produkte-geliefert:v1', 'tisch:18', 77, now() - interval '2 hours 57 minutes', '{"lieferungId": "c0018000-0000-0000-0000-000000000026","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000063","menge": 2},{"positionId": "b0018000-0000-0000-0000-000000000064","menge": 1},{"positionId": "b0018000-0000-0000-0000-000000000065","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.zahlung-registriert:v1', 'tisch:18', 78, now() - interval '2 hours 53 minutes', '{"zahlungId": "d0018000-0000-0000-0000-000000000026","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000063","menge": 2},{"positionId": "b0018000-0000-0000-0000-000000000064","menge": 1},{"positionId": "b0018000-0000-0000-0000-000000000065","menge": 4}],"gesamtZahlungCents": 2220,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:18', 79, now() - interval '2 hours 30 minutes', '{"bestellungId": "a0018000-0000-0000-0000-000000000027","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000066","varianteId": 33,"produktName": "Softdrinks","varianteName": "Spezi","kategorie": "beverage","einzelpreis": 280,"menge": 4},{"positionId": "b0018000-0000-0000-0000-000000000067","varianteId": 28,"produktName": "Weizen","varianteName": "Colaweizen Klein","kategorie": "beverage","einzelpreis": 300,"menge": 4},{"positionId": "b0018000-0000-0000-0000-000000000068","varianteId": 30,"produktName": "Weizen","varianteName": "Russ","kategorie": "beverage","einzelpreis": 300,"menge": 4}],"gesamtPreisCents": 3520,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.produkte-geliefert:v1', 'tisch:18', 80, now() - interval '2 hours 27 minutes', '{"lieferungId": "c0018000-0000-0000-0000-000000000027","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000066","menge": 4},{"positionId": "b0018000-0000-0000-0000-000000000067","menge": 4},{"positionId": "b0018000-0000-0000-0000-000000000068","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.zahlung-registriert:v1', 'tisch:18', 81, now() - interval '2 hours 25 minutes', '{"zahlungId": "d0018000-0000-0000-0000-000000000027","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000066","menge": 4},{"positionId": "b0018000-0000-0000-0000-000000000067","menge": 4},{"positionId": "b0018000-0000-0000-0000-000000000068","menge": 4}],"gesamtZahlungCents": 3520,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:18', 82, now() - interval '2 hours 5 minutes', '{"bestellungId": "a0018000-0000-0000-0000-000000000028","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000069","varianteId": 29,"produktName": "Weizen","varianteName": "Colaweizen Groß","kategorie": "beverage","einzelpreis": 400,"menge": 1},{"positionId": "b0018000-0000-0000-0000-000000000070","varianteId": 49,"produktName": "Hugo/Aperol","varianteName": "Aperol Spritz","kategorie": "beverage","einzelpreis": 550,"menge": 1},{"positionId": "b0018000-0000-0000-0000-000000000071","varianteId": 41,"produktName": "Saftschorle","varianteName": "Rhabarberschorle 0,5l","kategorie": "beverage","einzelpreis": 350,"menge": 2},{"positionId": "b0018000-0000-0000-0000-000000000072","varianteId": 3,"produktName": "Bratwurst","varianteName": "Currywurst","kategorie": "food","einzelpreis": 450,"menge": 1}],"gesamtPreisCents": 2100,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:18', 83, now() - interval '2 hours 1 minutes', '{"lieferungId": "c0018000-0000-0000-0000-000000000028","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000069","menge": 1},{"positionId": "b0018000-0000-0000-0000-000000000070","menge": 1},{"positionId": "b0018000-0000-0000-0000-000000000071","menge": 2},{"positionId": "b0018000-0000-0000-0000-000000000072","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:18', 84, now() - interval '1 hours 59 minutes', '{"zahlungId": "d0018000-0000-0000-0000-000000000028","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000069","menge": 1},{"positionId": "b0018000-0000-0000-0000-000000000070","menge": 1},{"positionId": "b0018000-0000-0000-0000-000000000071","menge": 2},{"positionId": "b0018000-0000-0000-0000-000000000072","menge": 1}],"gesamtZahlungCents": 2100,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.bestellung-aufgegeben:v1', 'tisch:18', 85, now() - interval '1 hours 41 minutes', '{"bestellungId": "a0018000-0000-0000-0000-000000000029","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000073","varianteId": 26,"produktName": "Weizen","varianteName": "Klein 0,3l","kategorie": "beverage","einzelpreis": 300,"menge": 2},{"positionId": "b0018000-0000-0000-0000-000000000074","varianteId": 49,"produktName": "Hugo/Aperol","varianteName": "Aperol Spritz","kategorie": "beverage","einzelpreis": 550,"menge": 4},{"positionId": "b0018000-0000-0000-0000-000000000075","varianteId": 29,"produktName": "Weizen","varianteName": "Colaweizen Groß","kategorie": "beverage","einzelpreis": 400,"menge": 4}],"gesamtPreisCents": 4400,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.produkte-geliefert:v1', 'tisch:18', 86, now() - interval '1 hours 36 minutes', '{"lieferungId": "c0018000-0000-0000-0000-000000000029","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000073","menge": 2},{"positionId": "b0018000-0000-0000-0000-000000000074","menge": 4},{"positionId": "b0018000-0000-0000-0000-000000000075","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.zahlung-registriert:v1', 'tisch:18', 87, now() - interval '1 hours 32 minutes', '{"zahlungId": "d0018000-0000-0000-0000-000000000029","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000073","menge": 2},{"positionId": "b0018000-0000-0000-0000-000000000074","menge": 4},{"positionId": "b0018000-0000-0000-0000-000000000075","menge": 4}],"gesamtZahlungCents": 4400,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.bestellung-aufgegeben:v1', 'tisch:18', 88, now() - interval '1 hours 12 minutes', '{"bestellungId": "a0018000-0000-0000-0000-000000000030","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000076","varianteId": 48,"produktName": "Hugo/Aperol","varianteName": "Hugo","kategorie": "beverage","einzelpreis": 550,"menge": 4}],"gesamtPreisCents": 2200,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.produkte-geliefert:v1', 'tisch:18', 89, now() - interval '1 hours 6 minutes', '{"lieferungId": "c0018000-0000-0000-0000-000000000030","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000076","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.zahlung-registriert:v1', 'tisch:18', 90, now() - interval '1 hours 3 minutes', '{"zahlungId": "d0018000-0000-0000-0000-000000000030","positionen": [{"positionId": "b0018000-0000-0000-0000-000000000076","menge": 4}],"gesamtZahlungCents": 2200,"kommentar": ""}');

-- ─────────────────────────────────────────────────────────────────────────────
-- TISCH 19: Stehtisch Eingang — schnelle Bestellungen, alle 3 Tage (78 Events)
-- ─────────────────────────────────────────────────────────────────────────────

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.bestellung-aufgegeben:v1', 'tisch:19', 1, now() - interval '50 hours', '{"bestellungId": "a0019000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000001","varianteId": 46,"produktName": "Kaffee","varianteName": "Espresso","kategorie": "beverage","einzelpreis": 180,"menge": 3},{"positionId": "b0019000-0000-0000-0000-000000000002","varianteId": 26,"produktName": "Weizen","varianteName": "Klein 0,3l","kategorie": "beverage","einzelpreis": 300,"menge": 1}],"gesamtPreisCents": 840,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.produkte-geliefert:v1', 'tisch:19', 2, now() - interval '49 hours 56 minutes', '{"lieferungId": "c0019000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000001","menge": 3},{"positionId": "b0019000-0000-0000-0000-000000000002","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.zahlung-registriert:v1', 'tisch:19', 3, now() - interval '49 hours 52 minutes', '{"zahlungId": "d0019000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000001","menge": 3},{"positionId": "b0019000-0000-0000-0000-000000000002","menge": 1}],"gesamtZahlungCents": 840,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.bestellung-aufgegeben:v1', 'tisch:19', 4, now() - interval '49 hours 45 minutes', '{"bestellungId": "a0019000-0000-0000-0000-000000000002","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000003","varianteId": 15,"produktName": "Salat","varianteName": "Caesar Salat","kategorie": "food","einzelpreis": 650,"menge": 1},{"positionId": "b0019000-0000-0000-0000-000000000004","varianteId": 28,"produktName": "Weizen","varianteName": "Colaweizen Klein","kategorie": "beverage","einzelpreis": 300,"menge": 1}],"gesamtPreisCents": 950,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.produkte-geliefert:v1', 'tisch:19', 5, now() - interval '49 hours 32 minutes', '{"lieferungId": "c0019000-0000-0000-0000-000000000002","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000003","menge": 1},{"positionId": "b0019000-0000-0000-0000-000000000004","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.zahlung-registriert:v1', 'tisch:19', 6, now() - interval '49 hours 8 minutes', '{"zahlungId": "d0019000-0000-0000-0000-000000000002","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000003","menge": 1},{"positionId": "b0019000-0000-0000-0000-000000000004","menge": 1}],"gesamtZahlungCents": 950,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:19', 7, now() - interval '48 hours 36 minutes', '{"bestellungId": "a0019000-0000-0000-0000-000000000003","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000005","varianteId": 28,"produktName": "Weizen","varianteName": "Colaweizen Klein","kategorie": "beverage","einzelpreis": 300,"menge": 1}],"gesamtPreisCents": 300,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.produkte-geliefert:v1', 'tisch:19', 8, now() - interval '48 hours 33 minutes', '{"lieferungId": "c0019000-0000-0000-0000-000000000003","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000005","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.zahlung-registriert:v1', 'tisch:19', 9, now() - interval '48 hours 29 minutes', '{"zahlungId": "d0019000-0000-0000-0000-000000000003","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000005","menge": 1}],"gesamtZahlungCents": 300,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:19', 10, now() - interval '48 hours 26 minutes', '{"bestellungId": "a0019000-0000-0000-0000-000000000004","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000006","varianteId": 8,"produktName": "Flammkuchen","varianteName": "Mediterran","kategorie": "food","einzelpreis": 750,"menge": 1},{"positionId": "b0019000-0000-0000-0000-000000000007","varianteId": 38,"produktName": "Wasser","varianteName": "Sprudel 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 1}],"gesamtPreisCents": 950,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.produkte-geliefert:v1', 'tisch:19', 11, now() - interval '48 hours 13 minutes', '{"lieferungId": "c0019000-0000-0000-0000-000000000004","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000006","menge": 1},{"positionId": "b0019000-0000-0000-0000-000000000007","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.zahlung-registriert:v1', 'tisch:19', 12, now() - interval '47 hours 51 minutes', '{"zahlungId": "d0019000-0000-0000-0000-000000000004","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000006","menge": 1},{"positionId": "b0019000-0000-0000-0000-000000000007","menge": 1}],"gesamtZahlungCents": 950,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.bestellung-aufgegeben:v1', 'tisch:19', 13, now() - interval '47 hours 14 minutes', '{"bestellungId": "a0019000-0000-0000-0000-000000000005","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000008","varianteId": 31,"produktName": "Softdrinks","varianteName": "Cola","kategorie": "beverage","einzelpreis": 280,"menge": 1}],"gesamtPreisCents": 280,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.produkte-geliefert:v1', 'tisch:19', 14, now() - interval '47 hours 9 minutes', '{"lieferungId": "c0019000-0000-0000-0000-000000000005","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000008","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.zahlung-registriert:v1', 'tisch:19', 15, now() - interval '47 hours 4 minutes', '{"zahlungId": "d0019000-0000-0000-0000-000000000005","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000008","menge": 1}],"gesamtZahlungCents": 280,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.bestellung-aufgegeben:v1', 'tisch:19', 16, now() - interval '46 hours 57 minutes', '{"bestellungId": "a0019000-0000-0000-0000-000000000006","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000009","varianteId": 7,"produktName": "Flammkuchen","varianteName": "Speck & Zwiebel","kategorie": "food","einzelpreis": 700,"menge": 1},{"positionId": "b0019000-0000-0000-0000-000000000010","varianteId": 43,"produktName": "Wein","varianteName": "Rotwein 0,2l","kategorie": "beverage","einzelpreis": 400,"menge": 1}],"gesamtPreisCents": 1100,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.produkte-geliefert:v1', 'tisch:19', 17, now() - interval '46 hours 49 minutes', '{"lieferungId": "c0019000-0000-0000-0000-000000000006","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000009","menge": 1},{"positionId": "b0019000-0000-0000-0000-000000000010","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.zahlung-registriert:v1', 'tisch:19', 18, now() - interval '46 hours 28 minutes', '{"zahlungId": "d0019000-0000-0000-0000-000000000006","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000009","menge": 1},{"positionId": "b0019000-0000-0000-0000-000000000010","menge": 1}],"gesamtZahlungCents": 1100,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.bestellung-aufgegeben:v1', 'tisch:19', 19, now() - interval '45 hours 58 minutes', '{"bestellungId": "a0019000-0000-0000-0000-000000000007","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000011","varianteId": 38,"produktName": "Wasser","varianteName": "Sprudel 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 2}],"gesamtPreisCents": 400,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.produkte-geliefert:v1', 'tisch:19', 20, now() - interval '45 hours 52 minutes', '{"lieferungId": "c0019000-0000-0000-0000-000000000007","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000011","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.zahlung-registriert:v1', 'tisch:19', 21, now() - interval '45 hours 49 minutes', '{"zahlungId": "d0019000-0000-0000-0000-000000000007","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000011","menge": 2}],"gesamtZahlungCents": 400,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.bestellung-aufgegeben:v1', 'tisch:19', 22, now() - interval '45 hours 20 minutes', '{"bestellungId": "a0019000-0000-0000-0000-000000000008","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000012","varianteId": 37,"produktName": "Wasser","varianteName": "Medium 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 3}],"gesamtPreisCents": 600,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.produkte-geliefert:v1', 'tisch:19', 23, now() - interval '45 hours 14 minutes', '{"lieferungId": "c0019000-0000-0000-0000-000000000008","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000012","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.zahlung-registriert:v1', 'tisch:19', 24, now() - interval '45 hours 10 minutes', '{"zahlungId": "d0019000-0000-0000-0000-000000000008","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000012","menge": 3}],"gesamtZahlungCents": 600,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:19', 25, now() - interval '44 hours 42 minutes', '{"bestellungId": "a0019000-0000-0000-0000-000000000009","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000013","varianteId": 48,"produktName": "Hugo/Aperol","varianteName": "Hugo","kategorie": "beverage","einzelpreis": 550,"menge": 3},{"positionId": "b0019000-0000-0000-0000-000000000014","varianteId": 45,"produktName": "Kaffee","varianteName": "Tasse","kategorie": "beverage","einzelpreis": 200,"menge": 3}],"gesamtPreisCents": 2250,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:19', 26, now() - interval '44 hours 37 minutes', '{"lieferungId": "c0019000-0000-0000-0000-000000000009","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000013","menge": 3},{"positionId": "b0019000-0000-0000-0000-000000000014","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:19', 27, now() - interval '44 hours 34 minutes', '{"zahlungId": "d0019000-0000-0000-0000-000000000009","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000013","menge": 3},{"positionId": "b0019000-0000-0000-0000-000000000014","menge": 3}],"gesamtZahlungCents": 2250,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:19', 28, now() - interval '44 hours 26 minutes', '{"bestellungId": "a0019000-0000-0000-0000-000000000010","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000015","varianteId": 4,"produktName": "Pommes","varianteName": "Klein","kategorie": "food","einzelpreis": 250,"menge": 1},{"positionId": "b0019000-0000-0000-0000-000000000016","varianteId": 40,"produktName": "Saftschorle","varianteName": "Johannisbeerschorle 0,5l","kategorie": "beverage","einzelpreis": 300,"menge": 1}],"gesamtPreisCents": 550,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:19', 29, now() - interval '44 hours 13 minutes', '{"lieferungId": "c0019000-0000-0000-0000-000000000010","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000015","menge": 1},{"positionId": "b0019000-0000-0000-0000-000000000016","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:19', 30, now() - interval '43 hours 54 minutes', '{"zahlungId": "d0019000-0000-0000-0000-000000000010","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000015","menge": 1},{"positionId": "b0019000-0000-0000-0000-000000000016","menge": 1}],"gesamtZahlungCents": 550,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:19', 31, now() - interval '28 hours', '{"bestellungId": "a0019000-0000-0000-0000-000000000011","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000017","varianteId": 43,"produktName": "Wein","varianteName": "Rotwein 0,2l","kategorie": "beverage","einzelpreis": 400,"menge": 1},{"positionId": "b0019000-0000-0000-0000-000000000018","varianteId": 14,"produktName": "Salat","varianteName": "Gemischter Salat","kategorie": "food","einzelpreis": 550,"menge": 1}],"gesamtPreisCents": 950,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.produkte-geliefert:v1', 'tisch:19', 32, now() - interval '27 hours 57 minutes', '{"lieferungId": "c0019000-0000-0000-0000-000000000011","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000017","menge": 1},{"positionId": "b0019000-0000-0000-0000-000000000018","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.zahlung-registriert:v1', 'tisch:19', 33, now() - interval '27 hours 52 minutes', '{"zahlungId": "d0019000-0000-0000-0000-000000000011","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000017","menge": 1},{"positionId": "b0019000-0000-0000-0000-000000000018","menge": 1}],"gesamtZahlungCents": 950,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.bestellung-aufgegeben:v1', 'tisch:19', 34, now() - interval '27 hours 22 minutes', '{"bestellungId": "a0019000-0000-0000-0000-000000000012","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000019","varianteId": 25,"produktName": "Bier","varianteName": "Maß 1,0l","kategorie": "beverage","einzelpreis": 850,"menge": 1}],"gesamtPreisCents": 850,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.produkte-geliefert:v1', 'tisch:19', 35, now() - interval '27 hours 17 minutes', '{"lieferungId": "c0019000-0000-0000-0000-000000000012","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000019","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.zahlung-registriert:v1', 'tisch:19', 36, now() - interval '27 hours 14 minutes', '{"zahlungId": "d0019000-0000-0000-0000-000000000012","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000019","menge": 1}],"gesamtZahlungCents": 850,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:19', 37, now() - interval '26 hours 54 minutes', '{"bestellungId": "a0019000-0000-0000-0000-000000000013","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000020","varianteId": 34,"produktName": "Softdrinks","varianteName": "Sprite","kategorie": "beverage","einzelpreis": 280,"menge": 2},{"positionId": "b0019000-0000-0000-0000-000000000021","varianteId": 2,"produktName": "Bratwurst","varianteName": "XXL","kategorie": "food","einzelpreis": 500,"menge": 1}],"gesamtPreisCents": 1060,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.produkte-geliefert:v1', 'tisch:19', 38, now() - interval '26 hours 49 minutes', '{"lieferungId": "c0019000-0000-0000-0000-000000000013","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000020","menge": 2},{"positionId": "b0019000-0000-0000-0000-000000000021","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.zahlung-registriert:v1', 'tisch:19', 39, now() - interval '26 hours 47 minutes', '{"zahlungId": "d0019000-0000-0000-0000-000000000013","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000020","menge": 2},{"positionId": "b0019000-0000-0000-0000-000000000021","menge": 1}],"gesamtZahlungCents": 1060,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.bestellung-aufgegeben:v1', 'tisch:19', 40, now() - interval '26 hours 25 minutes', '{"bestellungId": "a0019000-0000-0000-0000-000000000014","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000022","varianteId": 29,"produktName": "Weizen","varianteName": "Colaweizen Groß","kategorie": "beverage","einzelpreis": 400,"menge": 3}],"gesamtPreisCents": 1200,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.produkte-geliefert:v1', 'tisch:19', 41, now() - interval '26 hours 21 minutes', '{"lieferungId": "c0019000-0000-0000-0000-000000000014","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000022","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.zahlung-registriert:v1', 'tisch:19', 42, now() - interval '26 hours 16 minutes', '{"zahlungId": "d0019000-0000-0000-0000-000000000014","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000022","menge": 3}],"gesamtZahlungCents": 1200,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:19', 43, now() - interval '25 hours 57 minutes', '{"bestellungId": "a0019000-0000-0000-0000-000000000015","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000023","varianteId": 25,"produktName": "Bier","varianteName": "Maß 1,0l","kategorie": "beverage","einzelpreis": 850,"menge": 1}],"gesamtPreisCents": 850,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.produkte-geliefert:v1', 'tisch:19', 44, now() - interval '25 hours 51 minutes', '{"lieferungId": "c0019000-0000-0000-0000-000000000015","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000023","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.zahlung-registriert:v1', 'tisch:19', 45, now() - interval '25 hours 49 minutes', '{"zahlungId": "d0019000-0000-0000-0000-000000000015","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000023","menge": 1}],"gesamtZahlungCents": 850,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.bestellung-aufgegeben:v1', 'tisch:19', 46, now() - interval '25 hours 18 minutes', '{"bestellungId": "a0019000-0000-0000-0000-000000000016","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000024","varianteId": 32,"produktName": "Softdrinks","varianteName": "Fanta","kategorie": "beverage","einzelpreis": 280,"menge": 2},{"positionId": "b0019000-0000-0000-0000-000000000025","varianteId": 2,"produktName": "Bratwurst","varianteName": "XXL","kategorie": "food","einzelpreis": 500,"menge": 1}],"gesamtPreisCents": 1060,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.produkte-geliefert:v1', 'tisch:19', 47, now() - interval '25 hours 15 minutes', '{"lieferungId": "c0019000-0000-0000-0000-000000000016","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000024","menge": 2},{"positionId": "b0019000-0000-0000-0000-000000000025","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.zahlung-registriert:v1', 'tisch:19', 48, now() - interval '25 hours 12 minutes', '{"zahlungId": "d0019000-0000-0000-0000-000000000016","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000024","menge": 2},{"positionId": "b0019000-0000-0000-0000-000000000025","menge": 1}],"gesamtZahlungCents": 1060,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.bestellung-aufgegeben:v1', 'tisch:19', 49, now() - interval '24 hours 37 minutes', '{"bestellungId": "a0019000-0000-0000-0000-000000000017","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000026","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 1}],"gesamtPreisCents": 450,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.produkte-geliefert:v1', 'tisch:19', 50, now() - interval '24 hours 34 minutes', '{"lieferungId": "c0019000-0000-0000-0000-000000000017","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000026","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.zahlung-registriert:v1', 'tisch:19', 51, now() - interval '24 hours 32 minutes', '{"zahlungId": "d0019000-0000-0000-0000-000000000017","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000026","menge": 1}],"gesamtZahlungCents": 450,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.bestellung-aufgegeben:v1', 'tisch:19', 52, now() - interval '24 hours 1 minutes', '{"bestellungId": "a0019000-0000-0000-0000-000000000018","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000027","varianteId": 47,"produktName": "Tee","varianteName": "Verschiedene Sorten","kategorie": "beverage","einzelpreis": 200,"menge": 3}],"gesamtPreisCents": 600,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.produkte-geliefert:v1', 'tisch:19', 53, now() - interval '23 hours 58 minutes', '{"lieferungId": "c0019000-0000-0000-0000-000000000018","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000027","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.zahlung-registriert:v1', 'tisch:19', 54, now() - interval '23 hours 53 minutes', '{"zahlungId": "d0019000-0000-0000-0000-000000000018","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000027","menge": 3}],"gesamtZahlungCents": 600,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:19', 55, now() - interval '23 hours 30 minutes', '{"bestellungId": "a0019000-0000-0000-0000-000000000019","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000028","varianteId": 38,"produktName": "Wasser","varianteName": "Sprudel 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 2}],"gesamtPreisCents": 400,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.produkte-geliefert:v1', 'tisch:19', 56, now() - interval '23 hours 24 minutes', '{"lieferungId": "c0019000-0000-0000-0000-000000000019","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000028","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.zahlung-registriert:v1', 'tisch:19', 57, now() - interval '23 hours 21 minutes', '{"zahlungId": "d0019000-0000-0000-0000-000000000019","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000028","menge": 2}],"gesamtZahlungCents": 400,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:19', 58, now() - interval '22 hours 50 minutes', '{"bestellungId": "a0019000-0000-0000-0000-000000000020","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000029","varianteId": 38,"produktName": "Wasser","varianteName": "Sprudel 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 3},{"positionId": "b0019000-0000-0000-0000-000000000030","varianteId": 21,"produktName": "Brezel","varianteName": "mit Butter","kategorie": "food","einzelpreis": 300,"menge": 1}],"gesamtPreisCents": 900,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:19', 59, now() - interval '22 hours 46 minutes', '{"lieferungId": "c0019000-0000-0000-0000-000000000020","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000029","menge": 3},{"positionId": "b0019000-0000-0000-0000-000000000030","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:19', 60, now() - interval '22 hours 43 minutes', '{"zahlungId": "d0019000-0000-0000-0000-000000000020","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000029","menge": 3},{"positionId": "b0019000-0000-0000-0000-000000000030","menge": 1}],"gesamtZahlungCents": 900,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.bestellung-aufgegeben:v1', 'tisch:19', 61, now() - interval '4 hours 35 minutes', '{"bestellungId": "a0019000-0000-0000-0000-000000000021","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000031","varianteId": 32,"produktName": "Softdrinks","varianteName": "Fanta","kategorie": "beverage","einzelpreis": 280,"menge": 2},{"positionId": "b0019000-0000-0000-0000-000000000032","varianteId": 17,"produktName": "Waffeln","varianteName": "mit Puderzucker","kategorie": "food","einzelpreis": 300,"menge": 1}],"gesamtPreisCents": 860,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.produkte-geliefert:v1', 'tisch:19', 62, now() - interval '4 hours 30 minutes', '{"lieferungId": "c0019000-0000-0000-0000-000000000021","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000031","menge": 2},{"positionId": "b0019000-0000-0000-0000-000000000032","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.zahlung-registriert:v1', 'tisch:19', 63, now() - interval '4 hours 28 minutes', '{"zahlungId": "d0019000-0000-0000-0000-000000000021","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000031","menge": 2},{"positionId": "b0019000-0000-0000-0000-000000000032","menge": 1}],"gesamtZahlungCents": 860,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.bestellung-aufgegeben:v1', 'tisch:19', 64, now() - interval '3 hours 56 minutes', '{"bestellungId": "a0019000-0000-0000-0000-000000000022","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000033","varianteId": 49,"produktName": "Hugo/Aperol","varianteName": "Aperol Spritz","kategorie": "beverage","einzelpreis": 550,"menge": 1},{"positionId": "b0019000-0000-0000-0000-000000000034","varianteId": 1,"produktName": "Bratwurst","varianteName": "Normal","kategorie": "food","einzelpreis": 350,"menge": 1}],"gesamtPreisCents": 900,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.produkte-geliefert:v1', 'tisch:19', 65, now() - interval '3 hours 51 minutes', '{"lieferungId": "c0019000-0000-0000-0000-000000000022","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000033","menge": 1},{"positionId": "b0019000-0000-0000-0000-000000000034","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.zahlung-registriert:v1', 'tisch:19', 66, now() - interval '3 hours 47 minutes', '{"zahlungId": "d0019000-0000-0000-0000-000000000022","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000033","menge": 1},{"positionId": "b0019000-0000-0000-0000-000000000034","menge": 1}],"gesamtZahlungCents": 900,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:19', 67, now() - interval '3 hours 22 minutes', '{"bestellungId": "a0019000-0000-0000-0000-000000000023","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000035","varianteId": 37,"produktName": "Wasser","varianteName": "Medium 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 1}],"gesamtPreisCents": 200,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:19', 68, now() - interval '3 hours 18 minutes', '{"lieferungId": "c0019000-0000-0000-0000-000000000023","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000035","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:19', 69, now() - interval '3 hours 13 minutes', '{"zahlungId": "d0019000-0000-0000-0000-000000000023","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000035","menge": 1}],"gesamtZahlungCents": 200,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:19', 70, now() - interval '2 hours 52 minutes', '{"bestellungId": "a0019000-0000-0000-0000-000000000024","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000036","varianteId": 37,"produktName": "Wasser","varianteName": "Medium 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 1},{"positionId": "b0019000-0000-0000-0000-000000000037","varianteId": 8,"produktName": "Flammkuchen","varianteName": "Mediterran","kategorie": "food","einzelpreis": 750,"menge": 1}],"gesamtPreisCents": 950,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.produkte-geliefert:v1', 'tisch:19', 71, now() - interval '2 hours 46 minutes', '{"lieferungId": "c0019000-0000-0000-0000-000000000024","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000036","menge": 1},{"positionId": "b0019000-0000-0000-0000-000000000037","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.zahlung-registriert:v1', 'tisch:19', 72, now() - interval '2 hours 44 minutes', '{"zahlungId": "d0019000-0000-0000-0000-000000000024","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000036","menge": 1},{"positionId": "b0019000-0000-0000-0000-000000000037","menge": 1}],"gesamtZahlungCents": 950,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:19', 73, now() - interval '2 hours 28 minutes', '{"bestellungId": "a0019000-0000-0000-0000-000000000025","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000038","varianteId": 42,"produktName": "Wein","varianteName": "Weißwein 0,2l","kategorie": "beverage","einzelpreis": 400,"menge": 1},{"positionId": "b0019000-0000-0000-0000-000000000039","varianteId": 16,"produktName": "Kuchen","varianteName": "Stück","kategorie": "food","einzelpreis": 250,"menge": 1}],"gesamtPreisCents": 650,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.produkte-geliefert:v1', 'tisch:19', 74, now() - interval '2 hours 24 minutes', '{"lieferungId": "c0019000-0000-0000-0000-000000000025","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000038","menge": 1},{"positionId": "b0019000-0000-0000-0000-000000000039","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.zahlung-registriert:v1', 'tisch:19', 75, now() - interval '2 hours 21 minutes', '{"zahlungId": "d0019000-0000-0000-0000-000000000025","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000038","menge": 1},{"positionId": "b0019000-0000-0000-0000-000000000039","menge": 1}],"gesamtZahlungCents": 650,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.bestellung-aufgegeben:v1', 'tisch:19', 76, now() - interval '1 hours 53 minutes', '{"bestellungId": "a0019000-0000-0000-0000-000000000026","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000040","varianteId": 38,"produktName": "Wasser","varianteName": "Sprudel 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 2}],"gesamtPreisCents": 400,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.produkte-geliefert:v1', 'tisch:19', 77, now() - interval '1 hours 47 minutes', '{"lieferungId": "c0019000-0000-0000-0000-000000000026","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000040","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(7, 'Sophie Becker', 'tisch.zahlung-registriert:v1', 'tisch:19', 78, now() - interval '1 hours 45 minutes', '{"zahlungId": "d0019000-0000-0000-0000-000000000026","positionen": [{"positionId": "b0019000-0000-0000-0000-000000000040","menge": 2}],"gesamtZahlungCents": 400,"kommentar": ""}');

-- ─────────────────────────────────────────────────────────────────────────────
-- TISCH 20: Stehtisch Terrasse — Tag 2+3, gemischt (50 Events)
-- ─────────────────────────────────────────────────────────────────────────────

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.bestellung-aufgegeben:v1', 'tisch:20', 1, now() - interval '25 hours', '{"bestellungId": "a0020000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0020000-0000-0000-0000-000000000001","varianteId": 41,"produktName": "Saftschorle","varianteName": "Rhabarberschorle 0,5l","kategorie": "beverage","einzelpreis": 350,"menge": 1},{"positionId": "b0020000-0000-0000-0000-000000000002","varianteId": 38,"produktName": "Wasser","varianteName": "Sprudel 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 4}],"gesamtPreisCents": 1150,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.produkte-geliefert:v1', 'tisch:20', 2, now() - interval '24 hours 56 minutes', '{"lieferungId": "c0020000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0020000-0000-0000-0000-000000000001","menge": 1},{"positionId": "b0020000-0000-0000-0000-000000000002","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.zahlung-registriert:v1', 'tisch:20', 3, now() - interval '24 hours 54 minutes', '{"zahlungId": "d0020000-0000-0000-0000-000000000001","positionen": [{"positionId": "b0020000-0000-0000-0000-000000000001","menge": 1},{"positionId": "b0020000-0000-0000-0000-000000000002","menge": 4}],"gesamtZahlungCents": 1150,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.bestellung-aufgegeben:v1', 'tisch:20', 4, now() - interval '24 hours 19 minutes', '{"bestellungId": "a0020000-0000-0000-0000-000000000002","positionen": [{"positionId": "b0020000-0000-0000-0000-000000000003","varianteId": 33,"produktName": "Softdrinks","varianteName": "Spezi","kategorie": "beverage","einzelpreis": 280,"menge": 3},{"positionId": "b0020000-0000-0000-0000-000000000004","varianteId": 12,"produktName": "Grillplatte","varianteName": "Klein","kategorie": "food","einzelpreis": 800,"menge": 2}],"gesamtPreisCents": 2440,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.produkte-geliefert:v1', 'tisch:20', 5, now() - interval '24 hours 16 minutes', '{"lieferungId": "c0020000-0000-0000-0000-000000000002","positionen": [{"positionId": "b0020000-0000-0000-0000-000000000003","menge": 3},{"positionId": "b0020000-0000-0000-0000-000000000004","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.zahlung-registriert:v1', 'tisch:20', 6, now() - interval '24 hours 12 minutes', '{"zahlungId": "d0020000-0000-0000-0000-000000000002","positionen": [{"positionId": "b0020000-0000-0000-0000-000000000003","menge": 3},{"positionId": "b0020000-0000-0000-0000-000000000004","menge": 2}],"gesamtZahlungCents": 2440,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.bestellung-aufgegeben:v1', 'tisch:20', 7, now() - interval '23 hours 49 minutes', '{"bestellungId": "a0020000-0000-0000-0000-000000000003","positionen": [{"positionId": "b0020000-0000-0000-0000-000000000005","varianteId": 46,"produktName": "Kaffee","varianteName": "Espresso","kategorie": "beverage","einzelpreis": 180,"menge": 4},{"positionId": "b0020000-0000-0000-0000-000000000006","varianteId": 38,"produktName": "Wasser","varianteName": "Sprudel 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 1}],"gesamtPreisCents": 920,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.produkte-geliefert:v1', 'tisch:20', 8, now() - interval '23 hours 43 minutes', '{"lieferungId": "c0020000-0000-0000-0000-000000000003","positionen": [{"positionId": "b0020000-0000-0000-0000-000000000005","menge": 4},{"positionId": "b0020000-0000-0000-0000-000000000006","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.zahlung-registriert:v1', 'tisch:20', 9, now() - interval '23 hours 39 minutes', '{"zahlungId": "d0020000-0000-0000-0000-000000000003","positionen": [{"positionId": "b0020000-0000-0000-0000-000000000005","menge": 4},{"positionId": "b0020000-0000-0000-0000-000000000006","menge": 1}],"gesamtZahlungCents": 920,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.bestellung-aufgegeben:v1', 'tisch:20', 10, now() - interval '23 hours 17 minutes', '{"bestellungId": "a0020000-0000-0000-0000-000000000004","positionen": [{"positionId": "b0020000-0000-0000-0000-000000000007","varianteId": 37,"produktName": "Wasser","varianteName": "Medium 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 3},{"positionId": "b0020000-0000-0000-0000-000000000008","varianteId": 7,"produktName": "Flammkuchen","varianteName": "Speck & Zwiebel","kategorie": "food","einzelpreis": 700,"menge": 1}],"gesamtPreisCents": 1300,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.produkte-geliefert:v1', 'tisch:20', 11, now() - interval '23 hours 14 minutes', '{"lieferungId": "c0020000-0000-0000-0000-000000000004","positionen": [{"positionId": "b0020000-0000-0000-0000-000000000007","menge": 3},{"positionId": "b0020000-0000-0000-0000-000000000008","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.zahlung-registriert:v1', 'tisch:20', 12, now() - interval '23 hours 12 minutes', '{"zahlungId": "d0020000-0000-0000-0000-000000000004","positionen": [{"positionId": "b0020000-0000-0000-0000-000000000007","menge": 3},{"positionId": "b0020000-0000-0000-0000-000000000008","menge": 1}],"gesamtZahlungCents": 1300,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.bestellung-aufgegeben:v1', 'tisch:20', 13, now() - interval '22 hours 53 minutes', '{"bestellungId": "a0020000-0000-0000-0000-000000000005","positionen": [{"positionId": "b0020000-0000-0000-0000-000000000009","varianteId": 44,"produktName": "Wein","varianteName": "Rosé 0,2l","kategorie": "beverage","einzelpreis": 400,"menge": 3},{"positionId": "b0020000-0000-0000-0000-000000000010","varianteId": 39,"produktName": "Saftschorle","varianteName": "Apfelschorle 0,5l","kategorie": "beverage","einzelpreis": 300,"menge": 2},{"positionId": "b0020000-0000-0000-0000-000000000011","varianteId": 26,"produktName": "Weizen","varianteName": "Klein 0,3l","kategorie": "beverage","einzelpreis": 300,"menge": 1},{"positionId": "b0020000-0000-0000-0000-000000000012","varianteId": 3,"produktName": "Bratwurst","varianteName": "Currywurst","kategorie": "food","einzelpreis": 450,"menge": 1}],"gesamtPreisCents": 2550,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.produkte-geliefert:v1', 'tisch:20', 14, now() - interval '22 hours 50 minutes', '{"lieferungId": "c0020000-0000-0000-0000-000000000005","positionen": [{"positionId": "b0020000-0000-0000-0000-000000000009","menge": 3},{"positionId": "b0020000-0000-0000-0000-000000000010","menge": 2},{"positionId": "b0020000-0000-0000-0000-000000000011","menge": 1},{"positionId": "b0020000-0000-0000-0000-000000000012","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.zahlung-registriert:v1', 'tisch:20', 15, now() - interval '22 hours 46 minutes', '{"zahlungId": "d0020000-0000-0000-0000-000000000005","positionen": [{"positionId": "b0020000-0000-0000-0000-000000000009","menge": 3},{"positionId": "b0020000-0000-0000-0000-000000000010","menge": 2},{"positionId": "b0020000-0000-0000-0000-000000000011","menge": 1},{"positionId": "b0020000-0000-0000-0000-000000000012","menge": 1}],"gesamtZahlungCents": 2550,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:20', 16, now() - interval '22 hours 27 minutes', '{"bestellungId": "a0020000-0000-0000-0000-000000000006","positionen": [{"positionId": "b0020000-0000-0000-0000-000000000013","varianteId": 31,"produktName": "Softdrinks","varianteName": "Cola","kategorie": "beverage","einzelpreis": 280,"menge": 3},{"positionId": "b0020000-0000-0000-0000-000000000014","varianteId": 23,"produktName": "Bier","varianteName": "0,3l","kategorie": "beverage","einzelpreis": 300,"menge": 4},{"positionId": "b0020000-0000-0000-0000-000000000015","varianteId": 38,"produktName": "Wasser","varianteName": "Sprudel 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 1},{"positionId": "b0020000-0000-0000-0000-000000000016","varianteId": 20,"produktName": "Brezel","varianteName": "Normal","kategorie": "food","einzelpreis": 200,"menge": 2}],"gesamtPreisCents": 2640,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:20', 17, now() - interval '22 hours 21 minutes', '{"lieferungId": "c0020000-0000-0000-0000-000000000006","positionen": [{"positionId": "b0020000-0000-0000-0000-000000000013","menge": 3},{"positionId": "b0020000-0000-0000-0000-000000000014","menge": 4},{"positionId": "b0020000-0000-0000-0000-000000000015","menge": 1},{"positionId": "b0020000-0000-0000-0000-000000000016","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:20', 18, now() - interval '22 hours 17 minutes', '{"zahlungId": "d0020000-0000-0000-0000-000000000006","positionen": [{"positionId": "b0020000-0000-0000-0000-000000000013","menge": 3},{"positionId": "b0020000-0000-0000-0000-000000000014","menge": 4},{"positionId": "b0020000-0000-0000-0000-000000000015","menge": 1},{"positionId": "b0020000-0000-0000-0000-000000000016","menge": 2}],"gesamtZahlungCents": 2640,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:20', 19, now() - interval '21 hours 43 minutes', '{"bestellungId": "a0020000-0000-0000-0000-000000000007","positionen": [{"positionId": "b0020000-0000-0000-0000-000000000017","varianteId": 36,"produktName": "Wasser","varianteName": "Still 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 1},{"positionId": "b0020000-0000-0000-0000-000000000018","varianteId": 37,"produktName": "Wasser","varianteName": "Medium 0,5l","kategorie": "beverage","einzelpreis": 200,"menge": 1},{"positionId": "b0020000-0000-0000-0000-000000000019","varianteId": 2,"produktName": "Bratwurst","varianteName": "XXL","kategorie": "food","einzelpreis": 500,"menge": 2}],"gesamtPreisCents": 1400,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:20', 20, now() - interval '21 hours 40 minutes', '{"lieferungId": "c0020000-0000-0000-0000-000000000007","positionen": [{"positionId": "b0020000-0000-0000-0000-000000000017","menge": 1},{"positionId": "b0020000-0000-0000-0000-000000000018","menge": 1},{"positionId": "b0020000-0000-0000-0000-000000000019","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:20', 21, now() - interval '21 hours 35 minutes', '{"zahlungId": "d0020000-0000-0000-0000-000000000007","positionen": [{"positionId": "b0020000-0000-0000-0000-000000000017","menge": 1},{"positionId": "b0020000-0000-0000-0000-000000000018","menge": 1},{"positionId": "b0020000-0000-0000-0000-000000000019","menge": 2}],"gesamtZahlungCents": 1400,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:20', 22, now() - interval '21 hours 12 minutes', '{"bestellungId": "a0020000-0000-0000-0000-000000000008","positionen": [{"positionId": "b0020000-0000-0000-0000-000000000020","varianteId": 44,"produktName": "Wein","varianteName": "Rosé 0,2l","kategorie": "beverage","einzelpreis": 400,"menge": 2},{"positionId": "b0020000-0000-0000-0000-000000000021","varianteId": 28,"produktName": "Weizen","varianteName": "Colaweizen Klein","kategorie": "beverage","einzelpreis": 300,"menge": 2},{"positionId": "b0020000-0000-0000-0000-000000000022","varianteId": 25,"produktName": "Bier","varianteName": "Maß 1,0l","kategorie": "beverage","einzelpreis": 850,"menge": 3}],"gesamtPreisCents": 3950,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:20', 23, now() - interval '21 hours 6 minutes', '{"lieferungId": "c0020000-0000-0000-0000-000000000008","positionen": [{"positionId": "b0020000-0000-0000-0000-000000000020","menge": 2},{"positionId": "b0020000-0000-0000-0000-000000000021","menge": 2},{"positionId": "b0020000-0000-0000-0000-000000000022","menge": 3}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:20', 24, now() - interval '21 hours 4 minutes', '{"zahlungId": "d0020000-0000-0000-0000-000000000008","positionen": [{"positionId": "b0020000-0000-0000-0000-000000000020","menge": 2},{"positionId": "b0020000-0000-0000-0000-000000000021","menge": 2},{"positionId": "b0020000-0000-0000-0000-000000000022","menge": 3}],"gesamtZahlungCents": 3950,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.bestellung-aufgegeben:v1', 'tisch:20', 25, now() - interval '19 hours 10 minutes', '{"bestellungId": "a0020000-0000-0000-0000-000000000009","positionen": [{"positionId": "b0020000-0000-0000-0000-000000000023","varianteId": 7,"produktName": "Flammkuchen","varianteName": "Speck & Zwiebel","kategorie": "food","einzelpreis": 700,"menge": 2},{"positionId": "b0020000-0000-0000-0000-000000000024","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 4},{"positionId": "b0020000-0000-0000-0000-000000000025","varianteId": 48,"produktName": "Hugo/Aperol","varianteName": "Hugo","kategorie": "beverage","einzelpreis": 550,"menge": 2}],"gesamtPreisCents": 4300,"kommentar": "Terrasse Abend"}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.produkte-geliefert:v1', 'tisch:20', 26, now() - interval '18 hours 55 minutes', '{"lieferungId": "c0020000-0000-0000-0000-000000000009","positionen": [{"positionId": "b0020000-0000-0000-0000-000000000023","menge": 2},{"positionId": "b0020000-0000-0000-0000-000000000024","menge": 4},{"positionId": "b0020000-0000-0000-0000-000000000025","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.zahlung-registriert:v1', 'tisch:20', 27, now() - interval '18 hours 31 minutes', '{"zahlungId": "d0020000-0000-0000-0000-000000000009","positionen": [{"positionId": "b0020000-0000-0000-0000-000000000023","menge": 2},{"positionId": "b0020000-0000-0000-0000-000000000024","menge": 4},{"positionId": "b0020000-0000-0000-0000-000000000025","menge": 2}],"gesamtZahlungCents": 4300,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:20', 28, now() - interval '18 hours 16 minutes', '{"bestellungId": "a0020000-0000-0000-0000-000000000010","positionen": [{"positionId": "b0020000-0000-0000-0000-000000000026","varianteId": 1,"produktName": "Bratwurst","varianteName": "Normal","kategorie": "food","einzelpreis": 350,"menge": 3},{"positionId": "b0020000-0000-0000-0000-000000000027","varianteId": 23,"produktName": "Bier","varianteName": "0,3l","kategorie": "beverage","einzelpreis": 300,"menge": 4}],"gesamtPreisCents": 2250,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.produkte-geliefert:v1', 'tisch:20', 29, now() - interval '18 hours 2 minutes', '{"lieferungId": "c0020000-0000-0000-0000-000000000010","positionen": [{"positionId": "b0020000-0000-0000-0000-000000000026","menge": 3},{"positionId": "b0020000-0000-0000-0000-000000000027","menge": 4}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.zahlung-registriert:v1', 'tisch:20', 30, now() - interval '17 hours 50 minutes', '{"zahlungId": "d0020000-0000-0000-0000-000000000010","positionen": [{"positionId": "b0020000-0000-0000-0000-000000000026","menge": 3},{"positionId": "b0020000-0000-0000-0000-000000000027","menge": 4}],"gesamtZahlungCents": 2250,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.bestellung-aufgegeben:v1', 'tisch:20', 31, now() - interval '17 hours 30 minutes', '{"bestellungId": "a0020000-0000-0000-0000-000000000011","positionen": [{"positionId": "b0020000-0000-0000-0000-000000000028","varianteId": 25,"produktName": "Bier","varianteName": "Maß 1,0l","kategorie": "beverage","einzelpreis": 850,"menge": 2},{"positionId": "b0020000-0000-0000-0000-000000000029","varianteId": 49,"produktName": "Hugo/Aperol","varianteName": "Aperol Spritz","kategorie": "beverage","einzelpreis": 550,"menge": 2}],"gesamtPreisCents": 2800,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.produkte-geliefert:v1', 'tisch:20', 32, now() - interval '17 hours 17 minutes', '{"lieferungId": "c0020000-0000-0000-0000-000000000011","positionen": [{"positionId": "b0020000-0000-0000-0000-000000000028","menge": 2},{"positionId": "b0020000-0000-0000-0000-000000000029","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.zahlung-registriert:v1', 'tisch:20', 33, now() - interval '16 hours 59 minutes', '{"zahlungId": "d0020000-0000-0000-0000-000000000011","positionen": [{"positionId": "b0020000-0000-0000-0000-000000000028","menge": 2},{"positionId": "b0020000-0000-0000-0000-000000000029","menge": 2}],"gesamtZahlungCents": 2800,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.bestellung-aufgegeben:v1', 'tisch:20', 34, now() - interval '4 hours 10 minutes', '{"bestellungId": "a0020000-0000-0000-0000-000000000012","positionen": [{"positionId": "b0020000-0000-0000-0000-000000000030","varianteId": 45,"produktName": "Kaffee","varianteName": "Tasse","kategorie": "beverage","einzelpreis": 200,"menge": 1},{"positionId": "b0020000-0000-0000-0000-000000000031","varianteId": 16,"produktName": "Kuchen","varianteName": "Stück","kategorie": "food","einzelpreis": 250,"menge": 1}],"gesamtPreisCents": 450,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.produkte-geliefert:v1', 'tisch:20', 35, now() - interval '4 hours 6 minutes', '{"lieferungId": "c0020000-0000-0000-0000-000000000012","positionen": [{"positionId": "b0020000-0000-0000-0000-000000000030","menge": 1},{"positionId": "b0020000-0000-0000-0000-000000000031","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.zahlung-registriert:v1', 'tisch:20', 36, now() - interval '4 hours 4 minutes', '{"zahlungId": "d0020000-0000-0000-0000-000000000012","positionen": [{"positionId": "b0020000-0000-0000-0000-000000000030","menge": 1},{"positionId": "b0020000-0000-0000-0000-000000000031","menge": 1}],"gesamtZahlungCents": 450,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:20', 37, now() - interval '3 hours 36 minutes', '{"bestellungId": "a0020000-0000-0000-0000-000000000013","positionen": [{"positionId": "b0020000-0000-0000-0000-000000000032","varianteId": 27,"produktName": "Weizen","varianteName": "Groß 0,5l","kategorie": "beverage","einzelpreis": 400,"menge": 2},{"positionId": "b0020000-0000-0000-0000-000000000033","varianteId": 14,"produktName": "Salat","varianteName": "Gemischter Salat","kategorie": "food","einzelpreis": 550,"menge": 1}],"gesamtPreisCents": 1350,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.produkte-geliefert:v1', 'tisch:20', 38, now() - interval '3 hours 30 minutes', '{"lieferungId": "c0020000-0000-0000-0000-000000000013","positionen": [{"positionId": "b0020000-0000-0000-0000-000000000032","menge": 2},{"positionId": "b0020000-0000-0000-0000-000000000033","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(8, 'Markus Lehmann', 'tisch.zahlung-registriert:v1', 'tisch:20', 39, now() - interval '3 hours 25 minutes', '{"zahlungId": "d0020000-0000-0000-0000-000000000013","positionen": [{"positionId": "b0020000-0000-0000-0000-000000000032","menge": 2},{"positionId": "b0020000-0000-0000-0000-000000000033","menge": 1}],"gesamtZahlungCents": 1350,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.bestellung-aufgegeben:v1', 'tisch:20', 40, now() - interval '3 hours', '{"bestellungId": "a0020000-0000-0000-0000-000000000014","positionen": [{"positionId": "b0020000-0000-0000-0000-000000000034","varianteId": 45,"produktName": "Kaffee","varianteName": "Tasse","kategorie": "beverage","einzelpreis": 200,"menge": 2}],"gesamtPreisCents": 400,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.produkte-geliefert:v1', 'tisch:20', 41, now() - interval '2 hours 56 minutes', '{"lieferungId": "c0020000-0000-0000-0000-000000000014","positionen": [{"positionId": "b0020000-0000-0000-0000-000000000034","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(9, 'Anna Krause', 'tisch.zahlung-registriert:v1', 'tisch:20', 42, now() - interval '2 hours 54 minutes', '{"zahlungId": "d0020000-0000-0000-0000-000000000014","positionen": [{"positionId": "b0020000-0000-0000-0000-000000000034","menge": 2}],"gesamtZahlungCents": 400,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.bestellung-aufgegeben:v1', 'tisch:20', 43, now() - interval '2 hours 31 minutes', '{"bestellungId": "a0020000-0000-0000-0000-000000000015","positionen": [{"positionId": "b0020000-0000-0000-0000-000000000035","varianteId": 32,"produktName": "Softdrinks","varianteName": "Fanta","kategorie": "beverage","einzelpreis": 280,"menge": 2},{"positionId": "b0020000-0000-0000-0000-000000000036","varianteId": 4,"produktName": "Pommes","varianteName": "Klein","kategorie": "food","einzelpreis": 250,"menge": 1}],"gesamtPreisCents": 810,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.produkte-geliefert:v1', 'tisch:20', 44, now() - interval '2 hours 26 minutes', '{"lieferungId": "c0020000-0000-0000-0000-000000000015","positionen": [{"positionId": "b0020000-0000-0000-0000-000000000035","menge": 2},{"positionId": "b0020000-0000-0000-0000-000000000036","menge": 1}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.zahlung-registriert:v1', 'tisch:20', 45, now() - interval '2 hours 21 minutes', '{"zahlungId": "d0020000-0000-0000-0000-000000000015","positionen": [{"positionId": "b0020000-0000-0000-0000-000000000035","menge": 2},{"positionId": "b0020000-0000-0000-0000-000000000036","menge": 1}],"gesamtZahlungCents": 810,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.bestellung-aufgegeben:v1', 'tisch:20', 46, now() - interval '1 hours 58 minutes', '{"bestellungId": "a0020000-0000-0000-0000-000000000016","positionen": [{"positionId": "b0020000-0000-0000-0000-000000000037","varianteId": 41,"produktName": "Saftschorle","varianteName": "Rhabarberschorle 0,5l","kategorie": "beverage","einzelpreis": 350,"menge": 2}],"gesamtPreisCents": 700,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.produkte-geliefert:v1', 'tisch:20', 47, now() - interval '1 hours 54 minutes', '{"lieferungId": "c0020000-0000-0000-0000-000000000016","positionen": [{"positionId": "b0020000-0000-0000-0000-000000000037","menge": 2}],"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(6, 'Jan Hoffmann', 'tisch.zahlung-registriert:v1', 'tisch:20', 48, now() - interval '1 hours 52 minutes', '{"zahlungId": "d0020000-0000-0000-0000-000000000016","positionen": [{"positionId": "b0020000-0000-0000-0000-000000000037","menge": 2}],"gesamtZahlungCents": 700,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:20', 49, now() - interval '1 hours 36 minutes', '{"bestellungId": "a0020000-0000-0000-0000-000000000017","positionen": [{"positionId": "b0020000-0000-0000-0000-000000000038","varianteId": 24,"produktName": "Bier","varianteName": "0,5l","kategorie": "beverage","einzelpreis": 450,"menge": 2},{"positionId": "b0020000-0000-0000-0000-000000000039","varianteId": 20,"produktName": "Brezel","varianteName": "Normal","kategorie": "food","einzelpreis": 200,"menge": 3}],"gesamtPreisCents": 1500,"kommentar": ""}');

INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:20', 50, now() - interval '1 hours 21 minutes', '{"lieferungId": "c0020000-0000-0000-0000-000000000017","positionen": [{"positionId": "b0020000-0000-0000-0000-000000000038","menge": 2},{"positionId": "b0020000-0000-0000-0000-000000000039","menge": 3}],"kommentar": ""}');


-- =============================================================================
-- 5. IDENTITY-SEQUENZEN KORRIGIEREN
-- =============================================================================
SELECT setval(pg_get_serial_sequence('events', 'id'), (SELECT MAX(id) FROM events));

COMMIT;

-- =============================================================================
-- ZUSAMMENFASSUNG
-- =============================================================================
-- Gesamt-Events: 980
--   bestellung-aufgegeben: 333
--   produkte-geliefert: 331
--   produkte-storniert: 8
--   zahlung-registriert: 308
--
-- Benutzer:  11 (inkl. Admin aus Migration)
-- Tische:    22 (20 aktiv, 1 inaktiv, 1 gelöscht)
-- Produkte:  22 (18 aktiv, 2 inaktiv, 2 gelöscht [Langos, Suppe-Variante])
-- Varianten: 54
--
-- TISCH-ZUSTÄNDE:
--   Tisch 1                   |  99 Events | Offener Saldo (1750 Cent = 17.50 EUR)
--   Tisch 2                   |  73 Events | Offener Saldo (1950 Cent = 19.50 EUR)
--   Tisch 3                   |  18 Events | Abgeschlossen (Saldo = 0)
--   Tisch 4                   |  57 Events | Negativer Saldo (-1150 Cent)
--   Tisch 5                   |  95 Events | Negativer Saldo (-750 Cent)
--   Tisch 6                   |   9 Events | Offener Saldo (3390 Cent = 33.90 EUR)
--   Tisch 7                   |  23 Events | Offener Saldo (1350 Cent = 13.50 EUR)
--   Tisch 8                   | 102 Events | Offener Saldo (1500 Cent = 15.00 EUR)
--   Tisch 9                   |  83 Events | Offener Saldo (2000 Cent = 20.00 EUR)
--   Tisch 10                  |  12 Events | Negativer Saldo (-1050 Cent)
--   Tisch 11                  |   3 Events | Offener Saldo (4460 Cent = 44.60 EUR)
--   Tisch 12                  |  63 Events | Abgeschlossen (Saldo = 0)
--   Tisch 13                  |  18 Events | Abgeschlossen (Saldo = 0)
--   Tisch 14                  |   1 Events | Offener Saldo (4210 Cent = 42.10 EUR)
--   Tisch 15                  |   0 Events | Leer
--   Zelt A1                   |  75 Events | Offener Saldo (7720 Cent = 77.20 EUR)
--   Zelt A2                   |  31 Events | Offener Saldo (4000 Cent = 40.00 EUR)
--   Stehtisch Bar             |  90 Events | Abgeschlossen (Saldo = 0)
--   Stehtisch Eingang         |  78 Events | Abgeschlossen (Saldo = 0)
--   Stehtisch Terrasse        |  50 Events | Offener Saldo (1500 Cent = 15.00 EUR)
--
