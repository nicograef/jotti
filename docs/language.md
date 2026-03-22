# Ubiquitous Language — jotti

Dieses Dokument ist die **verbindliche Referenz** für die Ubiquitous Language des jotti-Projekts — für Entwickler, Agenten und alle Projektbeteiligten. Es definiert die Fachbegriffe der Domäne, ihre Code-Repräsentationen und die Sprachkonventionen pro Schicht.

Die Ubiquitous Language ist ein **Living Document**: Sie wird fortlaufend aktualisiert, wenn sich Begriffe, Strukturen oder Konventionen ändern.

## Sprachkonventionen

1. **Domänenbegriffe sind deutsch.** Alle Fachbegriffe des Kassenbetriebs, der Stammdaten und der Gastronomie-Domäne werden auf Deutsch benannt — in Code, Dokumentation und Kommunikation. Beispiele: `Bestellung`, `Tisch`, `Zahlung`, `Position`, `Ausgabe`, `Stornierung`, `Saldo`.

2. **Infrastruktur-Code bleibt englisch.** Authentifizierung, Konfiguration, HTTP-Framework und generische Sub-Domains verwenden englische Bezeichnungen. Beispiele: `User`, `Role`, `Token`, `Config`, `Middleware`. Technische Felder (z. B. `created_at`, `status`, `id`) bleiben in allen Schichten englisch.

3. **Benutzer-sichtbare Strings sind deutsch.** Alle UI-Labels, Fehlermeldungen, Platzhalter und Hilfetexte werden auf Deutsch formuliert. Im UI heißt es „Benutzer" (nicht „User"), „Einmalpasswort" (nicht „OnetimePassword"), „Getränke" (nicht „beverage").

4. **Commits sind auf Englisch.** Conventional Commits (`feat:`, `fix:`, `refactor:`, `docs:`, `test:`) mit englischen Nachrichten.

## Namenskonventionen pro Schicht

| Schicht                   | Sprache  | Konvention                 | Beispiel                                      |
| ------------------------- | -------- | -------------------------- | --------------------------------------------- |
| Go-Domain-Structs         | Deutsch  | PascalCase                 | `Bestellung`, `Tisch`, `Position`             |
| Go-Felder (Domäne)        | Deutsch  | PascalCase                 | `GesamtPreisCents`, `SaldoCents`              |
| TypeScript-Typen (Domäne) | Deutsch  | PascalCase                 | `Bestellung`, `Tisch`, `Zahlung`              |
| JSON-Keys (Domäne)        | Deutsch  | camelCase                  | `"gesamtPreisCents"`, `"saldoCents"`          |
| API-Pfade (Domäne)        | Deutsch  | kebab-case                 | `/bestellung-aufnehmen`, `/zahlung-kassieren` |
| DB-Tabellen (Domäne)      | Deutsch  | snake_case                 | `tische`, `produkte`, `produkt_varianten`     |
| DB-Tabellen (Infrastr.)   | Englisch | snake_case                 | `users`, `events`                             |
| DB-Spalten (Domäne)       | Deutsch  | snake_case                 | `kategorie`, `preis_cents`, `produkt_id`      |
| DB-Spalten (Infrastr.)    | Englisch | snake_case                 | `created_at`, `updated_at`, `status`, `id`    |
| Frontend-Routen           | Deutsch  | kebab-case                 | `/service/tische`, `/admin/produkte`          |
| Auth/Infrastruktur-Code   | Englisch | Sprachübliche Konventionen | `User`, `Role`, `Token`, `Config`             |

> **Pfadkonvention:** Dateipfade sind relativ angegeben — `domain/…` und `api/…` liegen unter `backend/`, `src/…` unter `frontend/`, `migrations/…` unter `database/`.

## Abweichungen: Ist-Zustand vs. Soll-Zustand

Alle bekannten Abweichungen zwischen Ist-Zustand und Soll-Zustand wurden behoben. Die folgende Tabelle dokumentiert bewusste Entscheidungen, die korrekt sind und keinen Handlungsbedarf haben.

### Kein Handlungsbedarf (bewusst korrekt)

