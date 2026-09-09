# ADR 12: Sprachschnitt der Bezeichner

- **Status:** akzeptiert (2026-09-09)
- **Kontext-Dokumente:** `docs/plans/plan-jotti-audit-fixes.md` Phase 14 (nach
  Merge gelöscht, siehe Git-Historie); `docs/language.md` Regeln 1, 2 und 5
  sowie die Schichtentabelle; `backend/api/admin.go`

## Kontext

`docs/language.md` schreibt die Sprache je Schicht fest. Drei Stellen sind für
diese Entscheidung maßgeblich:

- **Regel 1** (Zeile 9): Fachbegriffe der Kasse, der Stammdaten und der
  Gastronomie-Domäne sind deutsch — „in Code, Dokumentation und Kommunikation".
- **Regel 5** (Zeile 17): zustandsändernde Domänen-Commands („Vorgänge, die das
  Fiskalrecht oder ein Kassenwart benennt") tragen deutsche Fachverben; Queries,
  Derivationen, Persistence und Infrastruktur tragen englische Verben.
  Englisches Verb plus deutsches Nomen ist ausdrücklich „der Normalfall".
- **Schichtentabelle** (Zeile 27): API-Pfade der Domäne sind deutsch, Beispiele
  `/bestellung-aufnehmen` und `/zahlung-kassieren`.

### Endpunkt-Verben

`RegisterAdminRoutes()` registriert 56 Routen. 15 tragen ein deutsches Verb, 41
ein englisches. Nach dem Prinzip aus Regel 5 sortiert sich das so:

| Gruppe                          | Verb     | Beispiel (Zeile)               | Anzahl |
| ------------------------------- | -------- | ------------------------------ | ------ |
| Kassenführung, Fiskal, Bondruck | deutsch  | `/kasse-abschliessen` (102)    | 10     |
| Reihenfolge-Operationen         | deutsch  | `/verschiebe-produkt` (49)     | 3      |
| Betreiber-Meldung               | deutsch  | `/elster-meldung-setzen` (154) | 2      |
| Stammdaten-CRUD                 | englisch | `/create-produkt` (45)         | 13     |
| Konfigurations-Updates          | englisch | `/update-betreiber` (153)      | 3      |
| Benutzerverwaltung              | englisch | `/create-user` (34)            | 6      |
| Queries                         | englisch | `/get-all-produkte` (58)       | 17     |
| Export                          | englisch | `/export/dsfinvk` (91)         | 1      |
| TSE-Verbindungstest             | englisch | `/test-tse-verbindung` (136)   | 1      |

Die zehn deutschen Command-Routen benennen genau die Vorgänge, die ein
Kassenwart benennt: Kassensitzung eröffnen, Geldtransit buchen, Kasse
abschließen, TSE einrichten und übernehmen, Bons drucken und Druckaufträge
verwerfen. Die Reihenfolge-Operationen nennt Regel 5 namentlich als deutsch.
Die Benutzerverwaltung ist nach Regel 2 (Zeile 11) Infrastruktur. Die Queries
folgen Regel 5.

Dasselbe Bild in den übrigen Routendateien: `api/service.go` führt
`/bestellung-aufnehmen` (34), `/zahlung-kassieren` (36),
`/direktverkauf-taetigen` (61) neben `/get-aktive-tische` (75);
`api/serviceleitung.go` führt `/stornierung-erteilen` (24); `api/auth.go` führt
`/login` und `/set-password` (17–18).

Ein Ausreißer bleibt: `/favorit-hinzufuegen` und `/favorit-entfernen`
(`api/service.go:52-53`) sind deutsch, obwohl `docs/language.md:366` den
Favoriten als „einfache CRUD-Relation" führt — dieselbe Kategorie wie das
englisch benannte Stammdaten-CRUD.

### Frontend-Bezeichner

Hier gilt Regel 1 ohne Einschränkung, und hier wird sie verletzt. In
`frontend/src` (ohne `components/ui`) stehen 18 englische Bezeichner für
Domänenbegriffe mit 114 Vorkommen in 26 Dateien:

| Bezeichner                                                                       | Domänenbegriff        |
| -------------------------------------------------------------------------------- | --------------------- |
| `Receipt`, `ReceiptPosition`                                                     | Kassenbeleg, Position |
| `Products`, `ProductsProps`, `ProductItem`, `ProductItemProps`                   | Produkt               |
| `ProductList`, `ProductListSkeleton`, `ProductListComponentProps`                | Produktliste          |
| `VariantChip`, `VariantChipProps`, `VariantRow`, `VariantNamePreis`              | Variante              |
| `EditProductDialog`, `NewProductDialog`, `EditVariantDialog`, `NewVariantDialog` | Produkt, Variante     |
| `HistoryRow`                                                                     | Historie              |

15 Dateinamen tragen denselben Schnitt (`admin/products/Products.tsx`,
`ProductItem.tsx`, `productGrouping.ts`, `service/components/table/Receipt.tsx`,
`ProductList.tsx`, `components/common/VariantNamePreis.tsx` und weitere), dazu
fünf Verzeichnisse: `admin/products`, `admin/tables`, `service/product`,
`service/table`, `service/components/table`.

Die englischen Namen stehen neben ihren deutschen Gegenstücken in derselben
Datei: `VariantRow` (`service/components/table/ProductList.tsx:122`) rendert
`Variante`n, `ReceiptPosition` (`service/components/table/Receipt.tsx:3`) trägt
die Felder `einzelpreisCents` und `menge`, und `Receipt` nimmt die Props
`positionen` und `totalPrice` nebeneinander (Zeilen 9–14). Die Route
`/admin/produkte` (`AdminSidebar.tsx:136`) lädt `admin/products/AdminProductsPage`
(`routes.ts:108`).

Zwei Verzeichnisse sind nur außen englisch: `admin/tables` enthält
`Tisch.ts`, `Tische.tsx`, `TischItem.tsx`, `tischGrouping.ts`; `service/table`
enthält `Bestellung.ts`, `Stornierung.ts`, `Umbuchung.ts`, `Zahlung.ts`.
`admin/users` bleibt englisch — `User` ist die dokumentierte Ausnahme aus
Regel 2.

### Website

`website/src/lib/live-demo.ts` (200 Zeilen) ist durchgehend englisch:
`DemoProduct` (16), `Cart` (56), `addVariant` (60), `cartTotalCents` (73). Die
Datei simuliert eine Bestellung für die Marketing-Seite. `@jotti/website` ist
ein eigenes Paket, das keine Datei des Frontends importiert und keinen
Pfad-Alias dorthin hat (`website/tsconfig.json` erweitert nur
`astro/tsconfigs/strict`). `docs/language.md` erwähnt die Website an keiner
Stelle.

### Reichweite einer Umbenennung

| Fläche                        | Vorkommen                                  |
| ----------------------------- | ------------------------------------------ |
| Frontend-Bezeichner           | 114 in 26 Dateien, davon 15 Dateinamen     |
| Frontend-Routen               | keine — die Routen sind bereits deutsch    |
| E2E-Selektoren                | keine                                      |
| E2E-Kommentare                | 5 Zeilen in 3 Spec-Dateien                 |
| Anwender- und Architekturdoku | keine                                      |
| ADR 10                        | 5 Zeilen (`ProductList.tsx`, `VariantRow`) |

Die E2E-Suite verankert ihre Selektoren an gerenderten deutschen Texten, nicht
an Komponentennamen; die fünf Treffer in `variantenname-umbruch.mobile.spec.ts`,
`tischservice-viewport-ueberlauf.mobile.spec.ts` und
`produktliste-sticky-split.spec.ts` stehen ausschließlich in Kommentaren.

### Erwogene Alternativen

1. **Alles umbenennen — Endpunkte und Frontend.** Trifft 16 Command-Routen, die
   nach Regel 5 richtig heißen, und ändert damit einen Vertrag ohne Regelbezug.
2. **Nichts umbenennen, alles als Ausnahme in `docs/language.md` schreiben.**
   Schreibt 18 Bezeichner als Ausnahme fest und schwächt Regel 1 an ihrer
   Kernstelle. Präzedenz dagegen: `tisch_repo` wurde umbenannt statt
   dokumentiert.
3. **Nur die Frontend-Bezeichner umbenennen, Endpunkte lassen, Website
   ausnehmen.** Trifft genau die Menge, die eine Regel verletzt.

## Entscheidung

**Die Frontend-Bezeichner werden umbenannt, die Endpunkt-Verben bleiben, die
Website wird ausgenommen** (Alternative 3).

- **Korrektheit:** Keine der drei Alternativen behebt einen Fehler. Die
  Umbenennung im Frontend ist compilergeprüft; TypeScript findet jede
  Fundstelle. Eine Endpunkt-Umbenennung ist es nicht: die Pfade sind
  String-Literale, im Backend in `api/admin.go` und im Frontend in fünf
  Backend-Klassen (`ProduktBackend.ts`, `TischBackend.ts`, `TSEBackend.ts`,
  `DruckstationBackend.ts`, `BetreiberBackend.ts`).
- **Einfachheit:** Zwei Vokabulare in derselben Datei (`VariantRow` über
  `Variante`n) zwingen zur Übersetzung beim Lesen. Ein Vokabular ist einfacher.
  Die Endpunkte tragen kein zweites Vokabular: sie folgen einer Regel, die
  Verb und Nomen bewusst trennt.
- **Konsistenz:** Regel 1 gilt für Code ohne Einschränkung; 18 Bezeichner
  verletzen sie. Die Endpunkte erfüllen Regel 5 und die Schichtentabelle: jeder
  Vorgang, den ein Kassenwart benennt, hat einen deutschen Pfad.
- **Produkt-Konservatismus:** Die Umbenennung ist kein Feature und ändert kein
  Verhalten. Sie kostet Review-Aufmerksamkeit für einen Diff ohne
  Verhaltensänderung — deshalb als eigener Change, nicht nebenbei in einem
  fachlichen Umbau.

**Empfehlung: v1.1.**

## Konsequenzen

- Die Umbenennung landet als ein einzelner Change, der nichts anderes tut. Sein
  Diff besteht aus Bezeichnern, Dateinamen und Importpfaden.
- `docs/language.md` bekommt mit dieser Entscheidung den Geltungsbereich-Hinweis:
  die Konventionen gelten für Backend, Frontend und Datenbank, nicht für das
  Website-Paket. `live-demo.ts` bleibt englisch.
- Die 16 Command-Routen mit englischem Verb bleiben, ebenso `/favorit-hinzufuegen`
  und `/favorit-entfernen`. Wer den Ausreißer auflösen will, braucht ein eigenes
  ADR; er betrifft eine Service-Route, nicht die Admin-Endpunkte dieser
  Entscheidung.
- ADR 10 nennt `ProductList.tsx`, `ProductListSkeleton` und `VariantRow`. ADRs
  werden nicht umgeschrieben; diese fünf Zeilen zeigen nach der Umbenennung auf
  Namen, die es nicht mehr gibt. Das ist der Preis und kein Grund gegen die
  Entscheidung — ein ADR hält den Stand seines Entstehungszeitpunkts fest.
- Die fünf E2E-Kommentare ziehen im selben Change mit.
- **Vorziehen, wenn** ein fachlicher Change ohnehin `ProductList.tsx` oder die
  Produktverwaltung umschreibt: dann ist die Umbenennung dort billiger als
  separat, weil der Diff ohnehin gelesen wird.
- **Zurücknehmen, wenn** `docs/language.md` Regel 1 auf Domänen-_Typen_ verengt
  wird und Komponentennamen ausnimmt. Dann fällt die Grundlage weg, und es
  bleibt nur der Geltungsbereich-Hinweis für die Website.
