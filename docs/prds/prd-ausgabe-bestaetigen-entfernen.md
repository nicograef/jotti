# PRD: Ausgabe-Bestätigung entfernen

> Quelle: Praxistest eines Vereins am 09.07.2026. Rückmeldung der
> Servicekräfte: Die Ausgabe-Bestätigung stört im Betrieb mehr, als sie
> hilft. Alle technischen Aussagen in diesem PRD sind im Code
> verifiziert. Entscheidung dokumentiert in
> [ADR 01](../adrs/01_ausgabe-bestaetigen.md).

## Problem Statement

Beim ersten Praxiseinsatz hat sich die Ausgabe-Bestätigung (K-03) als
Hindernis erwiesen:

1. Servicekräfte müssen im stressigsten Moment des Abends zusätzliche
   Taps ausführen, um einen Status zu pflegen, von dem keine andere
   Funktion abhängt — nicht Kassieren, nicht Stornieren, nicht der
   Bondruck. Der Status ist rein informativ.
2. Real wird die Ausgabe über Arbeitsbons, im Kopf oder auf
   Vertrauensbasis verfolgt. Oft wird bei der Ausgabe direkt kassiert;
   die Bestätigung wäre dann eine Doppelerfassung derselben Handlung.
3. Die Ausgabe-UI belegt prominenten Platz im Bestellen-Tab der
   Tisch-Seite und verdrängt die eigentliche Kernfunktion
   (Bestellaufnahme), solange offene Positionen existieren.

Das widerspricht der Produktidentität von jotti: nur notwendige
Funktionen, diese aber technisch und in der UX hochwertig — jedes
zusätzliche Feature erhöht die Komplexität für ehrenamtliche Teams.

## Solution

jotti entfernt die Ausgabe-Bestätigung vollständig — UI, API, Event-Typ,
Projektion und Datenbestand. Der Bestellen-Tab zeigt nur noch die
Bestellaufnahme. „Offene Arbeit" bedeutet fortan „noch nicht kassiert":
Live-Reporting, Tischliste und Tisch-Badge zeigen den Zahlungsstand
statt des Ausgabestands. Die physische Ausgabe-Koordination übernehmen —
wie in der Praxis bewährt — die Arbeitsbons (K-12). Bestehende
Installationen werden beim Update automatisch bereinigt.

## User Stories

1. Als Servicekraft möchte ich im Bestellen-Tab ausschließlich die
   Bestellaufnahme sehen, damit ich unter Stress schnell bestellen kann,
   ohne einen irrelevanten Status pflegen zu müssen.
2. Als Servicekraft möchte ich am Tisch auf einen Blick den offenen
   Zahlbetrag sehen, damit ich weiß, ob an diesem Tisch noch kassiert
   werden muss.
3. Als Serviceleitung oder Admin möchte ich im Live-Reporting sehen, an
   welchen Tischen und bei welchen Servicekräften noch nicht kassiert
   wurde, damit ich gezielt nachfassen kann, bevor Gäste gehen.
4. Als Helfer an der Ausgabestation möchte ich mich weiterhin auf die
   Arbeitsbons verlassen, damit die Ausgabe ohne App-Interaktion
   koordiniert bleibt.
5. Als Betreiber einer bestehenden Installation möchte ich, dass das
   Update den Altdatenbestand automatisch bereinigt, damit die Anwendung
   danach ohne manuelle Eingriffe startet.

## Implementation Decisions

- **Vollständige Entfernung inklusive Lesepfade.** Der Event-Typ
  `ausgabe-bestaetigt:v1` verschwindet aus dem Event-Modell, der
  Projektion, der Historie, der fiskalischen Projektion und dem
  Event-JSON-Contract-Test. Die Event-Switches bleiben bewusst exklusiv
  (unbekannter Event-Typ → Fehler) — das Sicherheitsnetz gegen korrupte
  oder unerwartete Journaleinträge wird nicht aufgeweicht.
