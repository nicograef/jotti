import { expect, test } from '@playwright/test'

import { anmelden } from '../support/anmelden'
import { leseSollBestandCents } from '../support/kasse'
import { resetAndSeed } from '../support/seed'
import { abmelden, waehleVariante, zeileMit } from '../support/servicekraft'

// Pfand am Handy-Viewport: eine Servicekraft verkauft im Direktverkauf nur
// einen Aperol Spritz (5,50 €); dessen verknüpftes Pfand Glas (2,00 €) legt
// der Warenkorb selbst dazu (Auto-Pfand). Danach nimmt sie das Glas über
// „Pfand zurück" an. Der Admin liest den Soll-Bestand davor und danach; die
// Differenz ist Verkauf samt Pfand minus ausgezahltes Pfand.

const VERKAUF_CENTS = 750
const PFAND_CENTS = 200

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

    // Nur das Getränk wählen: 7,50 € statt 5,50 € beweist, dass das Pfand
    // automatisch mitkommt.
    await waehleVariante(page, 'Hugo/Aperol', 'Aperol Spritz')
    await page.getByRole('button', { name: /Kassieren.*7,50/ }).click()
    const kassierenDrawer = page.getByRole('dialog')
    await kassierenDrawer
      .getByRole('button', { name: 'Verkauf abschließen' })
      .click()
    await expect(page.getByText('Verkauf abgeschlossen.')).toBeVisible()

    // Ohne Auswahl öffnet der Dock-Button nur die Pfandrückgabe.
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
