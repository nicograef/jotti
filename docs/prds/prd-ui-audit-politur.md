# PRD: UI-Audit-Politur und -Härtung

> Grundlage: Mehr-Experten-UI-Audit vom 13.07.2026 (26 Befunde, Artifact-Report,
> Code-Root-Cause je Befund). Quelle und Detailtiefe: Memory `ui-audit-2026-07-13`
> und der Audit-Artifact. Datei-Zeilen-Referenzen wurden vor diesem PRD gegen den
> aktuellen Quellcode verifiziert (nur geringfügiger Zeilen-Drift).

## Problem Statement

Nach dem letzten Redesign hat die Oberfläche Bedien- und Darstellungsfehler, die
ausgerechnet die schwächste Nutzergruppe treffen: nicht-technische Helfer, die
unter Stress am eigenen Smartphone arbeiten.

- Beim Bestellen, Stornieren und Umbuchen liegen Gesamtsumme und Pflichtfeld
  (Kommentar, Ziel-Tisch) im scrollenden Drawer-Body, nur die Buttons sind
  sticky. Bei langen Bestellungen wandern gerade die wichtigsten, teils rechtlich
  verpflichtenden Steuerelemente unter den Fold, und die Primäraktion ist stumm
  deaktiviert ohne sichtbaren Grund. Für einen Helfer wirkt das wie ein
  Software-Fehler in einer Sackgasse.
- Historie, Kassieren, Direktverkauf-Historie und die Tischauswahl laufen auf
  schmalen Telefonen (390px) horizontal über, Beträge werden abgeschnitten.
- Die Finanzamt-Einrichtung schneidet im dritten Schritt die ELSTER-UUID und die
  Pflicht-Aktion „Als erledigt markieren“ ab.
- Der Dark Mode fällt genau auf Recovery- und Compliance-Screens (Bondrucker,
  Finanzamt) unter den WCAG-Kontrast.
- Microcopy driftet dort, wo Präzision compliance- und vertrauensrelevant ist:
  ein fehlgeschlagener Kassenbeleg (Gäste-Beleg) wird als Küchenproblem
  beschrieben, und ein kaputter Plural zeigt „Alle 1 Positionen“.

Unter all dem liegt eine gemeinsame Ursache: kein einheitliches Design-System.
Geld-Eingaben, Disabled-Zustände, Dialog-Footer, Badges, Karten-Füllungen und
Bestätigungsfarben driften pro Screen auseinander.

## Solution

Eine fokussierte Politur- und Härtungsrunde ohne neue Features. Die 26 Befunde
werden auf ihre sechs systemischen Ursachen gruppiert; wird die Ursache einmal
behoben, verschwinden mehrere Symptome zugleich. Sequenziert nach Hebel und
Risiko (P0 bis P2): zuerst der Blocker-Bereich (Drawer-Sticky-Footer und der
mobile Raster-Überlauf), dann die Kern-Härtung (Disabled-Affordanz, Preis-Spalte,
Compliance-Microcopy, Dark-Mode-Kontrast, begrenzte Fehl-Bon-Liste), zuletzt ein
breiter Design-System-Token-Pass, der die driftenden Primitive vereinheitlicht.

Produkt-Konservatismus als erste Linse: Jeder Befund ist ein Korrektheits- oder
Konsistenzfix, keiner ist ein neues Feature oder Gold-Plating. Deshalb bleiben
alle 26 Befunde als FIX in scope; nichts wird verschoben oder verworfen. Die
Scope-Grenze ist bewusst hart: vereinheitlicht werden nur die vom Audit benannten
Primitive. Keine neue Designsprache, keine Konfigurierbarkeit, kein Eingriff in
Backend oder Daten.

## User Stories

Akteure: Servicekraft (BYOD-Smartphone), Serviceleitung, Admin, Betreiber
(Compliance-Sicht).

1. Als Servicekraft möchte ich beim Bestellen, Stornieren und Umbuchen die
   Gesamtsumme, das Pflichtfeld und die Aktion immer sichtbar haben, ohne
   scrollen zu müssen, damit ich lange Vorgänge ohne Sackgasse abschließe.
