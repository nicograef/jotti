# Plan: Audit Remediation

> Source PRD: n/a

## Goal

Die im Code-Audit identifizierten Probleme in einer reviewbaren Reihenfolge beheben: korrekte Mutations-Metadaten, lokal reproduzierbare Qualitaetspruefung, konsistente Frontend/Backend-Vertraege, einheitliche Validierung und reduzierte Wartungs- und Performance-Hotspots.

## Architectural decisions

Durable decisions that apply across all phases:

- **API contract**: Bestehende POST-only-Routen und Response-Huellen bleiben erhalten; Backend-DTOs bleiben die Source of Truth fuer JSON-Formate.
- **Response strictness**: Frontend-Response-Parsing wird repo-weit zur Laufzeit strikt; unbekannte Felder sollen nicht mehr still verworfen werden.
- **Validation**: Geschaeftsregeln werden zuerst im Backend verbindlich definiert und danach im Frontend gespiegelt; Geldbetraege bleiben Cent-int.
- **Verification**: `make verify` bleibt der lokale und CI-nahe Einstiegspunkt; das lokale Bootstrap installiert `golang-migrate` per CI-aehnlichem Release-Tarball in einen PATH-Pfad, und Integrationstests rufen `migrate` ueber den PATH statt ueber einen Repo-Symlink auf.
- **Persistence metadata**: `updated_at` wird bei jeder fachlichen Mutation aktiv aktualisiert und nicht implizit vom Repository repariert.
- **Reporting aggregation**: Reporting-Aggregation bleibt im SQL-Layer und dedupliziert wiederholte JSON-Extraktion ueber gemeinsame SQL-Hilfsfunktionen im Schema.
- **Refactor scope**: Kein neues API-Format, keine neuen Dependencies, keine Event-Format-Migrationen.

## Inventory

- `backend/domain/user/user.go:125-193` - User-Mutatoren veraendern Status, Passwort oder Stammdaten, setzen `UpdatedAt` aber nicht konsistent fort.
- `backend/api/user/application/command.go:47-168` - Admin-Use-Cases laden, mutieren und persistieren `user.User`-Objekte unveraendert weiter.
- `backend/repository/user_repo/repo.go:57-71` - Persistiert `UpdatedAt` direkt aus dem Domain-Modell ins SQL-Update.
- `backend/sqlc/queries/users.sql:17-18` - Das `users`-Update erwartet `updated_at` explizit aus dem Anwendungscode.
- `test-integration.sh:11-14` - Cleanup- und Migrationspfad haengen an `database/migrate/migrate`.
- `test-integration.sh:43-49` - Der Integrationstest fuehrt Migrationen ueber denselben festen Symlink-Pfad aus.
- `scripts/setup-dev-tools.sh:1-106` - Lokales Verify-Setup installiert mehrere Tools, aber kein `golang-migrate`.
- `.github/workflows/ci.yml:157-166` - CI installiert `golang-migrate` separat und zeigt das beabsichtigte Ausfuehrungsmodell.
- `backend/api/product/http/query_handler.go:21-40` - Produkt- und Varianten-DTOs liefern mehr Felder als die aktuellen Frontend-Schemas modellieren.
- `backend/api/table/http/query_handler.go:28-36` - Tisch-DTOs liefern `updatedAt`, das im Admin-Frontend aktuell nicht modelliert ist.
- `backend/api/user/http/query_handler.go:20-30` - Benutzer-DTOs liefern `updatedAt`, waehrend das Frontend nur einen Teil des Vertrags abbildet.
- `frontend/src/admin/products/Produkt.ts:24-47` - Admin-Produktschemas bilden Preisgrenzen und Response-Felder nicht vollstaendig ab.
- `frontend/src/admin/tables/Tisch.ts:19-24` - Admin-Tischschema deckt den Backend-Response-Vertrag nicht vollstaendig ab.
- `frontend/src/admin/users/User.ts:45-52` - Admin-Benutzerschema deckt den Backend-Response-Vertrag nicht vollstaendig ab.
- `frontend/src/admin/products/ProduktBackend.ts:49-68` - Admin-Backends parsen Collection-Responses ueber Zod-Wrapper direkt in den Aufrufpfaden.
- `frontend/src/service/table/TischBackend.ts:36-153` - Auch Service-Flows parsen Listen, Historie und State mit denselben nicht-strikten Response-Schemas.
- `frontend/src/admin/reporting/ReportingBackend.ts:11-18` - Reporting verwendet direkte Response-Schema-Pruefung ueber denselben Backend-Client.
- `frontend/src/lib/Backend.ts:105-110` - Response-Parsing nutzt nicht-strikte Zod-Objects und maskiert Contract-Drift durch Strippen unbekannter Keys.
- `frontend/src/service/table/Bestellung.ts:27-31` - Bestellung erlaubt leere Kommentare bis zur Maximalgrenze.
- `frontend/src/service/table/Ausgabe.ts:18-20` - Ausgabe folgt derselben lockeren Kommentarregel wie Bestellung.
- `frontend/src/service/table/Zahlung.ts:18-21` - Zahlung folgt derselben lockeren Kommentarregel wie Bestellung.
- `frontend/src/service/table/Stornierung.ts:18-21` - Stornierung erzwingt mindestens drei Zeichen Kommentar.
- `frontend/src/service/table/Auszahlung.ts:15-18` - Auszahlung erzwingt mindestens drei Zeichen Kommentar.
- `backend/api/auth/application/command.go:58-97` - `SetNewPassword` ruft `u.SetPassword()` auf und persistiert anschliessend, setzt `UpdatedAt` aber nicht.
- `backend/api/table/http/command_handler.go:244-388` - Backend nimmt Kommentarfelder fuer Service-Aktionen entgegen, erzwingt aber keine gemeinsame Policy.
- `backend/api/table/application/command.go:289-318` - `BestellungAufnehmen` reichert Positionen per Einzelabfrage an (2 Roundtrips pro Position: GetVariant + GetProduct).
- `backend/sqlc/queries/reporting.sql:7-105` - Reporting aggregiert Eventdaten mit mehrfach kopierter `CASE WHEN`-Logik.
- `backend/api/table/application/command.go:239-282` - Die Tisch-Statusbefehle sind ueber einen Callback-Helfer abstrahiert.
- `backend/api/table/application/errors.go:49-61` - `fromRepositoryError` kapselt Logging und Mapping mit wenig eigenem Domain-Wert.
- `backend/api/relay.go:14-44` - `druckerRepoRelayAdapter` fuegt eine schmale Adapter-Schicht fuer eine einzelne Typkonvertierung ein.
- `frontend/src/lib/Backend.ts:31-65` - `TokenGetter` fuehrt eine Ein-Implementierungs-Abstraktion im Frontend-Backend-Client ein.

