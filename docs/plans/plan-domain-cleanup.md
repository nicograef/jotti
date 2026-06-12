# Plan: Domain-Layer Cleanup — alle Findings inkl. Strukturvorschläge

> Source PRD: n/a (from /cleanup report on `backend/domain/`, 2026-06-12)

## Goal

Alle Findings aus dem Cleanup-Review des Domain-Layers beheben: mechanische
Aufräumarbeiten (tote Hilfsfunktionen, falsche Doc-Kommentare, redundante
Validierungszweige, Stil-Drift), plus vier strukturelle Änderungen:

1. JWT-Minting aus der `User`-Entity in den Auth-Application-Layer verschieben
   (inkl. Entfernen des redundanten `"alg"`-Claims).
2. `Kassensitzung.Status` als typisierten String einführen.
3. Die "alle vier TSE-Felder zusammen"-Regel auf eine Stelle reduzieren
   (`tse.Credentials`).
4. Das persistierte Event-Payload-Schema vollständig in `domain/kasse`
   konsolidieren: Structs exportieren, `TSETxID`/`TSEAusfall` aufnehmen,
   Embed-Funktionen in die Domain ziehen, App-Layer-Kopien löschen,
   test-only `New…EventMitTSE`-Wrapper entfernen.

Kein Verhalten ändert sich — mit einer genehmigten Ausnahme: der `"alg"`-Claim
verschwindet aus dem JWT-Payload (wird von keinem Parser gelesen; laufende
Sessions bleiben gültig).

## Architectural decisions

Durable decisions that apply across all phases:

- **Event-Payload-Schema**: `domain/kasse` ist alleiniger Eigentümer der
  persistierten Event-Datenformen, inklusive der TSE-Felder `tseTxId`,
  `tseData`, `tseAusfall`. JSON-Keys bleiben byte-identisch (immutable
  events) — kein Feld wird umbenannt, keines hinzugefügt außer den bereits
  persistierten TSE-Feldern, die heute nur im App-Layer deklariert sind.
- **TSE-Embedding**: Pro signierbarem Event-Typ existiert eine exportierte
  Embed-Funktion in `domain/kasse` mit der Signatur
  `func(evt event.Event, txID string, tseData *kasse.TSEData) (event.Event, error)`
  (kompatibel zu `tseApp.EmbedTSE`). Die Signier-Orchestrierung
  (`Signierer.SignEvent`) bleibt in `api/tse/application`.
- **`TSEAusfall`** existiert nur dort, wo es heute persistiert wird:
  `zahlung-kassiert:v1` und `direktverkauf-getaetigt:v1`. Keine neuen Felder
  auf anderen Event-Typen.
- **JWT**: Die Entity prüft Status + Passwort (`User.VerifyPassword`);
  Token-Minting passiert ausschließlich in `api/auth/application` via
  `domain/jwt.GenerateJWTTokenForUser`. Das JWT-Payload enthält keinen
  `"alg"`-Claim mehr.
- **Status-Typen**: `kasse.KassensitzungStatus` als typed string mit den
  Konstanten `KassensitzungOffen`/`KassensitzungAbgeschlossen`, analog zu
  `product.Status`, `table.Status`, `user.Status`.

## Inventory

Domain (Findings):

- `backend/domain/tse/client.go:37-52` — `Credentials.Validate` mit
  unerreichbarem erstem Branch (`hasAny && !hasAll` ⊂ `!hasAll`).
- `backend/domain/product/variant.go:31-32,65,99` — "net price"-Doku auf
  Brutto-Preisen (Preise fließen als `Brutto` in `steuer.Aufteilen`).
- `backend/domain/user/user.go:20` — abgebrochener Doc-Kommentar
  ("…and products and .").
- `backend/domain/user/user.go:78` — `fmt.Errorf` für konstanten Fehler.
- `backend/domain/user/user.go:113` — `strings.ToLower` ist No-op nach
  `UsernameSchema` (`^[a-z0-9]+$`); `UpdateDetails` lowercased nicht.
