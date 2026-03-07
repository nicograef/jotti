# Ubiquitous Language — Domain-Begriffe in jotti

Dieses Dokument beschreibt die kanonische Fachsprache (Ubiquitous Language) von jotti im Sinne von Domain-Driven Design (DDD). Ziel ist eine einheitliche Terminologie in Code, Dokumentation und Kommunikation.

---

## Bounded Contexts

jotti hat zwei klar abgegrenzte Fachbereiche:

| Bounded Context | Deutsch | Inhalt |
|---|---|---|
| **Kassenbetrieb** | `service` | Tische, Bestellungen, Zahlungen, Lieferungen, Stornierungen |
| **Stammdaten** | `admin` | Produkte, Varianten, Tische (Stammdaten), Benutzer |

Auth (`/auth/*`) ist technische Infrastruktur, kein eigenständiger Bounded Context.

---

## Kanonische Begriffe

### Kassenbetrieb (Bounded Context: service)

| Fachbegriff (DE) | Code-Begriff (EN) | Beschreibung |
|---|---|---|
| **Tisch** | `table` | Abrechnungseinheit. Träger des Kassenjournals. |
| **Bestellung** | `order` | Auftrag des Gastes, aufgenommen von der Servicekraft. |
| **Position** | `line item` | Eine Produktvariante mit Menge und Preis innerhalb einer Bestellung. |
| **Lieferung** | `delivery` | Bestätigung, dass bestellte Positionen ausgeliefert wurden. |
| **Zahlung** | `payment` | Registrierung eines Zahlungseingangs (bar, Teilzahlung möglich). |
| **Stornierung** | `cancelation` | Rücknahme bestellter Positionen (nur Serviceleitung/Admin). |
| **Kassenjournal** | `event log` | Unveränderliche Folge aller Tischoperationen (Event Stream). |
| **Saldo** | `balance` | Offener Betrag an einem Tisch: Bestellsumme − Zahlungen − Stornierungen. |
| **Servicekraft** | `service user` | Rolle: nimmt Bestellungen auf, liefert aus, kassiert. |
| **Serviceleitung** | `senior service` | Rolle: alle Rechte der Servicekraft + Stornierung. |

### Stammdaten (Bounded Context: admin)

| Fachbegriff (DE) | Code-Begriff (EN) | Beschreibung |
|---|---|---|
| **Produkt** | `product` | Angebotener Artikel (z. B. „Cola", „Pommes"). |
| **Variante** | `variant` | Ausprägung eines Produkts mit eigenem Preis (z. B. „Klein 0,3l", „Groß 0,5l"). |
| **Kategorie** | `category` | Gruppierung von Produkten: Essen, Getränke, Sonstiges. |
| **Tisch** | `table` | Physischer Tisch mit Name und Status (aktiv/inaktiv). |
| **Benutzer** | `user` | Person mit Zugangsdaten und Rolle. |
| **Admin** | `admin` | Rolle: voller Zugriff auf Stammdaten und Kassenbetrieb. |

---

## Bekannte Inkonsistenzen und Empfehlungen

### 1. `cancelation` — Schreibfehler im Englischen

**Problem:** Im Code wird `cancelation` (ein `l`) verwendet. Die korrekte englische Schreibweise ist `cancellation` (zwei `l`).

**Aktueller Zustand:** Konsistent (aber falsch) in: `cancelation.go`, `productsCanceledEvent.go`, Event-Typ `table.variants-canceled:v1`, Tests, Responses.

**Empfehlung:** In einer Breaking-Change-Migration auf `cancellation` vereinheitlichen. Da Event-Typen versioniert sind (`table.variants-canceled:v1`), kann eine neue Version `table.variants-cancelled:v1` eingeführt werden, wenn Event-Migration ohnehin anfällt.

**Priorität:** Niedrig — solange intern konsistent, kein funktionaler Schaden.

---

### 2. „Bediener" vs. „Servicekraft"

**Problem:** Im Anforderungskatalog (`docs/requirements.md`) wird die Rolle als „Bediener" bezeichnet. In Code, AGENTS.md und allen anderen Dokumenten heißt dieselbe Rolle „Servicekraft".

**Empfehlung:** Kanonisch ist **„Servicekraft"** (Service-Kontext) und **„Bediener"** als informeller Oberbegriff in der Tagesabrechnung. Die Rollenspalte im Anforderungskatalog sollte auf „Servicekraft" umgestellt werden.

