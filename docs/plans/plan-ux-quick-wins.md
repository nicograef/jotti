# Plan: UX-Quick-Wins aus dem UX-Review (Juli 2026)

> Source PRD: docs/prds/prd-ux-quick-wins.md

## Goal

Die dreizehn User Stories des PRDs als Frontend-only-Feinschliff umsetzen: Vorgangs-State überlebt Tab-Wechsel, einheitliches ErfolgsPop-Feedback für alle Kassenjournal-Buchungen, Euro-Eingabe ohne Debounce-Überraschung, aktiver Anmelden-Button, AA-konforme Lösch-Bestätigungen, Copy-/Icon-/Farb-Konsistenz, sichtbarer Modus-Wechsel, Aufrunden-Chips und zwei kleine Flow-Verbesserungen. Keine Backend-, API- oder Schema-Änderung.

## Architectural decisions

Durable Entscheidungen, die für alle Phasen gelten:

- **Frontend-only**: Alle Änderungen unter `frontend/src/` und `e2e/`. Die Payloads von `zahlungKassieren` und `direktverkaufTaetigen` bleiben unverändert (ADR 08: Erhalten/Zielbetrag gehen nie an die API).
- **State-Heben (A1)**: Die beiden `useMengen`-Instanzen wandern von den Tab-Inhalten in `TablePage` — Bestell-Korb `Record<number, number>` (Schlüssel: Produkt-Variante-ID), Kassieren-Auswahl `Record<string, number>` (Schlüssel: Position-ID, gedeckelt auf unbezahlte Menge). Die Rückgabeobjekte (`mengen`, `add`, `remove`, `reset`, `setAll`) werden als Props an `Bestellung` und `Zahlung` durchgereicht. Kein `forceMount`; die `animate-fade-up`-Tab-Animation bleibt.
- **ErfolgsPop-Regel (A2)**: Alles, was das Kassenjournal verändert, bestätigt mit dem `ErfolgsPop`; der nachgelagerte Refetch läuft beim Schließen des Pops. Das gilt neben Tischseiten-Stornierung und -Umbuchung auch für den Direktverkauf-Storno (schließt heute ebenso kommentarlos).
- **Neuer Button-Variant `destructive-solid`** in `frontend/src/components/ui/button.tsx` plus neues Token-Paarstück `--destructive-solid-foreground` in `frontend/src/index.css` (Light: Weiß auf der dunklen Light-Destructive-Fläche; Dark: dunkles Rot um red-950 auf der hellen Dark-Destructive-Fläche), analog zum Paar `--warn`/`--warn-foreground`.
- **Neue Badge-Variante `warn`** (Amber-Soft-Tint) in `frontend/src/components/ui/badge.tsx`, nach dem Muster der bestehenden `destructive`-Soft-Tint-Variante.
- **Format-Helfer** in `frontend/src/lib/utils.ts`: neu `formatEuroMitVorzeichen(cents)` („+12,50 €" bei positiv, „0,00 €" ohne Vorzeichen, Minus liefert die bestehende Formatierung); `formatAlleAuswaehlenLabel` bekommt einen Varianten-Parameter (`'alle' | 'meine'`), damit Kassieren „Meine …" und die Umbuchung weiter „Alle …" sagt.
- **Neue Chip-Komponente `AufrundenChips`** unter `frontend/src/service/components/table/` mit purer Ableitungsfunktion `aufrundenVorschlaege(gesamtCents): number[]` (in `drawerUtils.ts`), gemeinsam genutzt von `ZahlungAbschluss` und `DirektverkaufAbschluss`.
- **Segmented Control als Navigation**: In `ServiceLayout` ersetzen zwei `Link`-Elemente in Tabs-Optik (`tabsListVariants`-Default-Look, Höhe mindestens 44 px) den statischen Titel; der aktive Zustand kommt aus der Route (`useMatch`), kein Radix-Tabs-Root. Die Routen `/service/tische` und `/service/direktverkauf` und die Persistenz über die Route-Loader in `frontend/src/routes.ts` bleiben unverändert.
- **Drawer-Sortierung**: `TischAuswahlDrawer` sortiert ausschließlich nach Tischname via `localeCompare(a.name, b.name, 'de', { numeric: true })` („Tisch 2" vor „Tisch 10").
- **Commit-Disziplin**: Jede Phase ist ein eigener commitbarer Schnitt; Phase 6 (Copy & Icons) ist bewusst ein gebündelter Sweep-Commit ohne Verhaltensänderung. Nach jeder Phase `make check`; E2E-relevante Phasen zusätzlich die betroffenen Specs.

## Inventory

Service (Tischseite und Flows):

