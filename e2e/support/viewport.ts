import type { Locator, Page } from '@playwright/test'
import { expect } from '@playwright/test'

// Misst scrollWidth des Scroll-Wurzelelements gegen innerWidth. scrollWidth
// kommt roh (kein "?? 0"): fehlt document.scrollingElement, ginge eine
// stillschweigende 0 als "kein Überlauf" durch — die Vorbedingung
// toBeGreaterThan(0) deckt das auf.
export async function erwarteKeinenHorizontalenUeberlauf(
  page: Page,
  screen: string,
): Promise<void> {
  const { scrollWidth, innerWidth } = await page.evaluate(() => ({
    scrollWidth: document.scrollingElement?.scrollWidth,
    innerWidth: window.innerWidth,
  }))
  expect(
    scrollWidth,
    `${screen}: Vorbedingung — document.scrollingElement muss existieren und eine scrollWidth tragen, sonst ist die Überlauf-Prüfung nicht aussagekräftig`,
  ).toBeGreaterThan(0)
  expect(
    scrollWidth,
    `${screen}: scrollWidth ${String(scrollWidth)} darf innerWidth ${innerWidth.toString()} nicht überschreiten`,
  ).toBeLessThanOrEqual(innerWidth)
}

// Prüft an der scrollenden Auswahl-Spalte (ab lg) zwei Symptome: kein
// horizontaler Überlauf (scrollWidth ≤ clientWidth) und bündiges Kleben der
// Kategorieleiste (Offset ≈ 0, nicht ≈ 56). Gemessen wird die Spalte, nicht
// document.scrollingElement: sie ist overflow-y-auto, ⇒ overflow-x rechnet zu
// auto, der Vollbreiten-Ausbruch träte also als Scrollbalken der Spalte auf.
// Startpunkt ist ein Kategorie-Chip statt eines Test-Hooks.
export async function erwarteBuendigeKategorieleisteImSplit(
  kategorieChip: Locator,
  screen: string,
): Promise<void> {
  const messung = await kategorieChip.evaluate((start) => {
    const istScrollbar = (el: HTMLElement) => {
      const overflowY = getComputedStyle(el).overflowY
      return overflowY === 'auto' || overflowY === 'scroll'
    }
    let leiste: HTMLElement | null = start.parentElement
    while (leiste && getComputedStyle(leiste).position !== 'sticky') {
      leiste = leiste.parentElement
    }
    if (!leiste) return null
    let spalte: HTMLElement | null = leiste.parentElement
    while (spalte && !istScrollbar(spalte)) {
      spalte = spalte.parentElement
    }
    if (!spalte) return null

    const scrollWidth = spalte.scrollWidth
    const clientWidth = spalte.clientWidth
    // Ans Ende scrollen: bei scrollTop 0 läge die Leiste ohnehin bündig.
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

  // Soft, damit sich die beiden Symptome im Fehlerfall nicht maskieren.
  expect
    .soft(
      messung.scrollWidth,
      `${screen}: scrollWidth ${messung.scrollWidth.toString()} der Auswahl-Spalte darf clientWidth ${messung.clientWidth.toString()} nicht überschreiten`,
    )
    .toBeLessThanOrEqual(messung.clientWidth)

  // Hart: ohne vertikalen Scroll wäre die Klebe-Prüfung bedeutungslos.
  expect(
    messung.scrollTop,
    `${screen}: Vorbedingung — die Auswahl-Spalte muss vertikal scrollen (scrollTop > 0), sonst ist die Klebe-Prüfung nicht aussagekräftig`,
  ).toBeGreaterThan(0)

  expect
    .soft(
      Math.abs(messung.klebeOffset),
      `${screen}: Kategorieleiste muss bündig oben an der Spalte kleben (Offset ${messung.klebeOffset.toString()} px ≈ 0, nicht ≈ 56)`,
    )
    .toBeLessThanOrEqual(2)
}

// Ungekürzt heißt scrollWidth ≤ clientWidth des Namensknotens: eine CSS-Kürzung
// (overflow hidden + nowrap) ließe scrollWidth wachsen, während textContent den
// vollen Namen trägt — der Text allein beweist die Lesbarkeit nicht.
export async function erwarteVollstaendigLesbarenNamen(
  nameKnoten: Locator,
  erwarteterName: string,
  screen: string,
): Promise<void> {
  const messung = await nameKnoten.evaluate((el) => ({
    scrollWidth: el.scrollWidth,
    clientWidth: el.clientWidth,
    text: (el.textContent ?? '').trim(),
  }))

  expect(messung.text, `${screen}: gemessener Knoten trägt den Namen`).toBe(
    erwarteterName,
  )
  expect(
    messung.scrollWidth,
    `${screen}: „${erwarteterName}" wird gekürzt — scrollWidth ${messung.scrollWidth.toString()} überschreitet clientWidth ${messung.clientWidth.toString()}`,
  ).toBeLessThanOrEqual(messung.clientWidth)
}

// Die beiden Größen, die der erste Tap auf eine Variantenzeile nicht verändern
// darf: Oberkante der Folgezeile und Breite des Namensknotens. Dokumentbezogen
// (scrollY eingerechnet), damit Scrollen zwischen zwei Messungen nicht
// verfälscht.
export async function zeilenGeometrie(
  folgeZeile: Locator,
  nameKnoten: Locator,
): Promise<{ folgeZeileY: number; nameBreite: number }> {
  const folgeZeileY = await folgeZeile.evaluate(
    (el) => el.getBoundingClientRect().top + window.scrollY,
  )
  const nameBreite = await nameKnoten.evaluate(
    (el) => el.getBoundingClientRect().width,
  )
  return { folgeZeileY, nameBreite }
}
