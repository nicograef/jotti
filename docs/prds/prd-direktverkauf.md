# PRD: Direktverkauf

## Problem Statement

Bei fast jedem Vereinsfest gibt es neben den bedienten Tischen eine **Theke**, an
der Gäste direkt bestellen und **sofort bezahlen** — Getränkeausschank, Kuchen- oder
Essenstheke. Der Gast bestellt, zahlt bar, erhält (bei Bedarf) einen Bon und holt
seine Ware ab. Bestellung, Zahlung und Ausgabe fallen in **einen einzigen Vorgang**
zusammen. Es gibt keinen offenen Saldo, keine spätere separate Zahlung, keine spätere
Ausgabe-Bestätigung.

Der bediente Tisch ist das genaue Gegenteil: ein **Aggregat mit Lebenszyklus** —
offener Saldo über die gesamte Sitzdauer, Teilzahlungen, Teilausgaben, ein
Event-Stream pro Tisch mit fortlaufender Versionierung und optimistischer
Nebenläufigkeitskontrolle (OCC).

### Warum der Direktverkauf **keine** Tisch-Variante sein darf

Ein früherer Ansatz wollte den Direktverkauf als bedienten Tisch mit einem
`direktverkauf`-Häkchen modellieren und pro Barverkauf die drei bestehenden Events
(`bestellung-aufgenommen` → `zahlung-kassiert` → `ausgabe-bestaetigt`) im
Tisch-Stream erzeugen. Das presst zwei fachlich entgegengesetzte Konzepte in ein
Modell und erbt den gesamten Aggregat-Apparat des Tisches, ohne ihn zu nutzen:

1. **Semantische Verzerrung.** Das Kassenjournal soll fachliche Fakten festhalten.
   Drei im selben Augenblick erzeugte Events erzählen eine Sequenz („bestellt, dann
   später bezahlt, dann später ausgegeben"), die nie stattfand. Die Historie müsste
   anschließend künstlich wieder zu „einem Verkauf" zusammengefasst werden.
2. **OCC-Hotspot an der schnellsten Kasse.** Separate Tisch-Subjects existieren
   gerade deshalb, weil `UNIQUE(subject, version)` alle Schreibvorgänge eines Subjects
   serialisiert (siehe `docs/handbuch.md` §3.3). Eine Theke ist die
   durchsatzstärkste Kasse — als _ein_ Tisch-Subject mit _drei_ Versionen pro Verkauf
   führt jeder parallele Verkauf zu Versionskonflikten und Retry-Schleifen.
3. **Flag-Kopplung an reine Stammdaten.** Der `Tisch` ist laut `docs/language.md`
   eine „reine Stammdaten-Entität: ein physischer Ort, an dem Gäste sitzen". Ein
   Direktverkauf hat keinen Ort und keine sitzenden Gäste. Ein `direktverkauf`-Flag
   verzweigt das Verhalten an vielen Stellen und vermischt zwei Konzepte in einer
   Entität.
4. **TSE-Mapping-Problem (Compliance).** § 146a AO verlangt, **jeden
   Geschäftsvorfall einzeln** per TSE-Transaktion zu schützen. Ein Barverkauf ist
   **ein** Geschäftsvorfall — als drei Events entsteht ein unklares 3:1-Mapping auf
   die spätere TSE-Transaktion.
5. **Ungenutzter Aggregat-Apparat (YAGNI invertiert).** Unbezahlte/ausstehende
   Positionen, Teilzahlung, Teilausgabe, negativer Saldo, die `tisch_sessions`-
   Projektion — beim Direktverkauf alles bedeutungslos, aber dauerhaft mitgeschleppt.

### Kernerkenntnis

> Der Direktverkauf braucht das **Event** (für das revisionssichere, append-only
> Kassenjournal), aber **nicht das Aggregat** des Tisches (Saldo, Projektion,
> langlebiger OCC-Stream). Beides wird sauber getrennt.

## Solution

Der Direktverkauf erhält ein **eigenes, schlankes Domänenmodell**, strikt getrennt
vom bedienten Tisch — **ohne** eigene Stammdaten-Entität:

- **Genau eine, implizite Theke.** jotti modelliert den Direktverkauf bewusst als
  **einen einzigen, namenlosen Verkaufskanal** — keine `Verkaufsstelle`-Entität, kein
  CRUD, keine Stammdatenpflege. Ein Vereinsfest hat in aller Regel genau eine Theke;
  mehrere benannte Verkaufsstellen wären vorzeitige Komplexität (YAGNI). Erreichbar als
  **einzelner Menüpunkt „Direktverkauf"** im Service-Bereich.
- **Neuer Geschäftsvorfall `Direktverkauf`** als eigenes, kleines Event-Aggregat mit
  genau zwei Event-Typen: `direktverkauf-getaetigt:v1` und (positionsgenaues)
  `direktverkauf-storniert:v1`. Kein Recycling der Tisch-Events.
- **Ein eigener Event-Stream pro Verkauf** (`kassensitzung-{nr}/direktverkauf-{uuid}`).
  Jeder Verkauf braucht ohnehin eine UUID (für Historie und positionsgenauen Storno);
  sie zugleich als Subject zu nutzen kostet nichts und macht jeden Verkauf zu einem
  eigenständig ladbaren Aggregat — **kein geteilter Versionsraum, Storno als `version
2` des eigenen Streams**, strukturell konfliktfrei (analog zum eigenen Subject jedes
  Tisches).
- **Keine Saldo-Projektion.** Es gibt nichts Veränderliches zu materialisieren — kein
  Projektions-Pendant zu `tisch_sessions`. Storno-Validierung läuft per günstigem
  On-Demand-Replay des einzelnen Verkauf-Streams (wenige Events).
- **Bargeldwirksam und compliance-konform.** Der Verkauf vereinnahmt Bargeld sofort;
  ein Storno gibt es sofort zurück. Beides fließt direkt in Kassenbestand und Z-Bon.
- **Aggregierte Reporting-Kennzahl** „Direktverkauf" (Anzahl + Umsatz, abzüglich
  Storno) — eine Zeile in den bestehenden Reporting-Kennzahlen, **keine**
  Aufschlüsselung pro Theke (es gibt nur eine).
