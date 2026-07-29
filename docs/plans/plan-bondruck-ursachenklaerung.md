# Plan: Ursache der fehlenden Bons klären

> Quell-PRD: n/a (aus Feldproblem (b) des ersten Festtags)
> Grundlage: `docs/plans/review-v0.17.2.md` (Feldproblem (b), zwingende Korrektur 3,
> Abschnitt „Schwerwiegend") und die vormalige Phase 6 aus
> `docs/plans/plan-v0.17.2-release.md`, die mit dem Abschluss jenes Plans hierher gezogen ist.

## Ziel

Klären, ob es das Problem überhaupt gibt und woran es liegt. Erst danach wird über den
archivierten Relay-Umbau entschieden — nicht umgekehrt.

Dieser Plan setzt **nicht** voraus, dass der Umbau zurückkommt. „Ursache nicht bestimmbar"
und „kein Umbau" sind gültige Ergebnisse und schließen den Plan ab.

## Ausgangslage

Tag 1 des Vereinsfests zeigte 1–3 Bons, die nicht gedruckt wurden. Tag 2 verlief ohne eine
einzige Druckerstörung. Das ist die gesamte Datenlage: zwei Tage, ein Störungstag, kein
reproduzierbarer Ablauf.

Kandidaten für die Ursache, keiner davon belegt: leere Papierrolle, WLAN-Aussetzer,
Stromsparmodus des Druckerrechners, Bedienung (der Bon wurde nie ausgelöst), oder
tatsächlich der Zustellpfad.

v0.17.2 hat den Zustellpfad **nicht** angefasst: `windows/relay/` ist gegenüber `v0.17.1`
unverändert. Es gilt weiterhin `windows/relay/main.go — druckeAuftrag()`: Papierprüfung per
DLE EOT, dann `conn.Write` — und der Auftrag gilt als gedruckt, sobald der Write fehlerfrei
zurückkehrt. Eine Quittung vom Drucker holt dieser Pfad nicht ein.

Der Umbau, der genau das ändern sollte, liegt im Tag `archiv/main-vor-v0.17.2`: die Commits
`f75bc9ab`, `7c2c37eb`, `80a6a819`, `b500915c`, `95324ef9`, `19a542d6`, `481b6169`. Er ändert
rund 400 Zeilen in `windows/relay/main.go` und rund 750 im zugehörigen Test.

## Resolved decisions

- **Erst Ursache, dann Entscheidung.** Phase 1 schreibt keinen Produktivcode. Der Grund für
  den v0.17.2-Schnitt gilt unverändert: Der Nutzen des Umbaus ist unbelegt, während das
  Review für genau diesen Umbau Verschlechterungen bestätigt hat.
- **Der Umbau bleibt archiviert, bis eine Ursache im Zustellpfad belegt ist.** Liegt die
  Ursache woanders — Papier, Netz, Bedienung —, ist er erledigt und wird nicht zurückgeholt.
- **Kommt er zurück, dann nicht als Block.** Vorher sind die beiden bestätigten Defekte zu
  beheben und die Migration auf die nächste freie Nummer zu ziehen (Phase 2).
- **Dieser Plan ist absichtlich selbsttragend.** `docs/plans/review-v0.17.2.md` wird mit
  v0.17.3 abgearbeitet und gelöscht (`docs/plans/plan-v0.17.3.md`, Phase 7); die beiden
  Defekte stehen deshalb hier vollständig, nicht als Verweis. Die Datei- und Zeilenangaben
  des Reviews zeigen ohnehin in den Archivbaum, nicht in diesen.

## Risiken

- **Die Datenlage kann die Frage nicht beantworten.** Ein still verlorener Bon ist in
  `druckauftraege` nicht von einem gedruckten zu unterscheiden (siehe Phase 1). Bleibt es
  dabei, ist „nicht bestimmbar" das Ergebnis — es darf nicht durch einen Umbau ersetzt
  werden, der dann wieder auf einer Vermutung stünde.
- **Belege können bereits gelöscht sein.** „Nochmal drucken" (`RetryDruckauftrag` in
  `backend/sqlc/queries/relay.sql`) setzt `versuche = 0` und `letzter_fehler = NULL`. Hat der
  Betreiber während des Festes nachgedruckt, ist die Fehlerhistorie dieser Zeilen weg.
- **Der Umbau kann Feldproblem (b) verschlimmern.** Das Review nennt den Quittungs-Fallback
  „die einzige Stelle, an der das Release Feldproblem (b) verschlimmern kann": Auf v0.17.1
  ging bei einem verschwundenen Drucker höchstens der eine gerade geschriebene Bon still
  verloren; mit dem Umbau kann dieselbe Lage eine ganze Sechsergruppe als `gedruckt`
  quittieren.
- **Das nächste Fest ist die bessere Datenquelle als eine Rekonstruktion.** Bleibt Phase 1
  ohne Befund, ist Abwarten die konservative Wahl: ein zweiter Störungstag liefert mehr als
  400 Zeilen neuer Transportcode auf Verdacht.

