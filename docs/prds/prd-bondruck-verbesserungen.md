# PRD: Bondruck-Verbesserungen — Druckstationen vereinheitlichen, Outbox robust machen

## Problem Statement

Beim ersten Usability-Test (10. Juni 2026, Vereinsheim, zwei Bondrucker) sind zwei Problemfelder sichtbar geworden:

1. **Ein ausgefallener Drucker legt den gesamten Bondruck lahm.** Das Print-Relay verarbeitet alle Druckaufträge sequenziell und wiederholt einen nicht zustellbaren Auftrag bis zu 60-mal mit je 5 Sekunden Wartezeit (bis zu 5 Minuten pro Auftrag). Während ein Drucker im falschen Subnetz hing, wurde der zweite, erreichbare Drucker nicht bedient. Der einzige Ausweg im Betrieb war ein manueller Eingriff in die Datenbank — für ehrenamtliche Teams inakzeptabel.

2. **Die Drucker-Konfiguration ist auf zwei Seiten verteilt und schwer verständlich.** Arbeitsbon-Stationen werden auf der Druckstationen-Seite konfiguriert, Kassenbeleg- und Abholbon-Drucker dagegen in einer separaten „Bondruck"-Sektion der Einstellungen-Seite — mit einem dreistufigen Direktverkauf-Modus-Select und einem konditional ausgegrauten IP-Feld. Admins müssen verstehen, wie zwei Konfigurationsorte und ein globaler Modus zusammenspielen.

Zusätzlich gibt es keinerlei Sichtbarkeit über hängende oder nicht zustellbare Druckaufträge — weder für Admins noch für Servicekräfte.

## Solution

1. **Eine Konfigurationsseite, ein Modell:** Alle Drucker werden als Druckstationen auf einer Seite konfiguriert — drei Produktkategorien (Essen, Getränk, Sonstiges) plus zwei neue Kategorien **Kassenbeleg** und **Abholbon**. Die Singleton-Tabelle `bondruck_einstellungen` entfällt ersatzlos.

2. **Direktverkauf-Routing per Ableitung statt Modus-Select:** Der explizite Direktverkauf-Modus (`kein_bon` / `abholbon` / `an_stationen`) entfällt. Stattdessen gilt: Ist die Abholbon-Station konfiguriert, erzeugt ein Direktverkauf einen Abholbon. Sonst gehen die Positionen an die Produktstationen (sofern konfiguriert). Ist gar nichts konfiguriert, wird nichts gedruckt. Erklärungstexte auf der Seite machen die Regel sichtbar.

3. **Robuste Outbox mit Fehlversuchs-Zählung:** Das Relay versucht pro Poll-Zyklus genau einen Zustellversuch pro Auftrag (kurzer TCP-Timeout, kein Sleep-Retry) und bedient Drucker unabhängig voneinander. Fehlversuche werden ans Backend gemeldet und dort gezählt; nach drei Fehlversuchen wechselt der Auftrag in den Status `fehlgeschlagen` und wird nicht weiter ausgeliefert. Ein toter Drucker kostet pro Zyklus nur noch einen kurzen Verbindungsversuch — andere Drucker drucken ungestört weiter.

4. **Fehlgeschlagene Aufträge sichtbar und verwaltbar:** Die Druckstationen-Seite zeigt fehlgeschlagene Druckaufträge mit Details und bietet „Erneut versuchen" und „Verwerfen" an. Kein manueller DB-Eingriff mehr nötig.

## User Stories

### Konfiguration