- `frontend/src/service/TablePage.tsx — TablePage, StatusBadgeInhalt` — Tabs (order/payment/history), ErfolgsPop-Wiring (`zeigeErfolg`, `erfolgSchliessen` mit `reload()`), Ladezustand „Tisch ??"/„?", `data-slot="tisch-saldo"`
- `frontend/src/components/ui/tabs.tsx — Tabs, TabsList, TabsTrigger, tabsListVariants` — Radix-Tabs; inaktiver Inhalt wird ausgehängt (Ursache des State-Verlusts)
- `frontend/src/hooks/use-mengen.ts — useMengen` — Mengen-State mit `add`/`remove`/`reset`/`setAll` und optionalem `max`
- `frontend/src/service/components/table/Bestellung.tsx — Bestellung` — hält heute den Bestell-Korb (`useMengen<number>`)
- `frontend/src/service/components/table/Zahlung.tsx — Zahlung` — hält heute die Kassieren-Auswahl (`useMengen<string>`), Auswählen-Button über `meinePositionen`
- `frontend/src/service/components/table/BestellungDrawer.tsx / ZahlungDrawer.tsx / drawerUtils.ts` — mobile Sheets, `toBestellungData`, `selectPositionen`, `calculateZahlungsbetraege`
- `frontend/src/service/components/table/BestellungAbschluss.tsx / ZahlungAbschluss.tsx — warLeer-Muster` — Idempotenz-Schlüssel (`bestellungId`) bzw. Erhalten/Zielbetrag/Hinweistext
- `frontend/src/service/components/table/TischHistorie.tsx — TischHistorie` — Prop-Kette `onStornierungErteilt`/`onBestellungUmgebucht`
- `frontend/src/service/components/table/HistorieStornierungDrawer.tsx / HistorieUmbuchungDrawer.tsx` — Storno schließt kommentarlos, Umbuchung mit `toast.success('Bestellung umgebucht.')`
- `frontend/src/service/components/ErfolgsPop.tsx — ErfolgsPop` — Vollbild-Bestätigung, Auto-Dismiss 1400 ms
- `frontend/src/service/DirektverkaufPage.tsx — DirektverkaufPage` — eigener ErfolgsPop, Storno-Pfad refetcht heute sofort
- `frontend/src/service/components/direktverkauf/Direktverkauf.tsx / DirektverkaufAbschluss.tsx / DirektverkaufHistorie.tsx / DirektverkaufStornoDrawer.tsx` — Direktverkauf-Flow; Abschluss ruft `calculateZahlungsbetraege(total, erhalten, 0)` ohne Zielbetrag; Storno-Drawer schließt kommentarlos
- `frontend/src/service/components/TischAuswahlDrawer.tsx — sortiereTische` — heute Favoriten → Saldo absteigend → Name
- `frontend/src/service/components/MeinTischCard.tsx — statusFarbe, countOffenePositionen` — Punkt-Logik rot/amber/grün
- `frontend/src/service/TableSelectionPage.tsx — TableSelectionPage` — Empty-State-Text mit Moduswechsel-Erklärung
- `frontend/src/service/ServiceLayout.tsx — ServiceLayout` — Header-Titel-Ternary, Backlink auf Tisch-Detail; kein Test vorhanden
- `frontend/src/components/common/UserDropdown.tsx — moduswechselEintrag` — Menü-Zweitweg mit `ArrowRightLeft`
- `frontend/src/lib/arbeitsmodus.ts — getArbeitsmodus, setArbeitsmodus` — Geräte-Persistenz (localStorage)
- `frontend/src/routes.ts — ServiceTischauswahlLoader, ServiceDirektverkaufLoader, ServiceIndexRedirect` — Loader persistieren den Modus

Gemeinsame Bausteine und Admin:

- `frontend/src/components/common/EuroInput.tsx — EuroInput, cleanInput` — 1-Sekunden-Debounce (`debounceRef`, Cleanup-Effect, Timeout-Zweig), onBlur-Normalisierung
- `frontend/src/components/common/LoginForm.tsx / PasswordForm.tsx` — Submit-Button `disabled={loading || !form.formState.isValid}`; keine Komponententests vorhanden
- `frontend/src/components/ui/button.tsx — buttonVariants` — Variants inkl. soft `destructive` und solid `warn`
- `frontend/src/components/ui/badge.tsx — badgeVariants` — keine Amber-Variante
- `frontend/src/components/ui/alert-dialog.tsx — AlertDialogAction` — reicht `variant` bereits an `Button` durch
- `frontend/src/index.css` — Token-Definitionen (`--warn`/`--warn-foreground`, `--destructive`) plus `@theme inline`-Wiring
- Ad-hoc-Lösch-Buttons (`className="bg-destructive text-white hover:bg-destructive/90"` auf `AlertDialogAction`): `frontend/src/admin/products/ProductItem.tsx`, `frontend/src/admin/products/EditVariantDialog.tsx`, `frontend/src/admin/tables/EditTischDialog.tsx`, `frontend/src/admin/users/UserRow.tsx`, `frontend/src/admin/settings/DruckstationConfigPage.tsx — AlleVerwerfenDialog`
- `frontend/src/admin/reporting/LiveReportingSection.tsx` — Refresh-Button „Jetzt"
- `frontend/src/admin/AdminSidebar.tsx` — `LogOut`-Icon sowohl für „Zum Service-Bereich" als auch „Abmelden"
- `frontend/src/admin/kasse/KasseAbschliessenSection.tsx — liveDifferenzCents` — Differenz-Anzeige, rot nur bei Fehlbetrag
- `frontend/src/admin/settings/DruckstationConfigPage.tsx — DruckstationCard, speichereIp` — on-blur-Speichern nur mit Toast
- `frontend/src/admin/finanzamt/EinrichtungSection.tsx — handleCopy` — Kopier-Bestätigung: Check-Icon für 2 Sekunden (Vorbild für die Drucker-IP)
- `frontend/src/lib/utils.ts — formatCents, formatEuro, formatAlleAuswaehlenLabel` — Format-Helfer
- `frontend/src/components/ui/skeleton.tsx — Skeleton` — vorhandener Baustein für Ladezustände