- `backend/domain/user/user.go:204-218` — `GenerateJWTToken` auf der Entity.
- `backend/domain/user/user_test.go:10-25` — toter Helper `makeTestUser`.
- `backend/domain/event/event.go:13-16` — ID-Felddoku klebt am Struct.
- `backend/domain/jwt/jwt.go:14` — redundanter `"alg"`-Claim im Payload;
  `:17` Kommentar `// 12 hours validity` wiederholt den Code.
- `backend/domain/jwt/jwt_test.go:20` — `interface{}` statt `any`.
- `backend/domain/kasse/bestellung.go:34-40` — triviale Cast-Wrapper
  `toPositionEventData`/`fromPositionEventData`, nur in den Plural-Loops
  benutzt.
- `backend/domain/kasse/bestellung.go:63-64` — Kategorie-/Steuersatz-Literale
  duplizieren `product.Kategorie`- und `steuer.Steuersatz`-Konstanten.
- `backend/domain/kasse/historie.go:28` — `GetHistoryFromEvents` (englisch)
  liefert `[]HistorieEintrag` (deutsch); einziger Produktions-Caller:
  `backend/api/table/application/query.go`.
- `backend/domain/kasse/tisch_session_events.go:111-113` und
  `backend/domain/kasse/direktverkauf_events.go:68-70` — Konstruktoren
  mutieren das übergebene Positionen-Slice (PositionID-Zuweisung).
- `backend/domain/kasse/kassensitzung.go:6-19` — `Status string` untypisiert.
- `backend/domain/kasse/kommentar_test.go:51-62,112-123,141-152` — identisches
  Fat-Position-Fixture dreimal inline.
- `backend/domain/settings/betreiber.go:46`,
  `backend/domain/settings/tse_konfiguration.go:60` — `time.Now()` ohne
  `.UTC()` (einzige Domain-Packages ohne UTC-Normalisierung).
- `backend/domain/settings/tse_konfiguration.go:17-52` — dupliziert die
  Vier-Felder-Regel aus `tse.Credentials.Validate` (zweimal: `Validate` und
  `IstKonfiguriert`).
- `backend/domain/product/product.go:67`, `backend/domain/product/variant.go:10,57`
  — Doc-Kommentare nennen falsche Bezeichner (`NewProduct`/`NewVariant`/
  "product variant" für gemeinsam genutzten `Status`).

App-Layer (Payload-Struct-Kopien und TSE-Verdrahtung):

- `backend/api/table/application/command.go:90-135` — Kopien von
  `zahlungKassiertV1Data`, `zahlungPositionData`, `bestellungAufgenommenV1Data`,
  `stornierungErteiltV1Data`, `auszahlungGeleistetV1Data` mit zusätzlichen
  Feldern `TSETxID` (alle) und `TSEAusfall` (nur Zahlung).
- `backend/api/table/application/kassenbeleg_command.go:68-85,264,293,329` —
  Kopien von `direktverkaufGetaetigtV1Data` (mit `TSEAusfall`) und
  `direktverkaufStorniertV1Data`; liest `TSETxID`/`TSEAusfall` für den Beleg.
- `backend/api/table/application/tse_signing.go:53-72` — vier
  `EmbedTSEInData`-Closures.
- `backend/api/direktverkauf/application/tse_signing.go:13-30,54-60` — Kopien
  + zwei Closures (Getaetigt mit `TSEAusfall`).
- `backend/api/kasse/application/tse_signing.go:14-42,65-75` — Kopien von
  Geldtransit/Differenz/Tagesabschluss (je `TSETxID`) + drei Closures.
- `backend/api/tse/application/signing.go:47-73` — `EmbedTSE`-Typ und
  generisches `EmbedTSEInData[T]`; `:97-172` `Signierer.SignEvent`
  validiert `TSEData` bereits vor dem Embed; `:130-135` baut
  `tse.Credentials` per Hand aus `settings.TSEKonfiguration`.
- `backend/api/auth/application/command.go:22-56` — einziger Caller von
  `User.GenerateJWTToken`; mappt Domain-Fehler auf App-Fehler.
- `backend/repository/kassensitzungen_repo/repo.go:43,67`,
  `backend/repository/kassensitzungen_repo/types.go:24` — einzige Stellen,
  die `Kassensitzung.Status` schreiben.

Test-Nutzung der `MitTSE`-Konstruktoren (nur Tests, keine Produktion):

