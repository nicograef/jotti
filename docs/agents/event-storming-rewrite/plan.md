# Plan: Event-Storming-Simulation neu schreiben

## Ziel

Das bestehende Dokument `docs/event-storming.md` (1032 Zeilen) komplett neu schreiben. Die bisherige Simulation ist überholt — Anforderungen (`docs/anforderungen.md`), Produktbeschreibung (`docs/produktbeschreibung.md`) und Nutzungsbestimmungen (`docs/lizenz-und-nutzung.md`) wurden grundlegend aktualisiert. Die neue Simulation muss alle aktuellen Anforderungen widerspiegeln und die Teilnehmerrunde vereinfachen.

**Grundprinzip: Unvoreingenommene Session.** Die Simulation soll bewusst frei von bestehender Implementierung, existierendem System-Design, sprachlichen Festlegungen und Anforderungsstatus sein. Fachsprache (Ubiquitous Language), Architekturentscheidungen und technische Patterns sollen in der Session selbst emergieren — nicht aus vorhandenen Dokumenten übernommen werden. Die Teilnehmer erkunden die Domäne so, als gäbe es noch kein System.

## Kontext

### Betroffene Dateien

- **Zu ersetzen:** `docs/event-storming.md` — bleibt als Strukturreferenz, wird am Ende durch `docs/event-storming-neu.md` ersetzt
- **Referenzdokumente (Source of Truth):**
  - `docs/anforderungen.md` — alle funktionalen und querschnittlichen Anforderungen mit IDs (ohne Implementierungsstatus)
  - `docs/produktbeschreibung.md` — Positionierung, Personas, Kernfeatures, Abgrenzung
  - `docs/lizenz-und-nutzung.md` — AGPL-3.0 + Non-Commercial, berechtigte Nutzer
- **Bewusst NICHT als Input (Unvoreingenommenheit):**
  - `docs/language.md` — Fachsprache soll in der Session emergieren, nicht vorgegeben werden
  - `docs/design.md`, `docs/cqrs.md` — Architektur- und Pattern-Entscheidungen sollen ergebnisoffen diskutiert werden
  - Bestehender Code / Implementierungsstatus — die Session tut so, als gäbe es noch kein System

### Änderungen gegenüber dem alten Dokument

#### 1. Teilnehmer vereinfachen (14 → 10 Personen)

| Alt                                  | Neu                                                                  | Begründung                                               |
| ------------------------------------ | -------------------------------------------------------------------- | -------------------------------------------------------- |
| FAC (Facilitator Lisa)               | **FAC** (Lisa) — absorbiert auch Scrum-Master-Aufgaben               | SCR-Rolle wegfallen lassen                               |
| SCR (Scrum Master Markus)            | ❌ entfällt                                                          | Aufgaben in FAC integriert                               |
| DEV1 (Backend Anna)                  | **DEV1** (Anna) — Senior Fullstack, Go + React, DDD/Event-Sourcing   | Fullstack statt Backend-only                             |
| DEV2 (Frontend Tim)                  | **DEV2** (Tim) — Senior Fullstack, Go + React, Architekturkenntnisse | Fullstack statt Frontend-only, absorbiert ARC            |
| ARC (Architekt Stefan)               | ❌ entfällt                                                          | Architekturwissen verteilt auf DEV1 + DEV2               |
| POS (POS-Experte Petra)              | ❌ entfällt                                                          | POS-Wissen über DOM und DEV1/DEV2 abgedeckt              |
| VER2 (Vereinsmitglied Sandra)        | ❌ entfällt                                                          | Buchhaltungs-/Export-Perspektive in KAS (Eva) integriert |
| DOM, SRV1, SRV2, ADM, KAS, SRL, VER1 | bleiben unverändert                                                  | Einzigartige Perspektiven                                |

**Neue Teilnehmerliste (10 Personen):**