2. Als Servicekraft möchte ich bei deaktivierter Aktion den Grund sehen oder zum
   fehlenden Feld geführt werden, damit ich nicht an einem toten Button hänge.
3. Als Servicekraft möchte ich auf meinem Smartphone keine horizontal
   überlaufenden Listen, damit Beträge und Labels vollständig lesbar bleiben.
4. Als Servicekraft möchte ich Preise in einer festen Spalte, damit ich sie beim
   schnellen Bestellen zuverlässig erfasse, unabhängig von der Namenslänge.
5. Als Servicekraft möchte ich korrekte Formulierungen („1 Position“ statt „Alle 1
   Positionen“), damit die Oberfläche vertrauenswürdig wirkt.
6. Als Admin möchte ich die Finanzamt-Einrichtung vollständig lesen und bedienen
   (ELSTER-UUID abtippbar, „Als erledigt markieren“ klickbar), damit ich die
   Kassenmeldung korrekt durchführe.
7. Als Admin möchte ich die Liste fehlgeschlagener Bons begrenzt und scrollbar,
   damit die Stationskonfiguration erreichbar bleibt.
8. Als Betreiber möchte ich, dass ein fehlgeschlagener Kassenbeleg nicht als
   Küchenproblem beschrieben wird, damit compliance-relevante Meldungen fachlich
   korrekt sind.
9. Als Nutzer im Dark Mode (BYOD im Freien) möchte ich ausreichende Kontraste auf
   Recovery- und Compliance-Screens, damit ich alles sicher lesen kann.
10. Als Admin möchte ich konsistente Geld-Eingaben, Dialog-Footer, Badges, Karten
    und Bestätigungsfarben, damit die Oberfläche ruhig und vorhersehbar wirkt.
11. Als Admin möchte ich beim endgültigen Kassenabschluss eine Bestätigung, deren
    Farbe „irreversibel“ signalisiert, ohne fälschlich „gefährlich/löschen“ zu
    meinen.
12. Als Serviceleitung möchte ich denselben Sticky-Footer-Komfort in der
    Stornierung wie beim Kassieren, damit die Pflicht-Begründung nie unter dem
    Fold verschwindet.

## Implementation Decisions

### Aufgelöste Produktentscheidungen (Decision Gate)

- #4 Tisch-Löschen: Ein-Zeilen-Footer (Ghost-Löschen links, Abbrechen/Speichern
  rechts), wie der EditUserDialog. Ersetzt die getrennte Vollbreiten-Zeile.
- #5 Passwort-Reset: nur der Einstieg im Zeilen-„…“-Menü bleibt; der Einstieg im
  Bearbeiten-Dialog entfällt. Das deckt sich mit der Hilfe-Karte.
- NEU07 Bestätigungsfarbe: eigenes Warn-Treatment für „irreversibel“, weder
  destruktiv-rot noch Standard-Grün. Als ADR festhalten (siehe Further Notes).
- Muster 03 Design-System: breiter Token-Pass als eigener Workstream über
  Geld-Input, Disabled-Zustand, Dialog-Footer, Badges, Karten-Füllung und
  Bestätigungsfarbe.
- #8 Preis-Spalte: gemeinsames Name/Preis/Stepper-Layout jetzt über Admin und
  Service, nicht später.
- Polish-Rest: alle vier Bündel in scope (Terminologie und Fehlertexte,
  Form-Feinschliff, Interaktion NEU13, Theme-Toggle NEU14).

### Befund-Übersicht mit Zuordnung und Entscheidung

Entscheidung ist durchgängig FIX (Begründung: Solution). Prio ist Sequenzierung,
keine Wichtigkeit.

