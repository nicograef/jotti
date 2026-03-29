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

Alle bekannten Abweichungen zwischen Ist-Zustand und Soll-Zustand wurden behoben. Die folgende Tabelle dokumentiert bewusste Entscheidungen, die korrekt sind und keinen Handlungsbedarf haben.

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

### Abgeschlossene Migrationen

Die folgenden Umbenennungen wurden im Kassenjournal-Redesign abgeschlossen und sind im Code vollständig umgesetzt.

| Bereich          | Alt (vor Redesign) | Neu (aktueller Code)  | Status           |
| ---------------- | ------------------ | --------------------- | ---------------- |
| DB-Tabelle       | `events`           | `kassenjournal`       | ✅ Abgeschlossen |
| DB-Projektion    | `table_state`      | `tisch_sessions`      | ✅ Abgeschlossen |
| Domain-Paket     | `domain/table/`    | `domain/kasse/`       | ✅ Abgeschlossen |
| Repository-Paket | `event_repo/`      | `kassenjournal_repo/` | ✅ Abgeschlossen |
| Frontend-Typen   | Englische Typen    | Deutsche UL-Typen     | ✅ Abgeschlossen |

> **Hinweis `domain/table/`:** Das Paket existiert weiterhin für Tisch-Stammdaten (`tisch.go`). Die Kasse-Logik (Event-Sourcing, Tisch-Sessions, Kassensitzung) wurde nach `domain/kasse/` ausgelagert — `domain/table/` enthält nur noch die CRUD-Entität `Tisch`.

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

| Rolle          | Code-Wert        | Berechtigungen                                 |
| -------------- | ---------------- | ---------------------------------------------- |
| Admin          | `admin`          | Produkte, Tische, Benutzer verwalten + Service |
| Serviceleitung | `serviceleitung` | Service-Funktionen + Stornierung + Auszahlung  |
| Servicekraft   | `service`        | Bestellen, Ausgabe bestätigen, Kassieren       |

Go-Typ: `Role` mit `AdminRole`, `ServiceleitungRole`, `ServiceRole` · DB-Enum: `UserRole` (`'admin'`, `'serviceleitung'`, `'service'`)

#### Einmalpasswort

Vom Admin generiertes 6-stelliges numerisches Passwort für die Erstanmeldung oder das Zurücksetzen eines Passworts.

Go-Feld: `OnetimePasswordHash` · DB-Spalte: `onetime_password_hash` · TS-Schema: `OnetimePasswordSchema`

#### Token

JWT (JSON Web Token) mit Benutzer-ID und Rolle. 12 Stunden gültig. Reiner Infrastruktur-Begriff — Englisch im Code ist korrekt.

---

### Kasse (Core Domain)

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

#### Bestellung

Ein Vorgang, bei dem eine Servicekraft Positionen für einen Tisch aufnimmt. Erzeugt ein `BestellungAufgenommen`-Event.

| Go-Struct    | TS-Typ       | Event-Typ                   |
| ------------ | ------------ | --------------------------- |
| `Bestellung` | `Bestellung` | `bestellung-aufgenommen:v1` |

#### Position

Ein einzelner Posten innerhalb einer Bestellung: Produktvariante + Menge + Einzelpreis. Alle Felder werden als Fat Event eingefroren.

| Go-Struct  | TS-Typ     | JSON-Keys (Schlüsselfelder)                                                                    |
| ---------- | ---------- | ---------------------------------------------------------------------------------------------- |
| `Position` | `Position` | `positionId`, `varianteId`, `produktName`, `varianteName`, `kategorie`, `einzelpreis`, `menge` |

#### Ausgabe

Bestätigung, dass bestellte Positionen dem Gast übergeben wurden. Erzeugt ein `AusgabeBestaetigt`-Event.

| Go-Struct | TS-Typ    | Event-Typ               |
| --------- | --------- | ----------------------- |
| `Ausgabe` | `Ausgabe` | `ausgabe-bestaetigt:v1` |

#### Zahlung

Kassierung einer Barzahlung. Kann sich auf einzelne Positionen beziehen. Erzeugt ein `ZahlungKassiert`-Event.

| Go-Struct | TS-Typ    | Event-Typ             |
| --------- | --------- | --------------------- |
| `Zahlung` | `Zahlung` | `zahlung-kassiert:v1` |

#### Stornierung

