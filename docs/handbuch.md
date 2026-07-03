# Entwickler-Handbuch: jotti

> **Zweck:** Architektur-Referenz: Bounded Contexts, Aggregate, Invarianten und Design-Entscheidungen. Feld-Schemata und Implementierungsdetails stehen kanonisch im Code (`backend/domain/`, `database/migrations/`); Start und Betrieb im [README](../README.md) und in [betrieb/](betrieb/).

## 1. Überblick

### 1.1 Systemvision

jotti ist ein self-hosted mPOS-System (Go-Backend, React-Frontend, PostgreSQL, Docker Compose). Servicekräfte nutzen ihre eigenen Smartphones (BYOD) im Browser. Das Kassenjournal basiert auf Event-Sourcing; Stammdaten sind CRUD. Produktvision, Zielgruppe und Positionierung: siehe [produktbeschreibung.md](produktbeschreibung.md).

### 1.2 Designziele

| Ziel                    | Bedeutung                                                                                                                |
| ----------------------- | ------------------------------------------------------------------------------------------------------------------------ |
| Radikale Einfachheit    | Minimaler Funktionsumfang, der genau das abdeckt, was ein Vereinsfest braucht, nicht mehr.                              |
| Mobile-first            | Alle Interaktionen sind für Smartphone-Browser und Touch-Bedienung optimiert.                                            |
| Lückenlose Transparenz  | Jede Transaktion ist unveränderlich protokolliert. Kein Datenverlust, keine Manipulation.                                |
| Null Kosten             | Keine Hardware, keine Abo-Gebühren, keine externe Abhängigkeit.                                                          |
| Volle Datenhoheit       | Self-hosted, alle Daten auf dem eigenen Server.                                                                          |
| Niedrige Einstiegshürde | Keine Schulung, keine App-Installation. Browser öffnen, einloggen, loslegen.                                             |
| Nachvollziehbarkeit     | Event-Sourcing im Kassenjournal: Jede Bestellung, Zahlung, Stornierung und Kassenbewegung ist jederzeit nachvollziehbar. |

### 1.3 Bewusste Abgrenzung

