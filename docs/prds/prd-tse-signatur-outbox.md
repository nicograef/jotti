# PRD: TSE-Signatur über Outbox und Signatur-Worker

> Umbau der TSE-Integration vom synchronen Signieren im Kassier-Pfad auf ein
> Outbox-Modell: Jeder signaturpflichtige Vorgang erzeugt im selben Commit
> einen Signaturauftrag, ein Worker ist der einzige Sprecher zur TSE, alle
> Signaturen liegen in einer Seitentabelle. Der heutige Nachsignier-Pfad
> (Auftragstabelle, Worker mit Healing und Backoff, Signatur-Seitentabelle)
> wird damit vom Ausnahme- zum Normalfall befördert.
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
2. Ein Signatur-Worker ist der einzige Sprecher zur TSE. Er wird nach jedem
   Commit sofort angestoßen (Polling nur als Fallback), arbeitet die Aufträge
   in Reihenfolge ab und legt das Ergebnis in der Signatur-Seitentabelle ab.
   Healing (Ist-Abfrage vor erneutem Signieren), Backoff und die
   Admin-Aktionen des heutigen Nachsignier-Workers bleiben erhalten.
3. Der Kassenbeleg ist der einzige Ort, der auf die Signatur wartet: Beim
   Beleg-Abruf wartet das System begrenzt (rund zehn Sekunden) auf die
   Signatur oder den Ausgang des laufenden Signierversuchs. Im Normalbetrieb
   liegt die Signatur beim Druck längst vor; ein Beleg ohne TSE-Daten
   entsteht nur bei dokumentiertem Ausfall. Bei bloßem Rückstau ohne Ausfall
   meldet das System den Beleg als ausstehend, statt einen rechtlich
   ungedeckten Ausfallvermerk zu drucken.
4. Ein TSE-Ausfall ist kein Event-Fakt mehr, sondern ein Queue-Zustand:
   Ausfall ist jede dokumentierte Verhinderung zeitnahen Signierens, ob
   durch Fehlversuche gegen die TSE oder durch Rückstand über der
   Ausfall-Schwelle (deckt auch hängenden Worker und App-Fehler ab). Die
   Auftragstabelle dokumentiert Beginn, Ende und Grund, Belege tragen den
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
   drucken können, sobald ich ihn anfordere, und dafür höchstens eine kurze,
   begrenzte Wartezeit in Kauf nehmen. Dauert es ausnahmsweise länger,
   möchte ich eine klare Ausstehend-Meldung statt einer Blockade.
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
10. Als Vereins-Admin möchte ich den Tagesabschluss nur dann durchführen,
    wenn alle Vorgänge signiert sind oder nur dokumentierte Ausfall-Reste
    verbleiben, und im Blockadefall eine klare Meldung sehen, was noch offen
    ist.
11. Als Vereins-Admin möchte ich, dass der Tagesabschluss bei einem
    mehrstündigen TSE-Ausfall trotzdem möglich ist, damit der Ausfall nicht
    den ganzen Veranstaltungstag blockiert.
12. Als Vereins-Admin möchte ich im Dashboard gewarnt werden, wenn die
    Signatur-Queue wächst oder Aufträge endgültig fehlschlagen, damit ich
    reagieren kann.
13. Als Vereins-Admin möchte ich, dass Vorgänge aus der Zeit vor der
    TSE-Einrichtung nach dem Einrichten automatisch nachsigniert werden,
    damit der Bestand ohne Handarbeit vollständig abgesichert ist.
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
  TSE-Transaktions-ID.
- Die Statusmaschine des Auftrags bleibt: offen, erledigt, fehlgeschlagen
  (nach Maximalversuchen), verworfen. Übergänge: Der Worker quittiert offen
  zu erledigt oder zählt Fehlversuche bis fehlgeschlagen; der Admin setzt
  fehlgeschlagen auf offen zurück oder verwirft offene wie fehlgeschlagene
  Aufträge mit Begründung. Der Backoff startet im Sekundenbereich (etwa 5,
  15, 45 Sekunden) und wächst dann exponentiell in Minuten bis zum Deckel
  von 30 Minuten; die heutige Minuten-Kurve stammt aus dem
  Nachsignier-Sonderfall und wäre im Normalpfad zu träge. Die
  Auftragstabelle ist zugleich die TSE-Ausfalldokumentation (Beginn, Ende,
  Grund je Auftrag) und die Datenbasis der Latenz-Metrik; sie ist
  aufbewahrungspflichtig: Aufträge werden nie gelöscht, Verwerfen ist ein
  protokollierter Statuswechsel mit Grund, Benutzer und Zeitpunkt (GoBD).
- processType und processData werden beim Einreihen als Snapshot im Auftrag
  gespeichert (friert ein, was zu signieren war; der Worker bleibt frei von
  Event-Schema-Wissen).
