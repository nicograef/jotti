# Plan: Nacharbeit zum Review der Client-Server-Robustheit

> Quelle: `docs/plans/review-client-server-robustheit.md` (Review zu PR #103,
> Branch `claude/jotti-client-server-issues-wrcxx8`).
> Der Ursprungsplan `docs/plans/plan-client-server-robustheit.md` ist vollständig
> abgehakt; dieser Plan arbeitet ausschließlich die Review-Befunde nach.

## Ziel

Alle 16 Review-Befunde abarbeiten: den Blocker, die vier schwerwiegenden, die
acht geringfügigen und die drei Cleanup-Punkte. Nach diesem Plan ist der Branch
mergefähig.

Drei inhaltliche Schwerpunkte:

1. **Das globale 8-Sekunden-Zeitlimit darf keinen langlaufenden Vorgang mitten
   im Lebenszyklus abbrechen.** Betroffen ist vor allem `admin/tse-einrichten`:
   Ein Abbruch dort hinterlässt eine bezahlte LIVE-TSS, deren Admin-PIN und PUK
   verloren sind.
2. **Der Idempotenz-Schlüssel wird serverseitig an die Nutzdaten gebunden.**
   Heute entscheidet der Client, was „derselbe Vorgang" ist — beide bisherigen
   Varianten sind falsch: die Rotation bei Nutzdaten-Änderung (Branch) erzeugt
   Doppelbuchungen, der stabile Schlüssel ohne Serverprüfung (main) verschluckt
   geänderte Einreichungen. Alle sieben buchenden Endpunkte laufen künftig über
   einen Mechanismus.
3. **Geteilter Kassen- und Tischzustand ist nie veraltet, und ein
   fehlgeschlagener Hintergrund-Refetch verwirft keine gültigen Daten.**

## Architekturentscheidungen

Durchgängig gültig für alle Phasen.

### Idempotenz-Mechanismus

- **Eine Tabelle für alle buchenden Vorgänge:** `vorgang_idempotenz`
  (`vorgang_id UUID PRIMARY KEY`, `art TEXT NOT NULL CHECK`,
  `user_id INT NOT NULL REFERENCES users(id)`, `payload_hash BYTEA NOT NULL`,
  `created_at TIMESTAMPTZ NOT NULL`). Die Zeile entsteht im selben Commit wie
  die Events des Vorgangs und **vor** deren Insert.
- **`art` deckt sieben Werte ab:** `bestellung`, `zahlung`, `stornierung`,
  `umbuchung`, `direktverkauf`, `direktverkauf-stornierung`, `geldtransit`.
- **`payload_hash` ist ein SHA-256 über `art` und die Nutzdaten des Vorgangs**,
  gebildet in der Application-Schicht aus den Kommando-Argumenten (nicht aus dem
  rohen HTTP-Body — dessen Bytes sind zwischen zwei Einreichungen nicht
  garantiert identisch). Die Art ist Teil der Bindung, weil sich alle sieben
  Arten einen Schlüsselraum teilen und mehrere Kommandos dieselbe Feldmenge
  einreichen (Zahlung und Stornierung serialisieren byteidentisch); derselbe
  Schlüssel mit anderer Art ist damit kein Duplikat, sondern
  `VorgangDatenAbweichend`.
- **Drei Ausgänge der Prüfung**, in Go als `kassenjournal_repo.VorgangStatus`:
  - `VorgangNeu` — regulär buchen.
  - `VorgangDuplikat` (gleiche ID, gleicher Hash) — stille Erfolgsantwort ohne
    zweite Buchung.
  - `VorgangDatenAbweichend` (gleiche ID, anderer Hash) — HTTP 409 mit dem
    Fehlercode `vorgang_daten_abweichend`.
- **Die Prüfung läuft an zwei Stellen** mit derselben Funktion: als Vorprüfung
  vor der fachlichen Validierung und noch einmal, wenn der Insert in der
  Schreibtransaktion am Primärschlüssel scheitert (Rennen zweier gleichzeitiger
  Anfragen).
- **Die partiellen Unique-Indexe auf dem Event-JSON**
  (`idx_kassenjournal_bestellung_id`, `idx_kassenjournal_verkauf_id`,
  `idx_kassenjournal_geldtransit_id` in `01_initial.up.sql`) bleiben
  unverändert als zweite Absicherung. Für alle ab Migration `08` geschriebenen
  Vorgänge lösen sie nicht mehr aus, weil die Idempotenz-Zeile vor den Events
  entsteht. Für Bestandsdaten aus der Zeit davor gilt das nicht — siehe
  „Risiken".

### Datenbank

- **`payload_hash` wird direkt in `08_vorgang_idempotenz.up.sql` ergänzt**,
  ebenso die erweiterte `art`-CHECK-Liste — also direkt in der Migration, die
  die Tabelle anlegt, und nicht in einer nachgelagerten `09`.
  Begründung: v0.17.1 (produktiv) enthält nur die Migrationen `01`–`05`;
  `07_favoriten_cleanup.up.sql` und `08_vorgang_idempotenz.up.sql` sind beide
  neu auf diesem Branch und liefen auf keiner Instanz. Der Freeze schützt
  persistierte Daten auf echten Instanzen — hier gibt es keine. Damit ist
  `payload_hash BYTEA NOT NULL` ohne Nullable-Zwischenzustand und ohne einen
  Fallback-Zweig möglich, der nie erreicht würde.
- **Event-JSON-Contracts bleiben unberührt.** Kein Event-Feld kommt hinzu, kein
  Event wird umgedeutet; `event_json_contract_test.go` muss unverändert grün
  bleiben.

### Frontend-Transport

- **`BackendClient` bekommt ein Optionsobjekt als letzten Parameter:**
  `post<T>(endpoint, body, responseSchema?, optionen?)` und
  `download(endpoint, body, optionen?)` mit
  `interface RequestOptionen { zeitlimitMs?: number }`.
- **Zwei benannte Zeitlimits neben dem Standard:**
  `REQUEST_TIMEOUT_MS = 8000` (unverändert Standard),
  `TSE_TIMEOUT_MS = 150_000` für die vier fiskaly-Endpunkte,
  `EXPORT_TIMEOUT_MS = 330_000` für den DSFinV-K-Export.
- **Das Client-Budget ist das serverseitige Schreibbudget plus Netzreserve:**
  150 s gegen `tseSetupWriteTimeout = 2 Minuten`, 330 s gegen
  `exportWriteTimeout = 5 Minuten`, je 30 s Aufschlag. Der Aufschlag gilt der
  Antwort, nicht der Reihenfolge des Aufgebens: Ein spät, aber erfolgreich
  geschriebener Antwort-Body soll den Client noch erreichen. Wäre das
  Client-Budget nicht größer als das Schreibbudget des Servers, hätte der Client
  in genau dem Fenster schon aufgegeben, in dem der Server gerade noch schreibt.
  Eine abgelaufene Schreibfrist bricht keinen Handler ab — sie lässt nur einen
  laufenden Schreibvorgang scheitern; für die beiden entkoppelten TSE-Handler
  ist der einzige serverseitige Aufgabepunkt der Leck-Wächter (10 Minuten).
- **Das Zeitlimit umfasst das Lesen des Antwort-Bodys**, nicht nur die
  Header-Phase.

### Aktualität der Queries

- **Der Standard `staleTime` ist `0`; `30_000` ist die begründete Ausnahme.**
  Die Voreinstellung in `createQueryClient` entfällt; ein benanntes
  `STAMMDATEN_AKTUALITAET_MS = 30_000` wird ausschließlich auf den
  Stammdaten-Hooks gesetzt. Ein neuer Hook ist damit per Voreinstellung frisch,
  und die Traffic-Ersparnis ist eine bewusste Entscheidung pro Hook statt einer
  stillen Voreinstellung für alles.
- **Stammdaten** sind Daten, die sich während einer Veranstaltung nicht ändern:
  Produkte, Benutzer, Betreiber, Kassenidentität, Druckstationen,
  TSE-Konfiguration. Alles, was Kassen- oder Tischzustand liefert, ist es nicht —
  auch keine Tisch-Query, weil jede von ihnen `saldoCents` der laufenden
  Kassensitzung mitführt.

### HTTP

- **Alle API-Endpunkte bleiben POST-only**, keine neuen Routen.
- **Langlaufende Handler verlängern die Schreibfrist der Verbindung** — über
  einen gemeinsamen Helfer statt duplizierter
  `http.NewResponseController`-Blöcke. Weil die Frist eine absolute Zeit ab
  Request-Start ist und während der gesamten Handler-Laufzeit läuft, setzen sie
  sie zweimal: am Eingang für die frühen Fehlerpfade, und noch einmal
  unmittelbar vor dem Schreiben, damit der Schreibvorgang ein eigenes Budget
  hat. Die Frist ist kein Abbruchmechanismus: Sie beendet keinen Handler,
  sondern lässt nur einen Schreibvorgang scheitern, der nach ihrem Ablauf noch
  läuft.
- **Höchstens ein Schreiber auf der TSE-Konfiguration.** Seit der Lebenszyklus
  den Client überlebt, kann der Admin die Einrichtung ein zweites Mal starten,
  während die erste noch läuft — der zweite Lauf sähe ein leeres fiskaly-Konto
  und legte eine zweite, bezahlte TSS an. Ein prozessinternes Schloss in der
  Application-Schicht (jotti läuft als eine einzige Backend-Instanz) lehnt den
  zweiten Aufruf sofort mit HTTP 409 `tse_setup_laeuft_bereits` ab, statt ihn
  warten zu lassen. Dasselbe Schloss nimmt der manuelle Zugangsdaten-Wechsel
  (`UpdateTSEKonfiguration`): Er schreibt über denselben `SaveEinrichtung` und
  liegt in der Oberfläche direkt unter dem Wizard — ohne das Schloss gewänne
  der letzte Schreiber, und die Instanz signierte danach gegen eine
  TSS/Client-Kombination, die nicht die eingerichtete ist.