1. Als Admin möchte ich alle Drucker (Produktstationen, Kassenbeleg, Abholbon) auf einer einzigen Seite konfigurieren, damit ich nicht zwischen zwei Seiten wechseln und deren Zusammenspiel verstehen muss.
2. Als Admin möchte ich pro Produktkategorie eine Drucker-IP und einen Bonmodus (pro Position / pro Bestellung) festlegen, damit Arbeitsbons an der richtigen Station in der passenden Stückelung ankommen.
3. Als Admin möchte ich den Kassenbeleg-Drucker als eigene Druckstation konfigurieren, damit Kassenbelege gedruckt werden können.
4. Als Admin möchte ich den Abholbon-Drucker als eigene Druckstation konfigurieren, damit Direktverkäufe einen Abholbon erzeugen.
5. Als Admin möchte ich eine Druckstation durch Leeren der IP deaktivieren, damit jotti auch ganz oder teilweise ohne Drucker betrieben werden kann.
6. Als Admin möchte ich keinen separaten Direktverkauf-Modus mehr pflegen, damit sich das Druckverhalten allein aus der Stationskonfiguration ergibt.
7. Als Admin möchte ich auf der Seite kurze Erklärungstexte je Station sehen (insbesondere zur Abholbon-Ableitungsregel), damit ich ohne Handbuch verstehe, was meine Konfiguration bewirkt.
8. Als Admin möchte ich bei einer ungültigen IP-Adresse sofort eine verständliche Feldmeldung sehen, damit Tippfehler nicht erst als generischer Speicherfehler auffallen.

### Direktverkauf-Routing

9. Als Servicekraft möchte ich, dass ein Direktverkauf einen Abholbon erzeugt, wenn die Abholbon-Station konfiguriert ist, damit Gäste ihre Bestellung an der Ausgabe abholen können.
10. Als Servicekraft möchte ich, dass Direktverkauf-Positionen an die Produktstationen gehen, wenn keine Abholbon-Station konfiguriert ist, damit Küche und Bar die Zubereitung sehen.
11. Als Servicekraft möchte ich Direktverkäufe auch ohne einen einzigen konfigurierten Drucker abschließen können, damit der Verkauf nie am Bondruck scheitert.

### Robustheit der Druck-Pipeline

12. Als Servicekraft möchte ich, dass meine Bestellung sofort an der erreichbaren Station gedruckt wird, auch wenn ein anderer Drucker gerade ausgefallen ist, damit der Service nicht ins Stocken gerät.
13. Als Admin möchte ich, dass ein nicht erreichbarer Drucker die Druckwarteschlange nicht blockiert, damit ein Konfigurationsfehler an einem Gerät nicht das ganze Fest betrifft.
14. Als Admin möchte ich, dass ein Druckauftrag nach drei erfolglosen Zustellversuchen als fehlgeschlagen markiert wird, damit die Warteschlange sich selbst bereinigt statt endlos zu wiederholen.
15. Als Admin möchte ich, dass ein vorübergehend nicht erreichbarer Drucker (z. B. Papierwechsel) Aufträge nicht sofort scheitern lässt, damit kurze Aussetzer folgenlos bleiben (ein Versuch pro Poll-Zyklus, drei Zyklen Karenz).
16. Als Admin möchte ich im laufenden Betrieb niemals direkt in die Datenbank eingreifen müssen, damit auch technisch nicht versierte Helfer das System betreiben können.

### Verwaltung fehlgeschlagener Aufträge

17. Als Admin möchte ich fehlgeschlagene Druckaufträge mit Bon-Art, Ziel-IP, Referenz, Zeitpunkt, Versuchszahl und letztem Fehler auf der Druckstationen-Seite sehen, damit ich Probleme erkennen und einordnen kann.
18. Als Admin möchte ich einen fehlgeschlagenen Auftrag erneut in die Warteschlange geben können (Zähler zurückgesetzt), damit der Bon nach Beheben des Druckerproblems doch noch gedruckt wird.
19. Als Admin möchte ich einen fehlgeschlagenen Auftrag verwerfen können, damit veraltete Bons (z. B. längst ausgegebene Bestellungen) die Liste nicht verstopfen.
20. Als Betreiber möchte ich, dass verworfene Aufträge in der Datenbank nachvollziehbar bleiben (Status statt Löschung), damit die technische Warteschlange im Zweifel rekonstruierbar ist.
21. Als Servicekraft möchte ich beim Kassenbeleg-Druck eine klare Fehlermeldung erhalten, wenn keine Kassenbeleg-Station konfiguriert ist, damit ich weiß, dass es ein Konfigurations- und kein Bedienproblem ist.

