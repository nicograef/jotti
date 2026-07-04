# PRD: TSE-Signatur über Outbox und Signatur-Worker

> Umbau der TSE-Integration vom synchronen Signieren im Kassier-Pfad auf ein
> Outbox-Modell: Jeder signaturpflichtige Vorgang erzeugt im selben Commit
> einen Signaturauftrag, ein Worker ist der einzige Sprecher für
> Signaturtransaktionen zur TSE, alle Signaturen liegen direkt am
> Signaturauftrag. Der heutige Nachsignier-Pfad
> (Auftragstabelle, Worker mit Healing und Backoff, Signatur-Seitentabelle)
> wird damit vom Ausnahme- zum Normalfall befördert; die Seitentabelle
> geht in der Auftragstabelle auf.
> Rechtliche Prüfung: siehe Further Notes; die Konformitätsbedingungen sind
> als Anforderungen eingearbeitet.

## Problem Statement

Die TSE ist heute tief in den Kern von jotti verwoben. Jedes signaturpflichtige
Event durchläuft im Kassier-Request eine synchrone Signierung mit Deadline,
bevor es persistiert wird. Das hat vier spürbare Folgen:

1. Kassieren wartet auf fiskaly. Jede Zahlung und jede Bestellung zahlt die
   Cloud-Roundtrips als Latenz, bei Vereins-Internet (Festzelt, Mobilfunk)
   spürbar; im Störungsfall bis zur vollen Deadline. Die Wartezeit trifft
   ehrenamtliche Servicekräfte im hektischsten Moment.
2. Zwei Signaturquellen. Erfolgreiche Signaturen liegen im Event-Payload,
   nachsignierte in der Seitentabelle. Jeder Leser (Kassenbeleg, DSFinV-K-
   Export) muss beide Quellen zusammenführen; der Ausfall ist als eingefrorenes
   Event-Flag dokumentiert, obwohl er später geheilt wird.
3. Der Kern kennt die TSE. Drei Kassen-Module tragen eigene Signierhelfer,
   das Journal-Repository eine kombinatorische Sammlung von Write-Methoden
   (Event mit/ohne Nachsignier-Auftrag, mit/ohne Druckaufträge, atomare
   Mehrfach-Writes), und die Event-Payloads führen Infrastrukturdaten mit.
   Command-Tests brauchen TSE-Verdrahtung, obwohl sie Kassenlogik testen.
4. Ein Korrektheitsloch. Stirbt der Prozess zwischen TSE-Signatur und
   Event-Write, existiert eine signierte TSE-Transaktion ohne
   Kassen-Gegenstück, eine erklärungsbedürftige Waise in der Verprobung
   TSE-Export gegen DSFinV-K.

Gleichzeitig existiert für den Ausfallpfad bereits die halbe Ziel-Architektur:
Auftragstabelle, Worker mit Ist-Abfrage-Healing und Backoff, Signatur-
Seitentabelle, Admin-Verwaltung. Sie wird aber nur im Sonderfall benutzt, der
Normalfall läuft daran vorbei.

## Solution

Die Signierung wird vollständig vom Kassier-Pfad entkoppelt:

1. Ein Kassen-Command validiert, baut das Event (ohne TSE-Felder) und schreibt
   es zusammen mit genau einem Signaturauftrag in einer Datenbank-Transaktion.
   Die Antwort an die Servicekraft kommt sofort nach dem Commit.
2. Ein Signatur-Worker ist der einzige Sprecher für Signaturtransaktionen
   zur TSE. Er wird nach jedem Commit sofort angestoßen (Polling nur als
   Fallback), arbeitet die Aufträge in Reihenfolge ab und legt das Ergebnis
   direkt am Auftrag ab.
   Healing (Ist-Abfrage vor erneutem Signieren), Backoff und die
   Admin-Aktionen des heutigen Nachsignier-Workers bleiben erhalten.
3. Der Beleg-Abruf antwortet sofort aus dem Signaturstatus, ohne Warten im
   Backend. Im Normalbetrieb liegt die Signatur beim Druck längst vor; liegt
   sie ausnahmsweise noch nicht vor, meldet das System den Beleg als
   ausstehend und die UI fasst selbsttätig nach. Ein Beleg ohne TSE-Daten
   entsteht nur bei dokumentiertem Ausfall, nie bei bloßer Queue-Latenz,
   denn der Ausfallvermerk ist rechtlich nur für echte Ausfälle gedeckt.
4. Ein TSE-Ausfall ist kein Event-Fakt mehr, sondern ein Störungszeitraum:
   Ein Störungsprotokoll dokumentiert Beginn, Ende und Grund jeder
   Verhinderung zeitnahen Signierens, ob durch TSE-weite Fehler, durch
   Rückstand über der Ausfall-Schwelle (deckt auch hängenden Worker und
   App-Fehler ab) oder durch fehlende TSE-Konfiguration. Belege tragen den
   Ausfallvermerk, und nach Rückkehr der TSE arbeitet der Worker den
   Rückstand ab. Nachsignierte Belege tragen beim Nachdruck einen Vermerk.

