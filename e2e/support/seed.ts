import type { APIRequestContext } from '@playwright/test'
import { expect } from '@playwright/test'

export interface Zugangsdaten {
  username: string
  password: string
}

// Antwort von POST /api/test/reset-and-seed — keine Zugangsdaten hart kodieren.
export interface SeedZugangsdaten {
  admin: Zugangsdaten
  serviceleitung: Zugangsdaten
  service: Zugangsdaten
}

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
  return (await response.json()) as SeedZugangsdaten
}