Nachträgliche Aufhebung bestellter Positionen. Nur durch Serviceleitung oder Admin. `Kommentar` ist Pflichtfeld. Erzeugt ein `StornierungErteilt`-Event.

| Go-Struct     | TS-Typ        | Event-Typ                |
| ------------- | ------------- | ------------------------ |
| `Stornierung` | `Stornierung` | `stornierung-erteilt:v1` |

#### Auszahlung

Auszahlung an den Gast, um einen negativen Saldo auszugleichen — entsteht, wenn bereits kassierte Positionen nachträglich storniert wurden (K-05). Kein Positionsbezug; freier Betrag. `Kommentar` ist Pflichtfeld. Erzeugt ein `AuszahlungGeleistet`-Event.

| Go-Struct    | TS-Typ       | Event-Typ                 |
| ------------ | ------------ | ------------------------- |
| `Auszahlung` | `Auszahlung` | `auszahlung-geleistet:v1` |

#### Saldo

Offener Betrag einer Tisch-Session: Bestellungen − Zahlungen − Stornierungen + Auszahlungen. Immer in Cent.

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

Die Kassensitzung und der Kassenbestand gehören zum Core-Domain-Kontext **Kasse** und nutzen dieselbe Persistenzstrategie: **Event-Sourcing im Kassenjournal**. Kassensitzung-Events werden unter dem Subject `kassensitzung-{nr}` geschrieben.

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

#### DruckerKonfiguration

Zuordnung einer Produktkategorie zu einer Drucker-IP und einem Bonmodus. CRUD-Entität.

DB-Tabelle: `kategorie_drucker`

#### Bonmodus

Druckmodus für Bestellbons: einzelner Bon pro Position oder ein gesammelter Bon pro Bestellung.

DB-Enum: `'pro_position'`, `'pro_bestellung'`

#### Relay

Separater Dienst (`cmd/relay/`), der Druckaufträge an lokale ESC/POS-Drucker weiterleitet.

---

### Gastronomie & Betrieb

Operative Fachbegriffe aus dem Gastronomie- und Festzeltbetrieb, die für Konfiguration, Buchführung und Compliance in jotti relevant sind.

- **Inhaus / Außerhaus:** Gesetzlich vorgeschriebene Unterscheidung, ob Speisen vor Ort verzehrt werden (voller MwSt.-Satz, 19 %) oder mitgenommen werden (ermäßigter Satz, 7 %). In jotti über den `Steuersatz` pro Produkt konfiguriert.

- **Trinkgeld:** Zuzahlung des Gastes. Buchhalterisch relevant: Trinkgeld an den Verein ist voll steuerpflichtig; Trinkgeld direkt an die Servicekraft ist in der Regel steuerfrei. jotti unterscheidet dies aktuell nicht — Hinweis für Betreiber in `docs/compliance.md`.

- **BYOD (Bring Your Own Device):** Servicekräfte nutzen ihre eigenen Smartphones. Das System ist Mobile-first konzipiert und läuft vollständig im Browser — keine App-Installation nötig.

- **Geldkatze / Kellnerportemonnaie:** Physische Geldtasche, die Servicekräfte im mobilen Festzeltbetrieb bei sich tragen, um direkt am Tisch kassieren zu können. Kein Code-Bezug — betriebliche Infrastruktur des Betreibers.

- **Belegausgabepflicht (Bonpflicht):** Gesetzliche Pflicht nach § 146a Abs. 2 AO, bei jedem abgeschlossenen Kassiervorgang einen Beleg auszustellen. In jotti: Bondruck via ESC/POS (Phase 1) + TSE-Signaturfelder (Phase 2). Siehe → **Kassenbeleg**.

- **eBeleg:** Digitaler Kassenbon (z. B. als PDF über QR-Code) als rechtskonformer, papierloser Ersatz für den Ausdruck. Phase-3-Feature in jotti — siehe `docs/roadmap.md`.

- **Kassensturzfähigkeit:** Anforderung, dass der berechnete Soll-Bestand an Bargeld jederzeit mit dem physisch vorhandenen Ist-Bestand übereinstimmt. Voraussetzung für GoBD-Konformität.

- **DifferenzSollIst:** Geschäftsvorfalltyp (`GV_TYP` im DSFinV-K), mit dem Fehlbeträge oder Überschüsse beim Kassensturz zwingend ausgebucht werden müssen.

