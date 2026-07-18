import type { Locator, Page } from '@playwright/test'
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

// erwarteBuendigeKategorieleisteImSplit prüft am gerenderten DOM beide Symptome
// der Split-Layout-Regression (ab lg) an der scrollenden Auswahl-Spalte:
//   1. kein horizontaler Überlauf — scrollWidth ≤ clientWidth der Spalte. Der
//      Vollbreiten-Ausbruch der Leiste träte als Scrollbalken der Spalte auf,
//      nicht am Dokument (die Spalte ist overflow-y-auto, ⇒ overflow-x rechnet
//      zu auto), daher wird die Spalte gemessen, nicht document.scrollingElement.
//   2. bündiges Kleben — nach dem Scrollen liegt die Oberkante der
//      Kategorieleiste an der Oberkante der Spalte (Offset ≈ 0, nicht ≈ 56).
// Startpunkt ist ein Kategorie-Chip (zugänglich, produktiv) statt eines
// Test-Hooks; von dort wird zur klebenden Leiste (position: sticky) und zur
// scrollenden Spalte (overflow-y auto/scroll) hochgelaufen — kein Markup-
// Eingriff, keine Klassen-Selektion. Die Spalte muss vertikal überlaufen, sonst
// wäre die Klebe-Prüfung nicht aussagekräftig; das sichert die Vorbedingung
// (scrollTop bewegt sich beim Scrollen ans Ende) ab.
export async function erwarteBuendigeKategorieleisteImSplit(
  kategorieChip: Locator,
  screen: string,
): Promise<void> {
  const messung = await kategorieChip.evaluate((start) => {
    const istScrollbar = (el: HTMLElement) => {
      const overflowY = getComputedStyle(el).overflowY
      return overflowY === 'auto' || overflowY === 'scroll'
    }
    // Nächster Vorfahr mit position: sticky = die klebende Kategorieleiste.
    let leiste: HTMLElement | null = start.parentElement
    while (leiste && getComputedStyle(leiste).position !== 'sticky') {
      leiste = leiste.parentElement
    }
    if (!leiste) return null
    // Nächster Vorfahr mit scrollbarem overflow-y = die Auswahl-Spalte.
    let spalte: HTMLElement | null = leiste.parentElement
    while (spalte && !istScrollbar(spalte)) {
      spalte = spalte.parentElement
    }
    if (!spalte) return null

    const scrollWidth = spalte.scrollWidth
    const clientWidth = spalte.clientWidth
    // Bis ans Ende scrollen, damit die klebende Leiste ihren Klebe-Offset
    // tatsächlich einnimmt (bei scrollTop 0 läge sie ohnehin bündig).
    spalte.scrollTop = spalte.scrollHeight
    const scrollTop = spalte.scrollTop
    const klebeOffset =
      leiste.getBoundingClientRect().top - spalte.getBoundingClientRect().top
    return { scrollWidth, clientWidth, scrollTop, klebeOffset }
  })

  if (messung === null) {
    throw new Error(
      `${screen}: klebende Kategorieleiste oder scrollende Auswahl-Spalte nicht gefunden`,
    )
  }

  // Beide Symptome unabhängig prüfen (soft): Der Überlauf-Fehlschlag darf den
  // Totzonen-Fehlschlag nicht maskieren (und umgekehrt) — im Fehlerfall werden
  // beide Symptome zugleich gemeldet, und jede Prüfung ist nachweislich
  // wirksam, statt dass die zweite hinter der ersten verborgen bleibt.
  expect
    .soft(
      messung.scrollWidth,
      `${screen}: scrollWidth ${messung.scrollWidth.toString()} der Auswahl-Spalte darf clientWidth ${messung.clientWidth.toString()} nicht überschreiten`,
    )
    .toBeLessThanOrEqual(messung.clientWidth)

  // Vorbedingung hart: Ohne vertikalen Scroll (scrollTop 0) läge die Leiste
  // ohnehin an der Oberkante — die Klebe-Prüfung wäre dann bedeutungslos.
  expect(
    messung.scrollTop,
    `${screen}: Vorbedingung — die Auswahl-Spalte muss vertikal scrollen (scrollTop > 0), sonst ist die Klebe-Prüfung nicht aussagekräftig`,
  ).toBeGreaterThan(0)

  // Bündig: Offset ≈ 0 (kleine Sub-Pixel-Toleranz), keine 56-px-Totzone.
  expect
    .soft(
      Math.abs(messung.klebeOffset),
      `${screen}: Kategorieleiste muss bündig oben an der Spalte kleben (Offset ${messung.klebeOffset.toString()} px ≈ 0, nicht ≈ 56)`,
    )
    .toBeLessThanOrEqual(2)
}
