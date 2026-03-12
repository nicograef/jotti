# Ubiquitous Language — jotti

Dieses Dokument ist die **verbindliche Referenz** für die Ubiquitous Language des jotti-Projekts — für Entwickler, Agenten und alle Projektbeteiligten. Es definiert die Fachbegriffe der Domäne, ihre Code-Repräsentationen und die Sprachkonventionen pro Schicht.

Die Ubiquitous Language ist ein **Living Document**: Sie wird fortlaufend aktualisiert, wenn sich Begriffe, Strukturen oder Konventionen ändern. Ursprung ist [Entwurf Abschnitt 12 „Ubiquitous Language"](entwurf.md#12-ubiquitous-language), der eine Zusammenfassung enthält. Die vollständige und aktuelle Definition wird hier gepflegt.

## Sprachkonventionen

1. **Domänenbegriffe sind deutsch.** Alle Fachbegriffe des Kassenbetriebs, der Stammdaten und der Gastronomie-Domäne werden auf Deutsch benannt — in Code, Dokumentation und Kommunikation. Beispiele: `Bestellung`, `Tisch`, `Zahlung`, `Position`, `Lieferung`, `Stornierung`, `Saldo`.

2. **Infrastruktur-Code bleibt englisch.** Authentifizierung, Konfiguration, Datenbank-Schicht, HTTP-Framework und generische Sub-Domains verwenden englische Bezeichnungen. Beispiele: `User`, `Role`, `Token`, `Config`, `Middleware`.

3. **Benutzer-sichtbare Strings sind deutsch.** Alle UI-Labels, Fehlermeldungen, Platzhalter und Hilfetexte werden auf Deutsch formuliert. Im UI heißt es „Benutzer" (nicht „User"), „Einmalpasswort" (nicht „OnetimePassword"), „Getränke" (nicht „beverage").

4. **Commits sind auf Englisch.** Conventional Commits (`feat:`, `fix:`, `refactor:`, `docs:`, `test:`) mit englischen Nachrichten.

## Namenskonventionen pro Schicht

| Schicht                   | Sprache  | Konvention                 | Beispiel                                        |
| ------------------------- | -------- | -------------------------- | ----------------------------------------------- |
| Go-Domain-Structs         | Deutsch  | PascalCase                 | `Bestellung`, `Tisch`, `Position`               |
| Go-Felder (Domäne)        | Deutsch  | PascalCase                 | `GesamtPreisCents`, `SaldoCents`                |
| TypeScript-Typen (Domäne) | Deutsch  | PascalCase                 | `Bestellung`, `Tisch`, `Zahlung`                |
| JSON-Keys (Domäne)        | Deutsch  | camelCase                  | `"gesamtPreisCents"`, `"saldoCents"`            |
| API-Pfade (Domäne)        | Deutsch  | kebab-case                 | `/bestellung-aufgeben`, `/zahlung-registrieren` |
| DB-Tabellen               | Englisch | snake_case                 | `tables`, `products`, `events`                  |
| DB-Spalten                | Englisch | snake_case                 | `price_cents`, `created_at`                     |
| Frontend-Routen           | Englisch | kebab-case                 | `/service/tables`, `/admin/products`            |
| Auth/Infrastruktur-Code   | Englisch | Sprachübliche Konventionen | `User`, `Role`, `Token`, `Config`               |

## Kassenbetrieb (Core Domain)

### Tisch

Zentrales Aggregat im Kassenbetrieb. Trägt einen Event Stream, aus dem sich der aktuelle Zustand (Saldo, offene Positionen) berechnet.

| Schicht             | Repräsentation                                                                                                                  | Datei                          |
| ------------------- | ------------------------------------------------------------------------------------------------------------------------------- | ------------------------------ |
| Go-Struct           | `Tisch` (Felder: `ID`, `Name`, `Status`, `CreatedAt`)                                                                           | `domain/table/tisch.go`        |
| TypeScript-Typ      | `Tisch`                                                                                                                         | `src/service/table/Tisch.ts`   |
| DB-Tabelle          | `tables`                                                                                                                        | `migrations/01_initial.up.sql` |
| API-Pfade (Admin)   | `/create-tisch`, `/update-tisch`, `/activate-tisch`, `/deactivate-tisch`, `/get-all-tische`                                     | `api/admin.go`                 |
| API-Pfade (Service) | `/get-tisch`, `/get-aktive-tische`, `/get-tisch-historie`, `/get-tisch-saldo`, `/get-tisch-unbezahlt`, `/get-tisch-ungeliefert` | `api/service.go`               |
| Frontend-Hooks      | `useTisch()`, `useAktiveTische()`, `useTischHistorie()`, `useTischSaldo()`, `useTischUnbezahlt()`, `useTischUngeliefert()`      | `src/service/table/hooks.ts`   |
| Frontend-Route      | `/service/tables/:tableId`                                                                                                      | `src/routes.ts`                |

