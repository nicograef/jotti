import { cleanup, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { FakeBackend } from '@/test/FakeBackend'
import { renderWithBackend } from '@/test/render'

import { TSEKonfigurationSection } from './TSEKonfigurationSection'

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

afterEach(() => {
  cleanup()
})

function backend(): FakeBackend {
  return new FakeBackend()
    .respond('admin/get-tse-konfiguration', {
      apiKeyGesetzt: true,
      apiSecretGesetzt: true,
      tssId: 'tss-1',
      clientId: 'client-1',
      istKonfiguriert: true,
    })
    .respond('admin/update-tse-konfiguration', {})
}

describe('TSEKonfigurationSection', () => {
  it('leert die Zugangsdaten erst nach Bestätigung', async () => {
    const fake = backend()
    const user = userEvent.setup()
    renderWithBackend(<TSEKonfigurationSection />, fake)

    await user.click(
      await screen.findByRole('button', { name: 'Alle Felder leeren' }),
    )
    await user.click(screen.getByRole('button', { name: 'Abbrechen' }))
    expect(fake.bodies('admin/update-tse-konfiguration')).toEqual([])

    await user.click(screen.getByRole('button', { name: 'Alle Felder leeren' }))
    await user.click(screen.getByRole('button', { name: 'Leeren' }))
    await waitFor(() => {
      expect(fake.bodies('admin/update-tse-konfiguration')).toEqual([
        { apiKey: '', apiSecret: '', tssId: '', clientId: '' },
      ])
    })
  })
})
