import { expect, test } from '@playwright/test'
import type { Page } from '@playwright/test'

import { anmelden } from '../support/anmelden'
import { resetAndSeed } from '../support/seed'
import { zeileMit } from '../support/servicekraft'

// Tracer-Bullet-Spec: der Kernpfad einer Servicekraft am Tisch — anmelden,
// eine Bestellung aufnehmen, kassieren und den sichtbaren Betrag prüfen. Läuft
// in beiden Viewport-Projekten (Desktop-Admin und Mobile-Service), damit der
// Durchstich Backend, Reverse-Proxy, Frontend und Seed-Reset gemeinsam abdeckt.

// „Tisch 1" ist im Demo-Drehbuch unbenutzt und startet daher ohne offene
// Positionen — das hält die Beträge im Test deterministisch.
const TISCH = 'Tisch 1'
// Bratwurst XXL kostet 5,00 € (Seed-Stammdaten).
const PRODUKT = 'Bratwurst'
const VARIANTE = 'XXL'
const PREIS = '5,00'

test.describe('Servicekraft nimmt eine Bestellung auf und kassiert', () => {
  test('Bestellung aufnehmen und kassieren zeigt den erwarteten Betrag', async ({
    page,
    request,
  }) => {
    // Jede Spec startet vom bekannten Seed-Zustand.
    const zugangsdaten = await resetAndSeed(request)

    await anmelden(page, zugangsdaten.service)

    // Tischauswahl öffnen, nach „Tisch 1" filtern und die Tisch-Zeile wählen.
    // Der Tisch-Button trägt den Saldo (…€) im Namen; so unterscheidet er sich
    // vom danebenliegenden Favoriten-Stern („… zu/aus Favoriten …").
    await page.goto('/service/tische')
    await page.getByRole('button', { name: 'Alle Tische' }).click()
    await page.getByPlaceholder('Tisch suchen...').fill(TISCH)
    await page
      .getByRole('button', { name: new RegExp(`^${TISCH}\\b.*€`) })
      .click()

    // Auf der Tischseite angekommen: die Tab-Leiste des Tisches ist sichtbar.
    await expect(page.getByRole('tab', { name: 'Bestellen' })).toBeVisible()

    // Eine Bratwurst XXL bestellen.
    await bestelleEinProdukt(page)

    // Auf den Kassieren-Tab wechseln und die eben bestellte Position auswählen.
    await page.getByRole('tab', { name: 'Kassieren' }).click()

    const position = zeileMit(page, `${PRODUKT} ${VARIANTE}`, 'Produkt hinzufügen')
    await expect(position).toBeVisible()
    await position.getByRole('button', { name: 'Produkt hinzufügen' }).click()

    // Kassieren-Leiste zeigt den Betrag und öffnet den Zahlungs-Drawer.
    const kassierenLeiste = page.getByRole('button', { name: /Kassieren/ })
    await expect(kassierenLeiste).toContainText(PREIS)
    await kassierenLeiste.click()

    // Der Zahlungs-Drawer bestätigt den sichtbaren Betrag …
    const drawer = page.getByRole('dialog')
    await expect(drawer.getByText(new RegExp(`${PREIS}\\s*€`)).first()).toBeVisible()

    // … und die Zahlung wird kassiert.
    await drawer.getByRole('button', { name: 'Kassieren' }).click()

    await expect(page.getByText('Zahlung erfolgreich.')).toBeVisible()

    // Nach dem Kassieren ist der Tisch wieder ausgeglichen.
    await expect(page.getByText('0,00 €').first()).toBeVisible()
  })
})

// bestelleEinProdukt fügt eine Bratwurst XXL zur Bestellung hinzu und nimmt sie
// über den Bestell-Drawer auf.
async function bestelleEinProdukt(page: Page) {
  // Sicherstellen, dass der Bestellen-Tab aktiv ist (Default), dann Produkt aufklappen.
  await page.getByRole('tab', { name: 'Bestellen' }).click()
  await page.getByText(PRODUKT, { exact: false }).first().click()

  const variante = zeileMit(page, VARIANTE, 'Variante hinzufügen')
  await variante.getByRole('button', { name: 'Variante hinzufügen' }).click()

  // Sticky-Leiste „Bestellung überprüfen" öffnet den Drawer.
  await page.getByRole('button', { name: /Bestellung überprüfen/ }).click()

  const drawer = page.getByRole('dialog')
  await drawer.getByRole('button', { name: 'Bestellung aufnehmen' }).click()

  await expect(page.getByText('Bestellung wurde aufgenommen.')).toBeVisible()
}
