import type { Page, Route } from '@playwright/test'

// Simuliert Serverfehler (500) und Netzabbrüche per Route-Interception. Nur
// Anfragen an /api/** werden abgefangen — das Frontend selbst rendert normal,
// nur die Datenantwort fehlt.

// Vergleicht den Endpunkt-Pfad ohne führendes /api/, z. B. „get-tisch-state".
export type EndpointMatcher = string | RegExp

function matches(pathname: string, endpoint: EndpointMatcher): boolean {
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

// Bricht ab, bevor eine Antwort ankommt: das Frontend sieht einen fetch-Fehler,
// keinen HTTP-Status.
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
