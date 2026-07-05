# Plan: Audit-Fixes B1–B7 + Randnotizen

> Source PRD: n/a (aus Code-Audit vom 2026-07-05, Stand `d0e3602`)

## Goal

Alle sieben Audit-Befunde (B1–B7) und alle sieben Randnotizen (R1–R7) sauber
beheben. Maßstab ist Qualität, Compliance und UX — nicht Aufwand oder
Breaking-Change-Vermeidung. Nach jeder Phase bleibt `make verify` grün.

## Architectural decisions

Durable decisions, die über alle Phasen gelten:

- **Keine neuen Routen, keine neuen Event-Typen.** Alle Fixes bleiben in den
  bestehenden POST-Endpunkten. B3 nutzt die vorhandenen Event-Typen
  (`stornierung-erteilt:v1` = Warenrücknahme, `bestellung-korrigiert:v1` =
  geldneutrale Korrektur); der Storno-Art-Diskriminator wird zur **Lesezeit**
  aus dem Event-Typ abgeleitet (`buildStornierungFromEvent` vs.
  `buildKorrekturFromEvent`), nie im Event gespeichert. Die Events bleiben
  immutable und unverändert.
- **Diskriminator-Feldname `barRueckgabe` (bool).** Das Reporting kennt dieses
  Feld bereits (`GetStornierungen`-SQL `reporting.sql:78`, DTO
  `query_handler.go`, Frontend `types.ts` `StornierungDetailSchema`). Die
  Tisch-Historie spiegelt denselben Namen und dieselbe Semantik
  (Warenrücknahme + Direktverkauf-Storno = `true`), statt einen neuen Begriff
  einzuführen.
- **0 €-Anfangsbestand ist gültig.** Die UI wird an das Backend angeglichen
  (Backend + fiskalische Projektion sind die fiskalisch korrekte Seite:
  Anfangsbestand 0 → keine Bareinlage → kein signaturpflichtiger Vorgang).
- **Login-Schutz: In-Memory-Soft-Throttle pro Konto.** Kein Schema, keine
  persistente Sicherheits-Statushaltung; automatisch ablaufender
  exponentieller Cooldown, gespiegelt am Muster von `RateLimitMiddleware`
  (`middleware.go:78-131`). Injiziert in die Auth-`Command` als Interface
  (Infrastruktur, kein Domain-State).
- **OTP-Zähler wird transaktional.** Der Set-Password-Pfad läuft als eine
  Transaktion mit Zeilensperre (`SELECT … FOR UPDATE` auf die User-Zeile),
  nach dem Callback-in-TX-Muster von
  `kassenjournal_repo.EroeffneKassensitzung` (`repo.go:146-179`).

## Inventory

Relevante bestehende Dateien und Muster:

- `frontend/src/admin/kasse/Kassensitzung.ts:15-40` — `KassensitzungStatus`-Enum
  und `KassensitzungSchema` (nur `offen | abgeschlossen`).
- `frontend/src/admin/reporting/ReportingFilter.tsx:54` — Status-Rendering
  (`🟢 offen` / `🔴 sonst`), Konsument der Kassensitzungs-Liste.
- `backend/api/reporting/http/query_handler.go:262` /
  `backend/api/kasse/kassenfuehrung/http/query_handler.go:52-56` — geben
  `status` roh als String durch (DB-CHECK: `offen |
  wird_abgeschlossen | abgeschlossen`, `01_initial.up.sql:127`).
- `frontend/src/admin/kasse/EroeffnenSection.tsx:44-58` +
  `frontend/src/admin/kasse/KasseBackend.ts:15-18` — erzwingen `betragCents ≥ 1`;
  Backend erlaubt `≥ 0` (`command_handler.go:34`,
  `fiskalische_projektion.go:94`).
- `backend/domain/kasse/historie.go:55-70` — mappt `stornierung-erteilt` und
  `bestellung-korrigiert` einheitlich auf `HistorieEintragStornierung`.
