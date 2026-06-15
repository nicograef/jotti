# PRD: Finanzamt-Seite

> Scope: Diese PRD legt die Admin-Seite „Finanzamt" als gemeinsamen Ort für alle Belange gegenüber dem Finanzamt an und zieht die bereits vorhandene fiskalische Oberfläche dorthin um. Die Einstellungen-Seite, die heute nur diese fiskalischen Sektionen enthält, entfällt. Kein neues Backend-Kontext, keine neuen Endpunkte, keine Prüfungsbereitschafts- oder TSE-Readiness-Logik. Die schweren Folge-Features (DSFinV-K F-04, ELSTER F-05, GoBD-Integritätsnachweis F-08, Archiv-Bundle F-10, Verfahrensdokumentation F-11) behalten je ihre eigene PRD und ergänzen ihre Sektion erst, wenn das Feature steht.
> Quellen: [anforderungen.md §6](../anforderungen.md), [compliance.md](../compliance.md), Artikel „Kassen-Nachschau in Deutschland".

## Problem Statement

jotti unterliegt als elektronisches Aufzeichnungssystem der KassenSichV-Pflicht. Bei einer Kassen-Nachschau (§ 146b AO) erscheint ein Prüfer unangekündigt während der Veranstaltung, darf Daten aus dem Kassensystem auslesen und Unterlagen einsehen. Der Verein muss dann sofort die geforderten Angaben und Nachweise vorzeigen können. Zielgruppe sind nicht-technische Vereinsvorstände und Helfer.

Heute sind die fiskalischen Belange über mehrere Admin-Seiten verstreut: die TSE-Warnungen auf dem Reporting-Dashboard, die Betreiber-Stammdaten, die Kassenidentität, die TSE-Konfiguration und die Nachsignier-Liste in den Einstellungen. Keine Seite ist „die Kasse gegenüber dem Finanzamt". Wenn der Prüfer klingelt, muss der Vorstand zwischen Seiten suchen, statt einen Ort zu öffnen.

## Solution

Eine neue Admin-Seite „Finanzamt" unter `/admin/finanzamt` wird der eine Ort für alles, was das Finanzamt betrifft. Statt eine leere Schale mit Platzhaltern zu bauen, zieht diese PRD die bereits funktionierenden Sektionen dorthin um:

- Betreiber-Stammdaten samt Bearbeitung (Vereinsname, Adresse, Steuernummer, USt-ID).
- Kassenidentität read-only (Seriennummer, Anlegedatum), mit Kopier-Funktion für die ELSTER-Meldung.
- TSE-Ausfalldokumentation: die Liste der TSE-Nachsignier-Vorgänge samt ihrer Recovery-Aktionen (wieder einreihen, verwerfen). Sie zieht vollständig um, kein read-only Spiegel.
- Eine faktische TSE-Anbindungs-Zeile (konfiguriert ja/nein, Umgebung) mit Link zur TSE-Einrichtung.
- Eine kompakte Sektion „Dokumente und Pflichten" mit Klartext-Hinweis auf die 10-Jahres-Aufbewahrung und Links zu den vorhandenen Betreiber- und Compliance-Unterlagen.

Die Einstellungen-Seite enthält heute ausschließlich diese fiskalischen Sektionen. Nachdem die TSE-Konfiguration auf die TSE-Einrichtung umgezogen ist (TSE-Wizard-PRD) und die übrigen Sektionen auf die Finanzamt-Seite, entfällt die Einstellungen-Seite samt Sidebar-Eintrag.

Das Reporting-Dashboard behält nur noch einen kompakten Warn-Banner (TSE nicht konfiguriert oder offene Nachsignierungen) mit Link auf die Finanzamt-Seite.

Es entsteht kein neues Backend-Kontext und kein neuer Endpunkt. Die Seite komponiert die bestehenden Endpunkte. Jedes schwere Folge-Feature ergänzt seine Sektion in seiner eigenen PRD, wenn es steht; bis dahin zeigt die Seite nur funktionierende Inhalte.

## User Stories

### Vereins-Admin