E2E:

- `e2e/tests/admin-kontrast-axe.spec.ts — pruefeKontrast` — axe color-contrast auf drei Admin-Seiten, Light+Dark, öffnet heute keine Dialoge
- `e2e/support/anmelden.ts — anmelden`, `e2e/support/seed.ts — resetAndSeed`, `e2e/support/servicekraft.ts — oeffneTisch, warteAufTischGeladen, abmelden` — Helfer zum Erweitern
- Copy-sensible Specs (Assertions matchen sichtbare Texte): `stornierung-serviceleitung.mobile.spec.ts`, `umbuchung.mobile.spec.ts`, `direktverkauf-storno.mobile.spec.ts`, `admin-live-reporting.spec.ts`, `admin-produkte-verwalten.spec.ts`, `bestellen-kassieren.spec.ts`, `tischservice-teilzahlung.mobile.spec.ts`

## Resolved decisions

- **Direktverkauf-Storno fällt unter die ErfolgsPop-Regel** (Phase 2): Das PRD formuliert die Regel „Alles, was das Kassenjournal verändert, bestätigt mit dem ErfolgsPop"; der `DirektverkaufStornoDrawer` schließt heute genauso kommentarlos wie der Tischseiten-Storno, und die `DirektverkaufPage` hat den Pop-Mechanismus bereits. Er wird mit verdrahtet (Text „Stornierung gebucht.", Refetch beim Pop-Schließen).
- **axe-Spec öffnet vier der fünf Lösch-Dialoge** (Produkt, Variante, Tisch, Helfer — alle mit Seed-Daten erreichbar). Der fünfte (Druckaufträge-„Alle verwerfen") erscheint nur bei mehr als einem fehlgeschlagenen Druckauftrag; dafür gibt es keine E2E-Maschinerie, und er nutzt denselben `destructive-solid`-Variant — Abdeckung über die vier geprüften Dialoge.
- **Label-Helfer**: `formatAlleAuswaehlenLabel` bekommt einen Varianten-Parameter statt eines zweiten Helfers; Singular der Meine-Variante: „Meine Position auswählen · X €". Der Toggle-Zustand „Auswahl aufheben" bleibt unverändert.
- **Amber-Soft-Tint als Badge-Variante mit festen Amber-Klassen** (z. B. `bg-amber-500/15 text-amber-700 dark:bg-amber-500/20 dark:text-amber-400`): Der Status-Punkt in `MeinTischCard` nutzt bereits feste Amber-Klassen; ein neues Soft-Amber-Token-Paar nur für einen Badge wäre Over-Engineering. Zentralisiert als Variante in `badge.tsx`, nicht ad hoc an der Call-Site.
- **Neutraler Status-Punkt** („nur fremde offene Positionen"): `bg-muted-foreground` (Token), grün bleibt `bg-green-600`, amber bleibt `bg-amber-500`.
- **Vorzeichen-Helfer**: „+" nur bei Werten > 0; 0 bleibt „0,00 €"; das Minus negativer Beträge liefert die bestehende `formatCents`-Formatierung. Alle Anzeigen der Kassensturz-Differenz nutzen den Helfer; die Farblogik (rot nur bei Fehlbetrag) bleibt.
- **Chips-Ableitung liefert immer genau zwei Vorschläge**: das kleinste 1-€-Vielfache echt über dem Gesamtbetrag und das kleinste 5-€-Vielfache echt über dem Gesamtbetrag; fallen beide zusammen, ersetzt das übernächste 5-€-Vielfache den Doppelgänger. Beispiele: 12,30 € → [13 €, 15 €]; 13,00 € → [14 €, 15 €]; 4,50 € → [5 €, 10 €]. Das deckt „zwei bis drei glatte Beträge, Duplikate entfernt, bei glattem Betrag die nächsthöheren" aus dem PRD deterministisch ab.
- **Kein Verhaltens-Reset beim Tab-Wechsel, aber pro Tisch**: `TablePage` bleibt bei einem `:tischId`-Wechsel gemountet (react-router ändert nur den Param). Der gehobene Mengen-State wird deshalb pro Tisch zurückgesetzt (z. B. `key={tischId}` auf dem State-tragenden Teilbaum oder Reset-Effekt auf `tischId`).

## Open questions / Risks

- **Copy-Kopplung der E2E-Specs**: Label-Änderungen (Phasen 2, 6, 8) ziehen Playwright-Assertions mit. Vor jedem Phasen-Abschluss nach den alten Texten greppen (`Jetzt`, `Umbenennen`, `Alle .* Positionen`, `Bestellung umgebucht`, `Tisch ??`, Header-Titel).
- **Seed-Abdeckung für den Varianten-Lösch-Dialog**: Der axe-Spec braucht ein Produkt mit Variante. Falls der Seed keine hergibt, legt der Spec sie sich selbst über die Admin-UI an (Muster: `admin-produkte-verwalten.spec.ts`).
- **`warteAufTischGeladen`** wartet auf `[data-slot="tisch-saldo"]` mit Betrags-Pattern — beim Skeleton-Umbau (Phase 6) muss der geladene Saldo diesen Slot behalten.

---

## Phase 1: Vorgangs-State überlebt Tab-Wechsel (A1)

**User stories**: 1

### Context

- `frontend/src/service/TablePage.tsx — TablePage` — künftiger Ort beider `useMengen`-Instanzen; hat den Tisch-State (Positionen) für die Kassieren-Max-Funktion
- `frontend/src/hooks/use-mengen.ts — useMengen` — der zu hebende State-Hook
- `frontend/src/service/components/table/Bestellung.tsx — Bestellung` / `frontend/src/service/components/table/Zahlung.tsx — Zahlung` — geben ihre lokalen Hook-Aufrufe ab und nehmen die Rückgabeobjekte als Props
- `frontend/src/service/components/table/BestellungDrawer.tsx / ZahlungDrawer.tsx` — bekommen `mengen` bereits als Prop, Kette bleibt
- `frontend/src/service/components/table/BestellungAbschluss.tsx / ZahlungAbschluss.tsx — warLeer-Muster` — unverändert lassen (Idempotenz-Schlüssel, Reset der Eingaben)

### What to build

Die beiden Mengen-Auswahlen (Bestell-Korb nach Variante-ID, Kassieren-Auswahl nach Position-ID) werden nach `TablePage` gehoben und als Props durchgereicht, sodass sie das Aus- und Wiedereinhängen der Radix-Tab-Inhalte überleben. Übriger UI-State der Tab-Flächen (z. B. Auf-/Zuklappen der Fremd-Positionen) bleibt lokal und darf weiterhin beim Tab-Wechsel verloren gehen. Das Leeren der Auswahl nach Erfolg (`reset()` in den Erfolgs-Callbacks) und das warLeer-Muster der Abschluss-Komponenten bleiben unverändert. Beim Wechsel auf einen anderen Tisch (`:tischId`-Änderung) startet die Auswahl leer.

### Acceptance criteria

- [x] Bestell-Korb: Mengen wählen, zu Historie wechseln, zurück zu Bestellen — die Mengen sind unverändert da (neuer Test in `TablePage.test.tsx` oder `Bestellung.test.tsx`)
- [x] Kassieren-Auswahl: analog über einen Tab-Wechsel hinweg erhalten (neuer Test)
- [x] Nach erfolgreichem Bestellen bzw. Kassieren ist die jeweilige Auswahl weiterhin geleert (bestehende Tests bleiben grün, ADR-08-Querschnittskriterium)
- [x] Bei Wechsel des Tisches (`:tischId`) startet die Auswahl leer
- [x] Kein `forceMount`; `animate-fade-up` läuft weiterhin bei Tab-Aktivierung
- [x] Sheet (< 1024 px) und Abschluss-Spalte (≥ 1024 px) verhalten sich identisch (bestehende Drawer-/Abschluss-Tests angepasst, nicht gelöscht)
- [x] `make check` grün

---

## Phase 2: ErfolgsPop für Stornierung und Umbuchung (A2)

**User stories**: 2

### Context

- `frontend/src/service/TablePage.tsx — zeigeErfolg, erfolgSchliessen` — Pop-Mechanismus mit Refetch beim Schließen; der Kommentar „Storno/Umbuchung lädt sofort" wird hinfällig
- `frontend/src/service/components/table/TischHistorie.tsx — TischHistorie` — Prop-Kette zu beiden Drawern
- `frontend/src/service/components/table/HistorieStornierungDrawer.tsx / HistorieUmbuchungDrawer.tsx` — Erfolgs-Callbacks; der Umbuchungs-Toast entfällt
- `frontend/src/service/DirektverkaufPage.tsx — DirektverkaufPage` / `frontend/src/service/components/direktverkauf/DirektverkaufHistorie.tsx / DirektverkaufStornoDrawer.tsx` — Direktverkauf-Storno-Pfad, heute sofortiger Refetch ohne Feedback

### What to build

Stornierung und Umbuchung der Tischseite sowie der Direktverkauf-Storno werden an den jeweils vorhandenen ErfolgsPop angeschlossen (gleiches Prop-Muster wie Bestellen/Kassieren: Erfolgs-Callback mit Text durch die Historien-Komponente an die Drawer). Texte: „Stornierung gebucht." und „Auf {Zielname} umgebucht." (Zielname = Name des Ziel-Tischs, den der Umbuchungs-Drawer aus seiner Tischauswahl kennt). Der bisher sofortige Refetch dieser Pfade folgt dem Pop-Muster: Refetch erst beim Schließen des Pops. `toast.success('Bestellung umgebucht.')` entfällt.

### Acceptance criteria

- [x] Nach erfolgreicher Stornierung (Tischseite) erscheint der ErfolgsPop mit „Stornierung gebucht."; der Drawer ist zu (Test in `HistorieStornierungDrawer.test.tsx`/`TischHistorie.test.tsx`)
- [x] Nach erfolgreicher Umbuchung erscheint „Auf {Zielname} umgebucht."; kein Toast mehr (Test angepasst)
- [x] Nach erfolgreichem Direktverkauf-Storno erscheint der Pop mit „Stornierung gebucht."
- [x] Refetch (Tisch-State/Historie) läuft erst beim Schließen des Pops, nicht schon beim Buchungserfolg
- [x] Betroffene E2E-Specs (`stornierung-serviceleitung.mobile.spec.ts`, `umbuchung.mobile.spec.ts`, `direktverkauf-storno.mobile.spec.ts`) matchen auf die Pop-Texte statt Toast/kommentarloses Schließen
- [x] `make check` grün

---

## Phase 3: Euro-Eingabe ohne Debounce-Reformat (A3)

**User stories**: 3

### Context

- `frontend/src/components/common/EuroInput.tsx — EuroInput` — `debounceRef`, Unmount-Cleanup-Effect und der `setTimeout`-Reformat-Zweig im onChange
- `frontend/src/components/common/EuroInput.test.tsx` — bestehende Sanitisierungs- und Blur-Tests

### What to build

Das Debounce-Reformat entfällt ersatzlos: Timer-Ref, Cleanup-Effect und Timeout-Zweig werden entfernt. Die Eingabe-Sanitisierung (`cleanInput`) bei jedem Tastendruck und die Normalisierung beim Verlassen des Felds (onBlur) bleiben unverändert. Alle Nutzer des Felds (Kassieren, Direktverkauf, `EuroField`-Wrapper im Admin) profitieren ohne eigene Anpassung.

### Acceptance criteria

- [x] Pause-Fall: „1" tippen, über eine Sekunde warten (Fake-Timer), Feld zeigt weiter „1"; danach „5" tippen ergibt „15"; Blur normalisiert zu „15,00" (neuer Test)
- [x] Bestehende `EuroInput`-Tests (Sanitisierung, Blur-Normalisierung, Punkt-als-Komma, Dezimalstellen-Kappung) bleiben grün
- [x] In `EuroInput.tsx` existiert kein `setTimeout`/Timer-Ref mehr
- [x] `make check` grün

---

## Phase 4: Aktiver Anmelden- und Passwort-Button (A4)

**User stories**: 4

### Context

- `frontend/src/components/common/LoginForm.tsx — LoginForm` — `disabled={loading || !form.formState.isValid}`
- `frontend/src/components/common/PasswordForm.tsx — PasswordForm` — gleiches Muster
- `frontend/src/lib/AuthBackend.ts — LoginSchema, SetPasswordSchema` — Zod-Validierung (unverändert)

### What to build

Beide Submit-Buttons sind nur noch während des Ladens deaktiviert (`disabled={loading}`). Ein Tap auf den Button bei ungültigem Formular löst `handleSubmit` aus; React Hook Form zeigt die Feldfehler, es geht kein Request raus. Doppel-Submits verhindert weiterhin der Ladezustand.

### Acceptance criteria

- [ ] Tap auf „Anmelden" bei leerem/ungültigem Formular zeigt die Feldfehler; der Login-Backend-Aufruf wird nicht ausgelöst (neuer `LoginForm.test.tsx` — bisher existiert keiner)
- [ ] Analog für „Passwort festlegen" (neuer `PasswordForm.test.tsx`)
- [ ] Während des Ladens ist der Button deaktiviert (kein Doppel-Submit)
- [ ] Gültige Eingaben submitten weiterhin erfolgreich (Happy Path im Test)
- [ ] `make check` grün

---

## Phase 5: Button-Variant `destructive-solid` und axe-Dialog-Prüfung (B1)

**User stories**: 5

### Context

- `frontend/src/components/ui/button.tsx — buttonVariants` — neuer Variant neben `warn`
- `frontend/src/index.css` — neues Token `--destructive-solid-foreground` (`:root` und `.dark`) plus `@theme inline`-Wiring
- `frontend/src/components/ui/alert-dialog.tsx — AlertDialogAction` — reicht `variant` bereits durch, Call-Sites brauchen nur den Variant statt `className`
- Die fünf Call-Sites: `frontend/src/admin/products/ProductItem.tsx`, `frontend/src/admin/products/EditVariantDialog.tsx`, `frontend/src/admin/tables/EditTischDialog.tsx`, `frontend/src/admin/users/UserRow.tsx`, `frontend/src/admin/settings/DruckstationConfigPage.tsx — AlleVerwerfenDialog`
- `e2e/tests/admin-kontrast-axe.spec.ts — pruefeKontrast` — zu erweitern um Dialog-Öffnung
- `docs/adrs/04_warn-bestaetigung.md` — das Warn-Muster als Vorbild (solide Fläche trägt ihren Kontrast selbst)

### What to build

Neuer Button-Variant `destructive-solid`: solide destruktive Fläche (`bg-destructive`), Textfarbe über das neue Token `--destructive-solid-foreground` (Light: Weiß — Light-Destructive ist dunkel genug; Dark: dunkles Rot um red-950 auf der aufgehellten Dark-Destructive-Fläche), Hover/Focus analog zum `warn`-Variant. Die fünf Ad-hoc-Call-Sites stellen auf `<AlertDialogAction variant="destructive-solid">` um und verlieren ihre `className`-Override; danach existiert die Kombination `bg-destructive text-white` nicht mehr im Repo. Der axe-Kontrast-Spec öffnet zusätzlich die vier per Seed erreichbaren Lösch-Bestätigungen (Produkt, Variante, Tisch, Helfer) und prüft sie in Light und Dark.

### Acceptance criteria

- [ ] Variant `destructive-solid` existiert mit Token-getragenem Kontrast in Light und Dark (≥ 4,5:1)
- [ ] Alle fünf Call-Sites nutzen den Variant; `grep -r "bg-destructive text-white" frontend/src` liefert nichts mehr
- [ ] `admin-kontrast-axe.spec.ts` öffnet die Lösch-Dialoge Produkt, Variante, Tisch und Helfer und findet in beiden Themes keine color-contrast-Verstöße
- [ ] Bestehende axe-Prüfungen (drei Seiten, beide Themes) bleiben grün
- [ ] `make check` grün

---

## Phase 6: Copy- und Icon-Sweep (B2)

**User stories**: 6, 7, 8

### Context

- `frontend/src/lib/utils.ts — formatAlleAuswaehlenLabel` — bekommt den Varianten-Parameter (`'alle' | 'meine'`)
- `frontend/src/service/components/table/Zahlung.tsx — Zahlung` — Auswählen-Button (wählt nur `meinePositionen`) → Meine-Variante
- `frontend/src/service/components/table/HistorieUmbuchungDrawer.tsx` — behält die Alle-Variante (dort stimmt „Alle")
- `frontend/src/admin/reporting/LiveReportingSection.tsx` — „Jetzt" → „Aktualisieren"
- `frontend/src/admin/products/ProductItem.tsx` — Dropdown „Umbenennen" → „Bearbeiten" (gleiches Label wie der Stift-Tooltip, beide öffnen denselben Dialog)
- `frontend/src/admin/AdminSidebar.tsx` — „Zum Service-Bereich" bekommt `ArrowRightLeft` (konsistent mit `UserDropdown`); `LogOut` bleibt exklusiv beim Abmelden
- `frontend/src/admin/kasse/KasseAbschliessenSection.tsx — liveDifferenzCents` — Differenz-Anzeige nutzt den neuen Vorzeichen-Helfer
- `frontend/src/service/TablePage.tsx — TablePage` — Ladezustand „Tisch ??"/„?" → `Skeleton`-Bausteine (`frontend/src/components/ui/skeleton.tsx`)

### What to build

Ein gebündelter Sweep ohne Verhaltensänderung: ehrliches Kassieren-Label „Meine N Positionen auswählen · X €" (Singular „Meine Position auswählen · X €"), Umbuchung behält „Alle"; Refresh-Button „Aktualisieren"; ein Begriff fürs Bearbeiten im Produkt-Dropdown; eigenes Icon für den Bereichswechsel; neuer Helfer `formatEuroMitVorzeichen` in `frontend/src/lib/utils.ts` für die Kassensturz-Differenz („+12,50 €" bei Überschuss, Farblogik unverändert, angewendet auf alle Kassensturz-Differenz-Anzeigen); Skeleton-Platzhalter statt „Tisch ??"/„?" beim Laden der Tischseite (`data-slot="tisch-saldo"` bleibt am geladenen Saldo, wegen `warteAufTischGeladen`).

### Acceptance criteria

- [ ] Kassieren-Button sagt „Meine N Positionen auswählen · X €" (Test angepasst); Umbuchungs-Drawer sagt weiter „Alle …"
- [ ] Live-Reporting-Refresh heißt „Aktualisieren" (Test in `LiveReportingSection.test.tsx` angepasst)
- [ ] Produkt-Dropdown-Eintrag heißt „Bearbeiten"; in `ProductItem.tsx` kommt „Umbenennen" nicht mehr vor (der Hinweistext der Tische-Verwaltung nutzt das Wort legitim weiter)
- [ ] „Zum Service-Bereich" nutzt `ArrowRightLeft`; `LogOut` erscheint nur noch bei Abmelden-Aktionen
- [ ] Kassensturz-Differenz zeigt bei Überschuss „+…" (Unit-Test für `formatEuroMitVorzeichen`: positiv, negativ, null); rot bleibt nur der Fehlbetrag
- [ ] Tischseite lädt mit Skeletons; die Strings „Tisch ??" und der Saldo-Platzhalter „?" existieren nicht mehr; `warteAufTischGeladen` (E2E) funktioniert unverändert
- [ ] Betroffene E2E-Assertions (grep nach den alten Labels) angepasst; `make check` grün

---

## Phase 7: „Unbezahlt" verlässt das Gefahren-Rot (B3)

**User stories**: 9

### Context

- `frontend/src/components/ui/badge.tsx — badgeVariants` — neue `warn`-Variante (Amber-Soft-Tint)
- `frontend/src/service/TablePage.tsx — StatusBadgeInhalt` — „N unbezahlt" nutzt heute `variant="destructive"`
- `frontend/src/service/components/MeinTischCard.tsx — statusFarbe` — Punkt-Logik rot/amber/grün
- `docs/adrs/04_warn-bestaetigung.md` — Farbsemantik: Rot bleibt Gefahr/Storno/Fehlern vorbehalten

### What to build

Neue Badge-Variante `warn` nach dem Soft-Tint-Muster der `destructive`-Variante, mit festen Amber-Klassen und AA-Kontrast in beiden Themes (z. B. `bg-amber-500/15 text-amber-700 dark:bg-amber-500/20 dark:text-amber-400`). Der „N unbezahlt"-Badge stellt darauf um („wartet auf dich"-Semantik). Der Status-Punkt in „Meine Tische" wird abgestuft: eigene offene Positionen `bg-amber-500`, nur fremde offene `bg-muted-foreground`, alles erledigt `bg-green-600`. Rot (`bg-destructive`) verschwindet aus beiden Stellen und bleibt Storno-Beträgen und Fehlerzuständen vorbehalten.

### Acceptance criteria

- [ ] „N unbezahlt"-Badge nutzt die neue `warn`-Variante, nicht mehr `destructive` (Test angepasst)
- [ ] Status-Punkt: eigene offene → amber, nur fremde offene → neutral, alles erledigt → grün (`MeinTischCard.test.tsx` angepasst)
- [ ] `bg-destructive` kommt in `MeinTischCard` und `StatusBadgeInhalt` nicht mehr vor
- [ ] „Alles bezahlt"-Badge und übrige Badge-Verwendungen unverändert
- [ ] `make check` grün

---

## Phase 8: Segmented Control „Tische | Theke" im Service-Header (C1)

**User stories**: 10

### Context

- `frontend/src/service/ServiceLayout.tsx — ServiceLayout` — Header-Titel-Ternary wird ersetzt; Backlink auf `/service/tische/:tischId` bleibt
- `frontend/src/components/ui/tabs.tsx — tabsListVariants` — Optik-Vorbild (Default-Variante, Pill-Look); kein Radix-Root nötig
- `frontend/src/routes.ts — ServiceTischauswahlLoader, ServiceDirektverkaufLoader` — Persistenz bleibt in den Loadern, keine Routing-Änderung
- `frontend/src/components/common/UserDropdown.tsx — moduswechselEintrag` — bleibt als Zweitweg
- `frontend/src/service/TableSelectionPage.tsx — TableSelectionPage` — Empty-State-Text wird gekürzt

### What to build

Auf den beiden Modus-Startrouten ersetzt eine Segmented Control „Tische | Theke" den statischen Header-Titel: zwei react-router-`Link`s in Tabs-Optik (Styling aus `tabsListVariants`/`TabsTrigger`-Klassen abgeleitet), aktiver Zustand aus der Route (`useMatch`), Tap-Ziele mindestens 44 px hoch, `aria-current` am aktiven Segment. Auf der Tisch-Detailseite bleibt der „‹ Meine Tische"-Backlink statt des Switchers. Der Menüeintrag im Benutzermenü bleibt. Die Arbeitsmodus-Persistenz bleibt unverändert (die Route-Loader setzen sie bereits). Im Empty-State der Tischauswahl entfällt der Satz, der den Moduswechsel übers Benutzermenü erklärt.

### Acceptance criteria

- [ ] Auf `/service/tische` und `/service/direktverkauf` zeigt der Header die Segmented Control; ein Tap navigiert in beide Richtungen (neuer `ServiceLayout.test.tsx` — bisher existiert keiner: Navigation und aktiver Zustand)
- [ ] Der aktive Zustand entspricht der Route (`aria-current` oder äquivalentes sichtbares Merkmal, im Test prüfbar)
- [ ] Tisch-Detailseite zeigt weiterhin den Backlink, keinen Switcher
- [ ] Nach Navigation über die Control ist der Arbeitsmodus persistiert (bestehende Loader-Tests in `routes.test.ts` bleiben grün)
- [ ] Benutzermenü-Eintrag funktioniert unverändert (`UserDropdown.test.ts` grün)
- [ ] Empty-State-Text erwähnt das Benutzermenü nicht mehr; E2E-Specs, die auf die alten Header-Titel matchen, sind angepasst
- [ ] Tap-Ziele der Segmente ≥ 44 px; `make check` grün

---

## Phase 9: Aufrunden-Chips fürs Trinkgeld (C2)

**User stories**: 11

### Context

- `frontend/src/service/components/table/drawerUtils.ts — calculateZahlungsbetraege` — bestehende clientseitige Rechnung; hier entsteht `aufrundenVorschlaege(gesamtCents)`
- `frontend/src/service/components/table/ZahlungAbschluss.tsx — ZahlungAbschluss` — Zielbetrag-Feld plus dreizeiliger Hinweis werden durch Chips ersetzt; Feld bleibt hinter „Anderer …"
- `frontend/src/service/components/direktverkauf/DirektverkaufAbschluss.tsx — DirektverkaufAbschluss` — bekommt Zielbetrag-State und damit erstmals Trinkgeld-Anzeige (heute `calculateZahlungsbetraege(total, erhalten, 0)`)
- `docs/adrs/08_service-split-screen.md` — Sheet/Spalte identisch; Eingabe-State pro logischem Vorgang (warLeer-Reset)

### What to build

Neue gemeinsame Komponente `AufrundenChips` (unter `frontend/src/service/components/table/`): ein Chip „{Betrag} € genau" für den exakten Gesamtbetrag, zwei Aufrunden-Vorschläge aus `aufrundenVorschlaege(gesamtCents)` (Ableitungsregel siehe Resolved decisions), plus „Anderer …", der das bisherige Euro-Feld einblendet. Ein Tap setzt den Zielbetrag; ein erneuter Tap auf den aktiven Chip wählt ab (zurück zu „genau"). Chips mindestens 44 px hoch, Ziffern `tabular-nums`. `ZahlungAbschluss` ersetzt Zielbetrag-Feld und dreizeiligen Hinweistext durch die Chips (das freie Feld erscheint nur nach „Anderer …"); `DirektverkaufAbschluss` bekommt Zielbetrag-State, die Chips und damit Trinkgeld- und Rückgeld-Zeilen wie am Tisch. Der Trinkgeld-Buchungshinweis („wird nicht als Kasseneinnahme gebucht …") bleibt. Der warLeer-Reset umfasst den neuen Zielbetrag-State. Keine Payload-Änderung; Sheet und Spalte verhalten sich identisch (die Komponente ist Variant-agnostisch).

### Acceptance criteria

- [ ] `aufrundenVorschlaege`: Unit-Tests für die Beispiele 12,30 → [13, 15]; 13,00 → [14, 15]; 4,50 → [5, 10] (Dedup-Fall)
- [ ] Chip-Tap setzt den Zielbetrag, Trinkgeld- und Rückgeld-Zeilen rechnen wie bisher clientseitig; erneuter Tap wählt ab (Tests)
- [ ] „Anderer …" blendet das Euro-Feld ein; freie Beträge funktionieren wie bisher
- [ ] Der dreizeilige Zielbetrag-Hinweistext existiert nicht mehr; der Trinkgeld-Buchungshinweis bleibt
- [ ] Direktverkauf zeigt erstmals Trinkgeld bei aufgerundetem Zielbetrag (Test in `DirektverkaufAbschluss.test.tsx`)
- [ ] Kassieren- und Direktverkauf-Payloads unverändert (bestehende Submit-Tests grün, kein neues Feld)
- [ ] warLeer-Reset leert auch den Chip-/Zielbetrag-Zustand beim Start eines neuen Vorgangs
- [ ] Chips ≥ 44 px hoch, Ziffern tabular; `make check` grün

---

## Phase 10: Drawer-Sortierung und Drucker-IP-Inline-Bestätigung (C3)

**User stories**: 12, 13

### Context

- `frontend/src/service/components/TischAuswahlDrawer.tsx — sortiereTische` — heute Favoriten → Saldo absteigend → Name
- `frontend/src/admin/settings/DruckstationConfigPage.tsx — DruckstationCard, speichereIp` — on-blur-Speichern, heute nur Toast
- `frontend/src/admin/finanzamt/EinrichtungSection.tsx — handleCopy` — Vorbild: Check-Icon für 2 Sekunden

### What to build

Der „Alle Tische"-Drawer sortiert durchgehend nach Tischname mit numerischem Vergleich (`localeCompare(…, 'de', { numeric: true })`, „Tisch 2" vor „Tisch 10"); die Favoriten-Gruppierung und die Saldo-Sortierung entfallen im Drawer, Favoriten-Stern und Saldo-Anzeige pro Zeile bleiben, die Hauptseite bleibt unverändert. Kein zweites Suchfeld. Beim on-blur-Speichern der Drucker-IP erscheint zusätzlich zum Toast eine kurze Inline-Bestätigung am Feld (Check-Icon für etwa zwei Sekunden, nach dem Muster der TSE-Kopier-Bestätigung).

### Acceptance criteria

- [ ] Drawer-Sortierung: „Tisch 2" vor „Tisch 10", Favoriten nicht mehr vorgezogen (Test in `TischAuswahlDrawer.test.tsx` angepasst)
- [ ] Favoriten-Stern und Saldo pro Zeile funktionieren unverändert
- [ ] Nach erfolgreichem IP-Speichern erscheint für ~2 Sekunden eine Inline-Bestätigung am Feld; der Toast bleibt (Test)
- [ ] Bei Validierungsfehler der IP erscheint keine Erfolgs-Bestätigung
- [ ] `make check` grün