- **Bondruck über die gemeinsame Outbox-Infrastruktur** (→ `docs/handbuch.md` §4.6) —
  der Direktverkauf speist dieselbe Arbeitsbon-Policy wie die Tischbestellung und kann
  auf Anforderung denselben Kassenbeleg erzeugen. Genau eine zusätzliche
  Konfigurationsentscheidung (`direktverkauf_modus`) steuert den operativen Druck:
  `kein_bon` (nichts), `abholbon` (ein kombinierter Bon für den Gast zur Selbstabholung,
  festes Label „Direktverkauf") oder `an_stationen` (Positionen nach Kategorie an die
  Druckstationen, wie ein Tisch).

## Ubiquitous Language (neue Begriffe)

| Begriff                       | Bedeutung                                                                                                                                                                                      |
| ----------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Direktverkauf**             | Geschäftsvorfall: ein in einem Schritt abgeschlossener Barverkauf an der Theke (bestellen+zahlen+ausgeben zugleich).                                                                           |
| **Verkauf**                   | Die konkrete Instanz eines Direktverkaufs, identifiziert durch eine UUID. Bildet einen eigenen Event-Stream.                                                                                   |
| **Direktverkauf-Stornierung** | Positionsgenaue Korrektur/Rückgabe eines Verkaufs durch Serviceleitung/Admin. Gibt Bargeld sofort zurück.                                                                                      |
| **Abholbon**                  | Operativer, nicht-fiskalischer Direktverkauf-Bon **für den Gast** (Selbstabholung). Festes Label „Direktverkauf", keine Preise. **Kein** Beleg i. S. v. § 146a AO (→ `docs/handbuch.md` §4.6). |

**Abgrenzung:** Der Direktverkauf ist **personalbedient** (ein Vereinsmitglied
bedient das Gerät). Der **Selbstbedienungs-Kiosk** (Gast bedient das Gerät selbst)
bleibt laut Produktbeschreibung ausdrücklich **out of scope**.

## Domänenmodell

### Direktverkauf (Event-Aggregat — im Kasse-Kontext `domain/kasse/`)

Der Direktverkauf ist ein Geschäftsvorfall der Kasse und wird ins **Kassenjournal**
geschrieben (wie alle finanziellen Vorfälle). Das **Aggregat ist der einzelne
Verkauf** — er hat einen minimalen Lebenszyklus: _getätigt_ → optional _(teil-)storniert_.
Es gibt **keine** Direktverkauf-Stammdaten-Entität: die Theke ist ein einziger,
impliziter Kanal ohne Namen, Status oder ID.

> **Designprinzip:** Die Stream-Ebene ist immer das Aggregat. Beim Tisch ist das
> Aggregat die Tisch-Session (`…/tisch-{id}`). Beim Direktverkauf ist das Aggregat der
> **Verkauf** (`…/direktverkauf-{uuid}`). Der Verkauf identifiziert sich allein über
> seine UUID; es gibt keine darüberliegende Verkaufsstelle.

#### Subject-Design

