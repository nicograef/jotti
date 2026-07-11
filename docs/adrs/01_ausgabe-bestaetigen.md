# ADR 01: Ausgabe-Bestätigung entfernen

- **Status:** akzeptiert (2026-07-10)
- **Kontext-Dokumente:**
  PRD `prd-ausgabe-bestaetigen-entfernen.md` (nach Merge gelöscht,
  siehe Git-Historie), `docs/produktbeschreibung.md`

## Kontext

jotti bot seit den ersten Versionen die Funktion „Ausgabe bestätigen"
(K-03): Servicekräfte markieren bestellte Positionen als an den Gast
übergeben (Event `ausgabe-bestaetigt:v1`). Der Praxistest bei einem
Verein am 09.07.2026 hat gezeigt, dass die Funktion im realen Betrieb
stört statt hilft:

- Der Ausgabe-Status ist rein informativ. Keine andere Funktion hängt
  davon ab — Stornierung, Kassieren, Umbuchen und Bondruck arbeiten
  unabhängig davon. Das Event ist nicht signaturpflichtig und erscheint
  nicht im DSFinV-K-Export.
- Real koordinieren die Teams die Ausgabe über Arbeitsbons (K-12), im
  Kopf oder auf Vertrauensbasis. Häufig wird bei der Ausgabe direkt
  kassiert — die Bestätigung wäre eine Doppelerfassung.
- Die Ausgabe-UI verdrängt im Bestellen-Tab die Kernfunktion
  (Bestellaufnahme), solange offene Positionen existieren.

jotti ist bewusst reduziert: nur notwendige Funktionen, diese aber
hochwertig. Ein Feature, das der erste Praxiseinsatz als Ballast
identifiziert, widerspricht dieser Identität.

Erwogene Alternativen:

1. **Nur UI entfernen, Backend behalten** — hinterlässt toten Code und
   einen API-Endpunkt ohne Nutzer; die Komplexität bleibt.
2. **Schreibpfad entfernen, Lesepfad als Legacy behalten** — hält den
   Event-Typ, den Contract-Test und die Projektion dauerhaft am Leben,
   nur um historische Testdaten lesen zu können.
3. **Optional per Einstellung** — Konfigurationskomplexität für ein
   Feature, das schon als Pflichtbestandteil niemand wollte.
4. **Vollständige Entfernung inklusive Datenbestand** — sauberster
   Schnitt, erfordert aber das Löschen bestehender Events.

## Entscheidung

Die Ausgabe-Bestätigung wird **vollständig entfernt** (Alternative 4):
UI, API-Endpunkt, Event-Typ, Projektion, Contract-Eintrag und
Altdatenbestand.

Die Event-Lesepfade (Projektion, Historie, fiskalische Projektion) sind
exklusiv implementiert — ein unbekannter Event-Typ ist ein Fehler, und
beim Backend-Start wird die Tisch-Session-Projektion vollständig aus dem
Journal neu aufgebaut. Verbleibende `ausgabe-bestaetigt:v1`-Events
würden den Start also verhindern. Deshalb löscht eine Release-Migration
die Alt-Events (und die Projektionsspalte der ausstehenden Positionen).
Da das Kassenjournal auch auf DB-Ebene append-only ist (`BEFORE`-Trigger
blockieren DELETE für alle Rollen, einschließlich des Table-Owners),
deaktiviert die Migration den Delete-Trigger innerhalb ihrer Transaktion
und aktiviert ihn vor dem COMMIT wieder; sie läuft als Schema-Owner und
kann das. Der DB-seitige Schreibschutz bleibt nach der Migration
unverändert bestehen.

Diese **einmalige Ausnahme vom Append-only-Prinzip** des Kassenjournals
und vom Event-Contract-Freeze ist vertretbar, weil:

- das Projekt vor v1.0.0 steht und Breaking Changes ausdrücklich in
  Kauf genommen werden;
- die einzige bekannte produktive Instanz im Testbetrieb ohne TSE läuft
  und keine aufbewahrungspflichtigen Echtdaten trägt;
- der Event-Typ nicht signaturpflichtig ist: keine TSE-Signaturen, keine
  DSFinV-K-Relevanz, keine Geldbeträge — die Löschung berührt keine
  fiskalische Nachvollziehbarkeit.

Die Exklusivität der Event-Switches (unbekannter Typ → Fehler) bleibt
als Integritäts-Sicherheitsnetz unverändert bestehen.

Folgeentscheidungen:

- „Offene Arbeit" (Live-Reporting, Tischlisten-Signal, Tisch-Badge)
  bedeutet fortan „noch nicht kassiert" statt „noch nicht ausgegeben".
- K-13 (Küchendisplay) und K-15 (Zubereitungsstatus) werden von der
  Roadmap gestrichen — sie bauen auf demselben, im Praxistest
  verworfenen Ausgabe-Tracking auf. Eine Wiederaufnahme bliebe über ein
  neues ADR möglich.

## Konsequenzen

- Der Bestellen-Tab enthält nur noch die Bestellaufnahme; die
  Tisch-Seite wird spürbar einfacher. Servicekräfte pflegen keinen
  Status mehr, von dem nichts abhängt.
- Die physische Ausgabe-Koordination liegt allein bei den Arbeitsbons
  (K-12) — bewusst, weil die Praxis genau das bereits so lebt.
- Das Kassenjournal enthält nach der Migration keinen
  `ausgabe-bestaetigt:v1`-Eintrag mehr; Versionslücken je Subject sind
  unschädlich (die optimistische Nebenläufigkeitskontrolle nutzt nur die
  Maximal-Version).
- Nach v1.0.0 wäre diese Art der Entfernung nicht mehr möglich; dann
  gälte wieder: Schreibpfad entfernen, Lesepfad als Legacy behalten.
- Mit diesem ADR wird `docs/adrs/` als Ort für Entscheidungen mit
  langfristiger Tragweite etabliert (Konventionen: siehe README im
  Verzeichnis).
