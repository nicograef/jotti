// App-Screenshots und OG-Bild für die Marketing-Website.
//
// Modi (Standard: beide nacheinander):
//   app  — gegen den e2e-Stack (JOTTI_ENABLE_TEST_API=1): Seed zurücksetzen,
//          anmelden, jedes Motiv in Hell UND Dunkel aufnehmen. Ziel:
//          `website/src/assets/screenshots/`.
//   og   — `make website-build`, `dist/` hinter dem Static-Server aus
//          `csp-server.mjs` servieren und den Hero (hell) als 1200×630-Bild
//          aufnehmen. Ziel: `website/src/assets/og-startseite.png`.
//
// Die App folgt der Systempräferenz (Theme-Default „system"), daher genügt
// Playwrights `emulateMedia({ colorScheme })` — kein Theme-Schalter-State.
//
// E2E_BASE_URL überschreibt den Compose-Default http://localhost:8080.
// Seed- und Login-Helfer kommen aus der e2e-Suite (`support/*.ts`); deshalb
// läuft das Skript mit `node --experimental-strip-types` (Make-Target
// `website-screenshots`).

import { spawnSync } from 'node:child_process'
import { mkdirSync } from 'node:fs'
import { join, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { devices } from '@playwright/test'

import { launchBrowser } from './browser.mjs'
import { anmelden } from '../support/anmelden.ts'
import { resetAndSeed } from '../support/seed.ts'
import {
  bestellePosition,
  oeffneHistorienDetail,
  oeffneTisch,
  waehleAlleVollAus,
  waehleVariante,
  zeileMit,
} from '../support/servicekraft.ts'

const repoRoot = resolve(fileURLToPath(new URL('../..', import.meta.url)))
const BASE = process.env.E2E_BASE_URL ?? 'http://localhost:8080'
const SHOT_OUT =
  process.env.SHOT_OUT ??
  join(repoRoot, 'website', 'src', 'assets', 'screenshots')
const OG_OUT =
  process.env.OG_OUT ??
  join(repoRoot, 'website', 'src', 'assets', 'og-startseite.png')

const mode = process.argv[2] ?? 'all'

// Wartet auf Fonts und einen Repaint nach dem Theme-Wechsel und nimmt den Fokus
// vom aktiven Element, damit kein Fokus-Ring in die Aufnahme gebrannt wird.
async function settle(page) {
  await page.evaluate(() => document.fonts.ready)
  await page.evaluate(() => document.activeElement?.blur?.())
  await page.waitForTimeout(450)
}

// Gleicher DOM-Zustand, nur `prefers-color-scheme` gekippt.
async function captureLightDark(page, name) {
  for (const scheme of ['light', 'dark']) {
    await page.emulateMedia({ colorScheme: scheme })
    await settle(page)
    await page.screenshot({ path: join(SHOT_OUT, `${name}-${scheme}.png`) })
    console.log(`  ✓ ${name}-${scheme}.png`)
  }
}

async function login(context, zugangsdaten) {
  const page = await context.newPage()
  await anmelden(page, zugangsdaten)
  return page
}

// Querformat-Tablet (≥ 1024 px): löst das zweispaltige Layout ab `lg` aus — den
// Service-Split-Screen (siehe docs/decisions.md D08) und das Sidebar-Layout der
// Admin-Motive. `produkte` (Tablet) und `produktverwaltung` (Desktop) zeigen
// deshalb bewusst dasselbe Design auf zwei Geräteklassen.
const tabletLandscape = {
  viewport: { width: 1194, height: 834 },
  deviceScaleFactor: 2,
}

async function captureApp() {
  mkdirSync(SHOT_OUT, { recursive: true })
  const browser = await launchBrowser()
  try {
    const apiContext = await browser.newContext({ baseURL: BASE })
    const zugangsdaten = await resetAndSeed(apiContext.request)
    await apiContext.close()

    // ---- Service-Motive (Handy, Servicekraft „maria") ----
    const phone = await browser.newContext({
      baseURL: BASE,
      ...devices['Pixel 7'],
    })
    const phonePage = await login(phone, zugangsdaten.service)

    await phonePage.goto('/service/tische')
    await phonePage.getByText('Meine Tische').first().waitFor()
    await phonePage.getByText('Noch offen', { exact: false }).first().waitFor()
    await captureLightDark(phonePage, 'tischuebersicht')

    // Lokaler Warenkorb, noch nicht abgeschickt.
    await oeffneTisch(phonePage, 'Tisch 1')
    await phonePage.getByRole('tab', { name: 'Bestellen' }).click()
    await waehleVariante(phonePage, 'Bratwurst', 'Normal', 2)
    await waehleVariante(phonePage, 'Pommes', 'Groß', 2)
    await phonePage
      .getByRole('button', { name: /Bestellung überprüfen/ })
      .waitFor()
    await captureLightDark(phonePage, 'bestellansicht')

    await oeffneTisch(phonePage, 'Tisch 2')
    await phonePage.getByRole('tab', { name: 'Kassieren' }).click()
    const vonAnderen = phonePage.getByRole('button', { name: /^Von anderen ·/ })
    if (await vonAnderen.isVisible().catch(() => false))
      await vonAnderen.click()
    await waehleAlleVollAus(phonePage)
    await phonePage.getByRole('button', { name: /Kassieren/ }).click()
    const zahlungDrawer = phonePage.getByRole('dialog')
    await zahlungDrawer.getByText(/€/).first().waitFor()
    await captureLightDark(phonePage, 'zahlung')
    await phonePage.keyboard.press('Escape')
    await phone.close()

    // ---- Direktverkauf (Querformat-Tablet, Split-Screen, Servicekraft „maria") ----
    const tabletService = await browser.newContext({
      baseURL: BASE,
      ...tabletLandscape,
    })
    const tabletServicePage = await login(tabletService, zugangsdaten.service)
    await tabletServicePage.goto('/service/direktverkauf')
    await tabletServicePage.getByRole('tab', { name: 'Verkaufen' }).waitFor()
    const dvZeile = zeileMit(
      tabletServicePage,
      'Currywurst',
      'Variante hinzufügen',
    )
    await dvZeile.getByRole('button', { name: 'Variante hinzufügen' }).click()
    await dvZeile.getByRole('button', { name: 'Variante hinzufügen' }).click()
    await tabletServicePage
      .getByRole('button', { name: 'Verkauf abschließen' })
      .waitFor()
    await captureLightDark(tabletServicePage, 'direktverkauf')
    await tabletService.close()

    // ---- Stornierung (Handy, Serviceleitung „felix") ----
    const phoneSL = await browser.newContext({
      baseURL: BASE,
      ...devices['Pixel 7'],
    })
    const phoneSLPage = await login(phoneSL, zugangsdaten.serviceleitung)
    // „Tisch 15" ist im Sonntags-Drehbuch unbenutzt (wie in der Storno-Spec).
    await oeffneTisch(phoneSLPage, 'Tisch 15')
    await bestellePosition(phoneSLPage, 'Pommes', 'Klein')
    await phoneSLPage.getByRole('tab', { name: 'Historie' }).click()
    const detail = await oeffneHistorienDetail(
      phoneSLPage,
      /Bestellung.*\+2,50/,
    )
    await detail.getByRole('button', { name: /Stornieren…/ }).click()
    const stornoDrawer = phoneSLPage.getByRole('dialog')
    await zeileMit(stornoDrawer, 'Pommes Klein', 'hinzufügen')
      .getByRole('button', { name: /hinzufügen/ })
      .click()
    await stornoDrawer
      .getByPlaceholder('Kommentar (erforderlich)')
      .fill('Falsch bestellt, storniert')
    // Den „Bestellung wurde aufgenommen."-Toast aus dem Bestellschritt abklingen
    // lassen, damit er nicht über der Storno-Aufnahme hängt.
    await phoneSLPage
      .getByText('Bestellung wurde aufgenommen.')
      .first()
      .waitFor({ state: 'hidden' })
      .catch(() => {})
    await captureLightDark(phoneSLPage, 'stornierung')
    await phoneSL.close()

    // ---- Admin-Motive (Querformat-Tablet, Admin „thomas") ----
    const tabletAdmin = await browser.newContext({
      baseURL: BASE,
      ...tabletLandscape,
    })
    const tabletAdminPage = await login(tabletAdmin, zugangsdaten.admin)

    await tabletAdminPage.goto('/admin/produkte')
    await tabletAdminPage.getByText('Produkte & Preise').first().waitFor()
    await captureLightDark(tabletAdminPage, 'produkte')

    await tabletAdminPage.goto('/admin/benutzer')
    await tabletAdminPage.getByText('Helfer & Zugänge').first().waitFor()
    await captureLightDark(tabletAdminPage, 'benutzer')

    // Der Betrag ist Pflichtfeld: ausgefüllt, damit die Blur-Validierung in
    // settle keinen Fehlerzustand in die Aufnahme brennt.
    await tabletAdminPage.goto('/admin/kasse')
    await tabletAdminPage.getByRole('button', { name: 'Geld einlegen' }).click()
    await tabletAdminPage.getByRole('dialog').waitFor()
    await tabletAdminPage.getByLabel('Betrag').fill('50,00')
    await tabletAdminPage.getByLabel('Kommentar').fill('Wechselgeld Nachschub')
    await captureLightDark(tabletAdminPage, 'geldtransit')
    await tabletAdminPage.keyboard.press('Escape')
    await tabletAdmin.close()

    // ---- Desktop-Motive (Browser-Rahmen, Admin „thomas") ----
    const desktop = await browser.newContext({
      baseURL: BASE,
      viewport: { width: 1360, height: 850 },
      deviceScaleFactor: 2,
    })
    const desktopPage = await login(desktop, zugangsdaten.admin)
    await desktopPage.goto('/admin/produkte')
    await desktopPage.getByText('Produkte & Preise').first().waitFor()
    await captureLightDark(desktopPage, 'produktverwaltung')

    // Auswertung: historischer Tagesbericht (deterministisch Nr. 2 gewählt).
    await desktopPage.goto('/admin/kassenberichte')
    await desktopPage.getByText('Berichte & Export').first().waitFor()
    await desktopPage.getByText('Sommerfest 26 Samstag').first().click()
    await desktopPage.getByText('Umsatz nach Steuersatz').first().waitFor()
    await captureLightDark(desktopPage, 'auswertung')
    await desktop.close()
  } finally {
    await browser.close()
  }
  console.log('App-Screenshots fertig →', SHOT_OUT)
}

async function captureOg() {
  console.log('Baue Website (make website-build) …')
  const build = spawnSync('make', ['website-build'], {
    cwd: repoRoot,
    stdio: 'inherit',
  })
  if (build.status !== 0) throw new Error('make website-build fehlgeschlagen')

  const { startStaticServer } = await import('./csp-server.mjs')
  const distDir = join(repoRoot, 'website', 'dist')
  const server = await startStaticServer(distDir)
  const browser = await launchBrowser()
  try {
    const context = await browser.newContext({
      viewport: { width: 1200, height: 630 },
      deviceScaleFactor: 1,
      colorScheme: 'light',
    })
    const page = await context.newPage()
    // Reduzierte Bewegung erzwingen, damit der animierte Hero-Crossfade
    // deterministisch auf seinem festen ersten Frame steht (stabiles OG-Bild).
    await page.emulateMedia({ reducedMotion: 'reduce' })
    await page.goto(server.url + '/', { waitUntil: 'networkidle' })
    await settle(page)
    await page.screenshot({ path: OG_OUT })
    console.log('  ✓ OG-Bild', OG_OUT)
  } finally {
    await browser.close()
    await server.close()
  }
}

if (mode === 'app' || mode === 'all') await captureApp()
if (mode === 'og' || mode === 'all') await captureOg()
