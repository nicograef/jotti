# PRD: Demo-Daten-Seeder

> Ersetzt die handgepflegte `database/seed.sql` (3-Tage-Sommerfest, ~993 Events) durch einen
> deterministischen Go-Seeder mit vollständiger Feature-Abdeckung.
> Entscheidungen geklärt am 2026-06-12 (Seed-Format, TSE-Abbildung, Bondruck, Szenario-Umfang,
> Modulschnitt, Tests, Guard-Verhalten).

## Problem Statement

Die Demo-Daten (eingespielt über `make seed`, genutzt für die lokale Entwicklung und die
öffentliche Demo auf jotti.rocks) stammen aus einer frühen Produktphase und decken die seither
gebauten Features nicht ab. Wer jotti heute über die Demo kennenlernt — ein Vereins-Admin, der
das System evaluiert, oder ein Interessent auf jotti.rocks — sieht ein unvollständiges und
teilweise widersprüchliches Bild:

- **Direktverkauf**: kein einziger Direktverkauf in den Demo-Daten — das Feature wirkt unbenutzt.
- **Umbuchung**: keine Umbuchung zwischen Tischen vorhanden.
- **TSE**: kein Event trägt TSE-Signaturdaten, alle TSE-Tabellen sind leer. Für ein
  Kassensystem, dessen Compliance-Story (KassenSichV, § 146a AO) zentral ist, fehlt damit
  ausgerechnet das Pflicht-Feature in der Demo: kein QR-Code auf Belegen, keine
  Ausfalldokumentation, leere TSE-Ansichten in den Einstellungen.
- **Betreiber-Stammdaten**: leer — obwohl sie fachlich Voraussetzung für die erste
  Kassensitzung sind. Die Demo widerspricht damit ihrer eigenen Invariante.
- **Druckerkonfiguration und Bondruck-Fehlerliste**: keine Drucker-IPs konfiguriert, keine
  Druckaufträge — Druckstationsseite und Fehlerliste sind in der Demo nicht demonstrierbar.
- **Kassenführung**: Geldtransit, Kassensturz und Soll/Ist-Differenz kommen nicht vor; die
  Tagesabschluss-Events enthalten Null-Summen statt der tatsächlichen Tagesumsätze.
- **Tisch-Favoriten**: leer.

Dazu kommt ein Wartungsproblem: Die seed.sql ist 872 KB groß, jede Event-Payload ist
handgepflegtes JSON. Bei jeder Änderung an Event-Formaten (laut Arbeitsregeln ausdrücklich
erlaubt und erwünscht) veraltet die Datei stillschweigend; Summen und Querbezüge (z. B.
Tagesabschluss-Beträge, der Zusammenfassungs-Kommentar am Dateiende) sind bereits heute
inkonsistent. Mit TSE-Daten in jedem fiskalischen Event würde sich die Datei etwa verdoppeln —
handgepflegt ist das nicht mehr beherrschbar.

## Solution

Ein **Go-Seeder** ersetzt die seed.sql vollständig. Er läuft als Subkommando des bestehenden
Backend-Binaries (analog zu `rebuild-projections`), erzeugt das Demo-Szenario programmatisch
über die Domain-Event-Konstruktoren und schreibt es direkt in die Datenbank. `make seed` und
`scripts/prod-reset-and-seed.sh` rufen das Subkommando auf; die seed.sql wird gelöscht.

Das Szenario bleibt das bewährte **3-Tage-Sommerfest des „TSV Musterstadt e.V."** (Freitag und
Samstag abgeschlossen, Sonntag offen, ~20 aktive Tische, ~1000 Events), wird aber zu einem
vollständigen, in sich konsistenten Betriebsablauf ausgebaut:

- **Betreiber-Stammdaten** des TSV Musterstadt e.V. sind gepflegt.
- **Alle fiskalischen Events tragen plausible Fake-TSE-Daten** (fortlaufende
  Transaktionsnummern und Signaturzähler, Signaturen, QR-Code-Daten im erwarteten Format).
  Die TSE-Konfiguration selbst bleibt leer — der Nachsignier-Worker bleibt dadurch inaktiv.
- **TSE-Ausfall am Samstagabend**: In einem erkennbaren Zeitfenster zur Stoßzeit tragen die
  Events keine TSE-Daten; stattdessen existieren Nachsignier-Aufträge (überwiegend erledigt
  mit nachgetragenen Signaturen in der Signatur-Seitentabelle, einzelne fehlgeschlagen mit
  Fehlertext, einer verworfen) — die gesetzlich geforderte Ausfalldokumentation ist damit
  sichtbar demonstrierbar.
