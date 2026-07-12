# Plan: Service-Redesign Nacharbeit (Cleanup-Befunde + offene Design-Befunde)

> Source PRD: n/a (Cleanup-Review über fff129c..HEAD und Absprache mit Nico, 2026-07-12)

## Goal

Die 12 Befunde des Cleanup-Reviews nach dem Service-Redesign beheben, den
Umbuchungs-Autotext entdoppeln, die ungelesenen EigeneUebersicht-Felder
vollständig entfernen und die abgesprochenen offenen Design-Befunde umsetzen
(1f, 1j, 1k, Dark-Destructive) bzw. prüfen (1g). Befund 1h (Mengenkorrektur im
Prüf-Drawer) bleibt bewusst offen.

## Architectural decisions

- **Benennung Umbuchungs-Freitext**: An der HTTP-Grenze heißt der optionale
  Freitext durchgängig `benutzerKommentar` — Interface-Parameter, Feld im
  Request-Struct, JSON-Key und Frontend-Request-Schema. Das Event-JSON ist
  eingefroren und bleibt unverändert: `kommentar` = Richtungs-Autotext,
  `benutzerKommentar` = Freitext (beide Felder existieren dort bereits so).
- **Autotext-Wortlaut** (nur neue Events): `Umbuchung von <Tischname>`
  (Zugangs-Seite) und `Umbuchung auf <Tischname>` (Abgangs-Seite). Persistierte
  Events werden nie umgedeutet; Alt-Texte mit "von Tisch"/"auf Tisch" bleiben in
  Historie und DSFinV-K-Export sichtbar und korrekt.
- **EigeneUebersicht** wird auf die vier Statistikfelder reduziert
  (`anzahlBestellungen`, `bestellungenCents`, `anzahlZahlungen`,
  `zahlungenCents`) — Domain-Struct, Application-Query, Response-DTO und
  Zod-Schema. `ComputeOffeneArbeitRollup` bleibt bestehen (wird vom
  Admin-Live-Reporting über `ComputeOffeneArbeitProServicekraft` genutzt).
- **Vorauswahl-Muster vereinheitlicht**: Alle drei Positionslisten-Drawer
  (Kassieren, Storno, Umbuchung) starten mit leerer Auswahl; die Umbuchung
  erhält einen "Alle auswählen"-Button nach dem bestehenden Muster in
  `Zahlung.tsx — alleAuswaehlen()`.
- **Neue geteilte Frontend-Bausteine**: `HistorieRowSkeleton` (geteiltes
  Lade-Skeleton für Tisch- und Direktverkauf-Historie), `HistorieDetail`
  (extrahierter Detail-Drawer-Inhalt der Tisch-Historie), `quelleZeitpunkt()`
  in `drawerUtils.ts` (Gegenstück zu `quelleTitel()`).

## Inventory

- `backend/api/kasse/tischgeschaeft/http/command_handler.go — command`-Interface
  und `bestellungUmbuchenRequest` — Parameter/Feld heißen heute `kommentar`.
- `backend/api/kasse/tischgeschaeft/application/command.go —
  BestellungUmbuchen(), buildUmbuchungKommentar()` — Impl-Parameter
  `benutzerKommentar`; Autotexte `"Umbuchung auf Tisch "` / `"Umbuchung von
  Tisch "`.
- `backend/api/fiskal/dsfinvk/mapper_test.go —
  TestMapUmbuchungGeldneutralMitReferenz,
  TestMapUmbuchungNotizMitBenutzerKommentar` — die
  `Z_SE_BARZAHLUNGEN`-Assertion sitzt im falschen Test.
- `backend/domain/reporting/reporting.go — EigeneUebersicht` — Felder
  `OffeneTische`, `AlleErledigt`.
- `backend/api/reporting/application/query.go — GetEigeneUebersicht()` — lädt
  Tisch-Sessions + alle Tische nur für den Rollup der beiden Felder;
  `TischSessionRepo`/`TischRepo` bleiben als Deps (GetLiveReporting nutzt sie).
- `backend/api/reporting/http/query_handler.go — eigeneUebersichtResponse,
  GetEigeneUebersichtHandler()` — DTO-Felder + Füll-Schleife.
- `docs/language.md` — Umbuchungs-Abschnitt beschreibt den Freitext noch als
  `kommentar`; `benutzerKommentar` fehlt.