> **Hinweis:** DB-Tabelle (`tables`) und Frontend-Route (`/service/tables`) verwenden Englisch — korrekt per Infrastruktur-Konvention.

### Bestellung

Ein Vorgang, bei dem eine Servicekraft Positionen für einen Tisch aufgibt. Erzeugt ein `BestellungAufgegeben`-Event.

| Schicht             | Repräsentation                                                                                                | Datei                              |
| ------------------- | ------------------------------------------------------------------------------------------------------------- | ---------------------------------- |
| Go-Struct           | `Bestellung` (Felder: `ID`, `UserID`, `TischID`, `Positionen`, `GesamtPreisCents`, `Comment`, `AufgegebenAm`) | `domain/table/bestellung.go`       |
| Go-Event-Typ        | `EventTypeBestellungAufgegebenV1` = `"tisch.bestellung-aufgegeben:v1"`                                        | `domain/table/events.go`           |
| Go-Command          | `BestellungAufgeben()`                                                                                        | `api/table/application/command.go` |
| TypeScript-Typ      | `Bestellung`, `BestellungAufgeben`                                                                            | `src/service/table/Bestellung.ts`  |
| API-Pfad            | `/service/bestellung-aufgeben`                                                                                | `api/service.go`                   |
| Frontend-Komponente | `<Bestellung>`, `<BestellungDrawer>`                                                                          | `src/service/components/table/`    |
| UI-Labels           | „Bestellen" (Tab), „Bestellung aufgeben" (Button), „Bestellung wurde aufgegeben." (Toast)                     |                                    |

### Position

Ein einzelner Posten innerhalb einer Bestellung: Produktvariante + Menge + Einzelpreis.

| Schicht        | Repräsentation                                              | Datei                             |
| -------------- | ----------------------------------------------------------- | --------------------------------- |
| Go-Struct      | `Position` (Felder: `ID`, `Name`, `PreisCents`, `Quantity`) | `domain/table/bestellung.go`      |
| JSON-Keys      | `"id"`, `"name"`, `"preisCents"`, `"quantity"`              | `domain/table/bestellung.go`      |
| TypeScript-Typ | `Position`                                                  | `src/service/table/Bestellung.ts` |

> **Soll-Hinweis:** Das Feld `Quantity` ist englisch — soll langfristig `Menge` heißen (Go-Feld, JSON-Key `"menge"`, TS-Feld). Das Feld `PreisCents` ist bereits korrekt deutsch.

### Lieferung

Die Bestätigung, dass bestellte Positionen dem Gast übergeben wurden. Erzeugt ein `ProdukteGeliefert`-Event.

| Schicht             | Repräsentation                                                                          | Datei                                        |
| ------------------- | --------------------------------------------------------------------------------------- | -------------------------------------------- |
| Go-Struct           | `Lieferung` (Felder: `ID`, `UserID`, `TischID`, `Positionen`, `Comment`, `GeliefertAm`) | `domain/table/lieferung.go`                  |
| Go-Event-Typ        | `EventTypeProdukteGeliefertV1` = `"tisch.produkte-geliefert:v1"`                        | `domain/table/events.go`                     |
| Go-Command          | `ProdukteLiefern()`                                                                     | `api/table/application/command.go`           |
| TypeScript-Typ      | `Lieferung`, `ProdukteLiefern`                                                          | `src/service/table/Lieferung.ts`             |
| API-Pfad            | `/service/produkte-liefern`                                                             | `api/service.go`                             |
| Frontend-Komponente | `<Lieferung>`                                                                           | `src/service/components/table/Lieferung.tsx` |
| UI-Labels           | „Produkte liefern" (Button), „Auslieferung" (Historie)                                  |                                              |

### Zahlung

Die Registrierung einer Barzahlung. Kann sich auf einzelne Positionen beziehen (Teilzahlung). Erzeugt ein `ZahlungRegistriert`-Event.