- `backend/domain/kasse/tse_roundtrip_test.go` (9×),
  `backend/domain/kasse/tisch_session_test.go` (3×),
  `backend/api/table/application/command_test.go` (3×),
  `backend/api/kasse/application/command_test.go` (Nutzung via Subjekt-Tests),
  `backend/api/direktverkauf/application/command_test.go` (1×).

Verifikation: `make test` (Backend-Unit-Tests), `make lint-backend`,
`make test-integration` für Phase 5.

## Resolved decisions

- **Payload-Schema**: Domain owns schema + embed (Option A). Structs werden
  exportiert, TSE-Felder wandern in die Domain, Embed-Funktionen pro Event-Typ
  in `domain/kasse`, App-Kopien werden gelöscht, `New…EventMitTSE`-Wrapper
  entfallen.
- **JWT**: `User.VerifyPassword(password) error` (Status + Hash, Domain-Fehler
  `ErrNotActive`/`ErrNoPassword`/`ErrInvalidPassword` bleiben);
  Minting in `api/auth/application`. `User.GenerateJWTToken` wird gelöscht.
- **`"alg"`-Claim**: wird entfernt (Token-Bytes ändern sich; akzeptiert).
- **Preis-Doku**: Preise sind brutto — die "net price"-Kommentare und
  -Fehlermeldungen in `variant.go` sind falsch und werden korrigiert
  (Code-Beleg: `Position.Einzelpreis` → `SteuermatrixPosition.Brutto`).
- **Credentials-Dedup**: `settings.TSEKonfiguration` erhält
  `Credentials() tse.Credentials`; der Sonderfall "komplett leer = gültig
  (TSE nicht konfiguriert)" bleibt in `settings`, weil die Semantik dort
  bewusst abweicht (`tse.Credentials.Validate` lehnt leer ab).

## Open questions / Risks

- **Phase 5 ist die breiteste Änderung** (10 Struct-Exporte, 9 Embed-Funktionen,
  4 App-Dateien, ~16 Test-Callsites). Die bestehenden Roundtrip-Tests
  (`tse_roundtrip_test.go`) sind das Sicherheitsnetz für Byte-Stabilität der
  Event-JSON — sie müssen vor und nach der Umstellung grün sein, ergänzt um
  Assertions für `tseTxId`/`tseAusfall`.
- `settings.TSEKonfiguration.Validate` liefert nach Phase 4 für den Fall
  "teilweise gesetzt" `tse.ErrUnvollstaendigeCredentials` statt des bisherigen
  `fmt.Errorf("all tse fields must be set together")` — Fehler-String ändert
  sich; betroffene Tests/Handler auf String-Matching prüfen.
- `NewVariante`/`UpdateDetails`-Fehlertexte ändern sich von "invalid net price"
  zu "invalid price" — interne Fehlerstrings, aber Tests gegenprüfen.

---

## Phase 1: Mechanische Bereinigung (alle S-Findings)

### Context

- `backend/domain/tse/client.go:37-52` — redundanter Validierungszweig
- `backend/domain/kasse/bestellung.go:34-64` — Cast-Wrapper + Literal-Duplikate
- `backend/domain/kasse/tisch_session_events.go:103-118`,
  `backend/domain/kasse/direktverkauf_events.go:61-75` — Input-Mutation
- `backend/domain/kasse/historie.go:28`,
  `backend/api/table/application/query.go` — Rename-Callsites
- `backend/domain/user/user.go:20,78,113`, `backend/domain/user/user_test.go:10-25`
- `backend/domain/event/event.go:13-16`
- `backend/domain/jwt/jwt.go:17`, `backend/domain/jwt/jwt_test.go:20`
- `backend/domain/product/product.go:67`, `backend/domain/product/variant.go:10,31-32,57,65,99`
- `backend/domain/settings/betreiber.go:46`, `backend/domain/settings/tse_konfiguration.go:60`
- `backend/domain/kasse/kommentar_test.go:36-46,51-62,112-123,141-152`

### What to build

Rein mechanische Korrekturen ohne API-Änderung (Ausnahme: ein Funktions-Rename
mit einem Produktions-Caller):

- `Credentials.Validate` auf die vier `has*`-Checks plus
  `if !hasAll { return ErrUnvollstaendigeCredentials }` reduzieren.
