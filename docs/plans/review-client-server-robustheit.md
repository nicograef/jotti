# Review: Client-Server-Robustheit (PR #103)

> Gegenstand: Branch `claude/jotti-client-server-issues-wrcxx8` gegen `origin/main`,
> 80 Dateien, +4426/-251. Grundlage der Änderungen ist
> `docs/plans/plan-client-server-robustheit.md`.
> Durchgeführt am 2026-07-27.

## Urteil

Nicht mergen wie er ist. Der PR verbessert Transportverhalten und Fehlerzustände
messbar, führt aber ein globales 8-Sekunden-Zeitlimit ein, das auch für
`admin/tse-einrichten` gilt und dort eine bezahlte LIVE-TSS unbrauchbar
zurücklassen kann. Zusätzlich ändert die späte `useVorgangId`-Umstellung den
Schlüssel-Lebenszyklus zweier Endpunkte, die nie Teil von Phase 7 waren, und
erzeugt dort einen neuen Doppelbuchungspfad.

Ein Blocker, vier schwerwiegende Befunde, acht geringfügige, drei
Cleanup-Punkte. `make check` läuft lokal grün.

## Methode

15 Experten-Linsen über den Diff (Idempotenz-Transaktion, Idempotenz-Semantik,
Favoriten-Pfad, Middleware und Export, Backend-Client-Transport, Query-Policy
und Offline, `useVorgangId`, Fehlerzustände, Plan gegen Code, Testqualität
Backend und Frontend, Freeze und Compliance, Scope und tote Pfade,
Betriebssicht, Build und Generierung), dazu fünf Cleanup-Durchläufe nach dem
Cleanup-Skill.

Jeder Rohbefund wurde von drei unabhängigen Widerlegern mit verschiedenen Linsen
angegriffen (Erreichbarkeit, bestehende Absicherung, Belegtreue); bestätigt gilt
ein Befund erst, wenn höchstens einer der drei ihn widerlegen konnte. 51
Rohbefunde, 28 bestätigt. 21 Cleanup-Vorschläge, 3 bestätigt.

Blocker und die vier schwerwiegenden Befunde wurden anschließend von Hand am
Quellcode nachgeprüft.

## Blocker

### Das 8-Sekunden-Zeitlimit storniert die TSE-Einrichtung mitten im Lebenszyklus

`frontend/src/lib/Backend.ts:151-177` setzt das Zeitlimit unkonditional für
jeden Endpunkt. `frontend/src/admin/tse/TSEBackend.ts:219-228` ruft darüber
`admin/tse-einrichten`. Serverseitig reicht
`backend/api/fiskal/setup/http/command_handler.go:120` `r.Context()` an
`RichteTSEEin` durch, und `backend/repository/tse_repo/fiskaly_client.go:464`
baut jeden fiskaly-Aufruf mit `http.NewRequestWithContext`. Der Client-Abbruch
storniert damit den laufenden Aufruf und alle folgenden.

`backend/api/fiskal/setup/application/setup.go:67-140` fährt sequenziell
`ListTSS`, `CreateTSS`, `vollendeLebenszyklus` (vier Aufrufe) und erst danach
`saveEinrichtung`. Das sind sieben bis neun HTTPS-Roundtrips zu fiskaly, je mit
10 s Timeout und bis zu drei Wiederholungen mit Backoff
(`fiskaly_client.go:22-23,458-524`).

Ablauf: Der Admin startet die Einrichtung in LIVE. Nach 8 s bricht der Browser
ab, Go storniert den Request-Kontext, `saveEinrichtung` läuft nie. Bei fiskaly
steht dann eine bezahlte TSS, deren zufällige Admin-PIN nur im verlorenen
Response-Body existierte und deren PUK nur im Zustand `CREATED` abrufbar war.
Der zweite Versuch scheitert an `hatAktiveTSS` (`setup.go:455`) mit
`tse_bereits_eingerichtet`, die Übernahme (`setup.go:290`) an fehlender PIN. Die
Instanz ist ohne fiskaly-Dashboard oder fiskaly-Support nicht mehr legal
betreibbar.

Auf `main` gibt es keinen AbortController, und `WriteTimeout: 10 * time.Second`
(`backend/app/app.go:32`) setzt nur eine Schreibfrist auf der Verbindung, ohne
`r.Context()` zu stornieren. Der Handler lief dort zu Ende und speicherte. Der
PR verschlechtert diesen Pfad also eindeutig.