```
kassensitzung-{nr}/direktverkauf-{uuid}
```

- Jeder Verkauf erzeugt genau **einen** Stream. `direktverkauf-getaetigt:v1` ist
  `version = 1`; spätere Teilstornos sind `version = 2, 3, …` im **selben** Stream
  (das Storno mutiert genau diesen Verkauf).
- Da unterschiedliche Verkäufe unterschiedliche Subjects haben, gibt es **keinen
  Versionskonflikt zwischen parallelen Verkäufen** — selbst wenn ausnahmsweise zwei
  Helfer gleichzeitig kassieren, kollidiert kein Versionsraum.
- Cross-Stream-Aggregation (Reporting, Kassenbestand) läuft kanonisch über die
  Spalte `kassensitzung_nr` — exakt die in `docs/handbuch.md` §3.3 vorgeschriebene
  Strategie, **nicht** über fragile Subject-LIKE-Patterns.

#### Event: `direktverkauf-getaetigt:v1`

Ein Vereinsmitglied schließt einen Barverkauf an der Theke ab.

| Feld                 | Typ        | Beschreibung                                                |
| -------------------- | ---------- | ----------------------------------------------------------- |
| `verkauf_id`         | UUID       | Eindeutige ID des Verkaufs                                  |
| `positionen[]`       | Position[] | Mindestens 1 Position (gleiche `Position` wie beim Tisch)   |
| `gesamtbetrag_cents` | int        | Summe aller Positionen in Cent (= sofort kassierter Betrag) |
| `kommentar`          | string?    | Optional, max. 100 Zeichen                                  |

`Position` wird **unverändert wiederverwendet** (`position_id`, `variante_id`,
`produkt_name`, `variante_name`, `kategorie`, `einzelpreis`, `menge` — Fat Event).

#### Event: `direktverkauf-storniert:v1`

Positionsgenaue Korrektur/Rückgabe durch Serviceleitung/Admin.

| Feld                       | Typ           | Beschreibung                              |
| -------------------------- | ------------- | ----------------------------------------- |
| `stornierung_id`           | UUID          | Eindeutige ID                             |
| `verkauf_id`               | UUID          | Referenz auf den stornierten Verkauf      |
| `positionen[]`             | PositionRef[] | `position_id` (UUID) + `menge` (int, ≥ 1) |
| `gesamt_stornierung_cents` | int           | Summe der stornierten Positionen in Cent  |
| `kommentar`                | string        | **Pflicht**, min. 3, max. 100 Zeichen     |

#### Invarianten

- **Kassensitzung-Invariante.** Ein Direktverkauf erfordert eine offene
  Kassensitzung (gleiche Regel wie für Tische). Keine offene KS → HTTP 409.
- **Mindestmengen-Invariante.** Ein Verkauf hat mindestens eine Position mit
  `menge ≥ 1`. `gesamtbetrag_cents` ist die konsistente Summe der Positionen.
- **Vollständigkeits-Invariante.** Ein Direktverkauf ist immer vollständig bezahlt —
  es gibt keinen Saldo, keine Teilzahlung, keinen offenen Restbetrag (konzeptionell
  stets 0).
- **Storno-Invariante (positionsgenau).** Storniert werden können nur Positionen
  dieses Verkaufs, die noch **nicht** storniert wurden, und höchstens in der
  ursprünglich verkauften Menge. Validierung per On-Demand-Replay des Verkauf-Streams
  (siehe unten). Mehrere Teilstornos pro Verkauf sind zulässig.
- **Rolleninvariante.** Direktverkauf tätigen: `admin`, `serviceleitung`, `service`.
  Direktverkauf stornieren: nur `admin`, `serviceleitung` — identisch zur bestehenden
  Storno-Kontrolltrennung beim Tisch.
- **Kassenwirksamkeit.** `direktverkauf-getaetigt` erhöht den Soll-Kassenbestand um
  `gesamtbetrag_cents` (Bargeld rein). `direktverkauf-storniert` verringert ihn um
  `gesamt_stornierung_cents` (Bargeld als Rückgabe sofort raus). **Anders als beim
  Tisch** ist hier **keine separate `auszahlung-geleistet`-Buchung** nötig — die
  Rückgabe ist Teil des Storno-Vorgangs selbst, weil es keinen aufzulösenden Saldo
  gibt.

## User Stories

### Verkaufen am Direktverkauf

1. Als Vereinsmitglied möchte ich den Direktverkauf über einen einzelnen Menüpunkt im
   Service-Bereich öffnen, damit ich Barverkäufe an der Theke erfassen kann.
