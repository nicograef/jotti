# Review: v0.17.2 (Freigabe für die Installation mitten im laufenden Fest)

> Gegenstand: `main` (HEAD `4040cd21`) gegen Tag `v0.17.1` (`3d334de2`),
> 180 Dateien, +15870/-1325. Bewertungsmaßstab ist Regressionsfreiheit gegenüber
> dem seit mehreren Tagen produktiv laufenden v0.17.1, nicht Vollständigkeit der
> Verbesserung.
> Durchgeführt am 2026-07-29.

## Urteil

**Nicht so installieren.** Fünf Korrekturen sind zwingend vorher nötig, drei davon
in Code, zwei in der Betriebsdokumentation.

Der schwerste Punkt ist kein Codefehler, sondern eine Lücke im Rollout: Der
dokumentierte Update-Ablauf tauscht das Print-Relay nicht aus, und der
Wire-Contract `/relay/poll` ist unverändert. Das alte Relay läuft nach dem Update
unbemerkt weiter — der gesamte Kern der Bon-Korrektur erreicht die
Produktivinstanz nicht. Feldproblem (b) bliebe unbehoben, während das Release
genau dafür eingespielt wird.

Der zweite Punkt ist eine harte Versionsinkompatibilität: `vorgangId` ist auf vier
buchenden Endpunkten neu Pflichtfeld. Jedes Helfer-Handy, das die Seite nach dem
Container-Tausch nicht neu lädt, kann nicht mehr kassieren, stornieren oder
umbuchen — Bestellen läuft weiter, die App wirkt also gesund, während der Geldpfad
tot ist.

Fünf zwingende Korrekturen, drei weitere schwerwiegende Befunde, elf geringfügige,
achtzehn geprüfte Cleanup-Punkte (alle nach dem Fest). Feldproblem (a) ist gelöst,
(c) ist gelöst, (b) ist so, wie das Release eingespielt würde, nicht gelöst.

## Methode

16 Experten-Linsen über den Diff (Relay-Transport, Relay-Rollout,
Druck-Warteschlange, Druck-Bestätigung, Idempotenz-Semantik,
Idempotenz-Transaktion, Frontend-Transport, Frontend-Service-Flow,
Migration/Daten, Freeze/Compliance, Regression-Sweep, Reporting-SQL,
Test-Substanz, Storno-Zuordnung, Middleware/Export, Betriebssicht), dazu eine
Vollständigkeitskritik über die von keiner Linse berührten Diff-Bereiche und
mehrere Cleanup-Durchläufe.

Jeder schwerwiegende Rohbefund wurde von drei unabhängigen Widerlegern mit
verschiedenen Linsen angegriffen; bestätigt gilt er erst, wenn höchstens einer
ihn widerlegen konnte. 26 bestätigte Befunde, 3 verworfene, 18 geprüfte
Cleanup-Vorschläge.

Alle Blocker- und Schwer-Befunde wurden für diesen Bericht eigenhändig am
Quellcode nachgeprüft; die Gliederung folgt der Ursache, nicht der Datei —
sechs Linsen hatten dieselbe `vorgangId`-Ursache gemeldet, vier dieselbe
Backoff-Ursache.

Nicht gelaufen: `make check`, `make verify`, Frontend-Suite, `make test-tse-live`.
Der Bericht stützt sich ausschließlich auf gelesenen Quellcode.

## Zwingende Korrekturen vor dem Einspielen

### 1. Der Update-Ablauf tauscht das Print-Relay nicht aus

