import { expect, test } from '@playwright/test'

import { anmelden } from '../support/anmelden'
import { resetAndSeed } from '../support/seed'
import { oeffneTisch, waehleVariante } from '../support/servicekraft'

// Regression für NEU13: Bei offenem Service-Drawer (Radix-Dialog, modal) darf die
// Tab-Leiste des ServiceDock nicht mehr interaktiv oder per Tastatur erreichbar
// sein. Radix macht die Umgebung des Dialogs per aria-hidden + Fokus-Falle inert;
// dieser Test hält das Verhalten fest, damit es nicht unbemerkt zerbricht.

test.describe('Tab-Leiste bei offenem Drawer', () => {
  test('ist nicht interaktiv und nicht per Tastatur fokussierbar', async ({
    page,
    request,
  }) => {
    const zugangsdaten = await resetAndSeed(request)
    await anmelden(page, zugangsdaten.service)
    await oeffneTisch(page, 'Tisch 1')

    await page.getByRole('tab', { name: 'Bestellen' }).click()
    // Vor dem Öffnen sind die Tabs regulär erreichbar.
    await expect(page.getByRole('tab', { name: 'Kassieren' })).toBeVisible()

    await waehleVariante(page, 'Bratwurst', 'Normal')
    await page.getByRole('button', { name: /Bestellung überprüfen/ }).click()
    await expect(page.getByRole('dialog')).toBeVisible()

    // aria-hidden entfernt die Tab-Leiste aus dem Accessibility-Baum, solange der
    // Drawer offen ist: per Rolle ist kein Tab mehr auffindbar.
    await expect(page.getByRole('tab')).toHaveCount(0)

    // Zeigergesteuert ist die Leiste ebenfalls tot (pointer-events: none).
    const pointerEvents = await page.evaluate(() => {
      const tab = [...document.querySelectorAll('[role="tab"]')].find((t) =>
        t.textContent?.includes('Kassieren'),
      ) as HTMLElement | undefined
      return tab ? getComputedStyle(tab).pointerEvents : null
    })
    expect(pointerEvents).toBe('none')

    // Tastatur-Fokus bleibt im Dialog gefangen: mehrfaches Tab erreicht nie einen
    // Tab-Trigger.
    for (let i = 0; i < 12; i++) {
      await page.keyboard.press('Tab')
      const aufTab = await page.evaluate(
        () => document.activeElement?.getAttribute('role') === 'tab',
      )
      expect(aufTab).toBe(false)
    }

    // Nach dem Schließen ist die Tab-Leiste wieder regulär bedienbar.
    await page.getByRole('button', { name: 'Abbrechen' }).click()
    await expect(page.getByRole('dialog')).toBeHidden()
    await expect(page.getByRole('tab', { name: 'Kassieren' })).toBeVisible()
    await page.getByRole('tab', { name: 'Kassieren' }).click()
  })
})
