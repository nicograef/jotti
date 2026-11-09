import type { APIRequestContext } from '@playwright/test'
import { expect } from '@playwright/test'

// startseite: Muster der URL, auf die die Anmeldung je Rolle weiterleitet.
export interface Zugangsdaten {
  username: string
  password: string
  startseite: RegExp
}

export interface SeedZugangsdaten {
  admin: Zugangsdaten
  serviceleitung: Zugangsdaten
  service: Zugangsdaten
}

// Admins landen auf der Auswertung. Servicekräfte und Serviceleitung landen
// im zuletzt genutzten Arbeitsmodus (frischer Browser: Tischauswahl).
export const STARTSEITE_ADMIN = /\/admin\/auswertung$/
export const STARTSEITE_SERVICE = /\/service\/(tische|direktverkauf)$/

// Antwort von POST /api/test/reset-and-seed — keine Zugangsdaten hart kodieren.
type SeedAntwort = Record<
  keyof SeedZugangsdaten,
  Omit<Zugangsdaten, 'startseite'>
>

// Nur registriert, wenn der Stack mit JOTTI_ENABLE_TEST_API=1 läuft. Jede Spec
// ruft das als Erstes auf.
export async function resetAndSeed(
  request: APIRequestContext,
): Promise<SeedZugangsdaten> {
  const response = await request.post('/api/test/reset-and-seed')
  expect(
    response.ok(),
    `reset-and-seed muss erfolgreich sein (Status ${response.status().toString()})`,
  ).toBeTruthy()
  const antwort = (await response.json()) as SeedAntwort
  return {
    admin: { ...antwort.admin, startseite: STARTSEITE_ADMIN },
    serviceleitung: {
      ...antwort.serviceleitung,
      startseite: STARTSEITE_SERVICE,
    },
    service: { ...antwort.service, startseite: STARTSEITE_SERVICE },
  }
}