- **Geldtransit / Privatentnahme:** Geschäftsvorfalltypen für Barentnahmen aus der Kasse (z. B. Bankeinzahlung), die gebucht werden müssen, um die Kassensturzfähigkeit aufrechtzuerhalten.

---

### Fiskalkonformität (Compliance Sub-Domain)

Begriffe für die gesetzlich vorgeschriebene Fiskalisierung nach § 146a AO und KassenSichV. Diese Sub-Domain wird phasenweise implementiert — siehe `docs/roadmap.md`.

> **Sprachkonvention:** Fiskal-Fachbegriffe folgen der deutschen Gesetzessprache und DSFinV-K-Spezifikation. Technische Interface-Namen bleiben englisch (Go-Konvention).

#### Gesetzliche Grundlagen

- **AO (Abgabenordnung):** Zentrales deutsches Steuergesetz. § 146a AO regelt die Pflichten (TSE, Belegausgabe, Kassenmeldung) für alle Betreiber elektronischer Aufzeichnungssysteme — und damit für jeden jotti-Betreiber.

- **KassenSichV (Kassensicherungsverordnung):** Auf der AO basierende Verordnung, die technische Detailanforderungen an manipulationssichere Kassen, TSE und Belege vorschreibt.

- **GoBD:** „Grundsätze zur ordnungsmäßigen Führung und Aufbewahrung von Büchern, Aufzeichnungen und Unterlagen in elektronischer Form" — Bundesfinanzministerium-Schreiben. Steuerrelevante Daten müssen 10 Jahre lang lückenlos und unveränderbar gespeichert werden. jottis Event-Sourcing-Architektur (Append-only) erfüllt diese Unveränderbarkeitsanforderung strukturell.

- **BSI (Bundesamt für Sicherheit in der Informationstechnik):** Deutsche Bundesbehörde, die technische Richtlinien (TR-03153) und Schutzprofile für die TSE-Zertifizierung definiert.

#### TSE & Kryptografie

- **TSE (Technische Sicherheitseinrichtung):** Zwingend vorgeschriebenes, vom BSI zertifiziertes Sicherheitsmodul, das jeden Kassiervorgang kryptografisch signiert. In jotti über ein Adapter-Pattern eingebunden (`TSEClient`-Interface, BYOT-Modell).

- **Cloud-TSE / Hardware-TSE:** Zwei Bereitstellungsformen: Eine Hardware-TSE ist ein physisches Speichermedium (USB/SD) an der Kasse; bei der Cloud-TSE werden Transaktionen über eine API in einem zertifizierten Rechenzentrum signiert. jotti unterstützt Cloud-TSE (z. B. fiskaly) über das BYOT-Modell.

- **SMAERS (Security Module Application for Electronic Record Keeping System):** Software-Komponente der TSE, die Kassendaten aufbereitet, den Signatur-Input zusammenstellt und mit dem kryptografischen Provider kommuniziert.

- **CSP (Cryptographic Service Provider):** Hardware- oder Cloud-Einheit innerhalb der TSE, die die eigentliche kryptografische Signatur vornimmt.

- **Transaktionsnummer (TSE_TANR):** Eindeutige, fortlaufende Nummer der TSE für jeden Kassiervorgang. Dient der Lückenerkennung.

- **Signaturzähler (TSE_TA_SIGZ):** Stetig ansteigender Zähler, der bei jedem Signaturvorgang hochgezählt wird — beweist die lückenlose kryptografische Kette. Go-Typ: `uint64` · JSON-Key: `signature_counter` · Pflichtfeld auf dem Kassenbeleg.

- **Prüfwert / Signatur:** Kryptografischer Hash-Wert (z. B. ECDSA-SHA256), der den Vorgang absiegelt und auf dem Kassenbeleg abgedruckt werden muss.

#### Kasse & Identifikation

| Begriff                     | Bedeutung                                                                                                                       |
| --------------------------- | ------------------------------------------------------------------------------------------------------------------------------- |
| **KassenID / Seriennummer** | Eindeutige UUID-v4 der jotti-Instanz. Für ELSTER-Meldung und TSE-Protokoll. **(nicht implementiert)** — siehe `docs/roadmap.md` |
| **TSEClient**               | Go-Interface mit Methoden `StartTransaction`, `UpdateTransaction`, `FinishTransaction`. **(nicht implementiert)**               |