- `frontend/src/service/table/Umbuchung.ts — BestellungUmbuchenSchema` —
  Request-Feld `kommentar` (Freitext).
- `frontend/src/service/table/Tisch.ts — EigeneUebersichtSchema,
  OffeneArbeitTischSchema` — Schema-Felder `offeneTische`/`alleErledigt`;
  `OffeneArbeitTischSchema` hat im Service-Scope keinen weiteren Nutzer
  (Admin-Reporting hat eine eigene Kopie in `admin/reporting/types.ts`).
- `frontend/src/service/table/hooks.ts — DEFAULT_EIGENE_UEBERSICHT,
  useEigeneUebersicht()` — Default enthält die beiden Felder.
- `frontend/src/service/components/table/HistorieUmbuchungDrawer.tsx —
  createDefaultMengen()` — roher `useState` + handgeschriebenes
  `onAdd`/`onRemove` statt `useMengen`; Voll-Vorauswahl.
- `frontend/src/hooks/use-mengen.ts — useMengen()` — geteilter Hook mit
  `add`/`remove`/`reset`/`setAll` und optionalem `max`-Cap.
- `frontend/src/service/components/table/Zahlung.tsx — alleAuswaehlen()` —
  Vorbild für den Alle-auswählen-Button (Button-Text mit Anzahl und Summe).
- `frontend/src/service/components/table/drawerUtils.ts — quelleTitel()` —
  Vorbild für `quelleZeitpunkt()`.
- `frontend/src/service/components/table/TischHistorie.tsx` — 68-zeilige
  Inline-IIFE für den Detail-Drawer; `ItemSkeleton`; nie aktive
  `disabled={primaryAction?.loading}`-Bindung an den Umbuchen/Stornieren-Buttons.
- `frontend/src/service/components/table/HistorieStornierungDrawer.tsx` und
  `HistorieUmbuchungDrawer.tsx` — wortgleicher DrawerTitle-Block mit
  `quelle.art === 'bestellung' ? aufgenommenAm : umgebuchtAm`.
- `frontend/src/service/components/table/CommentField.tsx — KommentarField` —
  Pflicht-Hinweis erst nach `touched`; `required` genutzt von
  `HistorieStornierungDrawer.tsx` und `direktverkauf/DirektverkaufStornoDrawer.tsx`.
- `frontend/src/service/components/direktverkauf/DirektverkaufHistorie.tsx —
  HistorieItemSkeleton` — byte-identische Kopie des TischHistorie-Skeletons;
  Zeilen tragen nur den Betrag als Titel, `verkauf.positionen` liegt in der
  Response bereits vor.
- `frontend/src/service/components/direktverkauf/DirektverkaufDrawer.tsx` —
  toter `crypto.randomUUID()`-Initializer für `verkaufId` (Regeneration erfolgt
  immer in `onOpenChange(true)`).
- `frontend/src/service/TableSelectionPage.tsx` — Leere-Suche-Meldung schließt
  mit ASCII-`"` statt typografischem `“`.
- `frontend/src/index.css` — Dark-Token `--destructive: oklch(0.704 0.191 22.216)`.
- `e2e/tests/umbuchung.mobile.spec.ts` — "gesperrt"-Kommentar über der
  `fill`-Zeile statt über den Disabled-Assertions; Regex und Erklär-Kommentar
  zur Autotext-Doppelung; Flow erwartet Voll-Vorauswahl.
- `e2e/tests/stornierung-serviceleitung.mobile.spec.ts` — identische
  Heading-Assertion einzeilig statt wie in umbuchung.mobile.spec.ts umbrochen.
- `backend/domain/kasse/event_json_contract_test.go` — pinnt persistierte
  Event-Formate; Fixtures repräsentieren Alt-Events und bleiben in allen Phasen
  unverändert.

## Resolved decisions

- Autotext: beide Richtungen ändern ("von" und "auf"), nur neue Events.
- EigeneUebersicht: Vollschnitt inkl. Domain-Struct-Feldern und Rollup-Block in
  der Query, nicht nur DTO/Schema.
- Offene Design-Befunde: 1f (Positions-Zusammenfassung), 1j
  (Validierungshinweis) und Dark-Destructive-Kontrast werden umgesetzt.
