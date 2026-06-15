# PRD: Vereinfachter Kassenabschluss

## Problem Statement

Der Kassenwart (Rechner) eines Vereins muss am Ende eines Veranstaltungstags zwei
getrennte Schritte ausführen, um die Kasse zu schließen. Zuerst den Kassensturz: das
Bargeld zählen und den gezählten Ist-Bestand eintragen. Danach, in einem zweiten Block
weiter unten auf der Seite, den Tagesabschluss: einen destruktiven Button drücken, der die
Kassensitzung unwiderruflich schließt. Beide Schritte hängen logisch zusammen ("Ich schließe
die Kasse für heute ab"), wirken in der Oberfläche aber wie zwei unabhängige Vorgänge.

Für nicht-technische Vereinshelfer entstehen daraus mehrere Probleme:

- Der Zusammenhang der beiden Schritte ist nicht erkennbar. Nach dem Kassensturz ist nicht
  klar, dass noch ein Pflichtschritt folgt.
- Die beim Kassensturz berechnete Differenz zwischen Soll und Ist wird nirgends angezeigt.
  Der Helfer erfährt nicht, ob und wie hoch die Kassendifferenz war, obwohl genau diese
  Differenz fiskalisch ehrlich zu dokumentieren ist.
- Der Soll-Bestand steht in einer Info-Karte ganz oben auf der Seite, das Ist-Eingabefeld
  weiter unten. Beim Zählen und Abgleichen muss der Helfer zwischen beiden hin- und
  herscrollen.
- Der Tagesabschluss ist unwiderruflich, hat aber keinen Bestätigungsdialog. Ein Fehltipp
  auf dem Smartphone schließt den Tag. Weniger kritische Löschaktionen an anderer Stelle im
  Admin-Bereich nutzen dagegen bereits einen Bestätigungsdialog.
- Die Fiskal-Begriffe (Geldtransit, Kassensturz, Tagesabschluss, Soll-Kassenbestand) stehen
  ohne Erklärung in der Oberfläche. Nur der Tagesabschluss hat einen einzigen Hinweissatz.
  Die fiskalisch heikelste Seite ist damit die am wenigsten erklärte, während die
  Einstellungen-Seite jeden Block mit einem kurzen Erklärabsatz versieht.
- Beim Geldtransit nutzt die Auswahl zwischen Einlage und Entnahme kleine native
  Radio-Schaltflächen. Auf dem Smartphone sind ihre Trefferflächen knapp.

Zusätzlich besteht ein technisches Problem ohne direkten Nutzerbezug, das beim Verstehen des
Ablaufs auffiel: Das Event `tagesabschluss-erstellt:v1` schreibt die Felder
`umsatzGesamtCents`, `stornierungCents`, `auszahlungenCents` und `geldtransitCents`
produktiv immer als 0. Diese Felder werden im Produktivcode nirgends gelesen (das Reporting
kommt vollständig aus separaten SQL-Aggregationen). Der Seed-Generator füllt dieselben Felder
dagegen mit echten Tagessummen. Seed-Events und Produktions-Events haben damit
unterschiedliche Semantik für dieselben Felder, und ein Prüfer oder Entwickler, der den
Event-Stream liest, sieht produktiv irreführende Nullen.

Eine kleinere, ebenfalls entwicklerseitige Unstimmigkeit betrifft die Benennung: Die
Frontend-Komponente für den Geldtransit heißt intern KassenbewegungSection, obwohl der
nutzersichtbare Titel und der verbindliche Fachbegriff Geldtransit lauten.

## Solution

Der Kassensturz und der Tagesabschluss werden zu einem geführten Schritt "Kasse abschließen"
zusammengelegt. Aus Sicht des Kassenwarts gibt es am Tagesende nur noch einen Vorgang:

- Der Helfer zählt das Bargeld und trägt den Ist-Bestand in ein Feld ein.
- Beim Klick auf "Kasse abschließen" öffnet sich ein Bestätigungsdialog, der Soll-Bestand,
  gezählten Ist-Bestand und die Differenz gegenüberstellt, zusammen mit den Tagessummen als
  Z-Bon-Vorschau. Der Dialog macht klar, dass der Abschluss unwiderruflich ist.
- Bestätigt der Helfer, führt das System in einem Vorgang alles aus: die Differenzbuchung
  (falls die Kasse nicht exakt stimmt), den Z-Bon (Tagesabschluss) und das Schließen der
  Kassensitzung.
- Bricht der Helfer ab, passiert nichts. Er kann nachzählen und es erneut versuchen. Damit
  ist der Dialog zugleich die Gelegenheit, eine unerwartet hohe Differenz vor dem Abschluss
  zu prüfen.

Der separate Kassensturz-Schritt entfällt aus der Bedienung. Fiskalisch bleibt der
Zählvorgang erhalten: Im Hintergrund wird weiterhin ein Kassensturz-Event geschrieben, und
eine von Null verschiedene Differenz erzeugt weiterhin eine eigene, signierte
Differenzbuchung. Die Reihenfolge und die Pflichten aus `docs/compliance.md` bleiben
unangetastet. Die Sperre, dass alle Tische einen Saldo von Null haben müssen, bleibt
ebenfalls bestehen.

Begleitend erhält der Kassen-Admin-Bereich durchgehend kurze Erklärabsätze im bereits
etablierten Stil der Einstellungen-Seite. Jeder Block (Eröffnen, Geldtransit, Abschluss)
bekommt einen Hilfetext, und die Fiskal-Begriffe werden mit einer Alltagsübersetzung
versehen, ohne die verbindlichen Fachbegriffe selbst zu ersetzen.

Die Auswahl zwischen Einlage und Entnahme beim Geldtransit wird auf größere,
fingerfreundliche Auswahlflächen umgestellt, damit sie auf dem Smartphone sicher zu treffen
ist.

Der Code-Smell der Null-Beträge wird behoben, indem das Tagesabschluss-Event mit den echten
Tagessummen befüllt wird. Das Event wird damit zu einer in sich geschlossenen, signierten
Z-Bon-Momentaufnahme, und dieselben Summen können direkt im Bestätigungsdialog als Vorschau
dienen. Die Inkonsistenz zwischen Seed und Produktion verschwindet.

## User Stories

### Kassenwart / Rechner (Abschluss)

1. Als Kassenwart möchte ich die Kasse am Tagesende in einem einzigen geführten Schritt
   abschließen, damit ich nicht zwei getrennte Vorgänge als zusammengehörig erkennen muss.
2. Als Kassenwart möchte ich nach dem Zählen den gezählten Ist-Bestand in ein Feld eingeben,
   damit das System die Differenz zum Soll selbst berechnet.
3. Als Kassenwart möchte ich vor dem endgültigen Abschluss eine Gegenüberstellung von Soll,
   Ist und Differenz sehen, damit ich erkenne, ob die Kasse stimmt.
4. Als Kassenwart möchte ich im selben Dialog die Tagessummen (Umsatz, Stornierungen,
   Auszahlungen, Geldtransit) als Z-Bon-Vorschau sehen, damit ich den Abschluss prüfen kann,
   bevor ich ihn bestätige.
5. Als Kassenwart möchte ich den unwiderruflichen Abschluss ausdrücklich bestätigen müssen,
   damit ich ihn nicht versehentlich per Fehltipp auslöse.
6. Als Kassenwart möchte ich den Abschluss aus dem Dialog heraus abbrechen können, damit ich
   bei einer unerwarteten Differenz erst nachzählen kann, ohne etwas zu verändern.
7. Als Kassenwart möchte ich, dass eine von Null verschiedene Differenz automatisch und
   ehrlich gebucht wird, damit meine Kasse nicht künstlich "auf den Cent genau" wirkt.
8. Als Kassenwart möchte ich eine klare Rückmeldung erhalten, dass die Kasse abgeschlossen
   ist, damit ich weiß, dass der Tag vollständig erfasst wurde.
9. Als Kassenwart möchte ich, dass der Abschluss abgelehnt wird, solange noch ein Tisch einen
   offenen Saldo hat, damit ich keinen Tag abschließe, während noch Geld aussteht.
10. Als Kassenwart möchte ich beim Abschluss-Block sehen, was zu tun ist (Bargeld zählen,
    gezählten Betrag eintragen, dass kleine Differenzen normal sind), damit ich den Vorgang
    ohne Vorwissen korrekt durchführe.

### Kassenwart / Rechner (laufender Betrieb und Eröffnung)

11. Als Kassenwart möchte ich beim Eröffnen erklärt bekommen, dass der Anfangsbestand das
    vorab gezählte Wechselgeld ist, damit ich den richtigen Betrag eintrage.
12. Als Kassenwart möchte ich beim Geldtransit kurz erklärt bekommen, was eine Einlage und
    eine Entnahme ist und wann ich sie buche, damit ich Bargeldbewegungen korrekt erfasse.
13. Als Kassenwart möchte ich zu den Fiskal-Begriffen eine Alltagsübersetzung sehen, damit
    ich die Oberfläche ohne Steuerwissen verstehe.
14. Als Kassenwart möchte ich Einlage und Entnahme auf dem Smartphone mit einer großen Fläche
    sicher antippen können, damit ich mich beim Geldtransit nicht vertippe.

### Betriebsprüfer / Auditierbarkeit

15. Als Betriebsprüfer möchte ich, dass das Tagesabschluss-Event die tatsächlichen
    Tagessummen signiert enthält, damit der Z-Bon als in sich geschlossene Momentaufnahme
    nachvollziehbar ist und nicht aus irreführenden Nullwerten besteht.

### Entwickler / Wartung

16. Als Entwickler möchte ich, dass Seed-Daten und produktive Daten dieselbe Semantik für die
    Tagesabschluss-Felder haben, damit ich beim Lesen des Event-Streams keine widersprüchliche
    Bedeutung interpretieren muss.
17. Als Entwickler möchte ich einen einzigen Abschluss-Command statt zweier getrennter
    Commands, damit die Reihenfolge der Events und die Invarianten an einer Stelle liegen.
18. Als Entwickler möchte ich, dass die Geldtransit-Komponente im Frontend nach ihrem
    Fachbegriff benannt ist, damit Code-Benennung und nutzersichtbarer Begriff übereinstimmen.

## Implementation Decisions

### Backend

- Ein neuer Anwendungs-Command `KasseAbschliessen` ersetzt die bisherigen Commands
  `KassensturzDurchfuehren` und `TagesabschlussErstellen`. Eingabe ist der gezählte
  Ist-Bestand in Cent. Der Command ermittelt die offene Kassensitzung, liest den
  Soll-Bestand, berechnet die Differenz, prüft die Tisch-Saldo-Sperre und schreibt in fester
  Reihenfolge die Events `kassensturz-durchgefuehrt:v1`, bei Differenz ungleich Null
  zusätzlich `differenz-soll-ist-gebucht:v1`, und abschließend `tagesabschluss-erstellt:v1`.
- Die bisherige Invariante "Kassensturz erforderlich" entfällt, weil der Zählvorgang Teil des
  Abschlusses ist und damit immer vorliegt. Die Invariante "alle Tische Saldo Null" bleibt.
- Das `tagesabschluss-erstellt:v1`-Event wird mit den echten Tagessummen befüllt
  (Gesamtumsatz, Stornierungen, Auszahlungen, Geldtransit). Die Aggregationslogik existiert
  bereits pro Kassensitzung im Reporting-Repository (Summary aus `GetReporting`); der Command
  nutzt sie als Datenquelle. Die Event-Felder bleiben unverändert bestehen, sie werden nur
  nicht länger mit Null geschrieben. Die Geldtransit-Summe stützt sich auf die bereits
  vorhandene SQL-Extraktion, die auch der Soll-Kassenbestand verwendet.
- Die Event-Reihenfolge folgt dem bestehenden sequentiellen Schreibmuster mit optimistischer
  Nebenläufigkeitskontrolle, wie es der heutige Zwei-Event-Kassensturz bereits nutzt. Schlägt
  ein Schreibvorgang nach dem ersten Event fehl, bleibt die Kassensitzung offen und der
  Abschluss kann wiederholt werden. Dieses Verhalten wird dokumentiert; ein Umbau auf eine
  einzige umschließende Datenbanktransaktion ist nicht Teil dieser PRD.
- Die TSE-Signierung der signierungspflichtigen Events (Differenzbuchung, Tagesabschluss)
  bleibt unverändert inklusive Nachsignierungs-Fallback bei TSE-Ausfall.
- HTTP: Ein neuer Endpunkt `kasse-abschliessen` ersetzt die Endpunkte
  `kassensturz-durchfuehren` und `tagesabschluss-erstellen`. Der Fehlercode
  `kassensturz_erforderlich` wird obsolet und entfällt. Der Fehlercode `tische_saldo_offen`
  bleibt. Breaking Changes an API und Events sind in der aktuellen Pre-Release-Phase laut
  Repo-Konventionen ausdrücklich erlaubt.

### Frontend

- Die Komponenten für Kassensturz und Tagesabschluss werden zu einer Abschluss-Komponente
  zusammengeführt. Sie enthält das Ist-Bestand-Eingabefeld und einen Button "Kasse
  abschließen".
- Der Klick öffnet den vorhandenen Bestätigungsdialog (das `alert-dialog`-Primitive, das
  bereits für Löschaktionen im Admin-Bereich genutzt wird). Der Dialog zeigt Soll, Ist und
  Differenz sowie die Tagessummen als Z-Bon-Vorschau und weist auf die Unwiderruflichkeit
  hin. Den Soll-Bestand kennt das Frontend bereits über den bestehenden
  Kassenbestand-Abruf; die Tagessummen stammen aus dem bereits vorhandenen
  Live-Reporting-Abruf.
- Die Backend-Klasse im Frontend ersetzt die Methoden `kassensturzDurchfuehren` und
  `tagesabschlussErstellen` durch eine Methode `kasseAbschliessen`, die den Ist-Bestand
  übergibt.
- Inline-Hilfetexte werden für alle Blöcke des Kassenbereichs ergänzt (Eröffnen, Geldtransit,
  Abschluss), im Stil der Einstellungen-Seite (gedämpfter, kleiner Erklärabsatz unter dem
  Titel). Die verbindlichen Fachbegriffe bleiben als Titel erhalten und erhalten eine
  Alltagsübersetzung als Untertitel.
- Die Einlage/Entnahme-Auswahl beim Geldtransit wird von kleinen Radio-Schaltflächen auf
  größere, fingerfreundliche Auswahlflächen umgestellt. Die Funktion bleibt gleich, nur die
  Trefferfläche wächst.
- Die Frontend-Komponente des Geldtransits wird passend zum Fachbegriff benannt. Es ist eine
  reine Umbenennung ohne Verhaltensänderung.

### Domänensprache

- Die Fachbegriffe der Ubiquitous Language (Kassensturz, Tagesabschluss, Geldtransit,
  Soll-Kassenbestand) werden nicht umbenannt. Sie werden in der Oberfläche erklärt, nicht
  ersetzt. Der nutzersichtbare Abschluss-Vorgang trägt die verständliche Bezeichnung "Kasse
  abschließen", deckt aber fachlich Kassensturz und Tagesabschluss ab.

## Testing Decisions

Ein guter Test prüft beobachtbares Verhalten an der Modulgrenze, nicht die interne
Umsetzung. Für den Abschluss bedeutet das, das Ergebnis am Event-Stream und am Status der
Kassensitzung zu prüfen, nicht einzelne interne Aufrufe.

- Schwerpunkt ist der Backend-Command `KasseAbschliessen` als das tiefe, in Isolation
  testbare Modul. Abzudeckende Fälle:
  - Abschluss bei exakt stimmender Kasse (Differenz Null): es wird ein Kassensturz-Event und
    ein Tagesabschluss-Event geschrieben, keine Differenzbuchung, die Kassensitzung erhält
    den Status abgeschlossen.
  - Abschluss bei Differenz ungleich Null: zusätzlich wird genau eine Differenzbuchung
    geschrieben.
  - Das Tagesabschluss-Event enthält die echten Tagessummen statt Nullwerten.
  - Ablehnung, wenn ein Tisch einen Saldo ungleich Null hat (Tisch-Saldo-Sperre).
  - Ablehnung, wenn keine Kassensitzung offen ist.
- Prior Art sind die bestehenden Command-Tests im Kasse-Anwendungspaket, die den bisherigen
  Kassensturz und Tagesabschluss abdecken, sowie die vorhandenen Mock-Repositories für
  Kassenjournal und Kassensitzungen. Die neuen Tests ersetzen die alten für die beiden
  zusammengelegten Commands.
- Das Frontend wird leichtgewichtig getestet (Eingabe des Ist-Bestands, korrekte Anzeige von
  Soll, Ist und Differenz im Dialog), passend zum bestehenden Vitest- und
  Testing-Library-Setup. Die Inline-Hilfetexte werden nicht eigens getestet.

## Out of Scope

- Eine eigene Anleitungs- oder Onboarding-Seite sowie Info-Popovers oder Tooltips. Die Hilfe
  bleibt bei Inline-Erklärabsätzen.
- Das Beibehalten eines separaten Kassensturz-Schritts für Zwischenzählungen. Der Kassensturz
  geht vollständig im Abschluss auf.
- Eine Umstellung des Event-Schreibens auf eine einzige umschließende Datenbanktransaktion.
- Änderungen an der Eröffnung und am Geldtransit über die ergänzten Hilfetexte, die
  fingerfreundlichere Einlage/Entnahme-Auswahl und die Komponenten-Umbenennung hinaus.
- Das Andrucken der TSE-Daten auf den Beleg, der QR-Code und der DSFinV-K-Export. Diese
  Themen haben eigene Anforderungen.
- Umbenennungen der verbindlichen Fachbegriffe der Ubiquitous Language.

## Further Notes

- Die fiskalischen Pflichten und Invarianten bleiben vollständig erhalten. Maßgeblich sind
  die Kassensturz-Dokumentation und die Z-Bon-Pflicht in `docs/compliance.md`, insbesondere
  dass Differenzen ehrlich zu buchen sind und ein Z-Bon je Veranstaltungstag erstellt wird.
- Der Bestätigungsdialog ist bewusst die Stelle, an der eine auffällige Differenz vor dem
  endgültigen Abschluss sichtbar wird. Er ersetzt aber nicht das vom Betreiber separat zu
  führende Zählprotokoll.
- Diese PRD entstand aus einem UX-Review des Kassen-Admin-Bereichs und betrifft
  ausschließlich den Admin. Die Service-Oberfläche der Helfer ist nicht betroffen.
