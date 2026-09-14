// Version dieses Clients, zur Bauzeit eingebrannt (siehe src/version.d.ts).
export const CLIENT_VERSION = __CLIENT_VERSION__

// Muster eines echten Release-Tags. Wortgleich mit der Prüfung im Makefile
// (Target `prod-up` gegen JOTTI_VERSION), damit es im Repo nur eine Definition
// von „echtes Release" gibt. Vorabversionen wie `v1.2.3-rc1` zählen mit.
const RELEASE_MUSTER = /^v[0-9]+\.[0-9]+\.[0-9]+([.+-].*)?$/

/**
 * Abweichung genau dann, wenn beide Seiten echte Release-Versionen sind und
 * sich unterscheiden. Alles andere schaltet den Vergleich still ab: In Dev,
 * E2E und Tests steht auf beiden Seiten der Default `dev` (oder `dev-<sha>`),
 * und ein Client gegen einen ungetaggten Server soll sich nicht als veraltet
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
