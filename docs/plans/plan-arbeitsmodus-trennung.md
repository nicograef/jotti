# Plan: Arbeitsmodus-Trennung (Tischservice / Direktverkauf)

> Source PRD: ../prds/prd-arbeitsmodus-trennung.md

## Goal

Der Service-Bereich erhält zwei gleichrangige Arbeitsmodi (Tischservice,
Direktverkauf). Ein Helfer arbeitet immer in genau einem Modus, gemerkt pro
Gerät. Beim Öffnen landet er im zuletzt genutzten Modus; gewechselt wird
ausschließlich über das Benutzermenü. Reiner Frontend-Schnitt, kein
Backend-Anteil.

## Architectural decisions

Durable decisions that apply across all phases:

- **Routen**: Keine neuen Routen. Bestehende Service-Routen bleiben:
  `/service` (Index/Einstieg), `/service/tische` (Tischservice),
  `/service/tische/:tischId` (Tischdetail, gehört zum Tischservice),
  `/service/direktverkauf`. Der Service-Index leitet künftig auf die Route des
  zuletzt genutzten Modus weiter statt fix auf `tische`.
- **Persistenz**: Geräte-Einstellung in `localStorage` (überlebt Logout/Login,
  BYOD). Zwei Werte: Tischservice, Direktverkauf. Default ohne gespeicherte
  Präferenz: Tischservice. Gekapselt in genau einem Frontend-Modul (Lese-/
  Schreib-/Hook-Schnittstelle), Vorlage ist die Theme-Präferenz.
- **„Persist on visit" über Route-Loader**: Der Modus wird beim Besuch einer
  Modus-Route geschrieben (auch Deep-Link/Lesezeichen). Realisiert in benannten
  Loadern analog zu `ServiceTableGuard`, damit Einstieg und Persistenz über
  Routing-Tests prüfbar sind.
- **Benennung (UI)**: Kopfbereich-Titel behält die vertrauten Titel „Meine
  Tische" / „Direktverkauf". Der Wechsel-Eintrag im Menü ist als Aktion
  formuliert (Verb + Wechsel-Icon) und nennt den Workflow-Begriff:
  „Zu Direktverkauf wechseln" / „Zu Tischservice wechseln".
- **Ableitung aus Route, nicht aus Zustand**: Kopfbereich-Titel und Menü-Label
  leiten sich aus der aktuellen Route ab (kein reaktives Mitlesen des
  gespeicherten Werts nötig). Der gespeicherte Wert wird nur im Service-Index
  gelesen und beim Routenbesuch geschrieben.

## Inventory

- `frontend/src/routes.ts:51-152` — Router. Service-Teilbaum mit Index-Redirect
  `{ index: true, loader: () => redirect('tische') }` (Z. 126) und den drei
  Content-Routen (Z. 127-147). `ServiceTableGuard` (Z. 44-49) ist die Vorlage
  für einen benannten, testbaren Loader.
- `frontend/src/routes.test.ts:1-26` — Prior Art für Routing-Tests
  (`ServiceTableGuard`, Redirect-Response-Assertions).
- `frontend/src/components/theme-provider.tsx:25-86` — Vorlage für das
  Präferenz-Modul: `localStorage.getItem/setItem` mit `storageKey`, Default-
  Fallback, schmale Schnittstelle plus Hook (`useTheme`).
- `frontend/src/lib/Auth.test.ts:26,78-92,154-166` — Prior Art für
  localStorage-Persistenztests (set → re-init → wert überlebt).
- `frontend/src/service/ServiceLayout.tsx:1-34` — Kopfbereich. `useMatch` für
  Tischdetail und Direktverkauf (Z. 7-9), Back-Link-Logik (Z. 15-26), bindet
  `UserDropdown` ein (Z. 27). Heute zeigt der Titel immer „Meine Tische".
- `frontend/src/service/TableSelectionPage.tsx:1,21-26,34-50` — bildschirm-
  breiter Direktverkauf-Button (Z. 21-26) mit `ShoppingCart`-Import (Z. 1);
  `EmptyState` „Keine Tische markiert" (Z. 34-50) als Ort für den Hinweistext.