| Schicht             | Repräsentation                                                                                                | Datei                              |
| ------------------- | ------------------------------------------------------------------------------------------------------------- | ---------------------------------- |
| Go-Struct           | `Zahlung` (Felder: `ID`, `UserID`, `TischID`, `Positionen`, `GesamtZahlungCents`, `Comment`, `RegistriertAm`) | `domain/table/zahlung.go`          |
| Go-Event-Typ        | `EventTypeZahlungRegistriertV1` = `"tisch.zahlung-registriert:v1"`                                            | `domain/table/events.go`           |
| Go-Command          | `ZahlungRegistrieren()`                                                                                       | `api/table/application/command.go` |
| TypeScript-Typ      | `Zahlung`, `ZahlungRegistrieren`                                                                              | `src/service/table/Zahlung.ts`     |
| API-Pfad            | `/service/zahlung-registrieren`                                                                               | `api/service.go`                   |
| Frontend-Komponente | `<ZahlungDrawer>`                                                                                             | `src/service/components/table/`    |
| UI-Labels           | „Bezahlen" (Tab), „Zahlung registrieren" (Button), „Zahlung erfolgreich." (Toast)                             |                                    |

### Stornierung

Die nachträgliche Aufhebung bestellter Positionen. Nur durch Serviceleitung oder Admin. Erzeugt ein `ProdukteStorniert`-Event.

| Schicht             | Repräsentation                                                                                                      | Datei                                                |
| ------------------- | ------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------- |
| Go-Struct           | `Stornierung` (Felder: `ID`, `UserID`, `TischID`, `Positionen`, `GesamtStornierungCents`, `Comment`, `StorniertAm`) | `domain/table/stornierung.go`                        |
| Go-Event-Typ        | `EventTypeProdukteStorniertV1` = `"tisch.produkte-storniert:v1"`                                                    | `domain/table/events.go`                             |
| Go-Command          | `ProdukteStornieren()`                                                                                              | `api/table/application/command.go`                   |
| TypeScript-Typ      | `Stornierung`, `ProdukteStornieren`                                                                                 | `src/service/table/Stornierung.ts`                   |
| API-Pfad            | `/senior_service/produkte-stornieren`                                                                               | `api/senior_service.go`                              |
| Frontend-Komponente | `<StornierungDrawer>`                                                                                               | `src/service/components/table/StornierungDrawer.tsx` |
| UI-Labels           | „Stornierung" (Drawer-Titel), „Produkte stornieren" (Button), „Stornierung erfolgreich." (Toast)                    |                                                      |

### Saldo

Der offene Betrag eines Tisches: Summe der Bestellungen − Summe der Zahlungen − Summe der Stornierungen. Immer in Cent.

| Schicht          | Repräsentation             | Datei                           |
| ---------------- | -------------------------- | ------------------------------- |
| Go-Funktion      | `GetSaldoFromEvents()`     | `domain/table/events.go`        |
| Go-Snapshot-Feld | `SaldoCents`               | `domain/table/snapshotEvent.go` |
| API-Pfad         | `/service/get-tisch-saldo` | `api/service.go`                |
| Frontend-Hook    | `useTischSaldo()`          | `src/service/table/hooks.ts`    |
| UI-Label         | „offen" (Badge)            |                                 |

### Historie

Der vollständige, unveränderliche Event Stream eines Tisches in chronologischer Reihenfolge.

**Synonym: Kassenjournal.** „Kassenjournal" ist der formale Fachbegriff aus dem Entwurf. „Historie" ist der im Code und UI verwendete Begriff und für ehrenamtliche Helfer verständlicher. Beide Begriffe bezeichnen denselben Sachverhalt.

| Schicht               | Repräsentation                | Datei                                            |
| --------------------- | ----------------------------- | ------------------------------------------------ |
| Go-Funktion           | `GetHistoryFromEvents()`      | `domain/table/events.go`                         |
| Go-Query              | `GetTischHistorie()`          | `api/table/application/query.go`                 |
| API-Pfad              | `/service/get-tisch-historie` | `api/service.go`                                 |
| TypeScript-Komponente | `<TischHistorie>`             | `src/service/components/table/TischHistorie.tsx` |
| UI-Label              | „Historie" (Tab)              |                                                  |

### Kommentar

Optionale Freitextnotiz zu einer Bestellung, Zahlung, Lieferung oder Stornierung (max. 100 Zeichen).

