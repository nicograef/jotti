# PRD: Bestellung umbuchen (K-09)

## Problem Statement

Beim Aufnehmen einer Bestellung wählt die Servicekraft zuerst den Tisch und stellt
dann die Positionen zusammen. Im Trubel eines Vereinsfests passiert der häufigste
Eingabefehler genau hier: Die Bestellung landet auf dem **falschen Tisch** — „Tisch 5"
statt „Tisch 15" angetippt, oder zwei nebeneinanderliegende Tische verwechselt. Heute
gibt es dafür nur einen umständlichen Workaround: Die Serviceleitung muss am falschen
Tisch jede Position einzeln **stornieren** (mit Pflichtgrund) und die komplette
Bestellung am richtigen Tisch **neu aufnehmen**. Das sind zwei getrennte, manuelle
Vorgänge an zwei Tischen, fehleranfällig (man vergisst eine Position, tippt einen
falschen Preis, wählt erneut den falschen Tisch) und langsam — während Gäste warten.

Es fehlt eine **gezielte Korrektur**, die eine bereits aufgenommene, **noch nicht
bezahlte** Bestellung in einem Schritt vom Quell- auf den Ziel-Tisch verschiebt, ohne
dass dabei Geld bewegt wird oder die Kassenführung durcheinandergerät.

## Solution

Serviceleitung und Admin erhalten die Aktion **„Umbuchen"** direkt an jeder Bestellung
in der Tisch-Historie — gleichberechtigt neben der bestehenden **„Stornieren"**-Aktion
und mit derselben Bedienlogik. Der Ablauf aus Sicht der Serviceleitung:

1. In der Historie des **Quell-Tischs** die betroffene Bestellung öffnen und „Umbuchen"
   wählen.
2. Es werden ausschließlich die **noch nicht bezahlten** Positionen dieser Bestellung
   angeboten (mit Mengenauswahl, standardmäßig alle vorausgewählt).
3. Den **Ziel-Tisch** aus den aktiven Tischen wählen (der Quell-Tisch ist
   ausgeschlossen).
4. Bestätigen — fertig.

Im Hintergrund erzeugt jotti **atomar** (eine Transaktion) genau zwei Vorgänge im
Kassenjournal:

- eine **Stornierung** der gewählten Positionen am **Quell-Tisch**, automatisch
  kommentiert mit „Umbuchung auf Tisch …",
- eine **neue Bestellung** mit denselben Positionen, Mengen und **Original-Preisen** am
  **Ziel-Tisch**, automatisch kommentiert mit „Umbuchung von Tisch …".

Die Umbuchung ist auf **unbezahlte** Positionen beschränkt. Dadurch ist sie immer
**bargeld-neutral**: Eine Tisch-Stornierung und eine Tisch-Bestellung verändern nur den
jeweiligen Tisch-Saldo, **nie** die Kassenlade (die `GetKassenbestand`-Berechnung kennt
weder Bestellung noch Tisch-Storno). Quell- und Ziel-Tisch bleiben in sich konsistent,
der Tagesabschluss (alle Tische Saldo 0) wird nicht blockiert, und es entstehen keine
Phantom-Bargeldbewegungen. Beide Tische tragen den Vorgang revisionssicher und
nachvollziehbar in ihrer Historie.

Bewusst **kein** erneuter Arbeitsbon-Druck am Ziel-Tisch (K-12): Die Umbuchung ist eine
**abrechnungsseitige Korrektur**, keine neue Zubereitungsanforderung. Der ursprüngliche
Arbeitsbon wurde bei der ersten Bestellaufnahme bereits gedruckt; ein zweiter Bon würde
in der Küche/Theke eine Doppel-Zubereitung auslösen.

### Ubiquitous Language (neue/relevante Begriffe)

| Begriff         | Bedeutung                                                                                         |
| --------------- | ------------------------------------------------------------------------------------------------- |
| **Umbuchung**   | Geschäftsvorfall-Paar: atomare Stornierung am Quell-Tisch + neue Bestellung am Ziel-Tisch (K-09). |
| **Quell-Tisch** | Tisch, an dem die fehlerhaft erfasste Bestellung aktuell liegt (wird storniert).                  |
| **Ziel-Tisch**  | Tisch, auf den die Positionen umgebucht werden (erhält die neue Bestellung).                      |

