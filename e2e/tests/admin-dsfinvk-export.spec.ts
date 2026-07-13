import { stat } from 'node:fs/promises'

import { expect, test } from '@playwright/test'

import { anmelden } from '../support/anmelden'
import { resetAndSeed } from '../support/seed'

// Deckt den Export-Klickpfad ab: eine abgeschlossene Kassensitzung auswählen
// und das DSFinV-K-Archiv über die Oberfläche herunterladen. Die inhaltliche
// Prüfung des Archivs übernimmt der DSFinV-K-Validator, nicht diese Spec.

test.describe('Admin lädt den DSFinV-K-Export herunter', () => {
  test('DSFinV-K-Export einer abgeschlossenen Kassensitzung herunterladen', async ({
    page,
    request,
  }) => {
    const zugangsdaten = await resetAndSeed(request)
    await anmelden(page, zugangsdaten.admin)

    await page.goto('/admin/kassenberichte')
    await expect(
      page.getByRole('heading', { name: 'Berichte & Export' }),
    ).toBeVisible()

    // Eine abgeschlossene Kassensitzung wählen (Seed: Freitag/Samstag sind
    // abgeschlossen, Sonntag ist die laufende Sitzung). Die Sitzungsliste ist
    // eine Spalte wählbarer Karten.
    await page
      .getByRole('button', { name: /Sommerfest 26 Samstag/ })
      .click()

    const downloadPromise = page.waitForEvent('download')
    await page
      .getByRole('button', { name: 'Archiv herunterladen (ZIP)' })
      .click()
    const download = await downloadPromise

    await expect(page.getByText('DSFinV-K-Archiv heruntergeladen.')).toBeVisible()

    const downloadPath = await download.path()
    expect(downloadPath, 'Download muss eine lokale Datei erzeugen').toBeTruthy()
    const stats = await stat(downloadPath ?? '')
    expect(stats.size, 'ZIP-Datei darf nicht leer sein').toBeGreaterThan(0)
    expect(download.suggestedFilename()).toMatch(/\.zip$/)
  })
})
