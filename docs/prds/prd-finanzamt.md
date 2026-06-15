# PRD: Finanzamt-Seite (Fiskal-Cockpit)

> Scope: Seiten-Schale. Diese PRD baut die Admin-Seite „Finanzamt" als integrierendes Cockpit und das lesende Backend-Kontext `fiskal` mit der Prüfungsbereitschafts-Query. Die schweren Fiskal-Features bekommen jeweils ein eigenes PRD: F-04 (DSFinV-K-Export), F-05 (ELSTER-Meldung/Meldestatus), F-08 (GoBD-Integritätsnachweis), F-10 (Archiv-Bundle), F-11 (Verfahrensdokumentation). Diese Seite ist deren gemeinsamer Einstiegspunkt.
> Quellen: [anforderungen.md §6](../anforderungen.md), [compliance.md](../compliance.md), Artikel „Kassen-Nachschau in Deutschland".

## Problem Statement

jotti unterliegt als elektronisches Aufzeichnungssystem der KassenSichV-Pflicht. Bei einer Kassen-Nachschau (§ 146b AO) erscheint ein Prüfer unangekündigt während der Veranstaltung, darf Daten aus dem Kassensystem auslesen und Unterlagen einsehen. Der Verein muss dann sofort nachweisen können, dass die Kasse konform ist, und die angeforderten Artefakte erzeugen (Datenexport, Nachweise, Meldedaten).

Die Zielgruppe sind nicht-technische Vereinsvorstände und Helfer. Heute sind die fiskalischen Belange über drei Admin-Seiten verstreut: TSE-Warnungen auf dem Reporting-Dashboard, Betreiber-Stammdaten, Kassenidentität, TSE-Konfiguration und die Nachsignier-Liste in den Einstellungen, der Tagesabschluss in der Kassensitzung. Keine Seite beantwortet die Frage „Sind wir auf eine Prüfung vorbereitet?" und keine bündelt die prüfungsrelevanten Daten und Funktionen an einem Ort. Mehrere Pflicht-Funktionen fehlen zudem noch ganz (DSFinV-K-Export, ELSTER-Meldestatus, Integritätsnachweis, Archiv-Bundle) und hätten ohne eine solche Seite keinen natürlichen Platz.

## Solution

Eine neue Admin-Seite „Finanzamt" unter `/admin/finanzamt` bündelt alle fiskalischen Angaben, Nachweise und Funktionen als Cockpit für die Prüfungsbereitschaft. Sie ist kein Konfigurationsformular, sondern beantwortet zuerst „Bist du bereit, wenn das Finanzamt klingelt?" und liefert darunter die Artefakte, die ein Prüfer anfordert. Klartext statt Paragrafen, ein Ampel-Status pro Punkt, jede Zeile mit konkretem nächsten Schritt.

Diese PRD liefert die Schale und den tragenden Mechanismus:

- Eine Prüfungsbereitschafts-Übersicht, die aus den bereits vorhandenen Signalen einen Statusbericht ableitet (TSE konfiguriert und signierfähig, keine offenen Nachsignierungen, Betreiber-Stammdaten vollständig).
- Eine read-only Kassenidentitäts-Sektion (Seriennummer, Inbetriebnahmedatum), die die für die ELSTER-Meldung nötige Identität sichtbar macht und zum Bearbeiten in die Einstellungen verlinkt.
- Eine read-only Ausfalldokumentation, die die TSE-Nachsignier-Vorgänge als prüfungsrelevantes Protokoll zeigt. Die Recovery-Aktionen (zurücksetzen, verwerfen) bleiben in den Einstellungen.
- Platzhalter-Sektionen für die noch offenen Features (DSFinV-K-Export, Archiv-Bundle, Integritätsnachweis, ELSTER-Meldung), die ihren Zweck erklären und als „in Vorbereitung" markiert sind.
- Eine Dokumenten-Sektion mit Links zu Verfahrensdokumentation, Betreiber-Leitfaden und Compliance-Überblick.

Das Reporting-Dashboard zeigt die TSE-Warnungen künftig nur noch als kompakten Banner mit Link auf die Finanzamt-Seite, die der kanonische Ort für Status und Behebung wird.

Tragend ist ein neues, rein lesendes Backend-Kontext `fiskal` mit einer Prüfungsbereitschafts-Query. Sie ist bewusst erweiterbar: Jedes Folge-PRD ergänzt seinen eigenen Bereitschafts-Punkt und füllt seine Sektion. Das Kontext hält keine eigene Datenhoheit, sondern liest aus den bestehenden Kontexten.

## User Stories

### Vereins-Admin: Prüfungsbereitschaft