| ID | Schwere | Ursache/Workstream | Prio |
| --- | --- | --- | --- |
| #1 | Blocker | M05 Finanzamt-Layout + Raster | P0 |
| #9 | Blocker | M01 Drawer-Sticky-Footer | P0 |
| #10 | Blocker | M01 Drawer-Sticky-Footer + M02 Affordanz | P0 |
| #11 | Major | M01 Drawer-Sticky-Footer + M02 Affordanz | P0 |
| #12 | Major | M04 Raster-Basisspalte | P0 |
| NEU04 | Major | M01 (Pflichtfeld sichtbar) + M02 (Grund sichtbar) | P0/P1 |
| #6 | Major | Fehl-Bon-Liste begrenzen | P1 |
| #8 | Major | M04 Preis-Spalte | P1 |
| NEU01 | Major | M05 Dark-Mode-Kontrast | P1 |
| NEU02 | Major | M06 Microcopy (Kassenbeleg ≠ Küche) | P1 |
| NEU03 | Major | M06 Microcopy (Plural) | P1 |
| NEU05 | Major | M05 Dark-Mode-Kontrast | P1 |
| NEU06 | Major | M03 Disabled-Token + M05 | P1 |
| #3 | Major | M03 Geld-Eingabe | P2 |
| #4 | Major | M03 Dialog-Footer | P2 |
| #5 | Major | M03 Dialog-Konsistenz | P2 |
| #7 | Minor | M06 Terminologie | P2 |
| NEU07 | Minor | M03 Warn-Bestätigung + Kassenabschluss-Ausrichtung | P2 |
| NEU08 | Minor | M03 Badges | P2 |
| NEU10 | Minor | M03 Karten-Füllung | P2 |
| NEU11 | Minor | M06 Drucker-Fehlertexte | P2 |
| NEU09 | Minor | Form-Feinschliff (Stepper) | P2 |
| #2 | Polish | M03 Button-Schatten | P2 |
| NEU12 | Polish | Form-Feinschliff (Zählhilfe-Spinner) | P2 |
| NEU13 | Polish | Interaktion (Tab-Leiste hinter Drawer) | P2 |
| NEU14 | Polish | Theme-Toggle | P2 |

### Module

Der Fokus liegt auf wenigen tiefen, isoliert testbaren Modulen, die viele
Call-Sites bedienen.

- Drawer-Layout-Vertrag (bestehend, ADR 03): `DrawerBody` ist der einzige
  Scrollbereich, `DrawerFooter` bleibt immer sichtbar. Der Fix wendet dieses
  bereits existierende Muster auf Bestellung, Stornierung und Umbuchung an;
  Gesamtsumme, Pflichtfeld und Primäraktion wandern in den Sticky-Bereich. Der
  Kassieren-Tab ist die Referenz, die es schon richtig macht. Kein neues Modul,
  sondern konsistente Anwendung.
- Disabled-Affordanz: ein gemeinsames Muster, das bei deaktivierter Primäraktion
  den Grund als Live-Hinweis anzeigt und optional zum fehlenden Feld führt. Kleine
  testbare Schnittstelle: Validierungszustand rein, Hinweistext raus.
- Geld-Eingabe: eine Komponente mit nahtloser €-Affordanz (kein separater
  umrandeter Präfix-Kasten), Wert in Cent, ersetzt die drei heutigen Stile. Tiefes
  Modul: versteckt Formatierung und Cent-Handling hinter `value`/`onChange`.
- Preis-Zeile: gemeinsames Name/Preis/Stepper-Layout mit fester Preis-Spalte,
  wiederverwendet in Admin (Produkt-Chips) und Service (Bestellen).
- Design-Token-Ebene: benannte CSS-Variablen für Disabled-Zustand,
  Warn-Bestätigung, Badge und Karten-Füllung, jeweils in Light und Dark und beide
  WCAG AA. Kein neues Widget, nur die vorhandenen Primitive an eine gemeinsame
  Token-Quelle gebunden.
- Raster-Basisspalte: mechanischer Sweep, kein Modul. Wo ein Grid Spalten nur an
  Breakpoints definiert (`lg:`/`sm:`/`2xl:` ohne Basis-Track), fehlt der mobile
  Einzelspalten-Track und die einzelne implizite `auto`-Spur wächst auf
  max-content. Fix: Basis `grid-cols-1` ergänzen. Raster mit bewusstem
  2-Spalten-Mobile (`grid-cols-2`) sind gebunden und bleiben unberührt.

