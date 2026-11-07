---
title: Datenaufbewahrung
description: 'Kassendaten 10 Jahre sicher aufbewahren: DSFinV-K-Export je Kassensitzung und das Datenbank-Backup als Sicherheitsnetz.'
---

Alle Kassendaten müssen 10 Jahre vollständig, lesbar und unveränderbar aufbewahrt
werden (§§ 146, 147 AO). jotti stellt sie in offenen, ohne Spezialsoftware lesbaren
Formaten bereit.

Die sichere Aufbewahrung selbst ist eure Aufgabe als Betreiber. Zwei Dinge gehören
ins Archiv.

**DSFinV-K-Export je Kassensitzung (das Wichtigste).** Das ist die vom Finanzamt
erwartete Standardform eurer Kassendaten, lesbar mit jeder Tabellenkalkulation. So
sichert ihr ihn nach jedem Veranstaltungstag:

1. Im Admin-Bereich unter „Berichte & Export" die abgeschlossene Kassensitzung
   auswählen.
2. Auf „Archiv herunterladen (ZIP)" klicken. jotti lädt eine ZIP-Datei in den
   Download-Ordner.
3. Prüfen: die ZIP-Datei öffnen. Darin liegen mehrere CSV-Dateien, eine `index.xml`
   und eine `.dtd`. Lässt sie sich öffnen und sind die CSV-Dateien nicht leer, ist
   der Export in Ordnung.
4. An zwei getrennte Orte kopieren, etwa auf einen USB-Stick und zusätzlich in einen
   Cloud-Speicher. Ein einzelner Speicherort genügt nicht.

Maßgeblich ist der Export, den ihr direkt nach dem Veranstaltungstag archiviert.
Ein späterer Export nimmt die dann aktuellen Betreiberdaten, Tischnamen und die
jotti-Version.

Ändert ihr Betreiberdaten oder Tischnamen, exportiert und archiviert vorher alle
abgeschlossenen Kassensitzungen.

**Datenbank-Backup als Sicherheitsnetz.** Es enthält das vollständige Kassenjournal
im Rohformat samt TSE-Signaturen und Stammdaten.

Auf dem Kassenrechner zieht jotti vor jedem Update automatisch ein Backup. Geht ein
Update schief, stellt `jotti-restore.cmd` (Doppelklick) das letzte zurück.

Diese Backups liegen auf demselben Rechner und gehen mit ihm verloren. Euer vom
Rechner unabhängiges Archiv ist deshalb der DSFinV-K-Export oben.

Für eine Kopie der Rohdaten außer Haus kopiert ihr den Ordner `C:\ProgramData\jotti\backups`
([Aktualisieren](aktualisieren.md#warum-eure-daten-erhalten-bleiben)) auf einen USB-Stick
oder in eine Cloud. Ein weiteres Backup auf Wunsch zeigt die `KURZANLEITUNG.md` im ZIP.

Wer die Rohdaten laufend automatisch außer Haus sichern will, betreibt jotti auf
einem Server (siehe [Backups](aktualisieren-backups.md#backups) im Experten-Weg).

Ebenfalls aufbewahren: die Z-Bons (Tagesabschlüsse, im DSFinV-K-Export enthalten)
und die Zählprotokolle vom Kassensturz. Sorgt dafür, dass nur berechtigte Personen
Zugriff auf Rechner und Daten haben.

**TSE-Export vor Stilllegung oder Vertragsende.** jotti speichert die Signaturen,
aber nicht die vollständigen Protokolle der TSE. Auch diese müsst ihr 10 Jahre
aufbewahren (§ 146a Abs. 1 Satz 4 AO).

Lasst euch deshalb vor der Stilllegung der TSS und vor dem Ende des
fiskaly-Vertrags den TSE-Export (TAR-Datei) bei fiskaly erstellen. Legt ihn wie
den DSFinV-K-Export an zwei getrennten Orten ab.
