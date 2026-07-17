# PRD: UX-Quick-Wins aus dem UX-Review (Juli 2026)

## Problem Statement

Das UX-Review vom Juli 2026 (Design-Handoff `design_handoff_jotti_ux_review/`) hat auf Basis des Frontend-Codes Reibungspunkte in Service, Admin und Login identifiziert. Alle Befunde wurden am 2026-07-17 gegen den aktuellen Stand von `main` verifiziert; keiner ist inzwischen behoben. Aus Nutzersicht:

- Servicekraft: Wer mitten in der Bestellung kurz in die Historie schaut, kommt zu einem kommentarlos geleerten Warenkorb zurück, weil die Tabs inaktive Inhalte auswerfen und die Auswahl lokal in den Tab-Flächen lebt. Gleiches gilt für die Kassieren-Auswahl. Im Gedränge ist das der teuerste Moment der App: neu zusammentippen, während der Gast wartet.
- Serviceleitung: Die Stornierung, ausgerechnet die geldrelevante, rollengeschützte Korrektur, bestätigt gar nicht (der Drawer schließt kommentarlos). Umbuchung bestätigt mit Toast, Bestellen/Kassieren/Direktverkauf mit Vollbild-Pop: drei Muster für dieselbe Sache.
- Servicekraft: Das Euro-Eingabefeld normalisiert per 1-Sekunden-Debounce mitten im Tippen. Wer langsam „15" tippt, hat nach der Pause „1,00" im Feld; die nächste Ziffer wird verworfen.
- Alle Benutzer: Der Anmelden-Button ist bei ungültigem Formular gesperrt; ein Tap darauf löst keine Validierung aus, der Nutzer erfährt nie, warum nichts passiert. Gleiches beim Passwort-festlegen-Formular.
- Admin (Dark Mode): Fünf Lösch-Bestätigungen stylen ad hoc weißen Text auf der im Dark Mode aufgehellten Destructive-Fläche, Kontrast ca. 2,9:1 und damit unter WCAG AA. Es ist zugleich das einzige Styling im Projekt, das am Token-System vorbeigeht.
- Kleinere Inkonsistenzen: Der Auswählen-Button beim Kassieren sagt „Alle N Positionen", wählt aber nur die eigenen; Refresh-Button „Jetzt" ist kryptisch; Stift-Tooltip „Bearbeiten" und Menüpunkt „Umbenennen" öffnen denselben Dialog; das LogOut-Icon steht sowohl für den Bereichswechsel als auch fürs Abmelden; die Kassensturz-Differenz zeigt bei Überschuss kein Vorzeichen; die Tischseite lädt mit wörtlich „Tisch ??" statt Skeletons; der Normalzustand „unbezahlt" trägt das per ADR 04 für Gefahr reservierte Rot.
- Servicekraft: Der Wechsel Tischservice/Direktverkauf ist im Benutzermenü versteckt; der Empty-State muss ihn per Fließtext erklären. Helfer, die zwischen Theke und Tischen rotieren, wechseln mehrmals pro Schicht.
- Servicekraft: „Machen wir 15" erfordert heute ein zweites Euro-Feld plus dreizeiligen Hinweistext; der Direktverkauf kennt gar kein Trinkgeld. Im „Alle Tische"-Drawer stehen Namen nicht beieinander (Sortierung nach Saldo), gezieltes Finden erfordert Scrollen. Die Drucker-IP speichert unsichtbar on-blur, einziger Hinweis ist ein flüchtiger Toast.

## Solution

Ein Frontend-only-Feinschliff-Paket im bestehenden Token- und Komponentensystem (shadcn/ui, Tailwind 4; ADR 04, 07 und 08 bleiben gültig), kein neues Design, keine Backend-Änderung:

1. Vorgangs-State überlebt Tab-Wechsel: Bestell-Korb und Kassieren-Auswahl bleiben beim Wechsel zwischen den Tabs der Tischseite erhalten.
2. Einheitliches Erfolgs-Feedback: Stornierung und Umbuchung bestätigen mit demselben Vollbild-Pop wie Bestellen, Kassieren und Direktverkauf. Regel: Alles, was das Kassenjournal verändert, bestätigt mit dem ErfolgsPop.
3. Euro-Eingabe ohne Überraschungen: Normalisierung nur noch beim Verlassen des Felds.
4. Aktiver Anmelden-Button: Ein Tap löst den Submit aus und zeigt die Feldfehler.
5. AA-konforme Lösch-Bestätigungen: neuer Button-Variant `destructive-solid` nach dem Warn-Muster, der seinen Kontrast in Light und Dark selbst trägt.
6. Copy- und Icon-Konsistenz: ehrliches „Meine N Positionen"-Label, „Aktualisieren", ein Begriff für Bearbeiten, eigenes Icon für den Bereichswechsel, Kassensturz-Differenz immer mit Vorzeichen, Skeletons statt „Tisch ??".
7. „Unbezahlt" verlässt das Gefahren-Rot: Amber-Soft-Tint für den Badge, abgestufte Punkt-Farben in „Meine Tische"; Rot bleibt Storno und Fehlern vorbehalten.
8. Sichtbarer Modus-Wechsel: Segmented Control „Tische | Theke" im Service-Header, ein Tap in beide Richtungen.
9. Aufrunden-Chips fürs Trinkgeld in beiden Abschlüssen (Tisch-Kassieren und Direktverkauf), das freie Feld bleibt hinter „Anderer …" erreichbar.
10. Kleine Flow-Verbesserungen: Namens-Sortierung im „Alle Tische"-Drawer, kurze Inline-Bestätigung beim Speichern der Drucker-IP.

## User Stories

1. Als Servicekraft möchte ich, dass meine Bestell- und Kassieren-Auswahl einen Tab-Wechsel (etwa einen kurzen Blick in die Historie) übersteht, damit ich im Gedränge nichts neu zusammentippen muss.
2. Als Serviceleitung möchte ich nach Stornierung und Umbuchung dieselbe unübersehbare Erfolgsbestätigung wie beim Bestellen und Kassieren, damit ich sicher weiß, dass die Korrektur gebucht wurde.
3. Als Servicekraft möchte ich Euro-Beträge in meinem Tempo eintippen können, ohne dass das Feld die Eingabe zwischendurch umformatiert und Ziffern verwirft.
4. Als Benutzer möchte ich beim Tap auf „Anmelden" bzw. „Passwort festlegen" sehen, welche Felder fehlen oder falsch sind, statt vor einem stumm gesperrten Button zu stehen.
5. Als Admin möchte ich Lösch-Bestätigungen auch im Dark Mode klar lesen können (WCAG AA), damit ich destruktive Aktionen sicher erkenne.
6. Als Servicekraft möchte ich, dass der Auswählen-Button beim Kassieren „Meine N Positionen auswählen" sagt, weil er nur meine eigenen Positionen wählt.
7. Als Admin möchte ich eindeutige Labels und Icons (Aktualisieren, ein Begriff fürs Bearbeiten, eigenes Icon für den Bereichswechsel) und eine Kassensturz-Differenz mit Vorzeichen, damit ich nichts interpretieren muss.
8. Als Servicekraft möchte ich beim Laden der Tischseite Skeleton-Platzhalter statt „Tisch ??" sehen, damit die App nicht kaputt wirkt.
9. Als Servicekraft möchte ich, dass „unbezahlt" nicht wie ein Fehler aussieht, damit Rot seine Warnwirkung für Storno und echte Fehler behält.
10. Als Servicekraft möchte ich mit einem Tap im Service-Header zwischen Tischen und Theke wechseln, weil ich mehrmals pro Schicht rotiere.
11. Als Servicekraft möchte ich „machen wir 15" mit einem Tap auf einen Aufrunden-Chip erfassen (auch an der Theke), damit Trinkgeld und Rückgeld sofort dastehen.
12. Als Servicekraft möchte ich im „Alle Tische"-Drawer die Tische nach Namen sortiert sehen, damit ich einen bestimmten Tisch ohne Schließen des Drawers finde.
13. Als Admin möchte ich beim Speichern der Drucker-IP eine kurze Bestätigung direkt am Feld sehen, damit ich nicht auf einen flüchtigen Toast angewiesen bin.