Für Servicekräfte heißt das: Buchen blockiert nie auf die TSE, auch nicht bei
Störungen. Für Admins: ein Ort für Signaturstatus, Ausfalldokumentation und
Rückstand. Für Entwickler: ein Signaturpfad statt zwei, ein Leseweg statt
Merge-Logik, ein TSE-freier Kassen-Kern.

## User Stories

1. Als Servicekraft möchte ich, dass Bestellungen und Zahlungen sofort gebucht
   sind, ohne auf die TSE zu warten, damit ich im Stoßbetrieb zügig kassieren
   kann.
2. Als Servicekraft möchte ich, dass eine TSE-Störung meinen Bestell- und
   Kassierfluss nicht unterbricht (keine Fehlermeldung, keine Blockade), damit
   der Betrieb auf dem Fest weiterläuft.
3. Als Servicekraft möchte ich einen Kassenbeleg mit vollständigen TSE-Daten
   drucken können, sobald ich ihn anfordere. Liegt die Signatur ausnahmsweise
   noch nicht vor, möchte ich eine klare Ausstehend-Meldung sehen, während
   die Kasse selbsttätig nachfasst, statt einer Blockade.
4. Als Servicekraft möchte ich bei einem echten TSE-Ausfall trotzdem einen
   Beleg ausgeben können, der den Ausfall klar ausweist, damit die
   Belegausgabepflicht erfüllt bleibt.
5. Als Servicekraft möchte ich beim Nachdruck eines nachträglich signierten
   Belegs einen Hinweis auf die Nachsignierung sehen, damit abweichende
   TSE-Zeiten auf dem Beleg erklärbar sind.
6. Als Vereins-Admin möchte ich alle Signaturaufträge mit Status, Versuchen
   und letztem Fehler einsehen, damit ich den Zustand der TSE-Anbindung
   jederzeit beurteilen kann.
7. Als Vereins-Admin möchte ich sehen, wie aktuell die Signatur-Queue ist
   (Anzahl offener Aufträge, Alter des ältesten), damit ich Störungen erkenne,
   bevor Belege betroffen sind.
8. Als Vereins-Admin möchte ich bei einem TSE-Ausfall eine automatisch
   geführte Ausfalldokumentation (Beginn, Ende, Grund) haben, damit die
   Nachweispflicht ohne Handarbeit erfüllt ist.
9. Als Vereins-Admin möchte ich fehlgeschlagene Signaturaufträge erneut
   einreihen oder begründet verwerfen können, damit ich nach längeren
   Ausfällen aufräumen kann.
10. Als Vereins-Admin möchte ich den Kassenabschluss nur dann durchführen,
    wenn alle Vorgänge signiert sind oder nur dokumentierte Ausfall-Reste
    verbleiben, und im Blockadefall eine klare Meldung sehen, was noch offen
    ist.
11. Als Vereins-Admin möchte ich, dass der Kassenabschluss bei einem
    mehrstündigen TSE-Ausfall trotzdem möglich ist, damit der Ausfall nicht
    den ganzen Veranstaltungstag blockiert.
12. Als Vereins-Admin möchte ich im Dashboard gewarnt werden, wenn die
    Signatur-Queue wächst oder Aufträge endgültig fehlschlagen, damit ich
    reagieren kann.
13. Als Vereins-Admin möchte ich, dass Vorgänge aus der Zeit ohne
    TSE-Konfiguration einen endgültigen, dokumentierten Status erhalten und
    nicht automatisch nachsigniert werden; will ich den Bestand nach der
    Einrichtung doch absichern, möchte ich ihn bewusst zur Nachsignierung
    zurücksetzen können.
14. Als Betriebsprüfer möchte ich im DSFinV-K-Export jede TSE-Transaktion
    einem Kassenvorgang zuordnen können und umgekehrt, damit die Verprobung
    ohne Waisen und Lücken aufgeht.
15. Als Betriebsprüfer möchte ich nicht signierte Vorgänge im Export mit
    einer Fehlererläuterung vorfinden, damit Ausfälle nachvollziehbar sind.
16. Als Betriebsprüfer möchte ich anhand der Auftrags- und Signaturdaten
    nachvollziehen können, dass die Absicherung im Regelbetrieb unmittelbar
    erfolgte (Auftragszeit vs. TSE-Zeit), damit die Konformität belegbar ist.
17. Als Entwickler möchte ich Kassen-Commands ohne TSE-Verdrahtung testen,
    damit Kassenlogik-Tests einfach und schnell bleiben.
18. Als Entwickler möchte ich genau einen Signaturpfad pflegen, damit
    Fehlerbehandlung, Idempotenz und Retry an einer Stelle leben.
19. Als Entwickler möchte ich Signaturen über genau einen Leseweg beziehen,
    damit Beleg und Export keine Merge-Logik über zwei Quellen brauchen.
20. Als Entwickler möchte ich den TSE-Anbieter hinter dem Worker austauschen
    können, ohne den Kassen-Kern anzufassen.
21. Als Entwickler möchte ich, dass jede TSE-Transaktion garantiert einem
    persistierten Event zugeordnet ist (erst committen, dann signieren),
    damit das Waisen-Szenario konstruktiv ausgeschlossen ist.

