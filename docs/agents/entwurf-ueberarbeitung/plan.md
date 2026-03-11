# Plan: Überarbeitung des Software-Entwurfs (entwurf.md)

## Ziel

Den high-level Software-Entwurf (`docs/design/entwurf.md`) vollständig neu schreiben — unvoreingenommen, basierend ausschließlich auf den Ergebnissen des Event Stormings (`docs/design/event-storming.md`), den Anforderungen (`docs/anforderungen.md`) und der Produktbeschreibung (`docs/produktbeschreibung.md`). Der aktuelle Implementierungsstand wird dabei bewusst ignoriert; der Entwurf beschreibt das Zielbild.

## Kontext

### Quelldokumente (Input)

| Dokument            | Pfad                            | Relevanz                                                                                                         |
| ------------------- | ------------------------------- | ---------------------------------------------------------------------------------------------------------------- |
| Event Storming      | `docs/design/event-storming.md` | Primärquelle: Aggregate, Events, Read Models, Bounded Contexts, Hotspots, Policies, Ubiquitous Language          |
| Anforderungen       | `docs/anforderungen.md`         | Vollständiger Anforderungskatalog mit Akzeptanzkriterien (K-01…K-13, S-01…S-03, A-01…A-03, Q-01…Q-08, R-01…R-06) |
| Produktbeschreibung | `docs/produktbeschreibung.md`   | Positionierung, Zielgruppe, Abgrenzung, USPs, Lizenzmodell                                                       |

### Zieldokument (Output)

- **Datei:** `docs/design/entwurf.md`
- **Zweck:** High-level Software-Entwurf als zentrale Architektur-Referenz für die Entwicklung.
- **Prinzip:** Rein fachlich/architektonisch motiviert — keine Implementierungsdetails, keine Tech-Stack-Festlegung auf Bibliotheksebene.

### Analyse des bestehenden Entwurfs

Der aktuelle Entwurf (`docs/design/entwurf.md`, ~950 Zeilen, 11 Kapitel) hat folgende Struktur:

1. Systemvision und Designziele
2. Bounded Contexts und Domain Map
3. Domänenmodell (Aggregate, Events, Read Models, Ubiquitous Language)
4. Persistenzstrategie (Event Store, CRUD-Tabellen, Snapshots, Replay)
5. Schichtenarchitektur (Backend-Schichten, Verzeichnisstruktur, Request-Lebenszyklus)
6. API-Entwurf (POST-only, Endpunkt-Tabellen, Fehlerbehandlung)
7. Frontend-Architektur (Technologie, Guards, Seitenstruktur, UI-Patterns, Backend-Kommunikation)
8. Sicherheitsarchitektur (Authentifizierung, Autorisierung, weitere Maßnahmen)
9. Infrastruktur und Deployment (Docker Compose, Komponenten, Deployment)
10. Querschnittskonzepte (Validierung, Geldbeträge, Mehrbenutzerfähigkeit, Datenintegrität, Mobile-first, Logging)
11. Priorisierung und Ausbaustufen (Stufe 1–3 mit Reihenfolge)
12. Anhang: Offene Entwurfsfragen (H1–H8 aus Event-Storming-Hotspots)

### Identifizierte Schwächen und Verbesserungspotenzial

#### 1. Vermischung von High-Level-Entwurf und Implementierungsdetails

Der aktuelle Entwurf vermischt architektonische Entscheidungen mit konkreten Implementierungsdetails:

- **Kapitel 4.2/4.3**: Enthält vollständige SQL-DDL-Statements (`CREATE TABLE events (…)`). Das gehört in eine Migration oder Database-Instruktion, nicht in einen Architektur-Entwurf.
- **Kapitel 5.2**: Listet die konkrete Verzeichnisstruktur (`backend/domain/table/`, `backend/api/auth/http/`) — das ist Implementierung, nicht Entwurf.
- **Kapitel 6.2**: Zählt jeden einzelnen API-Endpunkt mit URL auf (`POST /admin/produkt-anlegen`). Ein High-Level-Entwurf sollte API-Design-Prinzipien und die Struktur nach Bounded Contexts beschreiben, nicht jeden Endpunkt.
- **Kapitel 7**: Enthält konkrete Technologienamen (React 19, Vite 7, TypeScript 5.9, shadcn/ui, Zod 4) — das gehört in AGENTS.md oder eine Technik-Referenz, nicht in den fachlichen Entwurf.

