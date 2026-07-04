# PRD: Modulschnitt nach Kontextgrenzen

> Ausrichtung des Backends und Frontends am Subdomänen-Zielbild: Kontext-Ordner
> im Backend (`kasse`, `fiskal`, `druck`, `stammdaten`, `reporting`), Auflösung
> des God-Moduls `api/table`, deutsche Namen für Fachmodule, Aufspaltung des
> settings-Sammelmoduls, Tagesabschluss-Aggregation als Domänenfunktion statt
> Reporting-Abhängigkeit, Frontend-Feinschnitt und Aktualisierung der
> Architektur-Dokumentation. Reiner Strukturumbau: keine API-Pfad-, Schema-
> oder UI-Änderungen.
>
> **Voraussetzung:** Die PRD „TSE-Signatur über Outbox und Signatur-Worker"
> (prd-tse-signatur-outbox.md) ist zuerst umgesetzt. Diese PRD baut auf deren
> Ergebnis auf (TSE-freier Kassen-Kern, Signaturauftrags-Outbox, Signatur-
> Worker, Signaturstatus, Störungsprotokoll) und ordnet es in den
> Fiskal-Kontext ein.

## Problem Statement

Die Architektur-Dokumentation nennt drei Bounded Contexts (Kasse, Stammdaten,
Auth), der Code lebt aber faktisch mit sechs Subdomänen: Kasse (Core),
Fiskalisierung, Druck/Ausgabe, Stammdaten, Reporting (Supporting) und Auth
(Generic). Diese Diskrepanz erzeugt konkrete Strukturprobleme (die Zahlen-
und Dependency-Angaben beschreiben den Ist-Stand vor der Outbox-PRD; deren
Umbau verkleinert einzelne Dependency-Sätze — etwa entfallen TSESignierer
und SettingsRepo aus den Kassen-Commands —, ändert aber nichts an der
Diagnose):

1. **`api/table` ist ein God-Modul.** Ein einziger Command-Struct mit neun
   Dependencies vereint vier Kontexte: Tisch-Stammdaten-CRUD, das
   Kern-Kassengeschäft (Bestellen, Kassieren, Ausgabe, Storno, Umbuchung),
   den Kassenbeleg-Druck und die Favoriten-Verwaltung. Die Wiring-Dateien
   injizieren je Einsatzort unterschiedliche Teilmengen der Dependencies —
   ein Indiz für einen zu breiten Schnitt. Das Modul widerspricht der im
   Handbuch dokumentierten Kontextgrenze (Tisch-Stammdaten ≠ Tisch-Session)
   und trägt einen englischen Namen für einen deutschen Kernbegriff.
2. **Der Kassenabschluss hängt am Reporting.** Der Abschluss-Command holt die
   Z-Bon-Tagessummen über das Reporting-Repository — eine
   Read-Model-Abhängigkeit im Schreibpfad des Kerns, obwohl die
   Tagesabschluss-Aggregation eine Domänenberechnung über das Kassenjournal
   ist.
3. **DSFinV-K liegt fälschlich in der Domain-Schicht.** Der Mapper
   importiert vier Domain-Pakete und übersetzt Events in ein amtliches
   Exportformat — ein Downstream-Mapper der Fiskalisierung, kein
   Domänenkern.
4. **Das settings-Modul ist ein Sammelbecken.** Betreiber-Stammdaten,
   TSE-Konfiguration/-Setup/-Status und Kassenidentität teilen sich ein
   Domain-Paket, ein Repository und im Frontend eine 14-Methoden-Klasse über
   drei Kontexte hinweg.
5. **Kleinere Inkonsistenzen:** Fachmodule mit englischen Namen (product,
   table) neben deutschen (kasse, direktverkauf), obwohl die Ubiquitous
   Language deutsche Fachbegriffe vorschreibt und die DB-Tabellen längst
   deutsch heißen; duplizierte Repo-Konstruktion in den Wiring-Dateien; ein
   Cross-Modul-Adapter-Shim auf api-Ebene; im Frontend liegen zwei
   admin-exklusive Backend-Klassen im geteilten lib-Ordner, und die
   Kassensitzungs-Seite bündelt drei Formular-Sektionen in einer Datei.

