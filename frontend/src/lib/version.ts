// This client's version, set at build time (see src/version.d.ts).
export const CLIENT_VERSION = __CLIENT_VERSION__

// A real release tag, matching the Makefile `prod-up` check on JOTTI_VERSION so the repo defines it once.
// Pre-releases like `v1.2.3-rc1` count.
const RELEASE_MUSTER = /^v[0-9]+\.[0-9]+\.[0-9]+([.+-].*)?$/

/**
 * A mismatch only when both sides are real release versions and differ. Dev and E2E carry the default
 * `dev` (or `dev-<sha>`) on both sides, and a client must not report itself outdated against an untagged server.
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
