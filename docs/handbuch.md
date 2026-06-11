# Entwickler-Handbuch — jotti

## Inhaltsverzeichnis

1. [Überblick](#1-überblick)
   - [1.1 Systemvision](#11-systemvision)
   - [1.2 Designziele](#12-designziele)
   - [1.3 Bewusste Abgrenzung](#13-bewusste-abgrenzung)
2. [Bounded Contexts](#2-bounded-contexts)
   - [2.1 Kontextübersicht](#21-kontextübersicht)
   - [2.2 Beziehungen zwischen Kontexten](#22-beziehungen-zwischen-kontexten)
3. [Kasse (Core Domain)](#3-kasse-core-domain)
   - [3.1 Kassensitzung und Abrechnungskreis](#31-kassensitzung-und-abrechnungskreis)
   - [3.2 Kassenjournal (Event Store)](#32-kassenjournal-event-store)
   - [3.3 Subject-Design: Hierarchische Subjects](#33-subject-design-hierarchische-subjects)
   - [3.4 Tisch-Session (Abrechnungskreis-Aggregat)](#34-tisch-session-abrechnungskreis-aggregat)
   - [3.5 Kassensitzung-Lifecycle](#35-kassensitzung-lifecycle)
   - [3.6 Domain Events](#36-domain-events)
   - [3.7 Invarianten](#37-invarianten)
   - [3.8 Synchrone Projektion, CRUD-Entität und Event Replay](#38-synchrone-projektion-crud-entität-und-event-replay)
   - [3.9 Kassenbestand (Read Model)](#39-kassenbestand-read-model)
   - [3.10 Kassensturz](#310-kassensturz)
   - [3.11 Tagesabschluss (Z-Bon)](#311-tagesabschluss-z-bon)
   - [3.12 Policies](#312-policies)
   - [3.13 TSE-Architektur](#313-tse-architektur)
4. [Stammdaten](#4-stammdaten)
   - [4.1 Produkt-Aggregat](#41-produkt-aggregat)
   - [4.2 Tisch-Stammdaten](#42-tisch-stammdaten)
   - [4.3 Benutzer-Aggregat](#43-benutzer-aggregat)
   - [4.4 Tisch-Favoriten](#44-tisch-favoriten)
   - [4.5 Persistenz (CRUD)](#45-persistenz-crud)
   - [4.6 Bondruck: Arbeitsbon und Kassenbeleg (K-12)](#46-bondruck-arbeitsbon-und-kassenbeleg-k-12)
5. [Auth und Rollen](#5-auth-und-rollen)
   - [5.1 Rollen und Berechtigungsmatrix](#51-rollen-und-berechtigungsmatrix)
   - [5.2 Onboarding-Ablauf](#52-onboarding-ablauf)
6. [Architekturprinzipien](#6-architekturprinzipien)
   - [6.1 Schichtenarchitektur](#61-schichtenarchitektur)
   - [6.2 API-Design](#62-api-design)
   - [6.3 Frontend-Architektur](#63-frontend-architektur)
   - [6.4 Validierung](#64-validierung)
   - [6.5 Geldbeträge](#65-geldbeträge)
   - [6.6 Mehrbenutzerfähigkeit (OCC)](#66-mehrbenutzerfähigkeit-occ)
   - [6.7 Sicherheit](#67-sicherheit)
7. [Read Models](#7-read-models)
8. [Ubiquitous Language](#8-ubiquitous-language) → [language.md](language.md)

---

## 1. Überblick

### 1.1 Systemvision

jotti ist ein self-hosted mPOS-System (Go-Backend, React-Frontend, PostgreSQL, Docker Compose). Servicekräfte nutzen ihre eigenen Smartphones (BYOD) im Browser. Das Kassenjournal basiert auf Event-Sourcing; Stammdaten sind CRUD. Produktvision, Zielgruppe und Positionierung: siehe [produktbeschreibung.md](produktbeschreibung.md).

### 1.2 Designziele

| Ziel                        | Bedeutung                                                                                                                |
| --------------------------- | ------------------------------------------------------------------------------------------------------------------------ |
| **Radikale Einfachheit**    | Minimaler Funktionsumfang, der genau das abdeckt, was ein Vereinsfest braucht — nicht mehr.                              |
| **Mobile-first**            | Alle Interaktionen sind für Smartphone-Browser und Touch-Bedienung optimiert.                                            |
| **Lückenlose Transparenz**  | Jede Transaktion ist unveränderlich protokolliert. Kein Datenverlust, keine Manipulation.                                |
| **Null Kosten**             | Keine Hardware, keine Abo-Gebühren, keine externe Abhängigkeit.                                                          |
| **Volle Datenhoheit**       | Self-hosted, alle Daten auf dem eigenen Server.                                                                          |
| **Niedrige Einstiegshürde** | Keine Schulung, keine App-Installation. Browser öffnen, einloggen, loslegen.                                             |
| **Nachvollziehbarkeit**     | Event-Sourcing im Kassenjournal: Jede Bestellung, Zahlung, Stornierung und Kassenbewegung ist jederzeit nachvollziehbar. |

### 1.3 Bewusste Abgrenzung

Kartenzahlung, Reservierungen, Warenwirtschaft, Lieferservice, Multi-Standort, CRM und Kiosk-Modus sind bewusst ausgeschlossen. Vollständige Liste mit Begründung: siehe [produktbeschreibung.md §7.2](produktbeschreibung.md#72-was-jotti-bewusst-nicht-ist).

> **TSE / KassenSichV:** jotti unterliegt der TSE-Pflicht nach § 146a AO. Die TSE-Integration wird phasenweise implementiert — siehe [anforderungen.md](anforderungen.md) und [compliance.md](compliance.md).

---

## 2. Bounded Contexts

### 2.1 Kontextübersicht

| Context        | Typ                   | Beschreibung                                                                                                                            | Persistenz                     |
| -------------- | --------------------- | --------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------ |
| **Kasse**      | Core Domain           | Alle finanziellen Geschäftsvorfälle: Bestellen, Ausgabe bestätigen, Bezahlen, Stornieren, Kassenbewegungen, Kassensturz, Tagesabschluss | Event-Sourcing (Kassenjournal) |
| **Stammdaten** | Supporting Sub-Domain | Verwaltung von Produkten, Tischen, Benutzern, Betreiber-Stammdaten (CRUD)                                                               | CRUD                           |
| **Auth**       | Generic Sub-Domain    | Login, Logout, Passwort-Management, Token-Verwaltung                                                                                    | Infrastruktur                  |

> **Bondruck** (K-12) ist kein eigenständiger Bounded Context, sondern eine **Policy** innerhalb des Kasse-Context (→ [3.12 Policies](#312-policies)). Abrechnung/Reporting sind Read Models innerhalb der Kasse — kein eigener Context.

### 2.2 Beziehungen zwischen Kontexten

| Upstream   | Downstream | Beziehungstyp           | Beschreibung                                                                     |
| ---------- | ---------- | ----------------------- | -------------------------------------------------------------------------------- |
| Stammdaten | Kasse      | Customer/Supplier + ACL | Kasse liest Produkte/Tische, friert Daten zum Bestellzeitpunkt in Fat Events ein |
| Auth       | Kasse      | Open Host Service       | Token mit Benutzer-ID und Rolle                                                  |
| Auth       | Stammdaten | Open Host Service       | Token mit Benutzer-ID und Rolle                                                  |

Der Kasse-Kontext schützt sich über eine **Anti-Corruption Layer (ACL)** vor Stammdaten-Änderungen: Bestellungs-Events enthalten alle relevanten Produktdaten zum Zeitpunkt der Bestellung (Fat Events). Spätere Preisänderungen haben keinen Einfluss auf historische Bestellungen. Reporting-Projektionen aggregieren direkt über das Kassenjournal — keine Cross-Context-Kommunikation nötig.

> **Stammdaten-Änderungen während offener Kassensitzung:** Fat Events frieren Produktdaten zum Bestellzeitpunkt ein — Änderungen wirken erst in künftigen Bestellungen (Änderungssperre für Steuersätze folgt mit Compliance-Phase 1).

---

## 3. Kasse (Core Domain)

Der Kasse-Kontext vereint alle finanziellen Geschäftsvorfälle mit Event-Sourcing über das **Kassenjournal**: tischbezogene Vorgänge (Bestellen, Ausgabe bestätigen, Bezahlen, Stornieren, Auszahlen) und kassenführungsbezogene Vorgänge (Kassensitzung eröffnen, Anfangsbestand, Kassenbewegungen, Kassensturz, Tagesabschluss).

### 3.1 Kassensitzung und Abrechnungskreis

| Begriff              | Scope                            | DSFinV-K-Feld                  | Beschreibung                                                                                                                            |
| -------------------- | -------------------------------- | ------------------------------ | --------------------------------------------------------------------------------------------------------------------------------------- |
| **Kassensitzung**    | Global, 1× pro Veranstaltungstag | `Z_NR` (Kassenabschlussnummer) | Der administrative Rahmen: Eröffnung durch Admin, Anfangsbestand, Kassenbewegungen, Kassensturz, Tagesabschluss (Z-Bon).                |
| **Abrechnungskreis** | Pro Tisch pro Kassensitzung      | `ABRECHNUNGSKREIS`             | Die buchhalterische Einheit: Alle Bestellungen, Zahlungen, Stornierungen und Auszahlungen an einem Tisch innerhalb einer Kassensitzung. |

Die Kassensitzung ist der Container, der Abrechnungskreis (= Tisch-Session) ist der Inhalt. Der `ABRECHNUNGSKREIS` ist pro Tisch pro Tag (DSFinV-K).

### 3.2 Kassenjournal (Event Store)

Das Kassenjournal ist die zentrale, append-only Tabelle für alle finanziellen Geschäftsvorfälle — chronologische, vollständige, unveränderbare Aufzeichnung im Sinne von § 146 AO.

```
Kassenjournal (Tabelle: kassenjournal)
├── id                (int — DB-generiert, eindeutige Event-ID)
├── user_id           (int — wer hat die Aktion ausgeführt)
├── user_name         (string — Fat Event: Name zum Zeitpunkt der Aktion)
├── type              (string — Event-Typ, z. B. "bestellung-aufgenommen:v1")
├── subject           (string — Stream-Schlüssel, z. B. "kassensitzung-1/tisch-42")
├── version           (int — aufsteigende Version pro Subject, für OCC)
├── timestamp         (datetime — Zeitpunkt der Erzeugung)
├── data              (JSONB — Event-spezifische Daten)
└── kassensitzung_nr  (int — Zuordnung zur Kassensitzung für Cross-Stream-Queries)
```

Schema, Immutabilitäts-Trigger und OCC-Mechanismus (`UNIQUE(subject, version)`) sind identisch zum bisherigen Event Store. Die `kassensitzung_nr`-Spalte ermöglicht robuste Cross-Stream-Aggregationen (Reporting, Kassenbestand) ohne fragile LIKE-Patterns auf Subjects.

### 3.3 Subject-Design: Hierarchische Subjects

Subjects folgen einer hierarchischen Konvention mit zwei Ebenen:

```
kassensitzung-{nr}                             → Globaler Betriebstag (Kassensitzung)
kassensitzung-{nr}/tisch-{tischId}              → Abrechnungskreis (Tisch-Session)
kassensitzung-{nr}/direktverkauf-{uuid}         → Direktverkauf (ein Stream pro Verkauf)
```

**Kassensitzung-Subject:** `kassensitzung-1` — Nummer aus `kassensitzungen.z_nr`.

**Tisch-Session-Subject:** `kassensitzung-1/tisch-42` — entsteht implizit mit der ersten Bestellung (kein „Tisch-Öffnen"-Event).

**Direktverkauf-Subject:** `kassensitzung-1/direktverkauf-<uuid>` — ein eigener Stream pro Barverkauf an der Theke. Der Stream-Typ `direktverkauf` schreibt ausschließlich ins Kassenjournal (keine Projektion). `direktverkauf-getaetigt:v1` ist `version = 1`; positionsgenaue Stornierungen sind Folge-Versionen im selben Stream (OCC über `UNIQUE(subject, version)`).

Die **Storno-Validierung** läuft ohne Projektion per On-Demand-Replay des einzelnen Verkauf-Streams (`ReadEventsBySubject` → `ComputeNichtStornierteVerkaufPositionen`): Es lassen sich nur Positionen stornieren, die noch nicht (vollständig) storniert wurden, höchstens in der ursprünglich verkauften Menge. `direktverkauf-storniert:v1` speichert die stornierten Positionen als **Fat-Positionen** (wie der Tisch-Storno `stornierung-erteilt:v1`); die API nimmt `PositionRef` (`positionId` + `menge`) entgegen und reichert im Command an. `gesamtStornierungCents` (vom Command aus den nicht-stornierten Positionen berechnet) trägt den Geldbetrag. Anders als beim Tisch ist die Bargeld-Rückgabe **Teil des Storno-Vorgangs selbst** und mindert den Soll-Kassenbestand direkt — es gibt **keine** separate `auszahlung-geleistet`-Buchung, weil ein Direktverkauf keinen aufzulösenden Saldo hat. Die kompakte Direktverkauf-Historie (eine Zeile pro Verkauf, inkl. noch offener Positionen) entsteht durch Cross-Stream-Replay aller `direktverkauf-*`-Events der offenen Kassensitzung (`kassensitzung_nr`).

Separate Tisch-Subjects sind notwendig, weil der OCC-Constraint `UNIQUE(subject, version)` bei einem einzigen Subject alle Schreibvorgänge serialisieren würde — bei 5–30 Servicekräften nicht praktikabel.

**Kanonische Query-Strategie:**

| Zugriffsmuster                                                  | Kanonische Strategie                  | Beispiel                                        |
| --------------------------------------------------------------- | ------------------------------------- | ----------------------------------------------- |
| **Single-Stream-Replay** (ein Tisch, eine KS)                   | Exakter `subject`-Match               | `WHERE subject = 'kassensitzung-1/tisch-42'`    |
| **Cross-Stream-Aggregation** (Reporting, Kassenbestand, Export) | `kassensitzung_nr`                    | `WHERE kassensitzung_nr = $1`                   |
| **Tischübersicht** (alle Tische einer KS)                       | `kassensitzung_nr` + `tisch_sessions` | JOIN auf Projektion                             |
| **Globale Queries** (alle KS eines Tisches, Debug)              | Subject-LIKE                          | `WHERE subject LIKE 'kassensitzung-%/tisch-42'` |

### 3.4 Tisch-Session (Abrechnungskreis-Aggregat)

Die Tisch-Session ist die transaktionale Grenze für tischbezogene Vorgänge. Jeder Tisch innerhalb einer Kassensitzung bildet einen eigenständigen Abrechnungskreis mit eigenem Event-Stream (→ [3.6](#36-domain-events)), eigener Versionierung und eigenem Saldo. Das Projektions-Modell (`TischSession`) materialisiert den aktuellen Zustand (→ [3.8](#38-synchrone-projektion-crud-entität-und-event-replay)). Die Projektion ist session-scoped — jede KS startet mit leerer Projektion. Produktdaten werden als Fat Events zum Bestellzeitpunkt eingefroren.

### 3.5 Kassensitzung-Lifecycle

Die Kassensitzung durchläuft: **Eröffnung** (Datum + Bezeichnung) → **Anfangsbestand** (Wechselgeld) → **Betrieb** (Bestellungen, Ausgaben, Zahlungen, Stornierungen, Kassenbewegungen) → **Kassensturz** (Soll-Ist-Abgleich) → **Tagesabschluss** (Z-Bon, KS schließen). Alle KS-Events werden im selben Kassenjournal wie Tisch-Events gespeichert — Subject `kassensitzung-{nr}`.

### 3.6 Domain Events

Alle Events sind unveränderlich (append-only) und werden im Kassenjournal persistiert. Namenskonvention: deutsch, Partizip-Form, Pattern `{Substantiv}-{Partizip}:v{N}`.

#### Tisch-Session-Events (Subject: `kassensitzung-{nr}/tisch-{id}`)

##### BestellungAufgenommen (`bestellung-aufgenommen:v1`)

Servicekraft nimmt eine Bestellung am Tisch auf.

| Feld                 | Typ     | Beschreibung                   |
| -------------------- | ------- | ------------------------------ |
| `bestellung_id`      | UUID    | Eindeutige ID der Bestellung   |
| `gesamt_preis_cents` | int     | Summe aller Positionen in Cent |
| `kommentar`          | string? | Optional, max. 100 Zeichen     |
| `positionen[]`       | Array   | Mindestens 1 Position          |

**Position:**

| Feld            | Typ    | Beschreibung                                   |
| --------------- | ------ | ---------------------------------------------- |
| `position_id`   | UUID   | Eindeutige ID der Position                     |
| `variante_id`   | int    | FK auf Variante                                |
| `produkt_name`  | string | Fat Event — eingefroren                        |
| `variante_name` | string | Fat Event — eingefroren                        |
| `kategorie`     | enum   | `essen` · `getraenk` · `sonstiges` — Fat Event |
| `einzelpreis`   | int    | Cent, Fat Event — eingefroren                  |
| `menge`         | int    | ≥ 1                                            |

##### AusgabeBestaetigt (`ausgabe-bestaetigt:v1`)

Positionen als ausgegeben markieren. Teilausgaben möglich.

| Feld           | Typ           | Beschreibung                              |
| -------------- | ------------- | ----------------------------------------- |
| `ausgabe_id`   | UUID          | Eindeutige ID                             |
| `positionen[]` | PositionRef[] | `position_id` (UUID) + `menge` (int, ≥ 1) |
| `kommentar`    | string?       | Optional, max. 100 Zeichen                |

##### ZahlungKassiert (`zahlung-kassiert:v1`)

Barzahlung kassieren. Betrag = Summe der gewählten Positionen. Teilzahlungen möglich.

| Feld                   | Typ           | Beschreibung                              |
| ---------------------- | ------------- | ----------------------------------------- |
| `zahlung_id`           | UUID          | Eindeutige ID                             |
| `positionen[]`         | PositionRef[] | `position_id` (UUID) + `menge` (int, ≥ 1) |
| `gesamt_zahlung_cents` | int           | Cent — Summe der gewählten Positionen     |
| `kommentar`            | string?       | Optional, max. 100 Zeichen                |

##### StornierungErteilt (`stornierung-erteilt:v1`)

Stornierung durch Serviceleitung/Admin. Unabhängig vom Ausgabe-/Bezahlstatus.

| Feld                       | Typ           | Beschreibung                              |
| -------------------------- | ------------- | ----------------------------------------- |
| `stornierung_id`           | UUID          | Eindeutige ID                             |
| `positionen[]`             | PositionRef[] | `position_id` (UUID) + `menge` (int, ≥ 1) |
| `gesamt_stornierung_cents` | int           | Cent — Summe der stornierten Positionen   |
| `kommentar`                | string        | **Pflicht**, min. 3, max. 100 Zeichen     |

##### AuszahlungGeleistet (`auszahlung-geleistet:v1`)

Auszahlung durch Serviceleitung/Admin zum Ausgleich eines negativen Saldos (K-05). Freier Betrag, kein Positionsbezug.

| Feld            | Typ    | Beschreibung                          |
| --------------- | ------ | ------------------------------------- |
| `auszahlung_id` | UUID   | Eindeutige ID                         |
| `betrag_cents`  | int    | ≥ 1 Cent, kein Positionsbezug         |
| `kommentar`     | string | **Pflicht**, min. 3, max. 100 Zeichen |

#### Kassensitzung-Events (Subject: `kassensitzung-{nr}`)

##### KassensitzungEroeffnet (`kassensitzung-eroeffnet:v1`)

| Feld            | Typ    | Beschreibung                  |
| --------------- | ------ | ----------------------------- |
| `datum`         | date   | YYYYMMDD                      |
| `bezeichnung`   | string | z. B. „Sommerfest 2026 Tag 1" |
| `eroeffnet_von` | int    | User-ID des Admins            |

##### AnfangsbestandGesetzt (`anfangsbestand-gesetzt:v1`)

| Feld           | Typ | Beschreibung        |
| -------------- | --- | ------------------- |
| `betrag_cents` | int | Wechselgeld in Cent |
| `gesetzt_von`  | int | User-ID             |

##### GeldtransitGebucht (`geldtransit-gebucht:v1`)

Geldtransit (Einlage oder Entnahme) als Event im Kassenjournal.

| Feld           | Typ    | Beschreibung                          |
| -------------- | ------ | ------------------------------------- |
| `bewegung_id`  | UUID   | Eindeutige ID                         |
| `richtung`     | enum   | `einlage` · `entnahme`                |
| `betrag_cents` | int    | ≥ 1 Cent                              |
| `kommentar`    | string | **Pflicht**, min. 3, max. 200 Zeichen |
| `gebucht_von`  | int    | User-ID                               |

##### KassensturzDurchgefuehrt (`kassensturz-durchgefuehrt:v1`)

| Feld                 | Typ | Beschreibung     |
| -------------------- | --- | ---------------- |
| `soll_bestand_cents` | int | Errechneter Soll |
| `ist_bestand_cents`  | int | Gezählter Ist    |
| `differenz_cents`    | int | Soll − Ist       |
| `durchgefuehrt_von`  | int | User-ID          |

##### DifferenzSollIstGebucht (`differenz-soll-ist-gebucht:v1`)

Nur wenn `differenz_cents ≠ 0`.

| Feld           | Typ | Beschreibung                               |
| -------------- | --- | ------------------------------------------ |
| `betrag_cents` | int | Positiv = Überschuss, negativ = Fehlbetrag |
| `gebucht_von`  | int | User-ID                                    |

##### TagesabschlussErstellt (`tagesabschluss-erstellt:v1`)

| Feld                  | Typ      | Beschreibung                   |
| --------------------- | -------- | ------------------------------ |
| `z_nr`                | int      | Fortlaufend, nie zurücksetzbar |
| `zeitraum_von`        | datetime | Beginn der Kassensitzung       |
| `zeitraum_bis`        | datetime | Ende der Kassensitzung         |
| `umsatz_gesamt_cents` | int      | Gesamtumsatz                   |
| `stornierungen_cents` | int      | Summe Stornierungen            |
| `auszahlungen_cents`  | int      | Summe Auszahlungen             |
| `geldtransit_cents`   | int      | Netto-Geldtransit              |
| `erstellt_von`        | int      | User-ID                        |

### 3.7 Invarianten

#### Tisch-Session-Invarianten

$$\text{Saldo} = \sum \text{Bestellungen} - \sum \text{Zahlungen} - \sum \text{Stornierungen} + \sum \text{Auszahlungen}$$

Alle Beträge in Cent (Integer). Saldo = 0 bedeutet: alle Positionen bezahlt oder storniert. Ein Saldo < 0 entsteht, wenn bereits kassierte Positionen nachträglich storniert werden; `AuszahlungGeleistet` gleicht diesen negativen Saldo wieder aus.

- **Kassensitzung-Invariante:** Jeder schreibende Tisch-Vorgang erfordert eine offene Kassensitzung. Prüfung via `kassensitzungen`-Entität im Application Service. Keine offene KS → HTTP 409.
- **Ausgabe-Invariante:** Nur bestellte, nicht-stornierte Positionen können ausgegeben werden. Bereits ausgegebene Positionen nicht erneut ausgebbar. Teilausgaben zulässig.
- **Bezahl-Invariante:** Nur bestellte, nicht-stornierte, nicht-bezahlte Positionen können bezahlt werden. Der Zahlungsbetrag ergibt sich aus der Summe der gewählten Positionen — Überzahlung nicht möglich. Teilzahlungen zulässig.
- **Stornierungsinvariante:** Nur bestellte, nicht-stornierte Positionen können storniert werden — **unabhängig vom Ausgabe- und Bezahlstatus**. Bei Stornierung bereits bezahlter Positionen kann der Saldo temporär negativ werden (bewusstes Design). Kommentar ist **Pflichtfeld** (min. 3 Zeichen).
- **Auszahlungs-Invariante:** Betrag muss ≥ 1 Cent sein. Kommentar ist **Pflichtfeld** (min. 3, max. 100 Zeichen). Es gibt keine Obergrenze für den Auszahlungsbetrag (Freifeld). Nur `serviceleitung` und `admin` dürfen Auszahlungen leisten.
- **Rolleninvariante:** Stornierungen und Auszahlungen nur durch `serviceleitung` und `admin`. Alle anderen Tischoperationen (Bestellen, Ausgabe bestätigen, Bezahlen) stehen allen drei Rollen zur Verfügung.
- **Mindestmengen-Invariante:** Jede positionsbasierte Operation erfordert mindestens eine Position. Bestellung, Ausgabe, Zahlung oder Stornierung ohne Positionen sind ungültig.

#### Kassensitzung-Invarianten

| Invariante                    | Regel                                                                                                     |
| ----------------------------- | --------------------------------------------------------------------------------------------------------- |
| **Einzigkeits-Invariante**    | Maximal eine Kassensitzung darf `offen` sein.                                                             |
| **Nummern-Invariante**        | `z_nr` ist fortlaufend und lückenlos (`max(z_nr) + 1`). Wird beim INSERT in `kassensitzungen` berechnet.  |
| **Anfangsbestand-Invariante** | Pro Kassensitzung genau ein `AnfangsbestandGesetzt`. Wiederholter Aufruf wird abgelehnt.                  |
| **Kassensturz-Reihenfolge**   | `KassensturzDurchgefuehrt` ist Voraussetzung für `TagesabschlussErstellt`.                                |
| **Tisch-Saldo-Sperre**        | `TagesabschlussErstellt` ist nur möglich, wenn **alle** Tisch-Sessions der Kassensitzung Saldo = 0 haben. |
| **Abschluss-Invariante**      | `TagesabschlussErstellt` schließt die KS → Status `abgeschlossen`. Danach keine Events mehr im Stream.    |

> **Keine Bewegungs-Invariante:** Kassenbewegungen werden ohne Prüfung des Soll-Bestands gebucht.

### 3.8 Synchrone Projektion, CRUD-Entität und Event Replay

Eine synchrone Projektion (`tisch_sessions`) + eine CRUD-Entität (`kassensitzungen`) werden in derselben Transaktion wie das Event-INSERT aktualisiert (Write-Through). Ein expliziter `StreamType`-Parameter steuert das Routing — kein Subject-String-Parsing im Repository-Layer.

**Routing via StreamType:**

| `streamType`      | Kassenjournal-INSERT | `kassensitzungen` | `tisch_sessions` |
| ----------------- | -------------------- | ----------------- | ---------------- |
| `"kassensitzung"` | ✅                   | ✅ INSERT/UPDATE  | —                |
| `"tisch-session"` | ✅                   | —                 | ✅ UPSERT        |

**Write-Through-Ablauf:** `WriteEvent()` führt in einer einzigen Transaktion das Kassenjournal-INSERT und — je nach `streamType` — das `kassensitzungen`-UPDATE oder das `tisch_sessions`-UPSERT durch. Die `ApplyEvent()`-Funktion (`backend/domain/kasse/tisch_session.go`) ist eine reine Funktion in der Domain-Schicht (kein DB-Zugriff): nimmt `TischSession` + `Event` entgegen, gibt den neuen Zustand zurück.

#### `kassensitzungen` — CRUD-Entität (Hot-Path)

| Spalte   | Typ      | Beschreibung                       |
| -------- | -------- | ---------------------------------- |
| `z_nr`   | INT (PK) | Fortlaufende Kassenabschlussnummer |
| `datum`  | DATE     | Datum der Kassensitzung            |
| `status` | TEXT     | `offen` oder `abgeschlossen`       |

Diese CRUD-Entität wird bei **jedem** Tisch-Schreibvorgang gelesen (Kassensitzung-Sperre). Alle weiteren Kassensitzung-Daten (Anfangsbestand, Bezeichnung, Kassenbewegungen) werden bei Bedarf per In-Memory-Replay der wenigen KS-Events berechnet.

#### `tisch_sessions` — Session-scoped Tisch-Projektion

| Spalte                   | Typ       | Beschreibung                                     |
| ------------------------ | --------- | ------------------------------------------------ |
| `subject`                | TEXT (PK) | `kassensitzung-{nr}/tisch-{id}`                  |
| `tisch_id`               | INT (FK)  | Referenz auf `tische.id`                         |
| `kassensitzung_nr`       | INT       | Denormalisiert für schnelle Queries              |
| `saldo_cents`            | INT       | Aktueller Tisch-Saldo in Cent                    |
| `unbezahlte_positionen`  | JSONB     | `[]Position` — noch nicht bezahlte Positionen    |
| `ausstehende_positionen` | JSONB     | `[]Position` — noch nicht ausgegebene Positionen |
| `gesamt_zahlungen_cents` | INT       | Summe aller Zahlungen in Cent                    |
| `last_event_id`          | INT (FK)  | ID des zuletzt verarbeiteten Events              |
| `last_event_version`     | INT       | Version des zuletzt verarbeiteten Events         |

**Apply-Tabelle:**

| Event-Typ             | Zustandsänderung                                                                                             |
| --------------------- | ------------------------------------------------------------------------------------------------------------ |
| BestellungAufgenommen | Positionen zu `unbezahlte_positionen` und `ausstehende_positionen` hinzufügen, Saldo erhöhen                 |
| AusgabeBestaetigt     | Referenzierte Mengen aus `ausstehende_positionen` subtrahieren (Eintrag entfernen bei Menge 0)               |
| ZahlungKassiert       | Referenzierte Mengen aus `unbezahlte_positionen` subtrahieren, Saldo und `gesamt_zahlungen_cents` anpassen   |
| StornierungErteilt    | Referenzierte Mengen aus `unbezahlte_positionen` und `ausstehende_positionen` subtrahieren, Saldo reduzieren |
| AuszahlungGeleistet   | Saldo um `betrag_cents` erhöhen (negativen Saldo ausgleichen) — keine Positionslisten-Änderung               |

**Lesezugriff:** Operative Queries lesen direkt aus `tisch_sessions`; Historie liest den Event Stream via `ReadEventsBySubject()`. Bei Inkonsistenz kann `tisch_sessions` jederzeit aus dem Kassenjournal reberechnet werden (Single Source of Truth).

### 3.9 Kassenbestand (Read Model)

SQL-Aggregation über das Kassenjournal (eine `SELECT`-Query über `kassensitzung_nr`):

$$\text{Soll} = \text{Anfangsbestand}_{\text{KS}} + \sum_{\text{Tische}} \text{Zahlungen} - \sum_{\text{Tische}} \text{Auszahlungen} + \sum \text{Direktverkauf} - \sum \text{Direktverkauf-Storno} + \text{Kassenbewegungen}_{\text{netto}} + \text{DifferenzSollIst}$$

Alle Summanden stammen aus dem Kassenjournal. Keine Cross-Context-Projektion. Direktverkauf-Events (`direktverkauf-getaetigt:v1`, `direktverkauf-storniert:v1`) haben keine eigene Projektion, sind aber vollständig kassenwirksam und fließen in den Soll-Bestand ein.

### 3.10 Kassensturz

Am Ende einer Schicht vergleicht der Admin den errechneten Soll-Bestand mit dem physisch gezählten Ist-Bestand. Der Application Service schreibt beim Kassensturz **zwei Events in derselben Transaktion**, wenn `differenz_cents ≠ 0`:

| Version | Event                           | Wann                   |
| ------- | ------------------------------- | ---------------------- |
| N       | `kassensturz-durchgefuehrt:v1`  | Immer                  |
| N+1     | `differenz-soll-ist-gebucht:v1` | Nur wenn Differenz ≠ 0 |

Das `DifferenzSollIstGebucht`-Event bekommt eine eigene `kassenjournal.id` — direkt exportierbar als Zeile in `businesscases.csv` mit `GV_TYP = DifferenzSollIst`.

**Kassensturz — rechtliche Grundlagen:**

- **Soll-Bestand:** Anfangsbestand + Bareinnahmen − Barausgaben (vom System errechnet).
- **Ist-Bestand:** Manuell gezählter Bargeldbestand, erfasst im Zählprotokoll.
- **Invariante:** Kassenbestand darf niemals negativ sein.
- **Differenzbuchung:** Abweichung darf nicht per `UPDATE` überschrieben werden — Eigenbeleg erzeugen, TSE signieren, als `GV_TYP = DifferenzSollIst` in `businesscases.csv` exportieren. Fehlbetrag = negativer Wert, Überschuss = positiver Wert. Umsatzsteuerlich neutral.
- **Aufbewahrungspflicht:** Z-Bons und Zählprotokolle 10 Jahre (GoBD).
- **Hinweis Betriebsprüfung:** Eine Kasse, die „immer auf den Cent genau stimmt", gilt als manipulationsverdächtig — die Finanzverwaltung sucht mit IDEA-Software gezielt nach fehlenden Differenzen.

### 3.11 Tagesabschluss (Z-Bon)

Der Z-Bon ist das Ergebnis des `TagesabschlussErstellt`-Events — er aggregiert alle Geschäftsvorfälle einer Kassensitzung und erhält eine fortlaufende, nie zurücksetzbare `z_nr`.

**Invarianten:** `z_nr` strikt aufsteigend und lückenlos. Voraussetzung: Kassensturz durchgeführt + alle Tisch-Sessions Saldo = 0. Das Event schließt die KS (→ Status `abgeschlossen`). Z-Bons müssen 10 Jahre aufbewahrt werden (GoBD).

**Z-Bon vs. X-Bon:** Der Z-Bon setzt den Kassenumsatzspeicher auf null zurück; ein Zwischenbericht (X-Bon) ersetzt ihn rechtlich nicht. **Stammdaten-Snapshot:** Zu jedem Abschluss müssen die aktuell gültigen Stammdaten (Steuersätze, TSE-Zertifikate, Kassen-IDs) eingefroren werden — vor jeder Stammdaten-Änderung zunächst Kassenabschluss durchführen.

**Betreiber-Ablauf:**

1. Bargeld zählen (Zählprotokoll).
2. Z-Bon im System erstellen (kein X-Bon).
3. Ist mit Soll abgleichen.
4. Differenzen buchen (→ [§3.10](#310-kassensturz)).
5. Z-Bon und Zählprotokoll archivieren (10 Jahre).

### 3.12 Policies

- **Stornierungsberechtigung (K-04):** Nur `serviceleitung` und `admin` dürfen `StornierungErteilen`. Die Berechtigung wird in der Anwendungsschicht geprüft, bevor der Command an das Aggregat geht.
- **Arbeitsbon-Druck nach Kategorie (K-12):** Jedes `bestellung-aufgenommen:v1`-Event löst im Backend die **Arbeitsbon-Policy** aus (`backend/api/bondruck`). Sie gruppiert die Positionen nach Kategorie, schlägt für jede Kategorie mit konfigurierter Druckstation IP und Bonmodus (`pro_position` oder `pro_bestellung`) aus der `druckstationen`-Tabelle nach, formatiert den ESC/POS-Payload und reiht je einen **Druckauftrag** in die Outbox (`druckauftraege`) ein. Kategorien ohne Druckstation erzeugen keinen Auftrag. Der Arbeitsbon trägt **keine Preise** und ist **kein** Beleg i. S. v. § 146a AO (→ [4.6 Bondruck](#46-bondruck-arbeitsbon-und-kassenbeleg-k-12)). Das Print-Relay ist reiner Transport: Es holt offene Aufträge via `POST /relay/poll` und bestätigt gedruckte via `POST /relay/quittieren`.
- **Umbuchung (K-09):** Verschiebt eine Bestellung von Quell- auf Ziel-Tisch (= Stornierung + neue Bestellung). Cross-Aggregat-Transaktion — Atomarität auf Anwendungsebene sicherstellen. Nur `serviceleitung` und `admin`.

### 3.13 TSE-Architektur

> Compliance-spezifische Architektur-Entscheidungen für die TSE-Integration. Allgemeine Architekturprinzipien (Event-Sourcing, Kassenjournal, Projektion, OCC, Fat Events) sind in den vorherigen §3-Abschnitten beschrieben. Rechtliche Grundlagen → [compliance.md §3–§8](compliance.md).

#### TSE-Integration

```
┌─────────────────────────────────────────────┐
│                  jotti Backend               │
│                                              │
│  ┌──────────┐    ┌──────────────────────┐   │
│  │ Service-  │───▶│  TSE-Service          │   │
│  │ Handler   │    │  (Application Layer)  │   │
│  └──────────┘    └──────────┬───────────┘   │
│                              │               │
│                    ┌─────────▼─────────┐     │
│                    │  TSEClient        │     │
│                    │  (Interface)      │     │
│                    └─────────┬─────────┘     │
│                              │               │
└──────────────────────────────┼───────────────┘
                               │
                    ┌──────────▼──────────┐
                    │   Cloud-TSE API     │
                    │   (z.B. fiskaly)    │
                    └─────────────────────┘
```

**Interface-Design:**

```go
type TSEClient interface {
    StartTransaction(ctx context.Context, kassenID string, processType string, processData string) (StartResult, error)
    // UpdateTransaction nur für processType="Bestellung-V1" und "SonstigerVorgang-V1" zulässig.
    // Für "Kassenbeleg-V1" verboten (BMF-FAQ).
    UpdateTransaction(ctx context.Context, kassenID string, transactionNumber int, processData string) error
    FinishTransaction(ctx context.Context, kassenID string, transactionNumber int, processType string, processData string) (FinishResult, error)
}

type StartResult struct {
    TransactionNumber int
    LogTime           time.Time // TSE-interner Zeitstempel
    SerialNumberTSE   string
    SignatureCounter  int
}

type FinishResult struct {
    Signature        string
    LogTime          time.Time
    SignatureCounter int
}
```

#### Mapping: jotti-Vorgänge → TSE-Transaktionen (Atomares Modell)

Für das Festzelt-Muster gilt: Jeder Vorgang ist eine **eigenständige, sofort geschlossene** TSE-Transaktion.

| jotti-Vorgang                            | TSE-Operation             | processType           | Anmerkung                                                         |
| ---------------------------------------- | ------------------------- | --------------------- | ----------------------------------------------------------------- |
| Bestellung aufnehmen                     | `Start` + sofort `Finish` | `Bestellung-V1`       | Positionen in processData                                         |
| Zahlung kassieren (Teilzahlung)          | `Start` + sofort `Finish` | `Kassenbeleg-V1`      | Betrag + Zahlungsart in processData; **kein** UpdateTransaction   |
| Zahlung kassieren (Vollzahlung)          | `Start` + sofort `Finish` | `Kassenbeleg-V1`      | Wie oben                                                          |
| Positions-Storno                         | `Start` + sofort `Finish` | `Kassenbeleg-V1`      | Negative Menge/Betrag; BON_STORNO=1 im DSFinV-K                   |
| Bon-Storno (nach Zahlung)                | `Start` + sofort `Finish` | `Kassenbeleg-V1`      | Negativer Gesamtbetrag; BON_STORNO=1, REF_BON_ID gesetzt          |
| Auszahlung (negativen Saldo ausgleichen) | `Start` + sofort `Finish` | `Kassenbeleg-V1`      | Bargeldabfluss; negativer Betrag, Zahlungsart bar                 |
| Geldtransit (Einlage/Entnahme)           | `Start` + sofort `Finish` | `SonstigerVorgang-V1` | Kassenwirksame Geldbewegung ohne Umsatz                           |
| Kassendifferenz (Kassensturz)            | `Start` + sofort `Finish` | `SonstigerVorgang-V1` | Eigenbeleg `DifferenzSollIst`; umsatzsteuerlich neutral (→ §3.10) |
| Direktverkauf                            | `Start` + sofort `Finish` | `Kassenbeleg-V1`      | Bestellen + Zahlen in einem Schritt; 1 Verkauf = 1 Transaktion    |
| Direktverkauf-Storno                     | `Start` + sofort `Finish` | `Kassenbeleg-V1`      | Negativer Betrag; BON_STORNO=1, REF_BON_ID gesetzt                |
| Tagesabschluss (Z-Bon)                   | `Start` + sofort `Finish` | `SonstigerVorgang-V1` | Tagesaggregat in processData                                      |

**Alle Transaktionen eines Tisches** teilen denselben `ABRECHNUNGSKREIS`-Wert im DSFinV-K-Export.

#### Event-Store-Erweiterung

Die TSE-Rückgabewerte müssen in den bestehenden Event-Daten persistiert werden:

```go
// Erweiterung der Event-Data-Structs um TSE-Felder
type TSEData struct {
    TransactionNumber int    `json:"tseTransactionNumber"`
    LogTimeStart      string `json:"tseLogTimeStart"`  // logTime von StartTransaction
    LogTimeEnd        string `json:"tseLogTimeEnd"`    // logTime von FinishTransaction
    SignatureCounter  int    `json:"tseSignatureCounter"`
    Signature         string `json:"tseSignature"`
    SerialNumberTSE   string `json:"tseSerialNumber"`
    ProcessType       string `json:"tseProcessType"`
}

// Zusätzlich in TischSession-Daten:
type TischSession struct {
    AbrechnungskreisID     string    // z.B. "Tisch 42" (Tischname)
    ErsteBestellungLogTime time.Time // logTime der ersten Bestellung-V1 für den Bon-Aufdruck
}
```

#### DSFinV-K Export-Architektur

```
┌──────────────────────────────────────────────┐
│              DSFinV-K Exporter                │
│                                               │
│  ┌────────────┐  ┌────────────┐  ┌────────┐ │
│  │ Stammdaten- │  │ Einzelauf- │  │ Z-Bon- │ │
│  │ modul       │  │ zeichnungs-│  │ Modul  │ │
│  │             │  │ modul      │  │        │ │
│  └──────┬─────┘  └──────┬─────┘  └───┬────┘ │
│         │               │             │       │
│  ┌──────▼───────────────▼─────────────▼────┐ │
│  │   CSV-Generator (offizielle Dateinamen) │ │
│  │   transactions.csv, lines.csv,          │ │
│  │   cashregister.csv, tse.csv, …          │ │
│  └──────────────────┬──────────────────────┘ │
│                     │                         │
│  ┌──────────────────▼──────────────────────┐ │
│  │     index.xml Generator + ZIP-Builder  │ │
│  └─────────────────────────────────────────┘ │
└──────────────────────────────────────────────┘
```

#### Anbieter- und Meldeweg-Entscheidungen

TSE-Anbieter (fiskaly als erster Zielanbieter; anbieter-agnostisches `TSEClient`-Interface gegen Vendor-Lock-in) und Kassenmeldungs-Weg (Phase 1 manuell über ELSTER, Phase 2 ERiC oder fiskaly-Submission-API) sind mitsamt Begründung und Abwägung (ERiC vs. Submission-API) in [compliance.md §3.5 und §7](compliance.md) dokumentiert.

---

## 4. Stammdaten

Alle Stammdaten verwenden Soft-Delete via `status = 'deleted'`. Datensätze werden nie physisch gelöscht — referenzielle Integrität und historische Nachvollziehbarkeit (Fat Events im Kassenjournal) erfordern dies.

### 4.1 Produkt-Aggregat

Das Produkt-Aggregat verwaltet den Produktkatalog der Veranstaltung. Jedes Produkt gehört zu einer Kategorie und kann beliebig viele Varianten besitzen — jede Variante mit eigenem Namen und Preis.

```
Produkt
├── produkt_id       (int — DB-generiert)
├── name             (string — nicht leer)
├── kategorie        (essen | getraenk | sonstiges)
├── status           (active | deleted)
└── varianten[]
    ├── variante_id  (int — DB-generiert)
    ├── name         (string — nicht leer)
    ├── preis        (int, Cent — ≥ 0)
    └── status       (active | inactive | deleted)
```

**Invarianten:**

- Produktname darf nicht leer sein.
- Kategorie muss ein gültiger Wert sein (`essen`, `getraenk`, `sonstiges`).
- Jede Variante benötigt einen nicht-leeren Namen und einen Preis ≥ 0 (in Cent).
- Varianten können unabhängig vom Produkt deaktiviert werden (`inactive`). Inaktive Varianten erscheinen nicht im Service-Katalog.

### 4.2 Tisch-Stammdaten

Das Tisch-Stammdaten-Aggregat verwaltet die Basisdaten eines Tisches (Name + Status). Strikt von der Tisch-Session im Kasse-Kontext (→ [3.4](#34-tisch-session-abrechnungskreis-aggregat)) zu unterscheiden.

```
Tisch (Stammdaten)
├── tisch_id    (int — DB-generiert)
├── name        (string — nicht leer, z. B. „Tisch 1", „Stehtisch Eingang")
└── status      (active | inactive | deleted)
```

**Invarianten:**

- Name darf nicht leer sein.
- Nur aktive Tische (`active`) erscheinen in der Tischübersicht der Servicekräfte.

**Abgrenzung zum Kasse-Kontext:** In den Stammdaten ist der Tisch eine CRUD-Entität (Name + Status); im Kasse-Kontext eine Event-Sourced Tisch-Session (Abrechnungskreis). Beide teilen `tisch_id`.

### 4.3 Benutzer-Aggregat

Das Benutzer-Aggregat verwaltet die Zugangsdaten und Rollen der Helfer und Admins.

```
Benutzer
├── benutzer_id                (int — DB-generiert)
├── name                       (string — Anzeigename)
├── benutzername               (string — eindeutig, Login-Name)
├── passwort_hash              (string — Argon2id, NULL bei Neuanlage)
├── einmalpasswort_hash        (string — Argon2id, NULL nach Passwort-Vergabe)
├── rolle                      (admin | serviceleitung | service)
└── status                     (active | inactive | deleted)
```

**Invarianten:**

- Benutzername muss systemweit eindeutig sein.
- Rolle muss ein gültiger Wert sein (`admin`, `serviceleitung`, `service`).
- Passwort wird mit Argon2id gehasht — Klartext-Passwörter nie persistiert.
- Deaktivierte (`inactive`) und entfernte (`deleted`) Benutzer können sich nicht anmelden.
- Neue Benutzer: Status `inactive`, 6-stelliges Einmalpasswort (→ [5.2](#52-onboarding-ablauf)).

### 4.4 Tisch-Favoriten

Tisch-Favoriten sind eine einfache CRUD-Relation im Stammdaten-Kontext. Sie verknüpfen einen Benutzer mit einem oder mehreren Tischen und steuern, welche Tische auf dem Service-Dashboard als "Meine Tische" angezeigt werden.

```
tisch_favoriten
├── user_id     (int — FK users(id), NOT NULL)
├── tisch_id    (int — FK tische(id), NOT NULL)
└── created_at  (timestamptz — DEFAULT NOW())

PRIMARY KEY (user_id, tisch_id)
INDEX idx_tisch_favoriten_user_id ON tisch_favoriten(user_id)
```

Direktes CRUD ohne Event-Sourcing (kein Aggregat, keine Events). Benutzerspezifisch, idempotente Operationen (`ON CONFLICT DO NOTHING`), nur aktive Tische erlaubt. Fremdschlüssel sichern referenzielle Integrität; physisches Löschen durch Soft-Delete ausgeschlossen. Repository `favorit_repo` kapselt drei Operationen: `Add`, `Remove`, `GetByUser`.

### 4.5 Persistenz (CRUD)

Stammdaten (Produkte, Tische, Benutzer) werden mit klassischem CRUD verwaltet. Event-Sourcing ist hier nicht nötig — die historischen Daten stecken bereits in den Fat Events des Kasse-Context. Alle Stammdaten tragen `erstellt_am` und `aktualisiert_am` Zeitstempel.

### 4.6 Bondruck: Arbeitsbon und Kassenbeleg (K-12)

**Bondruck** umfasst zwei fachlich getrennte Bon-Familien auf einer gemeinsamen Druck-Infrastruktur. Sie teilen **keinen** Auslöser, Inhalt oder Rechtsstatus — nur die Druckauftrags-Outbox (`druckauftraege`) als Transport.

| Familie         | Auslöser                                        | Rechtsstatus           | Inhalt                                          |
| --------------- | ----------------------------------------------- | ---------------------- | ----------------------------------------------- |
| **Arbeitsbon**  | `bestellung-aufgenommen:v1` (automatisch)       | nicht-fiskalisch       | Ware ohne Preise (Küche/Theke)                  |
| **Kassenbeleg** | `POST /service/beleg-drucken` (auf Anforderung) | fiskalisch (§ 146a AO) | Positionen mit Preisen, Vereinsdaten, Kassen-ID |

**Arbeitsbon (operativ, K-12):** Eine Policy im Kasse-Context (→ [3.12 Policies](#312-policies)). Bei `bestellung-aufgenommen:v1` gruppiert die Arbeitsbon-Policy (`backend/api/bondruck`) die Positionen nach Kategorie, schlägt IP und Bonmodus aus `druckstationen` nach, formatiert den ESC/POS-Payload und reiht je einen Druckauftrag (`bon_art = 'arbeitsbon'`) in die Outbox ein. Kategorien ohne konfigurierte Druckstation erzeugen keinen Auftrag. **Inhalt:** Tischnummer (groß), Position (Art + Menge, z. B. „3x Pommes (groß)"), Kommentar, Uhrzeit, Servicekraft — **keine Preise**. Kein Beleg i. S. v. § 146a AO. KDS (K-13) und Zubereitungsstatus (K-15) sind noch offen.

**Kassenbeleg (fiskalisch, auf Anforderung):** Ein Service-Command (`POST /service/beleg-drucken`) erzeugt pro Anforderung genau **einen** Druckauftrag (`bon_art = 'kassenbeleg'`) an den Kassenbeleg-Drucker. Als Datenquelle dient entweder eine Tischzahlung (`zahlung-kassiert:v1`, Request: `tischId` + `zahlungId`) oder ein Direktverkauf (`direktverkauf-getaetigt:v1`, Request: `verkaufId`). **Inhalt:** Vereinsdaten (Betreiber, K-20), Kassen-Seriennummer (F-01), Datum/Uhrzeit, Positionen mit Einzelpreis × Menge, Gesamtbetrag, Zahlungsart „bar", Bon-Nummer (Event-ID des referenzierten Vorgangs). Die Outbox-Referenz ist entsprechend `zahlung-kassiert:{eventId}` oder `direktverkauf-getaetigt:{eventId}`. Erneuter Aufruf druckt nach, ohne den Vorgang fachlich zu wiederholen. Am Fest wird der Beleg selten verlangt (Belegausgabe-Befreiung → [compliance.md §5.1](compliance.md)). Steuer-Aufschlüsselung folgt mit F-07, TSE-Pflichtfelder mit F-02.

**Druckauftrags-Outbox (`druckauftraege`):** Single Source of Truth für alle Druckjobs — technische Warteschlange, **kein** fiskalisches Journal. Einziger Statusübergang: `offen → gedruckt`.

```
druckauftraege
├── id           (serial — PK)
├── ziel_ip      (string — IPv4 des Zieldruckers)
├── payload      (text — Base64-kodierter ESC/POS-Byte-String)
├── status       (offen | gedruckt)
├── bon_art      (arbeitsbon | kassenbeleg)
├── referenz     (text — fachliche Referenz, z. B. Event-ID)
├── erstellt_am  (timestamptz)
└── gedruckt_am  (timestamptz — NULL bis quittiert)
```

**Druckstationen (`druckstationen`-Tabelle):** Arbeitsbon-Stationen je Produktkategorie. Genau drei Zeilen (Seed: essen, getraenk, sonstiges). Admin-Konfiguration über `/admin/get-druckstationen` und `/admin/update-druckstationen`. Validierung: strikte IPv4-Validierung im Backend (zog `.IPv4()`), identisch im Frontend (Zod `z.ipv4()`).

```
druckstationen
├── kategorie   (essen | getraenk | sonstiges — PK)
├── drucker_ip  (string — IPv4, leer = kein Drucker)
├── bonmodus    (pro_position | pro_bestellung)
└── updated_at  (timestamptz)
```

**Bondruck-Einstellungen (`bondruck_einstellungen`, Singleton):** zentrale Bondruck-Konfiguration, administriert über `/admin/get-bondruck-einstellungen` und `/admin/update-bondruck-einstellungen`.

```
bondruck_einstellungen
├── kassenbeleg_drucker_ip  (string — IPv4, leer = kein Kassenbeleg-Drucker)
├── direktverkauf_modus     (kein_bon | abholbon | an_stationen)
├── abholbon_drucker_ip     (string — IPv4, leer = kein Abholbon-Drucker)
└── updated_at              (timestamptz)
```

`direktverkauf_modus` steuert den Bondruck für `direktverkauf-getaetigt:v1`:

- `kein_bon`: kein Druckauftrag.
- `abholbon`: genau ein kombinierter Abholbon (festes Label „Direktverkauf“, keine Preise) an `abholbon_drucker_ip`.
- `an_stationen`: Arbeitsbons nach Kategorie über `druckstationen` (identische Routing-Logik wie bei Tisch-Bestellungen).

Für den fiskalischen Kassenbeleg bleibt `kassenbeleg_drucker_ip` maßgeblich. Fehlt diese IP, schlägt `POST /service/beleg-drucken` mit klarer Fehlermeldung fehl.

**Relay = Transport:** Das Print-Relay (`cmd/relay/main.go`) holt offene Aufträge via `POST /relay/poll` (Antwort `{auftraege: [{id, zielIp, payload}]}`), druckt sie und bestätigt die gedruckten IDs via `POST /relay/quittieren` (`{token, gedruckteIds}`); das Backend setzt daraufhin `status = 'gedruckt'`. Das Relay formatiert nichts, kennt keine Kategorien und führt keinen Cursor — ESC/POS-Formatierung und Fachlogik liegen vollständig im Backend.

**Drucker-Voraussetzungen:** ESC/POS-Bondrucker mit Ethernet-Anschluss (TCP Port 9100), 80-mm-Thermodrucker (48 Zeichen/Zeile). Statische IP-Adresse empfohlen. Getestet mit MUNBYN ITPP047P-UE.

**Relay starten:**

```bash
RELAY_AUTH_TOKEN="<RELAY_AUTH_TOKEN>" \
RELAY_BACKEND_URL="https://jotti.meinverein.de/api" \
RELAY_POLL_SECONDS="2" \
./jotti-relay
```

| Umgebungsvariable       | Beschreibung                                                    | Standard                |
| ----------------------- | --------------------------------------------------------------- | ----------------------- |
| `RELAY_AUTH_TOKEN`      | Authentifizierungs-Token (aus `.env` des Servers)               | (erforderlich)          |
| `RELAY_BACKEND_URL`     | URL des jotti-Servers inkl. `/api`                              | `https://localhost/api` |
| `RELAY_POLL_SECONDS`    | Abfrageintervall in Sekunden                                    | `2`                     |
| `RELAY_TLS_SKIP_VERIFY` | TLS-Zertifikatspruefung ueberspringen (`1/true` oder `0/false`) | bei Local-Default aktiv |

**TLS-Verhalten:**

- **Local (ein Geraet, selbstsigniertes Zertifikat):** Ohne weitere Env-Variablen nutzt das Relay `https://localhost/api` und deaktiviert die Zertifikatspruefung automatisch. Optional explizit: `RELAY_TLS_SKIP_VERIFY=1`.
- **Cloud (gueltiges Zertifikat):** `RELAY_BACKEND_URL="https://<deine-domain>/api"` setzen und **kein** `RELAY_TLS_SKIP_VERIFY=1` setzen, damit Zertifikate verifiziert werden.

**Fehlerverhalten:** Bei nicht erreichbarem Drucker versucht das Relay mehrfach (bis zu 5 Minuten); danach bleibt der Auftrag `offen` und wird beim nächsten Poll erneut geliefert. Der DB-Status ist autoritativ: Stürzt das Relay zwischen Druck und Quittierung ab, kann ein Auftrag erneut gedruckt werden — beim nicht-fiskalischen Arbeitsbon unkritisch.

**Schnelltest:**

```bash
curl -X POST http://localhost:3000/relay/poll \
     -H "Content-Type: application/json" \
     -d '{"token":"<RELAY_AUTH_TOKEN>"}'
```

Erwartung: `200` mit `auftraege` bei gültigem Token, `401` bei ungültigem.

---

## 5. Auth und Rollen

### 5.1 Rollen und Berechtigungsmatrix

jotti kennt drei Rollen mit abgestuften Berechtigungen. Die Rollenprüfung erfolgt serverseitig anhand des JWT.

| Rolle              | Code-Bezeichnung | Beschreibung                                                         |
| ------------------ | ---------------- | -------------------------------------------------------------------- |
| **Admin**          | `admin`          | Voller Zugriff auf Stammdaten, Kasse (inkl. Kassensitzung) und Admin |
| **Serviceleitung** | `serviceleitung` | Kasse-Tischoperationen einschließlich Stornierung                    |
| **Servicekraft**   | `service`        | Kasse-Tischoperationen ohne Stornierung                              |

**Berechtigungsmatrix:**

| Aktion                         | Admin | Serviceleitung | Servicekraft |
| ------------------------------ | :---: | :------------: | :----------: |
| _Stammdaten_                   |       |                |              |
| Produkte verwalten             |   ✔   |                |              |
| Tische verwalten               |   ✔   |                |              |
| Benutzer verwalten             |   ✔   |                |              |
| Passwort zurücksetzen          |   ✔   |                |              |
| Betreiber-Stammdaten verwalten |   ✔   |                |              |
| _Kasse — Tisch-Operationen_    |       |                |              |
| Bestellung aufnehmen           |   ✔   |       ✔        |      ✔       |
| Ausgabe bestätigen             |   ✔   |       ✔        |      ✔       |
| Zahlung kassieren              |   ✔   |       ✔        |      ✔       |
| Stornierung erteilen           |   ✔   |       ✔        |              |
| Auszahlung leisten             |   ✔   |       ✔        |              |
| Tischübersicht einsehen        |   ✔   |       ✔        |      ✔       |
| Kassenjournal einsehen         |   ✔   |       ✔        |      ✔       |
| _Kasse — Kassensitzung_        |       |                |              |
| Kassensitzung eröffnen         |   ✔   |                |              |
| Anfangsbestand setzen          |   ✔   |                |              |
| Kassenbestand einsehen         |   ✔   |                |              |
| Kassenbewegung buchen          |   ✔   |                |              |
| Kassensturz durchführen        |   ✔   |                |              |
| Tagesabschluss (Z-Bon)         |   ✔   |                |              |
| _Abrechnung_                   |       |                |              |
| Tagesabrechnung einsehen       |   ✔   |                |              |
| Datenexport                    |   ✔   |                |              |
| _Allgemein_                    |       |                |              |
| Abmelden                       |   ✔   |       ✔        |      ✔       |

Die Rollenhierarchie ist inklusiv: Admin kann alles, was Serviceleitung kann. Serviceleitung kann alles, was Servicekraft kann — plus Stornierung.

### 5.2 Onboarding-Ablauf

Neue Benutzer durchlaufen einen zweistufigen Onboarding-Prozess, der sicherstellt, dass nur der Benutzer sein eigenes Passwort kennt:

1. **Benutzer anlegen:** Admin erstellt Benutzer (Name, Benutzername, Rolle, Status `inactive`). System generiert 6-stelliges Einmalpasswort, das der Admin dem Benutzer mitteilt.
2. **Erstanmeldung + Passwort setzen:** Benutzer meldet sich mit Einmalpasswort an. System erkennt am Zustand `einmalpasswort_hash ≠ NULL ∧ passwort_hash = NULL` den Onboarding-Status und leitet zur Passwort-Vergabe weiter (min. 8 Zeichen, Argon2id-Hash). Danach reguläre Anmeldung.

**Passwort-Reset:** Admin-Reset generiert neues Einmalpasswort, leert `passwort_hash` → Benutzer durchläuft Onboarding erneut.

---

## 6. Architekturprinzipien

### 6.1 Schichtenarchitektur

Das Backend ist in vier Schichten gegliedert: **HTTP** → **Application** → **Domain** → **Repository/Infra**.

- **HTTP-Schicht:** Request-Parsing, Response-Serialisierung, eigene DTOs mit `json`-Tags. Domain-Modelle nie direkt serialisiert — dedizierte Mapper. Keine Business-Logik.
- **Application-Schicht:** Use-Case-Koordination: fachliche Validierung (zog), Aggregat-State laden, Domain-Logik aufrufen, persistieren. Übersetzt Domain-Fehler in Fehlercodes.
- **Domain-Schicht:** Invarianten, Event-Konstruktion, Zustandsberechnung. Kein DB-, HTTP- oder JSON-Zugriff. Keine `json`-Tags (Ausnahme: Event-Data-Structs).
- **Repository/Infra-Schicht:** Datenbankzugriffe. Kasse: Kassenjournal (append-only) + synchrone Projektion + CRUD-Entität. Stammdaten: CRUD. Basis: sqlc-generierte Queries.

### 6.2 API-Design

**POST-only:** Alle API-Endpunkte sind POST-Endpunkte. Jede Aktion wird explizit benannt (z. B. `/service/bestellung-aufnehmen` statt `PUT /tables/5`).

**Bewusste Ops-Ausnahme:** `GET /health` ist explizit erlaubt, damit Container-Orchestrierung und Reverse-Proxy den Backend-Status per Healthcheck pruefen koennen. Diese Ausnahme gilt nur fuer `/health`; alle fachlichen Endpunkte bleiben strikt POST-only.

**JSON:** Request- und Response-Bodies sind JSON.

**Authentifizierung:** Jeder Endpunkt (außer `/auth/*`) erwartet ein gültiges JWT im `Authorization: Bearer <token>`-Header. Die Middleware prüft Signatur und Gültigkeit.

**Fehlerformat:** `{ "code": "<string>", "details": "<optional>" }`. HTTP-Statuscodes: `400` Client-Fehler, `401` fehlende/ungültige Auth, `403` unzureichende Rechte, `500` Server-Fehler.

**Bereichsgliederung:**

| Bereich        | Pfad-Präfix         | Auth                                                     |
| -------------- | ------------------- | -------------------------------------------------------- |
| Auth           | `/auth/*`           | — (öffentlich)                                           |
| Admin          | `/admin/*`          | JWT, Rolle `admin`                                       |
| Service        | `/service/*`        | JWT, Rolle `service`/`serviceleitung`/`admin`            |
| Senior Service | `/serviceleitung/*` | JWT, Rolle `serviceleitung`/`admin`                      |
| Relay          | `/relay/*`          | Statischer Token im Body (`RELAY_AUTH_TOKEN`) — kein JWT |

### 6.3 Frontend-Architektur

**Route Guards:** Zwei Guards schützen die Bereiche:

- `AdminGuard` — prüft, ob der eingeloggte Benutzer die Rolle `admin` hat.
- `ServiceGuard` — prüft, ob der Benutzer eingeloggt ist (Rolle `service`, `serviceleitung` oder `admin`).

Nicht autorisierte Zugriffe werden auf `/login` umgeleitet.

**Seitenstruktur:**

| Bereich   | Seiten                                                                                                                                                                                              |
| --------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Service   | Tischübersicht → Tisch-Detail (Tabs: Bestellen, Bezahlen, Historie). Ausgabe bestätigen ist in den Bestellen-Tab integriert; Stornieren ist für `serviceleitung`/`admin` im Bezahlen-Tab verfügbar. |
| Admin     | Produkte verwalten · Tische verwalten · Benutzer verwalten · **Druckerkonfiguration** (`DruckerConfigPage` — IP und Bonmodus pro Kategorie konfigurieren)                                           |
| Allgemein | Login · Passwort setzen (Erstanmeldung)                                                                                                                                                             |

**UI-Patterns:** Karten für Produkte/Tische, Drawer (Bottom-Sheet) für Bestell-/Bezahl-/Storno-Bestätigung, Tab-Navigation im Tisch-Detail, Plus/Minus-Buttons für Mengenauswahl (Touch-optimiert).

**BackendClient:** Das Frontend kommuniziert ausschließlich über Backend-Klassen, die das `BackendClient`-Interface verwenden. Direktes `fetch()` ist verboten.

### 6.4 Validierung

Alle Eingaben werden auf beiden Seiten unabhängig validiert: Frontend (Zod, vor Absenden) + Backend (zog, bei jedem Request). Das Backend ist die Single Source of Truth — das Frontend-Schema ist eine UX-Optimierung, keine Sicherheitsmaßnahme.

### 6.5 Geldbeträge

Alle Geldbeträge sind ganzzahlige Cent-Werte (`int` / `INTEGER` / JSON-Zahl) — durchgehend von Datenbank über Backend und API bis Frontend und Events. Keine Fließkommazahlen. Darstellung als „3,50 €“ erfolgt ausschließlich im Frontend (`formatCents()`).

### 6.6 Mehrbenutzerfähigkeit (OCC)

Das System verwendet zwei Persistenzstrategien:

| Bereich                               | Strategie      | Begründung                                                                |
| ------------------------------------- | -------------- | ------------------------------------------------------------------------- |
| Kasse (Tisch-Session + Kassensitzung) | Event-Sourcing | Geschichte ist fachlich relevant (Kassenjournal, Buchhaltung, Compliance) |
| Stammdaten (Produkt, Tisch, Benutzer) | CRUD           | Nur aktueller Zustand benötigt; Fat Events decken historische Daten ab    |

Mehrere Servicekräfte arbeiten gleichzeitig — auch am selben Tisch. Schreibkonflikte werden über Optimistic Concurrency Control gelöst (Subject- und OCC-Modell → [3.3](#33-subject-design-hierarchische-subjects)). Für den Mehrbenutzerbetrieb relevant ist der Retry: Jeder Schreibvorgang sendet die erwartete `event_version` mit; bei einem Konflikt lädt die Anwendungsschicht den Tischzustand neu und wiederholt den Vorgang.

### 6.7 Sicherheit

| Maßnahme                   | Umsetzung                                                                                                                       | Anforderung |
| -------------------------- | ------------------------------------------------------------------------------------------------------------------------------- | ----------- |
| HTTPS / TLS                | nginx terminiert TLS, Let's Encrypt-Zertifikat, HTTP → HTTPS-Redirect                                                           | Q-06        |
| Rate Limiting              | Login-Endpunkt ist durch Rate Limiting geschützt (Brute-Force-Schutz)                                                           | Q-07        |
| Security Headers           | Reverse Proxy setzt HSTS, X-Frame-Options, X-Content-Type-Options, CSP                                                          | Q-08        |
| Input-Validierung          | Frontend (Zod) + Backend (zog) — beide Seiten unabhängig voneinander                                                            | Q-03        |
| Passwort-Hashing           | Argon2id mit zufälligem Salt                                                                                                    | A-01        |
| Generische Fehlermeldungen | Fehlgeschlagene Logins geben keine Auskunft, ob Benutzer oder Passwort falsch war                                               | A-01        |
| Keine Secrets im Code      | Alle Secrets (JWT-Schlüssel, DB-Passwort, `RELAY_AUTH_TOKEN`) werden über Umgebungsvariablen konfiguriert                       | —           |
| JWT-Gültigkeit             | Tokens sind 12 Stunden gültig — kurze Lebensdauer begrenzt den Schaden bei Verlust                                              | A-01        |
| Relay-Token                | Statischer Token für `POST /relay/poll` und `POST /relay/quittieren` — kein JWT, kein Benutzerkontext. Relay ist kein Benutzer. | K-12        |

---

## 7. Read Models

Read Models sind aufbereitete Lese-Ansichten — reine Projektionen über vorhandene Daten (Events, Projektionstabelle oder Stammdaten). Sie werden nicht direkt geschrieben, sondern durch Events oder CRUD-Operationen aktualisiert.

### 7.1 Service-Ansichten

| Name             | ID   | Quelle                                           | Inhalt (Kurzfassung)                                                                                                                                                                              |
| ---------------- | ---- | ------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Tischübersicht   | K-06 | `tisch_sessions` + Stammdaten                    | Pro aktivem Tisch: Name, Saldo, Anzahl unbezahlter und ausstehender Positionen. Startseite des Service-Bereichs. JOIN auf `kassensitzung_nr`.                                                     |
| Tischdetails     | K-06 | `tisch_sessions`                                 | Alle Positionen mit Status, gruppiert nach Bestellung. Tabs: Übersicht, Bestellen, Ausgabe bestätigen, Bezahlen, Stornieren, Historie.                                                            |
| Produktkatalog   | —    | Produkt-Stammdaten                               | Aktive Produkte und Varianten, nach Kategorie gruppiert. Im Bestellvorgang geladen (kein eigenes Navigationsziel).                                                                                |
| Kassenjournal    | K-07 | Kassenjournal (Event Stream, Replay per Subject) | Chronologische Liste aller Vorgänge am Tisch: Zeitstempel, Typ, Positionen, Betrag, Servicekraft, Kommentar. Unveränderlich.                                                                      |
| Eigene Übersicht | R-06 | `kassenjournal` (SQL-Aggregation)                | KPIs der eigenen Servicekraft: Anzahl und Summe eigener Bestellungen sowie kassierter Zahlungen. Gefiltert auf `user_id` und `kassensitzung_nr`. Endpunkt: `POST /service/get-eigene-uebersicht`. |

Die operativen Ansichten (Tischübersicht, Tischdetails) lesen aus der synchronen Projektionstabelle `tisch_sessions` — kein Event-Replay nötig. Das Kassenjournal (Historie) liest weiterhin den vollständigen Event Stream via `ReadEventsBySubject()`. Details zur Projektionsarchitektur: [§3.8 Synchrone Projektion, CRUD-Entität und Event Replay](#38-synchrone-projektion-crud-entität-und-event-replay).

### 7.2 Admin-Ansichten (Reporting)

Alle Reporting-Ansichten aggregieren über `kassenjournal` und `tisch_sessions` (nur Admins, on-demand per SQL-Aggregation). Konsolidierter Endpoint `POST /admin/get-reporting` mit Sektionen `summary`, `breakdowns`, `stornierungen`. Filtert nach `kassensitzung_nr` statt Zeitraum. Kein Live-Dashboard, kein Polling.

Die `summary`-Sektion enthält zusätzlich die Direktverkauf-Kennzahl als aggregierte Sitzungsmetrik: `anzahlDirektverkaeufe` und `direktverkaufUmsatzCents` (netto aus Verkauf minus Storno).

| Name                        | ID   | Inhalt (Kurzfassung)                                                                               |
| --------------------------- | ---- | -------------------------------------------------------------------------------------------------- |
| Reporting (Unified)         | R-01 | KPIs (inkl. offene Tische), Umsatz pro Servicekraft/Tisch, Stornierungsübersicht, offene Betraege  |
| Abrechnung pro Tisch        | R-03 | Alle Bestellungen, Zahlungen, Ausgaben, Stornierungen chronologisch; Gesamt-Saldo pro Tisch        |
| Abrechnung pro Servicekraft | R-04 | Umsatz pro Servicekraft, Anzahl Bestellungen, Anzahl und Betrag der Stornierungen                  |
| Produktumsatz               | R-05 | Verkaufte Menge pro Produkt/Variante (abzgl. Stornierungen), Ranking, Gesamteinnahmen pro Variante |

### 7.3 Ausgabe-Ansichten

Der Relay-Poll-Endpunkt (`POST /relay/poll`) liefert die offenen Druckaufträge aus der `druckauftraege`-Outbox an das Print-Relay (reiner Transport, → [4.6 Bondruck](#46-bondruck-arbeitsbon-und-kassenbeleg-k-12)). KDS-Ansicht (K-13) und Zubereitungsstatus (K-15) sind noch offen.

---

## 8. Ubiquitous Language

Alle Fachbegriffe, Namenskonventionen pro Schicht, Code-Mappings und Ist-vs-Soll-Abweichungen: siehe **[Ubiquitous Language (language.md)](language.md)**.

Anforderungen mit Priorisierung (Must/Should/Nice-to-have) und Akzeptanzkriterien: siehe **[anforderungen.md](anforderungen.md)**.
