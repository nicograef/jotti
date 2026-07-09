---
title: Datenaufbewahrung
description: 'Kassendaten 10 Jahre sicher aufbewahren: DSFinV-K-Export je Kassensitzung und das Datenbank-Backup als Sicherheitsnetz.'
---

Alle Kassendaten müssen 10 Jahre vollständig, lesbar und unveränderbar aufbewahrt
werden (§§ 146, 147 AO). jotti stellt sie in offenen, ohne Spezialsoftware lesbaren
Formaten bereit; die sichere Aufbewahrung selbst ist eure Aufgabe als Betreiber.
Zwei Dinge gehören ins Archiv.

**DSFinV-K-Export je Kassensitzung (das Wichtigste).** Das ist die vom Finanzamt
erwartete Standardform eurer Kassendaten, lesbar mit jeder Tabellenkalkulation. So
sichert ihr ihn nach jedem Veranstaltungstag:

1. Im Admin-Bereich unter „Auswertungen" das „Dashboard" öffnen und im Abschnitt
   „Historische Auswertung" die abgeschlossene Kassensitzung auswählen.
2. Auf „DSFinV-K-Export" klicken. jotti lädt eine ZIP-Datei in den Download-Ordner.
3. Prüfen: die ZIP-Datei öffnen. Darin liegen mehrere CSV-Dateien, eine `index.xml`
   und eine `.dtd`. Lässt sie sich öffnen und sind die CSV-Dateien nicht leer, ist
   der Export in Ordnung.
4. An zwei getrennte Orte kopieren, etwa auf einen USB-Stick und zusätzlich in einen
   Cloud-Speicher. Ein einzelner Speicherort genügt nicht.

**Datenbank-Backup als Sicherheitsnetz.** Es enthält das vollständige Kassenjournal
im Rohformat samt TSE-Signaturen und Stammdaten. Auf dem Kassenrechner zieht jotti
vor jedem Update automatisch ein Backup; geht ein Update schief, stellt
`jotti-restore.cmd` (Doppelklick) das letzte zurück. Diese Backups liegen auf
demselben Rechner und gehen mit ihm verloren. Euer vom Rechner unabhängiges Archiv
ist deshalb der DSFinV-K-Export oben. Wer zusätzlich die Rohdaten außer Haus sichern
will, braucht dafür nicht zwingend einen Server: Auf dem Windows-Rechner könnt ihr
den Ordner `C:\ProgramData\jotti\backups` komplett auf einen USB-Stick oder in eine
Cloud kopieren; dorthin spiegelt jotti die automatischen Pre-Update-Backups, und die
`KURZANLEITUNG.md` im ZIP zeigt einen Befehl für ein weiteres Backup auf Wunsch. Wer
die Rohdaten laufend automatisch außer Haus sichern will, betreibt jotti auf einem
Server (siehe [Backups](aktualisieren-backups.md#backups) im Experten-Weg).

Ebenfalls aufbewahren: die Z-Bons (Tagesabschlüsse, im DSFinV-K-Export enthalten)
und die Zählprotokolle vom Kassensturz. Sorgt dafür, dass nur berechtigte Personen
Zugriff auf Rechner und Daten haben.
