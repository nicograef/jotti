# Plan: Bestellung umbuchen (K-09)

> Source PRD: ../prds/prd-bestellung-umbuchen.md

## Goal

Serviceleitung und Admin erhalten an jeder Bestellung in der Tisch-Historie die
Aktion **„Umbuchen"**, die eine noch **nicht bezahlte** Bestellung in einem Schritt
vom Quell- auf den Ziel-Tisch verschiebt. Im Hintergrund entstehen **atomar** (eine
Transaktion) genau zwei Vorgänge: eine **Stornierung** am Quell-Tisch und eine **neue
Bestellung** mit Original-Preisen am Ziel-Tisch. Die Umbuchung ist bargeld-neutral
(verändert die Kassenlade nicht), beschränkt auf unbezahlte Positionen, und löst am
Ziel-Tisch **keinen** neuen Arbeitsbon aus.

## Architectural decisions

Durable decisions that apply across all phases:

- **Route**: `POST /serviceleitung/bestellung-umbuchen` (registriert in
  `backend/api/serviceleitung.go` neben `stornierung-erteilen`; nur `serviceleitung`
  und `admin`).
- **Request-Vertrag**:
  `{ "quellTischId": int, "zielTischId": int, "positionen": [{ "positionId": uuid, "menge": int }] }`.
  **Kein** `kommentar`-Feld (serverseitig automatisch gesetzt). Antwort: leerer Erfolg.
- **Keine neuen Event-Typen, kein DB-Schema-Change**: Wiederverwendung von
  `bestellung-aufgenommen:v1` (Ziel) und `stornierung-erteilt:v1` (Quelle). Beide
  `tisch_sessions`-Projektionen werden in derselben Transaktion aktualisiert.
- **Atomarität & OCC**: Eine neue Repo-Methode schreibt beide Events + beide
  Projektionen in **einer** Transaktion. Für beide Subjects wird die nächste Version
  via `GetMaxVersion` vergeben; ein `UNIQUE(subject, version)`-Verstoß rollt die ganze
  Transaktion zurück (→ `ErrConflict`).
- **Eligibility = unbezahlt**: Validierung gegen `state.UnbezahltePositionen` (nicht
  gegen die nicht-stornierten Positionen wie bei der Stornierung). Auslieferungsstatus
  ist irrelevant.
- **Werterhaltung**: Ziel-Positionen werden direkt aus den aufgelösten Quell-Positionen
  gebildet (Original-`Einzelpreis`); kein erneuter Produkt-/Varianten-Lookup. Frische
  `PositionID`s am Ziel.
- **Kein Arbeitsbon**: schlichter Event-Write-Pfad (nicht der
  `WriteEventWithDruckauftraege`-Outbox-Pfad).

## Inventory

Relevante bestehende Dateien, Muster und Integrationspunkte:

- `backend/api/serviceleitung.go:16-43` — Router; verdrahtet `tableApp.Command` und
  registriert `stornierung-erteilen` + `auszahlung-leisten`. Hier kommt die neue Route hin.
- `backend/repository/kassenjournal_repo/repo.go:32-135` — `WriteEvent`,
  `WriteEventWithDruckauftraege`, `writeEventInTx`, `handleTischSessionEvent`. Muster für
  die neue atomare Methode (beide nutzen denselben transaktionalen `writeEventInTx`-Kern).
- `backend/repository/kassenjournal_repo/repo.go:308-318` — `GetMaxVersion` (OCC-Basis).
- `backend/repository/kassenjournal_repo/mock.go:49-72` — `MockRepo.WriteEvent` /
  `WriteEventWithDruckauftraege`; hier kommt die Mock-Erweiterung hin.
- `backend/repository/kassenjournal_repo/repo_test.go:225-318` — Prior art für
  Commit/Rollback-Integrationstests (`TestWriteEventWithDruckauftraege_*`).
- `backend/api/table/application/command.go:120-145` — `writeEventOCC` (Versionierung +
  Conflict-Mapping), das Muster für die Zwei-Subject-Versionierung.
