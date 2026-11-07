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

async function buche(betrag: string, kommentar = 'Wechselgeld') {
  const user = userEvent.setup()
  await user.clear(screen.getByLabelText('Betrag'))
  await user.type(screen.getByLabelText('Betrag'), betrag)
  await user.clear(screen.getByLabelText('Kommentar'))
  await user.type(screen.getByLabelText('Kommentar'), kommentar)
  await user.click(screen.getByRole('button', { name: 'Geld einlegen' }))
}

function fehlschlagDannErfolg(): FakeBackend {
  return new FakeBackend().respond(
    'admin/geldtransit-buchen',
    vi
      .fn()
      .mockImplementationOnce(() => {
        throw new BackendError(503, 'weg')
      })
      .mockReturnValue({}),
  )
}

function geldtransitIds(fake: FakeBackend): unknown[] {
  return fake
    .bodies('admin/geldtransit-buchen')
    .map((body) => (body as { geldtransitId: unknown }).geldtransitId)
}

describe('GeldtransitDialog', () => {
  it('meldet das angefangene Formular und gibt es beim Schließen frei', async () => {
    const user = userEvent.setup()
    const { rerender } = renderWithBackend(dialog(true), new FakeBackend())
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(0)

    await user.type(screen.getByLabelText('Kommentar'), 'Wechselgeld')
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(1)

    // Beim Schließen bleiben die Werte stehen, das nächste Öffnen verwirft sie
    // — der Vorgang ist damit erledigt.
    rerender(dialog(false))
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(0)
  })

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

  // Nach verlorener Antwort hätte das Backend den ersten Versuch schon gebucht:
  // gleicher Inhalt muss denselben Schlüssel tragen, geänderter einen neuen.
  it.each([
    ['gleichem Inhalt', '25', 'Wechselgeld', true],
    ['geändertem Betrag', '30', 'Wechselgeld', false],
    ['geändertem Kommentar', '25', 'Wechselgeld neu', false],
  ])(
    'behält den geldtransitId beim Retry mit %s nur bei gleichem Inhalt',
    async (_fall, betrag, kommentar, gleicherKey) => {
      const fake = fehlschlagDannErfolg()
      renderWithBackend(dialog(true), fake)

      await buche('25')
      await waitFor(() => {
        expect(geldtransitIds(fake)).toHaveLength(1)
      })
      await buche(betrag, kommentar)
      await waitFor(() => {
        expect(geldtransitIds(fake)).toHaveLength(2)
      })

      const [erster, zweiter] = geldtransitIds(fake)
      expect(zweiter === erster).toBe(gleicherKey)
    },
  )
})
