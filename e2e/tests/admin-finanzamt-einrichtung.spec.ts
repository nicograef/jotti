import type { Locator, Page } from '@playwright/test'
import { expect, test } from '@playwright/test'

import { anmelden } from '../support/anmelden'
import { resetAndSeed } from '../support/seed'

// Layout-Regression für die Finanzamt-Einrichtung (Phase 3, Befund #1). Die
// drei Einrichtungsschritte lagen in einem lg:grid-cols-3, das im max-w-4xl-
// Container zwischen 1024 und 1440px zu schmal wurde: die Aktion „Als erledigt
// markieren" und die ELSTER-Seriennummer wurden abgeschnitten. Jetzt greifen
// die drei Spalten erst ab xl (~1280px) und stapeln darunter vertikal, die
// WarnKarte hat min-w-0 und die Seriennummer ein eigenes, umbrechendes Feld.
//
// Der Seed lässt Schritt 3 (ELSTER-Kassenmeldung) offen: elsterGemeldetAm ist
// null, die Kassenidentität existiert. Die WarnKarte mit Aktion und
// Seriennummer wird daher gerendert.

// erwarteKeinenHorizontalenUeberlauf misst am gerenderten DOM, ob die Seite
// horizontal überläuft (verhaltensbasiert, keine Klassennamen-Prüfung).
async function erwarteKeinenHorizontalenUeberlauf(
  page: Page,
  screen: string,
): Promise<void> {
  const { scrollWidth, innerWidth } = await page.evaluate(() => ({
    scrollWidth: document.scrollingElement?.scrollWidth ?? 0,
    innerWidth: window.innerWidth,
  }))
  expect(
    scrollWidth,
    `${screen}: scrollWidth ${scrollWidth.toString()} darf innerWidth ${innerWidth.toString()} nicht überschreiten`,
  ).toBeLessThanOrEqual(innerWidth)
}

// erwarteVollstaendigImViewport prüft, dass der rechte Rand eines Elements die
// Viewport-Breite nicht überschreitet — das Element ist also nicht rechts
// abgeschnitten und damit klickbar.
async function erwarteVollstaendigImViewport(
  element: Locator,
  viewportBreite: number,
  screen: string,
): Promise<void> {
  await expect(element).toBeVisible()
  const box = await element.boundingBox()
  expect(box, `${screen}: Element hat keine Bounding-Box`).not.toBeNull()
  if (box === null) return
  expect(
    box.x + box.width,
    `${screen}: rechter Rand ${(box.x + box.width).toString()} darf Viewport-Breite ${viewportBreite.toString()} nicht überschreiten`,
  ).toBeLessThanOrEqual(viewportBreite)
}

test.describe('Finanzamt-Einrichtung bleibt bei jeder Breite bedienbar', () => {
  test('„Als erledigt markieren" und die ELSTER-Seriennummer sind bei 1440px und 1024px voll sichtbar, ohne Überlauf bei 390px', async ({
    page,
    request,
  }) => {
    const zugang = await resetAndSeed(request)
    await anmelden(page, zugang.admin)
    await page.goto('/admin/finanzamt')

    const markieren = page.getByRole('button', { name: 'Als erledigt markieren' })
    const seriennummerFeld = page.getByText(
      'Seriennummer des elektronischen Aufzeichnungssystems',
    )
    const seriennummer = page.locator('code')

    // Bei 1440px (3 Spalten aktiv) und 1024px (gestapelt) müssen Aktion und
    // Seriennummer vollständig innerhalb des Viewports liegen und das Feld darf
    // nicht über seine eigene Breite hinauslaufen (umbrechen statt kürzen).
    for (const width of [1440, 1024]) {
      await page.setViewportSize({ width, height: 900 })

      await erwarteVollstaendigImViewport(markieren, width, `Aktion ${width}px`)
      await expect(seriennummerFeld).toBeVisible()
      await expect(seriennummer).toBeVisible()

      const feldUeberlauf = await seriennummer.evaluate(
        (el) => el.scrollWidth - el.clientWidth,
      )
      expect(
        feldUeberlauf,
        `Seriennummer-Feld läuft bei ${width.toString()}px über (scrollWidth − clientWidth = ${feldUeberlauf.toString()})`,
      ).toBeLessThanOrEqual(1)

      await erwarteKeinenHorizontalenUeberlauf(page, `Finanzamt ${width}px`)
    }

    // Am schmalen Handy-Viewport (390px) darf die Seite nicht horizontal
    // überlaufen (Basis-Track grid-cols-1 gesetzt).
    await page.setViewportSize({ width: 390, height: 844 })
    await erwarteKeinenHorizontalenUeberlauf(page, 'Finanzamt 390px')

    // Die Aktion ist tatsächlich klickbar: bei 1024px anklicken markiert die
    // Meldung als erledigt und quittiert mit einem Toast.
    await page.setViewportSize({ width: 1024, height: 900 })
    await markieren.click()
    await expect(
      page.getByText('Kassenmeldung als erledigt markiert.'),
    ).toBeVisible()
  })
})