#### Transaktions-Lifecycle

Die TSE-Kommunikation folgt einem strikten Lifecycle. Jeder Kassiervorgang durchläuft diese Schritte:

- **StartTransaction:** API-Aufruf an die TSE bei Beginn eines neuen Vorgangs. Eröffnet die Transaktion und gibt die `TSE_TANR` zurück.
- **UpdateTransaction:** Optionaler API-Aufruf, um einer offenen Transaktion neue Daten hinzuzufügen. Nur bei bestimmten Vorgangsarten erlaubt.
- **FinishTransaction:** Zwingender Abschluss-Aufruf, der die finale Signatur (Prüfwert) der TSE generiert.
- **processType:** Strikt normierter String, der der TSE die Art des Vorgangs mitteilt (z. B. `Kassenbeleg-V1`, `Bestellung-V1`, `SonstigerVorgang-V1`).
- **processData:** Payload-String im BSI-Format (UTF-8, Punkt als Dezimaltrenner) mit Beträgen, Mengen und Steuersätzen.

#### Abrechnungsstruktur

Fiskal-Begriffe der Abrechnungsstruktur — Definitionen und Code-Mappings siehe oben unter → Kasse (Kassensitzung und Kassenbestand).

- **Abrechnungskreis:** → Tisch-Session (Abrechnungskreis)
- **Tagesabschluss / Z-Bon:** → Z-Bon (Tagesabschluss)
- **X-Bon:** Zwischenbericht, kein Tagesabschluss im Rechtssinne.
- **Bonkopf / Bonpos:** DSFinV-K-Aufteilung: Bonkopf enthält Metadaten und Gesamtsummen; Bonpos listet die Artikelzeilen.

#### Steuern

| Begriff          | Bedeutung                                                                                                                                            |
| ---------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Steuersatz**   | Steuerklasse eines Produkts. Enum-Werte: `standard` (19 %), `ermaessigt` (7 %), `befreit` (0 %). **(nicht implementiert)** — siehe `docs/roadmap.md` |
| **Steuerbetrag** | Berechneter Steuerbetrag in Cent für eine Position oder einen Vorgang. Immer in Cent, niemals als Float. **(nicht implementiert)**                   |
| **Nettobetrag**  | Betrag vor Steuerabzug. Immer in Cent. **(nicht implementiert)**                                                                                     |

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

Die folgenden Begriffe sind definiert, aber noch nicht im Code implementiert. Details und Priorisierung in `docs/roadmap.md`.

- **Bon:** Gedruckter Beleg mit Tisch, Servicekraft, Positionen, Mengen, Zeitstempel.
- **Küchendisplay (KDS):** Echtzeit-Anzeige offener Bestellungen an der Ausgabestation.
- **Zubereitungsstatus:** Status einer Position: offen → in Zubereitung → fertig.
- **Ausgabestation:** Physischer Ort (Küche, Getränketheke), an dem Positionen ausgegeben werden.
- **Tagesabrechnung:** Übersicht über Gesamtumsatz, Stornierungen und Umsatz pro Servicekraft.
- **Umsatz:** Summe aller registrierten Zahlungen in einem Zeitraum. Immer in Cent.
- **Stornoquote:** Verhältnis Stornierungsbetrag zu Bestellsumme.
- **Export:** CSV-Download von Umsätzen, Bestellungen und Artikeldaten für die Buchhaltung.

> **Hinweis:** Der Tagesabschluss (Z-Bon) ist kein Reporting-Vorgang, sondern eine transaktionale Operation des Kasse-Kontexts. Siehe → Z-Bon (Tagesabschluss).

---

## Erweitertes Fach-Glossar

Begriffe aus dem POS/Gastronomie-Umfeld, die **nicht im Scope von jotti** sind (bewusste Abgrenzung — siehe `docs/anforderungen.md` §6).

**Vereinswesen:** Zuwendungsbestätigung, Sponsoring, Aufwandsspende, Rückspende — Details zu steuerlichen Sphären siehe → Vereinswesen & Steuerliche Sphären.

**Warenwirtschaft:** Warenwirtschaft (WaWi), Inventur, Warengruppe, Pfand/Pfandrückzahlung, Einzweck-/Mehrzweckgutschein — explizit nicht im Scope.

**Hardware:** Kartenterminal (EFT-Terminal), Kassenlade, Gangsteuerung — explizit nicht im Scope.