- `frontend/src/service/DirektverkaufPage.tsx:20-22` — `<h1>Direktverkauf</h1>`
  (Z. 22), die separate Seitenüberschrift, die entfällt.
- `frontend/src/components/common/UserDropdown.tsx:15-63` — geteiltes
  Benutzermenü (Admin + Service). `location.pathname.startsWith('/admin')`
  (Z. 34) ist das Muster, um den Wechsel-Eintrag auf Service-Routen zu
  begrenzen. Trigger ist `Button size="icon"` (Z. 28, derzeit < 44 px).
- `frontend/src/admin/AdminLayout.tsx` — nutzt dasselbe `UserDropdown`; der
  Wechsel-Eintrag darf hier nicht erscheinen.
- `docs/language.md:105-118` — Direktverkauf-Eintrag; hier ist „Arbeitsmodus"
  als neuer Oberflächenbegriff aufzunehmen (keine `Verkaufsstelle`-Entität).

## Resolved decisions

- **Benennung**: Header „Meine Tische" / „Direktverkauf"; Menü-Aktion
  „Zu Direktverkauf wechseln" / „Zu Tischservice wechseln" (Verb + ⇄-Icon).
- **Persistenz-Stelle**: benannte Route-Loader (Index liest, Modus-Routen
  schreiben), nicht Effekte in Komponenten.
- **Granularität**: 2 Phasen (testbarer Kern, dann Layout-/UX-Bereinigung).
- **Reaktivität**: Titel/Label aus der Route ableiten, kein reaktiver Store.

## Open questions / Risks

- **Entdeckbarkeit** ist die Hauptkostenstelle (nicht der Code): der Wechsel
  liegt nun hinter einem Icon-Menü. Mitigiert durch den Hinweistext im leeren
  Tischauswahl-Zustand (Phase 2). Falls reale Vereine weiter Schwierigkeiten
  melden, wäre der nächste Schritt ein Org-/Benutzer-Toggle (in der PRD bewusst
  out of scope) — das Arbeitsmodus-Modul ist die einzige Stelle, an der die
  Bezugsquelle der Präferenz dann ausgetauscht würde.

---

## Phase 1: Modus-Kernschleife (Modul + modusabhängiger Einstieg + Menü-Wechsel)

**User stories**: 1, 4, 5, 8, 9, 10, 11, 14, 15

### Context

- `frontend/src/components/theme-provider.tsx:25-86` — Vorlage für die
  localStorage-gekapselte Präferenz (Default-Fallback, Schnittstelle + Hook).
- `frontend/src/routes.ts:44-49,119-149` — `ServiceTableGuard` als Loader-
  Muster; Service-Index-Redirect (Z. 126) wird modusabhängig; Modus-Routen
  bekommen einen Loader, der den Modus persistiert.
- `frontend/src/routes.test.ts:1-26` und `frontend/src/lib/Auth.test.ts:154-166`
  — Prior Art für Routing- bzw. localStorage-Persistenztests.
- `frontend/src/components/common/UserDropdown.tsx:15-63` — Wechsel-Eintrag,
  per Routen-Pfad auf den Service-Bereich begrenzt (Muster Z. 34).

### What to build

Ein neues Arbeitsmodus-Modul, das den Modus pro Gerät persistiert: ohne
gespeicherte Präferenz gilt Tischservice; gesetzte Werte überleben ein
Neu-Initialisieren. Der Einstieg in den Service-Bereich leitet je nach
gespeichertem Modus auf die Tischauswahl bzw. den Direktverkauf weiter. Jeder
Besuch einer Modus-Route (Tischauswahl, Tischdetail, Direktverkauf) schreibt
den zugehörigen Modus, sodass auch Deep-Links und Lesezeichen als „zuletzt
genutzt" zählen (Tischdetail zählt als Tischservice). Das Benutzermenü erhält
innerhalb des Service-Bereichs einen Wechsel-Eintrag, der in den jeweils
anderen Modus wechselt, die Geräte-Präferenz setzt und dorthin navigiert; im
Admin-Bereich erscheint der Eintrag nicht. End-to-end: wechseln → persistieren
→ beim nächsten Öffnen im richtigen Modus landen.

### Acceptance criteria

