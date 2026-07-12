# PRD: Admin-UI-Redesign „Ein Festtag, ein roter Faden"

> Grundlage ist das Design-Handoff in
> `docs/prds/design_handoff_admin_redesign/` (README plus Design-Boards).
> Das Handoff ist die visuelle Spezifikation: Layout, Abstände, Typografie
> und Copy der „Vorschlag"-Frames sind final gemeint und nutzen ausschließlich
> die bestehenden Design-Token. Diese PRD hält den funktionalen Scope, die
> beschlossenen Abweichungen vom Handoff und die technischen Entscheidungen
> fest. Bei Aussehens-Detailfragen gilt das Handoff, bei Scope-Fragen diese
> PRD.

## Problem Statement

Der Admin-Bereich funktioniert, ist aber für die Zielgruppe (ehrenamtliche
Vereinskassiere, die jotti zwei- bis dreimal im Jahr bedienen) in Struktur
und Sprache ein Entwicklerwerkzeug. Das Design-Review aller acht Admin-Seiten
hat sechs übergreifende Befunde ergeben:

1. **System-Sprache statt Nutzer-Sprache.** „Druckstationen",
   „Kassensitzung" und die Sidebar-Gruppen „Auswertungen/Verwaltung" folgen
   der Code-Struktur, nicht dem mentalen Modell der Ehrenamtlichen
   (Vorbereitung, Festtag, nach dem Fest).
2. **Kein globaler Systemstatus.** Ob die Kasse offen ist, die TSE signiert
   und die Drucker laufen, sieht man nur auf dem Dashboard; auf jeder
   anderen Seite ist der Zustand unsichtbar.
3. **Mobile-FAB auf dem Desktop.** Die Anlegen-Buttons (Produkt, Tisch,
   Benutzer) schweben fixiert unten rechts, weit weg von Überschrift und
   Liste, und verdecken beim Scrollen Inhalte.
4. **Löschen ist so prominent wie Bearbeiten.** Jede Karte trägt einen roten
   Papierkorb, obwohl Deaktivieren fast immer die richtige Aktion wäre.
   Verschärfend: Löschen und Deaktivieren sind im Backend heute völlig
   ungeschützt; ein Produkt mit Verkäufen oder ein Tisch mit offenem Saldo
   lässt sich jederzeit löschen.
5. **Inkonsistente Seitenköpfe.** Unterschiedliche Titel-Muster, Status als
   Emoji (🟢🔴), eine Seite ganz ohne H1.
6. **Fehler wohnen am falschen Ort.** Fehlgeschlagene Bons und TSE-Rückstand
   erscheinen als Banner auf dem Dashboard; die Behebung liegt zwei Klicks
   entfernt am Seitenende einer anderen Seite.

Dazu kommen Lücken in den Tages-Workflows: Der Soll-Bestand, die zentrale
Zahl beim Kassenführen, erscheint erst im Bestätigungs-Dialog des
Abschlusses; gebuchte Einlagen/Entnahmen sind nirgends nachprüfbar; dem
Kassenbericht fehlen die formalen Z-Bon-Eckdaten (Eröffnungs-/Abschlusszeit,
abschließender User, Kassensturz-Differenz); Drucker lassen sich beim Aufbau
am Festmorgen nicht testen, sondern fallen erst im Betrieb auf.

## Solution

Der komplette Admin-Bereich wird nach dem Design-Handoff umgebaut, innerhalb
des bestehenden Systems (React 19, Tailwind CSS 4, shadcn/ui, vorhandene
Token, Lucide-Icons, keine neuen Farben/Fonts/Radii):

1. **Sidebar nach Festablauf.** Gruppen „Heute" (Übersicht, Kassentag),
   „Vorbereitung" (Produkte & Preise, Tische, Helfer & Zugänge, Bondrucker),
   „Nach dem Fest" (Berichte & Export, Finanzamt & TSE), „Service". Dazu ein
   Event-Status-Chip im Sidebar-Header (Kasse offen/geschlossen, verlinkt
   auf den Kassentag) und Statuspunkte an Menüpunkten (Kassentag grün bei
   offener Kasse, Bondrucker rot bei fehlgeschlagenen Bons, Finanzamt rot
   bei TSE-Problemen). Alle Routen bleiben unverändert.
2. **Einheitlicher Seitenkopf** auf allen Seiten: H1, erklärende Unterzeile,
   Aktions-Slot rechts (Primary-Button statt FAB).
