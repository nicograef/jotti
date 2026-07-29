# Plan: Release v0.17.3 — Versions-Handshake und Aufräumarbeiten

> Quell-PRD: n/a (aus `docs/plans/review-v0.17.2.md`, der Commit-Einzelprüfung zum
> v0.17.2-Schnitt und drei gezielten Code-Erkundungen)
> Basis: der v0.17.2-Stand. Dieser Plan setzt voraus, dass
> `docs/plans/plan-v0.17.2-release.md` umgesetzt ist. Dort steht noch genau eine
> Checkbox offen (Zeile 397: PR #101 als „nach dem Fest" markieren) — eine reine
> Verwaltungsaufgabe ohne Code-Anteil, die diesen Plan nicht blockiert.

## Ziel

Zwei Dinge in einem Release, das nach dem Fest entsteht und ohne Zeitdruck getestet
werden kann:

1. **Der Versions-Handshake.** Ein Client, dessen Stand nicht mehr zum Server passt, soll
   das erkennen und sich neu laden — ohne dass jemand es ansagen muss, und **ohne je eine
   Eingabe zu vernichten**. v0.17.2 musste diesen Reload noch von Hand durchsetzen.
2. **Die Aufräumarbeiten**, die für das Fest-Release bewusst zurückgestellt wurden:
   Cleanups, zwei kleine Verbesserungen, und der Langläufer-Schutz der TSE-Einrichtung.

Der Maßstab bleibt der von v0.17.2: keine Regression in einem Pfad, der heute
funktioniert. Anders als dort darf hier neuer Code entstehen — es ist kein
Mid-Fest-Release.

## Architekturentscheidungen

- **Kein neuer Endpunkt, keine neue Route, keine neue Dependency.** Der Handshake nutzt
  den bestehenden `/health`-Endpunkt (`backend/api/health/health.go — HealthCheck.Handler()`,
  registriert in `backend/app/app.go — SetupRoutes()`), der bereits ein `version`-Feld
  liefert und außerhalb der JWT-Kette hängt. Zu beachten: Der Handler pingt bei **jedem**
  Aufruf die Datenbank (`health.go — PingContext`) und antwortet bei Ausfall mit 503. Eine
  regelmäßige Abfrage von jedem Helfer-Handy macht `/health` zum meistgerufenen Endpunkt
  des Systems; das Abfrageintervall ist deshalb eine bewusste Entscheidung (Phase 4), keine
  Nebensache.
- **Die Clientversion ist die echte Build-Version**, eingebrannt zur Bauzeit, analog zur
  Backend-Kette (`backend/Dockerfile` → `-ldflags -X main.version`). Begründung unter
  „Resolved decisions".
- **Vergleich nur zwischen echten Releases.** Aktiv ausschließlich, wenn **beide** Seiten
  dem Muster `v<major>.<minor>.<patch>` entsprechen. `dev` und `dev-<sha>` schalten den
  Vergleich still ab. Dieselbe Diskriminante gibt es bereits — nicht als Makefile-Variable,
  sondern als Shell-Prüfung im Target `prod-up` gegen `JOTTI_VERSION` aus der `.env`
  (`Makefile:173-178`, Regex `^v[0-9]+\.[0-9]+\.[0-9]+([.+-].*)?$`). Dieser Regex lässt
  Vorabversionen wie `v1.2.3-rc1` zu; der Handshake übernimmt ihn unverändert, damit es nur
  eine Definition von „echtes Release" im Repo gibt.
  Ein `MODE`/`DEV`-Check wäre falsch: Vite stempelt jedem `pnpm build` `MODE=production`
  auf, und `docker-compose.e2e.yml` wie `docker-compose.local.yml` bauen aus dem Quellcode
  ohne `VERSION`-Build-Arg — dort meldet der Stack `production`, obwohl `dev` gegen `dev`
  steht.
- **Vorgangs-Register als Modul-Singleton** mit Zähler und Listener-Set, gelesen über
  `useSyncExternalStore`, geschrieben über genau einen Hook. Kein globaler State-Store,
  kein vierter Context-Provider — die Regel „Nur React Hooks + Singletons" aus `AGENTS.md`
  ist wörtlich erfüllt, und es fügt sich in `frontend/src/lib/Auth.ts — AuthSingleton` und
  `frontend/src/lib/Backend.ts — BackendSingleton`.
- **Ein Montagepunkt für alles**: `frontend/src/App.tsx — App()`, neben dem `Toaster`.
  Zwischen `App` und den Bereichen gibt es keine gemeinsame Layout-Ebene, und
  `AdminLayout` wie `ServiceLayout` sind lazy geladen — ein Guard dort läge doppelt in
  zwei Chunks.
- **Reload-Regel**: Bei Versionsabweichung wird sofort neu geladen, **wenn** das
  Vorgangs-Register leer ist. Sonst erscheint ein nicht wegklickbarer Hinweis, und der
  Reload erfolgt automatisch, sobald das Register leer wird.
- **Höchstens ein Reload je Zielversion.** Vor dem Neuladen wird die Serverversion, auf die
  hin geladen wird, in `sessionStorage` vermerkt. Beim Start prüft der Guard diesen Vermerk:
  Passt die eigene Version danach immer noch nicht zur vermerkten Zielversion, wird **nicht
  erneut** neu geladen — dann bleibt es bei einem Hinweis mit eigenem Text und einer
  Schaltfläche zum Neuladen von Hand. Ohne diese Bremse entsteht im Update-Fenster eine
  Reload-Schleife; Begründung unter „Risiken".
- **Reihenfolge ist fachlich erzwungen**: Der Langläufer-Schutz der TSE-Einrichtung muss
  **vor** dem scharfen Handshake liegen. Begründung unter „Risiken".

## Inventar

**Bestehende Versionskette (Server)**
- `.github/workflows/release.yml` — leitet die Version aus dem Tag ab, reicht sie als
  Build-Arg **nur an das Backend-Image** weiter.
- `backend/Dockerfile` — `ARG VERSION=dev`, eingebrannt per `-ldflags`.
- `backend/main.go — version` — Default `"dev"`.
- `backend/app/app.go — SetupRoutes()` — registriert `/health` außerhalb der Routentabelle.
- `backend/api/health/health.go — HealthCheck` — liefert `{status, timestamp, version}`.
- `frontend/src/lib/HealthBackend.ts — getVersion()` — ruft es ab.
- `frontend/src/admin/hooks.ts — useVersion()` — `staleTime: Infinity`, genau ein Abruf je
  Seitenladen. Einziger Konsument: `frontend/src/admin/AdminSidebar.tsx`.

**Frontend-Build**
- `frontend/Dockerfile` — zweistufig, **ohne** `ARG`.
- `frontend/vite.config.ts` — kein `define`, kein `envPrefix`.
- `frontend/vitest.config.ts` — **eigenständige Config, die `vite.config.ts` ersetzt**;
  dupliziert Plugin und Alias bereits von Hand.
- `frontend/tsconfig.app.json` — inkludiert nur `src`; es existiert **keine** `.d.ts`-Datei
  im Repo.
- `frontend/nginx.conf` — `index.html` mit `no-cache`, `/assets/` `immutable`. Kein Service
  Worker, keine Offline-Cache-Schicht — ein Reload holt also verlässlich den neuen Stand.
- `frontend/public/manifest.webmanifest` — jotti **ist** als App installierbar
  (`"display": "standalone"`, verlinkt in `frontend/index.html`, eigene nginx-Regel in
  `nginx.conf:38-41`). Das ist kein Nebenschauplatz, sondern der Fall, den der Handshake am
  stärksten verbessert: In der installierten App gibt es weder Adresszeile noch
  Neu-laden-Pfeil, weshalb `docs/leitfaden/aktualisieren.md` heute verlangt, die App ganz
  wegzuwischen. Ein programmatischer Reload muss dort nachweislich funktionieren.

**Träger für das Vorgangs-Register**
- `frontend/src/hooks/use-mengen.ts — useMengen()` — trägt Bestell-Korb,
  Kassieren-Auswahl, Direktverkauf-Korb sowie die Mengenauswahl der Storno- und
  Umbuchungs-Drawer.
- `frontend/src/hooks/use-action-submit.ts — useActionSubmit()` und
  `frontend/src/hooks/use-form-action-submit.ts — useFormActionSubmit()` — tragen jede
  laufende Buchung, jeden Formular-Submit und den Beleg-Nachfasslauf
  (`frontend/src/service/beleg.ts — belegDruckenMitNachfassen()`). Beide halten ihren
  Zustand in schlichtem `useState`, laufen also **nicht** über react-query.
- **Zwei Submits gehen an diesen Trägern vorbei** und müssen überall dort mitgedacht
  werden, wo der Plan „die drei generischen Träger" sagt:
  `frontend/src/components/common/LoginForm.tsx:26` hält seinen Submit-Zustand von Hand,
  und `frontend/src/admin/reporting/hooks.ts:58 — useDsfinvkExport()` läuft über
  react-query `useMutation`.

**Sperr- und Reload-Vorlagen**
- `frontend/src/components/common/ErrorBoundary.tsx` — die einzige bestehende
  Ganzseiten-Sperre und der einzige `window.location.reload()`-Aufruf im Repo.
- `frontend/src/components/ui/alert-dialog.tsx` — Radix-Primitiv; ohne `AlertDialogCancel`
  und mit gesteuertem `open` ohne `onOpenChange` nicht wegklickbar. Präzedenz für
  gesteuerte, triggerlose Nutzung in `frontend/src/admin/kasse/KasseAbschliessenSection.tsx`.
- `frontend/src/pages/HydrateFallbackPage.tsx` — ganzseitige Wartefläche, passendes Bild
  für die Sekunde zwischen Entscheidung und Reload.

**Zurückgestellte Arbeit im Archiv**
- Tag `archiv/main-vor-v0.17.2` — enthält den Langläufer-Schutz (Commit `9ef41ab6`) und
  die Transporthärtung (`d85cf2ae`).
- `docs/plans/review-v0.17.2.md` — Abschnitte „Cleanup" und „Geringfügig".

## Resolved decisions

- **Handshake und Aufräumarbeiten kommen gemeinsam in v0.17.3**, nicht getrennt.
- **Reload-Strategie**: blockieren statt nur hinweisen; sofortiger Reload nur bei leerem
  Vorgangs-Register. Ein bloßes Banner mit Schaltfläche wurde verworfen — es ist
  ignorierbar, und ein ignorierter Hinweis lässt genau den inkompatiblen Client
  weiterlaufen, den der Handshake abfangen soll.
- **Die Cleanup-Liste aus `review-v0.17.2.md` wird nicht portiert, sondern der Durchlauf
  gegen den v0.17.2-Baum wiederholt.** Die alte Verifikation galt einem anderen Umfeld:
  von den 16 Punkten zielen mehrere auf Code, der mit dem Relay- und Idempotenz-Block
  entfallen ist, und mindestens einer beschreibt eine Form, die im neuen Baum anders
  aussieht. (Das Review-Dokument spricht an drei Stellen von „18 Punkten"; seine Liste
  enthält tatsächlich 16. Für diesen Plan folgenlos, weil der Durchlauf ohnehin neu läuft.)
- **Der TSE-Langläufer-Schutz ist dabei**, obwohl beim Verein die TSE eingerichtet ist. Er
  behebt einen Defekt, der heute schon besteht, und der Handshake verschärft ihn.
- **Die Idempotenz-Maschinerie bleibt zurückgestellt und kommt in diesem Release nicht
  wieder.** Sie liegt im Archiv (`archiv/main-vor-v0.17.2`, Commits `c6d9f840`, `187fbfea`,
  `73aee693` sowie die Nutzlast-Bindung des Schlüssels aus `9ef41ab6`) und macht `vorgangId`
  zum Pflichtfeld auf den buchenden Endpunkten — jedes Handy, das nicht neu geladen hat,
  bucht danach nicht mehr. Genau diesen Nachteil hebt der Handshake auf: Ab v0.17.3 erzwingt
  das System den Reload selbst, ein neues Pflichtfeld ist dann kein Feldrisiko mehr. Das ist
  ein Grund, sie **nach** der Auslieferung von v0.17.3 neu zu bewerten, kein Grund, sie hier
  mitzunehmen — sie löst ein Problem, das im Betrieb nie beobachtet wurde, und
  Produkt-Konservatismus lässt sie draußen, bis eines auftritt.
- **Der Guard bleibt auf der Anmeldeseite aktiv.** Eine frühere Fassung dieses Plans
  schaltete ihn dort ab („dort ist nichts offen, ein Reload wäre nur Lärm"). Die Begründung
  dreht das Argument um: Gerade *weil* dort nichts offen ist, ist die Anmeldeseite der
  billigste Ort für den Reload. Mit abgeschaltetem Guard meldet sich der veraltete Client
  erst an, navigiert weiter, und wird eine Sekunde später mitten im Bereichswechsel neu
  geladen — der Anmeldezustand überlebt das zwar (`frontend/src/lib/Auth.ts:95`, Token im
  `localStorage`), aber das Flackern trifft den Helfer nach dem Login statt davor. Aktiv
  gelassen greift der Reload in aller Regel beim Aufschlagen der Seite, bevor überhaupt
  jemand tippt. Der Restfall — der Reload feuert, während jemand das Passwort eintippt —
  kostet einmaliges Neutippen auf einer Seite, auf der sonst nichts auf dem Spiel steht.
- **Die Clientversion ist die echte Build-Version, nicht die beim ersten Laden gesehene
  Serverversion.** Die billigere Variante — die erste gesehene Serverversion als eigene
  Referenz merken — käme ohne jede Änderung an Build-Kette, Vite-Konfiguration und
  Typdeklaration aus. Sie hat aber ein Loch genau im Deployment-Fenster: Kommt das Backend
  hoch, bevor der Frontend-Container ersetzt ist, lädt ein Client das alte Bundle, fragt
  `/health`, bekommt die **neue** Version und schreibt sie als seine eigene fest. Er läuft
  danach dauerhaft alt gegen neu und wird nie eine Abweichung sehen — der Fehlerfall, den
  der Handshake verhindern soll. Ein Mechanismus, dessen einzige Aufgabe eine
  Kompatibilitätsgarantie ist, darf im Deployment keine Rennbedingung haben.

## Risiken

- **Reload-Schleife im Update-Fenster — der schwerwiegendste Fehlerfall.** Beim
  Selbsthosting-Weg (`make prod-update` → `scripts/prod-update.sh:154,159`) läuft
  `docker compose pull` und `up -d` **ohne vorheriges Stoppen**, also als rollierendes
  Neuanlegen in Abhängigkeitsreihenfolge. `docker-compose.prod.yml:110-112` bindet
  `frontend` per `depends_on` an `backend`, damit wird das Backend **garantiert vor** dem
  Frontend ersetzt. In dem Fenster dazwischen — Sekunden bis Zehnersekunden — antwortet das
  neue Backend auf `/health` mit der neuen Version, während der noch alte
  Frontend-Container über den noch alten Reverse-Proxy weiterhin das alte Bundle
  ausliefert. Ein Client, der darin abfragt, lädt neu, bekommt dieselbe alte `index.html`,
  sieht dieselbe Abweichung und lädt wieder neu: eine enge Schleife auf jedem Helfer-Handy,
  ausgerechnet in der heikelsten Minute eines Updates. Startet der Frontend-Container gar
  nicht, ist die Schleife unbegrenzt.
  Das Kriterium „der Reload löst genau einmal aus" schützt **nicht** davor — es gilt je
  Seitenleben, und jeder Reload beginnt ein neues. Deshalb der `sessionStorage`-Vermerk aus
  den Architekturentscheidungen: höchstens ein Reloadversuch je Zielversion, danach
  degradiert der Mechanismus auf den Hinweis aus Phase 4.
  Der Windows-Weg ist nicht betroffen: `jotti-stop.cmd` fährt alles herunter, und der
  Reverse-Proxy wartet beim Start auf `backend: service_healthy` **und**
  `frontend: service_started` — erreichbar wird der Stack erst, wenn beide neu sind.
- **Der erzwungene Reload ist selbst ein neuer Abbruchgrund.** Ein Reload schließt die
  laufende Verbindung, `net/http` storniert daraufhin `r.Context()` — und der trägt heute
  den kompletten fiskaly-Lebenszyklus der TSE-Einrichtung. Ohne den Langläufer-Schutz
  bleibt bei einem Reload nach dem Setzen der Admin-PIN eine bezahlte, halbfertige LIVE-TSS
  zurück, deren PUK und Admin-PIN nur in der nie ausgelieferten Antwort standen; der zweite
  Anlauf prallt an der Bereits-eingerichtet-Prüfung ab. Deshalb Phase 3 vor Phase 6.
- **Der Schutz deckt kein Deployment ab.** `context.WithoutCancel` überlebt keinen
  Serverneustart, weil `backend/app/app.go` dem Shutdown nur eine begrenzte Gnadenfrist
  gibt. Eine TSE-Ersteinrichtung während eines Updates bleibt gefährlich — das gehört als
  Warnung in den Leitfaden, nicht in Code.
- **Der zentrale Fehler-Toast feuert bei jedem Query-Fehler.** `frontend/src/lib/queryClient.ts`
  wirft ihn im `QueryCache.onError` mit fester ID. Eine dauerpollende Versionsabfrage
  erzeugt bei Netzabbruch periodisch genau diesen Toast. Muss in Phase 4 bewusst behandelt
  werden, sonst verschlechtert der Handshake den Alltag statt ihn zu verbessern.
- **Zähler-Leck im Vorgangs-Register.** An- und Abmeldung müssen im `useEffect`-Cleanup
  laufen. Zwei reale Auslöser: der Tischwechsel, bei dem `frontend/src/service/TablePage.tsx`
  gemountet bleibt und nur Zustand zurücksetzt, und das Umschalten zwischen Drawer- und
  Spaltenlayout, das in `frontend/src/service/components/table/Bestellung.tsx` ganze
  Teilbäume tauscht. Ein geleckter Zähler blockiert den Reload dauerhaft — der Handshake
  wäre dann wirkungslos und niemand würde es merken.
- **`vitest.config.ts` ersetzt `vite.config.ts`, es ergänzt sie nicht.** Eine
  Build-Konstante nur in `vite.config.ts` ist im Test nicht definiert und lässt jeden Test
  scheitern, der die betroffene Datei importiert.
- **Docker-Layer-Cache.** Ein deklariertes, aber im `RUN` nicht konsumiertes Build-Argument
  invalidiert die Build-Schicht nicht — dann liegt ein altes Bundle mit alter Version im
  neuen Image, und der Handshake vergleicht Unsinn. Das ist eine harte Anforderung an die
  Formulierung im `frontend/Dockerfile`, keine Feinheit.
- **Ein Pfad bleibt ungeschützt.** `frontend/src/lib/Backend.ts` setzt bei HTTP 401 hart
  `window.location.href` auf die Login-Seite und vernichtet dabei heute schon einen
  gefüllten Warenkorb ohne Rückfrage. Der Handshake schützt daneben, nicht dort. Kein
  Blocker für dieses Release, aber ein Kandidat für den nächsten Durchgang.

---

## Phase 1: Cleanups gegen den neuen Baum

### Context

- `docs/plans/review-v0.17.2.md` — Abschnitt „Cleanup", 16 Punkte (das Dokument selbst
  behauptet 18), verifiziert gegen einen Baum, den es nicht mehr gibt. Abschnitt
  „Geringfügig": 8 Punkte.
- `/home/nico/.claude/skills/cleanup/` — Referenzdateien des Cleanup-Skills
  (`readability.md`, `readability-de.md`, `code-smells.md`, `principles.md`,
  `architecture.md`).

### What to build

Der Cleanup-Durchlauf wird auf dem v0.17.2-Stand neu ausgeführt, nicht aus der alten Liste
übertragen. Umfang ist der Diff `v0.17.1..v0.17.2` plus die Dateien, die die alte Liste
benannt hat und die es im neuen Baum noch gibt.

Es gelten die Grenzen des Skills unverändert: niemals Verhalten ändern, niemals
umschreiben, nur subtrahieren oder vereinfachen, keine Kommentare oder Abstraktionen
hinzufügen. Jeder Vorschlag wird vor der Anwendung auf Beleg, Verhaltenserhalt und
tatsächlichen Nutzen geprüft; Geschmacksfragen fallen raus.

Erwartbar überleben etwa neun Punkte. Die folgenden sechs sind am v0.17.2-Stand bereits
nachgeprüft und existieren dort unverändert:

- doppelt vergebener SQL-Alias `u` — CTE `ursprung u` (`backend/sqlc/queries/reporting.sql:98-115`)
  gegen Tabellen-Alias `LEFT JOIN users u` (`:65`, `:156`)
- uneinheitliche Join-Form in derselben Datei — Komma-Join (`reporting.sql:104-105`) gegen
  `CROSS JOIN LATERAL` (`:122`)
- toter Prop in der Storno-Anzeige — `className` in
  `frontend/src/admin/reporting/StornoServicekraft.tsx:35-43`
- doppelte Assertion-Blöcke in **einer** Testdatei (nicht, wie die alte Liste nahelegt, in
  zwei): die beiden Funktionen in `backend/api/fiskal/export/http/handler_test.go`
- zwei reine Weiterleitungen im Kassen-Command —
  `backend/api/kasse/tischgeschaeft/application/command.go:126`
  (`writeEventWithDruckauftraege`, ein Aufrufer) und `:89` (`writeEventOCC` gibt eine
  ungenutzte `eventID` zurück)
- Ladeflagge ohne sprechenden Namen — `isPending` in `frontend/src/service/TablePage.tsx:97`
  neben `stateLoading`/`historieLoading`

Das ist eine Erwartung, kein Soll — null Punkte wären ein gültiges Ergebnis.

**Reihenfolge-Hinweis:** `backend/api/fiskal/export/http/handler_test.go` wird in Phase 3
ohnehin erweitert. Diesen einen Punkt deshalb entweder überspringen oder erst nach Phase 3
anfassen, statt dieselbe Datei zweimal umzuschreiben.

### Acceptance criteria

- [ ] Der Cleanup-Durchlauf lief gegen den v0.17.2-Stand, nicht gegen die alte Liste
- [ ] Jeder angewandte Punkt ist am Code belegt und nachweislich verhaltensneutral
- [ ] Punkte der alten Liste, deren Ziel nicht mehr existiert, sind als solche vermerkt
      statt still verschwunden
- [ ] `make verify` und `make test-frontend` laufen grün
- [ ] Der Diff enthält keine Verhaltensänderung — Rückgabewerte, Fehlerpfade und
      Reihenfolge von Seiteneffekten sind unverändert

---

## Phase 2: Retry-Politik und Korrelations-ID

### Context

- `frontend/src/lib/queryClient.ts — createQueryClient()` — setzt heute keine
  `defaultOptions`; es gilt der Bibliotheks-Default von drei Wiederholungen für **jeden**
  Fehler.
- `frontend/src/hooks/use-action-submit.ts — useActionSubmit()` — der Pfad, über den
  Buchungen laufen. Am Code bestätigt: reiner `useState`, **nicht** react-query. Einzige
  Ausnahme im Repo ist `frontend/src/admin/reporting/hooks.ts:58 — useDsfinvkExport()`
  (`useMutation`); eine Retry-Politik unter `defaultOptions.queries` erreicht Mutations
  nicht, der Export bleibt also unberührt. Das ist der Grund, warum die Politik
  ausdrücklich nur unter `queries` gesetzt wird und nicht unter `mutations`.
- `archiv/main-vor-v0.17.2`, Commit `d85cf2ae` — enthält beide Änderungen als Vorlage.

### What to build

Zwei kleine, unabhängige Verbesserungen, jede als eigener Commit, neu geschrieben statt
gepickt.

**Erstens die Retry-Politik.** Ein Fehler aus der 4xx-Familie wird heute dreimal
nachgeschleift, bevor die Meldung erscheint — die Helferin wartet auf eine Auskunft, die
schon beim ersten Versuch feststand. Künftig werden nur Netzfehler und Serverfehler
wiederholt, und seltener. Der Eingriff bleibt auf `defaultOptions.queries` begrenzt;
Buchungen sind nicht betroffen, weil sie an react-query vorbeilaufen, und der
DSFinV-K-Export nicht, weil Query-Defaults für Mutations nicht gelten.

**Zweitens die Korrelations-ID.** Der zentrale Fehler-Toast zeigt künftig die Referenz aus
der Backend-Antwort, damit ein gemeldeter Fehler im Server-Log auffindbar ist.

### Acceptance criteria

- [ ] Ein 4xx-Fehler in einer Lese-Query erscheint ohne Wiederholungen
- [ ] Ein Netzfehler wird weiterhin wiederholt
- [ ] Am Code belegt und durch einen Test gepinnt, dass die Politik keine Buchung berührt
- [ ] Der zentrale Fehler-Toast trägt die Korrelations-ID, wenn die Antwort eine liefert,
      und bleibt ohne sie unverändert lesbar
- [ ] `make test-frontend` läuft grün

---

## Phase 3: Langläufer-Schutz der TSE-Einrichtung

### Context

- `backend/api/fiskal/setup/http/command_handler.go` — **drei** schreibende Endpunkte,
  alle reichen heute `r.Context()` unverändert durch:
  `UpdateTSEKonfigurationHandler()` (Zeile 87), `RichteTSEEinHandler()` (113) und
  `UebernimmTSEHandler()` (157). Den fiskaly-Lebenszyklus fahren nur die letzten beiden —
  sie bekommen die Kontext-Entkopplung. Das Schloss legt sich über alle drei.
  `query_handler.go` liest nur und bleibt unberührt.
- `backend/api/fiskal/export/http/handler.go` — **hier existiert die Fristverlängerung
  bereits** (Zeile 68, `http.NewResponseController(w).SetWriteDeadline(...)`,
  `exportWriteTimeout = 5 * time.Minute`, aus Commit `f3fb8aea`). Der neue Helfer wird
  daraus **extrahiert**, nicht danebengestellt. Die Lücke, die bleibt: Der Aufruf steht
  *hinter* dem Fehler-Switch (Zeilen 50-61) — eine Fehlerantwort nach einem langen,
  gescheiterten Archivbau läuft also weiterhin gegen die globale 10-Sekunden-Frist. Genau
  das schließt der zweite Aufruf am Handler-Eingang.
- `backend/api/fiskal/setup/application/setup.go — RichteTSEEin()` — fährt den fiskaly-
  Lebenszyklus sequenziell und speichert erst danach; PUK und Admin-PIN entstehen im Lauf,
  werden nirgends persistiert und gehen genau einmal in die Antwort.
- `backend/app/app.go` — die globale Schreibfrist ist deutlich kürzer als ein
  Einrichtungslauf.
- `backend/api/helper/http.go` — Ort für die neue Fristverlängerung, neben den bestehenden
  Antwort-Helfern.
- `backend/api/middleware/middleware.go — responseWriter.Unwrap()` — bereits im Baum, macht
  die Fristverlängerung hinter der Logging-Middleware überhaupt erst wirksam.
- `frontend/src/lib/errorMessages.ts` — `commonErrorMessages` (Zeilen 10-94), die
  Zuordnung Fehlercode → deutscher Klartext. `tse_setup_laeuft_bereits` fehlt dort
  naturgemäß noch.
- `archiv/main-vor-v0.17.2`, Commit `9ef41ab6` — Vorlage. `9ef41ab6` ist **kein Vorfahr**
  von `HEAD`; die Historie ist auseinandergelaufen. Nachgeprüft: Für alle von `9ef41ab6`
  berührten Dateien dieser Phase (`setup/http/*`, `setup/application/*`, `helper/http.go`,
  `export/http/*`) ist `HEAD` byte-identisch mit `9ef41ab6^`. Die Vorlage lässt sich also
  sauber übernehmen; alle benötigten Test-Helfer existieren bereits.

### What to build

Drei zusammengehörige Schutzmaßnahmen für die drei schreibenden TSE-Setup-Endpunkte und
den DSFinV-K-Export.

**Fristverlängerung.** Ein Helfer kapselt das Heraufsetzen der Schreibfrist auf der
Verbindung und läuft weiter, wenn der Writer es nicht kann. Er entsteht durch Extraktion
der bereits im Export-Handler vorhandenen Fassung nach `backend/api/helper/http.go`, nicht
als zweite, parallele Implementierung. Er wird an jedem betroffenen
Handler zweimal aufgerufen: einmal am Eingang, damit auch frühe Fehlerantworten den Client
erreichen, und einmal unmittelbar vor dem Schreiben der Antwort. Zwei Aufrufe sind nötig,
weil die Frist eine absolute Zeit ab dem Lesen des Request-Headers ist und während der
Handler-Laufzeit mitverstreicht.

**Kontext-Entkopplung.** Die beiden schreibenden TSE-Endpunkte fahren den Lebenszyklus auf
einem Kontext, den ein Client-Abbruch nicht mehr storniert, mit eigener Obergrenze. Damit
erreicht ein einmal begonnener Lauf sein Speichern, statt eine bezahlte Ruine zu
hinterlassen.

**Prozess-Schloss.** Die Entkopplung macht einen zweiten, überlappenden Einrichtungslauf
denkbar — der würde eine zweite bezahlte TSS anlegen. Ein Schloss über alle drei
schreibenden Pfade lehnt den zweiten Aufruf mit dem Fehlercode `tse_setup_laeuft_bereits`
ab, den das Frontend als Klartext zeigt: dass gerade eine Einrichtung läuft und gewartet
werden muss. Der Eintrag gehört in `frontend/src/lib/errorMessages.ts` nach
`commonErrorMessages` (Zeilen 10-94); ohne ihn greift der Fallback in Zeile 107,
`` `${actionLabel} fehlgeschlagen. Bitte erneut versuchen.` `` — genau der Rat, der die
zweite bezahlte TSS erzeugt.

Der Kommentar, der die gewählten Fristen begründet, wird neu formuliert: die Vorlage rechnet
gegen Client-Zeitlimits, die es im Zielbaum nicht gibt.

### Acceptance criteria

- [ ] Ein Client-Abbruch während der TSE-Einrichtung lässt den Lebenszyklus zu Ende laufen
      und die Einrichtung speichern
- [ ] Ein zweiter, überlappender Einrichtungsversuch wird mit HTTP 409 und eigenem
      Fehlercode abgelehnt und legt keine zweite TSS an
- [ ] Das Frontend zeigt für diesen Code eine verständliche deutsche Meldung, nicht den
      generischen Fallback
- [ ] Fehlerantworten des DSFinV-K-Exports erreichen den Client auch dann, wenn der
      Archivbau die globale Schreibfrist überschritten hat
- [ ] Die vier neuen Testdateien laufen; das Schloss wird nach jedem Test zurückgesetzt,
      sodass ein Fehlschlag keine Folgetests mitreißt
- [ ] `make verify` läuft grün

---

## Phase 4: Clientversion und Erkennung

### Context

- `.github/workflows/release.yml` — reicht die Version bisher nur an das Backend-Image.
- `frontend/Dockerfile` — ohne `ARG`; das Argument muss im `RUN` konsumiert werden, sonst
  greift der Layer-Cache und das alte Bundle landet im neuen Image.
- `frontend/vite.config.ts` und `frontend/vitest.config.ts` — zwei eigenständige Configs;
  die Konstante muss in **beiden** stehen.
- `frontend/tsconfig.app.json` — inkludiert nur `src`; eine Deklarationsdatei muss dort
  liegen.
- `frontend/src/admin/hooks.ts — useVersion()` — heute `staleTime: Infinity` und im
  Admin-Bereich verortet.
- `frontend/src/admin/reporting/hooks.ts` — einzige Polling-Präzedenz im Repo.

### What to build

Die Clientversion wird zur Bauzeit eingebrannt, analog zur Backend-Kette, mit Default `dev`
auf beiden Seiten.

Die Versionsabfrage zieht von `src/admin/` nach `src/lib/`, damit der Service-Bereich sie
nutzen kann, ohne die Bereichs-Trennung zu brechen. Sie fragt künftig regelmäßig und beim
Zurückkehren in den Vordergrund nach, statt genau einmal je Seitenladen. Der zentrale
Fehler-Toast darf davon nicht ausgelöst werden — eine Versionsabfrage, die bei Funkloch
periodisch eine rote Meldung wirft, wäre eine Verschlechterung.

**Das Intervall ist 30 Sekunden**, gleichauf mit der einzigen bestehenden Polling-Stelle
(`frontend/src/admin/reporting/hooks.ts:52`). Das ist eine bewusste Festlegung, keine
Nebensache: `/health` pingt bei jedem Aufruf die Datenbank, und mit dreißig Helfer-Handys
wird das der meistgerufene Endpunkt des Systems. Dreißig Sekunden sind für Postgres
belanglos und für einen Versionswechsel schnell genug — wer ein Update ansagt, wartet
ohnehin länger. Kürzer wird es nicht ohne Anlass.

Diese Phase liefert **noch keinen Reload**. Sichtbares Ergebnis ist ein
nicht-blockierender Hinweis, dass eine neue Version bereitsteht. Das macht die ganze Kette
vom Tag bis ins Bundle überprüfbar, ohne den gefährlichen Teil scharf zu schalten.

Der Vergleich ist nur zwischen zwei echten Release-Versionen aktiv. In Dev, in E2E und in
Tests steht auf beiden Seiten derselbe Default — dort schlägt er per Konstruktion nie an.

### Acceptance criteria

- [ ] Ein aus einem Tag gebautes Frontend-Image trägt dieselbe Version wie das
      Backend-Image desselben Laufs
- [ ] Ein erneuter lokaler Build mit geänderter Version erzeugt nachweislich ein neues
      Bundle und nicht das gecachte alte
- [ ] `pnpm build` und `make test-frontend` laufen; kein Test scheitert an einer
      undefinierten Build-Konstante
- [ ] Die Versionsabfrage bemerkt einen Serverwechsel ohne Neuladen der Seite
- [ ] Die Versionsabfrage läuft alle 30 Sekunden und beim Zurückkehren in den Vordergrund
- [ ] Eine fehlgeschlagene Versionsabfrage erzeugt keinen Fehler-Toast
- [ ] Der Versionsvergleich ist eine reine Funktion und durch einen Tabellentest gepinnt:
      `dev`/`dev`, `dev`/Release, Release/`dev`, gleiche Releases, verschiedene Releases,
      `dev-<sha>` und eine Vorabversion — nur „zwei verschiedene echte Releases" ergibt
      eine Abweichung
- [ ] In Dev, E2E und Tests bleibt der Hinweis aus
- [ ] Die Versionsanzeige in der Admin-Seitenleiste funktioniert unverändert

---

## Phase 5: Vorgangs-Register

### Context

- `frontend/src/lib/Auth.ts — AuthSingleton`, `frontend/src/lib/Backend.ts — BackendSingleton`
  — die bestehende Singleton-Landschaft, deren Muster das Register folgt.
- `frontend/src/hooks/use-mengen.ts — useMengen()` — generischer Träger für alle Körbe und
  Mengenauswahlen.
- `frontend/src/hooks/use-action-submit.ts`, `frontend/src/hooks/use-form-action-submit.ts`
  — generische Träger für jede laufende Buchung und jeden Formular-Submit.
- `frontend/src/admin/tse/TSEEinrichtungWizard.tsx`,
  `frontend/src/admin/kasse/ZaehlhilfeDialog.tsx`,
  `frontend/src/admin/kasse/GeldtransitDialog.tsx`,
  `frontend/src/admin/kasse/KasseAbschliessenSection.tsx`,
  `frontend/src/admin/reporting/hooks.ts` — die Zustände, die keinen generischen Träger
  haben und sich einzeln melden müssen.
- `frontend/src/service/components/table/HistorieStornierungDrawer.tsx` und die beiden
  Schwester-Drawer `HistorieUmbuchungDrawer.tsx` sowie
  `frontend/src/service/components/direktverkauf/DirektverkaufStornoDrawer.tsx` — ihre
  Kommentarfelder sind Pflichtangaben und damit meldepflichtig. (Ihre Mengenauswahl läuft
  bereits über `useMengen`; hinzu kommt nur der Kommentar.)

### What to build

Ein Register, das zählt, wie viele Vorgänge gerade offen sind, und Interessenten
benachrichtigt, wenn sich das ändert. Geschrieben wird ausschließlich über einen Hook, der
sich im Cleanup wieder abmeldet; gelesen wird über die React-Schnittstelle für externe
Zustände.

„Offen" heißt: etwas, dessen Verlust eine Helferin ärgern würde. Der weitaus größte Teil
wird über die drei generischen Träger erfasst — ein gefüllter Korb, eine getroffene
Mengenauswahl, eine laufende Buchung, ein laufender Formular-Submit, der Beleg-Nachfasslauf.
Einzeln zu melden bleiben der TSE-Wizard, die Zählhilfe, das Geldtransit-Formular, das
Tagesabschluss-Formular, der laufende DSFinV-K-Export und die Pflicht-Kommentarfelder der
Storno- und Umbuchungs-Drawer.

Reine Anzeigen melden sich nicht: Detail-Ansichten der Historie, die Tischauswahl mit ihrem
Suchfeld, die Erfolgsmeldung mit ihrer kurzen Standzeit.

**`frontend/src/components/common/LoginForm.tsx` meldet bewusst nichts**, obwohl es seinen
Submit-Zustand von Hand hält und damit an den generischen Trägern vorbeiläuft. Meldete es
sich, würde der Reload bis nach der erfolgreichen Anmeldung warten und dann genau dort
feuern, wo er am meisten stört. Ohne Meldung greift er beim Aufschlagen der Anmeldeseite,
bevor jemand tippt — siehe die Entscheidung zum Guard auf der Anmeldeseite oben.

Diese Phase ist für sich genommen unsichtbar. Sie wird über Tests abgenommen und über eine
temporäre Anzeige während der Entwicklung, die vor dem Abschluss der Phase wieder
verschwindet.

### Acceptance criteria

- [ ] Ein gefüllter Bestell-Korb, eine getroffene Kassieren-Auswahl und ein gefüllter
      Direktverkauf-Korb melden je einen offenen Vorgang
- [ ] Jede laufende Buchung meldet einen offenen Vorgang und gibt ihn nach Erfolg **und**
      nach Fehlschlag wieder frei
- [ ] Die sechs Zustände ohne generischen Träger melden sich einzeln
- [ ] Der Zähler kehrt beim Verlassen einer Seite, beim Tischwechsel und beim Umschalten
      zwischen Drawer- und Spaltenlayout zuverlässig auf null zurück
- [ ] Ein Test schlägt fehl, wenn eine Abmeldung aus dem Cleanup entfernt wird
- [ ] Das Register lässt sich zwischen Tests zurücksetzen
- [ ] Reine Anzeigen melden keinen Vorgang

---

## Phase 6: Sperrhinweis und erzwungener Reload

### Context

- `frontend/src/App.tsx — App()` — der Montagepunkt.
- `frontend/src/components/common/ErrorBoundary.tsx` — Vorlage für Text und Reload-Aufruf.
- `frontend/src/components/ui/alert-dialog.tsx` — Vorlage für den nicht wegklickbaren
  Hinweis.
- `docker-compose.prod.yml:110-112` und `scripts/prod-update.sh:154,159` — die
  Reihenfolge, aus der die Reload-Schleife entsteht und gegen die die Bremse gebaut wird.
- `frontend/src/lib/Auth.ts:95` — das Token liegt im `localStorage`, ein Reload meldet also
  niemanden ab.
- Das Vorgangs-Register aus Phase 5 und die Erkennung aus Phase 4.

### What to build

Der nicht-blockierende Hinweis aus Phase 4 wird scharf geschaltet.

Erkennt der Guard eine Abweichung und ist das Register leer, lädt die Seite sofort neu.
Ist das Register nicht leer, erscheint ein nicht wegklickbarer Hinweis, der in klarem
Deutsch sagt, was zu tun ist: den laufenden Vorgang abschließen oder verwerfen, danach wird
automatisch neu geladen. Der Hinweis blockiert keine Eingabe, die zum Abschließen nötig ist
— er erklärt und wartet. Wird das Register leer, löst der Reload von selbst aus.

Der Guard entscheidet außerhalb des Renders und darf niemals mehrfach auslösen. Er greift
nur bei einer erfolgreich beantworteten Versionsabfrage, nie bei einem Fehler — ein
Serverneustart darf keinen Reload erzwingen, nur ein tatsächlicher Versionswechsel.

**Die Schleifenbremse.** Unmittelbar vor `window.location.reload()` vermerkt der Guard die
Serverversion, auf die hin er lädt, in `sessionStorage`. Beim nächsten Start liest er den
Vermerk: Trägt der Client immer noch nicht diese Zielversion, war der Reload wirkungslos —
dann wird **nicht** erneut geladen, sondern nur der Hinweis gezeigt. Stimmt die eigene
Version mit dem Vermerk überein, wird er gelöscht und der Mechanismus steht wieder scharf.
`sessionStorage` und nicht `localStorage`: Der Vermerk soll den Reload überleben, aber
nicht das Schließen des Tabs.

Diese Bremse ist kein Feinschliff, sondern der Unterschied zwischen „lädt einmal neu" und
„lädt im Update-Fenster in Endlosschleife neu" — die Herleitung steht unter „Risiken".
Sie verwandelt den einzigen unbegrenzten Fehlerfall des Handshakes in genau einen
wirkungslosen Reload plus stehenden Hinweis.

**Der gebremste Hinweis braucht einen eigenen Text und eine Schaltfläche.** Greift die
Bremse, wäre die Zusage „es lädt gleich von selbst neu" schlicht gelogen — dieser Client
lädt nicht mehr von selbst. Im gebremsten Fall sagt der Hinweis deshalb, dass das
automatische Erneuern nicht durchgekommen ist und von Hand nachgeholt werden muss, und
trägt eine Schaltfläche „Jetzt neu laden". Das ist die einzige Stelle, an der eine
Schaltfläche richtig ist: Im Normalfall wäre sie ignorierbar und damit wertlos (siehe
„Resolved decisions"), im gebremsten Fall ist sie der einzige ehrliche Ausweg. Sie ist
zugleich der Rettungsanker für den Fall, dass ein Client dauerhaft an einer alten
Auslieferung hängt.

**Auf der Anmeldeseite bleibt der Guard aktiv.** Das Register ist dort leer, der Reload
greift also beim Aufschlagen der Seite — der billigste Zeitpunkt überhaupt. Begründung
unter „Resolved decisions"; `LoginForm` meldet bewusst keinen Vorgang (Phase 5).

### Acceptance criteria

- [ ] Bei Versionsabweichung und leerem Register lädt die Seite ohne Zutun neu
- [ ] Bei Versionsabweichung und offenem Vorgang erscheint der Sperrhinweis, und der Reload
      unterbleibt
- [ ] Der Vorgang lässt sich mit stehendem Hinweis noch abschließen oder verwerfen
- [ ] Sobald das Register leer wird, lädt die Seite von selbst neu
- [ ] Der Hinweis ist nicht wegklickbar — weder über Escape noch über einen Klick daneben
- [ ] Eine fehlgeschlagene Versionsabfrage löst keinen Reload aus
- [ ] Der Reload löst innerhalb eines Seitenlebens genau einmal aus, auch wenn mehrere
      Abfragen dieselbe Abweichung melden
- [ ] **Ein wirkungsloser Reload wiederholt sich nicht:** Lädt der Client neu und trägt
      danach immer noch nicht die vermerkte Zielversion, bleibt es beim Hinweis — durch
      einen Test gepinnt, der den `sessionStorage`-Vermerk vorbelegt
- [ ] Der gebremste Hinweis verspricht kein automatisches Neuladen mehr und bietet
      stattdessen „Jetzt neu laden" an; die Schaltfläche lädt tatsächlich neu
- [ ] Stimmt die eigene Version mit dem Vermerk überein, ist der Vermerk gelöscht und ein
      späterer Versionswechsel löst wieder einen Reload aus
- [ ] Auf der Anmeldeseite greift der Guard wie überall sonst
- [ ] Ein Durchlauf am laufenden Stack bestätigt das Verhalten mit einem echten
      Versionswechsel, nicht nur im Test
- [ ] Der Reload funktioniert auch in der als App installierten Ansicht
      (`"display": "standalone"`, ohne Adresszeile und Neu-laden-Pfeil) — der Fall, den der
      Leitfaden heute nur durch Wegwischen der App lösen kann

---

## Phase 7: Leitfaden und Release

### Context

- `docs/leitfaden/aktualisieren.md` — enthält seit v0.17.2 den von Hand anzusagenden
  Reload-Schritt.
- `docs/handbuch.md` — Architekturdokumentation.
- `docs/plans/review-v0.17.2.md` — wird mit diesem Release abgearbeitet.

### What to build

Der Reload-Schritt verschwindet aus dem Aktualisierungs-Leitfaden — ab v0.17.3 erledigt ihn
das System. Betroffen sind der Abschnitt „Danach: jedes betroffene Gerät einmal neu laden"
(`docs/leitfaden/aktualisieren.md`, Zeilen 47-105) samt der versionsspezifischen
Unterabschnitte und der Punkt 6 der Mitten-im-Fest-Reihenfolge. An ihre Stelle tritt die
Beschreibung dessen, was die Helfer erwarten dürfen: dass sich die App nach einem Update
von selbst erneuert, und dass ein stehender Hinweis bedeutet, den laufenden Vorgang zuerst
abzuschließen.

Ausdrücklich mit weg kann die umständlichste Passage: dass die als App installierte jotti
„ganz geschlossen" und weggewischt werden muss. Genau diesen Fall löst der Handshake ohne
Zutun.

Ein kurzer Absatz beschreibt den Ausnahmefall: Steht ein Hinweis mit der Schaltfläche
„Jetzt neu laden", ist das automatische Erneuern nicht durchgekommen — einmal antippen
genügt. Das ersetzt keine Anweisung an das Team, sondern erklärt, was ein Helfer sieht,
wenn er es sieht.

Zwei Punkte kommen hinzu. Erstens die Warnung, während eines Updates keine TSE-Einrichtung
zu starten — der Schutz aus Phase 3 deckt einen Serverneustart nicht ab. Zweitens ein Absatz
im Handbuch, der den Handshake als Mechanismus festhält, damit er bei künftigen Änderungen
an Antwortformaten mitgedacht wird.

Anschließend Release nach der üblichen Mechanik: Tag setzen, CI bauen lassen, Artefakte
prüfen.

### Acceptance criteria

- [ ] Der von Hand anzusagende Reload-Schritt ist aus dem Leitfaden entfernt
- [ ] Der Leitfaden beschreibt das neue Verhalten aus Sicht der Helfer
- [ ] Die Warnung zur TSE-Einrichtung während eines Updates steht im Leitfaden
- [ ] Das Handbuch beschreibt den Versions-Handshake
- [ ] `docs/plans/review-v0.17.2.md` ist abgearbeitet und gelöscht
- [ ] `make verify`, `make test-frontend` und `make test-e2e` laufen grün
- [ ] Tag `v0.17.3` ist gesetzt und CI ist grün