1. Als Vereins-Admin möchte ich auf einer eigenen Seite „Finanzamt" auf einen Blick sehen, ob meine Kasse prüfungsbereit ist, damit ich bei einer unangekündigten Kassen-Nachschau ruhig bleiben kann.
2. Als Vereins-Admin möchte ich pro Bereitschafts-Punkt einen klaren Ampel-Status (in Ordnung, Warnung, offen) sehen, damit ich Probleme sofort erkenne, ohne Fachbegriffe deuten zu müssen.
3. Als Vereins-Admin möchte ich zu jedem Punkt eine kurze Erklärung in Klartext und einen konkreten nächsten Schritt sehen, damit ich weiß, was zu tun ist.
4. Als Vereins-Admin möchte ich von einem Bereitschafts-Punkt direkt zur Stelle springen können, an der ich ihn behebe (z. B. Einstellungen für die Stammdaten), damit ich nicht suchen muss.
5. Als Vereins-Admin möchte ich, dass die Übersicht erkennt, wenn die TSE nicht konfiguriert oder nicht signierfähig ist, damit ich die Kasse nicht versehentlich ohne fiskalische Absicherung betreibe.
6. Als Vereins-Admin möchte ich gewarnt werden, wenn TSE-Nachsignierungen offen sind, damit ich weiß, dass noch Vorgänge auf ihre Signatur warten.
7. Als Vereins-Admin möchte ich erkennen, wenn meine Betreiber-Stammdaten unvollständig sind, damit Belege und Meldedaten korrekt sind.

### Vereins-Admin: Identität und Meldedaten

8. Als Vereins-Admin möchte ich die Kassen-Seriennummer und das Inbetriebnahmedatum auf der Finanzamt-Seite sehen und kopieren können, damit ich sie für die ELSTER-Meldung griffbereit habe.
9. Als Vereins-Admin möchte ich von der Kassenidentität aus zu den Einstellungen gelangen, falls etwas zu ändern ist, damit die Bearbeitung an einem Ort bleibt.
10. Als Vereins-Admin möchte ich auf der Seite einen klar markierten Bereich für die ELSTER-Meldung sehen, auch solange er noch in Vorbereitung ist, damit ich weiß, dass diese Pflicht existiert und wo sie künftig erledigt wird.

### Vereins-Admin: Daten für die Prüfung

11. Als Vereins-Admin möchte ich auf der Seite einen Bereich „Daten für die Prüfung" sehen, in dem später der DSFinV-K-Export und das Archiv-Bundle bereitstehen, damit ich bei einer Prüfung weiß, wo ich die geforderten Dateien erzeuge.
12. Als Vereins-Admin möchte ich, dass noch nicht verfügbare Export-Funktionen klar als „in Vorbereitung" gekennzeichnet sind, damit ich nicht auf eine Funktion warte, die noch nicht existiert.

### Vereins-Admin: Nachweise und Dokumente

13. Als Vereins-Admin möchte ich die TSE-Ausfalldokumentation (welche Vorgänge wann nachsigniert wurden) als Protokoll einsehen, damit ich Ausfallzeiten gegenüber einem Prüfer belegen kann.
14. Als Vereins-Admin möchte ich, dass die Ausfalldokumentation auf der Finanzamt-Seite read-only ist, damit ich sie nicht versehentlich verändere; die Verwaltung bleibt im TSE-Bereich der Einstellungen.
15. Als Vereins-Admin möchte ich Links zur Verfahrensdokumentation, zum Betreiber-Leitfaden und zum Compliance-Überblick an einem Ort finden, damit ich einem Prüfer die geforderten Unterlagen schnell vorlegen kann.

### Betriebsprüfer

16. Als Betriebsprüfer möchte ich, dass der Verein mir die Kassenidentität, die TSE-Ausfalldokumentation und die geforderten Datenexporte an einer Stelle vorzeigen kann, damit die Nachschau zügig abläuft.

### Querschnitt und Navigation

17. Als Admin möchte ich die Finanzamt-Seite über einen eigenen Eintrag in der Seitenleiste erreichen, damit sie als eigenständiger Bereich erkennbar ist.
18. Als Admin möchte ich die Finanzamt-Seite auf dem Smartphone bedienen können, damit ich auch während der Veranstaltung darauf zugreifen kann.
19. Als Admin möchte ich auf dem Reporting-Dashboard weiterhin eine knappe Warnung sehen, wenn die TSE nicht einsatzbereit ist, mit einem Link zur Finanzamt-Seite für die Details, damit mir das Problem auch im Tagesgeschäft auffällt.
20. Als Entwickler möchte ich, dass die Prüfungsbereitschafts-Query so aufgebaut ist, dass jedes Folge-PRD seinen eigenen Bereitschafts-Punkt ergänzen kann, ohne die bestehende Logik umzubauen, damit die Seite mit den Features mitwächst.

