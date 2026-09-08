import { cleanup, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { VorgangsRegisterSingleton } from '@/lib/VorgangsRegister'

import { GeldtransitDialog } from './GeldtransitDialog'
import type { GeldtransitRichtung } from './Kassensitzung'

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

const { geldtransitBuchen } = vi.hoisted(() => ({
  geldtransitBuchen:
    vi.fn<
      (
        geldtransitId: string,
        richtung: GeldtransitRichtung,
        betragCents: number,
        kommentar: string,
      ) => Promise<void>
    >(),
}))

vi.mock('./hooks', () => ({
  kasseBackend: { geldtransitBuchen },
}))

beforeEach(() => {
  VorgangsRegisterSingleton.zuruecksetzen()
  geldtransitBuchen.mockReset()
})

afterEach(() => {
  cleanup()
})

function renderDialog(open: boolean) {
  return render(
    <GeldtransitDialog
      open={open}
      onOpenChange={vi.fn()}
      richtung="einlage"
      onSuccess={vi.fn()}
    />,
  )
}

async function buche(betrag: string) {
  const user = userEvent.setup()
  await user.type(screen.getByLabelText('Betrag'), betrag)
  await user.type(screen.getByLabelText('Kommentar'), 'Wechselgeld')
  await user.click(screen.getByRole('button', { name: 'Geld einlegen' }))
}

describe('GeldtransitDialog', () => {
  it('erneuert den geldtransitId beim erneuten Öffnen nach einem Fehlversuch', async () => {
    geldtransitBuchen
      .mockRejectedValueOnce(new Error('kaputt'))
      .mockResolvedValue(undefined)
    const { rerender } = renderDialog(true)

    await buche('25')
    await waitFor(() => {
      expect(geldtransitBuchen).toHaveBeenCalledTimes(1)
    })
    const ersterKey = geldtransitBuchen.mock.calls[0][0]

    // Schließen und erneut öffnen ist ein neuer Vorgang: Das Formular startet
    // leer, mit dem alten Schlüssel verwürfe das Backend die Buchung als
    // Duplikat.
    rerender(
      <GeldtransitDialog
        open={false}
        onOpenChange={vi.fn()}
        richtung="einlage"
        onSuccess={vi.fn()}
      />,
    )
    rerender(
      <GeldtransitDialog
        open
        onOpenChange={vi.fn()}
        richtung="einlage"
        onSuccess={vi.fn()}
      />,
    )

    await buche('30')
    await waitFor(() => {
      expect(geldtransitBuchen).toHaveBeenCalledTimes(2)
    })
    expect(geldtransitBuchen.mock.calls[1][0]).not.toBe(ersterKey)
  })
})