- `backend/api/table/application/command.go:500-539` — `StornierungErteilen` als nächster
  Nachbar; `command.go:401-447` — `resolvePositions` (PositionRef → fat Position + Summe).
- `backend/api/table/application/command.go:181-219` — `loadTischState` (offene
  Kassensitzung, Tisch-Existenz/Status, Subject, Projektion).
- `backend/api/table/application/errors.go:10-49` — Fehlerkatalog; hier kommen
  `ErrPositionNichtUmbuchbar` und `ErrUmbuchungGleicherTisch` hinzu.
- `backend/domain/kasse/tisch_session_events.go:89-164` — `NewBestellungAufgenommenEvent`
  (frische PositionIDs, Summe aus Einzelpreisen) und `NewStornierungErteiltEvent`.
- `backend/domain/kasse/tisch_session_events.go:50-62` — `stornierungErteiltV1DataSchema`
  (`Kommentar` `Min(3).Max(100)`) und `bestellungAufgenommenV1DataSchema:27-33`
  (`Kommentar` `Max(100)`). Auto-Kommentar muss in `Max(100)` passen.
- `backend/domain/kasse/subject.go:14-16` — `TischSessionSubject`.
- `backend/domain/table/tisch.go:46` — `TischNameSchema` erlaubt bis 100 Zeichen
  (Grund für die Kommentar-Kürzung).
- `backend/api/table/http/command_handler.go:17-31` — `command`-Interface (Methode
  ergänzen); `command_handler.go:466-510` — `stornierungErteilen{Request,Schema,Handler}`
  als Vorlage; `command_handler.go:252-271,278-291` — `positionRefRequest` + `toPositionRefs`.
- `backend/api/table/http/command_handler_test.go:52,267-277` — Mock-Command + Handler-Test-Muster.
- `frontend/src/service/table/TischBackend.ts:75-80` — `stornierungErteilen`-Methode als Vorlage.
- `frontend/src/service/table/Stornierung.ts:17-22` — `StornierungErteilenSchema` als Vorlage.
- `frontend/src/service/components/table/HistorieStornierungDrawer.tsx:1-205` — Vorlage für
  den neuen Drawer (Mengenauswahl, `useActionSubmit`, `toPositionRefs`).
- `frontend/src/service/components/table/TischHistorie.tsx:151-163,300-352` — „Stornieren"-
  Aktion + privater Helper `getStornierbarePositionen` (wird nach `drawerUtils.ts` verschoben).
- `frontend/src/service/components/table/drawerUtils.ts:1-72` — Ziel für beide Helfer;
  `drawerUtils.test.ts:1-79` — Vitest-Vorlage.
- `frontend/src/service/table/hooks.ts:10-19` — `useAktiveTische` (Quelle der aktiven
  Tische für den Ziel-Tisch-Picker; Quell-Tisch ausschließen).
- `frontend/src/service/TablePage.tsx:150-160` — Einbindung von `TischHistorie`
  (`onStornierungErteilt` → `reload`); analoge Verdrahtung für Umbuchung.
- `frontend/src/lib/errorMessages.ts` — zentrale Fehlercodes (`kasse_nicht_geoeffnet` bereits
  zentral; `position_nicht_umbuchbar` erhält eine Drawer-lokale Meldung via `byCode`).

## Resolved decisions

- **Auto-Kommentar-Länge**: Das `Kommentar`-Limit bleibt bei `Max(100)`. Der
  Auto-Kommentar (`„Umbuchung auf Tisch <Name>"` / `„Umbuchung von Tisch <Name>"`) wird
  so gebildet, dass er garantiert ≤ 100 Zeichen ist (Tischname bei Bedarf kürzen). Kein
  globaler Schema-Change über K-09 hinaus.
- **Frontend-Helfer**: Sowohl der bestehende `getStornierbarePositionen` als auch der
  neue `getUmbuchbarePositionen` liegen in `drawerUtils.ts` und werden per Vitest
  getestet (`getStornierbarePositionen` wird dabei aus `TischHistorie.tsx` herausgezogen).
