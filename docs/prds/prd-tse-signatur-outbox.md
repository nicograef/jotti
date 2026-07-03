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
   Beleg-Abruf wartet das System auf die Signatur oder den Ausgang des
   laufenden Signierversuchs. Im Normalbetrieb liegt die Signatur beim Druck
   längst vor; ein Beleg ohne TSE-Daten entsteht nur bei nachweisbarem
   Fehlversuch, also einem echten, automatisch dokumentierten TSE-Ausfall.
4. Ein TSE-Ausfall ist kein Event-Fakt mehr, sondern ein Queue-Zustand:
   Solange Aufträge fehlschlagen, dokumentiert die Auftragstabelle den Ausfall
   (Beginn, Ende, Grund), Belege tragen den Ausfallvermerk, und nach
   Rückkehr der TSE arbeitet der Worker den Rückstand ab. Nachsignierte
   Belege tragen beim Nachdruck einen Vermerk.

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
   begrenzte Wartezeit in Kauf nehmen.
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
13. Als Betriebsprüfer möchte ich im DSFinV-K-Export jede TSE-Transaktion
    einem Kassenvorgang zuordnen können und umgekehrt, damit die Verprobung
    ohne Waisen und Lücken aufgeht.
14. Als Betriebsprüfer möchte ich nicht signierte Vorgänge im Export mit
    einer Fehlererläuterung vorfinden, damit Ausfälle nachvollziehbar sind.
15. Als Betriebsprüfer möchte ich anhand der Auftrags- und Signaturdaten
    nachvollziehen können, dass die Absicherung im Regelbetrieb unmittelbar
    erfolgte (Auftragszeit vs. TSE-Zeit), damit die Konformität belegbar ist.
16. Als Entwickler möchte ich Kassen-Commands ohne TSE-Verdrahtung testen,
    damit Kassenlogik-Tests einfach und schnell bleiben.
17. Als Entwickler möchte ich genau einen Signaturpfad pflegen, damit
    Fehlerbehandlung, Idempotenz und Retry an einer Stelle leben.
18. Als Entwickler möchte ich Signaturen über genau einen Leseweg beziehen,
    damit Beleg und Export keine Merge-Logik über zwei Quellen brauchen.
19. Als Entwickler möchte ich den TSE-Anbieter hinter dem Worker austauschen
    können, ohne den Kassen-Kern anzufassen.
20. Als Entwickler möchte ich, dass jede TSE-Transaktion garantiert einem
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
  (nach Maximalversuchen mit exponentiellem Backoff), verworfen. Die
  Auftragstabelle ist zugleich die TSE-Ausfalldokumentation (Beginn, Ende,
  Grund je Auftrag) und die Datenbasis der Latenz-Metrik.
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
  Polling-Tick bleibt nur als Fallback; Abarbeitung in Auftragsreihenfolge
  (FIFO, hält das TSE-Log chronologisch); Ist-Abfrage-Healing vor erneutem
  Signieren und atomare Quittierung (Signatur ablegen + Auftrag erledigen)
  werden vom Nachsignier-Worker übernommen; der TSE-Client samt Auth-Token
  wird über Aufträge hinweg wiederverwendet.
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
  genau eines von drei Ergebnissen: Signatur vorhanden (regulärer
  TSE-Abschnitt), Signatur vorhanden mit Nachsigniert-Kennzeichen, oder
  Ausfall mit belegbarem Grund (Ausfallvermerk).
- Ausfallvermerk-Politik (Entscheidung: nur nach Fehlversuch): Der
  Beleg-Abruf wartet auf den Ausgang des laufenden Signierversuchs, begrenzt
  durch dessen Deadline. Ein Beleg ohne TSE-Daten entsteht nur, wenn ein
  Fehlversuch am Auftrag protokolliert ist oder eine Aufholphase nach
  dokumentiertem Ausfall läuft. Bloße Queue-Latenz erzeugt nie einen
  Ausfallvermerk, denn der Vermerk ist rechtlich nur für echte Ausfälle
  gedeckt.
- Nachsigniert-Vermerk (Entscheidung: beibehalten, Kriterium Fehlversuch
  oder verspätet): Der Vermerk erscheint, wenn am Auftrag mindestens ein
  Fehlversuch protokolliert wurde oder die Signatur später als rund eine
  Minute nach Auftragserstellung entstand. Er ist keine Rechtspflicht, aber
  er erklärt TSE-Zeitpunkte, die vom Belegdatum abweichen, und erscheint mit
  diesem Kriterium nur in echten Ausfall- und Aufholszenarien.