2. Als Vereinsmitglied möchte ich eine auf den Direktverkauf zugeschnittene
   „Verkaufen"-Oberfläche sehen (Produktauswahl + ein Abschluss-Button), nicht die
   Tabs Bestellen/Bezahlen, damit ich pro Verkauf nur einen Schritt brauche.
3. Als Vereinsmitglied möchte ich Produkte nach Kategorien auswählen und Mengen
   festlegen, damit ich die Bestellung des Gastes zusammenstelle.
4. Als Vereinsmitglied möchte ich die laufende Summe sehen, damit ich dem Gast den zu
   zahlenden Betrag nennen kann.
5. Als Vereinsmitglied möchte ich den Verkauf mit **einer** Bestätigung abschließen
   (bestellen + bezahlen + ausgeben zugleich), damit der Vorgang flüssig läuft.
6. Als Vereinsmitglied möchte ich nach Erfolg sofort eine leere Eingabe für den
   nächsten Gast haben, damit ich zügig weiterarbeite.
7. Als Vereinsmitglied möchte ich optional einen Kommentar erfassen (max. 100
   Zeichen), damit ich Besonderheiten festhalte.
8. Als Vereinsmitglied möchte ich, dass ein Verkauf erst nach erfolgreicher
   Speicherung als abgeschlossen gilt, damit kein Verkauf verloren geht oder doppelt
   bucht.
9. Als Vereinsmitglied möchte ich optional den erhaltenen Betrag eingeben und das
   Rückgeld angezeigt bekommen (bestehende Rückgeldberechnung K-10, rein
   clientseitig), damit ich mich nicht verrechne.

### Bon / Ausgabe (bestehende Bondruck-Infrastruktur, → `docs/handbuch.md` §4.6)

10. Als Betreiber möchte ich konfigurieren, ob beim Direktverkauf **kein Bon**, ein
    **Abholbon** an den Gast oder **Bons an die Stationen** gedruckt werden
    (`direktverkauf_modus`), damit ich die Ausgabe nach meiner Vereinsorganisation
    steuere.
11. Als Gast am Direktverkauf möchte ich (im `abholbon`-Modus) **einen** kombinierten
    Bon mit allen meinen Artikeln und dem festen Label „Direktverkauf" (statt eines
    Tischnamens) erhalten, mit dem ich Essen und Getränke selbst abhole.
12. Als Vereinsmitglied möchte ich (im `an_stationen`-Modus), dass meine
    Direktverkauf-Positionen genauso an Küche/Theke gedruckt werden wie eine
    Tischbestellung, damit die Ausgabe dort vorbereitet wird.
13. Als Betreiber möchte ich (im `kein_bon`-Modus) gar nichts drucken, weil bei uns
    alles auf Zuruf und Vertrauen läuft.
14. Als Gast am Direktverkauf möchte ich **zusätzlich** auf Anforderung einen
    fiskalischen Kassenbeleg erhalten können (§ 146a AO, derselbe On-Demand-Beleg wie
    beim Tisch), unabhängig vom Abholbon.
15. Als Vereinsmitglied möchte ich, dass nach dem Verkauf nichts als „ausstehende
    Ausgabe" offen bleibt, damit die Kasse keine vermeintlich offenen Posten zeigt.

### Korrektur / Storno / Rückgabe

16. Als Serviceleitung möchte ich einen fehlerhaften Verkauf positionsgenau
    stornieren können, damit ich Vertipper oder Rückgaben korrigieren kann, ohne den
    ganzen Verkauf rückgängig machen zu müssen.
17. Als Serviceleitung möchte ich, dass das zurückgegebene Bargeld durch die
    Stornierung selbst kassenwirksam wird (Kassenbestand sinkt), ohne dass ich eine
    separate Auszahlung buchen muss.
18. Als Vereinsmitglied (`service`) möchte ich selbst **keine** Stornierungen
    auslösen können, damit die Kontroll-Trennung erhalten bleibt.
19. Als Serviceleitung möchte ich die Direktverkauf-Historie einsehen (eine Zeile pro
    Verkauf), um zu stornierende Verkäufe zu finden.

### Übersicht / Reporting

20. Als Servicekraft möchte ich den Direktverkauf klar von Tischen unterscheiden
    (eigener Menüpunkt, kein Saldo), damit die Übersicht ehrlich bleibt.
21. Als Admin möchte ich den Direktverkauf-Umsatz als **eine aggregierte Kennzahl**
    (Anzahl + Umsatz) in den Reporting-Kennzahlen sehen, damit ich den Beitrag der
    Theke erkenne — ohne Aufschlüsselung pro Theke (es gibt nur eine).