- [x] Zeitlimit pro Aufruf konfigurierbar machen (optionaler Parameter an
      `BackendClient.post` und `BackendClient.download`, Standard 8000 ms).
- [x] `admin/tse-einrichten`, `admin/tse-uebernehmen`,
      `admin/test-tse-verbindung` und den DSFinV-K-Export auf ein deutlich
      höheres Limit ziehen.

## Schwerwiegend

### Nutzdaten-gebundener Schlüssel doppelbucht Bestellung und Direktverkauf

`frontend/src/hooks/use-vorgang-id.ts:25` rotiert den Schlüssel bei jeder
Nutzdaten-Änderung. `BestellungAbschluss.tsx:64` und
`DirektverkaufAbschluss.tsx:74` binden ihn an `positionen` und `kommentar`. Auf
`main` blieb der Schlüssel über eine ganze Zusammenstellung stabil und wechselte
nur beim Übergang Leerzustand zu Auswahl (`warLeerRef`).

Beide Endpunkte waren nie Teil von Phase 7. Sie deduplizieren über die
partiellen Unique-Indexe auf dem Event-JSON
(`database/migrations/01_initial.up.sql:161-168`), also rein über die ID, und
hängen nur an; es gibt keine Domäneninvariante, die eine Zweitbuchung abfängt.

Ablauf: Zwei Bier im Korb, „Verkauf abschließen“. Der Server committet, die
Antwort geht verloren, der Client bricht nach 8 s ab. Der Gast bestellt eine
Cola nach, der Helfer tippt sie ein und schließt erneut ab. Geänderte Nutzdaten,
neuer Schlüssel, zweites `direktverkauf-getaetigt:v1` mit eigener
TSE-Signatur. Im Journal stehen zwei Bier plus zwei Bier und Cola, kassiert
wurde einmal. Das ergibt einen Fehlbetrag im Kassensturz und einen
Phantomumsatz im DSFinV-K-Export.

Ein Revert ist die falsche Antwort: Auf `main` verschluckte der Server die
geänderte Einreichung und meldete Erfolg, ohne die Cola zu buchen. Beide
Varianten sind falsch. Die Bindung gehört auf die Serverseite.

- [x] Spalte `payload_hash` in `vorgang_idempotenz` (additive Migration `08`).
      Gleicher Schlüssel und gleicher Hash ergibt die stille Erfolgsantwort,
      gleicher Schlüssel und anderer Hash einen expliziten Fehlercode statt
      eines stillen Erfolgs.
      *Umgesetzt direkt in Migration `07` statt in einer `08`: `07` ist neu auf
      diesem Branch und lief auf keiner Instanz, der Freeze schützt hier also
      keine persistierten Daten. Begründung im Nacharbeitsplan unter
      „Datenbank".*
- [x] Äquivalenten Abgleich für `bestellung-aufnehmen` und
      `direktverkauf-taetigen`, deren Schlüssel im Event-JSON liegt.

### TSE-Live-Suite scheitert ab sofort im Teardown

`backend/api/fiskal/tse_live/tse_live_suite_test.go:124-138`: Die
Statement-Liste von `cleanLiveDB` endet mit `DELETE FROM users` und enthält
kein `DELETE FROM vorgang_idempotenz`. Die Suite schreibt über die neuen
`uuid.NewString()`-Argumente (Zeilen 383, 430, 440, 453, 466, 475, 500) sieben
Idempotenz-Zeilen mit `user_id` des Testbenutzers.
`database/migrations/07_vorgang_idempotenz.up.sql:17` deklariert
`user_id INT NOT NULL REFERENCES users(id)` ohne `ON DELETE`, das Löschen
bricht also mit 23503 ab und `t.Fatalf` meldet den Lauf als FAIL, obwohl alle
fiskalischen Assertions grün sind.

Die Schwestersuiten wurden korrekt nachgezogen
(`backend/api/kasse/tischgeschaeft/application/idempotenz_integration_test.go:35`,
`backend/api/kasse/direktverkauf/application/idempotenz_integration_test.go:34`),
nur diese nicht. Kein Produktionsrisiko, aber `make test-tse-live` ist der
einzige Ende-zu-Ende-Nachweis, dass jeder signaturpflichtige Vorgang signiert
wird, und läuft nicht in CI.