- 1g (Erklärung negativer Saldo): kein UI-Hinweis. Stattdessen prüfen, ob
  negativer Saldo per Invariante systemisch unmöglich ist; Ergebnis
  dokumentieren, bei erreichbarem Negativ-Saldo Befund an Nico melden — kein
  stiller Fix.
- 1k: Umbuchung startet leer + "Alle auswählen"-Button; kombiniert mit dem
  useMengen-Refactor.
- 1h (Mengenkorrektur im Prüf-Drawer): bleibt bewusst offen, Kandidat fürs
  Folge-Review.
- Naming an der HTTP-Grenze: `benutzerKommentar` durchziehen (statt den
  Impl-Parameter auf `kommentar` zurückzubenennen), damit der Begriff nicht mit
  dem Autotext-`kommentar` kollidiert.
- Der Push des Redesigns ist bereits erfolgt (`origin/main` == HEAD); dieser
  Plan setzt auf dem gepushten Stand auf.

## Open questions / Risks

- Phase 9 kann ergeben, dass ein negativer Saldo doch erreichbar ist — dann
  endet die Phase mit einem Bericht an Nico, nicht mit einem Fix.
- Der Ziel-Wert des Dark-Destructive-Tokens wird visuell abgestimmt
  (Screenshot-Abnahme), der Plan legt keinen festen oklch-Wert fest.
- Phase 5 ändert den E2E-Umbuchungs-Flow (Leer-Start); der Spec-Umbau muss im
  selben Schritt erfolgen, sonst ist die Suite zwischenzeitlich rot.

---

## Phase 1: Backend-Benennung und Test-Hygiene (benutzerKommentar an der HTTP-Grenze)

### Context

- `backend/api/kasse/tischgeschaeft/http/command_handler.go — command`-Interface,
  `bestellungUmbuchenRequest`, `bestellungUmbuchenSchema` — Umbenennungsziel.
- `backend/api/kasse/tischgeschaeft/application/command.go — BestellungUmbuchen()` —
  Impl-Parameter heißt bereits `benutzerKommentar`.
- `frontend/src/service/table/Umbuchung.ts — BestellungUmbuchenSchema` und
  `frontend/src/service/table/TischBackend.ts` — Frontend-Seite des Request-Felds.
- `frontend/src/service/components/table/HistorieUmbuchungDrawer.tsx` — baut den
  Request beim Submit.
- `backend/api/fiskal/dsfinvk/mapper_test.go` — verrutschte Assertion.
- `docs/language.md` — Umbuchungs-Abschnitt und Begriffstabelle.

### What to build

Den Freitext-Parameter der Umbuchung an der HTTP-Grenze auf `benutzerKommentar`
vereinheitlichen: Interface-Parameter im `command`-Interface, Feld und JSON-Key
im Request-Struct samt zog-Schema, und spiegelbildlich das Frontend-Request-Feld
in `BestellungUmbuchenSchema` plus dessen Erzeugung im Umbuchungs-Drawer.
Event-Formate und Response-DTOs bleiben unangetastet. Zusätzlich die
`Z_SE_BARZAHLUNGEN`-Assertion aus `TestMapUmbuchungNotizMitBenutzerKommentar`
zurück ans Ende von `TestMapUmbuchungGeldneutralMitReferenz` verschieben und
`docs/language.md` ergänzen: Bei der Umbuchung ist `Kommentar` der
Richtungs-Autotext und `BenutzerKommentar` der optionale Freitext.

### Acceptance criteria

- [x] Interface, Request-Struct, JSON-Key und Frontend-Schema nennen den
      Freitext einheitlich `benutzerKommentar`; eine Umbuchung mit Kommentar
      funktioniert Ende-zu-Ende.
- [x] `event_json_contract_test.go` ist unverändert und grün (Event-Format
      nicht berührt).
- [x] Die Geldneutralitäts-Prüfung (`Z_SE_BARZAHLUNGEN`) liegt wieder in
      `TestMapUmbuchungGeldneutralMitReferenz`; der Notiz-Test endet mit den
      BON_NOTIZ-Assertions.
- [x] `docs/language.md` beschreibt beide Umbuchungs-Kommentarbegriffe.
- [x] `make check` grün.

---

## Phase 2: Verhaltensneutrale Frontend- und E2E-Cleanups

### Context