- Eine zentrale fiskalische Projektion bildet Event auf Signaturpflicht,
  processType und processData ab und ersetzt die heute auf drei Module
  verteilten Signierhelfer. Der Kassen-Kern entscheidet damit weiterhin, was
  fiskalisch abzusichern ist (Domänenwissen), aber nicht mehr wie, wann oder
  womit signiert wird.
- Der Signatur-Worker ist der einzige Sprecher zur TSE: Sofort-Trigger nach
  jedem Commit über eine In-Process-Benachrichtigung, der bestehende
  Polling-Tick bleibt nur als Fallback und fängt nach einem Absturz
  verlorene Trigger auf; Abarbeitung in Auftragsreihenfolge (FIFO als
  Soll-Eigenschaft: im Regelbetrieb bleibt das TSE-Log chronologisch);
  differenzierte Fehlerbehandlung nach Fehlerklasse: Ein auftragsspezifischer
  Fehler (etwa von fiskaly abgelehnte processData) wird übersprungen und
  kommt per Backoff wieder dran, ein Gift-Auftrag staut also nie die Queue;
  ein TSE-weiter Fehler (Verbindung, 5xx) bricht den Durchlauf ab, statt
  Fehlversuche über den ganzen Rückstand zu kaskadieren; Ist-Abfrage-Healing
  vor erneutem Signieren und atomare Quittierung (Signatur ablegen + Auftrag
  erledigen) werden vom Nachsignier-Worker übernommen; der TSE-Client samt
  Auth-Token wird über Aufträge hinweg wiederverwendet. Der In-Process-
  Trigger setzt das Single-Prozess-Deployment von jotti voraus; ein
  Scale-out (LISTEN/NOTIFY, FOR UPDATE SKIP LOCKED) ist bewusst nicht Teil
  dieses Umbaus.
- Signaturaufträge entstehen auch ohne TSE-Konfiguration (Erst-Setup,
  Testbetrieb): Der Queue-Zustand heißt dann „wartet auf TSE-Konfiguration",
  ein dokumentierter Dauerzustand ohne Fehlversuchszählung und ohne
  Rückstands-Warnung. Belege tragen bis dahin den Vermerk, dass keine TSE
  konfiguriert ist; solche Aufträge blockieren den Tagesabschluss nicht.
  Wird die TSE später eingerichtet, arbeitet der Worker den gesamten
  Bestand automatisch nach; das TSE-Setup-Wizard-Vorhaben erhält die
  Nachsignierung des Altbestands damit ohne Zusatzaufwand.
- Die Signatur-Seitentabelle wird der einzige Signatur-Store. Event-Payloads
  verlieren sämtliche TSE-Felder (Signaturdaten, Transaktions-ID,
  Ausfall-Flag); der Ausfall ist künftig ein Queue-Zustand zur Lesezeit, kein
  eingefrorener Event-Fakt. Beleg und DSFinV-K-Export lesen ausschließlich
  die Seitentabelle. Keine Datenmigration: Pre-Release, Breaking Changes
  sind laut Repo-Regeln ausdrücklich erlaubt, alte Events werden nicht
  migriert; Schemaänderungen erfolgen direkt in der Initial-Migration, der
  Seed wird angepasst.
- Das Journal-Repository bietet einen Event-Write mit optionalem
  Signaturauftrag statt der heutigen Methoden-Kombinatorik; die atomaren
  Mehrfach-Writes (Storno-Aufteilung, Umbuchung, Sitzungseröffnung) nehmen
  je Event ihren Auftrag entgegen. TSE-Aufrufe innerhalb offener
  DB-Transaktionen entfallen vollständig.

Beleg und Wartepunkt

- Einziger Wartepunkt ist der Kassenbeleg-Abruf (Entscheidung: Beleg wartet,
  nicht die Buchungs-Response). Ein Warte-Modul kapselt die Logik und liefert
  genau eines von vier Ergebnissen: Signatur vorhanden (regulärer
  TSE-Abschnitt), Signatur vorhanden mit Nachsigniert-Kennzeichen, Ausfall
  mit belegbarem Grund (Ausfallvermerk), oder Signatur ausstehend (kein
  Beleg, die UI fordert erneut an). Die Gesamtwartezeit des Beleg-Abrufs ist
  auf rund zehn Sekunden begrenzt; damit ist auch der Rückstau-Fall
  definiert, in dem der Auftrag noch hinter anderen wartet und weder
  Signatur noch Fehlversuch vorliegt.
