-- =============================================================================
-- jotti Seed-Daten: "Sommerfest TSV Musterstadt"
-- =============================================================================
-- Szenario: Ein Vereinsfest mitten im Betrieb. Verschiedene Tische in
-- verschiedenen Zuständen — frisch bestellt, teilgeliefert, teilbezahlt,
-- komplett abgeschlossen, storniert, leer.
--
-- Voraussetzung: Frische DB nach Schema-Migration (01_initial.up.sql).
-- Der Admin-User "nico" (id=1) existiert bereits aus der Migration.
--
-- Passwort aller Seed-User: jotti123
-- Einmalpasswort von "nico" (aus Migration): 123456
-- =============================================================================

BEGIN;

-- =============================================================================
-- 1. BENUTZER
-- =============================================================================
-- Passwort-Hash für "jotti123" (Argon2id, m=65536, t=2, p=2)
-- Alle Seed-User haben password_hash gesetzt und muss_passwort_setzen implizit = false

INSERT INTO users (id, name, username, password_hash, role, status, created_at) VALUES
  (2, 'Thomas Müller',   'thomas', '$argon2id$v=19$m=65536,t=2,p=2$OSImYG1ms0Phs26KwwMwkQ$rkoWKOIjsPz7y6ps/W2pVEhn5vTc0N95SyiveQCn404', 'admin',           'active',   now()),
  (3, 'Felix Weber',     'felix',  '$argon2id$v=19$m=65536,t=2,p=2$OSImYG1ms0Phs26KwwMwkQ$rkoWKOIjsPz7y6ps/W2pVEhn5vTc0N95SyiveQCn404', 'serviceleitung',  'active',   now()),
  (4, 'Maria Schmidt',   'maria',  '$argon2id$v=19$m=65536,t=2,p=2$OSImYG1ms0Phs26KwwMwkQ$rkoWKOIjsPz7y6ps/W2pVEhn5vTc0N95SyiveQCn404', 'service',         'active',   now()),
  (5, 'Lisa Braun',      'lisa',   '$argon2id$v=19$m=65536,t=2,p=2$OSImYG1ms0Phs26KwwMwkQ$rkoWKOIjsPz7y6ps/W2pVEhn5vTc0N95SyiveQCn404', 'service',         'active',   now()),
  (6, 'Jan Hoffmann',    'jan',    '$argon2id$v=19$m=65536,t=2,p=2$OSImYG1ms0Phs26KwwMwkQ$rkoWKOIjsPz7y6ps/W2pVEhn5vTc0N95SyiveQCn404', 'service',         'inactive', now()),
  (7, 'Sophie Becker',   'sophie', '$argon2id$v=19$m=65536,t=2,p=2$OSImYG1ms0Phs26KwwMwkQ$rkoWKOIjsPz7y6ps/W2pVEhn5vTc0N95SyiveQCn404', 'service',         'deleted',  now());

SELECT setval(pg_get_serial_sequence('users', 'id'), (SELECT MAX(id) FROM users));


-- =============================================================================
-- 2. TISCHE
-- =============================================================================

INSERT INTO tables (id, name, status, created_at) VALUES
  (1,  'Tisch 1',            'active',   now()),
  (2,  'Tisch 2',            'active',   now()),
  (3,  'Tisch 3',            'active',   now()),
  (4,  'Tisch 4',            'active',   now()),
  (5,  'Tisch 5',            'active',   now()),
  (6,  'Tisch 6',            'active',   now()),
  (7,  'Tisch 7',            'active',   now()),
  (8,  'Tisch 8',            'active',   now()),
  (9,  'Stehtisch Eingang',  'active',   now()),
  (10, 'Stehtisch Terrasse', 'active',   now()),
  (11, 'Reserviert',         'inactive', now()),
  (12, 'Alter Tisch',        'deleted',  now());

SELECT setval(pg_get_serial_sequence('tables', 'id'), (SELECT MAX(id) FROM tables));


-- =============================================================================
-- 3. PRODUKTE & VARIANTEN
-- =============================================================================

INSERT INTO products (id, name, category, status, created_at) VALUES
  (1, 'Bratwurst',    'food',     'active',  now()),
  (2, 'Pommes',       'food',     'active',  now()),
  (3, 'Flammkuchen',  'food',     'active',  now()),
  (4, 'Kuchen',       'food',     'active',  now()),
  (5, 'Bier',         'beverage', 'active',  now()),
  (6, 'Spezi',        'beverage', 'active',  now()),
  (7, 'Wasser',       'beverage', 'active',  now()),
  (8, 'Kaffee',       'beverage', 'active',  now()),
  (9, 'Glühwein',     'beverage', 'inactive', now()),  -- inaktives Produkt (nicht Saison)
  (10, 'Waffeln',     'food',     'deleted',  now());  -- gelöschtes Produkt

INSERT INTO product_variants (id, product_id, name, price_cents, status, created_at) VALUES
  -- Bratwurst (Produkt 1)
  (1,  1, 'Normal',        350, 'active',   now()),
  (2,  1, 'XXL',           500, 'active',   now()),
  -- Pommes (Produkt 2)
  (3,  2, 'Klein',         250, 'active',   now()),
  (4,  2, 'Groß',          350, 'active',   now()),
  (5,  2, 'mit Ketchup',   300, 'inactive', now()),  -- inaktive Variante
  -- Flammkuchen (Produkt 3)
  (6,  3, 'Classic',       600, 'active',   now()),
  (7,  3, 'Speck & Zwiebel', 700, 'active', now()),
  -- Kuchen (Produkt 4)
  (8,  4, 'Stück',         250, 'active',   now()),
  -- Bier (Produkt 5)
  (9,  5, '0,3l',          300, 'active',   now()),
  (10, 5, '0,5l',          450, 'active',   now()),
  -- Spezi (Produkt 6)
  (11, 6, '0,3l',          250, 'active',   now()),
  (12, 6, '0,5l',          350, 'active',   now()),
  -- Wasser (Produkt 7)
  (13, 7, '0,5l',          200, 'active',   now()),
  -- Kaffee (Produkt 8)
  (14, 8, 'Tasse',         200, 'active',   now()),
  -- Glühwein (Produkt 9, inaktiv)
  (15, 9, 'Tasse',         300, 'active',   now()),
  -- Waffeln (Produkt 10, gelöscht)
  (16, 10, 'mit Puderzucker', 250, 'deleted', now()),
  (17, 10, 'mit Kirschen',   350, 'deleted', now());