---

## Phase 1: Zustellhistorie der beiden Festtage auswerten

### Context

- `druckauftraege` auf der Produktivinstanz — `status`, `versuche`, `letzter_fehler`,
  `naechster_versuch_ab`, `erstellt_am`, `gedruckt_am`, `bon_art`, `referenz`, `ziel_ip`
  (`database/migrations/01_initial.up.sql` und `02_druckauftrag_backoff.up.sql`).
- **Was die Tabelle belegen kann:** Jeder gemeldete Fehlversuch trägt seinen Text.
  `windows/relay/main.go — druckeAuftrag()` erzeugt vier unterscheidbare Formen:
  `papier leer (status=0x..)` und `status-abfrage fehlgeschlagen: …` aus `checkPrinter()`,
  `nicht erreichbar: …` aus dessen `net.DialTimeout`, und `senden fehlgeschlagen: …` aus
  `sendToPrinter()`. Das trennt Papierrolle, Netz beziehungsweise Stromsparmodus und
  Schreibabbruch sauber.
- **Was sie nicht belegen kann:** Der Zustellpfad meldet Erfolg, sobald `conn.Write`
  fehlerfrei zurückkehrt; `MarkDruckauftragGedruckt` setzt die Zeile daraufhin endgültig auf
  `gedruckt`. Ein Bon, der im Sendepuffer verschwand, sieht in der Tabelle aus wie ein
  gedruckter.
- **Relay-Log: es gibt keine Datei.** `windows/relay/main.go` loggt ausschließlich über
  `log.Printf` nach stderr, ohne `log.SetOutput`. Der Verlauf existiert nur im
  Konsolenfenster von `jotti-relay.exe`, solange es offen ist.
- **Backend-Log rotiert eng.** Die Request-Middleware protokolliert jede Relay-Anfrage,
  `docker-compose.prod.yml` begrenzt aber auf `max-size: 5m` bei `max-file: 3`, und das Relay
  pollt alle 2 s (`windows/relay/main.go — defaultPollSeconds`). Schätzung: Das trägt eine
  Größenordnung von ein bis zwei Tagen — ob Tag 1 noch darin steht, ist zuerst zu prüfen,
  bevor damit geplant wird.

### What to build

Keine Codeänderung. Eine Auswertung mit schriftlichem Ergebnis.

Zuerst sichern, was flüchtig ist: die `druckauftraege`-Zeilen beider Festtage exportieren,
das Backend-Log ziehen, und beim Betreiber nachfragen, ob das Relay-Fenster noch steht.

Dann die Zeilen beider Tage gegenüberstellen: Verteilung über `status`, alle Zeilen mit
`versuche > 0` samt `letzter_fehler`, und die Zeitstempel rund um die Störungsmeldungen.
Zeigt sich dort eine der vier Fehlerformen, ist die Ursache benannt und dieser Plan
beantwortet. Steht dagegen alles auf `gedruckt` mit `versuche = 0`, sind Papier und
Erreichbarkeit zum Zustellzeitpunkt ausgeschlossen; übrig bleiben der stille Write-Erfolg und
die Bedienung — letztere prüfbar, indem die `referenz` der Bons gegen die zugehörigen
Kassenjournal-Einträge gehalten wird.

Tag 2 ist die Kontrollgruppe, aber nicht in der Datenbank: Was war am zweiten Tag anders —
frische Rolle, anderer Standort, anderes WLAN, anderer Drucker? Das ist eine Frage an den
Betreiber, und sie ist genauso viel wert wie die SQL.

### Acceptance criteria

- [ ] Die `druckauftraege`-Zeilen beider Festtage liegen exportiert vor, bevor irgendetwas an
      der Instanz geändert wird
- [ ] Für jede Zeile mit `versuche > 0` ist der `letzter_fehler` einer der vier bekannten
      Formen zugeordnet oder ausdrücklich als unbekannt vermerkt
- [ ] Der Unterschied zwischen Tag 1 und Tag 2 ist beim Betreiber erfragt und festgehalten
- [ ] Die Ursache ist benannt oder ausdrücklich als „nicht bestimmbar" dokumentiert —
      inklusive der Angabe, welche Quelle gefehlt hat
- [ ] Das Ergebnis steht schriftlich in diesem Plan, nicht nur in einem Chatverlauf

---

## Phase 2: Über den archivierten Relay-Umbau entscheiden

### Context

- Tag `archiv/main-vor-v0.17.2` — die sieben Commits aus der Ausgangslage.
- `backend/repository/druckauftrag_repo/repo.go — backoffDauer()` — auf **diesem** Stand
  5/15/30/60/180 s bei `MaxDruckversuche = 6`. Die Staffel bremst hier nur die betroffene
  Zeile: `GetOffeneDruckauftraege` filtert pro Zeile
  (`naechster_versuch_ab IS NULL OR naechster_versuch_ab <= NOW()`). Die warteschlangenweite
  Bremse kommt erst mit dem Umbau.