- **Direktverkauf** als eigener Erzählstrang (z. B. Festbändchen- und Kuchenverkauf am
  Eingang): mehrere Direktverkäufe über alle drei Tage, mindestens ein Storno.
- **Umbuchungen** zwischen Tischen (Storno-/Bestellungs-Paar mit den Standard-Kommentaren),
  z. B. eine Gruppe, die vom Stehtisch an einen freien Tisch wechselt.
- **Kassenführung komplett**: Geldtransit (Entnahme zur Abschöpfung am Samstag, Einlage
  Wechselgeld am Sonntag), Kassensturz am Samstagabend mit kleiner Soll/Ist-Differenz und
  zugehöriger Differenz-Buchung, Tagesabschlüsse für Freitag und Samstag mit **aus den
  tatsächlichen Tages-Events berechneten Summen**.
- **Druckstationen** sind mit realistischen LAN-IPs und Bonmodi konfiguriert; **Druckaufträge**
  existieren in allen Status — insbesondere fehlgeschlagene mit Fehlertext für die
  Bondruck-Fehlerliste.
- **Tisch-Favoriten** für mehrere Service-Benutzer.

Der Seeder ist **deterministisch** (fester Zufalls-Seed, Zeitstempel relativ zum
Ausführungszeitpunkt): Jeder Lauf erzeugt dieselben Demo-Zustände, sodass Doku-Screenshots,
manuelle Tests und die jotti.rocks-Demo reproduzierbar bleiben. Läuft der Seeder gegen eine
Datenbank, die bereits Kassenjournal-Events enthält, bricht er mit einer klaren Fehlermeldung
ab und verweist auf den Reset-Weg (`make clean && make dev && make seed` bzw.
prod-reset-and-seed.sh).

## User Stories

### Interessent / Demo-Besucher (jotti.rocks)

1. Als Interessent möchte ich in der Demo ein realistisches, mehrtägiges Vereinsfest mit
   vollem Betrieb vorfinden, damit ich beurteilen kann, wie jotti im Ernstfall aussieht.
2. Als Interessent möchte ich auf gezahlten Belegen TSE-Signaturdaten und einen QR-Code sehen,
   damit ich erkenne, dass jotti die KassenSichV-Pflichten abdeckt.
3. Als Interessent möchte ich alle beworbenen Features (Direktverkauf, Umbuchung, Bondruck,
   TSE, Reporting) mit echten Beispieldaten belegt sehen, damit kein Feature leer oder
   unbenutzt wirkt.

### Vereins-Admin (Demo-Login als Admin)

4. Als Vereins-Admin möchte ich gepflegte Betreiber-Stammdaten (Vereinsname, Anschrift,
   Steuernummer) in den Einstellungen vorfinden, damit die Demo ihrer eigenen Regel
   „Betreiber vor erster Kassensitzung" genügt und ich sehe, wie die Daten auf Belegen landen.
5. Als Vereins-Admin möchte ich zwei abgeschlossene Kassensitzungen mit korrekten
   Tagesabschluss-Summen (Umsatz, Stornierungen, Auszahlungen, Geldtransit) sehen, damit ich
   dem Tagesabschluss-Feature vertrauen kann — statt der heutigen Null-Beträge.
6. Als Vereins-Admin möchte ich eine offene Kassensitzung mit laufendem Betrieb vorfinden,
   damit ich Kassenbestand, Live-Reporting und das Abschließen selbst ausprobieren kann.
7. Als Vereins-Admin möchte ich im Reporting drei Betriebstage vergleichen können (ruhiger
   Freitag, voller Samstag, laufender Sonntag), damit die Auswertung realistische Verläufe
   zeigt.
8. Als Vereins-Admin möchte ich im Reporting Umsätze nach Steuersätzen getrennt sehen
   (Regelsatz, ermäßigt, befreit), damit ich die steuerliche Auswertung nachvollziehen kann.
9. Als Vereins-Admin möchte ich Geldtransit-Buchungen (Abschöpfung, Wechselgeld-Einlage) in
   der Kassenhistorie sehen, damit ich verstehe, wie Bargeldbewegungen dokumentiert werden.
10. Als Vereins-Admin möchte ich einen dokumentierten Kassensturz mit Soll/Ist-Differenz und
    zugehöriger Differenz-Buchung vorfinden, damit ich den Umgang mit Kassendifferenzen sehe.