## Implementation Decisions

### Neues Backend-Kontext `fiskal`

- Ein neues, rein lesendes Bounded Context `fiskal`. Es hält keine eigene Datenhoheit und schreibt nichts zurück; es liest aus den bestehenden Kontexten `settings` und `tse` (später zusätzlich `kasse`, `steuer`, `product` für die Export-Features). Rolle im DDD-Sinn: Customer/Conformist gegenüber diesen Kontexten, später Open-Host-Service für den DSFinV-K-Export.
- Namensabgrenzung: `fiskal` (das Kontext) ist zu unterscheiden vom Vendor `fiskaly` (TSE-Anbieter, erscheint im Code nur als Eigenname, z. B. in der TSE-Repository-Schicht). Die neuen Begriffe (Finanzamt, Fiskal, Prüfungsbereitschaft, Ausfalldokumentation, Kassen-Nachschau) werden in `docs/language.md` ergänzt.

### Prüfungsbereitschafts-Query (Deep Module)

- Eine Query liefert einen strukturierten Prüfungsbereitschafts-Report: eine Liste von Punkten, jeder mit einem stabilen Schlüssel, einem Status (in Ordnung, Warnung, offen) und optionalen Detailangaben (z. B. Anzahl offener Nachsignierungen).
- Die Domänenlogik (Ableitung der Punkte aus den Repo-Zuständen) liegt in der Domänenschicht des `fiskal`-Kontexts und liest über die bestehenden Repository-Schnittstellen. Sie ist isoliert testbar.
- Anfangs enthaltene Punkte:
  - TSE einsatzbereit: abgeleitet aus der TSE-Konfiguration (konfiguriert ja/nein). Status offen, wenn nicht konfiguriert.
  - Nachsignierungen: Status Warnung, wenn offene Nachsignierungen vorhanden sind, sonst in Ordnung; Anzahl als Detail.
  - Betreiber-Stammdaten vollständig: Pflichtfelder Vereinsname, Straße, PLZ, Ort gesetzt. Status offen, wenn unvollständig.
- Die Report-Struktur ist erweiterbar: Folge-PRDs fügen weitere Punkte hinzu (ELSTER-Meldestatus, Integritätsergebnis, Verfahrensdokumentation), ohne bestehende Punkte zu verändern.
- Endpunkt: ein POST-only Admin-Endpunkt, der den Report zurückgibt (Konvention der Codebase: flacher Verb-Stil unter `/admin/*`, z. B. `get-fiskal-status`). Antwort über ein Response-DTO mit `json`-Tags in der HTTP-Schicht; die Domänen-Structs werden nie direkt serialisiert.

### Frontend: Seite „Finanzamt"

- Neue Seite unter `/admin/finanzamt`, lazy geladen wie die übrigen Admin-Seiten, hinter dem `AdminGuard`.
- Neuer Eintrag „Finanzamt" in der Seitenleiste in der Gruppe „Verwaltung" mit passendem Icon (z. B. Landmark oder ShieldCheck).
- Die Seite komponiert Sektionen, mobile-first, Karten stapeln auf kleinen Bildschirmen:
  - Prüfungsbereitschaft: rendert den Report der Query; pro Punkt Status, Klartext-Erklärung und ggf. Deep-Link.
  - Kassenidentität: read-only Anzeige von Seriennummer und Inbetriebnahmedatum (über den bestehenden Endpunkt), mit Kopier-Funktion und Link zu den Einstellungen.
  - ELSTER-Meldung: Platzhalter-Sektion mit Zweckerklärung, als „in Vorbereitung" markiert (Detailumsetzung in der F-05-PRD).
  - Daten für die Prüfung: Platzhalter-Sektionen für DSFinV-K-Export (F-04) und Archiv-Bundle (F-10), als „in Vorbereitung" markiert.
  - Integritätsnachweis: Platzhalter-Sektion (F-08), als „in Vorbereitung" markiert.
  - Ausfalldokumentation: read-only Liste der TSE-Nachsignier-Vorgänge über den bestehenden Endpunkt; keine Aktions-Buttons.
  - Dokumente und Pflichten: Links zu Verfahrensdokumentation, Betreiber-Leitfaden und Compliance-Überblick; Klartext-Hinweis auf die 10-Jahres-Aufbewahrung als Betreiberpflicht.
- Eine neue Backend-Klasse kapselt den Aufruf der Prüfungsbereitschafts-Query über das bestehende `BackendClient`-Interface; ein Hook stellt den Status der Seite bereit. Frontend-Validierung der Antwort mit Zod.
- Die read-only Sektionen wiederverwenden die bestehenden Endpunkte für Kassenidentität, Betreiber und Nachsignier-Aufträge; es werden dafür keine neuen Endpunkte gebaut.

