BEGIN;

ALTER TABLE druckauftraege
    ADD COLUMN naechster_versuch_ab TIMESTAMPTZ NULL;

COMMENT ON COLUMN druckauftraege.naechster_versuch_ab IS
    'Fruehester Zeitpunkt des naechsten Zustellversuchs (Backoff-Nachdruck). NULL = sofort faellig; wird nach jedem Fehlversuch auf NOW() + Backoff-Wartezeit gesetzt.';

COMMENT ON COLUMN druckauftraege.status IS
    'Druckstatus: offen -> gedruckt | fehlgeschlagen (nach 6 Fehlversuchen) -> verworfen oder zurueck auf offen.';

COMMENT ON COLUMN druckauftraege.versuche IS
    'Anzahl gemeldeter Fehlversuche; ab 6 wird der Auftrag fehlgeschlagen.';

COMMIT;
