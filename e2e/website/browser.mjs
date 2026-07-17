// Shared Chromium launcher for the website-tooling scripts (`csp-check.mjs`,
// `screenshots.mjs`). Falls back to the preinstalled browser build if the
// pinned Playwright build mismatches it.
//
// If the pinned Playwright build mismatches the preinstalled browser, set
// CHROMIUM_EXECUTABLE to a chrome binary.

import { chromium } from '@playwright/test'

export async function launchBrowser() {
  const executablePath = process.env.CHROMIUM_EXECUTABLE
  try {
    return await chromium.launch(executablePath ? { executablePath } : {})
  } catch (err) {
    // Fall back to the preinstalled browser build if the pinned one is missing.
    const fallback = '/opt/pw-browsers/chromium-1194/chrome-linux/chrome'
    if (!executablePath) return await chromium.launch({ executablePath: fallback })
    throw err
  }
}
