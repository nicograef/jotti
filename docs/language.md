# Ubiquitous Language — jotti

Dieses Dokument ist die **verbindliche Referenz** für die Ubiquitous Language des jotti-Projekts — für Entwickler, Agenten und alle Projektbeteiligten. Es definiert die Fachbegriffe der Domäne, ihre Code-Repräsentationen und die Sprachkonventionen pro Schicht.

Die Ubiquitous Language ist ein **Living Document**: Sie wird fortlaufend aktualisiert, wenn sich Begriffe, Strukturen oder Konventionen ändern. Die vollständige und aktuelle Definition wird hier gepflegt.

## Sprachkonventionen

1. **Domänenbegriffe sind deutsch.** Alle Fachbegriffe des Kassenbetriebs, der Stammdaten und der Gastronomie-Domäne werden auf Deutsch benannt — in Code, Dokumentation und Kommunikation. Beispiele: `Bestellung`, `Tisch`, `Zahlung`, `Position`, `Ausgabe`, `Stornierung`, `Saldo`.

2. **Infrastruktur-Code bleibt englisch.** Authentifizierung, Konfiguration, HTTP-Framework und generische Sub-Domains verwenden englische Bezeichnungen. Beispiele: `User`, `Role`, `Token`, `Config`, `Middleware`. Technische Felder (z. B. `created_at`, `status`, `id`) bleiben in allen Schichten englisch.

3. **Benutzer-sichtbare Strings sind deutsch.** Alle UI-Labels, Fehlermeldungen, Platzhalter und Hilfetexte werden auf Deutsch formuliert. Im UI heißt es „Benutzer" (nicht „User"), „Einmalpasswort" (nicht „OnetimePassword"), „Getränke" (nicht „beverage").

4. **Commits sind auf Englisch.** Conventional Commits (`feat:`, `fix:`, `refactor:`, `docs:`, `test:`) mit englischen Nachrichten.

## Namenskonventionen pro Schicht

| Schicht                   | Sprache  | Konvention                 | Beispiel                                      |
| ------------------------- | -------- | -------------------------- | --------------------------------------------- |
| Go-Domain-Structs         | Deutsch  | PascalCase                 | `Bestellung`, `Tisch`, `Position`             |
| Go-Felder (Domäne)        | Deutsch  | PascalCase                 | `GesamtPreisCents`, `SaldoCents`              |
| TypeScript-Typen (Domäne) | Deutsch  | PascalCase                 | `Bestellung`, `Tisch`, `Zahlung`              |
| JSON-Keys (Domäne)        | Deutsch  | camelCase                  | `"gesamtPreisCents"`, `"saldoCents"`          |
| API-Pfade (Domäne)        | Deutsch  | kebab-case                 | `/bestellung-aufnehmen`, `/zahlung-kassieren` |
| DB-Tabellen (Domäne)      | Deutsch  | snake_case                 | `tische`, `produkte`, `produkt_varianten`     |
| DB-Tabellen (Infrastr.)   | Englisch | snake_case                 | `users`, `events`                             |
| DB-Spalten (Domäne)       | Deutsch  | snake_case                 | `kategorie`, `preis_cents`, `produkt_id`      |
| DB-Spalten (Infrastr.)    | Englisch | snake_case                 | `created_at`, `updated_at`, `status`, `id`    |
| Frontend-Routen           | Englisch | kebab-case                 | `/service/tables`, `/admin/products`          |
| Auth/Infrastruktur-Code   | Englisch | Sprachübliche Konventionen | `User`, `Role`, `Token`, `Config`             |

> **Pfadkonvention:** Dateipfade in den Tabellen sind relativ angegeben — `domain/…` und `api/…` liegen unter `backend/`, `src/…` unter `frontend/`, `migrations/…` unter `database/`.

## Abweichungen: Ist-Zustand vs. Soll-Zustand

Die folgende Tabelle dokumentiert Abweichungen zwischen den aktuellen Code-Bezeichnungen und den durch die Ubiquitous Language definierten Soll-Bezeichnungen.

### Handlungsbedarf (Backend behoben, Frontend-Rename ausstehend)