- `backend/domain/kasse/tisch_session_events.go:338-404` —
  `buildStornierungFromEvent` (Warenrücknahme) und `buildKorrekturFromEvent`
  (Korrektur); `backend/domain/kasse/stornierung.go` — `Stornierung`-Struct.
- `backend/api/kasse/tischgeschaeft/http/query_handler.go:184-208` — HTTP-DTO
  `stornierung`.
- `frontend/src/service/table/Stornierung.ts` — `StornierungSchema`
  (Response) mit `kommentar` min 3 (R6); `StornierungErteilenSchema` (Request).
- `frontend/src/service/components/table/TischHistorie.tsx:190-220` — Storno-
  Rendering ohne Beleg-Button; `frontend/src/service/table/TischBackend.ts:69-82`
  — `belegDrucken(tischId, zahlungId)` (nur Zahlung).
- `frontend/src/service/components/direktverkauf/DirektverkaufHistorie.tsx:115-165`
  — **Vorlage** für den Storno-Beleg-Button (`belegAnfordern({verkaufId,
  stornierungId})`).
- `backend/api/druck/beleg/application/kassenbeleg_command.go:261-296` — der
  fertige, bislang ungenutzte Tisch-Warenrücknahme-Beleg-Pfad
  (`{tischId, stornierungId}`).
- `backend/api/middleware/middleware.go:139-146` — `clientIP` ohne
  `SplitHostPort` (B4); `:85-131` — `RateLimitMiddleware` (Muster für Throttle).
- `backend/api/auth/application/command.go:23-112` — `GenerateJWTToken`,
  `SetNewPassword`; `backend/domain/user/user.go:182-217` — `SetPassword`
  (nicht-atomarer Zähler).
- `backend/domain/kasse/tagesabschluss_summen.go:30-73` —
  `BerechneAbschlussSummen` (still überspringend, R1);
  `backend/api/kasse/kassenfuehrung/application/command.go:366` — Aufrufer.
- `backend/domain/kasse/tisch_session.go:207-278` —
  `accumulatePositionen`/`reduceByPosition`/`tagBesteller` (mutierend, R2).
- `backend/repository/kassensitzungen_repo/repo.go:99-110` +
  `mock.go:80` — toter `InsertKassensitzung` (B5).
- `backend/domain/kasse/bestellung.go:72` +
  `backend/domain/kasse/fiskalische_projektion.go:167` — doppelte
  Positions-Mapping-Funktion (B6).
- `backend/api/kasse/tischgeschaeft/application/command.go` —
  `AusgabeBestaetigen` (dupliziert `persistTischEvent`, B7);
  `query.go:124-177` — `GetMeineTischeState` (N+1, R3).
- `backend/api/druck/beleg/http/command_handler.go:75-…` — Vier-Body-Form-
  Dispatch (R5).
- `backend/app/app.go:94-119` — `fmt.Println`/`log.Printf` neben zerolog (R7).

## Resolved decisions

- **B2:** UI auf `≥ 0` lockern (Backend ist die fiskalisch korrekte Seite).
- **R4:** Soft-Throttle pro Konto (auto-ablaufend), zusätzlich zum reparierten
  IP-Limit; keine harte Konto-Sperre (Footgun für Helfer im Event).
- **B3:** Storno-Arten in der Historie **sichtbar unterscheiden**
  (Warenrücknahme vs. geldneutrale Korrektur), Beleg-Button nur bei der
  kassenwirksamen Warenrücknahme. Überschreibt die frühere „einheitlich
  Storno"-Darstellung bewusst zugunsten von Transparenz.

## Open questions / Risks

- **R1-Verhaltensänderung:** Ein streng fehlschlagender Tagesabschluss blockiert
  den Abschluss bei einem nicht-parsebaren Event. Das kann bei validierten
  Single-Version-Events praktisch nicht auftreten (Schutz gegen einen
  korrupten Store); im Ernstfall ist Blockieren korrekter als ein stiller,
  falscher Z-Bon. Wird in der Commit-Message dokumentiert.
