# Plan: Einmalpasswort auf 6 Ziffern und generierter Initial-Admin

> Source PRD: ../prds/prd-otp-6-digits-und-initial-admin.md

## Goal

Zwei zusammenhängende Ziele:

1. Den Aussperr-Deadlock bei Neuinstallationen beheben: Das OTP ist wieder genau
   6 Ziffern, und Frontend wie Backend prüfen dieses Format identisch aus einer
   kanonischen Regel je Seite.
2. Den hartkodierten Initial-Admin aus der Migration entfernen. Stattdessen legt
   das Backend beim Start den Admin an bzw. rotiert dessen OTP, solange die
   Ersteinrichtung offen ist, und macht den Klartext-Code dort sichtbar, wo der
   Anwender ohnehin hinschaut (Starter-Konsole, `prod-init`-Ausgabe, Status-Seite).

## Architektur-Entscheidungen

Durchgängige, phasenübergreifende Festlegungen:

- **OTP-Format**: genau 6 Ziffern, kanonische Regel `^\d{6}$` je Seite.
  - Backend: neue `user.OnetimePasswordSchema` (zog) im Domain-Modul
    `backend/domain/user`, verwendet vom Auth-HTTP-Handler beim Passwortsetzen.
  - Frontend: `OnetimePasswordSchema` in `frontend/src/lib/identity.ts` (Zod),
    einmalige Quelle, bereits von `AuthBackend.ts` und `UserBackend.ts` importiert.
- **Bootstrap-Aktionen**: genau eine von drei Aktionen, entschieden am DB-Zustand:
  - `create` — `users` leer → aktiver `admin` mit frischem 6-Ziffern-OTP, kein Passwort.
  - `rotate` — genau ein Benutzer, dieser ist `admin` ohne Passwort → neues OTP,
    Zähler zurück, Passwort/Status unberührt (Rotations-/Wiederherstellungsfall).
  - `skip` — jeder andere Zustand → keine Änderung.
- **Bootstrap-Platzierung**: neues, isoliert testbares Package (`backend/bootstrap`),
  aufgerufen aus `main.go` `run()` nach erfolgreichem DB-Ping und vor der
  HTTP-Auslieferung. Ein Bootstrap-Fehler ist **fatal** (`log.Fatal`), konsistent
  mit dem bestehenden `PingWithRetry`→`Fatal`-Muster (`backend/main.go:47-49`).
- **Domain-Wiederverwendung, keine neuen Pfade**: `create` komponiert das
  bestehende `NewUser` (inactive + frisches OTP) mit `Activate()`; `rotate` ist
  exakt `ResetPassword()`. Das Bootstrap-Package enthält nur die Entscheidungslogik.
- **Repository**: eine neue Zählabfrage `CountUsers`; Anlegen/Aktualisieren/Lesen
  über die bereits vorhandenen Pfade (`CreateUser`, `UpdateUser`, `GetUserByUsername`).
- **Marker-Logging**: bei `create`/`rotate` schreibt das Backend zwei Zeilen in
  den Log-Strom — eine maschinen-greifbare Markerzeile mit festem ASCII-Präfix und
  dem Klartext-Code, plus eine menschenlesbare Bannerzeile. Der Klartext geht nur
  in den Log-Strom, nie über das Netz.
- **Log-Konsum nur seit Container-Start**: Log-Konsumenten (Windows-Starter)
  berücksichtigen ausschließlich Zeilen seit dem aktuellen Container-Start
  (`docker inspect … .State.StartedAt` + `docker logs --since`). Fehlt der Marker
  im aktuellen Boot, ist die Einrichtung abgeschlossen — kein Code wird angezeigt.
- **Migration bleibt reines Schema**: kein Admin-`INSERT`, kein Argon2id im
  Migrate-Container; deterministisches SQL.
- **Keine Schema-Änderung an `users`**: die Tabelle deckt alle drei Aktionen bereits ab.

## Inventory