**Empfehlung:** Der Entwurf sollte _was_ und _warum_ beantworten, nicht _wie genau_. SQL-Schemas, Verzeichnisbäume und vollständige Endpunkt-Listen gehören in separate, implementierungsnahe Dokumente.

#### 2. FreibonAusgestellt-Event fehlt im Domänenmodell

Das Event Storming identifiziert `FreibonAusgestellt` als eigenes Domain Event des Tisch-Aggregats (Phase 4, Abschnitt 4.4, Abschnitt 5.2). Im aktuellen Entwurf (Kapitel 3.2) werden nur vier Event-Typen definiert: `BestellungAufgegeben`, `ProdukteGeliefert`, `ZahlungRegistriert`, `ProdukteStorniert` + `SnapshotErstellt`. Der `FreibonAusgestellt`-Event fehlt.

**Empfehlung:** `FreibonAusgestellt` als fünften fachlichen Event-Typ aufnehmen, mit Struktur (Tisch-ID, Bezeichnung, Preis in Cent, opt. Kommentar, Benutzer, Zeitstempel).

#### 3. Unvollständiges Tisch-Aggregat-Zustandsmodell

Das aktuelle Zustandsmodell des Tisch-Aggregats (Kapitel 3.1) bildet den Zustand als flache Positionsliste ab. Es fehlen:

- **Bestellungs-Gruppierung:** Positionen gehören zu einer Bestellung (über `bestellung_id`), was im Zustand nicht prominent modelliert ist. Das Event Storming betont die Bestellung als Konzept (Bezeichnung, Kommentar, Bestellzeitpunkt).
- **Bezeichnung pro Bestellung (K-07):** Der optionale Gruppenname fehlt im Zustandsmodell.
- **Kommentar pro Bestellung:** Wird im Event erwähnt, aber nicht im Zustand reflektiert.

**Empfehlung:** Den Zustand zweistufig modellieren: `Tisch → Bestellungen → Positionen`. Jede Bestellung hat eine ID, optionale Bezeichnung, Kommentar und einen Zeitstempel.

#### 4. Ausgabe-Context inhaltlich zu dünn

Der bestehende Entwurf erwähnt den Ausgabe-Context in der Context Map (Kapitel 2), aber beschreibt ihn nicht weiter. Es fehlen:

- Konkretisierung des KDS-Read-Models
- Architekturentscheidung zum Echtzeit-Mechanismus (Polling vs. SSE vs. WebSockets)
- Modellierung des Zubereitungsstatus (eigenes Aggregat oder Teil des Tisch-Aggregats?)
- Bon-Routing-Logik (Policy: Bestellung → Bons nach Kategorie)

**Empfehlung:** Dem Ausgabe-Context ein eigenes Unterkapitel widmen, das die offenen Architekturentscheidungen klar als solche benennt (Verweis auf Hotspots H3, H4, H5 aus dem Event Storming).

#### 5. Abrechnung-Context ohne Detailbeschreibung

Der Abrechnung-Context wird in der Context Map erwähnt und die Read Models in Kapitel 3.3 tabellarisch aufgelistet, aber es fehlt:

- Beschreibung der Aggregationsstrategie (on-the-fly vs. vorberechnete Projektionen — Hotspot H8)
- Tagesabschluss-Prozess (Hotspot H7: offene Tische, System-Reset, Veranstaltungskonzept)
- Datenexport-Architektur (Welche Daten? Welches Format? Nur CSV?)

**Empfehlung:** Eigenes Unterkapitel für den Abrechnung-Context mit Read-Model-Beschreibungen und Verweis auf offene Architekturentscheidungen.

#### 6. Fehlende Darstellung des Veranstaltungs-Konzepts

