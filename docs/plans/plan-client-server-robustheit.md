# Plan: Robustheit der Client-Server-Kommunikation

> Quell-PRD: n/a (aus Praxis-Meldung und Codeanalyse)

## Ziel

Verbindungs- und Ladefehler im Service-Betrieb sollen für ehrenamtliche Helfer
**erkennbar, korrekt benannt und selbst behebbar** sein — statt als plausible
Falschaussage über den Arbeitszustand zu erscheinen.

Auslöser: Ein Helfer meldete, das Service-Dashboard lade seine markierten Tische
nicht. Die App zeigte dabei „Keine Tische markiert" (Leerzustand), obwohl drei
Favoriten existierten. Der Verein hatte zuvor Tische gelöscht — das ist die
bestätigte Ursache: ein gelöschter Favoriten-Tisch lässt
`GetMeineTischeState` dauerhaft mit 500 fehlschlagen, und die Tischübersicht
stellt jeden Fehler als „leer" dar.

Der Plan behebt diese Ursache und härtet danach die darunterliegenden
Schwächen des Kommunikationspfads.

## Architekturentscheidungen

Durchgängig gültig für alle Phasen:

- **Endpunkte**: keine neuen Routen. Alle Änderungen erweitern bestehende
  POST-Endpunkte oder deren Antworten.
- **Fehlerformat**: Jede Fehlerantwort der API — ausnahmslos, auch aus
  Middleware — nutzt `errorResponse` (`backend/api/helper/http.go`) mit
  `{ "code": "<snake_case>" }`. Plain-Text-Fehlerantworten sind ein Defekt.
- **Fehlerdarstellung im Frontend**: Eine fehlgeschlagene Query wird nie als
  Leerzustand gerendert. Kanonisches Muster ist `LadefehlerAlert`
  (`frontend/src/components/common/LadefehlerAlert.tsx`), wie bereits in
  `frontend/src/service/TablePage.tsx` verwendet.
- **Idempotenz-Schlüssel**: Der Client erzeugt pro fachlichem Vorgang eine
  UUID und sendet sie als `vorgangId`. „Vorgang" ist der etablierte Begriff der
  Ubiquitous Language (`docs/language.md`). Der Schlüssel ist **kein**
  Event-Feld — die Event-JSON-Contracts bleiben unberührt (Freeze-Disziplin,
  Guard: `backend/domain/kasse/event_json_contract_test.go`). Er wird in einer
  eigenen Tabelle im selben Commit wie die Events festgehalten.
- **Schema**: neue Tabelle `vorgang_idempotenz`
  (`vorgang_id UUID PRIMARY KEY`, `art TEXT NOT NULL`,
  `user_id INT REFERENCES users(id) NOT NULL`, `created_at TIMESTAMPTZ NOT NULL`).
- **Migrationen**: forward-only, additiv, transaktional geklammert
  (`database/migrations/README.md`). Nächste freie Nummern: `06`, `07`.
- **Betriebsmodus**: jotti läuft ausschließlich im LAN-Mode (Server steht am
  Veranstaltungsort, jedes Helfer-Gerät hat eine eigene IP). Eine Umstellung
  der Rate-Limits von IP- auf Benutzer-Ebene ist damit gegenstandslos und
  **nicht** Teil dieses Plans; korrigiert wird nur das Antwortformat des
  Limiters.
- **Offline**: kein Service Worker, keine neue Dependency. Das Offline-Signal
  kommt aus `navigator.onLine` und dem `onlineManager` von
  `@tanstack/react-query`.

## Inventar

Backend:

- `backend/api/kasse/tischgeschaeft/application/query.go — Query.GetMeineTischeState()`
  — bricht die gesamte Antwort ab, wenn ein Favorit nicht auflösbar ist.
- `backend/repository/kassenjournal_repo/repo.go — Repository.ReadFavoritenTischStates()`
  — Batch-Join, filtert `t.status != 'deleted'`.
- `backend/api/stammdaten/tisch/application/command.go — Command.TischLoeschen()`,
  `Command.applyTischStatusChange()` — Soft-Delete ohne Favoriten-Cleanup.
- `backend/api/kasse/tischgeschaeft/application/command.go — Command.ZahlungKassieren()`,
  `Command.StornierungErteilen()`, `Command.BestellungUmbuchen()`,
  `Command.BestellungAufnehmen()` — letzterer ist die Referenz-Implementierung
  für Idempotenz.
- `backend/api/kasse/direktverkauf/application/command.go — Command.DirektverkaufStornieren()`,
  `Command.persistVerkaufEvent()` — der vierte zu härtende Vorgang.