| Bereich              | Ist (Code)                                      | Begründung                                                                                      |
| -------------------- | ----------------------------------------------- | ----------------------------------------------------------------------------------------------- |
| DB-Tabellen (Infra)  | `users`, `events`                               | Englisch ist korrekt — Infrastruktur / Generic Sub-Domain.                                      |
| DB-Tabellen (Domain) | `tische`, `produkte`, `produkt_varianten`       | Deutsch ist korrekt — Domänenbegriffe sind vertikal konsistent.                                 |
| Frontend-Routen      | `/admin/produkte`, `/service/tische`            | Deutsch ist korrekt — Routen repräsentieren Domänenkonzepte.                                    |
| Auth-Code            | `User`, `Role`, `OnetimePassword`               | Englisch ist korrekt — Generic Sub-Domain.                                                      |
| Auth-Routen          | `/login`, `/set-password`                       | Englisch ist korrekt — Auth ist Infrastruktur.                                                  |
| Status-Enums         | `active`, `inactive`, `deleted`                 | Englisch ist korrekt — technische Lifecycle-States, kein Domänenbegriff.                        |
| Kassenjournal        | `Historie` (Code) vs. `Kassenjournal` (Entwurf) | Bewusste Abweichung: „Historie" ist im Code und UI etabliert, beide Begriffe sind dokumentiert. |

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

| Go-Struct | TS-Typ | DB-Tabelle | API-Pfade (Admin)                                                                                         |
| --------- | ------ | ---------- | --------------------------------------------------------------------------------------------------------- |
| `User`    | `User` | `users`    | `/create-user`, `/update-user`, `/activate-user`, `/deactivate-user`, `/reset-password`, `/get-all-users` |

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

### Kassenbetrieb (Core Domain)

#### Tisch

Zentrales Aggregat im Kassenbetrieb. Trägt einen Event Stream, aus dem sich der aktuelle Zustand (Saldo, offene Positionen) berechnet. Im Festzeltbetrieb entspricht ein Tisch einer Gästegruppe, die über die gesamte Veranstaltung offen bleiben kann.

| Go-Struct | TS-Typ  | DB-Tabelle | API-Pfade (Admin)                                                                           | API-Pfade (Service)                                                                                                            |
| --------- | ------- | ---------- | ------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------ |
| `Tisch`   | `Tisch` | `tische`   | `/create-tisch`, `/update-tisch`, `/activate-tisch`, `/deactivate-tisch`, `/get-all-tische` | `/get-tisch`, `/get-aktive-tische`, `/get-tisch-historie`, `/get-tisch-saldo`, `/get-tisch-unbezahlt`, `/get-tisch-ausstehend` |

#### Bestellung

Ein Vorgang, bei dem eine Servicekraft Positionen für einen Tisch aufnimmt. Erzeugt ein `BestellungAufgenommen`-Event.

| Go-Struct    | TS-Typ       | Event-Typ                         | API-Pfad                        |
| ------------ | ------------ | --------------------------------- | ------------------------------- |
| `Bestellung` | `Bestellung` | `tisch.bestellung-aufgenommen:v1` | `/service/bestellung-aufnehmen` |

#### Position

Ein einzelner Posten innerhalb einer Bestellung: Produktvariante + Menge + Einzelpreis. Alle Felder werden als Fat Event eingefroren.

| Go-Struct  | TS-Typ     | JSON-Keys (Schlüsselfelder)                                                                    |
| ---------- | ---------- | ---------------------------------------------------------------------------------------------- |
| `Position` | `Position` | `positionId`, `varianteId`, `produktName`, `varianteName`, `kategorie`, `einzelpreis`, `menge` |

#### Ausgabe

Bestätigung, dass bestellte Positionen dem Gast übergeben wurden. Erzeugt ein `AusgabeBestaetigt`-Event.

| Go-Struct | TS-Typ    | Event-Typ                     | API-Pfad                       |
| --------- | --------- | ----------------------------- | ------------------------------ |
| `Ausgabe` | `Ausgabe` | `tisch.ausgabe-bestaetigt:v1` | `/service/ausgabe-bestaetigen` |

