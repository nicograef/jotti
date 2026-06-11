# Plan: Positions-Steuersatz im HTTP-DTO ergänzen (Service-State-Bug beheben)

> Source PRD: n/a (Bug-Analyse aus Live-Betrieb im Vereinsheim)

## Goal

Die Service-Tisch-Seite zeigt `0 €` trotz offener Bestellungen, lädt fehlerhaft,
und die Service-Startseite zeigt keine markierten Tische („Favoriten weg").
Ursache ist ein **stiller Schema-Mismatch**: Das Frontend-`PositionSchema`
verlangt ein Pflichtfeld `steuersatz`, das die Backend-HTTP-DTO `position`
nicht mitsendet. Jede positions-tragende Response scheitert an der
Zod-Validierung, der Query landet im Error-State, und die Hooks fallen auf
ihre Default-Werte zurück (0 €, leere Tischliste).

Ziel: Positionen tragen **überall** ihren Steuersatz end-to-end, der akute
Bug ist behoben, meine vorherigen Symptom-Pflaster sind zurückgenommen, und
künftige Mismatches dieser Art werden früh sichtbar.

## Root-Cause-Kette (verifiziert)

1. Domain `kasse.Position` **hat** `Steuersatz string`
   (`backend/domain/kasse/bestellung.go:10-18`), und das Event-Data-Struct
   persistiert ihn (`backend/domain/kasse/bestellung.go:21-31`).
2. Die HTTP-DTO `position` und ihr Mapper `toPosition` **lassen das Feld weg**
   (`backend/api/table/http/query_handler.go:111-130`).
3. Geteilter Mapper `toPositionen` versorgt `get-tisch-state`,
   `get-tisch-historie` und `get-meine-tische-state` — also fehlt `steuersatz`
   in **allen** dreien.
4. Frontend `PositionSchema` verlangt `steuersatz` als Pflicht-Enum
   (`frontend/src/service/table/Bestellung.ts:3-13`), genutzt von
   `TischSessionSchema` (`frontend/src/service/table/Tisch.ts:18-25`) und den
   Historie-Schemas.
5. `Backend.post` macht aus dem Zod-Fehler einen generischen
   `ResponseBodyError` und loggt nur `console.error`
   (`frontend/src/lib/Backend.ts:104-111`) → im Network-Tab sieht das JSON
   „korrekt" aus, die UI fällt still auf Defaults zurück.
6. Folge: `useTischState` → `DEFAULT_TISCH_STATE` (0 €);
   `useMeineTischeState` → `[]` → `TableSelectionPage` zeigt „Keine Tische
   markiert" (`frontend/src/service/TableSelectionPage.tsx:15-16,33-48`).

Der Favoriten-Drawer selbst (`useAktiveTischeMitFavoriten`,
`AktiverTischMitFavoritSchema`) trägt **keine** Positionen und funktioniert —
das wahrgenommene „Favoriten weg" stammt aus der leeren „Meine Tische"-Liste.

## Architectural decisions

- **Positionen tragen ihren Steuersatz in jeder HTTP-Response.** Das HTTP-DTO
  spiegelt das Domain-Modell (das den Steuersatz bereits führt). Konsistent
  über Tisch- und Direktverkauf-Bereich.
- **Steuersatz-Werte:** `regel` | `ermaessigt` | `befreit` | `kombi`
  (`backend/domain/steuer/steuer.go:5-11`) — exakt deckungsgleich mit dem
  Frontend-Enum. Keine Übersetzung nötig.
- **POST-only, Geld in Cent (int)** bleiben unverändert.
- **Beide Seiten validieren** (zog/Zod) — bestehendes Prinzip; der Fix stellt
  die Übereinstimmung wieder her statt eine Seite aufzuweichen.

## Inventory

Backend:

- `backend/api/table/http/query_handler.go:111-130` — `position` DTO +
  `toPosition` (Bug: kein `Steuersatz`)
- `backend/api/table/http/query_handler.go:298-336` — `getTischStateResponse`
  - Handler
- `backend/api/table/http/query_handler.go:380-420` — `getMeineTischeState`
  (nutzt `toPositionen`)
- `backend/api/direktverkauf/http/query_handler.go:20-47` — eigenes
  `position` DTO + `toPosition` (gleiche Auslassung, aktuell konsistent zum
  FE-Schema → latent)
- `backend/domain/kasse/bestellung.go:10-31` — Domain `Position` +
  `positionEventData` (führen `Steuersatz` bereits)
- `backend/domain/steuer/steuer.go:5-11` — Steuersatz-Konstanten

Frontend:

- `frontend/src/service/table/Bestellung.ts:3-13` — `PositionSchema`
  (verlangt `steuersatz`)
- `frontend/src/service/table/Tisch.ts:18-25` — `TischSessionSchema`
- `frontend/src/service/direktverkauf/Direktverkauf.ts:38-46` —
  `VerkaufPositionSchema` (ohne `steuersatz`)
- `frontend/src/service/TableSelectionPage.tsx:15-16` — `useMeineTischeState`
- `frontend/src/service/components/table/drawerUtils.test.ts` — Fixtures
  führen `steuersatz` bereits
- `frontend/src/lib/Backend.ts:104-111` — Zod-Fehler-Behandlung (zu still)

Zurückzunehmende Symptom-Pflaster (aus vorherigen Turns):

- `frontend/src/lib/Backend.ts` — `AbortSignal.timeout(15_000)`
- `frontend/src/main.tsx` — `QueryClient`-Defaults (`staleTime`,
  `refetchOnWindowFocus`, `retry`)
- `frontend/src/service/TablePage.tsx` — Fehlerbanner, `hasData`/`loadFailed`,
  `isError`-Verzweigungen
- `frontend/src/service/table/hooks.ts` — `enabled`, `refetchInterval`,
  `isError` in `useTischState`/`useTischHistorie`

## Resolved decisions (aus Clarify)

- **Fix-Umfang:** Alle Positions-DTOs konsistent — Tisch **und** Direktverkauf
  im Backend, plus `VerkaufPositionSchema` im Frontend.
- **Pflaster:** Alle vorherigen Symptom-Änderungen vollständig zurücknehmen;
  nur der Root-Cause-Fix bleibt.
- **Härtung:** (1) Regressionstest, der `steuersatz` in der Backend-Response
  prüft; (2) Zod-Fehler im Frontend im Dev sichtbarer machen (Feldpfad/Endpoint
  in Fehlermeldung statt nur generisch).

## Open questions / Risks

- **Verifiziert:** Das Read-Model liefert den Steuersatz bereits.
  `backend/api/table/application/command.go:453` setzt beim Bestellen
  `Steuersatz: string(prod.Steuersatz)`; die Tisch-Session-Projektion baut
  Positionen mit Steuersatz (`backend/domain/kasse/tisch_session_test.go:76`),
  und die Domain-Validierung verlangt ihn
  (`backend/domain/kasse/bestellung.go:64`). Einzige Lücke ist das HTTP-DTO →
  Phase 1 ist exakt „Feld + Mapper ergänzen", nichts Größeres.
- `make sqlc` ist **nicht** nötig — keine Query-Änderung.
- Dev-DB neu aufsetzen ist nicht erforderlich; bestehende Events tragen
  `steuersatz` bereits im Event-Data.

---

## Phase 1: Tisch-Positionen tragen Steuersatz (behebt den akuten Bug)

### Context

- `backend/api/table/http/query_handler.go:111-130` — DTO + `toPosition`
- `backend/api/table/http/query_handler.go:298-336,380-420` — betroffene
  Responses
- `backend/domain/kasse/bestellung.go:10-18` — Domain liefert `Steuersatz`
- `frontend/src/service/table/Bestellung.ts:3-13` — erwartet `steuersatz`

### What to build

Das HTTP-DTO `position` im Tisch-Bereich erhält das Feld `Steuersatz string`
mit JSON-Key `steuersatz`, und `toPosition` mappt `p.Steuersatz`. Da
`toPositionen` geteilt ist, erscheinen Steuersätze automatisch in
`get-tisch-state`, `get-tisch-historie` und `get-meine-tische-state`. Damit
parst das Frontend `TischSession`/`Position` wieder erfolgreich; die
Tisch-Seite zeigt echten Saldo und offene Positionen, die Service-Startseite
zeigt die markierten Tische.

### Acceptance criteria

- [x] Response von `service/get-tisch-state` enthält pro Position einen
      `steuersatz` aus `{regel, ermaessigt, befreit, kombi}`.
- [x] `get-tisch-historie` und `get-meine-tische-state` enthalten ebenfalls
      `steuersatz`.
- [x] Regressionstest im Handler-Test prüft, dass `steuersatz` im Response
      vorhanden und nicht leer ist.
- [x] `make test` (Backend Unit-Tests) grün.
- [x] Manuell im lokalen Setup: Tisch mit offenen Positionen zeigt korrekten
      Saldo statt 0 €; Startseite zeigt markierte Tische.

---

## Phase 2: Symptom-Pflaster zurücknehmen (sauberer Baseline-Zustand)

### Context

- `frontend/src/lib/Backend.ts`, `frontend/src/main.tsx`,
  `frontend/src/service/TablePage.tsx`, `frontend/src/service/table/hooks.ts`

### What to build

Alle in vorherigen Turns hinzugefügten Symptom-Änderungen werden entfernt, bis
diese vier Dateien funktional dem Stand von `origin/main` entsprechen
(abgesehen vom Härtungs-Diff aus Phase 4 in `Backend.ts`). Konkret: Fetch-
Timeout, QueryClient-Defaults, TablePage-Fehlerbanner samt `hasData`/
`loadFailed`/`isError`-Logik, sowie `enabled`/`refetchInterval`/`isError` in
den Hooks.

### Acceptance criteria

- [x] `TablePage.tsx` und `hooks.ts` entsprechen wieder dem ursprünglichen
      Verhalten (kein Fehlerbanner, keine `hasData`-Gates).
- [x] `main.tsx` nutzt wieder `new QueryClient()` ohne Custom-Defaults.
- [x] `frontend` Lint + Typecheck grün (`make lint`).
- [x] Verhalten der Tisch-Seite identisch zum Stand vor den Pflastern — nur
      dass dank Phase 1 die Daten jetzt korrekt sind.

---

## Phase 3: Direktverkauf-Positionen tragen Steuersatz (Konsistenz)

### Context

- `backend/api/direktverkauf/http/query_handler.go:20-47` — DTO + `toPosition`
- `frontend/src/service/direktverkauf/Direktverkauf.ts:38-46` —
  `VerkaufPositionSchema`

### What to build

Das Direktverkauf-`position`-DTO erhält ebenfalls `Steuersatz` (JSON
`steuersatz`), `toPosition` mappt es, und `VerkaufPositionSchema` ergänzt
`steuersatz: z.enum([...])`. Damit tragen Positionen im Direktverkauf- und
Tisch-Bereich einheitlich ihren Steuersatz — abgestimmt auf Domain, Events und
die TSE/Steuer-Roadmap.

### Acceptance criteria

- [x] Direktverkauf-Query-Responses enthalten `steuersatz` pro Position.
- [x] `VerkaufPositionSchema` verlangt `steuersatz`; Direktverkauf-Ansichten
      parsen und rendern weiterhin korrekt.
- [x] Direktverkauf-Tests (Backend + Frontend) grün.

---

## Phase 4: Schema-Mismatches im Dev früh sichtbar machen (Härtung)

### Context

- `frontend/src/lib/Backend.ts:104-111` — `responseSchema.safeParse` →
  `ResponseBodyError`, nur `console.error(error.message)`

### What to build

Bei fehlgeschlagener Response-Validierung wird die Fehlermeldung aussagekräftig:
Endpoint **und** der/die betroffenen Zod-Feldpfade (`issues[].path`) werden in
die `ResponseBodyError`-Message bzw. das Dev-Log aufgenommen. So zeigt ein
künftiger Mismatch sofort „Response of service/get-tisch-state is invalid:
ausstehendePositionen[0].steuersatz (invalid_type)" statt einer generischen
Meldung. Kein Verhaltenswechsel in Produktion außer einer präziseren Meldung.

### Acceptance criteria

- [ ] Bei einer simulierten Schema-Abweichung nennt die Fehlermeldung Endpoint
      und Feldpfad.
- [ ] Bestehende `Backend`-Unit-Tests bleiben grün; ggf. ein Test, der die
      angereicherte Meldung prüft.
- [ ] Kein Leak sensibler Daten in die Meldung (nur Pfade/Typen, keine Werte).