## User Stories

1. Als Serviceleitung möchte ich an einer Bestellung in der Tisch-Historie eine Aktion
   „Umbuchen" sehen, damit ich eine falsch erfasste Bestellung korrigieren kann.
2. Als Admin möchte ich dieselbe „Umbuchen"-Aktion nutzen können, da ich vollen
   Kassenzugriff habe.
3. Als Servicekraft (Rolle `service`) möchte ich die „Umbuchen"-Aktion **nicht** sehen,
   damit Korrekturen mit Saldo-Wirkung der Serviceleitung/Admin vorbehalten bleiben
   (gleiche Berechtigung wie Stornierung).
4. Als Serviceleitung möchte ich beim Umbuchen nur die **noch nicht bezahlten**
   Positionen der Bestellung angeboten bekommen, damit ich keine bereits kassierten
   Positionen verschiebe und so die Kassenführung durcheinanderbringe.
5. Als Serviceleitung möchte ich die **Mengen** je Position auswählen können (Default:
   alle umbuchbaren Positionen), damit ich auch fehl-gruppierte Bestellungen (z. B.
   2 von 4 Bier auf einen anderen Tisch) teilweise umbuchen kann.
6. Als Serviceleitung möchte ich den **Ziel-Tisch** aus den aktiven Tischen auswählen,
   wobei der Quell-Tisch nicht auswählbar ist, damit ich nicht versehentlich auf
   denselben Tisch umbuche.
7. Als Serviceleitung möchte ich, dass die Umbuchung **atomar** ausgeführt wird, damit
   niemals nur die Stornierung am Quell-Tisch **oder** nur die neue Bestellung am
   Ziel-Tisch entsteht (kein halber Zustand).
8. Als Serviceleitung möchte ich, dass die neue Bestellung am Ziel-Tisch die
   **Original-Preise** der Positionen übernimmt, damit eine zwischenzeitliche
   Preisänderung im Katalog den umgebuchten Betrag nicht verfälscht.
9. Als Serviceleitung möchte ich, dass am Quell-Tisch eine **Stornierung** mit dem
   automatischen Kommentar „Umbuchung auf Tisch &lt;Ziel&gt;" erscheint, damit in der
   Historie sofort erkennbar ist, warum und wohin storniert wurde.
10. Als Serviceleitung möchte ich, dass am Ziel-Tisch eine **Bestellung** mit dem
    automatischen Kommentar „Umbuchung von Tisch &lt;Quell&gt;" erscheint, damit die
    Herkunft der Positionen nachvollziehbar ist.
11. Als Serviceleitung möchte ich **keinen** Pflicht-Grund eintippen müssen, da der
    Ziel-Tisch die Korrektur bereits erklärt (Unterschied zur normalen Stornierung).
12. Als Serviceleitung möchte ich, dass die Umbuchung die **Kassenlade nicht verändert**
    (bargeld-neutral), damit kein Kassensturz-Unterschied und keine erfundene Auszahlung
    entsteht.
13. Als Admin möchte ich, dass eine Umbuchung den **Tagesabschluss nicht blockiert** —
    Quell- und Ziel-Tisch bleiben nach der Umbuchung jeweils auf einem korrekten,
    erfüllbaren Saldo.
14. Als Serviceleitung möchte ich, dass am Ziel-Tisch die umgebuchten Positionen als
    **offen/unbezahlt** und **ausstehend** geführt werden, damit ich sie dort regulär
    kassieren und die Ausgabe bestätigen kann.
15. Als Serviceleitung möchte ich, dass am Ziel-Tisch **kein** neuer Arbeitsbon gedruckt
    wird, damit Küche/Theke die bereits angeforderte Ware nicht doppelt zubereiten.
16. Als Serviceleitung möchte ich eine klare Fehlermeldung erhalten, wenn eine Position
    zwischenzeitlich (durch parallele Zahlung oder Stornierung) **nicht mehr umbuchbar**
    ist, damit ich meine Auswahl aktualisieren kann.
17. Als Serviceleitung möchte ich bei einem **Nebenläufigkeitskonflikt** (paralleler
    Schreibvorgang an Quell- oder Ziel-Tisch) eine verständliche Konflikt-Rückmeldung
    erhalten und den Vorgang erneut auslösen können, ohne dass ein inkonsistenter
    Zustand zurückbleibt.