- `frontend/src/service/TableSelectionPage.tsx` — Anführungszeichen der
  Leere-Suche-Meldung.
- `frontend/src/service/components/table/TischHistorie.tsx` — Detail-IIFE,
  `ItemSkeleton`, tote `disabled`-Bindungen.
- `frontend/src/service/components/direktverkauf/DirektverkaufHistorie.tsx —
  HistorieItemSkeleton` — Skeleton-Kopie.
- `frontend/src/service/components/table/drawerUtils.ts — quelleTitel()` —
  Vorbild für die neue Zeitpunkt-Helper-Funktion.
- `frontend/src/service/components/table/HistorieStornierungDrawer.tsx` /
  `HistorieUmbuchungDrawer.tsx` — duplizierter DrawerTitle-Block.
- `frontend/src/service/components/direktverkauf/DirektverkaufDrawer.tsx` —
  `verkaufId`-Initializer.
- `e2e/tests/umbuchung.mobile.spec.ts`, `e2e/tests/stornierung-serviceleitung.mobile.spec.ts`
  — Kommentar-Platzierung und Umbruch-Konsistenz.

### What to build

Sieben verhaltensneutrale Cleanups in einem Schritt: (1) schließendes
Anführungszeichen der Leere-Suche-Meldung auf `“` korrigieren; (2) die
Detail-Drawer-IIFE der Tisch-Historie in eine Komponente `HistorieDetail`
extrahieren (erhält das Detail, den Tisch, das Backend und die Setter, leitet
die drei Locals intern ab); (3) die toten `disabled={primaryAction?.loading}`-
Bindungen an den Umbuchen/Stornieren-Buttons entfernen (der Schließen-Button
behält seine); (4) das Lade-Skeleton als geteiltes `HistorieRowSkeleton`
extrahieren und in beiden Historien verwenden — die Zeilen-/Detail-Komponenten
selbst bleiben getrennt; (5) `quelleZeitpunkt()` in `drawerUtils.ts` ergänzen
und den DrawerTitle-Block beider Historie-Drawer darauf umstellen; (6) den
`verkaufId`-Initializer auf `useState('')` reduzieren; (7) in den E2E-Specs den
"gesperrt"-Kommentar über die Disabled-Assertions verschieben und die
Heading-Assertion in stornierung-serviceleitung dreizeilig umbrechen wie in
umbuchung.

### Acceptance criteria

- [x] Kein sichtbares oder funktionales Verhalten ändert sich (reine
      Extraktionen, tote Ausdrücke, Typografie); bestehende Komponenten-Tests
      laufen unverändert grün.
- [x] `TischHistorie.tsx` enthält keine Inline-IIFE mehr; das Skeleton existiert
      genau einmal.
- [x] Beide Historie-Drawer beziehen Titel und Zeitpunkt aus `drawerUtils`.
- [x] `make check` und die beiden angefassten E2E-Specs grün.

---

## Phase 3: Autotext-Wortlaut entdoppeln (beide Richtungen)

### Context

- `backend/api/kasse/tischgeschaeft/application/command.go —
  BestellungUmbuchen()` — die beiden `buildUmbuchungKommentar`-Aufrufe.
- `backend/api/kasse/tischgeschaeft/application/command_test.go` — Assertions
  auf Literal und Präfix.
- `backend/seed/engine_test.go` — Präfix-Assertion auf Seed-Daten.
- `e2e/tests/umbuchung.mobile.spec.ts` — Regex `Umbuchung von Tisch <Name>` und
  der Erklär-Kommentar zur Doppelung.
- `backend/domain/kasse/event_json_contract_test.go` — Fixtures repräsentieren
  persistierte Alt-Events und bleiben unverändert.

### What to build

Die Autotext-Präfixe in `BestellungUmbuchen` von `"Umbuchung von Tisch "` /
`"Umbuchung auf Tisch "` auf `"Umbuchung von "` / `"Umbuchung auf "` ändern.
Neue Events tragen damit `Umbuchung von <Tischname>` bzw.
`Umbuchung auf <Tischname>`; bei Tischen namens "Tisch 3" entsteht keine
Doppelung mehr. Die Assertions in `command_test.go` und `seed/engine_test.go`
sowie Regex und Kommentar im E2E-Umbuchungs-Spec nachziehen. Fixtures, die
Alt-Events darstellen (Contract-Test, DSFinV-K-Mapper-Test,
Domain-/Repo-Testdaten mit frei gewählten Strings), bleiben unverändert —
Alt-Texte bleiben gültig.

