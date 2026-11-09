import type { Page } from '@playwright/test'
import { expect } from '@playwright/test'

// Liest den Soll-Bestand der laufenden Kassensitzung von /admin/kasse in Cent.
// Betrag und Label „Soll-Bestand" teilen einen Container, der Betrag steht
// vorn; bis der Kassenbestand geladen ist, zeigt die Seite „—".
export async function leseSollBestandCents(page: Page): Promise<number> {
  await page.goto('/admin/kasse')
  const soll = page.getByText(/^Soll-Bestand/).locator('..')
  await expect(soll).toHaveText(/^\d+,\d{2}\s*€/)
  const text = (await soll.textContent()) ?? ''
  const treffer = /^(\d+),(\d{2})/.exec(text)
  expect(treffer, `Soll-Bestand nicht lesbar (Text: „${text}")`).not.toBeNull()
  const [, euro = '', cent = ''] = treffer ?? []
  return Number(euro) * 100 + Number(cent)
}