## Resolved decisions

- Der Plan umfasst alle Audit-Funde in einer gemeinsamen Roadmap.
- Die Umsetzung wird in 5 duennen, reviewbaren Phasen geschnitten.
- Korrektheit und Reproduzierbarkeit kommen vor Performance- und Lesbarkeitsrefactors.
- Response-Contract-Drift wird repo-weit ueber striktes Runtime-Parsing im Frontend sichtbar gemacht, nicht nur in Admin-Hotspots.
- Die repo-weite Strictness wird in einer einzigen Phase ueber die gesamte Frontend-Response-Flaeche eingefuehrt.
- Reporting-Deduplizierung bleibt im SQL-Layer und nutzt gemeinsame SQL-Hilfsfunktionen im Schema.
- `deleted` bleibt aus den Frontend-User-Typen ausgeschlossen; abgesichert wird das ueber Backend-Query-Tests plus Frontend-Contract-Tests.
- Das lokale Bootstrap installiert `golang-migrate` per Release-Tarball analog zu CI in einen PATH-Pfad; `test-integration.sh` nutzt anschliessend `migrate` aus dem PATH.
- Aus Compliance-Gruenden sind Kommentare bei Stornierungen und Auszahlungen verpflichtend; bei Bestellungen, Ausgaben und Zahlungen bleiben sie optional.
- **Phase 1 (SetPassword-Scope):** `UpdatedAt`-Fix umfasst auch `backend/api/auth/application/command.go` (`SetNewPassword`), da der Auth-Pfad dieselbe Invariante brechen wuerde.
- **Phase 2 (Symlink):** Der Symlink `database/migrate/migrate` wird entfernt; das Verzeichnis `database/migrate/` bleibt fuer das Dockerfile erhalten.
- **Phase 3 (Strictness-Mechanismus):** Alle bekannten Backend-Felder werden explizit in den Zod-Schemas modelliert; `.strict()` wird nicht verwendet, um Robustheit bei kuenftigen Backend-Erweiterungen zu bewahren. Fehlende Felder werden in bestehenden Schemas ergaenzt.
- **Phase 3 (Preisgrenze):** Die Frontend-Obergrenze fuer `preisCents` wird auf `max(99999)` gesetzt ‒ entspricht `LTE(99999)` im Backend (`backend/domain/product/variant.go`).
- **Phase 5 (Batch-Strategie):** `BestellungAufnehmen` laedt alle benoetigten Produkte und Varianten in je einer Batch-Abfrage statt per Einzellookup; dafuer werden neue `sqlc`-Queries benoetigt.
- **Phase 5 (Reporting-SQL):** Wiederholte Event-Extraktionslogik wird als `CREATE FUNCTION`-Definitionen in `01_initial.up.sql` ausgelagert; bestehende Queries referenzieren danach die Hilfsfunktionen.
- **Phase 5 (`fromRepositoryError`):** Behalten ‒ konsolidiert strukturiertes Logging mit Tisch-ID-Kontext und Error-Mapping an einer einzigen Stelle; der Nutzen ueberwiegt die Indirektion.
- **Phase 5 (`druckerRepoRelayAdapter`):** Behalten ‒ vermeidet einen Import-Zyklus zwischen dem `relay/application`-Paket und `drucker_repo`; eine andere Loesung wuerde eine Paketumstrukturierung erfordern.
- **Phase 5 (`TokenGetter`-Interface):** Behalten ‒ ermoeglicht Unit-Tests der `Backend`-Klasse ohne echte Auth-Abhaengigkeit; eine Ein-Implementierungs-Abstraktion ist hier bewusst sinnvoll.
- **Phase 5 (`applyTischStatusChange`):** Entfernen ‒ die drei Status-Commands (`TischAktivieren`, `TischDeaktivieren`, `TischLoeschen`) werden direkt inline geschrieben; der Callback-Helfer verdeckt die tatsaechliche Logik ohne nennenswerte DRY-Ersparnis.