## Implementation Decisions

Architektur

- Die bestehende Nachsignier-Auftragstabelle wird zur allgemeinen
  Signaturauftrags-Outbox generalisiert: genau ein Auftrag je
  signaturpflichtigem Event, angelegt in derselben Datenbank-Transaktion wie
  das Event (zwangsläufige Auslösung, kein Ermessen, kein Batching). Ein
  Auftrag referenziert sein Event eindeutig und trägt die selbst erzeugte
  TSE-Transaktions-ID (eine zufällige UUID; der veraltete Schema-Kommentar,
  der sie deterministisch nennt, wird dabei korrigiert).
- Die Statusmaschine des Auftrags: offen, erledigt, fehlgeschlagen (nach
  Maximalversuchen), verworfen, TSE nicht konfiguriert (endgültig, siehe
  unten). Übergänge: Der Worker quittiert offen zu erledigt, zählt
  auftragsspezifische Fehlversuche bis fehlgeschlagen oder markiert offen
  als TSE nicht konfiguriert; der Admin setzt fehlgeschlagene und
  TSE-nicht-konfigurierte Aufträge auf offen zurück (einzeln oder gesamt)
  oder verwirft offene wie fehlgeschlagene Aufträge mit Begründung.
  Fehlversuche und Auftrags-Backoff gelten nur auftragsspezifischen
  Fehlern: Backoff im Sekundenbereich (etwa 5, 15, 45 Sekunden) und wenige
  Maximalversuche (etwa drei), denn solche Fehler sind fast immer
  deterministisch; langes Retrying verzögert nur die Dashboard-Warnung.
  TSE-weite Fehler zählen nicht auf den Auftrag, sondern schalten den
  Worker in einen Störungszustand (siehe Worker); die heutige Minuten-Kurve
  mit zehn Versuchen stammt aus dem Nachsignier-Sonderfall und entfällt.
  Die Auftragstabelle ist die Datenbasis der Latenz-Metrik und
  aufbewahrungspflichtig: Aufträge werden nie gelöscht, Verwerfen ist ein
  protokollierter Statuswechsel mit Grund, Benutzer und Zeitpunkt (GoBD).
- processType und processData werden beim Einreihen als Snapshot im Auftrag
  gespeichert (friert ein, was zu signieren war; der Worker bleibt frei von
  Event-Schema-Wissen).
- Eine zentrale fiskalische Projektion bildet Event auf Signaturpflicht,
  processType und processData ab, auch datenabhängig (die Sitzungseröffnung
  ist nur bei Anfangsbestand über null signaturpflichtig), und ersetzt die
  heute auf drei Module verteilten Signierhelfer. Der Kassen-Kern
  entscheidet damit weiterhin, was
  fiskalisch abzusichern ist (Domänenwissen), aber nicht mehr wie, wann oder
  womit signiert wird. Beim Einreihen plausibilisiert die Projektion die
  erzeugten processData (Schema, Vorzeichen, Summen) und protokolliert
  Verstöße, ohne den Kassiervorgang zu blockieren; Gift-Aufträge werden so
  zu Testfehlern statt Laufzeitfällen.
- Der Signatur-Worker ist der einzige Sprecher für Signaturtransaktionen
  (Start, Finish, Ist-Abfrage); TSE-Setup-Flow und TSE-Status-Abfrage der
  Admin-Einstellungen sprechen weiterhin eigenständig mit fiskaly,
  zwangsläufig auch ohne fertige Konfiguration. Arbeitsweise des Workers:
  Sofort-Trigger nach jedem Commit über eine In-Process-Benachrichtigung,
  der bestehende Polling-Tick bleibt nur als Fallback und fängt nach
  einem Absturz verlorene Trigger auf; Abarbeitung in Auftragsreihenfolge
  (FIFO als
  Soll-Eigenschaft: im Regelbetrieb bleibt das TSE-Log chronologisch);
  differenzierte Fehlerbehandlung nach expliziter Fehlertaxonomie: Ein
  auftragsspezifischer Fehler (etwa 400/422, von fiskaly abgelehnte
  processData) zählt einen Fehlversuch am Auftrag und wird übersprungen,
  ein Gift-Auftrag staut also nie die Queue und scheitert nach wenigen
  Versuchen endgültig; ein TSE-weiter Fehler (Verbindung, 5xx, 429 samt
  Retry-After, 401/403, TSS-Zustandsfehler) bricht den Durchlauf ab, zählt
  keine Auftrags-Fehlversuche und schaltet den Worker in einen
  Störungszustand mit eigenem Backoff (Sekunden bis zum Deckel von wenigen
  Minuten) und Half-Open-Wiedereinstieg: Nach Ablauf versucht der Worker
  einen einzelnen Probe-Auftrag, bei Erfolg endet der Störungszeitraum und
  die volle Aufarbeitung beginnt. So markiert ein mehrstündiger Ausfall
  keine Aufträge als fehlgeschlagen, die Erholung wird binnen Minuten
  erkannt, und fiskaly wird während der Störung nicht mit dem ganzen
  Rückstand bombardiert. Beide Backoffs kommen bewusst ohne Jitter aus:
  Ein einzelner serieller Worker hat nichts zu desynchronisieren, und die
  Backoff-Tests bleiben deterministisch. Jeder Durchlauf hat eine
  Deadline; das Ist-Abfrage-Healing vor erneutem Signieren wird vom
  Nachsignier-Worker übernommen; die Quittierung (Signaturdaten ablegen
  und Auftrag erledigen) ist ein einzelnes Update am Auftrag; der
  TSE-Client samt Auth-Token wird über
  Aufträge hinweg wiederverwendet. Der In-Process-Trigger setzt das
  Single-Prozess-Deployment von jotti voraus; ein Advisory Lock beim
  Worker-Start sichert diese Annahme gegen eine versehentlich doppelt
  gestartete Instanz ab; ein Scale-out (LISTEN/NOTIFY, FOR UPDATE SKIP
  LOCKED) ist bewusst nicht Teil dieses Umbaus.
