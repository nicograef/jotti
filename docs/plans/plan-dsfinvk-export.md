# Plan: DSFinV-K-Export (F-04)

> Source PRD: [../prds/prd-dsfinvk-export.md](../prds/prd-dsfinvk-export.md)

## Goal

Ein Admin kann für eine gewählte Kassensitzung ein vollständiges, IDEA-lesbares
DSFinV-K-Archiv (ZIP mit CSV-Dateien, `index.xml`, `gdpdu-01-09-2004.dtd`)
herunterladen, ohne Internet und ohne Steuerberater. Der Export liest nur
vorhandene Daten (Kassenjournal, Stammdaten, TSE-Signaturen) und transformiert
sie seiteneffektfrei. Mit erledigt werden die zwei offenen F-06-Kriterien
(Abrechnungskreis je Bon) und die Voraussetzung, dass die TSE-Stammdaten
(Algorithmus, Public Key, Zertifikat) für die `tse.csv` lokal vorliegen.

## Architectural decisions

Durable decisions, die über alle Phasen gelten:

- **Route**: `GET /admin/export/dsfinvk` mit optionalem Query-Parameter
  `kassensitzung=N` (Default: aktuelle bzw. jüngste Sitzung). Innerhalb von
  `NewAdminApi` als `/export/dsfinvk` registriert; der `/admin/`-Mount erzwingt
  bereits Rolle `admin` per JWT-Middleware ([app/app.go:53-55](../../backend/app/app.go#L53-L55)),
  also liefert die Plattform den `403` ohne Admin-Rolle. Der Handler liefert
  `200` mit `application/zip` und `Content-Disposition`-Dateiname (Seriennummer,
  Kassensitzung, Zeitstempel), `404` bei unbekannter Sitzung, `400` bei
  ungültigem Parameter.
- **Schema**: neue Singleton-Tabelle `tse_stammdaten` (Signaturalgorithmus,
  Public Key, Zertifikat, Log-Time-Format, Versionsangaben), `id = 1 CHECK`, leer
  vorbefüllt nach dem Muster von `tse_konfiguration`
  ([migration:419-440](../../database/migrations/01_initial.up.sql#L419-L440)).
  Keine Änderung am Kassenjournal. Zusätzlich eine neue Lese-Query
  `ReadEventsByKassensitzung(kassensitzungNr)`, die alle Events einer Sitzung
  (Kassensitzungs-, Tisch-Session- und Direktverkauf-Streams) nach `id`
  geordnet liefert; eine solche Query existiert heute noch nicht.
- **Modulschnitt (Deep Modules, rein/dünn wie im PRD)**: ein reines
  `dsfinvk`-Mapper-Paket (Events plus Stammdaten-Snapshot zu typisierten
  Zeilen-Kollektionen je Tabelle), ein generischer CSV-Serializer, ein
  `index.xml`/DTD-Generator (statische DTD per `go:embed`), ein dünner
  ZIP-Packer, ein Export-Orchestrator (App-Service) und der Admin-Handler. Der
  fiskalisch heikle Teil liegt vollständig im Mapper und ist über Golden-Files
  prüfbar.
- **Schlüsselfelder**: `Z_KASSE_ID` = Kassen-Seriennummer (UUID aus
  `kassenidentitaet`), `Z_NR` = `kassensitzung_nr`, `Z_ERSTELLUNG` aus dem
  `tagesabschluss-erstellt:v1`-Event (geschlossene Sitzung) oder als
  Exportzeitpunkt synthetisiert (offene Sitzung). `ABRECHNUNGSKREIS` aus dem
  Subject abgeleitet (Tisch-ID zu Tischname).
- **Beträge**: intern durchgehend Cent (`int`) aus dem Journal, erst der
  Serializer stellt sie als Dezimalzahl mit Punkt und zwei Nachkommastellen dar.
- **TSE-Daten-Vereinigung**: pro Vorgang wird `TSEData` aus dem Event-Payload
  genutzt; fehlt sie (Persistenz im TSE-Ausfall), greift die
  `tse_signaturen`-Seitentabelle per `tx_id`
  ([repo.go:452-470](../../backend/repository/kassenjournal_repo/repo.go#L452-L470)).
- **Versionsstring**: DSFinV-K v2.5, konfigurierbar gehalten (in `index.xml`
  bzw. `cashregister.csv`), da die Tabellenstruktur seit v2.0 stabil ist.
- **Drittsystem-Grenze**: das `SetupClient`-Interface bekommt eine lesende
  Operation auf der fiskaly-TSS-Ressource (`signature_algorithm`, `public_key`,
  `certificate`, `log_time_format`); diese Felder liefert die fiskaly SIGN DE API
  auf der TSS-Ressource (verifiziert gegen `temp/fiskaly_sign_de_api_spec.json`).

## Inventory

- [backend/domain/kasse/tisch_session_events.go:14-20](../../backend/domain/kasse/tisch_session_events.go#L14-L20),
  [kassensitzung_events.go:12-18](../../backend/domain/kasse/kassensitzung_events.go#L12-L18),
  [direktverkauf_events.go:12-15](../../backend/domain/kasse/direktverkauf_events.go#L12-L15):
  alle Event-Typen und ihre `*V1Data`-Structs (Quelle des Mappings).
- [backend/domain/kasse/tse_data.go:10-19](../../backend/domain/kasse/tse_data.go#L10-L19):
  `TSEData` (Transaktionsnummer, Signaturzähler, Signatur, LogTimes, ProcessType,
  QRCodeData) im Event-Payload.
- [backend/domain/kasse/bestellung.go:12-40](../../backend/domain/kasse/bestellung.go#L12-L40):
  `Position`/`PositionEventData` (ProduktName, Steuersatz, Einzelpreis, Menge).
- [backend/domain/kasse/subject.go:36-66](../../backend/domain/kasse/subject.go#L36-L66):
  `ParseTischIDFromSubject`/`ParseZNrFromSubject` (Quelle des Abrechnungskreises).
- [backend/domain/steuer/steuer.go:44-63](../../backend/domain/steuer/steuer.go#L44-L63):
  `Aufteilen` inkl. `kombi`-Entfaltung 70/7 % und 30/19 %; Golden-Werte in
  [steuer_test.go](../../backend/domain/steuer/steuer_test.go).
- [backend/domain/settings/betreiber.go:10-18](../../backend/domain/settings/betreiber.go#L10-L18),
  [kassenidentitaet.go:9-12](../../backend/domain/settings/kassenidentitaet.go#L9-L12),
  [tse_konfiguration.go:11-17](../../backend/domain/settings/tse_konfiguration.go#L11-L17):
  Stammdaten für `location.csv`, `cashregister.csv`, `tse.csv`.
- [backend/domain/tse/setup.go:66-97](../../backend/domain/tse/setup.go#L66-L97):
  `SetupClient`-Interface; `TSSInfo` ([:36-40](../../backend/domain/tse/setup.go#L36-L40))
  trägt heute nur `ID`/`State`, daher braucht die Stammdaten-Persistenz eine neue
  Leseoperation.
- [backend/repository/tse_repo/fiskaly_setup.go:63-66](../../backend/repository/tse_repo/fiskaly_setup.go#L63-L66):
  `tssDetailResponse` (heute nur `admin_puk`/`state`); wird um die
  Stammdaten-Felder erweitert.
- [backend/api/settings/application/setup.go:129-142](../../backend/api/settings/application/setup.go#L129-L142)
  (Neuanlage-Speichern), [:306-317](../../backend/api/settings/application/setup.go#L306-L317)
  (Übernahme-Speichern), PUK-Reset [:280-294](../../backend/api/settings/application/setup.go#L280-L294):
  der gemeinsame Speicher-Schritt, an den die Stammdaten-Persistenz hängt.
- [backend/api/admin.go:94-102](../../backend/api/admin.go#L94-L102)
  (Reporting-Registrierung), [:151-180](../../backend/api/admin.go#L151-L180)
  (Settings-/Setup-Handler mit `NewTSESetupClient`-Factory): Einhängepunkt für
  Endpoint und Stammdaten-Hook.
- [backend/repository/kassenjournal_repo/repo.go:419-447](../../backend/repository/kassenjournal_repo/repo.go#L419-L447):
  `ReadEventsBySubject`/`ReadDirektverkaufEvents` als Vorlage für die neue
  Sitzungs-weite Query; [sqlc/queries/kassenjournal.sql](../../backend/sqlc/queries/kassenjournal.sql).
- [backend/repository/kassensitzungen_repo/repo.go:28-59](../../backend/repository/kassensitzungen_repo/repo.go#L28-L59):
  `GetOffeneKassensitzung`/`GetAllKassensitzungen` für Default und Selektor.
- [backend/sqlc/queries/tables.sql](../../backend/sqlc/queries/tables.sql) `GetTisch`:
  Tischname für den Abrechnungskreis.
- [backend/sqlc/queries/tse_signaturen.sql](../../backend/sqlc/queries/tse_signaturen.sql):
  `GetTSESignaturByTxID` (Vereinigungspunkt der Nachsignier-Seitentabelle).
- [database/migrations/01_initial.up.sql:359-391](../../database/migrations/01_initial.up.sql#L359-L391):
  Singleton-/Write-Guard-Muster (kassenidentitaet) für die neue Tabelle.
- Reporting-Prior-Art für Aggregations- und Steuer-Golden-Tests:
  [api/reporting/application/query.go](../../backend/api/reporting/application/query.go),
  [domain/reporting/reporting.go](../../backend/domain/reporting/reporting.go).
- Frontend-Reporting-Bereich: [frontend/src/admin/](../../frontend/src/admin/) und
  die Reporting-Lib-Aufrufe (Einhängepunkt für Selektor und Download-Button).
- Build/Test: `make test` (Unit), `make check` (Komplettprüfung ohne DB),
  `make sqlc` (nach SQL-Änderungen), `make test-integration` für `//go:build
  integration` env-gated fiskaly-Tests.

## Resolved decisions

- **Umfang**: Backend plus Admin-UI. Der Reporting-Bereich bekommt einen
  Kassensitzung-Selektor und einen Download-Button (Self-Service für
  nicht-technische Vereins-Admins, US 15).
- **Offene Sitzung**: exportierbar. Für eine noch nicht abgeschlossene Sitzung
  wird `cashpointclosing.csv` stand-jetzt synthetisiert (`Z_ERSTELLUNG` =
  Exportzeitpunkt, Summen aus allen Events bis jetzt); die Sitzung bleibt offen.
  Das deckt die Kassen-Nachschau im laufenden Betrieb ab (US 2).
- **Default-Sitzung**: ist eine Sitzung offen, ist sie vorausgewählt; sonst die
  jüngste abgeschlossene.
- **Zielversion**: v2.5, Versionsstring konfigurierbar.
- **Ein Export = genau eine Kassensitzung** (`Z_NR` = `kassensitzung_nr`).
- **Nicht zutreffende Tabellen** (`slaves.csv`, `pa.csv`) werden weggelassen und
  nicht in `index.xml` deklariert.
- **Tests**: Golden-File-Tests für den Mapper, Formatregel-Tests für den
  Serializer, Descriptor-Test für den `index.xml`/DTD-Generator. Kein dedizierter
  Integrationstest für Orchestrator und Handler (bewusste PRD-Entscheidung).

## Open questions / Risks

- **Reihenfolge der Stammdaten-Persistenz**: Phase 1 berührt die gerade
  abgeschlossenen Setup-Pfade ([plan-tse-setup-recovery.md](plan-tse-setup-recovery.md):
  `setup.go`, `SetupClient`, `fiskaly_setup.go`) und hängt bewusst am gemeinsamen
  Speicher-Schritt, nicht am Anlage-Lebenszyklus. Sonst bekämen per Übernahme
  onboardete Instanzen ein unvollständiges `tse.csv`. Zuerst umsetzen, solange der
  Setup-Code frisch ist.
- **Eine LIVE-TSS pro Kasse**: die harte LIVE-Sperre des Recovery-Plans sichert
  genau eine Stamm_TSE-Zeile im Produktivfall. TEST kann mehrere TSS sammeln, ist
  aber steuerlich ungültig.
- **Golden-Files als Vertrag**: bei DSFinV-K-Detailabweichungen (Spaltenreihenfolge,
  GV-Typ-Codes) ändern sich viele Golden-Dateien gleichzeitig. Anhang C/E der
  Spezifikation bleibt die maßgebliche Quelle; Abweichungen werden im Mapper
  zentralisiert, nicht in den Tests dupliziert.
- **Keine CI-Validierung gegen IDEA** (PRD Out of Scope). Eine manuelle
  Stichprobe gegen ein echtes Prüftool bleibt wünschenswert.

---

## Phase 1: TSE-Stammdaten-Persistenz

**User stories**: 24, 25

### Context

- [backend/domain/tse/setup.go:66-97](../../backend/domain/tse/setup.go#L66-L97):
  `SetupClient`; eine neue Leseoperation für die TSS-Stammdaten kommt hinzu.
- [backend/repository/tse_repo/fiskaly_setup.go:63-66](../../backend/repository/tse_repo/fiskaly_setup.go#L63-L66):
  `tssDetailResponse` wird um `signature_algorithm`, `public_key`, `certificate`,
  `log_time_format` erweitert.
- [backend/api/settings/application/setup.go:129-142](../../backend/api/settings/application/setup.go#L129-L142),
  [:306-317](../../backend/api/settings/application/setup.go#L306-L317),
  [:280-294](../../backend/api/settings/application/setup.go#L280-L294):
  der gemeinsame Speicher-Schritt aller Einrichtungspfade.
- [database/migrations/01_initial.up.sql:419-440](../../database/migrations/01_initial.up.sql#L419-L440):
  Singleton-Muster (`tse_konfiguration`).

### What to build

Eine neue Singleton-Tabelle `tse_stammdaten` mit Signaturalgorithmus, Public Key,
Zertifikat, Log-Time-Format und Versionsangaben, leer vorbefüllt. Das
`SetupClient`-Interface bekommt eine lesende Operation, die die
Stammdaten-Felder der TSS-Ressource von fiskaly liest; der Setup-Orchestrator
ruft sie am gemeinsamen Speicher-Schritt auf (derselbe atomare Abschluss, den
Neuanlage, Übernahme und PUK-Reset durchlaufen) und persistiert die Stammdaten
zusammen mit der `tse_konfiguration`. Schlägt der Stammdaten-Abruf fehl, wird der
Setup-Erfolg nicht zurückgenommen, aber der Fehlzustand verständlich protokolliert
(die Stammdaten lassen sich beim nächsten Verbinden nachziehen).

### Acceptance criteria

- [ ] Nach erfolgreicher Neuanlage (`RichteTSEEin`) liegen Algorithmus, Public Key,
      Zertifikat und Log-Time-Format in `tse_stammdaten`.
- [ ] Nach einer Übernahme (`UebernimmTSE`, inkl. PIN-freier F8-Übernahme und
      PUK-Reset) sind dieselben Stammdaten vollständig persistiert.
- [ ] Die Persistenz hängt am gemeinsamen Speicher-Schritt, nicht am
      Anlage-Lebenszyklus (per Unit-Test gegen den `SetupClient`-Fake belegt).
- [ ] Die fiskaly-Leseoperation ist gegen den Fake-Server kontraktgetestet
      (Pfad/Body), ein env-gated `//go:build integration`-Test deckt den echten
      Abruf ab.
- [ ] `make sqlc` aktuell, `make check` grün.

---

## Phase 2: Tracer bullet — minimaler gültiger Export

**User stories**: 1, 2, 4, 5, 6, 7, 8, 9, 10, 12, 15, 16, 17, 19, 22

### Context

- [backend/api/admin.go:94-102](../../backend/api/admin.go#L94-L102): Reporting-/
  Settings-Registrierung; der neue Handler wird analog eingehängt.
- [app/app.go:53-55](../../backend/app/app.go#L53-L55): `/admin/`-Mount erzwingt
  Rolle `admin`.
- [backend/repository/kassenjournal_repo/repo.go:419-447](../../backend/repository/kassenjournal_repo/repo.go#L419-L447):
  Vorlage für die neue Query `ReadEventsByKassensitzung`.
- [backend/repository/kassensitzungen_repo/repo.go:28-59](../../backend/repository/kassensitzungen_repo/repo.go#L28-L59):
  Default-Sitzung (offen bzw. jüngste).
- [backend/domain/steuer/steuer.go:44-63](../../backend/domain/steuer/steuer.go#L44-L63):
  Steueraufteilung für die USt-Zeilen.

### What to build

Die gesamte Pipeline durch alle Schichten für den einfachsten Vorgang (eine
`zahlung-kassiert` ohne Storno, ohne `kombi`): das reine Mapper-Paket erzeugt die
minimal nötigen Zeilen-Kollektionen für ein gültiges Archiv (Stammdaten
`location`, `cashregister`, `vat`, `tse`; `cashpointclosing` stand-jetzt
synthetisiert; `transactions`, `transactions_vat`, `datapayment`, `lines`,
`lines_vat`, `transactions_tse`). Der generische CSV-Serializer schreibt diese
DSFinV-K-konform (Semikolon, CRLF, Header, Dezimalpunkt, Spaltenreihenfolge,
UTF-8). Der `index.xml`/DTD-Generator deklariert genau die vorhandenen Tabellen
und bettet die statische `gdpdu-01-09-2004.dtd` ein. Der ZIP-Packer bündelt alles.
Der Orchestrator lädt Events plus Stammdaten-Snapshot und ruft die Module. Der
Admin-Handler streamt das ZIP unter `/export/dsfinvk` mit sprechendem Dateinamen.
Das Frontend bekommt im Reporting-Bereich einen Download-Button für die aktuelle
Sitzung.

### Acceptance criteria

- [ ] `GET /admin/export/dsfinvk` ohne Parameter liefert für die aktuelle Sitzung
      ein `200 application/zip` mit `Content-Disposition`-Dateiname (Seriennummer,
      Kassensitzung, Zeitstempel).
- [ ] Das Archiv enthält die CSV-Dateien mit offiziellen englischen Kleinschreib-
      Namen, eine `index.xml` und die `gdpdu-01-09-2004.dtd`; `slaves.csv`/`pa.csv`
      fehlen und sind nicht deklariert.
- [ ] Eine unbekannte Sitzung liefert `404`, ein ungültiger Parameter `400`; eine
      leere Sitzung eine verständliche Meldung statt eines defekten Archivs.
- [ ] Golden-Test (einfacher Barverkauf) prüft die erzeugten Zeilen je Tabelle
      inkl. Brutto/Netto/Steuer, `Z_KASSE_ID`, `Z_NR`, `BON_ID` und der
      TSE-Transaktionsdaten.
- [ ] Serializer-Formattest (Semikolon, CRLF, Dezimalpunkt, keine Tausendertrenner,
      Header, Spaltenreihenfolge, Escaping) und `index.xml`-Descriptor-Test grün.
- [ ] Der Download-Button lädt im Reporting-Bereich ein Archiv der aktuellen
      Sitzung; `make check` grün.

---

## Phase 3: Gastro-Tischablauf, Abrechnungskreis (F-06-Abschluss) und Sitzungsauswahl

**User stories**: 3, 13, 14 (teilweise), 23

### Context

- [backend/domain/kasse/subject.go:36-48](../../backend/domain/kasse/subject.go#L36-L48):
  `ParseTischIDFromSubject` für die Ableitung des Abrechnungskreises.
- [backend/sqlc/queries/tables.sql](../../backend/sqlc/queries/tables.sql) `GetTisch`:
  Tischname (z. B. `Tisch 42`).
- [backend/domain/kasse/tisch_session_events.go:24-55](../../backend/domain/kasse/tisch_session_events.go#L24-L55):
  `bestellung-aufgenommen` (Forderungsentstehung) und `zahlung-kassiert`
  (Forderungsauflösung).
- [backend/repository/kassensitzungen_repo/repo.go:14-26](../../backend/repository/kassensitzungen_repo/repo.go#L14-L26):
  `GetAllKassensitzungen` für den Selektor.

### What to build

Der Mapper bildet den gastronomischen Tisch-Ablauf konform ab:
`bestellung-aufgenommen` als Forderungsentstehung (processType `Bestellung-V1`,
kein sofortiger Umsatz), `zahlung-kassiert` als Forderungsauflösung und
Umsatzrealisierung (`Kassenbeleg-V1`). Jeder Bon wird über sein Subject einem
`ABRECHNUNGSKREIS` (Tischname) zugeordnet und in `allocation_groups.csv`
ausgewiesen. Damit sind die zwei offenen F-06-Akzeptanzkriterien prüfbar. Das
Frontend bekommt den Kassensitzung-Selektor (offen plus abgeschlossene), sodass
auch vergangene Betriebstage exportierbar sind; für abgeschlossene Sitzungen kommt
`Z_ERSTELLUNG` aus dem `tagesabschluss-erstellt`-Event.

### Acceptance criteria

- [ ] Bestellung und zugehörige Zahlung erscheinen als getrennte Geschäftsvorfälle
      (Forderungsentstehung bzw. -auflösung), Golden-Test belegt die Trennung.
- [ ] Jeder Bon trägt in `allocation_groups.csv` den korrekten `ABRECHNUNGSKREIS`
      (Tischname); jede TSE-Transaktion ist über ihren Bon zugeordnet.
- [ ] Eine abgeschlossene Sitzung lässt sich per Selektor wählen und exportieren;
      `Z_ERSTELLUNG` stammt aus dem Tagesabschluss-Event.
- [ ] Der Default zeigt die offene Sitzung, sonst die jüngste abgeschlossene.
- [ ] `make check` grün.

---

## Phase 4: Storno und Direktverkauf (Radierverbot)

**User stories**: 11, 20, 21

### Context

- [backend/domain/kasse/tisch_session_events.go:57-71](../../backend/domain/kasse/tisch_session_events.go#L57-L71):
  `stornierung-erteilt` (fette Positionen, `GesamtStornierungCents`).
- [backend/domain/kasse/direktverkauf_events.go:17-53](../../backend/domain/kasse/direktverkauf_events.go#L17-L53):
  `direktverkauf-getaetigt`/`-storniert` mit `VerkaufID`-Referenz.
- [backend/domain/steuer/steuer.go:49-57](../../backend/domain/steuer/steuer.go#L49-L57):
  `kombi`-Entfaltung 70 % zu 7 % und 30 % zu 19 %.

### What to build

Der Mapper bildet Stornierungen als eigene Negativ-Datensätze mit `BON_STORNO` und
`REF_BON_ID` auf den Ursprungsbon ab, mit eigener TSE-Signatur und gleichem
Abrechnungskreis; die Referenz landet in `references.csv`. `kombi`-Positionen
werden in ihre Steueranteile entfaltet (70/7 %, 30/19 %), sodass die
USt-Aufschlüsselung je Zeile und je Bon stimmt. Direktverkäufe werden als eigene
Belege abgebildet (positiver Verkauf und Storno je eigener Beleg mit Referenz);
Direktverkäufe tragen keinen Abrechnungskreis (Feld optional, dokumentiert).

### Acceptance criteria

- [ ] Eine Stornierung erzeugt einen Negativ-Datensatz mit `BON_STORNO` und
      `REF_BON_ID` auf den Ursprung; die Referenz steht in `references.csv`
      (Golden-Test).
- [ ] Eine `kombi`-Position erscheint korrekt in 70/7 % und 30/19 % aufgeteilt in
      `lines_vat.csv` und `transactions_vat.csv` (Golden-Test).
- [ ] Direktverkauf und Direktverkauf-Storno erscheinen als eigene Belege mit
      Referenz (Golden-Tests je Szenario).
- [ ] `make check` grün.

---

## Phase 5: Kassenabschlussmodul und übrige Geschäftsvorfalltypen

**User stories**: 14

### Context

- [backend/domain/kasse/kassensitzung_events.go:12-18](../../backend/domain/kasse/kassensitzung_events.go#L12-L18):
  `geldtransit-gebucht`, `differenz-soll-ist-gebucht`, `tagesabschluss-erstellt`.
- [backend/domain/kasse/kassensitzung_events.go:22-27](../../backend/domain/kasse/kassensitzung_events.go#L22-L27):
  `kassensitzung-eroeffnet` mit `BetragCents` (Anfangsbestand).
- [backend/domain/kasse/tisch_session_events.go:85-97](../../backend/domain/kasse/tisch_session_events.go#L85-L97):
  `auszahlung-geleistet`.
- Prior Art Aggregation: [api/reporting/application/query.go:48-100](../../backend/api/reporting/application/query.go#L48-L100).

### What to build

Das Kassenabschlussmodul aggregiert die Sitzung: `businesscases.csv` (Z_GV_TYP,
Beträge je Geschäftsvorfalltyp nach Steuersätzen), `payment.csv` (Z_Zahlart,
aggregierte Zahlart-Summen) und `cash_per_currency.csv` (Bargeldbestand EUR). Die
übrigen Geschäftsvorfalltypen werden gemäß DSFinV-K Anhang C/E gemappt:
Geldtransit, Differenz Soll/Ist (Kassensturz), Auszahlung und Anfangsbestand (aus
dem Eröffnungs-Event). Trinkgeld und Rückgeld bleiben außen vor (clientseitig,
nicht fiskalisch persistiert).

### Acceptance criteria

- [ ] `businesscases.csv` und `payment.csv` lassen die Tagessumme gegen die
      Einzelbons abgleichen (Golden-Test mit gemischter Sitzung).
- [ ] Geldtransit, Differenz Soll/Ist, Auszahlung und Anfangsbestand erscheinen mit
      korrektem Geschäftsvorfalltyp.
- [ ] `cash_per_currency.csv` weist den EUR-Bestand aus.
- [ ] `make check` grün.

---

## Phase 6: Nachsignier-Robustheit

**User stories**: 18

### Context

- [backend/repository/kassenjournal_repo/repo.go:452-470](../../backend/repository/kassenjournal_repo/repo.go#L452-L470):
  `GetTSESignaturByTxID` (Seitentabelle nach `tx_id`).
- [backend/sqlc/queries/tse_signaturen.sql](../../backend/sqlc/queries/tse_signaturen.sql):
  `tse_signaturen` als Quelle nachsignierter Vorgänge.
- [backend/domain/kasse/tse_data.go:10-19](../../backend/domain/kasse/tse_data.go#L10-L19):
  `TSEData` im Event-Payload als Happy-Path-Quelle.

### What to build

Der Mapper vereinigt für jeden signierten Vorgang die `TSEData` aus dem
Event-Payload mit der `tse_signaturen`-Seitentabelle: liegt im Event keine
Signatur vor (Persistenz während eines TSE-Ausfalls, später nachsigniert), wird
sie per `tx_id` aus der Seitentabelle nachgeladen. So erscheinen auch nachsignierte
Vorgänge vollständig in `transactions_tse.csv`. Der Orchestrator reicht die nötige
Lesefunktion an den reinen Mapper durch, ohne dass dieser I/O kennt.

### Acceptance criteria

- [ ] Ein während eines TSE-Ausfalls ohne Event-Signatur persistierter und später
      nachsignierter Vorgang erscheint vollständig in `transactions_tse.csv`
      (Transaktionsnummer, Signaturzähler, Signatur), belegt per Golden-Test.
- [ ] Vorgänge mit Signatur im Event-Payload bleiben unverändert; die Seitentabelle
      wird nur als Fallback genutzt.
- [ ] `make check` grün.