#### Zahlung

Kassierung einer Barzahlung. Kann sich auf einzelne Positionen beziehen. Erzeugt ein `ZahlungKassiert`-Event.

| Go-Struct | TS-Typ    | Event-Typ                   | API-Pfad                     |
| --------- | --------- | --------------------------- | ---------------------------- |
| `Zahlung` | `Zahlung` | `tisch.zahlung-kassiert:v1` | `/service/zahlung-kassieren` |

#### Stornierung

Nachträgliche Aufhebung bestellter Positionen. Nur durch Serviceleitung oder Admin. `Kommentar` ist Pflichtfeld. Erzeugt ein `StornierungErteilt`-Event.

| Go-Struct     | TS-Typ        | Event-Typ                      | API-Pfad                               |
| ------------- | ------------- | ------------------------------ | -------------------------------------- |
| `Stornierung` | `Stornierung` | `tisch.stornierung-erteilt:v1` | `/serviceleitung/stornierung-erteilen` |

#### Auszahlung

Auszahlung an den Gast, um einen negativen Saldo auszugleichen — entsteht, wenn bereits kassierte Positionen nachträglich storniert wurden (K-05). Kein Positionsbezug; freier Betrag. `Kommentar` ist Pflichtfeld. Erzeugt ein `AuszahlungGeleistet`-Event.

| Go-Struct    | Event-Typ                       | API-Pfad                             |
| ------------ | ------------------------------- | ------------------------------------ |
| `Auszahlung` | `tisch.auszahlung-geleistet:v1` | `/serviceleitung/auszahlung-leisten` |

#### Saldo

Offener Betrag eines Tisches: Bestellungen − Zahlungen − Stornierungen + Auszahlungen. Immer in Cent.

Go-Snapshot-Feld: `SaldoCents` · Go-Projektion: `ApplyEvent()` → `TischState` · API: `/service/get-tisch-state`

#### Splitrechnung / Teilzahlung

Kassiervorgang, bei dem der Gesamtsaldo eines Tisches auf mehrere Gäste aufgeteilt wird. In jotti über positionsbezogene Zahlungen abgebildet: Eine `Zahlung` kann sich auf eine Teilmenge der offenen Positionen beziehen.

#### Historie

Vollständiger, unveränderlicher Event Stream eines Tisches in chronologischer Reihenfolge.

**Synonym: Kassenjournal.** „Kassenjournal" ist der formale Fachbegriff; „Historie" ist der im Code und UI etablierte Begriff. Beide bezeichnen denselben Sachverhalt.

Go-Funktion: `GetHistoryFromEvents()` · Go-Query: `GetTischHistorie()` · API: `/service/get-tisch-historie`

#### Kommentar

Freitextnotiz zu Tischoperationen. Pflichtfeld bei Stornierung und Auszahlung, optional bei Bestellung, Ausgabe und Zahlung.

Go-Feld: `Kommentar` · JSON-Key: `"kommentar"` · TS-Feld: `kommentar`

#### Menge

Anzahl einer Produktvariante innerhalb einer Position.

Go-Feld: `Menge` · JSON-Key: `"menge"` · TS-Feld: `menge`

#### EigeneUebersicht

Kompakte KPI-Übersicht einer Servicekraft: Anzahl und Summe eigener Bestellungen sowie kassierter Zahlungen. Read Model, berechnet aus dem Event Store gefiltert auf `user_id`.

| Go-Struct          | TS-Typ             | JSON-Keys                                                                      | API-Pfad                         |
| ------------------ | ------------------ | ------------------------------------------------------------------------------ | -------------------------------- |
| `EigeneUebersicht` | `EigeneUebersicht` | `anzahlBestellungen`, `bestellungenCents`, `anzahlZahlungen`, `zahlungenCents` | `/service/get-eigene-uebersicht` |

---

### Kassenführung (Supporting Sub-Domain)

Die Kassenführung umfasst den vollständigen Lifecycle der Registerkasse — von der Eröffnung eines Abrechnungskreises über die laufende Kassenbestandsüberwachung bis zum formellen Tagesabschluss (Z-Bon). **Persistenzstrategie:** Immutable Records (INSERT-only, DB-Trigger-geschützt).

