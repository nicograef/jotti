# PRD: Einheitlicher Positionsname „Produkt Variante" (ohne Klammern)

## Problem Statement

Eine Bestellposition besteht in jotti aus einem Produkt (z. B. „Pommes", „Bier")
und einer Variante (z. B. „mit Ketchup", „mit Mayo", „groß", „klein"). Produkt-
und Variantenname sind bewusst so angelegt, dass sie zusammengesetzt einen
natürlichen, vollständigen Positionsnamen ergeben: „Pommes mit Ketchup",
„Bier groß".

Heute ist die Darstellung dieses Positionsnamens uneinheitlich:

- Der **Bondruck** (ESC/POS-Formatter) druckt Positionen als
  `{Produktname} ({Variantenname})` — also mit Klammern, z. B.
  „Pommes (mit Ketchup)". Das wirkt gestelzt und entspricht nicht der
  natürlichen Sprechweise.
- In der **Auswertung/Reporting**-Ansicht (Live-Reporting, Reporting-Ergebnisse)
  werden Positionen ebenfalls mit Klammern dargestellt:
  `{Produktname} ({Variantenname})`.
- Beim **Bestellung-Aufnehmen** (BestellungDrawer → Receipt) wird auf dem
  Bon-Vorschau-Beleg nur der **Variantenname** angezeigt (z. B. „mit Ketchup"
  statt „Pommes mit Ketchup"), wodurch unklar ist, um welches Produkt es geht.
- Andere Drawer (Zahlung, Ausgabe, Stornierung, Umbuchung, Direktverkauf-Storno)
  zeigen bereits korrekt `{Produktname} {Variantenname}` mit Leerzeichen — aber
  jede Stelle implementiert die Zusammensetzung selbst, sodass die Regel
  mehrfach (und teils abweichend) im Code dupliziert ist.

Der Anwender möchte überall in jotti denselben, natürlichen Positionsnamen sehen:
`{Produktname} {Variantenname}` mit einem einfachen Leerzeichen, ohne Klammern.

## Solution

Es gibt genau **eine kanonische Regel** für die Zusammensetzung des
Positionsnamens, die überall in jotti angewandt wird:

> **Positionsname = Produktname + " " + Variantenname**, mit einem einzelnen
> Leerzeichen verbunden und an den Rändern getrimmt. Keine Klammern, keine
> Sonderfälle, keine Zusammenfassung (Dedup) bei gleichlautenden Namen.

Beispiele:

- „Pommes" + „mit Ketchup" → **„Pommes mit Ketchup"**
- „Bier" + „groß" → **„Bier groß"**
- „Cola" + „Cola" → **„Cola Cola"** (verbatim, kein Zusammenfassen)
- „Maß Bier" + „" (leer) → **„Maß Bier"** (durch Trim, kein Trailing-Space)

Diese Regel wird in genau einer kleinen, testbaren Funktion pro Stack gekapselt
(Backend: Methode auf der Position; Frontend: Helper-Funktion). Alle
Anzeige- und Ausgabepfade rufen diese Funktion auf:

- **Bondruck**: Arbeitsbon (einzeln), Sammelbon, Direktverkauf-Abholbon und der
  fiskalische Kassenbeleg/Stornobeleg drucken `{Produkt} {Variante}` ohne
  Klammern.
- **Reporting**: Live-Reporting und Reporting-Ergebnisse zeigen
  `{Produkt} {Variante}` ohne Klammern.
- **Bestellung aufnehmen**: Die Bon-Vorschau (Receipt) zeigt den vollständigen
  Positionsnamen inkl. Produktname statt nur der Variante.
- **TSE/DSFinV-K-processData**: Die Bezeichnung der signierten Transaktion
  verwendet dieselbe kanonische Funktion.
- Alle bereits korrekt darstellenden Drawer werden auf dieselbe Funktion
  umgestellt (visuell unverändert, aber zukünftig zentral änderbar).

## User Stories

1. Als Küchen-/Ausgabe-Mitarbeiter möchte ich auf dem Arbeitsbon „Pommes mit
   Ketchup" statt „Pommes (mit Ketchup)" lesen, damit der Bon wie natürliche
   Sprache klingt und schnell erfassbar ist.
2. Als Küchen-Mitarbeiter möchte ich auf dem Sammelbon alle Positionen im
   Format „{Produkt} {Variante}" ohne Klammern sehen, damit die Darstellung
   einheitlich zum Einzelbon ist.
3. Als Mitarbeiter am Direktverkauf möchte ich auf dem Abholbon Positionen im
   selben klammerlosen Format sehen wie überall sonst.
4. Als Gast/Käufer möchte ich auf dem fiskalischen Kassenbeleg „Bier groß"
   statt „Bier (groß)" lesen, damit der Beleg verständlich und natürlich wirkt.
5. Als Mitarbeiter, der einen Stornobeleg druckt, möchte ich die Positionen im
   selben klammerlosen Format sehen wie auf dem Original-Kassenbeleg.
6. Als Servicekraft, die eine Bestellung aufnimmt, möchte ich in der
   Bon-Vorschau (Receipt) den vollständigen Positionsnamen „Pommes mit Ketchup"
   sehen — nicht nur „mit Ketchup" —, damit ich vor dem Absenden sicher
   erkenne, welches Produkt bestellt wird.
7. Als Servicekraft möchte ich, dass die Bon-Vorschau exakt das anzeigt, was
   später auf dem gedruckten Bon steht, damit es keine Überraschungen gibt.
8. Als Vereinsverantwortlicher im Reporting möchte ich im Live-Reporting
   Positionen als „{Produkt} {Variante}" ohne Klammern sehen, damit die
   Auswertung mit den Belegen konsistent ist.
9. Als Vereinsverantwortlicher möchte ich in den Reporting-Ergebnissen
   (abgeschlossener Zeitraum) Positionen im selben klammerlosen Format sehen.
10. Als Servicekraft beim Kassieren (Zahlung) möchte ich Positionen weiterhin
    als „{Produkt} {Variante}" sehen — jetzt aus einer zentralen Regel, damit
    die Darstellung garantiert mit Bon und Reporting übereinstimmt.
11. Als Servicekraft bei der Ausgabe möchte ich Positionen im einheitlichen
    Format sehen.
12. Als Servicekraft bei einer Stornierung möchte ich die zu stornierenden
    Positionen im einheitlichen Format sehen.
13. Als Servicekraft bei einer Umbuchung möchte ich die umzubuchenden
    Positionen im einheitlichen Format sehen.
14. Als Mitarbeiter beim Direktverkauf-Storno möchte ich die Positionen und die
    zugehörigen Bedien-Beschriftungen (aria-label „… verringern/hinzufügen") im
    einheitlichen, klammerlosen Format vorfinden.
15. Als Finanz-/Prüfungsverantwortlicher möchte ich, dass die in der TSE
    signierte Positionsbezeichnung dieselbe Zusammensetzungsregel verwendet wie
    Bon und UI, damit es genau eine Quelle der Wahrheit für Positionsnamen gibt.
16. Als Entwickler möchte ich die Zusammensetzungsregel an genau einer Stelle
    pro Stack (Backend/Frontend) ändern können, damit künftige
    Format-Anpassungen nicht über Dutzende Dateien verteilt werden müssen.
17. Als Entwickler möchte ich eine kleine, isoliert testbare Funktion für die
    Positionsbezeichnung haben, damit ich die Regel (inkl. Randfälle) eindeutig
    abdecken kann.
18. Als Anwender mit einem Produkt, dessen Variante genauso heißt wie das
    Produkt selbst, akzeptiere ich, dass „Cola Cola" gedruckt wird — die Regel
    bleibt bewusst einfach und ohne Sonderfall.

## Implementation Decisions

### Kanonische Regel

- Positionsname = `Produktname` + einzelnes Leerzeichen + `Variantenname`,
  anschließend an den Rändern getrimmt.
- **Keine** Zusammenfassung/Dedup, wenn Variantenname leer ist oder dem
  Produktnamen entspricht (bewusste Entscheidung: verbatim verbinden). Das
  Trimmen verhindert lediglich ein überflüssiges Leerzeichen, wenn ein
  Variantenname (etwa in Reporting-Daten ohne Mindestlänge) leer ist.
- **Keine Klammern.**

### Backend-Modul (Go)

- Es wird eine Methode auf der Positions-Struktur im `kasse`-Domänenpaket
  eingeführt (Arbeitstitel: `Position.Bezeichnung()`), die genau die kanonische
  Regel umsetzt und den zusammengesetzten Namen als String zurückgibt.
- Dieses ist ein **tiefes, schmales Modul**: triviale, stabile Signatur
  (`() string`), kapselt die einzige Stelle, an der Produkt- und Variantenname
  verbunden werden.
- Konsumenten im Backend:
  - **Bondruck-Formatter** (ESC/POS): Die drei Stellen, die heute
    `"%dx %s (%s)"` mit `ProduktName`/`VarianteName` formatieren
    (Einzel-Arbeitsbon, Sammelbon, fiskalischer Kassenbeleg), verwenden künftig
    `"%dx %s"` mit `pos.Bezeichnung()`.
  - **TSE/DSFinV-K-processData**: Die inline-Zusammensetzung der Bezeichnung
    wird durch `pos.Bezeichnung()` ersetzt. Die TSE-spezifischen Aspekte
    (CSV-Quoting/Anführungszeichen-Verdopplung, `"Unbekannt"`-Fallback bei
    komplett leerem Namen) bleiben als CSV-Schicht **erhalten** und umschließen
    die kanonische Funktion.
- **Fiskale Konsequenz** (bewusst akzeptiert): Da die bisherige TSE-Logik bei
  variante == produkt zusammenfasste, ändert sich die signierte Bezeichnung in
  diesem (in der Praxis nicht vorgesehenen) Sonderfall künftig von z. B. „Cola"
  zu „Cola Cola". Variantennamen sind Pflicht und mind. 3 Zeichen lang; ein
  leerer Variantenname kommt im `kasse`-Pfad nicht vor. Bereits signierte
  (immutable) Events bleiben unverändert; betroffen sind nur neue
  Zusammensetzungen.

### Frontend-Modul (TypeScript)

- Es wird eine reine Helper-Funktion eingeführt (Arbeitstitel:
  `formatPositionName(produktName, varianteName): string`), angesiedelt bei den
  übrigen Formatierungs-Utilities (dort, wo `formatCents` lebt), die exakt die
  kanonische Regel umsetzt.
- Konsumenten im Frontend (alle Anzeigepfade rufen den Helper):
  - **Reporting**: Live-Reporting und Reporting-Ergebnisse ersetzen die
    bedingte Klammer-Darstellung `{produktName}{varianteName ? " (…)" : ""}`
    durch den Helper.
  - **Bestellung aufnehmen**: Die Mapping-Funktion, die aus `Produkt[]` +
    ausgewählten Mengen die Receipt-Positionen erzeugt, setzt das `name`-Feld
    künftig auf `formatPositionName(produkt.name, variante.name)` statt nur auf
    den Variantennamen. Die `Receipt`-Komponente bleibt unverändert (rendert
    weiterhin `position.name`).
  - **Bereits korrekte Drawer** (Zahlung, Ausgabe, Historie-Stornierung,
    Historie-Umbuchung, Direktverkauf-Storno inkl. aria-labels) sowie die
    bestehende `toReceiptItems`-Hilfsfunktion stellen ihre inline-Komposition
    `{produktName} {varianteName}` auf den Helper um. Visuell unverändert; Ziel
    ist die zentrale Wartbarkeit.
- **Auswahl-UI bleibt unverändert**: Die Produktliste (ProductList) zeigt das
  Produkt als aufklappbaren Gruppen-Header und die Varianten als auswählbare
  Unterzeilen. Ein zusammengesetzter Name wäre hier redundant; diese Ansicht
  wird nicht geändert.

### Konsistenz / Single Source of Truth

- Nach der Umsetzung existiert die Zusammensetzungsregel pro Stack an genau
  einer Stelle. Eine künftige Änderung des Formats (z. B. Trennzeichen) ist eine
  Ein-Zeilen-Änderung pro Stack.

## Testing Decisions

Ein guter Test prüft **externes Verhalten**, nicht die Implementierung: also den
zurückgegebenen/gerenderten String bzw. die erzeugte Bon-Ausgabe — nicht, ob
eine bestimmte Hilfsfunktion aufgerufen wurde.

Zu testende Module:

- **Backend `Position.Bezeichnung()`**: Unit-Tests für
  - Normalfall: „Pommes" + „mit Ketchup" → „Pommes mit Ketchup".
  - Gleichlautend: „Cola" + „Cola" → „Cola Cola" (kein Dedup).
  - Leerer Variantenname: „Maß Bier" + „" → „Maß Bier" (Trim, kein
    Trailing-Space).
  - Prior Art: Domänen-Tests im `kasse`-Paket (z. B. tisch_session_test.go).
- **Backend Bondruck-Formatter**: Die bestehenden Golden-/String-Assertions im
  ESC/POS-Formatter-Test werden vom Klammer-Format auf das klammerlose Format
  aktualisiert (Einzel-Arbeitsbon, Sammelbon, Kassenbeleg, Stornobeleg). Es wird
  geprüft, dass die Bon-Ausgabe „Pommes mit Ketchup" bzw. „… {Produkt}
  {Variante}" enthält und keine Klammern mehr um die Variante.
  - Prior Art: bestehender formatter_test.go.
- **Backend TSE-processData**: Die bestehenden processdata-Tests werden geprüft
  und ggf. an das verbatim-Verhalten angepasst (die vorhandenen Fälle mit leerer
  bzw. abweichender Variante bleiben unverändert; ein etwaiger Fall mit
  identischer Variante würde sich auf „… …"-verbatim ändern).
  - Prior Art: bestehender processdata_test.go.
- **Frontend `formatPositionName`**: Unit-Tests für Normalfall, leeren
  Variantennamen (nur Produktname, kein Trailing-Space) und gleichlautende
  Namen.
  - Prior Art: lib-/utils-nahe Tests (z. B. drawerUtils.test.ts,
    errorMessages.test.ts).
- **Frontend Bestellaufnahme-Mapping**: Test, dass die Mapping-Funktion für die
  Bon-Vorschau jetzt den zusammengesetzten Namen „{Produkt} {Variante}" im
  `name`-Feld liefert (statt nur der Variante).
  - Prior Art: drawerUtils.test.ts.

## Out of Scope

- Die Produkt-/Varianten-Auswahl-UI (ProductList): Produkt-Header + Varianten als
  Unterzeilen bleiben unverändert.
- Produktanlage/-bearbeitung (NewProductDialog/EditProductDialog): keine neuen
  Benennungs-Hinweise oder Validierungen für „komponierbare" Namen.
- Keine Änderung der Validierungsregeln (Variantenname Pflicht, Mindest-/
  Maximallänge) für Produkte oder Varianten.
- Keine rückwirkende Änderung bereits gespeicherter (immutabler) Events oder
  bereits signierter TSE-Transaktionen; betroffen sind nur künftig erzeugte
  Bons, UI-Darstellungen und Signaturen.
- Keine Änderung an Preis-/Steuerzeilen des Kassenbelegs; nur die Artikel-/
  Positionszeile (Bezeichnung) ändert sich.

## Further Notes

- Die Lösung setzt voraus, dass Vereine Produkt- und Variantennamen bewusst so
  wählen, dass sie sich natürlich zu einem vollständigen Positionsnamen
  zusammensetzen („Pommes" + „mit Ketchup"). Dies entspricht der heutigen
  Praxis. Eine optionale Live-Vorschau des zusammengesetzten Namens in der
  Produktanlage könnte Anwender künftig dabei unterstützen — bewusst Out of
  Scope, aber als Folgeidee notiert.
- Der gleichlautende Sonderfall („Cola" + „Cola" → „Cola Cola") ist eine
  bewusste, dokumentierte Akzeptanz zugunsten einer einfachen, sonderfallfreien
  Regel.
