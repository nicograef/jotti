-- Erlaubt die Bon-Art 'testbon' für Druckaufträge. Ein Testbon ist ein regulärer
-- Druckauftrag (Stationsname + Zeitstempel), den der Admin beim Aufbau am
-- Festmorgen an eine konfigurierte Station schickt, um Drucker und Netzwerk zu
-- prüfen. Er läuft über dieselbe Outbox wie Arbeitsbon und Kassenbeleg und
-- erscheint bei Fehlschlag wie jeder andere fehlgeschlagene Druckauftrag.
--
-- Der CHECK auf druckauftraege.bon_art (aus 01_initial.up.sql) lässt sich nicht
-- in place ändern; er wird gelöscht und mit dem zusätzlichen Wert neu angelegt.
-- Der Constraint-Name folgt der PostgreSQL-Konvention <tabelle>_<spalte>_check.
BEGIN;

ALTER TABLE druckauftraege DROP CONSTRAINT druckauftraege_bon_art_check;

ALTER TABLE druckauftraege
    ADD CONSTRAINT druckauftraege_bon_art_check
    CHECK (bon_art IN ('arbeitsbon', 'kassenbeleg', 'testbon'));

COMMIT;