## Inventar

### Backend — Idempotenz

- `database/migrations/08_vorgang_idempotenz.up.sql` — Tabelle
  `vorgang_idempotenz`, aktuell ohne `payload_hash`, `art`-CHECK mit vier Werten.
- `backend/sqlc/queries/vorgang_idempotenz.sql` — `InsertVorgangIdempotenz`,
  `ExistsVorgangIdempotenz`.
- `backend/repository/kassenjournal_repo/repo.go` — `Vorgang`,
  `ErrVorgangBereitsGebucht`, `VorgangArt*`-Konstanten,
  `VorgangBereitsGebucht()`, `insertVorgangIdempotenzInTx()`,
  `WriteEventMitVorgang()`, `WriteTischSessionEventsAtomicMitVorgang()`,
  `WriteUmbuchungMitVorgang()`, `writeSingleEvent()`,
  `writeTischSessionEventsAtomic()`, `EventExistsByTypeAndVorgangsID()`.
- `backend/repository/kassenjournal_repo/mock.go` — `MockRepo`,
  `vorgangBereitsGebucht()`, `recordVorgang()`, `GebuchterVorgang()`,
  `WriteEventMitVorgang()`, `EventExistsByTypeAndVorgangsID()`. Die
  `WriteEvent`-Pfade des Mocks schreiben die Projektion **nicht** fort — Ursache
  des geringfügigen Testbefunds.
- `backend/api/kasse/tischgeschaeft/application/command.go` —
  `vorgangBereitsGebucht()`, `persistTischEvent()`, `BestellungAufnehmen()`,
  `BestellungUmbuchen()`, `ZahlungKassieren()`, `StornierungErteilen()`,
  `writeEventWithDruckauftraege()`.
- `backend/api/kasse/direktverkauf/application/command.go` —
  `DirektverkaufTaetigen()`, `DirektverkaufStornieren()`,
  `persistVerkaufEvent()`, `writeVersionedEvent()`.
- `backend/api/kasse/kassenfuehrung/application/command.go` —
  `GeldtransitBuchen()`, `writeKassensitzungEvent()`.
- `backend/api/kasse/tischgeschaeft/http/command_handler.go`,
  `backend/api/kasse/direktverkauf/http/command_handler.go`,
  `backend/api/kasse/kassenfuehrung/http/command_handler.go` — Request-DTOs,
  zog-Schemas, Fehlerabbildung.
- `frontend/src/lib/errorMessages.ts` — `commonErrorMessages`,
  `getActionErrorMessage()`.

### Backend — Zeitlimits

- `backend/app/app.go` — `NewApp()` mit `WriteTimeout: 10 * time.Second`.
- `backend/api/fiskal/export/http/handler.go` — `exportWriteTimeout`,
  `ExportHandler()`; das Vorbild für die Verlängerung der Schreibfrist.
- `backend/api/fiskal/setup/http/command_handler.go` —
  `RichteTSEEinHandler()`, `UebernimmTSEHandler()`.
- `backend/api/fiskal/setup/http/query_handler.go` —
  `TestTSEVerbindungHandler()`, `CheckTSESetupHandler()`.
- `backend/api/fiskal/setup/application/setup.go` — `RichteTSEEin()`,
  `vollendeLebenszyklus()`, `saveEinrichtung()`, `hatAktiveTSS()`.
- `backend/repository/tse_repo/fiskaly_client.go` — `defaultHTTPTimeout` (10 s),
  `defaultRetryAttempts` (3, also bis zu vier Versuche je Aufruf).
- `backend/api/middleware/middleware.go` — `responseWriter` mit `Unwrap()`.
- `backend/api/helper/http.go` — `SendConflict()`, `SendClientError()`.

### Frontend

- `frontend/src/lib/Backend.ts` — `BackendClient`, `Backend`,
  `REQUEST_TIMEOUT_MS`, `request()`, `leseBody()`, `post()`, `download()`,
  `NetzwerkFehler`.
- `frontend/src/lib/queryClient.ts` — `createQueryClient()`,
  `AKTUALITAETSSCHWELLE_MS`, `sollWiederholen()`, `QueryCache.onError`
  (toastet bereits jeden Query-Fehler, auch den eines Hintergrund-Refetch).
- `frontend/src/hooks/use-vorgang-id.ts` — `useVorgangId()`.
- Aufrufstellen von `useVorgangId`: `ZahlungAbschluss.tsx`,
  `BestellungAbschluss.tsx`, `HistorieStornierungDrawer.tsx`,
  `HistorieUmbuchungDrawer.tsx`, `DirektverkaufAbschluss.tsx`,
  `DirektverkaufStornoDrawer.tsx` (alle unter `frontend/src/service/`).
- `frontend/src/admin/kasse/GeldtransitDialog.tsx` — `GeldtransitDialog` mit
  eigenem `useState(() => crypto.randomUUID())` statt `useVorgangId`.
- `frontend/src/service/table/hooks.ts` — `useAktiveTische()`,
  `useTischHistorie()`, `useTischState()`, `useAktiveTischeMitFavoriten()` +
  `AKTIVE_TISCHE_MIT_FAVORITEN_KEY`, `useMeineTischeState()` +
  `MEINE_TISCHE_STATE_KEY`, `useEigeneUebersicht()` + `EIGENE_UEBERSICHT_KEY`.
  Die Schlüssel `'tisch-state'`, `'tisch-historie'`, `'aktive-tische'` sind
  String-Literale ohne exportierte Konstante.
- `frontend/src/service/product/hooks.ts` — `useAktiveProdukte()`
  (`'aktive-produkte'`).
- `frontend/src/service/direktverkauf/hooks.ts` — `useDirektverkaufHistorie()`.
- `frontend/src/admin/reporting/hooks.ts` —
  `ABGESCHLOSSENE_KASSENSITZUNGEN_KEY`, `REPORT_KEY`, `LIVE_REPORTING_KEY`,
  `useAbgeschlosseneKassensitzungen()`, `useReport()`, `useLiveReporting()`.
- `frontend/src/admin/kasse/hooks.ts` — `useOffeneKassensitzung()`,
  `useKassenbestand()` + `KASSENBESTAND_KEY`, `useGeldtransitListe()` +
  `GELDTRANSIT_LISTE_KEY`.
- `frontend/src/admin/kasse/KassensitzungPage.tsx` — `invalidateKasse()`.
- `frontend/src/admin/kasse/KasseAbschliessenSection.tsx` —
  `handleAbschliessen()`.
- `frontend/src/service/TablePage.tsx` — `reload()`, `erfolgSchliessen()`.
- `frontend/src/service/TableSelectionPage.tsx` — `TableSelectionPage`,
  Fehler-Frühausstieg vor Suchfeld, Favoritenliste, Fußzeilen-Button und
  `TischAuswahlDrawer`.
- `frontend/src/service/components/table/Bestellung.tsx`,
  `frontend/src/service/components/direktverkauf/Direktverkauf.tsx`,
  `frontend/src/service/components/direktverkauf/DirektverkaufHistorie.tsx` —
  je ein Fehler-Frühausstieg mit `LadefehlerAlert`.
- `frontend/src/components/common/LadefehlerAlert.tsx` — `LadefehlerAlert`
  (Props `titel`, `onErneutVersuchen`, `className`).
- `frontend/src/admin/tse/TSEBackend.ts` — `richteTSEEin()`, `uebernimmTSE()`,
  `testTSEVerbindung()`, `checkTSESetup()`.
- `frontend/src/admin/reporting/ReportingBackend.ts` — `exportDsfinvk()`
  (einzige `download()`-Aufrufstelle).

### Tests und CI

- `backend/api/fiskal/tse_live/tse_live_suite_test.go` — `cleanLiveDB()`.
- `backend/api/kasse/tischgeschaeft/application/idempotenz_integration_test.go`,
  `backend/api/kasse/direktverkauf/application/idempotenz_integration_test.go`.
- `backend/api/kasse/tischgeschaeft/application/command_idempotenz_test.go`,
  `backend/api/kasse/direktverkauf/application/command_test.go` (mit
  `spyEventRepo`).
- `backend/repository/kassenjournal_repo/repo_vorgang_idempotenz_test.go`,
  `backend/repository/kassenjournal_repo/repo_test.go`.
- `frontend/src/hooks/use-vorgang-id.test.ts`,
  `frontend/src/lib/Backend.test.ts`, `frontend/src/lib/queryClient.test.ts`,
  `frontend/src/service/TableSelectionPage.test.tsx`,
  `frontend/src/service/components/table/ZahlungAbschluss.test.tsx`.
- `.github/workflows/ci.yml` — Job `upgrade-path` mit
  `PREVIOUS_VERSION: v0.14.0`.
- `database/migrations/README.md` — nennt v0.14.0 als aktuelle Vorversion.

## Getroffene Entscheidungen

- **Umfang:** alle 16 Befunde (Blocker, schwerwiegend, geringfügig, Cleanup).
- **`payload_hash` kommt in die Migration, die `vorgang_idempotenz` anlegt**,
  nicht in eine nachgelagerte. v0.17.1 enthält nur `01`–`05`; `07` und `08` sind
  neu auf diesem Branch und liefen auf keiner Instanz. Konsequenz: lokale Dev-DB
  und die Demo-Instanz jotti.rocks müssen einmal zurückgesetzt werden, weil
  golang-migrate diese Versionen dort bereits als angewendet führt
  (`make rocks-reset-db` bzw. lokaler Reset).