1. Als Vereins-Admin möchte ich eine eigene Seite „Finanzamt" über die Seitenleiste erreichen, auf der alle Angaben und Nachweise gegenüber dem Finanzamt an einem Ort liegen, damit ich bei einer Kassen-Nachschau nicht zwischen Seiten suchen muss.
2. Als Vereins-Admin möchte ich die Betreiber-Stammdaten auf der Finanzamt-Seite einsehen und bearbeiten, damit Name, Anschrift und Steuernummer für Belege und Meldungen stimmen.
3. Als Vereins-Admin möchte ich die Kassen-Seriennummer auf der Finanzamt-Seite sehen und kopieren können, damit ich sie für die ELSTER-Meldung griffbereit habe.
4. Als Vereins-Admin möchte ich die TSE-Ausfalldokumentation (welche Vorgänge auf Nachsignierung warten oder nachsigniert wurden) auf der Finanzamt-Seite einsehen und die Vorgänge dort auch wieder einreihen oder verwerfen können, damit ich Ausfallzeiten belegen und bereinigen kann, ohne den Ort zu wechseln.
5. Als Vereins-Admin möchte ich von der Finanzamt-Seite zur TSE-Einrichtung gelangen, damit ich die TSE-Anbindung von hier aus einrichten oder ändern kann.
6. Als Vereins-Admin möchte ich auf der Finanzamt-Seite einen Klartext-Hinweis zur 10-Jahres-Aufbewahrung und Links zu den Betreiber- und Compliance-Unterlagen finden, damit ich einem Prüfer die geforderten Hinweise schnell vorlegen kann.
7. Als Vereins-Admin möchte ich auf dem Reporting-Dashboard weiterhin eine knappe Warnung sehen, wenn die TSE nicht konfiguriert ist oder Nachsignierungen offen sind, mit Link auf die Finanzamt-Seite, damit mir das Problem auch im Tagesgeschäft auffällt.
8. Als Vereins-Admin möchte ich die Finanzamt-Seite auf dem Smartphone bedienen können, damit ich auch während der Veranstaltung darauf zugreifen kann.

### Betriebsprüfer

9. Als Betriebsprüfer möchte ich, dass der Verein mir die Betreiber-Stammdaten, die Kassenidentität und die TSE-Ausfalldokumentation an einer Stelle vorzeigen kann, damit die Nachschau zügig abläuft.

## Implementation Decisions

### Frontend: Seite „Finanzamt"

- Neue Route `/admin/finanzamt`, lazy geladen wie die übrigen Admin-Seiten, hinter dem `AdminGuard`. Mobile-first, Karten stapeln auf kleinen Bildschirmen.
- Neuer Eintrag „Finanzamt" in der Seitenleiste in der Gruppe „Verwaltung" (Icon z. B. Landmark). Der Eintrag „Einstellungen" wird entfernt.
- Die Seite komponiert die bestehenden Sektionen. Die Komponenten und Hooks aus der heutigen Einstellungen-Seite ziehen unverändert um und werden wiederverwendet: `BetreiberSection`/`BetreiberForm`, `KassenidentitaetSection`, `TSENachsignierSection`/`NachsignierAuftragRow` samt `useBetreiber`, `useKassenidentitaet`, `useTSENachsignierAuftraege`, `useTSEStatus`.
- TSE-Anbindung: eine faktische Status-Zeile (konfiguriert ja/nein, Umgebung) aus dem bestehenden `get-tse-status`, mit Link „Einrichten oder ändern" auf `/admin/tse-einrichtung`. Das ist Navigation und faktische Anzeige, ausdrücklich keine Bewertung der Prüfungsbereitschaft. Damit übernimmt die Finanzamt-Seite den Einstiegspunkt zur TSE-Einrichtung, den die TSE-Wizard-PRD ursprünglich in der Einstellungen-Sektion vorgesehen hatte.
- Kassenidentität: Seriennummer und das vorhandene Anlegedatum (`AngelegtAm`) werden als solche beschriftet. Das Anlegedatum wird nicht als das ELSTER-Inbetriebnahmedatum ausgegeben; ein eigenes, rechtlich gemeintes Inbetriebnahmedatum ist Sache von F-05.

### Änderung am Reporting-Dashboard

- Die beiden ausführlichen TSE-Warnungsblöcke werden durch einen kompakten Banner ersetzt, der bei nicht konfigurierter TSE oder offenen Nachsignierungen erscheint und auf die Finanzamt-Seite verlinkt. Die Auslöse-Bedingungen bleiben unverändert (`get-tse-status`); nur die Darstellung verkürzt sich und der Detailort verschiebt sich auf die Finanzamt-Seite.

### Wegfall der Einstellungen-Seite

- Die Einstellungen-Seite enthält heute nur Betreiber-Stammdaten, Kassenidentität, TSE-Konfiguration und die Nachsignier-Liste. Betreiber, Kassenidentität und Nachsignier-Liste ziehen auf die Finanzamt-Seite, die TSE-Konfiguration auf die TSE-Einrichtung. Danach werden die Einstellungen-Route und der Sidebar-Eintrag entfernt.

### Kein Backend-Umbau

- Reine Frontend-Umstrukturierung auf bestehenden Endpunkten: `get-betreiber`, `update-betreiber`, `get-kassenidentitaet`, `get-tse-nachsignier-auftraege`, `tse-nachsignier-auftrag-zuruecksetzen`, `tse-nachsignier-auftrag-verwerfen`, `get-tse-status`. Es werden keine Endpunkte, kein Kontext und kein Schema angefasst.
- Begriffe für `docs/language.md`: Finanzamt-Seite, Ausfalldokumentation, Kassen-Nachschau.