| Begriff   | Ist (Code)   | Soll         | Status                  | Scope                                            |
| --------- | ------------ | ------------ | ----------------------- | ------------------------------------------------ |
| Produkt   | `Product`    | `Produkt`    | ✅ Backend, ⏳ Frontend | Go-Struct, JSON-Keys behoben. TS-Typ noch offen. |
| Variante  | `Variant`    | `Variante`   | ✅ Backend, ⏳ Frontend | Go-Struct, JSON-Keys behoben. TS-Typ noch offen. |
| Kategorie | `Category`   | `Kategorie`  | ✅ Backend, ⏳ Frontend | Go-Typ + Konstanten behoben. TS-Typ noch offen.  |
| Preis     | `PriceCents` | `PreisCents` | ✅ Backend, ⏳ Frontend | Go-Feld + JSON-Key behoben. TS-Feld noch offen.  |
| Kommentar | `Comment`    | `Kommentar`  | ✅ Backend, ⏳ Frontend | Go-Feld, JSON-Key behoben. TS-Feld noch offen.   |
| Menge     | `Quantity`   | `Menge`      | ✅ Backend, ⏳ Frontend | Go-Feld, JSON-Key behoben. TS-Feld noch offen.   |

> **Hinweis:** Alle Änderungen sind Breaking Changes für die API (JSON-Keys ändern sich). Frontend und Backend müssen koordiniert umgestellt werden.

### Kein Handlungsbedarf (bewusst korrekt)

| Bereich              | Ist (Code)                                      | Begründung                                                                                      |
| -------------------- | ----------------------------------------------- | ----------------------------------------------------------------------------------------------- |
| DB-Tabellen (Infra)  | `users`, `events`                               | Englisch ist korrekt — Infrastruktur / Generic Sub-Domain.                                      |
| DB-Tabellen (Domain) | `tische`, `produkte`, `produkt_varianten`       | Deutsch ist korrekt — Domänenbegriffe sind vertikal konsistent.                                 |
| Frontend-Routen      | `/admin/products`, `/service/tables`            | Englisch ist korrekt — Routen sind Infrastruktur.                                               |
| Auth-Code            | `User`, `Role`, `OnetimePassword`               | Englisch ist korrekt — Generic Sub-Domain.                                                      |
| Status-Enums         | `active`, `inactive`, `deleted`                 | Englisch ist korrekt — technische Lifecycle-States, kein Domänenbegriff.                        |
| Kassenjournal        | `Historie` (Code) vs. `Kassenjournal` (Entwurf) | Bewusste Abweichung: „Historie" ist im Code und UI etabliert, beide Begriffe sind dokumentiert. |

## Kassenbetrieb (Core Domain)

### Tisch

Zentrales Aggregat im Kassenbetrieb. Trägt einen Event Stream, aus dem sich der aktuelle Zustand (Saldo, offene Positionen) berechnet.

| Schicht             | Repräsentation                                                                                                                 | Datei                          |
| ------------------- | ------------------------------------------------------------------------------------------------------------------------------ | ------------------------------ |
| Go-Struct           | `Tisch` (Felder: `ID`, `Name`, `Status`, `CreatedAt`)                                                                          | `domain/table/tisch.go`        |
| TypeScript-Typ      | `Tisch`                                                                                                                        | `src/service/table/Tisch.ts`   |
| DB-Tabelle          | `tables`                                                                                                                       | `migrations/01_initial.up.sql` |
| API-Pfade (Admin)   | `/create-tisch`, `/update-tisch`, `/activate-tisch`, `/deactivate-tisch`, `/get-all-tische`                                    | `api/admin.go`                 |
| API-Pfade (Service) | `/get-tisch`, `/get-aktive-tische`, `/get-tisch-historie`, `/get-tisch-saldo`, `/get-tisch-unbezahlt`, `/get-tisch-ausstehend` | `api/service.go`               |
| Frontend-Hooks      | `useTisch()`, `useAktiveTische()`, `useTischHistorie()`, `useTischSaldo()`, `useTischUnbezahlt()`, `useTischAusstehend()`      | `src/service/table/hooks.ts`   |
| Frontend-Route      | `/service/tables/:tableId`                                                                                                     | `src/routes.ts`                |

> **Hinweis:** DB-Tabelle (`tables`) und Frontend-Route (`/service/tables`) verwenden Englisch — korrekt per Infrastruktur-Konvention.

### Bestellung

Ein Vorgang, bei dem eine Servicekraft Positionen für einen Tisch aufnimmt. Erzeugt ein `BestellungAufgenommen`-Event.