| Schicht        | Repräsentation | Datei                                                                                                        |
| -------------- | -------------- | ------------------------------------------------------------------------------------------------------------ |
| Go-Feld (Ist)  | `Comment`      | `domain/table/bestellung.go`, `zahlung.go`, `lieferung.go`, `stornierung.go`, `bestellungAufgegebenEvent.go` |
| JSON-Key (Ist) | `"comment"`    | (alle oben genannten Dateien)                                                                                |
| TS-Feld (Ist)  | `comment`      | `src/service/table/Bestellung.ts`, `Zahlung.ts`, `Lieferung.ts`, `Stornierung.ts`                            |

> **Soll-Hinweis:** Im Code aktuell `Comment` (englisch) — soll langfristig `Kommentar` heißen (Go-Feld, JSON-Key `"kommentar"`, TS-Feld). Betrifft: `bestellung.go`, `zahlung.go`, `lieferung.go`, `stornierung.go`, `bestellungAufgegebenEvent.go` + alle TS-Typen.

### Menge

Anzahl einer Produktvariante innerhalb einer Position.

| Schicht        | Repräsentation | Datei                             |
| -------------- | -------------- | --------------------------------- |
| Go-Feld (Ist)  | `Quantity`     | `domain/table/bestellung.go`      |
| JSON-Key (Ist) | `"quantity"`   | `domain/table/bestellung.go`      |
| TS-Feld (Ist)  | `quantity`     | `src/service/table/Bestellung.ts` |

> **Soll-Hinweis:** Im Code aktuell `Quantity` (englisch) — soll langfristig `Menge` heißen (Go-Feld, JSON-Key `"menge"`, TS-Feld). Betrifft: `bestellung.go` Position-Struct + TS Position-Typ.

### Bezeichnung