18. Als Serviceleitung möchte ich nur umbuchen können, wenn der **Ziel-Tisch aktiv**
    ist, damit keine Positionen auf einen deaktivierten/gelöschten Tisch wandern.
19. Als Serviceleitung möchte ich nur umbuchen können, wenn eine **Kassensitzung offen**
    ist, da jeder Geschäftsvorfall einer offenen Kassensitzung zugeordnet sein muss.
20. Als Serviceleitung möchte ich, dass die „Umbuchen"-Aktion **nicht angeboten** wird,
    wenn die Bestellung keine umbuchbaren (unbezahlten) Positionen mehr enthält, damit
    ich keine leere Auswahl öffne.
21. Als Admin/Prüfer möchte ich, dass beide Vorgänge (Storno + Bestellung) als
    **eigenständige, unveränderliche Events** im Kassenjournal liegen und (perspektivisch)
    je eine TSE-Transaktion bzw. DSFinV-K-Zeile bilden, damit die Umbuchung
    revisionssicher bleibt.
22. Als Serviceleitung möchte ich nach erfolgreicher Umbuchung eine Bestätigung sehen und
    die aktualisierte Historie/den aktualisierten Saldo des Quell-Tischs erhalten, damit
    ich den Erfolg sofort erkenne.
23. Als Gast möchte ich, dass meine bereits an meinem (richtigen) Tisch erfassten und
    bezahlten Positionen **unangetastet** bleiben, wenn am Nachbartisch eine Umbuchung
    passiert (positionsgenaue Wirkung, keine Seiteneffekte).

## Implementation Decisions

### Bereich & Berechtigung

- Neuer Endpunkt **`POST /serviceleitung/bestellung-umbuchen`**, registriert in
  `api/serviceleitung.go` neben `stornierung-erteilen` — damit ausschließlich
  `serviceleitung` und `admin` (entsprechend `handbuch.md` §3.12 und der
  Stornierungsberechtigung K-04).
- Request-Vertrag: `{ "quellTischId": int, "zielTischId": int, "positionen":
[{ "positionId": uuid, "menge": int }] }`. **Kein** `kommentar`-Feld (serverseitig
  automatisch gesetzt). Antwort: leerer Erfolg (wie Stornierung).

### Deep Module: atomarer Cross-Aggregat-Write (Kern)

- Neue Repository-Methode **`WriteUmbuchung`** in `kassenjournal_repo`. Sie nimmt zwei
  bereits versionierte Tisch-Session-Events entgegen (Storno für den Quell-Subject,
  Bestellung für den Ziel-Subject) sowie die `kassensitzungNr` und schreibt in **einer**
  Transaktion: beide Events ins `kassenjournal` **und** beide `tisch_sessions`-
  Projektionen (Quelle und Ziel). Sie wiederverwendet die vorhandene interne
  `writeEventInTx`-Logik je Event (beide `StreamTypeTischSession`).
- **Atomarität & Rollback:** Schlägt einer der beiden Schritte fehl, wird die gesamte
  Transaktion zurückgerollt — es gibt nie nur die Storno-Seite oder nur die
  Bestell-Seite. Dies ist die in `handbuch.md` §3.12 geforderte
  „Cross-Aggregat-Transaktion … Atomarität auf Anwendungsebene sicherstellen".
- **OCC:** Die Anwendungsschicht ermittelt für **beide** Subjects via `GetMaxVersion`
  die nächste Version und setzt sie auf den Events (analog zum bestehenden
  `writeEventOCC`). Verletzt einer der beiden `UNIQUE(subject, version)` (paralleler
  Schreibvorgang an einem der Tische), rollt die Transaktion zurück und der Konflikt
  wird als `ErrConflict` (HTTP-Konflikt) abgebildet.
- Der `kassenjournal_repo`-Mock wird um `WriteUmbuchung` erweitert (für Unit-Tests der
  Anwendungsschicht).

### Orchestrierung: `application.BestellungUmbuchen`

