# ADR 04: Warn-Bestätigung für irreversible Routine-Aktionen

- **Status:** akzeptiert (2026-07-13)
- **Kontext-Dokumente:** PRD `prd-ui-audit-politur.md` und Umsetzungsplan
  `plan-ui-audit-politur.md` (Phase 9, Befund NEU07) — nach Merge gelöscht,
  siehe Git-Historie

## Kontext

jotti hatte für die Bestätigung „Kasse endgültig abschließen" (Kassensturz +
Tagesabschluss/Z-Bon in einem Schritt) den Button-Variant `destructive`
(rote Soft-Tint-Fläche) verwendet — sowohl für die auslösende Aktion als auch
für den `AlertDialogAction` im Bestätigungsdialog.

Der Kassenabschluss ist zwar **irreversibel** (er lässt sich nicht rückgängig
machen), aber eine **routinemäßige, erwartete** Handlung am Ende jedes
Kassentags — kein gefährlicher, destruktiver Vorgang wie das Löschen von
Daten. Das Mehr-Experten-UI-Audit vom 13.07.2026 (NEU07) hat festgehalten,
dass die rote Einfärbung die falsche Bedeutung transportiert: Rot signalisiert
„gefährlich/Fehler/Verlust", nicht „irreversibel, aber normal". Der Nutzer soll
innehalten und bewusst bestätigen, aber nicht das Gefühl bekommen, etwas
kaputtzumachen.

Die Farbebene kannte bis dahin nur zwei Aktions-Signale: primär-Grün
(`--primary`) für die reguläre Bestätigung und destruktiv-Rot (`--destructive`)
für gefährliche/löschende Aktionen. Für „irreversibel, aber routine" fehlte ein
eigenes Signal.

Erwogene Alternativen:

1. **Bei `destructive` (Rot) bleiben** — überlädt Rot mit zwei Bedeutungen
   (gefährlich *und* irreversibel-routine) und lässt beide ununterscheidbar.
2. **Primär-Grün verwenden** — nivelliert den Kassenabschluss zu einer
   beliebigen Bestätigung; das gewünschte Innehalten geht verloren.
3. **Eigenes Warn-Treatment (Amber)** — ein drittes, distinktes Aktions-Signal
   zwischen Grün und Rot, das „Achtung, irreversibel" ohne „gefährlich" trägt.

## Entscheidung

Ein eigenes **Warn-Treatment (Amber)** wird als Design-Token
(`--warn` / `--warn-foreground` in `frontend/src/index.css`, `:root` **und**
`.dark`) und als Button-Variant `warn`
(`frontend/src/components/ui/button.tsx`) eingeführt (Alternative 3). Die
Bestätigung „Kasse endgültig abschließen" nutzt es statt `destructive` —
auslösender Button und `AlertDialogAction` gleichermaßen.

Eigenschaften des Tokens:

- **Distinkt** von destruktiv-Rot und primär-Grün: solide Amber-Fläche
  (amber-500) mit dunklem Text (amber-950).
- **WCAG AA** in Light und Dark: die solide Amber-Fläche trägt ihren
  Text-Kontrast selbst (≈ 7:1, AA/AAA), unabhängig vom Theme-Hintergrund.

**Bindende Regel für künftige Bestätigungsdialoge:** Aktionen, die
**irreversibel, aber routinemäßig** sind (erwarteter Teil des normalen
Ablaufs, kein Datenverlust im Fehler-Sinn), nutzen das Warn-Treatment
(`variant="warn"`). Destruktiv-Rot (`variant="destructive"`) bleibt
gefährlichen, löschenden Aktionen vorbehalten (z. B. Tisch/Produkt/Benutzer
löschen). Primär-Grün bleibt reversiblen bzw. unkritischen Bestätigungen
vorbehalten.

## Konsequenzen

- Es gibt drei klar getrennte Aktions-Signale: Grün (regulär), Amber
  (irreversibel-routine), Rot (gefährlich/löschend). Jede Farbe trägt genau
  eine Bedeutung.
- Neue irreversible Routine-Bestätigungen (falls künftig hinzukommend) folgen
  diesem Muster ohne erneute Design-Entscheidung.
- Der automatisierte WCAG-AA-Kontrast-Durchlauf
  (`e2e/tests/admin-kontrast-axe.spec.ts`) deckt den Kassenabschluss-Screen in
  Light und Dark ab und sichert damit den AA-Kontrast des Warn-Buttons mit ab.
- Das Warn-Token ist bewusst nur für Aktionsflächen (Buttons) gedacht; die
  Warn-*Hinweiskarten* (`WarnKarte`) bleiben unberührt und tragen weiter ihre
  eigene, in Phase 8 auf AA gebrachte Einfärbung.
