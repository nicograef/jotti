import type { APIRequestContext } from '@playwright/test'
import { expect, test } from '@playwright/test'

import { anmelden } from '../support/anmelden'
import { resetAndSeed } from '../support/seed'
import { oeffneTisch, waehleVariante } from '../support/servicekraft'

// Deckt den Verwaltungspfad für die Bondrucker ab: die Drucker-IP einer
// Station ändern und einen Zusatzdrucker mit Produkt einrichten. Name und IP
// werden on-blur gespeichert (kein Speichern-Button). Eine Bestellung über die
// Oberfläche muss danach als Druckauftrag beim Relay ankommen.

// RELAY_AUTH_TOKEN aus docker-compose.e2e.yml.
const RELAY_TOKEN = 'e2e-relay-token-do-not-use-in-production'

interface Druckauftrag {
  id: number
  zielIp: string
  payload: string // Base64 ESC/POS
}

async function offeneDruckauftraege(
  request: APIRequestContext,
): Promise<Druckauftrag[]> {
  const response = await request.post('/api/relay/poll', {
    data: { token: RELAY_TOKEN },
  })
  expect(
    response.ok(),
    `Relay-Poll muss erfolgreich sein (Status ${response.status().toString()})`,
  ).toBeTruthy()
  return ((await response.json()) as { auftraege: Druckauftrag[] }).auftraege
}

// Meldet alle offenen Aufträge (Seed-Historie, Testbon) als gedruckt, damit
// danach nur die Aufträge der Bestellung offen sind. Der Poll liefert höchstens
// 200 Aufträge je Abruf, daher die Schleife.
async function quittiereOffeneDruckauftraege(
  request: APIRequestContext,
): Promise<void> {
  for (let runde = 0; runde < 10; runde++) {
    const offene = await offeneDruckauftraege(request)
    if (offene.length === 0) return
    const response = await request.post('/api/relay/ergebnis', {
      data: {
        token: RELAY_TOKEN,
        gedruckteIds: offene.map((auftrag) => auftrag.id),
        fehlversuche: [],
      },
    })
    expect(response.ok(), 'Relay-Ergebnis muss erfolgreich sein').toBeTruthy()
  }
  throw new Error('Offene Druckaufträge ließen sich nicht quittieren')
}

test.describe('Admin verwaltet Druckstationen', () => {
  test('Drucker-IP einer Station ändern', async ({ page, request }) => {
    const zugangsdaten = await resetAndSeed(request)
    await anmelden(page, zugangsdaten.admin)

    await page.goto('/admin/druckstationen')
    await expect(
      page.getByRole('heading', { name: 'Bondrucker' }),
    ).toBeVisible()

    // Die Essen-Station ist per Seed-Drehbuch mit 192.168.8.51 konfiguriert.
    const essenCard = page
      .locator('div')
      .filter({ hasText: 'Bons für die Essensausgabe' })
      .filter({ has: page.getByLabel('Drucker-IP') })
      .last()
    const ipInput = essenCard.getByLabel('Drucker-IP')
    await expect(ipInput).toHaveValue('192.168.8.51')

    // Neue IP eintragen und das Feld verlassen — das speichert on-blur.
    await ipInput.fill('192.168.8.99')
    await ipInput.blur()

    await expect(
      page.getByText(/Drucker-IP für .Essen. gespeichert\./),
    ).toBeVisible()

    // Nach dem Neuladen kommt der Wert vom Backend, nicht aus dem Eingabefeld.
    await page.reload()
    await expect(
      page.getByRole('heading', { name: 'Bondrucker' }),
    ).toBeVisible()
    await expect(ipInput).toHaveValue('192.168.8.99')
  })

  test('Zusatzdrucker mit Produkt einrichten', async ({ page, request }) => {
    const zugangsdaten = await resetAndSeed(request)
    await anmelden(page, zugangsdaten.admin)

    await page.goto('/admin/druckstationen')
    await page.getByRole('button', { name: 'Drucker hinzufügen' }).click()
    const anlegenDialog = page.getByRole('dialog')
    await anlegenDialog.getByLabel('Name').fill('Kaffee & Kuchen')
    await anlegenDialog
      .getByRole('button', { name: 'Drucker hinzufügen' })
      .click()
    await expect(
      page.getByText('Drucker „Kaffee & Kuchen“ angelegt.'),
    ).toBeVisible()

    // Die Karte ist das innerste div mit Titelzusatz und IP-Feld.
    const karte = page
      .locator('div')
      .filter({ hasText: '— Zusätzlicher Drucker' })
      .filter({ has: page.getByLabel('Drucker-IP') })
      .last()
    const ipInput = karte.getByLabel('Drucker-IP')
    await ipInput.fill('192.168.8.60')
    await ipInput.blur()
    await expect(
      page.getByText('Drucker-IP für „Kaffee & Kuchen“ gespeichert.'),
    ).toBeVisible()

    await karte.getByRole('button', { name: 'Produkt hinzufügen' }).click()
    const auswahl = page.getByRole('dialog')
    await auswahl
      .getByRole('listitem')
      .filter({ has: page.getByText('Kuchen', { exact: true }) })
      .getByRole('button')
      .click()
    await expect(
      page.getByText('„Kuchen“ druckt jetzt an „Kaffee & Kuchen“.'),
    ).toBeVisible()
    await expect(auswahl).toBeHidden()
    await expect(karte.getByText('Kuchen', { exact: true })).toBeVisible()

    await karte.getByRole('button', { name: 'Testbon' }).click()
    await expect(
      page.getByText('Testbon an „Kaffee & Kuchen“ gesendet.'),
    ).toBeVisible()

    // Nach dem Neuladen kommt die Zuordnung vom Backend.
    await page.reload()
    await expect(
      page.getByRole('heading', { name: 'Bondrucker' }),
    ).toBeVisible()
    await expect(karte.getByText('Kuchen', { exact: true })).toBeVisible()

    // Eine Kuchen-Bestellung über die Oberfläche erzeugt genau einen Bon am
    // Zusatzdrucker, mit Produkt und Bestellkommentar. „Tisch 13" ist im
    // Demo-Drehbuch des laufenden Tages unbenutzt.
    await quittiereOffeneDruckauftraege(request)
    await oeffneTisch(page, 'Tisch 13')
    await waehleVariante(page, 'Kuchen', 'Stück')
    // Ab lg steht der Abschluss als feste Spalte neben der Produktliste.
    await page.getByPlaceholder('Kommentar (optional)').fill('ohne Sahne')
    await page.getByRole('button', { name: 'Bestellung aufnehmen' }).click()
    await expect(
      page.getByText('Bestellung wurde aufgenommen.').first(),
    ).toBeVisible()

    const bons = (await offeneDruckauftraege(request))
      .filter((auftrag) => auftrag.zielIp === '192.168.8.60')
      .map((auftrag) =>
        Buffer.from(auftrag.payload, 'base64').toString('latin1'),
      )
    expect(bons, 'genau ein Bon am Zusatzdrucker').toHaveLength(1)
    expect(bons[0]).toContain('Kuchen')
    expect(bons[0]).toContain('ohne Sahne')
  })
})
