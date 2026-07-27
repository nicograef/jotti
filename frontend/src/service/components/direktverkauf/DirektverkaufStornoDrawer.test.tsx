import { cleanup, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type {
  DirektverkaufHistorieEintrag,
  DirektverkaufStornieren,
} from '../../direktverkauf/Direktverkauf'
import { DirektverkaufStornoDrawer } from './DirektverkaufStornoDrawer'

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

afterEach(() => {
  cleanup()
})

const positionId = '22222222-2222-2222-2222-222222222222'

const position = {
  positionId,
  varianteId: 1,
  produktName: 'Cola',
  varianteName: '0,5l',
  kategorie: 'getraenk' as const,
  steuersatz: 'regel' as const,
  einzelpreisCents: 500,
  menge: 2,
}

const verkauf: DirektverkaufHistorieEintrag = {
  verkaufId: '11111111-1111-1111-1111-111111111111',
  userName: 'Anna',
  getaetigtAm: '2026-06-08T10:00:00Z',
  positionen: [position],
  gesamtbetragCents: 1000,
  kommentar: '',
  offenePositionen: [position],
  gesamtStorniertCents: 0,
  stornierungen: [],
}

const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/

function renderDrawer(
  direktverkaufStornieren: (s: DirektverkaufStornieren) => Promise<void>,
) {
  render(
    <DirektverkaufStornoDrawer
      backend={{ direktverkaufStornieren }}
      verkauf={verkauf}
      onClose={vi.fn()}
      onStorniert={vi.fn()}
    />,
  )
}

describe('DirektverkaufStornoDrawer', () => {
  it('storniert mit genau einem Backend-Call samt Vorgangs-Schlüssel', async () => {
    const user = userEvent.setup()
    const direktverkaufStornieren = vi
      .fn<(s: DirektverkaufStornieren) => Promise<void>>()
      .mockResolvedValue(undefined)
    renderDrawer(direktverkaufStornieren)

    await user.click(screen.getByRole('button', { name: /hinzufügen/ }))
    await user.type(
      screen.getByPlaceholderText('Kommentar (erforderlich)'),
      'Rückgabe',
    )
    await user.click(
      screen.getByRole('button', { name: 'Stornierung erteilen' }),
    )

    await waitFor(() => {
      expect(direktverkaufStornieren).toHaveBeenCalledTimes(1)
    })
    expect(direktverkaufStornieren).toHaveBeenCalledWith({
      // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
      vorgangId: expect.stringMatching(UUID),
      verkaufId: verkauf.verkaufId,
      positionen: [{ positionId, menge: 1 }],
      kommentar: 'Rückgabe',
    })
  })

  it('behält die vorgangId über einen Retry und wechselt sie beim neuen Vorgang', async () => {
    const user = userEvent.setup()
    const direktverkaufStornieren = vi
      .fn<(s: DirektverkaufStornieren) => Promise<void>>()
      .mockRejectedValueOnce(new Error('kaputt'))
      .mockResolvedValue(undefined)
    renderDrawer(direktverkaufStornieren)

    const erteilen = () =>
      screen.getByRole('button', { name: 'Stornierung erteilen' })

    await user.click(screen.getByRole('button', { name: /hinzufügen/ }))
    await user.type(
      screen.getByPlaceholderText('Kommentar (erforderlich)'),
      'Rückgabe',
    )

    await user.click(erteilen())
    await waitFor(() => {
      expect(direktverkaufStornieren).toHaveBeenCalledTimes(1)
    })
    const ersterKey = direktverkaufStornieren.mock.calls[0][0].vorgangId
    expect(ersterKey).toMatch(UUID)

    // Wiederholversuch desselben Vorgangs: derselbe Schlüssel.
    await user.click(erteilen())
    await waitFor(() => {
      expect(direktverkaufStornieren).toHaveBeenCalledTimes(2)
    })
    expect(direktverkaufStornieren.mock.calls[1][0].vorgangId).toBe(ersterKey)

    // Neuer logischer Vorgang: Auswahl leeren und neu füllen.
    await user.click(screen.getByRole('button', { name: /verringern/ }))
    await user.click(screen.getByRole('button', { name: /hinzufügen/ }))
    await user.click(erteilen())
    await waitFor(() => {
      expect(direktverkaufStornieren).toHaveBeenCalledTimes(3)
    })
    expect(direktverkaufStornieren.mock.calls[2][0].vorgangId).not.toBe(
      ersterKey,
    )
  })

  it('wechselt die vorgangId, wenn die Auswahl nach einem Fehlversuch wächst', async () => {
    const user = userEvent.setup()
    const direktverkaufStornieren = vi
      .fn<(s: DirektverkaufStornieren) => Promise<void>>()
      .mockRejectedValueOnce(new Error('kaputt'))
      .mockResolvedValue(undefined)
    renderDrawer(direktverkaufStornieren)

    const erteilen = () =>
      screen.getByRole('button', { name: 'Stornierung erteilen' })

    await user.click(screen.getByRole('button', { name: /hinzufügen/ }))
    await user.type(
      screen.getByPlaceholderText('Kommentar (erforderlich)'),
      'Rückgabe',
    )
    await user.click(erteilen())
    await waitFor(() => {
      expect(direktverkaufStornieren).toHaveBeenCalledTimes(1)
    })
    const ersterKey = direktverkaufStornieren.mock.calls[0][0].vorgangId

    // Geänderte Nutzdaten nach dem Fehlversuch (Menge 1 → 2): neuer Vorgang,
    // neuer Schlüssel — der Server prüft die geänderte Auswahl regulär.
    await user.click(screen.getByRole('button', { name: /hinzufügen/ }))
    await user.click(erteilen())
    await waitFor(() => {
      expect(direktverkaufStornieren).toHaveBeenCalledTimes(2)
    })
    const zweiterAufruf = direktverkaufStornieren.mock.calls[1][0]
    expect(zweiterAufruf.vorgangId).not.toBe(ersterKey)
    expect(zweiterAufruf.positionen).toEqual([{ positionId, menge: 2 }])
  })
})