- `backend/repository/kassenjournal_repo/repo.go — Repository.WriteEvent()`,
  `Repository.WriteEventWithDruckauftraege()`,
  `Repository.WriteTischSessionEventsAtomic()`, `Repository.WriteUmbuchung()`
  — die Schreibpfade der zu härtenden Vorgänge; alle nutzen `db.WithTx`.
- `backend/api/middleware/middleware.go — RateLimitMiddleware()` — sendet
  Plain Text statt JSON.
- `backend/api/helper/http.go — SendJSONResponse()`, `errorResponse` — das
  kanonische Fehlerformat.
- `backend/app/app.go — NewApp()` — `WriteTimeout: 10s` für alle Antworten.
- `backend/api/fiskal/export/http/handler.go — ExportHandler()` — streamt das
  DSFinV-K-ZIP gegen ebendieses Limit.
- `backend/api/stammdaten/tisch/http/command_handler.go — TischLoeschenHandler()`.

Frontend:

- `frontend/src/lib/Backend.ts — Backend.request()`, `Backend.post()`,
  `Backend.throwIfNotOk()` — kein Timeout, ungeschützte Body-Deserialisierung.
- `frontend/src/lib/queryClient.ts — createQueryClient()` — nur Toast, keine
  `defaultOptions`.
- `frontend/src/lib/errorMessages.ts — getActionErrorMessage()`,
  `commonErrorMessages` — Code-zu-Text-Abbildung.
- `frontend/src/service/table/hooks.ts — useMeineTischeState()`,
  `useAktiveTischeMitFavoriten()`, `useEigeneUebersicht()` — verschlucken
  `isError`.
- `frontend/src/service/TableSelectionPage.tsx — TableSelectionPage()` — rendert
  den Fehler als Leerzustand.
- `frontend/src/service/product/hooks.ts — useAktiveProdukte()`,
  `frontend/src/service/direktverkauf/hooks.ts — useDirektverkaufHistorie()` —
  gleiches Muster.
- `frontend/src/service/components/table/BestellungAbschluss.tsx — BestellungAbschluss()`
  — Referenz für den Lebenszyklus eines client-erzeugten Vorgangs-Schlüssels.
- `frontend/src/service/table/TischBackend.ts — TischBackend` — Domain-Backend-Klasse
  der Tischvorgänge.
- `frontend/src/service/table/Zahlung.ts — ZahlungKassierenSchema`,
  `frontend/src/service/table/Stornierung.ts — StornierungErteilenSchema`,
  `frontend/src/service/table/Umbuchung.ts — BestellungUmbuchenSchema`.
- `frontend/src/App.tsx — App()` — globale Shell, Ort des Offline-Banners.
- `frontend/src/components/common/LadefehlerAlert.tsx — LadefehlerAlert()`.

Datenbank:

- `database/migrations/01_initial.up.sql` — `tisch_favoriten` (kein
  `ON DELETE`-Verhalten, Tische werden soft-deleted), die partiellen
  Unique-Indexe `idx_kassenjournal_bestellung_id`,
  `idx_kassenjournal_verkauf_id`, `idx_kassenjournal_geldtransit_id`.
- `backend/sqlc/queries/tische.sql — GetAktiveTischeMitFavoriten` — listet nur
  `status = 'active'`, ein gelöschter Favorit ist im Drawer unsichtbar und
  damit nicht abwählbar.

## Getroffene Entscheidungen

- **Betrieb ist LAN-only.** Keine Arbeit an IP-/NAT-Rate-Limits; nur das
  Antwortformat des Limiters wird korrigiert (Phase 6).
- **Offline: nur Banner, kein Service Worker.** Keine neue Dependency, keine
  eigene Cache-Invalidierung. Das Reload-im-Funkloch-Problem (weiße Seite)
  bleibt bewusst bestehen und wird nicht gelöst.
- **Idempotenz für alle drei angefragten Vorgänge** (Kassieren, Stornieren,
  Umbuchen). **Erweiterung gegenüber der Absprache:** Der Plan zieht die
  Direktverkauf-Stornierung (`serviceleitung/direktverkauf-stornieren`) mit —
  sie ist der vierte buchende Endpunkt ohne Schlüssel, strukturgleich zu den
  drei anderen. Streichbar, ohne den Rest der Phase zu berühren.
- **Verwaiste Favoriten: beides** — das Backend überspringt nicht auflösbare
  Favoriten defensiv, und eine Migration räumt die Bestandszeilen auf.
- **Der Idempotenz-Schlüssel wandert nicht ins Event-JSON.** „Stornieren" ist
  serverseitig eine 1:n-Aktion (ein `bestellung-korrigiert` plus je betroffener
  Zahlung ein `stornierung-erteilt`), „Umbuchen" schreibt zwei Events. Ein
  einzelnes Event-ID-Feld kann den Vorgang deshalb nicht eindeutig abbilden;
  eine eigene Tabelle deckt alle vier Vorgangsformen (ein Event, zwei Events,
  n Events) mit einem Mechanismus ab und lässt die eingefrorenen
  Event-Contracts unangetastet.