`docs/leitfaden/aktualisieren.md:16-22` kennt drei Schritte: `jotti-stop.cmd`,
ZIP entpacken, `jotti-start.exe`. Das Relay kommt in der Datei nicht vor
(`grep -n relay docs/leitfaden/aktualisieren.md` → kein Treffer), und
`packaging/windows/jotti-stop.cmd:9` führt nur `%COMPOSE% down` aus, stoppt also
ausschließlich Container. Das Relay ist ein eigener nativer Prozess
(`packaging/windows/KURZANLEITUNG.md:49` „Für den Bondruck zusätzlich
**`jotti-relay.exe`** doppelklicken").

`git diff v0.17.1..HEAD -- backend/api/druck/ backend/api/relay.go` ist leer: Der
Wire-Contract ist unverändert. Das alte Relay pollt nach dem Update also
weiterhin erfolgreich, findet seine unveränderte `.env` unter
`%PROGRAMDATA%\jotti` und arbeitet mit dem alten Zustellverfahren weiter —
`git show v0.17.1:windows/relay/main.go:396-407`: zwei TCP-Verbindungen je Bon,
`_, err = conn.Write(data); return err`, also Erfolgsmeldung nach dem Schreiben
ohne jede Quittung. Genau dieser Pfad ist die dokumentierte Ursache der 1-3
fehlenden Bons.

Mit den Containern kommt nur die Backend-Hälfte mit (Warteschlangen-Backoff,
Migration 06). Die behebt Reihenfolge und Überholen, nicht die Falschquittierung.

- [ ] In `docs/leitfaden/aktualisieren.md` einen Schritt zwischen 1 und 3
      aufnehmen: „`jotti-relay.exe` beenden (Fenster schließen) und nach dem
      Start **aus dem neuen Ordner** erneut doppelklicken." Dasselbe in
      `packaging/windows/KURZANLEITUNG.md`.
- [ ] Das Beenden muss ausdrücklich **vor** dem Neustart stehen: `/relay/poll`
      vergibt keine Lease (`backend/sqlc/queries/relay.sql:5-20`, kein
      `FOR UPDATE`, kein Statuswechsel beim Poll). Zwei parallel laufende Relays
      bekommen dieselben offenen Aufträge und drucken jeden Bon doppelt.
- [ ] Kontrolle im Relay-Fenster: die Startzeile aus `windows/relay/main.go:193`
      muss die neue Version zeigen, und je Drucker muss die neue Zeile aus
      `main.go:341-342` erscheinen („Drucker %s: %d/%d Bons gesendet, Quittung %s,
      Dauer %s").

### 2. `vorgangId` ist neu Pflichtfeld auf vier buchenden Endpunkten

`backend/api/kasse/tischgeschaeft/http/command_handler.go:131` (`zahlungKassierenSchema`),
`:190` (`stornierungErteilenSchema`), `:197` (`bestellungUmbuchenSchema`) und
`backend/api/kasse/direktverkauf/http/command_handler.go:125`
(`direktverkaufStornierenSchema`) tragen je `"VorgangID": z.String().UUID().Required(),`.

In v0.17.1 gibt es das Feld nicht
(`git show v0.17.1:backend/api/kasse/tischgeschaeft/http/command_handler.go`,
`zahlungKassierenSchema` mit nur `TischID`/`Positionen`/`Kommentar`), und der
produktiv ausgelieferte Client sendet es nicht
(`git show v0.17.1:frontend/src/service/table/Zahlung.ts:20-24`).

`backend/api/helper/http.go:129-131` weist ab, bevor der Command läuft →
`SendClientError(w, "validation_error", issues)` → `:51` HTTP 400. Das Repo hält
das selbst fest:
`backend/api/kasse/tischgeschaeft/http/command_handler_test.go:152` Fall
„zahlung-kassieren ohne vorgangId" erwartet 400 + `validation_error`. Angezeigt
wird `frontend/src/lib/errorMessages.ts:97` „Bitte die Eingaben prüfen und erneut
versuchen." — ein Text, aus dem die Ursache nicht ableitbar ist; jede Wiederholung
scheitert identisch.

Nichts erzwingt das Neuladen: kein Service Worker, kein Versions-Handshake, kein
Auto-Reload (`grep -rn "serviceWorker|location.reload|APP_VERSION" frontend/src`
findet nur `components/common/ErrorBoundary.tsx:39` hinter einem React-Crash).
Das JWT überlebt den Neustart (`backend/domain/jwt/jwt.go:16`, 12 h;
`docker-compose.yml:61` `JWT_SECRET: ${JWT_SECRET}`), also gibt es auch keinen
401-Redirect, der eine Vollnavigation auslösen würde.

**Bestellen läuft weiter**, weil `bestellungAufnehmenSchema`
(`command_handler.go:78-83`) unverändert `BestellungID` verlangt und der alte
Client dieses Feld bereits schickte. Die Kasse nimmt also weiter Bestellungen an
und kann an keinem Tisch mehr kassieren — der Ausfall sieht aus wie ein
Einzelfehler, nicht wie eine Versionsinkompatibilität.

Zweite Ausprägung im Admin: Die Reporting-Antworten haben sich umbenannt
(`backend/api/reporting/http/query_handler.go:393` `kassiertCents` gegen
`git show v0.17.1:frontend/src/admin/reporting/types.ts:132` `zahlungenCents`,
und `stornierungenProServicekraft` ist entfallen). Das alte Zod-Schema wirft
`ResponseBodyError`; das Live-Dashboard des Rechners zeigt nur noch einen Fehler.

Dritte Ausprägung: `frontend/src/routes.ts:85-88` lädt alle Routen per
dynamischem `import()`. Nach dem Deploy sind die alten gehashten Chunks weg — eine
bis dahin nicht besuchte Route scheitert zusätzlich am fehlgeschlagenen Import.

- [ ] `VorgangID` in den vier Schemata optional annehmen
      (`z.String().UUID()` ohne `.Required()`, UUID-Prüfung nur bei nicht-leerem
      Wert) und im Handler bei leerem Wert serverseitig `uuid.NewString()`
      einsetzen. Ein unbekannter Schlüssel ergibt immer einen neuen Vorgang: Der
      alte Client verhält sich damit exakt wie unter v0.17.1 (keine Idempotenz,
      aber auch kein harter Fehler), der neue bekommt die volle Bindung.
- [ ] Zusätzlich, nicht ersatzweise: Reload jedes Geräts als Pflichtschritt in
      `docs/leitfaden/aktualisieren.md`. Nur der Reload behebt das Admin-Dashboard
      und die fehlenden Lazy-Chunks; die Schema-Lockerung deckt allein den
      Geldpfad und das Fenster zwischen Deploy und Reload.

### 3. Der stumm verschwundene Drucker bekommt bis zu sechs Bons quittiert

`windows/relay/main.go:431-437` wirft weg, **ob** der Drucker geantwortet hat:

```go
antwort := make([]byte, 1)
if _, err := conn.Read(antwort); err != nil {
    // Nicht jeder Drucker beantwortet die DLE-EOT-Statusabfrage …
    return nil
}
```

`holeQuittung` stützt sich im Timeout-Zweig genau auf diese Nachprüfung
(`main.go:506-512`): Läuft auch sie in den Timeout, liefert sie `nil`, und
`holeQuittung` gibt `ausgangUnbeantwortet, nil` zurück. `stelleGruppeZu` wertet
das fehlerfreie Ergebnis als Zustellung — `main.go:393`
`ergebnis.gedruckteIDs = auftragsIDs(auftraege)` —, und
`backend/repository/druckauftrag_repo/repo.go:137-141` setzt jede gemeldete ID
endgültig auf `gedruckt`.

Ablauf: Der Drucker beantwortet die **erste** Papierprüfung (`main.go:372`), kann
DLE EOT also nachweislich. Danach verliert er Strom oder fällt aus dem WLAN, ohne
FIN/RST. Die Writes der bis zu sechs Bons landen im Sendepuffer und melden Erfolg,
`GS r` wird nie ausgeführt, die Quittung läuft in den Timeout, die zweite
Papierprüfung ebenfalls → ganze Gruppe als `gedruckt`. Die Bons tauchen weder im
Poll noch unter `GetFehlgeschlageneDruckauftraege` je wieder auf.

Der Code **kann** diese Lage unterscheiden, wirft die Information aber weg. Der
Bestandstest deckt nur den legitimen Fall ab: `windows/relay/main_test.go:689`
(`druckerOptionen{papierstatusAntwort: papierOK}`) lässt den Testdrucker **beide**
Papierabfragen beantworten; `papierstatusAntwort(abfrage)` (`main_test.go:569-574`)
kann „erste beantwortet, zweite stumm" gar nicht ausdrücken.

Regressionsumfang präzise: Ist nur das Druckwerk tot und der TCP-Stack am Leben,
quittierte auch v0.17.1 stillschweigend — dort kein Unterschied. Verschwindet der
Drucker dagegen vom Netz, scheiterte auf v0.17.1 der `net.DialTimeout` des
nächsten Bons (`git show v0.17.1:windows/relay/main.go:362,397`) und erzeugte
einen sichtbaren Fehlversuch; still verloren ging höchstens der eine gerade
geschriebene Bon. Auf HEAD kann dieselbe Lage eine ganze Sechsergruppe quittieren.

Das ist die einzige Stelle, an der das Release Feldproblem (b) verschlimmern kann
— und es wird eingespielt, um (b) zu beheben.

- [ ] `pruefePapier` eine zweite Rückgabe geben (`geantwortet bool, err error`)
      statt die Antwortlage wegzuwerfen. Das Ergebnis der **ersten** Prüfung in
      `stelleGruppeZu` merken und an `holeQuittung` durchreichen: Hat der Drucker
      dort geantwortet und schweigt jetzt, ist er weg → `ausgangAbgebrochen` plus
      `gruppenFehlversuche`. Hat er schon dort geschwiegen, bleibt es bei
      `ausgangUnbeantwortet`. Ein Modell ohne Statusunterstützung bleibt damit
      benutzbar (`main_test.go:696-699` bleibt grün).

### 4. Der Geldtransit-Dialog leert sein Formular nicht mehr beim Öffnen

v0.17.1 hatte den Reset
(`git show v0.17.1:frontend/src/admin/kasse/GeldtransitDialog.tsx:66-69`):

```js
// Bei jedem Öffnen ein sauberes Formular (Betrag/Kommentar leer).
useEffect(() => {
  if (open) form.reset({ betragCents: 0, kommentar: '' })
}, [open, form])
```

Er ist in `9ef41ab6` entfallen. Heute geleert wird nur im Erfolgspfad
(`frontend/src/admin/kasse/GeldtransitDialog.tsx:113`), und der Schlüssel hängt am
leeren Betragsfeld (`:83-84`
`const geldtransitId = useVorgangId(betragCents === 0)`). Beide Richtungen teilen
dieselbe montierte Komponenteninstanz — `LaufenderBetriebSection.tsx:150-155`
wechselt nur `open` und `richtung`, die Richtung ist kein Formularfeld
(`GeldtransitDialog.tsx:96-101` übergibt sie als Prop).

Ablauf: „+ Geld einlegen", 50,00 und Kommentar eintragen, abbrechen (es wurde
nichts gesendet, es gibt keine Idempotenz-Zeile). Später „- Geld entnehmen": Der
Dialog öffnet mit Titel „Geld entnehmen", aber `betragCents=5000` und dem
Kommentar aus der Gegenrichtung steht noch da. Ein Tastendruck bucht eine
Entnahme, die niemand entnommen hat — im append-only Kassenjournal nur per
Gegenbuchung korrigierbar, der Kassensturz weist eine Differenz aus. Das
Weiterleben ist als Verhalten festgeschrieben
(`GeldtransitDialog.test.tsx:148` erwartet nach Abbrechen und Wiederöffnen
`toHaveValue('25,00')`).

Die Trade-off-Begründung im Code (`GeldtransitDialog.tsx:104-112`) trägt nicht:
Reset beim Öffnen und stabiler Schlüssel schließen sich nicht aus, sie wurden nur
an dieselbe Zustandsgröße gekoppelt.

- [ ] Reset beim Öffnen wiederherstellen und die Schlüsselrotation entkoppeln:
      `useVorgangId` hier fallenlassen, stattdessen wie in v0.17.1
      `useState(() => crypto.randomUUID())` halten und erst im Erfolgspfad per
      `setGeldtransitId(crypto.randomUUID())` rotieren. Damit bleibt beides
      erhalten: Der Wiederholversuch nach verlorener Antwort trägt denselben
      Schlüssel, und jeder neu geöffnete Dialog startet leer.
- [ ] Den Test „behält den Schlüssel über Schließen und Wiederöffnen" auf den
      Schlüssel statt auf den stehengebliebenen Feldwert umstellen und einen Test
      „Einlage-Entwurf abgebrochen → Entnahme-Dialog startet leer" ergänzen.

### 5. `admin/kasse-abschliessen` läuft ohne eigenes Zeitlimit und kann die Barriere hängen lassen

`frontend/src/admin/kasse/KasseBackend.ts:104-108` ruft `this.backend.post(...)`
ohne viertes Argument `optionen`, es greift also `REQUEST_TIMEOUT_MS = 8000`
(`frontend/src/lib/Backend.ts:66`, Abbruch in `:189-191`). Serverseitig reicht
`backend/api/kasse/kassenfuehrung/http/command_handler.go:154` `r.Context()`
durch, ohne `helper.ExtendWriteDeadline` und ohne Kontext-Entkopplung.

Der Statuswechsel auf `wird_abgeschlossen` steht per Autocommit außerhalb jeder
Transaktion (`application/command.go:380`). Der `defer`-Reset benutzt **denselben**
stornierten Kontext (`command.go:396`
`c.KassensitzungenRepo.SetKassensitzungOffen(ctx, ks.ZNr)`) und scheitert
ebenfalls; der Fehler wird nur geloggt. Die Sitzung bleibt in
`wird_abgeschlossen`, und ab da lehnt der Status-Guard
(`backend/repository/kassenjournal_repo/repo.go:427-429`) jedes
Nicht-Abschluss-Event ab: Bestellen, Kassieren, Stornieren, Umbuchen,
Direktverkauf, Geldtransit und Belegdruck antworten mit 409
`kasse_wird_abgeschlossen`. Alle Geräte zeigen „Bitte warten, bis der Abschluss
fertig ist" (`errorMessages.ts:53-54`).

Auflösbar ist das durch einen erneuten Abschluss-Versuch — der Dialog bleibt im
Fehlerfall offen. Der Schaden ist das Fenster dazwischen: Der Fehler-Toast meldet
nur einen fehlgeschlagenen Aufruf, nicht dass die Kasse jetzt gesperrt ist. Glaubt
der Admin, es sei nichts passiert, und geht erst das WLAN richten, steht der ganze
Betrieb.

Regression: Barriere und `defer`-Reset auf `r.Context()` existierten in v0.17.1
identisch (`git show v0.17.1:backend/api/kasse/kassenfuehrung/application/command.go:316,332`).
Neu ist allein der deterministische Client-Abbruch nach 8 s
(`git show v0.17.1:frontend/src/lib/Backend.ts` enthält weder `AbortController`
noch `signal` noch `setTimeout`), der den vorher nur bei echtem Verbindungsverlust
erreichbaren Pfad regelmäßig auslöst. Der Tagesabschluss läuft am selben Abend.

- [ ] Den `defer`-Reset in
      `backend/api/kasse/kassenfuehrung/application/command.go:394-400` auf einen
      vom Client-Abbruch entkoppelten Kontext legen:
      `resetCtx, resetCancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)`.
- [ ] In `KasseBackend.kasseAbschliessen` ein `{ zeitlimitMs: … }` oberhalb des
      Standardbudgets mitgeben und im Handler `helper.ExtendWriteDeadline(w, r, …)`
      beim Eintritt und direkt vor dem Schreiben setzen — `WriteTimeout` ist mit
      10 s absolut ab Request-Start (`backend/app/app.go:32`).

## Schwerwiegend (kann warten, aber nicht lange)

### Die warteschlangenweite Bremse verlängert die Wiederanlaufzeit nach jeder Druckerstörung

`backend/sqlc/queries/relay.sql:13-18` überspringt die **ganze** Ziel-IP, solange
irgendein offener Auftrag dieses Druckers wartet:

```sql
WHERE status = 'offen'
  AND ziel_ip NOT IN (
    SELECT ziel_ip FROM druckauftraege
    WHERE status = 'offen' AND naechster_versuch_ab > NOW()
  )
```

v0.17.1 filterte pro Zeile
(`git show v0.17.1:backend/sqlc/queries/relay.sql`:
`AND (naechster_versuch_ab IS NULL OR naechster_versuch_ab <= NOW())`). Die
Backoff-Staffel wurde nicht mitgezogen: `backend/repository/druckauftrag_repo/repo.go:22-37`
liefert unverändert 5/15/30/60/180 s bei `MaxDruckversuche = 6` (`repo.go:16`).

Vier Ausprägungen, eine Ursache:

1. **Papierrolle leer** (der häufigste Störungspfad). Fünf Fehlversuche
   eskalieren auf 180 s. Legt der Helfer die Rolle mitten im 180-s-Fenster ein,
   bleibt der druckbereite Drucker bis zu ~172 s stumm — auch für inzwischen neu
   eingereihte Bestellbons. Auf v0.17.1 druckte er beim nächsten Poll (~2 s)
   weiter.
2. **„Nochmal drucken"** — die Aktion, die der Betreiber wegen der fehlenden Bons
   ausführt. `relay.sql:47-49` setzt `versuche = 0` zurück, lässt der Zeile aber
   ihre alte, kleine `id`; durch `ORDER BY id ASC` (`:19`) steht sie vor allen
   aktuellen Bons. Ist der Drucker weiterhin nicht erreichbar, durchläuft sie die
   volle Leiter erneut: 5+15+30+60+180 = 290 s Stillstand der gesamten
   Warteschlange dieses Druckers, ohne jede Anzeige — die Oberfläche quittiert nur
   „Druckauftrag erneut eingereiht.". Auf v0.17.1 war dieselbe Bedienung folgenlos.
3. **Testbon.** `backend/api/druck/station/application/command.go:75-82` reiht ihn
   mit `naechster_versuch_ab = NULL` ein, die Sperre hält ihn trotzdem zurück,
   während `DruckstationConfigPage.tsx:236-237` „Testbon an „…" gesendet." meldet.
   Der Admin schließt auf einen defekten Drucker und drückt erneut; jeder Klick
   reiht einen weiteren Testbon ein.
4. **Sendefehler mitten in der Gruppe.** `windows/relay/main.go:377-386` lastet
   ihn nur dem einzelnen Bon an (`ergebnis.fehler = []fehlversuch{{ID: a.ID, …}}`),
   dessen Backoff aber die ganze Ziel-IP sperrt; die bereits geschriebenen Bons
   1..k-1 bleiben `offen` und drucken bei jedem der sechs Zustellversuche erneut.
   Der Kommentar dort begründet das mit „etwa eine ungültige Payload" — diesen
   Fall gibt es nicht, die Payload kommt base64-kodiert aus dem Backend
   (`arbeitsbon_policy.go:78`, `kassenbeleg_command.go:312`).

Gegenrechnung: Die Zustellrate ist gestiegen (v0.17.1 verbrannte in derselben
Störung Versuche quer durch die Warteschlange und kippte Bons auf
`fehlgeschlagen`), die Wiederanlaufzeit im häufigsten Störungspfad ist deutlich
schlechter. Das Verhalten ist bewusst gebaut und dokumentiert
(`README.md:80`, `docs/handbuch.md:318`).

- [ ] Kleinste wirksame Maßnahme, ohne Schema und ohne `sqlc`: die lange
      Backoff-Flanke deckeln — `backoffDauer` in
      `backend/repository/druckauftrag_repo/repo.go:22-37` auf 5/15/30/30/30 s.
      Maximaler Stillstand der Warteschlange fällt von 180 s auf 30 s, der
      „Nochmal drucken"-Fall von 290 s auf ~110 s. Soll das dokumentierte
      ~5-Minuten-Fenster bis `fehlgeschlagen` erhalten bleiben,
      `MaxDruckversuche` anheben und `README.md:80` sowie `docs/handbuch.md:318`
      mitziehen.
- [ ] Danach: `versuche = 0` aus `relay.sql:48` streichen (dann bremst ein
      erneut scheiternder Nachdruck die Warteschlange gar nicht mehr) und den
      Testbon von der Sperre ausnehmen. Beides erfordert `make sqlc` — nach dem
      Fest.
- [ ] Sendefehler wie einen Gruppenfehler behandeln
      (`gruppenFehlversuche(auftraege, err)` in `main.go:378-383`), damit der
      Präfix nicht bei jedem Retry erneut druckt.

### Warenkorb und Idempotenz-Schlüssel des Direktverkaufs liegen unter der Tab-Grenze

`frontend/src/service/DirektverkaufPage.tsx:88,101` rendert `verkaufenInhalt` in
`<TabsContent value="verkaufen">`. `Direktverkauf.tsx:43`
(`const { mengen, add, remove, reset } = useMengen<number>()`) und
`DirektverkaufAbschluss.tsx:78` (`const verkaufId = useVorgangId(...)`) liegen
darunter. Radix hängt inaktive Tab-Inhalte aus (`components/ui/tabs.tsx:171-184`
setzt kein `forceMount`), beides wird also beim Blick in die Historie verworfen.

`frontend/src/hooks/use-vorgang-id.ts:25-29` warnt genau davor: „Wird er unterhalb
einer Grenze aufgerufen, die aus- und wieder eingehängt wird (etwa im Inhalt eines
Radix-Tabs), bekommt eine unveränderte Auswahl nach dem Wiedereinhängen einen
neuen Schlüssel — aus einem Wiederholversuch würde eine zweite Buchung."

Ablauf: 2 Bier kassieren, Antwort geht verloren, Abbruch nach 8 s. Der Helfer
wechselt auf „Historie", um nachzusehen — Korb und `verkaufId` sind weg. Zurück
auf „Verkaufen" erfasst er dieselben 2 Bier erneut: neuer Schlüssel, neuer
`payload_hash`, zweites `direktverkauf-getaetigt:v1` mit eigener TSE-Signatur.
Zweimal 2 Bier im Journal, einmal kassiert.

Keine Regression — auf v0.17.1 lag der Schlüssel ebenfalls unter der Tab-Grenze
und die Doppelbuchung trat sogar ohne Tab-Wechsel ein. Es ist die an dieser Stelle
unvollständig angewandte neue Absicherung: Der Tisch-Pfad wurde gehoben und
getestet (`TablePage.tsx:193,247`, `TablePage.test.tsx:456` „behält die
bestellungId über einen Tab-Wechsel hinweg"), der Theken-Pfad nicht.

- [ ] `useMengen<number>()` und `useVorgangId(...)` nach `DirektverkaufPage`
      oberhalb von `<Tabs>` heben und als Props durchreichen — analog
      `TablePage.tsx:193/247`. Spiegel-Test in `DirektverkaufPage.test.tsx`
      ergänzen.

### Vier von sieben Aufrufstellen behandeln `vorgang_daten_abweichend` nicht

`frontend/src/lib/errorMessages.ts:111-112` behauptet: „Die Ansicht ist
aktualisiert und die Auswahl geleert — bitte nur die Differenz erneut erfassen."

Mit Bereinigung (3): `BestellungAbschluss.tsx:79`, `ZahlungAbschluss.tsx:103`,
`DirektverkaufAbschluss.tsx:99` — je ein `onCode`-Eintrag plus
`props.vorgangBereitsGebucht()`.

Ohne jede Bereinigung (4): `HistorieStornierungDrawer.tsx:72-81`,
`HistorieUmbuchungDrawer.tsx:130-141`, `DirektverkaufStornoDrawer.tsx:68-80` (alle
drei ohne `onCode`) und `GeldtransitDialog.tsx:86-89` — dort kennt der Hook das
Feld gar nicht: `frontend/src/hooks/use-form-action-submit.ts:8-16` deklariert nur
`form`, `actionLabel`, `byCode`, `fieldErrorsByCode`, `onSuccess`, während
`use-action-submit.ts:17` `onCode?: Record<string, () => void>` hat.

Folge: Der Toast behauptet eine Bereinigung, die nicht stattfindet. Der Schlüssel
rotiert nur über den Leerzustand (`use-vorgang-id.ts:44`), und die
Schlüsselprüfung läuft vor jeder fachlichen Prüfung
(`tischgeschaeft/application/command.go:436-438`). Der in der Meldung empfohlene
Weg — „nur die Differenz erneut erfassen" — ändert nur den Hash, nicht den
Schlüssel, und läuft zwangsläufig wieder in denselben 409. Ausweg wäre, die
Auswahl vollständig auf null zurückzunehmen oder den Drawer zu schließen; beides
nennt die Meldung nicht. Beim Geldtransit-Dialog hilft nur der erste Weg, weil
`LaufenderBetriebSection.tsx:150` ihn dauerhaft montiert hält.

Teilregression: Unter v0.17.1 endete derselbe Wiederholversuch in
`position_nicht_stornierbar` („Bitte Auswahl aktualisieren"), und ein Anpassen der
Auswahl führte zum Erfolg. Im Punkt „Weiterarbeiten nach dem Fehler ohne Verlassen
des Drawers" war v0.17.1 besser; die Verbesserung (kein zweites Buchen bei
identischer Wiederholung) bleibt davon unberührt.

- [ ] `onCode` aus `use-action-submit.ts` in `use-form-action-submit.ts`
      übernehmen und an allen vier Stellen dieselbe Bereinigung nachziehen wie in
      `BestellungAbschluss`/`ZahlungAbschluss`: Auswahl bzw. Formular leeren, die
      Ansicht neu laden, bei den drei Drawern zusätzlich schließen.

## Geringfügig

- [ ] Ein Wiederholversuch, der genau das Upgrade überspannt, endet bei
      Bestellung, Direktverkauf und Geldtransit in 409 `conflict` statt in der
      stillen Erfolgsantwort. Der v0.17.1-Fallback über die alten JSON-Unique-Indexe
      wurde ersatzlos entfernt
      (`git show v0.17.1:backend/api/kasse/tischgeschaeft/application/command.go:235`
      `EventExistsByTypeAndVorgangsID`; `grep -rn EventExistsByTypeAndVorgangsID
      backend/ --include=*.go` liefert heute nichts), die Indexe stehen aber
      unverändert in `01_initial.up.sql:161-171` und Migration 08 räumt sie nicht ab
      — `08_vorgang_idempotenz.up.sql:21` kennt und akzeptiert das ausdrücklich.
      `conflict` steht in keinem `onCode`, also räumt das Frontend nichts ab und
      jeder weitere Versuch endet identisch. Doppelt gebucht wird nichts.
      Minimal: `conflict` an denselben `onCode`-Handler hängen wie
      `vorgang_daten_abweichend`.
- [ ] `TSE_TIMEOUT_MS = 150_000` (`frontend/src/admin/tse/TSEBackend.ts:21`) ist
      aus dem **Schreibbudget** des Servers hergeleitet (`tseSetupWriteTimeout`,
      2 min) statt aus dessen **Arbeitsbudget**:
      `backend/api/fiskal/setup/http/command_handler.go:66`
      `const tseSetupLebenszyklusTimeout = 10 * time.Minute`, mit einem im Code
      bezifferten Worst Case von rund 7,5 min (`:54`). Die Schreibfrist wird per
      `ExtendWriteDeadline` unmittelbar vor dem Schreiben neu gesetzt (`:238`),
      begrenzt den Antwortzeitpunkt also nicht. Bricht der Client bei 150 s ab,
      läuft der Handler dank `context.WithoutCancel` (`:102-104`) zu Ende und
      persistiert `tssId`/`clientId`, aber PUK und Admin-PIN gehen in eine
      geschlossene Verbindung und werden nirgends gespeichert (`setup.go:381`).
      Mitten im Fest irrelevant (TSE ist eingerichtet), aber die im Vorreview
      abgehakte Blocker-Korrektur ist an dieser Stelle unvollständig. Fix:
      `TSE_TIMEOUT_MS = 630_000` und den Herleitungskommentar korrigieren.
- [ ] Der Kommentar ist Teil des serverseitigen Nutzdaten-Hashes
      (`tischgeschaeft/application/command.go:290,560`), lebt aber kürzer als der
      Schlüssel: `BestellungAbschluss.tsx:51` und `ZahlungAbschluss.tsx:61` halten
      ihn unterhalb der Tab-Grenze, während Korb und Schlüssel in `TablePage`
      liegen. Ein aus Nutzersicht unveränderter Wiederholversuch nach einem
      Tab-Wechsel wird als abweichende Nutzdaten abgelehnt statt still bestätigt;
      zweite Ausprägung: Ein Tab-Wechsel verliert den getippten Kommentar, während
      der Korb sichtbar stehen bleibt. Kein Rückschritt gegenüber v0.17.1 (dort
      verwarf der Tab-Wechsel beides). Fix: Kommentar mit Korb und Schlüssel nach
      `TablePage` heben.
- [ ] `database/migrations/README.md:19` behauptet, eine in `BEGIN/COMMIT`
      geklammerte Migration hinterlasse bei Fehlschlag keinen `dirty`-Zustand. Das
      ist falsch: golang-migrate (gepinnt auf `v4.19.1`,
      `database/migrate/Dockerfile:3`) ruft `SetVersion(target, true)` **vor** dem
      Lauf in einer eigenen Transaktion und committet. Scheitert Migration 08 auf
      einer Bestandsinstanz, rollt das Schema sauber zurück, `schema_migrations`
      steht aber committet auf `(8, dirty=true)`; jeder weitere `migrate up` bricht
      mit `ErrDirty` ab und der Backend-Container startet gar nicht
      (`docker-compose.release.yml:94-96`
      `condition: service_completed_successfully`). Wer den Fehlschlag anhand der
      README einordnet, erwartet einen sauberen Neuversuch und braucht in Wahrheit
      `migrate force 7` oder den Backup-Restore. Keine Regression (Zeile
      unverändert seit vor v0.17.1), aber genau die Doku, die beim Mid-Fest-Update
      gebraucht wird.
- [ ] `README.md:78` behauptet „erst diese Quittung macht sie `gedruckt`". Der
      Code widerspricht: `windows/relay/main.go:511` `return ausgangUnbeantwortet, nil`
      plus `:393` `ergebnis.gedruckteIDs = auftragsIDs(auftraege)` — die Gruppe
      gilt ohne Quittung als zugestellt. Der eigene Folgesatz und
      `docs/handbuch.md:316` sagen es richtig. Ein Betreiber, der wegen fehlender
      Bons ins README schaut, schließt daraus fälschlich, dass jeder als `gedruckt`
      markierte Bon bestätigt wurde. Halbsatz streichen.
- [ ] Migration 07 (`07_favoriten_cleanup.up.sql:15-19`) hat keinen Test und
      keinen CI-Schritt, der ihre `WHERE`-Bedingung diskriminiert. Ein repo-weiter
      grep nach `favoriten_cleanup` findet außerhalb von `.git` nur zwei
      Plandokumente. Der `upgrade-path`-Job (`.github/workflows/ci.yml:453-505`)
      migriert v0.17.1 → HEAD, prüft aber keinen Datenbestand; die geseedeten
      v0.17.1-Demodaten enthalten zwar einen gelöschten Tisch, aber keine
      Favoriten-Zeile darauf. Mutation `WHERE NOT EXISTS` → `WHERE EXISTS` lässt
      alle Tests und CI-Jobs grün und löschte auf der Produktivinstanz **alle**
      Markierungen aller Servicekräfte. Die vorliegende Bedingung ist korrekt
      gelesen; ungeprüft ist ausschließlich der einmalige Datenbestands-Eingriff.
      Integrationstest ergänzen.
- [ ] Die Reihenfolge der Geldtransit-Idempotenzprüfung (Vorprüfung **vor** der
      fachlichen Validierung, `kassenfuehrung/application/command.go:243-269`) ist
      von keinem Test gedeckt: Entfernt man `DetermineVorgangStatus` samt beider
      Status-Zweige, bleiben alle vier Duplikat-Tests grün, weil der Schreibpfad
      dieselben Sentinels liefert. Das einzige exklusiv von der Vorprüfung
      gelieferte Verhalten — Wiederholversuch nach Schließen der Kassensitzung —
      prüft kein Test. Für `ZahlungKassieren` ist dieselbe Reihenfolge
      diskriminierend gepinnt (`command_idempotenz_test.go:330-335`). Kein
      Laufzeitdefekt.
- [ ] `frontend/src/service/TablePage.tsx:140` bindet die Produkte-Ladeflagge als
      `isPending`, während die beiden Schwestern in derselben Komponente
      `stateLoading` (`:136`) und `historieLoading` (`:146`) heißen. Ausgerechnet
      die unbenannte entscheidet 95 Zeilen später (`:234`), ob der Bestell-Korb
      bereinigt oder unangetastet bleibt.

## Cleanup

18 geprüfte, verhaltenserhaltende Vorschläge. **Keiner davon gehört in das
Mid-Fest-Release** — sie sind reines Umbau-Risiko auf einem Stand, der in eine
laufende Produktion geht. Nach dem Fest abarbeiten. Ausnahme: der README-Punkt
oben, der eine falsche Betriebsaussage korrigiert, und der deshalb unter
„Geringfügig" steht.

Backend/Relay:

- [ ] `windows/relay/main.go:335` `zustelleGruppe` ist eine Namens-Permutation von
      `stelleGruppeZu` (`:362`) in Aufrufer/Aufgerufener-Beziehung; `zustelle…` ist
      zudem keine gültige Bildung des trennbaren Verbs. In `neueGruppenZustellung`
      umbenennen (Fabrik-Rolle statt Handlung).
- [ ] `windows/relay/main.go:328` `gruppenErgebnis.fehler []fehlversuch` heißt in
      `zyklusErgebnis` `fehlversuche` (`:125`), obwohl der Wert direkt in den
      anderen fließt; zugleich meint `fehler` in Go üblicherweise einen einzelnen
      `error`. Zu `fehlversuche` umbenennen.
- [ ] `backend/api/kasse/tischgeschaeft/application/command.go:119-129`
      `writeEventWithDruckauftraege` ist eine reine Weiterleitung mit neun
      Parametern und genau einem Aufrufer (`:347`), während die Schwesterstellen
      (`:195`, `direktverkauf/application/command.go:191,322`) `writeEventOCC`
      direkt mit Closure aufrufen. Inlinen. (Zwei Linsen gemeldet, gleiche Ursache.)
- [ ] `backend/api/kasse/tischgeschaeft/application/command.go:89,116`
      `writeEventOCC` gibt eine Event-ID zurück, die kein Aufrufer liest. Auf
      `error` reduzieren, wie der Schwester-Helfer
      `direktverkauf/application/command.go:350`.
- [ ] `backend/sqlc/queries/reporting.sql`: Alias `u` bezeichnet in derselben
      Anweisung zwei Relationen — CTE `ursprung` (`:104-116`) und Tabelle `users`
      (`:144,150,156`). Im CTE nach `urs` umbenennen, dann `make sqlc`.
      (`bu` → `u` aus dem Ursprungsvorschlag **nicht** übernehmen, das wäre reine
      Zusatzänderung.)
- [ ] `backend/sqlc/queries/reporting.sql:105` nutzt einen Komma-Join für
      `jsonb_array_elements`, 15 Zeilen tiefer steht `CROSS JOIN LATERAL` (`:120`,
      ebenso `:174,216`). Angleichen, dann `make sqlc`.

Frontend:

- [ ] `frontend/src/service/components/table/ZahlungAbschluss.tsx:77-80,104-107,112-115`:
      derselbe Vierfach-Reset dreimal wörtlich. Lokales `eingabenLeeren()`
      einführen. Wörtlich dieselbe Dreifach-Kopie in
      `DirektverkaufAbschluss.tsx:66-69,100-103,108-111` — dort ein eigenes
      lokales Helferlein, keine geteilte Abstraktion über Dateigrenzen.
- [ ] `frontend/src/service/TablePage.tsx:140` `isPending` → `produkteLoading`
      (siehe „Geringfügig"); optional derselbe Handgriff in
      `DirektverkaufPage.tsx:21/59`.
- [ ] `frontend/src/admin/reporting/StornoServicekraft.tsx:35-43`: `className`-Prop
      und `mb-3`-Default sind tot — die einzige Aufrufstelle
      (`LiveReportingSection.tsx:288-291`) überschreibt den Default immer. Prop und
      `cn`-Import entfernen, Klassen literal schreiben.

Tests:

- [ ] `backend/repository/druckauftrag_repo/repo_test.go:225-231,314-320,382-388,394-406`
      bauen den im selben Change eingeführten Helfer `assertOffeneAuftraege`
      (`:767-782`) von Hand nach. Ersetzen.
- [ ] `backend/api/kasse/direktverkauf/application/idempotenz_integration_test.go`:
      13 Zählabfragen inline, obwohl der Zwillingstest
      (`tischgeschaeft/.../idempotenz_integration_test.go:239-246`) im selben
      Change `countSignaturauftraege`/`countByType`/`countVorgaenge` anlegt.
      Nachziehen, auch den vorbestehenden Block `:114-120`.
- [ ] `backend/api/kasse/direktverkauf/application/idempotenz_integration_test.go:246-259,306-319`:
      derselbe 14-zeilige `positionId`-Leseblock inkl. anonymer Struct doppelt.
      In `erstePositionID(...)` ziehen, Vorbild `offeneRefs`
      (`tischgeschaeft/.../idempotenz_integration_test.go:258`).
- [ ] `backend/api/fiskal/export/http/handler_test.go:84-107,115-138`: zwei Tests
      mit zwölf identischen Assertion-Zeilen, einziger Unterschied ist die
      Middleware-Kette (`:90` gegen `:121`). Assertions in einen Helfer ziehen; die
      Schwesterdatei `setup/http/write_deadline_test.go:69` löst denselben Fall über
      eine Tabelle.
- [ ] `frontend/src/service/components/table/HistorieUmbuchungDrawer.test.tsx:246-254,302-310,345-353`
      umgehen den seit v0.17.1 bestehenden `renderDrawer`-Helfer (`:70-83`), den
      der vierte im selben Change entstandene Test (`:387`) benutzt. Ersetzen.
- [ ] `frontend/src/admin/reporting/LiveReportingSection.test.tsx:19-23`: Der
      `stornierungen`-Parameter ist durch `overrides: Partial<LiveReportingData>`
      bereits abgedeckt und erzwingt an zwei von acht Aufrufen (`:192,323`) ein
      bedeutungsloses `[]`. Streichen, der einzige echte Nutzer (`:217`) übergibt
      benannt.
- [ ] `frontend/src/components/common/OfflineBanner.test.tsx:15-19`:
      Einzeiler-Helfer `stubNavigatorOnLine` mit genau einem Aufruf (`:52`).
      Inlinen, den erklärenden jsdom-Kommentar mitziehen.

## Feldprobleme

### (a) Blockierte Tischübersicht durch gelöschte Tische in den Favoriten — **gelöst**

Drei Schichten, alle verifiziert:

- `database/migrations/07_favoriten_cleanup.up.sql:15-19` räumt den Bestand ab
  (`DELETE FROM tisch_favoriten f WHERE NOT EXISTS (SELECT 1 FROM tische t WHERE
  t.id = f.tisch_id AND t.status != 'deleted')`).
- `backend/repository/tisch_repo/repo.go:157-172` `DeleteTableMitFavoriten` löscht
  Tisch und Markierungen in **einer** Transaktion, verhindert also neue Waisen.
- `backend/api/kasse/tischgeschaeft/application/query.go:151-160` überspringt einen
  unauflösbaren Favoriten und loggt ihn, statt die ganze Übersicht scheitern zu
  lassen.

Abgesichert durch `tischgeschaeft/application/query_test.go:153` und
`tisch_repo/repo_test.go:119`. Einziger offener Punkt: Der einmalige
Datenbestands-Eingriff von Migration 07 ist selbst ungetestet (siehe
„Geringfügig") — die Bedingung ist am Text korrekt, aber sie läuft auf einer
Instanz mit aufbewahrungspflichtigen Daten. Vor dem Einspielen ein Backup
sicherstellen (das erledigt `jotti-start.exe` automatisch, siehe
`docs/leitfaden/aktualisieren.md`).

### (b) 1-3 Bons, die nicht gedruckt wurden — **so, wie es eingespielt würde: nicht gelöst**

Drei unabhängige Gründe:

1. Das alte Relay läuft weiter (zwingende Korrektur 1). Ohne Austausch bleibt
   genau der Pfad aktiv, der die fehlenden Bons erzeugt: Erfolgsmeldung nach
   `conn.Write` ohne Quittung.
2. Auch mit neuem Relay kann ein mitten in der Gruppe verschwundener Drucker bis
   zu sechs Bons als `gedruckt` quittiert bekommen (zwingende Korrektur 3) — mehr
   als auf v0.17.1, wo derselbe Ausfall höchstens einen Bon still verschluckte.
3. Die Wiederanlaufzeit nach der häufigsten Störung (Papier leer) ist von ~2 s auf
   bis zu ~172 s gestiegen, und die Diagnose-Bedienung „Nochmal drucken" kann die
   Warteschlange 290 s stilllegen (Abschnitt „Schwerwiegend").

Was das Release an (b) tatsächlich verbessert: Ein Drucker, der nur **eine**
Verbindung gleichzeitig annimmt, verliert keine Folgeaufträge mehr
(`main.go:362-368`, eine Verbindung je Ziel-IP und Zyklus statt zwei je Bon), die
Reihenfolge bleibt erhalten, und ein bestätigender Drucker liefert jetzt einen
echten Nachweis (`ausgangBestaetigt`). Mit den Korrekturen 1 und 3 ist (b)
adressiert; ohne sie nicht.

### (c) Stornierungen dem falschen Servicekraft zugeordnet — **gelöst**

`backend/sqlc/queries/reporting.sql:107-152` löst die Zuordnung über drei
Ursprungspfade auf: `stornierung-erteilt:v1` über `zahlungId` auf
`zahlung-kassiert:v1`, `direktverkauf-storniert:v1` über `verkaufId` auf
`direktverkauf-getaetigt:v1`, `bestellung-korrigiert:v1` positionsweise über
`bestellung-aufgenommen:v1`. Lässt sich nichts auflösen, fällt die Zeile per
`COALESCE` (`:147-152`) auf den Akteur zurück — das deckt sich mit
`docs/handbuch.md:501` („fällt **die Zeile** auf den Akteur zurück").

Der einzige Einwand aus dem Review — der Rückfall greife pro Event statt pro
Position und lasse bei gemischten Quellen Positionen ohne Zuordnung — ist
widerlegt: Der einzige schreibende Pfad ist an genau einen Historien-Eintrag
gebunden (`HistorieStornierungDrawer.tsx:36,48,88`, einziger Aufrufer
`TischHistorie.tsx:210`). Ist die Quelle eine Bestellung, sind **alle** Positionen
auflösbar; ist sie ein Umbuchungs-Zugang, **keine**, weil
`backend/domain/kasse/tisch_session_events.go:216-218` jeder Zielposition eine
frische ID gibt. Der gemischte Fall ist nur durch einen handgebauten API-Aufruf
außerhalb des Clients erzeugbar.

Kein Geldfehler: `reporting.sql:139` setzt `bar_rueckgabe` für
`bestellung-korrigiert:v1` auf `false`, und
`backend/api/reporting/application/query.go:113-115` addiert Beträge nur bei
`s.BarRueckgabe`. `TischSessionSubject` (`subject.go:16`) bindet Storno und
zugehörige Zahlung an dieselbe `kassensitzung_nr`, eine Warenrücknahme kann also
nie mehr als einen Betroffenen erzeugen.

Einschränkung, die kein Defekt ist: Die SQL ist latent fragil — würde die Auswahl
je quellenübergreifend, wäre der Rückfall zu grob. Kein Grund, mitten im Fest die
Query anzufassen.

Nebenwirkung, die zur zwingenden Korrektur 2 gehört: Die Umbenennung der
Reporting-Antworten (`zahlungenCents` → `kassiertCents`, Wegfall von
`stornierungenProServicekraft`) bricht den **alten** Admin-Tab, bis er neu geladen
wird.

## Einspielen

Reihenfolge für den Betreiber, wenn die zwingenden Korrekturen umgesetzt sind:

1. **Bedienpause abstimmen.** Nicht während der Hauptlast. Der Kassenleitung
   ansagen, dass alle Handys gleich neu geladen werden müssen.
2. **`jotti-relay.exe` beenden** (Fenster schließen). Vor dem Start des neuen
   Relays — zwei parallele Relays drucken jeden Bon doppelt, weil `/relay/poll`
   keine Lease vergibt (`backend/sqlc/queries/relay.sql:5-20`).
3. **`jotti-stop.cmd`** doppelklicken. Niemals `docker compose down -v`.
4. **Neues Release-ZIP entpacken**, **`jotti-start.exe`** starten. Der Starter
   sichert die Datenbank automatisch vor der Migration; Migration 07 räumt
   verwaiste Favoriten-Markierungen ab, Migration 08 legt `vorgang_idempotenz` an.
   Beide sind additiv und rühren kein Kassenjournal an.
5. **Migrations-Log prüfen.** Läuft `migrate` non-zero durch, startet das Backend
   gar nicht (`docker-compose.release.yml:94-96`). `schema_migrations` steht dann
   auf `dirty=true` und braucht `migrate force 7` oder `jotti-restore.cmd` — die
   README-Regel, die etwas anderes behauptet, ist falsch (siehe „Geringfügig").
6. **`jotti-relay.exe` aus dem NEUEN Ordner starten.** Kontrolle im
   Relay-Fenster: Startzeile mit der neuen Version (`windows/relay/main.go:193`)
   und je Drucker die neue Zeile „Drucker … Bons gesendet, Quittung …, Dauer …"
   (`main.go:341-342`). Fehlt die Quittungszeile, läuft noch das alte Relay.
7. **Jedes Helfer-Handy und den Rechner-Tab einmal neu laden** (Pull-to-Refresh
   bzw. Tab schließen und öffnen), **bevor** wieder kassiert wird. `index.html`
   wird mit `Cache-Control "no-cache"` ausgeliefert (`frontend/nginx.conf:21`),
   der Reload holt das neue Bundle zuverlässig — ausgelöst wird er von nichts.
8. **Rauchtest:** an einem Tisch bestellen, kassieren, stornieren; einen
   Direktverkauf tätigen; im Admin einen Testbon drucken; das Live-Dashboard
   öffnen (dort muss die Storno-Zuordnung je Servicekraft erscheinen).

**Ja, das Print-Relay muss mit ausgetauscht werden** — sonst ist der Hauptzweck
des Releases nicht erfüllt.

**Rollback ist unsauber.** `backend/api/helper/http.go:107`
`decoder.DisallowUnknownFields()`: Ein zurückgerolltes v0.17.1-Backend antwortet
auf das neue Bundle mit 400 `invalid_json`. Nach einem Rollback müsste jedes Gerät
erneut neu geladen werden, und der Starter verweigert Downgrades ohnehin
(`docs/leitfaden/aktualisieren.md`). Der Rückweg ist `jotti-restore.cmd`, nicht
das Zurückkopieren des alten ZIPs.

## Nicht bestätigt

Damit sie nicht erneut aufgeworfen werden.

- **Bereits geschriebene Bons müssten bei einem Schreibfehler mitten in der Gruppe
  quittiert werden.** Die beiden tragenden Prämissen sind nicht gleichzeitig
  erfüllbar: Im einzigen realistischen Nicht-Hard-Error-Fall (Write-Timeout, also
  voller Empfangspuffer) sind die Bons 1..k-1 gerade **nicht** gedruckt
  (`main.go:453-456`, `main_test.go:820-822` „Es darf die Gruppe nicht bestätigen —
  der Drucker hat nichts verarbeitet."). Wo sie wirklich gedruckt haben
  (Papierende, Stromverlust), scheitert `holeQuittung` bzw. meldet `pruefePapier`
  „papier leer" — die vorgeschlagene Korrektur griffe dort nicht. Die einzige
  deterministische Auslöserquelle des Vorschlags (ungültiges Base64,
  `main.go:458-461`) ist im Produktivdatenbestand unerreichbar, weil jede Payload
  per `base64.StdEncoding.EncodeToString` im Backend entsteht
  (`arbeitsbon_policy.go:78,89,144,162`, `kassenbeleg_command.go:312`,
  `station/application/command.go:77` — alle `NeuerDruckauftrag`-Konstruktionsstellen).
  Die Korrektur würde zudem auf dem `ausgangUnbeantwortet`-Pfad genau die
  Falschquittierung zurückholen, die das Release beseitigt.
- **Der Upgrade-Absatz in Migration 08 (`:17-23`) gelte fälschlich für alle sieben
  Vorgangsarten.** Zeile 17 beginnt mit „Für Bestandsdaten gilt das nicht." —
  anaphorische Negation des vorangehenden Satzes, der ausschließlich die drei
  namentlich genannten Indexe betrifft (`:10-15`). Für die vier anderen Arten ist
  das beschriebene Szenario nicht bloß unzutreffend, sondern unmöglich: Ein
  Alt-Client-Retry scheitert dort an der Body-Validierung, erreicht also nie einen
  Event-Insert. Text und Wirklichkeit decken sich vollständig.
- **Der Rückfall der Storno-Zuordnung greife pro Event statt pro Position und
  lasse umgebuchte Positionen ohne Zuordnung.** Der gemischte Fall ist über den
  Client nicht erzeugbar (siehe „Feldprobleme (c)"). Beide Dokumente
  (`docs/handbuch.md:501`, `docs/language.md:168`) beschreiben das implementierte
  Verhalten korrekt.

## Was dieses Review nicht abgedeckt hat

- **Keine Messung realer Laufzeiten.** Ob `admin/kasse-abschliessen` über das
  Vereins-WLAN tatsächlich 8 s überschreitet, ist aus Timeouts abgeleitet, nicht
  gemessen. Ebenso die Backoff-Zeiten im Feld.
- **Kein Test am laufenden Stack und keine Gerätetests.** Ob Caddy den
  Client-Abbruch in jeder Konstellation an den Go-Handler durchreicht und wie sich
  die Handys nach einem Container-Tausch real verhalten, wurde nicht verifiziert.
- **Kein Test mit echter Hardware.** Alle Relay-Aussagen stammen aus Quellcode und
  den Test-Druckern in `main_test.go`; ob ein konkretes Bondrucker-Modell DLE EOT
  n=4 und GS r 1 so beantwortet wie angenommen, wurde nicht gemessen.
- **Keine Testläufe.** Weder `make check` noch `make verify`, Frontend-Suite oder
  `make test-tse-live` liefen in dieser Sitzung. Die Mutationsaussagen der
  Testbefunde stammen aus den Verifikationsläufen der Linsen.
- **Keine Prüfung der DSFinV-K-Feldabbildung** und keine erneute Prüfung der in
  `docs/plans/review-client-server-robustheit.md` bereits abgehakten Punkte,
  außer wo ein Befund die dortige Korrektur als unvollständig belegt
  (`TSE_TIMEOUT_MS`).
- **Kein Test der Migrationen gegen einen echten Produktivbestand.** Migration 07
  und 08 wurden am Text gelesen; ein Probelauf gegen eine Kopie der laufenden
  Instanz hat nicht stattgefunden und wäre vor dem Mid-Fest-Update die billigste
  Absicherung.
