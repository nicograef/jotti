// Chromium launcher for the website tooling (`screenshots.mjs`). Set
// CHROMIUM_EXECUTABLE to a chrome binary to override the pinned Playwright build.

import { chromium } from '@playwright/test'

export function launchBrowser() {
  const executablePath = process.env.CHROMIUM_EXECUTABLE
  return chromium.launch(executablePath ? { executablePath } : {})
}
