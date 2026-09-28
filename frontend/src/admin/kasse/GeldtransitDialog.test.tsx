import { cleanup, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { BackendError } from '@/lib/Backend'
import { VorgangsRegisterSingleton } from '@/lib/VorgangsRegister'
import { FakeBackend } from '@/test/FakeBackend'
import { renderWithBackend } from '@/test/render'

import { GeldtransitDialog } from './GeldtransitDialog'

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

beforeEach(() => {
  VorgangsRegisterSingleton.zuruecksetzen()
})

afterEach(() => {
  cleanup()
})

function dialog(open: boolean) {
  return (
    <GeldtransitDialog
      open={open}
      onOpenChange={vi.fn()}
      richtung="einlage"
      onSuccess={vi.fn()}
    />
  )
}

async function buche(betrag: string) {
  const user = userEvent.setup()
  await user.type(screen.getByLabelText('Betrag'), betrag)
  await user.type(screen.getByLabelText('Kommentar'), 'Wechselgeld')
  await user.click(screen.getByRole('button', { name: 'Geld einlegen' }))
}

function geldtransitIds(fake: FakeBackend): unknown[] {
  return fake
    .bodies('admin/geldtransit-buchen')
    .map((body) => (body as { geldtransitId: unknown }).geldtransitId)
}

describe('GeldtransitDialog', () => {
  it('erneuert den geldtransitId beim erneuten Öffnen nach einem Fehlversuch', async () => {
    const geldtransitBuchen = vi
      .fn()
      .mockImplementationOnce(() => {
        throw new BackendError(400, 'kaputt')
      })
      .mockReturnValue({})
    const fake = new FakeBackend().respond(
      'admin/geldtransit-buchen',
      geldtransitBuchen,
    )
    const { rerender } = renderWithBackend(dialog(true), fake)

    await buche('25')
    await waitFor(() => {
      expect(geldtransitIds(fake)).toHaveLength(1)
    })
    const [ersterKey] = geldtransitIds(fake)

    rerender(dialog(false))
    rerender(dialog(true))

    await buche('30')
    await waitFor(() => {
      expect(geldtransitIds(fake)).toHaveLength(2)
    })
    expect(geldtransitIds(fake)[1]).not.toBe(ersterKey)
  })
})