| Schicht             | Repräsentation                                                                                                   | Datei                              |
| ------------------- | ---------------------------------------------------------------------------------------------------------------- | ---------------------------------- |
| Go-Struct           | `Bestellung` (Felder: `ID`, `UserID`, `TischID`, `Positionen`, `GesamtPreisCents`, `Kommentar`, `AufgenommenAm`) | `domain/table/bestellung.go`       |
| Go-Event-Typ        | `EventTypeBestellungAufgenommenV1` = `"tisch.bestellung-aufgenommen:v1"`                                         | `domain/table/events.go`           |
| Go-Command          | `BestellungAufnehmen()`                                                                                          | `api/table/application/command.go` |
| TypeScript-Typ      | `Bestellung`, `BestellungAufnehmen`                                                                              | `src/service/table/Bestellung.ts`  |
| API-Pfad            | `/service/bestellung-aufnehmen`                                                                                  | `api/service.go`                   |
| Frontend-Komponente | `<Bestellung>`, `<BestellungDrawer>`                                                                             | `src/service/components/table/`    |
| UI-Labels           | „Bestellen" (Tab), „Bestellung aufnehmen" (Button), „Bestellung wurde aufgenommen." (Toast)                      |                                    |

### Position

Ein einzelner Posten innerhalb einer Bestellung: Produktvariante + Menge + Einzelpreis.

| Schicht        | Repräsentation                                                                                                      | Datei                             |
| -------------- | ------------------------------------------------------------------------------------------------------------------- | --------------------------------- |
| Go-Struct      | `Position` (Felder: `PositionID`, `VarianteID`, `ProduktName`, `VarianteName`, `Kategorie`, `Einzelpreis`, `Menge`) | `domain/table/bestellung.go`      |
| JSON-Keys      | `"positionId"`, `"varianteId"`, `"produktName"`, `"varianteName"`, `"kategorie"`, `"einzelpreis"`, `"menge"`        | `domain/table/bestellung.go`      |
| TypeScript-Typ | `Position`                                                                                                          | `src/service/table/Bestellung.ts` |

> **Hinweis:** Die Position wurde komplett redesigned (Fat Events). Alle Felder nutzen deutsche Ubiquitous Language.

### Ausgabe

Die Bestätigung, dass bestellte Positionen dem Gast übergeben wurden. Erzeugt ein `AusgabeBestaetigt`-Event.

| Schicht             | Repräsentation                                                                           | Datei                                      |
| ------------------- | ---------------------------------------------------------------------------------------- | ------------------------------------------ |
| Go-Struct           | `Ausgabe` (Felder: `ID`, `UserID`, `TischID`, `Positionen`, `Kommentar`, `AusgegebenAm`) | `domain/table/ausgabe.go`                  |
| Go-Event-Typ        | `EventTypeAusgabeBestaetigtV1` = `"tisch.ausgabe-bestaetigt:v1"`                         | `domain/table/events.go`                   |
| Go-Command          | `AusgabeBestaetigen()`                                                                   | `api/table/application/command.go`         |
| TypeScript-Typ      | `Ausgabe`, `AusgabeBestaetigen`                                                          | `src/service/table/Ausgabe.ts`             |
| API-Pfad            | `/service/ausgabe-bestaetigen`                                                           | `api/service.go`                           |
| Frontend-Komponente | `<Ausgabe>`                                                                              | `src/service/components/table/Ausgabe.tsx` |
| UI-Labels           | „Ausgabe bestätigen" (Button), „Ausgabe" (Historie)                                      |                                            |

### Zahlung

Die Kassierung einer Barzahlung. Kann sich auf einzelne Positionen beziehen (Teilzahlung). Erzeugt ein `ZahlungKassiert`-Event.

| Schicht             | Repräsentation                                                                                               | Datei                              |
| ------------------- | ------------------------------------------------------------------------------------------------------------ | ---------------------------------- |
| Go-Struct           | `Zahlung` (Felder: `ID`, `UserID`, `TischID`, `Positionen`, `GesamtZahlungCents`, `Kommentar`, `KassiertAm`) | `domain/table/zahlung.go`          |
| Go-Event-Typ        | `EventTypeZahlungKassiertV1` = `"tisch.zahlung-kassiert:v1"`                                                 | `domain/table/events.go`           |
| Go-Command          | `ZahlungKassieren()`                                                                                         | `api/table/application/command.go` |
| TypeScript-Typ      | `Zahlung`, `ZahlungKassieren`                                                                                | `src/service/table/Zahlung.ts`     |
| API-Pfad            | `/service/zahlung-kassieren`                                                                                 | `api/service.go`                   |
| Frontend-Komponente | `<ZahlungDrawer>`                                                                                            | `src/service/components/table/`    |
| UI-Labels           | „Bezahlen" (Tab), „Kassieren" (Button), „Zahlung kassieren" (actionLabel)                                    |                                    |