22. Als Admin möchte ich, dass Direktverkäufe im **Gesamtumsatz**, im
    **Kassenbestand** und im **Tagesabschluss (Z-Bon)** vollständig enthalten sind,
    damit die Kassenführung stimmt.
23. Als Admin/Serviceleitung möchte ich Direktverkäufe und ihre Stornierungen im
    Kassenjournal lückenlos nachvollziehen, damit Revisionssicherheit gewahrt bleibt.

### Robustheit / Randfälle

24. Als Vereinsmitglied möchte ich ohne offene Kassensitzung nicht verkaufen können
    (gleiche Invariante wie beim Tisch).
25. Als Entwickler möchte ich, dass der Tisch-Ablauf
    (`bestellung-aufnehmen`/`zahlung-kassieren`/`ausgabe-bestaetigen`) und der
    Direktverkauf-Ablauf sich **nicht vermischen** — getrennte Endpunkte, getrennte
    Events.
26. Als Serviceleitung möchte ich bei parallelem Storno-Konflikt auf denselben
    Verkauf (OCC) eine klare 409-Rückmeldung erhalten und es erneut versuchen können.
27. Als Vereinsmitglied möchte ich, dass beim Tagesabschluss Direktverkäufe **nie**
    als „offener Saldo" blockieren (sie sind per Definition abgeschlossen).

## Implementation Decisions

### Domäne / Datenmodell

- **Keine neue Stammdaten-Entität und keine neue Tabelle.** Der Direktverkauf ist ein
  reiner Geschäftsvorfall im Kasse-Kontext — es gibt keine `Verkaufsstelle`, kein CRUD,
  keine `verkaufsstellen`-Tabelle. Damit bleibt `database/migrations/01_initial.up.sql`
  in diesem Bereich unverändert.
- **Keine Projektionstabelle.** Es gibt kein Projektions-Pendant zu `tisch_sessions` —
  ein Direktverkauf hat keinen veränderlichen Zustand, der materialisiert werden müsste.
- **Validierung beidseitig** (zog im Backend, Zod im Frontend) für den
  Direktverkauf-Request und den Storno-Request.

### Event-Modell / Aggregat

- **Zwei neue Event-Typen** in `domain/kasse/direktverkauf.go`:
  `direktverkauf-getaetigt:v1`, `direktverkauf-storniert:v1`. Event-Data-Structs mit
  stabilen `json`-Keys (immutable). Erstellungsfunktionen validieren per zog (Muster
  wie `NewBestellungAufgenommenEvent` et al.).
- **`Position` wird wiederverwendet** (gleiche Fat-Event-Struktur). Positions-IDs
  werden beim Tätigen erzeugt (wie bei der Bestellung).
- **Ein Verkauf = ein Event.** Das Tätigen schreibt **genau ein** Event. Es ist
  **kein** atomarer Mehr-Event-Write nötig (großer Vereinfachungsgewinn gegenüber der
  verworfenen Drei-Event-Alternative — keine neue Transaktions-Bündelungs-API im
  Repository).
- **Subject:** `kasse.DirektverkaufSubject(zNr, verkaufUUID)` →
  `kassensitzung-{nr}/direktverkauf-{uuid}`. Neue Parser/Builder in
  `domain/kasse/subject.go`.

### Storno-Validierung ohne Projektion

- **Reine Domänenfunktion** in `domain/kasse` (z. B.
  `ComputeNichtStornierteVerkaufPositionen(events []Event) ([]Position, error)`):
  nimmt `direktverkauf-getaetigt`-Positionen, zieht alle `direktverkauf-storniert`-
  Positionen ab, liefert die noch stornierbaren Positionen. Strukturell analog zum
  bestehenden `ComputeNichtStorniertePositionen` des Tisches, aber auf den **einzelnen
  Verkauf-Stream** angewandt (sehr wenige Events → vernachlässigbar günstig).
- Der Storno-Command lädt den Verkauf-Stream per `ReadEventsBySubject`, validiert die
  angeforderten Positionen gegen das Ergebnis und schreibt `direktverkauf-storniert:v1`
  mit `version = maxVersion + 1` (OCC wie gehabt).

### Repository

- **Kein neuer atomarer Multi-Write.** Die bestehende `WriteEvent(ctx, e, streamType,
kassensitzungNr)` wird genutzt. Neuer `StreamType` `direktverkauf`, der **nur** das
  Kassenjournal-INSERT ausführt und **keine** Projektion aktualisiert (Routing-`switch`
  in `kassenjournal_repo` um einen Zweig ergänzt, der bewusst nichts projiziert).
- OCC über `GetMaxVersion(subject)` + `UNIQUE(subject, version)` unverändert.