### Entwicklung & Betrieb

22. Als Entwickler möchte ich, dass das Relay-Modul von `make check` mitgeprüft wird (Lint, Tests, Build), damit Regressionen in der Druck-Pipeline vor dem Merge auffallen.
23. Als Entwickler möchte ich die Relay-Zyklus-Logik isoliert testen können, damit das Fehlverhalten aus dem Usability-Test nie unbemerkt zurückkehrt.

## Implementation Decisions

### Datenmodell

- Die Tabelle `druckstationen` erhält einen neuen Kategorie-Typ mit fünf Werten: `essen`, `getraenk`, `sonstiges`, `kassenbeleg`, `abholbon` (fünf Seed-Zeilen, Kategorie bleibt Primärschlüssel). Der bisherige Produktkategorie-Enum bleibt unverändert für Produkte bestehen.
- `bonmodus` ist nur für die drei Produktkategorien belegt; für `kassenbeleg` und `abholbon` ist er NULL. Ein CHECK-Constraint erzwingt diese Kopplung.
- Die Tabelle `bondruck_einstellungen` entfällt ersatzlos — inklusive Domain-Typ, Repository-Methoden, beider Admin-Endpoints und der Frontend-Sektion. Der Direktverkauf-Modus existiert in keiner Schicht mehr.
- Die Outbox `druckauftraege` erhält: erweiterten Status-CHECK (`offen`, `gedruckt`, `fehlgeschlagen`, `verworfen`), eine Spalte `versuche` (Integer, Default 0) und `letzter_fehler` (Text, NULL).
- Schema-Änderungen direkt in der initialen Migration (Pre-Release-Regel, keine neuen Migrationsdateien).

### Domain & Validierung

- Ein einheitliches Validierungs-Muster für den Druck-Bereich: typisierte Kategorie- und Bonmodus-Werte mit Konstruktor-Validierung im Domain-Modell (wie bisher bei den Bondruck-Einstellungen), zog/Zod-Schemas weiterhin an den HTTP-Rändern beider Seiten.

### Bondruck-Routing (Policy)

- Die Arbeitsbon-Policy bleibt ein tiefes, isoliert testbares Modul: Eingabe ist ein Event plus die Map aller konfigurierten Druckstationen, Ausgabe sind fertige Druckaufträge.
- Routing-Regeln:
  - `bestellung-aufgenommen` → Produktstationen je Kategorie (unverändert).
  - `direktverkauf-getaetigt` → Abholbon-Station, falls konfiguriert; sonst Produktstationen je Kategorie; sind keine Stationen konfiguriert, entstehen keine Aufträge.