SELECT setval(pg_get_serial_sequence('products', 'id'), (SELECT MAX(id) FROM products));
SELECT setval(pg_get_serial_sequence('product_variants', 'id'), (SELECT MAX(id) FROM product_variants));


-- =============================================================================
-- 4. EVENTS
-- =============================================================================
-- Hinweis: Events sind append-only (INSERT only). Jeder Tisch hat sein eigenes
-- subject ("tisch:<id>") und eine eigene Versionsnummer (monoton steigend ab 1).
--
-- Position-UUIDs werden als deterministische UUIDs erstellt, damit Liefer-,
-- Zahlungs- und Stornierungsevents korrekt darauf verweisen können.
--
-- Zeitstempel sind relativ zu "jetzt minus Stunden", um ein realistisches
-- laufendes Fest abzubilden.
-- =============================================================================

-- ─────────────────────────────────────────────────────────────────────────────
-- TISCH 1: Komplett abgeschlossen (bestellt → geliefert → bezahlt, Saldo = 0)
-- Benutzer: Maria (4) und Felix (3)
-- ─────────────────────────────────────────────────────────────────────────────

-- Bestellung 1: 2x Bratwurst Normal, 2x Bier 0,5l → 2*350 + 2*450 = 1600
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:1', 1, now() - interval '3 hours 45 minutes', '{
  "bestellungId": "a0000001-0001-0001-0001-000000000001",
  "positionen": [
    {"positionId": "p0000001-0001-0001-0001-000000000001", "varianteId": 1, "produktName": "Bratwurst", "varianteName": "Normal", "kategorie": "food", "einzelpreis": 350, "menge": 2},
    {"positionId": "p0000001-0001-0001-0001-000000000002", "varianteId": 10, "produktName": "Bier", "varianteName": "0,5l", "kategorie": "beverage", "einzelpreis": 450, "menge": 2}
  ],
  "gesamtPreisCents": 1600,
  "kommentar": ""
}');

-- Lieferung: Alles geliefert
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:1', 2, now() - interval '3 hours 30 minutes', '{
  "lieferungId": "l0000001-0001-0001-0001-000000000001",
  "positionen": [
    {"positionId": "p0000001-0001-0001-0001-000000000001", "menge": 2},
    {"positionId": "p0000001-0001-0001-0001-000000000002", "menge": 2}
  ],
  "kommentar": ""
}');

-- Bestellung 2: 1x Flammkuchen Speck, 1x Spezi 0,3l → 700 + 250 = 950
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:1', 3, now() - interval '3 hours 15 minutes', '{
  "bestellungId": "a0000001-0001-0001-0001-000000000002",
  "positionen": [
    {"positionId": "p0000001-0001-0001-0001-000000000003", "varianteId": 7, "produktName": "Flammkuchen", "varianteName": "Speck & Zwiebel", "kategorie": "food", "einzelpreis": 700, "menge": 1},
    {"positionId": "p0000001-0001-0001-0001-000000000004", "varianteId": 11, "produktName": "Spezi", "varianteName": "0,3l", "kategorie": "beverage", "einzelpreis": 250, "menge": 1}
  ],
  "gesamtPreisCents": 950,
  "kommentar": ""
}');

-- Lieferung: Bestellung 2 geliefert
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:1', 4, now() - interval '3 hours', '{
  "lieferungId": "l0000001-0001-0001-0001-000000000002",
  "positionen": [
    {"positionId": "p0000001-0001-0001-0001-000000000003", "menge": 1},
    {"positionId": "p0000001-0001-0001-0001-000000000004", "menge": 1}
  ],
  "kommentar": ""
}');

-- Zahlung: Komplett bezahlt (1600 + 950 = 2550)
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.zahlung-registriert:v1', 'tisch:1', 5, now() - interval '2 hours 45 minutes', '{
  "zahlungId": "z0000001-0001-0001-0001-000000000001",
  "positionen": [
    {"positionId": "p0000001-0001-0001-0001-000000000001", "menge": 2},
    {"positionId": "p0000001-0001-0001-0001-000000000002", "menge": 2},
    {"positionId": "p0000001-0001-0001-0001-000000000003", "menge": 1},
    {"positionId": "p0000001-0001-0001-0001-000000000004", "menge": 1}
  ],
  "gesamtZahlungCents": 2550,
  "kommentar": ""
}');


-- ─────────────────────────────────────────────────────────────────────────────
-- TISCH 2: Bestellt + teilweise geliefert (offene Lieferungen)
-- Benutzer: Lisa (5)
-- ─────────────────────────────────────────────────────────────────────────────

-- Bestellung: 3x Pommes Groß, 2x Bier 0,3l, 1x Wasser → 3*350 + 2*300 + 200 = 1850
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.bestellung-aufgegeben:v1', 'tisch:2', 1, now() - interval '1 hour 30 minutes', '{
  "bestellungId": "a0000002-0002-0002-0002-000000000001",
  "positionen": [
    {"positionId": "p0000002-0002-0002-0002-000000000001", "varianteId": 4, "produktName": "Pommes", "varianteName": "Groß", "kategorie": "food", "einzelpreis": 350, "menge": 3},
    {"positionId": "p0000002-0002-0002-0002-000000000002", "varianteId": 9, "produktName": "Bier", "varianteName": "0,3l", "kategorie": "beverage", "einzelpreis": 300, "menge": 2},
    {"positionId": "p0000002-0002-0002-0002-000000000003", "varianteId": 13, "produktName": "Wasser", "varianteName": "0,5l", "kategorie": "beverage", "einzelpreis": 200, "menge": 1}
  ],
  "gesamtPreisCents": 1850,
  "kommentar": "Pommes ohne Salz bitte"
}');

-- Teillieferung: Nur Getränke geliefert (Bier + Wasser)
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.produkte-geliefert:v1', 'tisch:2', 2, now() - interval '1 hour 20 minutes', '{
  "lieferungId": "l0000002-0002-0002-0002-000000000001",
  "positionen": [
    {"positionId": "p0000002-0002-0002-0002-000000000002", "menge": 2},
    {"positionId": "p0000002-0002-0002-0002-000000000003", "menge": 1}
  ],
  "kommentar": "Essen kommt gleich"
}');

