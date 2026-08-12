-- Ergänzt Produkte und Varianten um eine explizite Anzeigereihenfolge. Bisher
-- richtete sich die Reihenfolge in Admin und Service nach der id, also nach dem
-- Anlagezeitpunkt. Bei Preislisten mit vielen Varianten je Produkt lässt sich
-- die Bestelliste dadurch nicht an den tatsächlichen Verkaufsablauf anpassen;
-- häufig bestellte Artikel stehen unten und erzwingen langes Scrollen auf dem
-- Service-Handy.
--
-- Die Reihenfolge gilt jeweils innerhalb ihres Geltungsbereichs: bei Produkten
-- innerhalb der Kategorie, bei Varianten innerhalb ihres Produkts. Sortiert
-- wird immer nach (reihenfolge, id); die id bleibt Tiebreaker und hält die
-- Sortierung auch bei gleichen Werten deterministisch.
--
-- Additiv mit Backfill nach id: die heutige Anzeigereihenfolge bleibt in jeder
-- bestehenden Instanz exakt erhalten, bis ein Admin sie bewusst ändert.
BEGIN;

ALTER TABLE produkte
    ADD COLUMN reihenfolge INT NOT NULL DEFAULT 0;

ALTER TABLE produkt_varianten
    ADD COLUMN reihenfolge INT NOT NULL DEFAULT 0;

UPDATE produkte SET reihenfolge = id;
UPDATE produkt_varianten SET reihenfolge = id;

COMMENT ON COLUMN produkte.reihenfolge IS 'Anzeigereihenfolge innerhalb der Kategorie (aufsteigend, Tiebreaker id)';
COMMENT ON COLUMN produkt_varianten.reihenfolge IS 'Anzeigereihenfolge innerhalb des Produkts (aufsteigend, Tiebreaker id)';

COMMIT;