- Die bisherige Direktverkauf-Konfigurationsstruktur (Modus + Abholbon-IP) entfällt aus der Policy-Schnittstelle.
- Der Kassenbeleg-Command bezieht die Ziel-IP aus der Kassenbeleg-Druckstation statt aus den Bondruck-Einstellungen; die bestehende Fehlersemantik („kein Kassenbeleg-Drucker konfiguriert") bleibt erhalten.

### Relay-Protokoll & Outbox-Statusübergänge

- Der Poll-Endpoint bleibt unverändert (offene Aufträge, älteste zuerst).
- Der Quittieren-Endpoint wird zum Ergebnis-Endpoint erweitert: Das Relay meldet pro Zyklus in einem Request die erfolgreich gedruckten IDs und die Fehlversuche (ID + Fehlertext).
- Das Backend besitzt die Fehlversuchs-Logik: Pro gemeldetem Fehlversuch wird `versuche` inkrementiert und `letzter_fehler` aktualisiert; erreicht ein Auftrag drei Versuche, wechselt er auf `fehlgeschlagen`. Das Limit ist eine Backend-Konstante.
- Statusübergänge: `offen → gedruckt` (Quittierung), `offen → fehlgeschlagen` (drittes Fehlversuchs-Inkrement), `fehlgeschlagen → offen` (erneut versuchen, Zähler 0), `fehlgeschlagen → verworfen` (verwerfen). Einträge werden nie gelöscht.
- At-least-once-Semantik bleibt bewusst erhalten: Das Relay druckt erst und meldet danach. Schlägt die Meldung fehl, kann ein Bon doppelt gedruckt werden — für Arbeitsbons unkritisch, für Kassenbelege wie bisher akzeptiert (erneuter Druck wiederholt keinen fachlichen Vorgang).

### Relay-Zyklus

- Aufträge werden pro Zyklus nach Ziel-IP gruppiert; die Gruppen werden parallel verarbeitet (ein Drucker kann einen anderen nicht blockieren). Innerhalb einer Ziel-IP bleibt die ID-Reihenfolge erhalten; zwischen Druckern gibt es keine Reihenfolge-Garantie.
- Pro Auftrag genau ein Zustellversuch pro Zyklus mit kurzem TCP-Timeout; das bisherige Sleep-Retry (60 × 5 s) entfällt vollständig.
- Schlägt der erste Verbindungsaufbau zu einer Ziel-IP fehl, werden die übrigen Aufträge dieser IP im selben Zyklus übersprungen; übersprungene Aufträge zählen nicht als Fehlversuch und werden nicht gemeldet.
- Die Zyklus-Logik wird testbar geschnitten: Druck- und Melde-Funktionen sind injizierbar, sodass die Gruppierungs-, Skip- und Melde-Logik ohne echte Drucker getestet werden kann.

### Admin-API & Frontend

- Neue admin-only POST-Endpoints: fehlgeschlagene Druckaufträge auflisten, Auftrag erneut versuchen, Auftrag verwerfen.
- Die Druckstationen-Seite wird die einzige Druck-Konfigurationsseite: fünf Stationszeilen mit kategoriespezifischen Feldern (Bonmodus nur bei Produktkategorien), Erklärungstexten zur Ableitungsregel und Inline-IPv4-Validierung mit Feldmeldung (statt clientseitigem Schema-Throw mit generischem Toast).
- Darunter ein Abschnitt „Fehlgeschlagene Druckaufträge" mit Liste (Bon-Art, Ziel-IP, Referenz, Zeitpunkt, Versuche, letzter Fehler) und den Aktionen „Erneut versuchen" und „Verwerfen".
- Die Einstellungen-Seite verliert ihre Bondruck-Sektion; die Frontend-Backend-Klassen werden entsprechend verschoben/erweitert.

### Aufräumarbeiten im Zuge der Umsetzung

- Die doppelten 1:1-Mappings zwischen Policy-Druckauftrag und Repository-Insert-Typ werden auf einen gemeinsamen Typ bzw. einen Helfer reduziert.
- Die reinen Forwarding-Schichten der Relay-API werden bei der Erweiterung um den Ergebnis-Endpoint gestrafft.
- Das Relay-Modul wird in `make check` aufgenommen (Lint, Tests, Build).
- Die Referenzdokumente (Handbuch-Abschnitt Bondruck, Sprachregelungen, Anforderungen) werden an das neue Modell angepasst (Statusmodell der Outbox, Wegfall der Bondruck-Einstellungen, Ableitungsregel).

## Testing Decisions

- **Grundsatz:** Tests prüfen externes Verhalten (Eingabe → beobachtbares Ergebnis), keine Implementierungsdetails. Schnittstellen der tiefen Module sind der Testansatzpunkt.
- **Bondruck-Policy (Unit):** Routing-Regeln vollständig abdecken — Abholbon-Ableitung (konfiguriert/nicht konfiguriert), Stations-Routing je Kategorie, Bonmodus-Gruppierung, leere Konfiguration. Prior Art: bestehende Policy-Unit-Tests im Bondruck-Modul, die erweitert werden.
- **Outbox-Statusübergänge:** Fehlversuchs-Zählung (1, 2, 3 → fehlgeschlagen), erneut versuchen (Zähler-Reset), verwerfen, idempotentes Quittieren. Als Repository-Integrationstests gegen echte Datenbank und Handler-Tests mit Mock-Repository. Prior Art: bestehende Integrationstests und httptest-basierte Handler-Tests.
- **Relay-Zyklus (Unit):** Gruppierung nach Ziel-IP, Skip nach Erstfehler ohne Fehlversuchs-Meldung, korrekte Erfolg-/Fehlversuchs-Meldung, Reihenfolge innerhalb einer IP — mit injizierten Fake-Funktionen statt echter Drucker. Schließt die größte bestehende Testlücke (bisher ist nur das Config-Parsing getestet).
- **Bewusst nicht geplant:** Frontend-Page-Tests für die erweiterte Druckstationen-Seite. Die Zod-Schemas der Backend-Klassen sind über das bestehende Test-Muster abgedeckt.

## Out of Scope

- **Küchen-Display (KDS, K-13) und Zubereitungsstatus (K-15)** — bleiben offene, separate Anforderungen.
- **Queue-Monitor für offene Aufträge** (Anzahl, ältester Auftrag) — die Admin-UI zeigt ausschließlich fehlgeschlagene Aufträge; gedruckte und verworfene Aufträge bekommen keine eigene Ansicht.
- **Beibehaltung des Direktverkauf-Modus** in irgendeiner Form — bewusster Verzicht: Der Fall „Produktstationen konfiguriert, aber Direktverkauf ohne Bons" ist nicht mehr abbildbar.
- **Drucker-Discovery** (mDNS o. ä.), Druckerstatus-Dashboards, Unterstützung mehrerer paralleler Relays mit Job-Locking.
- **Datenmigration** bestehender Konfigurationen oder Aufträge — Pre-Release, Breaking Changes erwünscht, Dev-Datenbanken werden neu aufgesetzt.
- **Frontend-Page-Tests** für die Druckstationen-Seite (siehe Testing Decisions).

## Further Notes

- **Verlorene Fähigkeit (bewusste Entscheidung):** Durch die Ableitungsregel kann ein Verein mit konfigurierten Produktstationen Direktverkauf-Bons nicht mehr unterdrücken. Sollte sich das im Praxiseinsatz als Problem zeigen, wäre ein Modus-Feld an der Abholbon-Zeile die Rückfalloption — das Datenmodell lässt diese Erweiterung zu.
- **Papierstatus-Prüfung im Relay:** Die bestehende „Papier fast leer"-Warnung wertet vermutlich die falschen Statusbits aus (Ende-Bit statt Near-End-Bits der DLE-EOT-4-Antwort). Beim Umbau der Zyklus-Logik gegen die ESC/POS-Dokumentation verifizieren und korrigieren.
- **UX-Review:** Nach der Implementierung ist ein UX-Review der neuen Druckstationen-Seite sinnvoll (Verständlichkeit der Ableitungsregel, Mobile-Darstellung der fünf Zeilen plus Fehlgeschlagen-Liste).
- **Audit-Bezug:** Dieses PRD adressiert die Befunde 1, 2, 4, 5, 6, 7, 8, 9 und 11 des Code-Audits vom 11. Juni 2026 (Relay-Blocking, Doppeldruck-Semantik, totes UpdatedAt-Feld, IP-Validierungs-UX, Umlaut-Texte, Mapping-Duplikate, Relay-Schichten, Validierungs-Muster, fehlende Relay-Checks).
