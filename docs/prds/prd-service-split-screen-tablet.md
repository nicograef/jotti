# PRD: Zweispaltiges Service-Layout für Tablet und größer

## Problem Statement

Der gesamte Service-Bereich ist heute mobile-first als einspaltige Ansicht
gebaut: Die Auswahl (Produktliste, offene Positionen, Historie) füllt eine
schmale Spalte, der jeweilige Abschluss (Warenkorb, Zahlung, Storno, Umbuchung)
steckt in einem Bottom-Sheet-Drawer, den ein fixierter Aktionsbutton am unteren
Rand öffnet. Dieses Layout gilt auf allen Viewports gleich, auch auf großen
Geräten.

Auf den eigenen Smartphones der Servicekräfte (BYOD) ist das richtig. Aber
sobald jotti auf einem stationären Gerät läuft, das der Verein als feste
Thekenkasse hinstellt (ein Tablet, ein Laptop, ein Surface im Pavillon), passt
die schmale Handy-Spalte nicht mehr: Die Bildschirmbreite bleibt ungenutzt, und
jeder Kassiervorgang kostet einen zusätzlichen Tipp, um den Warenkorb-Drawer zu
öffnen, bevor kassiert werden kann.

Ein Verein hat für genau dieses Szenario (Direktverkauf aus einem Pavillon auf
einem Surface) außerhalb von jotti ein eigenes zweispaltiges Thekenkassen-Frontend
gebaut (Produkte links, Warenkorb rechts, dauerhaft sichtbar) und im Einsatz als
deutlich angenehmer erlebt. Das bestätigt den Bedarf: Auf einem stationären
Kassengerät soll die Fläche genutzt und der Abschluss ohne Zwischenschritt
sichtbar sein.

## Solution

Der Service-Bereich wird responsiv zweispaltig. Ab einer Viewport-Breite von
1024px (Tailwind `lg`, Querformat-Tablets und größer) erhält jede
Zusammenstellungs-Fläche ein festes zweispaltiges Layout: links die Auswahl,
rechts eine dauerhaft sichtbare Abschluss-Spalte, die genau den Inhalt zeigt, der
heute im Drawer steckt (Belegvorschau, Erhalten/Rückgeld, Kommentar, primärer
Aktionsbutton). Der Bottom-Sheet-Drawer und die fixierte Dock-Leiste entfallen
auf diesen Breiten; die Abschluss-Spalte trägt den Aktionsbutton selbst.

Unterhalb von 1024px bleibt alles unverändert: dieselbe einspaltige Ansicht,
derselbe Bottom-Sheet-Drawer, dieselbe fixierte Dock-Leiste. Die BYOD-Smartphones
der Servicekräfte erleben keine Änderung, und der iOS-PWA-Bottom-Sheet aus ADR 03
bleibt auf Handy-Breiten unangetastet.

Betroffen sind drei Zusammenstellungs-Flächen, die dem gleichen Muster folgen
(Auswahl plus Abschluss):

1. **Direktverkauf, Reiter „Verkaufen":** Produktliste links, Warenkorb mit
   Erhalten/Rückgeld/Kommentar und „Verkauf abschließen" rechts.
2. **Tisch, Reiter „Bestellen":** Produktliste links, Bestellübersicht mit
   „Bestellen" rechts.
3. **Tisch, Reiter „Kassieren":** offene Positionen links, Zahlungsübersicht mit
   Erhalten/Zielbetrag/Rückgeld und „Zahlung abschließen" rechts.

Die zeilengebundenen Korrekturvorgänge (Stornieren und Umbuchen aus der Historie,
sowie der Storno aus der Direktverkauf-Historie) haben keine dauerhafte
Abschluss-Spalte, weil sie aus einer konkreten Historienzeile heraus gestartet
werden. Sie bleiben Overlays, werden ab 1024px aber als mittig zentrierter
Modal-Dialog statt als Bottom-Sheet dargestellt, damit auf großen Bildschirmen
nirgends mehr ein vom unteren Rand hereinfahrendes Sheet erscheint.

Reine Listen ohne Abschluss (Historie, Tischauswahl) bleiben wie heute: Sie
nutzen ihre bereits vorhandenen mehrspaltigen Rasterlayouts und bekommen keine
Abschluss-Spalte.

Fachlich ändert sich nichts. Validierung, Idempotenz, Fehlerbehandlung,
Erfolgs-Pop, Steuer- und TSE-Logik, alle Backend-Aufrufe und alle Event-Formate
bleiben identisch. Es ändert sich ausschließlich, wo der immer gleiche
Abschluss-Inhalt dargestellt wird: im Drawer (schmal) oder in einer festen Spalte
(breit).