| Kürzel | Rolle                      | Person | Hintergrund                                             |
| ------ | -------------------------- | ------ | ------------------------------------------------------- |
| FAC    | Facilitator & Moderatorin  | Lisa   | Event-Storming-Erfahrung, hält Timeboxen, moderiert     |
| DEV1   | Senior Fullstack Developer | Anna   | Go + React, DDD, Event-Sourcing-Erfahrung               |
| DEV2   | Senior Fullstack Developer | Tim    | Go + React, Systemarchitektur, CQRS-Erfahrung           |
| DOM    | Domänenexperte             | Rudi   | Langjähriger Vereinsvorstand, 15 Jahre Festorganisation |
| SRV1   | Servicekraft               | Jonas  | Bedient seit 3 Jahren Tische beim Vereinsfest           |
| SRV2   | Servicekraft               | Maria  | Neu dabei, erstes Vereinsfest                           |
| SRL    | Serviceleitung             | Felix  | Senior-Servicekraft, darf stornieren, koordiniert Team  |
| ADM    | Administrator              | Thomas | Software und Hardware beim Verein                       |
| KAS    | Kassenwart                 | Eva    | Finanzen, Tagesabrechnung, Vereinsbuchhaltung, Export   |
| VER1   | Vereinsmitglied (Ausgabe)  | Klaus  | Getränkeausgabe, Küchenmitarbeit                        |

#### 2. Anforderungen aktualisieren