Für Entwickler heißt das: Die Architektur ist im Dateibaum nicht ablesbar,
Kontextgrenzen werden nur durch Disziplin gehalten, und die Dokumentation
widerspricht dem Code (etwa die Aussage, Bondruck sei „kein eigenständiger
Bounded Context, sondern eine Policy" — bei zwei eigenen Tabellen, vier
API-Modulen (bondruck, druckauftrag, druckstation, relay) plus dem
Kassenbeleg-Command in api/table, eigener Admin-UI und einem externen
Relay-Prozess).

## Solution

Der Modulschnitt folgt künftig den Kontextgrenzen, und die Kontexte werden
als Ordner-Ebene sichtbar:

1. **Backend-Kontext-Ordner.** `api/` gliedert sich in `kasse/`
   (Tischgeschäft, Direktverkauf, Kassenführung), `fiskal/` (Signatur, Setup,
   Export, DSFinV-K), `druck/` (Bondruck, Beleg, Auftrag, Station, Relay),
   `stammdaten/` (Produkt, Tisch, User, Betreiber) und `reporting/`.
   Auth, Middleware, Helper und Health bleiben als generische Infrastruktur
   auf oberster Ebene. Architektur = Dateibaum.
2. **`api/table` wird entlang der Kontexte aufgelöst:** Das Tischgeschäft
   (Bestellen, Ausgabe, Kassieren, Storno, Umbuchung samt Tisch-Queries)
   zieht als Kernmodul in den Kasse-Kontext; Tisch-CRUD und Favoriten in die
   Stammdaten; der Kassenbeleg in den Druck-Kontext. Jeder Command behält
   einen konstanten, kleinen Dependency-Satz.
3. **Der Kassenabschluss wird reporting-frei.** Eine reine Domänenfunktion
   berechnet die Tagesabschluss-Summen per Replay aller Events der
   Kassensitzung; das Reporting bleibt reiner Lesekontext.
4. **Fachmodule heißen deutsch.** product→produkt, table→tisch (Domain,
   Repository, api-Module). `user` bleibt englisch (dokumentierter
   Infrastruktur-Begriff), ebenso Auth/Middleware/Config.
5. **settings wird aufgespalten:** Betreiber in die Stammdaten;
   TSE-Konfiguration, TSE-Stammdaten und Kassenidentität in den
   Fiskal-Kontext. Das Frontend teilt die Sammelklasse analog in
   TSE- und Betreiber-Backend-Klasse und verschiebt admin-exklusive Klassen
   aus lib in die Admin-Slices; die Kassensitzungs-Seite wird in ihre
   Sektionen zerlegt.
6. **Die Dokumentation wird dem Code angeglichen:** Kontextkarte mit sechs
   Subdomänen, korrigierte Bondruck-Aussage, Paket-Namenskonventionen und
   neue Modulbegriffe in der Ubiquitous Language, angepasste
   applyTo-Muster der Instructions-Dateien.

Nach außen ändert sich nichts: API-Pfade, JSON-Formate, DB-Schema (bis auf
keine) und UI bleiben identisch. Der Umbau ist für Nutzer unsichtbar und für
Entwickler die Landkarte, die bisher fehlte.

## User Stories

1. Als Entwickler möchte ich die Bounded Contexts im Dateibaum ablesen
   können, damit ich neue Features ohne Architektur-Archäologie im richtigen
   Modul beginne.
2. Als Entwickler möchte ich, dass jedes api-Modul genau einem Kontext
   dient, damit Command-Structs einen konstanten, kleinen Dependency-Satz
   haben und partielle Injektion verschwindet.
3. Als Entwickler möchte ich das Tischgeschäft (Bestellen, Kassieren,
   Ausgabe, Storno, Umbuchung) als eigenständiges Kernmodul vorfinden, damit
   Kassenlogik nicht mit Stammdaten-CRUD vermischt ist.
4. Als Entwickler möchte ich Tisch-CRUD und Favoriten bei den Stammdaten
   finden, wo sie laut Kontextkarte hingehören.
5. Als Entwickler möchte ich den Kassenbeleg im Druck-Kontext pflegen, damit
   Beleg-Rendering und Druck-Infrastruktur beieinander liegen.
6. Als Entwickler möchte ich, dass der Kassenabschluss keine
   Reporting-Abhängigkeit trägt, damit der Schreibpfad des Kerns nur von
   Kasse-Bausteinen abhängt.
7. Als Entwickler möchte ich die Tagesabschluss-Summen als reine, isoliert
   testbare Domänenfunktion haben, damit Z-Bon-Logik ohne Repository-Mocks
   prüfbar ist.
