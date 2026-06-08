# Ubiquitous Language — jotti

Dieses Dokument ist die **verbindliche Referenz** für die Ubiquitous Language des jotti-Projekts — für Entwickler, Agenten und alle Projektbeteiligten. Es definiert die Fachbegriffe der Domäne, ihre Code-Repräsentationen und die Sprachkonventionen pro Schicht.

Die Ubiquitous Language ist ein **Living Document**: Sie wird fortlaufend aktualisiert, wenn sich Begriffe, Strukturen oder Konventionen ändern.

## Sprachkonventionen

1. **Domänenbegriffe sind deutsch.** Alle Fachbegriffe der Kasse, der Stammdaten und der Gastronomie-Domäne werden auf Deutsch benannt — in Code, Dokumentation und Kommunikation. Beispiele: `Bestellung`, `Tisch`, `Zahlung`, `Position`, `Ausgabe`, `Stornierung`, `Saldo`, `Kassensitzung`, `Kassenjournal`.

2. **Infrastruktur-Code bleibt englisch.** Authentifizierung, Konfiguration, HTTP-Framework und generische Sub-Domains verwenden englische Bezeichnungen. Beispiele: `User`, `Role`, `Token`, `Config`, `Middleware`. Technische Felder (z. B. `created_at`, `status`, `id`) bleiben in allen Schichten englisch.

3. **Benutzer-sichtbare Strings sind deutsch.** Alle UI-Labels, Fehlermeldungen, Platzhalter und Hilfetexte werden auf Deutsch formuliert. Im UI heißt es „Benutzer" (nicht „User"), „Einmalpasswort" (nicht „OnetimePassword"), „Getränke" (nicht „beverage").

4. **Commits sind auf Englisch.** Conventional Commits (`feat:`, `fix:`, `refactor:`, `docs:`, `test:`) mit englischen Nachrichten.

## Namenskonventionen pro Schicht

| Schicht                   | Sprache  | Konvention                 | Beispiel                                                                       |
| ------------------------- | -------- | -------------------------- | ------------------------------------------------------------------------------ |
| Go-Domain-Structs         | Deutsch  | PascalCase                 | `Bestellung`, `Tisch`, `Position`                                              |
| Go-Felder (Domäne)        | Deutsch  | PascalCase                 | `GesamtPreisCents`, `SaldoCents`                                               |
| TypeScript-Typen (Domäne) | Deutsch  | PascalCase                 | `Bestellung`, `Tisch`, `Zahlung`                                               |
| JSON-Keys (Domäne)        | Deutsch  | camelCase                  | `"gesamtPreisCents"`, `"saldoCents"`                                           |
| API-Pfade (Domäne)        | Deutsch  | kebab-case                 | `/bestellung-aufnehmen`, `/zahlung-kassieren`                                  |
| DB-Tabellen (Domäne)      | Deutsch  | snake_case                 | `tische`, `produkte`, `produkt_varianten`, `tisch_sessions`, `kassensitzungen` |
| DB-Tabellen (Infra.)      | Englisch | snake_case                 | `users`, `kassenjournal`                                                       |
| DB-Spalten (Domäne)       | Deutsch  | snake_case                 | `kategorie`, `preis_cents`, `produkt_id`                                       |
| DB-Spalten (Infrastr.)    | Englisch | snake_case                 | `created_at`, `updated_at`, `status`, `id`                                     |
| Frontend-Routen           | Deutsch  | kebab-case                 | `/service/tische`, `/admin/produkte`                                           |
| Auth/Infrastruktur-Code   | Englisch | Sprachübliche Konventionen | `User`, `Role`, `Token`, `Config`                                              |

> **Pfadkonvention:** Dateipfade sind relativ angegeben — `domain/…` und `api/…` liegen unter `backend/`, `src/…` unter `frontend/`, `migrations/…` unter `database/`.

## Abweichungen: Ist-Zustand vs. Soll-Zustand

Die bekannten Abweichungen zwischen Ist-Zustand und Soll-Zustand wurden behoben — die Bondruck-Neuordnung (Druckstation, Druckauftrags-Outbox, Kassenbeleg, Relay-Transport) ist vollständig umgesetzt. Die folgende Tabelle dokumentiert die verbleibenden bewussten Entscheidungen, die korrekt sind und keinen Handlungsbedarf haben.

### Kein Handlungsbedarf (bewusst korrekt)

| Bereich              | Ist (Code)                                      | Begründung                                                                                                    |
| -------------------- | ----------------------------------------------- | ------------------------------------------------------------------------------------------------------------- |
| DB-Tabellen (Infra)  | `users`, `kassenjournal`                        | Englisch ist korrekt — Infrastruktur / Generic Sub-Domain.                                                    |
| DB-Tabellen (Domain) | `tische`, `produkte`, `produkt_varianten`       | Deutsch ist korrekt — Domänenbegriffe sind vertikal konsistent.                                               |
| Frontend-Routen      | `/admin/produkte`, `/service/tische`            | Deutsch ist korrekt — Routen repräsentieren Domänenkonzepte.                                                  |
| Auth-Code            | `User`, `Role`, `OnetimePassword`               | Englisch ist korrekt — Generic Sub-Domain.                                                                    |
| Auth-Routen          | `/login`, `/set-password`                       | Englisch ist korrekt — Auth ist Infrastruktur.                                                                |
| Status-Enums         | `active`, `inactive`, `deleted`                 | Englisch ist korrekt — technische Lifecycle-States, kein Domänenbegriff.                                      |
| Kassenjournal        | `Historie` (Code) vs. `Kassenjournal` (Entwurf) | Bewusste Abweichung: „Historie" ist im Code und UI etabliert, beide Begriffe sind dokumentiert.               |
| KassensitzungState   | `KassensitzungState` (Go + TS)                  | Domänenbegriff ist „Kassensitzung". Suffix `State` markiert den CRUD-Zustand der Entität — kein Rename nötig. |

