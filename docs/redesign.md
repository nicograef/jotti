# Redesign: Kassenjournal als vereinigte Core Domain

## Inhaltsverzeichnis

1. [Kontext und Motivation](#1-kontext-und-motivation)
2. [Analyse: Probleme des aktuellen Modells](#2-analyse-probleme-des-aktuellen-modells)
3. [Neukonzeption: Kassenjournal als Core Domain](#3-neukonzeption-kassenjournal-als-core-domain)
4. [Pro und Contra](#4-pro-und-contra)
5. [Empfehlung](#5-empfehlung)
6. [Umsetzungsplan](#6-umsetzungsplan)

---

## 1. Kontext und Motivation

### 1.1 Ausgangslage

Das aktuelle Domänenmodell (Stand: `handbuch.md`) trennt zwei Bounded Contexts:

| Context           | Typ                   | Persistenz        |
| ----------------- | --------------------- | ----------------- |
| **Kassenbetrieb** | Core Domain           | Event-Sourcing    |
| **Kassenführung** | Supporting Sub-Domain | Immutable Records |

Die Kassenführung (Abrechnungskreis, Kassenbewegungen, Kassensturz, Z-Bon) ist aktuell als **Supporting Sub-Domain** mit **Immutable Records** (INSERT-only) modelliert:

> _„Kein OCC-Bedarf, kein Replay-Bedarf, Einfachheit."_ — handbuch.md §5.1

Die Kassenführung ist **noch nicht implementiert** — es gibt weder Domain-Code, noch Repositories, noch DB-Tabellen. Das Fenster für ein Redesign ist optimal.

### 1.2 These

> **Kassenbetrieb und Kassenführung gehören zu einer einzigen Core Domain „Kasse". Der Event Store wird zum Kassenjournal — einer einzigen Tabelle für alle finanziellen Geschäftsvorfälle. Die Kassensitzung bildet den globalen Betriebstag ab. Jeder Tisch innerhalb einer Kassensitzung bildet einen eigenständigen Abrechnungskreis (= Tisch-Session). Der Tisch wird als Stammdaten-Entität von der Kasse getrennt.**

---

## 2. Analyse: Probleme des aktuellen Modells

### 2.1 Bidirektionale Abhängigkeit

Das Handbuch definiert:

- **Kassenbetrieb → Kassenführung:** Events (ZahlungKassiert, StornierungErteilt, AuszahlungGeleistet) fließen in Kassenbestand und Reporting
- **Kassenführung → Kassenbetrieb:** Abrechnungskreis-Status sperrt/gibt Tischoperationen frei

```mermaid
graph LR
    subgraph Ist-Zustand
        KF[Kassenführung<br><i>Supporting Sub-Domain</i>]
        KB[Kassenbetrieb<br><i>Core Domain</i>]
        KF -- "Sperre / Freigabe" --> KB
        KB -- "Events liefern" --> KF
        style KF fill:#ffeeba
        style KB fill:#c3e6cb
    end
```

Diese bidirektionale Kopplung deutet darauf hin, dass die Verantwortlichkeiten eng genug verwoben sind, um in einem gemeinsamen Bounded Context kohärenter zu sein. Kassenbetrieb und Kassenführung teilen dieselben Daten (Events/Vorgänge), dieselben Invarianten (Kassenbestand darf nicht negativ werden) und denselben Lebenszyklus (Kassensitzung). Die Trennung erzwingt Cross-Context-Kommunikation für etwas, das fachlich zusammengehört.

### 2.2 Der Abrechnungskreis ist pro Tisch, nicht global

Das `handbuch.md` definiert den Abrechnungskreis als globale Kassensitzung. Die `abrechnungskreis.md`-Erkenntnisse zeigen: Der `ABRECHNUNGSKREIS` im Sinne der DSFinV-K ist **pro Tisch und Tag** (`Tisch-{Nr}-{YYYYMMDD}`), nicht ein einzelner Schlüssel für den gesamten Tag aller Tische. Ein tagesbasierter Gesamt-Schlüssel wäre ein Verstoß gegen die GoBD-Anforderung der Nachvollziehbarkeit.

**Konsequenz:** Was das Handbuch „Abrechnungskreis" nennt, vermischt zwei verschiedene Konzepte:

| Konzept                  | Granularität                 | Verantwortung                                                       |
| ------------------------ | ---------------------------- | ------------------------------------------------------------------- |
| **Globaler Betriebstag** | 1× pro Veranstaltungstag     | Tageseröffnung, Anfangsbestand, Kassenbewegungen, Z-Bon             |
| **Abrechnungskreis**     | 1× pro Tisch pro Betriebstag | DSFinV-K `ABRECHNUNGSKREIS`, Bestellungen, Zahlungen, Stornierungen |

### 2.3 Doppelrolle des Tisches

Im aktuellen Modell erfüllt der „Tisch" zwei Rollen:

| Rolle                | Kontext       | Verantwortung                                                                                              |
| -------------------- | ------------- | ---------------------------------------------------------------------------------------------------------- |
| **Physischer Tisch** | Stammdaten    | Ein Ort, an dem Gäste sitzen. Hat einen Namen, Status (active/inactive/deleted), wird vom Admin verwaltet. |
| **Tisch-Aggregat**   | Kassenbetrieb | Transaktionsgrenze für Bestellungen, Zahlungen, Stornierungen. Event-Sourced, mit OCC und Invarianten.     |

Der physische Tisch Nr. 42 ist ein **zeitloses Stammdatum**. Das Tisch-Aggregat mit Subject `tisch:42` hingegen ist ein **unbegrenzter, stetig wachsender Event-Stream**, der alle Bestellungen seit Beginn der Zeitrechnung akkumuliert. Es fehlt der natürliche Lebenszyklus: Niemand erwartet, dass der Maihock-Saldo vom Mai beim Weihnachtsmarkt im Dezember weitergezählt wird.

### 2.4 Zwei Persistenzstrategien ohne Grund

Das Handbuch begründet die Wahl von Immutable Records für die Kassenführung mit drei Argumenten:

| Argument             | Gegenkritik                                                                                                                                                   |
| -------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| _Kein OCC-Bedarf_    | OCC ist bei ES kostenlos mitgeliefert und schadet nicht. ES hat unabhängig davon Vorteile (Auditierbarkeit, Replay).                                          |
| _Kein Replay-Bedarf_ | Doch: Der Kassenbestand ist eine Projektion über Kassenbetrieb-Events **plus** eigene Records. Bei einem Event Store entfällt die hybride Berechnung.         |
| _Einfachheit_        | Ein einzelner Event Store für die gesamte Kasse ist langfristig **einfacher** als zwei Persistenzstrategien (ES + Immutable Records) mit Cross-Context-Logik. |

### 2.5 Compliance-Realität: Ein Aufzeichnungssystem

Die KassenSichV definiert jotti als **ein** elektronisches Aufzeichnungssystem. Der DSFinV-K-Export verlangt eine einheitliche Vorgangs-Sicht: Jeder Vorgang (Bestellung, Zahlung, Storno, Kassenbewegung) muss in denselben CSV-Dateien (`Bonkopf.csv`, `Bonpos.csv`) erscheinen — verknüpft über denselben `ABRECHNUNGSKREIS`. Die künstliche Trennung in zwei Kontexte mit zwei Persistenzstrategien erschwert diesen Export.

---

## 3. Neukonzeption: Kassenjournal als Core Domain

### 3.1 Grundidee

```mermaid
graph TB
    subgraph "Kasse (Core Domain)"
        KJ[("📒 Kassenjournal<br><i>Append-Only</i><br>Eine Tabelle, alle Vorgänge")]
        KS["Kassensitzung-Stream<br>Subject: kassensitzung-20260322<br><i>Eröffnung · Anfangsbestand · Kassenbewegung · Z-Bon</i>"]
        TS["Tisch-Abrechnungskreis-Streams<br>Subject: kassensitzung-20260322-tisch-42<br><i>Bestellen · Ausgabe · Zahlung · Storno</i>"]
        KS --> KJ
        TS --> KJ
        KJ --> KS_PROJ["kassensitzung_state<br><i>Synchrone Projektion</i>"]
        KJ --> TSS_PROJ["tisch_session_state<br><i>Synchrone Projektion (session-scoped)</i>"]
    end
    subgraph "Stammdaten (Supporting)"
        TISCH["tische<br><i>CRUD</i>"]
        PROD["produkte + varianten<br><i>CRUD</i>"]
        USER["users<br><i>CRUD</i>"]
    end
    subgraph "Auth (Generic)"
        AUTH["JWT · Login · Rollen"]
    end
    TISCH -. "FK in Query" .-> TS
    PROD -. "Fat Events (ACL)" .-> KJ
    AUTH -. "Autorisierung" .-> KJ
    style KJ fill:#d4edda
    style KS_PROJ fill:#c3e6cb
    style TSS_PROJ fill:#c3e6cb
    style TISCH fill:#ffeeba
    style PROD fill:#ffeeba
    style USER fill:#ffeeba
    style AUTH fill:#e2e3e5
```

**Drei Bounded Contexts, eine klare Rolle für jeden:**

| Context        | Typ         | Verantwortung                                                                                          | Persistenz                     |
| -------------- | ----------- | ------------------------------------------------------------------------------------------------------ | ------------------------------ |
| **Kasse**      | Core Domain | Alle finanziellen Geschäftsvorfälle: Bestellen, Bezahlen, Stornieren, Kassenbewegungen, Tagesabschluss | Event-Sourcing (Kassenjournal) |
| **Stammdaten** | Supporting  | Tische, Produkte, Benutzer verwalten                                                                   | CRUD                           |
| **Auth**       | Generic     | Login, Token, Rollen                                                                                   | Infrastruktur                  |

**Kernänderungen gegenüber dem Status Quo:**

1. **Ein Kassenjournal statt `events`-Tabelle** — fachlicher Name, dasselbe Schema
2. **Kassensitzung als globaler Betriebstag** — Subject `kassensitzung-{YYYYMMDD}` für Tageseröffnung, Kassenbewegungen, Z-Bon
3. **Abrechnungskreis = Tisch-Session** — Subject `kassensitzung-{YYYYMMDD}-tisch-{tischId}` pro Tisch pro Kassensitzung
4. **Physischer Tisch nur in Stammdaten** — die Tisch-Session (Abrechnungskreis) ist das Aggregat im Kasse-Kontext
5. **Keine bidirektionale Abhängigkeit** — alles ist ein Kontext
6. **Session-scoped `tisch_session_state`-Projektion** — ersetzt die globale `table_state` durch eine kassensitzungsbezogene Projektion mit natürlichem Lebenszyklus

---

### 3.2 Begriffsklärung: Kassensitzung und Abrechnungskreis

Die beiden Konzepte standen im bisherigen Modell synonym. Im neuen Modell haben sie klar getrennte Bedeutungen:

| Begriff              | Scope                            | DSFinV-K-Feld                  | Beschreibung                                                                                                                                                                                                                                                                                                                                                   |
| -------------------- | -------------------------------- | ------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Kassensitzung**    | Global, 1× pro Veranstaltungstag | `Z_NR` (Kassenabschlussnummer) | Der administrative Rahmen: Eröffnung durch Admin, Anfangsbestand, Kassenbewegungen, Kassensturz, Tagesabschluss (Z-Bon). Die `Z_NR` ist ein separater, fortlaufender Zähler — in Phase 1 identisch mit der Kassensitzungsnummer (ein Z-Bon pro KS), aber designtechnisch entkoppelt für spätere Erweiterungen (z. B. Zwischen-Z-Bon bei Stammdatenänderungen). |
| **Abrechnungskreis** | Pro Tisch pro Kassensitzung      | `ABRECHNUNGSKREIS`             | Die buchhalterische Einheit: Alle Bestellungen, Zahlungen, Stornierungen und Auszahlungen an einem Tisch innerhalb einer Kassensitzung. Identisch mit der Tisch-Session.                                                                                                                                                                                       |

**Warum diese Trennung?**

- Die DSFinV-K verlangt einen `ABRECHNUNGSKREIS` pro nachvollziehbarer Buchungseinheit — das ist der Tisch, nicht der gesamte Tag.
- Gleichzeitig existiert in der DSFinV-K die `Kassensitzungsnummer` (`Z_NR`) als globale Nummerierung der Kassenabschlüsse.
- Der „Abrechnungszeitraum" (Zeitspanne der Kassensitzung) ist eine Eigenschaft der Kassensitzung.

**Faustregel:** Die Kassensitzung ist der Container, der Abrechnungskreis (= Tisch-Session) ist der Inhalt.

#### Subject und `ABRECHNUNGSKREIS` — von Anfang an entkoppelt

Das Subject ist der systeminterne Stream-Schlüssel (OCC, Replay). Das DSFinV-K-Feld `ABRECHNUNGSKREIS` ist ein Freitextfeld (max. 50 Zeichen) ohne vorgeschriebenes Format. jotti leitet den Export-Wert aus dem Tischnamen ab (z. B. `Tisch 42`) — unabhängig vom Subject-Format. So bleibt das Subject eine rein technische Angelegenheit, und der Export-Wert kann ohne Systemänderung angepasst werden.

---

### 3.3 Kassenjournal statt Event Store

Die aktuelle Tabelle `events` ist fachlich und rechtlich ein **Kassenjournal**:

- **§ 146 AO** verlangt chronologische, vollständige, unveränderbare Aufzeichnung aller Geschäftsvorfälle.
- **KassenSichV** definiert das elektronische Aufzeichnungssystem, das genau diese Tabelle IST.
- **DSFinV-K** exportiert die Daten aus genau dieser Tabelle in die Bonkopf/Bonpos-CSVs.

```sql
CREATE TABLE kassenjournal (
    id          INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    user_id     INT REFERENCES users(id) NOT NULL,
    user_name   TEXT NOT NULL,
    type        TEXT NOT NULL,
    subject     TEXT NOT NULL,
    version     INT NOT NULL,
    timestamp   TIMESTAMPTZ NOT NULL,
    data        JSONB NOT NULL,
    kassensitzung_nr  INT NOT NULL,       -- FK-artige Zuordnung zur Kassensitzung
    UNIQUE (subject, version)
);

CREATE INDEX idx_kassenjournal_ks_nr ON kassenjournal(kassensitzung_nr);
```

Schema, Immutabilitäts-Trigger und OCC-Mechanismus bleiben identisch. Die `event.Event`-Struct im Go-Code bleibt ebenfalls unverändert.

Die `kassensitzung_nr`-Spalte ermöglicht robuste Queries ohne fragile LIKE-Patterns auf Subjects. Subjects bleiben trotzdem hierarchisch — sie dienen der OCC (`UNIQUE(subject, version)`) und dem Event-Replay eines einzelnen Streams. Aber Aggregationen über mehrere Streams verwenden die `kassensitzung_nr`-Spalte.

**Befüllung:**

- **KS-Events:** `kassensitzung_nr` = Z_NR der Kassensitzung (aus `kassensitzung_state`).
- **Tisch-Events:** `kassensitzung_nr` = Z_NR der aktuell offenen Kassensitzung (aus Application Service bekannt, da die KS-Sperre bereits geprüft wird).

---

### 3.4 Subject-Design: Hierarchische Subjects

#### Das Prinzip

Subjects folgen einer hierarchischen Konvention. Jeder Tisch-Abrechnungskreis ist dem übergeordneten Kassensitzung-Subject zugeordnet — erkennbar am gemeinsamen Präfix:

```
kassensitzung-{YYYYMMDD}                          → Globaler Betriebstag
kassensitzung-{YYYYMMDD}-tisch-{tischId}           → Abrechnungskreis (Tisch-Session)
```

#### Kassensitzung: `kassensitzung-{YYYYMMDD}`

```
kassensitzung-20260322      → Sommerfest 2026 Tag 1
kassensitzung-20260323      → Sommerfest 2026 Tag 2
kassensitzung-20261219      → Weihnachtsmarkt 2026
```

Das Datum stammt aus dem `KassensitzungEroeffnet`-Event und ist der **fachliche Tag** — bei Betrieb über Mitternacht bleibt das Datum des Eröffnungstags.

Die fortlaufende, lückenlose `Z_NR` (DSFinV-K-Pflichtfeld) lebt als eigener Zähler in der `kassensitzung_state`-Projektion — unabhängig vom Subject-Format.

#### Warum Datum im Subject statt Z_NR?

Da der `ABRECHNUNGSKREIS` kein Datum erfordert (siehe §3.2), wäre auch ein Z_NR-basiertes Format denkbar (`ks-1-tisch-42` statt `kassensitzung-20260322-tisch-42`). Das Datum wurde aus drei Gründen gewählt:

1. **Selbstdokumentation:** In Logs, DB-Queries und Debugging ist `kassensitzung-20260322` sofort verständlich — `ks-1` erfordert einen Lookup.
2. **Keine DB-Abhängigkeit:** Das Datum ist beim Eröffnen sofort bekannt. Die `z_nr` wird erst beim UPSERT in `kassensitzung_state` als `max(z_nr) + 1` berechnet — das Subject müsste auf den INSERT warten.
3. **Hierarchische Queries:** Datums-Präfixe sind feste 8 Zeichen lang, was LIKE-Queries sicher macht. Numerische Zähler (`ks-1` vs. `ks-12`) haben variable Länge.

#### Tisch-Abrechnungskreis: `kassensitzung-{YYYYMMDD}-tisch-{tischId}`

```
kassensitzung-20260322-tisch-42    → Tisch 42 am 22.03.2026
kassensitzung-20260322-tisch-7     → Tisch 7 am 22.03.2026
kassensitzung-20260323-tisch-42    → Tisch 42 am 23.03.2026 (neuer Tag, frischer Start)
```

Der Stream entsteht **implizit** mit der ersten Bestellung. Es gibt kein explizites „Tisch-Öffnen"-Event — die erste Bestellung für ein Subject erzeugt den Stream (Version 1).

#### Hierarchische Query-Patterns

Die hierarchische Subject-Struktur ermöglicht effiziente Abfragen auf verschiedenen Ebenen:

```sql
-- Alle Events einer Kassensitzung (global + alle Tische)
WHERE subject LIKE 'kassensitzung-20260322%'

-- Nur Tisch-Events einer Kassensitzung (ohne globale Events)
WHERE subject LIKE 'kassensitzung-20260322-tisch-%'

-- Alle Events eines bestimmten Tisch-Abrechnungskreises
WHERE subject = 'kassensitzung-20260322-tisch-42'

-- Alle Events eines Tisches über alle Kassensitzungen hinweg
WHERE subject LIKE 'kassensitzung-%-tisch-42'

-- Nur globale Kassensitzung-Events (exakter Match)
WHERE subject = 'kassensitzung-20260322'
```

> **Hinweis:** `LIKE 'kassensitzung-20260322%'` ist sicher, da Datums-Strings feste Länge haben und kein Datum Präfix eines anderen ist.

Die hierarchischen Queries dienen primär dem Event-Replay einzelner Streams und der OCC (`UNIQUE(subject, version)`). Für Aggregationen über mehrere Streams (z. B. Tischübersicht, Kassenbestand) ist die `kassensitzung_nr`-Spalte im Kassenjournal der bevorzugte Zugang (siehe §3.3).

#### Kanonische Query-Strategie

Das Kassenjournal bietet zwei Zugriffswege: exakte/LIKE-basierte Subject-Queries und die denormalisierte `kassensitzung_nr`-Spalte. Um Doppeldokumentation und Ad-hoc-Entscheidungen zu vermeiden, legt die folgende Tabelle fest, welcher Weg für welches Zugriffsmuster kanonisch ist:

| Zugriffsmuster                                                  | Kanonische Strategie                       | Beispiel                                            |
| --------------------------------------------------------------- | ------------------------------------------ | --------------------------------------------------- |
| **Single-Stream-Replay** (ein Tisch, eine KS)                   | Exakter `subject`-Match                    | `WHERE subject = 'kassensitzung-20260322-tisch-42'` |
| **Cross-Stream-Aggregation** (Reporting, Kassenbestand, Export) | `kassensitzung_nr`                         | `WHERE kassensitzung_nr = $1`                       |
| **Tischübersicht** (alle Tische einer KS)                       | `kassensitzung_nr` + `tisch_session_state` | JOIN auf Projektion                                 |
| **Globale Queries** (alle KS eines Tisches, Debug)              | Subject-LIKE                               | `WHERE subject LIKE 'kassensitzung-%-tisch-42'`     |

**Faustregel:** `kassensitzung_nr` für alles, was über einen einzelnen Stream hinausgeht. Subject-LIKE als Fallback für Ad-hoc-/Debug-Queries und für Zugriffe, die keine KS-Zuordnung benötigen.

#### Beziehung zwischen den Subjects

```mermaid
graph LR
    KS["kassensitzung-20260322<br><i>Eröffnet, Anfangsbestand, Kassenbewegungen, Z-Bon</i>"]
    T1["kassensitzung-20260322-tisch-42<br><i>AK für Tisch 42</i>"]
    T2["kassensitzung-20260322-tisch-43<br><i>AK für Tisch 43</i>"]
    T3["kassensitzung-20260322-tisch-7<br><i>AK für Tisch 7</i>"]
    KS -.-> T1
    KS -.-> T2
    KS -.-> T3
    style KS fill:#ffeeba
    style T1 fill:#c3e6cb
    style T2 fill:#c3e6cb
    style T3 fill:#c3e6cb
```

#### Warum zwei Subject-Typen statt einem?

Ein einziges Subject für die gesamte Kassensitzung scheitert an OCC: Der UNIQUE-Constraint `(subject, version)` serialisiert alle Schreibvorgänge desselben Subjects. Bei 5–30 Servicekräften und 5–50 Tischen wäre das System ein Single-Writer.

| Aggregat-Schnitt                    | Concurrency                                               | Praxistauglichkeit                                                              |
| ----------------------------------- | --------------------------------------------------------- | ------------------------------------------------------------------------------- |
| Ein Subject für alles               | Maximal 1 gleichzeitiger Schreiber für ALLE Tische        | 🚫 Nicht praktikabel                                                            |
| Ein Subject pro Tisch-AK            | Maximal 1 gleichzeitiger Schreiber **pro Tisch**          | ✅ Realistisch (selten bestellen zwei Personen an denselben Tisch gleichzeitig) |
| Separates Subject für Kassensitzung | Admin-Operationen kollidieren nicht mit Tisch-Operationen | ✅ Ideal (Admin-Aktionen sind selten)                                           |

---

### 3.5 Der Tisch in Stammdaten — und nur dort

Der Tisch ist eine reine Stammdaten-Entität:

```
tische
├── id          (INT, PK)
├── name        ("Tisch 42", "Tresen", "VIP links")
├── status      (active | inactive | deleted)
├── created_at
└── updated_at
```

Keine Event-Sourcing-Logik, keine Projektionen, keine Invarianten. Was der Kasse-Kontext vom Tisch braucht:

1. Die **ID** — um das Subject zu bilden (`kassensitzung-{YYYYMMDD}-tisch-{id}`)
2. Den **Status** — nur aktive Tische dürfen Bestellungen erhalten
3. Den **Namen** — für die Tischübersicht (im Read Model aufgelöst)

---

### 3.6 Event-Katalog

#### Kassensitzung-Events (Subject: `kassensitzung-{YYYYMMDD}`)

| Event-Typ                       | Beschreibung                                |
| ------------------------------- | ------------------------------------------- |
| `kassensitzung-eroeffnet:v1`    | Neue Kassensitzung eröffnet (Admin)         |
| `anfangsbestand-gesetzt:v1`     | Wechselgeld als Anfangsbestand gesetzt      |
| `kassenbewegung-gebucht:v1`     | Geldtransit, Privatentnahme oder -einlage   |
| `kassensturz-durchgefuehrt:v1`  | Soll-/Ist-Vergleich des Kassenbestands      |
| `differenz-soll-ist-gebucht:v1` | Kassendifferenz als eigenständiger Beleg    |
| `tagesabschluss-erstellt:v1`    | Z-Bon erstellt, Kassensitzung abgeschlossen |

#### Tisch-Abrechnungskreis-Events (Subject: `kassensitzung-{YYYYMMDD}-tisch-{id}`)

| Event-Typ                   | Beschreibung                        |
| --------------------------- | ----------------------------------- |
| `bestellung-aufgenommen:v1` | Bestellung mit Fat-Event-Positionen |
| `ausgabe-bestaetigt:v1`     | Positionen als ausgegeben markiert  |
| `zahlung-kassiert:v1`       | Barzahlung kassiert                 |
| `stornierung-erteilt:v1`    | Positionen storniert                |
| `auszahlung-geleistet:v1`   | Negativen Saldo ausgeglichen        |

> **Namenskonvention:** Alle Event-Typen folgen dem Pattern `{Substantiv}-{Partizip}:v{N}` — deutsche Fachbegriffe, kein technischer Namespace. `kassensitzung-` in `kassensitzung-eroeffnet:v1` ist kein Präfix, sondern Teil des Fachbegriffs „Kassensitzung eröffnet“.

#### Event-Data-Strukturen

**KassensitzungEroeffnet**

```
├── datum                 (date — YYYYMMDD, bestimmt das Subject)
├── bezeichnung           (string — z. B. „Sommerfest 2026 Tag 1“)
└── eroeffnet_von         (int — User-ID des Admins)
```

Die `nr` (Z_NR) wird nicht im Event gespeichert — sie wird beim UPSERT in `kassensitzung_state` als `max(z_nr) + 1` berechnet.

**AnfangsbestandGesetzt**

```
├── betrag_cents          (int — Wechselgeld)
└── gesetzt_von           (int — User-ID)
```

**KassenbewegungGebucht**

```
├── bewegung_id           (UUID)
├── art                   (geldtransit | privatentnahme | privateinlage)
├── betrag_cents          (int — ≥ 1)
├── kommentar             (string — Pflicht, min. 3, max. 200 Zeichen)
└── gebucht_von           (int — User-ID)
```

**KassensturzDurchgefuehrt**

```
├── soll_bestand_cents    (int — errechneter Soll)
├── ist_bestand_cents     (int — gezählter Ist)
├── differenz_cents       (int)
└── durchgefuehrt_von     (int — User-ID)
```

Der Application Service schreibt beim Kassensturz **zwei Events in derselben Transaktion**, wenn `differenz_cents ≠ 0`:

| Version | Event                           | Wann                   |
| ------- | ------------------------------- | ---------------------- |
| N       | `kassensturz-durchgefuehrt:v1`  | Immer                  |
| N+1     | `differenz-soll-ist-gebucht:v1` | Nur wenn Differenz ≠ 0 |

**DifferenzSollIstGebucht**

```
├── betrag_cents              (int — positiv = Überschuss, negativ = Fehlbetrag)
└── gebucht_von               (int — User-ID, identisch mit durchgefuehrt_von)
```

Das Event bekommt eine eigene `kassenjournal.id` — direkt exportierbar als Zeile in `businesscases.csv` mit `GV_TYP = DifferenzSollIst`. Das `differenz_cents`-Feld im `KassensturzDurchgefuehrt`-Event bleibt als informatives Feld für die UI (Anzeige des Soll/Ist-Vergleichs) — es ist keine Buchung, nur eine Feststellung.

**TagesabschlussErstellt**

```
├── z_nr                  (int — fortlaufend)
├── zeitraum_von          (datetime)
├── zeitraum_bis          (datetime)
├── umsatz_gesamt_cents   (int)
├── stornierungen_cents   (int)
├── auszahlungen_cents    (int)
├── geldtransit_cents     (int)
└── erstellt_von          (int — User-ID)
```

> **Kein Stammdaten-Snapshot im Event:** Die Stammdaten-Änderungssperre (siehe §3.7) garantiert, dass Stammdaten während einer offenen Kassensitzung unverändert bleiben. Der Z-Bon-Report kann die Stammdaten zum Abfragezeitpunkt laden — sie sind per Invariante identisch mit dem Zustand während der Kassensitzung. Falls compliance-relevant (GoBD-Aufbewahrung), kann der Snapshot als separater Export-Artefakt beim Z-Bon-Druck erzeugt werden, nicht als Feld im Event.

---

### 3.7 Invarianten

#### Tisch-Abrechnungskreis-Invarianten

Die bestehenden Invarianten (Saldo, Ausgabe-, Bezahl-, Stornierungsinvariante) bleiben inhaltlich identisch. Ergänzung:

- **Kassensitzung-Invariante:** Jeder schreibende Tisch-Vorgang erfordert eine offene Kassensitzung. Prüfung via `kassensitzung_state`-Projektion im Application Service. Keine offene KS → HTTP 409.

#### Kassensitzung-Invarianten

| Invariante                    | Regel                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| ----------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Einzigkeits-Invariante**    | Maximal eine Kassensitzung darf `offen` sein.                                                                                                                                                                                                                                                                                                                                                                                                     |
| **Nummern-Invariante**        | `z_nr` ist fortlaufend und lückenlos (`max(z_nr) + 1`). Wird beim UPSERT in `kassensitzung_state` in derselben Transaktion wie das Event-INSERT berechnet. Die `UNIQUE`-Constraint auf `z_nr` verhindert Duplikate. Da Kassensitzungen nur von Admins eröffnet werden, ist Concurrency extrem unwahrscheinlich. Falls zwei Admins gleichzeitig eröffnen: Der zweite Request bekommt einen Constraint-Violation-Fehler und muss wiederholt werden. |
| **Anfangsbestand-Invariante** | Pro Kassensitzung genau ein `AnfangsbestandGesetzt`. Wiederholter Aufruf wird abgelehnt.                                                                                                                                                                                                                                                                                                                                                          |
| **Kassensturz-Reihenfolge**   | `KassensturzDurchgefuehrt` ist Voraussetzung für `TagesabschlussErstellt`.                                                                                                                                                                                                                                                                                                                                                                        |
| **Tisch-Saldo-Sperre**        | `TagesabschlussErstellt` ist nur möglich, wenn **alle** Tisch-AKs der Kassensitzung Saldo = 0 haben. Der Admin muss vor dem Z-Bon alle Tische abkassieren oder unbezahlte Positionen stornieren.                                                                                                                                                                                                                                                  |
| **Abschluss-Invariante**      | `TagesabschlussErstellt` schließt die KS → Status `abgeschlossen`. Danach keine Events im Stream.                                                                                                                                                                                                                                                                                                                                                 |

> **Keine Bewegungs-Invariante:** Kassenbewegungen (Geldtransit, Privatentnahme, Privateinlage) werden ohne Prüfung des Soll-Bestands gebucht. Stattdessen zeigt der Kassensturz eine **Warnung**, wenn der Soll-Bestand negativ ist: „Soll-Bestand ist negativ — bitte Kassenbewegungen prüfen.“ Der Kassensturz ist der fachlich korrekte Ort für die Soll-Ist-Abweichungserkennung. Eine Schreibsperre beim Buchen würde den Betrieb blockieren, wenn der Admin tatsächlich mehr Geld entnehmen muss als rechnerisch vorhanden (z. B. weil Wechselgeld privat nachgelegt wurde, ohne es als Privateinlage zu buchen).

> **Stammdaten-Änderungssperre (Compliance, Phase 1):** Wenn Steuersätze, TSE-Konfiguration oder Seriennummer geändert werden, muss die Software zuvor automatisch einen Tagesabschluss durchführen. Für Produktpreise ist dies nicht erforderlich, da Fat Events die Preise zum Bestellzeitpunkt einfrieren. Diese Invariante wird mit der Compliance-Phase 1 implementiert.

#### Kassensitzung-Sperre (Cross-Stream)

Die Sperre ist die einzige Stream-übergreifende Invariante:

```mermaid
sequenceDiagram
    participant S as Servicekraft
    participant AS as Application Service
    participant KSP as KS-Projektion
    participant KJ as Kassenjournal

    S->>AS: BestellungAufnehmen(tischId=42, ...)
    AS->>KSP: GetOffeneKassensitzung()
    alt Keine offene Kassensitzung
        KSP-->>AS: nil
        AS-->>S: HTTP 409 "Kasse ist noch nicht geöffnet"
    else Kassensitzung offen
        KSP-->>AS: KS{Nr: 1, Status: offen}
        Note over AS: Subject = "kassensitzung-20260322-tisch-42"
        AS->>KJ: WriteEvent(bestellung-aufgenommen:v1, ...)
        KJ-->>AS: OK
        AS-->>S: HTTP 200
    end
```

**Konsistenzgarantie:** Die `kassensitzung_state`-Projektion wird in derselben Transaktion wie das Event-INSERT aktualisiert (synchron Write-Through für KS-Events). Die Tisch-Events lesen die KS-Projektion VOR dem Event-Write (separate Transaktion). Es gibt ein minimales Race-Condition-Fenster: Ein Admin könnte die KS abschließen, während gleichzeitig eine Servicekraft eine Bestellung aufgibt. Das ist akzeptabel: Die Tisch-Saldo-Sperre erzwingt Saldo = 0 auf allen Tischen vor dem Abschluss — die Wahrscheinlichkeit, dass in diesem Moment noch eine Bestellung eingeht, ist vernachlässigbar.

---

### 3.8 Projektionsstrategie: Zwei synchrone Projektionen

Beide Projektionen werden als synchrone Write-Through-Updates in derselben Transaktion wie das Event-INSERT aktualisiert. Ein expliziter `StreamType`-Parameter steuert das Routing — kein Subject-String-Parsing im Repository-Layer.

#### Routing via `StreamType`

Der Application Service weiß bereits, welchen Stream-Typ er beschreibt. Statt im Repository das Subject zu parsen, übergibt er den Stream-Typ explizit:

```
WriteEvent(event, streamType: "kassensitzung" | "tisch-session")
```

Das Routing-Schema:

| `streamType`      | Kassenjournal-INSERT | `kassensitzung_state` | `tisch_session_state` |
| ----------------- | -------------------- | --------------------- | --------------------- |
| `"kassensitzung"` | ✅                   | ✅ UPSERT             | —                     |
| `"tisch-session"` | ✅                   | —                     | ✅ UPSERT             |

#### `kassensitzung_state` — Hot-Path-Projektion

Die Kassensitzung braucht eine **minimale** Projektion: Der Status (`offen`/`abgeschlossen`) wird bei **jedem** Tisch-Schreibvorgang geprüft (Kassensitzung-Sperre). Ohne Projektion müsste jeder Request einen Event-Replay des Kassensitzung-Streams durchführen.

Die Projektion speichert nur die Felder, die im Hot Path benötigt werden — Status und `z_nr` für die KS-Sperre. Alle anderen Kassensitzung-Daten (Anfangsbestand, Bezeichnung, Kassenbewegungen, Kassensturz-Ergebnisse) werden bei Bedarf per In-Memory-Replay der 3–5 KS-Events berechnet.

```sql
CREATE TABLE kassensitzung_state (
    subject            TEXT PRIMARY KEY,        -- 'kassensitzung-20260322'
    z_nr               INT UNIQUE NOT NULL,     -- 1, 2, 3 ... (fortlaufend, lückenlos)
    datum              DATE NOT NULL,           -- 2026-03-22 (aus Eröffnung)
    status             TEXT NOT NULL CHECK (status IN ('offen', 'abgeschlossen')),
    last_event_id      INT NOT NULL REFERENCES kassenjournal(id),
    last_event_version INT NOT NULL
);
```

> **Warum keine `bezeichnung`?** Das Feld „Sommerfest Tag 1" wird nie im Hot Path gelesen — nur in Admin-UIs, die ohnehin die KS-Events replayed. Die Projektion bleibt auf das Minimum beschränkt: Status-Check und `z_nr` für die KS-Sperre und Reporting-Zuordnung.

#### `tisch_session_state` — Session-scoped Tisch-Projektion

Das Service-Dashboard (Tischübersicht) ist der meistgenutzte Endpunkt — Servicekräfte pollen ihn alle paar Sekunden. Eine synchrone Projektion macht diesen Read-Pfad trivial, ohne das Redesign-Fundament zu verändern.

Im Gegensatz zur alten `table_state` (PK: `tisch_id`, unbegrenztes Wachstum) ist die neue Projektion **session-scoped**: Der PK ist das Subject (`kassensitzung-{YYYYMMDD}-tisch-{id}`). Jede Kassensitzung startet mit einer leeren Projektion — natürlicher Lebenszyklus, keine Altlasten.

```sql
CREATE TABLE tisch_session_state (
    subject                TEXT PRIMARY KEY,  -- 'kassensitzung-20260322-tisch-42'
    tisch_id               INT NOT NULL REFERENCES tische(id),
    kassensitzung_nr       INT NOT NULL,      -- Denormalisiert für schnelle Queries
    saldo_cents            INT NOT NULL DEFAULT 0,
    unbezahlte_positionen  JSONB NOT NULL DEFAULT '[]',
    ausstehende_positionen JSONB NOT NULL DEFAULT '[]',
    gesamt_zahlungen_cents INT NOT NULL DEFAULT 0,
    last_event_id          INT NOT NULL REFERENCES kassenjournal(id),
    last_event_version     INT NOT NULL,
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tisch_session_state_ks_nr ON tisch_session_state(kassensitzung_nr);
```

**Vorteile gegenüber der alten `table_state`:**

- **Session-scoped:** Jede Kassensitzung startet mit einer leeren Projektion. Kein unbegrenztes Wachstum.
- **`kassensitzung_nr`:** Ermöglicht Tischübersicht per KS ohne Subject-Parsing.
- **Natürlicher Lebenszyklus:** Die Projektion lebt so lange wie die Kassensitzung.

**Tischübersicht wird trivial:**

```sql
SELECT t.id, t.name, COALESCE(tss.saldo_cents, 0) AS saldo_cents
FROM tische t
LEFT JOIN tisch_session_state tss ON tss.tisch_id = t.id AND tss.kassensitzung_nr = $1
LEFT JOIN tisch_favoriten f ON f.tisch_id = t.id AND f.user_id = $2
WHERE t.status = 'active'
ORDER BY f.tisch_id IS NOT NULL DESC, t.name;
```

Kein JSONB-Parsing, kein Regex, kein CASE/WHEN. Ein einfacher JOIN.

**Einzeltisch-Read:**

```sql
SELECT * FROM tisch_session_state WHERE subject = $1;
```

Für Invarianten-Prüfung (Bezahl-/Stornierungsinvariante) ist der State sofort verfügbar — kein Replay nötig.

#### Write-Through-Pseudocode

```
WriteEvent(event, streamType):
  BEGIN TX
    1. INSERT INTO kassenjournal (...)
    2. IF streamType = "kassensitzung":
         IF type = "kassensitzung-eroeffnet:v1":
           → INSERT kassensitzung_state (subject, z_nr, datum, status='offen', ...)
         ELSE IF type = "tagesabschluss-erstellt:v1":
           → UPDATE kassensitzung_state SET status='abgeschlossen', last_event_id, last_event_version
         ELSE:
           → UPDATE kassensitzung_state SET last_event_id, last_event_version
    3. ELSE IF streamType = "tisch-session":
         → UPSERT tisch_session_state (subject, tisch_id, kassensitzung_nr,
             saldo_cents, unbezahlte_positionen, ausstehende_positionen,
             gesamt_zahlungen_cents, last_event_id, last_event_version)
    4. COMMIT
```

## **Trade-off:** Der Write-Pfad bleibt Zwei-Operationen-pro-TX (INSERT + UPSERT) — identisch mit dem bisherigen Pattern. Das ist der bewusste Kompromiss: Write-Pfad minimal komplexer, Read-Pfad drastisch einfacher — für den meistgenutzten Endpunkt die richtige Entscheidung.

### 3.9 Kassenbestand

Der Kassenbestand (Soll) ist die Summe aus Anfangsbestand, allen Zahlungen, Auszahlungen und Kassenbewegungen:

$$\text{Soll} = \text{Anfangsbestand}_{\text{KS}} + \sum_{\text{Tische}} \text{Zahlungen} - \sum_{\text{Tische}} \text{Auszahlungen} + \text{Kassenbewegungen}_{\text{netto}} + \text{DifferenzSollIst}$$

Alle Summanden stammen aus dem Kassenjournal — eine einzige Aggregation über die `kassensitzung_nr`:

```sql
SELECT COALESCE(SUM(CASE
    WHEN type = 'anfangsbestand-gesetzt:v1'
        THEN (data->>'betragCents')::INT
    WHEN type = 'zahlung-kassiert:v1'
        THEN (data->>'gesamtZahlungCents')::INT
    WHEN type = 'auszahlung-geleistet:v1'
        THEN -(data->>'betragCents')::INT
    WHEN type = 'kassenbewegung-gebucht:v1' AND data->>'art' = 'privateinlage'
        THEN (data->>'betragCents')::INT
    WHEN type = 'kassenbewegung-gebucht:v1' AND data->>'art' IN ('privatentnahme', 'geldtransit')
        THEN -(data->>'betragCents')::INT
    WHEN type = 'differenz-soll-ist-gebucht:v1'
        THEN (data->>'betragCents')::INT
    ELSE 0
END), 0) AS soll_bestand_cents
FROM kassenjournal
WHERE kassensitzung_nr = $1
  AND type IN (
    'anfangsbestand-gesetzt:v1',
    'zahlung-kassiert:v1',
    'auszahlung-geleistet:v1',
    'kassenbewegung-gebucht:v1',
    'differenz-soll-ist-gebucht:v1'
  );
```

Eine SQL-Query, eine Tabelle — keine Cross-Context-Projektion, kein Event-Bus, kein Cross-Join mit der Kassensitzung-Projektion.

#### Reporting: Filterung nach Kassensitzung statt Zeitraum

Alle aktuellen Reporting-Queries (`GetReportingStats`, `GetUmsatzProServicekraft`, `GetUmsatzProTisch`, `GetStornierungen`) filtern nach `timestamp BETWEEN $1 AND $2`. Im neuen Modell ist Filterung nach `kassensitzung_nr` fachlich korrekt — ein Event nach Mitternacht gehört noch zum Betriebstag, nicht zum Folgetag. Die Kassensitzung definiert den buchhalterischen Abrechnungszeitraum, nicht die Uhrzeit.

**Migration der Reporting-Queries:**

| Alt                                                 | Neu                              |
| --------------------------------------------------- | -------------------------------- |
| `WHERE timestamp BETWEEN $1 AND $2`                 | `WHERE kassensitzung_nr = $1`    |
| Zeitraum-Parameter (von/bis) im Application Service | `kassensitzung_nr` als Parameter |
| Edge Case: Event um 00:03 Uhr gehört zum Folgetag   | Event gehört zur offenen KS      |

Der Wechsel auf `kassensitzung_nr` als Filterkriterium eliminiert Zeitstempel-Edge-Cases (Mitternacht, Zeitzonenwechsel) und macht Reports exakt reproduzierbar.

---

### 3.10 DSFinV-K-Kompatibilität

| DSFinV-K Feld      | Quelle im neuen Modell                                                                                     |
| ------------------ | ---------------------------------------------------------------------------------------------------------- |
| `Z_KASSE_ID`       | Seriennummer (Config)                                                                                      |
| `Z_ERSTELLUNG`     | Timestamp des `tagesabschluss-erstellt:v1`-Events                                                          |
| `Z_NR`             | `kassensitzung_state.z_nr` (separater fortlaufender Zähler, in Phase 1 identisch mit Kassensitzungsnummer) |
| `BON_ID`           | `kassenjournal.id`                                                                                         |
| `BON_NR`           | `kassenjournal.version` pro Subject                                                                        |
| `BON_TYP`          | Mapping von `kassenjournal.type`                                                                           |
| `ABRECHNUNGSKREIS` | Abgeleiteter Tischname (z. B. `Tisch 42`), entkoppelt vom systeminternen Subject (siehe §3.2)              |
| `TISCH_NR`         | Aus Subject extrahiert (Segment nach `tisch-`)                                                             |

Alle Daten in **einer Tabelle** (Kassenjournal) + zwei Projektionen (`kassensitzung_state`, `tisch_session_state`). Kein Cross-Context-Join.

> **Hinweis:** Der `ABRECHNUNGSKREIS`-Wert wird beim DSFinV-K-Export aus dem Tischnamen abgeleitet — nicht aus dem Subject (siehe §3.2). Damit bleibt das Subject-Format eine rein systeminterne Angelegenheit.

---

### 3.11 Lebenszyklus einer Kassensitzung

```mermaid
sequenceDiagram
    participant A as Admin
    participant S as Servicekraft
    participant AS as Application Service
    participant KJ as Kassenjournal
    participant KSP as KS-Projektion
    participant TSS as Tisch-Session-Projektion

    Note over A,TSS: Admin eröffnet den Betriebstag

    A->>AS: KassensitzungEroeffnen("Sommerfest Tag 1")
    AS->>KJ: INSERT (subject="kassensitzung-20260322", version=1, type="kassensitzung-eroeffnet:v1")
    AS->>KSP: UPSERT kassensitzung_state (subject='kassensitzung-20260322', z_nr=1, datum='2026-03-22', status=offen)

    A->>AS: AnfangsbestandSetzen(300€)
    AS->>KJ: INSERT (subject="kassensitzung-20260322", version=2, type="anfangsbestand-gesetzt:v1")

    Note over A,TSS: Servicekräfte arbeiten

    S->>AS: BestellungAufnehmen(tischId=42, ...)
    AS->>KSP: GetOffeneKassensitzung() → KS{Nr: 1}
    Note over AS: Subject = "kassensitzung-20260322-tisch-42", Version = 1
    AS->>KJ: INSERT (subject="kassensitzung-20260322-tisch-42", version=1, type="bestellung-aufgenommen:v1")
    AS->>TSS: UPSERT tisch_session_state (saldo += Bestellwert)

    S->>AS: ZahlungKassieren(tischId=42, ...)
    AS->>KSP: GetOffeneKassensitzung() → KS{Nr: 1}
    Note over AS: Subject = "kassensitzung-20260322-tisch-42", Version = 2
    AS->>KJ: INSERT (subject="kassensitzung-20260322-tisch-42", version=2, type="zahlung-kassiert:v1")
    AS->>TSS: UPSERT tisch_session_state (saldo -= Zahlung)

    Note over A,TSS: Admin beendet den Betriebstag (alle Tische müssen Saldo = 0 haben)

    A->>AS: KassensturzDurchfuehren(ist=487.50€)
    AS->>KJ: INSERT (subject="kassensitzung-20260322", version=3, type="kassensturz-durchgefuehrt:v1")

    A->>AS: TagesabschlussErstellen()
    Note over AS: Prüfe Tisch-Saldo-Sperre: Alle Tisch-AKs müssen Saldo = 0 haben
    Note over AS: Prüfung via tisch_session_state WHERE kassensitzung_nr = 1
    AS->>KJ: INSERT (subject="kassensitzung-20260322", version=4, type="tagesabschluss-erstellt:v1")
    AS->>KSP: UPDATE kassensitzung_state SET status=abgeschlossen

    Note over A,TSS: Nächster Veranstaltungstag

    A->>AS: KassensitzungEroeffnen("Sommerfest Tag 2")
    AS->>KJ: INSERT (subject="kassensitzung-20260323", version=1, type="kassensitzung-eroeffnet:v1")

    S->>AS: BestellungAufnehmen(tischId=42, ...)
    AS->>KSP: GetOffeneKassensitzung() → KS{Nr: 2}
    Note over AS: Subject = "kassensitzung-20260323-tisch-42", Version = 1 (frischer Start!)
    AS->>KJ: INSERT (subject="kassensitzung-20260323-tisch-42", version=1, ...)
    AS->>TSS: UPSERT tisch_session_state (neue Projektion für KS 2)
```

Jede Kassensitzung beginnt für jeden Tisch bei Version 1. Keine Altlasten-Streams, keine unbegrenzt wachsenden Projektionen.

---

### 3.12 Paketstruktur

```
backend/domain/
├── kasse/                              (NEU — vereinigte Core Domain)
│   ├── tisch_session.go                (Tisch-AK State + In-Memory Replay)
│   ├── tisch_session_events.go         (Bestellung, Ausgabe, Zahlung, Storno, Auszahlung)
│   ├── kassensitzung.go                (KS State + Projection)
│   ├── kassensitzung_events.go         (Eröffnung, Anfangsbestand, Bewegung, Sturz, Abschluss)
│   ├── kassenbestand.go                (Soll-Berechnung)
│   └── invarianten.go                  (Shared: Positions-Logik, Saldo-Berechnung)
├── table/                              (REDUZIERT — nur noch Stammdaten)
│   └── tisch.go                        (Tisch-Entity mit CRUD-Methoden)
├── product/                            (unverändert)
├── user/                               (unverändert)
└── event/                              (unverändert — Event-Envelope)

backend/repository/
├── kassenjournal_repo/                 (NEU — ersetzt event_repo)
│   └── repo.go                         (WriteEvent mit StreamType-Routing, Projektions-UPSERTs)
├── table_repo/                         (REDUZIERT — nur noch Stammdaten-CRUD)
└── ...

backend/api/
├── kasse/                              (NEU — Kassensitzung-Endpunkte)
│   ├── http/handler.go                 (KS eröffnen, Kassenbewegung, Kassensturz, Z-Bon)
│   └── application/service.go
├── table/                              (ANGEPASST — Tisch-Operationen + Stammdaten)
│   ├── http/handler.go
│   └── application/
│       ├── command.go                  (Bestellen, Bezahlen, Stornieren — nutzt kasse/)
│       └── query.go                    (Tischübersicht, Historie — liest kasse + Stammdaten)
└── ...
```

`domain/table/` schrumpft auf eine Datei für die Stammdaten-Entität. Die gesamte Kassen-Logik zieht in `domain/kasse/`.

---

### 3.13 Auswirkungen im Überblick

#### Was sich ändert

| Bereich              | Ist                                                                   | Neu                                                                                             |
| -------------------- | --------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------- |
| DB-Tabelle           | `events`                                                              | `kassenjournal`                                                                                 |
| DB-Projektion        | `table_state` (PK: tisch_id)                                          | `kassensitzung_state` + `tisch_session_state` (session-scoped, PK: subject)                     |
| Subject-Format       | `tisch:42` (unbegrenzt)                                               | `kassensitzung-{YYYYMMDD}-tisch-{id}` (pro Kassensitzung)                                       |
| Event-Typen          | `tisch.bestellung-aufgenommen:v1`                                     | `bestellung-aufgenommen:v1` (Subject bestimmt den Kontext)                                      |
| Tisch-State          | Synchrone Projektion in `table_state`                                 | Synchrone Projektion in `tisch_session_state` (session-scoped, pro KS)                          |
| Tischübersicht       | `SELECT * FROM table_state`                                           | `SELECT ... FROM tisch_session_state JOIN tische` (einfacher JOIN)                              |
| Tisch-Historie       | `ReadEventsBySubject("tisch:42")`                                     | `ReadEventsBySubject("kassensitzung-20260322-tisch-42")` (kassensitzungsbezogen)                |
| Gesamthistorie Tisch | Alle Events im Stream                                                 | `WHERE subject LIKE 'kassensitzung-%-tisch-42'` (alle Kassensitzungen)                          |
| Relay/Bondruck       | Pollt `events`, parst `tisch:{id}`, `tisch.bestellung-aufgenommen:v1` | Pollt `kassenjournal`, parst `kassensitzung-{YYYYMMDD}-tisch-{id}`, `bestellung-aufgenommen:v1` |
| Reporting            | `WHERE timestamp BETWEEN $1 AND $2`                                   | `WHERE kassensitzung_nr = $1` (fachlich korrekt, keine Mitternacht-Edge-Cases)                  |
| Domain-Paket Tisch   | `domain/table/` (Stammdaten + Events + Projection)                    | `domain/table/` (nur Stammdaten), `domain/kasse/` (Events + Replay)                             |
| Kassenbestand        | Nicht implementiert                                                   | SQL-Query über Kassenjournal + Kassensitzung-Projektion                                         |
| Kassensitzung        | Nicht implementiert (war "Abrechnungskreis")                          | Event-Sourced mit synchroner Projektion                                                         |

#### Was gleich bleibt

- **Event-Envelope** (`event.Event`): ID, UserID, UserName, Type, Subject, Version, Timestamp, Data
- **OCC-Mechanismus**: `UNIQUE(subject, version)`, MaxVersion + 1
- **Fat Events**: Produktdaten eingefroren im Event
- **Invarianten**: Saldo-, Ausgabe-, Bezahl-, Stornierungsinvariante (Logik identisch)
- **Immutabilitäts-Trigger**: Kein UPDATE/DELETE auf dem Journal
- **API-Endpunkte**: Dieselben POST-Endpunkte, dieselben Request/Response-Formate
- **Frontend**: Minimale Änderungen (API bleibt gleich)

---

## 4. Pro und Contra

### Pro

| #   | Argument                                   | Erläuterung                                                                                                                                                                                             |
| --- | ------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| P1  | **Einheitliche Persistenz**                | Ein Kassenjournal, eine Schreibmechanik, ein Replay-Muster. Keine hybride Berechnung (ES + Immutable Records). Senkt kognitive Komplexität.                                                             |
| P2  | **Eliminiert bidirektionale Abhängigkeit** | Kein Upstream/Downstream-Mapping mehr. Kassenbetrieb und Kassenführung sind derselbe Kontext.                                                                                                           |
| P3  | **Compliance-Vorteil**                     | DSFinV-K verlangt eine einheitliche Vorgangs-Sicht. Alle Geschäftsvorfälle in einem Store erleichtert den Export. Kassensitzung = `Z_NR`, Tisch-AK = `ABRECHNUNGSKREIS`.                                |
| P4  | **Korrekte AK-Granularität**               | `ABRECHNUNGSKREIS` ist pro Tisch pro Kassensitzung — wie von der DSFinV-K gefordert, nicht fälschlich global.                                                                                           |
| P5  | **Natürlicher Lebenszyklus**               | `kassensitzung-20260322-tisch-42` hat einen klaren Anfang (erste Bestellung) und ein Ende (Tagesabschluss). Kein unbegrenztes Wachstum.                                                                 |
| P6  | **Konzeptuelle Klarheit**                  | Tisch = Ort (Stammdaten). Tisch-AK = Transaktion (Kasse). Kassensitzung = Betriebstag. Keine Doppelrollen.                                                                                              |
| P7  | **Fachlicher Tabellenname**                | `kassenjournal` statt `events` — selbsterklärend und compliance-konform.                                                                                                                                |
| P8  | **Session-scoped Tisch-Projektion**        | `tisch_session_state` ersetzt die globale `table_state` — session-scoped mit natürlichem Lebenszyklus. Tischübersicht ist ein einfacher JOIN, kein JSONB-Parsing. Write-Pfad bewährt (INSERT + UPSERT). |
| P9  | **Hierarchische Subjects**                 | Prefix-Queries ermöglichen flexible Abfragen auf jedem Granularitätslevel (Kassensitzung, alle Tische, ein Tisch).                                                                                      |
| P10 | **Perfektes Timing**                       | Kassenführung ist noch nicht implementiert. Kein Migrationsaufwand.                                                                                                                                     |

### Contra

| #   | Argument                              | Erläuterung                                                                       | Bewertung                                                                                                                                                                     |
| --- | ------------------------------------- | --------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| C1  | **Breaking Changes**                  | Subject-Format, DB-Tabellennamen, Paketstruktur ändern sich.                      | **Akzeptabel:** Pre-Release, keine produktiven Instanzen. Breaking Changes sind ausdrücklich erwünscht.                                                                       |
| C2  | **Größerer Kasse-Kontext**            | Ein Bounded Context mit mehr Verantwortung. Potenzielle „Big Ball of Mud"-Gefahr. | **Gering:** Zwei klar getrennte Event-Stream-Typen mit eigenen Replay-Funktionen und Invarianten.                                                                             |
| C3  | **Zwei Projektionen statt einer**     | `kassensitzung_state` + `tisch_session_state` statt nur `table_state`.            | **Gering:** Beide sind minimal und session-scoped. Write-Through-Pattern ist bewährt. Session-Scope verhindert unbegrenztes Wachstum.                                         |
| C4  | **Mehr Event-Typen**                  | 5 neue Event-Typen (Kassensitzung).                                               | **Moderat:** Überschaubar. Immutable Records hätten vergleichbare Gesamtkomplexität.                                                                                          |
| C5  | **Projektions-Routing in WriteEvent** | WriteEvent muss je nach `StreamType` die richtige Projektion aktualisieren.       | **Gering:** Expliziter `StreamType`-Parameter statt implizitem Subject-Parsing. Klares Routing-Schema (siehe §3.8).                                                           |
| C6  | **LIKE-Queries statt FK-Joins**       | Subject-basierte Filterung statt FK-basierter Joins für Cross-Stream-Aggregation. | **Akzeptabel:** PostgreSQL-Indizes auf `subject` machen LIKE-Prefix-Queries effizient. Ergänzend steht die `kassensitzung_nr`-Spalte für robuste Aggregationen zur Verfügung. |

---

## 5. Empfehlung

Die argumentative Bilanz fällt eindeutig aus:

- **10 Pro-Argumente**, davon mehrere gewichtig (korrekte AK-Granularität, Compliance, hierarchische Subjects, session-scoped Projektionen)
- **6 Contra-Argumente**, alle mit Bewertung „akzeptabel“ bis „moderat“ — keines ist ein Showstopper

**Empfehlung:**

> **Das Kassenjournal-Modell umsetzen: Eine Tabelle `kassenjournal` für alle finanziellen Geschäftsvorfälle, hierarchische Subjects (`kassensitzung-{YYYYMMDD}` für den Betriebstag, `kassensitzung-{YYYYMMDD}-tisch-{id}` für den Tisch-Abrechnungskreis), zwei synchrone Projektionen (`kassensitzung_state` für die KS-Sperre, `tisch_session_state` für den Tisch-Read-Pfad), explizites `StreamType`-Routing im Repository, und das Domain-Paket `domain/kasse/` als vereinigte Core Domain.**

Begründung:

1. Die Kassenführung ist noch nicht implementiert — **jetzt ist der richtige Zeitpunkt**.
2. Die Trennung von Kassensitzung (global) und Abrechnungskreis (pro Tisch) bildet die DSFinV-K-Realität korrekt ab.
3. Tagesbegrenzte Tisch-Streams bilden die fachliche Realität ab (Vereinsfeste sind Tagesereignisse).
4. Session-scoped `tisch_session_state` gibt dem meistgenutzten Endpunkt (Tischübersicht) seinen schnellen Read-Pfad — ohne das bewährte INSERT + UPSERT Pattern zu verlassen.
5. Hierarchische Subjects ermöglichen flexible Queries auf jedem Granularitätslevel.
6. `kassenjournal` als Name bringt Code und Compliance in Einklang.
7. Explizites `StreamType`-Routing statt implizitem Subject-Parsing hält die Repository-Schicht sauber.

---

## 6. Umsetzungsplan

### Phase A1: Schema-Migration + Rename (alle Tests grün, keine neue Funktionalität)

- [ ] `events` → `kassenjournal` in `01_initial.up.sql` umbenennen (+ Trigger, Indizes, FKs)
- [ ] `table_state` → `tisch_session_state` (session-scoped: PK = `subject`, mit `kassensitzung_nr`)
- [ ] Projektionstabelle `kassensitzung_state` in `01_initial.up.sql` ergänzen
- [ ] Subject-Format `kassensitzung-{YYYYMMDD}-tisch-{id}` implementieren
- [ ] Repository: `kassenjournal_repo/` (ersetzt `event_repo/`) mit explizitem `StreamType`-Routing
- [ ] Tischübersicht als JOIN auf `tisch_session_state` implementieren
- [ ] Reporting-Queries von `timestamp BETWEEN` auf `kassensitzung_nr = $1` migrieren
- [ ] sqlc-Queries anpassen (`make sqlc`)
- [ ] Alle bestehenden Tests anpassen

### Phase A1.5: Relay-Anpassung (eigene Arbeitseinheit)

Die Relay-Anpassung betrifft mehrere Komponenten und ist eine eigenständige Arbeitseinheit:

- [ ] Poll-Query: Tabelle `events` → `kassenjournal` umbenennen
- [ ] Event-Typ-Parsing: `tisch.bestellung-aufgenommen:v1` → `bestellung-aufgenommen:v1`
- [ ] Subject-Parser: `tisch:{id}` → `kassensitzung-{YYYYMMDD}-tisch-{id}` umbauen
- [ ] Tisch-ID-Extraktion: Aus neuem Subject-Format extrahieren (Segment nach `tisch-`)
- [ ] Tischname für Bon: Aus Stammdaten per `tisch_id` laden (da Fat Events den Tischnamen nicht enthalten)
- [ ] Tests für neues Subject-Parsing und Event-Typ-Mapping

### Phase A2: Kassensitzung (neue Domain-Logik)

- [ ] Domain-Modell: `domain/kasse/` mit Kassensitzung-Events, State, Projection
- [ ] Domain-Modell: `domain/kasse/` mit Tisch-Session-Events, In-Memory-Replay
- [ ] Application Service + HTTP Handler für `KassensitzungEroeffnen`, `AnfangsbestandSetzen`
- [ ] **Kassensitzung-Sperre** im Tisch-Application-Service (KS-Projektion abfragen)
- [ ] Tisch-Events von `domain/table/` nach `domain/kasse/` verschieben
- [ ] `domain/table/` auf reine Stammdaten-Entity (CRUD) reduzieren

### Phase B: Kassenbewegungen + Kassenbestand

- [ ] `KassenbewegungGebucht`-Event (Geldtransit, Privatentnahme, Privateinlage)
- [ ] Kassenbestand als SQL-Aggregation über Kassenjournal + Kassensitzung-Projektion
- [ ] Application Service + HTTP Handler für Kassenbewegungen
- [ ] Admin-UI: Kassenbewegungen erfassen, Kassenbestand anzeigen

### Phase C: Kassensturz + Tagesabschluss (Z-Bon)

- [ ] `KassensturzDurchgefuehrt`-Event + `DifferenzSollIstGebucht`-Event (zwei Events in einer TX bei Differenz ≠ 0)
- [ ] `TagesabschlussErstellt`-Event
- [ ] Z-Bon-Logik im Application Service (Aggregation über Kassenjournal)
- [ ] Admin-UI: Kassensturz durchführen, Z-Bon anzeigen
- [ ] DSFinV-K-Export vorbereiten (alle Daten im Kassenjournal)