## Risks

- Die repo-weite Strictness-Umstellung kann zusaetzliche, bisher verdeckte Contract-Drift ausserhalb der auditierten Admin-Pfade sichtbar machen und damit den Umfang von Phase 3 vergroessern.
- SQL-Hilfsfunktionen fuer Reporting veraendern das DB-Schema und muessen sauber mit lokaler Dev- und Test-Datenbankinitialisierung zusammenspielen.

---

## Phase 1: Repair User Mutation Metadata

### Context

- `backend/domain/user/user.go:125-193` - User-Mutatoren aktualisieren `UpdatedAt` aktuell nicht.
- `backend/api/user/application/command.go:47-168` - Die User-Kommandos persistieren genau den Zustand, den die Domain liefert.
- `backend/api/auth/application/command.go:58-97` - `SetNewPassword` ruft `u.SetPassword()` auf und persistiert, setzt `UpdatedAt` aber ebenfalls nicht.
- `backend/repository/user_repo/repo.go:57-71` - Das Repository uebernimmt `UpdatedAt` unveraendert in die DB.
- `backend/sqlc/queries/users.sql:17-18` - SQL erwartet ein korrekt gesetztes `updated_at` aus dem Anwendungspfad.

### What to build

Eine vollstaendige Korrektur des User-Mutationspfads, sodass jede fachliche Aenderung an Benutzerstammdaten, Status oder Passwort einen neuen `updated_at`-Wert erzeugt und dieser Wert bis zur Query-Response konsistent durchgereicht wird. Der Fix umfasst sowohl die Admin-Commands (`user/application`) als auch den Auth-Pfad (`auth/application`), da beide dieselbe `UpdateUser`-Repository-Methode aufrufen.

### Acceptance criteria

- [x] `UpdateUser`, `ActivateUser`, `DeactivateUser`, `DeleteUser` und `ResetPassword` im Admin-Pfad erzeugen jeweils ein neueres `UpdatedAt`.
- [x] `SetNewPassword` im Auth-Pfad erzeugt ebenfalls ein neueres `UpdatedAt`.
- [x] Das Domain-Modell setzt `UpdatedAt = time.Now().UTC()` in den betroffenen Mutatoren; das Repository und die SQL-Query bleiben unveraendert.
- [x] Tests decken mindestens Stammdaten-, Status- und Passwort-Mutationen gegen `UpdatedAt` ab.
- [x] Nach einer Mutation liefern User-Queries den aktualisierten Zeitstempel zurueck.

---

## Phase 2: Make Local Verification Reproducible

### Context