- Ein explizites Störungsprotokoll ist die TSE-Ausfalldokumentation: je
  Störung ein Zeitraum mit Beginn, Ende und Grund, gespeist aus drei
  Quellen: Worker-Störungszustand nach TSE-weitem Fehler, Rückstands-Ausfall
  ab der Zwei-Minuten-Schwelle und der Dauerzustand ohne TSE-Konfiguration.
  Den Rückstands-Zeitraum öffnet und schließt ein Watchdog-Ticker neben dem
  Worker, der periodisch das Alter des ältesten offenen Auftrags prüft;
  die Dokumentation hängt damit weder am Worker (der hängen kann) noch am
  Zufall des Leser-Traffics, und die Lesepfade bleiben rein lesend.
  Ein Zeitraum endet mit der ersten erfolgreichen Signatur, dem
  Unterschreiten der Schwelle beziehungsweise der Einrichtung der
  Konfiguration. Alle Leser (Beleg-Ausfallvermerk, Kassenabschluss-Gate,
  Admin-Ansicht) prüfen dasselbe Kriterium, den aktiven oder zuzurechnenden
  Störungszeitraum, über dieselbe Signaturstatus-Funktion (siehe Beleg und
  Signaturstatus), statt Zeiträume zur Lesezeit aus Auftragszeilen zu
  rekonstruieren. Das Störungsprotokoll ist wie die Auftragstabelle
  aufbewahrungspflichtig.
- Signaturaufträge entstehen auch ohne TSE-Konfiguration (Erst-Setup,
  Testbetrieb), immer als offen; der Kassen-Kern bleibt frei von
  Konfigurationswissen. Der Worker markiert sie ohne vorhandene
  Konfiguration mit dem endgültigen Status TSE nicht konfiguriert: keine
  Fehlversuchszählung, keine Rückstands-Warnung, keine automatische
  Wiederaufnahme, auch nicht nach späterer Einrichtung. Diese Endgültigkeit
  hängt nicht am Worker-Timing: Der Übergang von nicht konfiguriert zu
  konfiguriert markiert in derselben Transaktion alle noch offenen Aufträge
  endgültig; sonst würde ein beim Einrichten noch unmarkierter Auftrag
  (etwa nach Absturz oder hängendem Worker) vom Worker automatisch
  nachsigniert. Ein bloßer Wechsel der Zugangsdaten bei durchgehend
  vorhandener Konfiguration markiert nichts; offener Ausfall-Rückstand wird
  dann regulär nachsigniert. Eine
  Nachsignierungspflicht existiert nicht, und die nachträgliche Signatur
  heilt den Zeitraum ohne TSE rechtlich nicht; sie wäre reine
  Bestandshärtung. Wer den Bestand doch absichern will, setzt die Aufträge
  bewusst zurück (einzeln oder gesamt); automatisch signiert wird nichts,
  das erspart nach langem Testbetrieb den Signatur-Burst ins TSE-Log. Der
  Worker unterscheidet dabei fehlende Konfiguration (markieren) von nicht
  lesbarer Konfiguration (Fehler, nichts tun). Belege tragen den Vermerk,
  dass keine TSE konfiguriert ist; solche Aufträge blockieren den
  Kassenabschluss nicht. Damit eine Kassensitzung stets vollständig mit oder
  ohne TSE läuft, sind Änderungen der TSE-Konfiguration nur ohne offene
  Kassensitzung möglich.
- Die Auftragstabelle wird der einzige Signatur-Store: Die Signaturdaten
  (Transaktionsnummer, Signaturzähler, Seriennummer, logTimes, Signatur,
  QR-Code-Daten) liegen als Spalten direkt am Auftrag, NULL bis zur
  Quittierung und danach genau einmal beschrieben. Die heutige
  Signatur-Seitentabelle entfällt: Auftrag und Signatur sind über die
  Event-Referenz und die Transaktions-ID ohnehin strikt 1:1; eine eigene
  Tabelle brächte nur einen zweiten Join und eine zweite Schreibstelle.
  Event-Payloads verlieren sämtliche TSE-Felder (Signaturdaten,
  Transaktions-ID, Ausfall-Flag); der Ausfall ist künftig ein Queue-Zustand
  zur Lesezeit, kein eingefrorener Event-Fakt. Beleg und DSFinV-K-Export
  lesen ausschließlich die Auftragstabelle. Keine Datenmigration:
  Pre-Release, Breaking Changes
  sind laut Repo-Regeln ausdrücklich erlaubt, alte Events werden nicht
  migriert; Schemaänderungen erfolgen direkt in der Initial-Migration, der
  Seed wird angepasst.