- [x] Ohne gespeicherte Präferenz liefert das Modul Tischservice.
- [x] Nach Setzen eines Modus liefert das Modul diesen zurück; der Wert
      überlebt ein Neu-Initialisieren (Persistenz, von außen beobachtbar).
- [x] Der Einstieg in den Service-Bereich leitet bei gespeichertem
      Tischservice auf die Tischauswahl, bei Direktverkauf auf den
      Direktverkauf weiter.
- [x] Der Besuch einer Modus-Route aktualisiert den gespeicherten Modus
      (Direktverkauf-Route → Direktverkauf; Tischauswahl und Tischdetail →
      Tischservice).
- [x] Das Benutzermenü zeigt im Service-Bereich „Zu Direktverkauf wechseln"
      bzw. „Zu Tischservice wechseln" (Verb + Wechsel-Icon); der Klick setzt
      die Präferenz und navigiert in den anderen Modus.
- [x] Der Wechsel-Eintrag steht Servicekraft, Serviceleitung und Admin im
      Service-Bereich zur Verfügung und erscheint nicht im Admin-Bereich.
- [x] Keine Backend-, Schema-, API- oder Event-Änderung.
- [x] `make lint` ist grün; neue Tests laufen.

---

## Phase 2: Layout & UX-Feinschliff

**User stories**: 2, 3, 6, 7, 12, 13

### Context

- `frontend/src/service/ServiceLayout.tsx:6-34` — Kopfbereich-Titel und
  Back-Link-Logik; Back-Link künftig nur auf der Tischdetail-Seite.
- `frontend/src/service/TableSelectionPage.tsx:1,21-26,34-50` — Entfernen des
  Direktverkauf-Buttons (+ ungenutzter `ShoppingCart`-Import); `EmptyState`
  erhält den Hinweis auf den Moduswechsel.
- `frontend/src/service/DirektverkaufPage.tsx:20-22` — Entfernen der separaten
  `<h1>`-Überschrift.
- `frontend/src/components/common/UserDropdown.tsx:28` — Touch-Fläche des
  Menü-Triggers auf ≥ 44 px.
- `docs/language.md:105-118` — Eintrag „Arbeitsmodus" ergänzen.

### What to build

Der Kopfbereich des Service-Layouts zeigt den aktiven Modus als Titel („Meine
Tische" im Tischservice, „Direktverkauf" im Direktverkauf). Der Back-Link
erscheint nur noch auf der Tischdetail-Seite (zurück zur Tischauswahl); im
Direktverkauf gibt es keinen Back-Link mehr. Die separate Seitenüberschrift auf
der Direktverkauf-Seite entfällt (keine doppelte Benennung). Der
bildschirmbreite Direktverkauf-Button auf der Tischauswahl-Seite wird ersatzlos
entfernt. Der leere Tischauswahl-Zustand („Keine Tische markiert") nennt
zusätzlich den Moduswechsel im Benutzermenü, damit Theken-Helfer den
Direktverkauf ohne Schulung finden. Der Menü-Trigger erhält eine Touch-Fläche
von mindestens 44 px. „Arbeitsmodus" wird als Oberflächenbegriff in
`docs/language.md` aufgenommen.

### Acceptance criteria

- [x] Der Kopfbereich zeigt im Tischservice „Meine Tische", im Direktverkauf
      „Direktverkauf" als Titel.
- [x] Der Back-Link erscheint nur auf der Tischdetail-Seite; im Direktverkauf
      ist kein Back-Link mehr vorhanden.
- [x] Die Direktverkauf-Seite hat keine separate `<h1>`-Überschrift mehr.
- [x] Auf der Tischauswahl-Seite gibt es keinen Direktverkauf-Button mehr (kein
      toter Import).
- [x] Der leere Tischauswahl-Zustand weist auf den Moduswechsel im Benutzermenü
      hin.
- [x] Der Menü-Trigger im Service-Header hat eine Touch-Fläche von ≥ 44 px.
- [x] `docs/language.md` enthält den Eintrag „Arbeitsmodus" (Oberflächenbegriff,
      keine `Verkaufsstelle`-Entität).
- [ ] Visuell auf mobilem Viewport verifiziert; `make lint` ist grün.
      (`make lint`/Typecheck/Tests grün; visuelle Mobile-Abnahme steht aus.)