3. **Acht Seiten-Redesigns** nach Handoff-Abschnitten 1a bis 1h: Übersicht
   (Status-Zeile mit Beheben-Button, Hero-Kennzahl, Team- und
   Offene-Tische-Spalten, kompakte aufklappbare Storno-Zeile), Berichte &
   Export (Sitzungsliste statt Dropdown, formaler Berichtskopf,
   Steuersatz-Tabelle ohne Tabs, erklärter DSFinV-K-Export-Block, Drucken),
   Produkte & Preise (Preisliste nach Kategorie, Varianten-Chips mit
   Schalter, „···"-Menü), Tische (Kachel-Grid nach Namens-Präfix,
   Saldo-Schutz), Helfer & Zugänge (Rollen-Badges, Passwort-Reset an der
   Zeile, Onboarding- und Rollen-Panels), Kassentag (3-Schritte-Stepper,
   Soll-Bestand mit Aufschlüsselung, Bewegungsliste, Zählhilfe,
   Offene-Tische-Warnung vor dem Abschluss), Bondrucker (Alarm-Karte zuerst,
   Testbon, erklärende Bonmodus-Options-Karten), Finanzamt & TSE
   (Einrichtungs-Checkliste inkl. ELSTER-Meldung, „Läuft alles?"-Ampel,
   technische Details in Collapsibles).
4. **Deaktivieren vor Löschen, mit echten Schutzregeln.** Lösch-Einstiege
   wandern in „···"-Menüs hinter den bestehenden AlertDialog. Neu und
   verbindlich im Backend: Produkte mit Verkäufen sind nur deaktivierbar,
   Tische mit offenem Saldo weder deaktivier- noch löschbar.
5. **Gezielte Backend-Erweiterungen** (alle POST, additiv):
   Kassenbestand-Aufschlüsselung, Geldtransit-Liste der offenen Sitzung,
   Kassensitzungs-Metadaten für Sitzungsliste und Berichtskopf,
   `elsterGemeldetAm`-Flag am Betreiber, Testdruck-Endpoint, Schutzregeln
   samt neuen Listen-Feldern (`hatVerkaeufe`, `saldoCents`).

**Beschlossene Vereinfachungen gegenüber dem Handoff:**

- **Keine Serienanlage für Tische** (Handoff 1d): bewusst abgewählt, die
  Tische-Seite behält Einzel-Anlage über den bestehenden Dialog.