- **Die Migrationen dieses Branch sind beim Rebase auf `main` von `06`/`07` auf
  `07`/`08` gerückt.** `main` hat zwischenzeitlich ein eigenes
  `06_druckauftrag_backoff_warteschlange.up.sql` bekommen; zwei Dateien mit
  derselben Versionsnummer lehnt golang-migrate ab, und Git meldet dabei keinen
  Konflikt, weil die Dateinamen verschieden sind.
- **Alle sieben buchenden Endpunkte laufen über `vorgang_idempotenz`.**
  Bestellung, Direktverkauf und Geldtransit ziehen von der Idempotenz über den
  Event-JSON-Index auf die Tabelle um. Das ist nicht nur additiv: Es entfernt
  `EventExistsByTypeAndVorgangsID` samt sqlc-Query, drei
  Interface-Deklarationen, Mock- und Spy-Implementierung sowie die drei
  „UniqueViolation → nachschlagen"-Zweige.
- **Gleicher Schlüssel mit anderen Nutzdaten ergibt einen expliziten 409**
  (`vorgang_daten_abweichend`), keinen stillen Erfolg. Weder Doppelbuchung noch
  verschluckte Einreichung.
- **`useVorgangId` rotiert nicht mehr bei Nutzdaten-Änderung**, sondern pro
  fachlichem Vorgang. Ohne diesen Rückbau sähe der Server denselben Schlüssel
  nie zweimal und die Hash-Prüfung liefe ins Leere.
- **`staleTime`-Voreinstellung wird auf `0` gedreht**, `30_000` wird zur
  benannten Ausnahme auf Stammdaten-Hooks. Die vom Review vorgeschlagene
  Auswahl dreier Queries wäre eine Aufzählung ohne benennbares Kriterium; die
  Umkehr ist in einem Satz beschreibbar und macht die unsichere Variante zur
  bewussten Entscheidung.
- **Ein fehlgeschlagener Hintergrund-Refetch braucht keine neue UI-Komponente.**
  `QueryCache.onError` in `queryClient.ts` toastet bereits jeden Query-Fehler
  (unterdrückt nur offline, wo das Banner die Ursache trägt). Der Fix beschränkt
  sich darauf, die gültigen Cache-Daten stehen zu lassen.

### Über die Review-Liste hinaus

Sechs Punkte, die der Review nicht aufführt, ohne die die zugehörige Korrektur
aber unvollständig bliebe. Sie sind bewusst enthalten und hier benannt:

- **Serverseitige Schreibfrist der TSE-Endpunkte** (Phase 1). Ein erhöhtes
  Client-Zeitlimit allein genügt nicht: Bei `WriteTimeout: 10 * time.Second`
  stirbt die Antwort mit PUK und Admin-PIN auf der Verbindung, sobald der
  Handler länger als 10 s braucht. Das ist auf `main` bereits so; ohne diesen
  Teil bliebe der Blocker-Fix wirkungslos.
- **Abbruchschutz der TSE-Einrichtung** (Phase 1). Der eigentliche Blocker ist
  nicht das Zeitlimit, sondern dass ein Client-Abbruch `r.Context()` storniert
  und damit den laufenden fiskaly-Lebenszyklus mittendrin abschießt — bei jedem
  Zeitlimit. PUK und Admin-PIN werden nie persistiert, und erst
  `saveEinrichtung` schreibt am Ende tssId, clientId und Zugangsdaten: Läuft der
  Lebenszyklus zu Ende, bleibt die Instanz betriebsfähig; bricht er mittendrin
  ab, ist sie es nicht — `hatAktiveTSS` blockiert den zweiten Versuch mit
  `tse_bereits_eingerichtet`, und die Übernahme scheitert an der fehlenden PIN.
  Ohne diesen Teil verschöbe der Blocker-Fix die Frist nur, statt das Problem zu
  beheben.
- **Serialisierung der beiden schreibenden Einrichtungs-Kommandos** (Phase 1).
  Erst der Abbruchschutz macht sie nötig: Weil der Lebenszyklus den Client
  überlebt und der Button nach der Fehlermeldung sofort wieder frei ist, kann
  der Admin eine zweite Einrichtung starten, während die erste noch vor
  `CreateTSS` steht. Der zweite Lauf sähe in `ListTSS` ein leeres Konto,
  `hatAktiveTSS` meldete false, und er legte eine zweite bezahlte LIVE-TSS an;
  beide Läufe endeten in `saveEinrichtung`, der spätere gewinnt, und die
  angezeigten PUK/PIN gehörten zur nicht konfigurierten TSS. Ohne diesen Teil
  hätte der Abbruchschutz einen neuen Fehlablauf eingeführt.
- **`TablePage`s Fehlerzweig** (Phase 5). Er hat dieselbe Form wie die vier vom
  Review benannten Zweige und denselben Defekt. Ihn auszulassen hinterließe eine
  Inkonsistenz in derselben Änderung.
- **Die beiden verbliebenen Leer-Defaults des Service-Pfads** (Phase 5):
  `TischAuswahlDrawer` und `HistorieUmbuchungDrawer`. Beide zeigen bei einem
  gescheiterten Erstladen eine leere Liste statt einer Meldung und behaupten
  damit dasselbe Falsche wie die vom Review benannten Zweige — der Drawer „keine
  Tische vorhanden", die Ziel-Tisch-Auswahl „Kein aktiver Ziel-Tisch verfügbar".
  Die Begründung ist dieselbe wie bei `TablePage`: Der `TischAuswahlDrawer` ist
  der einzige Einstieg in einen Tisch, wenn die Servicekraft keine Tische
  markiert hat, und ohne Ziel-Tisch ist eine Umbuchung nicht ausführbar. Sie
  auszulassen hinterließe dieselbe Falschaussage-Klasse in derselben Änderung.
- **`PREVIOUS_VERSION` im CI-Job `upgrade-path`** (Phase 6). Der Pin steht auf
  v0.14.0, produktiv läuft v0.17.1. Dieser Branch bringt die ersten
  Schema-Migrationen seit v0.14.0; das Gate soll den Pfad prüfen, der real
  läuft.

### Nicht im Umfang

- **`kassensitzung-eroeffnen` und `kasse-abschliessen` bekommen keinen
  Idempotenz-Schlüssel.** Beide sind bereits durch eine fachliche Invariante
  gegen eine Zweitbuchung geschützt, anders als Bestellung und Direktverkauf,
  die rein anhängen:
  - `KassensitzungEroeffnen()` prüft über `GetAktiveKassensitzung()` auf eine
    aktive Sitzung und liefert sonst `ErrKasseAlreadyOpen`; darunter liegt der
    partielle Unique-Index `idx_kassensitzungen_eine_aktiv`. Es kann nie eine
    zweite Sitzung entstehen.
  - `KasseAbschliessen()` erkennt über `findeVorhandenenKassensturz()` einen
    Kassensturz aus einem abgebrochenen früheren Versuch und überspringt
    Schritt 1; liegt dazwischen eine Buchung, bricht der Abschluss mit
    `ErrBuchungenNachKassensturz` ab. Die Folge-Events laufen über OCC gegen
    `expectedVersion`. Das ist eine Wiederanlauf-Erkennung am fachlichen
    Zustand und damit stärker als ein Client-Schlüssel.

  Einen eigenständigen Kassensturz-Endpunkt gibt es nicht; der Kassensturz ist
  Schritt 1 von `KasseAbschliessen()`.

  Bekannte Restlücke, bewusst offen gelassen: Geht bei
  `kassensitzung-eroeffnen` die Antwort verloren, meldet der Wiederholversuch
  `kasse_bereits_geoeffnet` — ein falsches Negativ trotz erfolgreicher
  Eröffnung. Der Zustand ist auf derselben Seite sichtbar, es entstehen keine
  falschen Daten.
- **Keine Änderung an Event-Formaten, an den partiellen Unique-Indexen in
  `01_initial.up.sql` oder an Docker-/Caddy-Konfiguration.**

## Risiken

- **Lokale und Demo-Datenbanken müssen zurückgesetzt werden** (Änderung an
  Migration `08`). Auf Produktion existiert die Tabelle nicht, dort ist es
  folgenlos. Wird der Reset vergessen, fehlt `payload_hash` still und der erste
  buchende Vorgang schlägt fehl.
- **Ein Wiederholversuch, der das Upgrade überspannt, endet in HTTP 409 statt
  in der stillen Erfolgsantwort.** Migration `08` legt für Bestands-Events keine
  Idempotenz-Zeilen an (auf der nachgestellten Upgrade-Strecke v0.17.1 → seed →
  `migrate up`: 673 Zeilen `kassenjournal`, davon 324
  `bestellung-aufgenommen:v1`, und 0 Zeilen `vorgang_idempotenz`). Wurde die
  erste Einreichung noch vor dem Upgrade gebucht und die Wiederholung erst
  danach abgesetzt, liefert die Vorprüfung `VorgangNeu`, das Kommando bucht
  regulär, und der alte partielle Index auf dem Event-JSON schlägt zu
  (`duplicate key value violates unique constraint
  "idx_kassenjournal_bestellung_id"`). Der Verstoß landet im generischen
  `db.ErrAlreadyExists`-Zweig und wird zu `ErrConflict`. Doppelt gebucht wird
  nichts, und das Fenster schließt sich mit dem Upgrade. Der entfernte Lookup
  wird dafür bewusst nicht wiederhergestellt — er wäre ein dauerhafter zweiter
  Idempotenz-Mechanismus für ein Fenster von Minuten. Die Log-Meldung der drei
  betroffenen Schreibpfade ist deshalb neutral formuliert („Unique violation on
  event write" statt „OCC conflict"), damit sie beide Quellen deckt.