### Änderung am Reporting-Dashboard

- Die beiden ausführlichen TSE-Warnungsblöcke auf dem Reporting-Dashboard werden durch einen kompakten Banner ersetzt, der bei nicht konfigurierter TSE oder offenen Nachsignierungen erscheint und auf die Finanzamt-Seite verlinkt. Die Auslöse-Bedingungen bleiben unverändert; nur die Darstellung wird verkürzt und der Detailort verschiebt sich auf die Finanzamt-Seite.

### Abgrenzung der drei fiskal-nahen Seiten

- Einstellungen: Stammdaten und TSE-Zugangsdaten eingeben (Write).
- TSE-Einrichtung (eigene PRD, geplant): den TSE-Lebenszyklus einrichten (Wizard).
- Finanzamt (diese PRD): Prüfungsbereitschaft nachweisen, Identität und Meldedaten anzeigen, Artefakte erzeugen (read und produce).

## Testing Decisions

- Ein guter Test prüft beobachtbares Verhalten, nicht die innere Umsetzung. Für die Prüfungsbereitschafts-Query heißt das: gegebene Repo-Zustände führen zu erwarteten Bereitschafts-Punkten und -Status.
- Getestet wird die Prüfungsbereitschafts-Query (das Deep Module): unter anderem TSE konfiguriert gegenüber nicht konfiguriert, offene Nachsignierungen vorhanden gegenüber keine, Betreiber-Stammdaten vollständig gegenüber unvollständig. Geprüft wird, dass die zurückgegebenen Punkte den jeweils korrekten Status und die korrekten Detailangaben tragen.
- Vorbild für diese Tests sind die bestehenden Query- und Application-Tests im Backend (Settings- und Kasse-Kontext) mit in-memory beziehungsweise gefälschten Repositories.
- Die Frontend-Schale ist dünne Komposition und erhält in dieser PRD keine eigenen Tests; Tests zu den Sektionen entstehen mit den jeweiligen Feature-PRDs.

## Out of Scope

- DSFinV-K-Export (F-04): eigene PRD. Hier nur eine als „in Vorbereitung" markierte Platzhalter-Sektion.
- Archiv-Bundle / 10-Jahres-Archivierung (F-10): eigene PRD. Hier nur Platzhalter.
- GoBD-Integritätsnachweis (F-08): eigene PRD. Hier nur Platzhalter.
- ELSTER-Meldung und Meldestatus (F-05): eigene PRD. Hier nur Platzhalter und Anzeige der bereits vorhandenen Identitätsdaten. Die programmatische Übermittlung (ERiC, fiskaly-Submission) ist ebenfalls dort verortet.
- Verfahrensdokumentation (F-11): das Erstellen der Muster-Verfahrensdokumentation ist eine eigene Aufgabe; hier wird nur darauf verlinkt.
- TSE-Setup-Wizard: eigene PRD (TSE-Einrichtung).
- TSE-Datenandruck auf dem Beleg und QR-Code (Rest von F-03, F-09).
- Das Bearbeiten von Betreiber-Stammdaten und TSE-Konfiguration bleibt in den Einstellungen.
- Die Recovery-Aktionen für Nachsignier-Aufträge (zurücksetzen, verwerfen) bleiben in den Einstellungen.

## Further Notes

- Leitidee der Seite ist die Prüfungsbereitschaft für die Kassen-Nachschau nach § 146b AO. Der Artikel und compliance.md beschreiben, was ein Prüfer verlangt (manipulationssichere TSE, DSFinV-K-Datenexport, lückenlose und unveränderbare Aufzeichnung, Belege mit TSE-Daten, Meldung beim Finanzamt, Verfahrensdokumentation, 10-Jahres-Aufbewahrung, stimmiger Kassenbestand). Die Sektionen der Seite bilden diese Checkliste ab.
- Erweiterungs-Vertrag für Folge-PRDs: Jedes Feature ergänzt (a) seinen Punkt in der Prüfungsbereitschafts-Query und (b) füllt seine Platzhalter-Sektion mit der echten Funktion. So bleibt die Seite über die Phasen hinweg konsistent.
- Vorgesehene Verortung des ELSTER-Meldestatus (für die F-05-PRD): als Value Object im bestehenden `settings`-Kontext, da compliance.md §7.4 den Meldestatus ausdrücklich in den Stammdaten verortet. Damit bleibt `fiskal` auch nach F-05 rein lesend.
- Abrechnungskreis (F-06) ist teilweise umgesetzt und speist später den DSFinV-K-Export; die Steuer-Aufteilung inklusive Kombi-Splitting existiert bereits im `steuer`-Kontext.
