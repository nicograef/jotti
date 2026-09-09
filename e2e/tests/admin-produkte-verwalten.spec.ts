import { expect, test } from '@playwright/test'

import { anmelden } from '../support/anmelden'
import { resetAndSeed } from '../support/seed'

// Deckt den Verwaltungspfad für Produkte samt Varianten ab: Anlegen,
// Bearbeiten, eine Variante aktivieren und wieder deaktivieren, sowie das
// Umsortieren von Produkten und Varianten. Die Seite „Produkte & Preise"
// listet Produkte je Kategorie mit Varianten als Chips (Name, Preis,
// Mini-Switch); eine inaktive Variante trägt die „aus"-Markierung.

// Liest die Namen aus den aria-Labels der Bedienelemente: Sie tragen den Namen
// und stehen in der Reihenfolge im DOM, in der das Backend die Liste liefert.
function namenAus(labels: string[], muster: RegExp): string[] {
  return labels.map((label) => muster.exec(label)?.[1] ?? label)
}

test.describe('Admin verwaltet Produkte und Varianten', () => {
  test('Produkt samt Variante anlegen, ändern und Variante deaktivieren', async ({
    page,
    request,
  }) => {
    const zugangsdaten = await resetAndSeed(request)
    await anmelden(page, zugangsdaten.admin)

    await page.goto('/admin/produkte')
    await expect(
      page.getByRole('heading', { name: 'Produkte & Preise' }),
    ).toBeVisible()

    // Neues Produkt anlegen.
    await page.getByRole('button', { name: 'Neues Produkt' }).click()
    const newProductDialog = page.getByRole('dialog')
    await newProductDialog.getByLabel('Name').fill('Eistee')
    await newProductDialog
      .getByRole('button', { name: 'Produkt anlegen' })
      .click()
    await expect(
      page.getByText('Produkt "Eistee" wurde angelegt.'),
    ).toBeVisible()

    // Die Produktzeile über den Namen und ihren Bearbeiten-Button auflösen und
    // eine Variante anlegen (der gestrichelte „Variante"-Button je Zeile).
    const produktItem = page
      .locator('div')
      .filter({ has: page.getByRole('button', { name: 'Produkt bearbeiten' }) })
      .filter({ hasText: 'Eistee' })
      .last()
    await produktItem
      .getByRole('button', { name: 'Variante', exact: true })
      .click()

    const newVariantDialog = page.getByRole('dialog')
    await newVariantDialog.getByLabel('Name').fill('0,5l')
    await newVariantDialog.getByLabel('Preis').fill('2,80')
    await newVariantDialog
      .getByRole('button', { name: 'Variante anlegen' })
      .click()
    await expect(
      page.getByText('Variante "0,5l" wurde angelegt.'),
    ).toBeVisible()

    // Die neue Variante ist zunächst deaktiviert: Preis-Chip sichtbar, Switch
    // aus und die „aus"-Markierung gesetzt.
    await expect(produktItem.getByText('2,80')).toBeVisible()
    const ausMarker = produktItem.getByText('aus', { exact: true })
    await expect(ausMarker).toBeVisible()
    const variantSwitch = produktItem.getByRole('switch').last()
    await expect(variantSwitch).not.toBeChecked()

    // Variante aktivieren: Switch an, „aus"-Markierung verschwindet.
    await variantSwitch.click()
    await expect(variantSwitch).toBeChecked()
    await expect(ausMarker).toBeHidden()

    // Variante wieder deaktivieren.
    await variantSwitch.click()
    await expect(variantSwitch).not.toBeChecked()
    await expect(ausMarker).toBeVisible()

    // Produkt bearbeiten: Namen ändern.
    await produktItem
      .getByRole('button', { name: 'Produkt bearbeiten' })
      .click()
    const editDialog = page.getByRole('dialog')
    await editDialog.getByLabel('Name').fill('Eistee Pfirsich')
    await editDialog.getByRole('button', { name: 'Speichern' }).click()

    await expect(page.getByText('Eistee Pfirsich')).toBeVisible()
  })

  test('Varianten und Produkte umsortieren', async ({ page, request }) => {
    const zugangsdaten = await resetAndSeed(request)
    await anmelden(page, zugangsdaten.admin)

    await page.goto('/admin/produkte')
    await expect(
      page.getByRole('heading', { name: 'Produkte & Preise' }),
    ).toBeVisible()

    const bratwurst = page
      .locator('div')
      .filter({ has: page.getByRole('button', { name: 'Produkt bearbeiten' }) })
      .filter({ hasText: 'Bratwurst' })
      .last()

    const variantenReihenfolge = async () =>
      namenAus(
        await bratwurst
          .getByRole('button', { name: /^Variante .+ bearbeiten$/ })
          .evaluateAll((elemente) =>
            elemente.map((el) => el.getAttribute('aria-label') ?? ''),
          ),
        /^Variante „(.+)" bearbeiten$/,
      )

    // Seed-Reihenfolge der Bratwurst-Varianten.
    expect(await variantenReihenfolge()).toEqual([
      'Normal',
      'XXL',
      'Currywurst',
    ])

    await bratwurst
      .getByRole('button', { name: 'Variante „XXL" nach vorne' })
      .click()

    await expect
      .poll(variantenReihenfolge)
      .toEqual(['XXL', 'Normal', 'Currywurst'])

    // Produkte verschieben sich innerhalb ihrer Kategorie.
    const produktReihenfolge = async () =>
      namenAus(
        await page
          .getByRole('button', { name: /^Produkt .+ nach oben$/ })
          .evaluateAll((elemente) =>
            elemente.map((el) => el.getAttribute('aria-label') ?? ''),
          ),
        /^Produkt „(.+)" nach oben$/,
      )

    expect((await produktReihenfolge()).slice(0, 3)).toEqual([
      'Bratwurst',
      'Pommes',
      'Flammkuchen',
    ])

    await page
      .getByRole('button', { name: 'Produkt „Pommes" nach oben' })
      .click()

    await expect
      .poll(async () => (await produktReihenfolge()).slice(0, 3))
      .toEqual(['Pommes', 'Bratwurst', 'Flammkuchen'])

    // Der Pfeil neben dem Switch muss ein echtes 32-px-Ziel sein und die
    // unsichtbare Trefferfläche des Switch (after:-inset-x-3) darf seine
    // zugewandte Kante nicht verschlucken.
    const chevron = bratwurst.getByRole('button', {
      name: 'Variante „XXL" nach hinten',
    })
    const box = await chevron.boundingBox()
    expect(box).not.toBeNull()
    // box ist ab hier nicht mehr null — die Assertion oben hätte sonst schon
    // fehlgeschlagen; die "!" ersetzt ein "?? 0", das eine fehlgeschlagene
    // Messung stillschweigend auf Koordinate 0 gezogen hätte.
    expect(box!.width).toBeGreaterThanOrEqual(32)
    expect(box!.height).toBeGreaterThanOrEqual(32)

    const treffer = await page.evaluate(
      ({ x, y }) =>
        document
          .elementFromPoint(x, y)
          ?.closest('button')
          ?.getAttribute('aria-label') ?? '',
      { x: box!.x + 1, y: box!.y + box!.height / 2 },
    )
    expect(treffer).toBe('Variante „XXL" nach hinten')
  })
})
