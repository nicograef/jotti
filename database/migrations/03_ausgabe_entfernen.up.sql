-- Entfernt die Ausgabe-Bestätigung (Event ausgabe-bestaetigt:v1) vollständig aus
-- dem Datenbestand. Begründung und einmalige Ausnahme vom Append-only-Prinzip des
-- Kassenjournals: siehe docs/adrs/01_ausgabe-bestaetigen.md.
--
-- Die Event-Lesepfade sind exklusiv (unbekannter Event-Typ → Fehler) und die
-- tisch_sessions-Projektion wird beim Backend-Start vollständig aus dem Journal neu
-- aufgebaut. Verbliebene ausgabe-bestaetigt:v1-Events würden den Start verhindern,
-- daher werden sie hier gelöscht. Der Delete-Trigger schützt das Kassenjournal auch
-- gegen den Table-Owner; die Migration deaktiviert ihn nur innerhalb dieser
-- Transaktion und aktiviert ihn vor dem COMMIT wieder — der Schreibschutz besteht
-- danach unverändert.
BEGIN;

ALTER TABLE kassenjournal DISABLE TRIGGER kassenjournal_no_delete;

-- Projektion leeren (FK auf kassenjournal-Events); wird beim Start neu aufgebaut.
DELETE FROM tisch_sessions;

DELETE FROM kassenjournal WHERE type = 'ausgabe-bestaetigt:v1';

ALTER TABLE kassenjournal ENABLE TRIGGER kassenjournal_no_delete;

ALTER TABLE tisch_sessions DROP COLUMN ausstehende_positionen;

COMMIT;