- **Die bestehenden partiellen Unique-Indexe für `bestellungId`, `verkaufId`
  und `geldtransitId` bleiben unverändert.** Sie sind produktiv erprobt und
  liefern eine harte DB-Garantie auf den Event-Zeilen. Sie werden weder
  entfernt noch auf die neue Tabelle umgestellt — eine Umstellung erprobter
  Geldpfade ohne Anlass wiegt schwerer als die verbleibende Zweigleisigkeit.

## Risiken

- **Phase 7 berührt Geldpfade.** Reihenfolge innerhalb der Transaktion ist
  korrektheitsrelevant: Die Idempotenz-Zeile muss **vor** den Event-Inserts
  geschrieben werden, sonst ist ein Unique-Konflikt nicht mehr eindeutig einem
  Duplikat (statt einem echten OCC-Konflikt) zuzuordnen.
- **Phase 4 ändert globales Verhalten** (`staleTime`, `retry`) und wirkt damit
  auch im Admin-Bereich. Beabsichtigt, aber beim Review mitzuprüfen.
- **Migration 07 löscht Zeilen.** `tisch_favoriten` ist reine
  Benutzer-Präferenz und trägt keine aufbewahrungspflichtigen Kassendaten;
  betroffen sind ausschließlich Zeilen, deren Tisch bereits gelöscht ist.

---

## Phase 1: Gelöschte Tische legen das Service-Dashboard nicht mehr lahm

### Kontext

- `backend/api/kasse/tischgeschaeft/application/query.go — Query.GetMeineTischeState()`
  — bricht mit `ErrDatabase` ab, sobald ein Favorit nicht in der Batch-Antwort
  steht; das ergibt 500 für **alle** Favoriten des Helfers.
- `backend/repository/kassenjournal_repo/repo.go — Repository.ReadFavoritenTischStates()`
  — liefert gelöschte Tische bewusst nicht mit; das Verhalten bleibt.
- `backend/api/stammdaten/tisch/application/command.go — Command.TischLoeschen()`
  — setzt nur den Status, ohne Favoriten aufzuräumen.
- `database/migrations/README.md` — Regeln für neue Migrationen.

### Was gebaut wird

Ein gelöschter Tisch darf das Dashboard derjenigen Helfer, die ihn markiert
hatten, nicht unbrauchbar machen — weder künftig noch in Bestandsinstanzen.

Drei zusammengehörige Änderungen:

1. `GetMeineTischeState` überspringt einen nicht auflösbaren Favoriten mit einer
   Warn-Log-Zeile (Tisch-ID und Benutzer-ID), statt die gesamte Abfrage
   abzubrechen. Die übrigen Favoriten werden normal geliefert.
2. `TischLoeschen` entfernt die `tisch_favoriten`-Zeilen des gelöschten Tisches
   im selben Commit wie den Statuswechsel. Nur beim Löschen — ein
   deaktivierter Tisch bleibt markierbar und wird von
   `ReadFavoritenTischStates` weiterhin geliefert.
3. Migration `07_favoriten_cleanup.up.sql` löscht die bereits verwaisten Zeilen
   (`tisch_favoriten`-Einträge, deren Tisch `status = 'deleted'` trägt oder
   nicht mehr existiert).

### Akzeptanzkriterien

- [x] Ein Helfer mit einem gelöschten und zwei aktiven Favoriten erhält von
      `service/get-meine-tische-state` HTTP 200 mit den zwei aktiven Tischen.
- [x] Der übersprungene Favorit erzeugt genau eine Warn-Log-Zeile mit Tisch-ID
      und Benutzer-ID.
- [x] Nach `admin/delete-tisch` existiert keine `tisch_favoriten`-Zeile mehr
      für diesen Tisch; die Favoriten anderer Tische bleiben unberührt.
- [x] Ein **deaktivierter** (nicht gelöschter) Favoriten-Tisch wird weiterhin
      als Favorit geführt und in der Antwort geliefert.
- [x] Migration `07_favoriten_cleanup.up.sql` liegt vor, ist in `BEGIN; … COMMIT;`
      geklammert, hat keine `.down.sql` und läuft auf einer Datenbank mit
      verwaisten Favoriten fehlerfrei durch.
- [x] `make rebuild-projections` läuft nach der Migration fehlerfrei durch.
- [x] Unit-Tests decken ab: alle Favoriten auflösbar, ein Favorit fehlt, alle
      Favoriten fehlen (Ergebnis: leere Liste, kein Fehler).

