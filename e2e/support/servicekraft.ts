import type { Locator, Page } from '@playwright/test'
import { expect } from '@playwright/test'

// Zugängliche Selektoren statt Test-IDs; Ausnahme sind vollePositionsZeilen und
// tischSaldo, die keine zugängliche Alternative tragen (data-slot).

// Innerste Zeile (div) mit dem Text und einem Button dieses Namens; scope ist
// die Seite oder ein Drawer. Der „has"-Filter wird bewusst von der Seite aus
// gebaut: vom selben scope verkettet, matcht Playwright ihn nicht zuverlässig.
export function zeileMit(
  scope: Page | Locator,
  text: string,
  buttonName: string | RegExp,
): Locator {
  const page = 'page' in scope ? scope.page() : scope
  return scope
    .locator('div')
    .filter({ hasText: text })
    .filter({ has: page.getByRole('button', { name: buttonName }) })
    .last()
}

// Tippt die Historien-Zeile mit diesem Text an (z. B. „Bestellung … +2,50 €").
// Die ganze Zeile ist ein Button und öffnet den Detail-Drawer mit
// Umbuchen/Stornieren/Drucken; der Text muss sie eindeutig treffen (Betrag).
export async function oeffneHistorienDetail(
  page: Page,
  text: string | RegExp,
): Promise<Locator> {
  await page.getByRole('button', { name: text }).click()
  const drawer = page.getByRole('dialog')
  await expect(drawer).toBeVisible()
  return drawer
}

// Über die Hauptsuche, die alle aktiven Tische erfasst — unabhängig davon, ob
// der Tisch schon in „Meine Tische" steht.
export async function oeffneTisch(page: Page, tisch: string): Promise<void> {
  await page.goto('/service/tische')
  await page.getByPlaceholder('Tisch suchen').fill(tisch)
  await page
    .getByRole('button', { name: new RegExp(`^${tisch}\\b.*€`) })
    .click()
  await expect(page.getByRole('tab', { name: 'Bestellen' })).toBeVisible()
}

// Nur die aktive Kategorie steht im DOM; Essen ist per Default aktiv. Die Chips
// fehlen, wenn nur eine Kategorie belegt ist — daher jeder Wechsel tolerant.
const KATEGORIE_CHIPS = ['Essen', 'Getränke', 'Sonstiges']

// Variantennamen („Normal", „Klein") kommen in mehreren Produkten vor und stehen
// gleichzeitig im DOM — erst die Gruppe macht eine Zeile eindeutig.
export function produktGruppe(page: Page, produkt: string): Locator {
  return page
    .locator('div')
    .filter({ has: page.getByRole('heading', { name: produkt, exact: true }) })
    .filter({ has: page.getByRole('button', { name: 'Variante hinzufügen' }) })
    .last()
}

// Fügt die Variante zur Auswahl hinzu, ohne die Bestellung abzuschicken.
export async function waehleVariante(
  page: Page,
  produkt: string,
  variante: string,
  menge = 1,
): Promise<void> {
  const variantenZeile = zeileMit(
    produktGruppe(page, produkt),
    variante,
    'Variante hinzufügen',
  )
  // Der Produktname verrät die Kategorie nicht — reihum jeden Chip probieren,
  // bis die Zeile im DOM erscheint.
  if (!(await variantenZeile.isVisible().catch(() => false))) {
    for (const label of KATEGORIE_CHIPS) {
      const chip = page.getByRole('button', { name: label, exact: true })
      if (await chip.isVisible().catch(() => false)) {
        await chip.click()
        if (await variantenZeile.isVisible().catch(() => false)) break
      }
    }
  }
  await expect(variantenZeile).toBeVisible()

  for (let i = 0; i < menge; i++) {
    await variantenZeile
      .getByRole('button', { name: 'Variante hinzufügen' })
      .click()
  }
}