### Acceptance criteria

- [x] Eine neue Umbuchung erzeugt die Kommentare `Umbuchung von <Name>` (Zugang)
      und `Umbuchung auf <Name>` (Abgang); die 100-Runen-Kürzung über
      `buildUmbuchungKommentar` bleibt wirksam.
- [x] `event_json_contract_test.go` unverändert grün.
- [x] `command_test.go`, `seed/engine_test.go` und
      `umbuchung.mobile.spec.ts` (inkl. bereinigtem Doppelungs-Kommentar) grün.
- [x] `make check` grün.

---

## Phase 4: EigeneUebersicht-Vollschnitt

### Context

- `backend/domain/reporting/reporting.go — EigeneUebersicht` — Felder
  `OffeneTische`, `AlleErledigt` entfernen.
- `backend/api/reporting/application/query.go — GetEigeneUebersicht()` —
  Sessions-/Tische-Laden und Rollup-Befüllung entfernen; die Funktion schrumpft
  auf Kassensitzungs-Nr + Repo-Aufruf.
- `backend/api/reporting/http/query_handler.go — eigeneUebersichtResponse,
  GetEigeneUebersichtHandler()` — DTO-Felder und Füll-Schleife entfernen.
- `backend/api/reporting/application/query_test.go`,
  `backend/api/reporting/http/query_handler_test.go` — Rollup-bezogene
  Testteile entfernen.
- `frontend/src/service/table/Tisch.ts — EigeneUebersichtSchema,
  OffeneArbeitTischSchema` und `frontend/src/service/table/hooks.ts —
  DEFAULT_EIGENE_UEBERSICHT` — Schema und Default reduzieren;
  `OffeneArbeitTischSchema` entfernen, sofern der finale Grep keinen weiteren
  Service-Nutzer zeigt.
- `backend/domain/kasse/offene_arbeit.go — ComputeOffeneArbeitRollup()` —
  bleibt unverändert (Admin-Live-Reporting).

### What to build

Die vom Frontend nicht mehr gelesenen Felder `offeneTische`/`alleErledigt` auf
allen Ebenen entfernen: Domain-Struct, Query-Befüllung (inkl. der nur dafür
geladenen Tisch-Sessions und Tische), Response-DTO und Zod-Schema samt
Hook-Default. Frontend-Tests und Mocks, die die Felder noch stubben, bereinigen.
`ComputeOffeneArbeitRollup` und seine Tests bleiben, da das Admin-Live-Reporting
sie weiter nutzt; die `TischSessionRepo`-/`TischRepo`-Dependencies der Query
bleiben ebenfalls (von `GetLiveReporting` genutzt).

### Acceptance criteria

- [ ] Die `get-eigene-uebersicht`-Response enthält genau die vier
      Statistikfelder; "Meine Tische" (Stat-Card) funktioniert unverändert.
- [ ] `GetEigeneUebersicht` lädt keine Tisch-Sessions und keine Tische mehr.
- [ ] Kein Vorkommen von `OffeneTische`/`AlleErledigt` bzw.
      `offeneTische`/`alleErledigt` mehr im EigeneUebersicht-Pfad (Domain-Struct,
      Query, DTO, Zod-Schema, Hook, Tests); Admin-Reporting unberührt.
- [ ] `make verify` grün (inkl. Integrationstests).

---

## Phase 5: Umbuchung startet leer + Alle-auswählen (Befund 1k, inkl. useMengen-Refactor)

### Context

- `frontend/src/service/components/table/HistorieUmbuchungDrawer.tsx` — ersetzt
  `useState` + `createDefaultMengen` + `onAdd`/`onRemove` durch `useMengen`.
- `frontend/src/hooks/use-mengen.ts — useMengen()` — `max`-Cap und `setAll`
  existieren bereits.
- `frontend/src/service/components/table/Zahlung.tsx — alleAuswaehlen()` —
  Muster für Button-Verhalten und -Beschriftung (Anzahl + Summe, zweiter Tap
  leert).
- `frontend/src/service/components/table/HistorieUmbuchungDrawer.test.tsx`,
  `e2e/tests/umbuchung.mobile.spec.ts` — Flow-Anpassung an den Leer-Start.