Das Event Storming wirft in Hotspot H7 die Frage auf, ob es ein übergreifendes Konzept „Veranstaltung" geben sollte — als logischer Rahmen, der die Event-Streams einer Veranstaltung kapselt und den Tagesabschluss/Reset ermöglicht, ohne Events zu löschen. Der aktuelle Entwurf ignoriert diesen Aspekt vollständig.

**Empfehlung:** Das Konzept als offene Architekturentscheidung im Entwurf benennen und die Auswirkungen auf Event Store, Snapshots und Reporting skizzieren.

#### 7. Inkonsistente Behandlung der Stornierungsinvariante

Im aktuellen Entwurf (Kapitel 3.1) steht: „Nur bestellte, nicht bereits stornierte Positionen können storniert werden." Das Event Storming (Phase 4.1) diskutiert aber explizit den Fall, dass auch _bereits bezahlte_ Positionen stornierbar sein sollen (Reklamation nach Zahlung). Der Entwurf bildet diesen Fall nicht ab.

**Empfehlung:** Die Invariante differenzierter formulieren: Stornierbare Positionen sind bestellte, nicht-stornierte Positionen — unabhängig vom Bezahl-/Lieferstatus. Bei Stornierung bezahlter Positionen entsteht ein negativer Saldo-Anteil.

#### 8. SnapshotErstellt als Event-Typ fragwürdig

Der aktuelle Entwurf modelliert `SnapshotErstellt` als Event im Event Stream. Das Event Storming (Phase 5.3) betont, dass Snapshots rein technische Optimierungen ohne fachliche Bedeutung sind. Ein Snapshot als Event im selben Stream zu speichern vermischt fachliche und technische Concerns.

**Empfehlung:** Snapshots aus dem Event Stream heraushalten und als separate technische Speicherung modellieren (z.B. eigene Tabelle oder markierter Datensatz). Dies wird als Entwurfsentscheidung dokumentiert.

#### 9. Ubiquitous Language unvollständig

Die aktuelle Tabelle (Kapitel 3.4) enthält 14 Begriffe. Dem Event Storming (Abschnitt 8.3) zufolge gibt es deutlich mehr etablierte Begriffe, insbesondere aus den Kontexten Ausgabe und Abrechnung:

- Bon, Küchendisplay (KDS), Zubereitungsstatus, Ausgabestation
- Tagesabrechnung, Umsatz, Stornoquote, Tagesabschluss, Export
- Bezeichnung (Bestellung), Freibon
- Event-Sourcing, Fat Event, ACL, Append-only, BYOD

**Empfehlung:** Die Ubiquitous Language vollständig aus dem Event Storming übernehmen und nach Bounded Context gruppieren.

#### 10. Priorisierung zu implementierungsnah

Kapitel 11 enthält eine detaillierte Reihenfolge innerhalb der Stufen (`Auth → Stammdaten → Kassenbetrieb → Querschnitt`). Das ist eher ein Sprint-Plan als ein Architektur-Entwurf.

**Empfehlung:** Priorisierung auf Stufen-Ebene belassen (Must-have / Should-have / Nice-to-have mit Anforderungs-IDs). Die Implementierungsreihenfolge gehört in ein separates Planungsdokument.

## Implementierungsschritte

Der neue Entwurf wird als vollständiger Ersatz von `docs/design/entwurf.md` geschrieben. Die Gliederung orientiert sich am Event Storming und folgt dem Prinzip „Was und Warum, nicht Wie genau".

### Schritt 1: Neues Inhaltsverzeichnis und Einleitung

Die Kapitelstruktur des neuen Entwurfs:

```markdown
# Software-Entwurf — jotti

1. Systemvision und Designziele
   1.1 Systemvision
   1.2 Designziele
   1.3 Bewusste Abgrenzung

2. Bounded Contexts und Domain Map
   2.1 Kontextübersicht
   2.2 Klassifikation (Core / Supporting / Generic)
   2.3 Context Map (Diagramm)
   2.4 Beziehungen zwischen Kontexten

3. Kassenbetrieb (Core Domain)
   3.1 Tisch-Aggregat
   3.2 Invarianten
   3.3 Domain Events
   3.4 Zustandsberechnung (Event Replay)
   3.5 Snapshot-Strategie
   3.6 Policies

4. Stammdaten (Supporting Sub-Domain)
   4.1 Produkt-Aggregat
   4.2 Tisch-Stammdaten-Aggregat
   4.3 Benutzer-Aggregat
   4.4 Persistenzstrategie (CRUD mit Soft-Delete)

5. Ausgabe (Supporting Sub-Domain)
   5.1 Bondruck
   5.2 Küchendisplay (KDS)
   5.3 Zubereitungsstatus
   5.4 Offene Architekturentscheidungen

6. Abrechnung (Supporting Sub-Domain)
   6.1 Read Models und Projektionen
   6.2 Aggregationsstrategie
   6.3 Tagesabschluss
   6.4 Datenexport
   6.5 Offene Architekturentscheidungen

7. Auth (Generic Sub-Domain)
   7.1 Authentifizierung
   7.2 Autorisierung (Rollenmodell)
   7.3 Sicheres Onboarding

8. Read Models
   8.1 Service-Ansichten
   8.2 Admin-Ansichten (Reporting)
   8.3 Ausgabe-Ansichten

9. Persistenzstrategie
   9.1 Zwei Strategien, eine Datenbank
   9.2 Event Store (Prinzipien, keine DDL)
   9.3 Stammdaten (Prinzipien, keine DDL)
   9.4 Optimistic Concurrency Control

10. Architekturprinzipien
    10.1 Schichtenarchitektur (Übersicht, kein Verzeichnisbaum)
    10.2 API-Design-Prinzipien (POST-only, kein Endpunkt-Katalog)
    10.3 Frontend-Architektur (Prinzipien, keine Tech-Details)
    10.4 Validierung (doppelte Absicherung)
    10.5 Geldbeträge (Cent-Integer)
    10.6 Mehrbenutzerfähigkeit
    10.7 Mobile-first
    10.8 Sicherheit

11. Infrastruktur und Deployment
    11.1 Architekturüberblick
    11.2 Deployment-Modell

12. Ubiquitous Language

13. Priorisierung und Ausbaustufen
    13.1 Stufe 1 — Must-have (MVP)
    13.2 Stufe 2 — Should-have
    13.3 Stufe 3 — Nice-to-have

14. Offene Entwurfsfragen
```

**Wesentliche Strukturänderungen vs. aktueller Entwurf:**