Kartenzahlung, Reservierungen, Warenwirtschaft, Lieferservice, Multi-Standort, CRM und Kiosk-Modus sind bewusst ausgeschlossen. Vollständige Liste mit Begründung: siehe [produktbeschreibung.md §6.2](produktbeschreibung.md#62-was-jotti-bewusst-nicht-ist).

> **TSE / KassenSichV:** jotti unterliegt der TSE-Pflicht nach § 146a AO. Die TSE-Integration wird phasenweise implementiert: siehe [anforderungen.md](anforderungen.md) und [compliance.md](compliance.md).

---

## 2. Bounded Contexts

### 2.1 Kontextübersicht

| Context    | Typ                   | Beschreibung                                                                                                                                      | Persistenz                     |
| ---------- | --------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------ |
| Kasse      | Core Domain           | Alle finanziellen Geschäftsvorfälle: Bestellen, Ausgabe bestätigen, Bezahlen/Kassieren, Stornieren, Kassenbewegungen, Kassensturz, Tagesabschluss | Event-Sourcing (Kassenjournal) |
| Stammdaten | Supporting Sub-Domain | Verwaltung von Produkten, Tischen, Benutzern, Betreiber-Stammdaten (CRUD)                                                                         | CRUD                           |
| Auth       | Generic Sub-Domain    | Login, Logout, Passwort-Management, Token-Verwaltung                                                                                              | Infrastruktur                  |

> Bondruck (K-12) ist kein eigenständiger Bounded Context, sondern eine Policy innerhalb des Kasse-Context (→ [3.12 Policies](#312-policies)). Abrechnung/Reporting sind Read Models innerhalb der Kasse, kein eigener Context.

### 2.2 Beziehungen zwischen Kontexten

| Upstream   | Downstream | Beziehungstyp           | Beschreibung                                                                     |
| ---------- | ---------- | ----------------------- | -------------------------------------------------------------------------------- |
| Stammdaten | Kasse      | Customer/Supplier + ACL | Kasse liest Produkte/Tische, friert Daten zum Bestellzeitpunkt in Fat Events ein |
| Auth       | Kasse      | Open Host Service       | Token mit Benutzer-ID und Rolle                                                  |
| Auth       | Stammdaten | Open Host Service       | Token mit Benutzer-ID und Rolle                                                  |

Der Kasse-Kontext schützt sich über eine Anti-Corruption Layer (ACL) vor Stammdaten-Änderungen: Bestellungs-Events enthalten alle relevanten Produktdaten zum Zeitpunkt der Bestellung (Fat Events). Spätere Preis- oder Stammdaten-Änderungen haben keinen Einfluss auf historische Bestellungen und wirken erst in künftigen Bestellungen (Änderungssperre für Steuersätze folgt mit Compliance-Phase 1). Reporting-Projektionen aggregieren direkt über das Kassenjournal, keine Cross-Context-Kommunikation nötig.

---

## 3. Kasse (Core Domain)

Der Kasse-Kontext vereint alle finanziellen Geschäftsvorfälle mit Event-Sourcing über das Kassenjournal: tischbezogene Vorgänge (Bestellen, Ausgabe bestätigen, Bezahlen/Kassieren, Stornieren, Umbuchen) und kassenführungsbezogene Vorgänge (Kassensitzung eröffnen, Anfangsbestand, Kassenbewegungen, Kassensturz, Tagesabschluss).

### 3.1 Kassensitzung und Abrechnungskreis

| Begriff          | Scope                            | DSFinV-K-Feld                  | Beschreibung                                                                                                                            |
| ---------------- | -------------------------------- | ------------------------------ | --------------------------------------------------------------------------------------------------------------------------------------- |
| Kassensitzung    | Global, 1× pro Veranstaltungstag | `Z_NR` (Kassenabschlussnummer) | Der administrative Rahmen: Eröffnung durch Admin, Anfangsbestand, Kassenbewegungen, Kassensturz, Tagesabschluss (Z-Bon).                |
| Abrechnungskreis | Pro Tisch pro Kassensitzung      | `ABRECHNUNGSKREIS`             | Die buchhalterische Einheit: Alle Bestellungen, Zahlungen, Stornierungen und Umbuchungen an einem Tisch innerhalb einer Kassensitzung. |

Die Kassensitzung ist der Container, der Abrechnungskreis (= Tisch-Session) ist der Inhalt. Der `ABRECHNUNGSKREIS` ist pro Tisch pro Tag (DSFinV-K).

### 3.2 Kassenjournal (Event Store)

Das Kassenjournal (Tabelle `kassenjournal`) ist die zentrale, append-only Tabelle für alle finanziellen Geschäftsvorfälle, chronologische, vollständige, unveränderbare Aufzeichnung im Sinne von § 146 AO. Ein Immutabilitäts-Trigger verhindert UPDATE und DELETE.

Architektonisch tragende Spalten: `subject` (Stream-Schlüssel, → [3.3](#33-subject-design-hierarchische-subjects)), `version` (aufsteigend pro Subject; der Constraint `UNIQUE(subject, version)` realisiert OCC, → [6.6](#66-mehrbenutzerfähigkeit-occ)), `type` (Event-Typ, z. B. `bestellung-aufgenommen:v1`), `data` (JSONB), `user_id`/`user_name` (Fat Event: Name zum Zeitpunkt der Aktion) und `kassensitzung_nr`. Dieses Feld ermöglicht robuste Cross-Stream-Aggregationen (Reporting, Kassenbestand) ohne fragile LIKE-Patterns auf Subjects. Vollständiges Schema: `database/migrations/01_initial.up.sql`.

### 3.3 Subject-Design: Hierarchische Subjects

Subjects folgen einer hierarchischen Konvention mit zwei Ebenen:

```
kassensitzung-{nr}                             → Globaler Betriebstag (Kassensitzung)
kassensitzung-{nr}/tisch-{tischId}              → Abrechnungskreis (Tisch-Session)
kassensitzung-{nr}/direktverkauf-{uuid}         → Direktverkauf (ein Stream pro Verkauf)
```

**Kassensitzung-Subject:** `kassensitzung-1`, Nummer aus `kassensitzungen.z_nr`.

**Tisch-Session-Subject:** `kassensitzung-1/tisch-42`, entsteht implizit mit der ersten Bestellung (kein „Tisch-Öffnen"-Event).

**Direktverkauf-Subject:** `kassensitzung-1/direktverkauf-<uuid>`, ein eigener Stream pro Barverkauf an der Theke, ohne Projektion. `direktverkauf-getaetigt:v1` ist `version = 1`; positionsgenaue Stornierungen sind Folge-Versionen im selben Stream. Die Storno-Validierung läuft per On-Demand-Replay des einzelnen Verkauf-Streams: Es lassen sich nur Positionen stornieren, die noch nicht (vollständig) storniert wurden, höchstens in der ursprünglich verkauften Menge. Die Bargeld-Rückgabe ist Teil des Storno-Vorgangs selbst und mindert den Soll-Kassenbestand direkt, analog zur Warenrücknahme bezahlter Tisch-Positionen. Die kompakte Direktverkauf-Historie (eine Zeile pro Verkauf) entsteht durch Cross-Stream-Replay aller `direktverkauf-*`-Events der offenen Kassensitzung.

Separate Tisch-Subjects sind notwendig, weil der OCC-Constraint `UNIQUE(subject, version)` bei einem einzigen Subject alle Schreibvorgänge serialisieren würde, bei 5–30 Servicekräften nicht praktikabel.

**Kanonische Query-Strategie:**

| Zugriffsmuster                                              | Kanonische Strategie                  | Beispiel                                        |
| ---------------------------------------------------------- | ------------------------------------- | ----------------------------------------------- |
| Single-Stream-Replay (ein Tisch, eine KS)                  | Exakter `subject`-Match               | `WHERE subject = 'kassensitzung-1/tisch-42'`    |
| Cross-Stream-Aggregation (Reporting, Kassenbestand, Export) | `kassensitzung_nr`                    | `WHERE kassensitzung_nr = $1`                   |
| Tischübersicht (alle Tische einer KS)                      | `kassensitzung_nr` + `tisch_sessions` | JOIN auf Projektion                             |
| Globale Queries (alle KS eines Tisches, Debug)             | Subject-LIKE                          | `WHERE subject LIKE 'kassensitzung-%/tisch-42'` |

### 3.4 Tisch-Session (Abrechnungskreis-Aggregat)

Die Tisch-Session ist die transaktionale Grenze für tischbezogene Vorgänge. Jeder Tisch innerhalb einer Kassensitzung bildet einen eigenständigen Abrechnungskreis mit eigenem Event-Stream (→ [3.6](#36-domain-events)), eigener Versionierung und eigenem Saldo. Das Projektions-Modell (`TischSession`) materialisiert den aktuellen Zustand (→ [3.8](#38-synchrone-projektion-crud-entität-und-event-replay)). Die Projektion ist session-scoped, jede KS startet mit leerer Projektion. Produktdaten sind zum Bestellzeitpunkt eingefroren (Fat Events, → [2.2](#22-beziehungen-zwischen-kontexten)).

### 3.5 Kassensitzung-Lifecycle

Die Kassensitzung durchläuft: Eröffnung (Datum, Bezeichnung, Anfangsbestand/Wechselgeld) → Betrieb (Bestellungen, Ausgaben, Zahlungen, Stornierungen, Kassenbewegungen) → Kassensturz (Soll-Ist-Abgleich) → Tagesabschluss (Z-Bon, KS schließen). Alle KS-Events werden im selben Kassenjournal wie Tisch-Events gespeichert, Subject `kassensitzung-{nr}`.

### 3.6 Domain Events

Alle Events sind unveränderlich (append-only) und werden im Kassenjournal persistiert. Namenskonvention: deutsch, Partizip-Form, Pattern `{Substantiv}-{Partizip}:v{N}`. Die Feld-Schemata (Felder, Typen, Validierung) stehen kanonisch im Code: `backend/domain/kasse/*_events.go`.

**Tisch-Session-Events** (Subject `kassensitzung-{nr}/tisch-{id}`):

| Event                       | Semantik                                        | Tragende Constraints                                                                     |
| --------------------------- | ----------------------------------------------- | ---------------------------------------------------------------------------------------- |
| `bestellung-aufgenommen:v1` | Servicekraft nimmt eine Bestellung am Tisch auf | ≥ 1 Position; Produktname, Variante, Kategorie und Einzelpreis als Fat Event eingefroren |
| `ausgabe-bestaetigt:v1`     | Positionen als ausgegeben markiert              | Positionsbezug (`position_id` + `menge`); Teilausgaben möglich                           |
| `zahlung-kassiert:v1`       | Barzahlung kassiert                             | Betrag = Summe der gewählten Positionen; Teilzahlungen möglich                           |
| `stornierung-erteilt:v1`    | Kassenwirksame Warenrücknahme bezahlter Positionen | Genau eine `zahlungId` (FIFO je Zahlung); negativer Umsatz am Ursprungssteuersatz + Bar-Rückgabe; Kommentar Pflicht (min. 3 Zeichen) |
| `bestellung-korrigiert:v1`  | Geldneutrale Stornierung unbezahlter Positionen | Positionsbezug; ohne Geld- und Umsatzwirkung; Kommentar optional                         |
| `bestellung-umgebucht:v1`   | Geldneutrale Umbuchung unbezahlter Positionen zwischen zwei Tischen | Quell-/Zielstrom mit gemeinsamer `umbuchungId`; ohne Geldwirkung; Kommentar optional     |

**Direktverkauf-Events** (Subject `kassensitzung-{nr}/direktverkauf-{uuid}`):

| Event                        | Semantik                                                     | Tragende Constraints                                                                                                            |
| ---------------------------- | ------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------- |
| `direktverkauf-getaetigt:v1` | Barverkauf an der Theke: Bestellen + Zahlen in einem Schritt | Immer `version = 1` des Streams                                                                                                 |
| `direktverkauf-storniert:v1` | Positionsgenauer Storno eines Direktverkaufs                 | Folge-Version im selben Stream; Fat-Positionen; Bargeld-Rückgabe inklusive (→ [3.3](#33-subject-design-hierarchische-subjects)) |

**Kassensitzung-Events** (Subject `kassensitzung-{nr}`):

| Event                           | Semantik                                                              | Tragende Constraints                                                                         |
| ------------------------------- | --------------------------------------------------------------------- | -------------------------------------------------------------------------------------------- |
| `kassensitzung-eroeffnet:v1`    | Admin eröffnet die Kassensitzung (Datum, Bezeichnung, Anfangsbestand) | Anfangsbestand (`betragCents`) ist Teil der Eröffnung, kein eigenes Event                    |
| `geldtransit-gebucht:v1`        | Einlage oder Entnahme (`richtung`)                                    | Kommentar Pflicht (min. 3 Zeichen)                                                           |
| `kassensturz-durchgefuehrt:v1`  | Soll-Ist-Abgleich                                                     | Voraussetzung für den Tagesabschluss                                                         |
| `differenz-soll-ist-gebucht:v1` | Differenzbuchung nach Kassensturz                                     | Nur wenn Differenz ≠ 0; gleiche Transaktion wie der Kassensturz (→ [3.10](#310-kassensturz)) |
| `tagesabschluss-erstellt:v1`    | Z-Bon: aggregiert die Kassensitzung und schließt sie                  | `z_nr` fortlaufend, nie zurücksetzbar                                                        |

### 3.7 Invarianten

#### Tisch-Session-Invarianten

$$\text{Saldo} = \sum \text{Bestellungen} - \sum \text{Zahlungen} - \sum \text{Korrekturen} \pm \sum \text{Umbuchungen}$$

Alle Beträge in Cent (Integer); der Saldo ist die Summe der noch offenen (bestellten, nicht bezahlten, nicht korrigierten/umgebuchten) Positionen und stets ≥ 0. Saldo = 0 bedeutet: alle Positionen bezahlt, korrigiert oder umgebucht. Die kassenwirksame Warenrücknahme bereits bezahlter Positionen (`stornierung-erteilt:v1`) verändert den Saldo nicht; sie wirkt allein auf den Kassenbestand (Bar-Rückgabe).

- **Kassensitzung-Invariante:** Jeder schreibende Tisch-Vorgang erfordert eine offene Kassensitzung. Prüfung via `kassensitzungen`-Entität im Application Service. Keine offene KS → HTTP 409.
- **Ausgabe-Invariante:** Nur bestellte, nicht-stornierte Positionen können ausgegeben werden. Bereits ausgegebene Positionen nicht erneut ausgebbar. Teilausgaben zulässig.
- **Bezahl-Invariante:** Nur bestellte, nicht-stornierte, nicht-bezahlte Positionen können bezahlt werden. Der Zahlungsbetrag ergibt sich aus der Summe der gewählten Positionen, Überzahlung nicht möglich. Teilzahlungen zulässig.
- **Stornierungsinvariante:** Nur bestellte, nicht-stornierte Positionen sind stornierbar. Eine „Stornieren"-Anforderung wird serverseitig nach Bezahlstatus aufgeteilt: noch unbezahlte Positionen werden geldneutral korrigiert (`bestellung-korrigiert:v1`, Kommentar optional), bereits bezahlte als kassenwirksame Warenrücknahme zurückgenommen (`stornierung-erteilt:v1`, je betroffener Zahlung ein Event mit genau einer `zahlungId`, FIFO nach Zahlung, Kommentar Pflicht, min. 3 Zeichen). Die entstehenden Events werden atomar geschrieben, jedes mit eigener TSE-Transaktion; der Saldo bleibt stets ≥ 0.
- **Rolleninvariante:** Stornierungen nur durch `serviceleitung` und `admin`. Alle anderen Tischoperationen (Bestellen, Ausgabe bestätigen, Kassieren, Umbuchen) stehen allen drei Rollen zur Verfügung.
- **Mindestmengen-Invariante:** Jede positionsbasierte Operation erfordert mindestens eine Position. Bestellung, Ausgabe, Zahlung oder Stornierung ohne Positionen sind ungültig.

#### Kassensitzung-Invarianten

| Invariante                | Regel                                                                                                     |
| ------------------------- | --------------------------------------------------------------------------------------------------------- |
| Einzigkeits-Invariante    | Maximal eine Kassensitzung darf `offen` sein.                                                             |
| Nummern-Invariante        | `z_nr` ist fortlaufend und strikt aufsteigend (Identity-Sequenz beim INSERT in `kassensitzungen`); fehlgeschlagene Eröffnungen können technische Lücken hinterlassen. |
| Anfangsbestand-Invariante | Anfangsbestand ist `betragCents` in `kassensitzung-eroeffnet:v1`, kein eigenes Event.                     |
| Kassensturz-Reihenfolge   | `KassensturzDurchgefuehrt` ist Voraussetzung für `TagesabschlussErstellt`.                                |
| Tisch-Saldo-Sperre        | `TagesabschlussErstellt` ist nur möglich, wenn alle Tisch-Sessions der Kassensitzung Saldo = 0 haben.    |
| Abschluss-Invariante      | `TagesabschlussErstellt` schließt die KS → Status `abgeschlossen`. Danach keine Events mehr im Stream.    |

> **Keine Bewegungs-Invariante:** Kassenbewegungen werden ohne Prüfung des Soll-Bestands gebucht.

### 3.8 Synchrone Projektion, CRUD-Entität und Event Replay

Eine synchrone Projektion (`tisch_sessions`) + eine CRUD-Entität (`kassensitzungen`) werden in derselben Transaktion wie das Event-INSERT aktualisiert (Write-Through). Ein expliziter `StreamType`-Parameter steuert das Routing, kein Subject-String-Parsing im Repository-Layer.

**Routing via StreamType:**

| `streamType`      | Kassenjournal-INSERT | `kassensitzungen` | `tisch_sessions`     |
| ----------------- | -------------------- | ----------------- | -------------------- |
| `"kassensitzung"` | ✅                   | ✅ INSERT/UPDATE  | —                    |
| `"tisch-session"` | ✅                   | —                 | ✅ UPSERT            |
| `"direktverkauf"` | ✅                   | —                 | — (keine Projektion) |

Die Zustandsberechnung (`ApplyEvent()` in `backend/domain/kasse/tisch_session.go`) ist eine reine Funktion der Domain-Schicht (kein DB-Zugriff): Sie nimmt `TischSession` + `Event` entgegen und schreibt pro Event-Typ Saldo und Positionslisten fort.

**`kassensitzungen` (CRUD-Entität, Hot-Path):** hält nur `z_nr`, `datum` und `status` und wird bei jedem Tisch-Schreibvorgang gelesen (Kassensitzung-Sperre). Alle weiteren KS-Daten (Anfangsbestand, Bezeichnung, Kassenbewegungen) werden bei Bedarf per In-Memory-Replay der wenigen KS-Events berechnet.

**`tisch_sessions` (session-scoped Projektion):** pro Subject eine Zeile mit Tisch-Referenz, Saldo, unbezahlten und ausstehenden Positionen (JSONB) sowie der ID/Version des zuletzt verarbeiteten Events. Operative Queries lesen direkt aus der Projektion; die Historie liest den Event-Stream via `ReadEventsBySubject()`. Bei Inkonsistenz kann die Projektion jederzeit aus dem Kassenjournal neu berechnet werden (Single Source of Truth).

### 3.9 Kassenbestand (Read Model)

SQL-Aggregation über das Kassenjournal (eine `SELECT`-Query über `kassensitzung_nr`):

$$\text{Soll} = \text{Anfangsbestand}_{\text{KS}} + \sum_{\text{Tische}} \text{Zahlungen} - \sum_{\text{Tische}} \text{Warenrücknahmen} + \sum \text{Direktverkauf} - \sum \text{Direktverkauf-Storno} + \text{Kassenbewegungen}_{\text{netto}} + \text{DifferenzSollIst}$$

Alle Summanden stammen aus dem Kassenjournal. Keine Cross-Context-Projektion. Direktverkauf-Events (`direktverkauf-getaetigt:v1`, `direktverkauf-storniert:v1`) haben keine eigene Projektion, sind aber vollständig kassenwirksam und fließen in den Soll-Bestand ein.

### 3.10 Kassensturz

Am Ende einer Schicht vergleicht der Admin den errechneten Soll-Bestand (→ [3.9](#39-kassenbestand-read-model)) mit dem physisch gezählten Ist-Bestand. Der Application Service schreibt im Kassenabschluss (→ [3.11](#311-tagesabschluss-z-bon)) nacheinander `kassensturz-durchgefuehrt:v1` (immer) und `differenz-soll-ist-gebucht:v1` (nur wenn `differenz_cents ≠ 0`); eine umschließende Transaktion über die Abschluss-Events gibt es bewusst nicht. Differenzen werden nie per `UPDATE` korrigiert, sie sind eigene Geschäftsvorfälle: Das Differenz-Event bekommt eine eigene `kassenjournal.id` und ist direkt als Zeile in `businesscases.csv` exportierbar (`GV_TYP = DifferenzSollIst`).

Rechtliche Grundlagen und Betreiberpflichten (Zählprotokoll, Differenzbuchung, Aufbewahrung) → [compliance.md §4](compliance.md#4-gobd-konformität) und [§8](compliance.md#8-betreiberpflichten).

### 3.11 Tagesabschluss (Z-Bon)

Der Z-Bon ist das Ergebnis des `tagesabschluss-erstellt:v1`-Events, er aggregiert alle Geschäftsvorfälle einer Kassensitzung und erhält eine fortlaufende, nie zurücksetzbare `z_nr`.

**Invarianten:** `z_nr` fortlaufend und strikt aufsteigend, technische Lücken durch Fehlversuche möglich. Voraussetzung: Kassensturz durchgeführt + alle Tisch-Sessions Saldo = 0 (→ [3.7](#37-invarianten)). Das Event schließt die KS (→ Status `abgeschlossen`).

**Stammdaten-Snapshot:** Zu jedem Abschluss müssen die aktuell gültigen Stammdaten (Steuersätze, TSE-Zertifikate, Kassen-IDs) eingefroren werden, vor jeder Stammdaten-Änderung zunächst Kassenabschluss durchführen.

Rechtliche Grundlagen und Betreiber-Ablauf (Z-Bon statt X-Bon, Zählprotokoll, Aufbewahrung) → [compliance.md §8](compliance.md#8-betreiberpflichten).

### 3.12 Policies

- **Stornierungsberechtigung (K-04):** Nur `serviceleitung` und `admin` dürfen `StornierungErteilen`. Die Berechtigung wird in der Anwendungsschicht geprüft, bevor der Command an das Aggregat geht.
- **Arbeitsbon-Druck nach Kategorie (K-12):** Jedes `bestellung-aufgenommen:v1`-Event löst im Backend die Arbeitsbon-Policy aus, die Druckaufträge in die Outbox einreiht (→ [4.6 Bondruck](#46-bondruck-arbeitsbon-und-kassenbeleg-k-12)).
- **Umbuchung (K-09):** Verschiebt unbezahlte Positionen von Quell- auf Ziel-Tisch über ein geldneutrales `bestellung-umgebucht:v1` (Quell- und Zielstrom mit gemeinsamer `umbuchungId`). Cross-Aggregat-Transaktion, atomar geschrieben. Steht allen drei Rollen (`service`, `serviceleitung`, `admin`) zur Verfügung.

### 3.13 TSE-Architektur

> Compliance-spezifische Architektur-Entscheidungen für die TSE-Integration. Rechtliche Grundlagen → [compliance.md §3–§8](compliance.md).

**Signier-Interface:** Die Application-Schicht signiert über das anbieter-agnostische `TSEClient`-Interface (`StartTransaction` / `FinishTransaction`, → `backend/domain/tse/client.go`). Jeder jotti-Vorgang ist eine eigenständige, sofort geschlossene Transaktion (atomares „Festzelt-Muster"): `Start` ohne Inhalt, `Finish` mit dem finalen Schema. Ein `UpdateTransaction` ist nicht Teil des Interface, weil es nicht benötigt wird und für `Kassenbeleg-V1` ohnehin unzulässig ist (laut BMF-FAQ nur für `Bestellung-V1`/`SonstigerVorgang`). jotti adressiert die Transaktion über eine selbst erzeugte UUIDv4 (`tx_id`). Die Signaturdaten (`TSEData`: Transaktionsnummer, logTime, Signaturzähler, Signatur, Seriennummer, processType) werden in den Event-Daten persistiert; die Tisch-Session hält zusätzlich die logTime der ersten Bestellung für den Bon-Aufdruck. Orchestrierung und processData-Formatter liegen in `backend/api/tse/application`.

**Ausfallpfad (Nachsignieren):** Der Signierversuch läuft synchron im Kassier-Pfad mit kurzer Deadline. Schlägt er fehl (TSE-Ausfall/Timeout), wird der Vorgang trotzdem gebucht und ein Nachsignier-Auftrag (`tse_nachsignier_auftraege`) angelegt; ein Hintergrund-Worker signiert später nach und legt die Signatur in `tse_signaturen` ab. Admins können offene Aufträge zurücksetzen oder verwerfen (`/admin/get-tse-nachsignier-auftraege`, …). Fehlt beim Bon-Aufdruck noch eine logTime, greift die Event-Zeit als Fallback (AEAO 1.14.3).

**Vorgang → processType:** Bestellung aufnehmen, geldneutrale Korrektur (`bestellung-korrigiert`), Umbuchung (`bestellung-umgebucht`) → `Bestellung-V1`; Zahlung, kassenwirksame Warenrücknahme (`stornierung-erteilt`), Geldtransit, Kassendifferenz, Direktverkauf (inkl. Storno) → `Kassenbeleg-V1`; Tagesabschluss (Z-Bon) → `SonstigerVorgang`. Alle Transaktionen eines Tisches teilen denselben `ABRECHNUNGSKREIS`. Eigenbeleg- und Storno-Details im Export (BON_STORNO, REF_BON_ID, AEAO 2.2.3.6.1) → [compliance.md §6](compliance.md#6-dsfinv-k-export-schnittstelle).

**DSFinV-K-Export (geplant):** Ein Exporter soll Stammdaten-, Einzelaufzeichnungs- und Z-Bon-Daten als DSFinV-K-CSV (`transactions.csv`, `lines.csv`, `cashregister.csv`, `tse.csv`, …) mit `index.xml` und ZIP bündeln. Datei-Struktur und Pflichtfelder → [compliance.md §6](compliance.md#6-dsfinv-k-export-schnittstelle).

**Anbieter- und Meldeweg-Entscheidungen:** TSE-Anbieter (fiskaly als erster Zielanbieter; anbieter-agnostisches `TSEClient`-Interface gegen Vendor-Lock-in) und Kassenmeldungs-Weg (Phase 1 manuell über ELSTER, Phase 2 ERiC oder fiskaly-Submission-API) sind mitsamt Begründung und Abwägung in [compliance.md §3.5](compliance.md#35-tse-varianten-und-anbieter-entscheidung) und [§7](compliance.md#7-elektronische-meldepflicht-eric--elster) dokumentiert.

---

## 4. Stammdaten

Alle Stammdaten verwenden Soft-Delete via `status = 'deleted'`. Datensätze werden nie physisch gelöscht, referenzielle Integrität und historische Nachvollziehbarkeit (Fat Events im Kassenjournal) erfordern dies. Tabellen-Schemata: `database/migrations/01_initial.up.sql`.

### 4.1 Produkt-Aggregat

Das Produkt-Aggregat verwaltet den Produktkatalog der Veranstaltung. Jedes Produkt gehört zu einer Kategorie (`essen`, `getraenk`, `sonstiges`) und kann beliebig viele Varianten besitzen, jede Variante mit eigenem Namen und Preis (Cent, ≥ 0).

**Invarianten:** Produkt- und Variantennamen nicht leer; Kategorie gültig; Preis ≥ 0. Varianten können unabhängig vom Produkt deaktiviert werden (`inactive`) und erscheinen dann nicht im Service-Katalog.

### 4.2 Tisch-Stammdaten

Tisch-Stammdaten sind Name + Status. Nur aktive Tische (`active`) erscheinen in der Tischübersicht der Servicekräfte; der Name darf nicht leer sein.

**Abgrenzung zum Kasse-Kontext:** In den Stammdaten ist der Tisch eine CRUD-Entität; im Kasse-Kontext eine Event-Sourced Tisch-Session (Abrechnungskreis, → [3.4](#34-tisch-session-abrechnungskreis-aggregat)). Beide teilen `tisch_id`.

### 4.3 Benutzer-Aggregat

Das Benutzer-Aggregat verwaltet Zugangsdaten und Rollen (`admin`, `serviceleitung`, `service`) der Helfer und Admins.

**Invarianten:** Benutzername systemweit eindeutig; Rolle gültig; Passwörter nur als Argon2id-Hash persistiert, Klartext nie. Deaktivierte (`inactive`) und entfernte (`deleted`) Benutzer können sich nicht anmelden. Neue Benutzer starten mit Status `inactive` und einem Einmalpasswort aus 8 Zeichen (→ [5.2](#52-onboarding-ablauf)).

### 4.4 Tisch-Favoriten

Tisch-Favoriten sind eine CRUD-Relation Benutzer ↔ Tisch und steuern, welche Tische auf dem Service-Dashboard als „Meine Tische" angezeigt werden. Kein Aggregat, keine Events; Operationen idempotent (`ON CONFLICT DO NOTHING`), nur aktive Tische erlaubt.

### 4.5 Persistenz (CRUD)

Stammdaten (Produkte, Tische, Benutzer) werden mit klassischem CRUD verwaltet. Event-Sourcing ist hier nicht nötig, die historischen Daten stecken bereits in den Fat Events des Kasse-Context. Alle Stammdaten tragen `erstellt_am` und `aktualisiert_am` Zeitstempel.

### 4.6 Bondruck: Arbeitsbon und Kassenbeleg (K-12)

Bondruck umfasst zwei fachlich getrennte Bon-Familien auf einer gemeinsamen Druck-Infrastruktur. Sie teilen keinen Auslöser, Inhalt oder Rechtsstatus, nur die Druckauftrags-Outbox (`druckauftraege`) als Transport.

| Familie     | Auslöser                                        | Rechtsstatus           | Inhalt                                          |
| ----------- | ----------------------------------------------- | ---------------------- | ----------------------------------------------- |
| Arbeitsbon  | `bestellung-aufgenommen:v1` (automatisch)       | nicht-fiskalisch       | Ware ohne Preise (Küche/Theke)                  |
| Kassenbeleg | `POST /service/beleg-drucken` (auf Anforderung) | fiskalisch (§ 146a AO) | Positionen mit Preisen, Vereinsdaten, Kassen-ID |

**Arbeitsbon (operativ, K-12):** Eine Policy im Kasse-Context (→ [3.12 Policies](#312-policies)). Bei `bestellung-aufgenommen:v1` gruppiert die Arbeitsbon-Policy (`backend/api/bondruck`) die Positionen nach Kategorie, schlägt Drucker-IP und Bonmodus (`pro_position` oder `pro_bestellung`) aus der `druckstationen`-Tabelle nach (eine Zeile pro Kategorie; Admin-Konfiguration mit beidseitiger IPv4-Validierung), formatiert den ESC/POS-Payload und reiht je einen Druckauftrag in die Outbox ein. Kategorien ohne konfigurierte Druckstation erzeugen keinen Auftrag. Inhalt: Tischnummer, Positionen (Art + Menge), Kommentar, Uhrzeit, Servicekraft, keine Preise. Kein Beleg i. S. v. § 146a AO. KDS (K-13) und Zubereitungsstatus (K-15) sind noch offen.

**Kassenbeleg (fiskalisch, auf Anforderung):** Ein Service-Command (`POST /service/beleg-drucken`) erzeugt pro Anforderung genau einen Druckauftrag an den Kassenbeleg-Drucker. Als Datenquelle dient entweder eine Tischzahlung (`zahlung-kassiert:v1`) oder ein Direktverkauf (`direktverkauf-getaetigt:v1`); die Outbox-Referenz ist die Event-ID des referenzierten Vorgangs. Inhalt: Vereinsdaten (K-20), Kassen-Seriennummer (F-01), Datum/Uhrzeit, Positionen mit Einzelpreis × Menge, Gesamtbetrag, Zahlungsart „bar", Bon-Nummer. Erneuter Aufruf druckt nach, ohne den Vorgang fachlich zu wiederholen. Am Fest wird der Beleg selten verlangt (Belegausgabe-Befreiung → [compliance.md §5.1](compliance.md#51-gesetzliche-grundlage)). Der Beleg enthält die Steueraufteilung (F-07) und (sofern eine TSE konfiguriert ist) die TSE-Pflichtfelder inkl. QR-Code (F-02).

**Druckauftrags-Outbox (`druckauftraege`):** Single Source of Truth für alle Druckjobs, eine technische Warteschlange (Ziel-IP, ESC/POS-Payload, `bon_art`, fachliche Referenz), kein fiskalisches Journal. Statusmodell: `offen → gedruckt`; nach drei gemeldeten Fehlversuchen `fehlgeschlagen`, von dort `verworfen` oder zurück auf `offen`.

**Direktverkauf-Routing (Ableitungsregel):** Der Bondruck für `direktverkauf-getaetigt:v1` wird aus den konfigurierten Druckstationen abgeleitet: Ist die Abholbon-Station konfiguriert, entstehen Abholbons an dieser Station gemäß ihrem Bonmodus; sonst Arbeitsbons an die Produktstationen; ohne konfigurierte Stationen entsteht kein Auftrag. Der Kassenbeleg-Drucker ist die Druckstation `kassenbeleg`; fehlt ihre IP, schlägt `POST /service/beleg-drucken` mit klarer Fehlermeldung fehl.

**Relay = Transport:** Das Print-Relay (`windows/relay/main.go`) holt offene Aufträge via `POST /relay/poll`, druckt sie und meldet das Ergebnis via `POST /relay/ergebnis` (gedruckte IDs und Fehlversuche); das Backend setzt die Status entsprechend. Das Relay formatiert nichts, kennt keine Kategorien und führt keinen Cursor, der DB-Status ist autoritativ; noch offene Aufträge liefert der nächste Poll erneut (beim nicht-fiskalischen Arbeitsbon unkritisch). Start und Konfiguration → [README §Print-Relay](../README.md#print-relay).

---

## 5. Auth und Rollen

### 5.1 Rollen und Berechtigungsmatrix

jotti kennt drei Rollen mit abgestuften Berechtigungen. Die Rollenprüfung erfolgt serverseitig anhand des JWT.

| Rolle          | Code-Bezeichnung | Beschreibung                                                         |
| -------------- | ---------------- | -------------------------------------------------------------------- |
| Admin          | `admin`          | Voller Zugriff auf Stammdaten, Kasse (inkl. Kassensitzung) und Admin |
| Serviceleitung | `serviceleitung` | Kasse-Tischoperationen einschließlich Stornierung                    |
| Servicekraft   | `service`        | Kasse-Tischoperationen ohne Stornierung                              |

**Berechtigungsmatrix:**

| Aktion                         | Admin | Serviceleitung | Servicekraft |
| ------------------------------ | :---: | :------------: | :----------: |
| _Stammdaten_                   |       |                |              |
| Produkte verwalten             |   ✔   |                |              |
| Tische verwalten               |   ✔   |                |              |
| Benutzer verwalten             |   ✔   |                |              |
| Passwort zurücksetzen          |   ✔   |                |              |
| Betreiber-Stammdaten verwalten |   ✔   |                |              |
| _Kasse: Tisch-Operationen_     |       |                |              |
| Bestellung aufnehmen           |   ✔   |       ✔        |      ✔       |
| Ausgabe bestätigen             |   ✔   |       ✔        |      ✔       |
| Zahlung kassieren              |   ✔   |       ✔        |      ✔       |
| Stornierung erteilen           |   ✔   |       ✔        |              |
| Bestellung umbuchen            |   ✔   |       ✔        |      ✔       |
| Tischübersicht einsehen        |   ✔   |       ✔        |      ✔       |
| Kassenjournal einsehen         |   ✔   |       ✔        |      ✔       |
| _Kasse: Kassensitzung_         |       |                |              |
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

Die Rollenhierarchie ist inklusiv: Admin kann alles, was Serviceleitung kann. Serviceleitung kann alles, was Servicekraft kann, plus Stornierung.

### 5.2 Onboarding-Ablauf

Neue Benutzer durchlaufen einen zweistufigen Onboarding-Prozess, der sicherstellt, dass nur der Benutzer sein eigenes Passwort kennt:

1. **Benutzer anlegen:** Admin erstellt Benutzer (Name, Benutzername, Rolle, Status `inactive`). System generiert ein Einmalpasswort aus 8 Zeichen (Kleinbuchstaben und Ziffern ohne verwechselbare Zeichen), das der Admin dem Benutzer mitteilt.
2. **Erstanmeldung + Passwort setzen:** Benutzer meldet sich mit Einmalpasswort an. System erkennt am Zustand `einmalpasswort_hash ≠ NULL ∧ passwort_hash = NULL` den Onboarding-Status und leitet zur Passwort-Vergabe weiter (min. 6 Zeichen, Argon2id-Hash). Danach reguläre Anmeldung.

**Passwort-Reset:** Admin-Reset generiert neues Einmalpasswort, leert `passwort_hash` → Benutzer durchläuft Onboarding erneut.

---

## 6. Architekturprinzipien

### 6.1 Schichtenarchitektur

Das Backend ist in vier Schichten gegliedert: HTTP → Application → Domain → Repository/Infra.

- **HTTP-Schicht:** Request-Parsing, Response-Serialisierung, eigene DTOs mit `json`-Tags. Domain-Modelle nie direkt serialisiert, dedizierte Mapper. Keine Business-Logik.
- **Application-Schicht:** Use-Case-Koordination: fachliche Validierung (zog), Aggregat-State laden, Domain-Logik aufrufen, persistieren. Übersetzt Domain-Fehler in Fehlercodes.
- **Domain-Schicht:** Invarianten, Event-Konstruktion, Zustandsberechnung. Kein DB-, HTTP- oder JSON-Zugriff. Keine `json`-Tags (Ausnahme: Event-Data-Structs).
- **Repository/Infra-Schicht:** Datenbankzugriffe. Kasse: Kassenjournal (append-only) + synchrone Projektion + CRUD-Entität. Stammdaten: CRUD. Basis: sqlc-generierte Queries.

### 6.2 API-Design

**POST-only:** Alle API-Endpunkte sind POST-Endpunkte. Jede Aktion wird explizit benannt (z. B. `/service/bestellung-aufnehmen` statt `PUT /tables/5`).

**Bewusste Ops-Ausnahme:** `GET /health` ist explizit erlaubt, damit Container-Orchestrierung und Reverse-Proxy den Backend-Status per Healthcheck prüfen können. Diese Ausnahme gilt nur für `/health`; alle fachlichen Endpunkte bleiben strikt POST-only.

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
| Relay          | `/relay/*`          | Statischer Token im Body (`RELAY_AUTH_TOKEN`), kein JWT  |

### 6.3 Frontend-Architektur

**Route Guards:** Zwei Guards schützen die Bereiche:

- `AdminGuard`: prüft, ob der eingeloggte Benutzer die Rolle `admin` hat.
- `ServiceGuard`: prüft, ob der Benutzer eingeloggt ist (Rolle `service`, `serviceleitung` oder `admin`).

Nicht autorisierte Zugriffe werden auf `/login` umgeleitet.

**Seitenstruktur:**

| Bereich   | Seiten                                                                                                                                                                                               |
| --------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Service   | Tischübersicht → Tisch-Detail (Tabs: Bestellen, Kassieren, Historie). Ausgabe bestätigen ist in den Bestellen-Tab integriert; Stornieren ist für `serviceleitung`/`admin` im Historie-Tab verfügbar. |
| Admin     | Produkte verwalten · Tische verwalten · Benutzer verwalten · Druckerkonfiguration (`DruckerConfigPage`, IP und Bonmodus pro Kategorie konfigurieren)                                                 |
| Allgemein | Login · Passwort setzen (Erstanmeldung)                                                                                                                                                              |

**UI-Patterns:** Karten für Produkte/Tische, Drawer (Bottom-Sheet) für Bestell-/Bezahl-/Storno-Bestätigung, Tab-Navigation im Tisch-Detail, Plus/Minus-Buttons für Mengenauswahl (Touch-optimiert).

**BackendClient:** Das Frontend kommuniziert ausschließlich über Backend-Klassen, die das `BackendClient`-Interface verwenden. Direktes `fetch()` ist verboten.

### 6.4 Validierung

Alle Eingaben werden auf beiden Seiten unabhängig validiert: Frontend (Zod, vor Absenden) + Backend (zog, bei jedem Request). Das Backend ist die Single Source of Truth, das Frontend-Schema ist eine UX-Optimierung, keine Sicherheitsmaßnahme.

### 6.5 Geldbeträge

Alle Geldbeträge sind ganzzahlige Cent-Werte (`int` / `INTEGER` / JSON-Zahl), durchgehend von Datenbank über Backend und API bis Frontend und Events. Keine Fließkommazahlen. Darstellung als „3,50 €" erfolgt ausschließlich im Frontend (`formatCents()`).

### 6.6 Mehrbenutzerfähigkeit (OCC)

Das System verwendet zwei Persistenzstrategien:

| Bereich                               | Strategie      | Begründung                                                                |
| ------------------------------------- | -------------- | ------------------------------------------------------------------------- |
| Kasse (Tisch-Session + Kassensitzung) | Event-Sourcing | Geschichte ist fachlich relevant (Kassenjournal, Buchhaltung, Compliance) |
| Stammdaten (Produkt, Tisch, Benutzer) | CRUD           | Nur aktueller Zustand benötigt; Fat Events decken historische Daten ab    |

Mehrere Servicekräfte arbeiten gleichzeitig, auch am selben Tisch. Schreibkonflikte werden über Optimistic Concurrency Control gelöst (Subject- und OCC-Modell → [3.3](#33-subject-design-hierarchische-subjects)). Für den Mehrbenutzerbetrieb relevant ist der Retry: Jeder Schreibvorgang sendet die erwartete `event_version` mit; bei einem Konflikt lädt die Anwendungsschicht den Tischzustand neu und wiederholt den Vorgang.

### 6.7 Sicherheit

| Maßnahme                   | Umsetzung                                                                                                                     | Anforderung |
| -------------------------- | ----------------------------------------------------------------------------------------------------------------------------- | ----------- |
| HTTPS / TLS                | nginx terminiert TLS, Let's Encrypt-Zertifikat, HTTP → HTTPS-Redirect                                                         | Q-06        |
| Rate Limiting              | Login-Endpunkt ist durch Rate Limiting geschützt (Brute-Force-Schutz)                                                         | Q-07        |
| Security Headers           | Reverse Proxy setzt HSTS, X-Frame-Options, X-Content-Type-Options, CSP                                                        | Q-08        |
| Input-Validierung          | Frontend (Zod) + Backend (zog), beide Seiten unabhängig voneinander                                                           | Q-03        |
| Passwort-Hashing           | Argon2id mit zufälligem Salt                                                                                                  | A-01        |
| Generische Fehlermeldungen | Fehlgeschlagene Logins geben keine Auskunft, ob Benutzer oder Passwort falsch war                                             | A-01        |
| Keine Secrets im Code      | Alle Secrets (JWT-Schlüssel, DB-Passwort, `RELAY_AUTH_TOKEN`) werden über Umgebungsvariablen konfiguriert                     | —           |
| JWT-Gültigkeit             | Tokens sind 12 Stunden gültig, kurze Lebensdauer begrenzt den Schaden bei Verlust                                            | A-01        |
| Relay-Token                | Statischer Token für `POST /relay/poll` und `POST /relay/ergebnis`, kein JWT, kein Benutzerkontext. Relay ist kein Benutzer. | K-12        |

---

## 7. Read Models

Read Models sind aufbereitete Lese-Ansichten, reine Projektionen über vorhandene Daten (Events, Projektionstabelle oder Stammdaten). Sie werden nicht direkt geschrieben, sondern durch Events oder CRUD-Operationen aktualisiert.

### 7.1 Service-Ansichten

| Name             | ID   | Quelle                                           | Inhalt (Kurzfassung)                                                                                                                                                                              |
| ---------------- | ---- | ------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Tischübersicht   | K-06 | `tisch_sessions` + Stammdaten                    | Pro aktivem Tisch: Name, Saldo, Anzahl unbezahlter und ausstehender Positionen. Startseite des Service-Bereichs. JOIN auf `kassensitzung_nr`.                                                     |
| Tischdetails     | K-06 | `tisch_sessions`                                 | Alle Positionen mit Status, gruppiert nach Bestellung. Tabs: Übersicht, Bestellen, Ausgabe bestätigen, Bezahlen/Kassieren, Stornieren, Historie.                                                  |
| Produktkatalog   | —    | Produkt-Stammdaten                               | Aktive Produkte und Varianten, nach Kategorie gruppiert. Im Bestellvorgang geladen (kein eigenes Navigationsziel).                                                                                |
| Kassenjournal    | K-07 | Kassenjournal (Event Stream, Replay per Subject) | Chronologische Liste aller Vorgänge am Tisch: Zeitstempel, Typ, Positionen, Betrag, Servicekraft, Kommentar. Unveränderlich.                                                                      |
| Eigene Übersicht | R-06 | `kassenjournal` (SQL-Aggregation)                | KPIs der eigenen Servicekraft: Anzahl und Summe eigener Bestellungen sowie kassierter Zahlungen. Gefiltert auf `user_id` und `kassensitzung_nr`. Endpunkt: `POST /service/get-eigene-uebersicht`. |

Die operativen Ansichten (Tischübersicht, Tischdetails) lesen aus der synchronen Projektionstabelle `tisch_sessions`, kein Event-Replay nötig. Das Kassenjournal (Historie) liest weiterhin den vollständigen Event Stream via `ReadEventsBySubject()`. Details zur Projektionsarchitektur: [§3.8](#38-synchrone-projektion-crud-entität-und-event-replay).

### 7.2 Admin-Ansichten (Reporting)

Reporting ist Admin-only und wird on-demand per SQL-Aggregation über `kassenjournal` und `tisch_sessions` berechnet (kein Polling, kein eigener Schreibpfad). Zwei Endpunkte:

| Endpunkt                         | Scope                                  | Inhalt                                                                              |
| -------------------------------- | -------------------------------------- | ---------------------------------------------------------------------------------- |
| `POST /admin/get-live-reporting` | offene Kassensitzung (ohne Parameter)  | KPIs, offene Tische, offene Saldi, Stornierungen                                    |
| `POST /admin/get-abrechnung`     | bestimmte Kassensitzung (`kassensitzungNr`) | `summary`, `breakdowns` (Umsatz pro Servicekraft/Tisch), `umsatzProSteuersatz`, `stornierungen` |

Beide `summary`-Sektionen enthalten die Direktverkauf-Kennzahlen `anzahlDirektverkaeufe` und `direktverkaufUmsatzCents` (netto: Verkauf minus Storno). Anforderungs-IDs (R-01–R-05) → [anforderungen.md](anforderungen.md).

### 7.3 Ausgabe-Ansichten

Der Relay-Poll-Endpunkt (`POST /relay/poll`) liefert die offenen Druckaufträge aus der `druckauftraege`-Outbox an das Print-Relay (reiner Transport, → [4.6 Bondruck](#46-bondruck-arbeitsbon-und-kassenbeleg-k-12)). KDS-Ansicht (K-13) und Zubereitungsstatus (K-15) sind noch offen.

---

## 8. Ubiquitous Language

Alle Fachbegriffe, Namenskonventionen pro Schicht, Code-Mappings und Ist-vs-Soll-Abweichungen: siehe [Ubiquitous Language (language.md)](language.md).

Anforderungen mit Priorisierung (Must/Should/Nice-to-have) und Akzeptanzkriterien: siehe [anforderungen.md](anforderungen.md).