---

### 3. „Bon" vs. „Quittung" vs. „Rechnung"

**Problem:** Im Projekt tauchen drei Begriffe für gedruckte Belege auf — ohne klare Abgrenzung.

**Empfehlung:**

| Begriff | Bedeutung in jotti |
|---|---|
| **Bon** | Gedruckter Beleg für Küche/Ausgabe (Küchenbon) **oder** für den Gast (Kassenbon) |
| **Kassenbon** | Gast-Beleg mit Zusammenfassung und Gesamtbetrag |
| **Küchenbon** | Interner Bon für Küche/Getränkeausgabe |
| **Quittung** | Nicht verwenden — kein etablierter Begriff in der Gastronomie |
| **Rechnung** | Nicht verwenden — hat steuerrechtliche Bedeutung (MwSt.-Ausweis) |

---

### 4. `LineItem` — fehlende deutsche Entsprechung

**Problem:** Der Code-Begriff `LineItem` hat keine konsistente deutsche Entsprechung. In der UI wird manchmal „Position", manchmal „Artikel" verwendet.

**Empfehlung:** Kanonisch **„Position"** für den deutschen UI-Text. `LineItem` bleibt der englische Code-Begriff.

---

### 5. „Rechner" als Rollenbezeichnung im Anforderungskatalog

**Problem:** Der Anforderungskatalog führt „Rechner" als Rolle (für Umsatzberichte, Tagesabrechnung). Diese Rolle existiert nicht im Code und wird auch nicht als separater Login abgebildet — die Funktion liegt beim Admin.

**Empfehlung:** „Rechner" in der Anforderungsdokumentation durch „Admin (Abrechnung)" ersetzen, um deutlich zu machen, dass es sich um den bestehenden Admin-Benutzer handelt, der Abrechnungsfunktionen nutzt.

---

### 6. Event-Typ-Namensgebung

**Problem:** Event-Typen im Code folgen dem Muster `table.<verb>-<noun>:<version>`. Das Verb-Nomen-Muster ist uneinheitlich:
- `order-placed` → Nomen-Verb (Substantiv + Partizip Perfekt Passiv) ✅
- `payment-registered` → Nomen-Verb ✅
- `variants-canceled` → Nomen-Verb ✅
- `variants-delivered` → Nomen-Verb ✅

Das Muster `<subject>-<past-participle>` ist konsistent. **Keine Änderung nötig.**

---

### 7. Snapshot als Event-Typ

**Problem:** `table.snapshot:v1` ist kein Business-Event, sondern eine technische Optimierung. Es erscheint im gleichen Event-Stream wie fachliche Events.

**Empfehlung:** In der Dokumentation klar unterscheiden zwischen:
- **Domain Events** (fachlich): `order-placed`, `payment-registered`, `variants-canceled`, `variants-delivered`
- **System Events** (technisch): `snapshot` — dient nur der Performance, hat keine fachliche Bedeutung

---

## DDD-Strukturempfehlungen

### Aggregat-Grenzen

Das `Table`-Aggregat (Tisch) ist korrekt abgegrenzt: Es kapselt den gesamten Kassenbetrieb an einem Tisch. Alle Tischoperationen laufen über den Event-Stream des Tisches.

**Empfehlung:** Das `Table`-Aggregat explizit dokumentieren:
- **Aggregatwurzel**: `Table` (identifiziert durch `table:<id>`)
- **Domain Events**: `OrderPlaced`, `PaymentRegistered`, `VariantsCanceled`, `VariantsDelivered`
- **Invarianten**: Saldo ≥ 0 nach Stornierung, Zahlung nur für bestellte Positionen

### Repository-Benennung

Aktuell: `table_repo`, `user_repo`, `product_repo`, `event_repo`

**Empfehlung:** Konsistent, keine Änderung nötig. `event_repo` bezieht sich auf den generischen Event-Store — passend für ein Event-Sourcing-System.

### Application-Service vs. Domain-Service

Aktuell werden Queries wie `GetBalance`, `GetHistory` als freie Funktionen in `domain/table/events.go` implementiert. Das ist sinnvoll, da sie keinen Seiteneffekt haben.

**Empfehlung:** Diese Funktionen sind **Domain Services** im DDD-Sinne. Sie könnten explizit als solche dokumentiert werden, um die Abgrenzung zu Application Services (HTTP-Handler-nahe Orchestrierung) klarer zu machen.