---

## Phase 2: Ladefehler sind im Service-Bereich als Fehler sichtbar

### Kontext

- `frontend/src/service/table/hooks.ts — useMeineTischeState()`,
  `useAktiveTischeMitFavoriten()`, `useEigeneUebersicht()` — geben `isError` und
  `refetch` nicht nach außen.
- `frontend/src/service/TableSelectionPage.tsx — TableSelectionPage()` — rendert
  bei leerer Liste den `EmptyState` „Keine Tische markiert", unabhängig davon,
  ob die Query erfolgreich war.
- `frontend/src/service/TablePage.tsx — TablePage()` — enthält bereits das
  Zielmuster (`stateError || historieError` → `LadefehlerAlert`).
- `frontend/src/components/common/LadefehlerAlert.tsx — LadefehlerAlert()`.
- `frontend/src/service/components/EigeneUebersicht.tsx — EigeneUebersichtKarten()`
  — rendert Anzahl und Betrag; kennt heute nur „lädt" und „fertig".
- `frontend/src/service/product/hooks.ts — useAktiveProdukte()` und
  `frontend/src/service/direktverkauf/hooks.ts — useDirektverkaufHistorie()` —
  dasselbe Muster, gleiche Wirkung (leeres Sortiment bzw. leere Historie).

### Was gebaut wird

Jede Query im Service-Bereich reicht ihren Fehlerzustand samt `refetch` an die
Seite durch; die Seiten unterscheiden sichtbar zwischen „nichts vorhanden" und
„konnte nicht geladen werden".

Auf `TableSelectionPage` ersetzt ein `LadefehlerAlert` mit dem Titel
„Tische konnten nicht geladen werden" den Leerzustand, sobald eine der drei
Queries fehlgeschlagen ist. Der „Erneut versuchen"-Button löst den Refetch aller
fehlgeschlagenen Queries der Seite aus. Solange ein Ladefehler ansteht, werden
die Übersichtskarten (`EigeneUebersichtKarten`) nicht mit Nullwerten gerendert —
`0 · 0,00 €` ist als Fehlerdarstellung ebenso irreführend wie der leere
Tischbereich.

Dieselbe Behandlung erhalten die Produktliste im Bestellen-Tab und die
Direktverkaufs-Historie.

### Akzeptanzkriterien

- [x] `useMeineTischeState`, `useAktiveTischeMitFavoriten`, `useEigeneUebersicht`,
      `useAktiveProdukte` und `useDirektverkaufHistorie` geben je `isError` und
      `refetch` zurück.
- [x] Schlägt eine der drei Queries der Tischübersicht fehl, zeigt die Seite den
      `LadefehlerAlert` und **nicht** „Keine Tische markiert".
- [x] Der „Erneut versuchen"-Button der Tischübersicht stößt den Refetch an;
      nach erfolgreichem Refetch verschwindet der Alert und die Tische werden
      angezeigt.
- [x] Bei anstehendem Ladefehler zeigen die Übersichtskarten keine
      Null-Beträge.
- [x] Erfolgreiche Query mit tatsächlich null Favoriten zeigt weiterhin
      unverändert den Leerzustand „Keine Tische markiert".
- [x] Schlägt das Laden der Produkte fehl, zeigt der Bestellen-Tab einen
      Ladefehler statt einer leeren Produktliste.
- [x] Schlägt das Laden der Direktverkaufs-Historie fehl, zeigt die Seite einen
      Ladefehler statt einer leeren Historie.
- [x] Komponententests decken für die Tischübersicht ab: Fehlerfall,
      Leerfall und Erfolgsfall.

---

## Phase 3: Hängende und abgebrochene Requests scheitern sauber

### Kontext

- `frontend/src/lib/Backend.ts — Backend.request()` — ruft `fetch` ohne
  `AbortSignal`; eine im WLAN hängende Verbindung wird nie abgebrochen, die
  Query bleibt dauerhaft `isPending` und wird nie wiederholt.
- `frontend/src/lib/Backend.ts — Backend.post()`, `Backend.throwIfNotOk()` —
  `response.json()` und `response.text()` laufen ungeschützt; ein abgebrochener
  Body wirft einen rohen `SyntaxError`/`TypeError` statt eines Fehlers des
  Backend-Clients.
- `frontend/src/lib/Backend.ts — BackendError` — trägt bereits `referenz` aus
  dem `X-Correlation-ID`-Header.
- `frontend/src/lib/errorMessages.ts — getActionErrorMessage()` — bildet Codes
  auf deutsche Texte ab.
- `frontend/src/lib/queryClient.ts — createQueryClient()` — zeigt den
  Query-Fehler-Toast ohne Referenz.

### Was gebaut wird