- Das Journal-Repository bietet einen Event-Write mit optionalem
  Signaturauftrag statt der heutigen Methoden-Kombinatorik; die atomaren
  Mehrfach-Writes (Storno-Aufteilung, Umbuchung, Sitzungseröffnung) nehmen
  je Event ihren Auftrag entgegen. TSE-Aufrufe innerhalb offener
  DB-Transaktionen entfallen vollständig.

Beleg und Signaturstatus

- Der Beleg-Abruf antwortet sofort und wartet nie (Entscheidung:
  Sofortantwort statt Warte-Modul; es gibt keinen Backend-Wartepunkt). Eine
  Signaturstatus-Funktion ohne Zeitverhalten kapselt die Logik als einzige
  Implementierung des Ausfallbegriffs (auch das Kassenabschluss-Gate
  urteilt über sie, siehe Kassenabschluss) und liefert
  genau eines von vier Ergebnissen: Signatur vorhanden (regulärer
  TSE-Abschnitt), Signatur vorhanden mit Nachsigniert-Kennzeichen, Ausfall
  mit belegbarem Grund (Ausfallvermerk), oder Signatur ausstehend (kein
  Beleg). Bei Ausstehend fasst die UI selbsttätig nach (etwa alle ein bis
  zwei Sekunden für rund zehn Sekunden, danach auf Anforderung); das trägt
  auch den Direktverkauf, wo die Servicekraft den Beleg unmittelbar nach
  der Buchung anfordern kann und die erste Antwort dann regelmäßig
  Ausstehend ist. Im
  Rückstau-Fall wartet der Auftrag noch hinter anderen, ohne dass Signatur
  oder Störung vorliegt.
- Ausfallbegriff (Entscheidung: ursachenunabhängig, Basis ist das
  Störungsprotokoll): Ein dokumentierter Ausfall liegt vor, wenn ein
  Störungszeitraum aktiv ist (TSE-weiter Fehler, Rückstands-Ausfall ab zwei
  Minuten Alter des ältesten offenen Auftrags, fehlende TSE-Konfiguration)
  oder wenn am konkreten Auftrag auftragsspezifische Fehlversuche
  protokolliert sind (der Gift-Fall: die TSE funktioniert, nur dieser
  Vorgang lässt sich nicht signieren). Der Rückstands-Ausfall deckt auch
  hängenden Worker und App-Fehler ab, bei denen nie ein Fehlversuch
  entsteht; die Schwelle liegt deutlich über normalen Lastspitzen, damit
  Stoßbetrieb die Ausfalldokumentation nicht verwässert.
- Ausfallvermerk-Politik (Entscheidung: nur bei dokumentiertem Ausfall):
  Ein Beleg ohne TSE-Daten entsteht nur bei dokumentiertem Ausfall im
  obigen Sinn oder während einer Aufholphase nach dokumentiertem Ausfall.
  Bloße Queue-Latenz unterhalb der Schwelle erzeugt nie einen
  Ausfallvermerk, sondern das Ergebnis Signatur ausstehend, denn der
  Vermerk ist rechtlich nur für echte Ausfälle gedeckt. Der Rest-Fall
  bleibt benannt: Eine langsame, aber fehlerfreie TSE liefert bis zur
  Rückstands-Schwelle nur Ausstehend-Antworten; dieses Fenster von maximal
  zwei Minuten ist die bewusst in Kauf genommene Kehrseite der
  Sofortantwort.
- Nachsigniert-Vermerk (Entscheidung: beibehalten, Kriterium rein
  zeitbasiert): Der Vermerk erscheint, wenn die Signatur später als rund
  eine Minute nach Auftragserstellung entstand. Er ist keine Rechtspflicht,
  aber er erklärt TSE-Zeitpunkte, die vom Belegdatum abweichen; ein schnell
  überwundener Fehlversuch erzeugt keine erklärungsbedürftige Abweichung
  und deshalb keinen Vermerk. Mit diesem Kriterium erscheint der Vermerk
  nur in echten Ausfall- und Aufholszenarien.

Kassenabschluss

