import { expect, test } from '@playwright/test'

import { anmelden } from '../support/anmelden'
import { resetAndSeed } from '../support/seed'
import { oeffneTisch } from '../support/servicekraft'
import { erwarteBuendigeKategorieleisteImSplit } from '../support/viewport'

// Regression der Sticky-Kategorieleiste im Zweispalten-Service-Layout (ab lg).
// Die Leiste klebt bündig an der Oberkante der Auswahl-Spalte, nicht 56 px tiefer
// (top-14). Ihr Vollbreiten-Ausbruch erzeugt keinen horizontalen Scrollbalken der
// Spalte. Unter lg deckt tischservice-viewport-ueberlauf.mobile.spec.ts den
// Überlauf ab. Geprüft werden beide Bildschirme mit ProductList: Tisch-„Bestellen"
// und Direktverkauf.

// Zwei Regime ≥ lg: 1024 px (lg, nicht xl) und 1280 px (xl). Eine an xl gebundene
// Regel oder eine von der useIsMobile-Schwelle (1024 px) abweichende lg-Grenze
// fällt so bei 1024 px auf.
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
