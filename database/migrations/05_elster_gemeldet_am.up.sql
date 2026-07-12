-- Ergänzt die Betreiber-Stammdaten um das Datum der ELSTER-Kassenmeldung. Nach
-- § 146a Abs. 4 AO muss jede Kasse innerhalb eines Monats nach Inbetriebnahme
-- über ELSTER beim Finanzamt gemeldet werden. jotti meldet nicht automatisch;
-- der Admin führt die Meldung manuell im ELSTER-Portal durch und hakt sie hier
-- ab. Die Spalte hält fest, an welchem Tag das geschah (NULL = noch offen).
--
-- Additiv und nullbar: bestehende Betreiber-Zeilen (Singleton, id=1) bleiben
-- gültig; die Meldung gilt als offen, bis der Admin sie setzt.
BEGIN;

ALTER TABLE betreiber
    ADD COLUMN elster_gemeldet_am DATE NULL;

COMMENT ON COLUMN betreiber.elster_gemeldet_am IS 'Datum der ELSTER-Kassenmeldung (§ 146a Abs. 4 AO); NULL = noch nicht gemeldet.';

COMMIT;