- **Datenbereinigung per Migration.** Eine neue forward-only Migration
  löscht zuerst die Tisch-Session-Projektionszeilen (sie referenzieren
  Events per Fremdschlüssel und werden beim Backend-Start ohnehin
  vollständig neu aufgebaut), dann die Alt-Events des Typs
  `ausgabe-bestaetigt:v1`, und entfernt anschließend die Spalte der
  ausstehenden Positionen aus der Projektionstabelle. Das Kassenjournal
  ist zusätzlich auf DB-Ebene schreibgeschützt: `BEFORE`-Trigger
  blockieren UPDATE/DELETE/TRUNCATE für alle Rollen, auch den
  Table-Owner. Die Migration deaktiviert den Delete-Trigger deshalb
  innerhalb ihrer Transaktion und aktiviert ihn vor dem COMMIT wieder;
  sie läuft als Schema-Owner (jotti-migrate-Container), der `ALTER
  TABLE` ausführen darf und von den REVOKE-Grants nicht eingeschränkt
  wird. Der Schutz besteht nach der Migration unverändert fort. Die
  bewusste, einmalige Ausnahme vom Append-only-Prinzip des
  Kassenjournals ist
  vertretbar, weil das Projekt vor v1.0.0 steht, die einzige bekannte
  Instanz im Testbetrieb ohne TSE läuft und der Event-Typ nicht
  signaturpflichtig ist (garantiert keine TSE-Signatur-Referenzen);
  Begründung im Detail in ADR 01.
- **Offene Arbeit = unbezahlt.** Die Tisch-Session-Projektion verliert
  das Konzept „ausstehende Positionen". Die Kennzahl „offene Arbeit"
  (Live-Reporting pro Servicekraft und Tisch) und das Signal „für mich
  erledigt" in der Tischliste basieren nur noch auf unbezahlten
  Positionen.
- **API ersatzlos.** Der Endpunkt zum Bestätigen der Ausgabe, sein
  Fehlercode und die Ausgabe-/Ausstehend-Felder in den Antworten von
  Tisch-State und Historie entfallen. Keine API-Versionierung nötig —
  Frontend und Backend werden im selben Release ausgetauscht.
- **UI.** Der Bestellen-Tab enthält nur noch die Bestellkomponente. Das
  Badge im Tisch-Header wechselt von „X offen / Alles ausgegeben!" auf
  den Zahlungsstand (offener Saldo bzw. „Alles bezahlt"). Die
  Tisch-Historie kennt die Eintragsart „Ausgabe" nicht mehr.
- **Dokumentation.** K-03 entfällt aus dem Funktionsumfang; K-13
  (Küchendisplay) und K-15 (Zubereitungsstatus) werden von der Roadmap
  gestrichen, da sie auf demselben — im Praxistest verworfenen —
  Ausgabe-Tracking aufbauen. Handbuch (Event-Tabelle, Invarianten,
  Rollenmatrix, UI-Beschreibung), Glossar, Produktbeschreibung und
  README werden bereinigt. Mit ADR 01 wird das Verzeichnis `docs/adrs/`
  als Ort für Architektur-Entscheidungen etabliert.
- **Seed/Demo.** Die Seed-Szenarien erzeugen keine Ausgabe-Zustände
  mehr.

## Testing Decisions

- Gute Tests prüfen ausschließlich externes Verhalten (Projektions- und
  API-Ergebnisse, sichtbares UI-Verhalten), keine
  Implementierungsdetails.
- **Neue Tests:**
  - Unit-Tests für die Neudefinition der offenen Arbeit: Rollup und
    „für mich erledigt" ergeben sich allein aus unbezahlten Positionen.
  - Integrationstest für die Migration: Ein Journal mit
    Ausgabe-Alt-Events wird bereinigt, und der anschließende
    Projektions-Rebuild läuft fehlerfrei mit korrekten Endzuständen.
- **Anzupassende Tests:** Event-Contract-Test (Typ entfällt aus der
  Pinning-Liste), Replay-Fuzz, Command-/Handler-Tests,
  Frontend-Komponenten- und Seiten-Tests, E2E-Service-Ablauf ohne
  Ausgabe-Schritt.
- **Prior Art:** table-driven Domain-Tests der Kasse, Integrationstests
  des Kassenjournal-Repositories, bestehende E2E-Service-Flows.

## Out of Scope

- Kein Ersatzfeature: kein Küchendisplay, kein Zubereitungsstatus, kein
  optionales/konfigurierbares Ausgabe-Tracking.
- Keine Änderungen an Arbeitsbons oder am Drucksystem.
- Keine Änderungen an Kassieren, Stornieren oder Umbuchen über das
  Entfernen der Ausgabe-Bezüge hinaus.
- Keine Migrationswerkzeuge für unbekannte Fremdinstallationen; die
  Release-Migration deckt den bekannten Bestand ab.

## Further Notes

Die Bereinigung läuft vollständig über die Release-Migration im
normalen Update-Ablauf (Backup → Migrate-Container → App-Start); ein
manueller Eingriff ist nicht vorgesehen. Die Lücken in den
Versionsnummern je Subject, die die Event-Löschung hinterlässt, sind
unschädlich — die optimistische Nebenläufigkeitskontrolle setzt nur auf
der Maximal-Version je Subject auf, und der Startup-Rebuild erzeugt die
Projektion aus den verbleibenden Events.