// Setzt einen offenen Tisch voraus; bestätigt die Bestellung im Drawer.
export async function bestellePosition(
  page: Page,
  produkt: string,
  variante: string,
  menge = 1,
): Promise<void> {
  await page.getByRole('tab', { name: 'Bestellen' }).click()

  await waehleVariante(page, produkt, variante, menge)

  await page.getByRole('button', { name: /Bestellung überprüfen/ }).click()
  const drawer = page.getByRole('dialog')
  await drawer.getByRole('button', { name: 'Bestellung aufnehmen' }).click()
  // .first(): Sonner kann bei mehreren Bestellungen kurzzeitig zwei Toasts mit
  // demselben Text im DOM halten (neuer Toast, alter noch beim Ausblenden).
  await expect(
    page.getByText('Bestellung wurde aufgenommen.').first(),
  ).toBeVisible()
}

// Neun Varianten → neun Positionen, Summe 52,00 €. Die Namen sind über alle
// Produkte hinweg eindeutig; die langen darunter („Fr: Schnitzel mit Pommes")
// liefern den nicht-umbrechenden Text für Footer- und Überlauf-Regressionen.
export const LANGE_BESTELLUNG_POSITIONEN: [
  produkt: string,
  variante: string,
][] = [
  ['Bratwurst', 'Normal'],
  ['Bratwurst', 'XXL'],
  ['Bratwurst', 'Currywurst'],
  ['Pommes', 'Klein'],
  ['Pommes', 'Groß'],
  ['Flammkuchen', 'Classic'],
  ['Flammkuchen', 'Speck & Zwiebel'],
  ['Flammkuchen', 'Mediterran'],
  ['Tagesgericht', 'Fr: Schnitzel mit Pommes'],
]

export async function nimmLangeBestellungAuf(page: Page): Promise<void> {
  await page.getByRole('tab', { name: 'Bestellen' }).click()
  for (const [produkt, variante] of LANGE_BESTELLUNG_POSITIONEN) {
    await waehleVariante(page, produkt, variante)
  }

  await page.getByRole('button', { name: /Bestellung überprüfen/ }).click()
  const bestellDrawer = page.getByRole('dialog')
  await expect(bestellDrawer.getByText('Flammkuchen Mediterran')).toBeVisible()
  await bestellDrawer
    .getByRole('button', { name: 'Bestellung aufnehmen' })
    .click()
  await expect(
    page.getByText('Bestellung wurde aufgenommen.').first(),
  ).toBeVisible()
}

export async function kassierePosition(
  page: Page,
  positionName: string,
  menge = 1,
): Promise<void> {
  await page.getByRole('tab', { name: 'Kassieren' }).click()
  const position = zeileMit(page, positionName, 'Produkt hinzufügen')
  await expect(position).toBeVisible()
  for (let i = 0; i < menge; i++) {
    await position.getByRole('button', { name: 'Produkt hinzufügen' }).click()
  }

  const kassierenLeiste = page.getByRole('button', { name: /Kassieren/ })
  await kassierenLeiste.click()

  const drawer = page.getByRole('dialog')
  await drawer.getByRole('button', { name: 'Kassieren' }).click()
  await expect(page.getByText('Zahlung erfolgreich.').first()).toBeVisible()
}

// Alle Tische mit Saldo ungleich 0,00 € aus dem „Alle Tische"-Drawer. Name und
// Saldo stehen als eigene <span>-Kinder ohne Trennzeichen (z. B. „Tisch 10,00 €"
// für „Tisch 1" mit Saldo „0,00 €") — deshalb je ein eigener <span> statt des
// zusammengesetzten Button-Textes.
async function offeneTischNamen(page: Page): Promise<string[]> {
  await page.goto('/service/tische')
  await page.getByRole('button', { name: 'Alle Tische' }).click()
  const zeilen = await page
    .getByRole('button', { name: /^.+\d,\d{2}\s*€$/ })
    .all()
  const namen: string[] = []
  for (const zeile of zeilen) {
    const spans = zeile.locator('span')
    const name = (await spans.nth(0).textContent())?.trim() ?? ''
    const saldo = (await spans.nth(1).textContent())?.trim() ?? ''
    if (name && !/^0,00\s*€$/.test(saldo)) {
      namen.push(name)
    }
  }
  await page.keyboard.press('Escape')
  return namen
}

// Klappt die Gruppe „Von anderen" auf (Positionen fremder Servicekräfte, per
// Default eingeklappt). Der Kopf fehlt, wenn keine fremden Positionen offen sind.
async function zeigeAlleAn(page: Page): Promise<void> {
  const vonAnderenKopf = page.getByRole('button', { name: /^Von anderen ·/ })
  if (await vonAnderenKopf.isVisible().catch(() => false)) {
    await vonAnderenKopf.click()
  }
}

