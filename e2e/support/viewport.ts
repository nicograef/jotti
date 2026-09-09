import type { Locator, Page } from '@playwright/test'
import { expect } from '@playwright/test'

// erwarteKeinenHorizontalenUeberlauf misst am gerenderten DOM, ob die Seite
// horizontal überläuft: scrollWidth des Scroll-Wurzelelements gegen die
// Viewport-Breite (innerWidth). Verhaltensbasiert statt Klassennamen-Prüfung.
// scrollWidth kommt roh aus der Messung (kein "?? 0"): fehlt
// document.scrollingElement, wäre eine stillschweigende 0 kleiner als jede
// innerWidth und ließe die fehlgeschlagene Messung als "kein Überlauf"
// durchgehen. Die Vorbedingung toBeGreaterThan(0) deckt genau das auf.
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

// erwarteVollstaendigLesbarenNamen prüft am gerenderten DOM, dass ein
// Variantenname ungekürzt in seiner Zeile steht: scrollWidth ≤ clientWidth des
// Namensknotens. Eine CSS-Kürzung (overflow hidden + white-space nowrap) ließe
// den scrollWidth über die Boxbreite hinauswachsen, während textContent
// unverändert den vollen Namen trägt — der Text allein beweist die Lesbarkeit
// also nicht. Zusätzlich wird der Text geprüft, damit die Messung nachweislich
// am richtigen Knoten hängt.
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

// zeilenGeometrie liest die beiden Größen, die der erste Tap auf eine
// Variantenzeile nicht verändern darf: die Oberkante der Folgezeile (sie rutscht
// nach unten, sobald die getippte Zeile wächst) und die Breite des
// Namensknotens (sie schrumpft, wenn der Stepper beim Einblenden von Minus und
// Menge Platz vom Namen nimmt, wodurch der Name neu umbricht). Beide Werte sind
// dokumentbezogen (scrollY eingerechnet), damit ein Scrollen zwischen zwei
// Messungen sie nicht verfälscht.
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