- [x] `"DELETE FROM vorgang_idempotenz",` vor `"DELETE FROM users",` einfügen.

### Aktualitätsschwelle 30 s ohne Invalidierung des Tischzustands

`frontend/src/lib/queryClient.ts:16,47` setzt global `staleTime: 30_000`; auf
`main` gab es keine `defaultOptions`, also `staleTime: 0` und einen Refetch bei
jedem Mount. Eine repo-weite Suche nach `invalidateQueries` findet für
`['tisch-state', id]` und `['tisch-historie', id]`
(`frontend/src/service/table/hooks.ts:29,50`) keinen einzigen Treffer im
Produktionscode. `frontend/src/service/TablePage.tsx:127-135` invalidiert nur
die drei Übersichts-Schlüssel und lädt den gerade offenen Tisch neu.

Ablauf: Helfer A öffnet Tisch 5 (Saldo 10 EUR) und geht zurück. Helfer B bucht
dort zwei Bier (+7 EUR). Innerhalb von 30 s öffnet A Tisch 5 erneut, sieht
10 EUR und die alten Positionen, kassiert 10 EUR. Der Server akzeptiert das, der
Gast geht, 7 EUR bleiben offen. `command.go:429-435` verhindert nur
Doppelkassierungen bereits bezahlter Positionen, nicht die Unterkassierung.

Zweite Ausprägung: Nach einer Umbuchung wird der Ziel-Tisch nirgends
invalidiert. Öffnet der Helfer ihn innerhalb von 30 s, zeigt er den Stand vor
der Umbuchung, während die invalidierte Übersichtskarte desselben Tisches
bereits den neuen Saldo führt.

- [x] `staleTime: 0` für die Queries, die geteilten Tischzustand abbilden
      (`useTischState`, `useTischHistorie`, `useMeineTischeState`); die 30 s auf
      Stammdaten beschränken, wo die Traffic-Ersparnis anfällt.
- [x] In `TablePage.reload` zusätzlich die Präfixe `['tisch-state']` und
      `['tisch-historie']` invalidieren. Inaktive Queries werden dabei nur als
      veraltet markiert und kosten keine zusätzlichen Requests.

### Fehlgeschlagener Hintergrund-Refetch verwirft gültige Cache-Daten

Die neuen Fehlerzweige prüfen nur `isError`:
`frontend/src/service/TableSelectionPage.tsx:85`,
`frontend/src/service/components/table/Bestellung.tsx:45`,
`frontend/src/service/components/direktverkauf/Direktverkauf.tsx:42`,
`frontend/src/service/components/direktverkauf/DirektverkaufHistorie.tsx:97`.
react-query setzt `status: 'error'` auch dann, wenn ein Hintergrund-Refetch
scheitert und die zuvor geladenen Daten unverändert im Cache liegen; genau dafür
unterscheidet die Bibliothek `isLoadingError` von `isRefetchError`.

Ablauf: Der Helfer hat einen Korb offen, das Handy wird gesperrt und über 30 s
später entsperrt. Der Focus-Refetch von `aktive-produkte` läuft in das
8-s-Limit und zwei Wiederholungen. Die vollständige Produktliste liegt im
Cache, der Bestellen-Tab zeigt trotzdem „Produkte konnten nicht geladen
werden“. Auf `TableSelectionPage` überspringt der frühe Return zusätzlich
Suchfeld, Favoritenliste, die feste Schaltfläche „Alle Tische“ und den
`TischAuswahlDrawer` (Zeile 175), also alle Einstiege in einen Tisch, auch wenn
nur die rein informative `eigene-uebersicht` gescheitert ist.
`ServiceLayout.tsx:15-27` bietet auf `/service/tische` keine
Alternativnavigation.

- [x] Fehlerzweig nur ziehen, wenn nichts Brauchbares vorliegt (`isLoadingError`
      statt `isError`); einen gescheiterten Refetch bei vorhandenen Cache-Daten
      nicht-blockierend melden.
- [x] Auf `TableSelectionPage` den Alert auf den Bereich beschränken, dessen
      Daten fehlen, und Fußzeilen-Button samt Drawer in jedem Fall gemountet
      lassen.

## Geringfügig