- Kassenbetrieb, Stammdaten, Ausgabe, Abrechnung und Auth sind eigene Kapitel (statt in Kapitel 3 „Domänenmodell" zusammengefasst)
- Read Models bekommen ein eigenes zentrales Kapitel
- Architekturprinzipien (Schichten, API, Frontend, Security) werden kompakt zusammengefasst — ohne Endpunkt-Listen oder Verzeichnisbäume
- Ubiquitous Language bekommt ein eigenes Kapitel statt einer kleinen Tabelle
- SQL-DDL und Verzeichnisbäume entfallen

### Schritt 2: Kapitel 1 — Systemvision und Designziele

- **Quelle:** Produktbeschreibung (Abschnitt 1, 2, 7), Event Storming (Abschnitt 2.7)
- **Inhalt:** Übernahme von Kapitel 1 des aktuellen Entwurfs, da dieses bereits sauber und unvoreingenommen formuliert ist.
- **Änderungen:** Keine wesentlichen Änderungen nötig — Systemvision, 7 Designziele, 9-Punkte-Abgrenzungsliste bleiben.

### Schritt 3: Kapitel 2 — Bounded Contexts und Domain Map

- **Quelle:** Event Storming (Abschnitt 6: Bounded Contexts und Domain Map)
- **Inhalt:** Fünf Contexts (Kassenbetrieb, Stammdaten, Ausgabe, Abrechnung, Auth) mit Klassifikation und Context Map.
- **Änderungen gegenüber aktuellem Entwurf:**
  - Context-Map-Diagramm übernehmen aus Event Storming (Abschnitt 6.3) — identisch zum aktuellen, aber als verifizierte Quelle.
  - Beziehungstabelle erweitern: Auth → alle vier Contexts (nicht nur „alle anderen"), je eine Zeile.
  - ACL-Erklärung beibehalten (Fat Events schützen Kassenbetrieb vor Stammdaten-Änderungen).

### Schritt 4: Kapitel 3 — Kassenbetrieb (Core Domain)

- **Quelle:** Event Storming (Phase 4.1, 5.1, 5.2), Anforderungen (K-01…K-06, K-07, K-08, K-09)
- **Inhalt:**
  - **Tisch-Aggregat**: Zustand mit zweistufiger Modellierung (Tisch → Bestellungen → Positionen). Tisch-ID, Saldo, event_version. Jede Bestellung hat: bestellung_id, bezeichnung?, kommentar?, zeitstempel, positionen[]. Jede Position hat: position_id, variante_id, produkt_name, variante_name, kategorie, einzelpreis, menge, geliefert, bezahlt, storniert.
  - **Invarianten** (aus Event Storming Phase 5.1):
    - Saldo = Σ Bestellungen − Σ Zahlungen − Σ Stornierungen (in Cent)
    - Nur bestellte, nicht-stornierte Positionen können geliefert werden
    - Nur bestellte, nicht-stornierte, nicht-bezahlte Positionen können bezahlt werden
    - Nur bestellte, nicht-stornierte Positionen können storniert werden (unabhängig vom Bezahl-/Lieferstatus — korrigiert gegenüber aktuellem Entwurf)
    - Stornierung nur durch Rollen `senior_service` oder `admin`
    - Jede Operation erfordert mindestens eine Position
  - **Domain Events** (5 fachliche + 1 technischer):
    1. `BestellungAufgegeben` — mit Fat-Event-Daten (Produktname, Variantenname, Kategorie, Einzelpreis), opt. Bezeichnung (K-07), opt. Kommentar
    2. `ProdukteGeliefert` — Referenz auf Positionen, opt. Kommentar
    3. `ZahlungRegistriert` — Referenz auf bezahlte Positionen, Betrag in Cent, opt. Kommentar
    4. `ProdukteStorniert` — Referenz auf stornierte Positionen, Stornobetrag in Cent, opt. Kommentar
    5. `FreibonAusgestellt` — freie Bezeichnung, Preis in Cent, opt. Kommentar (**NEU** gegenüber aktuellem Entwurf)
    - Alle Events: event_id, tisch_id, benutzer_id, benutzer_name, zeitstempel, version
  - **Zustandsberechnung**: Event-Replay-Algorithmus (Snapshot → Events → Apply)
  - **Snapshot-Strategie**: Snapshots als _separate_ technische Speicherung, _nicht_ als Event im Stream (**GEÄNDERT** gegenüber aktuellem Entwurf, der SnapshotErstellt als Event modelliert)
  - **Policies**:
    - Policy: Stornierung nur durch `senior_service` / `admin` (aus K-04)
    - Policy: Automatischer Bon-Druck nach Bestellung, getrennt nach Kategorie (aus K-11)
    - Hinweis auf Umbuchung (K-08) als Cross-Aggregat-Transaktion → Verweis auf offene Fragen

### Schritt 5: Kapitel 4 — Stammdaten (Supporting Sub-Domain)

- **Quelle:** Event Storming (Phase 4.2, 5.1), Anforderungen (S-01…S-03)
- **Inhalt:**
  - **Produkt-Aggregat**: id, name, kategorie (food/beverage/other), status (active/deleted), varianten[] (id, name, preis in Cent, status active/inactive/deleted). Invarianten: Name nicht leer, Kategorie gültig, Variante braucht Name und Preis > 0, Soft-Delete.
  - **Tisch-Stammdaten-Aggregat**: id, name, status (active/inactive/deleted). Hinweis auf die strikte Trennung von Tisch-Stammdaten und Tisch-Aggregat im Kassenbetrieb.
  - **Benutzer-Aggregat**: id, name, benutzername (unique), passwort_hash, rolle (admin/senior_service/service), muss_passwort_setzen, status. Invarianten: Benutzername eindeutig, Rolle gültig, Soft-Delete.
  - **Persistenz**: CRUD mit Soft-Delete. Kein Event-Sourcing nötig, weil historische Stammdaten-Zustände irrelevant sind (Fat Events im Kassenbetrieb decken das ab).
- **Keine SQL-DDL** — nur das konzeptionelle Modell.

### Schritt 6: Kapitel 5 — Ausgabe (Supporting Sub-Domain)

- **Quelle:** Event Storming (Phase 4.4, 5.3), Anforderungen (K-11, K-12, K-13)
- **Inhalt:**
  - **Bondruck (K-11)**: Policy beschreiben (BestellungAufgegeben → automatischer Bon-Druck nach Kategorie). Bon-Inhalt: Tisch, Servicekraft, Positionen, Mengen, Zeitstempel, Kommentar. Freibon als Sonderfall. Drucker-Konfiguration als Admin-Aufgabe. Druckerfehler dürfen den Bestellvorgang nie blockieren (Fire-and-Forget + Retry). Nachdruck einzelner Positionen.
  - **KDS (K-12)**: Read Model für Echtzeit-Anzeige offener Positionen nach Kategorie, gruppiert nach Tisch. Offene Architekturentscheidung: Polling vs. SSE vs. WebSockets → Hotspot H4.
  - **Zubereitungsstatus (K-13)**: Beschreibung des Workflows (offen → in Zubereitung → fertig). Offene Architekturentscheidung: Domain Events im Tisch-Aggregat vs. eigenes Aggregat vs. transienter State → Hotspot H5.
  - **Offene Architekturentscheidungen**: Verweis auf Hotspots H3 (Drucker-Integration), H4 (KDS-Echtzeit), H5 (Zubereitungsstatus-Modellierung) — jeweils mit Pro/Contra-Skizze aus dem Event Storming.

### Schritt 7: Kapitel 6 — Abrechnung (Supporting Sub-Domain)

- **Quelle:** Event Storming (Phase 4.3, 5.3), Anforderungen (R-01…R-06)
- **Inhalt:**
  - **Read Models**: Tagesabrechnung (R-01), Abrechnung pro Tisch (R-03), Abrechnung pro Servicekraft (R-04), Produktumsatz (R-05). Beschreibung der Datenquellen und Inhalte.
  - **Aggregationsstrategie**: On-the-fly vs. vorberechnete Projektionen → Hotspot H8. Argumentation: Für die erwartete Größenordnung (500–2000 Events pro Veranstaltung) sollte on-the-fly SQL-Aggregation ausreichen.
  - **Tagesabschluss (R-06)**: Prozess (offene Tische prüfen → Abschlussbericht → optional System-Reset). Offene Fragen: Umgang mit offenen Saldi, Veranstaltungskonzept, Archivierung → Hotspot H7.
  - **Datenexport (R-02)**: CSV-Export von Umsätzen, Bestellungen, Artikeldaten. Admin-only.

### Schritt 8: Kapitel 7 — Auth (Generic Sub-Domain)

- **Quelle:** Event Storming (Phase 2.4), Anforderungen (A-01…A-03)
- **Inhalt:**
  - **Authentifizierung**: JWT-basiert, Passwort-Hashing (Argon2id), generische Fehlermeldungen. Token-Gültigkeit: 12 Stunden.
  - **Autorisierung**: Drei Rollen (admin, senior_service, service) mit Berechtigungsmatrix (übernommen aus anforderungen.md).
  - **Sicheres Onboarding**: Einmalpasswort (6-stellig) → Eigenes Passwort setzen bei Erstanmeldung.

### Schritt 9: Kapitel 8 — Read Models

- **Quelle:** Event Storming (Phase 5.3)
- **Inhalt:** Zentrales Kapitel, das alle Read Models aus dem Event Storming zusammenfasst.
  - **Service-Ansichten**: Tischübersicht, Tischdetails, Produktkatalog, Kassenjournal
  - **Admin-Ansichten (Reporting)**: Tagesabrechnung, Abrechnung pro Tisch, Abrechnung pro Servicekraft, Produktumsatz
  - **Ausgabe-Ansichten**: KDS-Ansicht, Zubereitungsstatus
- Für jedes Read Model: Datenquelle, Inhalt, Akteure — exakt wie im Event Storming (Phase 5.3) erarbeitet.
- **Kein** Read Model „Produktkatalog" im aktuellen Entwurf vorhanden → ergänzen.

### Schritt 10: Kapitel 9 — Persistenzstrategie

- **Quelle:** Event Storming (Phase 5.1: Event-Sourcing vs. CRUD)
- **Inhalt:**
  - Übersichtstabelle: Kassenbetrieb → Event-Sourcing, Stammdaten/Auth → CRUD.
  - **Event Store**: Append-only, Optimistic Concurrency Control über Version (UNIQUE tisch_id + version). **Keine SQL-DDL** — nur Prinzipien und Constraints.
  - **Snapshot-Speicherung**: Separate Speicherung (nicht im Event Stream). Automatisch nach N Events oder auf Admin-Anfrage. Rein technische Optimierung.
  - **CRUD-Tabellen**: Prinzipien (Soft-Delete, Timestamps, referenzielle Integrität). **Keine SQL-DDL**.
  - **Optimistic Concurrency**: Erklärung des Version-Mechanismus und des Retry-Patterns bei Konflikten.

### Schritt 11: Kapitel 10 — Architekturprinzipien

- **Inhalt:** Kompakte Darstellung der übergreifenden Architekturprinzipien — ohne Verzeichnisbäume, ohne Endpunkt-Listen, ohne Technologie-Versionen.
  - **Schichtenarchitektur**: 4-Schichten-Diagramm (HTTP → Application → Domain → Repository/Infra) mit kurzer Beschreibung jeder Schicht. Request-Lebenszyklus als Beispiel (Bestellung aufgeben).
  - **API-Design**: POST-only, JSON, JWT-Auth, Rollenprüfung in Middleware, Fehlerformat. Bereichsgliederung: Auth (`/auth/*`), Admin (`/admin/*`), Service (`/service/*`), Senior Service (`/senior-service/*`). **Keine vollständige Endpunkt-Liste** — nur Prinzipien und Beispiele.
  - **Frontend-Architektur**: Mobile-first SPA, Guards (AdminGuard, ServiceGuard), Seitenstruktur nach Bereichen (Service: Tischübersicht → Tisch-Detail mit Tabs; Admin: Produkte, Tische, Benutzer). UI-Patterns (Karten, Drawer, Tab-Navigation, Kategorie-Tabs, Plus/Minus). Backend-Kommunikation über Backend-Klassen mit BackendClient-Interface. **Keine Framework-Versionen**.
  - **Validierung**: Frontend (Schema-Validierung) + Backend (unabhängige Schema-Validierung). Backend = Single Source of Truth.
  - **Geldbeträge**: Immer in Cent (Integer). Durchgehend: DB, Backend, API, Frontend, Events.
  - **Mehrbenutzerfähigkeit**: Parallele Zugriffe an verschiedenen Tischen → kein Konflikt. Am selben Tisch → Optimistic Concurrency.
  - **Mobile-first**: ≥360px Breite, Touch-optimiert, Drawer-Konzept, kein Hover, kein App-Download.
  - **Sicherheit**: HTTPS, Rate Limiting auf Login, Security Headers, Input-Validierung auf beiden Seiten, keine Secrets im Code, Passwort-Hashing mit Argon2id.

### Schritt 12: Kapitel 11 — Infrastruktur und Deployment

- **Inhalt:** Architekturüberblick-Diagramm (Docker Compose: nginx, Backend, PostgreSQL, Frontend, Migrate-Container). Deployment-Modell: Self-hosted, VPS oder Raspberry Pi, Let's Encrypt, keine externe Abhängigkeit.
- Übernahme aus dem aktuellen Entwurf (Kapitel 9) — bereits angemessen abstrakt.

### Schritt 13: Kapitel 12 — Ubiquitous Language

- **Quelle:** Event Storming (Abschnitt 8.3)
- **Inhalt:** Vollständige Ubiquitous Language, gruppiert nach Bounded Context:
  - Kassenbetrieb: Tisch, Bestellung, Position, Lieferung, Zahlung, Stornierung, Saldo, Kassenjournal, Bezeichnung, Kommentar, Freibon
  - Stammdaten: Produkt, Variante, Kategorie, Preis, Soft-Delete
  - Auth: Rolle, Einmalpasswort, Token
  - Ausgabe: Bon, KDS, Zubereitungsstatus, Ausgabestation
  - Abrechnung: Tagesabrechnung, Umsatz, Stornoquote, Tagesabschluss, Export
  - Übergreifend: Event-Sourcing, Fat Event, ACL, Append-only, BYOD

### Schritt 14: Kapitel 13 — Priorisierung und Ausbaustufen

- **Quelle:** Event Storming (Abschnitt 8.2)
- **Inhalt:** Drei Stufen (Must-have, Should-have, Nice-to-have) — exakt die Tabellenform aus dem Event Storming. **Keine Implementierungsreihenfolge innerhalb der Stufen** — das gehört in Planungsdokumente.

### Schritt 15: Kapitel 14 — Offene Entwurfsfragen

- **Quelle:** Event Storming (Abschnitt 7: Hotspots H1–H8)
- **Inhalt:** Tabelle mit allen 8 Hotspots — Thema, Kernfrage, Priorität. Erweitert um die konkreten offenen Fragen aus den jeweiligen Hotspot-Abschnitten.

### Schritt 16: Review und Querverweise

- Prüfen, dass jede Anforderung aus `anforderungen.md` im Entwurf adressiert ist.
- Prüfen, dass alle Domain Events aus dem Event Storming abgebildet sind.
- Prüfen, dass alle Read Models aus dem Event Storming aufgeführt sind.
- Querverweise zu den Quelldokumenten konsistent halten.

## Offene Fragen / Risiken

1. **Detailgrad der Event-Strukturen:** Der neue Entwurf soll die Event-Strukturen als Pseudocode/Baumdarstellung beibehalten (wie im aktuellen Entwurf) — das ist der richtige Detailgrad für einen High-Level-Entwurf. Keine JSON-Schemas oder Go-Structs.

2. **Snapshot-Modellierung:** Die Entscheidung, Snapshots _nicht_ als Events im Stream zu speichern, weicht vom aktuellen Entwurf ab. Das ist eine bewusste Architekturentscheidung, die aus dem Event Storming abgeleitet ist (Snapshots sind „rein technisch, keine fachliche Bedeutung"). Die Alternative (Snapshot als Event) sollte kurz als verworfene Option erwähnt werden.

3. **FreibonAusgestellt-Abgrenzung:** Das Event Storming lässt offen, ob `FreibonAusgestellt` ein eigener Event-Typ oder eine Markierung im `BestellungAufgegeben`-Event ist (Hotspot H2). Der Entwurf sollte beide Optionen benennen, aber eine bevorzugte Variante empfehlen (eigener Event-Typ, weil semantisch klarer).

4. **Stornierung bezahlter Positionen:** Die erweiterte Stornierungsinvariante (auch bezahlte Positionen stornierbar) hat Auswirkungen auf den Saldo — er kann temporär negativ werden. Das muss im Entwurf als bewusstes Design dokumentiert werden.

5. **Dokumentlänge:** Der aktuelle Entwurf hat ~950 Zeilen. Der neue wird durch die Aufwertung der Bounded-Context-Kapitel und die vollständige Ubiquitous Language vermutlich ähnlich lang oder etwas länger. Das SQL-DDL und die Endpunkt-Listen fallen weg, aber die fachlichen Beschreibungen der Ausgabe- und Abrechnung-Contexts kommen hinzu. Ziel: 800–1100 Zeilen.

## Referenzen

| Dokument               | Pfad                            |
| ---------------------- | ------------------------------- |
| Aktueller Entwurf      | `docs/design/entwurf.md`        |
| Event Storming         | `docs/design/event-storming.md` |
| Anforderungen          | `docs/anforderungen.md`         |
| Produktbeschreibung    | `docs/produktbeschreibung.md`   |
| AGENTS.md (Tech-Stack) | `AGENTS.md`                     |