## Implementation Decisions

Alle Änderungen ausschließlich im Frontend, im bestehenden Designsystem; die Mockups aus dem Handoff sind Richtungs-Referenzen, kein Produktionscode.

Block A (vor 1.0):

- A1 Vorgangs-State: Die beiden Mengen-Auswahlen (Bestell-Korb und Kassieren-Auswahl) werden von den Tab-Inhalten auf die Tischseite gehoben und als Props durchgereicht. Nur die Mengen wandern; übriger UI-State der Tab-Flächen (etwa das Auf-/Zuklappen der Fremd-Positionen) bleibt lokal. Der Lebenszyklus der Abschluss-Komponenten (Idempotenz-Schlüssel, Leeren der Auswahl nach Erfolg, warLeer-Muster) bleibt unverändert. Kein forceMount; die fadeUp-Tab-Animation bleibt.
- A2 Erfolgs-Feedback: Stornierung und Umbuchung werden an den bestehenden ErfolgsPop-Mechanismus der Tischseite angeschlossen; der Erfolgs-Callback wird über die Historien-Komponente an beide Drawer durchgereicht (gleiches Prop-Muster wie bei Bestellen und Kassieren). Texte: „Stornierung gebucht." und „Auf {Zielname} umgebucht." (Zielname ist der Ziel-Tisch). Der bisher sofortige Refetch der Drawer folgt künftig dem Pop-Muster: Refetch erst beim Schließen des Pops. Der Umbuchungs-Toast entfällt.
- A3 Euro-Eingabe: Das Debounce-Reformat entfällt ersatzlos (Timer, Cleanup-Effect und Timeout-Zweig); normalisiert wird nur noch onBlur.
- A4 Login: Anmelden- und Passwort-festlegen-Button sind nur noch während des Ladens deaktiviert; ein Tap löst den Submit aus, React Hook Form zeigt die Feldfehler. Doppel-Submits verhindert weiterhin der Ladezustand.

Block B (Copy-&-Token-Sweep):

- B1 destructive-solid: Neuer Button-Variant `destructive-solid` nach dem Warn-Muster aus ADR 04: eine solide destruktive Fläche, die ihren Kontrast selbst trägt. Light: destruktive Fläche mit weißem Text (Light-Destructive ist dunkel genug). Dark: helle rote Fläche mit dunklem Text über ein neues Token `--destructive-solid-foreground` (Dark etwa red-950), analog zum Paar `--warn`/`--warn-foreground`. Die fünf verifizierten Ad-hoc-Call-Sites (Produkt, Variante, Tisch, Helfer, Druckaufträge; das Review nannte sechs, verifiziert sind fünf) stellen auf den Variant um; danach existiert die Ad-hoc-Klassenkombination nicht mehr im Repo. Der axe-Kontrast-E2E-Spec wird erweitert, die betroffenen Dialoge tatsächlich zu öffnen (heute prüft er nur geladene Seiten ohne Dialoge).
- B2 Copy und Icons (ein Commit, keine Verhaltensänderung): Kassieren-Button „Meine N Positionen auswählen · X €", die Umbuchung behält „Alle" (dort stimmt es); Refresh-Button „Aktualisieren" statt „Jetzt"; Produkt-Dropdown „Bearbeiten" statt „Umbenennen" (gleiches Label wie der Stift-Tooltip, beide öffnen denselben Dialog); Icon für „Zum Service-Bereich" wird ArrowRightLeft, LogOut bleibt exklusiv beim Abmelden; Kassensturz-Differenz immer mit Vorzeichen („+12,50 €" bei Überschuss) über eine kleine Format-Hilfsfunktion neben der bestehenden Euro-Formatierung, die Farblogik (Rot nur bei Fehlbetrag) bleibt; Ladezustand der Tischseite mit Skeleton-Bausteinen statt „Tisch ??"/„?".
- B3 Unbezahlt-Farben: Der „N unbezahlt"-Badge wird ein Amber-Soft-Tint (Entscheidung: Amber statt neutral, „wartet auf dich"-Semantik, konsistent zur Warn-Familie). Der Status-Punkt in „Meine Tische": eigene offene Positionen amber, nur fremde offene neutral, alles erledigt grün (heute rot/amber/grün). Rot bleibt Storno-Beträgen und Fehlerzuständen vorbehalten (ADR 04).

