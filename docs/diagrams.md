# Diagramme — jotti

Visuelle Dokumentation der Fachlichkeit (DDD) und Architektur von jotti.
Alle Diagramme bilden den **aktuellen Implementierungsstand** ab. Geplante Features (TSE, DSFinV-K, Betreiber-Stammdaten) sind mit _(geplant)_ gekennzeichnet.

> **Referenzen:** [anforderungen.md](anforderungen.md) · [handbuch.md](handbuch.md) · [language.md](language.md) · [roadmap.md](roadmap.md) · [compliance.md](compliance.md)

---

## Inhaltsverzeichnis

1. [Systemkontext](#1--systemkontext)
2. [Bounded Context Map](#2--bounded-context-map)
3. [Rollen und Berechtigungen](#3--rollen-und-berechtigungen)
4. [Kassenbetrieb — Tisch-Aggregat (Zustandsdiagramm)](#4--kassenbetrieb--tisch-aggregat-zustandsdiagramm)
5. [Kassenbetrieb — Domain Events und Saldo-Fluss](#5--kassenbetrieb--domain-events-und-saldo-fluss)
6. [Kassenbetrieb — Bestellvorgang (Sequenz)](#6--kassenbetrieb--bestellvorgang-sequenz)
7. [Stammdaten — Domain Model](#7--stammdaten--domain-model)
8. [Kassenführung — Lifecycle](#8--kassenführung--lifecycle)
9. [Kassenführung — Kassenbestand-Berechnung](#9--kassenführung--kassenbestand-berechnung)
10. [Auth und Onboarding](#10--auth-und-onboarding)
11. [Schichtenarchitektur (Backend)](#11--schichtenarchitektur-backend)
12. [Event Sourcing: Synchrone Projektion + CRUD-Entität](#12--event-sourcing-synchrone-projektion--crud-entität)
13. [Bondruck — Relay-Architektur](#13--bondruck--relay-architektur)
14. [Deployment-Architektur](#14--deployment-architektur)
15. [API-Bereichsgliederung](#15--api-bereichsgliederung)
16. [Fiskalkonformität — Compliance-Architektur](#16--fiskalkonformität--compliance-architektur)
17. [Datenbank-Schema (ER-Diagramm)](#17--datenbank-schema-er-diagramm)
18. [Frontend-Seitenstruktur](#18--frontend-seitenstruktur)

---

## 1 · Systemkontext

Wer interagiert mit jotti? Welche externen Systeme sind angebunden?

```mermaid
C4Context
    title Systemkontext — jotti mPOS

    Person(servicekraft, "Servicekraft", "Nimmt Bestellungen auf, bestätigt Ausgaben, kassiert — auf eigenem Smartphone (BYOD)")
    Person(serviceleitung, "Serviceleitung", "Wie Servicekraft, zusätzlich Stornierung und Auszahlung")
    Person(admin, "Admin / Kassenwart", "Verwaltet Produkte, Tische, Benutzer, Kassenführung, Reporting")

    System(jotti, "jotti", "Mobile-POS-System für Vereinsfeste — Self-hosted, Docker Compose, Mobile-first Web-App")

    System_Ext(tse, "Cloud-TSE (fiskaly)", "Zertifizierte Technische Sicherheitseinrichtung — signiert Transaktionen nach KassenSichV")
    System_Ext(drucker, "ESC/POS-Bondrucker", "Ethernet-Thermodrucker an Ausgabestationen (Küche, Theke)")
    System_Ext(elster, "ELSTER", "Elektronische Steuererklärung — Meldepflicht für Kassensysteme nach § 146a AO")
    System_Ext(browser, "Smartphone-Browser", "BYOD — Chrome, Safari, Firefox auf privaten Geräten der Helfer")

    Rel(servicekraft, jotti, "Bestellen, Ausgabe, Kassieren", "HTTPS / Browser")
    Rel(serviceleitung, jotti, "Stornieren, Auszahlen", "HTTPS / Browser")
    Rel(admin, jotti, "Admin-Verwaltung, Reporting, Kassenführung", "HTTPS / Browser")
    Rel(jotti, tse, "TSE-Transaktionen signieren", "REST API")
    Rel(jotti, drucker, "Bons drucken via Print-Relay", "ESC/POS über TCP")
    Rel(admin, elster, "Kassensystem melden (manuell oder ERiC)", "Webportal / API")
    Rel(servicekraft, browser, "Öffnet Web-App")
```

---

## 2 · Bounded Context Map

Die drei Bounded Contexts von jotti und ihre Beziehungen nach DDD-Patterns.

```mermaid
graph TB
    subgraph auth ["🔐 Auth<br/><i>Generic Sub-Domain</i>"]
        direction TB
        auth_desc["Login · Logout · Passwort-Management<br/>JWT-Token mit Benutzer-ID + Rolle<br/><b>Persistenz:</b> Infrastruktur"]
    end

    subgraph stammdaten ["📋 Stammdaten<br/><i>Supporting Sub-Domain</i>"]
        direction TB
        stamm_desc["Produkte · Tische · Benutzer<br/><b>Persistenz:</b> CRUD + Soft-Delete"]
    end

    subgraph kasse ["💰 Kasse<br/><i>Core Domain</i>"]
        direction TB
        kasse_desc["Bestellen · Ausgabe · Kassieren · Stornieren · Auszahlen<br/>Kassensitzung · Kassenbewegungen · Kassensturz · Z-Bon<br/>Kassenjournal (Event-Sourcing)<br/>Projektion: tisch_sessions · CRUD-Entität: kassensitzungen"]
    end

    stammdaten -->|"Customer/Supplier + ACL<br/>Fat Events frieren Produktdaten ein"| kasse
    auth -->|"Open Host Service<br/>JWT (Benutzer-ID + Rolle)"| kasse
    auth -->|"Open Host Service<br/>JWT (Benutzer-ID + Rolle)"| stammdaten

    style kasse fill:#e8f5e9,stroke:#2e7d32,stroke-width:3px
    style stammdaten fill:#e3f2fd,stroke:#1565c0,stroke-width:2px
    style auth fill:#f3e5f5,stroke:#6a1b9a,stroke-width:2px
```

---

## 3 · Rollen und Berechtigungen

Welche Rollen gibt es und was dürfen sie?

```mermaid
graph LR
    subgraph rollen ["Rollenhierarchie (inklusiv)"]
        admin["🔑 Admin<br/><code>admin</code>"]
        sl["👔 Serviceleitung<br/><code>serviceleitung</code>"]
        sk["🍽️ Servicekraft<br/><code>service</code>"]
        admin -->|"kann alles was"| sl
        sl -->|"kann alles was"| sk
    end

    subgraph bereiche ["Berechtigungsbereiche"]
        stamm["📋 Stammdaten<br/>Produkte · Tische · Benutzer"]
        basis["💰 Kasse (Basis)<br/>Bestellen · Ausgabe · Kassieren<br/>Tischübersicht · Kassenjournal"]
        erweitert["⚠️ Kasse (Erweitert)<br/>Stornieren · Auszahlen"]
        kf["📊 Kasse — Kassensitzung<br/>Kassensitzung eröffnen · Kassenbestand<br/>Kassensturz · Z-Bon<br/>Kassenbewegungen"]
        reporting["📈 Reporting<br/>Tagesabrechnung · Datenexport<br/>DSFinV-K-Export"]
    end

    admin --- stamm
    admin --- kf
    admin --- reporting
    sl --- erweitert
    sk --- basis

    style admin fill:#ffcdd2,stroke:#b71c1c
    style sl fill:#fff9c4,stroke:#f57f17
    style sk fill:#c8e6c9,stroke:#2e7d32
```

---

## 4 · Kasse — Tisch-Session (Zustandsdiagramm)

Die Tisch-Session (Abrechnungskreis) ist das Event-Sourced-Aggregat im Kasse-Kontext. Entsteht implizit mit der ersten Bestellung innerhalb einer Kassensitzung. Subject: `kassensitzung-{nr}/tisch-{id}`.

```mermaid
stateDiagram-v2
    [*] --> Leer: Erste Bestellung für<br/>Tisch in dieser Kassensitzung

    Leer --> HatBestellungen: BestellungAufgenommen<br/>Saldo += Σ Positionen

    HatBestellungen --> HatBestellungen: BestellungAufgenommen<br/>Saldo += Σ Positionen

    HatBestellungen --> HatBestellungen: AusgabeBestaetigt<br/>Positionen als ausgegeben markiert<br/>(kein Saldo-Effekt)

    HatBestellungen --> HatBestellungen: ZahlungKassiert<br/>Saldo -= Σ gewählte Positionen

    HatBestellungen --> HatBestellungen: StornierungErteilt<br/>Saldo -= Σ stornierte Positionen<br/>(kann negativ werden)

    HatBestellungen --> HatBestellungen: AuszahlungGeleistet<br/>Saldo += Betrag<br/>(gleicht neg. Saldo aus)

    HatBestellungen --> Leer: Alle Positionen bezahlt/storniert<br/>Saldo = 0

    note right of HatBestellungen
        TischSession (Projektion):
        • subject (PK)
        • tisch_id, kassensitzung_nr
        • saldo_cents
        • unbezahlte_positionen[]
        • ausstehende_positionen[]
        • gesamt_zahlungen_cents
        • last_event_version (OCC)
    end note
```

---

## 5 · Kasse — Domain Events

Alle Event-Typen des Kasse-Kontexts. Tisch-Events verändern den Saldo einer Tisch-Session, Kassensitzung-Events steuern den Betriebstag.

### Tisch-Session-Events (Subject: `kassensitzung-{nr}/tisch-{id}`)

```mermaid
graph TD
    subgraph events ["Tisch-Session Events (append-only)"]
        e1["<b>BestellungAufgenommen</b><br/>Positionen[] mit Fat-Event-Daten<br/>GesamtPreisCents"]
        e2["<b>AusgabeBestaetigt</b><br/>PositionRef[] (ID + Menge)<br/>— kein Saldo-Effekt —"]
        e3["<b>ZahlungKassiert</b><br/>PositionRef[] (ID + Menge)<br/>GesamtZahlungCents"]
        e4["<b>StornierungErteilt</b><br/>PositionRef[] (ID + Menge)<br/>GesamtStornierungCents<br/>Kommentar (Pflicht)"]
        e5["<b>AuszahlungGeleistet</b><br/>BetragCents (frei)<br/>Kommentar (Pflicht)<br/>— kein Positionsbezug —"]
    end

    subgraph saldo ["Saldo-Formel"]
        formel["<b>Saldo = Σ Bestellungen − Σ Zahlungen − Σ Stornierungen + Σ Auszahlungen</b><br/><br/>Saldo &gt; 0 → offene Beträge<br/>Saldo = 0 → alles beglichen<br/>Saldo &lt; 0 → Rückzahlung fällig (nach Storno bezahlter Positionen)"]
    end

    e1 -->|"+ erhöht Saldo"| formel
    e2 -.->|"kein Effekt"| formel
    e3 -->|"− reduziert Saldo"| formel
    e4 -->|"− reduziert Saldo"| formel
    e5 -->|"+ erhöht Saldo"| formel

    style e1 fill:#c8e6c9,stroke:#2e7d32
    style e2 fill:#e0e0e0,stroke:#616161
    style e3 fill:#bbdefb,stroke:#1565c0
    style e4 fill:#ffcdd2,stroke:#b71c1c
    style e5 fill:#fff9c4,stroke:#f57f17
```

### Kassensitzung-Events (Subject: `kassensitzung-{nr}`)

```mermaid
graph TD
    subgraph ks_events ["Kassensitzung Events (append-only)"]
        k1["<b>KassensitzungEroeffnet</b><br/>Datum, Bezeichnung"]
        k2["<b>AnfangsbestandGesetzt</b><br/>BetragCents (Wechselgeld)"]
        k3["<b>KassenbewegungGebucht</b><br/>Art (geldtransit | privatentnahme | privateinlage)<br/>BetragCents, Kommentar"]
        k4["<b>KassensturzDurchgefuehrt</b><br/>SollBestandCents, IstBestandCents<br/>DifferenzCents"]
        k5["<b>DifferenzSollIstGebucht</b><br/>BetragCents<br/>(nur wenn Differenz ≠ 0)"]
        k6["<b>TagesabschlussErstellt</b><br/>Z_NR, Zeitraum, Umsätze<br/>→ Kassensitzung abgeschlossen"]
    end

    k1 --> k2
    k2 --> k3
    k3 --> k4
    k4 -.->|"wenn Differenz ≠ 0"| k5
    k4 --> k6
    k5 --> k6

    style k1 fill:#e3f2fd,stroke:#1565c0
    style k2 fill:#e3f2fd,stroke:#1565c0
    style k3 fill:#fff3e0,stroke:#e65100
    style k4 fill:#fff3e0,stroke:#e65100
    style k5 fill:#ffcdd2,stroke:#b71c1c
    style k6 fill:#ffcdd2,stroke:#b71c1c
```

---

## 6 · Kasse — Bestellvorgang (Sequenz)

Vom Tap der Servicekraft bis zum persistierten Event. Vor jeder Tisch-Operation prüft der Application Service die Kassensitzung-Sperre. _(TSE-Signierung ist geplant, aber noch nicht implementiert.)_

```mermaid
sequenceDiagram
    actor SK as Servicekraft
    participant FE as Frontend (Browser)
    participant BE as Backend (Go)
    participant DB as PostgreSQL
    participant KSP as kassensitzungen
    participant TSS as tisch_sessions
    participant Relay as Print-Relay
    participant Drucker as Bondrucker

    SK->>FE: Tisch auswählen, Positionen wählen, "Bestellen"
    FE->>FE: Zod-Validierung (Positionen, Menge, Kommentar)
    FE->>BE: POST /service/bestellung-aufnehmen<br/>{tischId, positionen[], kommentar, version}
    BE->>BE: zog-Validierung
    BE->>KSP: GetOffeneKassensitzung()

    alt Keine offene Kassensitzung
        KSP-->>BE: nil
        BE-->>FE: 409 "Kasse ist noch nicht geöffnet"
    else Kassensitzung offen
        KSP-->>BE: KS{Nr: 1, Datum: 20260322, Status: offen}
        BE->>DB: Produkt-Stammdaten laden (Fat Event)

        rect rgb(240, 248, 255)
            Note over BE,TSS: Transaktion (BEGIN...COMMIT)
            Note over BE: Subject = "kassensitzung-1/tisch-42"
            BE->>DB: INSERT INTO kassenjournal (BestellungAufgenommen)
            DB-->>BE: event_id, version
            BE->>TSS: SELECT tisch_sessions
            BE->>BE: ApplyEvent(state, event) → neuer State
            BE->>TSS: UPSERT tisch_sessions
        end

        BE-->>FE: 200 OK {tischState}
        FE-->>SK: Bestellung bestätigt ✓
    end

    Note over Relay,Drucker: Asynchron (Polling)
    Relay->>BE: POST /relay/poll {token, lastEventId}
    BE-->>Relay: {auftraege[], cursor}
    Relay->>Drucker: ESC/POS-Payload (TCP)
    Drucker-->>Relay: ACK
```

---

## 7 · Stammdaten — Domain Model

Die CRUD-verwalteten Aggregate des Stammdaten-Context.

```mermaid
classDiagram
    class Produkt {
        +int ProduktId
        +string Name
        +Kategorie Kategorie
        +EntityStatus Status
        +Variante[] Varianten
    }

    class Variante {
        +int VarianteId
        +string Name
        +int PreisCents
        +EntityStatus Status
    }

    class Kategorie {
        <<enumeration>>
        essen
        getraenk
        sonstiges
    }

    class Tisch {
        +int TischId
        +string Name
        +EntityStatus Status
    }

    class Benutzer {
        +int BenutzerId
        +string Name
        +string Benutzername
        +string PasswortHash
        +string EinmalpasswortHash
        +Role Rolle
        +EntityStatus Status
    }

    class Role {
        <<enumeration>>
        admin
        serviceleitung
        service
    }

    class EntityStatus {
        <<enumeration>>
        active
        inactive
        deleted
    }

    class TischFavorit {
        +int UserId
        +int TischId
    }

    Produkt "1" *-- "1..*" Variante : enthält
    Produkt --> Kategorie
    Variante --> EntityStatus
    Produkt --> EntityStatus
    Tisch --> EntityStatus
    Benutzer --> Role
    Benutzer --> EntityStatus
    TischFavorit --> Benutzer
    TischFavorit --> Tisch
```

---

## 8 · Kassensitzung-Lifecycle

Der vollständige Ablauf einer Kassensitzung von Eröffnung bis Tagesabschluss — alle Vorgänge als Events im Kassenjournal.

```mermaid
flowchart TD
    Start([Veranstaltungsbeginn])

    Start --> KS["<b>1. Kassensitzung eröffnen</b><br/>Admin vergibt Bezeichnung<br/>(z.B. 'Sommerfest Tag 1')<br/>Subject: kassensitzung-{nr}<br/>Event: kassensitzung-eroeffnet:v1<br/>Max. 1 offene KS"]

    KS --> AB["<b>2. Anfangsbestand setzen</b><br/>Wechselgeld in Cent eingeben<br/>Event: anfangsbestand-gesetzt:v1<br/>Einmalig pro Kassensitzung"]

    AB --> Betrieb

    subgraph Betrieb ["Laufender Betrieb"]
        direction TB
        KB["<b>Tisch-Operationen</b><br/>Bestellen → Ausgabe → Zahlung<br/>→ Auszahlung → Stornierung<br/>Subject: kassensitzung-{nr}/tisch-{id}<br/>(Events im Kassenjournal)"]

        BW["<b>Kassenbewegungen</b><br/>Geldtransit · Privatentnahme · Privateinlage<br/>Event: kassenbewegung-gebucht:v1<br/>mit art-Feld (Art der Bewegung)"]

        Bestand["<b>Kassenbestand (SQL-Aggregation)</b><br/>Soll = Anfangsbestand<br/>+ Zahlungen − Auszahlungen<br/>− Geldtransit − Privatentnahmen<br/>+ Privateinlagen + DifferenzSollIst<br/><i>Alles aus einer Kassenjournal-Query</i>"]

        KB --> Bestand
        BW --> Bestand
    end

    Betrieb --> Saldo{"Alle Tische<br/>Saldo = 0?"}
    Saldo -->|Nein| Betrieb
    Saldo -->|Ja| KSturz

    KSturz["<b>3. Kassensturz</b><br/>Admin zählt Bargeld (Ist-Bestand)<br/>System zeigt Soll vs. Ist<br/>Event: kassensturz-durchgefuehrt:v1<br/>+ differenz-soll-ist-gebucht:v1 (wenn ≠ 0)"]

    KSturz --> ZBon["<b>4. Tagesabschluss (Z-Bon)</b><br/>Event: tagesabschluss-erstellt:v1<br/>• Fortlaufende z_nr<br/>• Aggregierte Umsätze nach Steuersätzen<br/>• Soll/Ist/Differenz<br/>→ Kassensitzung abgeschlossen"]

    ZBon --> Close["Kassensitzung abgeschlossen<br/>Status: abgeschlossen"]

    Close --> Aufbewahrung["📁 10 Jahre Aufbewahrungspflicht<br/>(GoBD)"]
    Close --> Export["📤 DSFinV-K-Export<br/>ZIP mit CSV-Dateien<br/>für Betriebsprüfer"]
    Close --> Nächster{Nächster Tag?}

    Nächster -->|Ja| Start
    Nächster -->|Nein| Ende([Veranstaltungsende])

    style KS fill:#e3f2fd,stroke:#1565c0
    style AB fill:#e3f2fd,stroke:#1565c0
    style Saldo fill:#fff3e0,stroke:#e65100
    style KSturz fill:#fff3e0,stroke:#e65100
    style ZBon fill:#ffcdd2,stroke:#b71c1c
    style Bestand fill:#e8f5e9,stroke:#2e7d32
```

---

## 9 · Kasse — Kassenbestand-Berechnung

Woher die Werte für den Soll-Kassenbestand kommen — eine einzige SQL-Aggregation über das Kassenjournal.

```mermaid
graph LR
    subgraph quellen ["Kassenjournal-Events (WHERE kassensitzung_nr = $1)"]
        AB["anfangsbestand-gesetzt:v1<br/>BetragCents"]
        ZK["Σ zahlung-kassiert:v1<br/>(aus Tisch-Sessions)"]
        AZ["Σ auszahlung-geleistet:v1<br/>(aus Tisch-Sessions)"]
        BW_PE["Σ kassenbewegung-gebucht:v1<br/>art = privateinlage"]
        BW_GT["Σ kassenbewegung-gebucht:v1<br/>art = geldtransit"]
        BW_PN["Σ kassenbewegung-gebucht:v1<br/>art = privatentnahme"]
        DSI["Σ differenz-soll-ist-gebucht:v1"]
    end

    subgraph berechnung ["Kassenbestand (SQL-Aggregation)"]
        formel["<b>Soll-Bestand</b><br/>=  Anfangsbestand<br/>+  Zahlungen<br/>−  Auszahlungen<br/>−  Geldtransit<br/>−  Privatentnahmen<br/>+  Privateinlagen<br/>+  DifferenzSollIst"]
    end

    AB -->|"+"| formel
    ZK -->|"+"| formel
    AZ -->|"−"| formel
    BW_GT -->|"−"| formel
    BW_PN -->|"−"| formel
    BW_PE -->|"+"| formel
    DSI -->|"±"| formel

    subgraph kassensturz ["Kassensturz"]
        soll["Soll-Bestand"]
        ist["Ist-Bestand<br/>(physisch gezählt)"]
        diff["Differenz<br/>= Soll − Ist"]
        soll --> diff
        ist --> diff
    end

    formel --> soll

    style formel fill:#e8f5e9,stroke:#2e7d32
    style diff fill:#fff3e0,stroke:#e65100
```

---

## 10 · Auth und Onboarding

Der zweistufige Onboarding-Prozess und der normale Login-Flow.

```mermaid
sequenceDiagram
    actor Admin
    actor Benutzer
    participant FE as Frontend
    participant BE as Backend
    participant DB as PostgreSQL

    Note over Admin,DB: 1. Benutzer anlegen
    Admin->>FE: Name, Benutzername, Rolle eingeben
    FE->>BE: POST /admin/create-user
    BE->>BE: 6-stelliges Einmalpasswort generieren
    BE->>DB: INSERT users (status=inactive,<br/>einmalpasswort_hash=Argon2id(OTP),<br/>passwort_hash=NULL)
    BE-->>FE: {onetimePassword: "482917"}
    FE-->>Admin: Einmalpasswort anzeigen
    Admin-->>Benutzer: Einmalpasswort mitteilen (mündlich/Zettel)

    Note over Admin,DB: 2. Admin aktiviert Benutzer
    Admin->>FE: Benutzer aktivieren
    FE->>BE: POST /admin/activate-user
    BE->>DB: UPDATE users SET status='active'

    Note over Benutzer,DB: 3. Erstanmeldung
    Benutzer->>FE: Benutzername + Einmalpasswort
    FE->>BE: POST /auth/login
    BE->>DB: SELECT user
    BE->>BE: Prüfe: einmalpasswort_hash ≠ NULL ∧ passwort_hash = NULL
    BE-->>FE: JWT + {mustSetPassword: true}
    FE-->>Benutzer: Weiterleitung → "Passwort setzen"

    Note over Benutzer,DB: 4. Eigenes Passwort vergeben
    Benutzer->>FE: Neues Passwort eingeben (min. 6 Zeichen)
    FE->>BE: POST /auth/set-password {onetimePassword, newPassword}
    BE->>BE: Verifiziere Einmalpasswort gegen Hash
    BE->>DB: UPDATE users SET<br/>passwort_hash=Argon2id(newPW),<br/>einmalpasswort_hash=NULL
    BE-->>FE: JWT (normale Session)
    FE-->>Benutzer: Eingeloggt ✓

    Note over Benutzer,DB: 5. Normale Anmeldung (ab jetzt)
    Benutzer->>FE: Benutzername + eigenes Passwort
    FE->>BE: POST /auth/login
    BE->>DB: SELECT user → Argon2id-Vergleich
    BE-->>FE: JWT {userId, role, exp: +12h}
    FE-->>Benutzer: Dashboard
```

---

## 11 · Schichtenarchitektur (Backend)

Die vier Schichten des Go-Backends mit Verantwortlichkeiten und Abhängigkeitsrichtung.

```mermaid
graph TB
    subgraph http ["HTTP-Schicht<br/><code>api/&lt;domain&gt;/http/</code>"]
        h1["Request-Parsing · Response-Serialisierung"]
        h2["Eigene Request/Response-DTOs mit json-Tags"]
        h3["Mapper: Domain ↔ HTTP"]
        h4["Fehlerresponses · Routing"]
    end

    subgraph app ["Application-Schicht<br/><code>api/&lt;domain&gt;/application/</code>"]
        a1["Use-Case-Orchestrierung"]
        a2["Fachliche Validierung (zog-Schemas)"]
        a3["Aggregat laden → Domain-Logik → Persistieren"]
        a4["Domain-Fehler → Anwendungs-Fehlercodes"]
    end

    subgraph domain ["Domain-Schicht<br/><code>domain/&lt;domain&gt;/</code>"]
        d1["Aggregat-Invarianten"]
        d2["Event-Konstruktion"]
        d3["Zustandsberechnung (ApplyEvent)"]
        d4["Keine DB · Kein HTTP · Keine json-Tags"]
    end

    subgraph repo ["Repository / Infra-Schicht<br/><code>repository/&lt;domain&gt;_repo/</code>"]
        r1["Kassenjournal (append-only) + Projektion + CRUD-Entität — Kasse"]
        r2["CRUD — Stammdaten"]
        r3["sqlc-generierte Queries · pgx/v5"]
    end

    http --> app
    app --> domain
    app --> repo

    style http fill:#e3f2fd,stroke:#1565c0
    style app fill:#fff3e0,stroke:#e65100
    style domain fill:#e8f5e9,stroke:#2e7d32,stroke-width:3px
    style repo fill:#f3e5f5,stroke:#6a1b9a
```

---

## 12 · Event Sourcing: Synchrone Projektion + CRUD-Entität

Write-Through: Event und Tisch-Session-Projektion in derselben Transaktion. Kassensitzung als CRUD-Entität (kein Projection-Update im Event-Pfad).

```mermaid
sequenceDiagram
    participant AS as Application Service
    participant Repo as Kassenjournal Repository
    participant DB as PostgreSQL
    participant KJ as kassenjournal (Event Store)
    participant KSP as kassensitzungen
    participant TSS as tisch_sessions
    participant Dom as Domain (ApplyEvent)

    AS->>Repo: WriteEvent(event, streamType)

    rect rgb(240, 248, 255)
        Note over Repo,Dom: BEGIN TRANSACTION
        Repo->>KJ: INSERT INTO kassenjournal (...)<br/>RETURNING event_id, version
        Note over KJ: UNIQUE (subject, version) → OCC
        KJ-->>Repo: event_id, version

        alt streamType = "tisch-session"
            Repo->>TSS: SELECT * FROM tisch_sessions<br/>WHERE subject = ?
            TSS-->>Repo: aktueller State (oder Zero-Value)
            Repo->>Dom: ApplyEvent(state, event)
            Note over Dom: Reine Funktion — kein DB-Zugriff
            Dom-->>Repo: neuer TischSession
            Repo->>TSS: UPSERT tisch_sessions SET<br/>saldo_cents, unbezahlte_positionen,<br/>ausstehende_positionen, ...

        Note over Repo,Dom: COMMIT
    end

    Repo-->>AS: event_id

    Note over AS,TSS: Lesezugriff (Query)
    AS->>TSS: SELECT saldo_cents,<br/>unbezahlte_positionen, ...<br/>FROM tisch_sessions<br/>WHERE subject = ?
    Note over TSS: Kein Event-Replay nötig!
```

---

## 13 · Bondruck — Relay-Architektur

Wie Bestellungen von der Datenbank zum Bondrucker gelangen.

```mermaid
flowchart LR
    subgraph backend ["jotti Backend"]
        EventStore["kassenjournal-Tabelle<br/>(BestellungAufgenommen)"]
        KatDrucker["kategorie_drucker<br/>(IP + Bonmodus)"]
        PollEndpoint["POST /relay/poll<br/>Auth: RELAY_AUTH_TOKEN"]
    end

    subgraph relay ["Print-Relay (jotti-relay)"]
        Poller["Poller<br/>(alle 2s)"]
        State["State-Datei<br/>(cursor + printed_ids)"]
        ESCPos["ESC/POS-Formatter"]
    end

    subgraph drucker ["Bondrucker (Ethernet)"]
        Küche["🍳 Küchendrucker<br/>Kategorie: essen"]
        Theke["🍺 Thekendrucker<br/>Kategorie: getraenk"]
    end

    Poller -->|"POST {token, lastEventId}"| PollEndpoint
    PollEndpoint --> EventStore
    PollEndpoint --> KatDrucker
    PollEndpoint -->|"{auftraege[], cursor}"| Poller
    Poller --> ESCPos
    Poller --> State
    ESCPos -->|"TCP"| Küche
    ESCPos -->|"TCP"| Theke

    style EventStore fill:#e8f5e9,stroke:#2e7d32
    style relay fill:#fff3e0,stroke:#e65100
    style Küche fill:#ffcdd2,stroke:#b71c1c
    style Theke fill:#bbdefb,stroke:#1565c0
```

**Bonformate:**

| Bonmodus         | Beschreibung                                                           |
| ---------------- | ---------------------------------------------------------------------- |
| `pro_position`   | 1 Bon pro Position — Tischname (groß/fett) + 1 Position (doppelt hoch) |
| `pro_bestellung` | 1 Sammelbon pro Kategorie — Tischname + alle Positionen der Kategorie  |

---

## 14 · Deployment-Architektur

Docker-Compose-Stack für Self-Hosting.

```mermaid
graph TB
    subgraph internet ["Internet"]
        Client["📱 Smartphone-Browser<br/>(BYOD)"]
    end

    subgraph server ["Server (Docker Compose)"]
        subgraph nginx ["nginx Reverse Proxy"]
            TLS["TLS-Terminierung<br/>Let's Encrypt"]
            Static["Statische Dateien<br/>(Frontend-Build)"]
            Proxy["Proxy-Pass → Backend"]
        end

        subgraph backend ["jotti Backend (Go)"]
            API["Go HTTP-Server<br/>net/http stdlib"]
        end

        subgraph db ["PostgreSQL 17"]
            Database["Datenbank<br/>Kassenjournal · Stammdaten · Projektion + CRUD-Entität"]
        end

        subgraph migrate ["golang-migrate"]
            Mig["Schema-Migrationen<br/>01_initial.up.sql"]
        end
    end

    subgraph lokal ["Lokales Netzwerk"]
        subgraph relayhost ["Relay-Host (beliebiger PC)"]
            PrintRelay["jotti-relay<br/>(Go-Binary)"]
        end

        Drucker1["🍳 Küchendrucker"]
        Drucker2["🍺 Thekendrucker"]
    end

    Client -->|"HTTPS"| TLS
    TLS --> Static
    TLS --> Proxy
    Proxy -->|"HTTP"| API
    API -->|"pgx/v5"| Database
    Mig -->|"Migrationen"| Database

    PrintRelay -->|"HTTPS<br/>POST /relay/poll"| TLS
    PrintRelay -->|"TCP<br/>ESC/POS"| Drucker1
    PrintRelay -->|"TCP<br/>ESC/POS"| Drucker2

    style nginx fill:#e3f2fd,stroke:#1565c0
    style backend fill:#e8f5e9,stroke:#2e7d32
    style db fill:#f3e5f5,stroke:#6a1b9a
    style relayhost fill:#fff3e0,stroke:#e65100
```

---

## 15 · API-Bereichsgliederung

Alle API-Endpunkte gruppiert nach Bereich und Berechtigung. Alle Endpunkte sind POST-only.

```mermaid
graph TB
    subgraph auth_api ["🔓 /auth/* — Öffentlich"]
        login["POST /auth/login"]
        setPW["POST /auth/set-password"]
    end

    subgraph admin_api ["🔑 /admin/* — Rolle: admin"]
        subgraph admin_stamm ["Stammdaten"]
            cp["create-produkt · update-produkt<br/>activate-produkt · deactivate-produkt · delete-produkt"]
            cv["create-variante · update-variante<br/>activate-variante · deactivate-variante · delete-variante"]
            ct["create-tisch · update-tisch<br/>activate-tisch · deactivate-tisch · delete-tisch"]
            cu["create-user · update-user<br/>activate-user · deactivate-user<br/>delete-user · reset-password"]
            gp["get-all-produkte · get-all-tische · get-all-users"]
        end
        subgraph admin_kf ["Kasse — Kassensitzung"]
            ak["kassensitzung-eroeffnen"]
            ab_set["anfangsbestand-setzen"]
            gks["get-offene-kassensitzung"]
            gkb["get-kassenbestand"]
            bw["kassenbewegung-buchen"]
            ks["kassensturz-durchfuehren"]
            ta["tagesabschluss-erstellen"]
        end
        subgraph admin_rep ["Reporting"]
            rep["get-abrechnung"]
        end
        subgraph admin_conf ["Konfiguration"]
            dc["get-drucker-konfiguration · update-drucker-konfiguration"]
        end
    end

    subgraph service_api ["🍽️ /service/* — Rolle: service+"]
        sb["bestellung-aufnehmen"]
        sa["ausgabe-bestaetigen"]
        sz["zahlung-kassieren"]
        st["get-tisch-state · get-aktive-tische<br/>get-aktive-tische-mit-favoriten<br/>get-meine-tische-state"]
        sh["get-tisch-historie"]
        sf["favorit-hinzufuegen · favorit-entfernen"]
        sp["get-aktive-produkte"]
        se["get-eigene-uebersicht"]
    end

    subgraph sl_api ["👔 /serviceleitung/* — Rolle: serviceleitung+"]
        sto["stornierung-erteilen"]
        aus["auszahlung-leisten"]
    end

    subgraph relay_api ["🖨️ /relay/* — Token-Auth"]
        poll["POST /relay/poll<br/>{token, lastEventId}"]
    end

    style auth_api fill:#f3e5f5,stroke:#6a1b9a
    style admin_api fill:#ffcdd2,stroke:#b71c1c
    style service_api fill:#c8e6c9,stroke:#2e7d32
    style sl_api fill:#fff9c4,stroke:#f57f17
    style relay_api fill:#e0e0e0,stroke:#616161
```

---

## 16 · Fiskalkonformität — Compliance-Architektur

Wie die Compliance-Komponenten (TSE, DSFinV-K, ELSTER) ins System integriert werden.

```mermaid
graph TB
    subgraph phase1 ["Phase 1 — Compliance-Grundlage"]
        SN["F-01: Seriennummer<br/>UUID-v4 in system_config<br/>Beim ersten Start generiert"]
        ST["F-07: Steuersätze<br/>standard (19%) · ermäßigt (7%) · befreit (0%)<br/>Pro Variante konfigurierbar"]
        AK_F["F-06: Abrechnungskreis<br/>Fortlaufend nummeriert<br/>DSFinV-K-Pflichtfeld"]
        Beleg1["F-03: Belegausgabe (Basis)<br/>Steuersatz · Betrag · Seriennummer<br/>Betreiberadresse · Zahlungsart"]
        ELSTER["F-05: ELSTER-Anleitung<br/>Manuelle Meldung über Webportal<br/>Hinweis im Admin-Dashboard"]
        BD["KF-09: Betreiber-Stammdaten<br/>Name · Adresse · Steuernummer"]
    end

    subgraph phase2 ["Phase 2 — TSE-Integration"]
        TSE_IF["F-02: TSEClient Interface<br/>StartTransaction<br/>UpdateTransaction<br/>FinishTransaction"]
        Fiskaly["FiskalyTSEClient<br/>(REST-Adapter)"]
        TSE_Mock["TSE-Mock<br/>(Tests + Bypass-Modus)"]
        Hooks["TSE-Hooks<br/>BestellungAufnehmen → TSE<br/>ZahlungKassieren → TSE<br/>StornierungErteilen → TSE<br/>Tagesabschluss → TSE"]
        Beleg2["F-03: Beleg + TSE-Felder<br/>Transaktionsnummer · Signaturzähler<br/>TSE-Seriennummer · QR-Code"]
        DSFinVK["F-04: DSFinV-K-Export<br/>ZIP mit CSV-Dateien<br/>Bonkopf · Bonpos · TSE_Transaktionen<br/>Stamm_Kassen · index.xml"]
    end

    subgraph phase3 ["Phase 3 — Erweiterungen"]
        ERiC["F-05: ERiC-Schnittstelle<br/>Programmatische ELSTER-Meldung"]
        HashChain["F-08: GoBD Hash-Chain<br/>SHA-256 über Event-Stream<br/>Integritätsprüfung"]
    end

    SN --> Beleg1
    BD --> Beleg1
    ST --> Beleg1
    TSE_IF --> Fiskaly
    TSE_IF --> TSE_Mock
    Hooks --> TSE_IF
    Hooks --> Beleg2
    DSFinVK --> AK_F
    DSFinVK --> SN
    DSFinVK --> ST

    style phase1 fill:#e8f5e9,stroke:#2e7d32
    style phase2 fill:#fff3e0,stroke:#e65100
    style phase3 fill:#f3e5f5,stroke:#6a1b9a
```

---

## 17 · Datenbank-Schema (ER-Diagramm)

Alle Tabellen des aktuellen Schemas — Domänen-Tabellen deutsch, Infrastruktur-Tabellen englisch.

```mermaid
erDiagram
    users {
        int id PK
        string name
        string username UK
        string password_hash
        string onetime_password_hash
        string role "admin | serviceleitung | service"
        string status "active | inactive | deleted"
        timestamptz created_at
        timestamptz updated_at
    }

    produkte {
        int id PK
        string name
        string kategorie "essen | getraenk | sonstiges"
        string status "active | inactive | deleted"
        timestamptz created_at
        timestamptz updated_at
    }

    produkt_varianten {
        int id PK
        int produkt_id FK
        string name
        int preis_cents
        string status "active | inactive | deleted"
        timestamptz created_at
        timestamptz updated_at
    }

    tische {
        int id PK
        string name
        string status "active | inactive | deleted"
        timestamptz created_at
        timestamptz updated_at
    }

    kassenjournal {
        int id PK
        int user_id FK
        string user_name
        string type "bestellung-aufgenommen:v1 | ..."
        string subject "kassensitzung-{nr}[/tisch-{id}]"
        int version "pro subject, für OCC"
        jsonb data "Event-spezifische Daten"
        int kassensitzung_nr "FK-artige Zuordnung zur KS"
        timestamptz timestamp
    }

    kassensitzungen {
        int z_nr PK
        date datum
        string bezeichnung
        string status
        timestamptz created_at
        timestamptz updated_at
    }

    tisch_sessions {
        string subject PK "kassensitzung-{nr}/tisch-{id}"
        int tisch_id FK
        int kassensitzung_nr
        int saldo_cents
        jsonb unbezahlte_positionen
        jsonb ausstehende_positionen
        int gesamt_zahlungen_cents
        int last_event_id FK
        int last_event_version
        timestamptz updated_at
    }

    tisch_favoriten {
        int user_id PK, FK
        int tisch_id PK, FK
        timestamptz created_at
    }

    kategorie_drucker {
        string kategorie PK "essen | getraenk | sonstiges"
        string drucker_ip
        string bonmodus "pro_position | pro_bestellung"
        timestamptz updated_at
    }

    produkte ||--o{ produkt_varianten : "hat Varianten"
    tische ||--o{ tisch_sessions : "hat Sessions"
    kassenjournal }o--|| kassensitzungen : "kassensitzung_nr → z_nr"
    tisch_sessions }o--|| kassensitzungen : "kassensitzung_nr → z_nr"
    tisch_sessions ||--o| kassenjournal : "letztes Event"
    users ||--o{ tisch_favoriten : "hat Favoriten"
    tische ||--o{ tisch_favoriten : "ist Favorit von"
```

---

## 18 · Frontend-Seitenstruktur

Navigation und Seitenstruktur der Mobile-first Web-App.

```mermaid
graph TB
    subgraph public ["Öffentlich"]
        Login["<b>/login</b><br/>Benutzername + Passwort"]
        SetPW["<b>/set-password</b><br/>Eigenes Passwort vergeben<br/>(Erstanmeldung / Reset)"]
    end

    subgraph service_guard ["ServiceGuard (service · serviceleitung · admin)"]
        Dashboard["<b>/service/tische</b><br/>Service-Dashboard<br/>Meine Tische (Favoriten) als Rich Cards<br/>+ Eigene Übersicht (KPIs)"]

        AlleTische["<b>Drawer: Alle Tische</b><br/>Alle aktiven Tische<br/>Schnellsuche (K-11)<br/>Favoriten-Toggle ★"]

        TischDetail["<b>/service/tische/:tischId</b><br/>Tisch-Detail mit 3 Tabs"]

        subgraph tabs ["Tabs im Tisch-Detail"]
            TabBestellen["<b>Tab: Bestellen</b><br/>Produktkatalog nach Kategorie<br/>+/−  Mengensteuerung<br/>Ausgabe bestätigen (integriert)"]
            TabBezahlen["<b>Tab: Bezahlen</b><br/>Unbezahlte Positionen wählen<br/>Rückgeldberechnung (K-10)<br/>Stornieren (serviceleitung/admin)"]
            TabHistorie["<b>Tab: Historie</b><br/>Kassenjournal (Event Stream)<br/>Chronologisch, unveränderlich"]
        end

        subgraph drawers ["Drawers (Mobile-Overlay)"]
            DrawerBestellen["Bestellung zusammenstellen<br/>→ Zusammenfassung → Bestätigen"]
            DrawerAusgabe["Ausgabe bestätigen<br/>Positionen wählen → Bestätigen"]
            DrawerBezahlen["Zahlung kassieren<br/>Positionen wählen → Bestätigen"]
            DrawerStorno["Stornierung erteilen<br/>Positionen + Pflichtkommentar"]
            DrawerAuszahlung["Auszahlung leisten<br/>Betrag + Pflichtkommentar"]
        end
    end

    subgraph admin_guard ["AdminGuard (nur admin)"]
        Produkte["<b>/admin/produkte</b><br/>Produkte & Varianten verwalten"]
        Tische["<b>/admin/tische</b><br/>Tische verwalten"]
        Benutzer["<b>/admin/benutzer</b><br/>Benutzer verwalten<br/>Einmalpasswort generieren"]
        Drucker["<b>/admin/drucker</b><br/>Druckerkonfiguration<br/>IP + Bonmodus pro Kategorie"]

        subgraph kf_admin ["Kassensitzung"]
            KF_Übersicht["<b>/admin/kasse</b><br/>Aktive Kassensitzung<br/>Kassenbestand · Bewegungen<br/>Z-Bon-Historie"]
            KF_AK["Kassensitzung eröffnen"]
            KF_AB["Anfangsbestand setzen"]
            KF_BW["Kassenbewegungen buchen"]
            KF_KS["Kassensturz durchführen"]
            KF_ZB["Tagesabschluss (Z-Bon)"]
        end

        subgraph rep_admin ["Reporting"]
            Rep_Tages["<b>/admin/auswertung</b><br/>Tagesabrechnung (KPIs)<br/>Umsatz pro Servicekraft/Tisch<br/>Stornierungsubersicht"]
        end
    end

    Login -->|"JWT"| Dashboard
    Login -->|"mustSetPassword"| SetPW
    SetPW -->|"JWT"| Dashboard
    Dashboard --> AlleTische
    Dashboard --> TischDetail
    AlleTische --> TischDetail
    TischDetail --> tabs
    TabBestellen --> DrawerBestellen
    TabBestellen --> DrawerAusgabe
    TabBezahlen --> DrawerBezahlen
    TabBezahlen --> DrawerStorno
    TabBezahlen --> DrawerAuszahlung

    style public fill:#f3e5f5,stroke:#6a1b9a
    style service_guard fill:#e8f5e9,stroke:#2e7d32
    style admin_guard fill:#ffcdd2,stroke:#b71c1c
```
