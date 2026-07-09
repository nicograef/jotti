# Datenbank-Migrationen

golang-migrate (v4), angewendet über das `jotti-migrate`-Image (`../migrate/Dockerfile`). Die Migrationen werden ins Image gebacken; beim Deploy läuft ausschließlich `migrate ... up`.

## Forward-only (seit v0.14.0, produktive Erstinstallation)

jotti fährt **forward-only: keine Down-Migrationen.** Neue Änderungen kommen als `NN_<name>.up.sql`, fortlaufend nummeriert, additiv und vorwärtskompatibel. Es gibt bewusst **keine** `.down.sql`.

**Warum kein down:**

- Das Kassenjournal ist fiskalisch append-only (Radierverbot, 10 Jahre Aufbewahrung). Ein `down`, das Spalten oder Tabellen mit Belegdaten droppt, zerstört aufbewahrungspflichtige Daten — auf Produktion ein Footgun.
- Das echte Rollback ist der Backup-Restore. `make prod-update` zieht vor jeder Migration ein Backup; schlägt die Migration oder der Health-Check fehl, wird das Backup eingespielt. `migrate down` wird auf Produktion nie ausgeführt.
- `down`-Migrationen, die Daten verwandeln, sind ohnehin nicht ehrlich umkehrbar (die verworfenen Daten kommen nicht zurück). Forward-only gibt vor, was zutrifft.

## Regeln für neue Migrationen

1. Dateiname `NN_<kurzname>.up.sql`, `NN` = nächste freie Nummer (aktuell zuletzt `02_druckauftrag_backoff`).
2. Additiv und vorwärtskompatibel. Bestehende Migrationen (insb. `01_initial.up.sql`) werden seit der produktiven Erstinstallation (v0.14.0) **nicht** mehr editiert.
3. In eine Transaktion klammern (`BEGIN; … COMMIT;`) — Postgres-DDL ist transaktional, so rollt ein Fehlschlag sauber zurück und hinterlässt keinen `dirty`-Zustand in `schema_migrations`.
4. Event-JSON-Contracts sind eingefroren (Guard: `backend/domain/kasse/event_json_contract_test.go`); Event-Änderungen additiv als neue Version (`:vN`), nie in-place.
5. Nach jeder Migration muss `make rebuild-projections` fehlerfrei durchlaufen (Projektionen werden aus Events neu gebaut).

## ENUM vs. TEXT+CHECK

jotti verwendet für Status- und Kategorie-Spalten TEXT+CHECK statt PostgreSQL-ENUMs. Begründung:

- **ENUMs sind DDL-Objekte.** Eine neue Ausprägung erfordert `ALTER TYPE ... ADD VALUE`, das in PostgreSQL nur außerhalb einer Transaktion oder mit bestimmten Einschränkungen läuft. Damit ist eine rein transaktionale Migration nicht möglich (verstößt gegen Regel 3).
- **Zwei-Migrations-Muster für ENUM-Erweiterungen** wäre nötig: (1) eine nicht-transaktionale Migration fügt den neuen Wert zum Typ hinzu, (2) eine zweite transaktionale Migration nutzt ihn. Das erhöht die Migrations-Komplexität und die Fehleranfälligkeit erheblich.
- **TEXT+CHECK ist einfacher erweiterbar:** Neuer Wert = `ALTER TABLE ... DROP CONSTRAINT ..., ADD CONSTRAINT ... CHECK (... IN (..., 'neu'))` — vollständig transaktional in einer Migration.
- **Ausnahmen** (`UserRole`, `EntityStatus`, `ProduktKategorie`, `Steuersatz`, `DruckstationKategorie`): Diese ENUMs existieren, weil sie bei Schema-Erstellung eingeführt wurden oder weil sqlc für ENUMs typsichere Go-Typen erzeugt (Compile-Zeit-Prüfung statt Laufzeit-String). Neue Status-/Kategorie-Spalten werden als TEXT+CHECK angelegt.

## Testen

- **Frischinstallation:** `migrate ... up` auf leerer DB (deckt der Integrationstest `scripts/test-integration.sh` ab).
- **Upgrade-Pfad:** `up` auf einer mit Vorversions-Daten befüllten DB, danach Boot + `make rebuild-projections`. Das ist der Pfad, der auf echten Instanzen läuft, und der wichtigste Migrations-Test. Automatisiert im CI-Job `upgrade-path` (`.github/workflows/ci.yml`): Migrationen und Seed-Daten werden mit den Release-Images der Vorversion eingespielt, danach laufen Migration, Boot und `rebuild-projections` mit dem aktuellen Checkout.

## Vorversions-Pinning (CI-Job `upgrade-path`)

- `PREVIOUS_VERSION` im Job pinnt die Vorversion auf das letzte veröffentlichte Release (aktuell `v0.14.0`); die Images kommen von `ghcr.io/nicograef` (`jotti-migrate`, `jotti-backend`).
- Nach jedem Release wird `PREVIOUS_VERSION` auf das neue Tag angehoben (Teil der Release-Mechanik).
- Seit der `02_druckauftrag_backoff`-Migration ist der Job das Pflicht-Gate für Schema-Änderungen: Er beweist, dass eine befüllte Bestandsinstanz das Upgrade übersteht (Release-Guide Gate 4 (b)).