11. Als Vereins-Admin möchte ich in den Einstellungen konfigurierte Druckstationen (IPs,
    Bonmodi) vorfinden, damit die Druckerkonfiguration nicht leer wirkt und ich sie ändern
    kann.
12. Als Vereins-Admin möchte ich in der Bondruck-Fehlerliste fehlgeschlagene Druckaufträge mit
    verständlichem Fehlertext sehen, damit ich „Erneut versuchen" und „Verwerfen" ausprobieren
    kann.
13. Als Vereins-Admin möchte ich in den TSE-Einstellungen Nachsignier-Aufträge in
    verschiedenen Status sehen (erledigt, fehlgeschlagen, verworfen), damit ich die
    TSE-Ausfalldokumentation (AEAO zu § 146a) nachvollziehen kann.
14. Als Vereins-Admin möchte ich erkennen können, dass am Samstagabend ein TSE-Ausfall
    stattfand (Events ohne Signatur, Nachsignier-Aufträge mit Zeitraum und Grund), damit das
    Ausfall-Szenario als zusammenhängende Geschichte erzählt wird.
15. Als Vereins-Admin möchte ich Benutzer in allen Lebenszyklus-Zuständen (aktiv, inaktiv,
    gelöscht) und beiden Service-Rollen vorfinden, damit die Benutzerverwaltung realistisch
    befüllt ist.
16. Als Vereins-Admin möchte ich Produkte und Tische mit Soft-Delete- und Inaktiv-Beispielen
    vorfinden, damit Filter- und Statusverhalten demonstrierbar bleiben.

### Servicekraft (Demo-Login als Service)

17. Als Servicekraft möchte ich Tische in allen Zuständen vorfinden (leer, frisch bestellt,
    teilgeliefert, teilbezahlt, mit Guthaben/Auszahlung, abgeschlossen), damit ich jeden
    Arbeitsschritt an einem passenden Tisch ausprobieren kann.
18. Als Servicekraft möchte ich vorbelegte Tisch-Favoriten haben, damit ich die
    Favoriten-Funktion sofort sehe statt sie erst einrichten zu müssen.
19. Als Servicekraft möchte ich am Direktverkaufsstand vergangene Direktverkäufe (inklusive
    eines Stornos) sehen, damit ich den Direktverkaufs-Ablauf inklusive Stornierung
    nachvollziehen kann.
20. Als Servicekraft möchte ich in der Tisch-Historie eine Umbuchung sehen (Storno mit
    „Umbuchung auf Tisch …" und Bestellung mit „Umbuchung von Tisch …"), damit ich verstehe,
    wie ein Tischwechsel dokumentiert wird.
21. Als Servicekraft möchte ich realistische Bestellkommentare und Positionskombinationen
    sehen, damit die Demo wie echter Festbetrieb wirkt und nicht wie generierte Daten.

### Serviceleitung (Demo-Login als Serviceleitung)

22. Als Serviceleitung möchte ich erteilte Stornierungen mit Begründungen in der Historie
    sehen, damit der Stornierungs-Workflow (Rolle, Kommentarpflicht) demonstrierbar ist.

### Entwickler

23. Als Entwickler möchte ich Demo-Daten mit einem einzigen Befehl (`make seed`) einspielen,
    der Events, Nebentabellen und Projektionen in einem Schritt konsistent aufbaut.
24. Als Entwickler möchte ich, dass die Event-Payloads vom Seeder über dieselben
    Domain-Konstruktoren erzeugt werden wie im Live-Betrieb, damit Demo-Daten nie mehr
    stillschweigend von den Event-Schemata abweichen können.
25. Als Entwickler möchte ich, dass eine Event-Format-Änderung den Seeder per Compiler-Fehler
    oder Test bricht statt per veralteter JSON-Datei, damit ich die Demo-Daten beim Refactoring
    automatisch mitziehe.
26. Als Entwickler möchte ich deterministische Seed-Läufe, damit Screenshots, manuelle Tests
    und Fehlerberichte auf identischen Zuständen basieren.
27. Als Entwickler möchte ich, dass der Seeder auf einer nicht-leeren Datenbank sauber
    abbricht und den Reset-Weg nennt, damit ich Demo- und Arbeitsdaten nicht versehentlich
    vermische.