-- Nachbestellung: 2x Bier 0,5l → 2*450 = 900
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.bestellung-aufgegeben:v1', 'tisch:2', 3, now() - interval '1 hour 10 minutes', '{
  "bestellungId": "a0000002-0002-0002-0002-000000000002",
  "positionen": [
    {"positionId": "p0000002-0002-0002-0002-000000000004", "varianteId": 10, "produktName": "Bier", "varianteName": "0,5l", "kategorie": "beverage", "einzelpreis": 450, "menge": 2}
  ],
  "gesamtPreisCents": 900,
  "kommentar": ""
}');

-- Saldo: 1850 + 900 = 2750 (offene Lieferungen: 3x Pommes, 2x Bier 0,5l)


-- ─────────────────────────────────────────────────────────────────────────────
-- TISCH 3: Bestellt + geliefert + teilweise bezahlt (offener Saldo)
-- Benutzer: Maria (4)
-- ─────────────────────────────────────────────────────────────────────────────

-- Bestellung: 2x Bratwurst XXL, 1x Flammkuchen Classic, 3x Bier 0,5l → 2*500 + 600 + 3*450 = 2950
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:3', 1, now() - interval '2 hours 30 minutes', '{
  "bestellungId": "a0000003-0003-0003-0003-000000000001",
  "positionen": [
    {"positionId": "p0000003-0003-0003-0003-000000000001", "varianteId": 2, "produktName": "Bratwurst", "varianteName": "XXL", "kategorie": "food", "einzelpreis": 500, "menge": 2},
    {"positionId": "p0000003-0003-0003-0003-000000000002", "varianteId": 6, "produktName": "Flammkuchen", "varianteName": "Classic", "kategorie": "food", "einzelpreis": 600, "menge": 1},
    {"positionId": "p0000003-0003-0003-0003-000000000003", "varianteId": 10, "produktName": "Bier", "varianteName": "0,5l", "kategorie": "beverage", "einzelpreis": 450, "menge": 3}
  ],
  "gesamtPreisCents": 2950,
  "kommentar": ""
}');

-- Komplett geliefert
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:3', 2, now() - interval '2 hours 15 minutes', '{
  "lieferungId": "l0000003-0003-0003-0003-000000000001",
  "positionen": [
    {"positionId": "p0000003-0003-0003-0003-000000000001", "menge": 2},
    {"positionId": "p0000003-0003-0003-0003-000000000002", "menge": 1},
    {"positionId": "p0000003-0003-0003-0003-000000000003", "menge": 3}
  ],
  "kommentar": ""
}');

-- Teilzahlung: Nur Bier bezahlt (3*450 = 1350)
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:3', 3, now() - interval '2 hours', '{
  "zahlungId": "z0000003-0003-0003-0003-000000000001",
  "positionen": [
    {"positionId": "p0000003-0003-0003-0003-000000000003", "menge": 3}
  ],
  "gesamtZahlungCents": 1350,
  "kommentar": "Bier wird separat bezahlt"
}');

-- Nachbestellung: 2x Kuchen, 2x Kaffee → 2*250 + 2*200 = 900
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:3', 4, now() - interval '1 hour 45 minutes', '{
  "bestellungId": "a0000003-0003-0003-0003-000000000002",
  "positionen": [
    {"positionId": "p0000003-0003-0003-0003-000000000004", "varianteId": 8, "produktName": "Kuchen", "varianteName": "Stück", "kategorie": "food", "einzelpreis": 250, "menge": 2},
    {"positionId": "p0000003-0003-0003-0003-000000000005", "varianteId": 14, "produktName": "Kaffee", "varianteName": "Tasse", "kategorie": "beverage", "einzelpreis": 200, "menge": 2}
  ],
  "gesamtPreisCents": 900,
  "kommentar": ""
}');

-- Nachbestellung geliefert
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:3', 5, now() - interval '1 hour 30 minutes', '{
  "lieferungId": "l0000003-0003-0003-0003-000000000002",
  "positionen": [
    {"positionId": "p0000003-0003-0003-0003-000000000004", "menge": 2},
    {"positionId": "p0000003-0003-0003-0003-000000000005", "menge": 2}
  ],
  "kommentar": ""
}');

-- Saldo: 2950 + 900 - 1350 = 2500 (offen: 2x Bratwurst XXL, 1x Flammkuchen, 2x Kuchen, 2x Kaffee)


-- ─────────────────────────────────────────────────────────────────────────────
-- TISCH 4: Bestellung mit Stornierung
-- Benutzer: Lisa (5), Stornierung durch Felix (3, serviceleitung)
-- ─────────────────────────────────────────────────────────────────────────────

-- Bestellung: 4x Bratwurst Normal, 2x Pommes Klein, 4x Bier 0,3l → 4*350 + 2*250 + 4*300 = 3100
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.bestellung-aufgegeben:v1', 'tisch:4', 1, now() - interval '2 hours', '{
  "bestellungId": "a0000004-0004-0004-0004-000000000001",
  "positionen": [
    {"positionId": "p0000004-0004-0004-0004-000000000001", "varianteId": 1, "produktName": "Bratwurst", "varianteName": "Normal", "kategorie": "food", "einzelpreis": 350, "menge": 4},
    {"positionId": "p0000004-0004-0004-0004-000000000002", "varianteId": 3, "produktName": "Pommes", "varianteName": "Klein", "kategorie": "food", "einzelpreis": 250, "menge": 2},
    {"positionId": "p0000004-0004-0004-0004-000000000003", "varianteId": 9, "produktName": "Bier", "varianteName": "0,3l", "kategorie": "beverage", "einzelpreis": 300, "menge": 4}
  ],
  "gesamtPreisCents": 3100,
  "kommentar": ""
}');

-- Stornierung: 2x Bratwurst storniert (Gast hat sich umentschieden) → 2*350 = 700
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.produkte-storniert:v1', 'tisch:4', 2, now() - interval '1 hour 55 minutes', '{
  "stornierungId": "s0000004-0004-0004-0004-000000000001",
  "positionen": [
    {"positionId": "p0000004-0004-0004-0004-000000000001", "menge": 2}
  ],
  "gesamtStornierungCents": 700,
  "kommentar": "Gast hat sich umentschieden"
}');