- `test-integration.sh:11-14` - Cleanup und Down-Migration haengen an einem repo-lokalen Binary-Pfad.
- `test-integration.sh:43-49` - Der Up-Migrationsschritt nutzt denselben festen Pfad.
- `scripts/setup-dev-tools.sh:1-106` - Das lokale Bootstrap-Skript installiert Verify-Werkzeuge, aber nicht `golang-migrate`.
- `.github/workflows/ci.yml:157-166` - CI installiert `golang-migrate` explizit in den PATH und nutzt es dort direkt.

### What to build

Einen lokal reproduzierbaren Verify-Pfad, der in einer frischen Dev-Container-Umgebung ohne defekten Symlink auskommt: das Bootstrap installiert `golang-migrate` per CI-aehnlichem Release-Tarball in einen PATH-Pfad, der Integrationspfad ruft `migrate` direkt ueber den PATH auf, und der tote Symlink `database/migrate/migrate` wird entfernt.

### Acceptance criteria

- [x] `test-integration.sh` ruft `migrate` aus dem PATH auf statt ueber einen repo-eigenen Symlink.
- [x] Der Symlink `database/migrate/migrate` ist entfernt; das Verzeichnis `database/migrate/` (Dockerfile) bleibt erhalten.
- [x] Das lokale Setup installiert alle fuer `make verify` benoetigten Werkzeuge, inklusive `golang-migrate`, verbindlich im Bootstrap.
- [x] Eine frische lokale Umgebung kann die Verifikationskette bis mindestens zum Start der Integrationstests reproduzierbar ausfuehren.
- [x] Der lokale Migrationspfad und der CI-Migrationspfad nutzen dieselbe Installations- und Aufruflogik fuer `golang-migrate`.

---

## Phase 3: Align Frontend Contracts And Enforce Runtime Strictness

### Context

- `backend/api/product/http/query_handler.go:21-40` - Produkt-Responses enthalten `status` und `updatedAt` auf Produkt- und Variantenebene.
- `backend/api/table/http/query_handler.go:28-36` - Tisch-Responses enthalten `updatedAt`.
- `backend/api/user/http/query_handler.go:20-30` - User-Responses enthalten `status`, `createdAt` und `updatedAt`.
- `frontend/src/admin/products/ProduktBackend.ts:49-68` - Admin-Produkte liefern Response-Wrapper, die aktuell permissiv geparst werden.
- `frontend/src/admin/products/Produkt.ts:24-47` - Produkt- und Varianten-Schemas bilden den Backend-Vertrag und die Preisgrenze nicht vollstaendig ab.
- `frontend/src/admin/tables/Tisch.ts:19-24` - Tischschema ist schmaler als der DTO-Vertrag.
- `frontend/src/admin/users/User.ts:45-52` - Userschema ist schmaler als der DTO-Vertrag.
- `frontend/src/service/table/TischBackend.ts:36-153` - Service-Backends verwenden denselben Parsing-Mechanismus fuer Listen-, Historien- und State-Responses.
- `frontend/src/admin/reporting/ReportingBackend.ts:11-18` - Auch Reporting-Responses laufen ueber direkte Schema-Pruefung im gemeinsamen Backend-Client.
- `frontend/src/lib/Backend.ts:105-110` - Nicht-striktes Parsing verschluckt unbekannte Keys still.

### What to build

Eine vollstaendige Ausrichtung der Frontend-Response-Schemas auf die Backend-DTOs: alle fehlenden Felder werden in den bestehenden Zod-Schemas ergaenzt (kein `.strict()`), damit kuenftige Backend-Felder keinen Parse-Fehler ausloesen. Contract-Drift wird dadurch sichtbar, weil Felder, die der Backend-DTO liefert aber das Frontend-Schema nicht kennt, durch explizite Modellierung zukuenftig nicht mehr unbemerkt fehlen.

### Acceptance criteria

- [x] `VarianteSchema` ergaenzt `updatedAt`.
- [x] `ProduktSchema` ergaenzt `status` und `updatedAt`.
- [x] `TischSchema` ergaenzt `updatedAt`.
- [x] `UserSchema` ergaenzt `updatedAt`.
- [x] `PreisCentsSchema` setzt eine Obergrenze von `max(99999)` ‒ entspricht `LTE(99999)` im Backend (`backend/domain/product/variant.go`).
- [x] Die ergaenzten Schemas decken Admin-, Service- und Reporting-Backends ab (überall wo die Typen importiert werden).
- [x] Die Entscheidung zum Umgang mit dem `deleted`-User-Status ist ueber Backend-Query-Tests plus Frontend-Contract-Tests explizit abgesichert.