---

## Begriffsdefinitionen

### Vereinswesen & Steuerliche Sphären

Das Finanzamt betrachtet einen Verein in vier steuerliche Bereiche (Sphären). Diese bestimmen Buchführungspflichten und Steuersätze — und damit direkt die korrekte Konfiguration von jotti.

- **Gemeinnützigkeit:** Steuerbegünstigter Status eines Vereins, der voraussetzt, dass die Tätigkeit der Allgemeinheit selbstlos zugutekommt und in der Satzung verankert ist.
- **Ideeller Bereich:** Steuerfreier Kernbereich des Vereins ohne wirtschaftliche Tätigkeit (finanziert durch Spenden und Mitgliedsbeiträge).
- **Wirtschaftlicher Geschäftsbetrieb (WGB):** In der Regel steuerpflichtiger Bereich, in dem der Verein wie ein normales Unternehmen agiert — z. B. Getränke- und Essensverkauf auf einem Vereinsfest. jotti ist primär für diesen Bereich konzipiert.
- **Zweckbetrieb:** Wirtschaftlicher Geschäftsbetrieb, der steuerbegünstigt ist, weil er unmittelbar dem gemeinnützigen Zweck dient (z. B. Eintrittsgelder für ein Sportturnier).
- **Vermögensverwaltung:** Steuerfreie, passive Einnahmen aus Vereinsvermögen (z. B. Zinsen, Mieteinnahmen).
- **Kleinunternehmerregelung (§ 19 UStG):** Vereine mit geringen Umsätzen können von der Umsatzsteuerpflicht befreit sein. Hat direkte Auswirkungen auf die korrekte Konfiguration des `Steuersatz` im System.

### Akteure & Rollen

> **Sprachkonvention:** Systemrollen verwenden englische Code-Bezeichnungen — Auth ist eine Generic Sub-Domain. Benutzer-sichtbare Strings im UI sind deutsch.

**Fachliche Akteure (kein Code-Mapping):**

- **Servicekraft / Bedienung:** Freiwillige Helfer, die im Festzeltbetrieb Bestellungen aufnehmen, kassieren und Ausgaben bestätigen. Entspricht der Systemrolle `service`.
- **Serviceleitung:** Erfahrene Servicekraft mit erweiterten Rechten für Stornierungen und Auszahlungen. Entspricht der Systemrolle `serviceleitung`.
- **Kassenwart / Schatzmeister:** Vorstandsmitglied, das für Finanzen, Buchhaltung und den korrekten DSFinV-K-Export sowie Steuererklärungen verantwortlich ist. Entspricht typischerweise der Systemrolle `admin`.
- **Vorstand:** Gesetzliches Vertretungsorgan des Vereins. Haftet persönlich für die Einhaltung steuerlicher Pflichten (GoBD, KassenSichV). Entspricht typischerweise der Systemrolle `admin`.

**Systemrollen (mit Code-Mapping):**

#### Benutzer

Person mit Zugang zum System, identifiziert durch einen eindeutigen Benutzernamen.

| Go-Struct | TS-Typ | DB-Tabelle |
| --------- | ------ | ---------- |
| `User`    | `User` | `users`    |

#### Rolle

Berechtigungsstufe eines Benutzers. Bestimmt, welche Aktionen im System verfügbar sind.

| Rolle          | Code-Wert        |
| -------------- | ---------------- |
| Admin          | `admin`          |
| Serviceleitung | `serviceleitung` |
| Servicekraft   | `service`        |

Go-Typ: `Role` mit `AdminRole`, `ServiceleitungRole`, `ServiceRole` · DB-Enum: `UserRole` (`'admin'`, `'serviceleitung'`, `'service'`)