- Ausfallbegriff (Entscheidung: ursachenunabhängig mit Rückstands-Schwelle):
  Ein dokumentierter Ausfall liegt vor, wenn Fehlversuche am Auftrag
  protokolliert sind oder der älteste offene Auftrag älter als zwei Minuten
  ist (Rückstands-Ausfall; deckt auch hängenden Worker und App-Fehler ab,
  bei denen nie ein Fehlversuch entsteht). Der Dauerzustand „wartet auf
  TSE-Konfiguration" zählt ebenfalls als dokumentierter Zustand. Die
  Schwelle liegt deutlich über normalen Lastspitzen, damit Stoßbetrieb die
  Ausfalldokumentation nicht verwässert.
- Ausfallvermerk-Politik (Entscheidung: nur bei dokumentiertem Ausfall):
  Der Beleg-Abruf wartet auf den Ausgang des laufenden Signierversuchs,
  höchstens bis zu seiner Gesamtdeadline. Ein Beleg ohne TSE-Daten entsteht
  nur bei dokumentiertem Ausfall im obigen Sinn oder während einer
  Aufholphase nach dokumentiertem Ausfall. Bloße Queue-Latenz unterhalb der
  Schwelle erzeugt nie einen Ausfallvermerk, sondern das Ergebnis Signatur
  ausstehend, denn der Vermerk ist rechtlich nur für echte Ausfälle
  gedeckt.
- Nachsigniert-Vermerk (Entscheidung: beibehalten, Kriterium Fehlversuch
  oder verspätet): Der Vermerk erscheint, wenn am Auftrag mindestens ein
  Fehlversuch protokolliert wurde oder die Signatur später als rund eine
  Minute nach Auftragserstellung entstand. Er ist keine Rechtspflicht, aber
  er erklärt TSE-Zeitpunkte, die vom Belegdatum abweichen, und erscheint mit
  diesem Kriterium nur in echten Ausfall- und Aufholszenarien.

Tagesabschluss

- Tagesabschluss-Gate (Entscheidung: warten, Ausfall-Reste zulässig): Der
  Abschluss wartet kurz auf das Leerlaufen der Queue. Ausfall-Reste sind
  endgültig fehlgeschlagene, verworfene und offene Aufträge, die einem
  dokumentierten, auch noch laufenden Ausfall zuzurechnen sind (mindestens
  ein Fehlversuch oder aktiver Rückstands-Ausfall); sie lassen den Abschluss
  zu, die Abschlussmeldung weist sie aus. Nur frische offene Aufträge ohne
  Ausfallbezug blockieren mit einer Meldung, die sie benennt. Aufträge im
  Zustand „wartet auf TSE-Konfiguration" blockieren nicht. Das
  Abschluss-Event selbst wird regulär über die Queue signiert.
- Reste nach dem Abschluss (Entscheidung: nachsignieren): Kehrt die TSE nach
  einem Abschluss mit Ausfall-Resten zurück, arbeitet der Worker offene
  Reste regulär nach; endgültig fehlgeschlagene kann der Admin zurücksetzen.
  Die Signatur landet in der Seitentabelle, der Export zeigt sie
  vollständig, Nachsigniert-Vermerk und Ausfalldokumentation erklären den
  Zeitversatz gegenüber dem Kassenabschluss. Lieber eine späte Signatur als
  dauerhaft keine (AEAO Nr. 1.14.4: schnellstmögliche Wiederherstellung des
  konformen Zustands).

Admin und Monitoring

- Die Nachsignier-Verwaltung wird zur Signaturauftrags-Verwaltung
  (Statusliste, Zurücksetzen, Verwerfen mit Begründung) und um den
  Queue-Zustand ergänzt: Anzahl offener Aufträge und Alter des ältesten
  offenen Auftrags als Latenz-Metrik. Das Admin-Dashboard warnt ab rund
  einer Minute Rückstand oder bei endgültig fehlgeschlagenen Aufträgen; ab
  zwei Minuten Rückstand beginnt automatisch der dokumentierte
  Rückstands-Ausfall. Die Ausfalldokumentations-Ansicht bleibt, speist sich
  unverändert aus der Auftragstabelle und fasst zusammenhängende Aufträge zu
  Ausfall-Zeiträumen zusammen (ein Ausfall ist für den Prüfer ein Zeitraum
  mit Grund, nicht hunderte Einzelaufträge).

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
- Tagesabschluss nur bei leerer Queue oder ausschließlich dokumentierten
  Ausfall-Resten (AEAO Nr. 2.2.3.3).
- Ausfalldokumentation automatisch über die Auftragstabelle (AEAO
  Nr. 1.14.1), ursachenunabhängig einschließlich Rückstands-Ausfall; die
  Verfahrensbeschreibung in der Compliance-Dokumentation erläutert
  Mechanismus, typische Latenz und mögliche Verzögerungen und erfüllt damit
  zugleich die Herstellerdokumentations-Pflicht aus BSI TR-03153-1
  Kap. 3.9.3.
