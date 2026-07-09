import type { APIRequestContext } from '@playwright/test'
import { expect } from '@playwright/test'

// Zugangsdaten eines Seed-Benutzers, wie der Reset-Endpoint sie zurückgibt.
export interface Zugangsdaten {
  username: string
  password: string
}

// Antwort von POST /api/test/reset-and-seed: die deterministischen Zugangsdaten
// der Seed-Rollen. Die Suite hängt nichts hart kodiert an — sie liest die Daten
// aus der Antwort.
export interface SeedZugangsdaten {
  admin: Zugangsdaten
  serviceleitung: Zugangsdaten
  service: Zugangsdaten
}

// resetAndSeed setzt die Datenbank auf den deterministischen Demo-Zustand
// zurück und liefert die Zugangsdaten für die Anmeldung. Der Endpoint ist nur
// registriert, wenn der Stack mit JOTTI_ENABLE_TEST_API=1 läuft (E2E-Umgebung).
// Jede Spec ruft dies als Erstes auf, damit sie vom bekannten Seed-Zustand startet.
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
