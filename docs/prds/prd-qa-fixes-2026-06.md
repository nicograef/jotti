# PRD: QA-Fixes Juni 2026

> Behebt 7 der 8 Befunde aus der QA-Session vom 2026-06-12 (`docs/qa-notes.md`).
> Befund 5 (Direktverkauf-Prominenz / Modus-Trennung) ist ein UX-Rework und bekommt ein
> eigenes PRD — siehe Out of Scope.
> Entscheidungen geklärt am 2026-06-12 (Scope Befund 5, Sticky-Action-Bar für Befund 7,
> 0-Cent-Eröffnung erlaubt für Befund 8, Testumfang Backend + gezieltes Frontend).
>
> Die Befunde sind voneinander unabhängig. Sie behalten in diesem PRD die Nummerierung aus
> den QA-Notizen und sind so geschnitten, dass jeder Befund im Implementierungsplan eine
> eigene Phase (vertikaler Slice) werden kann.

## Problem Statement

Eine QA-Session am 2026-06-12 hat acht Befunde ergeben, die das Vertrauen in jotti und die
Bedienbarkeit im Servicebetrieb beeinträchtigen. Sieben davon werden mit diesem PRD behoben:

- **Befund 1 — Interner Fehler statt Hilfestellung:** Ein Vereins-Admin, der jotti frisch
  aufgesetzt hat und seine erste Kassensitzung eröffnen will, bekommt einen 500-Fehler mit
  generischer Fehlermeldung — weil die Betreiber-Stammdaten noch nicht gepflegt sind. Der
  fachlich erwartbare Zustand „noch nicht konfiguriert" wird wie ein Datenbankfehler
  behandelt. Der Admin erfährt nicht, was er tun muss, und hält jotti für kaputt.
- **Befund 2 — Toasts ohne semantische Farben:** Alle Toast-Meldungen erscheinen neutral
  eingefärbt. Erfolg, Warnung und Fehler sind nur am Icon unterscheidbar — im hektischen
  Servicebetrieb auf dem Smartphone reicht das nicht für die Erkennung auf einen Blick.
- **Befund 3 — Produktvarianten springen:** Beim Aktivieren/Deaktivieren oder Bearbeiten
  einer Produktvariante ändert sich die Reihenfolge der Varianten in der Admin-Ansicht.
  Ursache: Die Varianten werden in den Produkt-Queries per `json_agg` ohne `ORDER BY`
  aggregiert — nach einem `UPDATE` ändert sich die physische Zeilenreihenfolge in PostgreSQL.
- **Befund 4 — Druckstationen springen:** Beim Speichern einer Druckstation (Drucker-IP,
  Bonmodus) ändert sich die Reihenfolge der Druckstationen. Ursache: Die
  Druckstationen-Query hat kein `ORDER BY`.
- **Befund 6 — Favoriten ohne Aktualisierung:** Markiert eine Servicekraft einen Tisch als
  Favorit (oder entfernt ihn), erscheint die Änderung unter „Meine Tische" erst nach einem
  harten Seiten-Reload. Ursache: Die Favoriten-Mutation invalidiert nur die Query der
  Tischauswahl-Liste, nicht die Query hinter „Meine Tische".
