import { expect, test } from '@playwright/test'

import { anmelden } from '../support/anmelden'
import { resetAndSeed } from '../support/seed'

// Regression für NEU14: Der Theme-Umschalter in der Admin-Sidebar hat ein
// stabiles Label (unabhängig vom aktiven Design) und bleibt nach einem
// Maus-Klick nicht mit haftendem Fokus/Highlight zurück, das im Dark Mode wie
// die aktive Navigations-Seite aussieht. Tastaturbedienung behält den Fokus.

test.describe('Theme-Umschalter (Admin-Sidebar)', () => {
  test('stabiles Label und kein haftendes Highlight nach Maus-Klick', async ({
    page,
    request,
  }) => {
    const zugangsdaten = await resetAndSeed(request)
    await anmelden(page, zugangsdaten.admin)
    await page.goto('/admin/auswertung')

    const toggle = page.getByRole('button', { name: 'Design wechseln' })
    await expect(toggle).toBeVisible()

    // Klick schaltet das Design um, das Label bleibt jedoch stabil.
    await toggle.click()
    await expect
      .poll(() =>
        page.evaluate(() =>
          document.documentElement.classList.contains('dark'),
        ),
      )
      .toBe(true)
    await expect(toggle).toHaveText('Design wechseln')

    // Zeiger weg: der Umschalter ruht neutral — kein data-active, keine
    // Accent-Fläche, kein haftender Fokus (das wäre das falsche Active-Highlight).
    await page.mouse.move(0, 0)
    const rest = await page.evaluate(() => {
      const t = [...document.querySelectorAll('button')].find((b) =>
        /Design wechseln/.test(b.textContent ?? ''),
      ) as HTMLElement | null
      return {
        dataActive: t?.getAttribute('data-active'),
        bg: t ? getComputedStyle(t).backgroundColor : null,
        focused: document.activeElement === t,
      }
    })
    expect(rest.dataActive).not.toBe('true')
    expect(rest.bg).toBe('rgba(0, 0, 0, 0)')
    expect(rest.focused).toBe(false)

    // Tastaturbedienung behält den Fokus (Orientierung bleibt erhalten).
    await toggle.focus()
    await page.keyboard.press('Enter')
    await expect(toggle).toBeFocused()
    await expect(toggle).toHaveText('Design wechseln')
  })
})
