# Ubiquitous Language: jotti

Dieses Dokument ist die verbindliche Referenz für die Ubiquitous Language des jotti-Projekts, für Entwickler, Agenten und alle Projektbeteiligten. Es definiert die Fachbegriffe der Domäne, ihre Code-Repräsentationen und die Sprachkonventionen pro Schicht.

Die Ubiquitous Language ist ein Living Document: Sie wird fortlaufend aktualisiert, wenn sich Begriffe, Strukturen oder Konventionen ändern.

## Sprachkonventionen

1. **Domänenbegriffe sind deutsch.** Alle Fachbegriffe der Kasse, der Stammdaten und der Gastronomie-Domäne werden auf Deutsch benannt, in Code, Dokumentation und Kommunikation. Beispiele: `Bestellung`, `Tisch`, `Zahlung`, `Position`, `Ausgabe`, `Stornierung`, `Saldo`, `Kassensitzung`, `Kassenjournal`.

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

> **Pfadkonvention:** Dateipfade sind relativ angegeben, `domain/…` und `api/…` liegen unter `backend/`, `src/…` unter `frontend/`, `migrations/…` unter `database/`.

## Begriffsdefinitionen

### Vereinswesen & Steuerliche Sphären

Das Finanzamt teilt einen Verein in vier steuerliche Sphären, die Buchführungspflichten und Steuersätze bestimmen, Details und jotti-Bezug: [steuerrecht.md §8](steuerrecht.md#8-bedeutung-für-jotti).

- **Gemeinnützigkeit:** Steuerbegünstigter Status eines Vereins (selbstlose, satzungsgemäße Tätigkeit für die Allgemeinheit).
- **Ideeller Bereich:** Steuerfreier Kernbereich ohne wirtschaftliche Tätigkeit (Spenden, Mitgliedsbeiträge).
- **Wirtschaftlicher Geschäftsbetrieb (WGB):** In der Regel steuerpflichtiger Bereich, in dem der Verein wie ein Unternehmen agiert (z. B. Getränke- und Essensverkauf auf dem Vereinsfest), jottis primärer Einsatzbereich.
- **Zweckbetrieb:** Steuerbegünstigter wirtschaftlicher Geschäftsbetrieb, der unmittelbar dem gemeinnützigen Zweck dient.
- **Vermögensverwaltung:** Steuerfreie, passive Einnahmen aus Vereinsvermögen (Zinsen, Mieten).
- **Kleinunternehmerregelung (§ 19 UStG):** Befreiung von der Umsatzsteuerpflicht bei geringen Umsätzen, beeinflusst die korrekte `Steuersatz`-Konfiguration.

### Akteure & Rollen

**Fachliche Akteure (kein Code-Mapping):**

- **Servicekraft / Bedienung:** Freiwillige Helfer, die Bestellungen aufnehmen, kassieren und Ausgaben bestätigen. Systemrolle `service`.
- **Serviceleitung:** Erfahrene Servicekraft mit erweiterten Rechten für Stornierungen. Systemrolle `serviceleitung`.
- **Kassenwart / Schatzmeister:** Vorstandsmitglied, verantwortlich für Finanzen, Buchhaltung und Steuererklärungen. Typischerweise Systemrolle `admin`.
- **Vorstand:** Gesetzliches Vertretungsorgan des Vereins; haftet persönlich für die Einhaltung steuerlicher Pflichten. Typischerweise Systemrolle `admin`.

**Systemrollen (mit Code-Mapping):**

#### Benutzer

Person mit Zugang zum System, identifiziert durch einen eindeutigen Benutzernamen.

Go-Struct: `User` · TS-Typ: `User` · DB-Tabelle: `users`

#### Rolle

Berechtigungsstufe eines Benutzers: Admin (`admin`), Serviceleitung (`serviceleitung`), Servicekraft (`service`).

Go-Typ: `Role` mit `AdminRole`, `ServiceleitungRole`, `ServiceRole` · DB-Enum: `UserRole` (`'admin'`, `'serviceleitung'`, `'service'`)

Die vollständige Berechtigungsmatrix steht in [handbuch.md §5.1](handbuch.md#51-rollen-und-berechtigungsmatrix).

#### Einmalpasswort

Vom Admin generiertes Passwort aus 8 Zeichen (Kleinbuchstaben und Ziffern ohne verwechselbare Zeichen) für die Erstanmeldung oder das Zurücksetzen eines Passworts.

Go-Feld: `OnetimePasswordHash` · DB-Spalte: `onetime_password_hash` · TS-Schema: `OnetimePasswordSchema`

#### Token

JWT (JSON Web Token) mit Benutzer-ID und Rolle. 12 Stunden gültig. Reiner Infrastruktur-Begriff, Englisch im Code ist korrekt.

---

### Kasse (Core Domain)

Die Event-Feldschemata (Felder, Typen, Constraints) aller Kasse-Events stehen kanonisch im Code: `backend/domain/kasse/*_events.go`. Eine kompakte Event-Übersicht (Typ, Subject, Semantik) gibt [handbuch.md §3.6](handbuch.md#36-domain-events). Die folgenden Einträge geben die Begriff↔Code-Mappings.

#### Tisch (Stammdaten)

Reine Stammdaten-Entität: ein physischer Ort, an dem Gäste sitzen. Hat einen Namen, Status (active/inactive/deleted) und wird vom Admin verwaltet. Im Kasse-Kontext wird der Tisch nur über seine ID referenziert, die `tisch_id` fließt in das Subject der Tisch-Session ein.

Go-Struct: `Tisch` · TS-Typ: `Tisch` · DB-Tabelle: `tische`

#### Tisch-Session (Abrechnungskreis)

Das Event-Sourced Aggregat im Kasse-Kontext. Bildet alle Geschäftsvorfälle (Bestellungen, Zahlungen, Stornierungen, Umbuchungen, Ausgaben) eines Tisches innerhalb einer Kassensitzung ab. Entsteht implizit mit der ersten Bestellung.

| Go-Struct      | TS-Typ         | DB-Projektion    | Subject-Format                       |
| -------------- | -------------- | ---------------- | ------------------------------------ |
| `TischSession` | `TischSession` | `tisch_sessions` | `kassensitzung-{nr}/tisch-{tischId}` |

> **Hinweis `domain/table/`:** Das Paket existiert weiterhin für Tisch-Stammdaten (`tisch.go`). Die Kasse-Logik (Event-Sourcing, Tisch-Sessions, Kassensitzung) liegt in `domain/kasse/`; `domain/table/` enthält die Tisch-Stammdaten-Entität `Tisch` sowie die Read-Model-Structs `AktiverTisch`/`AktiverTischMitFavorit`.

#### Direktverkauf

Schlankes Event-Sourced Aggregat im Kasse-Kontext für den Barverkauf an der Theke: bestellen, zahlen und ausgeben in einem Schritt, ohne Tisch und ohne Projektion. Jeder Verkauf ist ein eigener Event-Stream mit eigener UUID. Direktverkauf hat keine `Verkaufsstelle`-Stammdaten-Entität.

| Go-Struct | TS-Typ | Event-Typ (Verkauf)          | Subject-Format                            |
| --------- | ------ | ---------------------------- | ----------------------------------------- |
| —         | —      | `direktverkauf-getaetigt:v1` | `kassensitzung-{nr}/direktverkauf-{uuid}` |
| —         | —      | `direktverkauf-storniert:v1` | `kassensitzung-{nr}/direktverkauf-{uuid}` |

> **Verkauf:** die fachliche Einheit eines Direktverkaufs (ein Stream, ein `verkaufId`). Kein eigenes Domain-Struct, der Verkauf existiert nur als Event-Stream im Kassenjournal. `direktverkauf-getaetigt:v1` ist `version = 1`; positionsgenaue Stornierungen sind Folge-Versionen im selben Stream.

> **Direktverkauf-Stornierung:** positionsgenaue Korrektur/Rückgabe eines Verkaufs durch Serviceleitung/Admin (`direktverkauf-storniert:v1`, Fat-Positionen, sofort kassenwirksam, die Bar-Rückgabe ist Teil des Vorgangs). API-Input: `PositionRef`; Validierung per On-Demand-Replay (`ComputeNichtStornierteVerkaufPositionen`). Regeln im Detail → [handbuch.md §3.3](handbuch.md#33-subject-design-hierarchische-subjects).

> **Direktverkauf-Bondruck (Ableitungsregel):** Ist die Druckstation `abholbon` konfiguriert, erzeugt `direktverkauf-getaetigt:v1` Abholbon(s) gemäß deren Bonmodus; sonst Arbeitsbons an die Produktstationen; ohne konfigurierte Stationen entsteht kein Druckauftrag.

#### Abholbon

Nicht-fiskalischer Bon für die Warenübergabe beim Direktverkauf. Festes Label „Direktverkauf", keine Preise, `bon_art = 'arbeitsbon'`, gedruckt an der Druckstation `abholbon` (Bonmodus `pro_bestellung` = ein Sammel-Abholbon, `pro_position` = ein Bon je Position).

#### Arbeitsmodus

Reiner Oberflächenbegriff im Service-Frontend: Eine Servicekraft arbeitet im Service-Bereich in genau einem Modus, Tischservice oder Direktverkauf. Die Wahl wird pro Gerät gemerkt (BYOD, überlebt Logout/Login) und ausschließlich über das Benutzermenü gewechselt. Kein Domain-Konzept, kein Backend, keine `Verkaufsstelle`-Entität.

TS-Typ: `Arbeitsmodus` (`'tischservice' | 'direktverkauf'`) · Persistenz: `localStorage` (`frontend/src/lib/arbeitsmodus.ts`)

#### Bestellung

Ein Vorgang, bei dem eine Servicekraft Positionen für einen Tisch aufnimmt.

| Go-Struct    | TS-Typ       | Event-Typ                   |
| ------------ | ------------ | --------------------------- |
| `Bestellung` | `Bestellung` | `bestellung-aufgenommen:v1` |

#### Position

Ein einzelner Posten innerhalb einer Bestellung: Produktvariante + Menge + Einzelpreis. Alle Felder werden als Fat Event eingefroren.

| Go-Struct  | TS-Typ     | JSON-Keys (Schlüsselfelder)                                                                                  |
| ---------- | ---------- | ------------------------------------------------------------------------------------------------------------ |
| `Position` | `Position` | `positionId`, `varianteId`, `produktName`, `varianteName`, `kategorie`, `steuersatz`, `einzelpreis`, `menge` |

#### Besteller (bestellende Servicekraft)

Die Servicekraft, die eine Bestellung aufgenommen hat. Reines Projektions- und Anzeigekonzept: Jede offene `Position` trägt den Besteller als eingefrorenen Username aus dem Event-Umschlag des `bestellung-aufgenommen`-Events. Die Event-Form bleibt unverändert, spätere Umbenennungen ändern alte Positionen nicht. Grundlage für die persönliche Erledigt-Sicht, die Sortierung „eigene zuerst" beim Kassieren und Ausgeben und die Schichtende-Prüfung im Live-Dashboard.

Go-Projektion-Felder: `Position.BestellerUserID`, `Position.BestellerName` · JSON/TS: `bestellerUserId`, `bestellerName`

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

Nachträgliche Aufhebung bestellter Positionen, nur durch Serviceleitung oder Admin (Endpunkt `/stornierung-erteilen`). Eine „Stornieren"-Anforderung wird serverseitig nach Bezahlstatus aufgeteilt: noch unbezahlte Mengen werden geldneutral korrigiert, bereits bezahlte als kassenwirksame Warenrücknahme zurückgenommen. Die entstehenden Events werden atomar geschrieben (alles-oder-nichts), jedes mit eigener TSE-Transaktion.

**Warenrücknahme (kassenwirksam):** Rückgabe bereits bezahlter Positionen als negativer Umsatz am Ursprungssteuersatz mit Bar-Rückgabe im selben Beleg. `ZahlungID` referenziert genau die begleichende Zahlung, deren Mengen zurückgenommen werden (eine Warenrücknahme je betroffener Zahlung, FIFO nach Zahlung, älteste zuerst). Der Rückgabebetrag folgt aus den Positionen, ist nicht frei wählbar. `Kommentar` ist Pflichtfeld. Signiert als `Kassenbeleg-V1` mit negativem Bruttoumsatz je Steuersatz und negativer Bar-Zahlung.

**Korrektur (geldneutral):** Stornierung noch unbezahlter Positionen ohne Geld- und Umsatzwirkung. Reduziert den offenen Betrag und nimmt die Positionen aus den aktiven Listen. `Kommentar` ist optional. Signiert als `Bestellung-V1` (ohne Zahlungszeile). In der Historie erscheint die Korrektur als „Stornierung".

| Vorgang        | Go-Struct     | TS-Typ        | Event-Typ                  |
| -------------- | ------------- | ------------- | -------------------------- |
| Warenrücknahme | `Stornierung` | `Stornierung` | `stornierung-erteilt:v1`   |
| Korrektur      | `Stornierung` | `Stornierung` | `bestellung-korrigiert:v1` |

#### Umbuchung

Geldneutrale Verschiebung noch unbezahlter Positionen zwischen zwei Tischen, durch die Servicekraft (Endpunkt `/bestellung-umbuchen`, Rolle `service`). Quell- und Zielstrom erhalten je ein Event mit derselben `UmbuchungID`: der Quelltisch verliert die Positionen (Saldo sinkt), der Zieltisch nimmt sie geldneutral auf (Saldo steigt um denselben Betrag). Beide Events werden atomar geschrieben und als `Bestellung-V1` signiert (ohne Zahlungszeile). `Kommentar` ist optional. In Historie und Export erscheint der Vorgang als „Umbuchung", nicht als „Stornierung".

| Go-Struct   | TS-Typ      | Event-Typ                 |
| ----------- | ----------- | ------------------------- |
| `Umbuchung` | `Umbuchung` | `bestellung-umgebucht:v1` |

#### Saldo

Offener Betrag einer Tisch-Session: Summe der noch unbezahlten Positionen, immer in Cent und nie negativ. Die Warenrücknahme bezahlter Positionen verändert den Saldo nicht (sie wirkt auf den Kassenbestand), die geldneutrale Korrektur und die Umbuchung reduzieren ihn. Die Saldo-Formel ist kanonisch in [handbuch.md §3.7](handbuch.md#37-invarianten) definiert.

Go-Projektion-Feld: `SaldoCents` · Berechnung: `ApplyEvent()` → `TischSession`

#### Splitrechnung / Teilzahlung

Kassiervorgang, bei dem der Gesamtsaldo eines Tisches auf mehrere Gäste aufgeteilt wird. In jotti über positionsbezogene Zahlungen abgebildet: Eine `Zahlung` kann sich auf eine Teilmenge der offenen Positionen beziehen.

#### Historie

Vollständiger, unveränderlicher Event Stream einer Tisch-Session in chronologischer Reihenfolge.

**Synonym: Kassenjournal.** „Kassenjournal" ist der formale Fachbegriff (die DB-Tabelle `kassenjournal` enthält alle Events); „Historie" ist der im Code und UI etablierte Begriff für die Tisch-spezifische Ansicht.

Go-Funktion: `GetHistorieFromEvents()` · Application-Query: `GetTischHistorie()` · API: `/service/get-tisch-historie`

#### Weitere Typen und Felder (Kasse)

| Begriff              | Bedeutung                                                                 | Code-Mapping                                                                                                   |
| -------------------- | ------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------- |
| Kommentar            | Freitextnotiz; Pflicht bei Warenrücknahme (kassenwirksamer Storno), sonst optional | Go `Kommentar` · JSON/TS `kommentar`                                                                  |
| Menge                | Anzahl einer Produktvariante innerhalb einer Position                     | Go `Menge` · JSON/TS `menge`                                                                                   |
| PositionRef          | Referenz auf eine Position (ID + Menge) für Zahlung, Ausgabe, Stornierung | Go/TS `PositionRef` · JSON `positionId`, `menge`                                                               |
| HistorieEintrag      | Eintrag der Tisch-Historie, typisiert nach Art                            | Go `HistorieEintrag` · Enum `Art`: `bestellung`, `zahlung`, `stornierung`, `umbuchung`, `ausgabe`              |
| EigeneUebersicht     | KPI-Read-Model einer Servicekraft: eigene Bestellungen und Zahlungen      | Go/TS `EigeneUebersicht` · JSON `anzahlBestellungen`, `bestellungenCents`, `anzahlZahlungen`, `zahlungenCents` |
| AktiverTisch         | Kompakte Tisch-Darstellung mit Saldo für die Tischübersicht (Read Model)  | Go `AktiverTisch` · TS `AktiverTischMitFavorit` (mit `istFavorit`)                                             |
| BestellPositionInput | Frontend-Eingabetyp einer Bestellposition (Produkt + Variante + Menge)    | TS `BestellPositionInput` · JSON `produktId`, `varianteId`, `menge`                                            |

---

### Kasse: Kassensitzung und Kassenbestand

Kassensitzung-Events werden unter dem Subject `kassensitzung-{nr}` im Kassenjournal persistiert. Feldschemata kanonisch in `backend/domain/kasse/kassensitzung_events.go`.

#### Kassensitzung

Global nummerierter Betriebstag, der einen Abrechnungszeitraum (typischerweise einen Veranstaltungstag) abgrenzt. Maximal eine Kassensitzung ist gleichzeitig `offen`; ohne offene Kassensitzung ist der Kassenbetrieb gesperrt. Lifecycle und `z_nr`-Regeln → [handbuch.md §3.5](handbuch.md#35-kassensitzung-lifecycle).

| Go-Struct            | DB-Tabelle        | Subject-Format       | Eröffnungs-Event             |
| -------------------- | ----------------- | -------------------- | ---------------------------- |
| `Kassensitzung`      | `kassensitzungen` | `kassensitzung-{nr}` | `kassensitzung-eroeffnet:v1` |

#### Bezeichnung

Freitextlabel einer Kassensitzung (z. B. „Maihock 2026").

Go-Feld: `Bezeichnung` · DB-Spalte: `kassensitzungen.bezeichnung` · JSON-Key: `"bezeichnung"`

#### Abrechnungskreis

DSFinV-K-Pflichtfeld (`ABRECHNUNGSKREIS`), identisch mit der Tisch-Session: pro Tisch und Kassensitzung existiert ein Abrechnungskreis. Der Export-Wert wird aus dem Tischnamen abgeleitet (z. B. `Tisch 42`).

#### Anfangsbestand

Wechselgeld zu Beginn einer Kassensitzung. Wird bei der Eröffnung gesetzt, als Feld `betragCents` im Event `kassensitzung-eroeffnet:v1`, kein eigenes Event.

#### Kassenbestand

Erwarteter Bargeldbestand (Soll), on-demand per SQL-Aggregation aus dem Kassenjournal berechnet, kein eigenes Read Model. Formel → [handbuch.md §3.9](handbuch.md#39-kassenbestand-read-model).

JSON-Key: `sollBestandCents`

#### Geldtransit (Kassenbewegung)

Bargeld-Bewegung außerhalb des Tisch-Verkehrs: Einlage (z. B. Wechselgeld nachfüllen, erhöht den Soll-Bestand) oder Entnahme (z. B. Abschöpfung in Bank/Tresor, reduziert ihn). `Kommentar` ist Pflichtfeld. DSFinV-K-Geschäftsvorfalltyp: `Geldtransit`.

| Event-Typ                | JSON-Key `richtung`     | API-Pfad                    |
| ------------------------ | ----------------------- | --------------------------- |
| `geldtransit-gebucht:v1` | `einlage` \| `entnahme` | `/admin/geldtransit-buchen` |

#### Kassenabschluss

Ein-Klick-Operation, die die offene Kassensitzung beendet: Kassensturz, Differenzbuchung (bei Differenz ≠ 0) und Tagesabschluss in fester Schreibreihenfolge. Als erste Handlung setzt sie die Sitzung auf `wird_abgeschlossen` (Barriere: ab diesem Commit werden Buchungs-Events abgelehnt); eine umschließende Transaktion über die Abschluss-Events gibt es bewusst nicht (→ [handbuch.md §3.10](handbuch.md#310-kassensturz), [§3.11](handbuch.md#311-tagesabschluss-z-bon)).

Go: `KasseAbschliessen` · API-Pfad: `/admin/kasse-abschliessen`

#### Kassensturz

Vergleich des errechneten Soll-Bestands mit dem physisch gezählten Ist-Bestand; erster Schritt des Kassenabschlusses → [handbuch.md §3.10](handbuch.md#310-kassensturz).

| Event-Typen                                                                          | JSON-Keys                                               |
| ------------------------------------------------------------------------------------ | ------------------------------------------------------- |
| `kassensturz-durchgefuehrt:v1` (+ `differenz-soll-ist-gebucht:v1` bei Differenz ≠ 0) | `sollBestandCents`, `istBestandCents`, `differenzCents` |

#### DifferenzSollIst

Automatisch erzeugtes Event (`differenz-soll-ist-gebucht:v1`) beim Kassensturz, wenn Soll-Bestand ≠ Ist-Bestand. DSFinV-K-Pflicht-Geschäftsvorfalltyp.

#### Z-Bon (Tagesabschluss)

Formeller Tagesabschluss: aggregiert die Kassensitzung und schließt sie ab (Status → `abgeschlossen`). Kein Report, sondern das abschließende Event des Kassenabschlusses (→ [handbuch.md §3.11](handbuch.md#311-tagesabschluss-z-bon)).

| Event-Typ                    | DB-Feld                | JSON-Keys (Auszug)                                                                                                    |
| ---------------------------- | ---------------------- | --------------------------------------------------------------------------------------------------------------------- |
| `tagesabschluss-erstellt:v1` | `kassensitzungen.z_nr` | `zNr`, `zeitraumVon`, `zeitraumBis`, `umsatzGesamtCents`, `stornierungCents`, `geldtransitCents` |

#### X-Bon

Zwischenbericht: informativer Abruf des aktuellen Kassenstands ohne Abschluss. Kein Tagesabschluss im Rechtssinne, nicht gesetzlich vorgeschrieben.

---

### Stammdaten (Supporting Sub-Domain)

#### Produkt

Artikel im Produktkatalog. Gehört zu genau einer Kategorie und enthält eine oder mehrere Varianten mit je eigenem Preis.

Go-Struct: `Produkt` · TS-Typ: `Produkt` · DB-Tabelle: `produkte`

#### Variante

Konkrete Ausprägung eines Produkts mit eigenem Namen und Preis in Cent (z. B. Produkt „Cola" → Varianten „0,3 l" / „0,5 l").

Go-Struct: `Variante` · TS-Typ: `Variante` · DB-Tabelle: `produkt_varianten`

#### Kategorie

Gruppierung von Produkten. Aktuell drei feste Kategorien.

| Go-Typ      | DB-Enum (`ProduktKategorie`)           | Go-Konstanten                                               |
| ----------- | -------------------------------------- | ----------------------------------------------------------- |
| `Kategorie` | `'essen'`, `'getraenk'`, `'sonstiges'` | `EssenKategorie`, `GetraenkKategorie`, `SonstigesKategorie` |

#### Preis

Geldbeträge werden ausnahmslos als ganzzahlige Cent-Werte gespeichert, niemals als Fließkommazahlen. 3,50 € = 350 Cent.

**Konvention:** Alle Preis-Felder tragen das Suffix `*Cents`, z. B. `PreisCents`, `GesamtPreisCents`, `SaldoCents`, `GesamtZahlungCents`, `GesamtStornierungCents`.

#### Soft-Delete

Logisches Löschen: Datensätze werden nicht physisch entfernt, sondern durch den Status `deleted` markiert. Ermöglicht Referenzintegrität und historische Auswertung.

DB-Enum: `EntityStatus` (`'active'`, `'inactive'`, `'deleted'`) · Go-Konstanten: `ActiveStatus`, `InactiveStatus`, `DeletedStatus` (in `domain/table`, `domain/product`, `domain/user`)

#### Favorit

Benutzerspezifische Markierung einer Servicekraft für einen Tisch ("Meine Tische"). Kein Aggregat, keine Events, einfache CRUD-Relation.

Go-Package: `repository/favorit_repo/` · DB-Tabelle: `tisch_favoriten` · TS: `istFavorit: boolean` in `AktiverTischMitFavorit` (kein eigener Typ)

---

### Reporting (Read Model)

Reporting-Daten werden on-demand per SQL-Aggregation aus dem Kassenjournal berechnet. Kein eigener Event Stream, reines Read Model. Alle Typen existieren spiegelbildlich als Go-Struct (`domain/reporting/`) und TS-Typ.

| Begriff             | Bedeutung                                                                                                  |
| ------------------- | ---------------------------------------------------------------------------------------------------------- |
| ReportingData       | Vollständiger Reporting-Datensatz einer Kassensitzung: Summary + Breakdowns + Stornierungen                |
| Summary             | Aggregierte Kennzahlen einer Kassensitzung (Umsatz, Stornierungen, offene Salden, Anzahlen)                |
| Breakdowns          | Aufschlüsselung des Umsatzes: `UmsatzProServicekraft []UmsatzServicekraft`, `UmsatzProTisch []UmsatzTisch` |
| UmsatzServicekraft  | Umsatz einer einzelnen Servicekraft (Zahlungen, Anzahl)                                                    |
| UmsatzTisch         | Umsatz eines einzelnen Tisches (Zahlungen, Anzahl)                                                         |
| StornierungDetail   | Einzelne Stornierung im Reporting (Zeitpunkt, Tisch, Benutzer, Betrag, Kommentar, Positionen); `barRueckgabe` markiert die kassenwirksame Warenrücknahme gegenüber der geldneutralen Korrektur |
| StornierungPosition | Position innerhalb einer StornierungDetail (Produktname, Variantenname, Menge, Einzelpreis)                |

---

### Bondruck & Infrastruktur

Bondruck ist der Oberbegriff für zwei fachlich getrennte Bon-Familien auf einer gemeinsamen Druck-Infrastruktur: den operativen Arbeitsbon und den gesetzlichen Kassenbeleg. Architektur und Abgrenzung: [handbuch.md §4.6](handbuch.md#46-bondruck-arbeitsbon-und-kassenbeleg-k-12).

#### Arbeitsbon

Operativer, nicht-fiskalischer Bon an eine Ausgabestation (Küche, Theke). Trägt keine Preise, nur die zuzubereitende/auszugebende Ware (Quelle, Artikel, Menge, Kommentar, Uhrzeit, Bedienung). Wird automatisch bei Bestellaufnahme erzeugt. Kein Beleg im Sinne von § 146a AO.

#### Kassenbeleg

Fiskalischer Zahlungsbeleg (§ 146a Abs. 2 AO, § 6 KassenSichV) für den Gast: alle Positionen mit Preisen, Steueraufteilung (F-07), Vereinsdaten (K-20), Kassen-Seriennummer (F-01) und (sofern TSE konfiguriert) TSE-Pflichtfelder inkl. QR-Code (F-02). Wird auf Anforderung pro Kassiervorgang gedruckt, am Fest greift meist die Belegausgabe-Befreiung (→ [compliance.md §5.1](compliance.md#51-gesetzliche-grundlage)). DSFinV-K-`processType`: `Kassenbeleg-V1`.

#### Druckstation

Konfigurierter Drucker je Kategorie: drei Produktstationen für Arbeitsbons plus die Stationen `kassenbeleg` und `abholbon`. CRUD-Entität.

DB-Tabelle: `druckstationen` · DB-Enum `DruckstationKategorie`: `'essen'`, `'getraenk'`, `'sonstiges'`, `'kassenbeleg'`, `'abholbon'`

#### Bonmodus

Druckmodus für Arbeitsbons/Abholbons: einzelner Bon pro Position oder ein gesammelter Bon pro Bestellung. Für die Kassenbeleg-Station entfällt er (NULL).

DB-Enum: `'pro_position'`, `'pro_bestellung'`

#### Druckauftrag

Konkreter Druckjob in der Outbox, Single Source of Truth für alle Druckjobs, Arbeitsbon und Kassenbeleg. Das Backend reiht ein, das Relay leert.

DB-Tabelle: `druckauftraege` · Spalten u. a.: `ziel_ip`, `payload` (Base64-ESC/POS), `bon_art` (`'arbeitsbon'` | `'kassenbeleg'`), `referenz`, `status` (`offen` → `gedruckt`; nach 3 Fehlversuchen `fehlgeschlagen` → `verworfen` oder zurück auf `offen`)

#### Relay

Separater Dienst (`windows/relay/`, Repo-Root): reiner Transport ohne Fachlogik und ohne lokalen Zustand. Holt offene Druckaufträge (`POST /relay/poll`), druckt und meldet das Ergebnis zurück (`POST /relay/ergebnis`, gedruckte IDs und Fehlversuche). Die ESC/POS-Formatierung liegt im Backend.

---

### Gastronomie & Betrieb

- **Inhaus / Außerhaus:** Historische Unterscheidung des Verzehrorts, seit 1.1.2026 für die Steuersatzbestimmung irrelevant: Speisen einheitlich 7 %, Getränke 19 % (→ [steuerrecht.md](steuerrecht.md)).
- **Trinkgeld:** Trinkgeld an den Verein ist voll steuerpflichtig; direkt an die Servicekraft in der Regel steuerfrei (→ [compliance.md](compliance.md)).
- **BYOD (Bring Your Own Device):** Servicekräfte nutzen eigene Smartphones; kein App-Install nötig.
- **Belegausgabepflicht (Bonpflicht):** Gesetzliche Pflicht nach § 146a Abs. 2 AO, Beleg nach jedem Kassiervorgang. Siehe → Kassenbeleg.
- **eBeleg:** Digitaler Kassenbon als papierloser Beleg-Ersatz. Geplant (→ anforderungen.md F-09).
- **Kassensturzfähigkeit:** Der Soll-Bestand muss jederzeit mit dem Ist-Bestand abgleichbar sein; Voraussetzung für GoBD-Konformität.

---

### Fiskalkonformität (Compliance Sub-Domain)

Begriffe der gesetzlich vorgeschriebenen Fiskalisierung nach § 146a AO und KassenSichV. TSE-Integration, Steuersätze und Kassenbeleg sind umgesetzt; DSFinV-K-Export, ELSTER-Meldung und eBeleg sind offen (→ `docs/anforderungen.md`).

> **Sprachkonvention:** Fiskal-Fachbegriffe folgen der deutschen Gesetzessprache und DSFinV-K-Spezifikation.

#### Gesetzliche Grundlagen

Je ein Satz, Pflichten und Details: [compliance.md §2](compliance.md#2-rechtliche-grundlagen).

- **AO (Abgabenordnung):** Zentrales Steuergesetz; § 146a regelt TSE-, Belegausgabe- und Meldepflichten für jeden jotti-Betreiber.
- **KassenSichV:** Verordnung mit den technischen Detailanforderungen an manipulationssichere Kassen, TSE und Belege.
- **GoBD:** BMF-Schreiben, steuerrelevante Daten müssen 10 Jahre lückenlos und unveränderbar gespeichert werden; jottis Append-only-Journal erfüllt das strukturell.
- **BSI:** Bundesbehörde, die technische Richtlinien (TR-03153) und Schutzprofile für die TSE-Zertifizierung definiert.

#### TSE & Kryptografie

- **TSE (Technische Sicherheitseinrichtung):** Zwingend vorgeschriebenes, BSI-zertifiziertes Sicherheitsmodul, das jeden Kassiervorgang kryptografisch signiert. In jotti als Cloud-TSE über das `TSEClient`-Interface umgesetzt.
- **TSEClient:** Anbieter-agnostisches Go-Interface (`domain/tse/client.go`) mit `StartTransaction` und `FinishTransaction` (atomares Muster). Implementierung: `FiskalyTSEClient` (`repository/tse_repo/`).
- **TSESetupClient:** Zweiter fiskaly-Sprecher neben dem Signierpfad: Go-Interface `SetupClient` (`domain/tse/setup.go`) für geführte Einrichtung und Statusabfrage (TSS anlegen oder übernehmen, Admin-PIN, Client registrieren, Stammdaten, Verbindungstest), zwangsläufig auch ohne fertige Konfiguration nutzbar. Implementierung: `FiskalyTSESetupClient` (`repository/tse_repo/fiskaly_setup.go`); Endpunkte u. a. `/admin/tse-einrichten`, `/admin/tse-uebernehmen`, `/admin/get-tse-status`.
- **Signaturauftrag:** Transaktionale Outbox-Zeile der Signierung (Tabelle `tse_signaturauftraege`, genau ein Auftrag je Event). Jeder signaturpflichtige Vorgang schreibt seinen Auftrag im selben Commit wie das Event; der Signatur-Worker quittiert die Signatur direkt am Auftrag. Die Signaturspalten (Transaktionsnummer, Signaturzähler, TSE-Seriennummer, logTime Start/Ende, Signatur, QR-Code-Daten) liegen am Auftrag, nicht mehr in der Event-Payload. Status: `offen`, `erledigt`, `fehlgeschlagen`, `tse_nicht_konfiguriert`.
- **Fiskalische Projektion:** Zentrale Funktion Event → (signaturpflichtig, processType, processData) (`domain/kasse/fiskalische_projektion.go`); sie entscheidet beim Einreihen, ob ein Vorgang einen Signaturauftrag erhält, und liefert dessen processData-Snapshot.
- **Signatur-Worker:** Einziger Sprecher für TSE-Signaturtransaktionen (`backend/app/tse_signatur_worker.go`). Arbeitet die offenen Aufträge FIFO ab, wird nach jedem Commit sofort angestoßen (Polling-Tick als Fallback), heilt per Ist-Abfrage und quittiert mit einem einzelnen Update am Auftrag. Ein session-gebundener Advisory Lock sichert die Single-Prozess-Annahme.
- **Signaturstatus:** Zustandslose Funktion (`domain/tse/signaturstatus.go`, `BestimmeSignaturstatus`) mit genau vier Ergebnissen: Signatur vorhanden, vorhanden mit Nachsigniert-Kennzeichen, Ausfall mit Grund, Signatur ausstehend. Einzige Implementierung des Ausfallbegriffs; Beleg-Abruf und Kassenabschluss-Gate urteilen über sie.
- **Signatur ausstehend:** Der Auftrag ist offen und keine Störung dokumentiert; die Signatur wird in Kürze erwartet. Der Beleg-Abruf antwortet mit Status `ausstehend` (kein Druckauftrag, die UI fasst nach), der Kassenabschluss blockiert.
- **Nachsigniert:** Kennzeichen einer verspäteten Signatur (Signatur später als rund eine Minute nach Auftragserstellung, Konstante `NachsigniertSchwelle`). Der Kassenbeleg druckt „Nachsigniert am …", weil die TSE-Zeitpunkte dann sichtbar vom Belegdatum abweichen. Ersetzt den früheren Begriff „Nachsignierung".
- **Störungsprotokoll:** Aufbewahrungspflichtige Tabelle `tse_stoerungen` der TSE-weiten Signierstörungen; je Störung ein Störungszeitraum, höchstens einer aktiv, kein DELETE.
- **Störungszeitraum:** Zeitraum im Störungsprotokoll (Beginn, Ende, Fehlertext; Ende NULL solange aktiv). Grund-Arten: `tse_fehler` (TSE-weiter Fehler), `rueckstand` (Signatur-Rückstand über der Schwelle), `keine_konfiguration` (keine TSE eingerichtet). Ein offener Auftrag während eines aktiven Zeitraums gilt als Ausfall.
- **Rückstands-Ausfall:** Offener Auftrag während eines aktiven `rueckstand`-Störungszeitraums. Ein Rückstands-Watchdog (`backend/app/tse_rueckstand_watchdog.go`) öffnet den Zeitraum, sobald der älteste offene Auftrag die Zwei-Minuten-Schwelle (`RueckstandSchwelle`) überschreitet, und schließt ihn beim Unterschreiten.
- **Aufholphase:** Zeitraum, in dem der Worker nach einer behobenen Störung den aufgelaufenen Rückstand abarbeitet; bis der Rückstands-Zeitraum schließt, tragen die dabei erzeugten Belege den Ausfallvermerk.
- **TSE nicht konfiguriert:** Endstatus `tse_nicht_konfiguriert` eines Auftrags, den der Worker ohne vorhandene TSE-Konfiguration endgültig markiert (keine Fehlversuche, keine automatische Wiederaufnahme). Der Beleg trägt den Vermerk „keine TSE konfiguriert"; der Kassenabschluss blockiert nie. Der Dauerzustand ohne Konfiguration ist ein `keine_konfiguration`-Störungszeitraum, der mit der Einrichtung endet.
- **Transaktionsnummer (TSE_TANR):** Eindeutige, fortlaufende TSE-Nummer pro Kassiervorgang. Dient der Lückenerkennung.
- **Signaturzähler (TSE_TA_SIGZ):** Stetig ansteigender Zähler pro Signaturvorgang. Pflichtfeld auf dem Kassenbeleg; als Signaturspalte am Signaturauftrag gespeichert.
- **Prüfwert / Signatur:** Kryptografischer Signaturwert, der den Vorgang absiegelt und auf dem Kassenbeleg abgedruckt wird.

#### Kasse & Identifikation

| Begriff                 | Bedeutung                                                                                                                                                                                 |
| ----------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| KassenID / Seriennummer | Eindeutige UUID-v4 der jotti-Instanz. Für ELSTER-Meldung und TSE-Protokoll. Persistiert in Tabelle `kassenidentitaet` (insert-once), abrufbar über Endpunkt `admin/get-kassenidentitaet`. |
| TSE-Konfiguration       | TSE-Zugangsdaten (API-Key/-Secret, TSS-ID, Client-ID), über die Admin-Einstellungen gepflegt. Singleton-Tabelle `tse_konfiguration`.                                                      |

#### Steuern

| Begriff                    | Bedeutung                                                                                                                                                                                          |
| -------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Steuersatz                 | Steuerklasse eines Produkts. Enum: `regel` (19 %), `ermaessigt` (7 %), `befreit` (0 %), `kombi` (70/30-Aufteilung). Go: `domain/steuer` · DB: `produkte.steuersatz` · JSON-Key: `steuersatz`       |
| Steuerbetrag / Nettobetrag | Pro Steuersatz berechnete Beträge (`steuer.Aufteilung`: Brutto, Netto, Steuer), immer in Cent. Auf dem Kassenbeleg als Steueraufteilung ausgewiesen. Fachregeln → [steuerrecht.md](steuerrecht.md) |

#### Export & Meldung

| Begriff                | Bedeutung                                                                                                                                                      |
| ---------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| DSFinV-K               | „Digitale Schnittstelle der Finanzverwaltung für Kassensysteme", standardisiertes CSV-ZIP-Exportformat (Version 2.5) für Betriebsprüfungen. Geplant (→ F-04).  |
| TAR-Archiv             | Gesetzlich vorgeschriebenes Dateiformat für den Export der rohen, kryptografisch gesicherten TSE-Log-Nachrichten.                                              |
| Kassenmeldung / ELSTER | Pflicht nach § 146a Abs. 4 AO: Meldung jeder jotti-Instanz innerhalb eines Monats nach Inbetriebnahme über das ELSTER-Portal (→ F-05).                         |
| ERiC                   | „ELSTER Rich Client", Programmierschnittstelle für die automatisierte ELSTER-Kommunikation. Geplant.                                                           |

---

### Architekturprinzipien

Kurzdefinitionen, die kanonische Architektur-Erklärung steht im [handbuch.md](handbuch.md).

- **Event-Sourcing:** Zustand wird aus unveränderlichen Events berechnet, nicht direkt gespeichert (→ handbuch.md §3.2).
- **Fat Event:** Event friert alle relevanten Daten zum Aktionszeitpunkt ein, Produktname, Preis, Steuersatz (→ handbuch.md §2.2).
- **Anti-Corruption Layer (ACL):** Eingefrorene Stammdaten entkoppeln den Kassenbetrieb von späteren Produkt-Änderungen (→ handbuch.md §2.2).
- **Append-only:** Events werden nie geändert oder gelöscht; Korrekturen sind kompensierende Events. Entspricht dem GoBD-Radierverbot (→ handbuch.md §3.2).
- **Synchrone Projektion:** Der Tisch-Zustand wird als `tisch_sessions`-Zeile in derselben Transaktion wie das Event geschrieben (→ handbuch.md §3.8).

---

## Geplant (nicht implementiert)

Die folgenden Begriffe sind definiert, aber noch nicht im Code implementiert. Details und Priorisierung in `docs/anforderungen.md`.

- **Küchendisplay (KDS):** Echtzeit-Anzeige offener Bestellungen an der Ausgabestation (K-13).
- **Zubereitungsstatus:** Status einer Position: offen → in Zubereitung → fertig (K-15).
- **Ausgabestation:** Physischer Ort (Küche, Getränketheke), an dem Positionen ausgegeben werden (K-13/K-15).
- **Stornoquote:** Verhältnis Stornierungsbetrag zu Bestellsumme.
- **Export:** CSV-Download von Umsätzen, Bestellungen und Artikeldaten für die Buchhaltung (R-02).
- **Privatentnahme / Privateinlage:** eigene DSFinV-K-Geschäftsvorfalltypen für Bewegungen in den/aus dem privaten Bereich des Vereins (neben dem → Geldtransit); aktuell wird jede Bargeld-Bewegung als Geldtransit gebucht.