8. Als Entwickler möchte ich alle fiskalischen Bausteine (Signatur-Outbox,
   Worker, Setup, Export, DSFinV-K-Mapper) in einem Fiskal-Kontext finden,
   damit Compliance-Änderungen einen klaren Ort haben.
9. Als Entwickler möchte ich, dass der DSFinV-K-Mapper nicht mehr als
   Domain-Paket firmiert, damit die Domain-Schicht nur echte Domänenlogik
   enthält.
10. Als Entwickler möchte ich deutsche Modulnamen für Fachliches (produkt,
    tisch), damit Code, DB-Tabellen und Ubiquitous Language dieselbe Sprache
    sprechen.
11. Als Entwickler möchte ich Betreiber-Stammdaten und TSE-Konfiguration in
    getrennten Modulen pflegen, damit Stammdaten- und Compliance-Änderungen
    sich nicht berühren.
12. Als Entwickler möchte ich, dass Repositories und Signaturbausteine an
    einer Stelle konstruiert werden, damit die Wiring-Dateien keine
    duplizierte Konstruktion tragen.
13. Als Entwickler möchte ich, dass der Adapter-Shim zwischen Tisch- und
    Bondruck-Modul entfällt, damit Modulgrenzen ohne Übersetzungsschicht
    funktionieren.
14. Als Frontend-Entwickler möchte ich in lib nur Generisches finden
    (HTTP-Client, Auth, Utils), damit bereichs-spezifischer Code in seinem
    Bereich liegt.
15. Als Frontend-Entwickler möchte ich getrennte Backend-Klassen für TSE und
    Betreiber, damit nicht jede Einstellungs-Änderung durch eine
    14-Methoden-Sammelklasse führt.
16. Als Frontend-Entwickler möchte ich die Kassensitzungs-Seite in
    Sektions-Dateien vorfinden, damit Eröffnung, Geldtransit und Abschluss
    unabhängig les- und änderbar sind.
17. Als neuer Mitwirkender möchte ich eine Kontextkarte im Handbuch, die dem
    Code entspricht, damit ich der Dokumentation vertrauen kann.
18. Als Agent (Copilot) möchte ich applyTo-Muster, die auf die neuen Pfade
    zeigen, damit bereichs-spezifische Instructions weiterhin greifen.
19. Als Reviewer möchte ich, dass der Umbau keine Verhaltensänderung
    enthält, damit ich Struktur- von Logikänderungen getrennt prüfen kann.
20. Als Admin möchte ich von alledem nichts bemerken: gleiche URLs, gleiche
    Oberfläche, gleiche Daten.

## Implementation Decisions

Backend: Kontext-Ordner