- [x] Das Zeitlimit endet mit dem Antwort-Header, nicht mit dem Body.
      `frontend/src/lib/Backend.ts:158-177`: `clearTimeout` steht im `finally`
      von `await fetch`, und fetch löst mit den Headern auf. Stockt die
      Verbindung mitten im Body, hängt die Query bis zum TCP-Timeout des
      Betriebssystems. Keine Regression, aber das abgehakte
      Phase-3-Kriterium gilt nur für die Header-Phase, und
      `Backend.test.ts:207-231` testet auch nur diese.
- [x] `abgeschlossene-kassensitzungen` wird nirgends invalidiert
      (`frontend/src/admin/reporting/hooks.ts:26`). Wer den Tagesabschluss bucht
      und binnen 30 s auf „Berichte & Export“ wechselt, sieht die neue Sitzung
      nicht; der Archiv-Download liefert dann die vorherige Sitzung.
      `ABGESCHLOSSENE_KASSENSITZUNGEN_KEY` und `REPORT_KEY` im Erfolgspfad des
      Tagesabschlusses invalidieren.
- [x] `frontend/src/admin/kasse/GeldtransitDialog.tsx:50` ist der siebte
      Idempotenz-Schlüssel und wurde als einziger nicht auf `useVorgangId`
      gezogen. Korrigiert der Admin nach einem Fehlversuch den Betrag, greift
      der Unique-Index, die UI meldet Erfolg und der korrigierte Betrag
      verschwindet. Vorbestehend, durch das neue 8-s-Limit aber
      wahrscheinlicher.
- [x] Die vier Zweige, die `ErrVorgangBereitsGebucht` aus der Schreibtransaktion
      auf eine stille Erfolgsantwort abbilden
      (`tischgeschaeft/application/command.go:153,384,554`,
      `direktverkauf/application/command.go:210`), sind von keinem Test
      erreichbar: Die Vorprüfung greift beim zweiten sequentiellen Aufruf immer
      zuerst. Genau dieser Zweig sichert laut Plan das Rennen zweier
      gleichzeitiger Anfragen ab. Repo-Spy ergänzen, der den Fehler aus
      `WriteEventMitVorgang` liefert, ohne dass die Vorprüfung anschlägt.
- [x] Drei der vier neuen Duplikat-Unit-Tests
      (`command_idempotenz_test.go:36,203`,
      `direktverkauf/application/command_test.go:456`) bleiben grün, wenn man
      `vorgangBereitsGebucht` aus allen vier Commands entfernt. Ursache:
      `mock.go:170-180` schreibt die Projektion nach einem erfolgreichen Write
      nicht fort. Die Integrationsstufe fängt den Regress, `make check` nicht.
- [x] `kommentar` wird in keinem der vier „wechselt die vorgangId“-Tests
      variiert; Löschen von `kommentar` aus allen sechs `useVorgangId`-Aufrufen
      lässt die Suite grün (`ZahlungAbschluss.test.tsx:184`). Die zentrale
      Behauptung von `ad17ab9` ist damit nur zur Hälfte abgesichert.
- [x] Der Retry-Test in `TableSelectionPage.test.tsx:119` prüft nur
      `reloadMeineTische`; ein `reload`, das die anderen beiden Queries nicht
      mehr neu lädt, bliebe grün.
- [x] Der auf Render-State-Sync umgeschriebene Leerzustands-Reset in
      `ZahlungAbschluss.tsx:64-73` ist ungetestet; Löschen des Blocks lässt 15
      Dateien mit 111 Tests grün. Die baugleichen Blöcke in
      `DirektverkaufAbschluss` und `BestellungAbschluss` sind abgedeckt.

## Cleanup

Nur verhaltenserhaltende Punkte; 3 von 21 Vorschlägen haben die Prüfung
überstanden.

- [x] `backend/repository/kassenjournal_repo/repo.go`: `WriteUmbuchung`
      (Zeile 279) und `WriteTischSessionEventsAtomic` (Zeile 194) haben mit dem
      Wechsel des `eventRepo`-Interfaces auf die `MitVorgang`-Varianten ihren
      letzten Produktionsaufrufer verloren, `MockRepo.WriteUmbuchung`
      (`mock.go:134`) hat gar keinen mehr. Löschen und die drei Testaufrufe
      (`repo_test.go:557,653,751`) auf `WriteUmbuchungMitVorgang` umstellen;
      danach kann `writeTischSessionEventsAtomic` `vorgang Vorgang` statt
      `*Vorgang` nehmen.