## User Stories

1. Als Servicekraft an einer stationären Thekenkasse (Tablet/Laptop) will ich
   Produkte links und den Warenkorb rechts dauerhaft nebeneinander sehen, damit
   ich beim Direktverkauf ohne Zwischenschritt kassieren kann.
2. Als Servicekraft an einer stationären Kasse will ich beim Tisch-Bestellen die
   Produktliste und die entstehende Bestellung gleichzeitig sehen, damit ich
   Positionen erfasse, ohne einen Drawer zu öffnen.
3. Als Servicekraft an einer stationären Kasse will ich beim Kassieren die offenen
   Positionen und die Zahlungsübersicht (Erhalten/Rückgeld) nebeneinander sehen,
   damit ich schneller und sicherer abrechne.
4. Als Servicekraft auf meinem eigenen Smartphone will ich, dass die Bedienung
   exakt bleibt wie bisher (einspaltig, Bottom-Sheet), damit die BYOD-Nutzung
   unverändert vertraut und stabil ist.
5. Als Servicekraft auf einem großen Bildschirm will ich, dass Stornieren und
   Umbuchen als zentrierter Dialog statt als Bottom-Sheet erscheinen, damit die
   Korrekturvorgänge zum großen Layout passen.
6. Als Servicekraft will ich, dass die Erfolgsbestätigung, die Fehlermeldungen und
   die Betragsberechnung (Rückgeld, Restbetrag) in der breiten Ansicht genauso
   funktionieren wie in der schmalen, damit ich mich auf kein neues Verhalten
   umstellen muss.

## Implementation Decisions

- **Reiner Präsentations-Layer, keine fachliche Änderung.** Geändert wird nur die
  Darstellung des Abschluss-Inhalts (Drawer versus feste Spalte versus zentrierter
  Dialog) in Abhängigkeit vom Breakpoint. Backend, Endpunkte, Event-Formate,
  Validierung, Idempotenz und die Steuer-/TSE-Logik bleiben unberührt. Die Änderung
  ist damit freeze-konform (nur Frontend, kein Schema, keine Events, keine
  API-Verträge).

- **Breakpoint `lg` (1024px).** Ab 1024px zweispaltig, darunter unverändert. Ein
  Querformat-Surface (rund 1368px effektiv) fällt sicher in die breite Ansicht,
  Hochformat-Tablets und Smartphones bleiben einspaltig.

- **Abschluss-Inhalt wird darstellungsneutral extrahiert.** Der heutige
  Drawer-Body samt Footer jeder der drei Flächen (Belegvorschau, Eingaben,
  primärer Button) wird als darstellungsneutrale Inhaltskomponente herausgezogen,
  sodass identischer Inhalt und identische Logik sowohl im Bottom-Sheet-Drawer
  (schmal) als auch in der festen Abschluss-Spalte (breit) gerendert werden. Es
  gibt pro Fläche genau eine Quelle für den Abschluss-Inhalt; die beiden
  Darstellungen unterscheiden sich nur im umschließenden Container.

- **Responsive Service-Shell.** Der Service-Shell entscheidet je Fläche und
  Breakpoint über die Darstellung: unterhalb `lg` die bestehende Kombination aus
  scrollendem Listeninhalt, fixierter Dock-Leiste (Reiter plus Aktions-Slot) und
  Bottom-Sheet-Drawer; ab `lg` ein zweispaltiges Layout mit unabhängig
  scrollenden Spalten, Reitern am oberen Rand des Inhaltsbereichs und der
  Abschluss-Spalte rechts, die den Aktionsbutton selbst trägt. Die fixierte
  Dock-Leiste und ihr Portal-Aktions-Slot sind damit eine reine Handy-Darstellung.

- **Abschluss-Spalte nur für Zusammenstellungs-Flächen.** Genau die drei Flächen
  Direktverkauf-Verkaufen, Tisch-Bestellen und Tisch-Kassieren erhalten eine
  rechte Abschluss-Spalte. Historie und Tischauswahl bleiben volle Breite mit
  ihren vorhandenen Rasterlayouts.