### Application / API (POST-only, deutsch)

- **Keine Admin-Endpunkte für Stammdaten.** Es gibt keine Verkaufsstellen-Verwaltung,
  also auch keine `create`/`update`/`activate`/`deactivate`/`delete`/`get-all`-Routen.
- **Service-Command** `DirektverkaufTaetigen` (Rollen admin/serviceleitung/service):
  prüft die offene Kassensitzung, reichert Positionen wie bei der Bestellung an
  (Batch-Fetch Produkte/Varianten, Fat Events), baut und schreibt
  `direktverkauf-getaetigt:v1`. Endpunkt `POST /direktverkauf-taetigen` in `service.go`.
  Request-DTO: `positionen[]` (`produktId`, `varianteId`, `menge`), `kommentar?` —
  **kein** `verkaufsstelleId`.
- **Serviceleitung-Command** `DirektverkaufStornieren` (Rollen admin/serviceleitung):
  Endpunkt `POST /direktverkauf-stornieren` in `serviceleitung.go`. Request-DTO:
  `verkaufId`, `positionen[]` (`positionId`, `menge`), `kommentar` (Pflicht).
- **Query** `GetDirektverkaufHistorie` für die kompakte Verkaufs-/Storno-Historie der
  offenen KS (eine Zeile pro Verkauf). Endpunkt `POST /get-direktverkauf-historie`.
- **Response-DTOs in der HTTP-Schicht** definiert; Domain-Structs werden nie direkt
  serialisiert.

### Frontend

- **Service-Anbindung** über eine neue `DirektverkaufBackend`-Klasse auf Basis des
  `BackendClient`-Interfaces (`direktverkaufTaetigen`, `direktverkaufStornieren`,
  `getDirektverkaufHistorie`). Kein direktes `fetch()`.
- **Navigation:** **ein** Menüpunkt „Direktverkauf" in der Service-Gruppe der Sidebar
  (erreichbar für admin/serviceleitung/service) — keine Theken-Liste, da es nur eine
  Theke gibt.
- **Verkaufen-Seite:** kombinierte „Verkaufen"-Oberfläche (Produktauswahl +
  Abschluss-Button + optionale Rückgeldberechnung K-10) statt der Tabs
  Bestellen/Bezahlen. Nach Erfolg leert sich die Eingabe für den nächsten Gast.
- **Direktverkauf-Historie** kompakt (eine Zeile pro Verkauf), mit positionsgenauer
  Storno-Aktion für serviceleitung/admin.
- **Keine Admin-Verwaltungsseite** — es gibt keine Stammdaten zu pflegen.

### Bondruck (nutzt die bestehende Bondruck-Infrastruktur, → `docs/handbuch.md` §4.6)

Voraussetzung ist die bereits umgesetzte Bondruck-Infrastruktur (Druckauftrags-Outbox,
Arbeitsbon-Policy, Kassenbeleg-Command, `bondruck_einstellungen`-Singleton — →
`docs/handbuch.md` §4.6). Der
Direktverkauf konsumiert diese Infrastruktur und fügt **keinen** neuen Bon-Typ hinzu —
nur eine Routing-Entscheidung und einen Abholbon-Formatter.

- **`direktverkauf_modus` (Konfiguration).** `bondruck_einstellungen` wird um
  `direktverkauf_modus` (`kein_bon` | `abholbon` | `an_stationen`) und
  `abholbon_drucker_ip` erweitert. Beidseitig validiert (Enum + IPv4).