- Neue Command-Methode in `api/table/application`. Ablauf: offene Kassensitzung prüfen
  → **Gleicher-Tisch-Guard** (`quellTischId != zielTischId`) → Quell-Tisch laden
  (Existenz/Status, Tisch-Session-State) und Ziel-Tisch laden (Existenz/Status, Subject)
  → angeforderte `PositionRef`s gegen die **`UnbezahltePositionen`** des Quell-Tischs
  validieren (Eligibility: nur unbezahlt) → Positionen zu vollständigen (fat) Positionen
  auflösen → Storno-Event (Quelle) und Bestellung-Event (Ziel) bauen → beide versionieren
  → `WriteUmbuchung` aufrufen.
- **Eligibility = unbezahlt:** Validierung gegen `state.UnbezahltePositionen` (nicht gegen
  `nichtStorniert` wie bei der Stornierung). Begründung und Konsequenzen siehe „Further
  Notes". Auslieferungsstatus ist irrelevant.
- **Wiederverwendung bestehender Events, keine neuen Event-Typen:** Storno über
  `NewStornierungErteiltEvent`, neue Bestellung über `NewBestellungAufgenommenEvent`. Da
  beide bereits bestehende, projektions- und reporting-wirksame Events sind, funktionieren
  Saldo, Kassenbestand, Reporting und Tagesabschluss ohne Anpassung.
- **Werterhaltung / keine Neu-Bepreisung:** Die Ziel-Positionen werden direkt aus den
  aufgelösten Quell-Positionen gebildet (`ProduktName`, `VarianteName`, `Kategorie`,
  `Einzelpreis`, `VarianteID`, `Menge`). `NewBestellungAufgenommenEvent` vergibt dabei
  **frische** `PositionID`s und berechnet die Summe aus den **Original-Einzelpreisen** neu.
  Es findet **kein** erneuter Produkt-/Varianten-Lookup statt — der umgebuchte Betrag ist
  exakt der stornierte Betrag.
- **Auto-Kommentar:** Storno-Kommentar „Umbuchung auf Tisch &lt;Ziel-Tischname&gt;"
  (erfüllt das Storno-Schema `Min(3).Max(100)`), Bestellung-Kommentar „Umbuchung von
  Tisch &lt;Quell-Tischname&gt;". Die Tischnamen stammen aus den ohnehin geladenen
  Tisch-Entitäten.
- **Kein Arbeitsbon:** `WriteUmbuchung` nutzt den schlichten Event-Write-Pfad (nicht den
  `WriteEventWithDruckauftraege`-Outbox-Pfad). Am Ziel-Tisch wird damit **kein**
  Druckauftrag erzeugt.
- Neue Anwendungsfehler: `ErrPositionNichtUmbuchbar` (→ `position_nicht_umbuchbar`) und
  `ErrUmbuchungGleicherTisch` (→ `umbuchung_gleicher_tisch`). Wiederverwendet:
  `ErrConflict`, `ErrKasseNichtGeoeffnet`, `ErrTischNotFound`, `ErrTischNotActive`.

### HTTP-Schicht

- Neuer `BestellungUmbuchenHandler` in `api/table/http` mit Request-DTO und
  zog-Validierung (analog `stornierungErteilenSchema`, jedoch `quellTischId` +
  `zielTischId`, ohne `kommentar`). Die Methode wird dem `command`-Interface des Handlers
  hinzugefügt. Fehler-Mapping deckt alle oben genannten Fälle ab.

### Frontend

- `TischBackend.bestellungUmbuchen(cmd)` postet validiert (Zod-Schema
  `BestellungUmbuchenSchema`: `quellTischId`, `zielTischId`, `positionen`) an
  `serviceleitung/bestellung-umbuchen` über den `BackendClient` (kein direktes `fetch`).
- Neue Komponente **`HistorieUmbuchungDrawer`** (Spiegel von `HistorieStornierungDrawer`):
  Positions-/Mengenauswahl der umbuchbaren Positionen **plus** Ziel-Tisch-Auswahl,
  **ohne** Kommentarfeld. Die Liste der aktiven Tische (ohne Quell-Tisch) wird über die
  bestehende Aktive-Tische-Abfrage bereitgestellt.
- `TischHistorie` erhält je Bestellung eine **„Umbuchen"**-Aktion neben „Stornieren",
  sichtbar nur bei `AuthSingleton.canCancel` **und** wenn umbuchbare Positionen vorhanden
  sind. Neuer Helper **`getUmbuchbarePositionen(bestellung, historie)`** (analog
  `getStornierbarePositionen`, zieht je `PositionID` jedoch **storniert _und_ bezahlt** ab).