- `ls database/migrations/` — höchste vergebene Nummer ist `06_favoriten_cleanup.up.sql`.

### What to build

Eine begründete Entscheidung, kein Umbau auf Verdacht. Nur wenn Phase 1 eine Ursache im
Zustellpfad belegt, steht der Umbau überhaupt zur Debatte; sonst wird er mit schriftlicher
Begründung endgültig abgelegt.

Fällt die Entscheidung für ihn, sind vorher drei Dinge zu erledigen.

**Erstens der Quittungs-Fallback.** `pruefePapier` wirft weg, *ob* der Drucker geantwortet
hat. `holeQuittung` stützt sich im Timeout-Zweig genau auf diese Nachprüfung: Läuft auch sie
in den Timeout, liefert sie `ausgangUnbeantwortet, nil`. `stelleGruppeZu` wertet das
fehlerfreie Ergebnis als Zustellung (`ergebnis.gedruckteIDs = auftragsIDs(auftraege)`), und
das Repository setzt jede gemeldete ID endgültig auf `gedruckt`. Ein Drucker, der die erste
Papierprüfung noch beantwortet und danach ohne FIN/RST Strom oder WLAN verliert, bekommt so
eine ganze Gruppe von bis zu sechs Bons quittiert; sie tauchen weder im Poll noch unter
`GetFehlgeschlageneDruckauftraege` je wieder auf. Korrektur laut Review: `pruefePapier` eine
zweite Rückgabe geben (`geantwortet bool, err error`), das Ergebnis der **ersten** Prüfung in
`stelleGruppeZu` merken und an `holeQuittung` durchreichen — hat der Drucker dort geantwortet
und schweigt jetzt, ist er weg (`ausgangAbgebrochen` plus `gruppenFehlversuche`); hat er schon
dort geschwiegen, bleibt es bei `ausgangUnbeantwortet`, damit ein Modell ohne
Statusunterstützung benutzbar bleibt.

**Zweitens die Backoff-Flanke.** Der Umbau bringt eine warteschlangenweite Bremse mit: Sein
`GetOffeneDruckauftraege` überspringt die **ganze** Ziel-IP, solange irgendein offener Auftrag
dieses Druckers wartet. Die Staffel wurde nicht mitgezogen. Im häufigsten Störungspfad —
Papierrolle leer — eskalieren fünf Fehlversuche auf 180 s; legt der Helfer die Rolle mitten im
Fenster ein, bleibt der druckbereite Drucker bis zu ~172 s stumm, auch für inzwischen neu
eingereihte Bestellbons. Auf dem heutigen Stand druckt er beim nächsten Poll (~2 s) weiter.
Schlimmer bei der Diagnose-Bedienung „Nochmal drucken": `RetryDruckauftrag` setzt
`versuche = 0`, lässt der Zeile aber ihre alte, kleine `id`, die über `ORDER BY id ASC` vor
allen aktuellen Bons steht — ist der Drucker weiterhin nicht erreichbar, durchläuft sie die
volle Leiter erneut: 5+15+30+60+180 = 290 s Stillstand der gesamten Warteschlange dieses
Druckers, ohne jede Anzeige. Kleinste wirksame Maßnahme laut Review, ohne Schema und ohne
`sqlc`: `backoffDauer` auf 5/15/30/30/30 s deckeln — maximaler Stillstand 180 s → 30 s, der
„Nochmal drucken"-Fall 290 s → ~110 s. Soll das dokumentierte ~5-Minuten-Fenster bis
`fehlgeschlagen` erhalten bleiben, `MaxDruckversuche` anheben und `README.md` sowie
`docs/handbuch.md` mitziehen.

**Drittens die Migrationsnummer.** Der Block trägt
`06_druckauftrag_backoff_warteschlange.up.sql` — eine reine Kommentarkorrektur ohne
Schema-Eingriff, aber auf der Nummer, die v0.17.2 inzwischen mit `06_favoriten_cleanup.up.sql`
belegt. Sie muss auf die nächste freie Nummer (`07`). Git meldet dabei keinen Konflikt, weil
die Dateinamen verschieden sind; golang-migrate (v4.19.1, gepinnt in
`database/migrate/Dockerfile`) lehnt dagegen die **ganze** Quelle mit
`duplicate migration file: …` ab. Der Fehler entsteht beim Initialisieren des
Source-Treibers, es läuft also keine einzige Migration — und der Backend-Container startet
gar nicht erst, weil er auf `service_completed_successfully` des Migrationslaufs wartet.

### Acceptance criteria

- [ ] Es liegt eine schriftliche, begründete Entscheidung über den archivierten Umbau vor
- [ ] Ohne belegte Ursache im Zustellpfad kommt der Umbau nicht zurück
- [ ] Wird er zurückgeholt: Quittungs-Fallback und Backoff-Flanke sind vorher behoben, jeweils
      mit einem Test, der den bisherigen Fehlerfall abbildet
- [ ] Wird er zurückgeholt: seine Migration steht auf der nächsten freien Nummer, und der
      CI-Job `upgrade-path` läuft grün
