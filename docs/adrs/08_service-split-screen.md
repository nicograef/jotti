# ADR 08: Zweispaltiges Service-Layout ab `lg` mit responsiver Drawer-Präsentation

- **Status:** akzeptiert (2026-07-15)
- **Kontext-Dokumente:** `docs/prds/prd-service-split-screen-tablet.md`;
  `docs/plans/plan-service-split-screen.md`; ADR 03 (Drawer-Layoutvertrag),
  ADR 07 (Desktop-Schwelle `lg`)
- **Nachtrag (2026-07-28):** Die Feststellung unten unter „Entscheidung"
  („Tisch-Kassieren trägt bewusst **keinen** Client-Schlüssel
  (`ZahlungKassierenSchema` hat keinen); die Idempotenz bleibt dort
  zustandsbasiert") ist überholt. `ZahlungKassierenSchema` trägt inzwischen ein
  `vorgangId`; alle sieben buchenden Endpunkte tragen seit der Nacharbeit zur
  Client-Server-Robustheit einen client-erzeugten Idempotenz-Schlüssel, den der
  Server an die Nutzdaten des Vorgangs bindet (`vorgang_idempotenz`,
  Migration `07`). Der übrige Inhalt dieses ADR — Layout, Container-Wahl, eine
  Quelle je Abschluss-Inhalt — bleibt akzeptiert und unverändert gültig.

## Kontext

Der Service-Bereich (`/service/*`) war durchgängig mobile-first einspaltig:
Die Auswahl (Produktliste, offene Positionen, Historie) füllt eine schmale
Spalte, der jeweilige Abschluss (Warenkorb, Zahlung, Storno, Umbuchung) steckt
in einem Bottom-Sheet-Drawer, den ein fixierter Dock-Aktionsbutton öffnet.
Dieses Layout galt auf allen Viewports.

Auf den BYOD-Smartphones der Servicekräfte ist das richtig. Sobald jotti aber
auf einem stationären Kassengerät läuft (Tablet, Laptop, Surface im Pavillon),
bleibt die Bildschirmbreite ungenutzt, und jeder Kassiervorgang kostet einen
zusätzlichen Tipp, um den Warenkorb-Drawer zu öffnen. Ein Verein hat für genau
dieses Szenario außerhalb von jotti ein zweispaltiges Thekenkassen-Frontend
gebaut und als deutlich angenehmer erlebt — belegter Bedarf.

Damit stand die implizite Festlegung „der Service-Bereich ist durchgängig
einspaltig" für große Bildschirme zur Disposition. Zugleich sollte der bindende
Bottom-Sheet-Layoutvertrag aus ADR 03 (installierte iOS-PWA, nur Handy-Breiten)
unangetastet bleiben.

## Entscheidung

Der Service-Bereich wird ab der bestehenden Desktop-Schwelle `lg` (1024px,
ADR 07) responsiv zweispaltig; darunter bleibt alles unverändert.

- **Drei Zusammenstellungs-Flächen erhalten eine dauerhaft sichtbare
  Abschluss-Spalte:** Direktverkauf-Verkaufen, Tisch-Bestellen, Tisch-Kassieren.
  Links die Auswahl, rechts der Abschluss-Inhalt (Belegvorschau, Eingaben,
  primärer Button). Die fixierte Dock-Leiste und ihr Portal-Aktions-Slot sind
  damit eine reine Handy-Darstellung; die Reiter liegen ab `lg` oben im
  Inhaltsbereich.
- **Reine Listen behalten volle Breite:** Historie und Tischauswahl bekommen
  keine Abschluss-Spalte, sondern behalten ihre vorhandenen Rasterlayouts.
- **Eine Quelle je Abschluss-Inhalt.** Der heutige Drawer-Body samt Footer jeder
  Fläche wird als eine darstellungsneutrale Abschluss-Inhaltskomponente
  extrahiert (Beleg, Eingaben, Betragsberechnung, primärer Button, Submit-/
  Fehler-/Retry-Verhalten, Idempotenz-Schlüssel-Lebenszyklus). Diese eine
  Komponente rendert sowohl im Bottom-Sheet-Drawer (schmal) als auch in der
  festen Spalte (breit); die Varianten unterscheiden sich nur im umschließenden
  Container. Kein zweites Layout- oder Drawer-System.
- **Container-Wahl per `useIsMobile()`, nicht per CSS-Sichtbarkeit.** Der
  JS-Breakpoint entscheidet, welcher Container mountet, damit der
  Abschluss-Inhalt genau einmal existiert — eine State-Quelle, ein
  Idempotenz-Schlüssel — und nicht zwei parallel gemountete Bäume doppelten
  Zustand führen.
- **Idempotenz-Schlüssel je logischem Vorgang** (Direktverkauf `verkaufId`,
  Bestellen `bestellungId`): neu, sobald eine Zusammenstellung aus dem
  Leerzustand beginnt, und erneut nach jedem erfolgreichen Abschluss beim
  Zurücksetzen; über einen Retry stabil. Tisch-Kassieren trägt bewusst **keinen**
  Client-Schlüssel (`ZahlungKassierenSchema` hat keinen); die Idempotenz bleibt
  dort zustandsbasiert, und der `useActionSubmit`-Loading-Guard verhindert den
  Doppel-Submit wie zuvor.
- **Korrekturvorgänge als zentrierter Dialog ab `lg`.** Die zeilengebundenen
  Overlays (Stornieren und Umbuchen aus der Tisch-Historie, Storno aus der
  Direktverkauf-Historie) bleiben auf dem einen Drawer-Primitive (ADR 03),
  erhalten aber eine responsive Präsentation: Bottom-Sheet unterhalb `lg`,
  mittig zentrierter Modal-Dialog ab `lg`. Umgesetzt als responsive Variante im
  `DrawerContent` — kein Fork der Primitive.
- **Der ADR-03-Handy-Vertrag gilt unter `lg` wörtlich weiter** (85dvh,
  Safe-Area, ein Scrollbereich, kein Drag-Handle). Die installierte iOS-PWA läuft
  nur auf Handy-Breiten und ist von der Dialog-Darstellung nicht betroffen.

Reiner Präsentations-Layer: kein Backend, keine Endpunkte, keine Event-Formate,
keine Validierung, Idempotenz oder Steuer-/TSE-Logik ändern sich. Damit
freeze-konform (nur Frontend).

## Konsequenzen

- Auf großen Bildschirmen fährt in keiner Service-Fläche mehr ein Sheet vom
  unteren Rand herein; die Abschluss-Spalte macht den Warenkorb ohne
  Zwischenschritt sichtbar.
- Die sicherheits- und geldrelevanten Kassen-Oberflächen werden umgebaut. Die
  darstellungsneutrale Extraktion muss verhaltensgleich zum heutigen Drawer sein;
  die bestehenden Verhaltenstests ziehen auf die extrahierten Komponenten um
  bzw. laufen unverändert weiter (sie matchen auf Texte und Bedienelemente, nicht
  auf den Container).
- Zweispaltige Darstellung und Breakpoint-Verhalten sind manuelle Abnahme
  (schmal, Querformat-Tablet, Laptop; Hell/Dunkel), analog zur Spektral-Abnahme
  (ADR 06). Automatisierte Tests prüfen externes Verhalten, nicht welcher
  Container rendert; keine Snapshot-Tests.
- Das Motion-Inventar und die Typografie aus dem Spektral-Redesign (ADR 06)
  gelten in der zweispaltigen Ansicht unverändert weiter; die Layout-Änderung
  führt keine neuen Motion- oder Farb-Entscheidungen ein.
- **Eingabe-State ist vorgangs-, nicht container-gebunden.** Die extrahierten
  Abschluss-Komponenten sind dauerhaft gemountet (in der Spalte ohnehin, im Sheet
  als Kind des Drawers), und Erhalten/Zielbetrag/Kommentar werden je logischem
  Vorgang zurückgesetzt (Beginn aus dem Leerzustand bzw. nach erfolgreichem
  Abschluss), nicht beim Schließen des Drawers. Ein bewusster, vereinheitlichter
  Bruch mit dem alten Handy-Drawer, der Erhalten/Zielbetrag beim Schließen leerte:
  Ein Schließen ohne Auswahländerung beendet den Vorgang nicht, die Eingaben
  bleiben also erhalten (Handy und Spalte verhalten sich gleich). Kein
  Geld-/Nutzlast-Effekt — Erhalten/Zielbetrag steuern nur die sichtbare Rückgeld-/
  Trinkgeld-Berechnung und gehen nie an das Backend; die Felder sind kontrolliert,
  Angezeigtes und Berechnetes stimmen immer überein.
- Eine Rückkehr zum durchgängig einspaltigen Service-Bereich bräuchte ein neues
  ADR.