- **Arbeitsbon-Policy erweitert.** Die bestehende Arbeitsbon-Policy reagiert zusätzlich
  auf `direktverkauf-getaetigt:v1`:
  - `an_stationen` → Positionen nach Kategorie an die Druckstationen (identische Logik
    wie beim Tisch),
  - `abholbon` → **ein** kombinierter Abholbon-Druckauftrag an `abholbon_drucker_ip`
    (festes Label „Direktverkauf", keine Preise),
  - `kein_bon` → keine Outbox-Zeile.
- **Festes Bon-Label „Direktverkauf".** Wo der Tisch-Arbeitsbon den Tischnamen aus dem
  Subject ableitet, nutzt der Abholbon den konstanten Text „Direktverkauf" — kein
  Subject-Parsing, kein Name im Event nötig.
- **Direktverkauf-Kassenbeleg.** Der bestehende On-Demand-Kassenbeleg-Command
  akzeptiert zusätzlich eine **Verkauf-Referenz**; derselbe Formatter und dieselbe
  Outbox, nur die Datenquelle ist der Verkauf statt einer Tischzahlung. Unabhängig vom
  Abholbon.

### Reporting / Kassenbestand (Compliance-relevant)

- **Zwei neue SQL-Extraktoren** in der Migration:
  `kj_extract_direktverkauf_cents` (`direktverkauf-getaetigt` → `gesamtbetragCents`)
  und `kj_extract_direktverkauf_storno_cents` (`direktverkauf-storniert` →
  `gesamtStornierungCents`).
- **`GetKassenbestand` erweitern:** `+ direktverkauf-getaetigt`,
  `− direktverkauf-storniert` (Bargeld rein bzw. als Rückgabe raus). Damit ist der
  Direktverkauf im Soll-Bestand und folglich im Kassensturz/Z-Bon korrekt enthalten.
- **`GetReportingStats` erweitern:** Direktverkäufe erhöhen den Gesamtumsatz,
  Direktverkauf-Stornos mindern ihn — konsistent zur bestehenden Umsatzdefinition.
  Zusätzlich **eine** aggregierte Kennzahl (Anzahl Direktverkäufe + Direktverkauf-Umsatz)
  als eigene Zeile, **ohne** Gruppierung — kein `GetUmsatzProVerkaufsstelle`, kein JOIN.
- **Tagesabschluss-Sperre unberührt:** Die „alle Tisch-Sessions Saldo 0"-Bedingung
  gilt nur für Tische. Direktverkäufe haben keinen Saldo und blockieren den
  Tagesabschluss nie.

### Compliance / DSFinV-K / TSE-Roadmap

- **Ein Geschäftsvorfall = ein Verkauf = ein Event** → sauberes späteres **1:1-Mapping
  auf eine TSE-Transaktion** (`StartTransaction`/`FinishTransaction` sofort), im
  Gegensatz zur 3:1-Problematik der verworfenen Alternative.
- **`ABRECHNUNGSKREIS`:** Da jeder Verkauf sofort geschlossen ist, ist der Verkauf
  selbst der Abrechnungskreis; das feste Label „Direktverkauf" dient als Bezeichnung im
  DSFinV-K-Export. Finale Festlegung in der TSE-Phase; das append-only Event genügt
  schon jetzt für Revisionssicherheit (§ 146 AO).
- **Storno = eigener Geschäftsvorfall** (Korrektur/Rückgabe), append-only erfasst und
  kassenwirksam — vollständig nachvollziehbar.

## Testing Decisions

Tests prüfen **externes Verhalten** an der öffentlichen Modul-Schnittstelle, nicht
Implementierungsdetails. Priorisiert:

1. **Direktverkauf-Events + Storno-Replay (`domain/kasse`) — Hauptkandidat, isoliert
   testbar.** `direktverkauf-getaetigt` erzeugt genau ein Event mit konsistenter Summe
   und Positions-IDs; `ComputeNichtStornierteVerkaufPositionen` liefert korrekt die
   noch stornierbaren Positionen nach mehreren Teilstornos; Vollstorno → leere
   Restmenge. Prior Art: `tisch_session_test.go`, `historie_test.go`.
2. **Application-Commands.** `DirektverkaufTaetigen`: keine offene KS → Fehler;
   Happy Path schreibt ein Event. `DirektverkaufStornieren`: Storno über verfügbare
   Menge hinaus → Fehler; nicht-existenter Verkauf → Fehler; OCC-Konflikt → 409;
   `service`-Rolle → abgewiesen. Prior Art: `api/table/application/command_test.go`.
3. **Trennungs-Guards.** Der Tisch-Ablauf und der Direktverkauf-Ablauf teilen keinen
   Zustand und keine Endpunkte; ein Direktverkauf-Request landet nie im Tisch-Stream
   und umgekehrt.
4. **Reporting/Kassenbestand (Integrationstest).** Direktverkauf erhöht
   Soll-Kassenbestand und Gesamtumsatz; Storno mindert beide; die aggregierte
   Direktverkauf-Kennzahl zählt korrekt. Prior Art:
   `repository/kassenjournal_repo/repo_test.go`, Reporting-Repo-Tests.
5. **Direktverkauf-Bons (nutzt die bestehende Bondruck-Infrastruktur).** Die Arbeitsbon-Policy
   erzeugt aus `direktverkauf-getaetigt` je `direktverkauf_modus` die richtigen
   Druckaufträge: `an_stationen` → Station-Bons nach Kategorie, `abholbon` → genau ein
   Abholbon (Label „Direktverkauf", keine Preise), `kein_bon` → keine Outbox-Zeile; der
   Direktverkauf-Kassenbeleg erzeugt genau einen Druckauftrag für einen echten Verkauf.
   Prior Art: `backend/api/bondruck` (Formatter-/Policy-Tests).
6. **Frontend.** Verkaufen löst genau einen Backend-Aufruf aus; nach Erfolg leere
   Eingabe; die Verkaufen-Seite rendert die kombinierte Oberfläche statt Tabs; Historie
   kompakt. Prior Art: bestehende Service-Komponententests, `routes.test.ts`.

Mit dem Nutzer abzustimmen: ob alle sechs Bereiche getestet werden oder der Fokus
zunächst auf dem Backend-Kern (1–5) liegt.

## Out of Scope

- **Selbstbedienungs-Kiosk** (gastbedient) — bleibt ausgeschlossen; der Direktverkauf
  ist personalbedient.
- **Benannte / mehrere Verkaufsstellen** — es gibt genau eine implizite Theke. Eine
  `Verkaufsstelle`-Stammdatenentität (Name, Status, CRUD) und „Umsatz pro Theke" sind
  bewusst nicht enthalten (YAGNI; bei echtem Bedarf später nachrüstbar — „rule of three").
- **Kartenzahlung / Zahlungsgateway** — Direktverkauf ist Barzahlung.
- **Steuer-Aufschlüsselung und TSE-Pflichtfelder auf dem Kassenbeleg** — hängen an F-07
  bzw. F-02 (→ `docs/handbuch.md` §4.6); der Basis-Kassenbeleg des Direktverkaufs ist
  davon unabhängig druckbar.
- **Bondruck-Infrastruktur selbst** (Outbox, Relay-Transport, Druckstation-Rename,
  Kassenbeleg-Command) — bereits umgesetzt (→ `docs/handbuch.md` §4.6); diese PRD
  konsumiert sie nur.
- **Teilzahlung / offener Saldo beim Direktverkauf** — ein Verkauf ist immer
  vollständig.
- **Separate Ausgabe-Bestätigung / Küchendisplay (K-13)** — beim Direktverkauf nicht
  verfolgt.

## Further Notes

- **Abhängigkeit von der Bondruck-Infrastruktur.** Der Bon-/Beleg-Teil dieses Features (Abholbon,
  `direktverkauf_modus`, Stations-Routing, Direktverkauf-Kassenbeleg) setzt die
  bereits umgesetzte Bondruck-Infrastruktur voraus (Druckauftrags-Outbox,
  Arbeitsbon-Policy, Kassenbeleg-Command, `bondruck_einstellungen`-Singleton — →
  `docs/handbuch.md` §4.6). Der **Kern** des
  Direktverkaufs (Event-Aggregat, Storno, Reporting, Kassenwirksamkeit) ist davon
  unabhängig und kann eigenständig umgesetzt werden.

Bei der Umsetzung zu aktualisierende Dokumente:

- `docs/language.md` — neue Begriffe „Direktverkauf", „Verkauf",
  „Direktverkauf-Stornierung", „Abholbon"; Namenskonventionen pro Schicht; Abgrenzung
  zum „Selbstbedienungs-Kiosk". Klarstellen, dass es keine `Verkaufsstelle`-Entität gibt.
- `docs/handbuch.md` — Direktverkauf als eigenes Event-Aggregat (Subject
  `kassensitzung-{nr}/direktverkauf-{uuid}`), die zwei Event-Typen, Storno-Replay,
  Kassenwirksamkeit, neuer `StreamType` `direktverkauf`, fehlende Projektion.
- `docs/anforderungen.md` — neue Anforderung (Vorschlag: **K-23 · Direktverkauf**;
  K-22 ist die aktuell höchste Nummer).
- `docs/compliance.md` — Direktverkauf als einzelner Geschäftsvorfall, 1:1-TSE-Mapping,
  `ABRECHNUNGSKREIS`-Behandlung, Kassenwirksamkeit von Verkauf und Storno.
- `docs/produktbeschreibung.md` — klarstellen: Direktverkauf (personalbedient) ist
  enthalten; „Selbstbedienungs-Kiosk" (gastbedient) bleibt ausgeschlossen.

**Kontrast zur verworfenen Alternative (zur Begründung im Review):** Den Direktverkauf
als Tisch mit `direktverkauf`-Flag und drei Events pro Barverkauf zu modellieren, presst
zwei entgegengesetzte Konzepte in ein Modell. Das eigene, schlanke Event-Aggregat löst
das an der Wurzel: ehrliche Event-Semantik (1 Verkauf = 1 Event statt 3:1), eigener
Stream pro Verkauf (Storno als `version 2`, kein geteilter Versionsraum), klares
TSE-Mapping und kein ungenutzter Aggregat-Apparat — und das alles **ohne** eine eigene
Stammdaten-Entität.