---

## Phase 4: Standardize Service Validation Rules

### Context

- `frontend/src/service/table/Bestellung.ts:27-31` - Bestellung erlaubt leere Kommentare.
- `frontend/src/service/table/Ausgabe.ts:18-20` - Ausgabe erlaubt leere Kommentare.
- `frontend/src/service/table/Zahlung.ts:18-21` - Zahlung erlaubt leere Kommentare.
- `frontend/src/service/table/Stornierung.ts:18-21` - Stornierung verlangt mindestens drei Zeichen.
- `frontend/src/service/table/Auszahlung.ts:15-18` - Auszahlung verlangt mindestens drei Zeichen.
- `backend/api/table/http/command_handler.go:244-388` - Backend nimmt alle Kommentarwerte entgegen, ohne eine gemeinsame Regel zu erzwingen.

### What to build

Eine compliance-konforme Kommentar-Policy fuer Service-Aktionen, die zuerst im Backend verbindlich eingefuehrt und danach in allen Frontend-Request-Schemas gespiegelt wird: Stornierungen und Auszahlungen muessen vom Personal kommentiert werden, waehrend Bestellungen, Ausgaben und Zahlungen als Standardfaelle weiterhin ohne Kommentar moeglich bleiben.

### Acceptance criteria

- [x] Stornierungen und Auszahlungen werden serverseitig ohne gueltigen Kommentar abgewiesen.
- [x] Bestellungen, Ausgaben und Zahlungen bleiben serverseitig ohne Kommentar zulaessig.
- [x] Die Frontend-Schemas spiegeln diese aufgabenspezifische Pflicht exakt fuer alle betroffenen Service-Aktionen.
- [x] Tests decken die Pflichtfaelle und Optionalfaelle jeweils mit leerem, zu kurzem, gueltigem und maximal langem Kommentar ab.

---

## Phase 5: Reduce Hot-Path And Maintenance Overhead

### Context

- `backend/api/table/application/command.go:289-318` - `BestellungAufnehmen` fuehrt 2 DB-Roundtrips pro Position aus (GetVariant + GetProduct).
- `backend/sqlc/queries/reporting.sql:7-105` - Reporting wiederholt `CASE WHEN type = '...' THEN (data->>'...')::int END`-Logik in 4 von 5 Queries.
- `backend/api/table/application/command.go:239-282` - `applyTischStatusChange` abstrahiert drei Status-Commands ueber einen Callback; wird entfernt, Commands werden direkt inline geschrieben.
- `backend/api/table/application/errors.go:49-61` - `fromRepositoryError`: behalten ‒ konsolidiert Logging mit Tisch-ID-Kontext und Error-Mapping.
- `backend/api/relay.go:14-44` - `druckerRepoRelayAdapter`: behalten ‒ vermeidet Import-Zyklus zwischen `relay/application` und `drucker_repo`.
- `frontend/src/lib/Backend.ts:31-65` - `TokenGetter`: behalten ‒ ermoeglicht Unit-Tests der `Backend`-Klasse ohne echte Auth-Abhaengigkeit.

### What to build

Eine gezielte Cleanup-Phase: Batch-Anreicherung im Bestellpfad (alle Produkt- und Varianten-IDs in je einer Abfrage), Deduplizierung von Reporting-SQL-Logik als PostgreSQL-Funktionen in `01_initial.up.sql`, und Entfernung von `applyTischStatusChange` zugunsten direkt lesbarer inline Status-Commands.

### Acceptance criteria

- [ ] `BestellungAufnehmen` laedt alle Produkte und Varianten in je einer Batch-Abfrage; neue `sqlc`-Queries werden dafuer angelegt.
- [ ] Wiederholte `CASE WHEN`-Extraktionslogik im Reporting ist als PostgreSQL-Hilfsfunktionen in `01_initial.up.sql` definiert; bestehende Queries verwenden diese Funktionen.
- [ ] `applyTischStatusChange` ist entfernt; `TischAktivieren`, `TischDeaktivieren` und `TischLoeschen` sind direkt inline implementiert.
- [ ] `fromRepositoryError`, `druckerRepoRelayAdapter` und `TokenGetter` bleiben erhalten; je ein erklaerende Begruendung ist im Code-Kommentar festgehalten.
- [ ] Die bestehende Verifikationskette bleibt nach den Refactors gruen.