- Erfolgs-/Fehlerbehandlung über das bestehende `useActionSubmit`-Muster; der Fehlercode
  `position_nicht_umbuchbar` erhält eine verständliche deutsche Meldung
  (Auswahl aktualisieren), `umbuchung_gleicher_tisch` wird durch die UI ohnehin verhindert.

## Testing Decisions

**Was einen guten Test ausmacht:** Getestet wird **beobachtbares Außenverhalten**, nicht
die interne Umsetzung — also welche Events/Projektionen nach einer Operation existieren,
welche Salden sich ergeben und welche Fehler bei ungültigen Eingaben entstehen. Keine
Tests gegen private Hilfsfunktionen oder konkrete SQL-Formulierungen.

Getestet werden (vom Nutzer bestätigt):

1. **`WriteUmbuchung` — Integrationstest** (`//go:build integration`,
   `kassenjournal_repo/repo_test.go`):
   - **Commit:** Nach erfolgreicher Umbuchung existieren am Quell-Subject das
     Storno-Event und am Ziel-Subject das Bestellung-Event; beide `tisch_sessions`-
     Projektionen zeigen die korrekten Salden (Quelle reduziert, Ziel erhöht) und
     Positionslisten.
   - **Rollback/Atomarität:** Schlägt der Ziel-Write fehl (z. B. ungültige
     Version/forcierter Fehler), bleibt **auch** der Quell-Storno aus — kein halber
     Zustand.
   - **OCC-Konflikt:** Eine kollidierende Version an einem der beiden Subjects führt zum
     vollständigen Rollback.
   - Prior art: `TestWriteEventWithDruckauftraege_CommitsEventAndAuftrag`,
     `TestWriteEventWithDruckauftraege_RollsBackEventOnAuftragError`,
     `TestWriteEvent_MultipleEvents_ProjectionCorrect`, `TestWriteEvent_InvalidData_Rollback`.

2. **`BestellungUmbuchen` — Unit-Test** (`//go:build unit`,
   `api/table/application/command_test.go`, mock-basiert):
   - Happy Path: korrekte Storno- und Bestellung-Events mit richtigem Quell-/Ziel-Subject,
     übernommenen Positionen/Mengen/Preisen, frischen Ziel-`PositionID`s und den
     Auto-Kommentaren.
   - **Eligibility:** Anforderung einer bereits bezahlten Position → `ErrPositionNichtUmbuchbar`.
   - **Gleicher-Tisch-Guard:** `quellTischId == zielTischId` → `ErrUmbuchungGleicherTisch`.
   - Ziel-Tisch inaktiv/unbekannt → `ErrTischNotActive` / `ErrTischNotFound`.
   - Keine offene Kassensitzung → `ErrKasseNichtGeoeffnet`.
   - Versionskonflikt aus dem Repo → `ErrConflict`.
   - Prior art: bestehende Stornierungs-/Bestellungs-Tests in derselben Datei.

3. **`getUmbuchbarePositionen` — Frontend-Helper-Unit-Test** (Vitest):
   - Nur unbezahlte, nicht-stornierte Positionen je Bestellung werden zurückgegeben;
     korrekte Restmengen nach Teil-Zahlung und Teil-Stornierung; leere Liste, wenn alles
     bezahlt/storniert.
   - Prior art: `frontend/src/service/components/table/drawerUtils.test.ts`.

**Nicht automatisiert getestet** (bewusst, vom Nutzer so gewählt): Drawer-/Interaktions-
und Komponententests für `HistorieUmbuchungDrawer`.

## Out of Scope