-- Komplett geliefert (nur die verbleibenden: 2x Bratwurst, 2x Pommes, 4x Bier)
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.produkte-geliefert:v1', 'tisch:4', 3, now() - interval '1 hour 45 minutes', '{
  "lieferungId": "l0000004-0004-0004-0004-000000000001",
  "positionen": [
    {"positionId": "p0000004-0004-0004-0004-000000000001", "menge": 2},
    {"positionId": "p0000004-0004-0004-0004-000000000002", "menge": 2},
    {"positionId": "p0000004-0004-0004-0004-000000000003", "menge": 4}
  ],
  "kommentar": ""
}');

-- Teilzahlung: Bier bezahlt (4*300 = 1200)
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.zahlung-registriert:v1', 'tisch:4', 4, now() - interval '1 hour 30 minutes', '{
  "zahlungId": "z0000004-0004-0004-0004-000000000001",
  "positionen": [
    {"positionId": "p0000004-0004-0004-0004-000000000003", "menge": 4}
  ],
  "gesamtZahlungCents": 1200,
  "kommentar": ""
}');

-- Saldo: 3100 - 700 - 1200 = 1200 (offen: 2x Bratwurst, 2x Pommes)


-- ─────────────────────────────────────────────────────────────────────────────
-- TISCH 5: Mehrere Bestellungen, verschiedene Benutzer — realistischer Betrieb
-- Benutzer: Maria (4), Lisa (5), Felix (3)
-- ─────────────────────────────────────────────────────────────────────────────

-- Bestellung 1 (Maria): 2x Flammkuchen Speck, 2x Spezi 0,5l → 2*700 + 2*350 = 2100
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:5', 1, now() - interval '3 hours', '{
  "bestellungId": "a0000005-0005-0005-0005-000000000001",
  "positionen": [
    {"positionId": "p0000005-0005-0005-0005-000000000001", "varianteId": 7, "produktName": "Flammkuchen", "varianteName": "Speck & Zwiebel", "kategorie": "food", "einzelpreis": 700, "menge": 2},
    {"positionId": "p0000005-0005-0005-0005-000000000002", "varianteId": 12, "produktName": "Spezi", "varianteName": "0,5l", "kategorie": "beverage", "einzelpreis": 350, "menge": 2}
  ],
  "gesamtPreisCents": 2100,
  "kommentar": ""
}');

-- Lieferung Bestellung 1
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:5', 2, now() - interval '2 hours 45 minutes', '{
  "lieferungId": "l0000005-0005-0005-0005-000000000001",
  "positionen": [
    {"positionId": "p0000005-0005-0005-0005-000000000001", "menge": 2},
    {"positionId": "p0000005-0005-0005-0005-000000000002", "menge": 2}
  ],
  "kommentar": ""
}');

-- Bestellung 2 (Lisa): 1x Bratwurst Normal, 1x Bratwurst XXL, 2x Bier 0,5l → 350 + 500 + 2*450 = 1750
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.bestellung-aufgegeben:v1', 'tisch:5', 3, now() - interval '2 hours 30 minutes', '{
  "bestellungId": "a0000005-0005-0005-0005-000000000002",
  "positionen": [
    {"positionId": "p0000005-0005-0005-0005-000000000003", "varianteId": 1, "produktName": "Bratwurst", "varianteName": "Normal", "kategorie": "food", "einzelpreis": 350, "menge": 1},
    {"positionId": "p0000005-0005-0005-0005-000000000004", "varianteId": 2, "produktName": "Bratwurst", "varianteName": "XXL", "kategorie": "food", "einzelpreis": 500, "menge": 1},
    {"positionId": "p0000005-0005-0005-0005-000000000005", "varianteId": 10, "produktName": "Bier", "varianteName": "0,5l", "kategorie": "beverage", "einzelpreis": 450, "menge": 2}
  ],
  "gesamtPreisCents": 1750,
  "kommentar": ""
}');

-- Lieferung Bestellung 2
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.produkte-geliefert:v1', 'tisch:5', 4, now() - interval '2 hours 15 minutes', '{
  "lieferungId": "l0000005-0005-0005-0005-000000000002",
  "positionen": [
    {"positionId": "p0000005-0005-0005-0005-000000000003", "menge": 1},
    {"positionId": "p0000005-0005-0005-0005-000000000004", "menge": 1},
    {"positionId": "p0000005-0005-0005-0005-000000000005", "menge": 2}
  ],
  "kommentar": ""
}');

-- Zahlung Bestellung 1 (Felix): 2100
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.zahlung-registriert:v1', 'tisch:5', 5, now() - interval '2 hours', '{
  "zahlungId": "z0000005-0005-0005-0005-000000000001",
  "positionen": [
    {"positionId": "p0000005-0005-0005-0005-000000000001", "menge": 2},
    {"positionId": "p0000005-0005-0005-0005-000000000002", "menge": 2}
  ],
  "gesamtZahlungCents": 2100,
  "kommentar": ""
}');

-- Bestellung 3 (Maria): 3x Wasser, 1x Kuchen → 3*200 + 250 = 850
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:5', 6, now() - interval '1 hour 45 minutes', '{
  "bestellungId": "a0000005-0005-0005-0005-000000000003",
  "positionen": [
    {"positionId": "p0000005-0005-0005-0005-000000000006", "varianteId": 13, "produktName": "Wasser", "varianteName": "0,5l", "kategorie": "beverage", "einzelpreis": 200, "menge": 3},
    {"positionId": "p0000005-0005-0005-0005-000000000007", "varianteId": 8, "produktName": "Kuchen", "varianteName": "Stück", "kategorie": "food", "einzelpreis": 250, "menge": 1}
  ],
  "gesamtPreisCents": 850,
  "kommentar": ""
}');

-- Lieferung Bestellung 3
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:5', 7, now() - interval '1 hour 30 minutes', '{
  "lieferungId": "l0000005-0005-0005-0005-000000000003",
  "positionen": [
    {"positionId": "p0000005-0005-0005-0005-000000000006", "menge": 3},
    {"positionId": "p0000005-0005-0005-0005-000000000007", "menge": 1}
  ],
  "kommentar": ""
}');

-- Bestellung 4 (Lisa): Nachbestellung Bier — 3x Bier 0,3l → 3*300 = 900
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.bestellung-aufgegeben:v1', 'tisch:5', 8, now() - interval '1 hour 15 minutes', '{
  "bestellungId": "a0000005-0005-0005-0005-000000000004",
  "positionen": [
    {"positionId": "p0000005-0005-0005-0005-000000000008", "varianteId": 9, "produktName": "Bier", "varianteName": "0,3l", "kategorie": "beverage", "einzelpreis": 300, "menge": 3}
  ],
  "gesamtPreisCents": 900,
  "kommentar": ""
}');

