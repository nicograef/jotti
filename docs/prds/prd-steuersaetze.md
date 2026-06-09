# PRD: Steuersätze (F-07)

> **Status:** Entwurf · **Anforderung:** [F-07 · Steuersätze](../anforderungen.md) · **Rechtsgrundlage:** [steuerrecht.md](../steuerrecht.md) · **Phase:** 1 (Compliance-Grundlage) · **Prio:** Must-have

## Problem Statement

Als gemeinnütziger Verein verkauft die Organisation auf ihrem Fest Speisen und Getränke im **wirtschaftlichen Geschäftsbetrieb** und ist damit umsatzsteuerpflichtig. Der Kassenwart muss am Ende der Veranstaltung die abzuführende Umsatzsteuer pro Steuersatz ermitteln und gegenüber dem Finanzamt belegen können.

Heute kennt jotti keinen Steuersatz: Ein Produkt hat nur einen Preis, mehr nicht. Daraus ergeben sich konkrete Probleme:

- Die **Tagesabrechnung** zeigt einen Gesamtumsatz, aber keine Aufschlüsselung nach Steuersatz — die USt-Schuld lässt sich nicht ablesen.
- Ein **Kassenbeleg** (F-03) kann den gesetzlich geforderten Steuerausweis (Netto, Steuersatz, Steuerbetrag pro Position) nicht enthalten, weil die Information fehlt.
- Der **DSFinV-K-Export** (F-04) kann die nach Steuersätzen aufgeteilten Beträge nicht liefern.
- Selbst wenn der Steuersatz nachträglich ergänzt würde, wären **historische Bestellungen** nicht mehr korrekt zuordenbar, weil der zum Verkaufszeitpunkt geltende Satz nicht festgehalten wurde.

Ohne Steuersätze bleibt jotti rechtlich nicht einsetzbar für den Zweck, für den es gedacht ist.

## Solution

Jedes Produkt wird einem **Steuersatz** zugeordnet — einem Pflichtfeld mit den vier unterstützten Werten: `regel` (19 %), `ermaessigt` (7 %), `befreit` (0 %) und `kombi` (70/30-Pauschalierung: 70 % → 7 %, 30 % → 19 %). Der Admin wählt den Satz bei der Produktanlage und -bearbeitung; das Frontend schlägt basierend auf der Produktkategorie einen passenden Default vor (Essen → ermäßigt, Getränk/Sonstiges → Regelsteuersatz), der vom Admin geändert werden kann.