### What to build

Der Umbuchungs-Drawer startet mit leerer Mengen-Auswahl wie Kassieren und
Storno. Die Mengenlogik läuft über `useMengen` mit `max` = umbuchbare Menge der
Position; `createDefaultMengen` wird zur Basis des "Alle auswählen"-Buttons
(via `setAll`), die handgeschriebenen `onAdd`/`onRemove` entfallen. Der Button
folgt dem Zahlung-Muster: zeigt Anzahl und Summe der umbuchbaren Positionen,
ein Tap wählt alles voll aus, ein zweiter Tap leert die Auswahl (`reset`). Der
Ausführen-Button bleibt bei leerer Auswahl deaktiviert (bestehende
`noPositionenSelected`-Logik). Komponenten-Test und E2E-Spec bilden den neuen
Flow ab (erst "Alle auswählen" bzw. gezielte Stepper-Taps, dann Ziel-Tisch,
Kommentar, Ausführen).

### Acceptance criteria

- [ ] Der Umbuchungs-Drawer öffnet mit leerer Auswahl; "Umbuchung ausführen" ist
      deaktiviert, bis mindestens eine Position gewählt ist.
- [ ] Ein Tap auf "Alle auswählen" wählt alle umbuchbaren Positionen voll aus;
      ein zweiter Tap leert die Auswahl.
- [ ] Die Mengenlogik kommt aus `useMengen`; im Drawer existiert keine eigene
      Increment-/Decrement-Implementierung mehr.
- [ ] `HistorieUmbuchungDrawer.test.tsx` und `umbuchung.mobile.spec.ts` bilden
      den Leer-Start ab und sind grün; `make check` grün.

---

## Phase 6: Positions-Zusammenfassung in der Direktverkauf-Historie (Befund 1f)

### Context

- `frontend/src/service/components/direktverkauf/DirektverkaufHistorie.tsx` —
  Zeilen tragen nur Betrag + Zeitpunkt/Name; `verkauf.positionen` liegt in der
  Response bereits vor (wird heute nur im Detail-Drawer als Receipt genutzt).
- `frontend/src/service/components/direktverkauf/DirektverkaufHistorie.test.tsx`
  — bestehende Zeilen-Assertions.

### What to build

Jede Verkaufs-Zeile der Direktverkauf-Historie erhält eine zusätzliche
Unterzeile mit der Positions-Zusammenfassung, clientseitig aus
`verkauf.positionen` abgeleitet: Mengen und Namen kommasepariert
("2× Bier 0,5 l, 1× Brezel"), einzeilig und per CSS (`truncate`) gekürzt.
Reine Anzeige-Änderung; Detail-Drawer, Aktionen und Storno-Zeilen bleiben
unverändert. Ziel des Befunds: der richtige Verkauf ist beim Storno ohne Öffnen
des Details wiederauffindbar.

### Acceptance criteria

- [ ] Jede Verkaufs-Zeile zeigt die Positions-Zusammenfassung als einzeilige,
      CSS-gekürzte Unterzeile.
- [ ] Zeitpunkt-/Name-Zeile, Beträge, Storno-Darstellung und Detail-Drawer sind
      unverändert.
- [ ] `DirektverkaufHistorie.test.tsx` prüft die Zusammenfassung; `make check`
      grün.

---

## Phase 7: Pflichtkommentar-Hinweis von Anfang an sichtbar (Befund 1j)

### Context

- `frontend/src/service/components/table/CommentField.tsx — KommentarField` —
  Hinweis erscheint heute nur bei `required && touched && invalid`.
- `frontend/src/service/components/table/HistorieStornierungDrawer.tsx` und
  `frontend/src/service/components/direktverkauf/DirektverkaufStornoDrawer.tsx`
  — die beiden `required`-Nutzer (Storno, GoBD-Pflichtkommentar).

### What to build

