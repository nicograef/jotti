# Plan: TSE-Setup-Wiederaufnahme aus Sackgassen

> Folge-Plan zu [plan-tse-setup-wizard.md](plan-tse-setup-wizard.md). Adressiert die
> UX-Review-Findings F2 und F8: Wege aus den beiden Sackgassen, in denen ein
> Vereins-Admin ohne Admin-PIN feststeckt.

## Goal

Ein Admin, der eine vorhandene TSS im fiskaly-Konto vorfindet, kommt auch ohne die
verwahrte Admin-PIN aus der Einrichtung heraus:

- **F8 (Korrektheit, ohne Entscheidung):** Ist die vorgefundene TSS bereits
  `INITIALIZED` und ein passender Client bereits `REGISTERED`, ist die TSE faktisch
  einsatzbereit; jotti muss nur noch die lokale Konfiguration speichern. Dafür ist
  keine Admin-PIN nötig (es folgt keine privilegierte fiskaly-Operation). Heute
  verlangt die Übernahme trotzdem die PIN und läuft in eine Sackgasse. Dieser Fix
  macht die im Commit `acd86e3c` dokumentierte Wiederaufnahme („recoverable via
  takeover") für den häufigsten Orphan-Fall tatsächlich wahr.
- **F2 (Produktentscheidung):** In der TEST-Umgebung soll der Admin eine neue TSE
  anlegen können, auch wenn bereits eine (nicht nutzbare, weil PIN-lose) TSS im Konto
  liegt. Heute verweigert das Backend jede Neuanlage, sobald irgendeine nicht-
  `DISABLED` TSS existiert, was beim Ausprobieren der Normalfall ist und das Konto
  dauerhaft blockiert.

## Decision required (vor Phase 2)

F2 ist eine Produktentscheidung. Empfehlung **Option A**:

- **TEST:** Neuanlage trotz vorhandener übernehmbarer TSS erlauben, hinter einer
  expliziten Sekundäraktion „Stattdessen neue TSE anlegen". Kein Tipp-Zwang (TEST ist
  kostenlos und steuerlich ungültig).
- **LIVE:** harte Sperre beibehalten (`ErrTSEBereitsEingerichtet`). Eine
  versehentliche zweite LIVE-TSS verursacht laufende Kosten; eine echt verlorene
  LIVE-PIN ist ein Support-Fall, kein Self-Service.

Alternative (Option B): auch in LIVE erlauben, nur warnen; wegen der Kostenfolge
nicht empfohlen. **Diese Entscheidung ist offen und vor Beginn von Phase 2 mit dem
User zu bestätigen.**

## Architectural decisions

- **Keine Schema-Änderung.** Beide Findings betreffen nur den Setup-Orchestrator
  (`backend/api/settings/application/setup.go`), das Domain-Setup-Interface
  (`backend/domain/tse`) und den Wizard (`frontend/src/admin/tse/`).
- **„Einsatzbereit ohne Arbeit" ist der einzige PIN-freie Übernahmepfad.** F8 lockert
  die PIN-Pflicht ausschließlich, wenn `state == INITIALIZED` **und** ein passender
  Client mit `state == REGISTERED` existiert. Jeder Pfad, der eine privilegierte
  fiskaly-Operation auslöst (Initialisieren, Client registrieren/neu registrieren),
  braucht weiterhin die Admin-PIN.
- **`passenderClient` muss den Client-State berücksichtigen.** Heute matcht
  `passenderClient` (`backend/api/settings/application/query.go:209`) nur die
  `serial_number`. Für F8 muss der „kein Arbeit nötig"-Zweig zusätzlich
  `REGISTERED` prüfen; ein `DEREGISTERED` Client mit passender Seriennummer erfordert
  Neu-Registrierung und damit weiterhin die PIN.
- **F2 über expliziten Parameter, nicht über stille Lockerung.** `RichteTSEEin` bzw.
  der Endpoint `admin/tse-einrichten` erhält ein Flag `neuAnlegenTrotzVorhandener`
  (Default `false`). Die Sperre `hatAktiveTSS` wird nur übersprungen, wenn das Flag
  gesetzt ist **und** `umgebung == TEST`. So bleibt der LIVE-Schutz unangetastet und
  die Neuanlage ist immer eine bewusste Nutzeraktion.

## Inventory

- `backend/api/settings/application/setup.go:41-135` `RichteTSEEin` (Sperre
  `hatAktiveTSS` bei `:78`); `:153-266` `UebernimmTSE` (PIN-Pflicht bei `:240-242`);
  `:275-314` `vollendeLebenszyklus` (INITIALIZED-Zweig `:309-313`,
  `AuthentifiziereAdmin` vor `if !hatClient`); `:333-342` `hatAktiveTSS`
- `backend/api/settings/application/query.go:207-216` `passenderClient` (matcht nur
  serial_number, liefert aber `State`)
