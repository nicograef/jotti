import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { VersionsHinweis } from './VersionsHinweis'

const versionState = vi.hoisted<{ version: string | undefined }>(() => ({
  version: undefined,
}))

vi.mock('@/hooks/use-version', () => ({
  useVersion: () => versionState.version,
}))

// Die Clientversion ist im Test der Default `dev` (define in vitest.config.ts).
vi.mock('@/lib/version', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/lib/version')>()),
  CLIENT_VERSION: 'v1.2.3',
}))

afterEach(() => {
  cleanup()
  versionState.version = undefined
})

describe('VersionsHinweis', () => {
  it('zeigt den Hinweis bei einer Abweichung zweier Releases', () => {
    versionState.version = 'v1.2.4'

    render(<VersionsHinweis />)

    expect(screen.getByRole('status')).toHaveTextContent(
      'Der Server läuft mit Version v1.2.4, diese Seite mit einer anderen. Bitte die Seite neu laden.',
    )
  })

  it('bleibt bei gleicher Version aus', () => {
    versionState.version = 'v1.2.3'

    render(<VersionsHinweis />)

    expect(screen.queryByRole('status')).toBeNull()
  })

  // Solange die erste Abfrage läuft, ist noch nichts zu melden.
  it('bleibt ohne Serverantwort aus', () => {
    render(<VersionsHinweis />)

    expect(screen.queryByRole('status')).toBeNull()
  })
})