- Die Auftragstabelle ist Teil der aufbewahrungspflichtigen Unterlagen:
  kein Löschen, Verwerfen nur als protokollierter Statuswechsel mit Grund,
  Benutzer und Zeitpunkt (GoBD-Nachvollziehbarkeit).

Benennung und Dokumentation

- Neue Begriffe (Signaturauftrag, Signatur-Worker, Aufholphase,
  Rückstands-Ausfall, Signatur ausstehend) werden in der Ubiquitous Language
  ergänzt; Endpunkte und UI-Texte folgen der bestehenden deutschen
  Fachsprache.
- Handbuch (TSE-Architektur) und Compliance-Dokumentation (TSE-Integration,
  Verfahrensbeschreibung) werden auf das Outbox-Modell aktualisiert.

## Testing Decisions

Gute Tests prüfen ausschließlich Außenverhalten über die öffentliche
Schnittstelle des Moduls (Eingaben, Ergebnisse, persistierte Effekte), nie
Implementierungsdetails. Vorbilder im Repo: die Worker-Tests mit Fake-
TSE-Client, die tabellengetriebenen processData-Tests, die Command-Tests der
Kassen-Module und der Seed-Integrationstest.

Getestet werden alle Kernmodule:

- Fiskalische Projektion: tabellengetrieben je Event-Typ (signaturpflichtig
  ja/nein, processType, processData inklusive Vorzeichen-/Faktor-Fällen).
- Signatur-Worker: Erfolgsfall, Fehlversuch mit Sekunden-Backoff,
  differenzierte Fehlerbehandlung (auftragsspezifischer Fehler wird
  übersprungen, TSE-weiter Fehler bricht den Durchlauf ab), Healing-Fälle
  (Transaktion bei der TSE bereits abgeschlossen, noch aktiv, unbekannt),
  FIFO-Reihenfolge im Regelbetrieb, Quittierung atomar, Trigger-Verhalten,
  Crash-Recovery (Auftrag committet, Trigger verloren, der Polling-Fallback
  signiert nach).
- Signatur-Warte-Modul: Signatur rechtzeitig; dokumentierter Ausfall führt
  zum Ausfallergebnis mit Grund; Rückstau ohne Ausfall führt zum Ergebnis
  Signatur ausstehend; Überschreiten der Rückstands-Schwelle kippt
  Ausstehend in Ausfall; Aufholphase; verspätete Signatur führt zum
  Nachsigniert-Kennzeichen; keine falschen Ausfallvermerke bei bloßer
  Latenz.
- Tagesabschluss-Gate: leere Queue, frischer offener Auftrag blockiert mit
  Meldung, Ausfall-Reste (auch offene Aufträge im laufenden Ausfall) lassen
  den Abschluss zu, Aufträge im Zustand „wartet auf TSE-Konfiguration"
  blockieren nicht.
- TSE-nicht-konfiguriert-Fluss: Aufträge entstehen ohne Konfiguration, eine
  spätere Einrichtung arbeitet den Bestand vollständig nach.
- Angepasste Beleg- und Export-Pfade: Kassenbeleg-Erzeugung mit den vier
  Ergebnisarten des Warte-Moduls; DSFinV-K-Mapper liest nur noch die
  Seitentabelle und füllt das Fehlerfeld für unsignierte Vorgänge.
- Der bestehende Seed-Integrationstest wird auf das neue Schema und den
  Worker-Fluss umgestellt.

## Out of Scope

- Ein zweiter TSE-Anbieter oder eine Hardware-TSE. Der Umbau kapselt den
  Anbieter im Worker und bereitet den Austausch vor, implementiert ihn aber
  nicht.
- Ein konfigurierbarer Wartepunkt (Buchungs-Response wartet). Es gibt genau
  einen Wartepunkt: den Beleg-Abruf.
- Änderungen an der Belegausgabe-UX jenseits der Warte- und Vermerk-Logik
  sowie der TSE-Setup-Wizard (eigenes Vorhaben).
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
BSI TR-03153-1 Kap. 3.9.3 erkennt Durchführungszeiten und Verzögerungen der
Signaturerstellung ausdrücklich an und verpflichtet den Hersteller (MUSS),
in einer Herstellerdokumentation darüber zu informieren; jotti ist hier
Hersteller, die Verfahrensbeschreibung erfüllt diese Pflicht. Die daraus
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
Verfahrensbeschreibung dokumentiert den Mechanismus und die erwartete
Größenordnung. Durchsatzannahme: Der serielle Worker schafft bei
fiskaly-Latenzen von 300 bis 500 Millisekunden grob zwei bis drei Signaturen
pro Sekunde; der Spitzenbedarf eines Vereinsfests liegt weit darunter. Würde
die Annahme je verletzt, wäre paralleles Signieren unter Aufgabe der
FIFO-Chronologie der nächste Schritt.