#### Abrechnungskreis

Fortlaufend nummerierte Kassensitzung, die einen Abrechnungszeitraum (typischerweise einen Veranstaltungstag) abgrenzt. DSFinV-K-Pflichtfeld. Maximal ein Abrechnungskreis kann gleichzeitig `offen` sein.

| Go-Struct (geplant) | DB-Tabelle (geplant) | JSON-Key           | API-Pfade (geplant)                 |
| ------------------- | -------------------- | ------------------ | ----------------------------------- |
| `Abrechnungskreis`  | `abrechnungskreis`   | `abrechnungskreis` | `/admin/abrechnungskreis-eroeffnen` |

#### Anfangsbestand

Wechselgeld zu Beginn einer Veranstaltung/Schicht. Pro Abrechnungskreis darf genau ein Anfangsbestand gesetzt werden. Basis für die Kassenbestandsführung.

Go-Feld (geplant): `AnfangsbestandCents` · JSON-Key: `anfangsbestandCents`

#### Kassenbestand

Read Model über den erwarteten Bargeldbestand zu jedem Zeitpunkt. Berechnung: Anfangsbestand + Zahlungen − Auszahlungen − Geldtransit − Privatentnahmen + Privateinlagen.

| Go-Struct (geplant) | JSON-Keys (geplant)                                                                              | API-Pfad (geplant)         |
| ------------------- | ------------------------------------------------------------------------------------------------ | -------------------------- |
| `Kassenbestand`     | `anfangsbestandCents`, `einnahmenCents`, `ausgabenCents`, `geldtransitCents`, `sollBestandCents` | `/admin/get-kassenbestand` |

#### Kassenbewegung

Oberbegriff für Geldtransit, Privatentnahme und Privateinlage — Bargeld-Bewegungen außerhalb des Tisch-Verkehrs. Immutable Record.

| Go-Struct (geplant) | DB-Tabelle (geplant) | JSON-Key         | Werte für `art`                                  |
| ------------------- | -------------------- | ---------------- | ------------------------------------------------ |
| `Kassenbewegung`    | `kassenbewegungen`   | `kassenbewegung` | `geldtransit`, `privatentnahme`, `privateinlage` |

#### Geldtransit

Entnahme von Bargeld aus der Kasse zur Einzahlung bei Bank oder Tresor. Reduziert den Soll-Kassenbestand. DSFinV-K-Geschäftsvorfalltyp: `Geldtransit`.

API-Pfad (geplant): `/admin/geldtransit-buchen`

#### Privatentnahme

Entnahme von Bargeld in den privaten Bereich des Vereins (nicht Bank). Fachlich analog zu Geldtransit, aber anderer DSFinV-K-Geschäftsvorfalltyp: `Privatentnahme`.

API-Pfad (geplant): `/admin/privatentnahme-buchen`

#### Privateinlage

Einlage von Bargeld in die Kasse (z. B. Nachfüllen von Wechselgeld). Erhöht den Soll-Kassenbestand. DSFinV-K-Geschäftsvorfalltyp: `Privateinlage`.

API-Pfad (geplant): `/admin/privateinlage-buchen`

#### Kassensturz

Vergleich des errechneten Soll-Bestands mit dem physisch gezählten Ist-Bestand. Bei Abweichung wird automatisch ein `DifferenzSollIst`-Vorgang gebucht. Voraussetzung für den Tagesabschluss (Z-Bon).

| Go-Struct (geplant) | DB-Tabelle (geplant) | JSON-Keys (geplant)                                     | API-Pfad (geplant)   |
| ------------------- | -------------------- | ------------------------------------------------------- | -------------------- |
| `Kassensturz`       | `kassensturz`        | `sollBestandCents`, `istBestandCents`, `differenzCents` | `/admin/kassensturz` |

#### DifferenzSollIst

Automatisch gebuchter Geschäftsvorfall beim Kassensturz, wenn Soll-Bestand ≠ Ist-Bestand. DSFinV-K-Pflicht-Geschäftsvorfalltyp.

