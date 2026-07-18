// App-Screenshots und OG-Bild für die Marketing-Website (`website/`).
//
// Zwei Modi (Standard: beide nacheinander):
//   app  — fährt gegen den e2e-Stack (JOTTI_ENABLE_TEST_API=1), setzt den Seed
//          zurück, meldet sich an und nimmt jedes Website-Motiv deterministisch
//          in Hell UND Dunkel auf. Ziel: `website/src/assets/screenshots/`.
//   og   — baut die Website (`make website-build`), serviert `dist/` hinter
//          demselben lokalen Static-Server wie die CSP-Verifikation
//          (`csp-server.mjs`) und nimmt den neuen Hero (hell) als 1200×630-OG-
//          Bild auf. Ziel: `website/src/assets/og-startseite.png`.
//
// Die App folgt der Systempräferenz (Theme-Default „system", siehe
// frontend `theme-provider.tsx`): Playwrights `emulateMedia({ colorScheme })`
// kippt `data-theme` auf `<html>` — ein eigener Theme-Schalter-State ist NICHT
// nötig (offene Frage aus Phase 9 verifiziert).
//
// BASE-URL-agnostisch über E2E_BASE_URL (wie die e2e-Suite, siehe
// `e2e/playwright.config.ts`): Default ist der Compose-Stack auf
// http://localhost:8080 (E2E_HTTP_PORT-Default). Beispiel gegen einen anderen
// Port:  E2E_BASE_URL=http://localhost:8081 node e2e/website/screenshots.mjs
//
// Seed- und Login-Logik werden aus der e2e-Suite wiederverwendet
// (`support/seed.ts`, `support/anmelden.ts`, `support/servicekraft.ts`); daher
// läuft das Skript mit `node --experimental-strip-types` (siehe Make-Target
// `website-screenshots`).
//
// Falls der gepinnte Playwright-Build den vorinstallierten Browser verfehlt,
// zeigt CHROMIUM_EXECUTABLE auf ein Chrome-Binary.

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
  process.env.SHOT_OUT ?? join(repoRoot, 'website', 'src', 'assets', 'screenshots')
const OG_OUT =
  process.env.OG_OUT ?? join(repoRoot, 'website', 'src', 'assets', 'og-startseite.png')

const mode = process.argv[2] ?? 'all'

// settle wartet auf Fonts und einen Repaint nach dem Theme-Wechsel, damit die
// Aufnahme nicht mitten im Übergang entsteht. Zusätzlich wird der Fokus vom
// aktiven Element genommen, damit kein Fokus-Ring auf autofokussierten
// Dialog-Feldern in die Marketing-Aufnahme gebrannt wird (harmlos, wenn nichts
// fokussiert ist).
async function settle(page) {
  await page.evaluate(() => document.fonts.ready)
  await page.evaluate(() => document.activeElement?.blur?.())
  await page.waitForTimeout(450)
}

// captureLightDark nimmt den aktuellen Zustand einer Seite in Hell und Dunkel
// auf (gleicher DOM-Zustand, nur `prefers-color-scheme` gekippt).
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

// Querformat-Tablet (≥ 1024 px, eigener deviceScaleFactor): löst das
// zweispaltige Layout ab `lg` aus — den Service-Split-Screen (ADR 08) für den
// Direktverkauf und das Sidebar-Layout der Admin-Motive. Damit zeigen `produkte`
// (Tablet) und `produktverwaltung` (Desktop) bewusst dasselbe Design auf zwei
// Geräteklassen.
const tabletLandscape = { viewport: { width: 1194, height: 834 }, deviceScaleFactor: 2 }