// Der Button-Filter grenzt die Zeilen von der umschließenden ItemGroup ab, die
// ebenfalls alle Buttons enthält.
function vollePositionsZeilen(page: Page): Locator {
  return page.locator('[data-slot="item"]').filter({
    has: page.getByRole('button', { name: 'Produkt hinzufügen' }),
  })
}

async function leseAuswahlZaehler(
  zeile: Locator,
): Promise<{ text: string; treffer: RegExpExecArray | null }> {
  const text = (await zeile.textContent()) ?? ''
  const treffer = /(\d+) von (\d+) ausgewählt/.exec(text)
  return { text, treffer }
}

// Klickt jede Zeile voll aus („N von N ausgewählt"). Anders als „Alle
// auswählen" erfasst das auch fremde Zeilen (sofern über zeigeAlleAn
// aufgeklappt). Die Obergrenze verhindert eine Endlosschleife; die harte
// Nachbedingung verhindert, dass 50 erfolglose Klicks stillschweigend durchgehen.
export async function waehleAlleVollAus(page: Page): Promise<void> {
  const zeilen = vollePositionsZeilen(page)
  const anzahlZeilen = await zeilen.count()
  for (let i = 0; i < anzahlZeilen; i++) {
    const zeile = zeilen.nth(i)
    for (let klick = 0; klick < 50; klick++) {
      const { treffer } = await leseAuswahlZaehler(zeile)
      if (treffer && treffer[1] === treffer[2]) break
      await zeile.getByRole('button', { name: 'Produkt hinzufügen' }).click()
    }

    const { text, treffer } = await leseAuswahlZaehler(zeile)
    expect(
      treffer,
      `Zeile ${String(i)}: kein „N von N ausgewählt"-Text gefunden (Text: „${text}")`,
    ).not.toBeNull()
    expect(
      treffer?.[1],
      `Zeile ${String(i)}: nach 50 Klicks nicht voll ausgewählt (Text: „${text}")`,
    ).toBe(treffer?.[2])
  }
}

// Saldo im Tisch-Header (siehe TablePage).
export function tischSaldo(page: Page): Locator {
  return page.locator('[data-slot="tisch-saldo"]')
}

// Ready-Signal: TablePage zeigt den Header-Saldo bis zum Ende des State-Fetch
// als Skeleton; erst danach ist der Tab-Inhalt gerendert.
async function warteAufTischGeladen(page: Page): Promise<void> {
  await expect(tischSaldo(page)).toHaveText(/\d,\d{2}\s*€/)
}

// Kassiert jeden Tisch mit offenem Saldo leer — nötig, bevor der Kassenabschluss
// zulässig ist. Die Gruppe „Von anderen" wird vor jeder Zählung aufgeklappt,
// sonst bleiben fremde Positionen unbemerkt offen.
export async function settleAlleOffenenTische(page: Page): Promise<void> {
  const namen = await offeneTischNamen(page)
  for (const tisch of namen) {
    await oeffneTisch(page, tisch)

    // Ohne dieses Ready-Signal läsen die Prüfungen unten den noch leeren DOM und
    // übersprängen den Kassieren-Zweig stumm (Fetch-Race).
    await warteAufTischGeladen(page)

    // Der „Kassieren"-Button ist immer im DOM (nur deaktiviert) — ob unbezahlte
    // Positionen existieren, zeigt allein die Positionsliste.
    await page.getByRole('tab', { name: 'Kassieren' }).click()
    await zeigeAlleAn(page)
    if ((await vollePositionsZeilen(page).count()) > 0) {
      await waehleAlleVollAus(page)
      await page.getByRole('button', { name: /Kassieren/ }).click()
      const drawer = page.getByRole('dialog')
      await drawer.getByRole('button', { name: 'Kassieren' }).click()
      await expect(page.getByText('Zahlung erfolgreich.').first()).toBeVisible()
    }
  }
}

export async function abmelden(page: Page): Promise<void> {
  await page.getByRole('button', { name: 'Benutzermenü' }).click()
  await page.getByRole('menuitem', { name: 'Abmelden' }).click()
  await expect(page).toHaveURL(/\/login$/)
}
