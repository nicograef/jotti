// CSP verification for the built website (`website/dist`).
//
// Serves the artefact behind the production CSP (see `csp-server.mjs`) and drives
// headless Chromium over the landing page and two docs pages, capturing every
// `securitypolicyviolation` DOM event. Exits non-zero on any violation, so it can
// gate each phase of the website redesign (`docs/plans/plan-website-redesign.md`).
//
// Uses Playwright from the e2e package. If the pinned Playwright build mismatches
// the preinstalled browser, set CHROMIUM_EXECUTABLE to a chrome binary.
//
// Usage: node e2e/website/csp-check.mjs [distDir]
//   distDir defaults to website/dist relative to the repo root.

import { resolve, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { launchBrowser } from './browser.mjs'
import { startStaticServer } from './csp-server.mjs'

const repoRoot = resolve(fileURLToPath(new URL('../..', import.meta.url)))
const distDir = process.argv[2] ?? join(repoRoot, 'website', 'dist')

// Landing plus two docs pages — enough to cover the Starlight shell (theme init,
// search, sidebar) under CSP without walking the whole doc tree.
const PATHS = ['/', '/docs/leitfaden/was-ist-jotti/', '/docs/leitfaden/installation/']

const COLLECT_VIOLATIONS = `
  window.__cspViolations = [];
  document.addEventListener('securitypolicyviolation', (e) => {
    window.__cspViolations.push({
      directive: e.effectiveDirective || e.violatedDirective,
      blockedURI: e.blockedURI,
      source: e.sourceFile ? e.sourceFile + ':' + e.lineNumber : '(inline)',
      sample: e.sample,
    });
  });
`

const server = await startStaticServer(distDir)
const browser = await launchBrowser()
let failed = false

try {
  for (const path of PATHS) {
    const context = await browser.newContext()
    await context.addInitScript(COLLECT_VIOLATIONS)
    const page = await context.newPage()
    await page.goto(server.url + path, { waitUntil: 'networkidle' })
    // Give hydration / deferred scripts a beat to run and trip any late CSP checks.
    await page.waitForTimeout(400)
    const violations = await page.evaluate(() => window.__cspViolations)
    if (violations.length > 0) {
      failed = true
      console.error(`\n✗ ${path} — ${violations.length} CSP violation(s):`)
      for (const v of violations) {
        console.error(`    [${v.directive}] blocked=${v.blockedURI} @ ${v.source}${v.sample ? ` sample="${v.sample}"` : ''}`)
      }
    } else {
      console.log(`✓ ${path} — no CSP violations`)
    }
    await context.close()
  }
} finally {
  await browser.close()
  await server.close()
}

if (failed) {
  console.error('\nCSP verification FAILED.')
  process.exit(1)
}
console.log('\nCSP verification passed.')