- **Erfolgs-Feedback**: Nach erfolgreicher Umbuchung Erfolgs-Toast
  („Bestellung umgebucht.") **plus** Reload von Historie und Saldo (erfüllt US-22).

## Open questions / Risks

- **Konsistenz-Asymmetrie**: Die Umbuchung validiert gegen `UnbezahltePositionen`
  (Projektion), die Stornierung gegen on-demand-replay nicht-stornierte Positionen. Das
  ist bewusst (PRD „Further Notes") — beim Self-Review nicht versehentlich angleichen.

---

## Phase 1: Atomarer Cross-Tisch-Write (`WriteUmbuchung`)

**User stories**: US-7 (atomar), US-17 (Nebenläufigkeit/OCC-Rollback); Fundament für
US-12, US-13, US-21, US-23.

### Context

- `backend/repository/kassenjournal_repo/repo.go:32-135` — `WriteEvent` /
  `WriteEventWithDruckauftraege` / `writeEventInTx`: zeigen, wie Event-INSERT und
  Projektions-Update in einer Transaktion laufen; der neue Schreibpfad nutzt denselben
  `writeEventInTx`-Kern zweimal (beide `StreamTypeTischSession`).
- `backend/repository/kassenjournal_repo/mock.go:49-72` — Mock-Schreibmethoden, die um
  die neue Methode erweitert werden (für die Unit-Tests von Phase 2).
- `backend/repository/kassenjournal_repo/repo_test.go:225-318` — Commit- und
  Rollback-Integrationstests als Vorlage.

### What to build

Eine neue Repository-Methode, die zwei **bereits versionierte** Tisch-Session-Events
(Stornierung für den Quell-Subject, Bestellung für den Ziel-Subject) zusammen mit der
`kassensitzungNr` entgegennimmt und in **einer** Transaktion beide Events ins
`kassenjournal` schreibt **und** beide `tisch_sessions`-Projektionen aktualisiert. Bei
einem Fehler in einem der beiden Schritte (inkl. Versions-/`UNIQUE`-Konflikt) wird die
gesamte Transaktion zurückgerollt — es entsteht nie nur die Storno- oder nur die
Bestell-Seite. Der Mock des Repos wird um dieselbe Methode erweitert, sodass die
Anwendungsschicht in Phase 2 dagegen testen kann.

### Acceptance criteria

- [x] Neue Repo-Methode schreibt Storno- (Quelle) und Bestellung-Event (Ziel) plus beide
      Projektionen in einer einzigen DB-Transaktion.
- [x] Erfolgsfall (Integrationstest): Am Quell-Subject existiert das Storno-Event, am
      Ziel-Subject das Bestellung-Event; beide `tisch_sessions` zeigen korrekte Salden
      (Quelle reduziert, Ziel erhöht) und Positionslisten.
- [x] Rollback/Atomarität (Integrationstest): Schlägt der Ziel-Write fehl, bleibt auch
      der Quell-Storno aus (kein halber Zustand, keine Projektionsänderung).
- [x] OCC-Konflikt (Integrationstest): Eine kollidierende Version an einem der beiden
      Subjects führt zum vollständigen Rollback und wird als Konflikt erkennbar.
- [x] Der Mock implementiert die neue Methode konsistent (kein realer Outbox-/Druckpfad).
- [x] `make test` / Integrationstests grün; `make lint` ohne Befunde.

---

## Phase 2: Command `BestellungUmbuchen` + Endpunkt

**User stories**: US-1, US-2, US-3, US-4, US-5, US-6, US-8, US-9, US-10, US-11, US-12,
US-13, US-14, US-16, US-18, US-19, US-20, US-21, US-23.

### Context

- `backend/api/table/application/command.go:500-539` — `StornierungErteilen` (nächster
  Nachbar: Eligibility-Check, `resolvePositions`, Event-Bau, OCC-Write).
- `backend/api/table/application/command.go:120-145,181-219,401-447` — `writeEventOCC`,
  `loadTischState`, `resolvePositions` (wiederverwendbare Bausteine).
- `backend/domain/kasse/tisch_session_events.go:89-164` — `NewBestellungAufgenommenEvent`
  (frische PositionIDs, Summe aus Einzelpreisen) + `NewStornierungErteiltEvent`.
- `backend/api/table/application/errors.go:10-49` — Fehlerkatalog (neue Fehler ergänzen).
- `backend/api/serviceleitung.go:23-31` — `tableApp.Command`-Verdrahtung + Routen.
- `backend/api/table/http/command_handler.go:17-31,466-510` — `command`-Interface +
  `stornierungErteilen{Request,Schema,Handler}` als Vorlage.
- `backend/api/table/http/command_handler_test.go:52,267-277` — Mock-Command + Handler-Test.
- `backend/api/table/application/command_test.go:427-521` — Stornierungs-Unit-Tests als
  Vorlage (Mock-basiert).

### What to build

Eine neue Command-Methode in der Anwendungsschicht, die den vollständigen Umbuchungs-
Ablauf orchestriert: offene Kassensitzung prüfen → Gleicher-Tisch-Guard
(`quellTischId != zielTischId`) → Quell- und Ziel-Tisch laden (Existenz/Status/Subject)
→ angeforderte Positionen gegen die **unbezahlten** Positionen des Quell-Tischs
validieren → zu vollständigen (fat) Positionen mit Original-Preisen auflösen →
Storno-Event (Quelle) und Bestellung-Event (Ziel) mit den Auto-Kommentaren bauen → beide
Subjects versionieren → die atomare Repo-Methode aus Phase 1 aufrufen. Neue
Anwendungsfehler `ErrPositionNichtUmbuchbar` und `ErrUmbuchungGleicherTisch`;
wiederverwendet werden `ErrConflict`, `ErrKasseNichtGeoeffnet`, `ErrTischNotFound`,
`ErrTischNotActive`. Dazu der HTTP-Handler mit Request-DTO + zog-Schema (`quellTischId`,
`zielTischId`, `positionen`, **kein** `kommentar`), Fehler-Mapping für alle Fälle, und
die Registrierung der Route `POST /serviceleitung/bestellung-umbuchen`.

### Acceptance criteria

- [x] Endpunkt `POST /serviceleitung/bestellung-umbuchen` ist registriert und nur für
      `serviceleitung`/`admin` erreichbar; Request-Vertrag wie in den Architektur-
      entscheidungen, Antwort leerer Erfolg.
- [x] Happy Path (Unit-Test): korrekte Storno- (Quell-Subject) und Bestellung-Events
      (Ziel-Subject) mit übernommenen Positionen/Mengen/**Original-Preisen**, frischen
      Ziel-`PositionID`s und den Auto-Kommentaren („Umbuchung auf/von Tisch <Name>").
- [x] Auto-Kommentar ist garantiert ≤ 100 Zeichen (langer Tischname wird gekürzt) und
      besteht die Event-Schema-Validierung (Storno `Min(3).Max(100)`).
- [x] Eligibility (Unit-Test): Anforderung einer bereits **bezahlten** Position →
      `ErrPositionNichtUmbuchbar` (→ `position_nicht_umbuchbar`).
- [x] Gleicher-Tisch-Guard (Unit-Test): `quellTischId == zielTischId` →
      `ErrUmbuchungGleicherTisch` (→ `umbuchung_gleicher_tisch`).
- [x] Ziel-Tisch inaktiv/unbekannt → `ErrTischNotActive` / `ErrTischNotFound`; keine
      offene Kassensitzung → `ErrKasseNichtGeoeffnet`; Repo-Versionskonflikt →
      `ErrConflict` (HTTP-Konflikt).
- [x] Kein Druckauftrag am Ziel-Tisch (schlichter Write-Pfad, kein Outbox-Pfad).
- [x] `make test` grün; `make lint` ohne Befunde.

---

## Phase 3: Frontend-Umbuchung (Drawer + Aktion)

**User stories**: US-1, US-2, US-3, US-4, US-5, US-6, US-16, US-20, US-22.

### Context

- `frontend/src/service/table/TischBackend.ts:75-80` — `stornierungErteilen` (Vorlage für
  die neue Backend-Methode über den `BackendClient`).
- `frontend/src/service/table/Stornierung.ts:17-22` — `StornierungErteilenSchema`
  (Vorlage für das neue Zod-Schema mit `quellTischId`/`zielTischId`, ohne `kommentar`).
- `frontend/src/service/components/table/HistorieStornierungDrawer.tsx:1-205` — Vorlage für
  den neuen Drawer (Mengenauswahl, `useActionSubmit`, `toPositionRefs`); **ohne**
  Kommentarfeld, **mit** Ziel-Tisch-Auswahl.
- `frontend/src/service/components/table/TischHistorie.tsx:151-163,300-352` — „Stornieren"-
  Aktion + privater Helper `getStornierbarePositionen` (nach `drawerUtils.ts` verschieben).
- `frontend/src/service/components/table/drawerUtils.ts:1-72` /
  `drawerUtils.test.ts:1-79` — Ziel für beide Helfer + Vitest-Vorlage.
- `frontend/src/service/table/hooks.ts:10-19` — `useAktiveTische` (Ziel-Tisch-Liste; Quell-
  Tisch ausschließen).
- `frontend/src/service/TablePage.tsx:150-160` — Einbindung/Reload-Verdrahtung von
  `TischHistorie`.
- `frontend/src/lib/errorMessages.ts` — zentrale Fehlercodes; `position_nicht_umbuchbar`
  erhält eine verständliche Drawer-lokale Meldung (Auswahl aktualisieren).

### What to build

Der Frontend-Pfad für die Umbuchung. `TischBackend` erhält eine `bestellungUmbuchen`-
Methode, die ein neues Zod-Schema (`quellTischId`, `zielTischId`, `positionen`) validiert
und an `serviceleitung/bestellung-umbuchen` postet. Beide Positions-Helfer
(`getStornierbarePositionen` und neu `getUmbuchbarePositionen` — letzterer zieht je
`PositionID` **storniert und bezahlt** ab) liegen in `drawerUtils.ts` und sind
unit-getestet. Ein neuer Drawer (Spiegel von `HistorieStornierungDrawer`) bietet die
Mengenauswahl der umbuchbaren Positionen **plus** eine Ziel-Tisch-Auswahl aus den aktiven
Tischen (Quell-Tisch ausgeschlossen), **ohne** Kommentarfeld. In `TischHistorie` erhält
jede Bestellung eine „Umbuchen"-Aktion neben „Stornieren", sichtbar nur bei
`AuthSingleton.canCancel` **und** wenn umbuchbare Positionen vorhanden sind. Nach Erfolg:
Toast „Bestellung umgebucht." + Reload von Historie und Saldo.

### Acceptance criteria

- [x] `getUmbuchbarePositionen` liefert nur unbezahlte, nicht-stornierte Positionen je
      Bestellung mit korrekten Restmengen nach Teil-Zahlung/Teil-Stornierung; leere Liste,
      wenn alles bezahlt/storniert (Vitest). `getStornierbarePositionen` bleibt korrekt
      und ist nun ebenfalls getestet.
- [x] „Umbuchen"-Aktion erscheint nur bei `canCancel` **und** vorhandenen umbuchbaren
      Positionen; bei leerer Auswahl wird kein Drawer geöffnet (US-20).
- [x] Drawer erlaubt Mengenauswahl (Default: alle umbuchbaren Positionen) **und** Ziel-
      Tisch-Auswahl aus aktiven Tischen ohne den Quell-Tisch; kein Kommentarfeld.
- [x] `bestellungUmbuchen` validiert per Zod und ruft den Endpunkt über den
      `BackendClient` auf (kein direktes `fetch`).
- [x] Fehlercode `position_nicht_umbuchbar` zeigt eine verständliche deutsche Meldung
      (Auswahl aktualisieren) (US-16).
- [x] Nach Erfolg: Toast „Bestellung umgebucht." + Reload von Historie und Saldo (US-22).
- [x] `make lint-frontend` und `make test-frontend` grün.