Das alte Dokument referenziert veraltete Anforderungs-Nummern (#23–#40). Das neue Dokument muss die aktuellen IDs verwenden:

**Kassenbetrieb:**

- K-01: Bestellung aufgeben
- K-02: Zahlung registrieren
- K-03: Lieferung bestätigen
- K-04: Stornierung
- K-05: Tischübersicht und Navigation
- K-06: Kassenjournal (Historie)
- K-07: Bezeichnung pro Bestellung
- K-08: Bestellungen umbuchen
- K-09: Rückgeldberechnung
- K-10: Tisch-Schnellsuche
- K-11: Bondruck (inkl. Freibon)
- K-12: Küchendisplay (KDS)
- K-13: Ausgabestationen mit Zubereitungsstatus

**Stammdaten:**

- S-01: Produktverwaltung
- S-02: Tischverwaltung
- S-03: Benutzerverwaltung

**Auth:**

- A-01: Login
- A-02: Passwort setzen
- A-03: Logout

**Querschnitt:**

- Q-01 bis Q-08

**Reporting:**

- R-01: Tagesabrechnung
- R-02: Datenexport
- R-03: Abrechnung pro Tisch
- R-04: Abrechnung pro Servicekraft
- R-05: Produktumsatz-Reporting
- R-06: Tagesabschluss

#### 3. Neue Features in die Simulation einarbeiten

Folgende Features sind im alten Dokument nur als Hotspots oder gar nicht erwähnt, müssen jetzt vollständig eingearbeitet werden:

- **Bondruck (K-11)**: Automatischer Bon-Druck nach Bestellung, getrennt nach Kategorie (Küche/Getränke), Freibon mit freier Preiseingabe
- **Küchendisplay / KDS (K-12)**: Echtzeit-Anzeige offener Bestellungen nach Kategorie, gruppiert nach Tisch
- **Ausgabestationen mit Zubereitungsstatus (K-13)**: Positionen als „in Zubereitung" / „fertig" markieren
- **Reporting (R-01 bis R-06)**: Vollständige Reporting-Sektion mit 6 Anforderungen
- **Abgrenzung aktualisiert**: Gast-Benachrichtigung (🚫) ist neu in der Liste

#### 4. Lizenzmodell korrekt referenzieren

Das alte Dokument erwähnt die Lizenz nicht. Das neue Dokument soll bei Bedarf korrekt auf **AGPL-3.0-or-later mit Zusatzbedingungen (Source-Available, Non-Commercial)** referenzieren — z.B. wenn die Teilnehmer über Kosten/Nutzung sprechen.

### Bestehende Patterns und Konventionen

Die Struktur des alten Dokuments ist grundsätzlich gut und kann als Vorlage dienen:

1. **Setup & Teilnehmer** — Rahmenbedingungen, Notation, Teilnehmerliste
2. **Phase 1 — Big Picture** — Chaotische Exploration (alle kleben Events)
3. **Phase 2 — Clustering und Pivot Points** — Timeline, Cluster, Pivot Points
4. **Phase 3 — Process Modelling** — Commands, Akteure, Policies
5. **Phase 4 — Software Design** — Aggregate, Read Models
6. **Bounded Contexts und Domain Map** — Context Map, Sub-Domain-Klassifikation
7. **Hotspots und offene Fragen** — Ungelöste Designfragen
8. **Ergebnisse und nächste Schritte** — Zusammenfassung, Priorisierung
9. **Anhang A — Vollständige Event-Liste**
10. **Anhang B — Stickies-Legende**

## Implementierungsschritte

### Schritt 1: Datei vorbereiten

Neue Datei `docs/event-storming-neu.md` anlegen. Die bestehende `docs/event-storming.md` bleibt während des Schreibens als Strukturreferenz erhalten (nicht als inhaltliche Vorlage — nur für Stil und Aufbau). Nach Fertigstellung und Validierung: alte Datei löschen und `event-storming-neu.md` → `event-storming.md` umbenennen.

### Schritt 2: Header und Einleitung schreiben

- Gleiche Einleitung wie bisher (Event Storming nach Brandolini)
- Inhaltsverzeichnis aktualisieren (gleiche 8 Abschnitte + 2 Anhänge)

### Schritt 3: Setup & Teilnehmer (Abschnitt 1)

- **Datum:** Samstag, 11. März 2026 (neues Datum, da neue Simulation)
- **Ort:** Vereinsheim Sportverein Grüntal (kann gleich bleiben)
- **Dauer:** ca. 5 Stunden (kürzer, da weniger Teilnehmer = weniger Diskussion)
- **Notation:** Gleich wie bisher (6 Farben)
- **Teilnehmerliste:** 10 Personen (siehe oben)
  - SCR entfällt (FAC übernimmt Timeboxen)
  - DEV1 und DEV2 sind beide Senior Fullstack mit Architekturkenntnissen
  - POS entfällt (DOM hat genug Erfahrung, DEVs kennen kommerzielle Systeme)
  - VER2 entfällt (KAS deckt Buchhaltungsperspektive ab)

### Schritt 4: Phase 1 — Big Picture (Abschnitt 2)

- Jeder Teilnehmer klebt Events an die Wand (gleicher Stil: Dialogform)
- **Keine SCR-Dialoge** — FAC übernimmt Timer-Ansagen
- **Keine ARC/POS/VER2-Dialoge** — deren Perspektiven fließen über DEV1/DEV2/DOM/KAS ein
- Events müssen **alle Anforderungen** aus `docs/anforderungen.md` widerspiegeln — ohne Unterscheidung nach Implementierungsstatus:
  - Kassenbetrieb (K-01 bis K-13)
  - Stammdaten (S-01 bis S-03) und Auth (A-01 bis A-03)
  - Reporting (R-01 bis R-06) und Querschnitt (Q-01 bis Q-08)
  - Bewusste Abgrenzung (Kartenzahlung, TSE, Reservierungen etc.)
- **Neue Events die im alten Dokument fehlen:**
  - 🟠 Küchenbestellung auf Display angezeigt (K-12)
  - 🟠 Position als „in Zubereitung" markiert (K-13)
  - 🟠 Position als „fertig" markiert (K-13)
  - 🟠 Produktumsatz ausgewertet (R-05)
  - 🟠 Abrechnung pro Tisch erstellt (R-03)
  - 🟠 Abrechnung pro Servicekraft erstellt (R-04)
- **Technische vs. fachliche Events bewusst trennen** — Infrastruktur-Aspekte (Auth, Sessions) nicht mit Domänen-Events vermischen; welche technischen Events nötig sind, soll in der Session diskutiert werden

### Schritt 5: Phase 2 — Clustering und Pivot Points (Abschnitt 3)

- **Timeline:** Gleiche 3 Phasen (Vorbereitung → Betrieb → Abschluss)
- **Pivot Points:** Gleich (Veranstaltung eröffnet, Bestellung aufgegeben, Tischkonto ausgeglichen, Tagesabschluss)
- **Cluster aktualisieren:**
  - Cluster A (Stammdaten) — gleich
  - Cluster B (Kassenbetrieb) — gleich + K-07 (Bezeichnung), K-09 (Rückgeld), K-10 (Schnellsuche)
  - Cluster C (Stornierung & Umbuchung) — gleich
  - Cluster D (Ausgabestationen & Bons) — ausgebaut mit K-11, K-12, K-13
  - Cluster E (Abrechnung/Reporting) — ausgebaut mit R-01 bis R-06

### Schritt 6: Phase 3 — Process Modelling (Abschnitt 4)

Detaillierte Command-Event-Policy-Flows für jeden Hauptprozess. Dialogform mit den **neuen** Teilnehmern.

**4.1 Kassenbetrieb — Bestellung bis Zahlung:**

- Bestellung aufnehmen (K-01) — wie bisher, aber DEV2 statt ARC für Architektur-Kommentare
- Lieferung bestätigen (K-03) — wie bisher
- Zahlung registrieren (K-02) — wie bisher, Teilzahlung hervorheben
- Stornierung (K-04) — wie bisher, SRL + Policy-Diskussion

**4.2 Stammdaten-Verwaltung:**

- Produkte (S-01), Tische (S-02), Benutzer (S-03) — gleiche Flows, ohne ARC/POS-Dialoge

**4.3 Abrechnung und Reporting (NEU — ausbauen):**

- Umsatzübersicht (R-01)
- Abrechnung pro Tisch (R-03)
- Abrechnung pro Servicekraft (R-04)
- Produktumsatz-Reporting (R-05)
- Datenexport (R-02)
- Tagesabschluss (R-06) — mit Diskussion über offene Tische

**4.4 Bondruck und Ausgabestationen (NEU — ausbauen):**

- Bon-Druck (K-11) — automatisch nach Bestellung, getrennt nach Kategorie
- Freibon (Teil von K-11) — freie Preiseingabe
- KDS (K-12) — Echtzeit-Anzeige für Küche/Getränkeausgabe
- Zubereitungsstatus (K-13) — Positionen als „in Zubereitung" / „fertig"
- VER1 (Klaus) bringt hier die Ausgabestation-Perspektive ein

### Schritt 7: Phase 4 — Software Design (Abschnitt 5)

Die Teilnehmer identifizieren Aggregate, Event-Typen und Read Models ergebnisoffen — basierend auf den Events und Commands aus den vorherigen Phasen, nicht auf vorgegebenen Strukturen oder existierenden Benennungen.

**5.1 Aggregate identifizieren**

- Welche Aggregate ergeben sich aus den Clustern?
- Welche Invarianten schützt jedes Aggregat?
- Welche Events gehören zu welchem Aggregat?
- Diskussion: Event-Sourcing vs. CRUD — was eignet sich wofür?

**5.2 Event-Typen benennen**

- Die Teilnehmer benennen Event-Typen selbst — keine vorgegebenen Namen aus bestehenden Dokumenten
- Konventionen für Event-Benennung gemeinsam festlegen (Sprache, Format, Versionierung)
- Fachliche vs. technische Events unterscheiden

**5.3 Read Models / Projektionen**

- Welche Ansichten brauchen die verschiedenen Akteure?
- Die Teilnehmer leiten Read Models aus den Anforderungen ab:
  - Tischübersicht, Tischdetails, Kassenjournal
  - Reporting-Ansichten (R-01 bis R-06)
  - Küchen-/Ausgabe-Ansichten (K-12, K-13)
- Diskussion: Wie werden Read Models aus Events aufgebaut?

### Schritt 8: Bounded Contexts und Domain Map (Abschnitt 6)

Die Teilnehmer identifizieren Bounded Contexts ergebnisoffen aus den Clustern und Aggregaten:

- Welche fachlichen Bereiche haben eigene Sprache und eigene Regeln?
- Welche Contexts sind Core, Supporting oder Generic?
- Wie kommunizieren die Contexts miteinander? (Context Map)
- ASCII-Diagramm gemeinsam erarbeiten
- Erwartete Kandidaten aus den Anforderungen: Kassenbetrieb, Stammdaten, Auth, Ausgabestationen/Bons, Reporting — aber die Teilnehmer sollen eigene Schnitte finden

### Schritt 9: Hotspots und offene Fragen (Abschnitt 7)

Hotspots an aktuelle Anforderungen anpassen. Alte Anforderungs-Nummern (#23–#40) durch neue IDs ersetzen.

**Beizubehaltende Hotspots (mit neuen IDs):**

- Tischumbuchung → K-08 (Nice-to-have)
- Freibon → Teil von K-11 (Should-have)
- Bon-Druck / Drucker-Integration → K-11 (Should-have)
- Offline-Fähigkeit → Q-05 (Nice-to-have)
- Tagesabschluss mit offenen Tischen → R-06 (Nice-to-have)

**Zu entfernender Hotspot:**

- Rückgeldberechnung — nicht wirklich ein Hotspot, einfach (reiner Anzeigeaspekt), kann bei Phase 3 kurz erwähnt werden → K-09

**Neue/aktualisierte Hotspots:**

- KDS-Architektur (K-12) — Wie kommen Bestelldaten in Echtzeit zur Küche/Ausgabe?
- Zubereitungsstatus (K-13) — Wie wird der Status modelliert? Eigene Events oder UI-State?
- Reporting-Aggregation (R-01 bis R-05) — Wie werden Auswertungen aus den Rohdaten berechnet?

### Schritt 10: Ergebnisse und nächste Schritte (Abschnitt 8)

- **Gemeinsames Verständnis:** 5 Kernaussagen zusammenfassen
- **Priorisierte Erkenntnisse:** Tabelle mit Anforderungs-IDs, sortiert nach Priorität
- **Ubiquitous Language:** Vollständig aus der Session heraus entwickelte Fachbegriffe dokumentieren — als eigenständiges Ergebnis, nicht als Ergänzung bestehender Definitionen
- **Feedback der Teilnehmer:** Neue Teilnehmer, keine Zitate von SCR/ARC/POS/VER2

### Schritt 11: Anhänge

- **Anhang A — Vollständige Event-Liste:** Aktualisiert mit neuen IDs und neuen Events
- **Anhang B — Stickies-Legende:** Gleich wie bisher

### Schritt 12: Validierung

- Prüfen, dass alle IDs aus `docs/anforderungen.md` mindestens einmal referenziert werden
- Prüfen, dass keine alten Anforderungs-Nummern (#23, #24, #25 etc.) überbleiben
- Prüfen, dass keine entfernten Teilnehmer (SCR, ARC, POS, VER2) noch Dialogzeilen haben
- Prüfen, dass **kein Bezug** auf `docs/language.md`, `docs/design.md` oder `docs/cqrs.md` genommen wird
- Prüfen, dass **kein Implementierungsstatus** (✅/🔲) erwähnt wird
- Prüfen, dass Event-Typ-Namen, Aggregate und Bounded Contexts aus der Session selbst hervorgehen
- Prüfen, dass die bewusste Abgrenzung (§7 anforderungen.md) korrekt reflektiert ist

## Offene Fragen / Risiken

1. **Dokumentlänge:** Das alte Dokument hat 1032 Zeilen. Mit weniger Teilnehmern und strafferer Simulation könnte das neue Dokument kürzer ausfallen (~800–900 Zeilen). Das ist akzeptabel — weniger Redundanz.

2. **Datum der neuen Simulation:** Vorschlag: Samstag, 11. März 2026. Alternativ ein anderes fiktives Datum.

3. **KDS und Bondruck als Bounded Context:** Im alten Dokument „Ausgabestationen" als zukünftig/extern markiert. Da KDS (K-12) und Bondruck (K-11) wichtige Anforderungen sind, sollten sie im neuen Dokument als konkreter Bestandteil behandelt werden — nicht mehr rein hypothetisch.

4. **Reporting-Aggregation:** R-01 bis R-05 werfen Design-Fragen auf (wie werden Auswertungen aggregiert?). Diese gehören als Hotspot in die Session — die Teilnehmer sollen ergebnisoffen diskutieren, ohne vorgegebene Patterns.

## Referenzen

- `docs/event-storming.md` — bestehendes Dokument (zu ersetzen)
- `docs/anforderungen.md` — aktuelle Anforderungen mit IDs
- `docs/produktbeschreibung.md` — Positionierung, Personas, Kernfeatures, Abgrenzung
- `docs/lizenz-und-nutzung.md` — Lizenzmodell (AGPL-3.0 + Non-Commercial)

**Bewusst NICHT als Input/Referenz:**

- `docs/language.md` — Fachsprache soll in der Session emergieren
- `docs/design.md` — Architekturentscheidungen sollen ergebnisoffen sein
- `docs/cqrs.md` — Technische Patterns sollen nicht vorweggenommen werden
- `AGENTS.md` — Projektregeln sind für die Implementierung relevant, nicht für die Event-Storming-Simulation