- Doc-Fixes: Brutto statt "net price" (Kommentar + zwei Fehlermeldungen),
  `ServiceRole`-Satz beenden, ID-Doku ans `Event.ID`-Feld, Funktionsnamen in
  den `NewProdukt`/`NewVariante`/`Status`-Kommentaren korrigieren,
  `// 12 hours validity` löschen.
- `toPositionEventData`/`fromPositionEventData` inlinen
  (`positionEventData(p)` / `Position(p)` direkt in den Plural-Loops).
- `positionSchema`-OneOf-Listen aus `product.Kategorie`- und
  `steuer.Steuersatz`-Konstanten ableiten (Import `steuer` in `bestellung.go`).
- `strings.ToLower`-No-op in `NewUser` entfernen.
- Positionen-Slices in `newBestellungAufgenommenEvent` und
  `newDirektverkaufGetaetigtEvent` vor der PositionID-Zuweisung mit
  `slices.Clone` kopieren (alle Produktions-Caller verifiziert: niemand liest
  das Slice danach).
- `time.Now().UTC()` in `betreiber.go` und `tse_konfiguration.go`.
- Toten Test-Helper `makeTestUser` löschen; Kommentar-Test-Fixture als
  `kommentarTestPositionenMitID` hochziehen (drei Inline-Kopien ersetzen).
- Stil-Drift: `errors.New` für `ErrNotActive`, `any` statt `interface{}` im
  JWT-Test.
- Rename `GetHistoryFromEvents` → `GetHistorieFromEvents`
  (Domain + `historie_test.go` + `api/table/application/query.go`).

### Acceptance criteria

- [ ] `make test` und `make lint-backend` grün
- [ ] Kein Verhalten geändert: keine JSON-Keys, keine Schemas, keine
      Validierungsregeln berührt (nur die zwei dokumentierten
      Fehlerstring-Korrekturen "invalid net price" → "invalid price")
- [ ] `grep -rn "GetHistoryFromEvents" backend/` liefert keine Treffer
- [ ] `grep -rn "makeTestUser" backend/` liefert keine Treffer

---

## Phase 2: JWT — Domain/Infrastruktur-Trennung

### Context

- `backend/domain/user/user.go:204-218` — `GenerateJWTToken` (Status-,
  Passwort-Check + Minting)
- `backend/domain/jwt/jwt.go:12-30` — Minting inkl. `"alg"`-Claim (Zeile 14)
- `backend/api/auth/application/command.go:22-56` — einziger Caller,
  Fehler-Mapping (`ErrNotActive`/`ErrNoPassword`/`ErrInvalidPassword`/
  `ErrTokenGeneration`)
- `backend/domain/user/user_test.go`, `backend/domain/jwt/jwt_test.go`,
  `backend/api/auth/application/command_test.go` — betroffene Tests

### What to build

Die Entity behält die Domain-Logik, das Minting wandert in den
Application-Layer:

- `User.VerifyPassword(password string) error`: prüft `Status != ActiveStatus`
  → `ErrNotActive`, leeren `PasswordHash` → `ErrNoPassword`, sonst
  Argon2id-Verifikation → `ErrInvalidPassword`. `User.GenerateJWTToken` wird
  gelöscht; `domain/user` importiert `domain/jwt` nicht mehr.
- `api/auth/application.Command.GenerateJWTToken` ruft `u.VerifyPassword(...)`
  und danach `jwt.GenerateJWTTokenForUser(u.ID, u.Name, string(u.Role),
  c.JWTSecret)`; das bestehende Fehler-Mapping bleibt unverändert
  (Minting-Fehler → `ErrTokenGeneration`).
- `"alg"`-Claim aus dem Token-Payload entfernen (Header trägt den Algorithmus
  bereits; kein Parser liest den Claim).

### Acceptance criteria

- [ ] `make test` und `make lint-backend` grün
- [ ] Login-Fehlerfälle (inaktiv, kein Passwort, falsches Passwort, User
      unbekannt) liefern dieselben App-Fehler wie zuvor
      (`command_test.go` unverändert grün oder nur Mock-seitig angepasst)