#### Z-Bon (Tagesabschluss)

Formeller Tagesabschlussbon: aggregiert alle Transaktionen eines Abrechnungskreises nach Steuersätzen und Zahlarten, erstellt einen Stammdaten-Snapshot und erhält eine fortlaufende, nie zurücksetzbare `z_nr`. **Immutables Dokument** — kein Aggregat mit Lifecycle.

| Go-Struct (geplant) | DB-Tabelle (geplant) | JSON-Keys (geplant)                                                        | API-Pfad (geplant)      |
| ------------------- | -------------------- | -------------------------------------------------------------------------- | ----------------------- |
| `ZBon`              | `z_bons`             | `zNr`, `zeitraumVon`, `zeitraumBis`, `sollBestandCents`, `istBestandCents` | `/admin/tagesabschluss` |

> **Abgrenzung:** Der Z-Bon ersetzt die bisherige R-07-Anforderung. Er ist kein Report, sondern eine transaktionale Operation der Kassenführung (Supporting Sub-Domain).

#### X-Bon

Zwischenbericht: informativer Abruf des aktuellen Kassenstands ohne Rücksetzen. Kein Tagesabschluss im Rechtssinne. Nicht gesetzlich vorgeschrieben.

---

### Stammdaten (Supporting Sub-Domain)

#### Produkt

Artikel im Produktkatalog. Gehört zu genau einer Kategorie und enthält eine oder mehrere Varianten mit je eigenem Preis.

| Go-Struct | TS-Typ    | DB-Tabelle | API-Pfade                                                 |
| --------- | --------- | ---------- | --------------------------------------------------------- |
| `Produkt` | `Produkt` | `produkte` | `/create-produkt`, `/update-produkt`, `/get-all-produkte` |



#### Variante

