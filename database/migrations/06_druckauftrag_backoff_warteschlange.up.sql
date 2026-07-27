-- Korrigiert den Spaltenkommentar aus 02_druckauftrag_backoff: die Backoff-
-- Wartezeit bremst seit dem Wechsel auf den warteschlangenweiten Poll nicht mehr
-- nur die Zeile, auf der sie steht. GetOffeneDruckauftraege ueberspringt die
-- ganze Ziel-IP, solange irgendein offener Auftrag dieses Druckers wartet — ein
-- Auftrag mit naechster_versuch_ab = NULL ist also nicht mehr zwingend sofort
-- faellig. Der alte Kommentar behauptet das Gegenteil und ist ueber \d+
-- druckauftraege fuer Betreiber sichtbar.
--
-- Reine Kommentaraenderung: kein Schema-, kein Datenbestand betroffen.
BEGIN;

COMMENT ON COLUMN druckauftraege.naechster_versuch_ab IS
    'Fruehester Zeitpunkt des naechsten Zustellversuchs (Backoff-Nachdruck); wird nach jedem Fehlversuch auf NOW() + Backoff-Wartezeit gesetzt. Die Wartezeit gilt der gesamten Warteschlange der Ziel-IP: solange irgendein offener Auftrag dieses Druckers wartet, wird kein Auftrag dieser Ziel-IP ausgeliefert — auch keiner mit NULL.';

COMMIT;