- Kassenabschluss-Gate (Entscheidung: Sofortantwort, Ausfall-Reste
  zulässig): Das Gate schützt die gesamte Ein-Klick-Abschluss-Operation
  (Kassensturz, Differenzbuchung, Tagesabschluss) und prüft als deren
  erste Handlung, noch vor der wird-abgeschlossen-Barriere. Es prüft
  sofort und wartet nicht; sind noch Aufträge offen, meldet es sie mit
  Anzahl und Alter, und die UI kann erneut anfordern (dasselbe Muster wie
  der Beleg-Abruf). Das Gate klassifiziert offene Aufträge über dieselbe
  Signaturstatus-Funktion wie der Beleg-Abruf und blockiert genau dann,
  wenn mindestens ein Auftrag das Ergebnis Signatur ausstehend hat; die
  Zurechnung existiert nur einmal, Beleg und Gate können einander nicht
  widersprechen. Ausfall-Reste sind endgültig fehlgeschlagene und
  verworfene Aufträge (stets) sowie offene Aufträge, die einem
  dokumentierten, auch noch laufenden Störungszeitraum zuzurechnen sind
  oder auftragsspezifische Fehlversuche tragen; sie lassen den Abschluss
  zu, die Abschlussmeldung weist sie aus. Nur frische offene Aufträge ohne
  Ausfallbezug blockieren mit einer Meldung, die sie benennt. Aufträge im
  endgültigen Status TSE nicht konfiguriert blockieren nicht; schließt ein
  Tag vollständig ohne TSE, weist die Abschlussmeldung das deutlich aus.
  Die signaturpflichtigen Abschluss-Events (Differenzbuchung bei Differenz
  ungleich Null, Tagesabschluss) werden regulär über die Queue signiert.
- Reste nach dem Abschluss (Entscheidung: nachsignieren): Kehrt die TSE nach
  einem Abschluss mit Ausfall-Resten zurück, arbeitet der Worker offene
  Reste regulär nach; endgültig fehlgeschlagene kann der Admin zurücksetzen.
  Die Signatur landet am Auftrag, der Export zeigt sie
  vollständig, Nachsigniert-Vermerk und Ausfalldokumentation erklären den
  Zeitversatz gegenüber dem Kassenabschluss. Lieber eine späte Signatur als
  dauerhaft keine (AEAO Nr. 1.14.4: schnellstmögliche Wiederherstellung des
  konformen Zustands).

Admin und Monitoring

- Die Nachsignier-Verwaltung wird zur Signaturauftrags-Verwaltung
  (Statusliste, Zurücksetzen einzeln oder gesamt, Verwerfen mit Begründung)
  und um den Queue-Zustand ergänzt: Anzahl offener Aufträge, Alter des
  ältesten offenen Auftrags und Abarbeitungsrate (Signaturen pro Minute,
  Signierdauer p95), damit wachsender von schrumpfendem Rückstand
  unterscheidbar ist, gerade in der Aufholphase; die Raten werden on demand
  per SQL aus den gespeicherten Auftrags- und Signaturzeiten über ein
  gleitendes 15-Minuten-Fenster berechnet, ein eigenes Metrik-Subsystem
  gibt es nicht. Das Admin-Dashboard warnt
  ab rund einer Minute Rückstand oder bei endgültig fehlgeschlagenen
  Aufträgen; ab zwei Minuten Rückstand eröffnet automatisch ein
  Störungszeitraum. Ohne TSE-Konfiguration zeigt das Dashboard eine
  permanente, unübersehbare Warnung: kein Queue-Alarm, sondern ein
  Konfigurationsalarm, denn produktiver Betrieb ohne TSE ist nicht konform
  und darf nicht normal aussehen. Die Ausfalldokumentations-Ansicht zeigt
  die Störungszeiträume direkt aus dem Störungsprotokoll (ein Ausfall ist
  für den Prüfer ein Zeitraum mit Grund, nicht hunderte Einzelaufträge).

Konformitätsbedingungen (als Anforderungen verbindlich)

- Einreihen transaktional mit der Aufzeichnung (KassenSichV § 2 Satz 1:
  unmittelbare, zwangsläufige Auslösung je Aufzeichnung).
- Sofort-Trigger statt Polling als Normalpfad; Ziellatenz im Regelbetrieb im
  Sekundenbereich, p95 unter fünf Sekunden; die Latenz ist aus den
  gespeicherten Zeiten (Auftrag erstellt vs. TSE-logTime) jederzeit
  nachweisbar.
- Beleg ohne TSE-Daten nur bei dokumentiertem Ausfall im Sinne des
  ursachenunabhängigen Ausfallbegriffs (AEAO Nr. 1.14.2/1.14.3), nie bei
  bloßer Queue-Latenz unterhalb der Schwelle.
- Kassenabschluss nur bei leerer Queue oder ausschließlich dokumentierten
  Ausfall-Resten (AEAO Nr. 2.2.3.3).
- Ausfalldokumentation automatisch über das Störungsprotokoll (AEAO
  Nr. 1.14.1), ursachenunabhängig einschließlich Rückstands-Ausfall und
  fehlender TSE-Konfiguration; die Muster-Verfahrensdokumentation
  (verfahrensdokumentation.md, Abschnitt TSE-Anbindung) erläutert
  Mechanismus, typische Latenz und mögliche Verzögerungen und erfüllt damit
  zugleich die Herstellerdokumentations-Pflicht aus BSI TR-03153-1
  Kap. 3.9.3.
- Auftragstabelle und Störungsprotokoll sind Teil der
  aufbewahrungspflichtigen Unterlagen: kein Löschen, die Signaturspalten
  werden genau einmal beschrieben, Verwerfen nur als protokollierter
  Statuswechsel mit Grund, Benutzer und Zeitpunkt
  (GoBD-Nachvollziehbarkeit).

Benennung und Dokumentation