Optionaler Name einer Bestellung (z. B. „Familie Müller"), um Gruppen am selben Tisch zu unterscheiden.

> **Hinweis:** Noch nicht implementiert (geplant für K-07).

### Freibon

Sonderposition außerhalb des Produktkatalogs — mit freier Bezeichnung und Preiseingabe. Erzeugt ein `FreibonAusgestellt`-Event.

> **Hinweis:** Noch nicht implementiert.

## Stammdaten (Supporting Sub-Domain)

### Produkt

Artikel im Produktkatalog. Gehört zu genau einer Kategorie und enthält eine oder mehrere Varianten mit je eigenem Preis.

| Schicht              | Repräsentation                                                        | Datei                                  |
| -------------------- | --------------------------------------------------------------------- | -------------------------------------- |
| Go-Struct (Ist)      | `Product` (Felder: `ID`, `Name`, `Category`, `Variants`, `CreatedAt`) | `domain/product/product.go`            |
| DB-Tabelle           | `products`                                                            | `migrations/01_initial.up.sql`         |
| TypeScript-Typ (Ist) | `Product`                                                             | `src/admin/products/ProductBackend.ts` |
| API-Pfade            | `/create-produkt`, `/update-produkt`, `/get-all-produkte`             | `api/admin.go`                         |

> **Soll-Hinweis:** Go-Struct und TS-Typ heißen aktuell `Product` (englisch). Da Produkte zur Domäne gehören, sollen sie langfristig `Produkt` heißen. Die API-Pfade verwenden bereits korrekt die deutsche Form.

### Variante

Konkrete Ausprägung eines Produkts mit eigenem Namen und Preis in Cent (z. B. Produkt „Cola" → Varianten „0,3 l" und „0,5 l").

| Schicht              | Repräsentation                                                                       | Datei                                  |
| -------------------- | ------------------------------------------------------------------------------------ | -------------------------------------- |
| Go-Struct (Ist)      | `Variant` (Felder: `ID`, `Name`, `PriceCents`, `Status`, `CreatedAt`)                | `domain/product/variant.go`            |
| DB-Tabelle           | `product_variants`                                                                   | `migrations/01_initial.up.sql`         |
| TypeScript-Typ (Ist) | `Variant`                                                                            | `src/admin/products/ProductBackend.ts` |
| API-Pfade            | `/create-variante`, `/update-variante`, `/activate-variante`, `/deactivate-variante` | `api/admin.go`                         |

> **Soll-Hinweis:** Go-Struct und TS-Typ heißen aktuell `Variant` (englisch) — sollen langfristig `Variante` heißen. Feld `PriceCents` soll `PreisCents` heißen (Konsistenz mit `Position.PreisCents` im Kassenbetrieb).

### Kategorie

Gruppierung von Produkten. Aktuell drei feste Kategorien: Essen, Getränke, Sonstiges.

| Schicht         | Repräsentation                                                                | Datei                          |
| --------------- | ----------------------------------------------------------------------------- | ------------------------------ |
| Go-Typ (Ist)    | `Category` mit Konstanten `FoodCategory`, `BeverageCategory`, `OtherCategory` | `domain/product/product.go`    |
| DB-Enum         | `ProductCategory` (`'food'`, `'beverage'`, `'other'`)                         | `migrations/01_initial.up.sql` |
| Frontend-Labels | `'food'` → „Essen", `'beverage'` → „Getränke", `'other'` → „Sonstiges"        | `src/service/table/Product.ts` |

> **Soll-Hinweis:** Go-Typ heißt aktuell `Category` (englisch) — soll langfristig `Kategorie` heißen. Konstanten sollen `EssenKategorie`, `GetränkeKategorie`, `SonstigesKategorie` heißen.

### Preis

Geldbeträge werden ausnahmslos als ganzzahlige Cent-Werte gespeichert — niemals als Fließkommazahlen. 3,50 € = 350 Cent.

**Konvention:** Alle Preis-Felder tragen das Suffix `*Cents` (z. B. `PreisCents`, `GesamtPreisCents`, `SaldoCents`, `GesamtZahlungCents`, `GesamtStornierungCents`).

> **Hinweis:** Im Kassenbetrieb wird durchgängig `PreisCents` (deutsch) verwendet. Bei Varianten heißt das Feld aktuell `PriceCents` (englisch) — diese Inkonsistenz ist in der Abweichungstabelle dokumentiert.

### Soft-Delete

Logisches Löschen: Datensätze werden nicht physisch entfernt, sondern durch den Status `deleted` als gelöscht markiert. Der Datensatz bleibt für Referenzintegrität und historische Auswertungen erhalten.

| Schicht              | Repräsentation                                         | Datei                          |
| -------------------- | ------------------------------------------------------ | ------------------------------ |
| DB-Enum              | `EntityStatus` (`'active'`, `'inactive'`, `'deleted'`) | `migrations/01_initial.up.sql` |
| Go-Domain (Tisch)    | Konstanten `ActiveStatus`, `InactiveStatus`            | `domain/table/tisch.go`        |
| Go-Domain (Variante) | Konstanten `ActiveStatus`, `InactiveStatus`            | `domain/product/variant.go`    |
| Go-Domain (User)     | Konstanten `ActiveStatus`, `InactiveStatus`            | `domain/user/user.go`          |

> **Hinweis:** In den Go-Domain-Modellen existieren nur `active` und `inactive` als Konstanten. Der Status `deleted` wird ausschließlich auf DB-Ebene verwendet und ist in der Domain nicht als Konstante abgebildet.

## Authentifizierung (Generic Sub-Domain)

> **Grundregel:** Auth ist eine Generic Sub-Domain — der Code verwendet bewusst englische Bezeichnungen, konform mit AGENTS.md Regel 6 („Infrastruktur-Code bleibt englisch"). Benutzer-sichtbare Strings im UI sind deutsch.

### Benutzer

Person mit Zugang zum System, identifiziert durch einen eindeutigen Benutzernamen.

| Schicht    | Repräsentation                                                                                                  | Datei                          |
| ---------- | --------------------------------------------------------------------------------------------------------------- | ------------------------------ |
| Go-Struct  | `User` (Felder: `ID`, `Name`, `Username`, `Role`, `Status`, `PasswordHash`, `OnetimePasswordHash`, `CreatedAt`) | `domain/user/user.go`          |
| DB-Tabelle | `users`                                                                                                         | `migrations/01_initial.up.sql` |
| API-Pfade  | `/create-user`, `/update-user`, `/activate-user`, `/deactivate-user`, `/reset-password`, `/get-all-users`       | `api/admin.go`                 |
| UI-Label   | „Benutzer verwalten", „Benutzer"                                                                                | `src/admin/users/`             |

> **Hinweis:** Code verwendet bewusst `User` (englisch), da Auth eine Generic Sub-Domain ist. Im UI wird korrekt „Benutzer" angezeigt.

### Rolle

Berechtigungsstufe eines Benutzers. Bestimmt, welche Aktionen im System verfügbar sind.

| Rolle          | Code-Wert        | Berechtigungen                                        |
| -------------- | ---------------- | ----------------------------------------------------- |
| Admin          | `admin`          | Alles: Produkte, Tische, Benutzer verwalten + Service |
| Serviceleitung | `senior_service` | Service-Funktionen + Stornierung                      |
| Servicekraft   | `service`        | Bestellen, Liefern, Kassieren                         |

| Schicht  | Repräsentation                                             | Datei                          |
| -------- | ---------------------------------------------------------- | ------------------------------ |
| Go-Typ   | `Role` mit `AdminRole`, `SeniorServiceRole`, `ServiceRole` | `domain/user/user.go`          |
| DB-Enum  | `UserRole` (`'admin'`, `'senior_service'`, `'service'`)    | `migrations/01_initial.up.sql` |
| JWT-Feld | `role`                                                     | `src/lib/Auth.ts`              |

> **Hinweis:** Englisch im Code ist korrekt (Generic Sub-Domain).

### Einmalpasswort

Vom Admin generiertes 6-stelliges numerisches Passwort für die Erstanmeldung oder das Zurücksetzen des Passworts eines Benutzers.

| Schicht         | Repräsentation          | Datei                          |
| --------------- | ----------------------- | ------------------------------ |
| Go-Feld         | `OnetimePasswordHash`   | `domain/user/user.go`          |
| DB-Spalte       | `onetime_password_hash` | `migrations/01_initial.up.sql` |
| Frontend-Schema | `OnetimePasswordSchema` | `src/lib/AuthBackend.ts`       |
| UI-Label        | „Einmalpasswort"        | `src/admin/users/`             |

> **Hinweis:** Englisch im Code ist korrekt (Infrastruktur). Im UI wird „Einmalpasswort" angezeigt.

### Token

JWT (JSON Web Token) mit Benutzer-ID und Rolle, 12 Stunden gültig. Dient der Authentifizierung bei API-Aufrufen.

Reiner Infrastruktur-Begriff — Englisch im Code ist korrekt.

## Ausgabe (Supporting Sub-Domain)

### Bon

Gedruckter Beleg mit Tisch, Servicekraft, Positionen, Mengen, Zeitstempel und optionalem Kommentar.

> **Hinweis:** Noch nicht implementiert.

### Küchendisplay (KDS)

Echtzeit-Anzeige offener Bestellungen an der Ausgabestation, gruppiert nach Tisch und gefiltert nach Kategorie.

> **Hinweis:** Noch nicht implementiert.

### Zubereitungsstatus

Status einer Position an der Ausgabestation: offen → in Zubereitung → fertig.

> **Hinweis:** Noch nicht implementiert.

### Ausgabestation

Physischer Ort (Küche, Getränketheke), an dem Positionen zubereitet und ausgegeben werden.

> **Hinweis:** Noch nicht implementiert.

## Abrechnung (Supporting Sub-Domain)

### Tagesabrechnung

Übersicht über Gesamtumsatz, Stornierungen und Umsatz pro Servicekraft — jederzeit vom Admin abrufbar.

> **Hinweis:** Noch nicht implementiert.

### Umsatz

Summe aller registrierten Zahlungen in einem bestimmten Zeitraum. Immer in Cent.

> **Hinweis:** Noch nicht implementiert.

### Stornoquote

Verhältnis von Stornierungsbetrag zu Bestellsumme. Indikator für Fehler oder Unregelmäßigkeiten.

> **Hinweis:** Noch nicht implementiert.

### Tagesabschluss

Administrativer Vorgang zum Ende einer Veranstaltung: offene Tische prüfen, Abschlussbericht generieren, optional System zurücksetzen.

> **Hinweis:** Noch nicht implementiert.

### Export

CSV-Download von Umsätzen, Bestellungen und Artikeldaten für die Vereinsbuchhaltung.

> **Hinweis:** Noch nicht implementiert.

## Übergreifende Prinzipien

### Event-Sourcing

Persistenzmuster für den Kassenbetrieb: Zustand wird nicht direkt gespeichert, sondern aus unveränderlichen Events berechnet. Jeder Tisch hat einen eigenen Event Stream. Der aktuelle Zustand (Saldo, offene Positionen) ergibt sich aus dem Replay aller Events.

### Fat Event

Event, das alle relevanten Daten zum Zeitpunkt der Aktion enthält — inklusive Produktname und Preis. Damit ist das Event unabhängig von späteren Stammdaten-Änderungen auswertbar.

### Anti-Corruption Layer (ACL)

Schutzmechanismus zwischen Bounded Contexts: Der Kassenbetrieb friert Stammdaten (Produktname, Preis) in Events ein und ist damit unabhängig von nachträglichen Änderungen an Produkten oder Varianten.

### Append-only

Grundprinzip des Event Streams: Events werden nur hinzugefügt, nie geändert oder gelöscht. Dies gilt ohne Ausnahme — auch nicht für Korrekturen. Falsche Aktionen werden durch kompensierende Events (z. B. Stornierung) aufgehoben.

### Snapshot

Vorberechneter Zwischenstand des Tisch-Zustands. Rein technische Performance-Optimierung ohne fachliche Bedeutung — beim Replay wird nur ab dem letzten Snapshot gelesen statt ab dem ersten Event.

> **Hinweis:** Der Entwurf beschreibt Snapshots als „separat vom Event Stream gespeichert". In der Implementierung werden Snapshots als eigener Event-Typ (`tisch.snapshot:v1`) in der `events`-Tabelle gespeichert — also innerhalb des Event Streams. Diese Abweichung ist eine bewusste Vereinfachung der Persistenzschicht.

### BYOD

Bring Your Own Device — Servicekräfte nutzen ihre eigenen Smartphones. Das System ist Mobile-first konzipiert und läuft vollständig im Browser, ohne App-Installation.

## Abweichungen: Ist-Zustand vs. Soll-Zustand

Die folgende Tabelle dokumentiert Abweichungen zwischen den aktuellen Code-Bezeichnungen und den durch die Ubiquitous Language definierten Soll-Bezeichnungen.

### Handlungsbedarf (Refactoring geplant)

| Begriff   | Ist (Code)   | Soll         | Priorität | Scope                                                                      |
| --------- | ------------ | ------------ | --------- | -------------------------------------------------------------------------- |
| Produkt   | `Product`    | `Produkt`    | Mittel    | Go-Struct, TS-Typ, JSON-Keys. API-Pfade bereits korrekt.                   |
| Variante  | `Variant`    | `Variante`   | Mittel    | Go-Struct, TS-Typ, JSON-Keys. API-Pfade bereits korrekt.                   |
| Kategorie | `Category`   | `Kategorie`  | Mittel    | Go-Typ + Konstanten, TS-Typ.                                               |
| Preis     | `PriceCents` | `PreisCents` | Mittel    | Go-Feld + JSON-Key in `Variant`. Position ist bereits korrekt.             |
| Kommentar | `Comment`    | `Kommentar`  | Mittel    | Go-Feld, JSON-Key, TS-Feld in Bestellung, Zahlung, Lieferung, Stornierung. |
| Menge     | `Quantity`   | `Menge`      | Mittel    | Go-Feld, JSON-Key, TS-Feld in Position.                                    |

> **Hinweis:** Alle Änderungen sind Breaking Changes für die API (JSON-Keys ändern sich). Frontend und Backend müssen koordiniert umgestellt werden.

### Kein Handlungsbedarf (bewusst korrekt)

| Bereich         | Ist (Code)                                        | Begründung                                                                                      |
| --------------- | ------------------------------------------------- | ----------------------------------------------------------------------------------------------- |
| DB-Tabellen     | `tables`, `products`, `users`, `product_variants` | Englisch ist korrekt — DB-Schicht ist Infrastruktur.                                            |
| Frontend-Routen | `/admin/products`, `/service/tables`              | Englisch ist korrekt — Routen sind Infrastruktur.                                               |
| Auth-Code       | `User`, `Role`, `OnetimePassword`                 | Englisch ist korrekt — Generic Sub-Domain.                                                      |
| Status-Enums    | `active`, `inactive`, `deleted`                   | Englisch ist korrekt — DB-Enums sind Infrastruktur.                                             |
| Kassenjournal   | `Historie` (Code) vs. `Kassenjournal` (Entwurf)   | Bewusste Abweichung: „Historie" ist im Code und UI etabliert, beide Begriffe sind dokumentiert. |

## Änderungshistorie

| Datum      | Änderung                                                             |
| ---------- | -------------------------------------------------------------------- |
| 2026-03-12 | Initiale Version erstellt aus Entwurf Abschnitt 12 und Code-Analyse. |