-- Lieferung Bestellung 4
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.produkte-geliefert:v1', 'tisch:5', 9, now() - interval '1 hour', '{
  "lieferungId": "l0000005-0005-0005-0005-000000000004",
  "positionen": [
    {"positionId": "p0000005-0005-0005-0005-000000000008", "menge": 3}
  ],
  "kommentar": ""
}');

-- Saldo: 2100 + 1750 + 850 + 900 - 2100 = 3500
-- (offen: Best.2 1750 + Best.3 850 + Best.4 900 = 3500)


-- ─────────────────────────────────────────────────────────────────────────────
-- TISCH 6: Frische Bestellung, noch nichts geliefert
-- Benutzer: Maria (4)
-- ─────────────────────────────────────────────────────────────────────────────

-- Bestellung: 2x Pommes Groß, 1x Bratwurst Normal, 2x Spezi 0,3l → 2*350 + 350 + 2*250 = 1550
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:6', 1, now() - interval '10 minutes', '{
  "bestellungId": "a0000006-0006-0006-0006-000000000001",
  "positionen": [
    {"positionId": "p0000006-0006-0006-0006-000000000001", "varianteId": 4, "produktName": "Pommes", "varianteName": "Groß", "kategorie": "food", "einzelpreis": 350, "menge": 2},
    {"positionId": "p0000006-0006-0006-0006-000000000002", "varianteId": 1, "produktName": "Bratwurst", "varianteName": "Normal", "kategorie": "food", "einzelpreis": 350, "menge": 1},
    {"positionId": "p0000006-0006-0006-0006-000000000003", "varianteId": 11, "produktName": "Spezi", "varianteName": "0,3l", "kategorie": "beverage", "einzelpreis": 250, "menge": 2}
  ],
  "gesamtPreisCents": 1550,
  "kommentar": "Bratwurst bitte ohne Senf"
}');

-- Saldo: 1550 (alles offen, nichts geliefert)


-- ─────────────────────────────────────────────────────────────────────────────
-- TISCH 7: Alles bestellt, alles geliefert, nichts bezahlt
-- Benutzer: Lisa (5)
-- ─────────────────────────────────────────────────────────────────────────────

-- Bestellung 1: 2x Flammkuchen Classic, 4x Bier 0,5l → 2*600 + 4*450 = 3000
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.bestellung-aufgegeben:v1', 'tisch:7', 1, now() - interval '2 hours 30 minutes', '{
  "bestellungId": "a0000007-0007-0007-0007-000000000001",
  "positionen": [
    {"positionId": "p0000007-0007-0007-0007-000000000001", "varianteId": 6, "produktName": "Flammkuchen", "varianteName": "Classic", "kategorie": "food", "einzelpreis": 600, "menge": 2},
    {"positionId": "p0000007-0007-0007-0007-000000000002", "varianteId": 10, "produktName": "Bier", "varianteName": "0,5l", "kategorie": "beverage", "einzelpreis": 450, "menge": 4}
  ],
  "gesamtPreisCents": 3000,
  "kommentar": ""
}');

-- Geliefert
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.produkte-geliefert:v1', 'tisch:7', 2, now() - interval '2 hours 15 minutes', '{
  "lieferungId": "l0000007-0007-0007-0007-000000000001",
  "positionen": [
    {"positionId": "p0000007-0007-0007-0007-000000000001", "menge": 2},
    {"positionId": "p0000007-0007-0007-0007-000000000002", "menge": 4}
  ],
  "kommentar": ""
}');

-- Bestellung 2: 2x Kuchen, 2x Kaffee → 2*250 + 2*200 = 900
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.bestellung-aufgegeben:v1', 'tisch:7', 3, now() - interval '1 hour 45 minutes', '{
  "bestellungId": "a0000007-0007-0007-0007-000000000002",
  "positionen": [
    {"positionId": "p0000007-0007-0007-0007-000000000003", "varianteId": 8, "produktName": "Kuchen", "varianteName": "Stück", "kategorie": "food", "einzelpreis": 250, "menge": 2},
    {"positionId": "p0000007-0007-0007-0007-000000000004", "varianteId": 14, "produktName": "Kaffee", "varianteName": "Tasse", "kategorie": "beverage", "einzelpreis": 200, "menge": 2}
  ],
  "gesamtPreisCents": 900,
  "kommentar": ""
}');

-- Geliefert
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.produkte-geliefert:v1', 'tisch:7', 4, now() - interval '1 hour 30 minutes', '{
  "lieferungId": "l0000007-0007-0007-0007-000000000002",
  "positionen": [
    {"positionId": "p0000007-0007-0007-0007-000000000003", "menge": 2},
    {"positionId": "p0000007-0007-0007-0007-000000000004", "menge": 2}
  ],
  "kommentar": ""
}');

-- Saldo: 3000 + 900 = 3900 (alles geliefert, nichts bezahlt)


-- ─────────────────────────────────────────────────────────────────────────────
-- TISCH 8: Leer — keine Events (frischer aktiver Tisch)
-- ─────────────────────────────────────────────────────────────────────────────
-- (Keine Events — Tisch existiert nur als Stammdaten)


-- ─────────────────────────────────────────────────────────────────────────────
-- TISCH 9 (Stehtisch Eingang): Kleine Bestellung, komplett bezahlt
-- Benutzer: Maria (4)
-- ─────────────────────────────────────────────────────────────────────────────

-- Bestellung: 2x Bier 0,3l, 1x Bratwurst Normal → 2*300 + 350 = 950
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:9', 1, now() - interval '3 hours 30 minutes', '{
  "bestellungId": "a0000009-0009-0009-0009-000000000001",
  "positionen": [
    {"positionId": "p0000009-0009-0009-0009-000000000001", "varianteId": 9, "produktName": "Bier", "varianteName": "0,3l", "kategorie": "beverage", "einzelpreis": 300, "menge": 2},
    {"positionId": "p0000009-0009-0009-0009-000000000002", "varianteId": 1, "produktName": "Bratwurst", "varianteName": "Normal", "kategorie": "food", "einzelpreis": 350, "menge": 1}
  ],
  "gesamtPreisCents": 950,
  "kommentar": ""
}');