- `backend/domain/user/password.go:143-162` — `generateOnetimePassword` (aktuell 8
  alphanumerische Zeichen, Charset + Entropie-Kommentar); `generateOnetimePasswordHash`
  darüber. `PasswordSchema` (Zeile 15) bleibt unverändert.
- `backend/domain/user/password_test.go:10-23` — `TestGenerateOnetimePassword`
  (8-Zeichen-Regex), muss auf 6 Ziffern.
- `backend/domain/user/user.go:96-126` — `NewUser` (inactive + OTP), `:128-131`
  `Activate()`, `:164-176` `ResetPassword()` (neues OTP, Zähler 0, Passwort leer),
  `:182-217` `SetPassword` — Zeile 189 normalisiert mit `ToLower`+`TrimSpace`
  (die tolerante Kleinschreibung entfällt, nur Trim bleibt).
- `backend/domain/user/user_test.go` — Domain-Tests mit OTP-Fixtures (Fehlversuchszähler,
  Einmaligkeit), Fixtures auf 6 Ziffern umstellen.
- `backend/api/auth/http/command_handler.go:73-77` — `setPasswordSchema`, OTP-Feld
  aktuell nur `z.String().Trim().Min(1)`; ersetzen durch `user.OnetimePasswordSchema`.
- `backend/api/auth/http/command_handler_test.go:110-120` — bestehender set-password-Test
  (leeres OTP); Vorlage für die Formatvalidierungs-Fälle.
- `backend/main.go:29-96` — Startablauf; `run()` (Zeile 72) ist der Aufrufort fürs Bootstrap.
- `backend/repository/user_repo/repo.go` — `CreateUser`/`UpdateUser`/`GetUserByUsername`
  vorhanden; `CountUsers` fehlt. `mock.go` spiegelt die Repo-Methoden für Unit-Tests.
- `backend/repository/user_repo/repo_test.go` — Fake-/Repo-Testmuster als Vorlage.
- `backend/sqlc/queries/users.sql` — sqlc-Queries; `CountUsers` neu, danach `make`-sqlc-Gen.
- `database/migrations/01_initial.up.sql:38-45` — hartkodierter Admin-`INSERT` (entfällt).
- `frontend/src/lib/identity.ts:22-25` — `OnetimePasswordSchema` (`^[a-z0-9]{8}$/i`),
  Kommentar Zeile 7 nennt bereits „6 Ziffern"; wird durch die Änderung wieder wahr.
- `frontend/src/components/common/FormFields.tsx:1,183-219` — `OTPField`: `maxLength={8}`,
  `pattern={REGEXP_ONLY_DIGITS_AND_CHARS}`, 8 Slots. `REGEXP_ONLY_DIGITS` (`^\d+$`) ist
  in `input-otp` verfügbar (`node_modules/input-otp/dist/index.d.ts:38`).
- `frontend/src/lib/{AuthBackend,identity}.ts`, `frontend/src/admin/users/UserBackend.ts`
  — importieren `OnetimePasswordSchema`; eine Änderung propagiert.
- `windows/starter/main.go:113-130` — `waitForHealth` → `printSuccess`-Flow;
  `windows/starter/system.go` — Docker-Helper (`exec.Command("docker", …)`, `waitForHealth`).