- **Kein Druckstations-Status** („verbunden · letzter Druck", Handoff 1g):
  wäre nur eine „zuletzt"-Aussage und bräuchte neue Persistenz. Die
  Stationskarten kommen ohne Status-Zeile.
- **Kein Bon-Klartext** an fehlgeschlagenen Druckaufträgen (Handoff 1g):
  die Alarm-Karte nutzt die bestehende Referenz-Darstellung
  (`formatDruckauftragReferenz`) plus Fehlertext.
- **Auto-Refresh bleibt bei 30 s** (Handoff sagt 60 s): das Dashboard pollt
  heute schon alle 30 s; übernommen wird die sichtbare
  „Live · aktualisiert HH:MM"-Anzeige und der Jetzt-Button.

## User Stories

Akteur ist durchgehend der Admin (Vereinskassier); Helfer, Kassenprüfer und
Steuerberater kommen als Nutznießer vor.

1. Als Admin möchte ich eine Sidebar, die dem Ablauf eines Vereinsfests
   folgt (Heute / Vorbereitung / Nach dem Fest), damit ich Funktionen dort
   finde, wo ich sie im Festablauf erwarte.
2. Als Admin möchte ich auf jeder Admin-Seite am Sidebar-Chip und an
   Statuspunkten sehen, ob die Kasse offen ist und ob Drucker oder TSE
   Probleme haben, damit ich Störungen bemerke, ohne das Dashboard zu
   öffnen.
3. Als Admin möchte ich auf jeder Seite denselben Kopf (Titel, erklärende
   Unterzeile, Hauptaktion oben rechts statt schwebendem Button), damit ich
   Aktionen immer am selben Ort finde.
4. Als Admin möchte ich auf der Übersicht eine Status-Zeile (Kasse, TSE,
   Drucker) mit einem „Beheben"-Button direkt am Fehler, damit ich vom
   Problem ohne Suchen zur Behebung komme.
5. Als Admin möchte ich den kassierten Umsatz als hervorgehobene
   Hero-Kennzahl mit erklärenden Unterzeilen sehen, damit ich kassiert und
   bestellt nicht mehr verwechsle.
6. Als Admin möchte ich offene Tische und den Abrechnungsstand meines Teams
   nebeneinander sehen und Stornos als kompakte, aufklappbare Zeile, damit
   der Schichtende-Check auf einen Blick funktioniert.
7. Als Admin möchte ich, dass sich die Übersicht sichtbar selbst
   aktualisiert („Live · aktualisiert HH:MM" plus Jetzt-Button), damit ich
   weiß, wie frisch die Zahlen sind.
8. Als Admin möchte ich abgeschlossene Kassensitzungen als Liste mit Datum,
   Bezeichnung und Umsatz sehen, damit ich den richtigen Tagesbericht ohne
   Dropdown-Raten finde.
9. Als Admin möchte ich am Tagesbericht die formalen Z-Bon-Eckdaten sehen
   (Nr., Eröffnungs- und Abschlusszeit, wer abgeschlossen hat,
   Kassensturz-Differenz), damit der Bericht gegenüber dem Kassenprüfer
   vollständig ist.
10. Als Admin möchte ich Steuersatz-Tabelle, Servicekraft-Umsätze und
    Stornierungen untereinander statt hinter Tabs sehen und den Bericht
    drucken können, damit ich ihn mit dem Kassenprüfer durchgehen kann.
11. Als Admin möchte ich einen erklärten Export-Block „Für Steuerberater &
    Finanzamt", damit ich weiß, was das DSFinV-K-Archiv ist und wem ich es
    gebe.
12. Als Admin möchte ich Produkte als Preisliste gruppiert nach Kategorie
    sehen, mit Preis und Aktiv-Schalter jeder Variante direkt in der Zeile,
    damit ich Preise und Verfügbarkeit ohne Aufklappen prüfen und schalten
    kann.
13. Als Admin möchte ich, dass ausverkaufte Varianten per Schalter
    deaktiviert werden und Löschen nur für Produkte ohne Verkäufe möglich
    ist (im „···"-Menü, vom Backend erzwungen), damit Berichte vollständig
    bleiben und niemand im Stress Daten vernichtet.
14. Als Admin möchte ich Tische als kompakte Kacheln, gruppiert nach
    Namens-Präfix, damit auch 40 Tische übersichtlich bleiben.
15. Als Admin möchte ich, dass Tische mit offenem Saldo den Betrag anzeigen
    und gegen Deaktivieren und Löschen geschützt sind (vom Backend
    erzwungen), damit kein Tisch mit offenen Posten verschwindet.
16. Als Admin möchte ich Rollen als beschriftete Badges mit einer
    Rechte-Erklärung sehen, damit ich Storno-Rechte nicht aus
    Stern-Symbolen erraten muss.
17. Als Admin möchte ich den Passwort-Reset direkt an der Benutzerzeile,
    damit die häufigste Support-Anfrage am Festtag zwei Klicks kostet.
18. Als Admin möchte ich neben der Benutzerliste eine Schritt-für-Schritt-
    Anleitung für das Helfer-Onboarding (Einmalpasswort-Verfahren), damit
    ich vor dem Fest zehn Helfer ohne Handbuch einrichte.
19. Als Admin möchte ich den Kassentag als drei Schritte sehen (eröffnet,
    laufender Betrieb, abschließen), damit der wichtigste Workflow des
    Tages selbsterklärend ist.
20. Als Admin möchte ich den Soll-Bestand samt Aufschlüsselung
    (Anfangsbestand, Bareinnahmen, Einlagen, Entnahmen) jederzeit sehen,
    damit ich die Kasse laufend im Griff habe statt erst beim Abschluss.
21. Als Admin möchte ich die heutigen Kassenbewegungen als Liste sehen und
    Einlagen/Entnahmen über Buttons mit vorbelegter Richtung buchen, damit
    ich nachprüfen kann, ob die 200 € Wechselgeld schon gebucht sind.
22. Als Admin möchte ich eine Zählhilfe (Stückzahl je Münze/Schein, Summe
    wird übernommen), damit ich beim Kassensturz nicht im Kopf rechnen
    muss.
23. Als Admin möchte ich vor dem endgültigen Abschluss eine Live-Rechnung
    (Soll, Gezählt, Differenz) und eine deutliche Warnung bei offenen
    Tischen, damit ich den unwiderruflichen Schritt informiert gehe.
24. Als Admin möchte ich fehlgeschlagene Bons als Alarm-Karte ganz oben auf
    der Bondrucker-Seite mit „Nochmal drucken" und „Verwerfen" direkt daran,
    damit die Küche fehlende Bons schnell nachbekommt.
25. Als Admin möchte ich je Station einen Testbon drucken können, damit ich
    beim Aufbau am Festmorgen prüfe, ob jeder Drucker erreichbar ist.
26. Als Admin möchte ich den Bonmodus als erklärende Options-Karten („Pro
    Position: je Gericht ein Abreiß-Bon" / „Pro Bestellung: ein Sammelbon"),
    damit ich den Küchen-Workflow-Schalter verstehe.
27. Als Admin möchte ich auf der Finanzamt-Seite eine Checkliste
    „Einrichtung: x von 3 Schritten" (Vereinsdaten, TSE, Kassenmeldung),
    damit ich sehe, was noch fehlt und bis wann.
28. Als Admin möchte ich die ELSTER-Kassenmeldung nach Erledigung abhaken
    können („Gemeldet am {Datum}", persistiert), damit die rote Warnung
    verschwindet, sobald ich gemeldet habe.
29. Als Admin möchte ich den TSE-Zustand als Klartext-Ampel („Ja, TSE
    signiert normal") mit technischen Details und Störungsprotokoll in
    Collapsibles, damit ich den Normalfall auf einen Blick erkenne und
    Details nur bei Bedarf sehe.

## Implementation Decisions

**Vorgehen und Struktur**

- Erst drei bis vier kleine gemeinsame Layout-Bausteine, dann die Seiten in
  Handoff-Reihenfolge darauf umstellen: einheitlicher Seitenkopf (H1 +
  Unterzeile + Aktions-Slot), Hinweis-Karte (Info-Variante) und Warn-Karte
  (Destructive-Variante), Statuspunkt. Kein darüber hinausgehendes
  Komponenten-Framework.
- Routen, Guards und die Handler-Struktur des Backends bleiben unverändert;
  es ändern sich UI-Labels („Kassentag", „Helfer & Zugänge", „Bondrucker",
  „Berichte & Export", „Finanzamt & TSE"), nicht die Fachbegriffe in Code,
  API und Doku (die Ubiquitous Language nach `docs/language.md` gilt
  unverändert, z. B. bleibt die Entität Kassensitzung).
- Bestehende Hooks, Backend-Klassen, Dialoge und Bestätigungs-Flows werden
  weiterverwendet (TanStack-Query-Hooks, `use-action-submit`, AlertDialog,
  Toasts). Der FAB-Pattern und die zugehörige Scroll-Clearance entfallen
  ersatzlos; Anlegen-Aktionen wandern in den Seitenkopf.
- Auto-Refresh der Übersicht bleibt beim bestehenden 30-s-Intervall;
  ergänzt werden sichtbarer Stand („aktualisiert HH:MM" aus `dataUpdatedAt`)
  und manueller Refetch-Button.
- Sidebar-Status speist sich aus den vorhandenen Queries (offene
  Kassensitzung, fehlgeschlagene Druckaufträge, TSE-Status/-Queue); keine
  zusätzlichen Polling-Anforderungen über die bestehenden Intervalle hinaus.
- Dark Mode ausschließlich über Token (Alpha-Varianten via Tailwind-Syntax,
  z. B. `destructive/40`); keine Sonderbehandlung.
- Keine neuen Dependencies; Icons ausschließlich Lucide.

**Backend-Erweiterungen (alle additiv, POST-only)**

- **Kassenbestand-Aufschlüsselung:** Die bestehende Kassenbestand-Query
  liefert zusätzlich zum Aggregat die Komponenten Anfangsbestand,
  Bareinnahmen (Zahlungen + Direktverkauf − geldwirksame Stornos), Einlagen
  und Entnahmen. Die vorhandenen SQL-Extractor-Funktionen werden dafür
  einzeln ausgewertet; die Summe der Komponenten muss dem bestehenden
  Soll-Bestand entsprechen.
- **Geldtransit-Liste:** Neue Query über die `geldtransit-gebucht`-Events
  der offenen Kassensitzung; Antwort je Buchung: Zeitpunkt, Richtung
  (Einlage/Entnahme), Betrag in Cent, Kommentar, buchender Benutzer
  (Anzeigename). Reine Projektion aus dem Kassenjournal, keine neuen
  Event-Formate.
- **Kassensitzungs-Metadaten:** Die Liste abgeschlossener Kassensitzungen
  liefert zusätzlich Gesamtumsatz und Abschluss-Zeitpunkt; die
  Bericht-Query liefert zusätzlich Eröffnungs-Zeitpunkt,
  Abschluss-Zeitpunkt, abschließenden Benutzer und Kassensturz-Differenz.
  Quelle sind die vorhandenen Journal-Events (Eröffnung, Kassensturz,
  Tagesabschluss); keine Schema-Änderung.
- **ELSTER-Flag:** Neues nullbares Datum `elsterGemeldetAm` an den
  Betreiber-Stammdaten (additive Migration nach Freeze-Disziplin). Eigener
  Befehl zum Setzen und Zurücknehmen (Fehlklick korrigierbar), Feld wird in
  der Betreiber-Query mitgeliefert. Es wird nur der manuelle Vollzug
  dokumentiert; eine automatisierte Meldung bleibt Nicht-Ziel.
- **Testdruck:** Neuer Befehl „Testbon an Station drucken": erzeugt einen
  einfachen Testbon (Stationsname, Zeitstempel) als regulären Druckauftrag
  über die bestehende Outbox/Relay-Mechanik. Schlägt der Testdruck fehl,
  erscheint er wie jeder Auftrag in den fehlgeschlagenen Druckaufträgen
  (dort sichtbar in der Alarm-Karte); es gibt keinen eigenen
  Status-Rückkanal.
- **Schutzregeln (Backend als Single Source of Truth):**
  - Produkt löschen ist nur erlaubt, wenn das Produkt in keiner
    Kassensitzung je verkauft wurde (Prüfung gegen das Kassenjournal);
    sonst eigener Fehlercode. Die Produktliste liefert je Produkt ein Flag
    `hatVerkaeufe`, damit das Frontend die Aktion proaktiv deaktivieren und
    den Tooltip zeigen kann.
  - Tisch deaktivieren oder löschen ist nur erlaubt, wenn der Tisch keinen
    offenen Saldo hat (Prüfung gegen die Tisch-Sessions-Projektion); sonst
    eigener Fehlercode. Die Tischliste liefert je Tisch `saldoCents`, damit
    das Frontend Kachel-Anzeige und Schutz ohne Join mit dem Live-Reporting
    darstellen kann.
  - Der bestehende Selbstlösch-Schutz für Benutzer bleibt; das Frontend
    blendet Löschen am eigenen Account aus („das bist du"-Badge).
- Fehlercodes der neuen Guards werden wie üblich über die
  `errorMessages`-Map in deutsche Klartext-Meldungen übersetzt.

**Frontend-Einzelentscheidungen**

- Zählhilfe-Dialog als eigenständige, rein clientseitige Komponente:
  Stückzahlen je Nennwert (1 ct bis 200 €), Summenberechnung, Übernahme in
  das Ist-Bestand-Feld. Kern als reine Funktion (Stückzahlen → Cent-Summe)
  isoliert testbar.
- Druck des Tagesberichts über `window.print` mit Print-Stylesheet, das nur
  die Berichtsspalte druckt.
- Storno-Details der Übersicht werden zur eingeklappten Zeile; das
  Aufklappen zeigt die bestehende Storno-Detail-Darstellung inline.
- Der Kassenabschluss behält den bestehenden Bestätigungs-AlertDialog als
  zweite Stufe samt Retry-Logik bei ausstehenden Signaturen; neu davor:
  Live-Differenz-Rechnung und Offene-Tische-Warnung.
- Leerzustände bleiben beim bestehenden Empty-Muster (z. B. Übersicht ohne
  Kassensitzung mit Link zum Kassentag; Kassentag ohne Sitzung zeigt
  Schritt 1 als aktives Eröffnen-Formular).
- Drucker-IP speichert on-blur oder per Enter mit Erfolgs-Toast; der
  Speichern-Button je Feld entfällt.

## Testing Decisions

- Getestet wird ausschließlich äußeres Verhalten (sichtbare Texte, Aktionen,
  Navigation, aufgerufene Backend-Methoden), keine Implementierungsdetails
  wie CSS-Klassen oder Komponentenstruktur.
- **Frontend:** Jede umgebaute Seite bekommt bzw. behält einen
  Seitentest nach dem bestehenden Muster (Testing Library mit gemockten
  Backend-Klassen, wie die vorhandenen Tests für Dashboard, Kassensitzung,
  Druckstationen und Sidebar). Für Produkte, Tische, Benutzer und Finanzamt
  existieren heute keine Seitentests; sie werden im Zuge des Umbaus
  ergänzt. Schwerpunkte: Schutzregeln (deaktivierte Aktionen samt Tooltip),
  Stepper-Zustände des Kassentags, Alarm-Karte und Testbon-Aktion,
  Checklisten-Zustände der Finanzamt-Seite, Sidebar-Gruppen und
  Statuspunkte.
- Die Zählhilfe wird als isolierte Komponente getestet (Eingaben →
  Summe → Übernahme), die Differenz-Rechnung über den Seitentest des
  Kassentags.
- **Backend:** Neue Queries und Befehle bekommen Handler- und
  Integrationstests nach den bestehenden Mustern des jeweiligen Kontexts;
  die Projektions-Queries (Geldtransit-Liste, Sitzungs-Metadaten,
  Kassenbestand-Komponenten) werden gegen ein per Events aufgebautes
  Journal getestet. Die Guards werden je Fehlerfall getestet (Produkt mit
  Verkäufen, Tisch mit Saldo) inklusive Fehlercode; die
  Summen-Invariante der Kassenbestand-Komponenten wird explizit geprüft.
- Der Event-JSON-Contract-Guard bleibt unberührt; es entstehen keine neuen
  Event-Typen und keine Änderungen an bestehenden Event-Formaten.

## Out of Scope

- **Serienanlage für Tische** (Handoff 1d): bewusst abgewählt; kann bei
  belegtem Praxis-Bedarf als eigene kleine PRD nachgezogen werden.
- **Druckstations-Status** („verbunden / letzter Druck HH:MM") und
  **Bon-Klartext-Positionen** an fehlgeschlagenen Druckaufträgen
  (Handoff 1g): entfallen; Referenz-Darstellung bleibt.
- Änderungen an Routen, URLs, Auth, Guards oder am Service-Bereich.
- Redesign der TSE-Einrichtungsseite und des Wizards (die Finanzamt-Seite
  verlinkt weiterhin dorthin; die manuelle Experten-Konfiguration bleibt
  dort).
- Automatisierte ELSTER-Meldung (dauerhaftes Nicht-Ziel laut
  `docs/anforderungen.md`); das Flag dokumentiert nur den manuellen
  Vollzug.
- Neue Farben, Fonts, Radii, Icons außerhalb von Lucide oder neue
  Dependencies.
- Mobile-spezifisches Redesign: Zielgerät ist Desktop/Laptop; das
  bestehende responsive Verhalten (Sidebar offcanvas, Mobile-Header,
  Grids brechen unter `lg` um) und die 44-px-Touchziele bleiben erhalten.
- Änderungen an Event-Formaten, am Kassenjournal oder an bestehenden
  DB-Tabellen über die eine additive Betreiber-Migration hinaus.

## Further Notes

- Die Beispieldaten in den Design-Frames (Beträge, Namen wie „Sophie Renz")
  sind Dummy-Daten; alle Werte kommen aus den bestehenden bzw. hier
  beschlossenen APIs.
- Die Umsetzung folgt der im Handoff empfohlenen Reihenfolge (Sidebar →
  Seitenkopf-Bausteine → Übersicht → Stammdaten-Seiten → Berichte →
  Kassentag → Bondrucker → Finanzamt); jeder Schritt ist unabhängig
  shipbar. Für die Umsetzung sind mehrere Implementierungspläne zu
  erwarten; diese PRD ist die gemeinsame Klammer.
- Überschneidung mit `docs/plans/offene-punkte.md` (Audit-Rest „solide
  Destructive-Buttons im Admin"): Durch das Redesign wandern die
  destruktiven Aktionen in „···"-Menüs und AlertDialoge; der Punkt sollte
  nach der Umsetzung gegengeprüft und abgehakt werden.
- Die Sitzungsliste der Berichte-Seite zeigt die offene Sitzung als nicht
  wählbaren Eintrag, der zur Übersicht führt; Status-Emojis entfallen
  ersatzlos.
- Der Handoff-Ordner bleibt neben dieser PRD liegen, bis das Redesign
  vollständig umgesetzt ist; danach kann er wie erledigte Pläne entfernt
  werden (Git-Historie bewahrt ihn).