Der Wert `kombi` bildet **Kombinationsangebote** ab (z. B. „Schnitzel + 1 Bier" als ein Produkt mit Pauschalpreis). Rechtsgrundlage ist die 30/70-Pauschalierung nach Abschn. 10.1 Abs. 12 UStAE (siehe [steuerrecht.md §4](../steuerrecht.md)): 70 % des Bruttopreises werden dem ermäßigten Satz (7 %, Speisen) zugeordnet, 30 % dem Regelsteuersatz (19 %, Getränke). Die Aufteilung erfolgt vollständig im `steuer`-Rechenmodul; für Reporting und Downstream-Consumer (Kassenbeleg, DSFinV-K) werden **zwei Steuer-Anteile** erzeugt.

Der zum Bestellzeitpunkt geltende Steuersatz wird — wie alle anderen Produktdaten — **unveränderlich in jede Position eingefroren** (Fat Event). Spätere Änderungen am Produkt berühren bereits aufgenommene Bestellungen nicht.

Ein zentrales, isoliertes **Steuer-Rechenmodul** leitet aus dem Bruttopreis (= der Preis, den der Gast zahlt) und dem Steuersatz den Nettobetrag und den Steuerbetrag ab. Dieses Modul ist die gemeinsame Grundlage für die Tagesabrechnung (dieses PRD), den Kassenbeleg (F-03) und den DSFinV-K-Export (F-04).

Die **Tagesabrechnung** erhält einen neuen Abschnitt, der den kassierten Umsatz pro Steuersatz mit Brutto-, Netto- und Steuerbetrag ausweist.

Für die Servicekraft ändert sich **nichts**: Der Steuersatz fließt unsichtbar aus dem Produkt in die Bestellung; die Bestellaufnahme bleibt unverändert.

> **Klarstellung der Preis-Semantik:** Der gespeicherte Produktpreis ist ein **Bruttopreis** (inkl. USt) — der Betrag, den der Gast tatsächlich zahlt. Die bestehenden Feldnamen (`preis_cents`, `PreisCents`, `einzelpreis`) bleiben unverändert; sie sind neutral benannt und tragen bereits Brutto-Semantik. Eine Umbenennung findet nicht statt — die Brutto-Semantik wird lediglich in der Dokumentation explizit gemacht. Netto- und Steuerbetrag werden stets aus dem Bruttopreis abgeleitet, nie aufgeschlagen.

## User Stories

1. Als Admin möchte ich beim Anlegen eines Produkts einen Steuersatz auswählen, damit der Verkauf umsatzsteuerlich korrekt erfasst wird.
2. Als Admin möchte ich den Steuersatz eines bestehenden Produkts ändern können, damit ich Fehlkonfigurationen oder Gesetzesänderungen abbilden kann.
3. Als Admin möchte ich aus den vier relevanten Sätzen wählen — 19 % (Regelsteuersatz), 7 % (ermäßigt), 0 % (befreit), Kombi (70/30) — damit ich keine ungültigen Werte eingeben kann.
4. Als Admin möchte ich, dass der Steuersatz ein Pflichtfeld ist, damit kein Produkt versehentlich ohne steuerliche Zuordnung verkauft wird.
5. Als Admin möchte ich, dass beim Anlegen eines Produkts der Steuersatz passend zur Kategorie vorausgewählt ist (Essen → 7 %, Getränk/Sonstiges → 19 %), damit der Regelfall ohne manuellen Eingriff korrekt ist.
6. Als Admin möchte ich den vorgeschlagenen Steuersatz ändern können (z. B. Leitungswasser als Getränk mit 7 %, Festbändchen als Sonstiges mit 0 %), damit Sonderfälle abbildbar sind.
7. Als Admin möchte ich für Artikel des Zweckbetriebs (z. B. Festbändchen) oder bei Anwendung der Kleinunternehmerregelung den Satz „befreit" (0 %) setzen können, damit steuerbefreite Umsätze korrekt ausgewiesen werden.
8. Als Admin möchte ich für Kombinationsangebote (z. B. „Schnitzel + 1 Bier") den Satz „Kombi (70/30)" setzen können, damit der Pauschalpreis gemäß der Vereinfachungsregelung (Abschn. 10.1 Abs. 12 UStAE) steuerlich korrekt auf Speisen- und Getränke-Anteil aufgeteilt wird.
9. Als Servicekraft möchte ich Bestellungen wie gewohnt aufnehmen, ohne mich um Steuersätze kümmern zu müssen, weil der Satz automatisch aus dem Produkt übernommen wird.
10. Als System möchte ich den Steuersatz jeder Position zum Bestellzeitpunkt einfrieren, damit historische Auswertungen auch nach Produktänderungen korrekt bleiben.
11. Als System möchte ich den Steuersatz serverseitig aus dem Produktkatalog ermitteln, damit das Frontend den Satz nicht manipulieren kann.
12. Als System möchte ich den Steuersatz in allen positionsführenden Events (Bestellung, Zahlung, Ausgabe, Stornierung, Direktverkauf) konsistent mitführen, damit jede Auswertung über jeden Vorgang dieselbe Steuerinformation findet.
13. Als Kassenwart möchte ich in der Tagesabrechnung den Umsatz pro Steuersatz sehen, damit ich die abzuführende Umsatzsteuer ablesen kann.
14. Als Kassenwart möchte ich pro Steuersatz Brutto-, Netto- und Steuerbetrag sehen, damit ich die Beträge direkt in die Buchhaltung bzw. Voranmeldung übernehmen kann.
15. Als Kassenwart möchte ich, dass der kassierte Umsatz (Zahlungen + Direktverkäufe abzüglich Direktverkauf-Stornos) die Basis der Aufschlüsselung ist, damit die ausgewiesene Steuer der tatsächlich vereinnahmten Steuer entspricht.
16. Als Kassenwart möchte ich, dass ein steuerbefreiter Umsatz (0 %) als eigene Zeile mit Steuerbetrag 0 erscheint, damit auch befreite Umsätze nachvollziehbar dokumentiert sind.
17. Als Entwickler möchte ich ein isoliertes, gut getestetes Rechenmodul für die Brutto→Netto/Steuer-Aufteilung, damit Kassenbeleg (F-03) und DSFinV-K-Export (F-04) dieselbe korrekte Logik wiederverwenden können.
18. Als Entwickler möchte ich, dass die Steuer-Aufteilung kaufmännisch korrekt rundet, damit Netto + Steuerbetrag exakt den Bruttobetrag ergeben.
19. Als Betreiber möchte ich, dass die Steuerinformation Teil der unveränderlichen Event-Historie ist, damit die GoBD-Anforderung der Nachvollziehbarkeit erfüllt wird.
20. Als Entwickler möchte ich den Steuersatz als deutschen Enum-Schlüssel speichern und den Prozentsatz aus einem stabilen Domain-Mapping ableiten, damit die Events schlank bleiben und die Klassifikation eindeutig ist.
21. Als Admin möchte ich beim Bearbeiten eines Produkts den aktuell gesetzten Steuersatz vorausgewählt sehen, damit ich Änderungen bewusst vornehme.
22. Als Nutzer der Validierung (Frontend und Backend) möchte ich bei fehlendem oder ungültigem Steuersatz eine klare deutsche Fehlermeldung erhalten, damit ich den Fehler sofort beheben kann.

## Implementation Decisions

### Steuersatz und Rechenmodul (deep module)

- Es entsteht ein neues, schlankes Support-Paket **`steuer`** in der Domain-Schicht. Es ist bewusst kein Bounded-Context-Slice (kein zugehöriges `api/`- oder Repo-Paket), analog zu bestehenden Support-Paketen wie `event` und `jwt`.
- Das Paket enthält das **Wertobjekt `Steuersatz`** als deutschen Enum mit den Werten `regel`, `ermaessigt`, `befreit`, `kombi` sowie einer Methode `Prozent()`, die den Prozentsatz liefert (19 / 7 / 0 / —). Go-Konstanten: `RegelSteuersatz`, `ErmaessigtSteuersatz`, `BefreitSteuersatz`, `KombiSteuersatz`. Das Enum→Prozent-Mapping ist die einzige Quelle der Wahrheit und wird als append-only behandelt (ein Schlüssel wird nie umgewidmet). Für `kombi` gibt es keinen einzelnen Prozentsatz — stattdessen liefert die Aufteilungs-Rechnung zwei Anteile.
- Das Paket enthält die **Aufteilungs-Rechnung** als zentrale Funktion `Aufteilen(brutto int, satz Steuersatz) []Aufteilung`. Jede `Aufteilung` enthält: `Satz` (effektiver Steuersatz), `Brutto`, `Netto`, `Steuer`. Für `regel`/`ermaessigt`/`befreit` wird genau ein Element zurückgegeben. Für `kombi` werden **zwei Elemente** zurückgegeben:
  - Speisen-Anteil: `Brutto_Speisen = round(Brutto × 70 / 100)`, aufgeteilt mit 7 %
  - Getränke-Anteil: `Brutto_Getraenke = Brutto − Brutto_Speisen`, aufgeteilt mit 19 %
  - Die Summe der beiden Brutto-Anteile ergibt exakt den Gesamt-Bruttobetrag (kein Rundungsverlust durch die Restwert-Zuweisung).
- Für die einfachen Sätze gilt: `Steuerbetrag = round(Brutto × p / (100 + p))`, `Netto = Brutto − Steuerbetrag`. Gerundet wird kaufmännisch (round half up) auf ganze Cent. Für `befreit` ist der Steuerbetrag 0 und Netto = Brutto.
- Eine zog-Schema-Validierung für `Steuersatz` wird bereitgestellt und auf beiden relevanten Ebenen (Produkt, Position) verwendet.

### Produkt (Stammdaten)

- Der Steuersatz ist eine Eigenschaft des **Produkts**, nicht der Variante: Alle Varianten eines Produkts teilen denselben Satz (eine Cola ist unabhängig von der Größe mit demselben Satz belegt).
- Das Domain-Modell `Produkt` erhält ein Pflichtfeld `Steuersatz`. Die Erzeugungs- und Aktualisierungsoperationen des Produkts nehmen den Steuersatz entgegen und validieren ihn. Domain-Structs tragen weiterhin **keine** `json`-Tags.
- Die Datenbank erhält in der Tabelle `produkte` eine neue Spalte `steuersatz` (NOT NULL) auf Basis eines neuen Enum-Typs `Steuersatz` (`regel`, `ermaessigt`, `befreit`, `kombi`). Die Änderung erfolgt direkt in der initialen Migration; es werden keine neuen Migrationsdateien angelegt (Dev-DB-Reset-Modell).
- Die sqlc-Queries für Produkt-Anlage, -Aktualisierung und -Abfrage werden um den Steuersatz erweitert; generierter Code wird neu erzeugt. `sqlc/dbgen/` wird nicht von Hand editiert.
- Die Admin-API (Produkt anlegen/bearbeiten) nimmt den Steuersatz im Request-DTO entgegen; das Produkt-Response-DTO weist ihn aus. Domain-Modelle werden nicht direkt serialisiert.
- Das Admin-Frontend zeigt im Produktformular eine Auswahl des Steuersatzes mit deutschen Labels: „19 % (Regelsteuersatz)", „7 % (ermäßigt)", „0 % (befreit)", „Kombi 70/30 (Menü)". Beim Anlegen wird der Steuersatz **kategoriebasiert vorausgewählt**: Kategorie `essen` → `ermaessigt`, Kategorie `getraenk` oder `sonstiges` → `regel`. Der Admin kann den Vorschlag jederzeit ändern (das Feld ist kein Read-only). Beim Bearbeiten ist der gespeicherte Satz vorausgewählt. Die Validierung erfolgt zusätzlich mit Zod.

### Position und Events (Kasse)

- Die geteilte, „fette" Position der Kasse erhält ein Feld `steuersatz` (Enum-Schlüssel). Damit fließt der Satz automatisch in **alle** positionsführenden Events: `bestellung-aufgenommen:v1`, `zahlung-kassiert:v1`, `ausgabe-bestaetigt:v1`, `stornierung-erteilt:v1`, `direktverkauf-getaetigt:v1`, `direktverkauf-storniert:v1`.
- Der JSON-Key in den Event-Daten ist `steuersatz` (camelCase, Domänenbegriff). Event-Data-Structs behalten ihre `json`-Tags (Persistenz-Serialisierung).
- Der Steuersatz wird **serverseitig** beim Auflösen der Positionen aus dem Produktkatalog ermittelt — dort, wo heute bereits Produktname, Variantenname, Kategorie und Einzelpreis aus dem Produkt eingefroren werden. Der lean API-Request (Positionsreferenz aus Positions-ID/Variante + Menge) wird nicht um den Steuersatz erweitert; das Frontend sendet keinen Steuersatz.
- Das Positions-Schema (zog) macht `steuersatz` zu einem Pflichtfeld und validiert gegen die erlaubten Werte.

### Reporting (Tagesabrechnung)

- Eine neue SQL-Query aggregiert den Bruttoumsatz **pro Steuersatz**, indem sie die Positions-Arrays der Basis-Events entfaltet (`jsonb`-Unnest) und `einzelpreis × menge` je `steuersatz` summiert. Basis sind die kassierten Umsätze: `zahlung-kassiert:v1` plus `direktverkauf-getaetigt:v1` minus `direktverkauf-storniert:v1`, gefiltert auf die gewählte Kassensitzung.
- Die Geldaggregation (Brutto pro Satz) bleibt in SQL (konsistent mit den bestehenden Extraktionsfunktionen); die Ableitung von Netto und Steuerbetrag erfolgt in der Reporting-Anwendungsschicht über das `steuer`-Rechenmodul. Zur Vermeidung von Rundungsdrift wird Netto/Steuer **auf dem aggregierten Bruttobetrag pro Satz** berechnet, nicht je Einzelposition aufsummiert.
- **Kombi-Auflösung im Reporting:** SQL gruppiert `kombi`-Positionen separat. Die Reporting-Anwendungsschicht ruft `Aufteilen(brutto, kombi)` auf und verteilt die beiden Ergebnis-Anteile (7 % und 19 %) in die entsprechenden Steuersatz-Zeilen. In der Tagesabrechnung erscheint keine eigene „Kombi"-Zeile — die Beträge fließen anteilig in die 7-%- und 19-%-Zeilen ein.
- Das Reporting-Domain-Modell erhält einen neuen Typ, der pro Steuersatz Brutto-, Netto- und Steuerbetrag trägt; er wird in die Reporting-Daten (Tagesabrechnung und Live-Übersicht, soweit konsistent) aufgenommen und über das Response-DTO ausgewiesen.
- Die Tagesabrechnungs-Seite im Admin-Frontend erhält einen neuen Abschnitt „Umsatz nach Steuersatz" mit einer Zeile je tatsächlich vorkommendem Satz (Brutto, Netto, Steuerbetrag). Das Backend bleibt die Single Source of Truth der Aufbereitung; das Frontend zeigt nur an.
- **Dokumentierte Grenze:** Positionslose Auszahlungen (Ausgleich negativer Tischsaldi) tragen keinen Steuersatz und werden der Aufschlüsselung nicht zugeordnet. Die Summe der Brutto-Zeilen kann daher bei vorhandenen Auszahlungen vom Gesamt-Nettoumsatz der Kennzahlen abweichen. Eine vollständige Abstimmung erfolgt erst mit dem DSFinV-K-Export (F-04).

### Seed-Daten

- Die Seed-Produkte erhalten plausible Steuersätze (Getränke: `regel`, Speisen: `ermaessigt`, Festbändchen: `befreit`, Menü „Schnitzel + 1 Bier": `kombi`). Die in den Seed-Events enthaltenen Positionen werden um `steuersatz` ergänzt, damit Reporting-Aggregationen auf Seed-Daten funktionieren.

## Testing Decisions

**Was einen guten Test ausmacht:** Tests prüfen **externes Verhalten**, nicht Implementierungsdetails. Sie beschreiben, was ein Modul bei gegebener Eingabe garantiert, und bleiben stabil, wenn sich die innere Umsetzung ändert. Bevorzugt werden tabellengetriebene Tests für reine Funktionen und verhaltensorientierte Tests für Anwendungs- und Reporting-Logik.

Auf Wunsch werden folgende Module getestet:

1. **Steuer-Aufteilung (Rechenmodul):** Reine Unit-Tests, tabellengetrieben. Schwerpunkt auf Rundungs-Edgecases und der Invariante **Netto + Steuerbetrag = Brutto** für alle Sätze. Beispiele: Brutto 500 ct bei 19 % → Steuer 80 ct, Netto 420 ct; bei 7 % → Steuer 33 ct, Netto 467 ct; `befreit` → Steuer 0, Netto = Brutto; `kombi` 1500 ct → Speisen-Brutto 1050 ct (Steuer 69 ct) + Getränke-Brutto 450 ct (Steuer 72 ct); Betrag 0 → 0/0; Rundung an der Halben-Cent-Grenze; `kombi` mit ungeradem Betrag (Restwert-Zuweisung an Getränke-Anteil). Prior Art: bestehende reine Domain-Tests im Kasse-Kontext (z. B. die Subject- und Event-Tests).
2. **Positions-Anreicherung im Command:** Verhaltensorientierter Test der Anwendungsschicht, der nachweist, dass der Steuersatz **aus dem Produkt** in die Position übernommen und im Event eingefroren wird (und ein vom Client gesendeter Wert ignoriert würde). Prior Art: bestehende Command-Tests der Bestellaufnahme, die das Einfrieren der Produktdaten in Positionen prüfen.
3. **Reporting per-Steuersatz-Aggregation:** Integrationstest gegen die Datenbank, der aus gesetzten Events die korrekte Gruppierung und Summenbildung pro Steuersatz nachweist — inklusive der Basis-Regel (Zahlungen + Direktverkauf − Direktverkauf-Storno) und eines befreiten Umsatzes. Prior Art: bestehende Reporting-Integrationstests der Tagesabrechnung.

Für die Produkt-Validierung (Steuersatz als Pflichtfeld) ist kein dedizierter Test vorgesehen; sie folgt dem etablierten zog/Zod-Muster der übrigen Produktfelder.

## Out of Scope

- **Fiskalischer Kassenbeleg (F-03):** Das tatsächliche Rendern und Drucken des Kassenbelegs mit Steuerausweis ist eine eigene Must-have-Anforderung. Dieses PRD liefert nur die Datenbasis und das Rechenmodul, das F-03 konsumiert.
- **DSFinV-K-Export (F-04)** und **TSE-Integration (F-02):** Aufteilung der Beträge nach Steuersätzen im Exportformat bzw. in TSE-`processData` baut auf diesem PRD auf, ist aber nicht Teil davon.
- **eBeleg (F-09).**
- **Kalkulatorische Aufteilung (Methode A) für Kombinationsangebote:** Die 70/30-Pauschalierung (Methode B) ist implementiert. Eine alternative Methode — anteilige Aufteilung nach den regulären Einzelverkaufspreisen der Menü-Komponenten — ist steuerlich zulässig, aber deutlich komplexer (erfordert Komponentenstruktur pro Produkt) und wird hier nicht umgesetzt.
- **Steuersatz je Variante:** Der Satz bleibt auf Produktebene.
- **Frei konfigurierbare oder zeitlich versionierte Steuersätze** sowie eine **Kleinunternehmer-Umschaltung (§ 19 UStG):** Die vier festen Enum-Werte genügen für den Zielbetrieb. Eine spätere Erweiterung ist möglich, ist aber hier nicht vorgesehen.
- **Anzeige des Steuersatzes in der Service-Bestellansicht:** Der Satz fließt unsichtbar mit; die Bestellaufnahme-UI bleibt unverändert.
- **Migration historischer Events:** Bestehende Events werden nicht rückwirkend um Steuersätze ergänzt (Pre-Release, DB-Reset-Modell).
- **Zuordnung positionsloser Auszahlungen** zu Steuersätzen in der Aufschlüsselung.

## Further Notes

- **Satz-Stabilität:** Die deutschen Sätze sind stabil (ab 2026 gilt für Speisen in der Gastronomie wieder der ermäßigte Satz von 7 %). Sollte sich ein gesetzlicher Satz je ändern, wird dies durch Anpassung des Enum→Prozent-Mappings abgebildet; vorhandene Events behalten ihren Enum-Schlüssel. Bestehende Schlüssel werden nie umgewidmet.
- **Rundungsstrategie:** Für Auswertungen wird Netto/Steuer auf dem aggregierten Bruttobetrag pro Satz gerechnet (keine Summe gerundeter Einzelpositionen → kein Drift). Der spätere Kassenbeleg (F-03) kann pro Position Netto/Steuer ausweisen; dort sind kleine, übliche Cent-Rundungsdifferenzen zwischen Positionssumme und Belegsumme zulässig und werden von DSFinV-K toleriert.
- **Abhängigkeit:** Dieses PRD ist Voraussetzung für F-03, F-04 und das vollständige F-02. Das Rechenmodul ist bewusst als wiederverwendbares, anbieter- und ausgabeunabhängiges Modul geschnitten.
- **Sprachkonventionen:** `Steuersatz` ist ein Domänenbegriff (deutsch, vertikal konsistent: Go `Steuersatz`, TS `Steuersatz`, JSON `steuersatz`, DB-Spalte `steuersatz`, DB-Enum-Typ `Steuersatz`). Enum-Werte: `regel`, `ermaessigt`, `befreit`, `kombi`. Benutzer-sichtbare Labels sind deutsch.
- **Downstream-Anforderung für `kombi`:** Der Kassenbeleg (F-03) muss bei einer `kombi`-Position den Steuerausweis auf **zwei Zeilen** entfalten (7 %-Anteil und 19 %-Anteil) oder mindestens in der Steuermatrix beide Anteile korrekt aufführen. Der DSFinV-K-Export (F-04) muss in `lines_vat.csv` zwei VAT-Einträge pro `kombi`-Position erzeugen. Beide Consumer nutzen dafür `steuer.Aufteilen()` und müssen keine eigene Split-Logik implementieren.
