import type { Page } from '@playwright/test'
import { expect } from '@playwright/test'

// erwarteKeinenHorizontalenUeberlauf misst am gerenderten DOM, ob die Seite
// horizontal überläuft: scrollWidth des Scroll-Wurzelelements gegen die
// Viewport-Breite (innerWidth). Verhaltensbasiert statt Klassennamen-Prüfung.
export async function erwarteKeinenHorizontalenUeberlauf(
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
