# ADR 11: TSE-Einrichtungs-Wizard bleibt eine Datei

- **Status:** akzeptiert (2026-09-09)
- **Kontext-Dokumente:** `docs/plans/plan-jotti-audit-fixes.md` Phase 14 (nach
  Merge gelöscht, siehe Git-Historie);
  `frontend/src/admin/tse/TSEEinrichtungWizard.tsx`

## Kontext

`TSEEinrichtungWizard.tsx` ist mit 934 Zeilen die größte Datei unter
`frontend/src` (ohne `components/ui`). Die nächstgrößte Produktivdatei,
`admin/settings/DruckstationConfigPage.tsx`, hat 557 Zeilen. Die Datei trägt 17
Funktionen auf oberster Ebene: 13 Komponenten und vier Hilfsfunktionen
(`istUebernehmbar`, `istEinsatzbereit`, `brauchtPin`, `tssZustandKlartext`,
Zeilen 40–57 und 883). Exportiert ist genau eine: `TSEEinrichtungWizard`
(Zeile 72).

Vorgeschlagen ist ein Umbau in zwei Teilen: je Schritt eine eigene Datei, und
die fiskaly-Zugangsdaten über einen lokalen React-Context statt über Props.

### Wie die Zugangsdaten heute laufen

Die Wurzelkomponente hält `apiKey` und `apiSecret` als State (Zeilen 73–74).
Sechs Komponenten deklarieren die beiden Werte als Props:

| Komponente               | Props-Deklaration | Weitergaben ab der Wurzel |
| ------------------------ | ----------------- | ------------------------- |
| `ZugangsdatenSchritt`    | 146–147, 153–154  | 1 (Eingabefeld)           |
| `BefundSchritt`          | 202–203, 208–209  | 1                         |
| `UebernahmeSchritt`      | 280–281, 286–287  | 2                         |
| `NeueTseTrotzdemAnlegen` | 569–570, 573–574  | 2                         |
| `BestaetigungSchritt`    | 600–601, 606–607  | 2 oder 3                  |
| `PukReset`               | 454–455, 460–461  | 3                         |

Der längste Weg umfasst drei Weitergaben: Wurzel → `BefundSchritt`
(Zeilen 124–125) → `NeueTseTrotzdemAnlegen` (249–250) → `BestaetigungSchritt`
(588–589). Der zweite gleich lange Weg endet bei `PukReset`: Wurzel →
`BefundSchritt` → `UebernahmeSchritt` (240–241) → `PukReset` (379–380).

Der State muss an der Wurzel liegen, unabhängig vom Zuschnitt der Dateien.
Zeile 80 meldet mit `useOffenerVorgang(apiKey.trim() !== '' || apiSecret.trim()
!== '')` einen offenen Vorgang, solange etwas eingetippt ist. `zurueckZuZugangsdaten`
(Zeile 100) leert beide Felder und hängt an zwei Schritten (Zeilen 120 und 128).

### Die Schritte sind keine reinen Anzeige-Komponenten

`useOffenerVorgang` steht an fünf Stellen der Datei: 80 (Zugangsdaten), 298
(Admin-PIN), 472 (Admin-PUK), 617 (LIVE-Tippbestätigung) und 729 (die
zurückgegebenen Geheimnisse). Jeder Schritt hält also eigenen Geheimnis-State
und meldet ihn selbst an das Vorgangs-Register. Ein Context für die
Zugangsdaten erfasst genau einen dieser fünf Werte.

### Was die Tests sehen

`TSEEinrichtungWizard.test.tsx` (245 Zeilen, neun Tests) importiert in Zeile 13
ausschließlich `TSEEinrichtungWizard` und rendert jeden Schritt über die
Wurzel. Eine Aufteilung in Dateien ändert an dieser Suite nichts — sie gewinnt
keine Testbarkeit und verliert keine.

### Context im Frontend

Zwei lokale Contexts existieren: `components/theme-provider.tsx:23`
(`ThemeProviderContext`) und `service/components/ServiceDock.tsx:15`
(`DockSlotContext`). Beide überbrücken etwas, das Props nicht können — die
Theme-Wahl quer durch den Baum und ein Portal-Ziel. Ein Context für die
Zugangsdaten überbrückt drei typisierte Weitergaben.

### Erwogene Alternativen

1. **Je Schritt eine Datei plus lokaler Context für die Zugangsdaten** (der
   Vorschlag). Die 13 Komponenten müssten exportiert werden; aus einer privaten
   Datei würde eine Modulgrenze mit öffentlicher Oberfläche.
2. **Nur die Dateien aufteilen, Props behalten.** Kostet dieselben Exporte,
   spart keine Zeile und verteilt einen linearen Ablauf auf mehrere Dateien.
3. **Alles lassen.** Ein linearer Ablauf steht in Leserichtung in einer Datei;
   der Einstieg ist die exportierte Wurzel oben.

## Entscheidung

**Der Wizard bleibt eine Datei, die Zugangsdaten bleiben Props**
(Alternative 3).

- **Korrektheit:** Kein Fehler wird behoben. Die Weitergabe ist
  typgeprüft — eine vergessene Prop bricht den Build, ein vergessener
  Context-Provider bricht zur Laufzeit.
- **Einfachheit:** Ein Context tauscht drei sichtbare Weitergaben gegen eine
  unsichtbare Kopplung. Wer heute liest, wer den API-Key sieht, folgt sechs
  Props-Deklarationen; danach müsste er jede Komponente auf `use(...)` prüfen.
- **Konsistenz:** Die beiden vorhandenen Contexts lösen Probleme, für die Props
  nicht reichen. Ein dritter für einen Wert, der drei Ebenen tief reicht, setzt
  eine andere Regel.
- **Produkt-Konservatismus:** Die TSE-Einrichtung läuft einmal je Installation.
  Ein Umbau ohne Nutzerwirkung an der Stelle, die den API-Key des fiskaly-Kontos
  führt, ist der falsche Ort für Aufräumarbeit.

**Empfehlung: nie.**

## Konsequenzen

- `TSEEinrichtungWizard.tsx` bleibt die größte Frontend-Datei. Das ist keine
  Zusage, dass sie beliebig wachsen darf — die Auslöser unten setzen die Grenze.
- Neue Schritte kommen als weitere Komponente in dieselbe Datei, unterhalb der
  Wurzel und oberhalb der Hilfsfunktionen.
- Die zwölf Schritt-Komponenten bleiben privat; exportiert ist allein die
  Wurzel. Kein anderes Modul kann einen Schritt einzeln rendern, und keiner ist
  an eine fremde Stelle koppelbar.
- **Wieder aufgreifen, wenn** einer von drei Fällen eintritt: die Zugangsdaten
  erreichen eine vierte Weitergabe-Ebene; ein zweiter TSE-Anbieter bringt einen
  parallelen Zweig mit eigenen Schritten; oder ein Schritt wird außerhalb des
  Wizards gebraucht. Dann entscheidet der konkrete Zuschnitt, nicht die
  Zeilenzahl.
- Die Zeilenzahl allein ist kein Auslöser. Sie misst, wie lang ein Ablauf ist,
  nicht wie verschränkt er ist.