- Neue Begriffe (Signaturauftrag, Signatur-Worker, Signaturstatus,
  Störungsprotokoll, Störungszeitraum, Aufholphase, Rückstands-Ausfall,
  Signatur ausstehend, TSE nicht konfiguriert) werden in der Ubiquitous
  Language ergänzt; Endpunkte und UI-Texte folgen der bestehenden deutschen
  Fachsprache. Die Abschluss-Operation heißt durchgängig Kassenabschluss
  (der Ein-Klick-Ablauf), Tagesabschluss bezeichnet nur noch das Event und
  den Z-Bon; die veralteten Einträge zu Kassensturz und Z-Bon in der
  Ubiquitous Language (Zwei-Schritt-Modell) werden dabei auf die
  zusammengelegte Operation aktualisiert.
- Handbuch (TSE-Architektur), Compliance-Dokumentation (TSE-Integration)
  und Muster-Verfahrensdokumentation (TSE-Anbindung) werden auf das
  Outbox-Modell aktualisiert.

## Testing Decisions

Gute Tests prüfen ausschließlich Außenverhalten über die öffentliche
Schnittstelle des Moduls (Eingaben, Ergebnisse, persistierte Effekte), nie
Implementierungsdetails. Vorbilder im Repo: die Worker-Tests mit Fake-
TSE-Client, die tabellengetriebenen processData-Tests, die Command-Tests der
Kassen-Module und der Seed-Integrationstest.

Getestet werden alle Kernmodule:

- Fiskalische Projektion: tabellengetrieben je Event-Typ (signaturpflichtig
  ja/nein, processType, processData inklusive Vorzeichen-/Faktor-Fällen)
  samt datenabhängiger Signaturpflicht (Sitzungseröffnung mit und ohne
  Anfangsbestand).
- Signatur-Worker: Erfolgsfall, auftragsspezifischer Fehlversuch mit
  Sekunden-Backoff und schnellem endgültigem Fehlschlagen (Gift-Auftrag
  staut nie die Queue), TSE-weiter Fehler bricht den Durchlauf ab, zählt
  keine Auftrags-Fehlversuche und eröffnet einen Störungszeitraum,
  Half-Open-Probe beendet ihn bei Erfolg und startet die volle
  Aufarbeitung, Healing-Fälle (Transaktion bei der TSE bereits
  abgeschlossen, noch aktiv, unbekannt), FIFO-Reihenfolge im Regelbetrieb,
  Quittierung als einzelnes Update am Auftrag, Trigger-Verhalten,
  Crash-Recovery (Auftrag committet,
  Trigger verloren, der Polling-Fallback signiert nach).
- Signaturstatus-Funktion: Signatur vorhanden; dokumentierter Ausfall führt
  zum Ausfallergebnis mit Grund; Rückstau ohne Störung führt zum Ergebnis
  Signatur ausstehend; Überschreiten der Rückstands-Schwelle eröffnet den
  Störungszeitraum und kippt Ausstehend in Ausfall; Aufholphase; verspätete
  Signatur führt zum Nachsigniert-Kennzeichen; keine falschen
  Ausfallvermerke bei bloßer Latenz.
- Störungsprotokoll: TSE-weiter Fehler eröffnet einen Zeitraum, die erste
  erfolgreiche Signatur schließt ihn; der Watchdog öffnet und schließt den
  Rückstands-Zeitraum an der Schwelle, auch bei hängendem Worker; der
  Zeitraum ohne TSE-Konfiguration endet mit der Einrichtung.
- Kassenabschluss-Gate: Sofortantwort statt Warten; leere Queue, frischer
  offener Auftrag blockiert mit Meldung, Ausfall-Reste (auch offene
  Aufträge im laufenden Störungszeitraum) lassen den Abschluss zu, Aufträge
  im Status TSE nicht konfiguriert blockieren nicht; die Klassifikation
  läuft über die Signaturstatus-Funktion, nicht über einen zweiten
  Zurechnungspfad.
- TSE-nicht-konfiguriert-Fluss: Aufträge entstehen als offen und werden vom
  Worker endgültig markiert; der Übergang zu konfiguriert markiert auch
  noch unmarkierte offene Aufträge, ein reiner Zugangsdaten-Wechsel nicht;
  eine spätere Einrichtung fasst endgültig markierte Aufträge nicht an;
  das Admin-Zurücksetzen reiht sie bewusst wieder ein und der Worker
  signiert sie nach.
- Angepasste Beleg- und Export-Pfade: Kassenbeleg-Erzeugung mit den vier
  Ergebnisarten der Signaturstatus-Funktion; DSFinV-K-Mapper liest nur noch
  die Auftragstabelle und füllt das Fehlerfeld für unsignierte Vorgänge.
- Der bestehende Seed-Integrationstest wird auf das neue Schema und den
  Worker-Fluss umgestellt.

## Out of Scope

- Ein zweiter TSE-Anbieter oder eine Hardware-TSE. Der Umbau kapselt den
  Anbieter im Worker und bereitet den Austausch vor, implementiert ihn aber
  nicht.