- **Befund 7 — Primäraktion außer Reichweite:** Auf der Tischseite (Tabs „Bestellen" und
  „Kassieren") stehen die primären Aktions-Buttons über der langen Produkt- bzw.
  Positionsliste. Die Servicekraft scrollt beim Auswählen nach unten und muss zum Abschließen
  wieder ganz nach oben scrollen — bei jeder einzelnen Bestellung.
- **Befund 8 — Gültige 0-Werte werden abgelehnt:** Die Backend-Validierung (zog) behandelt
  Go-Zero-Values bei `Required()` als fehlend. Eine Kassensitzung mit 0 € Anfangsbestand
  wird mit 400 abgelehnt, obwohl Schema-Absicht und Frontend 0 zulassen. Derselbe Effekt
  träfe einen Kassensturz mit 0 € Ist-Bestand — eine leere Kasse ist aber ein gültiges
  Zählergebnis.

## Solution

Jeder Befund wird unabhängig behoben:

- **Befund 1:** Fehlende Betreiber-Stammdaten beim Eröffnen einer Kassensitzung führen zu
  einem Client-Fehler mit dem bestehenden Fehlercode `betreiber_nicht_konfiguriert`. Das
  Frontend zeigt dafür eine konkrete Handlungsanweisung: Betreiber-Stammdaten in den
  Einstellungen pflegen, dann Kassensitzung eröffnen.
- **Befund 2:** Der Toaster erhält semantische Farbgebung für success, info, warning und
  error (die Aufrufstellen nutzen bereits die semantischen Toast-Funktionen, die Icons sind
  bereits semantisch — es fehlt nur die Farbe). Funktioniert in Light- und Dark-Theme.
- **Befund 3:** Die Varianten-Aggregation in allen Produkt-Queries sortiert deterministisch
  nach Varianten-ID (Anlage-Reihenfolge). Die Reihenfolge bleibt über Updates stabil.
- **Befund 4:** Die Druckstationen-Queries sortieren deterministisch nach Kategorie.
- **Befund 6:** Nach dem Umschalten eines Favoriten werden alle Queries invalidiert, die
  Favoriten anzeigen — die Tischauswahl-Liste und „Meine Tische". Die Änderung ist sofort
  sichtbar, ohne Reload.
- **Befund 7:** Die primäre Aktion („Bestellung aufnehmen" bzw. „Kassieren") wird als
  Sticky-Leiste am unteren Viewport-Rand platziert — oberhalb der auf Mobilgeräten bereits
  fixierten Tab-Leiste. Die Leiste zeigt Positionsanzahl und Summe und öffnet wie bisher den
  Review-Drawer. Die Servicekraft erreicht die Aktion jederzeit mit dem Daumen, ohne zu
  scrollen.
- **Befund 8:** 0 € ist ein gültiger Anfangsbestand (Wechselgeld kann später per Geldtransit
  eingelegt werden) und ein gültiger Kassensturz-Ist-Bestand. Das Validierungsmuster wird so
  korrigiert, dass es „Feld fehlt" von „Wert ist 0" unterscheidet; fehlende und negative
  Beträge werden weiterhin mit 400 abgelehnt.

## User Stories

### Befund 1 — Betreiber nicht konfiguriert

1. Als Vereins-Admin möchte ich beim Eröffnen der ersten Kassensitzung ohne gepflegte
   Betreiber-Stammdaten eine verständliche Fehlermeldung sehen, die mir sagt, dass ich
   zuerst die Betreiber-Stammdaten in den Einstellungen pflegen muss, damit ich weiß, wie
   ich fortfahren kann.
2. Als Vereins-Admin möchte ich, dass ein erwartbarer Konfigurationszustand nicht als
   „Interner Fehler" erscheint, damit ich jotti beim ersten Einrichten nicht für defekt
   halte.
3. Als Entwickler möchte ich, dass fehlende Betreiber-Stammdaten als Client-Fehler mit dem
   Fehlercode `betreiber_nicht_konfiguriert` beantwortet werden, damit das Frontend gezielt
   darauf reagieren kann.
4. Als Betreiber möchte ich, dass im Backend-Log nur echte Datenbankfehler als Error
   erscheinen, damit das Log beim Troubleshooting aussagekräftig bleibt.

### Befund 2 — Semantische Toast-Farben

5. Als Servicekraft möchte ich Erfolgsmeldungen grün und Fehlermeldungen rot sehen, damit
   ich im hektischen Servicebetrieb auf einen Blick erkenne, ob meine Aktion geklappt hat.
6. Als Vereins-Admin möchte ich Warnungen und Hinweise farblich von Erfolg und Fehler
   unterscheiden können, damit ich die Dringlichkeit einer Meldung sofort einordnen kann.
7. Als Nutzer möchte ich, dass die Toast-Farben sowohl im Light- als auch im Dark-Theme gut
   lesbar sind, damit die Meldungen unabhängig von meiner Theme-Wahl funktionieren.

### Befund 3 — Stabile Produktvarianten-Reihenfolge

8. Als Vereins-Admin möchte ich, dass Produktvarianten ihre Reihenfolge behalten, wenn ich
   eine Variante aktiviere, deaktiviere oder bearbeite, damit ich mich in der Liste nicht
   nach jeder Aktion neu orientieren muss.
9. Als Servicekraft möchte ich, dass die Varianten eines Produkts in der Bestell-Ansicht
   immer in derselben Reihenfolge stehen, damit ich Produkte aus dem Muskelgedächtnis
   antippen kann.
10. Als Entwickler möchte ich, dass die Varianten-Reihenfolge vom Backend deterministisch
    geliefert wird, damit kein Frontend-Workaround nötig ist (Backend ist Single Source of
    Truth für Aufbereitung).

### Befund 4 — Stabile Druckstationen-Reihenfolge

11. Als Vereins-Admin möchte ich, dass die Druckstationen nach dem Speichern einer Station
    an derselben Position stehen bleiben, damit ich beim Konfigurieren mehrerer Stationen
    nicht den Überblick verliere.
12. Als Vereins-Admin möchte ich die Druckstationen in einer nachvollziehbaren, gleichbleibenden
    Reihenfolge (nach Kategorie) sehen, damit ich eine bestimmte Station schnell finde.

### Befund 6 — Favoriten sofort sichtbar

13. Als Servicekraft möchte ich, dass ein als Favorit markierter Tisch sofort unter „Meine
    Tische" erscheint, damit ich direkt mit ihm arbeiten kann, ohne die Seite neu zu laden.
14. Als Servicekraft möchte ich, dass ein entfernter Favorit sofort aus „Meine Tische"
    verschwindet, damit die Liste immer meinem aktuellen Arbeitsbereich entspricht.
15. Als Servicekraft möchte ich, dass der Favoriten-Stern in der Tischauswahl und die Liste
    „Meine Tische" nie widersprüchliche Zustände zeigen, damit ich der Anzeige vertrauen
    kann.

### Befund 7 — Sticky-Aktionsleiste auf der Tischseite

16. Als Servicekraft möchte ich beim Bestellen die Aktion „Bestellung aufnehmen" jederzeit
    am unteren Bildschirmrand erreichen, damit ich nach dem Auswählen der Produkte nicht
    nach oben scrollen muss.
17. Als Servicekraft möchte ich beim Kassieren die Aktion ebenso jederzeit am unteren
    Bildschirmrand erreichen, damit der Bezahlvorgang am Tisch schnell geht.
18. Als Servicekraft möchte ich in der Sticky-Leiste die Anzahl der gewählten Positionen und
    die Summe sehen, damit ich den Stand der Bestellung bzw. Zahlung im Blick habe, während
    ich durch die Liste scrolle.
19. Als Servicekraft möchte ich vor dem Absenden weiterhin den Review-Schritt (Drawer mit
    Bon-Vorschau) sehen, damit ich Eingabefehler vor dem Buchen erkenne.
20. Als Servicekraft möchte ich, dass die Sticky-Leiste auf Mobilgeräten die fixierte
    Tab-Leiste nicht verdeckt und kein Listeninhalt hinter den Leisten verschwindet, damit
    alle Bedienelemente erreichbar bleiben.

### Befund 8 — 0-Cent-Beträge gültig

21. Als Vereins-Admin möchte ich eine Kassensitzung mit 0 € Anfangsbestand eröffnen können,
    weil Wechselgeld auch nachträglich per Geldtransit eingelegt werden kann.
22. Als Vereins-Admin möchte ich einen Kassensturz mit 0 € Ist-Bestand erfassen können, weil
    eine leere Kasse ein gültiges Zählergebnis ist.
23. Als Nutzer möchte ich bei tatsächlich fehlenden oder negativen Beträgen weiterhin eine
    klare Validierungsfehlermeldung bekommen, damit echte Eingabefehler nicht durchrutschen.
24. Als Entwickler möchte ich ein Validierungsmuster, das Zero-Values von fehlenden Feldern
    unterscheidet, damit dieselbe zog-Falle nicht bei künftigen Betragsfeldern erneut
    auftritt.

## Implementation Decisions

### Befund 1 — Fehler-Mapping Betreiber-Stammdaten

- Im Kasse-Command wird beim Eröffnen der Kassensitzung der „nicht gefunden"-Fall des
  Settings-Repositories (Betreiber-Stammdaten nie gespeichert) als „Betreiber nicht
  konfiguriert" interpretiert — derselbe Fehler, der bereits für unvollständig gepflegte
  Stammdaten existiert. Nur echte Datenbankfehler bleiben Server-Fehler (500).
- Der HTTP-Handler mappt diesen Fehler bereits auf den Fehlercode
  `betreiber_nicht_konfiguriert` (Client-Fehler); am Handler ist keine Änderung nötig.
- Das Frontend hinterlegt auf der Kassensitzungs-Seite für diesen Fehlercode eine
  deutschsprachige Meldung mit Handlungsanweisung (Betreiber-Stammdaten in den Einstellungen
  pflegen) über den bestehenden `byCode`-Mechanismus der Submit-Hooks.

### Befund 2 — Toast-Farbgebung

- Der zentrale Toaster (sonner, shadcn/ui-Wrapper) erhält semantische Farbgebung für die
  Varianten success, info, warning und error — über die von sonner vorgesehene
  Theming-Möglichkeit (richColors bzw. CSS-Variablen), abgestimmt auf Light- und Dark-Theme.
- Die Aufrufstellen bleiben unverändert: Sie verwenden bereits durchgängig die semantischen
  Funktionen (`toast.success`, `toast.error`); neutrale Aufrufe existieren nicht.

### Befund 3 — Deterministische Varianten-Sortierung

- Alle Produkt-Queries, die Varianten per `json_agg` aggregieren (Einzelprodukt, alle
  Produkte, aktive Produkte), sortieren innerhalb des Aggregats nach Varianten-ID
  (Anlage-Reihenfolge).
- Die Sortierung erfolgt ausschließlich im Backend (SQL); kein Frontend-Sortier-Workaround.
- Nach der Query-Änderung wird der sqlc-Code neu generiert.

### Befund 4 — Deterministische Druckstationen-Sortierung

- Die Druckstationen-Queries (alle und konfigurierte) sortieren nach Kategorie.
- Die Kategorie ist der fachliche Schlüssel der Druckstation (Unique-Constraint) und damit
  die natürliche, stabile Ordnung.

### Befund 6 — Query-Invalidierung Favoriten

- Die Favoriten-Mutation im Tischauswahl-Drawer invalidiert nach Erfolg zusätzlich zur Query
  der Tischauswahl-Liste auch die Query hinter „Meine Tische" — dem bestehenden
  Invalidierungs-Muster der Settings-Hooks folgend.

### Befund 7 — Sticky-Aktionsleiste

- Auf der Tischseite wird in den Tabs „Bestellen" und „Kassieren" die primäre Aktion als
  Sticky-Leiste am unteren Viewport-Rand platziert, auf Mobilgeräten oberhalb der bereits
  fixierten Tab-Leiste. Desktop verwendet dasselbe Muster (konsistent, kein Sonderfall).
- Die Leiste zeigt Positionsanzahl und Summe der aktuellen Auswahl und ist deaktiviert,
  solange keine Position gewählt ist (bestehendes Verhalten des Buttons).
- Die Leiste ersetzt den bisherigen Button oberhalb der Liste; sie öffnet wie bisher den
  jeweiligen Review-Drawer (Bestellung bzw. Zahlung). Am Drawer selbst ändert sich nichts.
- Der Listeninhalt erhält ausreichend Abstand nach unten (analog zum bestehenden Padding für
  die fixierte Tab-Leiste), damit keine Einträge hinter den Leisten verschwinden.

### Befund 8 — Zero-Value-sichere Betragsvalidierung

- Fachliche Entscheidung: 0 Cent ist gültig für den Anfangsbestand der
  Kassensitzungs-Eröffnung und für den Ist-Bestand des Kassensturzes.
- Technische Entscheidung: Betragsfelder, die 0 zulassen (Validierungsregel „größer oder
  gleich 0"), werden in den Request-DTOs als Pointer-Felder modelliert. zogs `Required()`
  prüft dann die Anwesenheit des Feldes statt des Nicht-Zero-Werts — der von zog
  dokumentierte Weg für dieses Problem. Fehlende Felder und negative Werte führen weiterhin
  zu 400 mit Feld-Fehlermeldung.
- Felder mit Regel „größer oder gleich 1" (z. B. Geldtransit-Betrag, Mengen) sind von der
  zog-Falle nicht betroffen und bleiben unverändert.
- Das Frontend lässt 0 € bereits zu; dort ist keine Änderung nötig.

## Testing Decisions

- **Was ein guter Test ist:** Tests prüfen ausschließlich von außen beobachtbares Verhalten
  — HTTP-Status und Fehlercodes, Reihenfolge gelieferter Daten, sichtbare UI-Zustände. Keine
  Tests auf Implementierungsdetails (interne Funktionsaufrufe, CSS-Klassennamen,
  Query-Interna).
- **Backend (Prior Art: Command- und Handler-Tests im Kasse-Kontext):**
  - Befund 1: Command-Test — „Betreiber nicht gefunden" liefert den Fehler „Betreiber nicht
    konfiguriert", nicht den Datenbankfehler. Handler-Test — Response ist Client-Fehler mit
    Code `betreiber_nicht_konfiguriert`.
  - Befund 8: Handler-Tests — Eröffnung mit 0 Cent und Kassensturz mit 0 Cent werden
    akzeptiert; fehlendes Betragsfeld und negativer Betrag werden mit 400 abgelehnt.
- **Backend-Repository-Integrationstests (Prior Art: bestehende Repo-Tests gegen die
  Dev-Datenbank):**
  - Befund 3: Varianten-Reihenfolge bleibt nach einem Varianten-Update stabil (nach ID).
  - Befund 4: Druckstationen-Reihenfolge bleibt nach einem Upsert stabil (nach Kategorie).
- **Frontend (Prior Art: Komponententests im Service-Bereich, Hook-Tests):**
  - Befund 6: Komponententest — nach dem Favoriten-Toggle werden beide betroffenen Queries
    invalidiert bzw. die Liste „Meine Tische" zeigt den neuen Stand.
  - Befund 7: Komponententest — die Aktionsleiste zeigt Positionsanzahl und Summe und ist
    ohne Auswahl deaktiviert.
- **Nicht automatisiert getestet:** Toast-Farben (Befund 2, rein visuell — manuelle Prüfung
  in beiden Themes) und das Sticky-Positionierungsverhalten selbst (visuell, mobile
  Viewports).

## Out of Scope

- **Befund 5 (Direktverkauf-Prominenz / Modus-Trennung):** Die Trennung von Tischservice und
  Direktverkauf als Arbeitsmodi ist ein UX-Rework mit eigenen Entwurfsentscheidungen
  (Modus-Wahl, Persistenz, Navigation) und wird in einem eigenen PRD behandelt — analog zum
  TSE-Setup-Wizard. Der Direktverkauf-Button auf der Tischauswahl-Seite bleibt bis dahin
  unverändert.
- Ein darüber hinausgehendes Redesign der Tischseite (Tab-Struktur, Produktlisten-Layout).
- Ein inhaltliches Audit aller Toast-Meldungen (Texte, Semantik der Aufrufstellen).
- Änderungen an Betragsfeldern mit Mindestwert 1 Cent oder an der Frontend-Validierung.
- Migration alter Daten oder API-Versionierung (Pre-Release, Breaking Changes erlaubt).

## Further Notes

- Die Befunde sind unabhängig; der Implementierungsplan soll je Befund eine eigene Phase mit
  eigenen Akzeptanzkriterien bilden. Die Reihenfolge ist frei wählbar; die reinen
  Backend-Fixes (Befunde 1, 3, 4, 8) sind klein und eignen sich als erste Slices.
- `docs/qa-notes.md` ist mit diesem PRD vollständig überführt (7 Befunde hier, Befund 5 ins
  Folge-PRD) und kann nach Abschluss der Umsetzung gelöscht werden — die Git-Historie
  bewahrt die Notizen.
- Befund 8 dokumentiert eine generelle zog-Falle (Zero-Values bei `Required()`). Das
  Pointer-Muster aus diesem PRD ist die Referenz für künftige Felder, bei denen 0 ein
  gültiger Wert ist.