- **Der neue Fehlerzustand `vorgang_daten_abweichend` ist für Helfer neu.** Die
  Meldung muss ohne technisches Vorwissen zu einer sinnvollen Handlung führen
  („Tisch aktualisieren und die Differenz nachbuchen").
- **Die Umkehr der `staleTime`-Voreinstellung erhöht den Traffic** gegenüber dem
  Zustand des Branch. Bei 20–30 Helfern im Vereins-WLAN ist das die
  Gegenrichtung zu Phase 4 des Ursprungsplans; die Ersparnis bleibt nur dort
  erhalten, wo sie unkritisch ist.
- **Ein veralteter Tisch-Saldo bleibt kurzzeitig sichtbar und bedienbar.**
  `staleTime: 0` löst beim Mount einen Refetch aus, verbirgt die
  zwischengespeicherten Daten aber nicht: Innerhalb der `gcTime` (Standard 5
  Minuten, nirgends überschrieben) zeigt ein zweiter Mount den alten Saldo mit
  entsperrten Reitern weiter, bis der Refetch antwortet — und dauerhaft, wenn
  der Refetch scheitert, weil `isLoadingError` bei vorliegenden Daten false
  bleibt (genau die in Phase 5 gewollte Unterscheidung). Der Restweg wäre eine
  Reiter-Sperre bei `isError && !isFetching`; das ist eine eigene
  Produktentscheidung und bewusst nicht Teil dieses Plans.
- **Reale Laufzeiten sind weiterhin nicht gemessen.** Ob 2 Minuten Serverarbeit
  und 150 s Clientgeduld für die TSE-Einrichtung über einen Vereins-Uplink
  genügen, ist aus Timeout- und Retry-Budgets abgeleitet (bis zu zehn
  HTTP-Sequenzen inklusive Auth, je 10 s Timeout, bis zu vier Versuche mit
  Backoff), nicht gemessen. Die Werte sind begründete Schätzungen, keine Messungen. Ein zu knapp
  bemessenes Budget ist seit dem Abbruchschutz allerdings nicht mehr
  gefährlich: Der Lebenszyklus läuft auch nach einem Abbruch zu Ende und
  speichert, die Instanz bleibt betriebsfähig. Verloren wäre die einmalige
  Anzeige von PUK und Admin-PIN — schmerzhaft, aber kein blockierter Zustand.
- **Läuft der Lebenszyklus länger als das Client-Budget (150 s), sind PUK und
  Admin-PIN verloren**, obwohl die Instanz betriebsfähig bleibt: Sie existieren
  nur in der einen Antwort, die den Client dann nicht mehr erreicht. Der Admin
  kann die TSE weiter betreiben, aber keine privilegierte fiskaly-Operation mehr
  ausführen. Das ist der Restschaden des Abbruchschutzes; beseitigen würde ihn
  nur eine andere Auslieferungsstrategie für die beiden Geheimnisse (etwa
  einmaliges Abholen über einen zweiten Aufruf), die dieser Plan bewusst nicht
  aufmacht.
- **Der Abbruchschutz greift gegen den Client-Abbruch, nicht gegen ein
  Prozessende.** Ein Deploy oder Neustart während einer laufenden TSE-Einrichtung
  wartet über `http.Server.Shutdown` bis zu 30 s auf den noch laufenden Handler
  (`backend/app/app.go`); ein danach immer noch laufender Lebenszyklus wird mit
  dem Prozess mitgerissen, und der Endzustand ist wieder der Blocker: bezahlte
  TSS, `hatAktiveTSS` sperrt die Neuanlage, die Übernahme scheitert an der
  fehlenden PIN. Bekannter Restweg, organisatorisch abgedeckt: Während einer
  laufenden TSE-Einrichtung wird nicht deployt und nicht neu gestartet.
- **`make test-tse-live` läuft nicht in CI** und braucht echte
  fiskaly-Zugangsdaten. Der Teardown-Fix aus Phase 2 lässt sich nur lokal
  nachweisen.

---

## Phase 1: Langlaufende Endpunkte überleben das Zeitlimit

Behebt den Blocker und den geringfügigen Befund zur Body-Phase.

### Kontext

- `frontend/src/lib/Backend.ts — request()` — setzt `REQUEST_TIMEOUT_MS` (8000)
  unkonditional für jeden Endpunkt; `clearTimeout` steht im `finally` von
  `await fetch` und greift damit, sobald die Header da sind.
- `frontend/src/admin/tse/TSEBackend.ts — richteTSEEin(), uebernimmTSE(),
  testTSEVerbindung(), checkTSESetup()` — die vier Aufrufe, die hinter dem
  Backend zu fiskaly gehen.
- `frontend/src/admin/reporting/ReportingBackend.ts — exportDsfinvk()` — einzige
  `download()`-Aufrufstelle.
- `backend/api/fiskal/setup/application/setup.go — RichteTSEEin()` — sequenziell
  `ListTSS`, `CreateTSS`, `vollendeLebenszyklus` (sechs Aufrufe, darunter
  zweimal Admin-Auth), danach erst `saveEinrichtung` (Speichern plus
  Stammdaten-Abruf).
- `backend/repository/tse_repo/fiskaly_client.go — defaultHTTPTimeout,
  defaultRetryAttempts` — 10 s pro Aufruf, bis zu vier Versuche
  (`defaultRetryAttempts = 3`) mit Backoff.
- `backend/app/app.go — NewApp()` — `WriteTimeout: 10 * time.Second`; die Frist
  läuft ab dem Ende des Header-Lesens, storniert `r.Context()` nicht, tötet aber
  den Schreibvorgang der Antwort.
- `backend/api/fiskal/export/http/handler.go — exportWriteTimeout,
  ExportHandler()` — das bestehende Vorbild für die verlängerte Schreibfrist.
- `backend/api/middleware/middleware.go — responseWriter.Unwrap()` — macht
  `http.ResponseController` durch den Logging-Wrapper hindurch wirksam.

### Was gebaut wird

`BackendClient.post` und `BackendClient.download` nehmen ein optionales
`RequestOptionen`-Objekt mit `zeitlimitMs` entgegen; ohne Angabe bleibt es beim
bisherigen Standard von 8 s. Die vier fiskaly-Endpunkte und der DSFinV-K-Export
setzen ihr Zeitlimit explizit — jeweils auf das Schreibbudget des Servers plus
30 s Netzreserve. Der Aufschlag gilt der Antwort, nicht der Reihenfolge des
Aufgebens: Ein spät, aber erfolgreich geschriebener Antwort-Body soll den
Client noch erreichen. Wäre das Client-Budget nicht größer, hätte der Client in
genau dem Fenster schon aufgegeben, in dem der Server gerade noch schreibt.

Das Zeitlimit deckt künftig den gesamten Aufruf ab, einschließlich des Lesens
des Antwort-Bodys: Der `AbortController` wird nicht mehr beim Auflösen von
`fetch` freigegeben, sondern erst, wenn Body oder Blob vollständig gelesen sind.
Ein Abbruch in der Body-Phase bleibt eine `zeitueberschreitung`, ein
Verbindungsabriss ohne Abbruchsignal ein `verbindungsabbruch`.

Serverseitig verlängern die vier fiskaly-Handler die Schreibfrist der
Verbindung, bevor sie das erste Byte schreiben — sonst stirbt die Antwort mit
PUK und Admin-PIN an `WriteTimeout: 10s`, egal wie geduldig der Client wartet.
Der dafür nötige `http.NewResponseController(w).SetWriteDeadline(...)`-Block
samt Warn-Log bei Fehlschlag zieht in einen gemeinsamen Helfer in
`backend/api/helper`; `ExportHandler()` nutzt denselben Helfer statt seiner
eigenen Kopie.

Ein einziger Aufruf am Handler-Eingang genügt dabei nicht: Die Frist ist eine
absolute Zeit ab Request-Start, keine Stoppuhr für den Schreibvorgang. Sie läuft
während der gesamten Handler-Laufzeit, und ein Handler, der 130 s mit fiskaly
spricht, schreibt danach in eine Frist, die vor 10 s abgelaufen ist. Alle fünf
Handler setzen sie deshalb zweimal: am Eingang, damit die frühen
Validierungsfehler noch ankommen, und unmittelbar vor dem Schreiben der Antwort.
Bei den vier fiskaly-Handlern liegt der zweite Aufruf direkt nach der Rückkehr
aus Command bzw. Query und deckt damit Fehler- und Erfolgszweig zugleich ab; im
`ExportHandler()` liegt er zwischen `Erstellen()` und dem Setzen der Header.

Die beiden schreibenden TSE-Endpunkte werden zusätzlich gegen den Client-Abbruch
immunisiert. Sie führen ihren fiskaly-Lebenszyklus unter einem vom
Request-Kontext abgekoppelten Kontext aus (`context.WithoutCancel`, darüber ein
`context.WithTimeout` als Leck-Wächter mit `defer cancel()`): Schließt der
Client die Verbindung, storniert net/http `r.Context()` und der Lebenszyklus
bräche mitten in der fiskaly-Sequenz ab — zurück bliebe eine bezahlte,
halbfertige TSS, deren PUK und Admin-PIN es nur in der verlorenen Antwort gab.
Er muss also auch ohne Zuhörer zu Ende laufen und speichern. Der Wert des
Leck-Wächters liegt bewusst weit über dem Worst Case des Lebenszyklus (bis zu
elf HTTP-Sequenzen à rund 41 s, zusammen rund 7,5 Minuten): Er ist kein
Reaktionszeit-Budget — das trägt der Client — sondern verhindert nur, dass eine
hängende fiskaly-Verbindung den abgekoppelten Lebenszyklus dauerhaft offenhält.
Zu knapp gewählt wäre er der Blocker in neuer Form. Abgekoppelt ist dabei der
Kontext, nicht der Ablauf: Der Handler startet keine Goroutine, sondern fährt
den Lebenszyklus synchron und kehrt erst mit ihm zurück.

Die beiden lesenden TSE-Endpunkte behalten `r.Context()`: Sie sind idempotent
und jederzeit wiederholbar, ein Abbruch hinterlässt dort nichts.

### Akzeptanzkriterien

- [x] `BackendClient.post` und `BackendClient.download` akzeptieren ein
      optionales `RequestOptionen`-Objekt mit `zeitlimitMs`; ohne Angabe gilt
      `REQUEST_TIMEOUT_MS = 8000`.
- [x] `admin/tse-einrichten`, `admin/tse-uebernehmen`,
      `admin/test-tse-verbindung` und `admin/tse-setup-pruefen` laufen mit
      `TSE_TIMEOUT_MS = 150_000`.
- [x] `admin/export/dsfinvk` läuft mit `EXPORT_TIMEOUT_MS = 330_000`.
- [x] Beide Client-Zeitlimits liegen 30 s über dem jeweiligen serverseitigen
      Schreibbudget (`tseSetupWriteTimeout = 2 Minuten`,
      `exportWriteTimeout = 5 Minuten`); die Kommentare an allen vier Konstanten
      nennen die Beziehung Client-Budget = Schreibbudget des Servers +
      Netzreserve und begründen sie mit der Antwort, die den Client noch
      erreichen soll. Kein Kommentar behauptet, die Schreibfrist bringe den
      Server dazu, „zuerst aufzugeben"; das Wort „deckungsgleich" kommt in
      diesem Zusammenhang nirgends mehr vor.
- [x] Das Zeitlimit bricht auch einen in der Body-Phase stockenden Aufruf ab;
      ein Test in `frontend/src/lib/Backend.test.ts` deckt genau diesen Fall ab
      (Header kommen an, Body-Stream stockt) und erwartet einen
      `NetzwerkFehler` mit `art === 'zeitueberschreitung'`.
- [x] Ein Test belegt, dass ein Verbindungsabriss in der Body-Phase ohne
      abgelaufenes Zeitlimit weiterhin `art === 'verbindungsabbruch'` liefert.
- [x] `RichteTSEEinHandler()`, `UebernimmTSEHandler()`,
      `TestTSEVerbindungHandler()` und `CheckTSESetupHandler()` verlängern die
      Schreibfrist auf 2 Minuten, bevor sie schreiben.
- [x] Alle fünf langlaufenden Handler (die vier TSE-Handler und
      `ExportHandler()`) setzen die Schreibfrist zweimal: am Handler-Eingang und
      unmittelbar vor dem Schreiben der Antwort. Die Tests zählen die Aufrufe
      **vor** dem ersten Schreibvorgang und fordern zwei; sie werden rot, wenn
      man den zweiten Aufruf entfernt oder hinter das Schreiben verschiebt.
- [x] Der Doc-Kommentar von `helper.ExtendWriteDeadline` erklärt, dass die Frist
      absolut ab Request-Start läuft und ein langlaufender Handler sie deshalb
      zweimal setzt.
- [x] Der Helfer für die Schreibfrist liegt in `backend/api/helper`, wird von
      den vier TSE-Handlern und von `ExportHandler()` genutzt, und die
      bisherige Inline-Kopie in `ExportHandler()` ist entfernt.
- [x] Scheitert das Setzen der Schreibfrist, wird gewarnt und weitergearbeitet
      (kein Abbruchgrund) — wie bisher im Export-Handler.
- [x] `RichteTSEEinHandler()` und `UebernimmTSEHandler()` übergeben ihrem
      Command einen vom Request-Kontext abgekoppelten Kontext
      (`context.WithoutCancel` mit `context.WithTimeout` als Leck-Wächter und
      `defer cancel()`); `TestTSEVerbindungHandler()` und
      `CheckTSESetupHandler()` behalten `r.Context()`.
- [x] Ein Test belegt beides: Bei den beiden schreibenden Handlern bleibt der an
      das Command übergebene Kontext gültig, wenn der Request-Kontext storniert
      wird, bei den beiden lesenden nicht. Nimmt man `context.WithoutCancel`
      heraus, wird er rot.
- [x] Korrelations-ID und zerolog-Logger sind im abgekoppelten Kontext
      weiterhin verfügbar; der Test prüft das mit.
- [x] Die Doc-Kommentare der beiden schreibenden Handler bzw. der Konstanten
      erklären, warum: Ein Client-Abbruch darf keine bezahlte LIVE-TSS
      halbfertig zurücklassen; PUK und Admin-PIN existieren nur in dieser einen
      Antwort, der Lebenszyklus muss auch ohne Zuhörer zu Ende laufen und
      speichern. Der Wert des Leck-Wächters ist mit seiner Rechnung belegt; der
      ungedeckelte `Retry-After` des fiskaly-Clients ist als bekannte Lücke
      benannt, und die Zusage des abgekoppelten Kontexts ist auf den
      Client-Abbruch beschränkt (ein Prozessende beendet ihn).
- [x] Es schreibt höchstens einer auf der TSE-Konfiguration: `RichteTSEEin`,
      `UebernimmTSE` und `UpdateTSEKonfiguration` teilen sich ein
      prozessinternes Schloss in der Application-Schicht, das in jedem Pfad
      (auch bei Panik) wieder frei wird. Ein zweiter Aufruf wartet nicht,
      sondern endet sofort mit `ErrTSESetupLaeuftBereits` → HTTP 409
      `tse_setup_laeuft_bereits`; `frontend/src/lib/errorMessages.ts` führt den
      Admin zum Warten statt zum Wiederholen.
- [x] Nach dem Client-Zeitlimit rät die Meldung der beiden schreibenden
      TSE-Aufrufe nicht zum Wiederholen: `getActionErrorMessage` nimmt ein
      optionales `byNetzwerkArt`, das vor der allgemeinen Netzmeldung greift,
      und `TSEEinrichtungWizard` belegt damit die Zeitüberschreitung mit „läuft
      im Hintergrund weiter, Konfiguration prüfen".
- [x] Ein Test belegt, dass der zweite Aufruf während eines laufenden ersten den
      neuen Fehler liefert und dabei keinen fiskaly-Aufruf absetzt; ein zweiter
      belegt, dass das Schloss nach Abschluss und nach einem Fehler wieder frei
      ist.
- [x] `make check` und die Frontend-Suite laufen grün.

---

## Phase 2: Der Idempotenz-Schlüssel wird an die Nutzdaten gebunden

Behebt den schwerwiegenden Befund zur Doppelbuchung (Mechanismus und die vier
bereits über `vorgang_idempotenz` laufenden Endpunkte), den Teardown-Befund der
TSE-Live-Suite sowie vier geringfügige Test- und Cleanup-Befunde.

### Kontext

- `database/migrations/08_vorgang_idempotenz.up.sql` — Tabelle ohne
  `payload_hash`, `art`-CHECK mit vier Werten.
- `backend/sqlc/queries/vorgang_idempotenz.sql` — `InsertVorgangIdempotenz`,
  `ExistsVorgangIdempotenz`.
- `backend/repository/kassenjournal_repo/repo.go — Vorgang,
  VorgangBereitsGebucht(), insertVorgangIdempotenzInTx(),
  ErrVorgangBereitsGebucht` — die bestehende reine ID-Prüfung.
- `backend/api/kasse/tischgeschaeft/application/command.go —
  vorgangBereitsGebucht(), ZahlungKassieren(), StornierungErteilen(),
  BestellungUmbuchen(), persistTischEvent()` — die drei Vorprüfungen und die
  drei Zweige, die `ErrVorgangBereitsGebucht` aus der Schreibtransaktion auf
  eine stille Erfolgsantwort abbilden.
- `backend/api/kasse/direktverkauf/application/command.go —
  DirektverkaufStornieren()` — vierte Vorprüfung, vierter Zweig.
- `backend/repository/kassenjournal_repo/mock.go — MockRepo.WriteEvent(),
  recordVorgang(), vorgangBereitsGebucht()` — der Mock schreibt die Projektion
  nach einem erfolgreichen Write nicht fort, weshalb die Vorprüfung in den
  Unit-Tests nicht tragend ist.
- `backend/api/fiskal/tse_live/tse_live_suite_test.go — cleanLiveDB()` — die
  Statement-Liste endet mit `DELETE FROM users` und räumt
  `vorgang_idempotenz` nicht ab; `user_id` referenziert `users(id)` ohne
  `ON DELETE`, der Teardown bricht mit 23503 ab.
- `frontend/src/hooks/use-vorgang-id.ts — useVorgangId()` — rotiert bei jeder
  Nutzdaten-Änderung.
- `frontend/src/service/components/table/ZahlungAbschluss.tsx` — enthält bereits
  den `warLeer`-Render-State-Sync für den Reset der Eingaben; derselbe Übergang
  trägt künftig die Schlüssel-Rotation.

### Was gebaut wird

Die Tabelle `vorgang_idempotenz` bekommt `payload_hash BYTEA NOT NULL`
(direkt in Migration `08`). Die Application-Schicht bildet aus den
Kommando-Argumenten jedes buchenden Vorgangs einen SHA-256 über eine explizite,
pro Kommando deklarierte Nutzdaten-Struktur und übergibt ihn im `Vorgang`.

Die Prüfung liefert statt eines Booleans drei Zustände: neu, Duplikat (gleiche
ID, gleicher Hash) oder abweichende Nutzdaten (gleiche ID, anderer Hash). Ein
Duplikat ergibt weiterhin die stille Erfolgsantwort, abweichende Nutzdaten einen
409 mit dem Fehlercode `vorgang_daten_abweichend`. Dieselbe Prüffunktion wird an
beiden Stellen benutzt: als Vorprüfung vor der fachlichen Validierung und noch
einmal, wenn der Insert in der Schreibtransaktion am Primärschlüssel scheitert.
Die Nutzdaten werden in der übergebenen Reihenfolge gehasht; eine umsortierte
Einreichung gilt als abweichend und führt zum 409, nicht zu einer zweiten
Buchung.

Im Frontend rotiert der Schlüssel nicht mehr bei Nutzdaten-Änderungen, sondern
pro fachlichem Vorgang: `useVorgangId` bekommt statt der Nutzdaten die Angabe,
ob die Zusammenstellung gerade leer ist, und erzeugt einen neuen Schlüssel beim
Übergang von leer zu nicht leer. Ein Wiederholversuch nach einem Fehlschlag
behält den Schlüssel — auch mit geänderter Auswahl, denn genau diesen Fall soll
der Server jetzt erkennen und melden.

Die Signaturänderung erzwingt, dass **alle** Aufrufstellen in dieser Phase
umgestellt werden, auch die drei, deren Serverbindung erst Phase 3 bringt
(Bestellung, Direktverkauf, Geldtransit). Diese drei haben zwischen Phase 2 und
Phase 3 den Schlüssel-Lebenszyklus von `main` ohne Nutzdaten-Prüfung — eine
geänderte Zweiteinreichung würde dort noch still verschluckt. Der Branch darf
zwischen diesen beiden Phasen nicht ausgeliefert werden; Phase 3 schließt die
Lücke.

Die `art`-CHECK-Liste in Migration `08` wird schon hier auf alle sieben Werte
erweitert, damit die Migration nicht zweimal angefasst werden muss. Drei der
Werte werden erst in Phase 3 benutzt.

In diesem Zug werden vier Testbefunde erledigt: `cleanLiveDB` räumt
`vorgang_idempotenz` ab; der Mock schreibt seine Projektion nach einem
erfolgreichen Write fort, damit die Vorprüfung tragend wird; ein Spy liefert den
Duplikatfehler aus dem Write, ohne dass die Vorprüfung anschlägt, und macht die
bisher unerreichbaren Zweige testbar; die Tests zum Schlüssel-Lebenszyklus
decken den Kommentar mit ab.

Der Cleanup-Punkt zur `if … else`-Form der Vorprüfungen wird hier miterledigt,
weil dieselben Stellen ohnehin umgeschrieben werden.

### Akzeptanzkriterien

- [x] `08_vorgang_idempotenz.up.sql` deklariert `payload_hash BYTEA NOT NULL`;
      der Tabellen- und Spaltenkommentar erklärt die Nutzdaten-Bindung.
- [x] Der `art`-CHECK in derselben Migration deckt alle sieben Werte ab
      (`bestellung`, `zahlung`, `stornierung`, `umbuchung`, `direktverkauf`,
      `direktverkauf-stornierung`, `geldtransit`); die `VorgangArt*`-Konstanten
      in `kassenjournal_repo` spiegeln sie vollständig. Drei davon werden erst
      in Phase 3 benutzt.
- [x] `kassenjournal_repo.Vorgang` trägt `PayloadHash []byte`; eine
      Hash-Funktion im selben Paket bildet eine beliebige Nutzdaten-Struktur auf
      SHA-256 ab.
- [x] `kassenjournal_repo` stellt eine Prüffunktion bereit, die
      `VorgangNeu`, `VorgangDuplikat` oder `VorgangDatenAbweichend` liefert;
      `ExistsVorgangIdempotenz` ist durch eine Query ersetzt, die den
      gespeicherten `payload_hash` liest.
- [x] `make sqlc` ist gelaufen; `sqlc/dbgen/` ist ausschließlich generiert.
- [x] `ZahlungKassieren`, `StornierungErteilen`, `BestellungUmbuchen` und
      `DirektverkaufStornieren` bilden ihre Nutzdaten je über eine eigene,
      explizit deklarierte Struktur ab: Zahlung und Stornierung über
      `tischId`, `positionen` (`positionId`, `menge`) und `kommentar`;
      Umbuchung über `quellTischId`, `zielTischId`, `positionen` und
      `benutzerKommentar`; Direktverkauf-Stornierung über `verkaufId`,
      `positionen` und `kommentar`.
- [x] Gleiche `vorgangId` mit gleichem Hash ergibt weiterhin die stille
      Erfolgsantwort ohne zweite Buchung.
- [x] Gleiche `vorgangId` mit abweichendem Hash ergibt HTTP 409 mit dem Code
      `vorgang_daten_abweichend` — sowohl aus der Vorprüfung als auch aus dem
      Konflikt in der Schreibtransaktion.
- [x] `frontend/src/lib/errorMessages.ts` bildet `vorgang_daten_abweichend` auf
      eine deutsche Meldung ab, die zur richtigen Handlung führt (bereits
      gebucht, Tisch aktualisieren, Differenz nachbuchen).
- [x] `useVorgangId` nimmt keine Nutzdaten mehr entgegen; der Schlüssel bleibt
      über eine Zusammenstellung stabil und wechselt beim Übergang von leer zu
      nicht leer.
- [x] Alle sechs bestehenden Aufrufstellen sind auf die neue Signatur
      umgestellt: `ZahlungAbschluss`, `HistorieStornierungDrawer`,
      `HistorieUmbuchungDrawer`, `DirektverkaufStornoDrawer`,
      `BestellungAbschluss` und `DirektverkaufAbschluss`. Die letzten beiden
      erhalten ihre Serverbindung erst in Phase 3.
- [x] Ein Frontend-Test belegt: Schlüssel bleibt bei geänderter Auswahl und bei
      geändertem Kommentar gleich, und wechselt nach dem Leeren und erneuten
      Befüllen der Zusammenstellung.
- [x] Der Leerzustands-Reset in `ZahlungAbschluss` ist getestet: Ein Test wird
      rot, wenn der Block entfernt wird.
- [x] `cleanLiveDB()` löscht `vorgang_idempotenz` vor `users`.
- [x] Der Mock schreibt seine Projektion nach einem erfolgreichen Write fort;
      entfernt man die Vorprüfung aus einem der vier Kommandos, wird
      mindestens ein Unit-Test rot.
- [x] Ein Repo-Spy liefert den Duplikat- bzw. Abweichungsfehler aus dem Write,
      ohne dass die Vorprüfung anschlägt; die entsprechenden Zweige aller vier
      Kommandos sind damit erreicht.
- [x] Die Vorprüfungen in `tischgeschaeft/application/command.go` folgen der
      Form von `direktverkauf/application/command.go` — kein `else` an einem
      returnenden Fehlerzweig, Happy Path nicht im else-Zweig.
- [x] `make verify` läuft grün (inklusive der beiden
      `idempotenz_integration_test.go`), nachdem die lokale Datenbank
      zurückgesetzt wurde.

---

## Phase 3: Bestellung, Direktverkauf und Geldtransit nutzen denselben Mechanismus

Schließt den nachgewiesenen Doppelbuchungs-Pfad und den geringfügigen
Geldtransit-Befund; ersetzt die zweite Idempotenz-Mechanik vollständig.

### Kontext

- `database/migrations/01_initial.up.sql` — `idx_kassenjournal_bestellung_id`,
  `idx_kassenjournal_verkauf_id`, `idx_kassenjournal_geldtransit_id`; die
  partiellen Unique-Indexe bleiben unverändert.
- `backend/api/kasse/tischgeschaeft/application/command.go —
  BestellungAufnehmen(), writeEventWithDruckauftraege()` — Idempotenz über
  `ErrConflict` plus Nachschlagen per `bestellungId`.
- `backend/api/kasse/direktverkauf/application/command.go —
  DirektverkaufTaetigen(), persistVerkaufEvent()` — dieselbe Form über
  `verkaufId`.
- `backend/api/kasse/kassenfuehrung/application/command.go —
  GeldtransitBuchen(), writeKassensitzungEvent()` — dieselbe Form über
  `geldtransitId`.
- `backend/repository/kassenjournal_repo/repo.go —
  EventExistsByTypeAndVorgangsID(), WriteEventWithDruckauftraege()` — die
  Lookup-Funktion hat genau diese drei Produktions-Aufrufer.
- `frontend/src/admin/kasse/GeldtransitDialog.tsx — GeldtransitDialog` — eigener
  `useState(() => crypto.randomUUID())`, manuell nach Erfolg rotiert; der
  einzige der sieben Schlüssel ohne `useVorgangId`.
- `frontend/src/admin/kasse/KasseBackend.ts — geldtransitBuchen()`.

### Was gebaut wird

Die drei Endpunkte, deren Idempotenz-Schlüssel bisher im Event-JSON lag,
schreiben ihre Idempotenz-Zeile künftig wie alle anderen in
`vorgang_idempotenz` — vor den Events, in derselben Transaktion, mit
`payload_hash`. Damit gelten für sie dieselben drei Ausgänge wie in Phase 2, und
der Fall „gleicher Schlüssel, geänderte Nutzdaten" endet auch hier in einem
expliziten 409 statt in einer zweiten Buchung.

Weil die Idempotenz-Zeile vor den Events entsteht, können die partiellen
Unique-Indexe auf dem Event-JSON für alle ab Migration `08` geschriebenen
Vorgänge nicht mehr auslösen. Die drei
„UniqueViolation → per Fach-ID nachschlagen"-Zweige entfallen ersatzlos, und
mit ihnen `EventExistsByTypeAndVorgangsID` samt sqlc-Query, den drei
Interface-Deklarationen, der Mock- und der Spy-Implementierung. Die Indexe
selbst bleiben als zweite Absicherung stehen. Für Bestandsdaten aus der Zeit
vor der Migration bleibt eine Restlücke offen — sie steht unter „Risiken" und
rechtfertigt den Lookup nicht zurück.

Bestellung und Direktverkauf schreiben ihre Events zusammen mit Druckaufträgen;
dafür braucht der Repository-Layer eine `MitVorgang`-Variante des
Druckauftrags-Writes.

`GeldtransitDialog` wechselt auf `useVorgangId` und verliert seine manuelle
Rotation nach Erfolg — „leer" ist dort der Zustand ohne eingegebenen Betrag.

Der Cleanup-Punkt zu `persistVerkaufEvent` wird hier miterledigt, weil die
Funktion in dieser Phase ohnehin umgeschrieben wird.

### Akzeptanzkriterien

- [x] `BestellungAufnehmen`, `DirektverkaufTaetigen` und `GeldtransitBuchen`
      schreiben ihre Idempotenz-Zeile mit `payload_hash` vor den Events in
      derselben Transaktion; als `vorgang_id` dient die bereits vorhandene
      client-erzeugte `bestellungId`, `verkaufId` bzw. `geldtransitId`.
- [x] Die Nutzdaten-Strukturen sind explizit deklariert: Bestellung über
      `tischId`, `positionen` (`produktId`, `varianteId`, `menge`) und
      `kommentar`; Direktverkauf über `positionen` und `kommentar`;
      Geldtransit über `richtung`, `betragCents` und `kommentar`.
- [x] Gleiche ID mit abweichendem Hash ergibt bei allen drei Endpunkten HTTP 409
      mit `vorgang_daten_abweichend`; gleiche ID mit gleichem Hash die stille
      Erfolgsantwort.
- [x] Ein echter OCC-Konflikt ist weiterhin von einer Duplikat-Einreichung
      unterscheidbar und ergibt weiterhin `conflict`.
- [x] Das Repository stellt eine `MitVorgang`-Variante des
      Druckauftrags-Writes bereit; Bestellung und Direktverkauf nutzen sie.
      Events und Druckaufträge bleiben in einer Transaktion.
- [x] `EventExistsByTypeAndVorgangsID` ist samt sqlc-Query, den drei
      Interface-Deklarationen in den Application-Paketen, der Mock- und der
      Spy-Implementierung entfernt; eine repo-weite Suche findet keinen Treffer
      mehr.
- [x] Die partiellen Unique-Indexe in `01_initial.up.sql` sind unverändert.
- [x] `persistVerkaufEvent` ist in seinen einzigen Aufrufer inlined, analog zur
      Schwesterstelle im selben File.
- [x] `GeldtransitDialog` bezieht seinen Schlüssel über `useVorgangId`; die
      manuelle Rotation nach Erfolg ist entfernt. Damit nutzen alle sieben
      Aufrufstellen dieselbe Semantik, und `crypto.randomUUID()` kommt im
      Produktionscode nur noch in `use-vorgang-id.ts` vor.
- [x] Ein Integrationstest belegt für Bestellung und Direktverkauf: zweite
      Einreichung mit gleicher ID und geänderten Positionen bucht nicht und
      liefert `vorgang_daten_abweichend` — das ist der im Review beschriebene
      Fehlbetrags-Pfad.
- [x] `event_json_contract_test.go` ist unverändert grün.
- [x] `make verify` läuft grün.

---

## Phase 4: Geteilter Kassenzustand ist nie veraltet

Behebt den schwerwiegenden Aktualitäts-Befund und den geringfügigen Befund zu
den abgeschlossenen Kassensitzungen.

### Kontext

- `frontend/src/lib/queryClient.ts — createQueryClient(),
  AKTUALITAETSSCHWELLE_MS` — setzt `staleTime: 30_000` global; auf `main` gab es
  keine `defaultOptions`.
- `frontend/src/service/table/hooks.ts — useTischState(), useTischHistorie(),
  useMeineTischeState(), useEigeneUebersicht(),
  useAktiveTischeMitFavoriten(), useAktiveTische()` — die Schlüssel
  `'tisch-state'`, `'tisch-historie'` und `'aktive-tische'` sind String-Literale
  ohne exportierte Konstante.
- `frontend/src/service/TablePage.tsx — reload()` — lädt den offenen Tisch neu
  und invalidiert drei Übersichts-Schlüssel; die Präfixe `'tisch-state'` und
  `'tisch-historie'` werden nirgends invalidiert, auch nicht für den Ziel-Tisch
  einer Umbuchung.
- `frontend/src/admin/reporting/hooks.ts —
  ABGESCHLOSSENE_KASSENSITZUNGEN_KEY, REPORT_KEY` — werden nirgends
  invalidiert.
- `frontend/src/admin/kasse/KasseAbschliessenSection.tsx —
  handleAbschliessen()` und `frontend/src/admin/kasse/KassensitzungPage.tsx` —
  der Tagesabschluss löst nur ein `refetch` der offenen Kassensitzung aus.
- `frontend/src/admin/kasse/hooks.ts — useOffeneKassensitzung(),
  useKassenbestand(), useGeldtransitListe()`.

### Was gebaut wird

Die globale Aktualitätsschwelle wird umgedreht: Der Standard ist wieder `0`, und
`30_000` wird als benannte Konstante ausschließlich auf den Stammdaten-Hooks
gesetzt — Produkte, Benutzer, Betreiber, Kassenidentität, Druckstationen,
TSE-Konfiguration. Alles, was Kassen- oder Tischzustand liefert, lädt beim Mount
frisch. Damit verschwindet das Fenster, in dem ein Helfer einen
Tisch mit einem bis zu 30 s alten Saldo öffnet, ohne dass überhaupt nachgeladen
wird. Geschlossen ist das Fenster damit nicht: Ein zweiter Mount innerhalb der
`gcTime` (Standard 5 Minuten) rendert den zwischengespeicherten Saldo mit
entsperrten Reitern weiter, solange der ausgelöste Refetch läuft — und auch
dann, wenn er scheitert (`isLoadingError` bleibt false, weil Daten vorliegen).
Aus bis zu 30 s wird die Dauer des Refetch; der Rest steht unter „Risiken".

Die drei Tisch-Query-Schlüssel bekommen exportierte Konstanten, weil sie ab
jetzt an zwei Stellen benutzt werden. `TablePage.reload` invalidiert zusätzlich
die Präfixe für Tischzustand und Tisch-Historie; inaktive Queries werden dabei
nur als veraltet markiert und erzeugen keine zusätzlichen Requests. Damit ist
auch der Ziel-Tisch einer Umbuchung abgedeckt.

Der Tagesabschluss invalidiert im Erfolgspfad die Berichts-Schlüssel, damit ein
direkter Wechsel auf „Berichte & Export" die neue Sitzung zeigt und der
Archiv-Download nicht die vorherige liefert.

### Akzeptanzkriterien

- [x] `createQueryClient()` setzt keine `staleTime`-Voreinstellung mehr; die
      Retry-Policy bleibt unverändert.
- [x] `STAMMDATEN_AKTUALITAET_MS = 30_000` ist an einer Stelle definiert und
      wird ausschließlich von den Stammdaten-Hooks gesetzt:
      `useAllProdukte`, `useAktiveProdukte`, `useAllUsers`, `useBetreiber`,
      `useKassenidentitaet`, `useDruckstationen`, `useTSEKonfiguration`. Ein
      Kommentar an der Konstante nennt das Kriterium. Keine Tisch-Query gehört
      dazu — weder `useAllTische` noch `useAktiveTische`: Beide Nutzlasten tragen
      `saldoCents`, den offenen Tisch-Saldo der laufenden Kassensitzung. Bei
      `useAllTische` steuert er den Löschen-/Deaktivieren-Guard und die Kopfzeile
      „N mit offenem Saldo"; bei `useAktiveTische` wertet ihn heute niemand aus,
      die Regel gilt aber der Nutzlast statt dem aktuellen Verwendungszweck,
      damit niemand später versehentlich einen veralteten Saldo anzeigt.
- [x] Kein Hook, der Kassen- oder Tischzustand liefert, setzt eine
      Aktualitätsschwelle; `useVersion` behält `staleTime: Infinity`,
      `useLiveReporting` behält sein `refetchInterval`. Ein Test belegt, dass
      auch `useAktiveTische` beim zweiten Mount neu lädt.
- [x] Die Schlüssel `'tisch-state'`, `'tisch-historie'` und `'aktive-tische'`
      sind als benannte Konstanten in `frontend/src/service/table/hooks.ts`
      deklariert und werden überall über die Konstante referenziert. Exportiert
      werden nur `TISCH_STATE_KEY` und `TISCH_HISTORIE_KEY`, weil nur sie
      `TablePage.reload` invalidiert; `AKTIVE_TISCHE_KEY` bleibt dateilokal —
      ihre einzige Konsumentin (die Ziel-Tisch-Auswahl im
      `HistorieUmbuchungDrawer`) ist nur bei geöffnetem Drawer gemountet und
      lädt ohne Aktualitätsschwelle bei jedem Öffnen neu.
- [x] `TablePage.reload` invalidiert zusätzlich die Präfixe für Tischzustand und
      Tisch-Historie; ein Test belegt, dass nach einer Umbuchung auch der
      Ziel-Tisch als veraltet markiert ist.
- [x] Der Erfolgspfad des Tagesabschlusses invalidiert
      `ABGESCHLOSSENE_KASSENSITZUNGEN_KEY` und `REPORT_KEY`.
- [x] Ein Test belegt: Ein zweiter Mount eines Tischzustands lädt neu, ein
      zweiter Mount der Produktliste innerhalb von 30 s nicht.
- [x] Die Frontend-Suite läuft grün.

---

## Phase 5: Ladefehler und Hintergrund-Refetch sind unterscheidbar

Behebt den schwerwiegenden Befund zu den Fehlerzweigen und den geringfügigen
Befund zum Retry-Test.

### Kontext

- `frontend/src/service/TableSelectionPage.tsx — TableSelectionPage` — ein
  Frühausstieg auf `tischeError || alleTischeError || uebersichtError`
  überspringt Suchfeld, Favoritenliste, den Fußzeilen-Button „Alle Tische" und
  den `TischAuswahlDrawer` — also alle Einstiege in einen Tisch.
- `frontend/src/service/components/table/Bestellung.tsx — Bestellung`,
  `frontend/src/service/components/direktverkauf/Direktverkauf.tsx —
  Direktverkauf`,
  `frontend/src/service/components/direktverkauf/DirektverkaufHistorie.tsx —
  DirektverkaufHistorie` — je ein Frühausstieg auf dem `isError`-Flag.
- `frontend/src/service/TablePage.tsx — TablePage` — derselbe Frühausstieg auf
  `stateError || historieError`.
- `frontend/src/components/common/LadefehlerAlert.tsx — LadefehlerAlert`.
- `frontend/src/lib/queryClient.ts — QueryCache.onError` — toastet bereits jeden
  Query-Fehler mit fester Toast-ID und schweigt nur offline.
- `frontend/src/service/ServiceLayout.tsx` — bietet auf `/service/tische` keine
  Alternativnavigation.
- `frontend/src/service/TableSelectionPage.test.tsx` — der Retry-Test prüft nur
  das erneute Laden der eigenen Tische.

### Was gebaut wird

Die Fehlerzweige unterscheiden künftig ein gescheitertes Erstladen von einem
gescheiterten Hintergrund-Refetch: Nur wenn keine brauchbaren Daten im Cache
liegen, ersetzt der Ladefehler die Ansicht. Scheitert ein Refetch, während
gültige Daten vorliegen, bleibt die Ansicht stehen — die Meldung übernimmt der
bereits vorhandene zentrale Fehler-Toast, es braucht dafür keine neue
Komponente.

Auf der Tischauswahl wird der Ladefehler außerdem auf den Bereich beschränkt,
dessen Daten fehlen. Suchfeld, Fußzeilen-Button und `TischAuswahlDrawer` bleiben
in jedem Fall gemountet, damit ein Fehler der rein informativen eigenen
Übersicht nicht jeden Einstieg in einen Tisch blockiert.

### Akzeptanzkriterien

- [x] `TableSelectionPage`, `Bestellung`, `Direktverkauf`,
      `DirektverkaufHistorie`, `TablePage`, `TischAuswahlDrawer` und
      `HistorieUmbuchungDrawer` ziehen ihren Fehlerzweig nur bei einem
      gescheiterten Erstladen (`isLoadingError`), nicht bei einem gescheiterten
      Hintergrund-Refetch.
- [x] Ein Test je Ausprägung belegt: Bei gefülltem Cache und gescheitertem
      Refetch bleiben die Daten sichtbar; bei leerem Cache und gescheitertem
      Erstladen erscheint der `LadefehlerAlert`.
- [x] Auf `TableSelectionPage` ersetzt der Ladefehler nur den Bereich, dessen
      Query gescheitert ist; Suchfeld, Fußzeilen-Button „Alle Tische" und
      `TischAuswahlDrawer` sind in jedem Fehlerfall gemountet und bedienbar.
- [x] Ein Test belegt, dass bei ausschließlich gescheiterter eigener Übersicht
      der Einstieg in einen Tisch weiterhin möglich ist.
- [x] Der Retry-Test in `TableSelectionPage.test.tsx` prüft, dass alle drei
      Queries neu geladen werden; entfernt man eine davon aus `reload`, wird der
      Test rot.
- [x] Die Frontend-Suite läuft grün.

---

## Phase 6: Verwaiste Repository-Varianten und Dokumentation

Erledigt den verbliebenen Cleanup-Punkt und die Dokumentations-Nachträge.

### Kontext

- `backend/repository/kassenjournal_repo/repo.go — WriteUmbuchung(),
  WriteTischSessionEventsAtomic(), WriteEventWithDruckauftraege(),
  writeTischSessionEventsAtomic(), writeSingleEvent()` — mit dem Wechsel der
  Kommandos auf die `MitVorgang`-Varianten in Phase 2 und 3 verlieren die
  einfachen Varianten ihre Produktions-Aufrufer.
- `backend/repository/kassenjournal_repo/mock.go — MockRepo.WriteUmbuchung()` —
  hat schon jetzt keinen Aufrufer.
- `backend/repository/kassenjournal_repo/repo_test.go` — drei Aufrufe von
  `WriteUmbuchung`.
- `database/migrations/README.md` — Abschnitt „Vorversions-Pinning" nennt
  v0.14.0.
- `.github/workflows/ci.yml` — Job `upgrade-path` mit
  `PREVIOUS_VERSION: v0.14.0`; produktiv läuft v0.17.1.
- `docs/verfahrensdokumentation.md` — beschreibt die Verfahrens- und
  Herstellerdokumentation; die Idempotenz buchender Vorgänge ist dort bisher
  nicht beschrieben.

### Was gebaut wird

Nach Phase 2 und 3 schreibt jedes buchende Kommando über eine
`MitVorgang`-Variante. Die dadurch verwaisten exportierten Repository-Methoden
werden entfernt und ihre Tests auf die verbliebenen Varianten umgestellt. Wo
danach jeder Aufrufer einen Vorgang übergibt, nimmt die interne Funktion den
Vorgang als Wert statt als Zeiger. `WriteEvent` und `EroeffneKassensitzung`
bleiben erhalten — Kassensitzungs-Events tragen bewusst keinen
client-gelieferten Schlüssel.

Welche Methoden tatsächlich verwaist sind, wird per Suche über den
Produktionscode bestimmt und nicht aus diesem Plan übernommen; die Kandidaten
sind `WriteUmbuchung`, `WriteTischSessionEventsAtomic`,
`WriteEventWithDruckauftraege` und `MockRepo.WriteUmbuchung`.

Der CI-Pin der Vorversion wird auf das aktuell produktive Release gehoben, und
die Idempotenz buchender Vorgänge wird in der Verfahrensdokumentation
beschrieben — sie ist ein Merkmal, das die Vollständigkeit der Aufzeichnung
sichert und gehört damit in die Herstellerdokumentation.

### Akzeptanzkriterien

- [x] Eine Suche über den Produktionscode belegt für jede entfernte Methode,
      dass sie keinen Aufrufer mehr hat; die Testaufrufe sind auf die
      verbliebenen `MitVorgang`-Varianten umgestellt.
- [x] `writeTischSessionEventsAtomic` nimmt den Vorgang als Wert statt als
      Zeiger, sofern nach der Bereinigung jeder Aufrufer einen übergibt.
- [x] `WriteEvent` und `EroeffneKassensitzung` bleiben unverändert erhalten.
- [x] `PREVIOUS_VERSION` im Job `upgrade-path` steht auf v0.17.1; der Abschnitt
      „Vorversions-Pinning" in `database/migrations/README.md` nennt dieselbe
      Version.
- [x] Der `upgrade-path`-Job läuft grün: Migration, Boot und
      `make rebuild-projections` auf einer mit v0.17.1 befüllten Datenbank.
- [x] `docs/verfahrensdokumentation.md` beschreibt die Idempotenz buchender
      Vorgänge: Schlüssel vom Client, Nutzdaten-Bindung, drei Ausgänge, und
      dass die Zeile im selben Commit wie die Events entsteht.
- [x] `make verify` läuft grün; kein toter Code, keine ungenutzten Importe.

---

## Abnahme des Gesamtplans

- [x] `make check` und `make verify` laufen grün.
- [x] Die Frontend-Suite läuft grün.
- [ ] `make test-tse-live` läuft lokal mit echten fiskaly-Zugangsdaten grün —
      inklusive Teardown. Dieser Nachweis ist nicht in CI automatisierbar.
- [ ] Die TSE-Einrichtung ist gegen die fiskaly-TEST-Umgebung einmal
      durchgespielt: Das Ergebnis mit PUK und Admin-PIN erreicht die Oberfläche,
      auch wenn der Vorgang länger als 10 Sekunden dauert.
- [ ] Ein manueller Durchgang deckt den im Review beschriebenen Fehlbetrags-Pfad
      ab: Einreichung, verlorene Antwort, geänderte Auswahl, erneute Einreichung
      — das Ergebnis ist eine verständliche Meldung, keine zweite Buchung.
- [x] `docs/plans/review-client-server-robustheit.md` ist vollständig abgehakt.