- Ein Backend-Wartepunkt auf die Signatur, ob in der Buchungs-Response oder
  im Beleg-Abruf. Der Beleg-Abruf antwortet sofort aus dem Signaturstatus;
  nachgefasst wird in der UI.
- Eine Priorisierung belegkritischer Aufträge (Priority-Bump beim
  Beleg-Abruf): bewusst verworfen, weil die Zahlung im TSE-Log sonst vor
  ihren Bestellungen signiert würde und der Bump ohne Backend-Wartepunkt
  kaum Latenz gewinnt.
- Eine automatische Nachsignierung von Vorgängen aus der Zeit ohne
  TSE-Konfiguration; das bewusste Admin-Zurücksetzen ersetzt sie.
- Änderungen an der Belegausgabe-UX jenseits der Nachfass- und
  Vermerk-Logik sowie der TSE-Setup-Wizard (eigenes Vorhaben; er erhält
  keine automatische Altbestands-Nachsignierung mehr, siehe oben).
- DSFinV-K-Strukturänderungen über die Signaturquelle und das Fehlerfeld
  hinaus; ELSTER-Meldepflicht.
- Eine Datenmigration für bestehende Events oder Instanzen (Pre-Release,
  keine produktiven Nutzer).
- Ein Scale-out des Signatur-Workers über mehrere Prozesse (LISTEN/NOTIFY,
  FOR UPDATE SKIP LOCKED). Der Umbau setzt das heutige
  Single-Prozess-Deployment voraus und dokumentiert das als Annahme.

## Further Notes

Rechtliche Bewertung (geprüft am Wortlaut der lokalen Rechtsquellen, siehe
Index in der Rechtsquellen-Sammlung des Repos): Eine synchrone Signatur ist
nirgends vorgeschrieben; die Normen regeln Auslöser, Reihenfolge und
zeitliche Kopplung, nicht die Ausführungsarchitektur. Maßgeblich sind
KassenSichV § 2 Satz 1 (für jede Aufzeichnung unmittelbar eine neue
Transaktion), § 2 Satz 3 (TSE legt die maßgeblichen Zeitpunkte fest), § 6
(TSE-Pflichtangaben auf dem Beleg), AO § 146a Abs. 2 (Beleg in unmittelbarem
zeitlichem Zusammenhang), AEAO zu § 146a Nr. 2.2.2 (unmittelbar mit
Vorgangsbeginn starten; 45-Sekunden-Anker für Updates), Nr. 2.2.3.3 (vor
Belegausgabe und Kassenabschluss zwingend beenden), Nr. 2.5.7
(Belegausgabe unmittelbar nach Vorgangsende) und Nr. 1.14 (Ausfallregime).
BSI TR-03153-1 Kap. 3.9.3 verpflichtet den Hersteller (MUSS), in einer
Herstellerdokumentation über Durchführungszeiten und mögliche Verzögerungen
der Absicherung aufzuklären; der Wortlaut zielt auf gleichzeitig zu
bearbeitende Absicherungen, was den Queue-Rückstand des Outbox-Modells
einschließt. jotti ist hier Hersteller, die Muster-Verfahrensdokumentation
erfüllt diese Pflicht. Die daraus
abgeleiteten Konformitätsbedingungen sind oben als Anforderungen
festgeschrieben. Das verbleibende Auslegungsrisiko liegt im nicht
quantifizierten Begriff unmittelbar (KassenSichV § 2 Satz 1, AEAO
Nr. 2.2.2): Sekundenlatenz ist mit dem 45-Sekunden-Wertungsmaßstab und der
TR-Anerkennung von Verzögerungen gut vertretbar, und die jederzeit
nachweisbare Latenz (Auftragszeit vs. TSE-logTime) ist die
Verteidigungslinie. Eine Nachsignierungspflicht existiert nicht; die
Nachsignierung bleibt als freiwillige Härtung erhalten.

Das Outbox-Modell ist der Rechtslage in zwei Punkten näher als der heutige
Sync-Pfad: Die Vollständigkeit der Absicherung ist transaktional garantiert
(kein vergessener Auftrag), und das Waisen-Szenario (signiert, aber nicht
aufgezeichnet) ist konstruktiv ausgeschlossen, weil erst committet und dann
signiert wird.

Betriebshinweis: Die Queue-Latenz ist die zentrale Betriebsgröße des neuen
Modells. Sie fällt als Differenz zwischen Auftragserstellung und TSE-logTime
ohnehin an und soll im Admin-Dashboard sichtbar sein; die
Muster-Verfahrensdokumentation dokumentiert den Mechanismus und die
erwartete Größenordnung. Durchsatzannahme: Eine Signatur besteht aus zwei
sequenziellen fiskaly-Roundtrips (Start und Finish); bei 300 bis 500
Millisekunden je Roundtrip schafft der serielle Worker grob ein bis
anderthalb Signaturen pro Sekunde; der Spitzenbedarf eines Vereinsfests
liegt weit darunter. Würde die Annahme je verletzt, ist Pipelining der
erste Schritt (Start des Folgeauftrags parallel zum Finish des laufenden,
die Chronologie der Finish-Aufrufe bleibt erhalten), paralleles Signieren
unter Aufgabe der FIFO-Chronologie erst der zweite.
