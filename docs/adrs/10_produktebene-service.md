# ADR 10: Keine Produktebene über der Variantenliste im Service

- **Status:** abgelehnt (2026-09-07)
- **Kontext-Dokumente:** PR [#111](https://github.com/nicograef/jotti/pull/111)
  (extern, offen); `docs/plans/review-externe-prs.md` § PR #111;
  `docs/plans/plan-praxis-feedback.md` Phase 10; ADR 08 (zweispaltiges
  Service-Layout); `frontend/src/service/components/table/ProductList.tsx`

## Kontext

Ein externer Beitrag schlägt eine dritte Navigationsebene im Bestell-Bildschirm
vor: Kategorie → Produkt-Kacheln → Variantenliste. Ein Tipp auf eine Kachel
öffnet die Varianten, ein Zurück-Knopf führt zurück zum Kachelraster.

### Verhältnis zu den Kategorie-Pills

Der PR ersetzt die vorhandenen Kategorie-Pills nicht, er schiebt eine Ebene
darunter. Der Pill-Block des PR (Zeilen 76–101) ist Zeichen für Zeichen
identisch mit dem heutigen Stand (`ProductList.tsx`, Zeilen 56–81). Gefiltert
wird also weiterhin über die Pills; die Kacheln kommen zusätzlich.

### Variantenzahl je Kategorie

| Quelle                                     | Produkte         | Varianten     | Größtes Produkt      |
| ------------------------------------------ | ---------------- | ------------- | -------------------- |
| Praxis-Setup (Plan, Phase 10)              | 7                | ~50           | nicht berichtet      |
| PR-Text #111                               | 2–4 je Kategorie | nicht genannt | 14 bzw. 16 Varianten |
| Seed-Szenario (`backend/seed/szenario.go`) | 19 aktiv         | 50 aktiv      | 5 Varianten          |

Im Seed verteilen sich die 50 Varianten auf Essen (9 Produkte, 21 Varianten),
Getränke (9 / 27) und Sonstiges (1 / 2). Das Praxis-Setup bündelt eine ähnliche
Gesamtzahl auf weniger Produkte. Der Autor begründet den PR damit:

> Unsere Preisliste hat Produkte mit 14 bzw. 16 Varianten. […] Mit der
> Produktebene stehen pro Kategorie zwei bis vier Kacheln, und die
> Variantenliste ist immer kurz.

Er nennt die Kosten selbst und schränkt ein:

> Das kostet einen Tap pro Bestellvorgang […]. Ob das den Gewinn aufwiegt,
> hängt von der Größe der Preisliste ab — bei kurzen Listen vermutlich nicht.
> […] Ich reiche das als Vorschlag ein, nicht als Notwendigkeit.

### Sichtbare Zeilen am Handy (Pixel 7, 412 × 915)

Eine Variantenzeile ist 61 px hoch: Stepper 44 px (`size-11`), `py-2` zweimal
8 px, 1 px Rahmen. Fix belegt sind Kopfzeile 56 px (`h-14`, sticky),
Kategorieleiste 55 px (6 + 36 + 12 + 1, sticky) und Dock 144 px (`dockFreiraum`,
9 rem).

| Ansicht          | Rest nach Rändern | Kopf der Liste             | Sichtbare Zeilen         |
| ---------------- | ----------------- | -------------------------- | ------------------------ |
| Heute (flach)    | 644 px            | Produktüberschrift ≈ 24 px | (644 − 24) / 61 = **10** |
| Mit Produktebene | 644 px            | Zurück-Zeile ≈ 28 px       | (644 − 28) / 61 = **10** |

Rechenweg: 915 − 56 − 55 − 144 = 660 px, abzüglich `mt-4` (16 px) bleiben
644 px. Beide Ansichten zeigen zehn Variantenzeilen. Für ein Produkt mit 16
Varianten scrollt die Servicekraft mit Ebene genauso wie ohne.

Die Ebene spart also nicht das Scrollen **innerhalb** eines großen Produkts,
sondern nur das Scrollen an anderen Produkten **vorbei**. Genau dafür sortiert
der Admin seit #109 (Phase 5) Produkte und Varianten selbst.

### Tap-Kosten je Bestellung

Beispiel: sieben Einheiten, vier Produkte, zwei Kategorien (der Alltagsfall am
Tisch).

| Tap-Art          | ohne Ebene | mit Ebene |
| ---------------- | ---------- | --------- |
| Plus je Einheit  | 7          | 7         |
| Kategoriewechsel | 1          | 1         |
| Kachel öffnen    | 0          | 4         |
| Zurück           | 0          | 2         |
| **Summe**        | **8**      | **14**    |

Regel: je Produkt ein Tap für die Kachel, je Folgeprodukt derselben Kategorie
zusätzlich ein Tap zurück. Ein-Varianten-Produkte kosten den Kachel-Tap ohne
jeden Gewinn; der PR verzichtet bewusst auf einen Direktweg („Die Ebene ist
immer da"). Im Seed betrifft das zwei von 19 aktiven Produkten.

### Zustand des Beitrags

| Schwere | Befund                                                                                                 |
| ------- | ------------------------------------------------------------------------------------------------------ |
| Blocker | Der Produkt-`<h2>` entfällt, an dem `e2e/support/servicekraft.ts` jede Bestellzeile verankert.          |
| Blocker | Elf Spec-Dateien und `e2e/website/screenshots.mjs` brechen; keine Datei unter `e2e/` ist angefasst.     |
| Major   | Die Kachel rendert `{', '}` als sichtbaren Textknoten: „, 1 Variante".                                  |
| Major   | Die neue Ebene bringt keine einzige Assertion mit; vier Testdateien bekommen nur einen Klick eingefügt. |
| Major   | Der Zurück-Knopf misst rund 24 px und ist der einzige Ausgang; `Stepper.tsx` dokumentiert 44 px.        |
| Major   | `lg:grid-cols-4` hängt am Viewport, das Raster steckt in der ~584 px breiten Spalte aus ADR 08.         |
| Minor   | Ein Kategoriewechsel setzt die geöffnete Produktansicht nur implizit zurück.                            |

Zusätzlich scheitert `e2e/tests/produktliste-sticky-split.spec.ts` an seiner
harten Vorbedingung `scrollTop > 0`: Das Kachelraster läuft bei 1024 × 720 nicht
mehr über (zwei Tests, zwei Breiten). Die e2e-Suite läuft bei jedem PR in CI und
steckt weder in `make check` noch in `make verify`.

Gegen den heutigen Stand ist der Branch nicht mehr konfliktfrei. `git merge-tree
--write-tree afa5ce4 refs/remotes/pr/111` meldet fünf Konflikt-Hunks in zwei
Dateien: vier in `ProductList.tsx`, einer in `Stepper.tsx`. Phase 4 hat dieselben
Bereiche (`VariantRow`, `ProductListSkeleton`, Stepper-Slot) neu geschrieben.

### Erwogene Alternativen

1. **Produktebene mit Direktweg für Ein-Varianten-Produkte** (Stepper auf der
   Kachel). Beseitigt den nutzlosen Tap, nicht die Hauptkosten. Zehn sichtbare
   Zeilen bleiben zehn. Und das Raster trüge zwei Bedienarten nebeneinander.
2. **Nur Pills und Reihenfolge** (Stand nach Phase 5). Die Pills teilen die
   Liste in drei; der Admin sortiert die häufigen Produkte nach oben. Keine neue
   Ebene, kein Bildschirmwechsel, keine e2e-Anpassung.
3. **Aufklappbare Produktgruppen** statt Ebenenwechsel. Die Produkte blieben
   sichtbar, kosten aber weiterhin einen Tap je Produkt. Dazu käme ein
   Zustand je Zeile und ein springender Scroll-Anker beim Aufklappen.

## Entscheidung

**Die Bestellliste bleibt zweistufig: Kategorie-Pills über einer flachen
Variantenliste. PR #111 wird nicht übernommen.**

- **Der Gewinn ist nicht belegt.** Vor wie nach dem Umbau stehen zehn
  Variantenzeilen am Bildschirm.
- **Die Kosten sind belegt** und liegen auf dem häufigsten Pfad überhaupt: mehr
  Taps, ein Bildschirmwechsel je Produkt, nie zwei Produkte gleichzeitig
  sichtbar.
- **Das echte Problem löst die Reihenfolge.** Wer scrollt, sucht ein Produkt
  weiter unten; seit #109 bestimmt der Admin diese Position.
- **Ein-Varianten-Produkte** würden für die Einheitlichkeit einen Tap ohne
  Gegenwert zahlen. Ein Direktweg wäre die Reparatur eines selbst erzeugten
  Problems.
- **Produkt-Konservatismus.** Ein unbelegter Umbau des Kernpfads für
  ehrenamtliche Teams unter Zeitdruck wird im Zweifel weggelassen.

Das ist keine Wertung des Beitrags. Der belegte Teil desselben Branches — der
Umbruch der Variantennamen — ist über #110 mit Autorschaft übernommen.

## Konsequenzen

- Der Eigentümer schließt #111 mit Verweis auf dieses ADR. Ein Merge oder
  Cherry-Pick des Produktebenen-Commits findet nicht statt.
- Aus dieser Entscheidung folgt keine Codeänderung. `ProductList.tsx`, die
  e2e-Suite, `docs/handbuch.md` § 6.3, `website/src/components/LiveDemo.tsx` und
  die Website-Screenshots bleiben unverändert und beschreiben weiter den
  geltenden Stand.
- Die Kategorie-Pills bleiben die einzige Filterebene des Bestell-Bildschirms.
  Gegen lange Listen wirkt die Admin-Reihenfolge, nicht eine Navigationsebene.
- **Wieder aufgreifen, wenn** ein Verein aus dem laufenden Einsatz berichtet,
  dass die Reihenfolge das Scrollen nicht löst — mit Zahlen: Produkte je
  Kategorie, Varianten je Produkt (ab etwa 15) und Bestellungen je Tisch. Dann
  entscheidet ein Feldtest, nicht ein Code-Review.
- Ein solcher Entwurf müsste den Direktweg für Ein-Varianten-Produkte von
  vornherein mitbringen und die e2e-Helfer mitziehen.
- Eine Produktebene bräuchte dann ein neues ADR. Dieses hier wird nicht
  umgeschrieben.