async function captureApp() {
  mkdirSync(SHOT_OUT, { recursive: true })
  const browser = await launchBrowser()
  try {
    const apiContext = await browser.newContext({ baseURL: BASE })
    const zugangsdaten = await resetAndSeed(apiContext.request)
    await apiContext.close()

    // ---- Service-Motive (Handy, Servicekraft „maria") ----
    const phone = await browser.newContext({ baseURL: BASE, ...devices['Pixel 7'] })
    const p = await login(phone, zugangsdaten.service)

    // Tischübersicht („Meine Tische")
    await p.goto('/service/tische')
    await p.getByText('Meine Tische').first().waitFor()
    await p.getByText('Noch offen', { exact: false }).first().waitFor()
    await captureLightDark(p, 'tischuebersicht')

    // Bestellansicht (Hero): lokaler Warenkorb, noch nicht abgeschickt.
    await oeffneTisch(p, 'Tisch 1')
    await p.getByRole('tab', { name: 'Bestellen' }).click()
    await waehleVariante(p, 'Bratwurst', 'Normal', 2)
    await waehleVariante(p, 'Pommes', 'Groß', 2)
    await p.getByRole('button', { name: /Bestellung überprüfen/ }).waitFor()
    await captureLightDark(p, 'bestellansicht')

    // Zahlung: Kassieren-Drawer eines Tisches mit offenen Positionen.
    await oeffneTisch(p, 'Tisch 2')
    await p.getByRole('tab', { name: 'Kassieren' }).click()
    const vonAnderen = p.getByRole('button', { name: /^Von anderen ·/ })
    if (await vonAnderen.isVisible().catch(() => false)) await vonAnderen.click()
    await waehleAlleVollAus(p)
    await p.getByRole('button', { name: /Kassieren/ }).click()
    const zahlungDrawer = p.getByRole('dialog')
    await zahlungDrawer.getByText(/€/).first().waitFor()
    await captureLightDark(p, 'zahlung')
    await p.keyboard.press('Escape')
    await phone.close()

    // ---- Direktverkauf (Querformat-Tablet, Split-Screen, Servicekraft „maria") ----
    // Ab lg (1024 px) rendert der Service-Bereich zweispaltig (ADR 08): links die
    // Produktauswahl, rechts die dauerhaft sichtbare Beleg-/Kassieren-Spalte.
    const tabletService = await browser.newContext({ baseURL: BASE, ...tabletLandscape })
    const dv = await login(tabletService, zugangsdaten.service)
    await dv.goto('/service/direktverkauf')
    await dv.getByRole('tab', { name: 'Verkaufen' }).waitFor()
    const dvZeile = zeileMit(dv, 'Currywurst', 'Variante hinzufügen')
    await dvZeile.getByRole('button', { name: 'Variante hinzufügen' }).click()
    await dvZeile.getByRole('button', { name: 'Variante hinzufügen' }).click()
    await dv.getByRole('button', { name: 'Verkauf abschließen' }).waitFor()
    await captureLightDark(dv, 'direktverkauf')
    await tabletService.close()

    // ---- Stornierung (Handy, Serviceleitung „felix") ----
    const phoneSL = await browser.newContext({ baseURL: BASE, ...devices['Pixel 7'] })
    const sl = await login(phoneSL, zugangsdaten.serviceleitung)
    // „Tisch 15" ist im Sonntags-Drehbuch unbenutzt (wie in der Storno-Spec).
    await oeffneTisch(sl, 'Tisch 15')
    await bestellePosition(sl, 'Pommes', 'Klein')
    await sl.getByRole('tab', { name: 'Historie' }).click()
    const detail = await oeffneHistorienDetail(sl, /Bestellung.*\+2,50/)
    await detail.getByRole('button', { name: /Stornieren…/ }).click()
    const stornoDrawer = sl.getByRole('dialog')
    await zeileMit(stornoDrawer, 'Pommes Klein', 'hinzufügen')
      .getByRole('button', { name: /hinzufügen/ })
      .click()
    await stornoDrawer
      .getByPlaceholder('Kommentar (erforderlich)')
      .fill('Falsch bestellt, storniert')
    // Den „Bestellung wurde aufgenommen."-Toast aus dem Bestellschritt abklingen
    // lassen, damit er nicht über der Storno-Aufnahme hängt.
    await sl
      .getByText('Bestellung wurde aufgenommen.')
      .first()
      .waitFor({ state: 'hidden' })
      .catch(() => {})
    await captureLightDark(sl, 'stornierung')
    await phoneSL.close()

    // ---- Admin-Motive (Querformat-Tablet, Admin „thomas") ----
    const tabletAdmin = await browser.newContext({ baseURL: BASE, ...tabletLandscape })
    const pa = await login(tabletAdmin, zugangsdaten.admin)

    await pa.goto('/admin/produkte')
    await pa.getByText('Produkte & Preise').first().waitFor()
    await captureLightDark(pa, 'produkte')

    await pa.goto('/admin/benutzer')
    await pa.getByText('Helfer & Zugänge').first().waitFor()
    await captureLightDark(pa, 'benutzer')

    // Geldtransit: „Geld einlegen"-Formular der offenen Kassensitzung mit
    // ausgefülltem, gültigem Beleg. Der Betrag ist ein Pflichtfeld: das
    // Ausfüllen zeigt einen realistischen Vorgang und verhindert, dass die
    // Blur-Validierung (settle) einen Fehlerzustand in die Aufnahme brennt.
    await pa.goto('/admin/kasse')
    await pa.getByRole('button', { name: 'Geld einlegen' }).click()
    await pa.getByRole('dialog').waitFor()
    await pa.getByLabel('Betrag').fill('50,00')
    await pa.getByLabel('Kommentar').fill('Wechselgeld Nachschub')
    await captureLightDark(pa, 'geldtransit')
    await pa.keyboard.press('Escape')
    await tabletAdmin.close()

    // ---- Desktop-Motive (Browser-Rahmen, Admin „thomas") ----
    const desktop = await browser.newContext({
      baseURL: BASE,
      viewport: { width: 1360, height: 850 },
      deviceScaleFactor: 2,
    })
    const pd = await login(desktop, zugangsdaten.admin)
    await pd.goto('/admin/produkte')
    await pd.getByText('Produkte & Preise').first().waitFor()
    await captureLightDark(pd, 'produktverwaltung')

    // Auswertung: historischer Tagesbericht (deterministisch Nr. 2 gewählt).
    await pd.goto('/admin/kassenberichte')
    await pd.getByText('Berichte & Export').first().waitFor()
    await pd.getByText('Sommerfest 26 Samstag').first().click()
    await pd.getByText('Umsatz nach Steuersatz').first().waitFor()
    await captureLightDark(pd, 'auswertung')
    await desktop.close()
  } finally {
    await browser.close()
  }
  console.log('App-Screenshots fertig →', SHOT_OUT)
}

async function captureOg() {
  // Website bauen, damit dist/ den neuen Hero mit den echten Screenshots enthält.
  console.log('Baue Website (make website-build) …')
  const build = spawnSync('make', ['website-build'], { cwd: repoRoot, stdio: 'inherit' })
  if (build.status !== 0) throw new Error('make website-build fehlgeschlagen')

  const { startStaticServer } = await import('./csp-server.mjs')
  const distDir = join(repoRoot, 'website', 'dist')
  const server = await startStaticServer(distDir)
  const browser = await launchBrowser()
  try {
    // OG-Format ist fix 1200×630; Hero hell aufnehmen.
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
