-- Erlaubt der Druckstation 'abholbon' den Bonmodus 'kein_bon': ein Direktverkauf
-- druckt dann gar keinen Bon, weder Abholbon noch Arbeitsbon an den
-- Produktstationen. Am Verkaufsort (Zapfanlage, Theke) braucht es oft keinen Bon.
-- Eine leere Abholbon-IP kann das nicht ausdrücken: Sie leitet den Direktverkauf
-- an die Produktstationen, die für die Tischbestellungen IPs tragen.
--
-- Der CHECK druckstationen_check (zuletzt aus 08_abholbon_pro_stueck.up.sql)
-- wird gelöscht und mit dem zusätzlichen Wert neu angelegt. Die neue Bedingung
-- ist eine Erweiterung: alle bestehenden Bonmodus-Werte bleiben gültig, keine
-- Zeile wird geändert.
BEGIN;

ALTER TABLE druckstationen DROP CONSTRAINT druckstationen_check;

ALTER TABLE druckstationen
    ADD CONSTRAINT druckstationen_check
    CHECK (
        (kategorie = 'abholbon' AND bonmodus IN ('pro_position', 'pro_bestellung', 'pro_stueck', 'kein_bon'))
        OR (kategorie IN ('essen', 'getraenk', 'sonstiges') AND bonmodus IN ('pro_position', 'pro_bestellung'))
        OR (kategorie = 'kassenbeleg' AND bonmodus IS NULL)
    );

COMMENT ON COLUMN druckstationen.bonmodus IS 'Bonmodus: pro_position (1 Bon pro Position) oder pro_bestellung (1 Sammelbon) für essen/getraenk/sonstiges/abholbon; nur für abholbon zusätzlich pro_stueck (1 Bon je Einheit) und kein_bon (Direktverkauf druckt keinen Bon); NULL für kassenbeleg.';

COMMIT;