### Stornierung

Die nachträgliche Aufhebung bestellter Positionen. Nur durch Serviceleitung oder Admin. Erzeugt ein `StornierungErteilt`-Event.

| Schicht             | Repräsentation                                                                                                        | Datei                                                |
| ------------------- | --------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------- |
| Go-Struct           | `Stornierung` (Felder: `ID`, `UserID`, `TischID`, `Positionen`, `GesamtStornierungCents`, `Kommentar`, `StorniertAm`) | `domain/table/stornierung.go`                        |
| Go-Event-Typ        | `EventTypeStornierungErteiltV1` = `"tisch.stornierung-erteilt:v1"`                                                    | `domain/table/events.go`                             |
| Go-Command          | `StornierungErteilen()`                                                                                               | `api/table/application/command.go`                   |
| TypeScript-Typ      | `Stornierung`, `StornierungErteilen`                                                                                  | `src/service/table/Stornierung.ts`                   |
| API-Pfad            | `/serviceleitung/stornierung-erteilen`                                                                                | `api/serviceleitung.go`                              |
| Frontend-Komponente | `<StornierungDrawer>`                                                                                                 | `src/service/components/table/StornierungDrawer.tsx` |
| UI-Labels           | „Stornierung" (Drawer-Titel), „Stornierung erteilen" (Button), „Stornierung erfolgreich." (Toast)                     |                                                      |

> **Hinweis:** `Kommentar` ist bei Stornierungen **Pflichtfeld** (min. 3, max. 100 Zeichen).

### Auszahlung

Auszahlung eines Betrags an den Gast, um einen negativen Saldo auszugleichen — entsteht, wenn bereits kassierte Positionen nachträglich storniert werden (K-05). Kein Positionsbezug; freier Betrag (≥ 1 Cent). Erzeugt ein `AuszahlungGeleistet`-Event.

| Schicht             | Repräsentation                                                                              | Datei                                               |
| ------------------- | ------------------------------------------------------------------------------------------- | --------------------------------------------------- |
| Go-Struct           | `Auszahlung` (Felder: `ID`, `UserID`, `TischID`, `BetragCents`, `Kommentar`, `GeleistetAm`) | `domain/table/auszahlungGeleistetEvent.go`          |
| Go-Event-Typ        | `EventTypeAuszahlungGeleistetV1` = `"tisch.auszahlung-geleistet:v1"`                        | `domain/table/events.go`                            |
| Go-Command          | `AuszahlungLeisten()`                                                                       | `api/table/application/command.go`                  |
| API-Pfad            | `/serviceleitung/auszahlung-leisten`                                                        | `api/serviceleitung.go`                             |
| Frontend-Komponente | `<AuszahlungDrawer>`                                                                        | `src/service/components/table/AuszahlungDrawer.tsx` |
| UI-Labels           | „Auszahlung leisten“ (Button), negativer Saldo-Badge in Tischkarte und Tischseite           |                                                     |

> **Hinweis:** `Kommentar` ist Pflichtfeld (min. 3, max. 100 Zeichen). Das UI befüllt den Betrag vor, wenn der Saldo negativ ist.

### Saldo

Der offene Betrag eines Tisches: Summe der Bestellungen − Summe der Zahlungen − Summe der Stornierungen + Summe der Auszahlungen. Immer in Cent.

| Schicht          | Repräsentation                | Datei                        |
| ---------------- | ----------------------------- | ---------------------------- |
| Go-Projektion    | `ApplyEvent()` → `TischState` | `domain/table/projection.go` |
| Go-Snapshot-Feld | `SaldoCents`                  | `domain/table/projection.go` |
| API-Pfad         | `/service/get-tisch-state`    | `api/service.go`             |
| Frontend-Hook    | `useTischState()`             | `src/service/table/hooks.ts` |
| UI-Label         | „offen" (Badge)               |                              |

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

Freitextnotiz zu Tischoperationen. Pflichtfeld bei Stornierung und Auszahlung (min. 3 Zeichen), optional bei Bestellung, Ausgabe und Zahlung. Max. 100 Zeichen.