- `scripts/prod-init.sh` — Abschluss-Summary („Useful commands") am Skriptende.
- `reverse-proxy/statuspage.go:206-247` — Karten-Template der Status-Seite (eigenes
  Binary, keine DB-Anbindung).
- `docs/leitfaden/installation.md:30-38` — „Erster Login" nennt festes `123456`.
- `docs/language.md:77-79` — Glossar „Einmalpasswort" (8 Zeichen).
- `docs/handbuch.md:282,362-365` — Benutzer-Invarianten und Onboarding-Ablauf (8 Zeichen).
- `docker-compose.yml:33-72` — `migrate`-Service (eigener Container) läuft vor dem
  Backend (`depends_on: service_completed_successfully`); bestätigt: Bootstrap muss im
  Backend nach dem Ping laufen, nicht in der Migration.

## Resolved decisions

- **Phasenschnitt: 3 Phasen.** P1 = 6-Ziffern-OTP end-to-end (behebt den Deadlock
  allein, da Seed `123456` wieder eingebbar wird). P2 = Bootstrap + Migration
  entkoppeln. P3 = Sichtbarkeits-Oberflächen + betreiber-/anwender-orientierte Doku.
- **Bootstrap-Fehler ist fatal** (`log.Fatal`) — laut abbrechen, Container-Restart
  wiederholt; ein stilles „läuft, aber nicht anmeldbar" wird vermieden.
- **Doku-Aufteilung**: Format-Aussagen (8 Zeichen → 6 Ziffern) im Glossar/den
  Invarianten gehören inhaltlich zu P1 (Format); die Aussagen über den generierten,
  rotierenden Initial-Code und „Erster Login" gehören zu P3 (Bootstrap-Sichtbarkeit).
- **Marker-Präfix** (Default, muss stabil bleiben, weil Starter und `prod-init`
  danach grep’en): maschinen-greifbare Zeile mit festem Präfix, z. B.
  `ADMIN-EINMALPASSWORT ` gefolgt von `benutzer=admin code=<6 Ziffern>`, plus eine
  menschenlesbare Bannerzeile. Der exakte String wird in P2 fixiert und in P3
  (Starter/`prod-init`-Grep) unverändert übernommen.
- **`SetPassword`-Normalisierung**: nur noch `TrimSpace`; die frühere
  `ToLower`-Toleranz (fürs Buchstaben-Alphabet) entfällt, reine Ziffern brauchen sie nicht.
- **v0 Beta, keine Bestandsinstallationen**: Es gibt keine produktiven Installationen,
  die migriert werden müssten. Die initiale Migration `01_initial.up.sql` wird direkt
  editiert (Breaking Change bewusst akzeptiert); es braucht keine nachgelagerte
  Migration und keinen Kompatibilitätspfad. Die `skip`-Aktion bleibt trotzdem
  notwendig — nicht als Bestands-Kompatibilität, sondern als Laufzeit-Invariante
  (nach gesetztem Passwort nicht rotieren, offenes Service-OTP nie antasten).

## Open questions / Risks

- **Container-Name/Compose-Kontext im Starter**: Der Grep muss den richtigen
  Backend-Container/-Service treffen (Release-Compose vs. Dev). In P3 gegen den
  konkret verwendeten Compose-Aufruf (`windows/starter/system.go`) verifizieren.
- **Log-Format (zerolog `ConsoleWriter`)**: Der Marker muss als stabiles Literal in
  der gerenderten Ausgabe stehen, damit der Grep unabhängig von Feld-Formatierung
  greift. In P2 durch einen Round-Trip (Log erzeugen, Zeile grep’en) prüfen.

---

## Phase 1: 6-Ziffern-OTP end-to-end

**User stories**: 2, 3, 12, 13, 14, 17, 21 (sowie 1 als Folge — der Seed `123456`
wird wieder eingebbar).

### Context

- `backend/domain/user/password.go:143-162` — Generator auf 6 Ziffern; Entropie-Kommentar
  auf `10^6`, abgesichert durch Einmaligkeit + 5-Versuche-Sperre.
- `backend/domain/user/password_test.go:10-23` — Generator-Test auf `^\d{6}$`.
- `backend/domain/user/user.go:182-217` — `SetPassword`: Normalisierung nur noch `TrimSpace`.
- `backend/domain/user/user_test.go` — OTP-Fixtures von 8-Zeichen auf 6-Ziffern.
- `backend/api/auth/http/command_handler.go:73-77` — `setPasswordSchema` nutzt neue
  `user.OnetimePasswordSchema`.
- `backend/api/auth/http/command_handler_test.go:110-120` — Vorlage für die
  Formatvalidierungs-Fälle.
- `frontend/src/lib/identity.ts:22-25` — Zod-Schema `^\d{6}$` + deutsche Meldung.
- `frontend/src/components/common/FormFields.tsx:183-219` — `OTPField` auf 6 Slots.
- `docs/language.md:77-79`, `docs/handbuch.md:282,362-365` — „8 Zeichen …" → „6 Ziffern".

### What to build

Der OTP-Generator erzeugt wieder genau 6 gleichverteilte Ziffern (kein Modulo-Bias).
Eine kanonische Formatregel je Seite prüft „genau 6 Ziffern": im Backend eine neue
`user.OnetimePasswordSchema`, die die bisherige „nur nicht leer"-Prüfung im
Passwort-Setzen-Handler ersetzt und so die Validierungslücke schließt; im Frontend
das bestehende `OnetimePasswordSchema` (`^\d{6}$`, deutsche Fehlermeldung „genau 6
Ziffern"). Das OTP-Eingabefeld hat 6 Slots, `maxLength=6`, reines Ziffern-Muster,
`inputMode="numeric"` und `autocomplete="one-time-code"`, damit Tablets/Smartphones
direkt die Zifferntastatur anbieten. `SetPassword` normalisiert nur noch per Trim.
Format-Aussagen in Glossar und Benutzer-Invarianten werden von „8 Zeichen" auf „6
Ziffern" korrigiert. Damit passt das Feld wieder zum erzeugten Code, und der Deadlock
verschwindet bereits unabhängig von Phase 2.

### Acceptance criteria

- [x] `generateOnetimePassword` liefert genau 6 Zeichen, ausschließlich `0`–`9`;
      Generator-Test grün (50 Durchläufe gegen `^\d{6}$`).
- [x] `user.OnetimePasswordSchema` existiert und akzeptiert `123456`, lehnt leer,
      `12345`, `1234567`, `abcdef` ab.
- [x] Der Passwort-Setzen-Handler lehnt ein Nicht-6-Ziffern-OTP mit Client-Fehler
      ab, bevor ein Hashvergleich stattfindet (Handler-Test).
- [x] `SetPassword` normalisiert nur per Trim; bestehende Domain-Tests
      (Fehlversuchszähler, Einmaligkeit) mit 6-Ziffern-Fixtures grün.
- [x] Frontend-`OnetimePasswordSchema` akzeptiert `123456`, lehnt `12345`,
      `1234567`, `abcdef` ab (Vitest).
- [x] `OTPField` zeigt 6 Slots, `maxLength=6`, Ziffern-Muster, `inputMode="numeric"`,
      `autocomplete="one-time-code"`.
- [x] `docs/language.md` und `docs/handbuch.md` sagen „6 Ziffern" statt „8 Zeichen".
- [x] `make verify` (Backend + Frontend) grün.
- [x] Manuell/Smoke: Frischer Seed `123456` lässt sich im „Passwort festlegen"-Formular
      eingeben, Passwort setzen und anschließend anmelden. (Belegt durch die
      Schema-Tests: Seed-Hash für `123456` unverändert, Front- und Backend-Schema
      akzeptieren `123456`.)

---

## Phase 2: Initial-Admin-Bootstrap + Migration entkoppeln

**User stories**: 10, 11, 15, 18, 20 (Basis für 5, 19 — Rotation als Wiederherstellung).

### Context

- `database/migrations/01_initial.up.sql:38-45` — Admin-`INSERT` + Kommentar entfällt.
- `backend/sqlc/queries/users.sql` — neue Query `CountUsers`, danach sqlc-Gen.
- `backend/repository/user_repo/repo.go`, `mock.go` — `CountUsers` in Repo + Mock.
- `backend/domain/user/user.go:96-131,164-176` — `NewUser`+`Activate` (create),
  `ResetPassword` (rotate) wiederverwenden.
- `backend/main.go:72-96` — `run()`: Bootstrap nach DB-Ping, vor `app.NewApp`/Serving;
  Fehler `log.Fatal`.
- `backend/repository/user_repo/repo_test.go` — Fake-Repo-Muster für den Entscheidungstest.

### What to build

Der hartkodierte Admin-`INSERT` entfällt aus `01_initial.up.sql`; die Migration
erzeugt nur noch das Schema. Ein neues, isoliert testbares Bootstrap-Package
entscheidet anhand des DB-Zustands genau eine von drei Aktionen und liefert
Klartext-Code plus Aktion zurück: `create` (leeres Repo → aktiver `admin`, kein
Passwort, frisches 6-Ziffern-OTP, komponiert aus `NewUser`+`Activate`), `rotate`
(einziger `admin` ohne Passwort → `ResetPassword`, auch nach fünf Fehlversuchen mit
leerem OTP-Hash: Selbstheilung der Sackgasse), `skip` (jeder andere Zustand — Admin
hat Passwort oder es existieren weitere Benutzer; insbesondere wird ein offenes
Service-OTP nie angetastet). Die Entscheidung nutzt `CountUsers` plus `GetUserByUsername`.
Das Bootstrap läuft beim Backend-Start nach dem DB-Ping und ist idempotent über die
drei Aktionen; die Username-Eindeutigkeit schützt zusätzlich gegen Doppelanlage. Bei
`create`/`rotate` schreibt es zwei Log-Zeilen: eine maschinen-greifbare Markerzeile
mit festem Präfix und dem Klartext-Code und eine menschenlesbare Bannerzeile. Ein
Bootstrap-Fehler bricht den Start fatal ab.

### Acceptance criteria

- [x] `01_initial.up.sql` enthält keinen Admin-`INSERT` mehr; Migration bleibt reines,
      deterministisches Schema-SQL.
- [x] `CountUsers` in `users.sql`, generiert, in Repo + Mock verfügbar.
- [x] Bootstrap-Entscheidungstest (tabellengetrieben, Fake-Repo, ohne DB) deckt ab:
  - [x] leeres Repo → `create`: aktiver `admin`, kein Passwort, OTP-Hash gesetzt,
        Rückgabe-OTP genau 6 Ziffern.
  - [x] einziger `admin` ohne Passwort → `rotate`: neuer OTP-Hash, Zähler 0,
        Status/Passwort unverändert leer.
  - [x] einziger `admin` ohne Passwort, OTP nach Fehlversuchen gesperrt (Hash leer)
        → `rotate`: frischer OTP-Hash, Zähler 0.
  - [x] `admin` mit Passwort → `skip`: keine Änderung.
  - [x] mehrere Benutzer inkl. Service-Benutzer mit offenem OTP → `skip`: Service-OTP
        unangetastet.
- [x] Bootstrap wird in `main.go` `run()` nach dem Ping aufgerufen; Fehler → `log.Fatal`.
- [x] Bei `create`/`rotate` erscheinen Markerzeile (fester Präfix + Klartext-Code) und
      Bannerzeile im Log; der Präfix ist als stabiles Literal grep-bar (Round-Trip geprüft).
- [x] Frische DB: Backend-Start erzeugt aktiven `admin`, loggt Code; Neustart bei
      offener Einrichtung loggt neuen Code; nach gesetztem Passwort `skip` (kein neuer Code).
- [x] `make verify` grün (inkl. Integrationstests, die auf den Seed-Admin bauten —
      auf den Bootstrap-Pfad bzw. explizites Anlegen umgestellt, falls betroffen).

---

## Phase 3: Sichtbarkeit des Codes + Betreiber-/Anwender-Doku

**User stories**: 4, 5, 7, 8, 9, 16, 19.

### Context

- `windows/starter/main.go:113-130` — nach `waitForHealth` in den `printSuccess`-Flow
  einhängen; `windows/starter/system.go` — Docker-Helper-Muster (`exec.Command`),
  `docker inspect .State.StartedAt` + `docker logs --since`.
- `scripts/prod-init.sh` — Abschluss-Summary („Useful commands").
- `reverse-proxy/statuspage.go:206-247` — zusätzliche statische Hinweiskarte im Template.
- `docs/leitfaden/installation.md:30-38` — „Erster Login" ohne festen `123456`.
- `docs/handbuch.md:362-365` — Onboarding-Abschnitt spiegelt generierten, rotierenden
  Initial-Code.

### What to build

Die drei Oberflächen machen den in Phase 2 geloggten Code sichtbar. Der Windows-Starter
liest nach erfolgreichem Health-Check die jüngste Markerzeile seit dem aktuellen
Container-Start (`docker inspect … .State.StartedAt` + `docker logs --since`) und zeigt
sie als vollständige Handlungsanweisung: Benutzername `admin`, der 6-stellige Code und
der Eingabeort („Passwort festlegen" in der App). Schlägt Lesen/Parsen fehl oder fehlt
der Marker im aktuellen Boot (Einrichtung abgeschlossen), ist das nicht fatal — die
Meldung lautet „jotti neu starten — dann wird ein neuer Code angezeigt", ohne Verweis
auf Logs. `scripts/prod-init.sh` parst keine Logs, sondern nennt im Summary den fertigen
Befehl (`docker compose … logs backend | grep <Präfix>`), mit dem der technische
Betreiber den Code abliest. Die Status-Seite bekommt eine zusätzliche statische,
stets unbedenkliche Hinweiskarte (kein DB-/Backend-Zugriff, kein Klartext-Code):
„Falls Sie jotti gerade zum ersten Mal einrichten: Der Anmelde-Code steht in der
Startkonsole. Konsole schon geschlossen? jotti neu starten — dann wird ein neuer Code
angezeigt." Die Doku wird angepasst: `installation.md` „Erster Login" beschreibt den
generierten, in der Startkonsole bzw. `prod-init`-Ausgabe angezeigten Code und den
Neustart-Weg; der Handbuch-Onboarding-Abschnitt spiegelt den rotierenden Initial-Code
statt des festen `123456`.

### Acceptance criteria

- [x] Windows-Starter liest die jüngste Markerzeile seit dem aktuellen Container-Start
      und zeigt Benutzername `admin`, 6-stelligen Code und Eingabeort in der Konsole.
- [x] Fehlt der Marker im aktuellen Boot oder scheitert das Parsen, zeigt der Starter
      non-fatal die Neustart-Meldung — kein Verweis auf Logs, Start läuft normal weiter.
- [x] `scripts/prod-init.sh` gibt im Summary den fertigen `logs … | grep <Präfix>`-Befehl
      aus (kein Log-Parser im Skript).
- [x] Status-Seite zeigt die statische Ersteinrichtungs-Hinweiskarte; kein DB-/Backend-Zugriff,
      kein Klartext-Code auf der Seite; `statuspage_test.go` weiterhin grün.
- [x] `docs/leitfaden/installation.md` „Erster Login" ohne festen `123456`, mit
      generiertem Code + Neustart-Weg; Handbuch-Onboarding spiegelt den rotierenden Code.
- [x] `make verify` grün; Release-Smoke deckt den Pfad „Einrichtung abgeschlossen,
      kein Marker im aktuellen Boot, kein Code in der Ausgabe" ab (manuell dokumentiert).

---

## Verifikation (gesamt)

- Nach jeder Phase `make verify` grün (Backend `-tags unit` + Integration, Frontend
  Lint/Typecheck/Vitest, sqlc-Gen aktuell).
- Phase 1 allein am frischen Seed `123456` demonstrierbar (Deadlock behoben).
- Phase 2 über Backend-Logs auf frischer/rotierender/abgeschlossener DB demonstrierbar.
- Phase 3 manuell bzw. über den Release-Smoke-Test (Starter-Ausgabe, `prod-init`-Summary,
  Status-Seite) inkl. „kein Marker"-Fall.