- `api/` erhält eine Kontext-Ebene; darunter behalten die Module den
  bewährten Schnitt `{application,http}`:
  - **kasse/** (Core): `tischgeschaeft` (Bestellen, Ausgabe bestätigen,
    Zahlung kassieren, Stornierung, Umbuchung, Tisch-Queries wie
    Tischübersicht, Tisch-State, Historie, Meine Tische),
    `direktverkauf` (Umzug unverändert), `kassenfuehrung` (Kassensitzung
    eröffnen, Geldtransit, Kassenabschluss, Kassenbestand; heute api/kasse).
  - **fiskal/** (Supporting): `signatur` (Signaturauftrags-Verwaltung,
    Signaturstatus, Störungsprotokoll — die Bausteine aus der Outbox-PRD),
    `setup` (TSE-Einrichtung, -Übernahme, -Status, Verbindungstest; heute in
    api/settings), `export` (DSFinV-K-Orchestrierung), `dsfinvk`
    (Mapper und Archiv-Builder; Umzug aus der Domain-Schicht, denn er ist
    Downstream-Mapper gegen das Event-Schema, keine Domänenlogik).
  - **druck/** (Supporting): `bondruck` (Arbeitsbon-Policy und ESC/POS),
    `beleg` (Kassenbeleg-Command aus api/table; bezieht TSE-Daten über den
    Signaturstatus aus der Outbox-PRD), `auftrag`
    (Druckauftrags-Verwaltung), `station` (Druckstationen-Konfiguration),
    `relay` (Poll- und Ergebnis-Endpunkte).
  - **stammdaten/** (Supporting): `produkt` (rename product), `tisch`
    (Tisch-CRUD aus api/table plus Favoriten — die Favoriten sind eine
    CRUD-Relation Benutzer↔Tisch und gehören zur Tisch-Verwaltung), `user`
    (bleibt englisch, dokumentierter Infrastruktur-Begriff), `betreiber`
    (aus settings).
  - **reporting/** bleibt unverändert (Read Models).
  - `auth`, `middleware`, `helper`, `health` bleiben als generische
    Infrastruktur direkt unter `api/`.
- Der Adapter-Shim zwischen Tisch- und Bondruck-Modul entfällt: Das
  Druckstation-Interface wird vom Konsumenten definiert (Consumer-Interface
  wie überall sonst), das Repository erfüllt es direkt.
- Die Wiring-Dateien (admin, service, serviceleitung, auth, relay) bleiben
  als Bereichs-Komposition erhalten, konstruieren Repositories und geteilte
  Bausteine aber nur noch an einer Stelle (ein gemeinsamer
  Abhängigkeiten-Bausatz, der an die Bereichs-Konstruktoren gereicht wird).

Backend: Domain-Schicht

- `domain/kasse` erhält die Tagesabschluss-Aggregation als reine Funktion:
  Eingabe alle Events der Kassensitzung (Bestellungen, Zahlungen,
  Warenrücknahmen, Korrekturen, Umbuchungen, Direktverkäufe samt Storni,
  Geldtransit, Differenzbuchung), Ausgabe die Summen, die das
  Tagesabschluss-Event braucht. Der Kassenabschluss-Command verliert damit
  die Reporting-Abhängigkeit; das Kassenjournal-Repository liefert die
  Events der Kassensitzung.
- **Bewusste Logik-Duplikation zur SQL-Schicht:** Die Reporting-Queries
  behalten ihre `kj_extract_*`-SQL-Funktionen (mengeneffiziente Reads über
  viele Kassensitzungen); die neue Domänenfunktion repliziert dieselbe
  Betrags-Extraktion in Go für den Schreibpfad einer einzelnen Sitzung.
  Der Äquivalenztest (siehe Testing Decisions) koppelt beide
  Implementierungen dauerhaft; jeder neue geldrelevante Event-Typ muss
  beide Stellen nachziehen — das ist der Preis dafür, dass das Reporting
  reiner Lesekontext bleibt.
- `domain/table` → `domain/tisch` (Tisch-Entität und
  AktiverTisch-Read-Models), `domain/product` → `domain/produkt`.
- `domain/settings` entfällt: Betreiber → neues `domain/betreiber`;
  TSE-Konfiguration, TSE-Stammdaten und Kassenidentität → `domain/tse`
  (die Kassenidentität ist die fiskalische Identität der Kasse für Beleg
  und Export und gehört zum Fiskal-Kontext).
- `domain/dsfinvk` zieht als Mapper-Paket in den Fiskal-Kontext der
  api-Schicht um (siehe oben) — die Domain-Schicht enthält danach nur noch
  Domänenlogik.
- `domain/user`, `domain/jwt`, `domain/event`, `domain/steuer`,
  `domain/reporting`, `domain/druckstation` bleiben unverändert.

Backend: Repository-Schicht

- Die Repository-Schicht bleibt flach (`repository/<x>_repo`); Umbenennungen
  für Sprachkonsistenz: product_repo → produkt_repo, table_repo →
  tisch_repo. settings_repo wird aufgeteilt: Betreiber-Zugriffe in ein
  neues betreiber_repo, TSE-Konfiguration/-Stammdaten/Kassenidentität in
  das bestehende tse_repo.
- Die sqlc-Query-Datei `tables.sql` wird zu `tische.sql` umbenannt (letzter
  englischer Ausreißer; die Query-Namen darin sind bereits deutsch). Danach
  `make sqlc`; der generierte Dateiname in `sqlc/dbgen/` folgt automatisch.
- Keine Schema-Änderungen: Die Tabellen (betreiber, tse_konfiguration,
  tse_stammdaten, kassenidentitaet, produkte, tische, users, …) sind
  bereits korrekt geschnitten und benannt.

Frontend

- lib enthält nur noch Generisches: HTTP-Client (BackendClient), Auth,
  Credential-Schemas (identity), Download, Utils, Fehlermeldungen,
  Arbeitsmodus. Die admin-exklusiven Klassen ziehen in ihre Slices:
  Druckstation-Backend zu den Druck-Einstellungen; die
  Einstellungen-Sammelklasse wird aufgeteilt in eine TSE-Backend-Klasse
  (Konfiguration, Setup, Status, Signaturaufträge) im TSE-Slice und eine
  Betreiber-Backend-Klasse (Betreiber, Kassenidentität) im Finanzamt-Slice.
- Die Kassensitzungs-Seite wird in ihre drei Sektionen (Eröffnung,
  Geldtransit, Abschluss) als eigene Dateien zerlegt; Verhalten und
  Erscheinungsbild bleiben identisch.
- Keine Änderungen an Routen, API-Aufrufen, Typen-Schnitt (admin/service
  behalten bewusst getrennte Modelle) oder UI-Texten.

API-Verträge und Daten

- Alle Endpunkt-Pfade, Request-/Response-Formate und JSON-Keys bleiben
  unverändert. Der Umbau ist ein reines internes Refactoring; das Frontend
  merkt davon nichts (außer den eigenen Datei-Umzügen).
- Keine Datenbank-Migrationen, keine Event-Format-Änderungen.

Dokumentation

- handbuch.md: Kontextübersicht auf sechs Subdomänen erweitert (Kasse als
  Core; Fiskalisierung, Druck/Ausgabe, Stammdaten, Reporting als
  Supporting; Auth als Generic) samt Beziehungen; die Aussage „Bondruck ist
  kein eigenständiger Bounded Context, sondern eine Policy" wird gestrichen
  und durch die Druck-Kontext-Beschreibung ersetzt; Schichten- und
  Modulbeschreibung auf die Kontext-Ordner aktualisiert; die
  TSE-Architektur-Abschnitte verweisen auf den Fiskal-Kontext (Basis ist
  der Stand nach der Outbox-PRD).
- language.md: Namenskonventionen um Go-Paketnamen ergänzt (deutsch für
  Fachmodule, englisch für Infrastruktur; `user` als dokumentierte
  Ausnahme); neue Modulbegriffe Tischgeschäft und Kassenführung als
  Fachbegriffe aufgenommen (deckungsgleich mit den im Handbuch etablierten
  „tischbezogenen" und „kassenführungsbezogenen Vorgängen"); veraltete
  Paket-Hinweise (etwa zu domain/table) aktualisiert.
- .github/instructions: applyTo-Muster auf die neuen Pfade angepasst
  (insbesondere Event-Sourcing-Instructions, die auf domain/table und
  domain/kasse zeigen); zusätzlich die Codebeispiele in
  backend.instructions.md, die alte Paketnamen tragen (`product.Produkt`,
  `product.Kategorie`), auf die neuen Paketnamen aktualisiert.
- AGENTS.md-Bereichsbeschreibung: Pfad-Referenzen prüfen und auf die
  Kontext-Ordner anpassen.

## Testing Decisions

Gute Tests prüfen ausschließlich Außenverhalten über die öffentliche
Schnittstelle (Eingaben, Ergebnisse, persistierte Effekte), keine
Implementierungsdetails. Da dieser Umbau fast ausschließlich Struktur
verschiebt, gilt: **Bestehende Tests ziehen mit ihren Modulen um und müssen
unverändert grün bleiben** — sie sind der Regressionsschutz des Umbaus.
`make verify` (inklusive Seed-Integrationstest) sichert das Gesamtsystem.

Neue Tests entstehen nur für neue Logik:

- **Tagesabschluss-Aggregation** (reine Domänenfunktion): tabellengetrieben
  über Event-Sequenzen — leere Kassensitzung, nur Bestellungen ohne
  Zahlung, Zahlungen über mehrere Steuersätze, Warenrücknahme, geldneutrale
  Korrektur und Umbuchung (dürfen die Summen nicht verändern),
  Direktverkauf samt Storno, Geldtransit in beide Richtungen,
  Differenzbuchung. Referenzwerte deckungsgleich mit den heutigen
  Z-Bon-Zahlen (Äquivalenz zum Reporting-Ergebnis für dieselbe
  Kassensitzung). Vorbild: die tabellengetriebenen Tests in domain/kasse
  (Storno-Aufteilung, Offene-Arbeit).
- Der Kassenabschluss-Command-Test verliert das Reporting-Mock und prüft
  stattdessen, dass die Event-Daten des Tagesabschlusses aus den
  Journal-Events berechnet werden.

Ausdrücklich keine neuen Testarten für die Umzüge selbst (kein
Architektur-/Import-Grenzen-Test — Entscheidung aus der Klärungsrunde).

## Out of Scope

- Jegliche Verhaltensänderung: keine neuen Features, keine geänderten
  API-Pfade, JSON-Formate, Fehlercodes, UI-Texte oder Abläufe.
- Datenbank-Schema- oder Event-Format-Änderungen; Datenmigrationen.
- Umbenennung der DB-Tabelle `tisch_sessions`: geprüft und bewusst
  belassen — der Name entspricht dem UL-Begriff Tisch-Session
  (language.md), abgegrenzt von der Kassensitzung.
- Die TSE-Signatur-Outbox selbst (eigene PRD, Voraussetzung dieser PRD).
- Umbenennung von `user`/Auth-Infrastruktur ins Deutsche (dokumentierte
  Ausnahme der Ubiquitous Language).
- Ein automatisierter Architektur-Test für Modulgrenzen (bewusst verworfen;
  die Kontext-Ordner selbst machen Grenzverletzungen im Review sichtbar).
- Umbau der Satelliten-Module (resolver, reverse-proxy, windows/relay,
  website) und der Frontend-Routen- oder Typen-Struktur (admin/service
  behalten getrennte Modelle — das ist gewollt).
- Scale-out-, Deployment- oder Infrastruktur-Änderungen.

## Further Notes

- **Reihenfolge:** Diese PRD setzt die Outbox-PRD voraus, weil deren Umbau
  die TSE-Verflechtung des Kerns auflöst (Events ohne TSE-Felder, ein
  Schreibprimitiv mit Outbox-Anhängen, zentrale fiskalische Projektion).
  Erst dadurch werden die Modul-Umzüge dieser PRD zu reinen Verschiebungen
  ohne Logikänderung.
- **Leitprinzip des Schnitts:** Modulgrenze = Kontextgrenze = Begriff der
  Ubiquitous Language. Der Kasse-Kern entscheidet über Geschäftsvorfälle;
  Fiskalisierung und Druck reagieren auf committete Events über ihre
  Outboxen (Signaturauftrag, Druckauftrag) — zwei Outboxen, ein Muster,
  bewusst ohne gemeinsame Abstraktion, aber mit parallelen Namen, Status
  und Admin-Aktionen.
- **Klassifikation:** Kasse ist Core Domain (Differenzierung: lückenlose
  Transparenz für Vereine); Fiskalisierung ist Supporting (gesetzlich
  zwingend, nicht differenzierend); Druck, Stammdaten, Reporting sind
  Supporting; Auth ist Generic. Diese Einordnung steuert, wo Sorgfalt und
  Eigenbau lohnen und wo Standardlösungen genügen.
- **Bewusste Nicht-Ziele der Architektur:** kein Microservices-Schnitt,
  kein Message-Broker, keine generische Outbox-Bibliothek. Bei 5–30
  gleichzeitigen Nutzern und Single-Prozess-Deployment ist der modulare
  Monolith das Optimum; die Arbeit liegt im Schnitt der Module, nicht in
  ihrer Verteilung.
- **Favoriten über Kontextgrenzen:** Die Favoriten-Commands ziehen zu
  stammdaten/tisch, aber die Tischgeschäft-Queries („Meine Tische",
  AktiverTischMitFavorit) lesen den Favoriten-Status weiterhin mit — das
  favorit_repo wird damit aus beiden Kontexten gelesen. Für Read Models
  legitim, aber eine bewusste Entscheidung, keine versehentliche.
- **Finanzamt-Seite mit zwei Backend-Klassen:** Die FinanzamtPage bündelt
  Betreiber-, Kassenidentitäts- und TSE-Sektionen und nutzt nach der
  Aufspaltung beide neuen Backend-Klassen — für die TSE-Sektionen als
  Cross-Slice-Import aus dem tse-Slice. Gewollt, da die UI unverändert
  bleibt.
- **Tagesabschluss und Same-Command-Events:** Der Abschluss-Command erzeugt
  Kassensturz- und ggf. Differenz-Event im selben Vorgang vor dem
  Tagesabschluss. Die reine Domänenfunktion muss diese noch nicht
  persistierten Events zusätzlich zu den Journal-Events als Eingabe
  erhalten (in-memory anhängen), sonst fehlt die Differenzbuchung in den
  Summen.
- **Risiko und Gegenmittel:** Das Hauptrisiko sind versehentliche
  Logikänderungen während der Umzüge. Gegenmittel: strikt mechanische
  Verschiebungen (Move + Import-Anpassung) getrennt von den wenigen echten
  Änderungen (Tagesabschluss-Funktion, settings-Aufspaltung,
  Deps-Bausatz), unveränderte Tests als Regressionsnetz und `make verify`
  nach jedem Schritt.