Block C (nach 1.0):

- C1 Mode-Switcher: Im Service-Header ersetzt eine Segmented Control „Tische | Theke" den bisherigen Titel (Tabs-Optik, aber Navigation über Links auf die beiden Modus-Startrouten; der aktive Zustand kommt aus der Route). Auf der Tisch-Detailseite bleibt der „‹ Meine Tische"-Backlink statt des Switchers. Der Menüeintrag im Benutzermenü bleibt als Zweitweg. Der Hinweistext im Empty-State der Tischauswahl wird entsprechend gekürzt. Die Arbeitsmodus-Persistenz (zuletzt genutzter Modus pro Gerät) bleibt unverändert; Tap-Ziele mindestens 44 px.
- C2 Aufrunden-Chips: Neue, von beiden Abschlüssen (Tisch-Kassieren und Direktverkauf) verwendete Chip-Komponente; der Direktverkauf bekommt damit erstmals eine Trinkgeld-Anzeige. Chips aus dem Gesamtbetrag abgeleitet: der exakte Betrag (etwa „13 € genau"), dann die nächsten zwei bis drei glatten Beträge (auf volle 1 € und volle 5 € aufgerundet, Duplikate entfernt; ist der Gesamtbetrag selbst glatt, dienen die nächsthöheren glatten Beträge als Vorschläge), plus „Anderer …", der das bisherige Euro-Feld einblendet. Ein Tap setzt den Zielbetrag; Trinkgeld- und Rückgeld-Zeilen rechnen wie bisher rein clientseitig. Ein aktiver Chip ist per erneutem Tap abwählbar (zurück zu „genau"). Der dreizeilige Hinweistext entfällt. Keine Backend- oder Payload-Änderung: Zielbetrag und Erhalten gehen weiterhin nicht an die API (ADR 08); Sheet unter 1024 px und Spalte ab 1024 px verhalten sich identisch. Chips mindestens 44 px hoch, Ziffern tabular.
- C3 Flow-Verbesserungen: Der „Alle Tische"-Drawer sortiert durchgehend nach Tischname (numerisch bewusst, „Tisch 2" vor „Tisch 10") statt Favoriten, dann Saldo; die Favoriten-Gruppierung entfällt im Drawer, die Hauptseite bleibt unverändert. Entscheidung: kein zweites Suchfeld im Drawer, die Konsolidierung auf die eine Hauptsuche bleibt bestehen. Beim on-blur-Speichern der Drucker-IP erscheint zusätzlich zum Toast eine kurze Inline-Bestätigung am Feld (etwa zwei Sekunden), nach dem bestehenden Muster der Kopier-Bestätigung an der TSE-Seriennummer.

## Testing Decisions

- Gute Tests prüfen von außen sichtbares Verhalten (was der Nutzer sieht und tut), keine Implementierungsdetails: Auswahl nach Tab-Wechsel noch vorhanden, Pop-Text erscheint nach Storno, Feldinhalt nach langsamer Eingabe, Feldfehler nach Submit-Tap, Label-Texte, Chip-Ableitung, Sortierreihenfolge.
- Erweitert werden die bestehenden komponenten-nahen Vitest/Testing-Library-Tests der betroffenen Bereiche (Tischseite, Bestellung, Zahlung, Euro-Eingabe, Historien-Drawer, Login- und Passwort-Formular, Tischauswahl-Drawer). Neu hinzu kommen: „Auswahl überlebt Tab-Wechsel", der Pause-Fall der Euro-Eingabe, die Aufrunden-Chips (Ableitung, Deduplizieren, Abwählen, „Anderer …") und die Segmented Control (Navigation und aktiver Zustand; für das Service-Layout existiert bisher kein Test).
- E2E: Der axe-Kontrast-Spec öffnet künftig die Lösch-Bestätigungen und prüft sie in Light und Dark. E2E-Specs und Unit-Tests matchen auf sichtbare Texte; die Copy-Änderungen aus Block B ziehen die betroffenen Assertions mit.
- Prior Art: die vorhandenen colocated `*.test.tsx` im Frontend und die Playwright-/axe-Specs im e2e-Verzeichnis.

## Out of Scope

- Backend-, API- oder Schema-Änderungen: Das gesamte Paket ist Frontend-only; die Kassieren- und Direktverkauf-Payloads bleiben unverändert.
- Direkte „Stornieren…"-Aktion an der Historien-Zeile (Handoff-Ticket C3.2): gestrichen. Der Zwei-Drawer-Weg bleibt für die seltene, rollengeschützte Korrektur (Produkt-Konservatismus, Präzedenz ADR 01).
- Zweites Suchfeld im „Alle Tische"-Drawer: Die frische Konsolidierung auf die Hauptsuche bleibt bestehen; es kommt nur die Namens-Sortierung.
- Amber-Einfärbung der positiven Kassensturz-Differenz (im Review nur als „ggf." erwähnt): Es kommt nur das Vorzeichen, die Farblogik bleibt.
- forceMount-Variante für die Tabs (im Handoff explizit ausgeschlossen; stattdessen wird der State gehoben).
- Pixel-Übernahme aus den Handoff-Mockups: Abstände und Varianten werden aus dem bestehenden Designsystem abgeleitet.
- Alles, was das Review ausdrücklich als Stärke benennt (ErfolgsPop blockiert bis zum Refetch, ActionHint-Muster, Drei-Farben-Semantik als solche), bleibt unangetastet.

## Further Notes

- Quelle: `docs/prds/design_handoff_jotti_ux_review/` (README, TICKETS.md, zwei HTML-Referenzen). Alle Ticket-Behauptungen wurden am 2026-07-17 gegen `main` verifiziert. Zwei Korrekturen gegenüber dem Handoff: Es sind fünf statt sechs Ad-hoc-Lösch-Buttons, und das Drawer-Suchfeld-Ticket kollidierte mit der einen Tag alten Konsolidierung der Tischsuche (aufgelöst durch die Namens-Sortierungs-Entscheidung).
- Die Handoff-Reihenfolge bleibt als Phasen-Empfehlung für den Implementierungsplan: Block A vor 1.0 (Verhalten, das Beta-Tester als Bug melden würden), dann Block B als gebündelter Sweep, dann Block C.
- Für alle Service-Tickets gilt ADR 08 als Querschnitts-Akzeptanzkriterium: Sheet (unter 1024 px) und dauerhafte Abschluss-Spalte (ab 1024 px) verhalten sich identisch; nach erfolgreichem Bestellen/Kassieren wird die jeweilige Auswahl weiterhin geleert.
- Das Handoff-Verzeichnis wird nach vollständiger Umsetzung gelöscht (wie frühere Handoff-Bundles); dieses PRD ist der maßgebliche Extrakt.