### Abgrenzung und Koordination der fiskal-nahen Seiten

- TSE-Einrichtung (TSE-Wizard-PRD): den TSE-Lebenszyklus einrichten (Wizard) und die Verbindung testen. Der Einstiegspunkt liegt nach diesem Umzug auf der Finanzamt-Seite statt in den Einstellungen. Das ersetzt die im TSE-Wizard-Plan vorgesehene „Status plus Link"-Sektion in den Einstellungen und behält die Entscheidung „kein eigener Sidebar-Eintrag" bei, da die Finanzamt-Seite dorthin verlinkt.
- Finanzamt (diese PRD): alles gegenüber dem Finanzamt, fiskalische Stammdaten lesen und bearbeiten, Nachweise vorzeigen.
- Einstellungen: entfällt.
- Reihenfolge: Der Umzug der TSE-Konfiguration (TSE-Wizard, Phase 2) und dieser Umzug müssen gemeinsam landen, damit die Einstellungen-Seite nicht halb leer zurückbleibt und der Einstiegspunkt zur TSE-Einrichtung nie verwaist ist.

## Testing Decisions

- Es entsteht keine neue Geschäftslogik, daher keine neuen Backend-Tests. Die bestehenden Endpunkt- und Query-Tests bleiben gültig.
- Die Seite ist dünne Komposition bereits funktionierender, anderswo getesteter Sektionen; nach Konvention der Codebase keine eigenen Frontend-Komponententests.
- Manuell zu verifizieren: jede umgezogene Sektion funktioniert auf der neuen Seite (Betreiber laden und speichern, Kassenidentität anzeigen und kopieren, Nachsignier-Vorgänge anzeigen, wieder einreihen und verwerfen); die Einstellungen-Route ist entfernt; die Seitenleiste zeigt „Finanzamt"; der Dashboard-Banner verlinkt korrekt; die Seite ist ab 360 px bedienbar.

## Out of Scope

- DSFinV-K-Export (F-04), Archiv-Bundle (F-10), GoBD-Integritätsnachweis (F-08), ELSTER-Meldung und Meldestatus (F-05), Verfahrensdokumentation (F-11): je eigene PRD. Jedes Feature ergänzt seine Sektion auf dieser Seite, wenn es steht. Keine Platzhalter-Sektionen jetzt.
- TSE-Setup-Wizard und die TSE-Konfigurations-Oberfläche: TSE-Wizard-PRD.
- TSE-Datenandruck auf dem Beleg und QR-Code (Rest von F-03, F-09).
- Jede Prüfungsbereitschafts- oder TSE-Readiness-Bewertung: bewusst nicht gebaut. Die Seite zeigt faktische Angaben und Nachweise, kein Ampel-Urteil.
- Ein neues Backend-Kontext `fiskal`: nicht nötig. Falls ein Folge-Feature eine echte lesende Aggregation über mehrere Kontexte braucht, wird das in dessen PRD entschieden.

## Further Notes

- Leitidee bleibt die Kassen-Nachschau nach § 146b AO: die Finanzamt-Seite ist der eine Ort, den der Vorstand öffnet, wenn das Finanzamt klingelt. Über die Phasen wächst sie um die schweren Nachweise (DSFinV-K, Archiv, Integrität, ELSTER).
- Erweiterungsvertrag für Folge-PRDs: Jedes Feature ergänzt seine Sektion auf dieser Seite und, falls nötig, seinen eigenen Endpunkt. Es gibt keine zentrale Registry und keine geteilte Query, die mitgepflegt werden müsste.
- ELSTER-Meldestatus (für die F-05-PRD): als Value Object im bestehenden `settings`-Kontext, da compliance.md §7.4 den Meldestatus in den Stammdaten verortet. Wird auf dieser Seite angezeigt, sobald F-05 steht.
- Inbetriebnahmedatum: Für die ELSTER-Meldung ist das tatsächliche Inbetriebnahmedatum maßgeblich, nicht der technische Anlegezeitpunkt der Datenbankzeile. Ein editierbares Inbetriebnahmedatum ist Sache von F-05; bis dahin zeigt die Seite nur das Anlegedatum als technische Angabe.
- Dokumente und Pflichten: verlinkt werden die erreichbaren Unterlagen (Betreiber-Leitfaden, Compliance-Überblick); die Verfahrensdokumentation kommt hinzu, sobald F-11 steht. Wie die Repository-Dokumente aus der App erreichbar gemacht werden (In-App-Hilfe oder externer Link), ist ein Umsetzungsdetail.