- [ ] `domain/user` hat keinen Import von `domain/jwt` mehr
- [ ] Ein mit dem neuen Code erzeugtes Token enthält keinen `alg`-Claim im
      Payload und wird von `ParseAndValidateJWTToken` akzeptiert

---

## Phase 3: Typisierter KassensitzungStatus

### Context

- `backend/domain/kasse/kassensitzung.go:6-19` — Entity + untypisierte
  Konstanten
- `backend/repository/kassensitzungen_repo/repo.go:43,67`,
  `backend/repository/kassensitzungen_repo/types.go:24` — Schreibstellen
- Lesende Vergleiche gegen `kasse.KassensitzungOffen` in
  `api/kasse`, `api/table`, `api/reporting`, `api/direktverkauf`,
  `repository/kassenjournal_repo` (kompilieren mit typed constants weiter)

### What to build

`type KassensitzungStatus string` in `domain/kasse` einführen, die beiden
Konstanten typisieren, `Kassensitzung.Status` auf den neuen Typ umstellen.
An den sqlc-Grenzen (`row.Status` ist `string`) explizit konvertieren
(`kasse.KassensitzungStatus(row.Status)` bzw. `string(ks.Status)` bei
Query-Parametern). Keine Validierung/kein Schema hinzufügen — reine
Typisierung analog `table.Status`.

### Acceptance criteria

- [ ] `make test` und `make lint-backend` grün
- [ ] `Kassensitzung.Status` ist `KassensitzungStatus`; beide Konstanten
      typisiert; keine String-Literale "offen"/"abgeschlossen" außerhalb der
      Konstanten-Definition und SQL (`grep -rn '"offen"' backend/ --include="*.go"`)

---

## Phase 4: TSE-Credentials-Regel deduplizieren

### Context

- `backend/domain/settings/tse_konfiguration.go:17-52` — `Validate` +
  `IstKonfiguriert` mit je eigener Vier-Felder-Prüfung
- `backend/domain/tse/client.go:30-52` — `Credentials` + kanonische Regel
- `backend/api/tse/application/signing.go:130-135` — manueller
  Feld-für-Feld-Bau von `tse.Credentials`
- `backend/domain/settings/tse_konfiguration_test.go` — bestehende Tests
  (Fehlerfall "teilweise gesetzt")

### What to build

`settings.TSEKonfiguration` delegiert an `tse.Credentials`:

- Neue Methode `Credentials() tse.Credentials` (Feld-Mapping an einer Stelle).
- `IstKonfiguriert()` delegiert: `t.Credentials().Validate() == nil`.
- `Validate()`: behält den Sonderfall "alle vier leer = gültig" (lokale
  `leer()`-Prüfung), delegiert den Rest an `t.Credentials().Validate()`;
  Längen- und `UpdatedAt`-Checks bleiben unverändert in `settings`.
- `signing.go:130-135` nutzt `conf.Credentials()` statt des manuellen Baus.

> **Hinweis:** Der Fehlerwert für "teilweise gesetzt" wird dadurch
> `tse.ErrUnvollstaendigeCredentials` — Tests, die auf den alten String
> matchen, anpassen.

### Acceptance criteria

- [ ] `make test` und `make lint-backend` grün
- [ ] Die Vier-Felder-Regel existiert nur noch in
      `tse.Credentials.Validate` (kein `hasApiKey && hasApiSecret && …`
      mehr in `settings`)
- [ ] Verhalten identisch: komplett leer → gültig + `IstKonfiguriert() ==
      false`; teilweise gesetzt → Fehler; vollständig → gültig + `true`

---

## Phase 5: Event-Payload-Schema in die Domain konsolidieren

### Context

- `backend/domain/kasse/tisch_session_events.go:23-91` — fünf unexportierte
  Payload-Structs + Schemas
- `backend/domain/kasse/direktverkauf_events.go:16-49` — zwei Structs
- `backend/domain/kasse/kassensitzung_events.go:22-99` — fünf Structs
- `backend/domain/kasse/bestellung.go:23-56` — `positionEventData` + Loops
- `backend/api/tse/application/signing.go:47-73` — `EmbedTSE`-Typ +
  generisches `EmbedTSEInData[T]` (zieht als unexportierter Helper nach
  `domain/kasse` um; der `EmbedTSE`-Typ bleibt in `api/tse/application`)
