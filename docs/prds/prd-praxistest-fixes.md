# PRD: Praxistest-Fixes Runde 1 (Drawer-Layout, Kassenabschluss-Supportbarkeit, Druck-Nachdruck)

> Quelle: Praxistest eines Vereins am 07.07.2026 auf v0.14.0 (lokale
> Installation, iPhone-Browser im Service, Windows-Laptop als Admin).
> Vier Rückmeldungen, alle im Code verifiziert. Ausgeliefert wird als
> vorgezogenes Patch-Release (Arbeitstitel v0.14.1), unabhängig vom
> v1.0.0-Zug.

## Problem Statement

Beim ersten echten Einsatz sind vier Probleme aufgetreten:

1. Große Bestellungen lassen sich nicht stornieren. Im
   Stornierungs-Drawer läuft die Positionsliste ab etwa 4 Positionen
   optisch über das Pflicht-Kommentarfeld und die Buttons hinweg; das
   Kommentarfeld ist nicht mehr bedienbar, die Stornierung damit
   blockiert.
2. Bei langen Produkt- und Variantennamen (z. B. "Süßgetränke
   Johannisbeersaftschorle 0,5L") wird die Zeile breiter als der
   Bildschirm; der Plus-Button rutscht unerreichbar nach rechts, es
   lassen sich keine Positionen für die Stornierung auswählen.
3. Beim Kassenabschluss erschien der generische Toast "Es ist ein
   unerwarteter Serverfehler aufgetreten". Der Server lief und hat
   bewusst mit 500 geantwortet; der konkrete Auslöser steht nur in den
   Server-Logs des Geräts, auf die kein Zugriff besteht. Der Verein
   kann den Fehler weder selbst einordnen noch sinnvoll melden.
4. Wird während des Betriebs die Papierrolle gewechselt, gehen die in
   dieser Zeit anfallenden Bons verloren: Das Relay versucht die
   Zustellung im 2-Sekunden-Takt, nach 3 Fehlversuchen (also nach etwa
   6 Sekunden) ist der Auftrag endgültig fehlgeschlagen und wird nie
   wieder automatisch gedruckt. Ein Rollenwechsel dauert aber 1 bis 3
   Minuten und lässt so reihenweise Aufträge in den Endzustand kippen.

Technische Root Causes (verifiziert):

- **Punkte 1 und 2** haben dieselbe Ursache: Die Radix-ScrollArea in den
  Drawern. Die Höhenkette ist gebrochen (Viewport mit 100 %-Höhe gegen
  einen Root mit automatischer Höhe), dadurch scrollt die Liste nie und
  läuft mangels overflow-hidden optisch über die nachfolgenden Elemente.
  Zusätzlich wickelt Radix den Inhalt in ein display:table-Div mit
  Max-Content-Breite, wodurch truncate nie greift und lange Namen die
  Zeile aufblähen. Betroffen sind fünf Stellen: Stornierungs-Drawer,
  Umbuchungs-Drawer, Direktverkauf-Storno-Drawer sowie der Beleg
  (Receipt) im Zahlungs- und im Ausgabe-Drawer. Der Kassieren-Flow hat
  bei vielen Positionen also dasselbe Problem wie der Storno.
- **Punkt 3**: v0.14.0 enthält weder die Panic-Recovery-Middleware noch
  den idempotenten Kassenabschluss-Wiederanlauf (beides liegt bereits
  auf main). Ein wiederholter Abschluss-Versuch nach Teilfehler konnte
  auf v0.14.0 ein zweites Kassensturz-Event ins Journal schreiben.
  Der generische 500er-Toast nennt keine Referenz, mit der sich ein
  gemeldeter Fehler später einem Log-Eintrag zuordnen ließe, obwohl
  jede Anfrage bereits eine Correlation-ID trägt.
- **Punkt 4**: Versuchszähler statt Zeitfenster. Die Grenze von 3
  Versuchen ist für dauerhafte Fehler gedacht, wird aber von jeder
  kurzen Unterbrechung (Rollenwechsel, Drucker-Neustart) in Sekunden
  aufgebraucht. Kassenbelege laufen über dieselbe Mechanik und sind
  genauso betroffen.

## Solution

1. **Drawer-Listen scrollen nativ.** Die Radix-ScrollArea wird in allen
   fünf Stellen durch natives overflow-y-auto mit dvh-basierter
   Maximalhöhe ersetzt (Vorbild: Tischauswahl-Drawer, der es bereits so
   macht). Kommentarfeld, Summen und Buttons liegen außerhalb des
   Scrollbereichs und sind immer sichtbar und bedienbar. Lange Namen
   werden per truncate abgeschnitten, Plus/Minus bleiben an ihrem Platz.
   Der dreifach duplizierte Zeilenblock (Name, Preis, Minus/Anzahl/Plus)
   wird als gemeinsame Komponente PositionAuswahlListe extrahiert.
2. **Kassenabschluss supportbar machen.** Die bereits auf main liegende
   Härtung (Panic-Recovery, idempotenter Wiederanlauf) wird mit dem
   Patch-Release ausgeliefert. Zusätzlich nennt der generische
   Serverfehler-Toast künftig die Correlation-ID der fehlgeschlagenen
   Anfrage als Fehler-Referenz, sodass eine Meldung aus dem Feld einem
   Log-Eintrag zugeordnet werden kann.
3. **Druckaufträge mit Backoff-Nachdruck.** Fehlversuche führen nicht
   mehr sofort zum nächsten Versuch im Poll-Takt: Nach jedem Fehlversuch
   wartet der Auftrag eine wachsende Zeitspanne (5s, 15s, 30s, 1m, 3m),
   bevor er wieder ausgeliefert wird. Nach dem Erstversuch und fünf
   Wiederholungen (rund 5 Minuten nach dem ersten Versuch) gilt er als
   endgültig fehlgeschlagen, denn ein noch späterer Bon ist in Küche
   und Ausschank wertlos. Ein Rollenwechsel wird damit vollständig überbrückt; die
   Aufträge drucken automatisch nach, sobald der Drucker wieder
   erreichbar ist. Für den Fall "Rolle war lange leer, keiner hat es
   gemerkt" bekommt die Druckstationen-Seite zusätzlich zum bestehenden
   Einzel-Retry einen Sammel-Button, der alle fehlgeschlagenen Aufträge
   erneut einreiht.

## User Stories

1. Als Servicekraft möchte ich auch aus einer großen Bestellung (10 und
   mehr Positionen) stornieren können, damit ein voller Tisch kein
   Sonderfall ist: Kommentarfeld und "Stornierung erteilen" sind immer
   erreichbar.
2. Als Servicekraft möchte ich die Positionsliste im Stornierungs-Drawer
   mit dem Finger scrollen können (natives Momentum-Scrolling), damit
   ich alle Positionen sehe und auswählen kann.
3. Als Servicekraft möchte ich bei langen Produkt- und Variantennamen
   weiterhin Plus- und Minus-Buttons erreichen, damit die Auswahl nicht
   an der Namenslänge scheitert; überlange Namen werden abgeschnitten
   dargestellt.
4. Als Servicekraft möchte ich dieselben Garantien beim Umbuchen auf
   einen anderen Tisch, damit auch dieser Drawer bei vielen Positionen
   bedienbar bleibt.
5. Als Servicekraft möchte ich dieselben Garantien beim
   Direktverkauf-Storno, damit alle drei Auswahl-Drawer identisch
   funktionieren.
6. Als Servicekraft möchte ich beim Kassieren eines vollen Tischs den
   Beleg scrollen können, während Trinkgeld- und Erhalten-Feld,
   Kommentar und "Kassieren" sichtbar bleiben, damit die Zahlung nicht
   an der Positionszahl scheitert.
7. Als Servicekraft möchte ich bei der Ausgabe-Bestätigung eines vollen
   Tischs dieselben Garantien, damit auch dort nichts verdeckt wird.
8. Als Servicekraft möchte ich, dass das Kommentarfeld beim Fokussieren
   nicht hinter der Bildschirmtastatur verschwindet, damit ich den
   Pflichtkommentar auf dem Handy tippen kann.
9. Als Kassenverantwortlicher möchte ich, dass ein erneuter
   Abschluss-Versuch nach einem Fehler kein zweites Kassensturz-Event
   ins Journal schreibt, damit das Kassenjournal fiskalisch sauber
   bleibt.
10. Als Kassenverantwortlicher möchte ich, dass ein Absturz einer
    Hintergrund-Komponente den Server nicht beendet, damit die Kasse im
    Betrieb nicht stehen bleibt.
11. Als Admin möchte ich bei einem unerwarteten Serverfehler eine kurze
    Fehler-Referenz im Toast sehen, damit ich sie bei einer Meldung an
    den Betreiber nennen kann.
12. Als Betreiber möchte ich eine gemeldete Fehler-Referenz direkt im
    Server-Log wiederfinden, damit ich Praxistest-Berichte ohne
    Rätselraten diagnostizieren kann.
13. Als Servicekraft möchte ich, dass Bons, die während eines
    Rollenwechsels anfallen, nach dem Wechsel automatisch nachgedruckt
    werden, damit keine Bestellung verloren geht.
14. Als Küchen- oder Ausschank-Helfer möchte ich, dass nachgedruckte
    Bons in ihrer ursprünglichen Reihenfolge kommen, damit die
    Abarbeitung stimmt.
15. Als Admin möchte ich, dass ein Auftrag nach rund 5 Minuten
    erfolgloser Versuche endgültig aufgibt, damit nicht irgendwann
    veraltete Bons drucken, die niemand mehr zuordnen kann.
16. Als Admin möchte ich auf der Druckstationen-Seite alle
    fehlgeschlagenen Aufträge mit einem Klick erneut einreihen können,
    damit ich nach einer länger unbemerkten Störung nicht jeden Auftrag
    einzeln anfassen muss.
17. Als Admin möchte ich weiterhin einzelne fehlgeschlagene Aufträge
    erneut einreihen oder verwerfen können, damit ich gezielt
    entscheiden kann, was noch gedruckt werden soll.
18. Als Admin möchte ich auf dem Dashboard einen Hinweis sehen, wenn
    Druckaufträge fehlgeschlagen sind, damit ich die Störung überhaupt
    bemerke.
19. Als Vereins-Admin möchte ich das Patch-Release über den normalen
    Update-Weg einspielen können, ohne Daten zu verlieren, damit der
    nächste Einsatz abgesichert ist.
20. Als Kassenbeleg-Empfänger möchte ich, dass auch Kassenbelege vom
    Backoff-Nachdruck profitieren, da sie über dieselbe Druckmechanik
    laufen.

## Implementation Decisions

- **PositionAuswahlListe** wird als gemeinsame Service-Komponente
  extrahiert: Zeile mit Name (truncate), Einzelpreis und Menge,
  Minus/Anzahl/Plus. Sie kapselt den Scrollbereich (natives
  overflow-y-auto, Maximalhöhe dvh-basiert statt fester rem-Wert) und
  wird von Stornierungs-, Umbuchungs- und Direktverkauf-Storno-Drawer
  verwendet. Die Mengenlogik (add/remove/Grenzen) bleibt im bestehenden
  Mengen-Hook.
- **Receipt** behält seine Schnittstelle (Positionen, Gesamtsumme) und
  ersetzt intern die ScrollArea durch natives Scrollen; damit sind
  Zahlungs- und Ausgabe-Drawer abgedeckt.
- **Die ScrollArea-UI-Komponente wird entfernt**, sobald der letzte
  Nutzer umgestellt ist. Radix ScrollArea bringt auf Touch-Geräten
  keinen Mehrwert und hat zwei strukturelle Fallen (Höhenkette,
  table-Wrapper); natives Scrollen ist das Vorbild-Muster aus dem
  Tischauswahl-Drawer.
- **Drawer-Gesamtlayout**: Header, Liste (scrollt), Summen,
  Kommentarfeld und Footer-Buttons; nur die Liste scrollt. Die
  dvh-Maximalhöhe der Liste ist so bemessen, dass Kommentarfeld und
  Buttons auch mit geöffneter Bildschirmtastatur erreichbar bleiben.
- **Druckauftrag-Backoff** lebt vollständig im Backend; das Relay bleibt
  unverändert dumm (pollt, druckt, meldet). Die Druckauftrags-Tabelle
  bekommt per Migration einen Zeitstempel "nächster Versuch ab". Die
  Abfrage offener Aufträge liefert nur fällige Aufträge; die
  Ergebnis-Meldung setzt bei einem Fehlversuch den nächsten
  Fälligkeitszeitpunkt anhand einer reinen Backoff-Funktion
  (Fehlversuchsnummer zu Wartezeit: 5s, 15s, 30s, 60s, 180s). Die
  maximale Versuchszahl steigt auf 6 (Erstversuch plus fünf
  Wiederholungen); die bestehende Endzustands-Semantik
  (fehlgeschlagen, Einzel-Retry setzt Versuche zurück) bleibt.
- **Sammel-Retry**: neuer Command und Endpoint "alle fehlgeschlagenen
  Aufträge erneut einreihen" (Status-Guard wie beim Einzel-Retry),
  dazu ein Button auf der Druckstationen-Seite mit Bestätigungsdialog.
- **Fehler-Referenz**: 500-Antworten enthalten die bereits existierende
  Correlation-ID der Anfrage. Der Frontend-Fehlerpfad reicht sie durch
  und hängt sie an die generische Serverfehler-Meldung an ("Referenz:
  a1b2c3d4"). Kein neues Tracing, keine neue Infrastruktur; dieselbe ID
  steht schon heute in jeder Log-Zeile der Anfrage.
- **Schema-Änderung als Migration**, nicht per Edit des
  Initial-Schemas: Seit der Erstinstallation am 07.07.2026 gilt der
  De-facto-Freeze; der Upgrade-Pfad wird von der bestehenden
  CI-Harness geprüft.
- **Release-Weg**: vorgezogenes Patch-Release von main (Arbeitstitel
  v0.14.1, endgültige Nummer nach Release-Konvention beim Schnitt, da
  Migration und Sammel-Retry eher ein Minor sind). Es enthält damit
  automatisch die bereits gemergten Nacharbeit-Blöcke inklusive
  Panic-Recovery und idempotentem Wiederanlauf.

> **Assumption:** Der konkrete Auslöser des 500ers beim Verein bleibt
> ungeklärt (kein Log-Zugriff). Die Maßnahme ist Härtung plus
> Supportbarkeit statt Root-Cause-Fix; tritt der Fehler erneut auf,
> macht die Fehler-Referenz ihn diagnostizierbar.

## Testing Decisions

- Gute Tests prüfen externes Verhalten über die öffentliche
  Schnittstelle, keine Implementierungsdetails; Layout-/CSS-Verhalten
  ist in jsdom nicht sinnvoll testbar und wird stattdessen per
  Playwright-Abnahme verifiziert.
- **Backoff-Logik**: Unit-Tests für die reine Backoff-Funktion
  (Versuchsnummer zu Wartezeit, Grenzen). Repository-/Integrationstests:
  nicht fällige Aufträge werden nicht ausgeliefert, Fehlversuch setzt
  die nächste Fälligkeit, 5. Fehlversuch führt in den Endzustand,
  Sammel-Retry reiht nur fehlgeschlagene Aufträge wieder ein. Prior
  Art: bestehende Journal- und Relay-Integrationstests sowie die
  vorhandenen Druckauftrags-Handler-Tests.
- **PositionAuswahlListe**: Komponententests mit Vitest und Testing
  Library: Zeilen rendern, Plus/Minus ändern Mengen innerhalb der
  Grenzen (0 bis Bestellmenge), Auswahl meldet sich an den Drawer.
  Prior Art: bestehende Drawer- und StickyActionBar-Tests im
  Service-Bereich.
- **Fehler-Referenz**: Test des Fehlermeldungs-Moduls, dass die
  generische 500er-Meldung die Referenz aus der Antwort anzeigt (und
  ohne Referenz unverändert bleibt); Backend-Test, dass 500-Antworten
  die Correlation-ID enthalten.
- **Playwright-Abnahme** am Mobil-Viewport (iPhone-Maße) für alle fünf
  umgestellten Stellen, jeweils mit vielen Positionen und überlangen
  Namen: Kommentarfeld und Buttons sichtbar, Liste scrollt, nichts
  läuft über. Prior Art: gebündeltes Chromium headless wie bei den
  bisherigen Screenshot-Abnahmen.

## Out of Scope

- Root-Cause-Diagnose des konkreten Kassenabschluss-500ers vom
  07.07.2026 (keine Logs verfügbar).
- Prüfung des Vereins-Journals auf ein doppeltes Kassensturz-Event aus
  einem möglichen v0.14.0-Wiederholungsversuch; das ist ein manueller
  Schritt beim nächsten Kontakt mit dem Gerät (SQL gegen das
  Kassenjournal), kein Produkt-Feature.
- Voller Mobile-UX-Audit über die verifizierten Muster hinaus (eigener
  ux-review-Durchlauf, separat).
- Kassenabschluss-Bedienhilfen wie eine Warnung bei großer
  Soll-Ist-Differenz oder ein Pflicht-Zählschritt (der Verein hat mit
  Ist-Bestand 0,00 € abgeschlossen); gehört thematisch zur
  Kassenabschluss-Vereinfachung und deren PRD.
- Live-Statusanzeige der Drucker (erreichbar/nicht erreichbar) im
  Admin-Bereich.
- Änderungen am Relay (Poll-Takt, Zustell-Logik); der Backoff wird
  vollständig serverseitig gesteuert.

## Further Notes

- Der Tischauswahl-Drawer ist das Vorbild für natives Scrollen und
  bleibt unverändert.
- Der Dashboard-Hinweis auf fehlgeschlagene Druckaufträge liegt bereits
  auf main und wird mit diesem Release erstmals ausgeliefert.
- Die Bestellseite (Produkt-/Variantenliste) ist vom ScrollArea-Problem
  nicht betroffen: Sie scrollt nativ auf Seitenebene und hat
  Touch-taugliche Buttongrößen.
- Beim Sammel-Retry drucken Aufträge in ID-Reihenfolge pro Drucker
  (bestehende Gruppierung im Relay), damit Bon-Reihenfolgen erhalten
  bleiben.