28. Als Entwickler möchte ich das Szenario („Drehbuch") getrennt von Engine und Persistenz
    lesen und anpassen können, damit ein neues Produkt oder ein zusätzlicher Tag eine kleine,
    lokale Änderung bleibt.
29. Als Entwickler möchte ich keine 872-KB-SQL-Datei mehr im Repo pflegen, damit Reviews und
    Diffs von Demo-Daten-Änderungen handhabbar bleiben.

### Demo-Betreiber (jotti.rocks)

30. Als Demo-Betreiber möchte ich die Demo-Datenbank per Skript zurücksetzen und neu seeden
    können (wie bisher prod-reset-and-seed.sh), damit die öffentliche Demo regelmäßig in einen
    definierten Zustand zurückkehrt.
31. Als Demo-Betreiber möchte ich, dass alle Demo-Logins das bekannte Passwort behalten,
    damit bestehende Anleitungen und die Login-Hinweise auf jotti.rocks gültig bleiben.
32. Als Demo-Betreiber möchte ich, dass die leere TSE-Konfiguration und die konfigurierten
    Drucker-IPs den Live-Demo-Betrieb nicht stören (kein Signierversuch, keine sichtbaren
    Fehler durch nicht erreichbare Drucker), damit die Demo unbeaufsichtigt stabil läuft.

## Implementation Decisions

### Format und Einbettung

- Der Seeder ist ein **Subkommando des Backend-Binaries** (analog zum bestehenden
  Projektions-Rebuild-Kommando). Kein eigenes Binary, keine neue Dependency.
- `make seed` ruft das Subkommando im Backend-Container auf; der bisherige separate
  Projektions-Rebuild-Schritt entfällt aus dem Make-Target, weil der Seeder den Rebuild
  selbst anstößt. `prod-reset-and-seed.sh` stellt vom psql-Import auf das Subkommando um.
- `database/seed.sql` wird **gelöscht**; alle Verweise (Makefile, Skripte, Doku) werden
  umgestellt.

### Modulschnitt (bestätigt)

1. **Szenario-Definition** — deklaratives, deterministisches „Drehbuch" des 3-Tage-Sommerfests:
   Stammdaten (Benutzer, Tische, Produkte/Varianten, Betreiber, Druckstationen,
   Tisch-Favoriten) und Ablauf pro Tag (Bestellzyklen pro Tisch, Direktverkäufe, Umbuchungen,
   Stornierungen, Auszahlungen, Geldtransit, Kassensturz, TSE-Ausfallfenster,
   Druckauftrags-Historie). Reine Daten, kein I/O.
2. **Szenario-Engine** — übersetzt das Drehbuch in korrekt versionierte Event-Streams:
   Subjects nach bestehender Konvention (Kassensitzung, Tisch-Session, Direktverkauf),
   fortlaufende Versionen pro Subject, Zeitstempel relativ zum Ausführungszeitpunkt
   (Freitag ≈ vor 2 Tagen, Sonntag = heute). Berechnet die Tagesabschluss-Summen aus den
   tatsächlich erzeugten Tages-Events. Tiefes Modul, isoliert testbar.
3. **Fake-TSE-Signierer** — versieht alle fiskalischen Events mit konsistenten TSE-Daten:
   global monoton steigende Transaktionsnummern und Signaturzähler, eine feste
   Fake-TSE-Seriennummer, plausible logTime-Paare, Signatur-Strings und QR-Code-Daten im
   selben Format, das der Belegdruck erwartet. Setzt das Ausfallfenster um: Events im Fenster
   erhalten keine TSE-Daten; stattdessen entstehen Nachsignier-Aufträge (überwiegend erledigt
   mit passenden Einträgen in der Signatur-Seitentabelle, einzelne fehlgeschlagen mit
   Fehlertext, einer verworfen). Tiefes Modul, isoliert testbar.
4. **Seed-Writer** — persistiert Stammdaten, Events, Druckaufträge und TSE-Seitentabellen
   transaktional, korrigiert die Identity-Sequenzen und stößt anschließend den vorhandenen
   Projektions-Rebuild an. Dünne Schicht über bestehenden Repositories/Queries; enthält den
   Guard (Abbruch bei nicht-leerem Kassenjournal mit Hinweis auf den Reset-Weg).
5. **CLI-Integration** — Verdrahtung des Subkommandos, Umstellung von Makefile und
   Reset-Skript, Löschung der seed.sql.

### Events und Konsistenz

- **Alle Events entstehen über die bestehenden Domain-Konstruktoren** (inklusive der
  MitTSE-Varianten). Der Seeder setzt anschließend nur den Zeitstempel auf den historischen
  Szenario-Zeitpunkt; Payload-Aufbau und Validierung bleiben vollständig in der Domain-Schicht.
- **Umbuchungen** nutzen dieselbe Semantik wie der Live-Befehl: ein Storno-Event auf dem
  Quell-Tisch und ein Bestellungs-Event auf dem Ziel-Tisch mit den Standard-Kommentaren
  („Umbuchung auf/von Tisch …").
- **Direktverkäufe** erhalten eigene Subjects mit fester (deterministischer) UUID je Verkauf.
- **Tagesabschluss-Summen** (Umsatz, Stornierungen, Auszahlungen, Geldtransit) werden von der
  Engine aus den Events des jeweiligen Tages berechnet — keine handgepflegten Beträge mehr.
- Die Verteilung der Events bleibt in der Größenordnung des bisherigen Szenarios
  (ruhiger Freitag ~160, voller Samstag ~700, laufender Sonntag ~130 Events), inklusive der
  bewährten Tisch-Dramaturgie (Stammtisch, Geburtstagsfeier, Stoßzeiten) plus der neuen
  Erzählstränge (Direktverkaufsstand, Umbuchungen, TSE-Ausfall, Kassensturz).

### TSE-Abbildung

- **TSE-Konfiguration bleibt leer.** Dadurch bleibt der Nachsignier-Worker nachweislich
  inaktiv und der Live-Demo-Betrieb versucht keine echten Signaturen. Konsequenz (akzeptiert):
  live erzeugte Events tragen keine TSE-Daten.
- **Fiskalische Events** (Bestellung, Zahlung, Stornierung, Auszahlung, Direktverkauf und
  -storno, Geldtransit, Differenz, Tagesabschluss) tragen Fake-TSE-Daten; nicht-fiskalische
  (Ausgabe, Kassensitzungs-Eröffnung) nicht — wie im Live-Betrieb.
- **Ausfallfenster**: Samstagabend zur Stoßzeit (ca. 45–90 Minuten). Die Nachsignier-Aufträge
  dokumentieren Beginn, Ende und Grund des Ausfalls (Felder der bestehenden Outbox-Tabelle).
  Für erledigte Aufträge existieren die nachgetragenen Signaturen in der Seitentabelle, sodass
  die Belegansicht sie auflösen kann.

### Bondruck

- **Druckstationen** werden mit realistischen LAN-IPs und gemischten Bonmodi konfiguriert
  (alle fünf Stationen, inklusive Kassenbeleg und Abholbon).
- **Druckaufträge** werden als Historie geseedet: überwiegend gedruckt, einige offen, mehrere
  fehlgeschlagen mit verständlichem Fehlertext (z. B. Drucker nicht erreichbar) für die
  Fehlerliste, einer verworfen. Beide Bon-Arten (Arbeitsbon, Kassenbeleg) kommen vor.
- Dass im Live-Demo-Betrieb neue Druckaufträge dauerhaft „offen" bleiben (kein Relay), ist
  akzeptiert — offene Aufträge sind in der Oberfläche nicht sichtbar, die Fehlerliste zeigt
  nur fehlgeschlagene.

### Determinismus und Guard

- Fester Zufalls-Seed; identische Läufe erzeugen identische Daten (bis auf den Zeitanker
  „jetzt"). Feste UUIDs nach erkennbarem Schema, wo Events IDs brauchen.
- Guard: Enthält das Kassenjournal bereits Events, bricht der Seeder ohne Änderungen ab und
  nennt den Reset-Weg. Kein `--force`, kein Löschen — das Kassenjournal ist bewusst
  append-only und schreibgeschützt.

## Testing Decisions

Gute Tests prüfen **beobachtbares Verhalten und Invarianten**, nicht Implementierungsdetails:
Sie behaupten nichts über interne Strukturen der Engine, sondern über Eigenschaften der
erzeugten Daten, die fachlich garantiert sein müssen.

- **Szenario-Engine (Unit, `-tags=unit`)**:
  - Versionen pro Subject beginnen bei 1 und sind lückenlos monoton.
  - Zeitstempel innerhalb eines Subjects sind monoton steigend; jeder Event-Zeitstempel liegt
    im Zeitraum seiner Kassensitzung.
  - Tagesabschluss-Summen stimmen mit den aus den Tages-Events aggregierten Werten überein.
  - Umbuchungen sind paarweise konsistent (Storno- und Bestellpositionen identisch,
    Kommentare nach Konvention).
  - Salden gehen auf: Für abgeschlossene Tische Bestellungen − Stornierungen − Zahlungen
    (± Auszahlung) = 0; die dokumentierten „offenen" Zustände am Sonntag haben die erwarteten
    Salden.
- **Fake-TSE-Signierer (Unit, `-tags=unit`)**:
  - Transaktionsnummern und Signaturzähler sind global streng monoton.
  - Genau die fiskalischen Event-Typen erhalten TSE-Daten; Events im Ausfallfenster keine.
  - Für jedes Event im Ausfallfenster existiert genau ein Nachsignier-Auftrag; erledigte
    Aufträge haben einen passenden Signatur-Eintrag.
  - Erzeugte TSE-Daten bestehen die bestehende Domain-Validierung.
- **Seeder Ende-zu-Ende (Integration, bestehendes Integrationstest-Setup gegen echte
  Postgres-DB)**:
  - Ein kompletter Lauf auf frischer DB schlägt nicht fehl; alle Feature-Tabellen sind befüllt
    (Betreiber, Druckstationen, Druckaufträge, Nachsignier-Aufträge, Signaturen, Favoriten).
  - Projektionen sind konsistent: ein erneuter Rebuild ändert nichts.
  - Guard: zweiter Lauf bricht mit Fehler ab und schreibt nichts.
- **Prior Art**: Unit-Tests der Domain-Schicht (Event-Konstruktoren, Direktverkauf,
  Tisch-Session) und die bestehenden Integrationstests über das Repo-übliche
  Integrationstest-Skript.

## Out of Scope

- **Echte TSE-Signierung in der Demo** (gefüllte TSE-Konfiguration, fiskaly-TEST-Credentials)
  — die Demo bleibt ohne Live-Signierung; siehe `docs/prds/prd-tse-setup-wizard.md` für den
  Einrichtungsweg echter Instanzen.
- **Print-Relay in der Demo-Umgebung** — geseedete Druckaufträge sind Historie; es wird nicht
  wirklich gedruckt.
- **Frontend-Änderungen** — alle Ansichten existieren bereits; dieses PRD befüllt sie nur.
- **Schema-Änderungen** — der Seeder arbeitet gegen das bestehende Schema.
- **Mehrere wählbare Szenarien oder Parametrisierung** (Größe, Dauer, Branche) — es gibt genau
  ein Drehbuch.
- **DSFinV-K-Export-Validierung der Fake-TSE-Daten gegen externe Prüfwerkzeuge** — die
  Fake-Daten müssen jottis eigene Validierung und Darstellung bedienen, nicht die
  Finanzverwaltung.
- **Automatischer periodischer Demo-Reset auf jotti.rocks** — das manuelle Reset-Skript bleibt
  der Weg.

## Further Notes

- **Bewusste Demo-Inkonsistenz:** Geseedete (historische) Events tragen TSE-Daten, live in der
  Demo erzeugte nicht — und live erzeugte Druckaufträge bleiben „offen". Beides ist die
  akzeptierte Folge davon, dass die Demo ohne echte TSE und ohne Drucker läuft. Sollte das in
  der Praxis verwirren, wäre ein Demo-Hinweis im Frontend ein separates, kleines Folgethema.
- **Demo-Logins:** Benutzerliste und Passwort (`jotti123`) bleiben unverändert, damit
  bestehende Anleitungen gültig bleiben.
- **Warum kein SQL mehr:** Der entscheidende Hebel ist nicht die Dateigröße, sondern die
  Kopplung an die Event-Schemata. Solange Demo-Daten als JSON-Literale gepflegt werden,
  veralten sie bei jeder erlaubten Breaking Change unbemerkt. Über die Domain-Konstruktoren
  bricht stattdessen der Build — der Seeder kann gar nicht veralten.
- **Zeitanker:** Wie bisher liegen die Tage relativ zu „jetzt" (Freitag ≈ −2 Tage, Sonntag =
  heute), damit die offene Kassensitzung immer „heute" ist und Live-Reporting sinnvoll wirkt.
- Die Zusammenfassung am Ende der alten seed.sql (inkl. Verweise auf nicht mehr existierende
  Event-Typen) entfällt ersatzlos; die Szenario-Definition im Code ist künftig die einzige
  Quelle der Wahrheit über das Drehbuch.
