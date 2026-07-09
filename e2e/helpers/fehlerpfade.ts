import type { Page, Route } from '@playwright/test'

// Hilfsfunktionen zur Simulation von Serverfehlern (5xx) und Netzabbrüchen per
// Playwright-Route-Interception. Sie fangen ausschließlich POST-Aufrufe an die
// Backend-API (/api/<endpoint>) ab — das Frontend selbst (HTML/JS/CSS) bleibt
// unangetastet, damit die Seite normal rendert und nur die Datenantwort fehlt.

// EndpointMatcher entscheidet anhand des Endpunkt-Pfads (ohne führendes /api/),
// ob eine Anfrage abgefangen wird. So lassen sich gezielt einzelne Endpunkte
// simulieren (z. B. nur „get-tisch-state", nicht „zahlung-kassieren").
export type EndpointMatcher = string | RegExp

function matches(pathname: string, endpoint: EndpointMatcher): boolean {
  // pathname enthält das führende „/api/"; der Endpunkt-Name selbst nicht.
  const endpointPath = pathname.replace(/^\/api\//, '')
  return typeof endpoint === 'string'
    ? endpointPath === endpoint
    : endpoint.test(endpointPath)
}

async function fulfillServerError(route: Route): Promise<void> {
  await route.fulfill({
    status: 500,
    contentType: 'application/json',
    headers: { 'X-Correlation-ID': 'e2e-test-correlation-id' },
    body: JSON.stringify({ code: 'internal_server_error' }),
  })
}

// simuliereServerfehler lässt jede POST-Anfrage an einen der angegebenen
// Endpunkte mit HTTP 500 (inkl. Correlation-ID) fehlschlagen — wie ein echter
// Backend-Fehler.
export async function simuliereServerfehler(
  page: Page,
  endpoints: EndpointMatcher[],
): Promise<void> {
  await page.route('**/api/**', async (route) => {
    const url = new URL(route.request().url())
    if (endpoints.some((endpoint) => matches(url.pathname, endpoint))) {
      await fulfillServerError(route)
      return
    }
    await route.continue()
  })
}

// simuliereNetzabbruch lässt jede POST-Anfrage an einen der angegebenen
// Endpunkte abbrechen (wie ein Verbindungsabbruch, bevor eine Antwort
// ankommt) — das Frontend erhält keinen HTTP-Status, sondern einen fetch-Fehler.
export async function simuliereNetzabbruch(
  page: Page,
  endpoints: EndpointMatcher[],
): Promise<void> {
  await page.route('**/api/**', async (route) => {
    const url = new URL(route.request().url())
    if (endpoints.some((endpoint) => matches(url.pathname, endpoint))) {
      await route.abort('failed')
      return
    }
    await route.continue()
  })
}