| Schicht       | Repräsentation | Datei                                                                                                       |
| ------------- | -------------- | ----------------------------------------------------------------------------------------------------------- |
| Go-Feld       | `Kommentar`    | `domain/table/bestellung.go`, `zahlung.go`, `ausgabe.go`, `stornierung.go`, `bestellungAufgenommenEvent.go` |
| JSON-Key      | `"kommentar"`  | (alle oben genannten Dateien)                                                                               |
| TS-Feld (Ist) | `comment`      | `src/service/table/Bestellung.ts`, `Zahlung.ts`, `Ausgabe.ts`, `Stornierung.ts`                             |

> **Hinweis:** Backend-Rename abgeschlossen (`Comment` → `Kommentar`). Frontend-Rename (`comment` → `kommentar`) ausstehend.

### Menge

Anzahl einer Produktvariante innerhalb einer Position.

| Schicht       | Repräsentation | Datei                             |
| ------------- | -------------- | --------------------------------- |
| Go-Feld       | `Menge`        | `domain/table/bestellung.go`      |
| JSON-Key      | `"menge"`      | `domain/table/bestellung.go`      |
| TS-Feld (Ist) | `quantity`     | `src/service/table/Bestellung.ts` |

> **Hinweis:** Backend-Rename abgeschlossen (`Quantity` → `Menge`). Frontend-Rename (`quantity` → `menge`) ausstehend.

### EigeneUebersicht

Kompakte KPI-Übersicht einer Servicekraft über ihre eigenen Aktivitäten: Anzahl und Summe eigener Bestellungen sowie Anzahl und Summe eigener kassierten Zahlungen. Read Model — berechnet aus dem Event Store, gefiltert auf `user_id`.

| Schicht             | Repräsentation                                                                                              | Datei                                         |
| ------------------- | ----------------------------------------------------------------------------------------------------------- | --------------------------------------------- |
| Go-Struct           | `EigeneUebersicht` (Felder: `AnzahlBestellungen`, `BestellungenCents`, `AnzahlZahlungen`, `ZahlungenCents`) | `domain/reporting/reporting.go`               |
| JSON-Keys           | `"anzahlBestellungen"`, `"bestellungenCents"`, `"anzahlZahlungen"`, `"zahlungenCents"`                      | Response-DTO in `api/reporting/http/`         |
| TypeScript-Typ      | `EigeneUebersicht`                                                                                          | `src/service/table/Tisch.ts`                  |
| API-Pfad            | `/service/get-eigene-uebersicht`                                                                            | `api/service.go`                              |
| Frontend-Komponente | `<EigeneUebersicht>`                                                                                        | `src/service/components/EigeneUebersicht.tsx` |
| UI-Label            | „Meine Übersicht" (Sektion), „Bestellungen" / „Kassiert" (Karten)                                           |                                               |

## Stammdaten (Supporting Sub-Domain)

### Produkt

Artikel im Produktkatalog. Gehört zu genau einer Kategorie und enthält eine oder mehrere Varianten mit je eigenem Preis.

| Schicht        | Repräsentation                                                                                 | Datei                           |
| -------------- | ---------------------------------------------------------------------------------------------- | ------------------------------- |
| Go-Struct      | `Produkt` (Felder: `ID`, `Name`, `Kategorie`, `Status`, `Varianten`, `CreatedAt`, `UpdatedAt`) | `domain/product/product.go`     |
| DB-Tabelle     | `produkte`                                                                                     | `migrations/01_initial.up.sql`  |
| TypeScript-Typ | `Produkt`                                                                                      | `src/admin/products/Product.ts` |
| API-Pfade      | `/create-produkt`, `/update-produkt`, `/get-all-produkte`                                      | `api/admin.go`                  |

> **Hinweis:** Backend- und DB-Rename abgeschlossen. Frontend-Typ-Rename (`Product` → `Produkt`) ausstehend.

### Variante