`KommentarField` zeigt bei `required` die Anforderung von Anfang an als
statischen, muted Hilfetext unter dem Feld ("Kommentar ist erforderlich
(mind. 3 Zeichen)."). Erst wenn das Feld berührt wurde und ungültig ist,
wechselt derselbe Text auf die destructive-Farbe — kein roter Fehler vor der
ersten Interaktion, aber die Anforderung ist sofort lesbar. Optionale
Kommentarfelder bleiben ohne Hilfetext. Beide Storno-Drawer erben das Verhalten
ohne eigene Änderung.

### Acceptance criteria

- [ ] Beim Öffnen eines Storno-Drawers ist die Kommentar-Anforderung sichtbar
      (muted), ohne dass das Feld berührt wurde.
- [ ] Nach Berühren mit ungültigem Inhalt wird derselbe Hinweis destructive;
      bei gültigem Inhalt verschwindet die Fehlerfarbe.
- [ ] Optionale Kommentarfelder (Bestellen, Kassieren, Umbuchen, Direktverkauf)
      zeigen weiterhin keinen Hilfetext.
- [ ] Betroffene Komponenten-Tests angepasst; `make check` grün.

---

## Phase 8: Dark-Destructive-Kontrast (G10)

### Context

- `frontend/src/index.css` — Dark-Token `--destructive: oklch(0.704 0.191 22.216)`.
- Sichtbare destructive-Elemente im Service: Stornieren-Button im
  Historie-Detail (`TischHistorie.tsx`), Soft-Destructive-Storno-Buttons und
  Unbezahlt-Badges; Pflichtkommentar-Hinweis aus Phase 7.

### What to build

Den Dark-Mode-Wert von `--destructive` so anheben (Lightness/Chroma), dass
destructive Elemente im Dunkeln wieder klar als Warnsignal lesbar sind —
insbesondere Soft-Destructive-Flächen (rot auf Rot-Anteil) und der
Disabled-Zustand müssen unterscheidbar bleiben. Der Zielwert wird visuell
bestimmt: Screenshot-Abnahme light + dark per Playwright-Chromium (headless)
auf den betroffenen Screens (Tisch-Historie mit Storno-Aktionen,
Storno-Drawer, Direktverkauf-Historie). Light-Mode-Token bleibt unverändert.

### Acceptance criteria

- [ ] Destructive Buttons/Badges sind im Dark Mode deutlich vom
      Umgebungs-Hintergrund und vom Disabled-Zustand unterscheidbar
      (Screenshot-Vergleich vorher/nachher).
- [ ] Light Mode ist pixel-unverändert auf den geprüften Screens.
- [ ] `make check` grün.

---

## Phase 9: Invarianten-Prüfung negativer Saldo (Befund 1g, Analyse ohne Code)

### Context

- `backend/domain/kasse/tisch_session.go` — `SaldoCents`-Projektion
  (+Bestellung, −Zahlung, −Storno, ±Umbuchung).
- `backend/api/kasse/tischgeschaeft/application/command.go —
  ZahlungKassieren(), StornierungErteilen(), BestellungUmbuchen()` — prüfen, ob
  Überzahlung/Überstorno gegen die offenen Positionen validiert wird.
- `docs/plans/` — Kontext: das Storno-/Erstattung-Rework hatte "Saldo nie
  negativ" als Ziel.
- `frontend/src/service/components/TischAuswahlDrawer.tsx` — rote
  Negativ-Färbung als defensive Anzeige.

### What to build

Kein Produktions-Code. Analysieren, ob die Kommando-Validierung einen negativen
`SaldoCents` systemisch ausschließt (jede Zahlung/Stornierung/Umbuchung nur
gegen tatsächlich offene Positionen und Mengen). Ergebnis als kurzer
Analyse-Bericht im Chat: Belegkette (Code-Stellen + vorhandene Domain-Tests).
Gilt die Invariante, ist Befund 1g erledigt (die rote Negativ-Färbung im
Alle-Tische-Drawer bleibt als harmlose Defensive bestehen). Ist ein negativer
Saldo erreichbar, wird der Pfad beschrieben und Nico zur Entscheidung
vorgelegt — kein stiller Fix, keine ungefragte UI-Änderung.

### Acceptance criteria

- [ ] Analyse-Bericht mit Belegkette liegt vor (Chat): Invariante gilt / gilt
      nicht, mit konkreten Code-/Test-Referenzen.
- [ ] Falls die Invariante gilt: 1g im Plan als erledigt markiert, keine
      Code-Änderung.
- [ ] Falls nicht: reproduzierbarer Pfad dokumentiert und Entscheidung von Nico
      eingeholt, bevor irgendetwas geändert wird.