- App-Kopien (werden gelöscht):
  `backend/api/table/application/command.go:90-135`,
  `backend/api/table/application/kassenbeleg_command.go:68-85`,
  `backend/api/direktverkauf/application/tse_signing.go:13-30`,
  `backend/api/kasse/application/tse_signing.go:14-42`
- Embed-Closures (werden durch Domain-Funktionen ersetzt):
  `backend/api/table/application/tse_signing.go:53-72`,
  `backend/api/direktverkauf/application/tse_signing.go:54-60`,
  `backend/api/kasse/application/tse_signing.go:65-75`
- Test-Callsites der `MitTSE`-Wrapper: `tse_roundtrip_test.go`,
  `tisch_session_test.go`, `api/table/application/command_test.go:1620,2139,2335`,
  `api/kasse/application/command_test.go`,
  `api/direktverkauf/application/command_test.go`

### What to build

`domain/kasse` wird alleiniger Eigentümer der persistierten Event-Payloads:

- **Structs exportieren** (Rename, JSON-Tags unverändert):
  `PositionEventData`, `BestellungAufgenommenV1Data`, `ZahlungKassiertV1Data`,
  `StornierungErteiltV1Data`, `AusgabeBestaetigtV1Data`,
  `AuszahlungGeleistetV1Data`, `DirektverkaufGetaetigtV1Data`,
  `DirektverkaufStorniertV1Data`, `GeldtransitGebuchtV1Data`,
  `DifferenzSollIstGebuchtV1Data`, `KassensitzungEroeffnetV1Data`,
  `KassensturzDurchgefuehrtV1Data`, `TagesabschlussErstelltV1Data`.
- **TSE-Felder aufnehmen**, exakt wie heute im App-Layer persistiert:
  `TSETxID string \`json:"tseTxId,omitempty"\`` auf allen neun signierbaren
  Typen; `TSEAusfall bool \`json:"tseAusfall,omitempty"\`` nur auf
  `ZahlungKassiertV1Data` und `DirektverkaufGetaetigtV1Data`.
- **Embed-Funktionen in der Domain** (eine pro signierbarem Typ, Signatur
  kompatibel zu `tseApp.EmbedTSE`), z. B.
  `kasse.EmbedTSEInZahlungKassiert(evt, txID, tseData)`. Gemeinsame Mechanik
  (Typ-Check, JSON-Roundtrip) als unexportierter generischer Helper in
  `domain/kasse`. Nicht-nil `tseData` wird beim Embed validiert (ersetzt die
  Validierung der bisherigen `MitTSE`-Konstruktoren).
- **App-Layer entschlacken**: die vier Kopie-Blöcke löschen;
  `tse_signing.go`-Dateien reichen nur noch die Domain-Embed-Funktionen an
  `SignEvent` durch; `kassenbeleg_command.go` parst mit den exportierten
  Domain-Structs.
- **Konstruktoren vereinfachen**: die neun `New…EventMitTSE`-Wrapper und die
  `tseData`-Parameter der privaten Implementierungen entfernen; private
  Implementierungen in die öffentlichen `New…Event` falten. Domain- und
  App-Tests stellen auf `New…Event` + Embed-Funktion um.
- **Roundtrip-Tests erweitern**: `tse_roundtrip_test.go` prüft zusätzlich,
  dass `tseTxId`/`tseAusfall` den JSON-Roundtrip verlustfrei überleben.

### Acceptance criteria

- [ ] `make test`, `make lint-backend` und `make test-integration` grün
- [ ] `grep -rn "V1Data struct" backend/api/` liefert keine Treffer
      (keine Payload-Struct-Kopien mehr im App-Layer)
- [ ] `grep -rn "EventMitTSE" backend/` liefert keine Treffer
- [ ] Persistierte Event-JSON byte-identisch zu vorher: gleiche Keys, gleiche
      omitempty-Semantik (Roundtrip-Tests + ein manueller Vergleich eines
      signierten und eines Ausfall-Events)
- [ ] `TSEAusfall` existiert nur auf `ZahlungKassiertV1Data` und
      `DirektverkaufGetaetigtV1Data`
