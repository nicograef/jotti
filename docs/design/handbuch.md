# Entwickler-Handbuch — jotti

> **Hinweis:** Dieses Handbuch ist die **verbindliche Architektur-Referenz** für jotti. Die ursprünglichen Entwurfsdokumente (`entwurf.md`, `event-storming.md`) sind historische Artefakte der Entwurfsphase und werden nicht aktuell gehalten.

---

## Inhaltsverzeichnis

1. [Überblick](#1-überblick)
   - [1.1 Systemvision](#11-systemvision)
   - [1.2 Designziele](#12-designziele)
   - [1.3 Bewusste Abgrenzung](#13-bewusste-abgrenzung)
2. [Bounded Contexts](#2-bounded-contexts)
   - [2.1 Kontextübersicht](#21-kontextübersicht)
   - [2.2 Beziehungen zwischen Kontexten](#22-beziehungen-zwischen-kontexten)
3. [Kassenbetrieb (Core Domain)](#3-kassenbetrieb-core-domain)
   - [3.1 Tisch-Aggregat](#31-tisch-aggregat)
   - [3.2 Invarianten](#32-invarianten)
   - [3.3 Domain Events](#33-domain-events)
   - [3.4 Synchrone Projektion und Event Replay](#34-synchrone-projektion-und-event-replay)
   - [3.5 Policies](#35-policies)
4. [Stammdaten](#4-stammdaten)
   - [4.1 Produkt-Aggregat](#41-produkt-aggregat)
   - [4.2 Tisch-Stammdaten](#42-tisch-stammdaten)
   - [4.3 Benutzer-Aggregat](#43-benutzer-aggregat)
   - [4.4 Persistenz (CRUD)](#44-persistenz-crud)
   - [4.5 Ausgabe und Abrechnung](#45-ausgabe-und-abrechnung)
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
8. [Priorisierung](#8-priorisierung)
9. [Ubiquitous Language](#9-ubiquitous-language) → [language.md](language.md)

---

## 1. Überblick

### 1.1 Systemvision

jotti ist ein Mobile-Point-of-Sale-System für temporäre Gastronomie-Veranstaltungen gemeinnütziger Organisationen. Ehrenamtliche Servicekräfte nehmen auf ihren eigenen Smartphones (BYOD) im Browser Bestellungen auf, liefern aus, kassieren und stornieren — alles pro Tisch. Admins verwalten Produkte, Tische und Benutzer. Das System ist self-hosted per Docker Compose.

### 1.2 Designziele

| Ziel                        | Bedeutung                                                                                                     |
| --------------------------- | ------------------------------------------------------------------------------------------------------------- |
| **Radikale Einfachheit**    | Minimaler Funktionsumfang, der genau das abdeckt, was ein Vereinsfest braucht — nicht mehr.                   |
| **Mobile-first**            | Alle Interaktionen sind für Smartphone-Browser und Touch-Bedienung optimiert.                                 |
| **Lückenlose Transparenz**  | Jede Transaktion ist unveränderlich protokolliert. Kein Datenverlust, keine Manipulation.                     |
| **Null Kosten**             | Keine Hardware, keine Abo-Gebühren, keine externe Abhängigkeit.                                               |
| **Volle Datenhoheit**       | Self-hosted, alle Daten auf dem eigenen Server.                                                               |
| **Niedrige Einstiegshürde** | Keine Schulung, keine App-Installation. Browser öffnen, einloggen, loslegen.                                  |
| **Nachvollziehbarkeit**     | Event-Sourcing für den Kassenbetrieb: Jede Bestellung, Zahlung und Stornierung ist jederzeit nachvollziehbar. |

### 1.3 Bewusste Abgrenzung

Folgende Features sind **bewusst nicht enthalten** — jedes zusätzliche Feature erhöht Komplexität für ehrenamtliche Teams:

- Kartenzahlung / Zahlungsgateway
- Zertifizierte TSE (KassenSichV)
- Reservierungssystem
- Warenwirtschaft / Inventory
- Lieferservice-Integration
- Multi-Standort-Verwaltung
- Kundenverwaltung / CRM
- Selbstbedienungs-Kiosk
- Trinkgeld-Tracking

---

## 2. Bounded Contexts

### 2.1 Kontextübersicht

| Context           | Typ                   | Beschreibung                                                          | Persistenz      |
| ----------------- | --------------------- | --------------------------------------------------------------------- | --------------- |
| **Kassenbetrieb** | Core Domain           | Tisch-basierte Vorgänge: Bestellen, Liefern, Bezahlen, Stornieren     | Event-Sourcing  |
| **Stammdaten**    | Supporting Sub-Domain | Verwaltung von Produkten, Tischen, Benutzern (CRUD)                   | CRUD            |
| **Ausgabe**       | Supporting Sub-Domain | Bondruck, Küchendisplay (KDS), Zubereitungsstatus                     | Event-getrieben |
| **Abrechnung**    | Supporting Sub-Domain | Tagesabrechnung, Umsatzberichte, Datenexport (Read-only-Projektionen) | Read-only       |
| **Auth**          | Generic Sub-Domain    | Login, Logout, Passwort-Management, Token-Verwaltung                  | Infrastruktur   |

### 2.2 Beziehungen zwischen Kontexten

| Upstream      | Downstream    | Beziehungstyp                     | Beschreibung                                                                             |
| ------------- | ------------- | --------------------------------- | ---------------------------------------------------------------------------------------- |
| Stammdaten    | Kassenbetrieb | Customer/Supplier + ACL           | Kassenbetrieb liest Produkte/Tische, friert Daten zum Bestellzeitpunkt in Fat Events ein |
| Kassenbetrieb | Ausgabe       | Published Language (Event-driven) | Bestellungs-Events triggern KDS-Anzeige und Bon-Druck                                    |
| Kassenbetrieb | Abrechnung    | Published Language (Event-driven) | Tisch-Events werden zu Auswertungen projiziert                                           |
| Auth          | Kassenbetrieb | Open Host Service                 | Token mit Benutzer-ID und Rolle                                                          |
| Auth          | Stammdaten    | Open Host Service                 | Token mit Benutzer-ID und Rolle                                                          |
| Auth          | Ausgabe       | Open Host Service                 | Token mit Benutzer-ID und Rolle                                                          |
| Auth          | Abrechnung    | Open Host Service                 | Token mit Benutzer-ID und Rolle                                                          |

Der Kassenbetrieb schützt sich über eine **Anti-Corruption Layer (ACL)** vor Stammdaten-Änderungen: Bestellungs-Events enthalten alle relevanten Produktdaten zum Zeitpunkt der Bestellung (Fat Events). Spätere Preisänderungen haben keinen Einfluss auf historische Bestellungen.

---

## 3. Kassenbetrieb (Core Domain)

### 3.1 Tisch-Aggregat

Das Tisch-Aggregat ist die zentrale transaktionale Grenze im Kassenbetrieb. Der Zustand wird über eine synchrone Projektion in der `table_state`-Tabelle vorgehalten, die in derselben Transaktion wie das Event-INSERT aktualisiert wird (→ [3.4](#34-synchrone-projektion-und-event-replay)).

**Event-Stream-Modell:** Der Event Stream enthält die vollständige Historie: Tisch → Bestellungen → Positionen.

```
Tisch (Event Stream)
├── tisch_id              (int — DB-generiert)
├── saldo                 (int, Cent — berechnet)
├── event_version         (int — letzte Event-Version)
└── bestellungen[]
    ├── bestellung_id     (UUID — im Event generiert)
    ├── kommentar?        (string, optional — max. 100 Zeichen)
    ├── zeitstempel       (datetime)
    ├── benutzer_id       (int)
    ├── benutzer_name     (string)
    └── positionen[]
        ├── position_id   (UUID — im Event generiert)
        ├── variante_id   (int)
        ├── produkt_name  (string — Fat Event)
        ├── variante_name (string — Fat Event)
        ├── kategorie     (essen | getraenk | sonstiges — Fat Event)
        ├── einzelpreis   (int, Cent — Fat Event)
        └── menge         (int, ≥ 1)
```

**Projektions-Modell (`TischState`):** Die synchrone Projektion ist eine flache Struktur — keine Bestellungs-Hierarchie, sondern mengenbasierte Positionslisten. Lieferung, Bezahlung und Stornierung reduzieren die Menge in diesen Listen (statt Boolean-Flags zu setzen).

```
TischState (Projektion)
├── saldo_cents                (int, Cent — berechnet)
├── unbezahlte_positionen[]    (Position mit verbleibender Menge)
├── ungelieferte_positionen[]  (Position mit verbleibender Menge)
├── gesamt_zahlungen_cents     (int, Cent)
├── last_event_id              (int)
└── last_event_version         (int)
```

**Fat Events:** Produktdaten (Name, Variantenname, Kategorie, Einzelpreis) werden zum Bestellzeitpunkt im Event eingefroren. Spätere Stammdatenänderungen haben keinen Einfluss auf historische Bestellungen — der Kassenbetrieb schützt sich so per ACL vor dem Stammdaten-Context.

### 3.2 Invarianten

$$\text{Saldo} = \sum \text{Bestellungen} - \sum \text{Zahlungen} - \sum \text{Stornierungen}$$

Alle Beträge in Cent (Integer). Saldo = 0 bedeutet: alle Positionen bezahlt oder storniert.

- **Liefer-Invariante:** Nur bestellte, nicht-stornierte Positionen können geliefert werden. Bereits gelieferte Positionen nicht erneut lieferbar. Teillieferungen zulässig.
- **Bezahl-Invariante:** Nur bestellte, nicht-stornierte, nicht-bezahlte Positionen können bezahlt werden. Der Zahlungsbetrag ergibt sich aus der Summe der gewählten Positionen — Überzahlung nicht möglich. Teilzahlungen zulässig.
- **Stornierungsinvariante:** Nur bestellte, nicht-stornierte Positionen können storniert werden — **unabhängig vom Liefer- und Bezahlstatus**. Bei Stornierung bereits bezahlter Positionen kann der Saldo temporär negativ werden (bewusstes Design).
- **Rolleninvariante:** Stornierungen nur durch `serviceleitung` und `admin`. Alle anderen Tischoperationen (Bestellen, Liefern, Bezahlen) stehen allen drei Rollen zur Verfügung.
- **Mindestmengen-Invariante:** Jede Operation erfordert mindestens eine Position. Bestellung, Lieferung, Zahlung oder Stornierung ohne Positionen sind ungültig.

### 3.3 Domain Events

Vier unveränderliche (append-only) Event-Typen. Namenskonvention: deutsch, Vergangenheitsform, PascalCase.

**Gemeinsame Event-Metadaten:**

```
event_id          (int — DB-generiert, eindeutige Event-ID)
tisch_id          (int — Aggregat-ID, abgeleitet aus subject "tisch:<id>")
benutzer_id       (int — wer hat die Aktion ausgeführt)
benutzer_name     (string — Fat Event: Name zum Zeitpunkt der Aktion)
zeitstempel       (datetime — Zeitpunkt der Erzeugung)
version           (int — aufsteigende Versionsnummer pro Tisch, für OCC)
```

#### BestellungAufgegeben

Servicekraft gibt eine Bestellung am Tisch auf.

```
BestellungAufgegeben
├── [Event-Metadaten]
├── bestellung_id        (UUID)
├── gesamt_preis_cents   (int, Cent — Summe aller Positionen)
├── kommentar?           (string, optional — max. 100 Zeichen)
└── positionen[]
    ├── position_id      (UUID)
    ├── variante_id      (int)
    ├── produkt_name     (string — Fat Event)
    ├── variante_name    (string — Fat Event)
    ├── kategorie        (essen | getraenk | sonstiges — Fat Event)
    ├── einzelpreis      (int, Cent — Fat Event)
    └── menge            (int, ≥ 1)
```

#### ProdukteGeliefert

Bestellte Positionen werden als geliefert markiert. Teillieferungen möglich.

```
ProdukteGeliefert
├── [Event-Metadaten]
├── lieferung_id      (UUID)
├── positionen[]      (PositionRef)
│   ├── position_id   (UUID)
│   └── menge         (int, ≥ 1)
└── kommentar?        (string, optional — max. 100 Zeichen)
```

#### ZahlungRegistriert

Barzahlung wird registriert. Betrag = Summe der gewählten Positionen. Teilzahlungen möglich.

```
ZahlungRegistriert
├── [Event-Metadaten]
├── zahlung_id            (UUID)
├── positionen[]          (PositionRef)
│   ├── position_id       (UUID)
│   └── menge             (int, ≥ 1)
├── gesamt_zahlung_cents  (int, Cent — Summe der gewählten Positionen)
└── kommentar?            (string, optional — max. 100 Zeichen)
```

#### ProdukteStorniert

Serviceleitung oder Admin storniert Positionen. Unabhängig vom Liefer-/Bezahlstatus.

```
ProdukteStorniert
├── [Event-Metadaten]
├── stornierung_id             (UUID)
├── positionen[]               (PositionRef)
│   ├── position_id            (UUID)
│   └── menge                  (int, ≥ 1)
├── gesamt_stornierung_cents   (int, Cent — Summe der stornierten Positionen)
└── kommentar?                 (string, optional — max. 100 Zeichen)
```

### 3.4 Synchrone Projektion und Event Replay

Der Tisch-Zustand wird über eine synchrone Projektion in der `table_state`-Tabelle vorgehalten. Die Projektion wird in derselben Transaktion wie das Event-INSERT aktualisiert (Write-Through):

```
BEGIN TX
  1. event ← INSERT INTO events (...)          → event_id
  2. zustand ← SELECT * FROM table_state       → aktueller Zustand (oder Zero-Value)
  3. zustand ← ApplyEvent(zustand, event)      → neuer Zustand (reine Go-Funktion)
  4. UPSERT INTO table_state (zustand)         → neuer Zustand persistiert
COMMIT TX
```

Die `ApplyEvent()`-Funktion (`backend/domain/table/projection.go`) ist eine reine Funktion in der Domain-Schicht — kein DB-Zugriff. Sie nimmt einen `TischState` und ein `Event` entgegen und gibt den neuen `TischState` zurück. Der Projektor ist ein internes Detail des Event-Repositories; der Command-Service ruft weiterhin nur `WriteEvent()` auf.

**`table_state`-Schema:**

| Spalte                    | Typ          | Beschreibung                                    |
| ------------------------- | ------------ | ----------------------------------------------- |
| `tisch_id`                | INT (PK, FK) | Referenz auf `tische.id`                        |
| `saldo_cents`             | INT          | Aktueller Tisch-Saldo in Cent                   |
| `unbezahlte_positionen`   | JSONB        | `[]Position` — noch nicht bezahlte Positionen   |
| `ungelieferte_positionen` | JSONB        | `[]Position` — noch nicht gelieferte Positionen |
| `gesamt_zahlungen_cents`  | INT          | Summe aller Zahlungen in Cent                   |
| `last_event_id`           | INT (FK)     | ID des zuletzt verarbeiteten Events             |
| `last_event_version`      | INT          | Version des zuletzt verarbeiteten Events        |
| `updated_at`              | TIMESTAMPTZ  | Zeitpunkt der letzten Aktualisierung            |

**Lesezugriff (Queries):** Operative Queries (Saldo, unbezahlte/ungelieferte Positionen) lesen direkt aus `table_state` — kein Event-Replay nötig. Das Kassenjournal (Historie) liest weiterhin den vollständigen Event Stream via `ReadEventsBySubject()`.

**Apply-Tabelle:**

| Event-Typ            | Zustandsänderung                                                                                              |
| -------------------- | ------------------------------------------------------------------------------------------------------------- |
| BestellungAufgegeben | Positionen zu `unbezahlte_positionen` und `ungelieferte_positionen` hinzufügen, Saldo erhöhen                 |
| ProdukteGeliefert    | Referenzierte Mengen aus `ungelieferte_positionen` subtrahieren (Eintrag entfernen bei Menge 0)               |
| ZahlungRegistriert   | Referenzierte Mengen aus `unbezahlte_positionen` subtrahieren, Saldo und `gesamt_zahlungen_cents` anpassen    |
| ProdukteStorniert    | Referenzierte Mengen aus `unbezahlte_positionen` und `ungelieferte_positionen` subtrahieren, Saldo reduzieren |

**Selbstheilung:** Bei Inkonsistenz kann `table_state` jederzeit aus allen Events reberechnet werden — der Event Stream bleibt die Single Source of Truth. Details zur Projektionsarchitektur: [ADR: CQRS](../adr/cqrs.md).

### 3.5 Policies

- **Stornierungsberechtigung (K-04):** Nur `serviceleitung` und `admin` dürfen stornieren. Die Berechtigung wird in der Anwendungsschicht geprüft, bevor der Command an das Aggregat geht.
- **Automatischer Bon-Druck nach Kategorie (K-11):** Bei `BestellungAufgegeben` wird ein Bon pro Kategorie an die zugeordnete Ausgabestation gesendet (Essen → Küchenbon, Getränke → Thekenbon). Kategorie-Drucker-Zuordnung in den Stammdaten konfiguriert.
- **Umbuchung (K-08):** Verschiebt eine Bestellung von Quell- auf Ziel-Tisch (= Stornierung + neue Bestellung). Cross-Aggregat-Transaktion — Atomarität auf Anwendungsebene sicherstellen. Nur `serviceleitung` und `admin`.

---

## 4. Stammdaten

### 4.1 Produkt-Aggregat

Das Produkt-Aggregat verwaltet den Produktkatalog der Veranstaltung. Jedes Produkt gehört zu einer Kategorie und kann beliebig viele Varianten besitzen — jede Variante mit eigenem Namen und Preis.

```
Produkt
├── produkt_id       (int — DB-generiert)
├── name             (string — nicht leer)
├── kategorie        (essen | getraenk | sonstiges)
├── status           (active | inactive | deleted)
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
- Soft-Delete: Produkte und Varianten werden durch Status-Änderung auf `deleted` entfernt, nicht physisch gelöscht. Historische Bestellungen bleiben valide, weil die Events die Produktdaten zum Bestellzeitpunkt enthalten (Fat Events).
- Varianten können unabhängig vom Produkt deaktiviert werden (`inactive`). Inaktive Varianten erscheinen nicht im Service-Katalog.

### 4.2 Tisch-Stammdaten

Das Tisch-Stammdaten-Aggregat verwaltet die Basisdaten eines Tisches: seinen Namen und seinen Status. Es ist strikt vom Tisch-Aggregat im Kassenbetrieb (→ [3.1](#31-tisch-aggregat)) zu unterscheiden.

```
Tisch (Stammdaten)
├── tisch_id    (int — DB-generiert)
├── name        (string — nicht leer, z. B. „Tisch 1", „Stehtisch Eingang")
└── status      (active | inactive | deleted)
```

**Invarianten:**

- Name darf nicht leer sein.
- Soft-Delete: Tische werden durch Status-Änderung auf `deleted` entfernt. Der Datensatz bleibt erhalten, damit historische Events valide bleiben.
- Nur aktive Tische (`active`) erscheinen in der Tischübersicht der Servicekräfte.

**Abgrenzung zum Kassenbetrieb:** Im Kassenbetrieb ist der Tisch ein Event-Sourced-Aggregat mit Bestellungen, Zahlungen und Saldo. In den Stammdaten ist er eine einfache CRUD-Entität mit Name und Status. Beide teilen sich die `tisch_id`, haben aber unterschiedliche Verantwortlichkeiten und Persistenzstrategien.

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
- Passwort wird mit Argon2id gehasht gespeichert — Klartext-Passwörter werden nie persistiert.
- Soft-Delete: Benutzer werden durch Status-Änderung auf `deleted` entfernt. Deaktivierte (`inactive`) und entfernte (`deleted`) Benutzer können sich nicht anmelden.
- Neue Benutzer werden initial mit Status `inactive` angelegt und müssen durch den Admin aktiviert werden.
- Bei Neuanlage oder Passwort-Reset wird ein 6-stelliges Einmalpasswort generiert und als `einmalpasswort_hash` gespeichert. Der reguläre `passwort_hash` wird geleert. Das System erkennt am Zustand `einmalpasswort_hash ≠ NULL ∧ passwort_hash = NULL`, dass der Benutzer ein eigenes Passwort vergeben muss (→ [5.2](#52-onboarding-ablauf)).

### 4.4 Persistenz (CRUD)

Stammdaten (Produkte, Tische, Benutzer) werden mit klassischem CRUD verwaltet. Event-Sourcing ist hier nicht nötig — die historischen Daten stecken bereits in den Fat Events des Kassenbetrieb-Context.

- **Soft-Delete statt physischem Löschen:** Datensätze werden durch Status-Änderung auf `deleted` entfernt. Physisches Löschen ist nicht vorgesehen, damit referenzielle Integrität und historische Nachvollziehbarkeit erhalten bleiben.
- **Timestamps:** Alle Stammdaten tragen `erstellt_am` und `aktualisiert_am` Zeitstempel.
- **Referenzielle Integrität:** Produkte und Varianten werden nie physisch gelöscht, damit Fremdschlüssel-Referenzen aus dem Event Store valide bleiben.

### 4.5 Ausgabe und Abrechnung

**Ausgabe (Supporting Sub-Domain):** Der Ausgabe-Context umfasst Bondruck, Küchendisplay (KDS) und Zubereitungsstatus. Bons werden automatisch bei Bestellungen nach Kategorie an Ausgabestationen gesendet (Essen → Küche, Getränke → Theke). Der Ausgabe-Context ist nicht Teil des MVP — Details in `entwurf.md` Kap. 5.

**Abrechnung (Supporting Sub-Domain):** Der Abrechnung-Context konsumiert Tisch-Events und projiziert sie in Read-only-Auswertungen: Tagesabrechnung, Abrechnung pro Tisch/Servicekraft, Produktumsatz und CSV-Datenexport. Alle Reporting-Ansichten sind reine Read Models ohne eigene Events. Die Abrechnung ist nicht Teil des MVP — Details in `entwurf.md` Kap. 6.

---

## 5. Auth und Rollen

### 5.1 Rollen und Berechtigungsmatrix

jotti kennt drei Rollen mit abgestuften Berechtigungen. Die Rollenprüfung erfolgt serverseitig anhand des JWT.

| Rolle              | Code-Bezeichnung | Beschreibung                                                                 |
| ------------------ | ---------------- | ---------------------------------------------------------------------------- |
| **Admin**          | `admin`          | Voller Zugriff auf Stammdaten (Produkte, Tische, Benutzer) und Kassenbetrieb |
| **Serviceleitung** | `serviceleitung` | Kassenbetrieb einschließlich Stornierung                                     |
| **Servicekraft**   | `service`        | Kassenbetrieb ohne Stornierung                                               |

**Berechtigungsmatrix:**

| Aktion                   | Admin | Serviceleitung | Servicekraft |
| ------------------------ | :---: | :------------: | :----------: |
| Produkte verwalten       |   ✔   |                |              |
| Tische verwalten         |   ✔   |                |              |
| Benutzer verwalten       |   ✔   |                |              |
| Passwort zurücksetzen    |   ✔   |                |              |
| Bestellung aufgeben      |   ✔   |       ✔        |      ✔       |
| Lieferung bestätigen     |   ✔   |       ✔        |      ✔       |
| Zahlung registrieren     |   ✔   |       ✔        |      ✔       |
| Stornierung durchführen  |   ✔   |       ✔        |              |
| Tischübersicht einsehen  |   ✔   |       ✔        |      ✔       |
| Kassenjournal einsehen   |   ✔   |       ✔        |      ✔       |
| Tagesabrechnung einsehen |   ✔   |                |              |
| Datenexport              |   ✔   |                |              |
| Tagesabschluss einleiten |   ✔   |                |              |
| Abmelden                 |   ✔   |       ✔        |      ✔       |

Die Rollenhierarchie ist inklusiv: Admin kann alles, was Serviceleitung kann. Serviceleitung kann alles, was Servicekraft kann — plus Stornierung.

### 5.2 Onboarding-Ablauf

Neue Benutzer durchlaufen einen zweistufigen Onboarding-Prozess, der sicherstellt, dass nur der Benutzer sein eigenes Passwort kennt:

1. **Benutzer anlegen:** Der Admin erstellt einen Benutzer mit Name, Benutzername und Rolle. Das System generiert ein 6-stelliges Einmalpasswort, das der Admin dem Benutzer mitteilt (z. B. mündlich oder auf einem Zettel). Der Benutzer wird mit Status `inactive` angelegt und muss vom Admin aktiviert werden.
2. **Erstanmeldung:** Der Benutzer meldet sich mit Benutzername und Einmalpasswort an. Das System erkennt am Zustand `einmalpasswort_hash ≠ NULL ∧ passwort_hash = NULL`, dass noch kein eigenes Passwort vergeben wurde, und leitet zur Seite „Passwort setzen" weiter.
3. **Eigenes Passwort setzen:** Der Benutzer vergibt ein eigenes Passwort (min. 8 Zeichen). Das Einmalpasswort wird gegen den gespeicherten Hash verifiziert. Das neue Passwort wird mit Argon2id gehasht als `passwort_hash` gespeichert, der `einmalpasswort_hash` wird geleert.
4. **Normale Anmeldung:** Ab jetzt meldet sich der Benutzer mit seinem selbst gewählten Passwort an.

**Passwort-Reset:** Bei einem Admin-Reset wird ein neues 6-stelliges Einmalpasswort generiert und als `einmalpasswort_hash` gespeichert. Der reguläre `passwort_hash` wird geleert. Der Benutzer durchläuft beim nächsten Login erneut Schritt 2 und 3.

---

## 6. Architekturprinzipien

### 6.1 Schichtenarchitektur

Das Backend ist in vier Schichten gegliedert:

```
┌─────────────────────────────────────────────────┐
│  HTTP-Schicht (Handler)                         │
│  Routing, Request-Parsing, Response-Serialisier.│
├─────────────────────────────────────────────────┤
│  Application-Schicht (Services)                 │
│  Use Cases, Orchestrierung, Fehler-Mapping      │
├─────────────────────────────────────────────────┤
│  Domain-Schicht                                 │
│  Aggregat-Logik, Invarianten, Domain Events     │
├─────────────────────────────────────────────────┤
│  Repository / Infra-Schicht                     │
│  Datenbankzugriff, Event Store, sqlc-Queries    │
└─────────────────────────────────────────────────┘
```

- **HTTP-Schicht:** Liest den Request-Body, validiert das Format und delegiert an den Application-Service. Definiert eigene Request- und Response-DTOs mit `json`-Tags. Domain-Modelle werden nie direkt serialisiert - dedizierte Mapper-Funktionen übersetzen zwischen Domain und HTTP. Gibt strukturierte Fehlerresponses zurück. Keine Business-Logik.
- **Application-Schicht:** Koordiniert den Use Case: validiert fachlich (zog-Schema), lädt Aggregat-State, ruft Domain-Logik auf und persistiert das Ergebnis. Übersetzt Domain-Fehler in anwendungsseitige Fehlercodes.
- **Domain-Schicht:** Enthält die fachlichen Regeln (Aggregat-Invarianten, Event-Konstruktion, Zustandsberechnung). Kennt keine Datenbank, kein HTTP und keine JSON-Serialisierung. Domain-Structs tragen keine `json`-Tags (Ausnahme: Event-Data-Structs für Event-Store-Persistenz).
- **Repository/Infra-Schicht:** Kapselt alle Datenbankzugriffe. Für das Tisch-Aggregat: Event Store (append-only). Für Stammdaten: CRUD. Implementiert auf Basis von sqlc-generierten Queries.

### 6.2 API-Design

**POST-only:** Alle API-Endpunkte sind POST-Endpunkte. Jede Aktion wird explizit benannt (z. B. `/service/bestellung-aufgeben` statt `PUT /tables/5`).

**JSON:** Request- und Response-Bodies sind JSON.

**Authentifizierung:** Jeder Endpunkt (außer `/auth/*`) erwartet ein gültiges JWT im `Authorization: Bearer <token>`-Header. Die Middleware prüft Signatur und Gültigkeit.

**Fehlerformat:**

```json
{ "code": "<string>", "details": "<optional>" }
```

HTTP-Statuscodes: `400` Client-Fehler, `401` fehlende/ungültige Auth, `403` unzureichende Rechte, `500` Server-Fehler.

**Bereichsgliederung:**

| Bereich        | Pfad-Präfix         | Rolle(n)                             |
| -------------- | ------------------- | ------------------------------------ |
| Auth           | `/auth/*`           | — (öffentlich)                       |
| Admin          | `/admin/*`          | `admin`                              |
| Service        | `/service/*`        | `service`, `serviceleitung`, `admin` |
| Senior Service | `/serviceleitung/*` | `serviceleitung`, `admin`            |

### 6.3 Frontend-Architektur

**Route Guards:** Zwei Guards schützen die Bereiche:

- `AdminGuard` — prüft, ob der eingeloggte Benutzer die Rolle `admin` hat.
- `ServiceGuard` — prüft, ob der Benutzer eingeloggt ist (Rolle `service`, `serviceleitung` oder `admin`).

Nicht autorisierte Zugriffe werden auf `/login` umgeleitet.

**Seitenstruktur:**

| Bereich   | Seiten                                                                                                                                                                                   |
| --------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Service   | Tischübersicht → Tisch-Detail (Tabs: Bestellen, Bezahlen, Historie). Liefern ist in den Bestellen-Tab integriert; Stornieren ist für `serviceleitung`/`admin` im Bezahlen-Tab verfügbar. |
| Admin     | Produkte verwalten · Tische verwalten · Benutzer verwalten                                                                                                                               |
| Allgemein | Login · Passwort setzen (Erstanmeldung)                                                                                                                                                  |

**UI-Patterns:**

- **Karten:** Produkte und Tische werden als Karten dargestellt.
- **Drawer (Bottom-Sheet):** Bestellen, Liefern, Bezahlen und Stornieren öffnen einen Drawer von unten mit Zusammenfassung und Bestätigung.
- **Tab-Navigation:** Im Tisch-Detail navigiert der Benutzer zwischen den Aktionen über Tabs.
- **Plus/Minus:** Mengenauswahl über Plus/Minus-Buttons (Touch-optimiert).

**BackendClient:** Das Frontend kommuniziert ausschließlich über Backend-Klassen, die das `BackendClient`-Interface verwenden. Direktes `fetch()` ist verboten.

### 6.4 Validierung

Alle Eingaben werden auf beiden Seiten unabhängig voneinander validiert:

| Seite    | Schema-Bibliothek | Zeitpunkt                                    |
| -------- | ----------------- | -------------------------------------------- |
| Frontend | Zod               | Vor dem Absenden — direktes Feedback am Feld |
| Backend  | zog               | Bei jedem Request — vor der Business-Logik   |

Das Backend ist die Single Source of Truth. Das Frontend-Schema ist eine UX-Optimierung (sofortiges Feedback), aber keine Sicherheitsmaßnahme — das Backend lehnt ungültige Anfragen unabhängig vom Frontend ab.

### 6.5 Geldbeträge

Alle Geldbeträge werden durchgehend als ganzzahlige Cent-Werte (Integer) gespeichert und verarbeitet. Fließkommazahlen werden für Geldbeträge nirgendwo verwendet.

| Ebene     | Datentyp       | Beispiel         |
| --------- | -------------- | ---------------- |
| Datenbank | `INTEGER`      | `350` (= 3,50 €) |
| Backend   | `int`          | `350`            |
| API       | JSON-Zahl      | `350`            |
| Frontend  | `number` (int) | `350`            |
| Events    | `int`          | `350`            |

Die Darstellung als „3,50 €" geschieht ausschließlich im Frontend als reine Formatierung (`formatCents()`).

### 6.6 Mehrbenutzerfähigkeit (OCC)

Das System verwendet zwei Persistenzstrategien:

| Bereich                               | Strategie      | Begründung                                                             |
| ------------------------------------- | -------------- | ---------------------------------------------------------------------- |
| Kassenbetrieb (Tisch)                 | Event-Sourcing | Geschichte ist fachlich relevant (Kassenjournal, Buchhaltung)          |
| Stammdaten (Produkt, Tisch, Benutzer) | CRUD           | Nur aktueller Zustand benötigt; Fat Events decken historische Daten ab |

Mehrere Servicekräfte arbeiten gleichzeitig — Schreibkonflikte am selben Tisch werden über Optimistic Concurrency Control gelöst:

1. Beim Laden eines Tisches wird die aktuelle `event_version` mitgegeben.
2. Beim Schreiben eines neuen Events wird die erwartete Version mitgeschickt.
3. Die Datenbank prüft via UNIQUE Constraint `(subject, version)`, ob die Version noch frei ist.
4. Ist die Version bereits vergeben, schlägt die Operation mit einem Konflikt-Fehler fehl.
5. Die Anwendungsschicht führt einen Retry durch: Tischzustand neu laden, Operation erneut anwenden, neuen Schreibversuch starten.

### 6.7 Sicherheit

| Maßnahme                   | Umsetzung                                                                             | Anforderung |
| -------------------------- | ------------------------------------------------------------------------------------- | ----------- |
| HTTPS / TLS                | nginx terminiert TLS, Let's Encrypt-Zertifikat, HTTP → HTTPS-Redirect                 | Q-06        |
| Rate Limiting              | Login-Endpunkt ist durch Rate Limiting geschützt (Brute-Force-Schutz)                 | Q-07        |
| Security Headers           | Reverse Proxy setzt HSTS, X-Frame-Options, X-Content-Type-Options, CSP                | Q-08        |
| Input-Validierung          | Frontend (Zod) + Backend (zog) — beide Seiten unabhängig voneinander                  | Q-03        |
| Passwort-Hashing           | Argon2id mit zufälligem Salt                                                          | A-01        |
| Generische Fehlermeldungen | Fehlgeschlagene Logins geben keine Auskunft, ob Benutzer oder Passwort falsch war     | A-01        |
| Keine Secrets im Code      | Alle Secrets (JWT-Schlüssel, DB-Passwort) werden über Umgebungsvariablen konfiguriert | —           |
| JWT-Gültigkeit             | Tokens sind 12 Stunden gültig — kurze Lebensdauer begrenzt den Schaden bei Verlust    | A-01        |

---

## 7. Read Models

Read Models sind aufbereitete Lese-Ansichten — reine Projektionen über vorhandene Daten (Events, Projektionstabelle oder Stammdaten). Sie werden nicht direkt geschrieben, sondern durch Events oder CRUD-Operationen aktualisiert.

### 7.1 Service-Ansichten

| Name           | ID   | Quelle                              | Inhalt (Kurzfassung)                                                                                                         |
| -------------- | ---- | ----------------------------------- | ---------------------------------------------------------------------------------------------------------------------------- |
| Tischübersicht | K-05 | `table_state` + Stammdaten          | Pro aktivem Tisch: Name, Saldo, Anzahl unbezahlter und ungelieferter Positionen. Startseite des Service-Bereichs.            |
| Tischdetails   | K-05 | `table_state`                       | Alle Positionen mit Status, gruppiert nach Bestellung. Tabs: Übersicht, Bestellen, Liefern, Bezahlen, Stornieren, Historie.  |
| Produktkatalog | —    | Produkt-Stammdaten                  | Aktive Produkte und Varianten, nach Kategorie gruppiert. Im Bestellvorgang geladen (kein eigenes Navigationsziel).           |
| Kassenjournal  | K-06 | Tisch-Events (Event Stream, Replay) | Chronologische Liste aller Vorgänge am Tisch: Zeitstempel, Typ, Positionen, Betrag, Servicekraft, Kommentar. Unveränderlich. |

Die operativen Ansichten (Tischübersicht, Tischdetails) lesen aus der synchronen Projektionstabelle `table_state` — kein Event-Replay nötig. Das Kassenjournal liest weiterhin den vollständigen Event Stream, da die Historie _der_ Event Stream ist. Details zur Projektionsarchitektur: [ADR: CQRS](../adr/cqrs.md).

### 7.2 Admin-Ansichten (Reporting)

Alle Reporting-Ansichten aggregieren Daten aus `events` und `table_state` tischübergreifend und sind nur für Admins zugänglich. Die Berechnung erfolgt on-demand per SQL-Aggregation (kein Background Worker, kein Eventual Consistency).

| Name                        | ID   | Inhalt (Kurzfassung)                                                                               |
| --------------------------- | ---- | -------------------------------------------------------------------------------------------------- |
| Tagesabrechnung             | R-01 | Gesamtumsatz, Umsatz pro Servicekraft, Stornierungsübersicht, offene Beträge                       |
| Abrechnung pro Tisch        | R-03 | Alle Bestellungen, Zahlungen, Lieferungen, Stornierungen chronologisch; Gesamt-Saldo pro Tisch     |
| Abrechnung pro Servicekraft | R-04 | Umsatz pro Servicekraft, Anzahl Bestellungen, Anzahl und Betrag der Stornierungen                  |
| Produktumsatz               | R-05 | Verkaufte Menge pro Produkt/Variante (abzgl. Stornierungen), Ranking, Gesamteinnahmen pro Variante |

### 7.3 Ausgabe-Ansichten

KDS-Ansicht (K-12) und Zubereitungsstatus (K-13) sind nicht Teil des MVP. Details in `entwurf.md` Kap. 8.3.

---

## 8. Priorisierung

Drei Stufen: Must-have (unverzichtbar für den ersten Einsatz), Should-have (wichtig, nicht blockierend) und Nice-to-have (iterativ ergänzbar). Innerhalb einer Stufe ist keine Reihenfolge vorgegeben.

### 8.1 Stufe 1 — Must-have (MVP)

| ID   | Anforderung                 |
| ---- | --------------------------- |
| K-01 | Bestellung aufgeben         |
| K-02 | Zahlung registrieren        |
| K-03 | Lieferung bestätigen        |
| K-04 | Stornierung                 |
| K-05 | Tischübersicht / Navigation |
| K-06 | Kassenjournal (Historie)    |
| S-01 | Produktverwaltung           |
| S-02 | Tischverwaltung             |
| S-03 | Benutzerverwaltung          |
| A-01 | Login                       |
| A-02 | Passwort setzen             |
| A-03 | Logout                      |
| Q-01 | Usability und Mobile-first  |
| Q-02 | Mehrbenutzerfähigkeit       |
| Q-03 | Validierung                 |
| Q-04 | Datenintegrität             |
| Q-06 | HTTPS / TLS                 |

### 8.2 Stufe 2 — Should-have

| ID   | Anforderung                 |
| ---- | --------------------------- |
| K-11 | Bondruck                    |
| K-12 | Küchendisplay (KDS)         |
| Q-07 | Rate Limiting               |
| Q-08 | Security Headers            |
| R-01 | Tagesabrechnung             |
| R-03 | Abrechnung pro Tisch        |
| R-04 | Abrechnung pro Servicekraft |
| R-05 | Produktumsatz-Reporting     |

### 8.3 Stufe 3 — Nice-to-have

| ID   | Anforderung                             |
| ---- | --------------------------------------- |
| K-08 | Bestellungen umbuchen                   |
| K-09 | Rückgeldberechnung                      |
| K-10 | Tisch-Schnellsuche                      |
| K-13 | Ausgabestationen mit Zubereitungsstatus |
| Q-05 | Offline-Fähigkeit                       |
| R-02 | Datenexport                             |
| R-06 | Tagesabschluss                          |

---

## 9. Ubiquitous Language

Alle Fachbegriffe, Namenskonventionen pro Schicht, Code-Mappings und Ist-vs-Soll-Abweichungen: siehe **[Ubiquitous Language (language.md)](language.md)**.