-- Geliefert
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:9', 2, now() - interval '3 hours 20 minutes', '{
  "lieferungId": "l0000009-0009-0009-0009-000000000001",
  "positionen": [
    {"positionId": "p0000009-0009-0009-0009-000000000001", "menge": 2},
    {"positionId": "p0000009-0009-0009-0009-000000000002", "menge": 1}
  ],
  "kommentar": ""
}');

-- Bezahlt
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:9', 3, now() - interval '3 hours 10 minutes', '{
  "zahlungId": "z0000009-0009-0009-0009-000000000001",
  "positionen": [
    {"positionId": "p0000009-0009-0009-0009-000000000001", "menge": 2},
    {"positionId": "p0000009-0009-0009-0009-000000000002", "menge": 1}
  ],
  "gesamtZahlungCents": 950,
  "kommentar": ""
}');

-- Neue Runde: 3x Bier 0,5l, 1x Pommes Klein → 3*450 + 250 = 1600
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:9', 4, now() - interval '2 hours 50 minutes', '{
  "bestellungId": "a0000009-0009-0009-0009-000000000002",
  "positionen": [
    {"positionId": "p0000009-0009-0009-0009-000000000003", "varianteId": 10, "produktName": "Bier", "varianteName": "0,5l", "kategorie": "beverage", "einzelpreis": 450, "menge": 3},
    {"positionId": "p0000009-0009-0009-0009-000000000004", "varianteId": 3, "produktName": "Pommes", "varianteName": "Klein", "kategorie": "food", "einzelpreis": 250, "menge": 1}
  ],
  "gesamtPreisCents": 1600,
  "kommentar": ""
}');

-- Geliefert
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:9', 5, now() - interval '2 hours 40 minutes', '{
  "lieferungId": "l0000009-0009-0009-0009-000000000002",
  "positionen": [
    {"positionId": "p0000009-0009-0009-0009-000000000003", "menge": 3},
    {"positionId": "p0000009-0009-0009-0009-000000000004", "menge": 1}
  ],
  "kommentar": ""
}');

-- Bezahlt
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.zahlung-registriert:v1', 'tisch:9', 6, now() - interval '2 hours 30 minutes', '{
  "zahlungId": "z0000009-0009-0009-0009-000000000002",
  "positionen": [
    {"positionId": "p0000009-0009-0009-0009-000000000003", "menge": 3},
    {"positionId": "p0000009-0009-0009-0009-000000000004", "menge": 1}
  ],
  "gesamtZahlungCents": 1600,
  "kommentar": ""
}');

-- Dritte Runde: 2x Spezi 0,5l → 2*350 = 700
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.bestellung-aufgegeben:v1', 'tisch:9', 7, now() - interval '2 hours', '{
  "bestellungId": "a0000009-0009-0009-0009-000000000003",
  "positionen": [
    {"positionId": "p0000009-0009-0009-0009-000000000005", "varianteId": 12, "produktName": "Spezi", "varianteName": "0,5l", "kategorie": "beverage", "einzelpreis": 350, "menge": 2}
  ],
  "gesamtPreisCents": 700,
  "kommentar": ""
}');

-- Geliefert
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.produkte-geliefert:v1', 'tisch:9', 8, now() - interval '1 hour 50 minutes', '{
  "lieferungId": "l0000009-0009-0009-0009-000000000003",
  "positionen": [
    {"positionId": "p0000009-0009-0009-0009-000000000005", "menge": 2}
  ],
  "kommentar": ""
}');

-- Saldo: 950 + 1600 + 700 - 950 - 1600 = 700 (offen: 2x Spezi 0,5l)


-- ─────────────────────────────────────────────────────────────────────────────
-- TISCH 10 (Stehtisch Terrasse): Stornierung nach Bezahlung (negativer Saldo)
-- Benutzer: Lisa (5), Stornierung durch Thomas (2, admin)
-- ─────────────────────────────────────────────────────────────────────────────

-- Bestellung: 2x Bratwurst Normal, 2x Bier 0,3l → 2*350 + 2*300 = 1300
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.bestellung-aufgegeben:v1', 'tisch:10', 1, now() - interval '1 hour 30 minutes', '{
  "bestellungId": "a0000010-0010-0010-0010-000000000001",
  "positionen": [
    {"positionId": "p0000010-0010-0010-0010-000000000001", "varianteId": 1, "produktName": "Bratwurst", "varianteName": "Normal", "kategorie": "food", "einzelpreis": 350, "menge": 2},
    {"positionId": "p0000010-0010-0010-0010-000000000002", "varianteId": 9, "produktName": "Bier", "varianteName": "0,3l", "kategorie": "beverage", "einzelpreis": 300, "menge": 2}
  ],
  "gesamtPreisCents": 1300,
  "kommentar": ""
}');

-- Geliefert
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.produkte-geliefert:v1', 'tisch:10', 2, now() - interval '1 hour 20 minutes', '{
  "lieferungId": "l0000010-0010-0010-0010-000000000001",
  "positionen": [
    {"positionId": "p0000010-0010-0010-0010-000000000001", "menge": 2},
    {"positionId": "p0000010-0010-0010-0010-000000000002", "menge": 2}
  ],
  "kommentar": ""
}');

-- Komplett bezahlt
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.zahlung-registriert:v1', 'tisch:10', 3, now() - interval '1 hour', '{
  "zahlungId": "z0000010-0010-0010-0010-000000000001",
  "positionen": [
    {"positionId": "p0000010-0010-0010-0010-000000000001", "menge": 2},
    {"positionId": "p0000010-0010-0010-0010-000000000002", "menge": 2}
  ],
  "gesamtZahlungCents": 1300,
  "kommentar": ""
}');

-- Stornierung NACH Bezahlung: 1x Bratwurst war schlecht → 350
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(2, 'Thomas Müller', 'tisch.produkte-storniert:v1', 'tisch:10', 4, now() - interval '50 minutes', '{
  "stornierungId": "s0000010-0010-0010-0010-000000000001",
  "positionen": [
    {"positionId": "p0000010-0010-0010-0010-000000000001", "menge": 1}
  ],
  "gesamtStornierungCents": 350,
  "kommentar": "Bratwurst war nicht in Ordnung, Geld zurück"
}');

