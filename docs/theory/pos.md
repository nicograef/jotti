# POS-Systeme & Gastronomie-Domäne — Theorie

Dieses Dokument ist ein theoretisches Nachschlagewerk für Point-of-Sale-Systeme mit Schwerpunkt auf der Gastronomie. Es erklärt die Geschichte von POS-Systemen, gängige Architektur-Patterns, gastronomiespezifische Workflows, Datenmodelle, Fiskalgesetzgebung, Payment Integration sowie den Unterschied zwischen Non-Profit- und kommerziellen POS-Lösungen.

---

## Inhaltsverzeichnis

1. [Warum POS-Architektur verstehen?](#1-warum-pos-architektur-verstehen)
2. [Geschichte der POS-Systeme](#2-geschichte-der-pos-systeme)
3. [POS-Architektur-Patterns](#3-pos-architektur-patterns)
4. [Gastronomie-POS im Detail](#4-gastronomie-pos-im-detail)
5. [Datenmodelle](#5-datenmodelle)
6. [Fiskalgesetzgebung](#6-fiskalgesetzgebung)
7. [Payment Integration](#7-payment-integration)
8. [Non-Profit vs. Commercial POS](#8-non-profit-vs-commercial-pos)
9. [POS-Marktlandschaft](#9-pos-marktlandschaft)
10. [Entscheidungsmatrix: POS-Architektur nach Anwendungsfall](#10-entscheidungsmatrix-pos-architektur-nach-anwendungsfall)
11. [Referenzen](#11-referenzen)

---

## 1. Warum POS-Architektur verstehen?

Ein Point-of-Sale-System ist das **operative Herzstück** jedes Gastronomiebetriebs. Es erfasst Transaktionen in Echtzeit, koordiniert mehrere Rollen (Servicekraft, Küche, Kasse, Admin) und muss auch bei hoher Last und Netzwerkproblemen korrekt funktionieren.

### 1.1 POS als kritisches System

POS-Systeme haben ungewöhnliche Anforderungen, die sie von typischen Web-Applikationen unterscheiden:

| Anforderung                | Erläuterung                                                       |
| -------------------------- | ----------------------------------------------------------------- |
| **Hochverfügbarkeit**      | Ein Ausfall während des Betriebs kostet direkt Umsatz             |
| **Offline-Fähigkeit**      | Netzwerkunterbrechungen dürfen den Kassenbetrieb nicht blockieren |
| **Datenintegrität**        | Buchungsfehler bei Transaktionen sind geschäftskritisch           |
| **Skalierbarkeit**         | Stoßzeiten (Mittagsrush, Events) erzeugen Lastspitzen             |
| **Fiskalkonformität**      | Gesetzliche Aufbewahrungspflichten, Manipulationsschutz           |
| **Usability under Stress** | Bedienoberflächen müssen unter Stress in Sekunden bedienbar sein  |

### 1.2 POS als Domänen-Schnittpunkt

POS-Systeme sind ein Schnittpunkt mehrerer Domänen:

```
┌────────────────────────────────────────────┐
│              POS-System                    │
├────────────┬───────────┬───────────────────┤
│  Gastronomie│  Buchhaltung│  Payment         │
│  Workflows │  & Reporting│  & Banking        │
├────────────┼───────────┼───────────────────┤
│  Inventory │  Personal-  │  Fiskal-          │
│  Management│  verwaltung │  gesetzgebung     │
└────────────┴───────────┴───────────────────┘
```

Diese Komplexität erklärt, warum kommerzielle POS-Systeme teuer sind und warum maßgeschneiderte, schlanke Lösungen für spezifische Anwendungsfälle (wie Vereinsveranstaltungen) durchaus sinnvoll sind.

---

## 2. Geschichte der POS-Systeme

### 2.1 Die mechanische Registrierkasse (1879–1960er)

**1879** erfand James Ritty — Saloon-Besitzer in Dayton, Ohio — die erste mechanische Registrierkasse, um Kassenbetrug durch Angestellte zu verhindern. Das Gerät zählte Transaktionen mechanisch und erzeugte am Tagesende eine Summe. 1884 übernahm John H. Patterson das Unternehmen und gründete die **National Cash Register Company (NCR)** — bis heute einer der bedeutendsten POS-Hersteller.

Frühe Registrierkassen:

- Rein mechanisch, keine Quittungen
- Öffnung der Geldschublade mit Klingelton als Verkaufsnachweis
- Summendruck erst durch Papierolle (Patterson-Innovation)

### 2.2 Electronic Cash Register — ECR (1970er–1980er)

**1973** stellte IBM das System IBM 3650/3660 vor: Ein Mainframe als Store-Controller, der bis zu 128 POS-Terminals verwalten konnte. Das System etablierte erstmals **Client-Server-Architektur**, Peer-to-Peer-Kommunikation und LAN im Kassenumfeld.

**1974** baute William Brobeck & Associates das erste Mikroprozessor-gesteuerte Kassensystem für McDonald's — basierend auf dem Intel 8008. Jede Station zeigte die vollständige Bestellung eines Tisches an und kommunizierte mit zwei vernetzten Computern.

**1986** präsentierte Gene Mosher auf der COMDEX die erste grafische POS-Software mit Touchscreen-Interface — auf dem Atari 520ST. Das **ViewTouch**-System setzte die Blaupause für moderne Touch-POS-Interfaces.

### 2.3 Windows-POS und Commodity-Hardware (1990er–2000er)

Mit dem Aufkommen preiswerter Windows-basierter Hardware entstanden flexible POS-Systeme für den Massenmarkt:

- **Touchscreen-Terminals** wurden erschwinglich
- **SQL-Datenbanken** ersetzten proprietäre Datenspeicher
- **Offene APIs** ermöglichten Integration mit Buchhaltung, Inventory, CRM
- POS-Pakete (Kassensoftware + Hardware) für ~4.000 USD pro Kasse

Nachteile des Windows-POS-Zeitalters: enge Hardware-Bindung, proprietäre Software, komplexe On-Premise-Installationen, hoher Wartungsaufwand.

### 2.4 Cloud-POS und SaaS (2010er)

**Square** (gegründet 2009) demokratisierte den Kartenzahlungsmarkt für Kleinunternehmer mit einem einfachen Dongle am Smartphone. Daraus entwickelte sich eine vollständige Cloud-POS-Plattform.

Merkmale der Cloud-POS-Ära:

- **SaaS-Modell**: monatliche Gebühren statt Einmalkauf
- **iPad/Tablet** als günstige Kassen-Hardware
- **Echtzeit-Berichte** aus der Cloud
- **Automatische Updates** ohne IT-Aufwand
- **API-Ökosystem**: Integrationen mit Buchhaltung, Lieferservice, Reservierung

Wichtige Cloud-POS-Systeme: Square, Toast, Lightspeed, Orderbird, Zettle (by PayPal).

### 2.5 Mobile POS — mPOS (2010er bis heute)

**Mobile POS (mPOS)** nutzt Smartphones oder Tablets als primäre Kassenhardware:

- Keine dedizierte Kassen-Hardware notwendig
- Servicekräfte nehmen Bestellungen am Tisch auf
- Kartenzahlung via Bluetooth-Kartenleser (SumUp, Zettle, Square Reader)
- Progressive Web Apps (PWA) ermöglichen App-ähnliche Erfahrung im Browser

**2020er: Contactless & Integrated Payments** — NFC-Zahlung, QR-Code-Bestellung (Self-Service), Delivery-Integration (UberEats, Lieferando) als Standard.

### 2.6 Zeitstrahl

```
1879  Ritty's Inkorrumpierbare Kassiererin (mechanisch)
1884  NCR — National Cash Register Company
1973  IBM 3650/3660 — Client-Server im Kassenumfeld
1974  Brobeck/McDonald's — erster Mikroprozessor-POS
1986  ViewTouch — erste Touch-GUI für POS
1990s Windows-POS, SQL-Datenbanken, Commodity-Hardware
2009  Square — Cloud-POS, Demokratisierung Kartenzahlung
2010s iPad/Tablet-POS, SaaS-Modell, Cloud-First
2015+ mPOS, NFC/Contactless, PWA-Kassen
2020+ Self-Order Kiosk, QR-Code-Bestellung, Delivery-Integration
```

---

## 3. POS-Architektur-Patterns

### 3.1 On-Premise POS (Traditionell)

```
┌──────────────────────────────────────────────┐
│  Lokal im Restaurant                         │
│                                              │
│  ┌──────────┐    ┌──────────┐               │
│  │  Kasse 1 │    │  Kasse 2 │               │
│  └────┬─────┘    └────┬─────┘               │
│       │               │                     │
│  ┌────┴───────────────┴──────────┐           │
│  │    Lokaler POS-Server         │           │
│  │  (SQL-DB, Geschäftslogik)     │           │
│  └───────────────────────────────┘           │
│                                              │
│  ┌──────────┐    ┌──────────┐               │
│  │Küchendrucker│  │Bondrucker│               │
│  └──────────┘    └──────────┘               │
└──────────────────────────────────────────────┘
```

**Vorteile:**

- Funktioniert ohne Internet
- Geringe Latenz (lokales Netzwerk)
- Volle Datenkontrolle

**Nachteile:**

- Hardware-Investition und Wartung
- Keine zentralen Updates
- Schwierige Multi-Standort-Verwaltung
- Hohe Einrichtungskosten

**Typische Einsatzbereiche:** Etablierte Restaurants mit stabiler IT, Umgebungen ohne zuverlässiges Internet.

### 3.2 Cloud-POS (SaaS)

```
┌────────────────────────────────────────────────────┐
│  Cloud                                             │
│  ┌──────────────────────────────────────────────┐  │
│  │  POS-Plattform                               │  │
│  │  (Multi-Tenant, API, DB, Reporting)          │  │
│  └──────────────────────────────────────────────┘  │
└────────────────────────────────────────────────────┘
           ↕ HTTPS                 ↕ HTTPS
┌──────────────────┐   ┌──────────────────────┐
│   iPad-Kasse     │   │  Smartphone-Service  │
│  (Browser/App)   │   │  (Browser/mPOS-App)  │
└──────────────────┘   └──────────────────────┘
```

**Vorteile:**

- Keine lokale Server-Infrastruktur
- Automatische Updates
- Zentrales Reporting über alle Standorte
- Schneller Einstieg (SaaS)

**Nachteile:**

- Abhängigkeit von Internetverbindung
- Monatliche Abo-Kosten
- Datenschutz: Transaktionsdaten beim Anbieter
- Kein oder begrenztes Offline-Modus

**Typische Einsatzbereiche:** Gastro-Ketten, Cafés, Einzelhändler ohne eigene IT.

### 3.3 Hybrid-POS

Hybrid-Systeme kombinieren lokale Verarbeitung mit Cloud-Synchronisation:

```
┌─────────────────────────────────────────┐
│  Lokal                                  │
│  ┌──────────┐   ┌─────────────────────┐ │
│  │  Kasse   │──→│  Lokaler Cache/DB   │ │
│  │ (Client) │   │  (Offline-fähig)    │ │
│  └──────────┘   └─────────┬───────────┘ │
└────────────────────────────┼────────────┘
                             │ Sync (wenn online)
                    ┌────────▼────────┐
                    │   Cloud-Backend │
                    │   (Reporting,   │
                    │   Backup, Multi-│
                    │   Standort)     │
                    └─────────────────┘
```

**Vorteile:**

- Offline-Betrieb möglich (lokaler Cache)
- Zentrale Verwaltung und Reporting
- Resilienz gegen Netzwerkausfälle

**Typische Einsatzbereiche:** Multi-Standort-Betriebe, Veranstaltungen ohne stabile Internetverbindung.

### 3.4 Mobile POS (mPOS)

mPOS nutzt Smartphones als vollwertige POS-Terminals:

```
┌───────────────────────────────────────────────────────┐
│  Smartphone-Browser (PWA)  oder  native App           │
│  ┌──────────────────────────────────────────────────┐ │
│  │  Bestellung aufnehmen → Backend-API aufrufen     │ │
│  └──────────────────────────────────────────────────┘ │
└───────────────────────────────────────────────────────┘
          ↕ HTTPS                     ↕ Bluetooth
┌──────────────────────┐   ┌──────────────────────────┐
│   Backend-Server     │   │  Kartenleser             │
│  (Go/Node/Rails...)  │   │  (SumUp, Zettle, Square) │
└──────────────────────┘   └──────────────────────────┘
```

**Vorteile:**

- Keine proprietäre Kassen-Hardware
- Servicekräfte nutzen eigene oder günstige Smartphones
- Schnelle Inbetriebnahme
- Mobile Bestellaufnahme direkt am Tisch

**Nachteile:**

- Kartenzahlungs-Hardware separat nötig (Bluetooth-Reader)
- Abhängigkeit von WLAN-Qualität
- Kleine Displays, eingeschränkte Ergonomie

**Typische Einsatzbereiche:** Kleinbetriebe, Vereinsveranstaltungen, Pop-up-Restaurants, Foodtrucks.

### 3.5 Self-Hosted POS

Self-Hosting kombiniert die Kontrolle eines On-Premise-Systems mit moderner Web-Technologie:

```
┌──────────────────────────────────────────────────────┐
│  Eigene VM / VPS                                     │
│                                                      │
│  ┌──────────────────────────────────────────────┐   │
│  │  Docker Compose                              │   │
│  │  ├── Reverse Proxy (nginx/Traefik/Caddy)    │   │
│  │  ├── Backend (Go/Node/...)                  │   │
│  │  └── Datenbank (PostgreSQL)                 │   │
│  └──────────────────────────────────────────────┘   │
│                                                      │
└──────────────────────────────────────────────────────┘
         ↕ HTTPS (Let's Encrypt)
┌────────────────────────────────────┐
│  Smartphones der Servicekräfte     │
│  (Browser, keine App-Installation) │
└────────────────────────────────────┘
```

**Vorteile:**

- Vollständige Datenkontrolle
- Keine laufenden SaaS-Kosten
- Anpassbar (Open Source)
- Kein Vendor Lock-in

**Nachteile:**

- Server-Infrastruktur und Betrieb selbst verantwortlich
- Updates müssen manuell oder per CI/CD eingespielt werden
- Kein professioneller Support

**Typische Einsatzbereiche:** Technisch affine Vereine und Non-Profit-Organisationen, Entwickler-Teams, spezifische Anwendungsfälle ohne passende kommerzielle Lösung.

---

## 4. Gastronomie-POS im Detail

### 4.1 Order Lifecycle im Restaurant

Der typische Lebenszyklus einer Restaurantbestellung:

```
Gast setzt sich hin
       │
       ▼
Bestellung aufnehmen (Servicekraft)
       │
       ▼
Bestellung an Küche übermitteln (Küchenbon / KDS)
       │
       ▼
Speisen zubereiten (Küche)
       │
       ▼
Lieferung bestätigen (Servicekraft oder Küche)
       │
       ▼
Weitere Bestellrunden (Getränke, Nachspeise...)
       │
       ▼
Zahlung anfordern (Gast)
       │
       ▼
Abrechnung (bar, Karte, getrennt, Rechnung)
       │
       ▼
Tisch freigeben / Reinigen
```

### 4.2 Tischbasierter Workflow

Der entscheidende Unterschied zwischen Einzel- und Restaurantkasse ist der **offene Tisch-Saldo**:

| Merkmal               | Einzelhandels-POS           | Gastronomie-POS                          |
| --------------------- | --------------------------- | ---------------------------------------- |
| **Transaktionsmodel** | Eine Transaktion = ein Kauf | Mehrere Bestellrunden auf einem Tisch    |
| **Saldo**             | Sofort abgerechnet          | Offener Saldo über mehrere Bestellrunden |
| **Rollen**            | Kassiererin                 | Service, Küche, Kassierer, Manager       |
| **Produkte**          | SKU-basiert, Barcode        | Menü-Items, Varianten, Modifikationen    |
| **Stornierung**       | Nachkauf-Rückgabe           | Positions-Stornierung im laufenden Tab   |

**Tischzustands-Modell:**

```
FREI → BESETZT → [BESTELLUNGEN...] → ABZURECHNEN → FREI
                        ↑
                   (Neue Runden)
```

Zugehörige Operationen je Zustand:

- **BESETZT**: Bestellung aufnehmen, Lieferung bestätigen, Stornierung
- **ABZURECHNEN**: Zahlung (ganz/teilweise), Rechnung drucken

### 4.3 Kitchen Display System (KDS)

Ein **Kitchen Display System** ersetzt den Küchenbon-Drucker durch ein digitales Display in der Küche:

```
Service-Tablet          KDS (Küchendisplay)
     │                         │
     │  Bestellung aufgeben    │
     │ ───────────────────────►│
     │                         │  Aufgabe erscheint
     │                         │  in der Queue
     │                         │
     │                         │  Koch bereitet zu
     │                         │
     │  Lieferung bestätigen   │
     │ ◄─────────────────────── │  „Ready" markiert
```

**KDS-Vorteile gegenüber Bondrucker:**

- Echtzeit-Statusanzeige (ausstehend / in Bearbeitung / fertig)
- Priorisierung nach Tisch und Wartezeit
- Kein Papierstau, kein Rollenwechsel
- Metriken: Durchschnittliche Zubereitungszeit, Rückstandsanzeige

**KDS-Implementierungsstrategien:**

1. **Polling**: KDS fragt regelmäßig neue Bestellungen ab (einfach, leichte Latenz)
2. **WebSocket/SSE**: Server-Push für Echtzeit-Updates (geringe Latenz, komplexer)
3. **LISTEN/NOTIFY (PostgreSQL)**: DB-seitige Events → Backend → KDS

### 4.4 Split Bills — Getrennte Abrechnung

Getrennte Abrechnung ist eine häufige Anforderung in der Gastronomie:

**Split-Strategien:**

- **Equal Split**: Gesamtbetrag durch Personenzahl
- **Item Split**: Jede Person zahlt ihre Positionen
- **Custom Split**: Beliebige Aufteilung (z. B. Paare)
- **Teilzahlung**: Betrag X jetzt, Rest später

**Datenmodell für Split:**

```
Tisch hat offene Positionen
  ├── Position A (Hauptgericht Gast 1)  → Zahlung 1
  ├── Position B (Hauptgericht Gast 2)  → Zahlung 2
  ├── Position C (Flasche Wein)          → Zahlung 1+2 (geteilt)
  └── Position D (Espresso)              → Zahlung 1
```

Herausforderung: Positionen können auf mehrere Zahlungen aufgeteilt werden. Das Datenmodell muss **atomare Zuweisung** sicherstellen und **Doppelzahlung** verhindern.

### 4.5 Floor Management

Größere Restaurants nutzen **Floor Management** (auch: Table Map):

- Visuelle Darstellung des Restaurants
- Echtzeit-Tischstatus (frei / besetzt / Reservierung)
- Drag-and-Drop für Tischzusammenlegung
- Kapazitätsplanung

**Reservierungssystem-Integration:**
POS und Reservierungssystem teilen die Tischbelegungs-Daten. Typische Integrations-APIs: OpenTable, Resy, TheFork.

---

## 5. Datenmodelle

### 5.1 Transaktionsbasiertes Modell (Klassisch)

Das klassische Datenmodell speichert nur den **aktuellen Zustand**:

```sql
-- Tisch-Saldo
CREATE TABLE tische (
  id          UUID PRIMARY KEY,
  name        VARCHAR(50),
  status      VARCHAR(20),  -- 'frei', 'besetzt', 'abzurechnen'
  saldo_cent  INTEGER DEFAULT 0
);

-- Bestellungen (Mutable: werden bei Stornierung geupdated)
CREATE TABLE bestellungen (
  id           UUID PRIMARY KEY,
  tisch_id     UUID REFERENCES tische(id),
  produkt_id   UUID REFERENCES produkte(id),
  menge        INTEGER,
  preis_cent   INTEGER,
  status       VARCHAR(20),  -- 'offen', 'geliefert', 'bezahlt', 'storniert'
  erstellt_am  TIMESTAMPTZ
);
```

**Vorteile:** Einfach, geringe Datenmenge.

**Nachteile:**

- Keine Nachvollziehbarkeit (wer hat was wann geändert?)
- Stornierungen und Korrekturen verändern historische Daten
- Manipulationsanfällig (fiscalrechtlich problematisch)
- Keine natürliche Audit-Trail

### 5.2 Journal-basiertes Modell

Viele POS-Systeme nutzen ein **Kassenjournal** — ein append-only Log aller Transaktionen:

```sql
CREATE TABLE kassenjournal (
  id            BIGSERIAL PRIMARY KEY,
  tisch_id      UUID,
  typ           VARCHAR(50),  -- 'bestellung', 'zahlung', 'stornierung', 'lieferung'
  betrag_cent   INTEGER,
  payload       JSONB,        -- typ-spezifische Daten
  erstellt_am   TIMESTAMPTZ,
  bediener_id   UUID
);
```

Der aktuelle Tisch-Saldo wird durch Aggregation (SUM über alle Einträge) berechnet.

**Vorteile:**

- Vollständiger Audit-Trail
- Manipulationssicher (kein DELETE/UPDATE)
- Basis für Kassenberichte

**Nachteile:**

- Saldo-Berechnung via Aggregation (teuer bei großem Journal)
- Optimierung durch Snapshots notwendig

### 5.3 Event-Sourcing-Modell

Das vollständige Event-Sourcing-Modell speichert alle Zustandsänderungen als **immutable Events**:

```
Subject: "table:42"

Event 1: table.order-placed:v1
  { positionen: [{produkt_id: "...", menge: 2, preis: 1200}] }

Event 2: table.order-placed:v1
  { positionen: [{produkt_id: "...", menge: 1, preis: 800}] }

Event 3: table.items-delivered:v1
  { positionen: [{produkt_id: "...", menge: 2}] }

Event 4: table.payment-registered:v1
  { betrag: 2000 }

Event 5: table.items-cancelled:v1
  { positionen: [{produkt_id: "...", menge: 1}] }
```

**State-Rekonstruktion:**

```
Saldo            = Σ(Bestellungen) − Σ(Zahlungen) − Σ(Stornierungen)
UnbezahltPos     = bestellt − bezahlt − storniert
UngeliefertPos   = bestellt − geliefert − storniert
```

**Snapshot-Optimierung:** Bei langen Event-Historien wird ein **Snapshot** gespeichert:

```
Event 1-100: [ältere Events]
Snapshot @100: { saldo: 15200, offene_positionen: [...] }
Event 101-110: [neuere Events seit Snapshot]

→ State = Snapshot + Replay(101..110)
```

**Vergleich der Modelle:**

| Eigenschaft              | Transaktional | Journal-basiert  | Event-Sourcing      |
| ------------------------ | ------------- | ---------------- | ------------------- |
| **Manipulationsschutz**  | ❌            | ✅ (append-only) | ✅ (immutable)      |
| **Audit-Trail**          | ❌            | ✅ (partiell)    | ✅ (vollständig)    |
| **Komplexität**          | Gering        | Mittel           | Hoch                |
| **Query-Performance**    | Hoch          | Mittel           | Erfordert Snapshots |
| **Fiskalkonformität**    | Problematisch | Geeignet         | Ideal               |
| **State-Rekonstruktion** | Direkt        | Via Aggregation  | Via Replay          |

### 5.4 Produktkatalog und Varianten

```
Produkt
├── id, name, kategorie, aktiv
├── preis_cent (Basispreis)
└── Varianten
    ├── Variante A: "Klein" → Aufpreis: 0
    ├── Variante B: "Groß" → Aufpreis: 100
    └── Variante C: "Pitcher" → Aufpreis: 500

Modifikatoren (optional)
├── "ohne Zwiebeln"
├── "extra scharf"
└── "mit Sauce extra"
```

**Herausforderung: Preisänderungen**

Wenn ein Produkt seinen Preis ändert, müssen alle historischen Bestellungen noch den alten Preis enthalten. Zwei Strategien:

1. **Preis bei Bestellung einfrieren**: Der Preis wird zum Bestellzeitpunkt in der Position gespeichert (Snapshot-Preis).
2. **Preisversionen**: Jedes Produkt hat eine Preis-Historie mit Gültigkeitszeitraum.

Empfehlung: **Preis-Snapshot** bei Bestellaufnahme (einfacher, klarer Audit-Trail).

---

## 6. Fiskalgesetzgebung

### 6.1 Überblick und Regulierungsziele

Fiskalgesetzgebung für Kassensysteme verfolgt primär **steuerrechtliche Ziele**:

1. **Manipulationsschutz**: Nachträgliche Änderungen von Buchungen verhindern
2. **Revisionssicherheit**: Alle Transaktionen müssen nachvollziehbar sein
3. **Kassenjournal-Pflicht**: Lückenlose Aufzeichnung aller Vorgänge
4. **Aufbewahrungspflicht**: Daten für mehrere Jahre verfügbar halten

Verschiedene Länder haben unterschiedliche Ansätze zur technischen Umsetzung.

### 6.2 Deutschland: KassenSichV und TSE

**Kassensicherungsverordnung (KassenSichV)**, in Kraft seit 2020, schreibt für elektronische Kassensysteme vor:

- **Technische Sicherheitseinrichtung (TSE)**: Hardware- oder Cloud-Modul, das jede Transaktion kryptografisch signiert
- **Transaktionsnummern**: Fortlaufende, lückenlose Nummerierung
- **Exportformat (DSFinV-K)**: Standardisiertes Exportformat für Betriebsprüfungen
- **Meldepflicht**: Kassensysteme müssen beim Finanzamt gemeldet werden

**TSE-Anforderungen:**

```
Transaktion
  ├── Inhalt (Positionen, Betrag, Zeit)
  ├── Transaktionsnummer (fortlaufend, nicht manipulierbar)
  ├── Signatur (kryptografische Signatur der TSE)
  └── TSE-Seriennummer

TSE speichert:
  ├── Signaturzähler (tamper-evident)
  ├── Zeitstempel (von vertrauenswürdiger Quelle)
  └── Signaturen aller Transaktionen
```

**TSE-Typen:**

- **Hardware-TSE**: USB-Stick oder SD-Karte mit Secure Element (z. B. Swissbit, Epson)
- **Cloud-TSE**: TSE-Dienst über das Internet (z. B. Deutsche Fiskal, fiskaly)

**Anbieter zertifizierter TSE-Lösungen:**

- Swissbit TSE (Hardware)
- Epson TSE (Hardware, in Bondrucker integriert)
- fiskaly (Cloud-TSE)
- Deutsche Fiskal / Bundesdruckerei (Cloud-TSE)

### 6.3 Österreich: RKSV

Die österreichische **Registrierkassensicherheitsverordnung (RKSV)**, in Kraft seit 2016, nutzt einen ähnlichen Ansatz:

- **Signaturerstellungseinheit (SEE)**: Entspricht der deutschen TSE
- **Kassenidentifikationsnummer**: Eindeutige Kennung je Kasse
- **Startbeleg**: Kasse wird zu Betriebsbeginn mit einem signierten Startbeleg registriert
- **Jahresbeleg**: Jährliche Prüfung per FinanzOnline
- **Belegaustauschformat**: Maschinenlesbare Quittungen (QR-Code)

### 6.4 Non-Profit-Ausnahmen

Non-Profit-Organisationen und Vereine sind in der Regel **von der Registrierkassenpflicht befreit** oder haben vereinfachte Regelungen, wenn:

- **Deutschland**: Jahresumsatz < 17.500 EUR (Kleinunternehmer, §19 UStG) oder ausschließlich steuerfreie Umsätze
- **Vereinsfeste**: Gelegentliche Veranstaltungen mit vereinfachten Buchführungsregeln (§ 64 Abs. 5 AO — Vereinfachung für satzungszweckgemäßen Geschäftsbetrieb)
- **Österreich**: Vereine mit Jahreseinnahmen < 15.000 EUR sind von der RKSV befreit

**Wichtig:** Diese Ausnahmen müssen im Einzelfall geprüft werden. Steuerliche Beratung ist empfohlen.

### 6.5 GoBD (Deutschland)

Die **Grundsätze zur ordnungsmäßigen Führung und Aufbewahrung von Büchern, Aufzeichnungen und Unterlagen in elektronischer Form (GoBD)** regeln ergänzend:

- **Unveränderbarkeit**: Einmal gebuchte Daten dürfen nicht gelöscht oder nachträglich verändert werden
- **Protokollierung**: Jede Änderung muss nachvollziehbar protokolliert werden
- **Aufbewahrungsfrist**: 10 Jahre für steuerlich relevante Unterlagen
- **Exportierbarkeit**: Daten müssen für Prüfer exportierbar sein (maschinelle Auswertbarkeit)

Event-Sourcing-basierte POS-Systeme erfüllen GoBD-Anforderungen natürlich: Immutable Events sind inhärent unveränderlich und vollständig protokolliert.

---

## 7. Payment Integration

### 7.1 Payment-Architektur-Überblick

```
┌──────────────────────────────────────────────────────────────────┐
│  POS-System                                                      │
│                                                                  │
│  ┌────────────────┐       ┌──────────────────────────────────┐  │
│  │  Kassensoftware│──────►│  Payment Service Layer           │  │
│  └────────────────┘       │  (Abstrahiert Payment-Anbieter)  │  │
│                           └───────────────┬──────────────────┘  │
└───────────────────────────────────────────┼─────────────────────┘
                                            │
              ┌─────────────────────────────┼──────────────────────┐
              │                             │                      │
              ▼                             ▼                      ▼
    ┌─────────────────┐          ┌──────────────────┐   ┌────────────────┐
    │  SumUp          │          │  Zettle (PayPal) │   │  Stripe        │
    │  (Kartenleser)  │          │  (Kartenleser)   │   │  (Online)      │
    └────────┬────────┘          └────────┬─────────┘   └───────┬────────┘
             │                            │                      │
             ▼                            ▼                      ▼
    ┌────────────────────────────────────────────────────────────────────┐
    │              Acquiring Bank / Payment Network                      │
    │              (Visa, Mastercard, SEPA, etc.)                        │
    └────────────────────────────────────────────────────────────────────┘
```

### 7.2 Kartenleser-APIs für mPOS

Die drei wichtigsten mPOS-Anbieter in Europa:

#### SumUp

- **Zielgruppe**: Kleinunternehmer, Marktstand, Events
- **Hardware**: SumUp Air, SumUp Solo (Bluetooth/USB)
- **API**: REST-API für Transaktionsmanagement, Produkt-Einrichtung
- **SDK**: iOS, Android für native App-Integration
- **Gebühren**: 1,69 % pro Transaktion (Stand 2024, DE)

```
SumUp-Zahlungsablauf (SDK):
1. POS erstellt Checkout über SumUp-API (Betrag, Beschreibung)
2. SDK leitet Zahlung an Kartenleser weiter
3. Kunde steckt/tippt Karte auf Reader
4. Reader verarbeitet Zahlung (Online-Autorisierung)
5. SumUp-API sendet Webhook: Zahlung erfolgreich/fehlgeschlagen
6. POS bucht Zahlung als erfolgreich
```

#### Zettle by PayPal

- **Zielgruppe**: Einzel- und Kleinhändler, Foodservice
- **Hardware**: Zettle Reader 2 (Bluetooth)
- **API**: Zettle Go SDK, REST-API
- **Integration**: PayPal-Ökosystem, Shopify, WooCommerce

#### Square

- **Zielgruppe**: US-fokussiert, international verfügbar
- **Hardware**: Square Reader (Audio, Lightning, Bluetooth)
- **API**: Square Connect API — umfangreichste POS-API
- **Besonderheit**: Vollständiges POS-Ökosystem (Inventory, Customers, Analytics)

### 7.3 PCI-DSS Compliance

**Payment Card Industry Data Security Standard (PCI-DSS)** ist ein Sicherheitsstandard für alle Systeme, die Kreditkartendaten verarbeiten.

**PCI-DSS Level-System:**

| Level | Transaktionen/Jahr | Anforderung                                       |
| ----- | ------------------ | ------------------------------------------------- |
| 1     | > 6 Mio.           | Jährliches Audit durch QSA (externe Prüfer)       |
| 2     | 1–6 Mio.           | Jährlicher SAQ + vierteljährl. Vulnerability Scan |
| 3     | 20.000–1 Mio.      | Jährlicher SAQ + vierteljährl. Vulnerability Scan |
| 4     | < 20.000           | Jährlicher SAQ                                    |

**Wichtigstes Prinzip:** Kartendaten **nie selbst speichern**. Durch Nutzung eines Payment-Gateways (SumUp, Stripe, Zettle) übernimmt der Anbieter die PCI-Compliance für die Kartendatenverarbeitung.

**SAQ = Self-Assessment Questionnaire:** Selbst-Audit-Formular für Händler ohne eigene Kartenverarbeitung.

### 7.4 NFC und Contactless Payments

**Near Field Communication (NFC)** ist die Technologie hinter Contactless-Zahlungen:

```
Gast          Kartenleser         Acquiring Bank
  │               │                     │
  │ Karte/Phone   │                     │
  │ nähert sich   │                     │
  │ ─────────────►│                     │
  │               │  ISO 14443 / ISO 7816│
  │               │  NFC-Handshake      │
  │               │  (< 4 cm)           │
  │               │                     │
  │               │  Authorization Req  │
  │               │ ────────────────────►
  │               │  Approval           │
  │               │ ◄────────────────────
  │               │                     │
  │  Bestätigung  │                     │
  │◄──────────────│                     │
```

**Aktuelle Standards:**

- **EMV Contactless**: Europäischer Standard (Chip + PIN für > 50 EUR)
- **Apple Pay / Google Pay**: Tokenisierte NFC-Zahlung via Wallet
- **Limit ohne PIN**: Je nach Bank/Land 50–100 EUR

---

## 8. Non-Profit vs. Commercial POS

### 8.1 Anforderungsvergleich

| Anforderung           | Kommerziell (Dauerbetrieb)       | Non-Profit (Gelegentliche Events) |
| --------------------- | -------------------------------- | --------------------------------- |
| **Verfügbarkeit**     | 365/24/7                         | Nur während Events                |
| **Skalierung**        | Wachsend, Multi-Standort         | Feste, kleine Teams               |
| **Fiskalkonformität** | TSE, GoBD, volle Compliance      | Oft vereinfacht / ausgenommen     |
| **Payment**           | Karte, NFC, EC, QR-Code          | Oft nur Bargeld                   |
| **Inventar**          | Echtzeit-Lagerverwaltung         | Nicht relevant                    |
| **CRM/Loyalty**       | Kundenprofile, Bonuspunkte       | Nicht relevant                    |
| **Reporting**         | Umsatz, Trends, Prognosen        | Tagesumsatz, Abrechnung           |
| **Hardware**          | Kassen-Terminal, Bondrucker, KDS | Smartphones (BYOD)                |
| **Setup-Aufwand**     | Wochen–Monate                    | Stunden                           |
| **Support**           | SLA, 24/7-Hotline                | Community / Selbsthilfe           |

### 8.2 Total Cost of Ownership (TCO)

**Kommerzielles POS-System (beispielhaft: 1 Kasse, 1 Jahr, DE):**

| Position                       | Kosten/Jahr            |
| ------------------------------ | ---------------------- |
| Software-Abo (z. B. Orderbird) | 600–1.200 EUR          |
| Hardware (iPad + Halterung)    | 400–800 EUR (einmalig) |
| Bondrucker                     | 150–400 EUR (einmalig) |
| TSE (Cloud-TSE Jahresgebühr)   | 60–120 EUR             |
| Kartenleser                    | 30–80 EUR (einmalig)   |
| Transaktionsgebühren (1,5–2%)  | je nach Umsatz         |
| **Gesamt (laufend)**           | ~700–1.400 EUR/Jahr    |

**Self-hosted Open-Source-POS:**

| Position                            | Kosten/Jahr      |
| ----------------------------------- | ---------------- |
| VPS (kleinste Instanz)              | 60–100 EUR       |
| Domain + TLS (Let's Encrypt)        | 10–15 EUR        |
| Entwicklung/Wartung (Eigenleistung) | — (ehrenamtlich) |
| **Gesamt (laufend)**                | ~70–115 EUR/Jahr |

### 8.3 Feature-Abgrenzung: Was Non-Profit-POS braucht

Ein Non-Profit-POS für Vereinsveranstaltungen benötigt:

**✅ Kern-Features:**

- Produkte und Kategorien verwalten
- Tischbasierte Bestellaufnahme per Smartphone
- Lieferung bestätigen
- Zahlungen (Bargeld) registrieren, Teilzahlungen
- Stornierungen (privilegiert: Serviceleitung)
- Tisch-Saldo und Kassenjournal
- Rollenverwaltung (Admin, Service, Serviceleitung)

**❌ Nicht nötig:**

- Kartenleserzahlung (oft aus Kostengründen)
- Inventarverwaltung
- CRM und Kundenbindung
- KDS (Küchenanzeige)
- Bon-Druck (optional, aber kein Muss)
- Multi-Tenant

**❌ Nicht relevant:**

- TSE/KassenSichV (Non-Profit-Ausnahme)
- Jahresabschluss, DATEV-Export

---

## 9. POS-Marktlandschaft

### 9.1 Kommerzielle Systeme

#### Tablet/iPad-POS (Europa)

| System         | Fokus             | Besonderheit                               | Preis (ca.)     |
| -------------- | ----------------- | ------------------------------------------ | --------------- |
| **Orderbird**  | Gastronomie (DE)  | iPad-POS, DATEV-Integration, TSE inklusive | ab 49 EUR/Monat |
| **Lightspeed** | Restaurant/Retail | Vollständiges System, Multi-Standort       | ab 69 EUR/Monat |
| **Gastrofix**  | Gastronomie (DE)  | Fokus auf Deutsche Steuergesetzgebung      | Auf Anfrage     |
| **Flyt POS**   | Gastronomie (UK)  | Open API, Lieferdienst-Integration         | Auf Anfrage     |

#### Amerikanische Systeme (int. verfügbar)

| System                     | Fokus                    | Besonderheit                             |
| -------------------------- | ------------------------ | ---------------------------------------- |
| **Toast**                  | Restaurants (USA)        | Android-basiert, umfangreiches Ökosystem |
| **Square for Restaurants** | Gastronomie              | Einfacher Einstieg, günstig              |
| **Revel Systems**          | Enterprise Restaurant    | iPad-POS, Franchise-Fähig                |
| **NCR Aloha**              | Gastronomie (Enterprise) | Marktführer USA, sehr umfangreich        |

### 9.2 Open-Source-Systeme

| System                | Sprache/Stack  | Fokus                 | Besonderheit                        |
| --------------------- | -------------- | --------------------- | ----------------------------------- |
| **UniCenta oPOS**     | Java           | Einzelhandel + Gastro | Aktives Projekt, freie Installation |
| **Floreant POS**      | Java           | Restaurant            | Fokus auf Restaurant-Workflows      |
| **OpenBravo POS**     | Java           | Einzelhandel          | Veraltet, kaum Weiterentwicklung    |
| **OpenPOS**           | Angular/Spring | Enterprise            | Moderner Tech-Stack                 |
| **GNU Cash Register** | Diverse        | Basic                 | Sehr einfach, Community-Projekt     |

**Bewertung Open-Source-Systeme:**

- Meist Java-basiert, ältere Tech-Stacks
- Installation erfordert technisches Know-how
- Fehlende moderne mPOS-Unterstützung
- Selten spezialisiert auf Non-Profit/Vereinsveranstaltungen

### 9.3 Self-Order und Kiosk-Systeme

**Self-Order Kiosk** ermöglicht Gästen, Bestellungen selbst aufzugeben:

```
Touchscreen-Terminal am Tisch oder Eingang
  ├── Menü durchsuchen
  ├── Bestellung zusammenstellen
  ├── Zahlung (Karte/NFC/QR)
  └── Bon (digital / gedruckt)
```

Bekannte Systeme: McDonald's, Burger King (proprietär), Lightspeed K-Series (Kiosk), Square Kiosk.

### 9.4 Cloud-Kitchen und Delivery-Integration

**Cloud-Kitchens** (auch Ghost Kitchens) sind reine Produktionsküchen ohne Gastraum — optimiert für Lieferdienste:

```
POS-System
  └── Delivery-Integration
      ├── UberEats API
      ├── Lieferando API
      ├── Deliveroo API
      └── Eigene Bestell-Website
```

Herausforderung: Bestellungen aus mehreren Kanälen (Lieferdienst + Vor-Ort) in einem System aggregieren.

---

## 10. Entscheidungsmatrix: POS-Architektur nach Anwendungsfall

| Kriterium                       | On-Premise | Cloud-POS | Hybrid | mPOS/Self-Hosted |
| ------------------------------- | ---------- | --------- | ------ | ---------------- |
| **Internet-unabhängig**         | ✅         | ❌        | ✅     | ❌               |
| **Geringe Initialkosten**       | ❌         | ✅        | ❌     | ✅               |
| **Keine laufenden SaaS-Kosten** | ✅         | ❌        | ❌     | ✅               |
| **Automatische Updates**        | ❌         | ✅        | ✅     | ❌               |
| **Keine Hardware nötig**        | ❌         | Teilw.    | Teilw. | ✅               |
| **Vollständige Datenkontrolle** | ✅         | ❌        | Teilw. | ✅               |
| **Multi-Standort**              | ❌         | ✅        | ✅     | ❌               |
| **TSE/KassenSichV bereit**      | ✅         | ✅        | ✅     | Offen            |
| **Einfache Einrichtung**        | ❌         | ✅        | ❌     | Mittel           |
| **Ideal für Vereinsevents**     | ❌         | ❌        | ❌     | ✅               |

**Empfehlungen:**

- **Dauergastronomie (Restaurant)**: Cloud-POS oder Hybrid → Orderbird, Lightspeed, Toast
- **Einzelhändler**: Cloud-POS → Square, Lightspeed Retail
- **Enterprise / Multi-Standort**: On-Premise oder Hybrid → NCR, Revel, Lightspeed Enterprise
- **Vereinsveranstaltungen / Non-Profit**: Self-Hosted mPOS → Open-Source-Lösungen oder eigene Entwicklung
- **Foodtruck / Marktstand**: Cloud-mPOS → Square, Zettle, SumUp mit App

---

## 11. Referenzen

### POS-Geschichte & Grundlagen

1. [Wikipedia: Point of Sale](https://en.wikipedia.org/wiki/Point_of_sale) — Geschichte, Terminologie, Systemtypen
2. [Wikipedia: Cash Register](https://en.wikipedia.org/wiki/Cash_register) — Entstehung der Registrierkasse (James Ritty, NCR)
3. [Square Developer Docs](https://developer.squareup.com/) — Payment Integration, POS-API-Design, mPOS-Patterns
4. [Toast Developer Platform](https://doc.toasttab.com/) — Restaurant POS API, KDS-Integration, Order Lifecycle

### Fiskalgesetzgebung

5. [KassenSichV — Kassensicherungsverordnung](https://dejure.org/gesetze/KassenSichV) — TSE-Anforderungen, Manipulationsschutz
6. [RKSV — Registrierkassensicherheitsverordnung Österreich](https://www.ris.bka.gv.at/GeltendeFassung.wxe?Abfrage=Bundesnormen&Gesetzesnummer=20009390) — Registrierkassenpflicht, Ausnahmen für Vereine
7. [Abgabenordnung (AO)](https://dejure.org/gesetze/AO) — Rechtsgrundlage für GoBD (Grundsätze ordnungsmäßiger Buchführung)

### Payment Integration

8. [SumUp Developer Docs](https://developer.sumup.com/) — SumUp API, Kartenleser-SDK
9. [Zettle Developer Portal](https://developer.zettle.com/) — Zettle SDK, Payment-Integration
10. [PCI Security Standards Council](https://www.pcisecuritystandards.org/) — PCI-DSS Anforderungen, SAQ-Formulare
11. [EMVCo Contactless Specifications](https://www.emvco.com/emv-technologies/contactless/) — EMV Contactless, NFC-Standards

### Marktlandschaft

12. [Orderbird](https://www.orderbird.com) — Gastronomie-POS (DE), iPad-basiert
13. [Lightspeed Restaurant](https://www.lightspeedhq.com) — Vollständiges Restaurantsystem
14. [Toast POS](https://pos.toasttab.com) — Gastronomie-POS (v.a. USA)
15. [Square for Restaurants](https://squareup.com/gb/en/restaurants) — Flexibles Gastro-POS
16. [UniCenta oPOS](https://unicenta.com) — Open-Source POS (Einzelhandel/Gastronomie)
17. [Floreant POS](https://floreant.org) — Open-Source Restaurant-POS

### Weiterführende Literatur

- **PCI SSC**: _PCI DSS v4.0 Requirements and Testing Procedures_ — vollständige PCI-Anforderungen
- **Martin Fowler**: _Patterns of Enterprise Application Architecture_ — Transaktionsmodelle, Journal-Patterns
- **Greg Young**: _CQRS Documents_ — Event-Sourcing als Grundlage für Kassenjournale