### Akzeptanzkriterien je Workstream

Muster 01 (P0) Drawer-Sticky-Footer (#9, #10, #11, NEU04-Teil):

- Bei einer langen Bestellung/Stornierung/Umbuchung (Body scrollt) sind
  Gesamtsumme, Pflichtfeld (Kommentar bzw. Ziel-Tisch) und die Primäraktion ohne
  Scrollen sichtbar.
- Der Layout-Vertrag aus ADR 03 bleibt gewahrt (`max-h-[85dvh]`, Safe-Area,
  `DrawerBody` einziger Scrollbereich). Verifikation auf einem echten iOS-Gerät
  als installierte PWA, weil Desktop und Browser-Tab den Fehler nicht zeigen.

Muster 02 (P1) Disabled-Affordanz (NEU04, #10, #11):

- Ist die Primäraktion deaktiviert, nennt die Oberfläche den Grund sichtbar (z. B.
  „Kommentar erforderlich“, „Ziel-Tisch wählen“) und/oder führt zum fehlenden
  Feld. Kein stummer, grundloser toter Button mehr in den Service-Drawern.

Muster 04 (P0/P1) Wert-Spalte und Raster-Disziplin:

- P0 Raster-Sweep: kein horizontaler Seiten-Überlauf (`scrollWidth` ≤ Viewport)
  bei 390px auf Historie, Kassieren, Direktverkauf-Historie, Tischauswahl,
  Finanzamt-Einrichtung und den übrigen im Sweep gefundenen Rastern. Betroffen ist
  der volle im Quellcode verifizierte Satz (rund 15 Raster ohne Basis-Track über
  Admin und Service), nicht nur die sieben ursprünglich benannten Stellen.
- P1 Preis-Spalte (#8): der Variantenpreis steht an einer festen Spaltenposition,
  unabhängig von der Namenslänge; umbrechende Namen lösen den Preis nicht ab.
  Dasselbe Layout in Admin und Service.

Muster 05 (P1) Dark-Mode-Kontrast (#1-Anteil, NEU01, NEU05, NEU06-Anteil):

- Grüne Aktionen („Nochmal drucken“), Outline-Buttons, Drucker-IP-Felder und
  deaktivierte Primäraktionen erreichen im Dark Mode WCAG AA (4,5:1 für Text, 3:1
  für UI-Ränder). Ein automatisierter Kontrast-Check bestätigt die Recovery- und
  Compliance-Screens in Light und Dark.

Muster 06 (P1/P2) Microcopy und Terminologie (NEU02, NEU03, #7, NEU11):

- NEU02: ein fehlgeschlagener Kassenbeleg (Gäste-Beleg) wird nicht als
  Küchenproblem beschrieben. Die generische Fehl-Bon-Warnung darf nicht
  pauschal „die Küche hat sie nicht“ behaupten; der Text richtet sich nach der
  tatsächlichen Bon-Art (Arbeitsbon vs. Kassenbeleg vs. Testbon).
- NEU03: Singular „1 Position“, Plural „N Positionen“. Die zwei betroffenen
  Produktionsstrings (Kassieren, Umbuchung) werden korrigiert; die Tests, die
  aktuell den kaputten String „Alle 1 Positionen“ festschreiben, werden
  mitgezogen.
- #7: Begriffe konsistent zu `docs/language.md`. „Kassenbeleg“ wird verwendet, wo
  es der gesetzliche Beleg ist; „Bon“ bleibt dem operativen Arbeitsbon
  vorbehalten. Das helfer-freundliche Sidebar-Label „Bondrucker“ darf bleiben; die
  URL `/druckstationen` ist nicht nutzersichtbar und wird nicht angefasst.
- NEU11: einheitliche Drucker-Fehlertexte (kein Mix aus Klartext und
  Techniker-Jargon).

Muster 03 (P2) Design-System-Token-Pass (#3, #4, #5, NEU06-Token, NEU07, NEU08,
NEU10, #2):

- #3: eine Geld-Eingabe-Komponente, nahtlose €-Affordanz, überall identisch
  (Neue Variante, Variante bearbeiten, Kasse abschließen, Zählhilfe).
- #4: Tisch-Bearbeiten-Dialog nutzt den Ein-Zeilen-Footer wie EditUserDialog.
- #5: nur ein Passwort-Reset-Einstieg (Zeilen-Menü).
- NEU06: ein einziges Disabled-Token für Primäraktionen, nicht Mint gegen Rosa.
- NEU07: eigenes Warn-Treatment-Token für irreversible Routine-Bestätigungen; die
  Bestätigung „Kasse endgültig abschließen“ nutzt es. Zusätzlich Ausrichtung von
  €-Eingabe, Soll/Gezählt-Differenz und CTA im Kassenabschluss.
- NEU08: Rollen-Badges in einem einheitlichen Pill-Stil, nicht drei Stile in einer
  Zeile.
- NEU10: die „Kassierter Umsatz“-Karte wirkt nicht grau/deaktiviert zwischen den
  weißen Karten; einheitliche Karten-Füllung.
- #2: `shadow-xs` konsistent über die Button-Varianten (einheitlich mit oder ohne,
  nicht nur auf `outline`).
- Grenze: nur diese Primitive. Keine neue Designsprache, keine zusätzlichen
  Varianten.

Form-Feinschliff und Interaktion (P2) (NEU09, NEU12, NEU13, NEU14):

- NEU09: der Minus-Stepper bei Menge 0 ist eindeutig deaktiviert, kein
  geisterhafter gestrichelter Kreis, der antippbar wirkt.
- NEU12: die Zählhilfe zeigt keine nativen Browser-Number-Spinner mehr, sondern
  nutzt das Formular-System wie überall.
- NEU13: die Tab-Leiste (Bestellen | Kassieren | Historie) ist bei offenem Drawer
  nicht interaktiv und nicht fokussierbar.
- NEU14: der Sidebar-Theme-Umschalter hat kein wechselndes Label und bekommt im
  Dark Mode kein Active-Page-Highlight.

Fehl-Bon-Liste begrenzen (P1) (#6):

- Die Liste fehlgeschlagener Bons ist in der Höhe begrenzt und scrollbar, sodass
  die Stationskonfiguration darunter immer erreichbar bleibt (auch bei vielen
  Einträgen). Microcopy dieses Bereichs siehe NEU02/NEU11.

Finanzamt-Layout (P0) (#1):

- Der Container der drei Schritte hat dieselbe Breite wie die Panels darunter, oder
  die Schritte stapeln unterhalb ~1100px vertikal. Die ELSTER-UUID steht in einem
  eigenen, vollbreiten Feld mit Kopier-Button; die Aktionszeile bricht um, sodass
  „Als erledigt markieren“ bei 1440px und 1024px vollständig sichtbar und klickbar
  ist. Dark-Mode-Kontrast AA (siehe Muster 05).

### Freeze-Sicherheit

Alle Änderungen sind Frontend (TSX/CSS) plus wenige nutzersichtbare Strings
(NEU02-Titel, NEU03-Plural, Terminologie) plus mitgezogene Frontend-Tests. Keine
Änderung an DB-Schema, Event-JSON oder persistierten Daten. Das gilt auch für den
Token-Pass (CSS-Variablen) und das ADR (nur Doku). Damit fällt nichts unter die
Freeze-Disziplin.

## Testing Decisions

Getestet wird ausschließlich externes Verhalten, nie Implementierungsdetails
(keine Assertions auf Klassennamen; stattdessen Sichtbarkeit, Textinhalt,
berechneter Kontrast).

- Mobile-Überlauf (E2E/Regression bei 390px): kein horizontaler Seiten-Überlauf
  auf Historie, Kassieren, Direktverkauf-Historie, Tischauswahl und Finanzamt.
  Prior Art: die bestehende Mobile-E2E-Regression, die den Primär-Button im
  Viewport hält (ADR 03), und die Playwright-Overflow-Messung des Audits.
- Drawer-Verhalten: bei langer Bestellung/Stornierung/Umbuchung bleiben
  Gesamtsumme, Pflichtfeld und Primäraktion sichtbar; der Enabled-Zustand der
  Aktion folgt der Validierung; im deaktivierten Fall erscheint ein Grund. Prior
  Art: die bestehenden Vitest/Testing-Library-Komponententests der Drawer.
- Plural (Unit): rendert „1 Position“ (Singular) und „N Positionen“ (Plural). Die
  bestehenden Tests, die derzeit den kaputten String „Alle 1 Positionen“
  festschreiben (Kassieren, Umbuchung), werden angepasst, weil sie den Bug
  kodieren.
- Kontrast: ein automatisierter axe/Kontrast-Durchlauf auf den Dark-Mode-Recovery-
  und Compliance-Screens bestätigt AA. Das ist die im Audit als offen markierte
  Lücke.
- Token-Komponenten (Unit): Geld-Eingabe (Cent rein, formatierter Wert und Cent
  raus) und die Disabled-Affordanz (Grund wird bei deaktivierter Aktion gerendert).

## Out of Scope

- Keine neuen Features, keine Konfigurierbarkeit.
- Keine Änderungen an Backend, API, DB-Schema, Event-JSON oder persistierten Daten.
- Keine neuen Abhängigkeiten.
- Kein Redesign über die benannten Primitive hinaus; keine neue Designsprache.
- Vollständiger Tastatur- und Screenreader-Audit über den Kontrast-Durchlauf
  hinaus. Der Audit hat diesen Pass als noch nicht durchgeführt markiert; er wird
  als eigene Folgearbeit empfohlen, nicht hier erledigt.
- Der reine Label/URL-Kosmetik-Drift „Bondrucker“ gegen `/druckstationen`. Die URL
  ist nicht nutzersichtbar; Routen-Umbenennungen sind verschiebbar.
- #4 und #5 öffnen nicht die breitere Dialog-Muster-Frage; sie wenden nur die
  gewählte Variante an (Ein-Zeilen-Footer, ein Reset-Einstieg).

## Further Notes

- Sequenzierung: P0 zuerst (Drawer-Sticky-Footer, Raster-Sweep, Finanzamt), weil
  höchster Hebel, geringstes Risiko und alle drei Blocker damit erledigt sind.
  Danach P1 (Kern-Härtung). Zuletzt P2 (Token-Pass und Feinschliff), am breitesten
  und mit dem geringsten Risiko für den Nutzer.
- Raster-Vollständigkeit: der Audit benannte sieben Stellen; die Verifikation fand
  den tatsächlichen Satz (rund 15 Raster ohne Basis-Track) über Admin und Service.
  Das Akzeptanzkriterium ist deshalb verhaltensbasiert (kein Überlauf bei 390px),
  damit Umsetzende jede Fläche prüfen, statt einer driftenden Liste zu vertrauen.
- ADR: NEU07 als neues ADR in `docs/adrs/` festhalten (Warn-Bestätigung für
  irreversible Routine-Aktionen; bindet künftige Bestätigungsdialoge). #4 und #5
  sind hier als Entscheidungen dokumentiert und brauchen kein eigenes ADR, da sie
  ein bestehendes Muster anwenden und keine architekturbindende Ausnahme sind.
- Querbezug: G10 in `docs/plans/offene-punkte.md` hat den Dark-Mode-Kontrast der
  destruktiven Buttons bereits aufgeworfen und teils gelöst. Der Token-Pass soll
  diesen Punkt abschließen statt ihn zu duplizieren.
- Risiken:
  - Drawer-Umbau kann die iOS-PWA-Regression aus ADR 03 zurückbringen, wenn der
    Layout-Vertrag verletzt wird. Prüfung auf echtem Gerät ist Pflicht.
  - Der Raster-Sweep darf bewusste Mehrspalten-Layouts nicht zu Einspaltern
    machen; nur Raster ohne Basis-Track erhalten `grid-cols-1`, jede Fläche wird
    visuell geprüft.
  - Dark-Mode-Token-Änderungen dürfen den Light-Mode-Kontrast nicht senken; beide
    Modi müssen AA halten (vgl. die Token-Abwägung in G10).