- **B3 berührt eine dokumentierte frühere Entscheidung** (Storno einheitlich).
  Vom Nutzer bestätigt (sichtbar unterscheiden), daher kein Risiko mehr, aber
  in `docs/` (Handbuch/Language) nachziehen.

---

## Phase 1: Kassensitzungs-Status im Frontend vervollständigen (B1)

### Context

- `frontend/src/admin/kasse/Kassensitzung.ts:15-40` — Enum + Schema kennen
  `wird_abgeschlossen` nicht → Zod wirft `ResponseBodyError`.
- `frontend/src/admin/reporting/ReportingFilter.tsx:54` — Status-Label.
- `backend/…/reporting/http/query_handler.go:262` — sendet den Roh-Status.

### What to build

Der transiente Status `wird_abgeschlossen` wird im Frontend ein erstwertiger
Status: `KassensitzungStatus`-Enum und alle `status`-Zod-Enums
(`KassensitzungSchema` in `kasse/` und die Re-Exports im Reporting) erhalten den
dritten Wert, und die Statusanzeige bekommt ein eigenes, unmissverständliches
Label (z. B. 🟡 „wird abgeschlossen…"). Damit überlebt die Admin-Abrechnung eine
Sitzung im Zwischenstatus statt komplett zu brechen.

### Acceptance criteria

- [x] `KassensitzungSchema.status` akzeptiert `offen | wird_abgeschlossen | abgeschlossen`.
- [x] Kein `ResponseBodyError` mehr, wenn `get-all-kassensitzungen` oder
      `get-offene-kassensitzung` eine Sitzung im Zwischenstatus liefert.
- [x] Die Statusanzeige (Filter/Liste) zeigt für `wird_abgeschlossen` ein
      eigenes Label, nicht das „abgeschlossen"-/🔴-Symbol.
- [x] Vitest deckt die drei Statuswerte im Schema und im Status-Label ab.

---

## Phase 2: 0 €-Anfangsbestand in der UI erlauben (B2)

### Context

- `frontend/src/admin/kasse/KasseBackend.ts:15-18` —
  `KassensitzungEroeffnenSchema` erzwingt `≥ 1`.
- `frontend/src/admin/kasse/EroeffnenSection.tsx:44-58` — Formular-Schema
  `betragCents ≥ 1`.
- Backend erlaubt bereits `≥ 0` (`command_handler.go:34`,
  `fiskalische_projektion.go:89-100`).

### What to build

Die UI-seitigen Validierungsgrenzen für den Anfangsbestand werden von `≥ 1` auf
`≥ 0` gesenkt (Formular-Schema und Backend-Klassen-Schema), passend zur
Backend-Grenze. Ein Verein ohne Wechselgeld kann eine Kassensitzung mit 0 €
eröffnen; es entsteht dann korrekt kein signaturpflichtiger Eröffnungsvorgang.
Die Fehlermeldung („Bitte einen Anfangsbestand eingeben") wird angepasst, sodass
0 eine gültige, bewusste Eingabe ist und nicht als Leerfeld gewertet wird.

### Acceptance criteria

- [x] `KassensitzungEroeffnenSchema` und das Formular-Schema akzeptieren
      `betragCents = 0`, lehnen negative Werte weiter ab.
- [x] Eine Eröffnung mit 0 € erzeugt genau eine offene Kassensitzung und
      **keinen** TSE-Signaturauftrag (bestehendes Backend-Verhalten, per Test
      abgesichert falls noch nicht vorhanden).
- [x] Leeres Feld bleibt ein Validierungsfehler; 0 ist gültig.
- [x] Vitest deckt 0 (gültig) und leer/negativ (ungültig) ab.

---

## Phase 3: Storno-Arten unterscheiden + Warenrücknahme-Beleg (B3, R6)

### Context

- `backend/domain/kasse/stornierung.go` — `Stornierung`-Struct (ohne
  Diskriminator); `tisch_session_events.go:338-404` — `buildStornierungFromEvent`
  (Warenrücknahme) vs. `buildKorrekturFromEvent` (Korrektur).
- `backend/domain/kasse/historie.go:55-70` — einheitliches Mapping.
- `backend/api/kasse/tischgeschaeft/http/query_handler.go:184-208` — Storno-DTO.
- `backend/api/druck/beleg/application/kassenbeleg_command.go:261-296` — fertiger
  Tisch-Warenrücknahme-Beleg-Pfad (`{tischId, stornierungId}`).
- `frontend/src/service/table/Stornierung.ts` — Response-Schema (kommentar
  min 3, R6); `TischBackend.ts:69-82` — `belegDrucken` (nur Zahlung).
- `frontend/src/service/components/table/TischHistorie.tsx:190-220` — Storno-
  Rendering; `…/direktverkauf/DirektverkaufHistorie.tsx:115-165` — Button-Vorlage.

### What to build

Ende-zu-Ende sichtbare Unterscheidung der beiden Storno-Arten plus Belegdruck
für die kassenwirksame Warenrücknahme:

- **Domain:** Die `Stornierung` bekommt ein abgeleitetes Feld `barRueckgabe`
  (bool). `buildStornierungFromEvent` setzt es `true`, `buildKorrekturFromEvent`
  `false`. Kein Event-Format ändert sich (Ableitung aus dem Event-Typ).
- **HTTP:** Das Tisch-Historie-DTO trägt `barRueckgabe`. Der Storno-Beleg wird
  über den bereits vorhandenen Backend-Pfad ausgelöst; das Frontend bekommt eine
  Backend-Klassen-Methode `stornobelegDrucken(tischId, stornierungId)` (Body
  `{tischId, stornierungId}`), analog zur Zahlungs-Variante.
- **Frontend:** Die Historie kennzeichnet Warenrücknahme (Bargeld zurück) vs.
  geldneutrale Korrektur sichtbar (Label/Icon). Nur bei der Warenrücknahme
  erscheint der „Stornobeleg drucken"-Button (mit Nachfass-Logik aus
  `beleg.ts`), gespiegelt am Direktverkauf-Muster.
- **R6:** `StornierungSchema` (Response) lockert `kommentar` auf `max(100)` ohne
  `min`, weil geldneutrale Korrekturen mit leerem Kommentar gültig sind
  (die Eingabepflicht `min 3` bleibt in `StornierungErteilenSchema`).
- **Docs:** Handbuch/Language-Notiz zur „einheitlich Storno"-Darstellung an die
  neue sichtbare Unterscheidung angleichen.

### Acceptance criteria

- [x] `Stornierung.barRueckgabe` ist `true` für `stornierung-erteilt:v1`,
      `false` für `bestellung-korrigiert:v1`; Domain-Test deckt beide ab.
- [x] Das Tisch-Historie-Response trägt `barRueckgabe`; Frontend-Schema
      akzeptiert es und ein leerer Kommentar bricht die Anzeige nicht mehr.
- [x] Die Historie unterscheidet Warenrücknahme und Korrektur sichtbar.
- [x] Der Beleg-Button erscheint **nur** bei Warenrücknahmen und löst über
      `{tischId, stornierungId}` einen Stornobeleg aus (negativer Betrag,
      Referenz auf den Zahlungsbeleg); bei ausstehender TSE greift dieselbe
      Nachfass-Logik wie beim Zahlungsbeleg.
- [x] Backend- und Frontend-Tests decken den neuen Beleg-Pfad und die
      Arten-Unterscheidung ab; `make verify` grün.

---

## Phase 4: Login-Härtung (B4, R4, OTP-Atomizität)

### Context

- `backend/api/middleware/middleware.go:139-146` — `clientIP` ohne Port-Strip
  (B4); `:85-131` — `RateLimitMiddleware` (In-Memory-Muster).
- `backend/api/auth/application/command.go:23-61` — `GenerateJWTToken` (kein
  kontobezogener Schutz).
- `backend/domain/user/user.go:182-217` — `SetPassword` (nicht-atomarer Zähler);
  `backend/api/auth/application/command.go:63-112` — `SetNewPassword` (read-
  modify-write über zwei Repo-Aufrufe).
- `backend/repository/kassenjournal_repo/repo.go:146-179` — Callback-in-TX-Muster
  als Vorlage.

### What to build

Drei zusammenhängende Sicherheitsfixes:

- **B4:** `clientIP` wendet `net.SplitHostPort` auf den `RemoteAddr`-Fallback an,
  sodass der Rate-Limiter-Key die IP ohne Port ist und das Limit ohne
  vertrauenswürdigen Proxy tatsächlich greift (kein neuer Limiter je
  Verbindung, keine unbegrenzt wachsende Map).
- **R4 (Soft-Throttle):** Eine In-Memory-`LoginThrottle`-Komponente (Mutex-
  geschützte Map `username → Fehlversuche + cooldownBis`, auto-ablaufend,
  exponentieller Backoff) wird als Interface in die Auth-`Command` injiziert.
  `GenerateJWTToken` fragt vor der Passwortprüfung `Allow(username)`, verbucht
  Fehlversuche und setzt bei Erfolg zurück. Der HTTP-Handler mappt eine
  gedrosselte Anmeldung auf `429`/Konflikt mit klarer Meldung. Keine dauerhafte
  Sperre.
- **OTP-Atomizität:** Der Set-Password-Pfad läuft als eine Transaktion mit
  Zeilensperre (`SELECT … FOR UPDATE` auf die User-Zeile), die
  `user.SetPassword` als Callback ausführt und das Ergebnis im selben Commit
  persistiert — konkurrierende Versuche für denselben Benutzer werden
  serialisiert, der Fehlversuchszähler kann nicht mehr unterzählen.

### Acceptance criteria

- [x] Ohne `X-Forwarded-For` ist der Limiter-Key die reine IP; ein Test mit
      wechselnden Ports desselben Clients teilt sich einen Limiter.
- [x] Nach mehreren Fehlanmeldungen für ein Konto wird der nächste Versuch
      kurz gedrosselt (auto-ablaufend); ein erfolgreicher Login setzt den
      Zähler zurück; ein anderes Konto ist nie betroffen.
- [x] Der Login-Handler antwortet bei aktiver Drosselung mit einem klaren
      Status/Code (nicht `invalid_credentials`).
- [x] Nebenläufige Set-Password-Versuche für denselben Benutzer zählen jeden
      Fehlversuch genau einmal (transaktionaler Test).
- [x] `make verify` grün.

---

## Phase 5: Event- und Abschluss-Korrektheit (R1, R2)

### Context

- `backend/domain/kasse/tagesabschluss_summen.go:30-73` —
  `BerechneAbschlussSummen` überspringt unparsebare Events still (R1);
  `backend/api/kasse/kassenfuehrung/application/command.go:366` — Aufrufer im
  Kassenabschluss.
- `backend/domain/kasse/tisch_session.go:207-278` — `accumulatePositionen`,
  `reduceByPosition`, `reduceByPositionStrict`, `tagBesteller` mutieren das
  Backing-Array des Eingabe-States (R2).

### What to build

Zwei Korrektheits-Härtungen im Event-Kern:

- **R1:** `BerechneAbschlussSummen` gibt `(AbschlussSummen, error)` zurück und
  meldet ein nicht-parsebares Event als Fehler statt es still zu überspringen.
  `KasseAbschliessen` propagiert den Fehler und bricht den Abschluss ab, statt
  einen falschen Z-Bon zu erzeugen. Konsistent mit der sonst strengen
  Fehlerkultur (`reduceByPositionStrict`, unbekannter Typ in der fiskalischen
  Projektion). Praktisch nur bei einem korrupten Store erreichbar — dort ist
  der Abbruch das korrekte Verhalten.
- **R2:** Die `ApplyEvent`-Helfer arbeiten nicht mehr auf dem Backing-Array des
  Eingabe-States: Positionslisten werden vor der Mutation geklont
  (bzw. `tagBesteller` liefert eine Kopie). Die Funktionen werden wert-sicher,
  wie ihre Signatur suggeriert; ein versehentliches Wiederverwenden des
  Eingabe-States kann keine stille Verfälschung mehr auslösen.

### Acceptance criteria

- [x] `BerechneAbschlussSummen` liefert bei einem unparsebaren Event einen
      Fehler; `KasseAbschliessen` gibt ihn nach oben und schreibt kein
      Tagesabschluss-Event.
- [x] Gültige Sitzungen liefern unveränderte Summen (bestehende Tests grün).
- [x] Ein Test weist nach, dass `ApplyEvent` den übergebenen State (inkl.
      seiner Slices) nicht mutiert.
- [x] `make verify` grün.

---

## Phase 6: Backend-Hygiene und Vereinfachung (B5, B6, B7, R3, R5, R7)

### Context

- `backend/repository/kassensitzungen_repo/repo.go:99-110` + `mock.go:80` —
  toter `InsertKassensitzung` (B5).
- `backend/domain/kasse/bestellung.go:72` (`fromPositionenEventData`) +
  `backend/domain/kasse/fiskalische_projektion.go:167`
  (`positionenFromEventData`) — identische Funktion (B6).
- `backend/api/kasse/tischgeschaeft/application/command.go` —
  `AusgabeBestaetigen` dupliziert `persistTischEvent` (B7).
- `backend/api/kasse/tischgeschaeft/application/query.go:124-177` —
  `GetMeineTischeState` N+1 (R3).
- `backend/api/druck/beleg/http/command_handler.go:75-…` — Vier-Body-Form-
  Dispatch in der HTTP-Schicht (R5).
- `backend/app/app.go:94-119` — gemischtes Logging (R7).

### What to build

Interne Qualität ohne Änderung des äußeren Verhaltens:

- **B5:** `InsertKassensitzung` (Repo + Mock) entfernen — die Eröffnung läuft
  ausschließlich über `EroeffneKassensitzung`.
- **B6:** Die doppelte Positions-Mapping-Funktion auf eine reduzieren
  (die exportierte in `bestellung.go` behalten, den Duplikat-Aufrufer
  umstellen).
- **B7:** `AusgabeBestaetigen` nutzt `persistTischEvent` statt des inline
  duplizierten Write-plus-Error-Mappings.
- **R3:** `GetMeineTischeState` liest Tischnamen und Sessions der Favoriten
  über eine Batch-Query (eine Abfrage statt N+1); Ergebnis unverändert.
- **R5:** Die Auswahl der vier Beleg-Body-Formen wandert aus dem HTTP-Handler in
  die Application-Schicht (ein typisiertes Kommando statt Pointer-Feld-Dispatch);
  der Handler liest/validiert nur noch und delegiert.
- **R7:** `app.go` nutzt durchgängig zerolog statt `fmt.Println`/`log.Printf`.

### Acceptance criteria

- [ ] Kein toter `InsertKassensitzung` mehr; Build grün.
- [ ] Nur noch eine Positions-Mapping-Funktion; alle Aufrufer nutzen sie.
- [ ] `AusgabeBestaetigen` und `ZahlungKassieren` teilen sich denselben
      Persist-Helfer; Fehler-Mapping identisch.
- [ ] `GetMeineTischeState` liefert dasselbe Ergebnis über eine Batch-Query
      (bestehende Query-Tests grün, N+1 verschwunden).
- [ ] Der Beleg-Handler enthält keine Geschäftslogik zur Body-Form-Auswahl mehr;
      Verhalten (alle vier Formen + Fehlercodes) unverändert per Test belegt.
- [ ] `app.go` loggt ausschließlich über zerolog.
- [ ] `make verify` grün.
