import type { Page } from '@playwright/test'
import { expect, test } from '@playwright/test'

import { anmelden } from '../support/anmelden'
import { resetAndSeed } from '../support/seed'
import { abmelden, waehleVariante, zeileMit } from '../support/servicekraft'

// Pfand am Handy-Viewport: eine Servicekraft verkauft im Direktverkauf Bier
// 0,5l (4,50 €) mit Pfand Glas (2,00 €) und nimmt das Glas danach über
// „Pfand zurück" an. Der Admin liest den Soll-Bestand davor und danach; die
// Differenz ist Verkauf minus ausgezahltes Pfand.

const VERKAUF_CENTS = 650
const PFAND_CENTS = 200

// Betrag und Label „Soll-Bestand" teilen einen Container, der Betrag steht
// vorn; bis der Kassenbestand geladen ist, zeigt die Seite „—".
async function leseSollBestandCents(page: Page): Promise<number> {
  await page.goto('/admin/kasse')
  const soll = page.getByText(/^Soll-Bestand/).locator('..')
  await expect(soll).toHaveText(/^\d+,\d{2}\s*€/)
  const text = (await soll.textContent()) ?? ''
  const treffer = /^(\d+),(\d{2})/.exec(text)
  expect(treffer, `Soll-Bestand nicht lesbar (Text: „${text}")`).not.toBeNull()
  const [, euro = '', cent = ''] = treffer ?? []
  return Number(euro) * 100 + Number(cent)
}

test.describe('Servicekraft verkauft mit Pfand und nimmt das Glas zurück', () => {
  test('der Soll-Bestand steigt um den Verkauf minus das ausgezahlte Pfand', async ({
    page,
    request,
  }) => {
    const zugangsdaten = await resetAndSeed(request)

    await anmelden(page, zugangsdaten.admin)
    const sollVorher = await leseSollBestandCents(page)
    await abmelden(page)

    await anmelden(page, zugangsdaten.service)
    await page.goto('/service/direktverkauf')
    await expect(page.getByRole('tab', { name: 'Verkaufen' })).toBeVisible()

    await waehleVariante(page, 'Bier', '0,5l')
    await waehleVariante(page, 'Pfand', 'Glas')
    await page.getByRole('button', { name: /Kassieren.*6,50/ }).click()
    const kassierenDrawer = page.getByRole('dialog')
    await kassierenDrawer
      .getByRole('button', { name: 'Verkauf abschließen' })
      .click()
    await expect(page.getByText('Verkauf abgeschlossen.')).toBeVisible()

    // Ohne Auswahl öffnet der Dock-Button nur noch die Pfandrückgabe.
    await page.getByRole('button', { name: /Pfand zurück/ }).click()
    const pfandDrawer = page.getByRole('dialog')
    await zeileMit(pfandDrawer, 'Glas', 'Variante hinzufügen')
      .getByRole('button', { name: 'Variante hinzufügen' })
      .click()
    await pfandDrawer.getByRole('button', { name: 'Pfand auszahlen' }).click()
    await expect(page.getByText('Pfand ausgezahlt.')).toBeVisible()
    await abmelden(page)

    await anmelden(page, zugangsdaten.admin)
    const sollNachher = await leseSollBestandCents(page)
    expect(sollNachher - sollVorher).toBe(VERKAUF_CENTS - PFAND_CENTS)
  })
})