- **Leerzustand der Abschluss-Spalte.** Solange nichts ausgewählt ist, zeigt die
  rechte Spalte einen Hinweiszustand (etwa „Produkte auswählen") und einen
  deaktivierten Aktionsbutton, gleichbedeutend mit dem heute deaktivierten
  Dock-Button.

- **Idempotenz-Schlüssel-Lebenszyklus in der festen Spalte.** Heute wird der
  Vorgangs-Schlüssel (`verkaufId` bzw. `zahlungId`) beim Öffnen des Drawers neu
  erzeugt. In der festen Spalte gibt es kein Öffnen; der Schlüssel wird stattdessen
  je logischem Vorgang erzeugt (neuer Schlüssel, wenn eine Zusammenstellung aus dem
  Leerzustand heraus beginnt, und erneut nach jedem erfolgreichen Abschluss beim
  Zurücksetzen). Ein Retry desselben Vorgangs behält seinen Schlüssel. Die
  bestehende Idempotenz-Garantie bleibt damit erhalten.

- **Korrekturvorgänge als zentrierter Dialog ab `lg`.** Der Drawer bleibt das eine
  Drawer-System (ADR 03), erhält aber eine responsive Präsentation: Bottom-Sheet
  unterhalb `lg`, mittig zentrierter Modal-Dialog ab `lg`. Der bindende
  Bottom-Sheet-Layoutvertrag aus ADR 03 (85dvh, Safe-Area, ein Scrollbereich, kein
  Drag-Handle) gilt unverändert für Handy-Breiten; die installierte iOS-PWA wird
  nur auf Handys betrieben und ist von der Dialog-Darstellung nicht betroffen.
  Betroffen sind die Storno- und Umbuchungs-Overlays aus Tisch- und
  Direktverkauf-Historie.

- **Erfolgs-Pop und Fehler-Feedback unverändert.** Der Erfolgs-Pop überlagert
  weiterhin bildschirmfüllend und wird nach dem Abschluss ausgelöst; der
  nachgelagerte Refetch läuft weiter beim Schließen des Pops. Fehler-Toasts und die
  Betragsberechnung (Rückgeld, Trinkgeld, Restbetrag) bleiben identisch.

- **Verhältnis zum Spektral-Redesign (ADR 06).** Das Spektral-Redesign des
  App-Frontends war ein reiner Restyling-Layer und ist bereits umgesetzt
  (ADR 06, akzeptiert 2026-07-14); es nannte Layout-/Strukturumbauten ausdrücklich
  als „out of scope". Dieses PRD ist die dazu komplementäre Layout-Änderung und
  baut auf dem umgesetzten Restyle auf: Das bestehende Motion-Inventar (Erfolgs-Pop,
  zählende Beträge, Listen-Eintritt, Statuswechsel) und die Typografie gelten in der
  zweispaltigen Ansicht unverändert weiter. Die extrahierten Abschluss-Komponenten
  behalten diese Animationen und Schriften bei; die zweispaltige Ansicht führt
  keine neuen Motion- oder Farb-Entscheidungen ein.

- **Neues ADR bei Umsetzung.** Die Änderung hebt die implizite Festlegung „der
  Service-Bereich ist durchgängig einspaltig/mobile-first" für große Bildschirme
  auf und führt die responsive Drawer-Präsentation ein. Das wird analog zu ADR 03
  als eigenes ADR festgehalten, inklusive des Breakpoints und der Regel, welche
  Flächen eine Abschluss-Spalte erhalten und welche als Dialog erscheinen.

- **Empfohlene Phasierung (für den Umsetzungsplan, nicht Teil des Zielbilds).**
  Zuerst Direktverkauf-Verkaufen (der belegte, konkrete Bedarf und die
  risikoärmste Fläche), dann Tisch-Bestellen und Tisch-Kassieren, zuletzt die
  Dialog-Darstellung der Korrekturvorgänge. Jede Scheibe ist für sich auslieferbar,
  weil unterhalb `lg` alles unverändert bleibt.

## Testing Decisions

- **Gute Tests prüfen externes Verhalten, nicht die Darstellung.** Geprüft werden
  sichtbare Texte, berechnete Beträge, Reihenfolge von Callbacks und die
  Backend-Aufrufe mit ihren Nutzlasten, nicht CSS-Klassen, Breakpoints oder
  welcher Container gerade rendert.

- **Extrahierte Abschluss-Inhalte werden getestet.** Für jede der drei Flächen
  wird die darstellungsneutrale Abschluss-Komponente isoliert geprüft: Sie rendert
  die ausgewählten Positionen und die Gesamtsumme, berechnet Rückgeld bzw.
  Restbetrag korrekt, ist im Leerzustand deaktiviert und ruft beim Abschluss das
  Backend genau einmal mit der erwarteten Nutzlast auf. Der
  Idempotenz-Schlüssel bleibt über einen Retry stabil und wechselt bei einem neuen
  logischen Vorgang.

- **Prior Art.** Das vorherrschende Muster sind Vitest-Komponententests mit
  Testing Library (etwa `DirektverkaufDrawer`, `ZahlungDrawer`, `Bestellung`,
  `Zahlung`). Weil der Abschluss-Inhalt aus den heutigen Drawern extrahiert wird,
  ziehen die bestehenden Verhaltenstests auf die extrahierten Komponenten um bzw.
  laufen unverändert weiter, da sie auf Texte und Bedienelemente matchen, nicht auf
  den Container.

- **Bestehende Suiten bleiben grün.** Da sich Inhalt und Bedienelemente nicht
  ändern, sondern nur ihr Ort, müssen die vorhandenen Service-Tests ohne
  inhaltliche Anpassung grün bleiben. Snapshot-Tests gibt es nicht und werden nicht
  eingeführt.

- **Zweispaltige Darstellung und Breakpoint-Verhalten sind manuelle Abnahme.**
  Welche Spalten bei welcher Breite erscheinen, die dauerhafte Sichtbarkeit der
  Abschluss-Spalte und die Dialog-statt-Bottom-Sheet-Darstellung der
  Korrekturvorgänge werden per Sichtabnahme über die Breakpoints geprüft (schmal,
  Querformat-Tablet, Laptop), in Hell und Dunkel, analog zur Abnahme des
  Spektral-Redesigns.

## Out of Scope

- Jede fachliche Änderung: keine neuen Endpunkte, keine Event-Änderungen, keine
  geänderte Validierung, Idempotenz, Steuer- oder TSE-Logik.
- Reine Listen ohne Abschluss: Historie und Tischauswahl bekommen keine
  Abschluss-Spalte (ihre vorhandenen Rasterlayouts bleiben).
- Der Admin-Bereich. Dieses PRD betrifft nur den Service-Bereich (`/service/*`).
- Das Handy-Layout unterhalb 1024px: unverändert, inklusive Bottom-Sheet-Drawer,
  fixierter Dock-Leiste und iOS-PWA-Layoutvertrag aus ADR 03.
- Visuelles Restyling (Spektral-Verlauf, Motion-Inventar, Typografie): bereits
  umgesetzt (ADR 06). Dieses PRD ändert daran nichts, sondern übernimmt das
  bestehende Verhalten in der zweispaltigen Ansicht.
- Weitere Rückmeldungen aus der Vereins-E-Mail, die getrennt behandelt werden:
  README-Hinweis zum Erst-Login (kleiner Doku-Fix), Bon-per-E-Mail (eigenes,
  greenfield PRD), Helferdeckel-Modus (gesonderte Bedarfsprüfung), sowie der
  Betrieb ohne Internet und die TSE-Nachsignierung (Betreiber-/Doku-Thema, keine
  Code-Änderung).

## Further Notes

- **Nutzenschwerpunkt und ehrliche Risikoeinordnung.** Der belegte Bedarf liegt
  beim Direktverkauf auf einem stationären Kassengerät; dort ist der Gewinn am
  größten. Tisch-Bestellen und Tisch-Kassieren laufen in der Praxis meist auf den
  BYOD-Smartphones der Servicekräfte (unterhalb `lg`, unverändert); der Gewinn
  dort greift nur, wenn Tischservice ausnahmsweise auf einem großen Gerät bedient
  wird. Der volle Service-Umfang ist bewusst gewählt, damit ein großes Gerät
  nirgends mehr in die Handy-Spalte fällt; der Preis ist ein größerer Umbau der
  meist genutzten und sicherheitskritischen Kassen-Oberfläche. Die empfohlene
  Phasierung liefert den konkreten Bedarf zuerst und hält jede Scheibe für sich
  auslieferbar.

- **Kein zweites UI-System.** Es entsteht kein paralleles Layout- oder
  Drawer-System. Der Abschluss-Inhalt bleibt pro Fläche eine einzige Komponente,
  und der Drawer bleibt das eine Drawer-System mit einer zusätzlichen responsiven
  Präsentation. Das hält den Wartungsaufwand und die Testfläche klein.

- **Referenz-Verhalten.** Der bestehende Drawer-Body und -Footer je Fläche
  definieren das Sollverhalten der extrahierten Abschluss-Komponenten
  (Belegvorschau, Erhalten/Rückgeld/Kommentar, primärer Button, Retry- und
  Fehlerverhalten). Die zweispaltige Ansicht darf davon nur in der Anordnung
  abweichen, nicht im Verhalten.
