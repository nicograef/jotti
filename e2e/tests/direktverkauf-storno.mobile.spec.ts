import { expect, test } from '@playwright/test'

import { anmelden } from '../support/anmelden'
import { resetAndSeed } from '../support/seed'
import {
  oeffneHistorienDetail,
  waehleVariante,
  zeileMit,
} from '../support/servicekraft'

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

    // Festbändchen Erwachsene (5,00 €) und Kinder (3,00 €) verkaufen —
    // waehleVariante aktiviert dafür den Sonstiges-Kategorie-Chip.
    await waehleVariante(page, 'Festbändchen', 'Erwachsene')
    await waehleVariante(page, 'Festbändchen', 'Kinder')

    // Der Dock-Aktionsbutton trägt Anzahl, Label und Summe („2 · Kassieren ·
    // 8,00 €") und öffnet den Kassieren-Drawer; „Verkauf abschließen" liegt dort.
    await page.getByRole('button', { name: /Kassieren.*8,00/ }).click()
    const kassierenDrawer = page.getByRole('dialog')
    await expect(kassierenDrawer).toBeVisible()
    await kassierenDrawer
      .getByRole('button', { name: 'Verkauf abschließen' })
      .click()
    await expect(page.getByText('Verkauf abgeschlossen.')).toBeVisible()

    // In der Historie erscheint der frische Verkauf mit dem Gesamtbetrag. Die
    // Zeile trägt keine Inline-Aktionen mehr: sie (per Betrag eindeutig) öffnet
    // den Detail-Drawer, dort liegt „Stornieren…".
    await page.getByRole('tab', { name: 'Historie' }).click()
    const detail = await oeffneHistorienDetail(page, /Verkauf.*8,00/)
    await detail.getByRole('button', { name: /Stornieren…/ }).click()

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