-- Saldo: 1300 - 1300 - 350 = -350 (negativer Saldo — Geld muss zurückgegeben werden)


-- ─────────────────────────────────────────────────────────────────────────────
-- TISCH 3: Weitere Aktivität — noch eine Runde
-- ─────────────────────────────────────────────────────────────────────────────

-- Bestellung 3: 2x Bier 0,3l → 2*300 = 600
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:3', 6, now() - interval '1 hour', '{
  "bestellungId": "a0000003-0003-0003-0003-000000000003",
  "positionen": [
    {"positionId": "p0000003-0003-0003-0003-000000000006", "varianteId": 9, "produktName": "Bier", "varianteName": "0,3l", "kategorie": "beverage", "einzelpreis": 300, "menge": 2}
  ],
  "gesamtPreisCents": 600,
  "kommentar": ""
}');

-- Geliefert
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:3', 7, now() - interval '50 minutes', '{
  "lieferungId": "l0000003-0003-0003-0003-000000000003",
  "positionen": [
    {"positionId": "p0000003-0003-0003-0003-000000000006", "menge": 2}
  ],
  "kommentar": ""
}');

-- Saldo Tisch 3 aktualisiert: 2950 + 900 + 600 - 1350 = 3100


-- ─────────────────────────────────────────────────────────────────────────────
-- TISCH 5: Stornierung einer Position + Nachbestellung
-- ─────────────────────────────────────────────────────────────────────────────

-- Stornierung: 1x Kuchen aus Bestellung 3 → 250
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.produkte-storniert:v1', 'tisch:5', 10, now() - interval '55 minutes', '{
  "stornierungId": "s0000005-0005-0005-0005-000000000001",
  "positionen": [
    {"positionId": "p0000005-0005-0005-0005-000000000007", "menge": 1}
  ],
  "gesamtStornierungCents": 250,
  "kommentar": "Kuchen war leider aus"
}');

-- Nachbestellung: 1x Flammkuchen Classic → 600
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.bestellung-aufgegeben:v1', 'tisch:5', 11, now() - interval '50 minutes', '{
  "bestellungId": "a0000005-0005-0005-0005-000000000005",
  "positionen": [
    {"positionId": "p0000005-0005-0005-0005-000000000009", "varianteId": 6, "produktName": "Flammkuchen", "varianteName": "Classic", "kategorie": "food", "einzelpreis": 600, "menge": 1}
  ],
  "gesamtPreisCents": 600,
  "kommentar": ""
}');

-- Geliefert
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:5', 12, now() - interval '35 minutes', '{
  "lieferungId": "l0000005-0005-0005-0005-000000000005",
  "positionen": [
    {"positionId": "p0000005-0005-0005-0005-000000000009", "menge": 1}
  ],
  "kommentar": ""
}');

-- Saldo Tisch 5: 2100 + 1750 + 850 + 900 + 600 - 2100 - 250 = 3850


-- ─────────────────────────────────────────────────────────────────────────────
-- TISCH 4: Restbetrag bezahlt — wird komplett abgeschlossen
-- ─────────────────────────────────────────────────────────────────────────────

-- Restzahlung: 2x Bratwurst + 2x Pommes = 2*350 + 2*250 = 1200
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.zahlung-registriert:v1', 'tisch:4', 5, now() - interval '45 minutes', '{
  "zahlungId": "z0000004-0004-0004-0004-000000000002",
  "positionen": [
    {"positionId": "p0000004-0004-0004-0004-000000000001", "menge": 2},
    {"positionId": "p0000004-0004-0004-0004-000000000002", "menge": 2}
  ],
  "gesamtZahlungCents": 1200,
  "kommentar": ""
}');

-- Neue Bestellung nach Bezahlung: 1x Kaffee, 1x Kuchen → 200 + 250 = 450
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.bestellung-aufgegeben:v1', 'tisch:4', 6, now() - interval '40 minutes', '{
  "bestellungId": "a0000004-0004-0004-0004-000000000002",
  "positionen": [
    {"positionId": "p0000004-0004-0004-0004-000000000004", "varianteId": 14, "produktName": "Kaffee", "varianteName": "Tasse", "kategorie": "beverage", "einzelpreis": 200, "menge": 1},
    {"positionId": "p0000004-0004-0004-0004-000000000005", "varianteId": 8, "produktName": "Kuchen", "varianteName": "Stück", "kategorie": "food", "einzelpreis": 250, "menge": 1}
  ],
  "gesamtPreisCents": 450,
  "kommentar": ""
}');

-- Geliefert
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.produkte-geliefert:v1', 'tisch:4', 7, now() - interval '30 minutes', '{
  "lieferungId": "l0000004-0004-0004-0004-000000000002",
  "positionen": [
    {"positionId": "p0000004-0004-0004-0004-000000000004", "menge": 1},
    {"positionId": "p0000004-0004-0004-0004-000000000005", "menge": 1}
  ],
  "kommentar": ""
}');

-- Saldo Tisch 4: 3100 + 450 - 700 - 1200 - 1200 = 450 (offen: 1x Kaffee, 1x Kuchen)


-- ─────────────────────────────────────────────────────────────────────────────
-- TISCH 2: Pommes endlich geliefert + neue Bestellung
-- ─────────────────────────────────────────────────────────────────────────────

-- Pommes geliefert
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.produkte-geliefert:v1', 'tisch:2', 4, now() - interval '55 minutes', '{
  "lieferungId": "l0000002-0002-0002-0002-000000000002",
  "positionen": [
    {"positionId": "p0000002-0002-0002-0002-000000000001", "menge": 3}
  ],
  "kommentar": ""
}');

-- Nachbestellung Bier geliefert
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.produkte-geliefert:v1', 'tisch:2', 5, now() - interval '50 minutes', '{
  "lieferungId": "l0000002-0002-0002-0002-000000000003",
  "positionen": [
    {"positionId": "p0000002-0002-0002-0002-000000000004", "menge": 2}
  ],
  "kommentar": ""
}');

-- Bestellung 3: 1x Flammkuchen Speck, 1x Wasser → 700 + 200 = 900
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.bestellung-aufgegeben:v1', 'tisch:2', 6, now() - interval '40 minutes', '{
  "bestellungId": "a0000002-0002-0002-0002-000000000003",
  "positionen": [
    {"positionId": "p0000002-0002-0002-0002-000000000005", "varianteId": 7, "produktName": "Flammkuchen", "varianteName": "Speck & Zwiebel", "kategorie": "food", "einzelpreis": 700, "menge": 1},
    {"positionId": "p0000002-0002-0002-0002-000000000006", "varianteId": 13, "produktName": "Wasser", "varianteName": "0,5l", "kategorie": "beverage", "einzelpreis": 200, "menge": 1}
  ],
  "gesamtPreisCents": 900,
  "kommentar": ""
}');