Konkrete Ausprägung eines Produkts mit eigenem Namen und Preis in Cent (z. B. Produkt „Cola" → Varianten „0,3 l" und „0,5 l").

| Schicht        | Repräsentation                                                                       | Datei                           |
| -------------- | ------------------------------------------------------------------------------------ | ------------------------------- |
| Go-Struct      | `Variante` (Felder: `ID`, `Name`, `PreisCents`, `Status`, `CreatedAt`, `UpdatedAt`)  | `domain/product/variant.go`     |
| DB-Tabelle     | `produkt_varianten`                                                                  | `migrations/01_initial.up.sql`  |
| TypeScript-Typ | `Variante`                                                                           | `src/admin/products/Product.ts` |
| API-Pfade      | `/create-variante`, `/update-variante`, `/activate-variante`, `/deactivate-variante` | `api/admin.go`                  |

> **Hinweis:** Backend- und DB-Rename abgeschlossen. Frontend-Typ-Rename (`Variant` → `Variante`) ausstehend.

### Kategorie

Gruppierung von Produkten. Aktuell drei feste Kategorien: Essen, Getränke, Sonstiges.

| Schicht         | Repräsentation                                                                         | Datei                            |
| --------------- | -------------------------------------------------------------------------------------- | -------------------------------- |
| Go-Typ          | `Kategorie` mit Konstanten `EssenKategorie`, `GetraenkKategorie`, `SonstigesKategorie` | `domain/product/product.go`      |
| DB-Enum         | `ProduktKategorie` (`'essen'`, `'getraenk'`, `'sonstiges'`)                            | `migrations/01_initial.up.sql`   |
| Frontend-Werte  | `Kategorie.ESSEN`, `Kategorie.GETRAENK`, `Kategorie.SONSTIGES`                         | `src/admin/products/Product.ts`  |
| Frontend-Labels | `'essen'` → „Essen", `'getraenk'` → „Getränke", `'sonstiges'` → „Sonstiges"            | `src/service/product/Product.ts` |

> **Hinweis:** Backend-, DB- und Frontend-Werte-Rename abgeschlossen.

### Preis

Geldbeträge werden ausnahmslos als ganzzahlige Cent-Werte gespeichert — niemals als Fließkommazahlen. 3,50 € = 350 Cent.

**Konvention:** Alle Preis-Felder tragen das Suffix `*Cents` (z. B. `PreisCents`, `GesamtPreisCents`, `SaldoCents`, `GesamtZahlungCents`, `GesamtStornierungCents`).

> **Hinweis:** Im Kassenbetrieb und bei Varianten wird durchgängig `PreisCents` (deutsch) verwendet.

### Soft-Delete

Logisches Löschen: Datensätze werden nicht physisch entfernt, sondern durch den Status `deleted` als gelöscht markiert. Der Datensatz bleibt für Referenzintegrität und historische Auswertungen erhalten.

| Schicht              | Repräsentation                                         | Datei                          |
| -------------------- | ------------------------------------------------------ | ------------------------------ |
| DB-Enum              | `EntityStatus` (`'active'`, `'inactive'`, `'deleted'`) | `migrations/01_initial.up.sql` |
| Go-Domain (Tisch)    | Konstanten `ActiveStatus`, `InactiveStatus`            | `domain/table/tisch.go`        |
| Go-Domain (Variante) | Konstanten `ActiveStatus`, `InactiveStatus`            | `domain/product/variant.go`    |
| Go-Domain (User)     | Konstanten `ActiveStatus`, `InactiveStatus`            | `domain/user/user.go`          |

> **Hinweis:** In den Go-Domain-Modellen existieren nur `active` und `inactive` als Konstanten. Der Status `deleted` wird ausschließlich auf DB-Ebene verwendet und ist in der Domain nicht als Konstante abgebildet.

### Favorit

Eine benutzerspezifische Markierung, die eine Servicekraft für einen Tisch setzt, um diesen auf dem Service-Dashboard als Rich Card anzuzeigen ("Meine Tische"). Kein Aggregat, keine Events — einfache CRUD-Relation in der DB (`tisch_favoriten`: `user_id` + `tisch_id` als Composite PK).

| Schicht        | Repräsentation                                                                     | Datei                             |
| -------------- | ---------------------------------------------------------------------------------- | --------------------------------- |
| Go-Package     | `backend/repository/favorit_repo/`                                                 | `repository/favorit_repo/repo.go` |
| DB-Tabelle     | `tisch_favoriten`                                                                  | `migrations/01_initial.up.sql`    |
| TypeScript-Typ | (kein eigener Typ — als `istFavorit: boolean` in `AktiverTischMitFavorit`)         | `src/service/table/Tisch.ts`      |
| API-Pfade      | `/service/favorit-hinzufuegen`, `/service/favorit-entfernen`                       | `api/service.go`                  |
| UI-Label       | „Meine Tische" (Dashboard-Überschrift), Stern-Toggle (★ / ☆) im Alle-Tische-Drawer |                                   |

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
| Serviceleitung | `serviceleitung` | Service-Funktionen + Stornierung                      |
| Servicekraft   | `service`        | Bestellen, Ausgabe bestätigen, Kassieren              |

| Schicht  | Repräsentation                                              | Datei                          |
| -------- | ----------------------------------------------------------- | ------------------------------ |
| Go-Typ   | `Role` mit `AdminRole`, `ServiceleitungRole`, `ServiceRole` | `domain/user/user.go`          |
| DB-Enum  | `UserRole` (`'admin'`, `'serviceleitung'`, `'service'`)     | `migrations/01_initial.up.sql` |
| JWT-Feld | `role`                                                      | `src/lib/Auth.ts`              |

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

> **Hinweis:** In der Implementierung werden Snapshots als eigener Event-Typ (`tisch.snapshot:v1`) in der `events`-Tabelle gespeichert — eine bewusste Vereinfachung der Persistenzschicht.

### BYOD

Bring Your Own Device — Servicekräfte nutzen ihre eigenen Smartphones. Das System ist Mobile-first konzipiert und läuft vollständig im Browser, ohne App-Installation.

## Geplant (nicht implementiert)

Die folgenden Begriffe sind in der Ubiquitous Language definiert, aber noch nicht im Code implementiert.

### Ausgabe (Supporting Sub-Domain)

| Begriff                 | Bedeutung                                                                                                       |
| ----------------------- | --------------------------------------------------------------------------------------------------------------- |
| **Bon**                 | Gedruckter Beleg mit Tisch, Servicekraft, Positionen, Mengen, Zeitstempel und optionalem Kommentar.             |
| **Küchendisplay (KDS)** | Echtzeit-Anzeige offener Bestellungen an der Ausgabestation, gruppiert nach Tisch und gefiltert nach Kategorie. |
| **Zubereitungsstatus**  | Status einer Position an der Ausgabestation: offen → in Zubereitung → fertig.                                   |
| **Ausgabestation**      | Physischer Ort (Küche, Getränketheke), an dem Positionen zubereitet und ausgegeben werden.                      |

### Abrechnung (Supporting Sub-Domain)

| Begriff             | Bedeutung                                                                                                                              |
| ------------------- | -------------------------------------------------------------------------------------------------------------------------------------- |
| **Tagesabrechnung** | Übersicht über Gesamtumsatz, Stornierungen und Umsatz pro Servicekraft — jederzeit vom Admin abrufbar.                                 |
| **Umsatz**          | Summe aller registrierten Zahlungen in einem bestimmten Zeitraum. Immer in Cent.                                                       |
| **Stornoquote**     | Verhältnis von Stornierungsbetrag zu Bestellsumme. Indikator für Fehler oder Unregelmäßigkeiten.                                       |
| **Tagesabschluss**  | Administrativer Vorgang zum Ende einer Veranstaltung: offene Tische prüfen, Abschlussbericht generieren, optional System zurücksetzen. |
| **Export**          | CSV-Download von Umsätzen, Bestellungen und Artikeldaten für die Vereinsbuchhaltung.                                                   |

---

## Fiskalkonformität (Compliance Sub-Domain)

Begriffe für die gesetzlich vorgeschriebene Fiskalisierung nach § 146a AO und KassenSichV. Diese Sub-Domain wird phasenweise implementiert — siehe `docs/roadmap.md`.

> **Sprachkonvention:** Fiskal-Fachbegriffe folgen der deutschen Gesetzessprache und DSFinV-K-Spezifikation. Technische Interface-Namen bleiben englisch (Go-Konvention).

### Kasse und Identifikation

| Begriff                                         | Go-Struct / Go-Typ    | DB-Tabelle / -Feld                      | JSON-Key            | Bedeutung                                                                                                                                                                    |
| ----------------------------------------------- | --------------------- | --------------------------------------- | ------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Seriennummer** / **KassenID**                 | `KassenID` (`string`) | `system_config.value` (key=`kassen_id`) | `kassen_id`         | Eindeutige UUID-v4 der jotti-Instanz. Beim ersten Containerstart generiert, dauerhaft gespeichert. Wird für ELSTER-Meldung und TSE-Protokoll benötigt.                       |
| **TSE** / **Technische Sicherheitseinrichtung** | Interface `TSEClient` | —                                       | —                   | Zertifiziertes Sicherheitsmodul nach BSI TR-03153. Signiert und protokolliert jeden Kassiervorgang kryptografisch. In jotti über ein Adapter-Pattern eingebunden (BYOT).     |
| **TSEClient**                                   | Interface (Go)        | —                                       | —                   | Go-Interface mit Methoden `StartTransaction`, `UpdateTransaction`, `FinishTransaction`. Wird von anbieter-spezifischen Implementierungen (z. B. `FiskalyTSEClient`) erfüllt. |
| **Signaturzähler**                              | `uint64`              | —                                       | `signature_counter` | Fortlaufender Zähler der TSE für jede signierte Transaktion. Pflichtfeld auf dem Kassenbeleg.                                                                                |

### Abrechnungsstruktur

| Begriff              | Go-Struct            | DB-Tabelle         | JSON-Key           | Bedeutung                                                                                                                               |
| -------------------- | -------------------- | ------------------ | ------------------ | --------------------------------------------------------------------------------------------------------------------------------------- |
| **ABRECHNUNGSKREIS** | `Abrechnungskreis`   | `abrechnungskreis` | `abrechnungskreis` | Fortlaufend nummerierte Kassensitzung, die einen Abrechnungszeitraum (typisch: einen Veranstaltungstag) abgrenzt. DSFinV-K-Pflichtfeld. |
| **Tagesabschluss**   | (s. Abrechnung oben) | —                  | —                  | Administrativer Abschluss eines ABRECHNUNGSKREIS. Erzeugt DSFinV-K-exportierbaren Datensatz.                                            |

### Steuern

| Begriff          | Go-Typ              | DB-Feld      | JSON-Key       | Bedeutung                                                                                                                                          |
| ---------------- | ------------------- | ------------ | -------------- | -------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Steuersatz**   | `Steuersatz` (Enum) | `steuersatz` | `steuersatz`   | Steuerklasse eines Produkts. Enum-Werte: `standard` (19 %), `ermaessigt` (7 %), `befreit` (0 %). Wird als Fat Event in die Bestellung eingefroren. |
| **Steuerbetrag** | `int` (Cent)        | —            | `steuerbetrag` | Berechneter Steuerbetrag in Cent für eine Position oder einen Vorgang. Immer in Cent, niemals als Float.                                           |
| **Nettobetrag**  | `int` (Cent)        | —            | `nettobetrag`  | Betrag vor Steuerabzug. Immer in Cent.                                                                                                             |

### Export und Meldung

| Begriff         | Bedeutung                                                                                                                                                                                                                                                                                             |
| --------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **DSFinV-K**    | „Digitale Schnittstelle der Finanzverwaltung für Kassensysteme" — standardisiertes CSV-ZIP-Exportformat (Version 2.4) für Betriebsprüfungen durch die Finanzverwaltung.                                                                                                                               |
| **ELSTER**      | Elektronisches Steuerportal der deutschen Finanzverwaltung. Jede jotti-Instanz muss innerhalb eines Monats nach Inbetriebnahme über ELSTER gemeldet werden (§ 146a Abs. 4 AO).                                                                                                                        |
| **ERiC**        | „ELSTER Rich Client" — Programmierschnittstelle (API) für die automatisierte ELSTER-Kommunikation. Phase-3-Feature.                                                                                                                                                                                   |
| **Kassenbeleg** | Pflichtbeleg, der nach jedem Kassiervorgang ausgestellt werden muss (§ 146a Abs. 2 AO). In jotti: Bondruck via ESC/POS + (Phase 2) TSE-Signaturfelder.                                                                                                                                                |
| **BYOT**        | „Bring Your Own TSE" — Betreiber schließen selbst einen Vertrag mit einem Cloud-TSE-Anbieter (z. B. fiskaly) ab und injizieren API-Schlüssel via `.env`-Datei.                                                                                                                                        |
| **GoBD**        | „Grundsätze zur ordnungsmäßigen Führung und Aufbewahrung von Büchern, Aufzeichnungen und Unterlagen in elektronischer Form" — Bundesfinanzministerium-Schreiben zu elektronischer Buchführung. Event-Sourcing erfüllt die Unveränderbarkeitsanforderung. Kryptografische Verkettung folgt in Phase 3. |