Tagesabschluss

- Tagesabschluss-Gate (Entscheidung: warten, Ausfall-Reste zulässig): Der
  Abschluss wartet kurz auf das Leerlaufen der Queue. Verbleiben nur
  fehlgeschlagene oder verworfene Aufträge eines dokumentierten Ausfalls,
  fährt er fort; verbleiben offene, noch signierbare Aufträge, bricht er mit
  einer Meldung ab, die die offenen Aufträge benennt. Das Abschluss-Event
  selbst wird regulär über die Queue signiert.

Admin und Monitoring

- Die Nachsignier-Verwaltung wird zur Signaturauftrags-Verwaltung
  (Statusliste, Zurücksetzen, Verwerfen) und um den Queue-Zustand ergänzt:
  Anzahl offener Aufträge und Alter des ältesten offenen Auftrags als
  Latenz-Metrik, mit Warnhinweis im Admin-Dashboard bei Rückstand oder
  endgültig fehlgeschlagenen Aufträgen. Die Ausfalldokumentations-Ansicht
  bleibt und speist sich unverändert aus der Auftragstabelle.

Konformitätsbedingungen (als Anforderungen verbindlich)

- Einreihen transaktional mit der Aufzeichnung (KassenSichV § 2 Satz 1:
  unmittelbare, zwangsläufige Auslösung je Aufzeichnung).
- Sofort-Trigger statt Polling als Normalpfad; Ziellatenz im Regelbetrieb im
  Sekundenbereich; die Latenz ist aus den gespeicherten Zeiten (Auftrag
  erstellt vs. TSE-logTime) jederzeit nachweisbar.
- Beleg ohne TSE-Daten nur bei nachweisbarem Fehlversuch bzw. dokumentiertem
  Ausfall (AEAO Nr. 1.14.2/1.14.3), nie bei bloßer Queue-Latenz.
- Tagesabschluss nur bei leerer Queue oder ausschließlich dokumentierten
  Ausfall-Resten (AEAO Nr. 2.2.3.3).
- Ausfalldokumentation automatisch über die Auftragstabelle (AEAO
  Nr. 1.14.1); die Verfahrensbeschreibung in der Compliance-Dokumentation
  erläutert Mechanismus und typische Latenz.

Benennung und Dokumentation

- Neue Begriffe (Signaturauftrag, Signatur-Worker, Aufholphase) werden in der
  Ubiquitous Language ergänzt; Endpunkte und UI-Texte folgen der bestehenden
  deutschen Fachsprache.
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
- Signatur-Worker: Erfolgsfall, Fehlversuch mit Backoff, Healing-Fälle
  (Transaktion bei der TSE bereits abgeschlossen, noch aktiv, unbekannt),
  FIFO-Reihenfolge, Quittierung atomar, Trigger-Verhalten.
- Signatur-Warte-Modul: Signatur rechtzeitig; Fehlversuch führt zum
  Ausfallergebnis mit Grund; Aufholphase; verspätete Signatur führt zum
  Nachsigniert-Kennzeichen; keine falschen Ausfallvermerke bei bloßer
  Latenz.
- Tagesabschluss-Gate: leere Queue, offener Auftrag blockiert mit Meldung,
  Ausfall-Reste lassen den Abschluss zu.
- Angepasste Beleg- und Export-Pfade: Kassenbeleg-Erzeugung mit den drei
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
Signaturerstellung ausdrücklich an. Die fünf daraus abgeleiteten
Konformitätsbedingungen sind oben als Anforderungen festgeschrieben. Eine
Nachsignierungspflicht existiert nicht; die Nachsignierung bleibt als
freiwillige Härtung erhalten.

Das Outbox-Modell ist der Rechtslage in zwei Punkten näher als der heutige
Sync-Pfad: Die Vollständigkeit der Absicherung ist transaktional garantiert
(kein vergessener Auftrag), und das Waisen-Szenario (signiert, aber nicht
aufgezeichnet) ist konstruktiv ausgeschlossen, weil erst committet und dann
signiert wird.

Betriebshinweis: Die Queue-Latenz ist die zentrale Betriebsgröße des neuen
Modells. Sie fällt als Differenz zwischen Auftragserstellung und TSE-logTime
ohnehin an und soll im Admin-Dashboard sichtbar sein; die
Verfahrensbeschreibung dokumentiert den Mechanismus und die erwartete
Größenordnung.