- **Umbuchen bereits bezahlter Positionen.** Bewusst ausgeschlossen (siehe „Further
  Notes"). Der seltene „erst nach der Zahlung bemerkt"-Fall wird mit den bestehenden
  Funktionen gelöst: Stornierung + **Auszahlung (echte Erstattung)** am Quell-Tisch sowie
  neue Bestellung + **Zahlung** am Ziel-Tisch — mit realer Bargeldbewegung.
- **Umbuchen über Kassensitzungen hinweg.** Quell- und Ziel-Tisch gehören immer zur
  aktuell offenen Kassensitzung (Subjects sind sitzungsskopiert).
- **Erneuter Arbeitsbon-/Belegdruck** am Ziel-Tisch.
- **Explizite Verknüpfungsfelder** in den Event-Daten (z. B. `umbuchungId`,
  Quell-/Ziel-Referenz). Die Nachvollziehbarkeit erfolgt über die Auto-Kommentare; eine
  maschinenlesbare Verknüpfung ist nicht Teil dieses PRD.
- **Zusammenführen mehrerer Bestellungen** in einem Umbuchungsvorgang. Die Umbuchung
  geht stets von **einer** Bestellung aus (mehrere nacheinander sind möglich).
- **Verschieben von Zahlungen oder Auszahlungen** zwischen Tischen.

## Further Notes

### Warum nur unbezahlte Positionen (Korrektheit & Compliance)

Eine Tisch-**Stornierung** und eine Tisch-**Bestellung** verändern ausschließlich den
jeweiligen **Tisch-Saldo**, **nie** die Kassenlade. Verifiziert in
`GetKassenbestand` (`backend/sqlc/queries/kassensitzungen.sql`): Der Soll-Bestand summiert
nur `kassensitzung-eroeffnet`, `zahlung-kassiert`, `auszahlung-geleistet`,
`direktverkauf-getaetigt/-storniert`, `geldtransit-gebucht` und
`differenz-soll-ist-gebucht` — **weder** Bestellung **noch** Tisch-Storno.

Daraus folgt: Würde man eine **bezahlte** Position umbuchen (nur Storno + Bestellung),
ergäbe sich am Quell-Tisch ein **negativer** Saldo („wir schulden dem Gast"), am
Ziel-Tisch ein **offener** Saldo („Gast schuldet erneut") — obwohl das Geld längst
korrekt in der Lade liegt. Um beide Tische für den Tagesabschluss (alle Tische Saldo 0)
zu schließen, wären eine **Auszahlung** (Quelle) und eine **Zahlung** (Ziel) nötig, die
**physisch nie stattgefunden haben** → Kassendifferenz und eine unrichtige Bargeldspur.
Genau diese Bargeld-Genauigkeit ist das Schutzziel von KassenSichV/GoBD.

Wichtig zur Einordnung: Die reine **Aufzeichnung** wäre auch dann konform (jeder
Geschäftsvorfall bleibt ein unveränderliches, einzeln nachvollziehbares Event). Das
Problem ist nicht Datenverlust, sondern die **erzwungene Phantom-Bargeldbewegung**. Die
Beschränkung auf unbezahlte Positionen hält die Umbuchung bargeld-neutral, beide Tische
konsistent und den Tagesabschluss erfüllbar — die einfache, korrekte und konforme Lösung.

### Verhältnis zur Stornierung

Die Umbuchung ist **keine** Erweiterung der Stornierung, sondern ein eigener
Geschäftsvorfall, der zwei bestehende Events atomar koppelt. Die Stornierung bleibt für
echte Stornos (Position fällt weg) erhalten und arbeitet weiterhin auf **allen
nicht-stornierten** Positionen (inkl. bezahlter), während die Umbuchung bewusst auf
**unbezahlte** Positionen begrenzt ist.

### Arbeitsbon-Restbetrachtung

Da am Ziel-Tisch kein neuer Arbeitsbon gedruckt wird, kennt die Ausgabestation den
korrigierten Tisch nicht aus einem Bon. Für jottis Zielbetrieb (kleine Feste, direkte
Absprache im Team) ist das akzeptabel und vermeidet Doppel-Zubereitung; der ursprüngliche
Arbeitsbon trug ohnehin bereits den (falschen) Quell-Tisch. Sollte sich später ein Bedarf
für einen „Umbuchungs-Hinweisbon" zeigen, kann das separat ergänzt werden.

### Anknüpfung Compliance (perspektivisch)

Beide erzeugten Events sind reguläre Kasse-Events und fügen sich in das künftige
TSE-Mapping (`handbuch.md` §3.13) und den DSFinV-K-Export ein: Storno und Bestellung
bilden je eine eigene Transaktion bzw. Export-Zeile, jeweils ihrem Abrechnungskreis
(Tisch-Session) zugeordnet. Es ist kein Sonderfall im Compliance-Modell nötig.