Konkrete Ausprägung eines Produkts mit eigenem Namen und Preis in Cent (z. B. Produkt „Cola" → Varianten „0,3 l" / „0,5 l").

| Go-Struct  | TS-Typ     | DB-Tabelle          | API-Pfade                                                                            |
| ---------- | ---------- | ------------------- | ------------------------------------------------------------------------------------ |
| `Variante` | `Variante` | `produkt_varianten` | `/create-variante`, `/update-variante`, `/activate-variante`, `/deactivate-variante` |



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

| Go-Package                 | DB-Tabelle        | API-Pfade                                                    |
| -------------------------- | ----------------- | ------------------------------------------------------------ |
| `repository/favorit_repo/` | `tisch_favoriten` | `/service/favorit-hinzufuegen`, `/service/favorit-entfernen` |

TS-Repräsentation: `istFavorit: boolean` in `AktiverTischMitFavorit` (kein eigener Typ).

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

| Begriff                     | Go-Typ                | DB-Feld / -Tabelle                      | JSON-Key    | Bedeutung                                                                                                                                                                         |
| --------------------------- | --------------------- | --------------------------------------- | ----------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **KassenID / Seriennummer** | `KassenID` (`string`) | `system_config.value` (key=`kassen_id`) | `kassen_id` | Eindeutige UUID-v4 der jotti-Instanz. Beim ersten Containerstart generiert. Für ELSTER-Meldung und TSE-Protokoll.                                                                 |
| **TSEClient**               | Interface (Go)        | —                                       | —           | Go-Interface mit Methoden `StartTransaction`, `UpdateTransaction`, `FinishTransaction`. Anbieter-spezifische Implementierungen (z. B. `FiskalyTSEClient`) erfüllen das Interface. |

#### Transaktions-Lifecycle

Die TSE-Kommunikation folgt einem strikten Lifecycle. Jeder Kassiervorgang durchläuft diese Schritte:

- **StartTransaction:** API-Aufruf an die TSE bei Beginn eines neuen Vorgangs. Eröffnet die Transaktion und gibt die `TSE_TANR` zurück.
- **UpdateTransaction:** Optionaler API-Aufruf, um einer offenen Transaktion neue Daten hinzuzufügen. Nur bei bestimmten Vorgangsarten erlaubt.
- **FinishTransaction:** Zwingender Abschluss-Aufruf, der die finale Signatur (Prüfwert) der TSE generiert.
- **processType:** Strikt normierter String, der der TSE die Art des Vorgangs mitteilt (z. B. `Kassenbeleg-V1`, `Bestellung-V1`, `SonstigerVorgang-V1`).
- **processData:** Payload-String im BSI-Format (UTF-8, Punkt als Dezimaltrenner) mit Beträgen, Mengen und Steuersätzen.

#### Abrechnungsstruktur

| Begriff              | Go-Struct / Go-Typ | DB-Feld / -Tabelle | JSON-Key           | Bedeutung                                                                                                                                                                                                         |
| -------------------- | ------------------ | ------------------ | ------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **ABRECHNUNGSKREIS** | `Abrechnungskreis` | `abrechnungskreis` | `abrechnungskreis` | Fortlaufend nummerierte Kassensitzung, die einen Abrechnungszeitraum (typisch: einen Veranstaltungstag) abgrenzt. DSFinV-K-Pflichtfeld. Verbindet logisch zusammengehörige Tisch-Vorgänge. Siehe → Kassenführung. |
| **Tagesabschluss**   | `ZBon` (geplant)   | `z_bons` (geplant) | `zNr`              | Formaler Abschluss eines ABRECHNUNGSKREIS. Erzeugt ein immutables Dokument (Z-Bon) mit fortlaufender `z_nr`. Gehört zur Kassenführung (Supporting Sub-Domain), nicht zum Reporting.                                         |
| **Z-Bon**            | `ZBon` (geplant)   | `z_bons` (geplant) | `zNr`              | Tagesabschlussbon: aggregiert alle Transaktionen nach Steuersätzen und Zahlarten (`businesscases.csv`). Immutables Dokument — kein Reset von Events.                                                              |
| **X-Bon**            | —                  | —                  | —                  | Zwischenbericht: informativer Abruf des aktuellen Kassenstands ohne Rücksetzen. Kein Tagesabschluss im Rechtssinne.                                                                                               |
| **Bonkopf / Bonpos** | —                  | —                  | —                  | DSFinV-K-Aufteilung: Bonkopf enthält Metadaten und Gesamtsummen des Belegs; Bonpos listet die einzelnen Artikelzeilen (`lines.csv`).                                                                              |

#### Steuern

| Begriff          | Go-Typ              | DB-Feld      | JSON-Key       | Bedeutung                                                                                                                                          |
| ---------------- | ------------------- | ------------ | -------------- | -------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Steuersatz**   | `Steuersatz` (Enum) | `steuersatz` | `steuersatz`   | Steuerklasse eines Produkts. Enum-Werte: `standard` (19 %), `ermaessigt` (7 %), `befreit` (0 %). Wird als Fat Event in die Bestellung eingefroren. |
| **Steuerbetrag** | `int` (Cent)        | —            | `steuerbetrag` | Berechneter Steuerbetrag in Cent für eine Position oder einen Vorgang. Immer in Cent, niemals als Float.                                           |
| **Nettobetrag**  | `int` (Cent)        | —            | `nettobetrag`  | Betrag vor Steuerabzug. Immer in Cent.                                                                                                             |

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

- **Snapshot:** Vorberechneter Zwischenstand des Tisch-Zustands als Performance-Optimierung. Beim Replay wird nur ab dem letzten Snapshot gelesen. Gespeichert als eigener Event-Typ (`tisch.snapshot:v1`) in der `events`-Tabelle.

---

## Geplant (nicht implementiert)

Die folgenden Begriffe sind in der Ubiquitous Language definiert, aber noch nicht im Code implementiert.

### Ausgabe (Teil des Kassenbetrieb-Context)

| Begriff                 | Bedeutung                                                                                                       |
| ----------------------- | --------------------------------------------------------------------------------------------------------------- |
| **Bon**                 | Gedruckter Beleg mit Tisch, Servicekraft, Positionen, Mengen, Zeitstempel und optionalem Kommentar.             |
| **Küchendisplay (KDS)** | Echtzeit-Anzeige offener Bestellungen an der Ausgabestation, gruppiert nach Tisch und gefiltert nach Kategorie. |
| **Zubereitungsstatus**  | Status einer Position an der Ausgabestation: offen → in Zubereitung → fertig.                                   |
| **Ausgabestation**      | Physischer Ort (Küche, Getränketheke), an dem Positionen zubereitet und ausgegeben werden.                      |

### Abrechnung (Teil der Kassenführung)

| Begriff             | Bedeutung                                                                                              |
| ------------------- | ------------------------------------------------------------------------------------------------------ |
| **Tagesabrechnung** | Übersicht über Gesamtumsatz, Stornierungen und Umsatz pro Servicekraft — jederzeit vom Admin abrufbar. |
| **Umsatz**          | Summe aller registrierten Zahlungen in einem bestimmten Zeitraum. Immer in Cent.                       |
| **Stornoquote**     | Verhältnis von Stornierungsbetrag zu Bestellsumme. Indikator für Fehler oder Unregelmäßigkeiten.       |
| **Export**          | CSV-Download von Umsätzen, Bestellungen und Artikeldaten für die Vereinsbuchhaltung.                   |

> **Hinweis:** Der Tagesabschluss (Z-Bon) ist kein Reporting-Vorgang, sondern eine transaktionale Operation der Kassenführung (Supporting Sub-Domain). Siehe → Kassenführung.

---

## Erweitertes Fach-Glossar

Begriffe aus dem POS/Gastronomie-Umfeld, die **nicht im Scope von jotti** sind (bewusste Abgrenzung — siehe `docs/anforderungen.md` §6). Für die Einarbeitung neuer Teammitglieder in den Fachkontext.

**Vereinswesen:**

- **Zuwendungsbestätigung (Spendenbescheinigung):** Amtliches Dokument, das nur gemeinnützige Vereine ausstellen dürfen, damit Spender ihre Zahlungen ohne Gegenleistung steuerlich absetzen können.
- **Sponsoring:** Leistungsaustausch (kein Spendencharakter): Unternehmen zahlt Geld oder Sachmittel, Verein erbringt als Gegenleistung Werbung.
- **Aufwandsspende:** Helfer verzichtet auf Kostenerstattung (z. B. Fahrtkosten) und überlässt den Betrag dem Verein als Spende.
- **Rückspende:** Verzicht auf vertraglich vereinbartes Honorar (z. B. Übungsleiterpauschale) zugunsten des Vereins; der Helfer erhält dafür eine Spendenbescheinigung.

**Warenwirtschaft:**

- **Warenwirtschaft (WaWi):** An die Kasse angebundenes System, das Lagerbestände führt und bei jedem Verkauf automatisch Artikel abbucht.
- **Inventur:** Physische Zählung von Lagerartikeln zum Abgleich von System-Soll-Bestand und realem Ist-Bestand.
- **Warengruppe:** Logische Artikelgruppierung für zentrale Steuersatzverwaltung und betriebswirtschaftliche Auswertung.
- **Pfand / Pfandrückzahlung:** Gesondert gebuchte Nebenleistung für Mehrwegbehälter (Aufschlag beim Verkauf, Erstattung als negativer Betrag bei Rückgabe).
- **Einzweck- / Mehrzweckgutschein:** Umsatzsteuerlich wichtige Unterscheidung: Einzweckgutscheine werden beim Verkauf versteuert, Mehrzweckgutscheine erst bei Einlösung.

**Hardware:**

- **Kartenterminal (EFT-Terminal):** Gerät für bargeldlose Zahlungen (EC-/Kreditkarte) mit ZVT- oder O.P.I.-Schnittstelle zur Kassensoftware.
- **Kassenlade (Geldlade):** Physische Geldschublade, die über den Bondrucker elektronisch angesteuert wird und sich bei Bar-Abschlüssen öffnet.
- **Gangsteuerung:** Softwarefunktion, mit der Servicekräfte der Küche die Reihenfolge der Gänge (Vorspeise → Hauptgang → Dessert) mitteilen.
