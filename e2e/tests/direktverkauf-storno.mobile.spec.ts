import { expect, test } from '@playwright/test'

import { anmelden } from '../support/anmelden'
import { resetAndSeed } from '../support/seed'
import { zeileMit } from '../support/servicekraft'

// Direktverkauf (Verkauf an der Theke ohne Tisch) und dessen Storno — beides
// am Handy-Viewport. Festbändchen Erwachsene (5,00 €) und Kinder (3,00 €) sind
// deterministisch: die Historie wird über den frisch getätigten Verkauf
// identifiziert (nicht über Position in der Liste). Storno ist nur
// Admin/Serviceleitung erlaubt (AuthSingleton.canCancel), daher meldet sich
// die Spec als Serviceleitung an — die auch selbst verkaufen darf.

test.describe('Servicekraft tätigt einen Direktverkauf und storniert ihn', () => {
  test('Verkauf erscheint in der Historie, Storno reduziert ihn auf den Restbetrag', async ({
    page,
    request,
  }) => {
    const zugangsdaten = await resetAndSeed(request)
    await anmelden(page, zugangsdaten.serviceleitung)

    await page.goto('/service/direktverkauf')
    await expect(page.getByRole('tab', { name: 'Verkaufen' })).toBeVisible()

    // Festbändchen Erwachsene (5,00 €) und Kinder (3,00 €) verkaufen.
    await page.getByText('Festbändchen', { exact: false }).first().click()
    const erwachsene = zeileMit(page, 'Erwachsene', 'Variante hinzufügen')
    await erwachsene.getByRole('button', { name: 'Variante hinzufügen' }).click()
    const kinder = zeileMit(page, 'Kinder', 'Variante hinzufügen')
    await kinder.getByRole('button', { name: 'Variante hinzufügen' }).click()

    await expect(page.getByText('8,00 €')).toBeVisible()

    await page.getByRole('button', { name: 'Verkauf abschließen' }).click()
    await expect(page.getByText('Verkauf abgeschlossen.')).toBeVisible()

    // In der Historie erscheint der frische Verkauf mit dem Gesamtbetrag.
    await page.getByRole('tab', { name: 'Historie' }).click()
    const verkaufsEintrag = zeileMit(page, '8,00', 'Stornieren')
    await expect(verkaufsEintrag).toBeVisible()

    await verkaufsEintrag.getByRole('button', { name: 'Stornieren' }).click()

    const drawer = page.getByRole('dialog')
    await expect(drawer.getByText('Verkauf stornieren')).toBeVisible()

    // Nur das Festbändchen Kinder (3,00 €) stornieren.
    const stornoKinder = zeileMit(page, 'Kinder', 'hinzufügen')
    await stornoKinder.getByRole('button', { name: /hinzufügen/ }).click()

    await drawer
      .getByPlaceholder('Kommentar (erforderlich)')
      .fill('Kind doch nicht mit')
    await drawer.getByRole('button', { name: 'Stornierung erteilen' }).click()

    // Der Verkauf zeigt nun den stornierten Anteil neben dem Gesamtbetrag.
    await expect(page.getByText(/3,00\s*€\s*storniert/)).toBeVisible()
  })
})