-- Geliefert
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.produkte-geliefert:v1', 'tisch:2', 7, now() - interval '25 minutes', '{
  "lieferungId": "l0000002-0002-0002-0002-000000000004",
  "positionen": [
    {"positionId": "p0000002-0002-0002-0002-000000000005", "menge": 1},
    {"positionId": "p0000002-0002-0002-0002-000000000006", "menge": 1}
  ],
  "kommentar": ""
}');

-- Teilzahlung: Erste Bestellung bezahlt (1850)
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.zahlung-registriert:v1', 'tisch:2', 8, now() - interval '15 minutes', '{
  "zahlungId": "z0000002-0002-0002-0002-000000000001",
  "positionen": [
    {"positionId": "p0000002-0002-0002-0002-000000000001", "menge": 3},
    {"positionId": "p0000002-0002-0002-0002-000000000002", "menge": 2},
    {"positionId": "p0000002-0002-0002-0002-000000000003", "menge": 1}
  ],
  "gesamtZahlungCents": 1850,
  "kommentar": ""
}');

-- Saldo Tisch 2: 1850 + 900 + 900 - 1850 = 1800 (offen: 2x Bier 0,5l + Flammkuchen + Wasser)


-- ─────────────────────────────────────────────────────────────────────────────
-- TISCH 7: Teilzahlung
-- ─────────────────────────────────────────────────────────────────────────────

-- Teilzahlung: Kaffee + Kuchen bezahlt (900)
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(3, 'Felix Weber', 'tisch.zahlung-registriert:v1', 'tisch:7', 5, now() - interval '20 minutes', '{
  "zahlungId": "z0000007-0007-0007-0007-000000000001",
  "positionen": [
    {"positionId": "p0000007-0007-0007-0007-000000000003", "menge": 2},
    {"positionId": "p0000007-0007-0007-0007-000000000004", "menge": 2}
  ],
  "gesamtZahlungCents": 900,
  "kommentar": ""
}');

-- Nachbestellung: 2x Bier 0,3l → 2*300 = 600
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(5, 'Lisa Braun', 'tisch.bestellung-aufgegeben:v1', 'tisch:7', 6, now() - interval '15 minutes', '{
  "bestellungId": "a0000007-0007-0007-0007-000000000003",
  "positionen": [
    {"positionId": "p0000007-0007-0007-0007-000000000005", "varianteId": 9, "produktName": "Bier", "varianteName": "0,3l", "kategorie": "beverage", "einzelpreis": 300, "menge": 2}
  ],
  "gesamtPreisCents": 600,
  "kommentar": ""
}');

-- Saldo Tisch 7: 3000 + 900 + 600 - 900 = 3600 (offen: 2x Flammkuchen, 4x Bier 0,5l, 2x Bier 0,3l ungeliefert)


-- ─────────────────────────────────────────────────────────────────────────────
-- TISCH 6: Getränke schnell geliefert, Essen bleibt offen
-- ─────────────────────────────────────────────────────────────────────────────

-- Teillieferung: Nur Getränke
INSERT INTO events (user_id, user_name, type, subject, version, timestamp, data) VALUES
(4, 'Maria Schmidt', 'tisch.produkte-geliefert:v1', 'tisch:6', 2, now() - interval '5 minutes', '{
  "lieferungId": "l0000006-0006-0006-0006-000000000001",
  "positionen": [
    {"positionId": "p0000006-0006-0006-0006-000000000003", "menge": 2}
  ],
  "kommentar": "Essen braucht noch etwas"
}');

-- Saldo Tisch 6: 1550 (unverändert, nichts bezahlt)


-- =============================================================================
-- 5. IDENTITY-SEQUENZEN KORRIGIEREN
-- =============================================================================
-- Nach expliziten ID-Inserts die Sequenzen auf den maximalen Wert setzen,
-- damit nachfolgende Auto-Inserts korrekt funktionieren.

SELECT setval(pg_get_serial_sequence('events', 'id'), (SELECT MAX(id) FROM events));

COMMIT;

-- =============================================================================
-- ZUSAMMENFASSUNG DER TISCH-ZUSTÄNDE
-- =============================================================================
--
-- Tisch 1:  Saldo =     0 | Komplett abgeschlossen (2 Best., alles geliefert, alles bezahlt)
-- Tisch 2:  Saldo =  1800 | Alles geliefert, teilweise bezahlt (3 Best., 1 Zahlung)
-- Tisch 3:  Saldo =  3100 | Alles geliefert, teilweise bezahlt (3 Best., 1 Zahlung)
-- Tisch 4:  Saldo =   450 | Stornierung + teilweise bezahlt (2 Best., 1 Storno, 2 Zahlungen)
-- Tisch 5:  Saldo =  3850 | Realistischer Betrieb (5 Best., 1 Storno, 1 Zahlung, mehrere User)
-- Tisch 6:  Saldo =  1550 | Frisch, Getränke geliefert, Essen offen, nichts bezahlt
-- Tisch 7:  Saldo =  3600 | Teilbezahlt + offene Nachbestellung (ungeliefert)
-- Tisch 8:  Saldo =     0 | Leer — keine Events
-- Tisch 9:  Saldo =   700 | Stehtisch: 2 Runden bezahlt, 3. Runde offen
-- Tisch 10: Saldo =  -350 | Negativer Saldo (Stornierung nach Bezahlung)
-- Tisch 11: (inaktiv)     | Nicht im Service sichtbar
-- Tisch 12: (gelöscht)    | Nicht im System sichtbar
--
-- Events gesamt: ~90
-- Benutzer: 7 (1 aus Migration + 6 Seed: 2 aktiv admin, 1 aktiv serviceleitung,
--              2 aktiv service, 1 inaktiv service, 1 gelöscht service)
-- Produkte: 10 (8 aktiv, 1 inaktiv, 1 gelöscht) mit 17 Varianten
-- Tische: 12 (10 aktiv, 1 inaktiv, 1 gelöscht)