Die vollständige Berechtigungsmatrix steht in [handbuch.md §5.1](handbuch.md#51-rollen-und-berechtigungsmatrix).

#### Einmalpasswort

Vom Admin generiertes 6-stelliges numerisches Passwort für die Erstanmeldung oder das Zurücksetzen eines Passworts.

Go-Feld: `OnetimePasswordHash` · DB-Spalte: `onetime_password_hash` · TS-Schema: `OnetimePasswordSchema`

#### Token

JWT (JSON Web Token) mit Benutzer-ID und Rolle. 12 Stunden gültig. Reiner Infrastruktur-Begriff — Englisch im Code ist korrekt.

---

### Kasse (Core Domain)

Die Event-Feldschemata (Felder, Typen, Constraints) aller Kasse-Events sind kanonisch in [handbuch.md §3.6](handbuch.md#36-domain-events) definiert. Die folgenden Einträge geben die Begriff↔Code-Mappings (Go-Struct, TS-Typ, Event-Typ, JSON-Keys).

#### Tisch (Stammdaten)

Reine Stammdaten-Entität: ein physischer Ort, an dem Gäste sitzen. Hat einen Namen, Status (active/inactive/deleted) und wird vom Admin verwaltet. Im Kasse-Kontext wird der Tisch nur über seine ID referenziert — die `tisch_id` fließt in das Subject der Tisch-Session ein.

| Go-Struct | TS-Typ  | DB-Tabelle |
| --------- | ------- | ---------- |
| `Tisch`   | `Tisch` | `tische`   |

#### Tisch-Session (Abrechnungskreis)

Das Event-Sourced Aggregat im Kasse-Kontext. Bildet alle Geschäftsvorfälle (Bestellungen, Zahlungen, Stornierungen, Ausgaben, Auszahlungen) eines Tisches innerhalb einer Kassensitzung ab. Entsteht implizit mit der ersten Bestellung. Subject-Format: `kassensitzung-{nr}/tisch-{tischId}`.

| Go-Struct      | TS-Typ         | DB-Projektion    | Subject-Format                       |
| -------------- | -------------- | ---------------- | ------------------------------------ |
| `TischSession` | `TischSession` | `tisch_sessions` | `kassensitzung-{nr}/tisch-{tischId}` |

> **Hinweis `domain/table/`:** Das Paket existiert weiterhin für Tisch-Stammdaten (`tisch.go`). Die Kasse-Logik (Event-Sourcing, Tisch-Sessions, Kassensitzung) wurde nach `domain/kasse/` ausgelagert — `domain/table/` enthält nur noch die CRUD-Entität `Tisch`.

#### Direktverkauf

Schlankes Event-Sourced Aggregat im Kasse-Kontext für den **Barverkauf an der Theke**: bestellen, zahlen und ausgeben in einem Schritt — **ohne** Tisch und **ohne** Projektion. Jeder **Verkauf** ist ein eigener Event-Stream mit eigener UUID. Der Stream-Typ `direktverkauf` schreibt ausschließlich ins Kassenjournal. Direktverkauf hat **keine** `Verkaufsstelle`-Stammdaten-Entität.

| Go-Struct | TS-Typ | Event-Typ (Verkauf)          | Subject-Format                            |
| --------- | ------ | ---------------------------- | ----------------------------------------- |
| —         | —      | `direktverkauf-getaetigt:v1` | `kassensitzung-{nr}/direktverkauf-{uuid}` |
| —         | —      | `direktverkauf-storniert:v1` | `kassensitzung-{nr}/direktverkauf-{uuid}` |

> **Verkauf:** die fachliche Einheit eines Direktverkaufs (ein Stream, ein `verkaufId`). Kein eigenes Domain-Struct — der Verkauf existiert nur als Event-Stream im Kassenjournal. `direktverkauf-getaetigt:v1` ist `version = 1`; positionsgenaue Stornierungen sind Folge-Versionen im selben Stream.

> **Direktverkauf-Stornierung:** positionsgenaue Korrektur/Rückgabe eines Verkaufs durch Serviceleitung/Admin (`direktverkauf-storniert:v1`, Folge-Version im selben Stream). Speichert die stornierten Positionen als **Fat-Positionen** (wie der Tisch-Storno, selbst-enthaltend fürs Reporting); die API nimmt `PositionRef` (`positionId` + `menge`) entgegen und reichert sie im Command an. `gesamtStornierungCents` ist die Summe der stornierten Positionen und sofort kassenwirksam (Bargeld-Rückgabe) — **ohne** separate `auszahlung-geleistet`-Buchung, da ein Verkauf keinen offenen Saldo hat. Validierung per On-Demand-Replay des Verkauf-Streams (`ComputeNichtStornierteVerkaufPositionen`): nur noch nicht stornierte Positionen, höchstens die ursprünglich verkaufte Menge. Mehrere Teilstornos pro Verkauf sind zulässig.

> **Direktverkauf-Modus (Bondruck):** steuert den nicht-fiskalischen Bonfluss bei `direktverkauf-getaetigt:v1` über `bondruck_einstellungen.direktverkauf_modus`: `kein_bon` (kein Auftrag), `abholbon` (genau ein kombinierter Abholbon), `an_stationen` (Routing nach Produktkategorie wie Tisch-Arbeitsbon).

#### Abholbon

Nicht-fiskalischer kombinierter Bon für Direktverkauf im Modus `abholbon`. Festes Label „Direktverkauf“, keine Preise, genau ein Druckauftrag (`bon_art = 'arbeitsbon'`) an `bondruck_einstellungen.abholbon_drucker_ip`.

#### Bestellung

Ein Vorgang, bei dem eine Servicekraft Positionen für einen Tisch aufnimmt.

| Go-Struct    | TS-Typ       | Event-Typ                   |
| ------------ | ------------ | --------------------------- |
| `Bestellung` | `Bestellung` | `bestellung-aufgenommen:v1` |

#### Position

Ein einzelner Posten innerhalb einer Bestellung: Produktvariante + Menge + Einzelpreis. Alle Felder werden als Fat Event eingefroren.

| Go-Struct  | TS-Typ     | JSON-Keys (Schlüsselfelder)                                                                    |
| ---------- | ---------- | ---------------------------------------------------------------------------------------------- |
| `Position` | `Position` | `positionId`, `varianteId`, `produktName`, `varianteName`, `kategorie`, `einzelpreis`, `menge` |

#### Ausgabe

Bestätigung, dass bestellte Positionen dem Gast übergeben wurden.

| Go-Struct | TS-Typ    | Event-Typ               |
| --------- | --------- | ----------------------- |
| `Ausgabe` | `Ausgabe` | `ausgabe-bestaetigt:v1` |

#### Zahlung

Kassierung einer Barzahlung. Kann sich auf einzelne Positionen beziehen.

| Go-Struct | TS-Typ    | Event-Typ             |
| --------- | --------- | --------------------- |
| `Zahlung` | `Zahlung` | `zahlung-kassiert:v1` |

#### Stornierung

Nachträgliche Aufhebung bestellter Positionen. Nur durch Serviceleitung oder Admin. `Kommentar` ist Pflichtfeld.

| Go-Struct     | TS-Typ        | Event-Typ                |
| ------------- | ------------- | ------------------------ |
| `Stornierung` | `Stornierung` | `stornierung-erteilt:v1` |

#### Auszahlung

Auszahlung an den Gast, um einen negativen Saldo auszugleichen — entsteht, wenn bereits kassierte Positionen nachträglich storniert wurden (K-05). Kein Positionsbezug; freier Betrag. `Kommentar` ist Pflichtfeld.

| Go-Struct    | TS-Typ       | Event-Typ                 |
| ------------ | ------------ | ------------------------- |
| `Auszahlung` | `Auszahlung` | `auszahlung-geleistet:v1` |

#### Saldo

Offener Betrag einer Tisch-Session. Immer in Cent. Die Saldo-Formel ist kanonisch in [handbuch.md §3.7](handbuch.md#37-invarianten) definiert.

Go-Projektion-Feld: `SaldoCents` · Berechnung: `ApplyEvent()` → `TischSession`

#### Splitrechnung / Teilzahlung

Kassiervorgang, bei dem der Gesamtsaldo eines Tisches auf mehrere Gäste aufgeteilt wird. In jotti über positionsbezogene Zahlungen abgebildet: Eine `Zahlung` kann sich auf eine Teilmenge der offenen Positionen beziehen.

#### Historie

Vollständiger, unveränderlicher Event Stream einer Tisch-Session in chronologischer Reihenfolge.

**Synonym: Kassenjournal.** „Kassenjournal“ ist der formale Fachbegriff (die DB-Tabelle `kassenjournal` enthält alle Events); „Historie“ ist der im Code und UI etablierte Begriff für die Tisch-spezifische Ansicht.

Go-Funktion: `GetHistoryFromEvents()` · Go-Query: `GetTischHistorie()` · API: `/service/get-tisch-historie`

#### Kommentar

Freitextnotiz zu Tischoperationen. Pflichtfeld bei Stornierung und Auszahlung, optional bei Bestellung, Ausgabe und Zahlung.

Go-Feld: `Kommentar` · JSON-Key: `"kommentar"` · TS-Feld: `kommentar`

#### Menge

Anzahl einer Produktvariante innerhalb einer Position.

Go-Feld: `Menge` · JSON-Key: `"menge"` · TS-Feld: `menge`

#### EigeneUebersicht

Kompakte KPI-Übersicht einer Servicekraft: Anzahl und Summe eigener Bestellungen sowie kassierter Zahlungen. Read Model, berechnet aus dem Kassenjournal gefiltert auf `user_id` und `kassensitzung_nr`.

| Go-Struct          | TS-Typ             | JSON-Keys                                                                      |
| ------------------ | ------------------ | ------------------------------------------------------------------------------ |
| `EigeneUebersicht` | `EigeneUebersicht` | `anzahlBestellungen`, `bestellungenCents`, `anzahlZahlungen`, `zahlungenCents` |

#### HistorieEintrag

Einzelner Eintrag in der Tisch-Historie, typisiert nach Art (Bestellung, Zahlung, Stornierung, Ausgabe, Auszahlung).

Go-Struct: `HistorieEintrag` · Feld `Art`: `HistorieEintragArt` (Enum: `"bestellung"`, `"zahlung"`, `"stornierung"`, `"ausgabe"`, `"auszahlung"`)

#### PositionRef

Referenz auf eine Position (ID + Menge) für Zahlung, Ausgabe und Stornierung.

Go-Struct: `PositionRef` · TS-Typ: `PositionRef` · JSON-Keys: `positionId`, `menge`

#### AktiverTisch

Kompakte Tisch-Darstellung mit Saldo für die Tischübersicht. Read Model.

Go-Struct: `AktiverTisch` (ID, Name, SaldoCents)

#### BestellPositionInput

Frontend-Eingabetyp für eine einzelne Bestellposition (Produkt + Variante + Menge).

TS-Typ: `BestellPositionInput` · JSON-Keys: `produktId`, `varianteId`, `menge`

---

### Kasse — Kassensitzung und Kassenbestand

Die Kassensitzung und der Kassenbestand gehören zum Core-Domain-Kontext **Kasse** und nutzen dieselbe Persistenzstrategie: **Event-Sourcing im Kassenjournal**. Kassensitzung-Events werden unter dem Subject `kassensitzung-{nr}` geschrieben. Die Event-Feldschemata sind kanonisch in [handbuch.md §3.6](handbuch.md#36-domain-events) definiert.

#### Kassensitzung

Global nummerierter Betriebstag, der einen Abrechnungszeitraum (typischerweise einen Veranstaltungstag) abgrenzt. Wird durch Admin-Aktion eröffnet (Event `kassensitzung-eroeffnet:v1`). Maximal eine Kassensitzung kann gleichzeitig `offen` sein. Ohne offene Kassensitzung ist der gesamte Kassenbetrieb gesperrt (HTTP 409). Die `z_nr` ist ein fortlaufender, lückenloser Zähler in der `kassensitzungen`-Entität.

| Go-Struct            | DB-Tabelle        | Subject-Format       |
| -------------------- | ----------------- | -------------------- |
| `KassensitzungState` | `kassensitzungen` | `kassensitzung-{nr}` |

#### Bezeichnung

Freitextlabel einer Kassensitzung (z. B. „Maihock 2026").

Go-Feld: `Bezeichnung` · DB-Spalte: `kassensitzungen.bezeichnung` · JSON-Key: `"bezeichnung"`

#### Abrechnungskreis

DSFinV-K-Pflichtfeld (`ABRECHNUNGSKREIS`). Im neuen Modell ist der Abrechnungskreis **identisch mit der Tisch-Session**: pro Tisch und Kassensitzung existiert ein Abrechnungskreis. Der DSFinV-K-Export-Wert wird aus dem Tischnamen abgeleitet (z. B. `Tisch 42`) — unabhängig vom Subject-Format.

#### Anfangsbestand

Wechselgeld zu Beginn einer Veranstaltung/Schicht. Pro Kassensitzung darf genau ein Anfangsbestand gesetzt werden. Wird als Event `anfangsbestand-gesetzt:v1` im Kassenjournal persistiert.

Go-Event-Data-Feld: `BetragCents` · JSON-Key: `betragCents`

#### Kassenbestand

SQL-Aggregation über den erwarteten Bargeldbestand. Berechnung: Anfangsbestand + Zahlungen − Auszahlungen + Kassenbewegungen (vorzeichenbehaftet). Kein eigenes Read Model — wird on-demand aus dem Kassenjournal aggregiert.

JSON-Key: `sollBestandCents`

#### Kassenbewegung

Oberbegriff für Geldtransit, Privatentnahme und Privateinlage — Bargeld-Bewegungen außerhalb des Tisch-Verkehrs. Wird als Event `kassenbewegung-gebucht:v1` mit Feld `art` im Kassenjournal persistiert.

| Event-Typ                   | JSON-Key `art`                                   |
| --------------------------- | ------------------------------------------------ |
| `kassenbewegung-gebucht:v1` | `geldtransit`, `privatentnahme`, `privateinlage` |

#### Geldtransit

Entnahme von Bargeld aus der Kasse zur Einzahlung bei Bank oder Tresor. Reduziert den Soll-Kassenbestand. DSFinV-K-Geschäftsvorfalltyp: `Geldtransit`.

API-Pfad: `/admin/kassenbewegung-buchen` (mit `art: geldtransit`)

#### Privatentnahme

Entnahme von Bargeld in den privaten Bereich des Vereins (nicht Bank). Fachlich analog zu Geldtransit, aber anderer DSFinV-K-Geschäftsvorfalltyp: `Privatentnahme`.

#### Privateinlage

Einlage von Bargeld in die Kasse (z. B. Nachfüllen von Wechselgeld). Erhöht den Soll-Kassenbestand. DSFinV-K-Geschäftsvorfalltyp: `Privateinlage`.

Alle Kassenbewegungen über `/admin/kassenbewegung-buchen` mit jeweiligem `art`-Wert.

#### Kassensturz

Vergleich des errechneten Soll-Bestands mit dem physisch gezählten Ist-Bestand. Der Application Service schreibt ein `kassensturz-durchgefuehrt:v1`-Event; bei Differenz ≠ 0 folgt ein `differenz-soll-ist-gebucht:v1`-Event in derselben Transaktion (Zwei-Event-Muster). Voraussetzung für den Tagesabschluss (Z-Bon).

| Event-Typ                                                        | JSON-Keys                                               |
| ---------------------------------------------------------------- | ------------------------------------------------------- |
| `kassensturz-durchgefuehrt:v1` + `differenz-soll-ist-gebucht:v1` | `sollBestandCents`, `istBestandCents`, `differenzCents` |

#### DifferenzSollIst

Automatisch erzeugtes Event (`differenz-soll-ist-gebucht:v1`) beim Kassensturz, wenn Soll-Bestand ≠ Ist-Bestand. DSFinV-K-Pflicht-Geschäftsvorfalltyp. Wird in derselben Transaktion wie das Kassensturz-Event geschrieben.

#### Z-Bon (Tagesabschluss)

Formeller Tagesabschlussbon: aggregiert alle Transaktionen einer Kassensitzung nach Steuersätzen und Zahlarten. Wird als `tagesabschluss-erstellt:v1`-Event im Kassenjournal persistiert und schließt die Kassensitzung ab (Status → `abgeschlossen`). Erhält eine fortlaufende, nie zurücksetzbare `z_nr`.

| Event-Typ                    | DB-Feld                | JSON-Keys                                                                                                             |
| ---------------------------- | ---------------------- | --------------------------------------------------------------------------------------------------------------------- |
| `tagesabschluss-erstellt:v1` | `kassensitzungen.z_nr` | `zNr`, `zeitraumVon`, `zeitraumBis`, `umsatzGesamtCents`, `stornierungCents`, `auszahlungenCents`, `geldtransitCents` |

> **Abgrenzung:** Der Z-Bon ersetzt die bisherige R-07-Anforderung. Er ist kein Report, sondern eine transaktionale Operation des Kasse-Kontexts.

#### X-Bon

Zwischenbericht: informativer Abruf des aktuellen Kassenstands ohne Rücksetzen. Kein Tagesabschluss im Rechtssinne. Nicht gesetzlich vorgeschrieben.

---

### Stammdaten (Supporting Sub-Domain)

#### Produkt

Artikel im Produktkatalog. Gehört zu genau einer Kategorie und enthält eine oder mehrere Varianten mit je eigenem Preis.

| Go-Struct | TS-Typ    | DB-Tabelle |
| --------- | --------- | ---------- |
| `Produkt` | `Produkt` | `produkte` |

#### Variante

Konkrete Ausprägung eines Produkts mit eigenem Namen und Preis in Cent (z. B. Produkt „Cola" → Varianten „0,3 l" / „0,5 l").

| Go-Struct  | TS-Typ     | DB-Tabelle          |
| ---------- | ---------- | ------------------- |
| `Variante` | `Variante` | `produkt_varianten` |

#### Kategorie

Gruppierung von Produkten. Aktuell drei feste Kategorien.

| Go-Typ      | DB-Enum (`ProduktKategorie`)           | Go-Konstanten                                               |
| ----------- | -------------------------------------- | ----------------------------------------------------------- |
| `Kategorie` | `'essen'`, `'getraenk'`, `'sonstiges'` | `EssenKategorie`, `GetraenkKategorie`, `SonstigesKategorie` |

#### Preis

Geldbeträge werden ausnahmslos als ganzzahlige Cent-Werte gespeichert — niemals als Fließkommazahlen. 3,50 € = 350 Cent.

**Konvention:** Alle Preis-Felder tragen das Suffix `*Cents` — z. B. `PreisCents`, `GesamtPreisCents`, `SaldoCents`, `GesamtZahlungCents`, `GesamtStornierungCents`.

#### Soft-Delete

Logisches Löschen: Datensätze werden nicht physisch entfernt, sondern durch den Status `deleted` markiert. Ermöglicht Referenzintegrität und historische Auswertung.

DB-Enum: `EntityStatus` (`'active'`, `'inactive'`, `'deleted'`) · Go-Konstanten: `ActiveStatus`, `InactiveStatus` (in `domain/table`, `domain/product`, `domain/user`)

#### Favorit

Benutzerspezifische Markierung einer Servicekraft für einen Tisch ("Meine Tische"). Kein Aggregat, keine Events — einfache CRUD-Relation.

| Go-Package                 | DB-Tabelle        |
| -------------------------- | ----------------- |
| `repository/favorit_repo/` | `tisch_favoriten` |

TS-Repräsentation: `istFavorit: boolean` in `AktiverTischMitFavorit` (kein eigener Typ).

---

### Reporting (Read Model)

Reporting-Daten werden on-demand per SQL-Aggregation aus dem Kassenjournal berechnet. Kein eigener Event Stream — reines Read Model.

#### ReportingData

Vollständiger Reporting-Datensatz einer Kassensitzung: Summary + Breakdowns + Stornierungen.

Go-Struct: `ReportingData` · TS-Typ: `ReportingData`

#### Summary

Aggregierte Kennzahlen einer Kassensitzung (Umsatz, Stornierungen, offene Salden, Anzahlen).

Go-Struct: `Summary` · TS-Schema: `SummarySchema`

#### Breakdowns

Aufschlüsselung des Umsatzes nach Servicekraft und Tisch.

Go-Struct: `Breakdowns` · Enthält: `UmsatzProServicekraft []UmsatzServicekraft`, `UmsatzProTisch []UmsatzTisch`

#### UmsatzServicekraft

Umsatz einer einzelnen Servicekraft (Zahlungen, Auszahlungen, Anzahl).

Go-Struct: `UmsatzServicekraft` · TS-Typ: `UmsatzServicekraft`

#### UmsatzTisch

Umsatz eines einzelnen Tisches (Zahlungen, Auszahlungen, Anzahl).

Go-Struct: `UmsatzTisch` · TS-Typ: `UmsatzTisch`

#### StornierungDetail

Detailansicht einer einzelnen Stornierung im Reporting (Zeitpunkt, Tisch, Benutzer, Betrag, Kommentar, Positionen).

Go-Struct: `StornierungDetail` · TS-Typ: `StornierungDetail`

#### StornierungPosition

Einzelne Position innerhalb einer StornierungDetail (Produktname, Variantenname, Menge, Einzelpreis).

Go-Struct: `StornierungPosition` · TS-Typ: `StornierungPosition`

---

### Bondruck & Infrastruktur

**Bondruck** ist der Oberbegriff für zwei fachlich getrennte Bon-Familien auf einer gemeinsamen Druck-Infrastruktur: den operativen **Arbeitsbon** (nicht-fiskalisch, automatisch beim Entstehen von Ware) und den gesetzlichen **Kassenbeleg** (fiskalisch, auf Anforderung beim Kassieren). Beide Familien teilen **keinen** Auslöser, Inhalt oder Rechtsstatus — nur den **Druckauftrag** (Outbox) als Transport.

#### Arbeitsbon

Operativer, nicht-fiskalischer Bon an eine Ausgabestation (Küche, Theke). Trägt **keine Preise**, nur die zuzubereitende/auszugebende Ware (Quelle, Artikel, Menge, Kommentar, Uhrzeit, Bedienung). Wird automatisch bei Bestellaufnahme erzeugt. **Kein Beleg** im Sinne von § 146a AO.

#### Kassenbeleg

Fiskalischer Zahlungsbeleg (§ 146a Abs. 2 AO, § 6 KassenSichV) für den Gast: alle Positionen mit Preisen, Vereinsdaten (Betreiber, K-20) und Kassen-Seriennummer (F-01). Wird **auf Anforderung** pro Kassiervorgang gedruckt — am Fest selten (Befreiung „Verkauf an eine Vielzahl unbekannter Personen", § 146a Abs. 2 Satz 2 AO). DSFinV-K-`processType`: `Kassenbeleg-V1`. Steuer-Aufschlüsselung folgt mit F-07, TSE-Pflichtfelder mit F-02.

#### Druckstation

Konfigurierter Drucker an einem Ausgabeort, je Produktkategorie — die Stationen für **Arbeitsbons**. CRUD-Entität.

DB-Tabelle: `druckstationen`

#### Bonmodus

Druckmodus für Arbeitsbons: einzelner Bon pro Position oder ein gesammelter Bon pro Bestellung.

DB-Enum: `'pro_position'`, `'pro_bestellung'`

#### Druckauftrag

Konkreter Druckjob in der Outbox: Ziel-IP, ESC/POS-Payload und Status (`offen` | `gedruckt`). Single Source of Truth für alle Druckjobs — Arbeitsbon **und** Kassenbeleg. Das Backend reiht ein, der Relay leert.

DB-Tabelle: `druckauftraege` _(geplant; aktuell ein transientes DTO ohne Persistenz)_

#### Relay

Separater Dienst (`cmd/relay/`), der Druckaufträge an lokale ESC/POS-Drucker weiterleitet. Soll: **reiner Transport** — holt offene Druckaufträge, druckt, quittiert; **keine** Fachlogik (ESC/POS-Formatierung liegt im Backend). _(Ist-Zustand: berechnet Druckaufträge noch beim Poll aus Events und hält einen lokalen Cursor — siehe Abweichungen.)_

---

### Gastronomie & Betrieb

- **Inhaus / Außerhaus:** Unterscheidung Vor-Ort-Verzehr (19 % MwSt.) vs. Mitnahme (7 % MwSt.); in jotti über `Steuersatz` pro Produkt konfiguriert.
- **Trinkgeld:** Trinkgeld an den Verein ist voll steuerpflichtig; direkt an die Servicekraft in der Regel steuerfrei. Hinweis für Betreiber: `docs/compliance.md`.
- **BYOD (Bring Your Own Device):** Servicekräfte nutzen eigene Smartphones; kein App-Install nötig.
- **Belegausgabepflicht (Bonpflicht):** Gesetzliche Pflicht nach § 146a Abs. 2 AO — Beleg nach jedem Kassiervorgang. Siehe → **Kassenbeleg**.
- **eBeleg:** Digitaler Kassenbon als papierloser Beleg-Ersatz. Phase-3-Feature — siehe `docs/anforderungen.md`.
- **Kassensturzfähigkeit:** Soll-Bestand muss jederzeit mit dem Ist-Bestand übereinstimmen; Voraussetzung für GoBD-Konformität.
- **DifferenzSollIst:** DSFinV-K-Geschäftsvorfalltyp für Fehlbeträge/Überschüsse beim Kassensturz.
- **Geldtransit / Privatentnahme:** DSFinV-K-Geschäftsvorfalltypen für Barentnahmen; müssen gebucht werden, um Kassensturzfähigkeit aufrechtzuerhalten.

---

### Fiskalkonformität (Compliance Sub-Domain)

Begriffe für die gesetzlich vorgeschriebene Fiskalisierung nach § 146a AO und KassenSichV. Diese Sub-Domain wird phasenweise implementiert — siehe `docs/anforderungen.md`.

> **Sprachkonvention:** Fiskal-Fachbegriffe folgen der deutschen Gesetzessprache und DSFinV-K-Spezifikation. Technische Interface-Namen bleiben englisch (Go-Konvention).

#### Gesetzliche Grundlagen

- **AO (Abgabenordnung):** Zentrales deutsches Steuergesetz. § 146a AO regelt die Pflichten (TSE, Belegausgabe, Kassenmeldung) für alle Betreiber elektronischer Aufzeichnungssysteme — und damit für jeden jotti-Betreiber.

- **KassenSichV (Kassensicherungsverordnung):** Auf der AO basierende Verordnung, die technische Detailanforderungen an manipulationssichere Kassen, TSE und Belege vorschreibt.

- **GoBD:** „Grundsätze zur ordnungsmäßigen Führung und Aufbewahrung von Büchern, Aufzeichnungen und Unterlagen in elektronischer Form" — Bundesfinanzministerium-Schreiben. Steuerrelevante Daten müssen 10 Jahre lang lückenlos und unveränderbar gespeichert werden. jottis Event-Sourcing-Architektur (Append-only) erfüllt diese Unveränderbarkeitsanforderung strukturell.

- **BSI (Bundesamt für Sicherheit in der Informationstechnik):** Deutsche Bundesbehörde, die technische Richtlinien (TR-03153) und Schutzprofile für die TSE-Zertifizierung definiert.

#### TSE & Kryptografie

- **TSE (Technische Sicherheitseinrichtung):** Zwingend vorgeschriebenes, vom BSI zertifiziertes Sicherheitsmodul, das jeden Kassiervorgang kryptografisch signiert. In jotti über ein Adapter-Pattern eingebunden (`TSEClient`-Interface, BYOT-Modell).
- **Cloud-TSE / Hardware-TSE:** Cloud-TSE: Signatur über API in zertifiziertem Rechenzentrum; Hardware-TSE: physisches Speichermedium (USB/SD). jotti unterstützt Cloud-TSE (z. B. fiskaly) über das BYOT-Modell.
- **Transaktionsnummer (TSE_TANR):** Eindeutige, fortlaufende TSE-Nummer pro Kassiervorgang. Dient der Lückenerkennung.
- **Signaturzähler (TSE_TA_SIGZ):** Stetig ansteigender Zähler pro Signaturvorgang. Go-Typ: `uint64` · JSON-Key: `signature_counter` · Pflichtfeld auf dem Kassenbeleg.
- **Prüfwert / Signatur:** Kryptografischer Hash-Wert (z. B. ECDSA-SHA256), der den Vorgang absiegelt und auf dem Kassenbeleg abgedruckt werden muss.

#### Kasse & Identifikation

| Begriff                     | Bedeutung                                                                                                                                                                                 |
| --------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **KassenID / Seriennummer** | Eindeutige UUID-v4 der jotti-Instanz. Für ELSTER-Meldung und TSE-Protokoll. Persistiert in Tabelle `kassenidentitaet` (insert-once), abrufbar über Endpunkt `admin/get-kassenidentitaet`. |
| **TSEClient**               | Go-Interface mit Methoden `StartTransaction`, `UpdateTransaction`, `FinishTransaction`. **(nicht implementiert)**                                                                         |

#### Steuern

| Begriff          | Bedeutung                                                                                                                                                  |
| ---------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Steuersatz**   | Steuerklasse eines Produkts. Enum-Werte: `standard` (19 %), `ermaessigt` (7 %), `befreit` (0 %). **(nicht implementiert)** — siehe `docs/anforderungen.md` |
| **Steuerbetrag** | Berechneter Steuerbetrag in Cent für eine Position oder einen Vorgang. Immer in Cent, niemals als Float. **(nicht implementiert)**                         |
| **Nettobetrag**  | Betrag vor Steuerabzug. Immer in Cent. **(nicht implementiert)**                                                                                           |

#### Export & Meldung

| Begriff                    | Bedeutung                                                                                                                                                               |
| -------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **DSFinV-K**               | „Digitale Schnittstelle der Finanzverwaltung für Kassensysteme" — standardisiertes CSV-ZIP-Exportformat (Version 2.4) für Betriebsprüfungen durch die Finanzverwaltung. |
| **TAR-Archiv**             | Gesetzlich vorgeschriebenes Dateiformat für den Export der rohen, kryptografisch gesicherten TSE-Log-Nachrichten.                                                       |
| **Kassenmeldung / ELSTER** | Pflicht nach § 146a Abs. 4 AO: Jede jotti-Instanz muss innerhalb eines Monats nach Inbetriebnahme über das ELSTER-Portal gemeldet werden.                               |
| **ERiC**                   | „ELSTER Rich Client" — Programmierschnittstelle für die automatisierte ELSTER-Kommunikation. Phase-3-Feature.                                                           |
| **Kassenbeleg**            | Pflichtbeleg nach § 146a Abs. 2 AO nach jedem Kassiervorgang. In jotti: Bondruck via ESC/POS + (Phase 2) TSE-Signaturfelder.                                            |
| **BYOT**                   | „Bring Your Own TSE" — Betreiber schließen selbst einen Vertrag mit einem Cloud-TSE-Anbieter (z. B. fiskaly) und injizieren API-Schlüssel via `.env`.                   |

---

### Architekturprinzipien

- **Event-Sourcing:** Persistenzmuster für den Kassenbetrieb: Zustand wird nicht direkt gespeichert, sondern aus unveränderlichen Events berechnet. Jeder Tisch hat einen eigenen Event Stream.

- **Fat Event:** Event, das alle relevanten Daten zum Zeitpunkt der Aktion enthält — inklusive Produktname und Preis. Damit ist das Event unabhängig von späteren Stammdaten-Änderungen auswertbar.

- **Anti-Corruption Layer (ACL):** Schutzmechanismus zwischen Bounded Contexts: Der Kassenbetrieb friert Stammdaten (Produktname, Preis) in Events ein und ist damit unabhängig von nachträglichen Änderungen an Produkten oder Varianten.

- **Append-only:** Grundprinzip des Event Streams: Events werden nur hinzugefügt, nie geändert oder gelöscht. Falsche Aktionen werden durch kompensierende Events (z. B. Stornierung) aufgehoben. Entspricht dem GoBD-Radierverbot für steuerrelevante Daten.

- **Synchrone Projektion:** Performance-Optimierung, die den Snapshot-as-Event-Ansatz ersetzt hat (siehe ADR: CQRS). Der Tisch-Zustand wird als `tisch_sessions`-Zeile synchron in derselben Transaktion wie das Event geschrieben. Kein Event-Replay beim Lesen nötig.

---

## Geplant (nicht implementiert)

Die folgenden Begriffe sind definiert, aber noch nicht im Code implementiert. Details und Priorisierung in `docs/anforderungen.md`.

- **Bon:** Gedruckter Beleg mit Tisch, Servicekraft, Positionen, Mengen, Zeitstempel.
- **Küchendisplay (KDS):** Echtzeit-Anzeige offener Bestellungen an der Ausgabestation.
- **Zubereitungsstatus:** Status einer Position: offen → in Zubereitung → fertig.
- **Ausgabestation:** Physischer Ort (Küche, Getränketheke), an dem Positionen ausgegeben werden.
- **Tagesabrechnung:** Übersicht über Gesamtumsatz, Stornierungen und Umsatz pro Servicekraft.
- **Umsatz:** Summe aller registrierten Zahlungen in einem Zeitraum. Immer in Cent.
- **Stornoquote:** Verhältnis Stornierungsbetrag zu Bestellsumme.
- **Export:** CSV-Download von Umsätzen, Bestellungen und Artikeldaten für die Buchhaltung.

> **Hinweis:** Der Tagesabschluss (Z-Bon) ist kein Reporting-Vorgang, sondern eine transaktionale Operation des Kasse-Kontexts. Siehe → Z-Bon (Tagesabschluss).
