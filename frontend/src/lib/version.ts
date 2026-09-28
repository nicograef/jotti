// Version dieses Clients, zur Bauzeit eingebrannt (siehe src/version.d.ts).
export const CLIENT_VERSION = __CLIENT_VERSION__

// A real release tag, matching the Makefile `prod-up` check on JOTTI_VERSION so the repo defines it once.
// Pre-releases like `v1.2.3-rc1` count.
const RELEASE_MUSTER = /^v[0-9]+\.[0-9]+\.[0-9]+([.+-].*)?$/

/**
 * Abweichung genau dann, wenn beide Seiten echte Release-Versionen sind und
 * sich unterscheiden. Alles andere schaltet den Vergleich still ab: In Dev
 * und E2E steht auf beiden Seiten der Default `dev` (oder `dev-<sha>`), und
 * ein Client gegen einen ungetaggten Server soll sich nicht als veraltet
 * melden.
 */
export function istVersionsabweichung(
  clientVersion: string,
  serverVersion: string,
): boolean {
  if (
    !RELEASE_MUSTER.test(clientVersion) ||
    !RELEASE_MUSTER.test(serverVersion)
  ) {
    return false
  }
  return clientVersion !== serverVersion
}
