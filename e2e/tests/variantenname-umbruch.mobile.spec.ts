import type { Page } from '@playwright/test'
import { expect, test } from '@playwright/test'

import { anmelden } from '../support/anmelden'
import { resetAndSeed } from '../support/seed'
import { oeffneTisch, produktGruppe, zeileMit } from '../support/servicekraft'
import {
  erwarteVollstaendigLesbarenNamen,
  zeilenGeometrie,
} from '../support/viewport'

// Kontrakt der Bestellliste am Handy. Kürzte sie den Variantennamen, sähen zwei
// Varianten desselben Produkts gleich aus und die Servicekraft griffe zur
// falschen — die Fehlbuchungs-Ursache aus dem Praxis-Feedback. Der Name bricht
// um, der Preis steht darunter, der Stepper sitzt in einem Slot fester Breite.
// Der erste Tap darf deshalb weder den Namen neu umbrechen noch die Zeilen
// darunter verschieben, sonst landet der zweite Tap auf der Nachbarzeile.
// Geprüft wird am gerenderten DOM, nicht an Klassennamen. Beide Bildschirme,
// die sich die ProductList teilen, sind abgedeckt: der Bestellen-Tab eines
// Tisches und der Direktverkauf.

// Zwei Handy-Lagen, beide unter der lg-Schwelle (1024 px) und damit im
// einspaltigen Layout: Hochformat ist die Regel-Haltung der Servicekraft,
// Querformat die Gegenprobe mit breitem, aber flachem Viewport.
const HANDY_LAGEN = [
  { lage: 'Hochformat', viewport: { width: 412, height: 915 } },
  { lage: 'Querformat', viewport: { width: 915, height: 412 } },
] as const

// Zwei Varianten desselben Seed-Produkts, deren Namen ein langes gemeinsames
// Ende teilen („…schorle 0,5l"). Kürzt die Namensspalte, bleibt für beide
// derselbe sichtbare Text übrig.
const PRODUKT = 'Saftschorle'
const ERSTE_VARIANTE = 'Johannisbeerschorle 0,5l'
const FOLGE_VARIANTE = 'Rhabarberschorle 0,5l'

// erwarteLesbareVariantenOhneShift prüft beides auf dem gerade offenen
// Bestell-Bildschirm: die vollständige Lesbarkeit beider Namen und die
// Ruhe der Liste beim ersten Tap.
async function erwarteLesbareVariantenOhneShift(
  page: Page,
  screen: string,
): Promise<void> {
  // Die Saftschorlen liegen in der Kategorie Getränke; nur die aktive Kategorie
  // steht im DOM.
  await page.getByRole('button', { name: 'Getränke', exact: true }).click()

  const gruppe = produktGruppe(page, PRODUKT)
  const ersteZeile = zeileMit(gruppe, ERSTE_VARIANTE, 'Variante hinzufügen')
  const folgeZeile = zeileMit(gruppe, FOLGE_VARIANTE, 'Variante hinzufügen')
  await expect(ersteZeile).toBeVisible()
  await expect(folgeZeile).toBeVisible()
  await ersteZeile.scrollIntoViewIfNeeded()

  const ersterName = ersteZeile.getByText(ERSTE_VARIANTE, { exact: true })
  const folgeName = folgeZeile.getByText(FOLGE_VARIANTE, { exact: true })

  // 1. Beide Namen stehen ungekürzt in ihrer Zeile; erst dadurch bleiben die
  //    beiden Varianten am Bildschirm unterscheidbar.
  await erwarteVollstaendigLesbarenNamen(
    ersterName,
    ERSTE_VARIANTE,
    `${screen}: erste Variante`,
  )
  await erwarteVollstaendigLesbarenNamen(
    folgeName,
    FOLGE_VARIANTE,
    `${screen}: zweite Variante`,
  )

  // 2. Der erste Tap blendet Minus und Menge ein. Weil der Stepper-Slot seine
  //    Breite behält, bleibt die Namensspalte gleich breit und die Folgezeile
  //    liegt danach dort, wo die Servicekraft sie gerade sieht.
  const vorher = await zeilenGeometrie(folgeZeile, ersterName)
  await ersteZeile.getByRole('button', { name: 'Variante hinzufügen' }).click()
  await expect(ersteZeile.getByText('1', { exact: true })).toBeVisible()
  const nachher = await zeilenGeometrie(folgeZeile, ersterName)

  expect(
    Math.abs(nachher.nameBreite - vorher.nameBreite),
    `${screen}: Namensspalte muss beim ersten Tap gleich breit bleiben (${vorher.nameBreite.toString()} → ${nachher.nameBreite.toString()} px)`,
  ).toBeLessThanOrEqual(1)
  expect(
    Math.abs(nachher.folgeZeileY - vorher.folgeZeileY),
    `${screen}: Folgezeile darf beim ersten Tap nicht verrutschen (${vorher.folgeZeileY.toString()} → ${nachher.folgeZeileY.toString()} px)`,
  ).toBeLessThanOrEqual(1)
}

for (const { lage, viewport } of HANDY_LAGEN) {
  test.describe(`Variantennamen am Handy im ${lage}`, () => {
    // Überschreibt den Pixel-7-Default des mobile-service-Projekts auf
    // Block-Ebene — dieselbe Technik wie im Split-Layout-Spec, nur für die
    // beiden Handy-Lagen.
    test.use({ viewport })

    test('Tisch-Bestellen: Namen ungekürzt, kein Versatz beim ersten Tap', async ({
      page,
      request,
    }) => {
      const zugangsdaten = await resetAndSeed(request)
      await anmelden(page, zugangsdaten.serviceleitung)

      await oeffneTisch(page, 'Tisch 3')

      await erwarteLesbareVariantenOhneShift(page, `Tisch-Bestellen @ ${lage}`)
    })

    test('Direktverkauf: Namen ungekürzt, kein Versatz beim ersten Tap', async ({
      page,
      request,
    }) => {
      const zugangsdaten = await resetAndSeed(request)
      await anmelden(page, zugangsdaten.serviceleitung)

      await page.goto('/service/direktverkauf')
      await expect(page.getByRole('tab', { name: 'Verkaufen' })).toBeVisible()

      await erwarteLesbareVariantenOhneShift(page, `Direktverkauf @ ${lage}`)
    })
  })
}