Ein Request, der nicht innerhalb von 8 Sekunden antwortet, wird clientseitig
abgebrochen und als Verbindungsfehler gemeldet — statt die Oberfläche dauerhaft
im Skeleton stehen zu lassen.

Dazu bekommt der Backend-Client eine neue Fehlerklasse `NetzwerkFehler` für
alle Fälle, in denen keine auswertbare Antwort zustande kam: Timeout,
Verbindungsabbruch und unvollständig übertragener Body. Sie trägt eine
Unterscheidung zwischen „Zeitüberschreitung" und „Verbindung unterbrochen",
damit die Meldung an den Helfer konkret bleibt.

`getActionErrorMessage` erhält für `NetzwerkFehler` einen eigenen Text
(„Keine Verbindung zum Server. Bitte WLAN prüfen und erneut versuchen."), und
der Query-Fehler-Toast zeigt die Korrelations-ID an, sofern eine vorliegt —
ohne sie ist eine gemeldete Störung im Backend-Log nicht auffindbar.

Ergänzend wird der 401-Pfad entschärft: Die Weiterleitung auf `/login` erfolgt
höchstens einmal, auch wenn mehrere parallele Queries gleichzeitig 401
erhalten.

### Akzeptanzkriterien

- [x] Jeder Request des Backend-Clients trägt ein Abbruch-Signal mit 8 Sekunden
      Zeitlimit; ein hängender Request endet als `NetzwerkFehler` vom Typ
      Zeitüberschreitung.
- [x] Ein abgebrochener oder unvollständiger Antwort-Body führt zu einem
      `NetzwerkFehler`, nicht zu einem rohen `SyntaxError`.
- [x] Eine syntaktisch gültige, aber schema-verletzende Antwort führt weiterhin
      zu `ResponseBodyError` (unverändertes Verhalten).
- [x] `getActionErrorMessage` liefert für `NetzwerkFehler` einen
      verbindungsbezogenen deutschen Text, nicht den Serverfehler-Text.
- [x] Der Query-Fehler-Toast zeigt die Korrelations-ID, wenn der Fehler eine
      trägt, und bleibt ohne Referenz unverändert.
- [x] Mehrere gleichzeitig mit 401 scheiternde Queries lösen genau eine
      Weiterleitung auf `/login` aus.
- [x] Unit-Tests in `frontend/src/lib/Backend.test.ts` decken Timeout,
      abgebrochenen Body und den unveränderten Schema-Fehlerpfad ab.

---

## Phase 4: Wiederholversuche und Aktualität gezielt steuern

### Kontext

- `frontend/src/lib/queryClient.ts — createQueryClient()` — setzt keine
  `defaultOptions`; es gelten die Vorgaben von `@tanstack/react-query`:
  drei Wiederholungen für **jeden** Fehler und `staleTime: 0`.
- `frontend/src/lib/Backend.ts — BackendError`, `ResponseBodyError` und der in
  Phase 3 ergänzte `NetzwerkFehler` — die Fehlerklassen, an denen die
  Wiederhol-Entscheidung hängt.
- `frontend/src/lib/queryClient.test.ts` — bestehende Testdatei.

### Was gebaut wird

Wiederholt wird nur, was Aussicht auf Erfolg hat: Netzwerkfehler und
Serverfehler ab Status 500. Ein `BackendError` mit 4xx (Validierung, fehlende
Berechtigung, Konflikt) und ein `ResponseBodyError` werden sofort gemeldet — sie
werden beim zweiten Versuch nicht plötzlich gelingen und verzögern die
Fehlermeldung nur.

Zusätzlich bekommen Queries eine Aktualitätsschwelle von 30 Sekunden. Heute
feuert jedes Entsperren des Handys sämtliche montierten Queries neu; bei
20–30 Helfern ist das der größte Teil des Traffics im Vereins-WLAN. Für
Situationen, in denen Aktualität wichtiger ist als Sparsamkeit, bleibt das
Verhalten überschreibbar — der Tisch-Zustand nach einer Buchung wird
weiterhin über `refetch` bzw. `invalidateQueries` sofort aktualisiert und
ist von der Schwelle nicht betroffen.

### Akzeptanzkriterien

- [x] Ein `BackendError` mit Status 4xx wird nicht wiederholt.
- [x] Ein `ResponseBodyError` wird nicht wiederholt.
- [x] Ein `NetzwerkFehler` und ein `BackendError` mit Status ab 500 werden
      wiederholt, höchstens zweimal, mit wachsendem Abstand.
- [x] Queries tragen eine Standard-Aktualitätsschwelle von 30 Sekunden.
- [x] Ein `refetch` bzw. `invalidateQueries` nach einer Buchung lädt trotz der
      Schwelle sofort neu — der Tisch-Saldo nach dem Kassieren ist unverändert
      aktuell.
- [x] Tests in `frontend/src/lib/queryClient.test.ts` decken die
      Wiederhol-Entscheidung je Fehlerklasse ab.

---

## Phase 5: Offline-Zustand ist sichtbar

### Kontext

- `frontend/src/App.tsx — App()` — globale Shell mit `ThemeProvider`,
  `TooltipProvider` und `Toaster`; der Ort für ein app-weites Banner.
- `frontend/src/lib/queryClient.ts — createQueryClient()` — der zentrale
  Query-Fehler-Toast, der bei Funkloch für jede Query feuern würde.
- `@tanstack/react-query` — `onlineManager` als Quelle des Online-Zustands.

### Was gebaut wird

Ein schmales, dauerhaft sichtbares Banner am oberen Rand meldet
„Keine Verbindung — Änderungen sind gerade nicht möglich", solange das Gerät
offline ist, und verschwindet bei Rückkehr der Verbindung von selbst.

Solange der Offline-Zustand aktiv ist, unterdrückt der zentrale Query-Fehler-
Toast seine Meldung: Die Ursache steht bereits im Banner, und ein zusätzlicher
Toast über einen „Serverfehler" ist in dieser Lage schlicht falsch.

Das Banner ist reine Anzeige — es blockiert keine Bedienung und puffert keine
Buchungen. Ein Reload im Funkloch führt weiterhin zu einer leeren Seite; das ist
eine bewusst offen gelassene Lücke (siehe „Getroffene Entscheidungen").

### Akzeptanzkriterien

- [x] Geht das Gerät offline, erscheint das Banner ohne Interaktion innerhalb
      einer Sekunde.
- [x] Kehrt die Verbindung zurück, verschwindet das Banner ohne Interaktion.
- [x] Das Banner erscheint in allen Bereichen (Service, Admin, Login).
- [x] Bei aktivem Offline-Zustand erzeugen fehlschlagende Queries keinen
      zusätzlichen Fehler-Toast.
- [x] Bei bestehender Verbindung erscheint der Query-Fehler-Toast unverändert.
- [x] Das Banner überdeckt keine Bedienelemente des Service-Bereichs (Kopfzeile
      und Fußleiste bleiben erreichbar).

---

## Phase 6: Rate-Limit-Antworten folgen dem Fehlerformat der API

### Kontext

- `backend/api/middleware/middleware.go — RateLimitMiddleware()` — antwortet mit
  `http.Error(w, "Rate limit exceeded", 429)`, also Plain Text.
- `backend/api/helper/http.go — SendJSONResponse()`, `errorResponse` — das
  kanonische Format `{ "code": ... }`.
- `frontend/src/lib/Backend.ts — Backend.throwIfNotOk()` — kann die
  Plain-Text-Antwort nicht parsen und erzeugt `code = "unknown"`.
- `frontend/src/lib/errorMessages.ts — getActionErrorMessage()` — bildet
  `"unknown"` auf „unerwarteter Serverfehler … Administrator kontaktieren" ab.
- `backend/api/middleware/middleware_test.go` — bestehende Tests der Middleware.

### Was gebaut wird

Der Rate-Limiter antwortet mit dem gleichen JSON-Fehlerformat wie jeder andere
Endpunkt: Status 429 mit `{ "code": "rate_limited" }`. Damit trifft der Helfer
auf eine wahre Aussage — der Server ist kurzzeitig überlastet, nicht defekt.

Das Frontend bildet `rate_limited` auf einen entsprechenden Text ab
(„Zu viele Anfragen in kurzer Zeit. Bitte einen Moment warten und erneut
versuchen.").

Betroffen sind die Bereiche `auth`, `relay` und — nur in der E2E-Umgebung —
`test`; die Limit-Werte selbst bleiben unverändert (LAN-Betrieb, jedes Gerät
hat eine eigene IP).

### Akzeptanzkriterien

- [x] Eine rate-limitierte Anfrage liefert Status 429, Content-Type
      `application/json` und den Body `{"code":"rate_limited"}`.
- [x] Der Backend-Client erzeugt daraus einen `BackendError` mit
      `code === "rate_limited"`, nicht `"unknown"`.
- [x] Das Frontend zeigt für `rate_limited` den Wartetext, nicht den
      Serverfehler-Text.
- [x] Ein Test in `backend/api/middleware/middleware_test.go` pinnt Status,
      Content-Type und Body der Limit-Antwort.

---

## Phase 7: Alle buchenden Vorgänge sind idempotent

### Kontext

- `backend/api/kasse/tischgeschaeft/application/command.go — Command.BestellungAufnehmen()`
  — Referenz-Implementierung: client-erzeugter Schlüssel, Duplikat wird zur
  stillen Erfolgsantwort.
- `backend/api/kasse/tischgeschaeft/application/command.go — Command.ZahlungKassieren()`
  (ein Event), `Command.StornierungErteilen()` (ein `bestellung-korrigiert` plus
  je betroffener Zahlung ein `stornierung-erteilt`), `Command.BestellungUmbuchen()`
  (zwei Events, Quelle und Ziel).
- `backend/api/kasse/direktverkauf/application/command.go — Command.DirektverkaufStornieren()`,
  `Command.persistVerkaufEvent()` — ein Event, Schreibpfad über `WriteEvent()`
  bzw. `WriteEventWithDruckauftraege()`.
- `backend/repository/kassenjournal_repo/repo.go — Repository.WriteEvent()`,
  `Repository.WriteEventWithDruckauftraege()`,
  `Repository.WriteTischSessionEventsAtomic()`, `Repository.WriteUmbuchung()` —
  die Schreibpfade der vier Vorgänge; alle nutzen `db.WithTx`.
- `backend/api/kasse/tischgeschaeft/http/command_handler.go` — Request-Structs
  und `zog`-Schemas der Tisch-Endpunkte;
  `backend/api/kasse/direktverkauf/http/command_handler.go` — dasselbe für den
  Direktverkauf.
- `frontend/src/service/components/table/BestellungAbschluss.tsx — BestellungAbschluss()`
  — Referenz für den Lebenszyklus des Schlüssels im UI: neu beim Beginn einer
  Zusammenstellung und nach jedem Erfolg, unverändert über einen Wiederholversuch.
- Die anzupassenden Abschluss-Komponenten:
  `frontend/src/service/components/table/ZahlungAbschluss.tsx — ZahlungAbschluss()`,
  `frontend/src/service/components/table/HistorieStornierungDrawer.tsx — HistorieStornierungDrawer()`,
  `frontend/src/service/components/table/HistorieUmbuchungDrawer.tsx — HistorieUmbuchungDrawer()`,
  `frontend/src/service/components/direktverkauf/DirektverkaufStornoDrawer.tsx — DirektverkaufStornoDrawer()`.
- `frontend/src/service/table/Zahlung.ts — ZahlungKassierenSchema`,
  `frontend/src/service/table/Stornierung.ts — StornierungErteilenSchema`,
  `frontend/src/service/table/Umbuchung.ts — BestellungUmbuchenSchema`,
  `frontend/src/service/direktverkauf/Direktverkauf.ts — DirektverkaufStornierenSchema`.
- `frontend/src/service/direktverkauf/DirektverkaufBackend.ts — DirektverkaufBackend`
  — ruft `serviceleitung/direktverkauf-stornieren`, den vierten buchenden
  Endpunkt ohne Idempotenz-Schlüssel.
- `database/migrations/01_initial.up.sql` — Muster der bestehenden
  Idempotenz-Indexe.

### Was gebaut wird

Heute gilt: Bricht die Verbindung ab, **nachdem** der Server eine Zahlung
gebucht hat, sieht der Helfer einen Fehler. Tippt er erneut, scheitert der
zweite Versuch mit „Mindestens eine Position ist nicht mehr bezahlbar" — er hat
kassiert und glaubt, es zweimal nicht geschafft zu haben. Doppelbuchungen
verhindert die Domänenprüfung zwar zuverlässig; die Rückmeldung ist trotzdem
falsch.

Der Client erzeugt deshalb pro fachlichem Vorgang eine `vorgangId` und sendet
sie mit. Der Server hält sie in der neuen Tabelle `vorgang_idempotenz` fest,
und zwar **im selben Commit** wie die Events des Vorgangs und **vor** deren
Insert. Trifft ein Wiederholversuch mit derselben `vorgangId` ein, kollidiert
der Primärschlüssel, und der Server antwortet mit Erfolg, ohne ein zweites Mal
zu buchen. Weil die Idempotenz-Zeile zuerst geschrieben wird, bleibt ein echter
OCC-Konflikt (fremde `vorgangId`, veraltete Stream-Version) eindeutig als 409
unterscheidbar.

Der vierte buchende Endpunkt ohne Schlüssel, `serviceleitung/direktverkauf-stornieren`,
wird gleich mitgezogen. Er hat dieselbe Schwäche, nutzt denselben Mechanismus
und kostet nur einen weiteren `art`-Wert — ihn auszulassen hieße, einen von vier
strukturgleichen Geldpfaden ungehärtet zu lassen. (Das Buchen selbst —
`bestellung-aufnehmen` und `direktverkauf-taetigen` — ist über die bestehenden
partiellen Unique-Indexe bereits abgesichert und bleibt unverändert.)

Migration `08_vorgang_idempotenz.up.sql` legt die Tabelle an. Die `art`-Spalte
hält den Vorgangstyp (`zahlung`, `stornierung`, `umbuchung`,
`direktverkauf-stornierung`) für die Nachvollziehbarkeit im Support.

Im Frontend erhalten die Abschluss-Komponenten der vier Vorgänge denselben
Schlüssel-Lebenszyklus wie `BestellungAbschluss`: eine neue UUID, sobald eine
Auswahl aus dem Leerzustand beginnt und nach jedem erfolgreichen Abschluss —
unverändert, solange derselbe Vorgang wiederholt wird.

Die Event-JSON-Contracts bleiben unangetastet: `vorgangId` ist kein Event-Feld.

### Akzeptanzkriterien

- [x] Migration `08_vorgang_idempotenz.up.sql` legt `vorgang_idempotenz`
      (`vorgang_id` UUID Primärschlüssel, `art`, `user_id`, `created_at`) an,
      transaktional geklammert, ohne `.down.sql`.
- [x] `service/zahlung-kassieren`, `serviceleitung/stornierung-erteilen`,
      `service/bestellung-umbuchen` und `serviceleitung/direktverkauf-stornieren`
      verlangen ein `vorgangId` im UUID-Format; eine Anfrage ohne oder mit
      ungültigem `vorgangId` wird mit `validation_error` abgelehnt.
- [x] Zweimaliges Senden derselben Kassier-Anfrage mit identischer `vorgangId`
      erzeugt genau ein `zahlung-kassiert:v1`-Event, genau einen
      Signaturauftrag und beide Male eine Erfolgsantwort.
- [x] Dasselbe gilt für Stornieren (die Event-Anzahl des Vorgangs bleibt
      unverändert, auch wenn er mehrere Events umfasst), für Umbuchen
      (genau ein Quell- und ein Ziel-Event) und für die
      Direktverkauf-Stornierung.
- [x] Ein echter OCC-Konflikt — neue `vorgangId`, veraltete Stream-Version —
      liefert weiterhin 409 mit `conflict` und schreibt keine Events.
- [x] Die Idempotenz-Zeile und die Events eines Vorgangs werden atomar
      geschrieben: Schlägt der Event-Insert fehl, existiert danach keine
      `vorgang_idempotenz`-Zeile für diesen Vorgang.
- [x] Der Guard-Test `backend/domain/kasse/event_json_contract_test.go` läuft
      unverändert durch — kein Event-Typ und kein JSON-Schlüssel hat sich
      geändert.
- [x] Im Frontend behält ein Wiederholversuch desselben Vorgangs seine
      `vorgangId`; nach erfolgreichem Abschluss wird eine neue erzeugt.
- [x] `make rebuild-projections` läuft nach der Migration fehlerfrei durch.

---

## Phase 8: Große Downloads werden nicht mehr abgeschnitten

### Kontext

- `backend/app/app.go — NewApp()` — setzt `WriteTimeout: 10 * time.Second` für
  den gesamten Server; die Frist gilt für das vollständige Schreiben jeder
  Antwort.
- `backend/api/fiskal/export/http/handler.go — ExportHandler()` — streamt das
  DSFinV-K-ZIP einer Kassensitzung und läuft damit gegen ebendiese Frist.
- Im gesamten Backend wird nirgends `http.ResponseController` verwendet, die
  Schreibfrist wird also nie verlängert.

### Was gebaut wird

Der Export-Handler verlängert seine eigene Schreibfrist über
`http.ResponseController`, bevor er mit dem Streamen beginnt. Der knappe
10-Sekunden-Wert bleibt für alle übrigen Endpunkte erhalten — er ist dort
richtig und schützt vor hängenden Verbindungen.

Ohne diese Änderung wird ein Export, dessen Übertragung länger als zehn
Sekunden dauert, serverseitig abgeschnitten; der Betreiber erhält ein
unvollständiges Archiv, ohne dass ein Fehler gemeldet wird. Bei
aufbewahrungspflichtigen Daten ist das der schlechteste denkbare Ausgang.

### Akzeptanzkriterien

- [x] Der Export-Handler setzt vor dem ersten Schreibvorgang eine großzügigere
      Schreibfrist (mindestens fünf Minuten).
- [x] Ein Export, dessen Übertragung mehr als zehn Sekunden dauert, kommt
      vollständig und als gültiges ZIP-Archiv beim Client an.
- [x] Die Schreibfrist aller übrigen Endpunkte bleibt bei zehn Sekunden.
- [x] Kann die Frist nicht gesetzt werden, wird das protokolliert und der
      Export läuft trotzdem — die Verlängerung ist eine Verbesserung, kein
      Abbruchgrund.