- [x] `backend/api/kasse/tischgeschaeft/application/command.go:310` (analog 416
      und 463): `if gebucht, err := …; err != nil { return } else if gebucht`
      ist die einzige Stelle im Nicht-Test-Backend, die `else` an einen
      returnenden Fehlerzweig hängt, und versteckt den Happy Path im
      else-Zweig. Auf die Form von
      `direktverkauf/application/command.go:154-163` ziehen.
- [x] `backend/api/kasse/direktverkauf/application/command.go:256`:
      `persistVerkaufEvent` ist nach dem Wegfall des zweiten Zweigs eine reine
      Weiterleitung mit einem Aufrufer, während die Schwesterstelle im selben
      File `writeVersionedEvent` direkt aufruft. Inlinen.

## Nicht bestätigt

Damit sie nicht erneut aufgeworfen werden.

- Das 8-s-Limit mache den DSFinV-K-Export unmöglich und hebe die neue
  5-Minuten-Schreibfrist auf. Falsch: `clearTimeout` läuft im `finally` direkt
  nach `await fetch`, also sobald die Header da sind; der Blob-Transfer läuft
  ohne Signal. Die beiden Fristen decken disjunkte Phasen ab, Phase 8 bleibt
  wirksam. Betroffen wäre nur eine Archiv-Erzeugung über 8 s, und die scheiterte
  wegen `WriteTimeout: 10s` auch vorher. Sechs Linsen haben diese Variante
  gemeldet, alle widerlegt.
- Das Offline-Banner schiebe die feste Service-Spalte unter den Sichtbereich.
  Nachgemessen bei 1024x768 und 1280x800: Der `calc`-Wert enthält das
  Bottom-Padding von `main` bereits, überstehen tun 16 px beziehungsweise 8 px,
  und das ist ausschließlich `DrawerFooter`-Padding. Kein Bedienelement wird
  beschnitten.
- Der `favorit_repo`-Mock mit `RemoveByTisch` ohne Produktions-Gegenstück sei
  totes Gerüst. Das Gegenstück ist die sqlc-Query `RemoveFavoritenByTisch`,
  aufgerufen in `tisch_repo/repo.go:158`; `repo_test.go:119,156` decken
  Transaktion und Rollback ab.
- `handbuch.md` und `verfahrensdokumentation.md` müssten `vorgang_idempotenz`
  nachziehen. Abschnitt 3.8 ist ein Projektions-Routing, kein
  Transaktions-Inventar; `tse_signaturauftraege` und `druckauftraege` stehen
  dort ebenfalls nicht, und die Idempotenz über partielle Unique-Indexe war
  auch vorher nicht dokumentiert.
- „Erneut versuchen“ gebe bis zu 27 s keine Rückmeldung. Nach einem
  gescheiterten Erstladen setzt query-core beim Refetch `status: 'pending'` und
  `error: null` zurück, der Alert weicht sofort dem Skeleton.
- Die Änderung des `responseWriter` sei eine riskante Grenzöffnung.
  `Unwrap() http.ResponseWriter` erweitert das Methodenset des Wrappers nicht;
  bestehende Typzusicherungen auf `http.Flusher` oder `http.Hijacker`
  verhalten sich unverändert. Nur `http.ResponseController` gewinnt Zugriff.

## Was dieses Review nicht abgedeckt hat

- Keine Messung realer Laufzeiten. Ob die TSE-Einrichtung über einen
  Vereins-Uplink tatsächlich 8 s überschreitet und wie lange die
  DSFinV-K-Erzeugung bei echtem Datenvolumen braucht, ist aus Timeouts und
  Retry-Budgets abgeleitet, nicht gemessen.
- Kein Test am laufenden Stack. Ob Caddy den Client-Abbruch in jeder
  Konstellation an den Go-Handler durchreicht, wurde nicht verifiziert.
- Keine Prüfung der DSFinV-K-Feldabbildung, der Print-Relay-Seite oder von
  Migrationen jenseits von `06` und `07`.
- Keine UI-Abnahme auf echten Geräten; die Layout-Aussagen stammen aus
  Headless-Messungen.
- Ausgeführt wurde in dieser Sitzung nur `make check` (grün). Weder
  `make verify` noch die Frontend-Suite noch `make test-tse-live` liefen; die
  Mutationsnachweise der Testbefunde stammen aus den Verifikationsläufen des
  Reviews.