- `backend/api/settings/application/setup_test.go`: Unit-Test-Muster gegen den Fake
- `backend/domain/tse/fake_client.go`: Fake mit aufgezeichneten Aufrufen (für die
  Assertion „AuthentifiziereAdmin nicht aufgerufen")
- `backend/api/settings/http/command_handler.go`: zog-Schema des Einrichten-Requests
  (neues Flag ergänzen)
- `frontend/src/admin/tse/TSEEinrichtungWizard.tsx`: `brauchtPin` (PIN-Gating),
  `BefundSchritt` (Verzweigung Übernahme/Neuanlage), `UebernahmeSchritt`,
  `BestaetigungSchritt`, `PinFehltHinweis` (Disclosure aus dem F1-Fix)
- `frontend/src/lib/EinstellungenBackend.ts`: `TSEEinrichten`-Typ, `richteTSEEin`

## Open questions / Risks

- **F2-Entscheidung offen** (TEST erlauben vs. auch LIVE), siehe oben.
- **TEST-Konto füllt sich** (bereits im Wizard-Plan vermerkt): F2 erleichtert die
  Neuanlage und erhöht damit den Anfall nicht löschbarer Test-TSS. Akzeptabel, da
  TEST; im Leitfaden erwähnen.
- **F8-Sicherheit:** Sorgfältig sicherstellen, dass der PIN-freie Pfad wirklich keine
  fiskaly-Schreiboperation auslöst; sonst würde eine nicht autorisierte Mutation
  versucht und schlüge fehl. Durch den Fake-Test mit Aufruf-Assertion absichern.

---

## Phase 1: F8 — PIN-freie Übernahme einer einsatzbereiten TSS

**Findings:** F8, heilt F7

### Context

- `backend/api/settings/application/setup.go:240-242`: PIN-Pflicht für
  `UNINITIALIZED`/`INITIALIZED` greift heute auch, wenn nichts mehr zu tun ist
- `:309-313`: `AuthentifiziereAdmin(pin)` läuft vor dem `if !hatClient`-Guard und
  gatet ausschließlich die danach übersprungene `RegistriereClient`
- `backend/api/settings/application/query.go:209-216`: `passenderClient` liefert
  `State`, wird aber für die hatClient-Entscheidung in `setup.go:216` ohne State-Check
  verwendet

### What to build

- Backend: In `UebernimmTSE` einen Zweig „bereits einsatzbereit" erkennen
  (`state == INITIALIZED` **und** passender Client **`REGISTERED`**). In diesem Fall
  die PIN-Pflicht (`:240-242`) nicht anwenden und in `vollendeLebenszyklus` den
  `AuthentifiziereAdmin`-Aufruf überspringen (keine fiskaly-Mutation). Anschließend
  wie gehabt atomar die Konfiguration speichern. Den State-Check in die
  hatClient-Bestimmung ziehen, sodass nur ein `REGISTERED` Client als „fertig" gilt.
- Frontend: `brauchtPin` so erweitern, dass es für eine `INITIALIZED` TSS mit
  passendem `REGISTERED` Client `false` liefert. Der `UebernahmeSchritt` zeigt dann
  kein PIN-Feld und keinen `PinFehltHinweis`, sondern bietet direkt „TSE übernehmen"
  an, das nur die Konfiguration speichert.

### Acceptance criteria

- [ ] Übernahme einer `INITIALIZED` TSS mit passendem `REGISTERED` Client gelingt mit
      leerer PIN; der Fake belegt, dass `AuthentifiziereAdmin` **nicht** aufgerufen
      wird (Unit-Test)
- [ ] Übernahme einer `INITIALIZED` TSS mit passendem **`DEREGISTERED`** Client
      verlangt weiterhin die PIN (Unit-Test)
- [ ] Übernahme einer `INITIALIZED` TSS **ohne** passenden Client verlangt weiterhin
      die PIN (Registrierung nötig); bestehendes Verhalten bleibt
- [ ] Wizard zeigt für die einsatzbereite TSS kein PIN-Feld und schließt die Übernahme
      mit einem Klick ab; `make check` grün

---

## Phase 2: F2 — Neue TSE in TEST trotz vorhandener TSS

**Findings:** F2 · **setzt die Decision oben voraus**

### Context

- `backend/api/settings/application/setup.go:78` + `:333-342`: `hatAktiveTSS`-Sperre,
  umgebungsunabhängig
- `frontend/src/admin/tse/TSEEinrichtungWizard.tsx` `BefundSchritt`: rendert
  `BestaetigungSchritt` (Neuanlage) heute nur bei `nurDisabledOderLeer`

### What to build

- Backend: Parameter `neuAnlegenTrotzVorhandener bool` in `RichteTSEEin` und im
  zog-Schema von `admin/tse-einrichten`. `hatAktiveTSS`-Sperre nur überspringen,
  wenn Flag gesetzt **und** `umgebung == UmgebungTest`. LIVE bleibt hart gesperrt.
- Frontend: In `BefundSchritt` zusätzlich zum Übernahme-Angebot in **TEST** eine
  klar untergeordnete Aktion „Stattdessen neue TSE anlegen" anbieten (ruft
  `richteTSEEin` mit gesetztem Flag). Den `PinFehltHinweis` (F1) auf diesen Weg
  verweisen lassen, sodass der PIN-lose Admin in TEST einen echten Self-Service-
  Ausweg hat. In LIVE bleibt es bei Übernahme bzw. Support-Verweis.

### Acceptance criteria

- [ ] In TEST legt „Stattdessen neue TSE anlegen" trotz vorhandener `INITIALIZED` TSS
      eine zweite, frische TSE an (Unit-Test mit gesetztem Flag)
- [ ] In LIVE wird die Neuanlage trotz Flag verweigert (`ErrTSEBereitsEingerichtet`,
      Unit-Test)
- [ ] Ohne Flag bleibt das bisherige Verhalten unverändert (Sperre greift in beiden
      Umgebungen)
- [ ] Wizard: der PIN-lose Admin findet in TEST den Neuanlage-Ausweg ohne Umweg über
      einen Fehlversuch; Leitfaden um den TEST-Hinweis ergänzt
