-- Erlaubt der Druckstation 'abholbon' den dritten Bonmodus 'pro_stueck': je
-- Einheit einer Position ein eigener Abholbon. Gäste und Gruppen kaufen im
-- Direktverkauf mehrere Einheiten auf einmal (3x Bier) und lösen sie einzeln
-- an der Theke ein; dafür braucht jede Einheit ihren eigenen Bon.
--
-- Die Produktstationen (essen, getraenk, sonstiges) drucken Arbeitsbons und
-- kennen den Modus nicht; der Kassenbeleg trägt weiterhin keinen Bonmodus.
--
-- Der CHECK auf druckstationen (aus 01_initial.up.sql) lässt sich nicht in
-- place ändern; er wird gelöscht und mit dem zusätzlichen Wert neu angelegt.
-- Er referenziert zwei Spalten und ist anonym, sein generierter Name folgt
-- daher der PostgreSQL-Konvention <tabelle>_check. Die neue Bedingung ist eine
-- Erweiterung: alle bestehenden Bonmodus-Werte bleiben gültig, keine Zeile
-- wird geändert.
BEGIN;

ALTER TABLE druckstationen DROP CONSTRAINT druckstationen_check;

ALTER TABLE druckstationen
    ADD CONSTRAINT druckstationen_check
    CHECK (
        (kategorie = 'abholbon' AND bonmodus IN ('pro_position', 'pro_bestellung', 'pro_stueck'))
        OR (kategorie IN ('essen', 'getraenk', 'sonstiges') AND bonmodus IN ('pro_position', 'pro_bestellung'))
        OR (kategorie = 'kassenbeleg' AND bonmodus IS NULL)
    );

COMMENT ON COLUMN druckstationen.bonmodus IS 'Bonmodus: pro_position (1 Bon pro Position) oder pro_bestellung (1 Sammelbon) für essen/getraenk/sonstiges/abholbon, zusätzlich pro_stueck (1 Bon je Einheit) nur für abholbon; NULL für kassenbeleg.';

COMMIT;
