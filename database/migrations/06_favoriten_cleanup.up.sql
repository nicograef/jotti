-- Räumt verwaiste Tisch-Markierungen ab: Zeilen in tisch_favoriten, deren Tisch
-- gelöscht ist (status = 'deleted'). Sie sind entstanden, weil das Löschen eines
-- Tisches bisher nur dessen Status setzte, ohne die Markierungen zu entfernen.
-- Eine solche Markierung ist für die betroffene Servicekraft weder sichtbar noch
-- abwählbar (gelöschte Tische erscheinen nicht in der Tischauswahl) und legte
-- zuvor ihre gesamte Tischübersicht lahm.
--
-- tisch_favoriten hält ausschließlich Benutzer-Präferenzen und keine
-- aufbewahrungspflichtigen Kassendaten; Kassenjournal und Projektionen bleiben
-- unberührt. Der Fremdschlüssel tisch_favoriten.tisch_id -> tische(id) schließt
-- Zeilen ohne zugehörigen Tisch aus; die Bedingung "kein nicht-gelöschter Tisch"
-- deckt den Bestand damit vollständig ab.
BEGIN;

DELETE FROM tisch_favoriten f
WHERE NOT EXISTS (
    SELECT 1 FROM tische t
    WHERE t.id = f.tisch_id AND t.status != 'deleted'
);

COMMIT;
