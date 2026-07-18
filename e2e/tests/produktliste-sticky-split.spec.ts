import { expect, test } from '@playwright/test'

import { anmelden } from '../support/anmelden'
import { resetAndSeed } from '../support/seed'
import { oeffneTisch } from '../support/servicekraft'
import { erwarteBuendigeKategorieleisteImSplit } from '../support/viewport'

// Regressionstest für die Sticky-Kategorieleiste im Zweispalten-Service-Layout
// (ab lg, 1024px). Der Fehler trat ausschließlich ≥ lg auf: Die klebende
// Kategorieleiste hing 56px zu tief in die linke Auswahl-Spalte (top-14 statt
// bündig), sodass Produktzeilen dahinter verschwanden, und ihr Vollbreiten-
// Ausbruch (negative Außenränder ohne kompensierende Ancestor-Padding im Split)
// erzeugte einen horizontalen Scrollbalken der Spalte. Unter lg (390px) deckt
// bereits tischservice-viewport-ueberlauf.mobile.spec.ts das (unveränderte)
// Verhalten ab; dieser Spec ergänzt ausschließlich die ≥ lg-Perspektive. Beide
// Bildschirme, die sich die ProductList teilen (Tisch-„Bestellen" und
// Direktverkauf), werden geprüft — ein Fix korrigiert beide.

// Zwei Breakpoint-Regime ≥ lg, in denen der Ausbruch vor dem Fix
// unterschiedlich groß war und die der lg-Fix beide abdecken muss:
//   - 1024px (lg, nicht xl): Ausbruch war md:-mx-8 = 32px, plus 56px-Totzone.
//   - 1280px (xl):           Ausbruch war xl:-mx-12 = 48px, plus 56px-Totzone.
// Beide Breiten zu prüfen macht den Guard trennscharf: Ein fälschlich an xl
// gebundener Fix (oder ein Auseinanderdriften der lg-Grenze von der
// useIsMobile-Schwelle 1024px) bliebe bei 1024px nicht unbemerkt, und das Band
// 1024–1279 ist nicht länger ungedeckt.
const LG_BREITEN = [1024, 1280] as const

test.describe('Sticky-Kategorieleiste im Split-Layout ab lg', () => {
  for (const breite of LG_BREITEN) {
    test.describe(`bei ${breite.toString()}px`, () => {
      // Überschreibt den Pixel-7-Default des mobile-service-Projekts auf
      // Block-Ebene — analog zum 390px-Spec, nur in die andere Richtung. Die
      // Phone-Emulationsflags des Projekts (isMobile/touch/mobile-UA) sind hier
      // inert: Die App gatet das Split-Layout über window.innerWidth
      // (useIsMobile-Schwelle 1024px), nicht über das Playwright-isMobile-Flag.
      // Die 720px-Höhe hält die Auswahl-Spalte scrollbar (Vorbedingung der
      // Klebe-Prüfung). Kein neues Playwright-Projekt.
      test.use({ viewport: { width: breite, height: 720 } })

      test('Tisch-Bestellen: Leiste klebt bündig, kein Überlauf der Auswahl-Spalte', async ({
        page,
        request,
      }) => {
        const zugangsdaten = await resetAndSeed(request)
        await anmelden(page, zugangsdaten.serviceleitung)

        await oeffneTisch(page, 'Tisch 3')

        // Der Bestellen-Tab ist Default und rendert ab lg das Split-Layout mit
        // der ProductList als linker Auswahl-Spalte. Der Essen-Chip ist die
        // Default-Kategorie (Seed belegt Essen/Getränke/Sonstiges ⇒ die Leiste
        // rendert) und dient als zugänglicher Startpunkt für die DOM-Messung.
        const essenChip = page.getByRole('button', {
          name: 'Essen',
          exact: true,
        })
        await expect(essenChip).toBeVisible()

        await erwarteBuendigeKategorieleisteImSplit(
          essenChip,
          `Tisch-Bestellen @ ${breite.toString()}px`,
        )
      })

      test('Direktverkauf: Leiste klebt bündig, kein Überlauf der Auswahl-Spalte', async ({
        page,
        request,
      }) => {
        const zugangsdaten = await resetAndSeed(request)
        await anmelden(page, zugangsdaten.serviceleitung)

        await page.goto('/service/direktverkauf')
        await expect(page.getByRole('tab', { name: 'Verkaufen' })).toBeVisible()

        // Der Verkaufen-Tab ist Default und rendert ab lg dasselbe Split-Layout
        // mit der ProductList als linker Auswahl-Spalte.
        const essenChip = page.getByRole('button', {
          name: 'Essen',
          exact: true,
        })
        await expect(essenChip).toBeVisible()

        await erwarteBuendigeKategorieleisteImSplit(
          essenChip,
          `Direktverkauf @ ${breite.toString()}px`,
        )
      })
    })
  }
})
